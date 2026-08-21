# Testing & Coverage Methodology

How every Valkyrja repository is tested, and how to reach **100% code coverage (line and branch)**. This is language-agnostic guidance
with PHP specifics called out; ports (Go, Python, TypeScript, Java) mirror the same structure and recipes against their
equivalent tooling (see [CI_TOOLS.md](CI_TOOLS.md)).

> **Porting rule:** port code **and its tests together**, not as a later pass. The goal is 100% coverage — line and branch — in every
> language. Mirror the source repo's test directory layout and map the test framework (e.g. PHPUnit → Vitest:
> `assertSame` → `expect().toBe`, data providers → `it.each`, `setUp` → `beforeEach`).

---

## 1. The definition of done — coverage is per file

The 100% rule is **per file, not an aggregate**, and it binds both directions:

- **Every file you add is at 100%** — line and branch — on its own, before the
  change is done. A repo-wide percentage is not evidence: one fully untested new
  class hides inside a large, well-covered codebase and barely moves the total.
- **Every file you touch stays at 100%.** Adding a branch to an existing file
  means adding the test for it in the same change.

**A green gate is not proof of coverage.** Every repo _runs_ coverage and
publishes a report, but no language's gate currently **fails** on it — a build
at 55% passes exactly like one at 100%. That is the deliberate state for now,
and gating may be added later; either way the 100% requirement does not depend
on a tool enforcing it. So **read the coverage report yourself** before calling
a change done, and check the per-file numbers for the files you added or
changed, not just the summary line. If gating does arrive, treat it as a
backstop for what you missed — never as the thing that defines the rule.

The only exception is an **explicitly documented** one: code that genuinely
cannot be covered (a process-exiting call, a blocking server loop) is excluded
in the coverage tool's own config, narrowly, with a comment saying why. Two
rules about exclusions: an accepted gap must be _written down_ where the tool
reads it, never merely tolerated in silence; and never lower a threshold to
accommodate a gap — a floor set to "whatever we happen to be at" legitimizes
the gap and defeats the point. Cover the code, or exclude it narrowly and say
why.

### A test that cannot fail proves nothing

Coverage counts the lines a test runs. It does not check that the test reports a
wrong implementation. A test that calls the method under test on both sides of an
assertion runs every line. It still passes when the code is wrong.

Write the expected value as a literal. Then name the wrong implementation the
test still passes under. When you can name a real one, the test needs a
different expected value.

```php
// Wrong — both sides call the method under test, so the assertion holds
// whatever the method writes.
self::assertSame($message->getFormattedText(), file_get_contents($filepath));
```

```php
// Right — the literal fails when the code writes the unformatted text.
self::assertSame("\e[97;42mtext\e[39;49m", file_get_contents($filepath));
```

---

## 2. Repository anatomy

Every PHP repo under `php/` shares this shape:

```
src/<Namespace>/...                  # production code (PSR-4: Vendor\Namespace\ -> src/Namespace/)
tests/
  bootstrap.php                      # requires ../vendor/autoload.php (root vendor: deps + src + tests autoload)
  Tests/
    Abstract/<Repo>TestCase.php      # per-repo base test case
    Unit/...                         # unit tests              (namespace Vendor\Tests\Unit\...)
    Functional/...                   # functional/integration  (namespace Vendor\Tests\Functional\...)
    Fixtures/...                     # reusable test fixtures/doubles (real classes, not *Test.php)
.github/ci/<tool>/                   # one dir per CI tool, each with its own composer.json + vendor
  phpunit/phpunit.xml.dist           # bootstrap, <testsuite>, <source> include/exclude
```

Key conventions:

- **Test classes** end in `Test.php` (the phpunit `<testsuite>` matches `suffix="Test.php"`) and are `final`.
- **Namespace ↔ dir:** `autoload-dev` maps `Vendor\Tests\` → `tests/Tests/`. Test namespaces mirror the class under
  test (`Vendor\Tests\Unit\<MirrorOfSrcPath>`).
- **CI binaries** live under `.github/ci/<tool>/vendor/bin/` (e.g. phpunit at `.github/ci/phpunit/vendor/bin/phpunit`).
  Always drive tools through the **root `composer.json` script shortcuts** (`composer phpunit`, `composer psalm`, …).
- **Coverage** is produced by `composer phpunit-coverage` (`phpunit --coverage-text`); the `<source>` block defines what
  counts toward coverage.

### Test organization (the `Tests` namespace)

All tests live under a single root namespace `Vendor\Tests\` mapped to `tests/Tests/` (PSR-4 `autoload-dev`). Inside it,
a fixed taxonomy of sub-namespaces separates tests by kind — **this layout is consistent across every repo and must be
preserved in ports**:

| Sub-namespace               | Directory                 | Holds                                                                                                                                                                                                                     | `*Test.php`? |
| --------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| `Vendor\Tests\Unit\…`       | `tests/Tests/Unit/`       | Unit tests for a single class in isolation. The path **mirrors the `src` path** of the class under test (e.g. `src/View/Provider/ViewServiceProvider.php` → `tests/Tests/Unit/View/Provider/ServiceProviderTest.php`).    | yes          |
| `Vendor\Tests\Functional\…` | `tests/Tests/Functional/` | Functional/integration tests that boot the app or exercise several units together.                                                                                                                                        | yes          |
| `Vendor\Tests\Abstract\…`   | `tests/Tests/Abstract/`   | Abstract base test cases (e.g. `<Repo>TestCase`) — **not** themselves tests.                                                                                                                                              | no           |
| `Vendor\Tests\Fixtures\…`   | `tests/Tests/Fixtures/`   | Reusable real classes used **by** tests: doubles, fixtures, stub providers/entities/commands, sample attribute/trait/enum classes. Subdivided by concept (`Fixtures/Provider`, `Fixtures/Trait`, `Fixtures/Contract`, …). | no           |
| `Vendor\Tests\<X>` (root)   | `tests/Tests/*.php`       | Shared top-level helpers, e.g. `EnvClass` (a test `Env` subclass) used across suites.                                                                                                                                     | no           |

Notes:

- A repo only creates the sub-namespaces it needs — e.g. `application` (skeleton) currently has just `Functional`;
  `sindri` has `Abstract`, `Fixtures`, `Unit`; `valkyrja` has all of them plus the root `EnvClass`.
- Things in `Fixtures/` are production-shaped classes (named `*Class`, `*Provider`, etc. — never `*Test`) so the
  architecture rules (PHPArkitect) can assert "testable classes are named appropriately and are not tests."
- Rector fixture data is the one exception that lives **outside** `tests/Tests/` (see §3) — it is `require`d data, not
  autoloaded test code.

### Test base classes

- `Valkyrja\PhpUnit\Abstract\ValkyrjaTestCase` (from the `valkyrja/ci-phpunit` package) — base for most repos.
- `Valkyrja\PhpUnit\Abstract\ServiceProviderTestCase` — boots a `Container` with a base `Env`, `ApplicationContract`
  stub, and `Config`; set `protected static string $provider`. Use for testing service providers' `publishers()` and
  `publishX()` methods.
- Each repo defines `tests/Tests/Abstract/<Repo>TestCase.php` extending one of the above (or PHPUnit `TestCase`).

---

## 3. The 100% coverage goal — recipes by code shape

### Plain classes / services

Unit-test each public method. Lock observable behavior with `assertSame`.

### Service providers

Use `ServiceProviderTestCase`. For each entry in `publishers()`, invoke the publisher callback against the container and
assert the singleton was registered as the expected type. Cover any branching in `publish()` (e.g. debug-mode vs not).

### Config / Data classes (constructors that build arrays)

Instantiate and assert the resulting structure. These are often large but single-method — one instantiation covers them.

### CI-tool config repos (php-cs-fixer, phparkitect, phpstan, psalm, phpcodesniffer)

Per `php/TODO.md`: **test that the expected rules exist and are configured exactly as expected.**

- Call the config builder (`Rules::getConfig(...)`, `Rules::getRules(...)`).
- Assert the returned object type and top-level settings (e.g. risky allowed, finder passed through).
- **Lock the full rule set** with a single `assertSame($expectedCompleteArray, $config->getRules())`. This makes any
  added/removed/changed rule fail the test. (php-cs-fixer's `Config::getRules()` returns the array as-passed, without
  expanding rule sets, so the literal comparison is stable.)
- Custom architecture expressions (e.g. PHPArkitect `NotHaveAttribute`): build a `ClassDescription` via
  `ClassDescriptionBuilder` and test **both** branches (`describe()`, and `evaluate()` with and without a match).

### Rector rules

- Integration: extend `Rector\Testing\PHPUnit\AbstractRectorTestCase`; `provideData()` →
  `yieldFilesFromDirectory(...)`; `provideConfigFilePath()` → a config registering only the rule under test.
- **Fixtures and config live OUTSIDE the PSR-4 test tree** (they are `require`d/data, not autoloaded classes):
  `tests/Fixture/<RuleName>/*.php.inc` and `tests/config/<RuleName>/configured_rule.php`. The `Test.php` stays in the
  namespaced `tests/Tests/...`.
- Fixture format: input, then a line `-----`, then expected output. Omit the separator for "no change" cases.
- Add fixtures for **every reachable branch** (each transform path + each "skip/keep" guard). Build permutations from
  valid PHP only — illegal constructs (e.g. duplicate import names) are not testable.
- Defensive guards and `getRuleDefinition()`/`getNodeTypes()` that fixtures cannot reach: cover with a **plain unit
  test** that instantiates the rule (`new Rule()`) and calls `refactor()` directly with synthetic PhpParser nodes
  (e.g. a `Nop` for the unsupported-node branch, degenerate `Use_` nodes for `count`/`isset` guards).
- Exclude fixtures from formatting/standards: add `<exclude-pattern>*/tests/Fixture/*</exclude-pattern>` to the
  phpcodesniffer ruleset (fixtures are intentionally non-conforming).

### Entry / worker packages (roadrunner, openswoole, frankenphp, jetty, netty, tomcat) — solved

`run()` is an infinite worker loop bound to a runtime-only dependency (`Worker::create()`, blocking `$server->start()`,
`frankenphp_handle_request()`, `server.join()`), so it cannot be unit-tested as written. The pattern below is now the
established approach in PHP and Java, and every worker entry follows it:

1. **Isolate each irreducible runtime call behind its own small overridable seam** — one method that does nothing but
   make that call (`sendSwooleBody()`, `startServer()`, `waitForRequest()`, `readBody()`). Mark it excluded from
   coverage (`@codeCoverageIgnore`; a JaCoCo package exclusion in Java) — it holds no logic to test.
2. **Put every decision outside the seams.** Request marshaling, response emission, header folding and the dispatch
   pipeline are ordinary methods, tested directly against synthetic input.
3. **Drive the loop from a `Fixtures` subclass** that overrides the seams with doubles, so the worker's branches and
   failure paths run without a live runtime.
4. **Prove it end to end once.** A smoke test boots a real application and drives the whole per-request path —
   convert the native request → dispatch → emit through the native response — with only the irreducible runtime call
   doubled. The starter applications go further and start the real server, make a live HTTP request, and assert the
   rendered response, each gated on that runtime being available so it skips cleanly where absent.

The payoff is that the defects this shape hides get caught: a worker that ignored the runtime's request and rebuilt it
from globals, and one that never wrote the response back to the runtime at all, both looked fine until a test made a
real request.

---

## 4. Coverage gotchas & their fixes

| Situation                                                                                                                                                 | Effect on coverage                                                                     | Fix                                                                                                                                                                    |
| --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Constant-only or empty classes (`Constant/*Info`, empty `Rules {}`)                                                                                       | 0 executable lines → counted as 0/0 (vacuously 100%)                                   | No test needed.                                                                                                                                                        |
| Example/template files (`*.example.php`) duplicating real classes                                                                                         | Never autoloaded → permanently uncovered                                               | Exclude from the phpunit `<source>`: `<exclude><directory suffix=".example.php">…/src</directory></exclude>`.                                                          |
| Closing `];` of a multiline array used as a `??` fallback                                                                                                 | PHP attributes no opcode to that line → unreachable, even when the default branch runs | Extract the default array into a dedicated **constant collection class** (e.g. `OrkaReplacementCollection::CORE`); the method line becomes a single covered statement. |
| Defensive guards required by static analysis but unreachable by real input (`!isset($x[0])` after `count===1`, `!$c instanceof Comment` on a typed array) | Permanently uncovered via normal input                                                 | Cover with synthetic-node unit tests that construct the degenerate state directly. Do **not** delete the guard (static analysis needs it).                             |
| `getAttribute()` / dynamic-property `mixed` (Psalm `MixedAssignment` at errorLevel 1)                                                                     | n/a (analysis, not coverage)                                                           | Contain the `mixed` behind a typed helper (`getOriginalName(Node): ?Name`) or a declared-`mixed` interface return; avoid inferred-mixed.                               |

---

## 5. CI gate to run on every change

Run from the repo root via composer scripts (see [CI_TOOLS.md](CI_TOOLS.md) for roles):

1. `composer phpstan` — no errors.
2. `composer psalm` — no errors, 100% inferred.
3. `composer phpcodesniffer` — no errors.
4. `composer phpcsfixer` (auto-fix) then `composer phpcsfixer-check` — it commonly reformats arrays (`=>` alignment) and
   enforces a trailing newline; apply its fixes before committing.
5. `composer rector-check` — no changes suggested.
6. `composer phpunit-coverage` — green and 100% (line and branch).

Other standing rules: every file ends with a trailing newline; American English in prose; improvements/bug-fixes target
the lowest affected `??.x` branch, new features/deprecations target `master`.

---

## 6. Repo-by-repo status & notes

| Repo                                                       | Shape                                                                                                | Coverage approach                                                                                                                                                                                    |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `valkyrja` (framework)                                     | 27 modules, ~1140 src files                                                                          | Unit + functional under `tests/Tests`; providers via `ServiceProviderTestCase`; reusable doubles in `tests/Tests/Fixtures`; worker abstracts tested via faux-loop `run(config, requestCount)`. 100%. |
| `sindri`                                                   | AST readers + data-file generators                                                                   | Unit/functional over readers, generators, providers, commands. 100%.                                                                                                                                 |
| `application` (skeleton)                                   | Example app: Http/Cli apps, controllers, providers, commands, configs, data, models, ORM entity/repo | Functional tests booting the app (`App::app()` / `App::directory()`); resolve services from the booted container; views ship in `app/resources/views`. **`*.example.php` excluded from coverage.**   |
| `ci/phpunit`                                               | Test-case helpers                                                                                    | 100%.                                                                                                                                                                                                |
| `ci/phpcsfixer`, `ci/phparkitect`                          | Rule/config builders + custom expressions                                                            | Full-rule-set `assertSame` lock + custom-expression branch tests. 100%.                                                                                                                              |
| `ci/rector`                                                | Custom Rector rule + config                                                                          | Fixture integration test + synthetic-node unit test for guards. 100%.                                                                                                                                |
| `ci/phpstan`, `ci/psalm`, `ci/phpcodesniffer`              | Empty `Rules {}` + constant `Info`                                                                   | No coverable code (vacuously 100%).                                                                                                                                                                  |
| `entry/roadrunner`, `entry/openswoole`, `entry/frankenphp` | Worker entrypoints                                                                                   | **Deferred** — `run()` loops need a test wrapper.                                                                                                                                                    |

---

## 7. Mental model for ports

When porting a module to another language, reproduce **three things together**: the source class, its test class
(mirrored path + name), and the coverage outcome (100%). Translate the recipe, not just the code:

- A PHP `ServiceProviderTestCase` test → the target's DI-container provider test.
- A Rector fixture suite → the target migration tool's fixture suite (or skip if the role has no equivalent — see
  CI_TOOLS.md gaps).
- The same gotchas recur: exclude example/template files, extract un-coverable default literals to named constants,
  and unit-test unreachable defensive guards with synthetic inputs.
