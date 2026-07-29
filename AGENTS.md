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
- **Route middleware is appended, never deduplicated.** Across every protocol
  (HTTP, CLI, gRPC), both the runtime collector and `sindri` codegen *append* each
  registered middleware in order — they never dedupe. If the same middleware is
  scheduled twice at a stage it runs twice (including the qualified- vs
  simple-name spelling of one class); a duplicate is the developer's bug, not the
  framework's to silently fix, and the generated cache must mirror reflection
  exactly. Never add `distinct` / `array_unique` / `!contains` to middleware
  collection.

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

### Coverage is per file, and you must measure it

The 100% rule is **per file, not an aggregate**, and it binds both directions:

- **Every file you add is at 100%** — line and branch — on its own, before the
  change is done. A repo-wide percentage is not evidence: one fully untested new
  class hides inside a large, well-covered codebase and barely moves the total.
  This is not hypothetical. A new class landed at 65% line / 50% branch while the
  repo-wide number only fell from 100% to 96%, and every local check passed.
- **Every file you touch stays at 100%.** Adding a branch to an existing file
  means adding the test for it in the same change.

**A green gate is not proof of coverage.** No language's gate enforces a coverage
floor today — each one generates a report and then ignores it, so a build at 55%
passes exactly like one at 100%. Until that is fixed, **read the coverage report
yourself** before calling a change done, and check the per-file numbers for the
files you added or changed, not just the summary line.

The only exception is an **explicitly documented** one: code that genuinely
cannot be covered (a process-exiting call, a blocking server loop) is excluded in
the coverage tool's own config, narrowly, with a comment saying why. Two rules
about exclusions: an accepted gap must be *written down* where the tool reads it,
never merely tolerated in silence; and never lower a threshold to accommodate a
gap — a floor set to "whatever we happen to be at" legitimizes the gap and
defeats the point. Cover the code, or exclude it narrowly and say why.

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
10. **Never comment transient state.** A comment in code or config must stay true
    indefinitely. Do not write one to explain something temporary, or something
    an automated process will later rewrite — a pinned version awaiting a
    release, a workaround pending a fix, a value some job regenerates. Automation
    rewrites values, not the prose around them, so the comment outlives what it
    described and becomes an assertion that is now false. That is worse than no
    comment, because the next reader trusts it.
    **Put it in the PR description instead — nothing is lost.** The squash merge
    writes the PR title as the commit subject and the *entire PR description* as
    the commit body, so the explanation lives in git history permanently,
    attached to the commit that introduced it and reachable by `git log` /
    `git blame`. (This is also why the "no commit body" rule in §7 governs only
    the commits you write; the merge commit's body comes from the PR.) The
    explanation is better placed there anyway: pinned to when it was true,
    instead of floating in a file where a later automated edit silently
    falsifies it.
    The test is whether the comment states a *decision or invariant* or a
    *current condition*. "This job asserts only generated code, so its coverage
    report is meaningless" is a decision — keep it. "Pinned ahead of the others
    until the next release bumps it" is a condition — the release automation will
    strand it, so it belongs in the PR.

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

### Application entry points

Entry points are named for the runtime that drives them. The framework and the
starter apps group them on **different axes**, so the same runtime lands in a
different place in each — but both put the runtime in the **class name**, and that
is the invariant, not the directory shape.

**Framework — group by adapter, protocol in the class name.** One adapter serves
several protocols, so the adapter is the segment and the protocol is the suffix:

```
Application/Entry/
├── Http                        ← default, in-core; one per protocol (Cli, Grpc, …)
├── Cli
├── Abstract/                   ← App, and a WorkerX base per protocol
│   ├── App
│   └── WorkerHttp
├── OpenSwoole/OpenSwooleHttp   ← and OpenSwooleGrpc, OpenSwoolePushWorkerQueue as they land
├── RoadRunner/RoadRunnerHttp
└── FrankenPhp/FrankenPhpHttp
```

Java spells the same tree `entry/netty/{NettyHttp, NettyGrpc}` and
`entry/jetty/{JettyHttp, JettyGrpc}`, over an `abstract_/{App, WorkerHttp,
WorkerGrpc}`.

**Starter app and `template` — group by protocol, runtime in the class name.** The
app's axis is the protocol module, which owns `Config`, the controllers, the routing
data, and the providers; only the server driving them differs. So the protocol is the
segment and the runtime is the prefix, and the default entry keeps the bare name `App`:

|         | PHP                      | Java                |
|---------|--------------------------|---------------------|
| default | `App\Http\App`           | `app.http.App`      |
| variant | `App\Http\OpenSwooleApp` | `app.http.JettyApp` |

**Never nest a starter-app entry under the runtime** (`App\Http\OpenSwoole\App`). It
yields several classes named `App` inside one protocol — a dozen once gRPC and Queue
land — and the variant axis is not always an adapter, so it cannot be a segment:
[`QUEUE.md`](QUEUE.md) ships non-adapter `PullQueue` / `PushQueue` defaults, which
become `App\Queue\App` and `App\Queue\PushApp` alongside a per-runtime
`App\Queue\OpenSwoolePushApp`.

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
| Concrete exception | implements Throwable | `*Exception` (Go: `*Error`) | `Exception\` | — |
| Any throwable | extends/implements Throwable | — | `Throwable\` | — |
| Base `*RuntimeException` / `*InvalidArgumentException` | — | as-is | `Abstract\` | abstract |
| Type / Model / Entity | extends the base | — | `Type\` / `Model\` / `Entity\` | — |
| Abstract class | is abstract | must **not** contain `Abstract` | `Abstract\` | abstract |
| Enum | is an enum | must **not** contain `Enum` | `Enum\` | enum |
| Trait | is a trait | must **not** contain `Trait` (src) | `Trait\` | trait |
| Test fixture | reusable double in `Fixtures\` | `*Fixture` | `Fixtures\` | final |

The relationships are **bidirectional**: everything in `Contract\` must be an
interface *and* named `*Contract`; everything in `Enum\` must be an enum; every
final constant lives in `Constant\`; and so on. For `Abstract`, `Enum`, and
`Trait` the *segment* carries the meaning, so the **name must not repeat it** — an
abstract `Stream` is `Abstract\Stream`, never `AbstractStream`.

Tests: concrete test classes are `final`, live in `Unit\`/`Functional\`, and are
named `*Test`; reusable doubles live in `Fixtures\`, named `*Fixture`, never
`*Test`. A fixture that is itself an enum, trait, or contract keeps that type's
naming (`*Enum` / `*Trait` / `*Contract`) — the type rule supersedes the fixture
marker, just as the segment does for `Abstract`/`Enum`/`Trait` above. No class
carries an `@author` docblock.

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
| **Fixtures**   | reusable doubles used *by* tests, named `*Fixture` (never `*Test`) |
| **Abstract**   | base test cases (not themselves tests)                       |

### The test root

The taxonomy above is nested under a **dedicated test root**, so tests are never
siblings of the framework's own namespaces: PHP `Valkyrja\Tests\`, TypeScript
`tests/Tests/`, Java `io.valkyrja.tests.`. The segment matters most where the
namespace is a global identifier — without it, `io.valkyrja.fixtures.http.routing`
reads exactly like a framework package.

**Go is the only permitted deviation.** Go requires a test to live in the package
it tests (`*_test.go` beside the source, in the same package or its `_test`
sibling), so it has no separate test root; its reusable doubles still live in a
`fixtures` package mirroring the source tree, with the `*Fixture` suffix. No other
port may drop the test root.

One narrow carve-out applies to languages with package-private access: a test that
genuinely needs it sits in the source package instead, and only that test. PHP has
no such concept, so its tree is uniform; Java's guide names the exception.

Rules that hold everywhere: unit-test paths mirror `src`; test classes/files use
the language's test-name convention; reusable doubles are production-shaped
classes in `Fixtures`, never named like tests. **Every code branch is tested, all
tests and the full CI gate pass, and coverage is 100% (line and branch) — per
file, for every file added or touched — and never drops** — see the Definition of
done in §3, which also covers why a green gate is not proof of coverage and how
an unreachable line may be excluded. Per-code-shape recipes and coverage
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
- **Tag casing is exact** — write each tag exactly as listed above; initialisms
  stay uppercase (`[CI]`, `[GitHub]`), never `[Ci]` / `[Github]`.
- No body / co-author lines unless explicitly asked. This governs the commits
  *you* write; the squash merge takes its subject from the PR title and its body
  from the PR description, which is why that description is where durable
  explanation belongs (see §3, rule 10).
- PR description follows the
  [PR template](https://github.com/valkyrjaio/.github/blob/master/.github/PULL_REQUEST_TEMPLATE.md)
  — fill **Description**, **Types of changes**, and **Changes** (bold
  file/component — em dash — what changed).

### Current working branch

The current working branch is always the current-year `??.x` branch (for 2026,
`26.x`) — never `master`/`main`. If no current-year branch exists, use the
previous year's `??.x`; if that does not exist either, fall back to `master`.

**Always `git fetch` before you branch.** The check below is not satisfiable from
memory: `origin/<branch>` is a local cache, so comparing against it without
fetching compares against whatever was true the last time you fetched. A stale
remote-tracking ref reports "up to date" and a branch cut from it silently starts
from old code — the conflict does not surface until later, on an open PR, and
every review round trip until then is against the wrong base.

So, in order: `git fetch`, then confirm the target branch is not behind its remote
(or the branch it should track); if it is behind, update it — or confirm with the
user how to proceed — and only then create the branch or worktree. Fetch **every
time you start a branch**, not once per session: another agent's PR, a release, or
a dependency-bump job may have merged while you were working, and in these repos
they frequently do.

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
