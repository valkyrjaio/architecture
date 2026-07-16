# AGENTS.md — Java (Layer 2)

Per-language guide for the **Java** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file only records the Java
**deltas**. PHP is the reference implementation; mirror it.

---

## Layout & naming

- **New repos** are scaffolded from the **`project-template-java`** repo — the
  source of truth for repo/file/class structure (canonical rule: §3.9).
- **Package root:** `io.valkyrja` (the `sindri` build tool uses `io.sindri`).
  Components map to `io.valkyrja.<component>` (`container`, `http`, `cli`,
  `dispatch`, `event`, `application`, `throwable`, …), with sub-packages by
  concern: `.contract` (interfaces), `.manager`/`.dispatcher`/… (impls),
  `.provider`, `.data`, `.annotation`.
- **Source:** `src/main/java/io/valkyrja/<component>/…`. Contracts named
  `*Contract`; the concrete implementation takes the bare name (`Router`,
  `Container`).
- **Java version:** 21 (Project Loom virtual threads). Every file starts with the
  license header; UTF-8 enforced; JSpecify nullness annotations (`@Nullable` /
  `@NonNull`) for static analysis.
- **Style:** concrete classes `final`; overrides `@Override`; framework data
  classes are **records** (compact constructors for defensive copying); deferred
  bindings are `Function<ContainerContract, ?>` lambdas — no reflection;
  `@Provides(Contract.class)` marks provider methods (runtime retention).

### Exceptions

All Valkyrja exceptions are **unchecked** (extend `RuntimeException`), no `throws`
declarations. Framework base `ValkyrjaInvalidArgumentException extends
java.lang.IllegalArgumentException` (keeps the parity name, extends the native
root). Then abstract `Component*` → concrete `Component<Specific>Exception`.
Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) is enforced by
**ArchUnit** (the PHPArkitect analog, run in the `archunit` CI build). Segments
are **lowercase** packages: `contract`, `provider`, `factory`, `constant`,
`exception`, `throwable`, `type`, `model`, `entity`, `security`, `command`.

Java nuances:

- **Reserved words → trailing underscore.** `abstract` and `enum` are Java
  keywords and cannot be package names — use **`abstract_`** and **`enum_`** for
  those segments. The §4 *name* rules still hold (an abstract class's name must
  not contain `Abstract`; an enum's must not contain `Enum`).
- **No traits.** Java has no trait construct, so there is no `trait` segment —
  share behavior via abstract classes or interface `default` methods.
- **Attributes → annotations.** The attribute marker is a Java annotation
  (`@interface`); annotation types live in an `annotation` package.
- Name suffixes are identical to §4 (`*Contract`, `*ServiceProvider`,
  `*Exception`, `*Factory`, …).

---

## Tests

- **Location:** tests live in the **JUnit CI build**, not the main source tree:
  `.github/ci/junit/src/test/java/io/valkyrja/`.
- **Packages mirror the PHP taxonomy** in a **parallel** package (prefer testing
  through the public API):
  - `io.valkyrja.unit.<ns>` — unit tests, `<Class>Test`
  - `io.valkyrja.functional.<ns>` — functional tests
  - `io.valkyrja.fixtures.<ns>` — reusable concrete doubles / marker types
  - (use the repo's own root where applicable, e.g. `io.sindri.unit.*`)
- If a test genuinely needs package-private access, place just that test in the
  source package instead.
- **Framework:** JUnit 5 (`junit-jupiter`) + **Mockito** for stubbing *contracts*;
  port concrete PHP `Fixtures` doubles to real Java classes under
  `io.valkyrja.fixtures.*`.
- **Coverage:** **JaCoCo — 100% (line and branch), never dropping** — every code
  branch has a test. Exclude only Java-only, non-unit-testable infra (e.g.
  `**/benchmark/**`, some `entry/*` adapters).

---

## Build & CI tools (Gradle)

Build system is **Gradle (Kotlin DSL)**. Each tool runs from its own
`.github/ci/<tool>/` via delegated tasks. Check the root `build.gradle.kts` for
exact task names.

| Role                     | Tool                          | Task                          |
|--------------------------|-------------------------------|-------------------------------|
| Formatting               | Spotless (Google Java Format) | `spotlessApply` / `spotlessCheck` |
| Architecture enforcement | ArchUnit                      | `archunit`                    |
| Static analysis          | ErrorProne + NullAway         | (compile-time)                |
| Static analysis + security | SpotBugs + FindSecBugs      | (bytecode)                    |
| Automated migration      | OpenRewrite                   | (recipe-based)                |
| Testing + coverage       | JUnit 5 + JaCoCo              | `junit`                       |

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).**
`./gradlew ci` runs the full gate: `spotlessCheck` → `archunit` → ErrorProne →
SpotBugs → `junit` (JaCoCo 100%). Use `./gradlew spotlessApply` to auto-format.

---

## Java-specific notes

- **Framework source shipping:** Java must publish a `-sources.jar` as a required
  build dependency (the cache-optional runtime needs source available).
- **`sindri` (build tool)** uses the Trees API + JavaPoet as an annotation
  processor to read `@Handler`/`@Provides` and generate the four cache data
  classes. Dev-only; the framework has zero AST/build deps.
- **`entry/*`** (jetty/netty/tomcat) are server adapters; Java's built-in
  `HttpServer` is the zero-dependency default.

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md).
