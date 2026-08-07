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

The cross-language taxonomy ([`../STRUCTURE.md`](../STRUCTURE.md)) applies, with
**PascalCase directory** segments mirroring PHP: `Contract/`, `Provider/`,
`Factory/`, `Constant/`, `Exception/`, `Throwable/`, `Abstract/`, `Enum/`,
`Type/`, `Model/`, `Entity/`, `Security/`.

TypeScript nuances:

- **No architecture linter** (no ArchUnit/PHPArkitect equivalent) — the taxonomy
  is enforced in **review**; keep it exact anyway.
- **No traits** — no `Trait/` segment; share behavior via an `Abstract/` base.
- **Interfaces erase at runtime** — contracts cannot be binding keys, so use the
  string-constant keys (see Layout & naming). Name suffixes match `STRUCTURE.md`
  (`*Contract`, etc.).

---

## Tests

- **Location:** separate `tests/Tests/{Unit,Functional,Fixtures}` parallel to
  `src/`; unit paths mirror `src/`.
- **Naming:** `*.test.ts` (not `.spec.ts`). Runner is **Vitest** (v8 coverage,
  `include: ['src/**/*.ts']`).
- **PHPUnit → Vitest mapping:**

  | PHPUnit              | Vitest                      |
  | -------------------- | --------------------------- |
  | `assertSame`         | `expect().toBe`             |
  | `assertTrue/False`   | `expect().toBe(true/false)` |
  | `assertInstanceOf`   | `expect().toBeInstanceOf`   |
  | `expectException`    | `expect(() => …).toThrow`   |
  | `@dataProvider`      | `it.each([...])`            |
  | `setUp` / `tearDown` | `beforeEach` / `afterEach`  |

- **Fixtures:** reusable doubles in `tests/Tests/Fixtures/…`, named `*Fixture`
  (e.g. a `ServiceFixture` with a static `make(container, args)` factory).
- **Coverage: 100% (line and branch), never dropping** — every code branch has a
  test.

---

## Build & CI tools (npm)

Each tool runs from its own `.github/ci/<tool>/`; the root `package.json` exposes
script shortcuts (`cd .github/ci/<tool> && npm run …`). Check `package.json` for
exact names.

| Role            | Tool                                       | Command(s)                            |
| --------------- | ------------------------------------------ | ------------------------------------- |
| Type checking   | `tsc --noEmit`                             | `npm run typescript` / `build`        |
| Static analysis | ESLint + typescript-eslint                 | `npm run eslint` / `eslint-check`     |
| Formatting      | Prettier (Biome is the arch-preferred alt) | `npm run prettier` / `prettier-check` |
| Dead code       | Knip                                       | (as configured)                       |
| Testing         | Vitest                                     | `npm run vitest` / `vitest-coverage`  |

Prettier config: 4-space indent, single quotes, `printWidth: 120`, trailing
commas `all`.

### CI gate (run before done)

**Every check green, all tests pass, coverage 100% (line and branch).** Run the
full gate, not a subset:

`npm run typescript` → `npm run eslint` (then `eslint-check`) → `npm run prettier`
(then `prettier-check`) → `npm run vitest-coverage`.

### A package publishes by hand until its trusted publisher is configured

Warning: the `publish` job fails on **every** release of a package whose trusted
publisher npmjs does not hold. It is not a defect in the repository or in the
workflow, and it is not limited to the first release.

`_ts-release-npm-publish.yml` publishes with npm trusted publishing, which
authenticates through OIDC. npm attaches a trusted publisher **to a package**, and
it rejects a publish from an identity it does not hold one for. npm answers with
`404` rather than `403`, because a `403` would tell a stranger that the name
exists:

```
npm error code E404
npm error 404 Not Found - PUT https://registry.npmjs.org/@valkyrjaio%2f<package>
npm error 404  … could not be found or you do not have permission to access it
```

**Run the release workflow anyway.** The `release` job does the work that the
repository keeps: it computes the version, rewrites `package.json`, rewrites the
`*Info.ts` constants, writes `CHANGELOG.md`, and cuts the tag and the GitHub
release. Only the `publish` job fails, and it runs after all of that.

Then publish the tag by hand:

1. Check the tag out on its own, so the publish carries the released commit and
   not a branch that has moved.
2. Install the TypeScript CI dependencies **first**. `prepublishOnly` runs
   `npm run build`, and `tsconfig.json` resolves `@types/node` through
   `typeRoots`, which points into `.github/ci/typescript/node_modules`. Without
   that install, `tsc` stops with error `TS2688`, and it reports that it cannot
   find the type definition file for `node`.
3. Publish. npm asks for a one-time password and opens a browser.

```bash
git worktree add .worktrees/publish --detach v26.1.0
cd .worktrees/publish
(cd .github/ci/typescript && npm ci)
npm publish --access public
```

Note that a version published this way carries **no provenance**. `--provenance`
needs the OIDC token that only CI holds.

Configure the trusted publisher on npmjs, and name this repository, to stop the
failure recurring. Until a person does that, every release repeats the manual
step above. After it, each release publishes from CI, with provenance, and no
person authenticates.

Warning: a re-run of the failed job does **not** tell you whether the identity is
authorized. npm rejects a duplicate version before it checks authorization, so a
re-run against a version that already published returns
`You cannot publish over the previously published versions: <version>` whatever
the authorization state is. Read that error as "this version exists", and never
as "the identity is now authorized".

This was measured on `ci-eslint-ts`. The v26.0.0 publish failed with `E404`, a
person published the tag by hand, and the re-run then returned the duplicate
version error. That reads as authorization, and it is not: v26.1.0 failed with
`E404` again, from the same workflow and the same identity, against a package
that by then existed.

Note that provenance is not the signal either. npm signs the provenance statement
and writes it to the transparency log **before** the registry rejects the publish,
so a failed run still reports a signed statement.

---

## TypeScript-specific notes

- **Framework source shipping:** ships `.ts` source **only** — never compiled
  `.js`. Consumers compile the framework together with their own app (a loader
  such as `tsx`, or their bundler). Source must be present because the
  cache-optional runtime relies on constructor references and `sindri` reads that
  same source through the compiler API; publishing `.js` alongside it would split
  those into two module graphs, so the runtime could load a different copy of a
  class than the generated data cache references.
- **`sindri` (build tool)** uses the TypeScript compiler API to generate the four
  cache data classes. Dev-only; the framework has zero AST/build deps.
- **Architecture-enforcement & security are known toolchain gaps** in TS
  (no strong ArchUnit/PHPArkitect equivalent, no dedicated taint scanner) —
  enforce those rules in review. See [`../CI_TOOLS.md`](../CI_TOOLS.md).
- **Discriminate contracts with reusable type guards, not inline `in` checks.**
  TS has no `instanceof` for interface contracts, so runtime discrimination
  (request vs response vs route, etc.) is structural. Put the check in one place —
  a `Contract.instanceOf(value)` (or `isX(value): value is X`) guard co-located
  with the contract — and reuse it, instead of scattering ad-hoc `'prop' in x`
  checks across dispatchers. One canonical guard per contract keeps the
  discriminating property in a single spot, so a shape change updates one guard
  rather than N call sites, and a wrong property can't silently misclassify at one
  site (e.g. `RequestHandler.dispatchRouter` once used `'getPath' in x` to detect a
  response, but requests have no `getPath`, so every request was treated as a
  response — see [`TODO.md`](TODO.md)).

- **Never name a class in a decorator argument — thunk it.** A decorator argument
  is evaluated at class-definition time, so `[Provider, 'method']` dereferences a
  binding that may still be initializing (circular imports, or a class naming
  itself) and throws `ReferenceError: Cannot access 'X' before initialization`.
  Always write `[() => Provider, 'method']`; creating a closure captures the
  binding without reading it. Full rationale, rejected alternatives, and the
  Python implications: [`DECORATORS.md`](DECORATORS.md).

- **Dynamic route regexes are stored as native anchored patterns.** TypeScript
  sets `Regex.START` / `Regex.END` to `^` / `$`, not the PHP reference's
  PCRE-delimited `/^` / `$/`. `Matcher` compiles the stored regex with
  `new RegExp(regex)`, and a `RegExp` built from a string takes no delimiters — it
  would match the leading and trailing `/` as literal characters, so a
  delimiter-framed regex never matches. Java made the same move for
  `java.util.regex.Pattern`; PHP keeps its delimiters. Unlike Java, TypeScript
  needs only **one** test guard for this: `sindri` computes a dynamic route's
  regex inside `AstHttpDataFileGenerator` by calling the framework `Processor`, so
  the generator's golden snapshot already sits on the production path and pins the
  framework's exact output. (Java precomputes it a layer above the generator and
  therefore carries a second, end-to-end guard — see
  [`../java/AGENTS.md`](../java/AGENTS.md). That asymmetry is deliberate; do not
  collapse it into parity.)

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`DECORATORS.md`](DECORATORS.md).
