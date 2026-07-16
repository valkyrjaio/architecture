# AGENTS.md — Valkyrja (cross-language canonical)

The **canonical, cross-language** operating guide for coding agents working in any
Valkyrja repository — PHP, Java, Go, Python, or TypeScript. It captures the rules
that are **the same in every language**.

This is **Layer 1** of a three-layer guide:

1. **This file** — cross-language rules that apply everywhere.
2. **`<language>/AGENTS.md`** (in this repo, next to this file) — the per-language
   deltas: exact CI commands, package roots, tool lists, test mapping, and the
   per-language spelling of the structure taxonomy (§4).
   → [`php`](php/AGENTS.md) · [`java`](java/AGENTS.md) · [`go`](go/AGENTS.md) ·
   [`python`](python/AGENTS.md) · [`typescript`](typescript/AGENTS.md) ·
   [`kotlin`](kotlin/AGENTS.md)
3. **A thin `AGENTS.md` in each framework repo** — says what that repo is and
   links back here.

> A fix to a rule that applies to all languages belongs **here**. A fix specific
> to one language belongs in that language's Layer-2 file. When those and a
> deeper architecture document disagree, the architecture document wins — fix the
> guide.

> **Before contributing, also read
> [`CONTRIBUTING.md`](https://github.com/valkyrjaio/.github/blob/master/CONTRIBUTING.md)**
> — the submission process, running CI locally, the commit/PR conventions, and
> branch targeting. This guide is the technical companion to it.

---

## 1. What Valkyrja is

Valkyrja is a single framework ported to five languages in priority order. PHP is
the **reference implementation**; every other port mirrors its structure,
naming, and tests.

| # | Language       | Status                                | Package root / namespace |
|---|----------------|---------------------------------------|--------------------------|
| 1 | **PHP**        | Production — reference implementation  | `Valkyrja\`              |
| 2 | **Java**       | In progress                           | `io.valkyrja`            |
| 3 | **Go**         | Proof of concept                      | `io/valkyrja`            |
| 4 | **Python**     | Planned                               | `valkyrja`               |
| 5 | **TypeScript** | Planned                               | `@valkyrjaio/valkyrja`   |
| 6 | **Kotlin**     | Planned (JVM — nearly free from Java) | `io.valkyrja`            |

Each language has parallel repos: the **framework** (runtime, zero build/AST
dependencies), **sindri** (the dev-only build tool that generates the cache), an
`application` example, a `template` skeleton, and `entry/*` server adapters. The
build tool is called `sindri` and is never a production dependency.

The **`template` repo is the structural source of truth** — it defines how a
repo's directories, files, and classes are laid out. Every new repo in that
language is scaffolded from it (see §3, rule 9).

Use the shared vocabulary (app, module, component, tool) consistently — see
[VOCABULARY.md](https://github.com/valkyrjaio/.github/blob/master/VOCABULARY.md).

---

## 2. Core architectural principles

These hold in **every** language. Do not violate them in a port.

- **Every language works without cache.** Providers expose class/constructor
  references (PHP/Java/Python `::class`/`.class`, TypeScript `new () => T`, Go
  interface methods) so the framework can walk the provider tree and register
  everything at runtime. Cache is a cold-start optimization, not a correctness
  requirement.
- **The framework has zero AST dependencies.** All source extraction and code
  generation lives in `sindri` (the build tool), never in the framework.
- **Four data classes for the whole app.** `sindri` aggregates every provider
  into exactly four generated classes — `AppContainerData`, `AppEventData`,
  `AppHttpRoutingData`, `AppCliRoutingData`. The framework loads four objects at
  boot.
- **Typed handler signatures, not dynamic dispatch.** Handlers are explicit
  typed closures — HTTP → `ResponseContract`, CLI → `OutputContract`, Listener →
  `any`. Parameters are `(ContainerContract, map<string, mixed>)`; request/route
  come from the container, not the signature. `#[Handler]` / `@Handler` /
  `@handler` is a **metadata marker only**, never an active registrar.
- **`AppConfig` is the build tool entry point.** No `valkyrja.yaml`. The app
  config class already lists the component providers; `sindri` reads it via AST.
- **No provider-reference constants class.** Provider references use
  `::class` / `.class` / class objects / constructor references directly so
  `sindri` can resolve them statically. (Binding-*key* constants files are fine
  and expected — see §4.)

Full detail: [`SUMMARY.md`](SUMMARY.md) and [`README.md`](README.md).

---

## 3. Golden rules for every change

**Definition of done — non-negotiable, across the board, in every language and
every repo.** A change is not finished until, for the repo you touched:

- **Every code branch is tested** — *branch* coverage, not just line coverage.
  Every path, guard, and error branch gets a test (use synthetic inputs to reach
  defensive guards that normal input can't). ("Branch" here means a code
  path, not a git branch.)
- **All tests pass.**
- **Every CI check passes** — the *full* gate (static analysis, formatting,
  architecture, migration, tests), never a subset.
- **Coverage is and stays 100%** — line *and* branch. It must never drop.

Then:

1. **Port code and its tests together**, never as a later pass. Mirror the source
   repo's test layout and map the framework (e.g. PHPUnit → Vitest: `assertSame`
   → `toBe`, data providers → `it.each`, `setUp` → `beforeEach`).
2. **End every file with a trailing newline.**
3. **American English** in all prose and identifiers ("color", "normalize").
4. **Every source file carries the license header** (see §5).
5. **Target the right branch** (see §7) — improvements/bug fixes go to the lowest
   affected `??.x`, new features/deprecations go to `master`.
6. **Run the full CI gate** for the language you touched before considering the
   work done — exact commands are in your language's Layer-2 guide.
7. **One branch and one PR per change.** Create a new branch off the correct
   target branch, then commit with the `[Component]` message, push, and open a PR
   (base = that target branch) with the template filled out. **Ask for
   confirmation before committing, before pushing, and before opening the PR.**
   Keep each branch/PR small and atomic. See §7.
8. **Cross-language changes propagate.** If a change affects more than one port,
   make it in every affected language in the *same* batch (code and tests
   together) and cross-link the sibling PRs. See §7.
9. **New repos are scaffolded from the language's `template` repo** — the source
   of truth for repo layout and file/class structure. Start from it; never
   hand-assemble a repo's structure. Your Layer-2 guide names the template repo.

---

## 4. Naming conventions (identical across languages)

### Throwables / exceptions

Recursive uniqueness rule: `Valkyrja*` → `ComponentName*` → `SubComponent*`, and
prepend parent names until the name is **unique across the entire framework**.

- All base and categorical exceptions are **abstract**; only concrete, specific
  exceptions are thrown.
- Every component always ships `ComponentRuntimeException` and
  `ComponentInvalidArgumentException`, even if currently unused.
- Each language maps the framework base onto its native root — see the Layer-2
  guide (e.g. Java extends `IllegalArgumentException`; TypeScript extends `Error`
  and sets `this.name`).

Detail: [`THROWABLES.md`](THROWABLES.md).

### Providers

Same recursive uniqueness rule — the forcing function is the single generated
data-cache file that references providers from many components at once, so
identical names collide and break compilation.

Pattern: `ComponentName[SubComponent]TypeProvider`, where `Type` is one of
`Component` (top-level aggregator), `Service` (container bindings), `HttpRoutes`,
`CliRoutes`, `Listeners`. Examples: `HttpComponentProvider`, `HttpServiceProvider`,
`HttpRoutingHttpRoutesProvider`, `CliRoutesProvider`. App-defined overrides
prepend `App`/`User` (e.g. `AppHttpServiceProvider`), never bare
`HttpServiceProvider` or `ServiceProvider`.

### Contracts

Interfaces are suffixed `Contract` (`ContainerContract`, `RouterContract`) and
live in a `Contract/` (or `.contract`) subpackage. The concrete implementation
takes the bare name (`Container`, `Router`); shared behavior goes in an
`Abstract/` base.

### Structure taxonomy (enforced)

A class's *kind* is encoded three ways at once — its **name suffix**, the
**segment** (namespace/package/directory) it lives in, and its **modifier** — and
all three must agree. This is the machine-verified spec (PHP's PHPArkitect
`Rules` is the reference; Java ArchUnit and Kotlin Konsist mirror it; where a
language has no architecture linter — Go, Python, TypeScript — it is enforced in
review). PHP segment spellings are shown; **each Layer-2 guide gives the
per-language spelling** (case + reserved-word handling + constructs a language
lacks).

| Kind | Identified by | Name | Segment | Modifier |
|------|---------------|------|---------|----------|
| Contract | is an interface | `*Contract` | `Contract\` | interface |
| Service provider | implements `ServiceProviderContract` | `*ServiceProvider` | `Provider\` | — |
| Component provider | implements `ComponentProviderContract` | `*ComponentProvider` | `Provider\` | — |
| Route provider | implements `Http`/`CliRouteProviderContract` | `*RouteProvider` | `Provider\` | — |
| Listener provider | implements `ListenerProviderContract` | `*ListenerProvider` | `Provider\` | — |
| Factory | — | `*Factory` | `Factory\` | — |
| Constant | — | `*Constant` | `Constant\` | final |
| Attribute / annotation | has the attribute marker | — | `Attribute\` | — |
| CLI command | — | `*Command` | `Cli\Command\` | — |
| Security | — | `*Security` | `Security\` | final |
| Concrete exception | implements Throwable | `*Exception` | `Exception\` | — |
| Any throwable | extends/implements Throwable | — | `Throwable\` | — |
| Base `*RuntimeException` / `*InvalidArgumentException` | — | as-is | `Abstract\` | abstract |
| Type / Model / Entity | extends the base | — | `Type\` / `Model\` / `Entity\` | — |
| Abstract class | is abstract | must **not** contain `Abstract` | `Abstract\` | abstract |
| Enum | is an enum | must **not** contain `Enum` | `Enum\` | enum |
| Trait | is a trait | must **not** contain `Trait` (src) | `Trait\` | trait |

The relationships are **bidirectional**: everything in `Contract\` must be an
interface *and* named `*Contract`; everything in `Enum\` must be an enum; every
final constant lives in `Constant\`; and so on. For `Abstract`, `Enum`, and
`Trait` the *segment* carries the meaning, so the **name must not repeat it** — an
abstract `Stream` is `Abstract\Stream`, never `AbstractStream`.

Tests: concrete test classes are `final`, live in `Unit\`/`Functional\`, and are
named `*Test`; reusable doubles live in `Fixtures\`, named `*Class`, never
`*Test`. No class carries an `@author` docblock.

### Binding-key constants

Per-component constants files (never one central file). String format
`io.valkyrja.{component}.{ClassName}`. PHP holds `::class` strings, Java holds
`.class` objects, Go/Python/TypeScript hold string literals. Detail:
[`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md).

### Port order for a new component

**Container → Dispatch → Event → Application → CLI → HTTP → Bin.**

---

## 5. License header (every source file)

Every file begins with the header (comment syntax adapted per language), and
languages that support it declare strictness first (PHP `declare(strict_types=1);`,
TypeScript `strict` via tsconfig, Java UTF-8 + JSpecify nullness):

```
This file is part of the Valkyrja Framework package.

(c) Melech Mizrachi <melechmizrachi@gmail.com>

For the full copyright and license information, please view the LICENSE
file that was distributed with this source code.
```

Other cross-language style: concrete classes are `final` where the language
supports it; override methods are marked (`#[Override]` / `@Override` /
`noImplicitOverride`); contracts are interfaces.

---

## 6. Testing (shared shape)

Every framework repo mirrors the same test taxonomy — the layout is consistent
across languages and must be preserved in ports:

| Kind           | Holds                                                         |
|----------------|--------------------------------------------------------------|
| **Unit**       | one class in isolation; path mirrors the `src` path          |
| **Functional** | boots the app / exercises several units together             |
| **Fixtures**   | reusable real doubles/stubs used *by* tests (never `*Test`)  |
| **Abstract**   | base test cases (not themselves tests)                       |

Rules that hold everywhere: unit-test paths mirror `src`; test classes/files use
the language's test-name convention; reusable doubles are production-shaped
classes in `Fixtures`, never named like tests. **Every code branch is tested, all
tests and the full CI gate pass, and coverage is 100% (line and branch) and never
drops** — see the Definition of done in §3. Per-code-shape recipes and coverage
gotchas:
[`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md). Exact directory paths, test
framework, and the PHPUnit→target mapping live in your Layer-2 guide.

---

## 7. Branch, commit, push & open a PR

Every change lands on its own branch as a pull request. **Ask for confirmation
before each write action — committing, pushing, and opening the PR** (creating
the branch needs no prompt). Per change:

1. **Branch** off the correct target branch with a `prefix/…` name (see **Branch
   names** below; e.g. `feature/contextual-bindings`).
2. **Commit** — after confirmation — using the `[Component]` message format.
3. **Push** the branch — after confirmation.
4. **Open a PR** — after confirmation — with its **base set to that same target
   branch** and the PR template filled out (see below).

Keep each branch and PR small and atomic — one focused change per PR.

- **Commit** (single line, trailing period required):
  `[Component] Short imperative description.`
- **PR title** (same tag, **no** trailing period): `[Component] Description`
- **Component tags:** `[Documentation]`, `[CI]`, `[GitHub]`, `[Git]`,
  `[Composer]`, `[Functions]`, `[Deprecation]`, module tags like `[Container]` /
  `[Http]` / `[Cli]`, version tags like `[25.x]`, `[Release]`.
- No body / co-author lines unless explicitly asked.
- PR description follows the template — fill **Description**, **Types of
  changes**, and **Changes** (bold file/component — em dash — what changed).

### Current working branch

The current working branch is always the current-year `??.x` branch (for 2026,
`26.x`) — never `master`/`main`. If no current-year branch exists, use the
previous year's `??.x`; if that does not exist either, fall back to `master`.

Before starting work on that branch, first check it is not behind its remote (or
the branch it should track). If it is behind, update it — or confirm with the
user how to proceed — before doing any work.

### Branch names

`prefix/descriptive-name`, kebab-case. When an issue tracks the work, include it:
`prefix/ISSUE-{number}-descriptive-name` (e.g. `fix/ISSUE-123-header-normalization`).
The `prefix` and the PR's base branch are both set by the change type:

| Change type     | Target (base) branch                           | Branch prefix  |
|-----------------|------------------------------------------------|----------------|
| Improvement     | Lowest major affected `??.x`                   | `improvement/` |
| Bug fix         | Lowest major affected `??.x`                   | `fix/`         |
| New feature     | `master`                                       | `feature/`     |
| Deprecation     | `master`                                       | `deprecation/` |
| Breaking change | `master` (unless a bug fix — open issue first) | `breaking/`    |
| Documentation   | Lowest major affected branch the docs apply to | `docs/`        |

### Cross-language changes

If a change affects more than one language port, make it in **every affected
language in the same batch** — never a deferred follow-up. A bug fixed in the PHP
reference implementation that also exists in Java/TypeScript/etc. is fixed there
at the same time, code and tests together. Open one PR per language repo and
**cross-link the sibling PRs**: each PR's Description lists the matching PRs for
the other languages.

Full detail:
[CONTRIBUTING.md](https://github.com/valkyrjaio/.github/blob/master/CONTRIBUTING.md).

---

## 8. Where to read more

Read these in order when starting or extending a port:

1. [`PORTS.md`](PORTS.md) — per-language characteristics
2. [`THROWABLES.md`](THROWABLES.md) — exception hierarchy
3. [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md) — binding keys & closures
4. [`DISPATCH.md`](DISPATCH.md) — handler contracts
5. [`DATA_CACHE.md`](DATA_CACHE.md) — provider contracts & cache generation
6. [`BUILD_TOOL.md`](BUILD_TOOL.md) — `sindri` implementation
7. [`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md) — testing & 100% coverage
8. `{language}/PROVIDER_CONTRACTS.md` — full contracts + examples
9. `{language}/README.md` — port notes & priority order
10. `{language}/AGENTS.md` — the Layer-2 agent guide for that language
