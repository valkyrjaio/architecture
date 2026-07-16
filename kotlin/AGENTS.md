# AGENTS.md — Kotlin (Layer 2)

Per-language guide for the **Kotlin** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file records the Kotlin
**deltas**. PHP is the reference implementation; mirror its behavior, adapting to
Kotlin idiom.

> **Brand-new port — no Kotlin repos exist yet.** The **first task** is to create
> **`project-template-kotlin`**, which will **mirror
> [`project-template-java`](https://github.com/valkyrjaio/project-template-java)
> very closely** — same layout, packages, and provider contracts — swapping in
> Kotlin idiom only where it clearly wins (`data class`, `sealed class`, `object`,
> Gradle Kotlin DSL). Once it exists it is the structural source of truth every
> other Kotlin repo is scaffolded from (canonical rule: §3.9). Until then there is
> **no** structural source of truth — do not hand-assemble repos.
>
> **Provisional.** Kotlin is "nearly free from Java" and shares the JVM runtime,
> package root, and provider contracts. Until a Kotlin decision doc exists, the
> **Java guide is the interim reference** — read [`../java/AGENTS.md`](../java/AGENTS.md)
> alongside this file; Kotlin follows it except where noted. Record Kotlin-only
> decisions here as they are made.

---

## Layout & naming

- **New repos** are scaffolded from **`project-template-kotlin`** (once created —
  see the note above) — the source of truth for repo/file/class structure
  (canonical rule: §3.9).
- **Package root:** `io.valkyrja` (shared with Java; JVM interop). Components map
  to lowercase packages exactly as Java.
- **Contracts:** `interface` named `*Contract`. **Abstract** → `abstract class`;
  **enums** → `enum class`; immutable data objects → `data class` (the
  Java-record equivalent); exception hierarchies may use `sealed class`.
- **No traits.** Share behavior via `abstract class` or interface default
  methods.
- Every file starts with the license header.

### Exceptions

All exceptions are **unchecked** (Kotlin has no checked exceptions). Framework
base `ValkyrjaInvalidArgumentException : IllegalArgumentException` (parity name,
native root), then abstract `Component*` → concrete
`Component<Specific>Exception`. Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) is enforced by
**Konsist** (Kotlin's ArchUnit analog). Segments are **lowercase** packages, same
as Java: `contract`, `provider`, `factory`, `constant`, `exception`, `throwable`,
`type`, `model`, `entity`, `security`, `command`.

Kotlin nuances:

- **Reserved words → trailing underscore, for JVM/Java parity.** Use
  **`abstract_`** and **`enum_`** as package segments (matching Java, since the two
  share packages). `abstract`/`enum` are Kotlin soft-keywords and could be
  backtick-escaped, but `_` keeps parity with the Java port. The §4 *name* rules
  still hold (no `Abstract`/`Enum` in the class name).
- **No traits** — no `trait` segment.
- **Attributes → annotations** in an `annotation` package.
- Name suffixes match §4 (`*Contract`, `*ServiceProvider`, `*Exception`, …).

---

## Tests

- **Layout:** mirror the Java port — tests in the JUnit CI build under
  `io.valkyrja.{unit,functional,fixtures}`, parallel packages, class names
  `<Class>Test`; reusable doubles in `io.valkyrja.fixtures.*`.
- **Framework:** JUnit 5; **MockK** for idiomatic Kotlin mocking (Mockito also
  interops).
- **Coverage:** **Kover** (or JaCoCo) — **100% (line and branch), never
  dropping**; every code branch has a test.

---

## Build & CI tools (Gradle)

Build system is **Gradle (Kotlin DSL)** on the JVM (Java 21). Each tool runs from
`.github/ci/<tool>/` as in the Java port.

| Role                     | Tool                    | Notes                          |
|--------------------------|-------------------------|--------------------------------|
| Formatting               | ktlint (via Spotless)   | Kotlin official style          |
| Static analysis / style  | detekt                  | Kotlin static analysis         |
| Architecture enforcement | Konsist                 | ArchUnit analog                |
| Automated migration      | OpenRewrite             | shared with Java               |
| Testing + coverage       | JUnit 5 + Kover         | 100% line and branch           |

- **Build tool (`sindri-kotlin`):** KSP (Kotlin Symbol Processing) — or the shared
  Java annotation processor — reads `@Handler`/`@Provides` and generates the four
  cache data classes. Dev-only; the framework has zero AST/build deps.

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate: ktlint/Spotless → detekt → Konsist → JUnit + Kover (100%).

---

Interim reference: [`../java/AGENTS.md`](../java/AGENTS.md),
[`../java/PROVIDER_CONTRACTS.md`](../java/PROVIDER_CONTRACTS.md). Kotlin appears in
[`../PORTS.md`](../PORTS.md) as a future JVM port.
