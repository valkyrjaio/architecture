<p align="center"><a href="https://valkyrja.io" target="_blank">
    <img src="https://raw.githubusercontent.com/valkyrjaio/art/refs/heads/master/long-banner/orange/default.png" width="100%">
</a></p>

# Valkyrja Architecture

The architectural record of how [Valkyrja][valkyrja url] is designed, why
decisions were made, and what needs to be built across all language ports.

This repository is not end-user documentation. It exists to keep the
architecture consistent as the framework expands to new languages and as the
existing PHP implementation evolves. Decisions that affect multiple ports,
cross-language contracts, and implementation roadmaps are recorded here.

What's Included
---------------

- **Architecture documents** — cross-cutting decisions on throwables,
  container bindings, handlers, cache generation, and the build tool
- **Language contracts** — per-language provider contract specifications for
  Java, Go, Python, and TypeScript
- **Implementation notes** — per-language README files with port priority
  order and implementation guidance
- **TODO checklists** — per-language change and port checklists
- **Port roadmap** — priority order and current status across all five
  languages

The Ports
---------

Valkyrja is being ported to five languages in priority order:

| # | Language       | Status                                | Build tool          |
|---|----------------|---------------------------------------|---------------------|
| 1 | **PHP**        | Production — reference implementation | `valkyrja-forge`    |
| 2 | **Java**       | In progress                           | `io.valkyrja:forge` |
| 3 | **Go**         | Proof of concept                      | `io/valkyrja/forge` |
| 4 | **Python**     | Planned                               | `valkyrja-forge`    |
| 5 | **TypeScript** | Planned                               | `@valkyrja/forge`   |

Future languages under consideration: Kotlin (nearly free from Java), Scala,
Rust, Ruby.

Architecture Documents
----------------------

Cross-cutting architectural decisions that apply to all ports.

| Document                                         | Description                                               |
|--------------------------------------------------|-----------------------------------------------------------|
| [`SUMMARY.md`](SUMMARY.md)                       | Full session summary — all decisions and reasoning        |
| [`PORTS.md`](PORTS.md)                           | Language port list, per-language notes, comparison tables |
| [`THROWABLES.md`](THROWABLES.md)                 | Exception naming convention, hierarchy, language mapping  |
| [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md) | Closure bindings, string constants, per-component files   |
| [`DISPATCH.md`](DISPATCH.md)                     | Handler contracts, typed signatures, dispatch deprecation |
| [`DATA_CACHE.md`](DATA_CACHE.md)                 | Cache architecture, provider contracts, build flows       |
| [`BUILD_TOOL.md`](BUILD_TOOL.md)                 | Build tool design, Bin extraction, AST implementations    |

Language Contracts
------------------

Per-language provider contract specifications.

| Document                                                               | Description                                       |
|------------------------------------------------------------------------|---------------------------------------------------|
| [`java/PROVIDER_CONTRACTS.md`](java/PROVIDER_CONTRACTS.md)             | Java provider contracts and implementations       |
| [`go/PROVIDER_CONTRACTS.md`](go/PROVIDER_CONTRACTS.md)                 | Go provider contracts and implementations         |
| [`python/PROVIDER_CONTRACTS.md`](python/PROVIDER_CONTRACTS.md)         | Python provider contracts and implementations     |
| [`typescript/PROVIDER_CONTRACTS.md`](typescript/PROVIDER_CONTRACTS.md) | TypeScript provider contracts and implementations |

Implementation Notes
--------------------

Per-language implementation notes and priority order.

| Document                                       | Description                                               |
|------------------------------------------------|-----------------------------------------------------------|
| [`php/README.md`](php/README.md)               | PHP — changes required to existing implementation         |
| [`java/README.md`](java/README.md)             | Java — port implementation notes and priority order       |
| [`go/README.md`](go/README.md)                 | Go — port implementation notes and priority order         |
| [`python/README.md`](python/README.md)         | Python — port implementation notes and priority order     |
| [`typescript/README.md`](typescript/README.md) | TypeScript — port implementation notes and priority order |

TODO Checklists
---------------

| Document                           | Description           |
|------------------------------------|-----------------------|
| [`php/TODO.md`](php/TODO.md)       | PHP change checklist  |
| [`python/TODO.md`](python/TODO.md) | Python port checklist |

Core Architectural Principles
-----------------------------

**Every language works without cache.** The provider contract design — class
references in PHP/Java/Python, constructor references in TypeScript, interface
methods in Go — allows the framework to traverse the provider tree and register
everything at runtime. Cache is a cold-start performance optimization, not a
correctness requirement.

**The framework has zero AST dependencies.** All source extraction and code
generation logic lives in the per-language build tool packages. The framework
only knows how to load cache data files if they exist and traverse the provider
tree if they don't.

**The build tool is a text generator.** It writes strings that are valid source
code. It never needs application classes compiled in — class names from AST are
written as text, the compiler resolves them later.

**Four data classes for the entire application.** The build tool aggregates
everything across all providers into exactly four classes — `AppContainerData`,
`AppEventData`, `AppHttpRoutingData`, `AppCliRoutingData`. The framework loads
four objects at boot.

**Typed handler signatures move errors before production.** Explicit closure
handlers with typed signatures catch wrong return types at compile time (Java,
Go, TypeScript) or CI time (PHP, Python). The dispatch approach discovered
these errors at request time in production.

**The AppConfig class is the build tool entry point.** No separate YAML file.
The application config already lists all component providers — the build tool
reads it via AST.

**Component provider constants classes do not exist.** Provider class
references use `::class` / `.class` / class objects directly. Constants classes
for provider references would break the build tool's static analysis. Binding
key constants files are unaffected.

Key Decisions At a Glance
-------------------------

### Throwables

- Naming: `Valkyrja*` → `ComponentName*` → `SubComponent*` or
  `ParentSubComponent*` (if shared)
- Rule: prepend parent names until unique across the entire framework
- All base and categorical exceptions are abstract
- Every component always ships `ComponentRuntimeException` and
  `ComponentInvalidArgumentException`
- See [`THROWABLES.md`](THROWABLES.md)

### Container Bindings

- All bindings use explicit closure factories — no reflection-based
  instantiation
- Per-component string constants files for cross-language binding key identity
- See [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md)

### Handlers

- Three typed handler signatures: HTTP → `ResponseContract`, CLI →
  `OutputContract`, Listener → `any`
- Parameters: `(ContainerContract, map<string, mixed>)` — `ServerRequestContract`
  and `RouteContract` available via container when needed, not explicit
  parameters
- `#[Handler]` / `@Handler` / `@handler` — metadata marker in all languages,
  never active registrar
- See [`DISPATCH.md`](DISPATCH.md)

### Cache Generation

- Build tool reads `AppConfig` class, walks provider tree via AST, generates
  four data classes
- Routes: `Parameter` objects carry segment constraints, `ProcessorContract`
  compiles regex, stored pre-compiled
- Python `@handler` is metadata only — `_valkyrja_handler` on the function,
  read by framework at bootstrap, skipped when cache loaded
- See [`DATA_CACHE.md`](DATA_CACHE.md)

### Build Tool

- Separate repository and package per language — dev dependency only, never
  production
- PHP `Bin` component extracted to `valkyrja-forge` — `nikic/php-parser` lives
  there, not in the framework
- Build tool is itself a Valkyrja application — validates the cache-optional
  architecture
- See [`BUILD_TOOL.md`](BUILD_TOOL.md)

PHP — Changes Required
----------------------

The PHP implementation is complete but requires alignment changes before other
ports diverge too far. See [`php/TODO.md`](php/TODO.md) for the full checklist.

Priority items:

1. Throwable renaming and abstraction
2. Provider contract interfaces
3. `publishers()` map migration
4. `#[Handler]` and `#[Parameter]` attributes
5. Bin extraction to `valkyrja-forge` — **must happen before handler logic
   ships** (existing `cache:generate` will break)

Starting a New Port
-------------------

Port components in this order:

**Container → Dispatch → Event → Application → CLI → HTTP → Bin**

Read these files in order:

1. [`PORTS.md`](PORTS.md) — language-specific characteristics and decisions
2. [`THROWABLES.md`](THROWABLES.md) — exception hierarchy for your language
3. [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md) — binding key constants and
   closure factories
4. [`DISPATCH.md`](DISPATCH.md) — handler contracts and typed closure
   signatures
5. [`DATA_CACHE.md`](DATA_CACHE.md) — provider contracts and cache generation
6. [`BUILD_TOOL.md`](BUILD_TOOL.md) — build tool implementation for your
   language
7. `{language}/PROVIDER_CONTRACTS.md` — full contract and implementation
   examples
8. `{language}/README.md` — implementation notes and priority order

Relationship to Framework Repositories
--------------------------------------

```
architecture   ← you are here — decisions and roadmaps
     │
     ├── valkyrja-php                 ← PHP framework (runtime, zero build deps)
     ├── valkyrja-forge-php           ← PHP build tool (nikic/php-parser)
     ├── valkyrja-java                ← Java framework (runtime)
     ├── valkyrja-forge-java          ← Java build tool (annotation processor)
     ├── valkyrja-go                  ← Go framework (runtime)
     ├── valkyrja-forge-go            ← Go build tool (go/analysis)
     ├── valkyrja-python              ← Python framework (runtime)
     ├── valkyrja-forge-python        ← Python build tool (ast + inspect)
     ├── valkyrja-ts                  ← TypeScript framework (runtime)
     └── valkyrja-forge-ts            ← TypeScript build tool (TS compiler API)
```

Each framework repository is runtime-only with zero AST or build tooling
dependencies. Each build tool (`valkyrja-forge-*`) is a dev-only dependency
containing all code generation logic for that language.

Contributing
------------

Contributions to the architecture — new decision records, port planning
updates, clarifications to existing decisions — are welcome. See
[`CONTRIBUTING.md`][contributing url] for the submission process,
[`REPOSITORY_NAMING.md`][repository naming url] for how repos are named, and
[`VOCABULARY.md`][vocabulary url] for terminology used across the project.

Architecture decisions should be recorded here *before* they land in code, so
that per-language ports can reference the canonical decision rather than
rediscovering it.

License
-------

Licensed under the [MIT license][MIT license url]. See
[`LICENSE.md`](./LICENSE.md).

[valkyrja url]: https://valkyrja.io

[contributing url]: https://github.com/valkyrjaio/.github/blob/master/CONTRIBUTING.md

[repository naming url]: https://github.com/valkyrjaio/.github/blob/master/REPOSITORY_NAMING.md

[vocabulary url]: https://github.com/valkyrjaio/.github/blob/master/VOCABULARY.md

[MIT license url]: https://opensource.org/licenses/MIT
