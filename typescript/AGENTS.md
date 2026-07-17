# AGENTS.md — TypeScript (Layer 2)

Per-language guide for the **TypeScript** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file only records the
TypeScript **deltas**. PHP is the reference implementation; mirror it.

---

## Layout & naming

- **New repos** are scaffolded from the **`project-template-ts`** repo — the
  source of truth for repo/file/class structure (canonical rule: §3.9).
- **Package scope:** `@valkyrjaio/*` (`@valkyrjaio/valkyrja`,
  `@valkyrjaio/sindri`, `@valkyrjaio/application`). Imports include the `.ts`
  extension, e.g.
  `import { Container } from '@valkyrjaio/valkyrja/Container/Manager/Container.ts'`.
- **Source:** `src/Valkyrja/<Module>/…` (PascalCase dirs mirroring PHP). Files:
  contracts `Contract/*Contract.ts`, abstract bases `Abstract/*.ts`, concrete
  exceptions `Exception/*.ts`, constants `Constant/*.ts`, data `Data/*.ts`,
  enums `Enum/*.ts`.
- Every file starts with the license header. `tsconfig` is **strict** (target
  ES2023, NodeNext) with `noUncheckedIndexedAccess`, `noImplicitOverride`,
  `exactOptionalPropertyTypes`.
- **Style:** contracts are `interface` (never `type` aliases); shared behavior in
  an `abstract class` implementing the contract; concrete classes extend it.
  Binding keys are `static readonly` string constants (no `::class` equivalent);
  provider class lists use **constructor references** (`Array<new () =>
  Contract>`) for direct runtime instantiation.

### Exceptions

Three abstract branches, all extending native `Error`:
`ValkyrjaThrowable` / `ValkyrjaRuntimeException` / `ValkyrjaInvalidArgumentException`
→ abstract `Component*` → concrete `Component<Specific>Exception`, which sets
`this.name` in its constructor. Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) applies, with
**PascalCase directory** segments mirroring PHP: `Contract/`, `Provider/`,
`Factory/`, `Constant/`, `Exception/`, `Throwable/`, `Abstract/`, `Enum/`,
`Type/`, `Model/`, `Entity/`, `Security/`.

TypeScript nuances:

- **No architecture linter** (no ArchUnit/PHPArkitect equivalent) — the taxonomy
  is enforced in **review**; keep it exact anyway.
- **No traits** — no `Trait/` segment; share behavior via an `Abstract/` base.
- **Interfaces erase at runtime** — contracts cannot be binding keys, so use the
  string-constant keys (see Layout & naming). Name suffixes match §4
  (`*Contract`, etc.).

---

## Tests

- **Location:** separate `tests/Tests/{Unit,Functional,Fixtures}` parallel to
  `src/`; unit paths mirror `src/`.
- **Naming:** `*.test.ts` (not `.spec.ts`). Runner is **Vitest** (v8 coverage,
  `include: ['src/**/*.ts']`).
- **PHPUnit → Vitest mapping:**

  | PHPUnit                | Vitest                       |
  |------------------------|------------------------------|
  | `assertSame`           | `expect().toBe`              |
  | `assertTrue/False`     | `expect().toBe(true/false)`  |
  | `assertInstanceOf`     | `expect().toBeInstanceOf`    |
  | `expectException`      | `expect(() => …).toThrow`    |
  | `@dataProvider`        | `it.each([...])`             |
  | `setUp` / `tearDown`   | `beforeEach` / `afterEach`   |

- **Fixtures:** reusable doubles in `tests/Tests/Fixtures/…`, named `*Fixture`
  (e.g. a `ServiceFixture` with a static `make(container, args)` factory).
- **Coverage: 100% (line and branch), never dropping** — every code branch has a
  test.

---

## Build & CI tools (npm)

Each tool runs from its own `.github/ci/<tool>/`; the root `package.json` exposes
script shortcuts (`cd .github/ci/<tool> && npm run …`). Check `package.json` for
exact names.

| Role            | Tool                       | Command(s)                          |
|-----------------|----------------------------|-------------------------------------|
| Type checking   | `tsc --noEmit`             | `npm run typescript` / `build`      |
| Static analysis | ESLint + typescript-eslint | `npm run eslint` / `eslint-check`   |
| Formatting      | Prettier (Biome is the arch-preferred alt) | `npm run prettier` / `prettier-check` |
| Dead code       | Knip                       | (as configured)                     |
| Testing         | Vitest                     | `npm run vitest` / `vitest-coverage`|

Prettier config: 4-space indent, single quotes, `printWidth: 120`, trailing
commas `all`.

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate, not a subset:

`npm run typescript` → `npm run eslint` (then `eslint-check`) → `npm run prettier`
(then `prettier-check`) → `npm run vitest-coverage`.

---

## TypeScript-specific notes

- **Framework source shipping:** must ship `.ts` source alongside compiled `.js`
  (the cache-optional runtime relies on constructor references being available).
- **`sindri` (build tool)** uses the TypeScript compiler API to generate the four
  cache data classes. Dev-only; the framework has zero AST/build deps.
- **Architecture-enforcement & security are known toolchain gaps** in TS
  (no strong ArchUnit/PHPArkitect equivalent, no dedicated taint scanner) —
  enforce those rules in review. See [`../CI_TOOLS.md`](../CI_TOOLS.md).

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md).
