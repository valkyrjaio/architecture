# AGENTS.md — Python (Layer 2)

Per-language guide for the **Python** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file records the Python
**deltas**. PHP is the reference implementation; mirror its behavior, adapting to
Python idiom. Authoritative port detail: [`README.md`](README.md),
[`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md).

**Python 3.14+ required** for modern typing and language features.

> **Cold-start note — the lazy-import plan changed.** PEP 690 (implicit lazy
> imports) was **withdrawn** and never shipped; Python 3.14 does **not** lazy-load
> imports by default. Its successor, **PEP 810 (explicit lazy imports)**, is still
> under discussion and unshipped — treat lazy imports as a *possible future
> optimization*, not a guarantee. The container design does **not** depend on it:
> correctness rests on string-constant binding keys + lambda-wrapped values (see
> Layout & naming), which avoid eager imports in any Python version. If explicit
> lazy imports land, they would only further defer provider-module loading in the
> generated cache. Until then, eager-import cold start is a known Python weakness —
> the Go/TS ports are the escape valve for Lambda-heavy workloads.

---

## Layout & naming

- **New repos** are scaffolded from the language's `template` repo
  (`project-template-python`, in progress) — the source of truth for
  repo/file/class structure (canonical rule: §3.9).
- **Package root:** `valkyrja` (lowercase). Components map to snake_case
  subpackages (`valkyrja.container`, `valkyrja.http`, `valkyrja.event`,
  `valkyrja.cli`).
- **Namespace package (PEP 420):** the `valkyrja` root is an *implicit namespace
  package* — **no `__init__.py` at `src/valkyrja/`** in any repo, so every
  component repo contributes into the shared `valkyrja` namespace (the direct
  analog of PHP's `Valkyrja\` PSR-4 spread across repos). Sub-packages
  (`valkyrja/container/`, …) are regular packages with `__init__.py`.
- **Contracts:** `ABC` + `@abstractmethod` (Python has no `interface`). The
  **name keeps the `*Contract` suffix** (`ContainerContract`,
  `ServiceProviderContract`).
- **Binding keys:** **string constants** (required — class objects would force
  eager imports and defeat lazy loading). Per-component constants files
  (`container/container_constants.py` → `ContainerConstants.CONTAINER`, …). The
  `class_()` helper builds an FQN string from a class (needed because `class` is
  reserved).
- **Data objects:** `@dataclass(frozen=True)` (the readonly-class equivalent).
- **Handler markers:** `@handler` / `@parameter` are **metadata only** — they
  attach `_valkyrja_handler` / `_valkyrja_parameters` to the function; the
  framework reads them at bootstrap and skips them when cache is loaded.
- Every file starts with the license header.

### Exceptions

Keep the **`*Exception` name parity** even though the native bases differ:
`ValkyrjaThrowable(BaseException, ABC)`, `ValkyrjaRuntimeException(RuntimeError,
ABC)`, `ValkyrjaInvalidArgumentException(ValueError, ABC)` → abstract
`Component*` → concrete `Component<Specific>Exception`. Detail:
[`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) applies with
**snake_case** module/package segments (`contract`, `provider`, `factory`,
`constant`, `exception`, `throwable`, `abstract`, `enum`, `type`, `model`,
`entity`, `security`). Name suffixes match §4 (`*Contract`, `*ServiceProvider`,
`*Exception`, …).

Python nuances:

- **No hard-keyword collisions**, but segment and identifier names can shadow the
  stdlib and builtins — `enum` and `type` shadow the stdlib module / builtin (Go
  hit the same `type` collision), and `type` / `id` are builtins to avoid as
  identifiers. Safe only under absolute imports; keep aware. `class` is reserved →
  use the `class_()` helper for class-reference strings.
- **`interface` → `ABC`**, **abstract class → `ABC` + `@abstractmethod`**, **enum
  → `enum.Enum`**. **No traits** — use mixins/multiple inheritance; no `trait`
  segment.
- **No strong architecture linter** (import-linter only checks the import graph) —
  enforce the taxonomy in **review**, as TypeScript does. Python's stdlib `ast`
  leaves room to add mechanical checks later: a `python/ci/pytest` package with
  `ast`-based rules could enforce the name↔segment agreement import-linter can't —
  the analog of PHP's `php/ci/phparkitect` `Rules`. Future enhancement, not
  required now.

---

## Tests

- **Location:** `tests/{unit,functional,fixtures,abstract}` parallel to `src/`
  (snake_case spelling of the cross-language §6 taxonomy); unit paths mirror
  `src/`.
- **Naming:** files `test_*.py`; test classes/functions per pytest convention;
  reusable doubles live in `tests/fixtures/` — production-shaped, never named like
  tests.
- **Framework:** `pytest`. PHPUnit → pytest mapping:

  | PHPUnit              | pytest                              |
  |----------------------|-------------------------------------|
  | `assertSame`         | `assert a == b` (`is` for identity) |
  | `assertTrue/False`   | `assert x` / `assert not x`         |
  | `assertInstanceOf`   | `assert isinstance(...)`            |
  | `expectException`    | `with pytest.raises(...)`           |
  | `@dataProvider`      | `@pytest.mark.parametrize`          |
  | `setUp` / `tearDown` | fixtures / `yield` fixtures         |

- **Coverage:** `coverage.py` / `pytest-cov` — **100% (line and branch), never
  dropping**; every code branch has a test.

---

## Build & CI tools

- **Build tool (`sindri-python`):** stdlib `ast` + `inspect` only (no external AST
  lib); reads `@handler`/`@parameter` metadata and the provider tree, generates
  the four cache data files. Dev-only; the framework has zero AST deps.
- **CI:** `import-linter` (module boundaries) · **Ruff** (lint + format) ·
  **mypy** (`strict`) · **Bandit** (security) · **pytest** (+ coverage).
- **Isolation & how to run:** each tool lives under its own `.github/ci/<tool>/`
  with its own `pyproject.toml` + `uv.lock` — the **uv** analog of PHP's per-tool
  `composer.json`/`vendor` and TS's `package.json`/`node_modules` (uv maps to those
  toolchains 1:1). Run a tool via `uv run --directory .github/ci/<tool> …`; the
  root `pyproject.toml` exposes shortcut tasks — check it first for exact names.

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate: `ruff format` + `ruff check` → `mypy` → `import-linter` → `bandit`
→ `pytest --cov` (100%).

---

## Python-specific notes

- **Namespace packages (PEP 420):** no `__init__.py` at `src/valkyrja/` — every
  repo shares the `valkyrja` namespace (see Layout & naming).
- **Source shipping is free.** Unlike Java (`-sources.jar`) and TypeScript (ship
  `.ts`), Python ships `.py` by default, so the cache-optional runtime always has
  source available for provider-tree traversal.
- **Entry adapters / deployment** are the Python analog of Java's `entry/*`
  (jetty/netty/tomcat): each `entry/*` wraps an **ASGI** server — **Uvicorn /
  Hypercorn / Granian** (Rust-based) — as the worker deployment model. Python
  additionally supports a **CGI / Lambda** mode (cache optional in dev, required in
  prod). One app, two entry points: `valkyrja.worker.run(app)` (ASGI) and
  `valkyrja.cgi.run(app)` (CGI). The GIL limits true thread parallelism — async
  ASGI is the idiomatic concurrency model.
- **`sindri` (build tool)** resolves classes to source with `inspect.getfile()`,
  parses via stdlib `ast`, and emits the four cache data classes through
  `string.Template`. Dev-only; the framework has zero AST/build deps.

---

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md), and the Python section of [`../CI_TOOLS.md`](../CI_TOOLS.md).
