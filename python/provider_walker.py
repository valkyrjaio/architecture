"""
Walks a Python implementation of ComponentProviderContract and extracts the
classes returned from each get_*_providers method, resolving each reference
back to the file that defines it.

The Python port is assumed to use:

    from myapp.foo import FooProvider
    from myapp.bar import BarProvider

    class MyComponentProvider(ComponentProviderContract):
        @staticmethod
        def get_component_providers(app: Application) -> list[type[ComponentProviderContract]]:
            return [FooProvider, BarProvider]
        # ...

We extract the identifiers from the returned list literal, match them back to
their import statements at the top of the file, and resolve each module to a
file path.
"""

from __future__ import annotations

import ast
import importlib.util
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterable

METHODS: dict[str, str] = {
    "component_providers": "get_component_providers",
    "container_providers": "get_container_providers",
    "event_providers":     "get_event_providers",
    "cli_providers":       "get_cli_providers",
    "http_providers":      "get_http_providers",
}


@dataclass
class ProviderRef:
    name: str            # local name as used in source
    module: str | None   # dotted module path it was imported from
    file: str | None     # resolved absolute path, or None


@dataclass
class AnalysisResult:
    class_name: str | None = None
    providers: dict[str, list[ProviderRef]] = field(default_factory=dict)


class ProviderWalker:
    """
    Static analyzer for ComponentProviderContract implementations.

    `search_paths` should be the list of roots that make up the project's
    Python path, in the order they'd appear in sys.path. For a typical setup
    this is just the project root containing the top-level package.
    """

    def __init__(self, search_paths: Iterable[Path | str]) -> None:
        self._search_paths = [str(Path(p).resolve()) for p in search_paths]

    def analyze(self, file_path: str | Path) -> AnalysisResult:
        file_path = Path(file_path)
        source = file_path.read_text(encoding="utf-8")
        tree = ast.parse(source, filename=str(file_path))

        # Build a map of local-name -> source module, e.g. "FooProvider"
        # -> "myapp.foo". Handles both `from pkg import Name` (possibly
        # aliased) and `import pkg.mod as alias`.
        import_map = self._build_import_map(tree, file_path)

        # Locate the first class in the file. For providers this is canonical;
        # if you have multi-class files this becomes a configuration concern.
        class_def = next(
            (n for n in tree.body if isinstance(n, ast.ClassDef)),
            None,
        )

        result = AnalysisResult(providers={k: [] for k in METHODS})
        if class_def is None:
            return result
        result.class_name = class_def.name

        for key, method_name in METHODS.items():
            method = next(
                (
                    n for n in class_def.body
                    if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef))
                       and n.name == method_name
                ),
                None,
            )
            if method is None:
                continue
            result.providers[key] = self._extract_refs(method, import_map)

        return result

    def _build_import_map(self, tree: ast.Module, file_path: Path) -> dict[str, str]:
        """
        Map every locally-bound name introduced by an import statement to the
        module it came from, so we can resolve identifiers in return lists
        without needing a full symbol table.
        """
        mapping: dict[str, str] = {}
        for node in tree.body:
            if isinstance(node, ast.ImportFrom):
                module = self._resolve_from_module(node, file_path)
                if module is None:
                    continue
                for alias in node.names:
                    local = alias.asname or alias.name
                    mapping[local] = module
            elif isinstance(node, ast.Import):
                for alias in node.names:
                    # `import a.b.c` binds `a` unless aliased; `import a.b as x`
                    # binds `x` to the full dotted path.
                    if alias.asname:
                        mapping[alias.asname] = alias.name
                    else:
                        mapping[alias.name.split(".")[0]] = alias.name
        return mapping

    def _resolve_from_module(self, node: ast.ImportFrom, file_path: Path) -> str | None:
        """
        Turn a `from X import Y` statement into the absolute module path X.
        Handles relative imports by walking up from the current file's package.
        """
        if node.level == 0:
            return node.module
        # Relative import — resolve against the current file's package.
        current_pkg = self._package_for_file(file_path)
        if current_pkg is None:
            return None
        parts = current_pkg.split(".")
        if node.level > len(parts):
            return None
        base = ".".join(parts[: len(parts) - node.level + 1])
        # level=1 means "same package", so we slice off level-1 components.
        if node.level > 1:
            base = ".".join(parts[: len(parts) - (node.level - 1)])
        if not base:
            return node.module
        return f"{base}.{node.module}" if node.module else base

    def _package_for_file(self, file_path: Path) -> str | None:
        """
        Derive the dotted package path for a file based on configured search
        paths. This mirrors what importlib would compute at runtime but does
        so without executing any code.
        """
        file_abs = str(file_path.resolve())
        for root in self._search_paths:
            if file_abs.startswith(root + "/"):
                rel = file_abs[len(root) + 1:]
                parts = rel.split("/")
                if parts[-1].endswith(".py"):
                    parts[-1] = parts[-1][:-3]
                if parts[-1] == "__init__":
                    parts = parts[:-1]
                # The package is everything up to (but not including) the module.
                return ".".join(parts[:-1]) if parts else None
        return None

    def _extract_refs(
            self,
            method: ast.FunctionDef | ast.AsyncFunctionDef,
            import_map: dict[str, str],
    ) -> list[ProviderRef]:
        """
        Walk the method body looking for `return [Name, Name, ...]` and pull
        each element's identifier. Anything that isn't a bare Name in the list
        (e.g. a call, conditional, or computed value) is intentionally skipped.
        """
        refs: list[ProviderRef] = []
        for node in ast.walk(method):
            if not isinstance(node, ast.Return):
                continue
            if not isinstance(node.value, (ast.List, ast.Tuple)):
                continue
            for elt in node.value.elts:
                if not isinstance(elt, ast.Name):
                    continue
                name = elt.id
                module = import_map.get(name)
                refs.append(ProviderRef(
                    name=name,
                    module=module,
                    file=self._resolve_module_to_file(module, name) if module else None,
                ))
        return refs

    def _resolve_module_to_file(self, module: str, symbol: str) -> str | None:
        """
        Resolve a dotted module path to a file. We use importlib.util.find_spec
        temporarily, but we adjust sys.path to only include our configured
        roots to avoid picking up unrelated installations.

        find_spec does not execute the module (it only loads the finder/loader
        pair), which is what we want for static-ish analysis.
        """
        saved_path = sys.path[:]
        try:
            sys.path = self._search_paths + [p for p in saved_path if p not in self._search_paths]
            spec = importlib.util.find_spec(module)
            if spec is None or spec.origin is None or spec.origin == "built-in":
                return None
            return str(Path(spec.origin).resolve())
        except (ImportError, ValueError, ModuleNotFoundError):
            return None
        finally:
            sys.path = saved_path
