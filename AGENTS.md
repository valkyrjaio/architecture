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
- **Typed handler signatures, not dynamic dispatch.** Handlers are explicit
  typed closures — HTTP → `ResponseContract`, CLI → `OutputContract`, Listener →
  `any`. Parameters are `(ContainerContract, map<string, mixed>)`; request/route
  come from the container, not the signature. `#[Handler]` / `@Handler` /
  `@handler` is a **metadata marker only**, never an active registrar.
- **`AppConfig` is the build tool entry point.** No `valkyrja.yaml`. The app
  config class already lists the component providers; `sindri` reads it via AST.
- **A component config holds only component-wide settings.** Each adapter gets
  its own config contract and its own default class. The component's service
  provider publishes each contract as a separate container binding. A component
  config that holds every adapter config forces an application to construct
  configuration for adapters that the application never resolves. An adapter
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

- **Every file you add is at 100%** — line and branch — on its own, before the
  change is done. A repo-wide percentage is not evidence: one fully untested new
  class hides inside a large, well-covered codebase and barely moves the total.
  This is not hypothetical. A new class landed at 65% line / 50% branch while the
  repo-wide number only fell from 100% to 96%, and every local check passed.
- **Every file you touch stays at 100%.** Adding a branch to an existing file
  means adding the test for it in the same change.

**A green gate is not proof of coverage.** Every repo _runs_ coverage and
publishes a report, but no language's gate currently **fails** on it — a build at
55% passes exactly like one at 100%. That is the deliberate state for now, and
gating may be added later; either way the 100% requirement above does not depend
on a tool enforcing it. So **read the coverage report yourself** before calling a
change done, and check the per-file numbers for the files you added or changed,
not just the summary line. If gating does arrive, treat it as a backstop for what
you missed — never as the thing that defines the rule.

The only exception is an **explicitly documented** one: code that genuinely
cannot be covered (a process-exiting call, a blocking server loop) is excluded in
the coverage tool's own config, narrowly, with a comment saying why. Two rules
about exclusions: an accepted gap must be _written down_ where the tool reads it,
never merely tolerated in silence; and never lower a threshold to accommodate a
gap — a floor set to "whatever we happen to be at" legitimizes the gap and
defeats the point. Cover the code, or exclude it narrowly and say why.

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
   affected `??.x`, new features/deprecations go to `master`.
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
10. **Never comment transient state.** A comment in code or config must stay true
    indefinitely. Do not write one to explain something temporary, or something
    an automated process will later rewrite — a pinned version awaiting a
    release, a workaround pending a fix, a value some job regenerates. Automation
    rewrites values, not the prose around them, so the comment outlives what it
    described and becomes an assertion that is now false. That is worse than no
    comment, because the next reader trusts it.
    **Put it in the PR description instead — nothing is lost.** The squash merge
    writes the PR title as the commit subject and the _entire PR description_ as
    the commit body, so the explanation lives in git history permanently,
    attached to the commit that introduced it and reachable by `git log` /
    `git blame`. (This is also why the "no commit body" rule in §7 governs only
    the commits you write; the merge commit's body comes from the PR.) The
    explanation is better placed there anyway: pinned to when it was true,
    instead of floating in a file where a later automated edit silently
    falsifies it.
    The test is whether the comment states a _decision or invariant_ or a
    _current condition_. "This job asserts only generated code, so its coverage
    report is meaningless" is a decision — keep it. "Pinned ahead of the others
    until the next release bumps it" is a condition — the release automation will
    strand it, so it belongs in the PR.
    State the decision in the description in a sentence or two. The description
    takes the decision, not the essay around it — its budget is in §7.
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
13. **A comment states what the code cannot show.** Keep a comment to one or two
    lines: the constraint, the invariant, or the reason the obvious approach
    fails. Do not write a comment that narrates what the next line does — the
    code shows it. When an explanation needs a paragraph, put the paragraph in
    the PR description or in a document, and shrink the comment to one sentence
    that states the conclusion.
    Density is itself a defect. A file where every block carries a comment
    paragraph is a wall of text. A reader skips walls, so the one warning that
    matters goes unread. Comment the one line in ten that needs a comment, and
    let that warning stand alone.

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
**segment** (namespace/package/directory) it lives in, and its **modifier** — and
all three must agree. This is the machine-verified spec (PHP's PHPArkitect
`Rules` is the reference; Java ArchUnit and Kotlin Konsist mirror it; where a
language has no architecture linter — Go, Python, TypeScript — it is enforced in
review). PHP segment spellings are shown; **each Layer-2 guide gives the
per-language spelling** (case + reserved-word handling + constructs a language
lacks).

| Kind                                                   | Identified by                                | Name                               | Segment                        | Modifier  |
| ------------------------------------------------------ | -------------------------------------------- | ---------------------------------- | ------------------------------ | --------- |
| Contract                                               | is an interface                              | `*Contract`                        | `Contract\`                    | interface |
| Service provider                                       | implements `ServiceProviderContract`         | `*ServiceProvider`                 | `Provider\`                    | —         |
| Component provider                                     | implements `ComponentProviderContract`       | `*ComponentProvider`               | `Provider\`                    | —         |
| Route provider                                         | implements `Http`/`CliRouteProviderContract` | `*RouteProvider`                   | `Provider\`                    | —         |
| Listener provider                                      | implements `ListenerProviderContract`        | `*ListenerProvider`                | `Provider\`                    | —         |
| Factory                                                | —                                            | `*Factory`                         | `Factory\`                     | —         |
| Constant                                               | —                                            | `*Constant`                        | `Constant\`                    | final     |
| Attribute / annotation                                 | has the attribute marker                     | —                                  | `Attribute\`                   | —         |
| CLI command                                            | —                                            | `*Command`                         | `Cli\Command\`                 | —         |
| Security                                               | —                                            | `*Security`                        | `Security\`                    | final     |
| Concrete exception                                     | implements Throwable                         | `*Exception` (Go: `*Error`)        | `Exception\`                   | —         |
| Any throwable                                          | extends/implements Throwable                 | —                                  | `Throwable\`                   | —         |
| Base `*RuntimeException` / `*InvalidArgumentException` | —                                            | as-is                              | `Abstract\`                    | abstract  |
| Type / Model / Entity                                  | extends the base                             | —                                  | `Type\` / `Model\` / `Entity\` | —         |
| Abstract class                                         | is abstract                                  | must **not** contain `Abstract`    | `Abstract\`                    | abstract  |
| Enum                                                   | is an enum                                   | must **not** contain `Enum`        | `Enum\`                        | enum      |
| Trait                                                  | is a trait                                   | must **not** contain `Trait` (src) | `Trait\`                       | trait     |
| Test fixture                                           | reusable double in `Fixtures\`               | `*Fixture`                         | `Fixtures\`                    | final     |

The relationships are **bidirectional**: everything in `Contract\` must be an
interface _and_ named `*Contract`; everything in `Enum\` must be an enum; every
final constant lives in `Constant\`; and so on. For `Abstract`, `Enum`, and
`Trait` the _segment_ carries the meaning, so the **name must not repeat it** — an
abstract `Stream` is `Abstract\Stream`, never `AbstractStream`.

Tests: concrete test classes are `final`, live in `Unit\`/`Functional\`, and are
named `*Test`; reusable doubles live in `Fixtures\`, named `*Fixture`, never
`*Test`. A fixture that is itself an enum, trait, or contract keeps that type's
naming (`*Enum` / `*Trait` / `*Contract`) — the type rule supersedes the fixture
marker, just as the segment does for `Abstract`/`Enum`/`Trait` above. No class
carries an `@author` docblock.

### What a data object holds

A data object holds data and the behavior of its own state. A message, a model,
an entity, a type, and a config are all data objects.

A data object does not hold a static method. A static method describes the type,
not the state, so it belongs to another class. Three rules place it:

- A **factory** takes construction. A named constructor such as `create` or
  `fromArray` belongs on the `*Factory` for that type.
- A **support class** takes a calculation or a rendering.
- An **enum** is exempt. Every language gives an enum static members of its own
  (PHP `from`, `cases`), and a method on an enum case reads that case.

One static method stays on the data object: a named constructor that returns its
own type. `Header::fromValue()` returns a `Header`, so `Header` keeps it. The
rule removes a static method that returns something else — a field name, a table
name, a format, or a validation result.

```php
// Wrong — the trait puts a rendering on the entity, and it returns a string.
trait Dateable
{
    public static function getFormattedDate(): string
    {
        return DateFactory::getFormattedDate(static::getDateFormat());
    }
}
```

```php
// Right — the factory renders the date. The entity does not repeat the factory.
$date = DateFactory::getFormattedDate(DateFormat::DEFAULT);
```

```php
// Right — the named constructor returns its own type, so the type keeps it.
class StringT extends Type implements StringContract
{
    public static function fromValue(mixed $value): static
    {
        return new static(StringFactory::fromMixed($value));
    }
}
```

The rule reaches every segment that puts a static method on a data object. A
`Contract\` interface declares the method, an `Abstract\` base holds a default,
and a `Trait\` trait mixes the method in. Each one obeys the rule.

Static metadata is the hard case. A method such as `getTableName` describes the
type and carries no instance state. [`STATIC_METHODS.md`](STATIC_METHODS.md)
records the replacement: a metadata registry that the developer fills in a
service provider, keyed by class token.

Warning: a call on a _variable_ class does not port, whatever the method returns.
PHP writes `$type::fromValue($value)`, and no other Valkyrja language can. The
exemption above covers the declaration, not that call.
[`STATIC_METHODS.md`](STATIC_METHODS.md) owns the replacement for it.

### Component config

A component gets one `ComponentNameConfigContract` for the settings that apply to
the whole component. The default adapter is the most common such setting. Each
adapter then gets its own `ComponentName<Adapter>ConfigContract`. Every contract
has a default implementation that drops the `Contract` suffix (`CacheConfig`,
`CacheRedisConfig`), and all of these live in the component's `Data\` segment.
The component's service provider publishes each contract as its own container
binding.

Two rules make the shape work:

1. **The component config does not hold the adapter configs.** The container
   resolves an adapter config only when something asks for that adapter. An
   application that uses one cache adapter never constructs the configuration for
   the other cache adapters.
2. **An adapter contract prefixes every property with the adapter name.** One
   application config class can implement several adapter contracts at once.
   Without the prefix, two adapters that both declare a `prefix` property
   collide.

```php
// Wrong — the component config holds every adapter config. An application that
// uses only the null cache still constructs the redis and the log configuration.
interface CacheConfigContract
{
    public string $defaultCache { get; }
    public CacheRedisConfig $redisCache { get; }
    public CacheLogConfig $logCache { get; }
    public CacheNullConfig $nullCache { get; }
}
```

```php
// Right — the component config holds the component-wide setting only.
interface CacheConfigContract
{
    /** @var class-string<CacheContract> */
    public string $defaultCache { get; }
}

// Right — each adapter has its own contract, and each property carries the
// adapter prefix, so one class can implement several contracts at once.
interface CacheRedisConfigContract
{
    public string $redisHost { get; }
    public int $redisPort { get; }
    public string $redisPrefix { get; }
}

interface CacheNullConfigContract
{
    public string $nullPrefix { get; }
}
```

The application implements only the contracts for the adapters that it uses:

```php
final class AppConfig extends Config implements CacheConfigContract, CacheRedisConfigContract
{
    public function __construct(
        public string $defaultCache = RedisCache::class,
        public string $redisHost = 'cache.internal',
        public int $redisPort = 6379,
        public string $redisPrefix = 'app:',
    ) {
        parent::__construct();
    }
}
```

The service provider binds the application config when the application config
implements the contract. If it does not, the service provider binds the
framework default:

```php
public static function publishRedisConfig(ContainerContract $container): void
{
    $config = $container->getSingleton(ConfigContract::class);

    if ($config instanceof CacheRedisConfigContract) {
        $container->setSingleton(CacheRedisConfigContract::class, $config);

        return;
    }

    $container->setSingleton(CacheRedisConfigContract::class, new CacheRedisConfig());
}
```

### Method naming

The prefix on a method name tells a reader what the method does, and whether the caller's
own value changes. `validate` reports a failure, `isValid` returns a boolean, `parse`
changes the argument, `getParsed` returns a copy, `set` modifies the host, and `with`
returns a clone.

Two per-language caveats matter most. A language without pass-by-reference cannot honor
the in-place family for an immutable type. **Go reports a failure with a returned
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
a script under `.github/ci/scripts/` are all shell. SonarCloud reads a `.sh` file
in each repository, and `shellcheck` reads one on a developer's machine. Most
rules below are a rule of one of those two tools, so a script that breaks one
reports a finding.

Warning: shell inside a GitHub Actions `run:` block is invisible to both tools.
Neither one reads a workflow file, so a rule below goes unenforced until the
shell moves into a `.sh` file. That is a reason to put the work in a script, not
a reason to relax the rule.

**Use `[[ ]]` for a test, never `[ ]`.** `[[ ]]` is a shell keyword rather than a
command, so it parses the test rather than expanding it first. An empty variable
and a variable holding a space are then safe without a quote.

```bash
# Wrong — `[ ]` is a command, and its arguments expand before it runs.
if [ -n "$BRANCH_EXISTS" ]; then
    echo "The branch is there already."
fi
```

```bash
# Right — `[[ ]]` parses the test.
if [[ -n "$BRANCH_EXISTS" ]]; then
    echo "The branch is there already."
fi
```

Warning: quote the right side of a `=` inside `[[ ]]`. An unquoted right side is
a pattern, so `[[ "$a" = $b ]]` matches a glob where `[ "$a" = "$b" ]` compared
two strings.

**Give every `case` a default branch.** A `case` with no `*)` says nothing about
the value it does not name, and a reader cannot tell a deliberate skip from a
missing branch.

```bash
# Right — the default states that every other value is ignored.
case "$OUTCOME" in
    success) released=$((released + 1)) ;;
    timeout) timed_out=$((timed_out + 1)) ;;
    *) ;;
esac
```

**A local variable takes `lower_case`.** A script mixes a local with a global
that a function reports through, and case is what separates them.

```bash
# Right — `base_sha` belongs to the function, and `BRANCH_EXISTS` to the caller.
create_branch_if_needed() {
    local base_sha
    base_sha=$(gh api "repos/$ORG/$REPO/git/refs/heads/$BASE_BRANCH" --jq '.object.sha')

    BRANCH_EXISTS="$base_sha"
}
```

**Put a list of arguments in an array, never in a string.** A string of
arguments needs word splitting to work, and word splitting breaks on a value
that holds a space.

```bash
# Wrong — the string splits on every space, including one inside a name.
REVIEWER_FLAGS="--assignee $REVIEWER --reviewer $REVIEWER"
gh pr create $REVIEWER_FLAGS
```

```bash
# Right — the array carries each argument, whatever it holds.
REVIEWER_FLAGS=(--assignee "$REVIEWER" --reviewer "$REVIEWER")
gh pr create "${REVIEWER_FLAGS[@]}"
```

**A suppression states its reason.** A linter is wrong sometimes, and the next
reader must be able to tell a considered suppression from a silenced one.

```bash
# shellcheck disable=SC2001 # `^` anchors each line, and `${var//}` anchors nothing.
NEW_CONTENT=$(echo "$CONTENT" | sed 's/^name: Reusable/name: Z Reusable/')
```

Warning: the `set` line of a script that a GitHub Actions workflow runs depends
on how the workflow runs it. A `run:` step that names no shell gives `bash -e`
alone, and an explicit `shell: bash` gives `-eo pipefail`. A script that adds
`pipefail` where the block had none does not do what the block did. The rule and
the table are in
[`.github/workflows/README.md`](https://github.com/valkyrjaio/.github/blob/26.x/.github/workflows/README.md).

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
- **PR title** — same root and type, **no** trailing period, and the issue
  reference is required when an issue exists.
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
- PR description follows the
  [PR template](https://github.com/valkyrjaio/.github/blob/26.x/.github/PULL_REQUEST_TEMPLATE.md)
  — fill **Description**, **Types of changes**, and **Changes** (bold
  file/component — em dash — what changed). When an issue tracks the work, put
  `Closes #123` in the description: it becomes the squash commit body, so that is
  both what closes the issue on merge and where the link durably lives.
- **Budget the description.** State what changed, why, and any trap — most pull
  requests fit in three to six sentences. Add a table or a verification note
  only when it carries something a reviewer would otherwise miss. Do not
  restate the diff, and do not narrate the process that produced it. Rule 10 in
  §3 moves decisions into the description, and a decision is a sentence, not an
  essay. A reviewer skips a wall of text, so every extra sentence hides the one
  the reviewer needs.
- **The description holds the what and the why.** A pull request answers five
  questions. GitHub records the when and the where, and the diff records the
  how. The description holds the two the reader cannot derive: what changed,
  and why the change is right. Warning: a description that restates a specific
  edit binds the description to the code. A later push then makes the
  description false until someone corrects both together. Write "Update the Python binding
  keys, because the old keys were wrong", not "Change the key from
  `valkyrja.container.ContainerContract` to
  `valkyrja.container.manager.ContainerContract`". The first form stays true
  through every revision of the pull request.
- **A stable name sets the level of detail.** Warning: too little and too much
  fail the same way — the reader cannot tell what the pull request is about.
  "Update a method for naming consistency" names no method, and a restated
  signature buries the answer in detail. A class name and a contract method
  name survive every revision, so the description names them. Write "Rename
  `checkRoute` to `isValidRoute` to follow the method-naming families". The
  signature stays in the diff.

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

STE has two parts: a set of writing rules, and a dictionary of approved words.
The dictionary is a licensed ASD document, so this project does not certify
against it. Apply the writing rules in this section, and apply the dictionary's
core discipline: **one word has one meaning, and one meaning has one word.**
Technical names (`ContainerContract`, `sindri`, `26.x`) and technical verbs
(`serialize`, `cache`) are always permitted — STE calls these Technical Names
and Technical Verbs.

### What the style governs

It governs prose that this project writes for a reader:

- Markdown documents — `AGENTS.md`, `README.md`, and every design document here.
- Doc comments in source — docblock, Javadoc, TSDoc, docstring.
- PR descriptions, issue text, and commit message text.
- Strings the framework prints to a person — CLI help and exception messages.

It does not govern:

- Code, identifiers, file paths, commands, and program output.
- Quoted material from another source. **Never edit a quote to make it
  compliant.**
- Fixed third-party text, such as the license header in §5.

### The rules

1. **One instruction per sentence.** Keep an instruction to 20 words or fewer,
   and a description to 25 words or fewer.
2. **Use the active voice.** Name the actor. Write "`sindri` reads the config",
   not "the config is read".
3. **Use simple tenses.** Prefer the simple present. Do not use an `-ing` form as
   the verb of a sentence.
4. **Keep the articles.** Write "the provider", not "provider".
5. **Use three nouns or fewer in a noun cluster.** Break a longer cluster with a
   preposition or a relative clause.
6. **Use one term for one thing.** A contract is a "contract" in every document;
   do not call it an "interface" in the next paragraph.
7. **Repeat the noun instead of a pronoun** when more than one noun could be the
   referent.
8. **One topic per paragraph, six sentences or fewer.** Put three or more
   conditions or steps in a vertical list.
9. **Put the warning before the instruction.** State what breaks, then state what
   to do.
10. **No slang, no idiom, no humor.** They do not translate, and an agent reads
    them as fact.

Rules 1, 2, and 7 carry the most weight. This paragraph breaks all three:

> Because the generated cache must mirror what reflection produces exactly, and
> because a duplicate registration is considered to be the developer's error
> rather than something that ought to be silently repaired by the framework,
> middleware is appended by both the runtime collector and `sindri` without any
> deduplication step being applied to it.

The same content in STE:

> The runtime collector and `sindri` append each middleware in order. Neither one
> dedupes. A duplicate registration is the developer's error. The framework does
> not correct the duplicate, because the generated cache must match reflection
> exactly.

**A shorter document is not the goal.** STE often makes a dense paragraph longer,
because one run-on sentence becomes three plain ones. Count re-reads, not words.
A paragraph that a reader understands on the first pass beats a shorter paragraph
that the same reader must parse twice.

This holds for a decision log too, where compression is most tempting. A reader
opens a decision log to answer one question: what did we consider, and why did we
reject it? A run-on sentence hides exactly that answer. **No section of a
document is exempt** — a design record that a reader consults years later has the
most to gain.

### Select before you write

STE governs how you write a sentence. This rule governs which sentences you
write. Include a sentence only when it changes what the reader does or decides.
A document is not more complete because it is longer — every sentence the reader
must skip hides the sentence the reader needs.

Cut these before they reach the page:

- **Process narration.** "This was found after the code was written", "every
  example was checked before commit" — state the conclusion, not the journey.
- **Restatement of the artifact.** A PR description does not re-teach the
  section it adds. A comment does not repeat the code below it. A summary does
  not restate the diff.
- **Background the reader can derive.** The diff, the file, and the linked
  document already carry it.
- **The evidence for a rule.** A guide states the rule. The case that produced
  the rule is context, and a reader does not act on it.

Warning: evidence goes stale before the rule does. A rule that carries its own
evidence sends a reader to check the evidence, and the document then states
something that is no longer true. State the rule, and keep the case that
produced it in the pull request description.

**Add context only when all three of these are true:**

- A reader needs the context.
- The context is relevant to the rule.
- Someone asks for the context.

> Wrong — the rule carries its evidence, and the evidence names a port status
> that changes:
>
> A port that carries a whole component in one pull request is too large to
> review. The Go port shows the cost. One Go pull request added every component
> at once, and it closed without a merge.

> Right — the rule states what breaks, and nothing else:
>
> A port that carries a whole component in one pull request is too large to
> review. A reviewer cannot read the full diff with care, so the review reports
> style and misses the design.

"A shorter document is not the goal" (above) and this rule do not conflict. That
rule forbids compressing a needed sentence into a dense one. This rule forbids
writing a sentence nobody needs. Write every needed sentence plainly, and no
others.

### Code examples

Every rule that has a code shape gets a code example. Prose states the rule; the
example shows it. A reader who does not yet know the rule must be able to copy
the example and be correct.

- **State the rule in prose first.** An example never replaces the rule. The
  prose carries the reason, and the reason is what a reader needs to apply the
  rule to a case the example does not cover.
- **One idea per example, 20 lines or fewer.** Split a longer example.
- **Show the wrong form, then the right form**, when a rule is easy to break.
  Mark each form, and say in a comment why it is wrong or right.
- **Use real names from the framework.** Do not invent `Foo` or `Bar`.
- **Tag every fence with its language** — `php`, `java`, `ts`, `go`, `python`,
  `bash`. Highlighting and tooling read the tag.
- **Show PHP first**, because PHP is the reference implementation. Add a second
  example only for a language whose spelling differs. A Layer-2 guide shows only
  its own language.
- **Keep every example valid.** An example that does not compile teaches the
  wrong thing. Copy from real source where you can, and link to the file instead
  of pasting code longer than 20 lines.

The taxonomy rule in §4 — "for `Abstract`, `Enum`, and `Trait` the segment
carries the meaning, so the name must not repeat it" — takes this example:

```php
// Wrong — the class name repeats the segment.
namespace Valkyrja\Log\Logger\Abstract;

abstract class AbstractLogger implements LoggerContract {}
```

```php
// Right — the segment says "abstract", so the name does not.
namespace Valkyrja\Log\Logger\Abstract;

abstract class Logger implements LoggerContract {}
```

### When you edit an existing document

Rewrite the paragraph you touch, not the whole file. A large style-only rewrite
hides the change that the PR is about, and it makes the diff impossible to
review. A `docs:` PR may rewrite a full document for style alone, but then it
changes nothing else.

---

## 9. Where to read more

Read these in order when starting or extending a port:

1. [`PORTS.md`](PORTS.md) — per-language characteristics
2. [`THROWABLES.md`](THROWABLES.md) — exception hierarchy
3. [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md) — binding keys & closures
4. [`DISPATCH.md`](DISPATCH.md) — handler contracts
5. [`DATA_CACHE.md`](DATA_CACHE.md) — provider contracts & cache generation
6. [`BUILD_TOOL.md`](BUILD_TOOL.md) — `sindri` implementation
7. [`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md) — testing & 100% coverage
8. [`METHOD_NAMING.md`](METHOD_NAMING.md) — method name prefixes
9. [`PACKAGE_NAMING.md`](PACKAGE_NAMING.md) — package, registry, and source namespace names
10. [`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md) — commit & PR title format
11. [`VERSIONING.md`](VERSIONING.md) — version scheme & release automation
12. `{language}/PROVIDER_CONTRACTS.md` — full contracts + examples
13. `{language}/README.md` — port notes & priority order
14. `{language}/AGENTS.md` — the Layer-2 agent guide for that language
