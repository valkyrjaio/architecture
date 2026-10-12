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

| Directory     | Holds                                                   |
| ------------- | ------------------------------------------------------- |
| `component/`  | one directory per component                             |
| `convention/` | a rule that holds across every component and every port |
| `language/`   | the per-port deltas, one directory per language         |

Each component's directory holds two files:

| File        | Is                                                                            |
| ----------- | ----------------------------------------------------------------------------- |
| `<NAME>.md` | the contract: the hierarchy, the names and the behavior every port implements |
| `README.md` | the template a port's own component `README.md` starts from                   |

A contract **carries no code example**. An example in one language becomes the
spelling a port copies, instead of the contract it has to satisfy. The examples
live in each port's own `README.md`, written from the template.

**The template carries headings, not prose.** It fixes the sections and their
order, so every port's document reads alike, and it leaves the prose in the
contract. A template that carried the prose would be copied into every port, and
the first local edit would put one copy out of step with the contract — which is
the drift this repository exists to prevent.

### A feature reaches the architecture first

A change to what a component **is** lands here before it lands in any port: the
contract changes, then each port implements it. A change made in one port first
makes that port the de facto definition, and every later port inherits whatever
that port happened to do.

So the order is: change the contract, change the template if the sections move,
then open one pull request per port.

## Start here

- [`AGENTS.md`](AGENTS.md) — the operating guide, and the golden rules
- [`LIFECYCLE.md`](convention/LIFECYCLE.md) — the pipeline stages, their order
  and what each one guarantees
- [`PORT_PARITY.md`](convention/PORT_PARITY.md) — the baseline every port
  reaches, and what parity means

## The ports

| #   | Language       | Status                          | Build tool                           |
| --- | -------------- | ------------------------------- | ------------------------------------ |
| 1   | **PHP**        | Production — most complete port | `valkyrja/sindri`                    |
| 2   | **Java**       | In progress                     | `io.valkyrja:sindri`                 |
| 3   | **Go**         | Proof of concept                | `github.com/valkyrjaio/sindri-go/vN` |
| 4   | **Python**     | Planned                         | `valkyrja-sindri`                    |
| 5   | **TypeScript** | In progress                     | `@valkyrjaio/sindri`                 |

**No port is the reference — the documents here are.** Every port is built from
`component/` and `convention/`, which is what keeps the ports from drifting
apart: one description of the framework, and five implementations of it. A port
being the most complete does not make it the authority.

Future languages under consideration: Kotlin (nearly free from Java), Scala,
Rust, Ruby.

## Components

Grouped as a reading order. The order a port **builds** them in is the tier
table in [`PORT_PARITY.md`](convention/PORT_PARITY.md), and that is the one to
follow when porting.

Two components have no document, and deliberately: a marker facility and a
reflection facility are shaped by what one language offers, so there is no
cross-port contract for either to state.

**Foundation** — everything depends on these.

| Document                                           | Covers                                      |
| -------------------------------------------------- | ------------------------------------------- |
| [`THROWABLE.md`](component/THROWABLE/THROWABLE.md) | the failure root, and the handler           |
| [`TYPE.md`](component/TYPE/TYPE.md)                | typed values, and identifiers               |
| [`SUPPORT.md`](component/SUPPORT/SUPPORT.md)       | the clock, behind something a test controls |

**Core** — nothing runs without these.

| Document                                                 | Covers                                        |
| -------------------------------------------------------- | --------------------------------------------- |
| [`CONTAINER.md`](component/CONTAINER/CONTAINER.md)       | registration, resolution, child scopes        |
| [`EVENT.md`](component/EVENT/EVENT.md)                   | dispatch, listeners, the event id             |
| [`APPLICATION.md`](component/APPLICATION/APPLICATION.md) | boot, the provider tree, config, entry points |

**Protocols** — each runs the pipeline in
[`LIFECYCLE.md`](convention/LIFECYCLE.md).

| Document                               | Covers                                 |
| -------------------------------------- | -------------------------------------- |
| [`HTTP.md`](component/HTTP/HTTP.md)    | requests, responses, routing           |
| [`CLI.md`](component/CLI/CLI.md)       | input, output, commands                |
| [`GRPC.md`](component/GRPC/GRPC.md)    | calls, status, streaming, cancellation |
| [`QUEUE.md`](component/QUEUE/QUEUE.md) | jobs, the wire envelope, outcomes      |

**Services** — one contract, adapters behind it, the application names one.

| Document                                              | Covers                                   |
| ----------------------------------------------------- | ---------------------------------------- |
| [`LOG.md`](component/LOG/LOG.md)                      | records, by severity                     |
| [`CACHE.md`](component/CACHE/CACHE.md)                | keyed values, lifetimes, tags            |
| [`SESSION.md`](component/SESSION/SESSION.md)          | per-visitor state, and the forgery token |
| [`FILESYSTEM.md`](component/FILESYSTEM/FILESYSTEM.md) | files, local or remote                   |
| [`CRYPT.md`](component/CRYPT/CRYPT.md)                | two-way encryption                       |
| [`JWT.md`](component/JWT/JWT.md)                      | signed tokens                            |
| [`MAIL.md`](component/MAIL/MAIL.md)                   | sending to people                        |
| [`SMS.md`](component/SMS/SMS.md)                      | sending to phones                        |
| [`BROADCAST.md`](component/BROADCAST/BROADCAST.md)    | pushing to live listeners                |
| [`VIEW.md`](component/VIEW/VIEW.md)                   | rendering templates                      |

**Application** — the components an application builds on.

| Document                                              | Covers                          |
| ----------------------------------------------------- | ------------------------------- |
| [`AUTH.md`](component/AUTH/AUTH.md)                   | who the caller is               |
| [`ORM.md`](component/ORM/ORM.md)                      | records in a relational store   |
| [`VALIDATION.md`](component/VALIDATION/VALIDATION.md) | rules, and what failed          |
| [`API.md`](component/API/API.md)                      | one envelope for every response |

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
| [`ATTRIBUTES.md`](convention/ATTRIBUTES.md)                   | what an attribute declares, and its name    |
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
