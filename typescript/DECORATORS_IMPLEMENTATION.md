# Task: Add routing decorator/attribute support to Valkyrja TypeScript (framework + sindri + starter app)

## Goal

Bring the TypeScript port to parity with PHP/Java: let controllers and CLI
commands declare routes **declaratively via decorators** (`@Route`,
`@DynamicRoute`, `@RouteHandler`, request-method decorators, CLI command
decorators), which Sindri scans and merges into the generated data caches —
**in addition to** the existing imperative `getRoutes()` path (already shipped).

This is a NET-NEW capability: it was consciously left out when the TS port was
first laid out because decorators were assumed unavailable. They are available
now (see "Decorator strategy" below).

Work spans three repos, each with its own PR, cross-linked, released in order.
Work on the current-year `??.x`/`master` branch as the branch-targeting table
dictates, in **separate worktrees**, and **ask before committing / pushing /
opening any PR**. Follow each repo's `AGENTS.md` and the architecture guides
(`~/Dropbox/Sites/Valkyrja/architecture/AGENTS.md` + `typescript/AGENTS.md`) for
naming, structure, testing (100% coverage), CI, and commit/PR conventions
(`[Component] Message.` commits, no co-author lines).

## Repos & local paths

- Framework: `~/Dropbox/Sites/Valkyrja/typescript/valkyrja` (npm `@valkyrjaio/valkyrja`, GitHub `valkyrjaio/valkyrja-ts` — confirm the remote).
- Sindri: `~/Dropbox/Sites/Valkyrja/typescript/sindri` (npm `@valkyrjaio/sindri`, GitHub `valkyrjaio/sindri-ts`).
- Starter app: `~/Dropbox/Sites/Valkyrja/typescript/application` (npm `@valkyrjaio/application`, GitHub `valkyrjaio/valkyrja-starter-app-ts`).
- PHP reference (source of truth for the pattern): `~/Dropbox/Sites/Valkyrja/php/*`.

Current released versions: `@valkyrjaio/valkyrja` and `@valkyrjaio/sindri` are at
**26.3.0** (26.3.0 already shipped the imperative-`getRoutes()` generation fix and
generator golden tests — see sindri-ts PR #62).

## Current state (verified) — what exists and what doesn't

- The TS framework ships **no routing decorators**. Grep confirms zero
  `@Route`/`RouteHandler`/`Get`/etc. exports under
  `src/Valkyrja/Http/Routing/`. Routes are registered **imperatively**: a route
  provider's `getRoutes()` returns concrete `new Route(...)` / `new DynamicRoute(...)`
  instances (CLI: `new Route(name, description, handler, ...)`).
- Sindri (26.3.0) implements the imperative path: `RouteProviderReader.extractRoutes()`
  returns the route objects, and the Http/Cli generators emit them verbatim and
  derive the `paths`/`dynamicPaths`/`regexes` maps (mirroring the runtime
  `RouteCollection`). This is correct and shipped — do not regress it.
- Sindri **already has the AST readers for the decorator path but they are
  half-baked and unwired for real use**:
  - `src/Sindri/Ast/HttpRouteAttributeReader.ts`, `CliRouteAttributeReader.ts`,
    `ListenerAttributeReader.ts` scan `@Route`/`@DynamicRoute`/`@RouteHandler`
    (and CLI/listener) decorators on controller classes.
  - `RouteProviderReader.readFile()` reads a provider's `getControllerClasses()`
    method (the list of controllers to scan for decorators).
  - `GenerateDataFromAst.generateHttpData()/generateCliData()` iterate
    `providerResult.controllerClasses`, run the attribute reader, and merge the
    results — this code path runs today but is never exercised because no TS app
    uses `getControllerClasses()` / decorators.
  - **Known bugs in the attribute path** (found while shipping the imperative
    path, left untouched): `HttpRouteAttributeReader.buildRouteExpr()` builds a
    `new DynamicRoute(...)` with arguments in an order that does NOT match the
    framework constructor (`DynamicRoute(path, name, regex, parameters, handler,
    requestMethods, ...)`) — it omits `regex` and puts `parameters` where `regex`
    belongs; and `AstHttpDataFileGenerator.getRoutesAsContent()` compensates by
    **appending the computed regex as the last argument**, which is also wrong.
    These must be fixed so attribute-generated `DynamicRoute` expressions are
    constructor-correct.

## The PHP/Java pattern to mirror

- PHP `App\Http\Controller\HomeController` methods carry
  `#[Route(path, name, requestMethods)]` + `#[RouteHandler([HttpRouteProvider::class, 'handler'])]`
  (and `#[Middleware(...)]`, `#[Parameter(...)]` for dynamic routes).
- `App\Http\Provider\HttpRouteProvider::getControllerClasses()` returns
  `[HomeController::class]`, and its `getRoutes()` returns `[]`.
- Sindri scans the controllers' attributes AND reads any imperative `getRoutes()`,
  then **merges** both into the generated routing data. Both sources coexist.
- Read `php/application/app/src/App/Http/Controller/HomeController.php` and
  `HttpRouteProvider.php` for the exact attribute shapes to port.

## Decorator strategy (design decisions to make first)

> **These decisions have since been made and recorded.** See
> [`DECORATORS.md`](DECORATORS.md) for the design rationale — which options were
> tried and rejected, why class references must be thunked
> (`[() => Provider, 'method']`), and what carries over to the Python port.

TypeScript supports decorators two ways; **decide and document the choice**:

1. **TC39 Stage-3 (standard) decorators** (TS 5+). Class/method/accessor/field
   only. Standardized; the preferred direction.
2. Legacy `experimentalDecorators`. Avoid unless a hard blocker appears.

Key constraints, established during investigation:

- **Sindri reads decorators from source via AST (ts-morph) — it never executes
  them.** So the *generator* side needs no runtime decorator support; the readers
  just need the decorator syntax to be present and parseable.
- The **app runs via `node --experimental-strip-types`**. Decorators are syntax
  (not types), so they survive type-stripping — but whether they *execute* at
  runtime depends on Node's native decorator support. **Do a small spike first**
  to determine the runtime story and pick one of:
  - **(a) Scan-only markers** — the decorators are simple no-op/metadata
    functions; routes are materialized only through Sindri's generated data cache
    at build time; at runtime the cached data is used (no runtime decorator
    execution needed beyond Node accepting the syntax). Simplest; fits the
    cached-data model. Preferred unless the spike shows a problem.
  - **(b) Runtime-executing decorators** — mirror PHP's reflection-based
    `AttributeRouteCollector` with a runtime collector that reads decorator
    metadata when the app runs uncached. More work; only if (a) is insufficient.
- **Parameter decorators are NOT part of Stage-3.** PHP uses `#[Parameter]` on
  route parameters for dynamic routes. In TS, **fold parameter metadata into the
  `@Route`/`@DynamicRoute` decorator options** (e.g. a `parameters: [...]` field)
  instead of a separate parameter decorator.

## Scope of work, by repo (suggested order)

### 1. Framework — `typescript/valkyrja` (new feature → `master`, `feature/` branch)

- Add a routing decorator module mirroring PHP `Valkyrja\Http\Routing\Attribute\*`:
  `@Route`, `@DynamicRoute`, `@RouteHandler`, request-method decorators
  (`@Get`/`@Post`/…), `@Middleware`, and the CLI equivalents. Fold parameter
  info into the route decorator per the constraint above.
- If the spike chooses runtime-executing decorators, add the runtime collector
  (port PHP's `AttributeRouteCollector`). If scan-only, the decorators can be
  metadata-only.
- Port the corresponding PHP tests; keep 100% coverage.
- Release. This unblocks the app compiling against the decorators.

### 2. Sindri — `typescript/sindri` (improvement/fix → `26.x`)

- Fix the attribute-path bugs: `HttpRouteAttributeReader.buildRouteExpr()`
  DynamicRoute argument order (emit a constructor-correct
  `new DynamicRoute(path, name, regex, parameters, handler, requestMethods, …)`)
  and remove the regex-append hack in `AstHttpDataFileGenerator.getRoutesAsContent()`.
- Ensure attribute routes and imperative `getRoutes()` routes **merge correctly**
  in `generateHttpData`/`generateCliData` (both populate one file; no clobbering).
- Add/extend unit tests + the golden snapshots
  (`tests/Tests/Unit/Generator/Ast/Golden/`) to cover decorator-derived output.
  The generator golden tests and the app's `sindri-vitest` check are your safety
  nets — update goldens intentionally when output shape changes.
- Validate end-to-end against the starter app once it uses decorators (step 3):
  the decorator-generated routing data must be equivalent to the imperative one.
- Release.

### 3. Starter app — `valkyrja-starter-app-ts` (improvement → `26.x`)

- Convert `src/App/Http/Controller/HomeController.ts` to decorator-based routes
  and `src/App/Http/Provider/HttpRouteProvider.ts` to implement
  `getControllerClasses()` (returning `[HomeController]`) with `getRoutes()`
  returning `[]` — mirroring PHP. Do the same for the CLI side
  (`TestCommand` / `CliRouteProvider`).
- Update the `*.example.ts` stubs and the normal `vitest` stub tests as needed.
- Extend the isolated `.github/ci/sindri-vitest` check (added in PR #74) to assert
  the decorator-derived generated routing data is populated and correct
  (same route names/paths/methods as before the conversion).
- Bump `@valkyrjaio/valkyrja` + `@valkyrjaio/sindri` to the releases from steps 1–2.

## Verification

- Build sindri (`npm run build`) and run it against the converted starter app
  (`node <sindri>/dist/bin/sindri.cjs generate src/App/Http/Config.ts`); confirm
  the generated `AppHttpRoutingData` has the same routes/paths/dynamicPaths/regexes
  as the current imperative output (7 HTTP routes incl. the dynamic
  `/{value}` route; CLI `test` command).
- Run each repo's full suite + tsc + eslint + prettier before proposing a PR;
  keep coverage at 100%. Cross-link the three PRs (framework + sindri + app) and
  the PHP/Java siblings.

## Out of scope (separate, tracked follow-up)

Container `deferredCallback` population is a **different** Sindri gap:
`ServiceProviderReader` can't yet resolve `publishers()` service-id keys that
reference framework constants (`ContainerServiceId.Data`) or same-file statics
(`ServiceProvider.HomeControllerId`), so container data generates empty. Do NOT
try to solve it here — but note that a parallel effort could add `@Service`-style
container decorators once routing decorators land.
