# AGENTS.md — Valkyrja (cross-language canonical)

The **canonical, cross-language** operating guide for coding agents working in any
Valkyrja repository — PHP, Java, Go, Python, or TypeScript. It captures the rules
that are **the same in every language**.

This is **Layer 1** of a three-layer guide:

1. **This file** — cross-language rules that apply everywhere.
2. **`<language>/AGENTS.md`** (in this repo, next to this file) — the per-language
   deltas: exact CI commands, package roots, tool lists, test mapping, and the
   per-language spelling of the structure taxonomy ([`STRUCTURE.md`](STRUCTURE.md)).
   → [`php`](php/AGENTS.md) · [`java`](java/AGENTS.md) · [`go`](go/AGENTS.md) ·
   [`python`](python/AGENTS.md) · [`typescript`](typescript/AGENTS.md) ·
   [`kotlin`](kotlin/AGENTS.md)
3. **A thin `AGENTS.md` in each framework repo** — says what that repo is and
   links back here.

> A fix to a rule that applies to all languages belongs **here**. A fix specific
> to one language belongs in that language's Layer-2 file. When those and a
> deeper architecture document disagree, the architecture document wins — fix the
> guide.

> Write a rule that governs a human as well as an agent in its topic document. A
> release rule goes in [`VERSIONING.md`](VERSIONING.md), and a commit rule goes in
> [`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md). Point at the rule from this
> guide. This guide holds in full only the rules for how an agent works.

> **Before contributing, also read
> [`CONTRIBUTING.md`](https://github.com/valkyrjaio/.github/blob/26.x/CONTRIBUTING.md)**
> — the submission process, running CI locally, the commit/PR conventions, and
> branch targeting. This guide is the technical companion to it.

---

## 1. What Valkyrja is

Valkyrja is a single framework ported to five languages in priority order. PHP is
the **reference implementation**; every other port mirrors its structure,
naming, and tests.

| #   | Language       | Status                                | Package root / namespace |
| --- | -------------- | ------------------------------------- | ------------------------ |
| 1   | **PHP**        | Production — reference implementation | `Valkyrja\`              |
| 2   | **Java**       | In progress                           | `io.valkyrja`            |
| 3   | **Go**         | Proof of concept                      | `valkyrja`               |
| 4   | **Python**     | Planned                               | `valkyrja`               |
| 5   | **TypeScript** | Planned                               | `@valkyrjaio/valkyrja`   |
| 6   | **Kotlin**     | Planned (JVM — nearly free from Java) | `io.valkyrja`            |

Each language has parallel repos: the **framework** (runtime, zero build/AST
dependencies), **sindri** (the dev-only build tool that generates the cache), an
`application` example, a `template` skeleton, and `entry/*` server adapters. The
build tool is called `sindri` and is never a production dependency.

The **`template` repo is the structural source of truth** — it defines how a
repo's directories, files, and classes are laid out. Every new repo in that
language is scaffolded from it (see §3, rule 9).

Use the shared vocabulary (app, module, component, tool) consistently — see
[VOCABULARY.md](https://github.com/valkyrjaio/.github/blob/26.x/VOCABULARY.md).

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
- **Handler signatures are typed.** Handlers are explicit
  typed closures — HTTP → `ResponseContract`, CLI → `OutputContract`, Listener →
  `any`. Parameters are `(ContainerContract, map<string, mixed>)`; request/route
  come from the container, not the signature. `#[RouteHandler]` / `@RouteHandler` /
  `@route_handler` is a **metadata marker only**, never an active registrar.
- **`AppConfig` is the build tool entry point.** No `valkyrja.yaml`. The app
  config class already lists the component providers; `sindri` reads it via AST.
- **A component config holds only component-wide settings.** Each adapter gets
  its own config contract and its own default implementation. The component's
  service provider publishes each contract as a separate container binding. A
  component config that holds every adapter config forces an application to
  construct configuration for adapters that the application never uses. An adapter
  contract prefixes every property with the adapter name, so one application
  config class can implement several adapter contracts. See §4.
- **No provider-reference constants class.** Provider references use
  `::class` / `.class` / class objects / constructor references directly so
  `sindri` can resolve them statically. (Binding-_key_ constants files are fine
  and expected — see §4.)
- **Route middleware is appended, never deduplicated.** Across every protocol
  (HTTP, CLI, gRPC), both the runtime collector and `sindri` codegen _append_ each
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

- **Every code branch is tested** — _branch_ coverage, not just line coverage.
  Every path, guard, and error branch gets a test (use synthetic inputs to reach
  defensive guards that normal input can't). ("Branch" here means a code
  path, not a git branch.)
- **All tests pass.**
- **Every CI check passes** — the _full_ gate (static analysis, formatting,
  architecture, migration, tests), never a subset.
- **Coverage is and stays 100%** — line _and_ branch. It must never drop.

### Coverage is per file, and you must measure it

The 100% rule is **per file, not an aggregate**, and it binds both directions:
every file you add is at 100% on its own, and every file you touch stays at
100%. A green gate is not proof — no language's gate currently fails on
coverage — so read the coverage report yourself, per file, before you call a
change done. Code that genuinely cannot be covered is excluded narrowly, in the
coverage tool's own config, with a comment saying why. The full rules:
[`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md).

Then:

1. **Port code and its tests together**, never as a later pass. Mirror the source
   repo's test layout and map the framework (e.g. PHPUnit → Vitest: `assertSame`
   → `toBe`, data providers → `it.each`, `setUp` → `beforeEach`).
2. **End every file with a trailing newline.**
3. **American English** in all prose and identifiers ("color", "normalize"), and
   **Simplified Technical English** in documentation prose, with a code example
   for every rule that has a code shape (see §8).
4. **Every source file carries the license header** (see §5).
5. **Target the right branch** (see §7) — improvements/bug fixes go to the lowest
   affected `??.x`, new features/deprecations go to `master`. Warning: an
   exception suspends part of this rule today. See
   [`VERSIONING.md`](VERSIONING.md).
6. **Run the full CI gate** for the language you touched before considering the
   work done — exact commands are in your language's Layer-2 guide.
7. **One branch and one PR per change.** Create a new branch off the correct
   target branch, then commit with the `[Root] type:` message, push, and open a PR
   (base = that target branch) with the template filled out. **Ask for
   confirmation before committing, before pushing, and before opening the PR.**
   Keep each branch/PR small and atomic. A new component is more than one
   change, and the contracts land in their own pull request first. See §7.
8. **Cross-language changes propagate.** If a change affects more than one port,
   make it in every affected language in the _same_ batch (code and tests
   together). Open each PR standalone, and never cross-link the siblings. See
   §7.
9. **New repos are scaffolded from the language's `template` repo** — the source
   of truth for repo layout and file/class structure. Start from it; never
   hand-assemble a repo's structure. Your Layer-2 guide names the template repo.
10. **Never comment transient state.** A comment in code or config must stay
    true indefinitely, so a temporary condition never goes in a comment.
    Automation rewrites values, not the prose around them, so the comment
    becomes an assertion that is now false. Put the explanation in the PR
    description instead ([`PR_DESCRIPTION.md`](PR_DESCRIPTION.md)). The squash
    merge writes the description into git history as the commit body, so
    nothing is lost. Full rules, the decision-or-condition test, and examples:
    [`COMMENTS.md`](COMMENTS.md).
11. **Update the documentation in the same pull request.** A change to behavior,
    to configuration, or to a public API also updates every document that
    describes it. The component's `README.md` is the usual one, because each
    component documents its configuration and its container bindings there.
    Never leave the documents for a later sweep.
    A document that describes the old behavior is worse than no document,
    because the reader trusts it. This is the same failure that rule 10
    describes for a comment: the file now asserts something that is false. A
    later sweep also loses the reason for each edit, because the sweep is one
    large commit that no longer sits next to the change it explains.
    **The documentation is part of the definition of done.** A pull request that
    changes an env constant, a config property, a container binding, or a
    contract is not complete while a `README.md` still lists the old one.
12. **A workflow caller in GitHub Actions obeys the organization's rule on
    secrets.** The rule forbids `secrets: inherit` and states what a caller
    passes instead. The rule governs every repo, because each repo has its own
    callers in `.github/workflows/ci.yml`. Read the rule before you add or edit
    a caller:
    [`AGENTS.md` in `valkyrjaio/.github`](https://github.com/valkyrjaio/.github/blob/26.x/AGENTS.md).
13. **A comment states what the code cannot show.** Keep a comment to one or
    two lines: the constraint, the invariant, or the reason the obvious
    approach fails. Density is itself a defect. Comment the one line in ten
    that needs a comment. Full rules and examples:
    [`COMMENTS.md`](COMMENTS.md).
14. **A doc comment stays true for every override.** A doc comment describes the
    method, not one implementation of the method, because every override
    inherits the comment. Keep the inherited block as it is when you override a
    method; change only the parameter and return annotations. Inside a body,
    clear code takes no comment — a comment explains what is unclear, or why
    the code does something this particular way. Rule 10 describes the same
    failure for a comment that states a current condition; rule 13 limits how
    much a comment says. Full rules and examples: [`COMMENTS.md`](COMMENTS.md).
15. **A type declaration carries no doc comment.** A class, a contract, a
    trait, an enum, and a struct explain themselves:
    [`STRUCTURE.md`](STRUCTURE.md) encodes the kind in the name and the
    segment, and each method's doc comment — one sentence that
    enhances the signature, plus the annotations — states what the type does.
    The one exception is a test fixture's one-line block, which says what the
    fixture is for. Full rules and examples: [`COMMENTS.md`](COMMENTS.md).
16. **Verify a claim before you write it.** A sentence about another file is a
    claim, and you check a claim rather than assert it. Open the file, and read
    the code the sentence describes. This governs a comment, a `README.md`, a
    pull request description, and a review reply. Warning: a wrong claim
    outlives wrong code, because the next reader trusts prose and does not
    check it.
17. **Know how a call fails before you wrap it.** Name each way the call fails
    before you write the guard. A call reports a failure through a return
    value, through a warning, and through a throwable. A call can also do part
    of the work and report the count instead of a failure. Handle each one. A
    guard written from memory handles the failure the author remembered. See
    §7.
18. **Push work that is ready to review.** Review your own diff before you push
    it. See §7.

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
| ------- | ------------------------ | ------------------- |
| default | `App\Http\App`           | `app.http.App`      |
| variant | `App\Http\OpenSwooleApp` | `app.http.JettyApp` |

**Never nest a starter-app entry under the runtime** (`App\Http\OpenSwoole\App`). It
yields several classes named `App` inside one protocol — a dozen once gRPC and Queue
land — and the variant axis is not always an adapter, so it cannot be a segment:
[`QUEUE.md`](QUEUE.md) ships non-adapter `PullQueue` / `PushQueue` defaults, which
become `App\Queue\App` and `App\Queue\PushApp` alongside a per-runtime
`App\Queue\OpenSwoolePushApp`.

### Structure taxonomy (enforced)

A class's _kind_ is encoded three ways at once — its **name suffix**, the
**segment** (namespace/package/directory) it lives in, and its **modifier** —
and all three must agree, bidirectionally. The spec is machine-verified where
the language has an architecture linter, and review enforces it elsewhere. For
`Abstract`, `Enum`, and `Trait` the segment carries the meaning, so the name
must not repeat it — an abstract `Stream` is `Abstract\Stream`, never
`AbstractStream`. Each Layer-2 guide gives the per-language spelling. The full
table, the test-class rules, the author-docblock rule, and the examples:
[`STRUCTURE.md`](STRUCTURE.md).

### What a data object holds

A data object holds data and the behavior of its own state. A message, a model,
an entity, a type, and a config are all data objects. A data object does not
hold a static method: a **factory** takes construction, a **support class**
takes a calculation or a rendering, and an **enum** is exempt. One static
method stays on the data object — a named constructor that returns its own
type. The full rules, the examples, and the replacement for static metadata:
[`STATIC_METHODS.md`](STATIC_METHODS.md).

### Component config

A component gets one `ComponentNameConfigContract` for the settings that apply
to the whole component, and each adapter gets its own
`ComponentName<Adapter>ConfigContract`. The component config does not hold the
adapter configs, and an adapter contract prefixes every property with the
adapter name. Every contract has a default implementation that drops the
`Contract` suffix.

The default implementations live in the component's `Data\` segment. The
contracts live in the component's `Data\Contract\` segment. The service
provider publishes each contract as its own container binding. The full rules
and examples: [`COMPONENT_CONFIG.md`](COMPONENT_CONFIG.md).

### Method naming

The prefix on a method name tells a reader what the method does, and whether the caller's
own value changes. `validate` reports a failure, `isValid` returns a boolean, `parse`
changes the argument, `getParsed` returns a copy, `set` modifies the host, and `with`
returns a clone.

Two per-language caveats matter most. A language without pass-by-reference cannot honor
the in-place convention for an immutable type. **Go reports a failure with a returned
`error`, not a throw**, so a Go signature carries an extra return value.

The full table and the per-language spelling: [`METHOD_NAMING.md`](METHOD_NAMING.md).

### Binding-key constants

Per-component constants files (never one central file). The key is modeled on
how the port imports the class: the source file's directory path, written the
way that language writes a namespace, with the `Contract` segment removed
because the class name ends in `Contract` already. The key is therefore
language-specific — TypeScript writes
`Valkyrja.Container.Manager.ContainerContract`, and Go and Python write
`valkyrja.container.manager.ContainerContract`. PHP holds `::class`
strings, Java holds `.class` objects, Go/Python/TypeScript hold string
literals. Detail: [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md).

### Shell scripts

A repository in every language holds shell. A CI helper, a pre-commit hook, and
a script under `.github/ci/scripts/` are all shell, and every script runs under
`bash`. The rules:

- Use `[[ ]]` for a test, never `[ ]`.
- Give every `case` a default branch.
- Write a local variable in `lower_case`.
- Put a list of arguments in an array, never in a string.
- Give every suppression a reason.
- Match the `set` line to how the workflow runs the script.

SonarCloud and `shellcheck` enforce most of these rules.

Warning: shell inside a GitHub Actions `run:` block is invisible to both tools,
so a rule goes unenforced until the shell moves into a `.sh` file. That is a
reason to put the work in a script, not a reason to relax the rule.

The rules and their examples: [`SHELL_SCRIPTS.md`](SHELL_SCRIPTS.md).

### Port order for a new component

**Container → Event → Application → CLI → HTTP → Bin.**

---

## 5. License header (every source file)

Every file begins with the header (comment syntax adapted per language), and
languages that support it declare strictness first (PHP `declare(strict_types=1);`,
TypeScript `strict` via tsconfig, Java UTF-8 + JSpecify nullness):

```
This file is part of the Valkyrja Framework package.

Copyright (c) 2016-present Melech Mizrachi

Released under the MIT License. See LICENSE.md for details.
```

The first line names the package, and each repository has its own identifier for
it. `COPYRIGHT_HEADER.md` in the `.github` repo holds the identifier for every
repository, and it is the source of truth for the header. The two lines that
follow it are the same in every repository and in every language.

The year is 2016, because the first commit in `valkyrja-php` dates to October
of 2016. Every repository uses that year, including a port that a later year
created, because each port is a translation of the same work. `LICENSE.md`
states the same year.

Other cross-language style: concrete classes are `final` where the language
supports it; override methods are marked (`#[Override]` / `@Override` /
`noImplicitOverride`); contracts are interfaces.

---

## 6. Testing (shared shape)

Every framework repo mirrors the same test taxonomy — the layout is consistent
across languages and must be preserved in ports:

| Kind           | Holds                                                              |
| -------------- | ------------------------------------------------------------------ |
| **Unit**       | one class in isolation; path mirrors the `src` path                |
| **Functional** | boots the app / exercises several units together                   |
| **Fixtures**   | reusable doubles used _by_ tests, named `*Fixture` (never `*Test`) |
| **Abstract**   | base test cases (not themselves tests)                             |

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
2. **Commit** — after confirmation — using the `[Root] type:` message format.
3. **Push** the branch — after confirmation.
4. **Ask whether to add the `claude-review` label**, and add it only if the user
   says so. See **Asking for a review** below.
5. **Open a PR** — after confirmation — with its **base set to that same target
   branch** and the PR template filled out (see below).

Keep each branch and PR small and atomic — one focused change per PR.

Commits and PR titles carry a **root** for where the change lands and a **type**
for what kind of change it is — neither may restate the other. Full rules, the
root kinds, and worked examples:
[`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md). The essentials:

```
[Root] type: Message.                 [Root] type!: Message.
[Root] type(#123): Message.           [Root] type(#123)!: Message.
```

- **Working-branch commit** — single line, trailing period required; the issue
  reference is permitted but not required. The period marks a ledger entry on a
  working branch; a permanent subject line takes none, which covers both the
  squashed PR title and any direct push to a protected branch.
- **PR title** — **no** trailing period, and the issue reference is required
  when an issue exists. The root and the type are chosen for the whole change,
  so neither has to match any one branch commit. The release bump reads the
  type on the title, and never the type on a branch commit. A change that adds
  a capability or marks the public API as deprecated takes `feat` or
  `deprecate` on the title. A change that breaks a public contract takes `!` on
  the title, even when no branch commit carries the marker.
- **Types:** `feat`, `fix`, `deprecate`, `docs`, `test`, `refactor`, `perf`,
  `style`, `build`, `ci`, `chore`, `revert`. Append `!` before the colon on
  anything that breaks a public contract. No type marks a change as automated —
  git records the author already.
- **Roots** are an **open vocabulary** governed by two rules: a root names a
  _thing_, never a kind of change or the actor that made it; and a root is never
  the repo's own identity (`[PhpCsFixer]` says nothing inside the phpcsfixer repo,
  everything inside the framework). That second rule is positional — `[Http]` is a
  root here because HTTP lives here. Examples: a module (`[Http]`), a concept
  (`[Provider]`, `[Middleware]`), `[Dependency]`, an external tool (`[Composer]`,
  `[PhpCsFixer]`), a port (`[Java]`), `[Git]` / `[Workflow]` / `[GitHub]` /
  `[Process]`, a version line (`[26.x]`), a release version (`[v26.6.1]`). Module
  roots take their source directory's spelling (`[Orm]`, not `[ORM]`). Expect the
  vocabulary to grow — when a repo has a thing worth naming, name it.
- **Breadth is not a root.** A change touching 20 modules is still about something —
  renaming every component throwable is `[Throwable]`, not `[All]`. One level down,
  the same logic replaces stacking: middleware across HTTP and CLI is
  `[Middleware]`, not `[Http][Cli]`. If no single root fits, the change is doing too
  much; split it. Stack only to narrow in a cross-language repo (`[Java][Http]`).
- There is no `[Documentation]`, `[Deprecation]`, or `[Tests]` root — the types
  carry that, which frees the root to name what the change is actually about,
  including on `ci` and `test` work (`[Http] ci:`, `[Container] test:`).
- **Root casing is exact** — write each root exactly as listed above; initialisms
  stay uppercase (`[CI]`, `[GitHub]`), never `[Ci]` / `[Github]`.
- The type is not decoration: `feat`, `deprecate`, and `!` drive the middle
  version component and everything else is a patch, so the type you choose is
  what determines the next release. See [`VERSIONING.md`](VERSIONING.md).
- No body / co-author lines unless explicitly asked. This governs the commits
  _you_ write; the squash merge takes its subject from the PR title and its body
  from the PR description, which is why that description is where durable
  explanation belongs (see §3, rule 10).
- **The PR description is permanent** — the squash merge writes it into git
  history as the commit body. The description follows the
  [PR template](https://github.com/valkyrjaio/.github/blob/26.x/.github/PULL_REQUEST_TEMPLATE.md),
  it holds the what and the why, and it keeps no sentence the diff already
  shows. Full rules:
  [`PR_DESCRIPTION.md`](PR_DESCRIPTION.md).

### Push work that is ready to review

**A push says the change is ready to review.** Review the diff yourself first,
against these guides, the way the reviewer reads it. A change that still moves
spends one review round for each push, and each round reports the same kind of
finding again.

Read the diff for these, because a self-review finds them and a compiler does
not:

- **A claim you did not check.** See §3, rule 16.
- **A test that passes for a wrong implementation.** See
  [`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md) §1.
- **A document that this change made false.** A `README.md`, a design document,
  or a doc comment describes behavior that the diff changed. A later commit on
  the branch also moves the code under prose that an earlier commit wrote. See
  §3, rule 11.
- **A rule with a shape you can count.** The word limits and the writing rules
  in [`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md).
- **A fix that answers part of the finding it came from.** A finding that lists
  criteria is a checklist. Read the fix back against the list before you resolve
  the thread.
- **A commit that reverses an earlier commit on the branch.** The diff of one
  commit does not show the reversal, and the pull request carries both.
- **A failure of the call you wrapped that no guard handles.** §3, rule 17. One
  call reports several failures, and each one takes a guard. This example
  writes the guards for the two the return value reports:

  ```php
  // fwrite returns false for a failed write, and returns a count below the
  // length for a partial write. Each return takes its own guard.
  $written = fwrite($stream, $data);

  if ($written === false) {
      throw new ComponentStreamWriteException('The stream write failed');
  }

  if ($written < strlen($data)) {
      throw new ComponentStreamWriteException('The stream wrote part of the data');
  }
  ```

**Apply a fix in the context of the whole change.** A finding quotes a few
lines. A fix that reads only those lines lands in a file that earlier fixes
already changed. The fix then reverses an earlier decision, or it falsifies
prose an earlier commit wrote. Read the whole file, and read the earlier
commits on the branch. Say in the message when a commit reverses an earlier
one.

**Write the prose last.** A `README.md` paragraph, a doc comment, and the pull
request description each describe code. Each one goes stale when the code
moves, so write each one when the code stops moving. Read each one again before
the push that carries it.

### Asking for a review

The `claude-review` label starts an automated review of the pull request. **Ask
the user before you add it, the same way you ask before a commit, a push, and a
pull request.** A review spends the user's own Claude subscription, so the user
decides, not you. Never add the label because the change looks worth reviewing.

Add the label **as you open the pull request**, when the user asks for one:

```bash
gh pr create --base 26.x --label claude-review --title "…" --body-file …
```

Warning: the label does not start a review on a pull request that is already
open. The trigger listens for `opened`, `ready_for_review`, `synchronize`, and
`reopened`, so a label added later is read on the next one of those. To start a
review then, close the pull request and reopen it:

```bash
gh pr close 123 && gh pr reopen 123
```

The label stays until it is removed. Each later push reviews the change again, so
tell the user that a labeled pull request keeps costing usage while the label is
on.

### Answering a review

**A pull request that carries the `claude-review` label is not done when it
opens. It is done when the review has reported and every finding has an
answer.** The label is a request you made, so the findings are work you asked
for.

Wait for the review. It runs after the other checks, and a busy organization can
hold it for many minutes. A pull request left the moment it opens is a pull
request whose findings nobody reads.

Warning: a finding can be wrong. **Verify each one against the code and the
documentation before you act on it.** A reviewer that states a fact you can check
has given you something to check, not something to obey. A wrong finding that you
"fix" makes the change worse, and it teaches the next reader that the wrong
statement was right.

Each finding gets one of three answers:

- **The finding is right.** Push the fix, then resolve the thread in the same
  turn. Do not wait for the reviewer to confirm.
- **The finding is wrong, or you disagree.** Leave the thread open and reply with
  the evidence. The user decides, not you.
- **The finding is right and out of scope.** Leave the thread open, open an
  issue, and reply with the number.

Warning: never resolve a thread to clear the list. An open thread means open
work, and resolving one that the code does not answer hides it. Resolve a thread
only when the change answers it.

The REST API cannot resolve a thread. Read the thread ids, then resolve each one
the change answers:

```bash
gh api graphql -f query='mutation { resolveReviewThread(input: {threadId: "PRRT_kwDOH…"}) { thread { isResolved } } }'
```

Resolving is an action on the pull request, so it uses the same account a push
uses.

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
| --------------- | ---------------------------------------------- | -------------- |
| Improvement     | Lowest major affected `??.x`                   | `improvement/` |
| Bug fix         | Lowest major affected `??.x`                   | `fix/`         |
| New feature     | `master`                                       | `feature/`     |
| Deprecation     | `master`                                       | `deprecation/` |
| Breaking change | `master` (unless a bug fix — open issue first) | `breaking/`    |
| Documentation   | Lowest major affected branch the docs apply to | `docs/`        |

Warning: the current-year branch overrides this table until the framework has
users. Every change targets that branch, including a new feature, a deprecation
and a breaking change. See [`VERSIONING.md`](VERSIONING.md).

### Contracts land first

A component lands as two pull requests or more, and the contracts land first:

1. A **contracts pull request** adds the contracts for a component, or for one
   sub-component, and nothing else.
2. An **implementation pull request** stacks on that contracts pull request, and
   it adds the classes that satisfy the contracts.

A port that carries a whole component in one pull request is too large to
review. A reviewer cannot read the full diff with care, so the review reports
style and misses the design. A contracts pull request is small, and it decides
the shape. A reviewer reads the contracts with no implementation in the diff,
and reviews the design on its own. The same reviewer then reads each
implementation against a contract that the project agreed already.

**Split further, per sub-component, when a component is large.** Each
sub-component then gets a contracts pull request and an implementation pull
request. A contracts pull request bases on the target branch, unless it needs a
contract that another contracts pull request adds. The stack for the Queue
component reads:

| Pull request                     | Branch                                       | Base                                         |
| -------------------------------- | -------------------------------------------- | -------------------------------------------- |
| Message contracts                | `feature/queue-message-contracts`            | `master`                                     |
| Message implementation           | `feature/queue-message`                      | `feature/queue-message-contracts`            |
| Middleware and routing contracts | `feature/queue-middleware-routing-contracts` | `feature/queue-message-contracts`            |
| Middleware implementation        | `feature/queue-middleware`                   | `feature/queue-middleware-routing-contracts` |
| Routing implementation           | `feature/queue-routing`                      | `feature/queue-middleware-routing-contracts` |

Warning: an implementation pull request bases on the contracts branch, not on
the target branch. GitHub moves the base to the target branch when the contracts
pull request merges.

```bash
# Right — the implementation pull request bases on the contracts branch.
gh pr create --base feature/queue-message-contracts \
  --title "[Queue] feat: Add the queue message implementation"
```

**What belongs in a contracts pull request.** The test is whether the file
declares the shape that a caller programs against, or whether the file does the
work:

- **A contract** — the contracts pull request. The contract is what the pull
  request is for.
- **An enum** — the contracts pull request. The cases are shared vocabulary, and
  an implementation depends on them.
- **A constant holder** — the contracts pull request. A binding key names a
  contract, so the key lands with the contract. See §4, _Binding-key constants_.
- **A throwable contract** — the contracts pull request. A caller catches the
  throwable contract, so the throwable contract is part of the shape.
- **A concrete exception** — the implementation pull request that throws it. The
  exception names a failure of one implementation.

**One contracts pull request carries two sets of mutually dependent contracts.**
Two sub-components can reference each other's contracts. The Queue middleware
contracts take a routing `RouteContract`, and HTTP has the same shape. Neither
set of contracts stands on its own, so one contracts pull request adds both. The
two implementations still land as two pull requests.

**A contracts pull request does not move coverage.** A contract declares no
executable line, so a contracts pull request adds no test. The 100% rule in §3
does not block a contracts pull request, and coverage does not drop.

Warning: an enum case that holds a method is executable code. The same pull
request adds the test for that method.

**The rule governs a pull request that adds a contract.** A pull request that
adds no contract is one pull request, as before. A new adapter behind an
existing contract is one pull request, and a bug fix is one pull request.

### Cross-language changes

If a change affects more than one language port, make it in **every affected
language in the same batch** — never a deferred follow-up. A bug fixed in the PHP
reference implementation that also exists in Java/TypeScript/etc. is fixed there
at the same time, code and tests together. Open one PR per language repo.

**Warning: never cross-link the sibling PRs.** The cost of a cross-link is a
wrong merge. GitHub expands a PR reference inline, next to the commits and the
threads of the PR that holds it. A reader then reads a sibling's commit as a
commit on this PR. The same reader reads a linked issue as closed while it is
open. This project merged a PR early for that reason. Keep each PR standalone.
Report the sibling list to the person who asked for the batch, and keep the list
out of the Description.

Full detail:
[CONTRIBUTING.md](https://github.com/valkyrjaio/.github/blob/26.x/CONTRIBUTING.md).

---

## 8. Documentation style — Simplified Technical English

Write documentation prose in **ASD-STE100 Simplified Technical English (STE)**,
the controlled English of aerospace maintenance manuals. Two groups read these
documents: contributors whose first language is not English, and coding agents.
Both groups fail on the same input — long sentences, passive voice, ambiguous
pronouns, and one idea written two ways. STE removes that input.

The essentials: one instruction per sentence, the active voice, one term for
one thing, and a repeated noun instead of an ambiguous pronoun. Write a full
sentence, and give a dash-joined afterthought its own sentence. Include a
sentence only when it changes what the reader does or decides. Every rule that
has a code shape gets a generic code example, and the example shows PHP first.
Write an example, never a copy of the source. When you edit an existing
document, rewrite the paragraph you touch, not the whole file.

The scope, the eleven writing rules, the selection rule, and the rules for code
examples: [`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md).

---

## 9. Where to read more

Read these in order when starting or extending a port:

1. [`PORTS.md`](PORTS.md) — per-language characteristics
2. [`STRUCTURE.md`](STRUCTURE.md) — the structure taxonomy
3. [`THROWABLES.md`](THROWABLES.md) — exception hierarchy
4. [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md) — binding keys & closures
5. [`HANDLERS.md`](HANDLERS.md) — handler contracts
6. [`DATA_CACHE.md`](DATA_CACHE.md) — provider contracts & cache generation
7. [`COMPONENT_CONFIG.md`](COMPONENT_CONFIG.md) — the component config shape
8. [`BUILD_TOOL.md`](BUILD_TOOL.md) — `sindri` implementation
9. [`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md) — testing & 100% coverage
10. [`METHOD_NAMING.md`](METHOD_NAMING.md) — method name prefixes
11. [`COMMENTS.md`](COMMENTS.md) — what a comment may state
12. [`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md) — the writing rules for documentation prose
13. [`PACKAGE_NAMING.md`](PACKAGE_NAMING.md) — package, registry, and source namespace names
14. [`SHELL_SCRIPTS.md`](SHELL_SCRIPTS.md) — the rules for shell
15. [`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md) — commit & PR title format
16. [`PR_DESCRIPTION.md`](PR_DESCRIPTION.md) — what a pull request description holds
17. [`VERSIONING.md`](VERSIONING.md) — version scheme & release automation
18. `{language}/PROVIDER_CONTRACTS.md` — full contracts + examples
19. `{language}/README.md` — port notes & priority order
20. `{language}/AGENTS.md` — the Layer-2 agent guide for that language
