# Provider Walker

Static analysis tooling for walking `ComponentProviderContract` implementations across all Valkyrja ports. Each walker
parses a provider source file, extracts the references returned from the five `get*Providers` methods, and resolves each
reference to a file path on disk. The resolved paths can then be fed back into the walker to traverse the provider graph
recursively.

## Contract

All five walkers expose a single `analyze(filePath)` entry point that returns a structure with the same five keys:

- `componentProviders`
- `containerProviders`
- `eventProviders`
- `cliProviders`
- `httpProviders`

Each key holds a list of `{reference, file}` pairs, where `reference` is the fully-qualified class name / import path /
dotted module path as appropriate to the language, and `file` is the absolute path to the file that declares it (or
`null` if resolution failed).

This shape is intentionally uniform so that downstream tooling — the dependency graph, Völundr's CI/CD automation,
pre-generated data caches, validation passes — can consume output from any port without branching on language.

## Framework invariants the walkers depend on

The walkers are purely static. They do not execute any user code. This only works because the
`ComponentProviderContract` already commits to a statically analyzable shape, and each port must preserve that
commitment.

1. Each `get*Providers` method must return a literal list / array / slice expression directly from a `return` statement.
   No conditionals, no computed values, no building the list across multiple statements.
2. Every element of that literal must be a direct class reference (PHP `::class`, Java `.class`, Go typed nil or
   zero-value, TS/Python bare identifier). No calls, no lookups, no indirection.
3. Every referenced class must be imported at the top of the file (or declared in the same file / same package, where
   the language allows). No dynamic imports, no string-based class loading.

If a port ever needs runtime logic to decide what it provides, that logic belongs in a separate method — the
`get*Providers` methods must stay boring. This is the same discipline that makes the PHP pre-generated data cache
classes possible, so it costs nothing extra to maintain.

## Per-language notes

### PHP — `nikic/php-parser` + Composer `ClassLoader`

Uses `NodeFinder` to locate the class/interface declaration and its `get*Providers` methods, then walks return
expressions for `ClassConstFetch` nodes with `::class`. The `NameResolver` visitor is run first so every `Name` node
carries its resolved FQCN, which means `use` aliasing and relative names are handled automatically.

File resolution delegates to `Composer\Autoload\ClassLoader::findFile()`, which respects PSR-4, PSR-0, and classmap
entries as configured in the project's `composer.json`. Boot the project's `vendor/autoload.php` to get a configured
loader, pass it into the walker's constructor.

Minimal example — parse a file, pull the first `::class` reference out, resolve it to a file path:

```php
$loader = require __DIR__ . '/vendor/autoload.php'; // returns ClassLoader

$parser = (new ParserFactory())->createForNewestSupportedVersion();
$ast    = $parser->parse(file_get_contents($filePath));

$traverser = new NodeTraverser();
$traverser->addVisitor(new NameResolver());
$ast = $traverser->traverse($ast);

$fetch = (new NodeFinder())->findFirstInstanceOf($ast, ClassConstFetch::class);
$fqcn  = $fetch->class->getAttribute('resolvedName')->toString();

$file = $loader->findFile($fqcn); // PSR-4 / PSR-0 / classmap, all handled
```

### Java — JavaParser with `CombinedTypeSolver`

Handles both `new Class[] { A.class, B.class }` and `List.of(A.class, B.class)` / `Arrays.asList(...)` return shapes.
The `JavaSymbolSolver` resolves type references to FQNs using the configured source roots; resolution failures fall back
to the textual name with scope so analysis can still proceed on partially-complete code.

File resolution is a deterministic package-to-path mapping (`io.valkyrja.foo.Bar` → `<root>/io/valkyrja/foo/Bar.java`)
checked against each configured source root.

Minimal example — parse a file, pull the first `.class` expression out, resolve it to a file path:

```java
CombinedTypeSolver solver = new CombinedTypeSolver(
        new ReflectionTypeSolver(),
        new JavaParserTypeSolver(sourceRoot.toFile())
);
StaticJavaParser.

getParserConfiguration()
    .

setSymbolResolver(new JavaSymbolSolver(solver));

CompilationUnit cu = StaticJavaParser.parse(filePath);
ClassExpr classExpr = cu.findFirst(ClassExpr.class).orElseThrow();
String fqn = classExpr.getType().asClassOrInterfaceType()
        .resolve().asReferenceType().getQualifiedName();

Path file = sourceRoot.resolve(fqn.replace('.', '/') + ".java");
```

**Alternative worth considering for the Java port specifically:** because the port already uses an annotation processor
with JavaPoet for data class generation, you could emit a `@Provides(Foo.class, Bar.class)` annotation and walk that
instead of method bodies. Annotations are more robust to refactoring than method-body AST patterns, and the annotation
processor can validate the shape at compile time.

### Go — `go/parser` + `golang.org/x/tools/go/packages`

Loads packages with full type info (`NeedTypes | NeedTypesInfo | NeedSyntax`), which means the walker gets free, correct
resolution of module paths, vendor directories, and GOPATH — all the things you don't want to reimplement.

Go's provider pattern is expected to use typed nils to build type lists, since Go doesn't have class literals:

```go
return []ComponentProvider{
(*foo.Provider)(nil),
(*bar.Provider)(nil),
}
```

The walker also handles `pkg.Type{}` zero-value composite literals as an alternative form. File paths come from
`types.Object.Pos()` via the `FileSet`.

Minimal example — load a package, pick a named type, resolve it to a file path:

```go
cfg := &packages.Config{
Mode: packages.NeedTypes | packages.NeedTypesInfo |
packages.NeedSyntax | packages.NeedFiles,
}
pkgs, _ := packages.Load(cfg, "./...")

// Given some ast.Expr referencing a type:
tv := pkgs[0].TypesInfo.Types[expr]
named := tv.Type.(*types.Named)
obj := named.Obj()

fqn  := obj.Pkg().Path() + "." + obj.Name()
file := pkgs[0].Fset.Position(obj.Pos()).Filename
```

### TypeScript — TS Compiler API

Builds a `Program` from the project's `tsconfig.json`, which means path mappings, `baseUrl`, `node_modules` resolution,
and `.d.ts` handling all work out of the box — the compiler has already solved all of this and we ride on top.

Identifiers in the return array are resolved via `TypeChecker.getSymbolAtLocation()`, with `getAliasedSymbol()` called
for imported symbols to reach the original declaration. The source file path comes from the symbol's first declaration
node.

Minimal example — build a Program, pick an identifier, resolve it to a file path:

```typescript
const config = ts.readConfigFile(tsconfigPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(
    config.config, ts.sys, path.dirname(tsconfigPath),
);
const program = ts.createProgram({
    rootNames: parsed.fileNames,
    options: parsed.options,
});
const checker = program.getTypeChecker();

// Given some ts.Identifier node in the source:
let symbol = checker.getSymbolAtLocation(identifier)!;
if (symbol.flags & ts.SymbolFlags.Alias) {
    symbol = checker.getAliasedSymbol(symbol);
}
const file = symbol.declarations?.[0].getSourceFile().fileName;
```

### Python — `ast` + `importlib.util.find_spec`

The only port that requires real work to resolve, because Python's import system is dynamic. The walker builds an import
map from the file's top-level `Import` and `ImportFrom` nodes (handling aliases and relative imports), matches
identifiers in return lists against that map to get dotted module paths, and then uses `importlib.util.find_spec` with a
restricted `sys.path` to locate the file.

Relative imports (`from ..foo import bar`) are resolved by walking up from the current file's derived package, which is
computed from the configured search paths rather than imported at runtime.

Minimal example — parse a file, extract the first import, resolve it to a file path:

```python
import ast, importlib.util, sys
from pathlib import Path

sys.path.insert(0, str(project_root))

tree = ast.parse(Path(file_path).read_text())
imp = next(n for n in tree.body if isinstance(n, ast.ImportFrom))
module = imp.module  # e.g. "myapp.foo"

spec = importlib.util.find_spec(module)
file = Path(spec.origin).resolve() if spec and spec.origin else None
```

## Caveats

### Static-analysis edge cases

- **Conditional returns.** A method with `if (...) return [A]; return [B];` will produce the union of both branches.
  This is usually desirable — you want to know every class that could be provided — but it can mislead you into thinking
  a class is always provided when it's only sometimes provided. If you need branch-sensitive analysis, extend each
  walker to track the enclosing control flow and emit conditional edges.
- **Inheritance and traits / mixins.** The walkers look at the first class/interface in the file and only read methods
  declared on that class directly. If a port ever uses a base class or trait/mixin that provides a default
  `get*Providers` implementation, the walker will miss those providers. Fix is either to forbid the pattern (cleanest)
  or to do a second pass that follows `extends` / embeds / inheritance and merges parent method bodies.
- **Generics and wildcards.** None of the walkers interpret type parameters. A method returning
  `List<Class<? extends ComponentProviderContract>>` is read element-by-element regardless of the bound, which is fine
  for extracting references but means the walker won't catch a type-bound violation. Use the compiler for that.

### Resolution failures are silent by design

All five walkers return `file: null` (or equivalent) when they can't resolve a reference, rather than throwing. This is
deliberate — a recursive graph walk shouldn't abort because of one missing file, and a partial graph is more useful than
no graph. Callers that want strict mode should check for nulls and fail the build themselves.

### PHP — Composer must be booted

The PHP walker needs a live `ClassLoader` instance. In a CLI tool you'll typically
`require __DIR__ . '/../vendor/autoload.php'` and grab the loader it returns. If you're analyzing a project that isn't
your own, point Composer at that project's `vendor/autoload.php` instead. The walker itself is stateless per-call, so
you can reuse one instance across thousands of files.

### Java — source roots only, no bytecode

The Java walker resolves against source directories, not classpath JARs. If a provider is declared inside a dependency
JAR (unusual for framework-internal providers but possible for integration modules like `valkyrja-tomcat`), the walker
will return `file: null` for it. If you need to follow into JARs, swap the `JavaParserTypeSolver` for a
`JarTypeSolver` — the rest of the walker doesn't change.

### Go — loader cost is real

`packages.Load` with full type info is not cheap on large modules. The walker caches the loaded package set for the
lifetime of the `Walker` instance, so analyze many files with one walker rather than constructing a new walker per file.
For incremental/watch mode you'll want to rebuild the walker on file changes; the `packages` loader does not have a
stable incremental API.

### TypeScript — Program construction is expensive

Same story as Go. Build the `Program` once per analysis session and hold it. Also note that the walker resolves to
whatever file the compiler considers authoritative for a symbol, which may be a `.d.ts` file rather than a `.ts` source
file if the compiler found declarations first. If you specifically need `.ts` sources, post-process the resolved paths
to strip `.d.ts` and look for the matching `.ts`.

### Python — `find_spec` has side effects

`importlib.util.find_spec` does not execute the target module, but it does import parent packages to find submodules.
This means `__init__.py` files along the path will execute. For analyzing your own framework code this is a non-issue.
For analyzing untrusted or expensive-to-import code, replace `find_spec` with a pure filesystem walker that mirrors
PSR-4-style conventions (root + dotted path → file) — the rest of the Python walker is unchanged. Jedi's static resolver
is another option if you want correctness without giving up safety.

### Python — first class wins

The walker takes the first `ClassDef` in the file as the provider. Provider files are expected to contain one provider
class each, consistent with the PHP convention. Multi-class files require per-file configuration to pick the right
class.

### The walkers don't validate the contract

A class that implements `ComponentProviderContract` is assumed to implement it correctly. The walkers don't check that
the `getComponentProviders` method returns only `ComponentProviderContract` subtypes, that `getHttpProviders` returns
only `HttpRouteProviderContract` subtypes, etc. That validation belongs in a separate pass — it's trivial to add once
you have the resolved file paths, because at that point you can analyze each referenced file and check its declared
interfaces against the expected bucket. This is probably the natural next feature for Völundr.

## Recursion

Because each walker returns resolved file paths, recursive traversal is language-agnostic:

```
queue = [entry_provider_file]
visited = set()
graph = {}

while queue:
    file = queue.pop()
    if file in visited: continue
    visited.add(file)
    result = walker.analyze(file)
    graph[file] = result
    for bucket in result.providers.values():
        for ref in bucket:
            if ref.file and ref.file not in visited:
                queue.append(ref.file)
```

The queue-and-visited pattern is the same in every language; only the walker construction differs. A reasonable output
format for the assembled graph is a JSON map of
`{file: {className, componentProviders: [...], containerProviders: [...], ...}}`, which is cheap to produce from any
walker's result and easy to consume downstream.
