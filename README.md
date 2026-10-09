<p align="center"><a href="https://valkyrja.io" target="_blank">
    <img src="https://raw.githubusercontent.com/valkyrjaio/art/refs/heads/26.x/long-banner/orange/default.png" width="100%">
</a></p>

# Valkyrja Architecture

The contract every [Valkyrja][valkyrja url] port implements. It states the
hierarchy, the names and the behavior that must be the same in every language, so
a reader who knows one port can read another.

This repository is not end-user documentation, and it is not a design journal. A
document here earns its place by stating a rule that holds **across** languages.
How one port spells that rule lives in that port, next to the code.

## How it is organized

| Directory     | Holds                                                         |
| ------------- | ------------------------------------------------------------- |
| `component/`  | one document per component: its hierarchy, names and behavior |
| `convention/` | a rule that holds across every component and every port       |
| `language/`   | the per-port deltas, one directory per language               |

A `component/` document **carries no code example**. An example in one language
becomes the spelling a port copies, instead of the contract it has to satisfy.
The examples live in that component's `README.md` in each port.

## Start here

- [`AGENTS.md`](AGENTS.md) — the operating guide, and the golden rules
- [`LIFECYCLE.md`](convention/LIFECYCLE.md) — the pipeline stages, their order
  and what each one guarantees
- [`PORT_PARITY.md`](convention/PORT_PARITY.md) — the baseline every port
  reaches, and where each one stands

## The ports

| #   | Language       | Status                                | Build tool                           |
| --- | -------------- | ------------------------------------- | ------------------------------------ |
| 1   | **PHP**        | Production — reference implementation | `valkyrja/sindri`                    |
| 2   | **Java**       | In progress                           | `io.valkyrja:sindri`                 |
| 3   | **Go**         | Proof of concept                      | `github.com/valkyrjaio/sindri-go/vN` |
| 4   | **Python**     | Planned                               | `valkyrja-sindri`                    |
| 5   | **TypeScript** | Planned                               | `@valkyrja/sindri`                   |

PHP is the reference implementation for every component it holds. **The reference
is per component**, so a component PHP does not hold takes the reference of the
port that built it first — Grpc is Java's today. Each `component/` document names
its own reference.

Future languages under consideration: Kotlin (nearly free from Java), Scala,
Rust, Ruby.

## Components

| Document                                     | Covers                                        |
| -------------------------------------------- | --------------------------------------------- |
| [`APPLICATION.md`](component/APPLICATION.md) | boot, the provider tree, config, entry points |
| [`CONTAINER.md`](component/CONTAINER.md)     | registration, resolution, child scopes        |
| [`EVENT.md`](component/EVENT.md)             | dispatch, listeners, the event id             |
| [`HTTP.md`](component/HTTP.md)               | requests, responses, routing                  |
| [`CLI.md`](component/CLI.md)                 | input, output, commands                       |
| [`GRPC.md`](component/GRPC.md)               | calls, status, streaming, cancellation        |
| [`QUEUE.md`](component/QUEUE.md)             | jobs, the wire envelope, outcomes             |

## Conventions

| Document                                                      | Covers                                      |
| ------------------------------------------------------------- | ------------------------------------------- |
| [`LIFECYCLE.md`](convention/LIFECYCLE.md)                     | the pipeline stages and their guarantees    |
| [`STRUCTURE.md`](convention/STRUCTURE.md)                     | the structure taxonomy                      |
| [`CONTRACTS.md`](convention/CONTRACTS.md)                     | what a contract is, per language mechanism  |
| [`PROVIDERS.md`](convention/PROVIDERS.md)                     | the provider hierarchy and its contracts    |
| [`THROWABLES.md`](convention/THROWABLES.md)                   | the exception hierarchy and its naming      |
| [`HANDLERS.md`](convention/HANDLERS.md)                       | handler signatures and registration         |
| [`CONTAINER_BINDINGS.md`](convention/CONTAINER_BINDINGS.md)   | binding keys, per language                  |
| [`COMPONENT_CONFIG.md`](convention/COMPONENT_CONFIG.md)       | how a component's config is split           |
| [`DATA_CACHE.md`](convention/DATA_CACHE.md)                   | the generated data classes                  |
| [`SINDRI.md`](convention/SINDRI.md)                           | `sindri`, and its output                    |
| [`METHOD_NAMING.md`](convention/METHOD_NAMING.md)             | what a method prefix promises               |
| [`PACKAGE_NAMING.md`](convention/PACKAGE_NAMING.md)           | package, registry and namespace names       |
| [`STATIC_METHODS.md`](convention/STATIC_METHODS.md)           | where a static method may live              |
| [`COMMENTS.md`](convention/COMMENTS.md)                       | what a comment may state                    |
| [`DOCUMENTATION_STYLE.md`](convention/DOCUMENTATION_STYLE.md) | the writing rules for prose                 |
| [`TESTING_METHODOLOGY.md`](convention/TESTING_METHODOLOGY.md) | testing, and 100% coverage                  |
| [`PORT_PARITY.md`](convention/PORT_PARITY.md)                 | the baseline every port reaches             |
| [`PORTS.md`](convention/PORTS.md)                             | per-language characteristics and transforms |
| [`ADDING_A_MODULE.md`](convention/ADDING_A_MODULE.md)         | adding a component                          |
| [`ADDING_NEW_LANGUAGE.md`](convention/ADDING_NEW_LANGUAGE.md) | adding a port                               |
| [`CI_TOOLS.md`](convention/CI_TOOLS.md)                       | the gate each language runs                 |
| [`SHELL_SCRIPTS.md`](convention/SHELL_SCRIPTS.md)             | the rules for shell                         |
| [`COMMIT_CONVENTION.md`](convention/COMMIT_CONVENTION.md)     | commit and pull request title format        |
| [`PR_DESCRIPTION.md`](convention/PR_DESCRIPTION.md)           | what a pull request description holds       |
| [`VERSIONING.md`](convention/VERSIONING.md)                   | the version scheme and release automation   |
| [`VERSION_SUPPORT.md`](convention/VERSION_SUPPORT.md)         | which versions are supported                |
| [`BRANCH_PROMOTION.md`](convention/BRANCH_PROMOTION.md)       | how a branch is promoted                    |

## Languages

Each directory holds that port's agent guide, its provider contracts, and its
port notes.

[`php`](language/php/AGENTS.md) · [`java`](language/java/AGENTS.md) ·
[`typescript`](language/typescript/AGENTS.md) ·
[`go`](language/go/AGENTS.md) · [`python`](language/python/AGENTS.md) ·
[`kotlin`](language/kotlin/AGENTS.md)

## Keeping it true

A document that describes the old behavior is worse than no document, because the
reader trusts it. Two rules follow, and
[`AGENTS.md`](AGENTS.md) §3 holds both in full:

- **Update the document in the same pull request as the change.** Never leave it
  for a later sweep.
- **Verify a claim before you write it.** A sentence about another file is a
  claim. Open the file and read the code the sentence describes.

[valkyrja url]: https://valkyrja.io
