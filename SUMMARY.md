# Session Summary

This document summarizes the architectural decisions and documentation produced during this planning session for
Valkyrja's cross-language ports.

---

## What We Accomplished

### 1. Language Port List Finalized

Established the priority-ordered port list:

1. **PHP** — original, reference implementation
2. **Java** — in progress
3. **Go** — proof of concept started
4. **Python** — planned
5. **TypeScript / Node.js** — planned

Future ports (lower priority): Kotlin (nearly free from Java), Scala, Rust, Ruby.

C++, Play (a framework not a language), and C# were evaluated and either dropped or deferred.

---

### 2. Exception / Throwable Naming Convention

Established a complete, recursive naming convention:

- Framework base: `Valkyrja*` (abstract, extends language root)
- Component: `ComponentName*` (abstract, always present)
- Unique subcomponent: `SubComponent*`
- Shared subcomponent: `ParentComponentSubComponent*`
- Recursive rule: prepend parent names until the name is unique across the entire framework

Every base and categorical exception is abstract. Every component always ships `ComponentRuntimeException` and
`ComponentInvalidArgumentException` even if unused. Only concrete specific exceptions are thrown.

**Java note:** `ValkyrjaInvalidArgumentException` extends `IllegalArgumentException` (not `InvalidArgumentException`
which doesn't exist in Java) while using the parity name for cross-port consistency.

**Produced:** `THROWABLES.md`

---

### 3. Container Bindings — Closures and Constants

Established that:

- All container bindings use explicit closure factories across all languages
- PHP and Java use `::class` / `.class` for type-safety but not dynamic dispatch
- Go, Python, and TypeScript use string constants (no `::class` equivalent)
- Per-component constants files — never a single central constants file
- String format: `io.valkyrja.{component}.{ClassName}`

**Produced:** `CONTAINER_BINDINGS.md`

---

### 4. Dispatch Deprecation

Established that the Dispatch component cannot be central across all ports — it relies on `::class` / `.class` dynamic
dispatch which doesn't exist in Go, TypeScript, or Python reliably.

**Replacement:** Closure-based handler contracts with fully typed signatures:

- HTTP: `(ContainerContract, map<string, mixed>) → ResponseContract`
- CLI: `(ContainerContract, map<string, mixed>) → OutputContract`
- Listener: `(ContainerContract, map<string, mixed>) → any`

`ServerRequestContract` and `RouteContract` are intentionally absent from handler signatures — available via container
when needed, keeping signatures minimal and concern-agnostic.

Each concern gets its own typed handler contract:

- `HttpHandlerContract` using `HttpHandlerFunc`
- `CliHandlerContract` using `CliHandlerFunc`
- `ListenerHandlerContract` using `ListenerHandlerFunc`

Dispatch retained as opt-in power feature for PHP and Java only.

**Produced:** `DISPATCH.md`

---

### 5. Provider Contracts

Established a complete provider contract hierarchy across all five languages:

- `ComponentProviderContract` — top-level aggregator, lists sub-providers by category
- `ServiceProviderContract` — container bindings via `publishers()` map
- `HttpRouteProviderContract` — HTTP routes via `getControllerClasses()` + `getRoutes()`
- `CliRouteProviderContract` — CLI routes (same shape as HTTP)
- `ListenerProviderContract` — event listeners

**Key decisions:**

- `getControllerClasses()` and `getListenerClasses()` are **absent** from Go and TypeScript — no annotation support
- All provider list methods must return simple literals — no conditional logic
- Python uses `list[type]` for class lists, `list[RouteContract]` for object lists
- TypeScript uses `Array<new () => Contract>` for provider class lists — constructor references allow direct
  instantiation at runtime

**Provider Naming Convention:**

All provider implementations — framework and application-defined — must be uniquely named across the entire framework.
The naming rule is identical to the throwable naming rule: prepend parent component (and subcomponent if needed) names
until the name is unique across the framework.

The forcing function is the generated data cache file. `AppContainerData`, `AppEventData`, `AppHttpRoutingData`, and
`AppCliRoutingData` each reference providers from multiple components in a single generated file. Identical class names
across components produce namespace collisions that prevent compilation. Unique names are a hard requirement, not a
style preference.

**Pattern:** `ComponentNameTypeProvider` or `ComponentNameSubComponentTypeProvider`

Where `Type` is one of:

- `Component` — top-level aggregator (one per component)
- `Service` — container bindings
- `HttpRoutes` — HTTP route definitions
- `CliRoutes` — CLI route definitions
- `Listeners` — event listener definitions

**Framework examples:**

```
HttpComponentProvider               — top-level HTTP aggregator
HttpServiceProvider                 — HTTP container bindings
HttpRoutingHttpRoutesProvider       — HTTP routing subcomponent routes
HttpRoutingListenersProvider        — HTTP routing subcomponent listeners

ContainerComponentProvider          — top-level Container aggregator
ContainerServiceProvider            — Container bindings

EventComponentProvider              — top-level Event aggregator
EventServiceProvider                — Event container bindings
EventListenersProvider              — Event listeners

CliComponentProvider                — top-level CLI aggregator
CliServiceProvider                  — CLI container bindings
CliRoutesProvider                   — CLI routes
```

**Application-defined providers** follow the same rule. A developer overriding `HttpServiceProvider` to customise the
router binding names their override `AppHttpServiceProvider` or `UserHttpServiceProvider` — never
`HttpServiceProvider` (conflicts with the framework class) or `ServiceProvider` (ambiguous across the entire codebase).

The uniqueness rule applies recursively — the same question as throwables: is this name unique across the entire
framework? If no, prepend the immediate parent name and ask again.

**Produced:** `CONTRACTS_JAVA.md`, `CONTRACTS_GO.md`, `CONTRACTS_PYTHON.md`, `CONTRACTS_TYPESCRIPT.md`

---

### 6. Data Cache Files

Established the complete cache generation architecture:

**Four output classes per application:**

- `AppContainerData` — all bindings
- `AppEventData` — all listeners
- `AppHttpRoutingData` — all HTTP routes (uses `HttpRoutingData` structure with routes, paths, dynamicPaths, regexes
  indexes)
- `AppCliRoutingData` — all CLI routes

**sindri tool:**

- Standalone tool, separate from the framework
- Reads application `AppConfig` class for top-level providers — no separate yaml file needed
- Walks static provider tree via AST — no application execution needed
- Constructs `ValkyrjaRoute` objects and runs `ProcessorContract::route()` for regex compilation
- Generates source code as text — no application classes needed at generation time
- All five languages work without cache via direct provider instantiation

**Route parameters:**

- `#[Parameter]` / `@Parameter` annotations carry name and pattern
- Build tool extracts parameters from AST
- Compiled regex pre-built and stored in cache data class

**All languages work without cache** — this is guaranteed by the provider contract design:

- PHP/Java/Python: class references allow direct instantiation
- TypeScript: constructor references (`Array<new () => Contract>`) allow direct instantiation
- Go: interface methods called directly on provider structs

**PHP CLI command note:** The existing `cache:generate` command will break when handler logic is implemented. Must
migrate to `sindri` before handler logic ships.

**Produced:** `DATA_CACHE.md`

---

### 7. Build Tool Architecture

Established that:

- `sindri` is a standalone tool, separate repo per language
- The build tool is itself a Valkyrja application — validates the cache-optional architecture
- It can generate its own cache for optimized subsequent runs (three-pass bootstrap)
- Language-specific AST implementations with practical code examples

**PHP (`sindri`, formerly `Bin`):**

- Bin component extracted from framework to separate repository
- `nikic/php-parser` as dependency of `sindri`, not the framework
- All file generation and scaffolding moves here
- Dev-only Composer dependency

**Per-language build tools:**

- PHP: `sindri` — nikic/php-parser
- Java: `io.valkyrja:sindri` — Trees API + JavaPoet (annotation processor)
- Go: `io/valkyrja/sindri` — go/analysis, go/ast (stdlib)
- Python: `sindri` — ast, inspect (stdlib)
- TypeScript: `@valkyrja/sindri` — TypeScript compiler API

**Framework source shipping policy:**

- PHP/Python/Go: source always available
- Java: must publish `-sources.jar` as required build dependency
- TypeScript: must ship `.ts` source alongside compiled `.js`

**Produced:** `BUILD_TOOL.md`

---

### 8. Port-Specific Notes

Documented per-language notes including:

- Python vs FastAPI differentiation — Valkyrja's compile-once cached bootstrap is a genuine differentiator
- Python CGI vs worker mode — both supported, identical codebase
- TypeScript constructor references enabling runtime without cache
- Go's single-binary compiled nature and near-instant startup
- Java Project Loom virtual threads for concurrency

**Produced:** `PORTS.md`

---

## Documents Produced

| File                      | Contents                                                                   |
|---------------------------|----------------------------------------------------------------------------|
| `PORTS.md`                | Language port list, per-language characteristics, comparison table         |
| `THROWABLES.md`           | Exception naming convention, hierarchy, language mapping, decision tree    |
| `CONTAINER_BINDINGS.md`   | Closure bindings, string constants, per-component constants files          |
| `DISPATCH.md`             | Handler contracts, typed closure signatures, dispatch deprecation          |
| `DATA_CACHE.md`           | Cache architecture, provider contracts, AppConfig entry point, build flows |
| `BUILD_TOOL.md`           | Build tool design, language AST implementations, Bin extraction            |
| `CONTRACTS_JAVA.md`       | Java contracts and implementations with full code examples                 |
| `CONTRACTS_GO.md`         | Go contracts and implementations with full code examples                   |
| `CONTRACTS_PYTHON.md`     | Python contracts and implementations with full code examples               |
| `CONTRACTS_TYPESCRIPT.md` | TypeScript contracts and implementations with full code examples           |
| `README_PHP.md`           | PHP implementation changes required, priority order                        |
| `README_JAVA.md`          | Java port implementation notes, priority order                             |
| `README_GO.md`            | Go port implementation notes, priority order                               |
| `README_PYTHON.md`        | Python port implementation notes, priority order                           |
| `README_TYPESCRIPT.md`    | TypeScript port implementation notes, priority order                       |

---

## Key Insights

**All languages work without cache.** The provider contract design — class references in PHP/Java/Python, constructor
references in TypeScript, interface methods in Go — allows the framework to traverse the provider tree and register
everything at runtime with no cache. Cache is a performance optimization, not a correctness requirement.

**The build tool is a text generator.** It writes strings that are valid source code. It never needs application classes
compiled in — class names from AST are written as text and the compiler resolves them later. This is how every
pre-compilation code generation step works in every language.

**Four data classes for the entire application.** The build tool aggregates everything across all providers into exactly
four classes — one per concern. The framework loads four objects at boot. No merging, no iteration.

**Typed handler signatures move errors to before production.** The dispatch approach had no type enforcement — wrong
method names were discovered at request time in production. Explicit closure handlers with typed signatures catch wrong
return types and wrong parameters at compile time (Java, Go, TypeScript) or CI time (PHP, Python).

**Bin belongs outside the framework, and becomes sindri.** File generation and scaffolding are build-time concerns.
Moving them to `sindri` removes all AST and build tooling from the framework's production dependency tree.

**The framework has zero AST dependencies.** All AST logic lives in the build tool. The framework only knows how to load
cache data classes if they exist and how to traverse the provider tree if they don't.

**No valkyrja.yaml.** The application config class is already the authoritative provider list — no separate file needed.
The build tool reads it via AST. This eliminates a duplicate source of truth.

**Python uses string constants for container binding keys** — same as Go and TypeScript. Using class objects as keys
forces module imports, defeating Python 3.14's lazy import mechanism. **Valkyrja's Python port requires Python 3.14
minimum** — lazy imports are the language-level solution to Python's cold start import problem. For Lambda workloads
where cold starts remain a concern, the Go or TypeScript port provides compiled binary startup times within the same
framework ecosystem.

**No existing Python framework does this.** FastAPI uses function objects as dependency identifiers with per-request
resolution — no container, no caching. Django has no DI container at all. Third-party containers like `lagom` use class
objects as keys, forcing imports. Valkyrja's string constant approach with lambda-wrapped bindings and Python 3.14 lazy
imports is the first Python framework design that achieves genuine lazy loading of service providers at the container
level.

Per-component constants files ship with the framework for all five languages — PHP holds `::class` strings, Java holds
`.class` objects, Go/Python/TypeScript hold string literals. Application-defined constants follow the same pattern,
written by the developer. Forge auto-generating application constants is a planned future enhancement.

**No component provider constants class.** Provider class references must use `::class` / `.class` / class objects
directly. A constants class would allow constant references that the build tool cannot resolve statically. Binding key
constants files are unaffected.
