# AGENTS.md — PHP (Layer 2)

Per-language guide for the **PHP** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file only records the PHP
**deltas**.

PHP is the **reference implementation** — when a port disagrees with PHP on
structure, naming, or tests, PHP is right unless the architecture docs say
otherwise.

---

## Layout & naming

- **New repos** are scaffolded from the **`project-template-php`** repo — the
  source of truth for repo/file/class structure (canonical rule: §3.9).
- **Namespace:** `Valkyrja\` (PSR-4, `src/Valkyrja/` → `Valkyrja\`). Build tool
  namespace is `Sindri\`. Tests: `Valkyrja\Tests\` → `tests/Tests/`.
- **Source:** `src/Valkyrja/<Module>/<SubComponent>/<Type>/<Class>.php`. Contracts
  in `Contract/` with the `Contract` suffix; shared behavior in `Abstract/`.
- **PHP version:** `>=8.4`. Every file starts with the license header and
  `declare(strict_types=1);`.
- **Style:** concrete classes `final`; overrides marked `#[Override]`. A service
  class implements its contract and carries no registration code — a service
  provider registers it. A static
  `make(ContainerContract $container, array $arguments = []): static` factory is
  an optional alternative, not a house rule; the framework does not use one. See
  [`README.md`](README.md).

### Class references

Reference a class with `::class`. Never write the name of a class as a string
literal.

PHP resolves `::class` against the `use` statements of the file, so the name is
correct when the file compiles. A string literal is text. The IDE does not rename
it, PHPStan and Psalm do not check it, and PHPArkitect does not see the
dependency. A typo in a string survives every check in the gate and fails at run
time.

```php
// Wrong — a rename leaves this string behind, and no tool reports it.
$container->setSingleton('Valkyrja\Jwt\Contract\JwtContract', $container->getSingleton($default));
```

```php
// Right — the name resolves through the use statement above.
$container->setSingleton(JwtContract::class, $container->getSingleton($default));
```

The rule holds everywhere a class is named: a container binding key, a provider
list, a test data provider, and an `@var` or `@param` annotation.

A configuration format is the one exception, because NEON, YAML and JSON have no
`::class`. Warning: a class name in a config file drifts from the code without a
failure. Keep the authoritative list in PHP with `::class`, and assert that the
config matches it. `valkyrja/ci-phpstan` does this — `Rules::getRules()` holds each
rule as `::class`, `rules.neon` registers the same rules for PHPStan, and a test
fails when the two disagree.

### Exceptions

`ValkyrjaThrowable` (interface) → abstract `ValkyrjaRuntimeException` /
`ValkyrjaInvalidArgumentException` → abstract `Component*` → concrete
`Component<Specific>Exception`. Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../STRUCTURE.md`](../STRUCTURE.md)) is **defined
here** and enforced by **PHPArkitect** (`composer phparkitect`; the rules live in
the `valkyrja/ci-phparkitect` package's `Rules` class). Segments are PascalCase
namespace parts exactly as in `STRUCTURE.md` — `Contract\`, `Provider\`, `Factory\`,
`Constant\`, `Attribute\`, `Exception\`, `Throwable\`, `Abstract\`, `Enum\`,
`Trait\`, `Type\`, `Model\`, `Entity\`, `Security\`, `Cli\Command\`. The other
languages adapt to PHP, not the reverse.

PHP nuances:

- Base `*RuntimeException` / `*InvalidArgumentException` are `abstract` and live
  in `Abstract\`; concrete exceptions are named `*Exception` in `Exception\`.
- Attributes carry `#[Attribute]` and live in `Attribute\`; non-attributes must
  not be attributed.
- **Test traits invert the src trait rule:** in tests a trait lives in `Trait\`
  **and is named `*Trait`**, whereas a src trait's name must _not_ contain
  `Trait`.

---

## Tests

- **Location:** `tests/Tests/{Unit,Functional,Fixtures,Abstract}` (plus a root
  `EnvClass.php`). Unit paths mirror `src/`.
- **Naming:** test classes end in `Test.php`, are `final`, methods `testX()`.
- **Base cases:** `Valkyrja\PhpUnit\Abstract\ValkyrjaTestCase`; providers via
  `ServiceProviderTestCase` (set `protected static string $provider`, assert each
  `publishers()` entry registers the expected singleton). Each repo also defines
  `tests/Tests/Abstract/<Repo>TestCase.php`.
- **Fixtures:** reusable, production-shaped doubles in `tests/Tests/Fixtures/…`,
  named `*Fixture` — never `*Test`. A fixture that is itself an enum, trait, or
  contract keeps that type's naming (`*Enum` / `*Trait` / `*Contract`).
- **Coverage: 100% (line and branch), never dropping** — every code branch has a
  test — via `composer phpunit-coverage`. On **`valkyrja`** that script measures
  lines only, because `--path-coverage` is too slow to bundle there; branches
  need `composer phpunit-path-coverage-parallel` and are nobody's gate, so check
  them before calling work done (see the CI tools section). Recipes & gotchas:
  [`../TESTING_METHODOLOGY.md`](../TESTING_METHODOLOGY.md).

---

## CI tools & how to run them

Every tool is isolated under `.github/ci/<tool>/` with its own `composer.json`;
binaries live at `.github/ci/<tool>/vendor/bin/`. **Always drive them through the
root `composer.json` script shortcuts** — check that file first for exact names.

| Role                         | Tool            | Command(s)                                      |
| ---------------------------- | --------------- | ----------------------------------------------- |
| Architecture enforcement     | PHPArkitect     | `composer phparkitect`                          |
| Static analysis              | PHPStan         | `composer phpstan`                              |
| Static analysis + taint      | Psalm           | `composer psalm` (`psalm-check`, `psalm-stats`) |
| Code standards               | PHP CodeSniffer | `composer phpcodesniffer`                       |
| Formatting                   | PHP-CS-Fixer    | `composer phpcsfixer` then `phpcsfixer-check`   |
| Automated migration          | Rector          | `composer rector` / `rector-check`              |
| Testing                      | PHPUnit         | `composer phpunit` / `phpunit-coverage`         |
| Branch coverage (`valkyrja`) | PHPUnit         | `composer phpunit-path-coverage-parallel`       |

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate, not a subset:

`phpstan` → `psalm` → `phpcodesniffer` → `phpcsfixer` then `phpcsfixer-check`
→ `rector-check` → `phpunit-coverage`.

If a `composer.json` changed: `composer validate --strict` (root) or
`composer validate --no-check-publish` (others).

### Check branch coverage before done

Every repo except `valkyrja` keeps `--path-coverage` inside its `coverage`
script, so `phpunit-coverage` already reports branches there and the gate covers
them. **On `valkyrja` it does not** — that script measures lines only, and
nothing in CI gates branch coverage, so a new uncovered branch will not fail
anything. **Check it yourself once the work is complete**, so nothing creeps in:

```bash
GAPS=1 composer phpunit-path-coverage-parallel
```

`GAPS=1` lists every file below 100% branch coverage, ranked by branches missing.
Compare against the same command on your base branch and account for any file
your change added to the list. This runs the suite as parallel per-component
shards (~3 min); the serial `phpunit-path-coverage` is 10x slower.

**Read per-shard numbers, never the merged ones, when deciding what is missing.**
Xdebug builds a different branch map for a function depending on whether it ran,
so merging shard data unions two incompatible maps and invents branches nothing
can execute — it has reported 9 missing branches for a file that is at 100% in
its own shard. Confirm any suspected gap against the owning shard first:

```bash
composer phpunit-path-coverage-shard Http
```

**Branch coverage counts blocks entered, not edges taken — so an always-true
guard reads 100%.** Xdebug marks a _basic block_ covered when execution enters
it. An `if` whose condition can never be false still runs entry → body →
after-if, so every block is hit and the function reports full branch coverage;
the never-taken edge is only visible in **path** coverage, which nothing gates.
`GAPS=1` therefore never lists an always-true guard, and a file's absence from
that list is not evidence its code is all reachable. A dead _sub-expression_
does show up, because it forms a block nothing enters. Both shapes came from the
same invariant in `Cli\Interaction`:

| Dead code                                              | branches                    | paths |
| ------------------------------------------------------ | --------------------------- | ----- |
| `Answer::isValidResponse()`'s unreachable first clause | 8/9 — listed by `GAPS=1`    | —     |
| `QuestionWriter::writeQuestion()`'s always-true `if`   | 3/3 — invisible to `GAPS=1` | 1/2   |

Per-function branch and path counts, which no script reports, come from a
`--coverage-php` dump:

```php
$cov = require 'coverage.cov';
$fn  = $cov['codeCoverage']->functionCoverage()[$file][$function];
// $fn->branches and $fn->paths; an entry is hit when ->hit !== []
```

A genuine gap is closed by a test. A branch that no test can reach is closed by
**changing the source**, not by excusing it — fold an exhaustive `match`'s last
arm into `default`, wrap an irreducible syscall in an overridable seam, delete a
condition that is provably dead. Reaching an otherwise-unreachable state by
subclassing to poke at `protected` state is _excusing_ it: it locks in semantics
the public API cannot produce, and the TypeScript port had exactly such a test
pinning "empty allowed responses accepts anything" until the dead clause was
removed. See the branch-coverage notes in [`TODO.md`](TODO.md) for the known
categories and their remedies.

---

## PHP-specific notes

- **`sindri` (build tool)** holds `nikic/php-parser` and all code generation. The
  legacy `cache:generate` command will break once handler logic ships — migrate
  to `sindri` before then. `sindri` is a dev-only dependency; the framework has
  zero AST deps.
- **CI-tool config repos** (`ci/*`) are tested by asserting the full rule set is
  configured exactly as expected (`assertSame` lock on `getRules()`), plus branch
  tests for any custom expressions/rules. See
  [`../TESTING_METHODOLOGY.md`](../TESTING_METHODOLOGY.md) §3.
- **Entry workers** (the per-runtime `Application\Entry\<Runtime>` HTTP workers —
  FrankenPHP, OpenSwoole, RoadRunner) reach **100% line + branch** coverage: each
  `run()` wraps its irreducible runtime call (`frankenphp_handle_request`,
  `$server->start()`, `Worker::create()`/`waitRequest()`) in a small overridable
  seam marked `@codeCoverageIgnore`, and a `Tests\Fixtures` subclass overrides the
  seams to drive the loop, its branches, and failure paths.
- **Dynamic route regexes keep their PCRE delimiters — never "fix" them away.**
  `Regex::START` / `Regex::END` are `/^` and `$/`, so a cached route regex reads
  `/^\/users\/(?<id>[0-9]+)$/`. That framing is **required**: `Matcher` calls
  `preg_match($regex, $path)`, and PCRE takes a delimited pattern — hand it a bare
  `^…$` and it errors ("Delimiter must not be alphanumeric, backslash, or NUL")
  rather than simply failing to match. Every other port strips the delimiters,
  because none of their engines take any — `java.util.regex.Pattern`,
  `new RegExp(string)`, Python `re.compile`, Go `regexp.Compile` — and each records
  that as a deliberate deviation in its own guide
  ([`../java/AGENTS.md`](../java/AGENTS.md),
  [`../typescript/AGENTS.md`](../typescript/AGENTS.md),
  [`../python/AGENTS.md`](../python/AGENTS.md),
  [`../go/AGENTS.md`](../go/AGENTS.md)). PHP is the odd one out **by necessity, not
  by neglect**. Working across ports, do not strip PHP's delimiters to match the
  others, and do not carry PHP's delimiters into a port that cannot use them.

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md).
