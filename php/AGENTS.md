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
- **Style:** concrete classes `final`; overrides marked `#[Override]`; classes
  implementing `ServiceContract` expose a static
  `make(ContainerContract $container, array $arguments = []): static` factory.

### Exceptions

`ValkyrjaThrowable` (interface) → abstract `ValkyrjaRuntimeException` /
`ValkyrjaInvalidArgumentException` → abstract `Component*` → concrete
`Component<Specific>Exception`. Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) is **defined
here** and enforced by **PHPArkitect** (`composer phparkitect`; the rules live in
the `valkyrja/phparkitect` package's `Rules` class). Segments are PascalCase
namespace parts exactly as in §4 — `Contract\`, `Provider\`, `Factory\`,
`Constant\`, `Attribute\`, `Exception\`, `Throwable\`, `Abstract\`, `Enum\`,
`Trait\`, `Type\`, `Model\`, `Entity\`, `Security\`, `Cli\Command\`. The other
languages adapt to PHP, not the reverse.

PHP nuances:

- Base `*RuntimeException` / `*InvalidArgumentException` are `abstract` and live
  in `Abstract\`; concrete exceptions are named `*Exception` in `Exception\`.
- Attributes carry `#[Attribute]` and live in `Attribute\`; non-attributes must
  not be attributed.
- **Test traits invert the src trait rule:** in tests a trait lives in `Trait\`
  **and is named `*Trait`**, whereas a src trait's name must *not* contain
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
  test — via `composer phpunit-coverage`. Recipes & gotchas:
  [`../TESTING_METHODOLOGY.md`](../TESTING_METHODOLOGY.md).

---

## CI tools & how to run them

Every tool is isolated under `.github/ci/<tool>/` with its own `composer.json`;
binaries live at `.github/ci/<tool>/vendor/bin/`. **Always drive them through the
root `composer.json` script shortcuts** — check that file first for exact names.

| Role                     | Tool         | Command(s)                                     |
|--------------------------|--------------|------------------------------------------------|
| Architecture enforcement | PHPArkitect  | `composer phparkitect`                         |
| Static analysis          | PHPStan      | `composer phpstan`                             |
| Static analysis + taint  | Psalm        | `composer psalm` (`psalm-check`, `psalm-stats`)|
| Code standards           | PHP CodeSniffer | `composer phpcodesniffer`                   |
| Formatting               | PHP-CS-Fixer | `composer phpcsfixer` then `phpcsfixer-check`  |
| Automated migration      | Rector       | `composer rector` / `rector-check`             |
| Testing                  | PHPUnit      | `composer phpunit` / `phpunit-coverage`        |

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate, not a subset:

`phpstan` → `psalm` → `phpcodesniffer` → `phpcsfixer` then `phpcsfixer-check`
→ `rector-check` → `phpunit-coverage`.

If a `composer.json` changed: `composer validate --strict` (root) or
`composer validate --no-check-publish` (others).

---

## PHP-specific notes

- **`sindri` (build tool)** holds `nikic/php-parser` and all code generation. The
  legacy `cache:generate` command will break once handler logic ships — migrate
  to `sindri` before then. `sindri` is a dev-only dependency; the framework has
  zero AST deps.
- **CI-tool config repos** (`ci/*`) are tested by asserting the full rule set is
  configured exactly as expected (`assertSame` lock on `getRules()`), plus branch
  tests for any custom expressions/rules. See
  [`../TESTING_METHODOLOGY.md`](../TESTING_METHODOLOGY.md) §2.
- **Entry workers** (`entry/*`) have infinite `run()` loops that are currently
  coverage-**deferred**; helper methods are testable.

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md).
