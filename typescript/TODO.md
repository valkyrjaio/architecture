# TypeScript

## TODOs

Missing badges for scrutinizer, coverage %, sonarcloud maintainability

### High priority — rename test `classes` → `fixtures`

**Cross-language change — mirror this in every port (Go, Java, PHP, Python) so
the test trees stay 1:1.** Rename the reusable-test-double directory from
`classes` to `fixtures`. "Fixtures" is the widely-understood term for reusable
test doubles/sample classes; "classes" is generic and reads oddly next to
`unit`/`functional`.

TypeScript specifics:

- Move `tests/classes/` (or wherever the reusable doubles live) → `tests/fixtures/`
  and update every import path across the Vitest suites.
- Update any path aliases / `tsconfig` `paths`, ESLint/Prettier globs, and Vitest
  coverage include/exclude that name the `classes` dir.
- Decide whether the `*Class` file/suffix convention also becomes `*Fixture`
  (preferred for full parity) or stays — pick one and apply it everywhere.
- Update `TESTING_METHODOLOGY.md` once the rename lands.

### Branch coverage in CI

Vitest coverage (istanbul or v8 provider) reports **branch coverage**. Set
`coverage.thresholds.branches` to **100** (alongside lines/functions/statements)
so every ternary / `&&`/`||` short-circuit / optional-chain / `switch` arm is
exercised both ways — line coverage can read 100% while a branch is half-tested.
PHP and Java are enforcing the same (see their `TODO.md` files).

## Port gaps (found while porting the Application tests)

These are places where the TS port lags PHP. Tests currently assert the **current
TS behavior**; update both the source and the tests when these land, so TS
matches PHP ~1:1.

- **Event module not ported.** `ApplicationComponentProvider.getComponentProviders()`
  returns `[Container]` (PHP: `[Container, Event]`). No `EventComponentProvider`.
  Kernel `getProviders()` for the default config yields 2 providers (PHP: 3).
- **Log module not ported.** PHP's `CliApplicationComponentProvider` /
  `HttpApplicationComponentProvider` include `LogComponentProvider`; TS does not.
- **View module not ported.** PHP's `HttpApplicationComponentProvider` includes
  `ViewComponentProvider` (+ renderer/template); TS does not.
- **`HttpApplicationComponentProvider` is a stub** — returns `[Container]` only
  (PHP returns 8 providers: HttpMessage/Middleware/Routing/Server/Log/View…).
- **`Valkyrja` kernel does not cache empty provider arrays** — `getEventProviders()`
  etc. recompute a fresh `[]` on every call (the cache guard is `length > 0`), so
  empty results are not reference-stable. PHP caches unconditionally.
- **`Config`/`CliConfig` use positional constructors** (11+ params); setting only
  `providers`/`callbacks` requires many `undefined`s. Consider an options-object
  constructor to match PHP named args.
- **Tests are not type-checked or linted.** `tsconfig.json` `include` is `src/**`
  only, and ESLint/Prettier run on `src` (+`bin`). `tests/**` is run by Vitest
  (esbuild, no type-check). Add `tests/**` to the type-check + lint + format scope.
- **PHP route/listener providers expose `getControllerClasses()`/`getListenerClasses()`;**
  TS omits them by design (no reliable annotations) — fixtures reflect the TS
  contracts (`getRoutes()` / `getListeners()` only).
- **No `Env` module / `Exiter`; functional run-loop half not ported.** PHP's
  `Functional/.../Entry/{Cli,Http}Test` drive a full `Cli::run()`/`Http::run()`
  with an `Env` (data-cache class names) + `Exiter`, asserting attribute-routed
  handlers fire and debug-mode data-publish behavior. TS has no `Env`/`Exiter`
  and no attribute routing, so only the **boot + container-service** assertions
  are ported; the route-running half is deferred until those land.
- **Response cache (`CacheResponseMiddleware`) not ported.** PHP caches responses
  and serves them on a later request. When porting it, build it the JSON way —
  PHP just switched off file generation: serialize the response to JSON (`class`,
  `statusCode`, `reasonPhrase`, `headers` as `{name, value}[]`, `body`, plus `uri`
  for redirects) on `terminated()`, and reconstruct it on `requestReceived()` by
  instantiating the stored response class with only its `headers` argument (the
  one constructor arg shared by every response subclass) and applying
  `withStatusCode`/`withReasonPhrase`/`withBody` (+`withUri` for redirects). Do
  **not** replicate PHP's old `ResponseFileGenerator`/`Support/Generator/FileGenerator`
  approach — those were removed in PHP; there is nothing to port from them. (TS
  currently only has `Http/Server/Middleware/SendingResponse/NoCacheResponseMiddleware`,
  which is unrelated.)

### Container namespace

- **No `NativeChildContainer`** — TS has only `ChildContainer`; the PHP
  `NativeChildContainerTest` has no TS counterpart (not ported).
- **No standalone `ProvidersAware`** — the providers-aware behavior (`register`,
  `publish`, deferred callbacks) is inlined into `Container`; PHP's
  `Manager/ProvidersAwareTest` is covered by the `Container` test instead.
- **No `Provides` trait** — TS service providers implement `ServiceProviderContract`
  directly; PHP's `Provider/ProvidesTest` has no TS counterpart.
- **`ChildContainer` does not inherit singleton *bindings* from the parent.** It
  overrides `isAlias`/`isService`/`isSingletonInstance`/`isDeferred`/`isPublished`
  (and the `get*WithoutChecks`) to fall back to the parent, but **not**
  `isSingletonBinding` — so a `bindSingleton` on the parent is visible to the
  child as a *service*, not a singleton binding. PHP inherits the binding.
- **`Container.getFallback` ignores `InvalidReferenceMode`** — it always throws
  `ContainerInvalidReferenceException`. PHP's `NEW_INSTANCE_OR_THROW_EXCEPTION`
  mode instead tries to instantiate the requested class. The `mode` parameter is
  currently a no-op.

### Event namespace (largely unported)

Only `EventData` and the `ListenerContract` / `ListenerProviderContract`
interfaces exist in TS. Missing (PHP has tests for all of these, with no TS
target yet):

- `Listener` data class (`Data/Listener`)
- `ListenerCollection` (`Collection/`)
- the Event **Dispatcher** (`Dispatcher/`)
- attribute-based listener **Collector** (`Collector/AttributesListenerCollector`)
- Event **ComponentProvider** / **ServiceProvider** (`Provider/`) — this is the
  missing `EventComponentProvider` referenced under the Application gap above
- `Listener` / `ListenerHandler` **attributes** (`Attribute/`) — no TS attributes

## Sindri

- Ship a standalone, downloadable executable on each release so Sindri can be
  used without installing it via npm.
    - TypeScript: bundle `bin/sindri` to a single JS file (esbuild/rollup) and
      produce a standalone binary — Node's **Single Executable Application**
      (SEA), or `bun build --compile` / `deno compile` / `pkg` — then attach it
      to the GitHub release as a release asset so it can be downloaded and run
      directly (`./sindri`).
    - This mirrors PHP shipping a **Phar** and Java shipping a runnable **jar**
      on release — see each language's `TODO.md` for the per-language task.
