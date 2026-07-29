# TypeScript

## TODOs

Missing badges for scrutinizer, coverage %, sonarcloud maintainability

### Cross-language testing-gap audit

Compare this port's test suite against the other languages' and either close each
difference or record it in `AGENTS.md` as deliberate. Out of scope for the work
that prompted it; tracked here so it is not lost.

Prompted by a concrete miss: `sindri-java`'s golden snapshot never exercised the
dynamic-route regex path, so a framework regex-format change rode through a
dependency bump silently, while `sindri-ts` caught the equivalent change at once
([sindri-java#54](https://github.com/valkyrjaio/sindri-java/pull/54)). Only that
single gap was checked across ports — nobody has compared the suites broadly.

What to look for:

- Behavior a sibling port asserts that this one does not.
- An assertion pinned to a fragment where a sibling pins the whole value — a
  fragment survives the framing around it changing, which is exactly how the miss
  above happened.
- A generator, adapter, or component carrying snapshot/branch coverage on one
  side and none here.
- Test tooling that differs: what the static analyzers and formatters actually
  cover, and whether the test tree is inside or outside that scope.

Not every difference is a defect — some are forced by the language. Read the
per-language notes in `AGENTS.md` before "aligning" anything; the dynamic route
regex framing is the worked example (PHP must keep its PCRE delimiters, every
other port must not).

Known starting point: `sindri-ts` was checked against four specific gaps and had
none of them — it was in fact the port whose golden caught the framework regex
change first. Its suite has never been compared against PHP's more broadly, which
is the actual work here.

### Reset the reused deps branch in the update-dependencies workflow

`_ts-update-dependencies.yml` reuses an open `deps/update-dependencies-*` branch across runs,
committing each run's updates on top of the previous run's tree. Nothing a run writes to that
branch can be walked back by a later run, so a bad version pinned once stays pinned until
somebody deletes the branch by hand.

Java hit this concretely. An unfiltered root `dependencyUpdates` report bumped
`io.netty:netty-codec-http` to the 2015-era `5.0.0.Alpha2` prerelease and turned CI red
([starter-app-java#53](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/53)). Fixing
the filter ([#54](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/54)) did **not**
repair the branch: `useLatestVersions` only ever upgrades, so the re-run reported the alpha as
merely "exceeding" the latest and changed nothing. The PR had to be closed and the branch deleted.
The Java workflow now resets the reused branch to the base commit before running the updates
([.github#151](https://github.com/valkyrjaio/.github/pull/151)), so each run recomputes from
scratch and the branch always holds exactly `base + today's updates`.

Confirm first that this port's updater can actually strand a bad version the way
`useLatestVersions` does — a tool that rewrites every constraint to "latest" on each run may
already self-correct, in which case record that in `AGENTS.md` instead of adding the reset.

Tradeoff to weigh: the reset discards any commit pushed onto the deps branch by hand.

### Branch coverage in CI

Vitest coverage (istanbul or v8 provider) reports **branch coverage**. Set
`coverage.thresholds.branches` to **100** (alongside lines/functions/statements)
so every ternary / `&&`/`||` short-circuit / optional-chain / `switch` arm is
exercised both ways — line coverage can read 100% while a branch is half-tested.
PHP and Java are enforcing the same (see their `TODO.md` files).

### Centralize contract type guards (the `instanceof` equivalent)

Runtime discrimination between contracts is done with inline structural checks —
`'getStatusCode' in x`, `'getPath' in x`, `'writeMessages' in x` — scattered
across dispatchers/handlers. Because TS has no `instanceof` for interface
contracts, each of these is really an ad-hoc `instanceof`. Provide **one reusable
guard per contract** (e.g. `ResponseContract.instanceOf`,
`ServerRequestContract.instanceOf`, `RouteContract.instanceOf`,
`OutputContract.instanceOf` — the `Contract.instanceOf(value)` namespace helper
pattern already used by some contracts) and use it everywhere, so the
discriminating property lives in a single place.

This is a correctness issue, not just tidiness: `RequestHandler.dispatchRouter`
detected a returned response with `'getPath' in requestAfterMiddleware`, but a
`ServerRequest` has no `getPath`, so **every** request was misclassified as a
response and returned without being dispatched (fixed in `valkyrja-ts#82` by
checking `'getStatusCode'`). A shared `ResponseContract.instanceOf` guard would
have kept the one true check in one place and prevented the bug class. Audit
`Http`/`Cli` dispatch + middleware handlers for the same inline-`in` pattern and
replace them with the contract guards.

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

## VLID — cross-language parity

**Cross-language change — mirror in every port (Go, Java, PHP, Python).** VLID
(`Type/Vlid`) is PHP-only today; port it here (code + tests). It is the source of the
queue envelope `id` (a **VLID V1** — the longest, most-random version). Lock
cross-language parity:

- Port `Type/Vlid`, then add a conformance test: generate a VLID for **each version
  V1–V4** from a **fixed input timestamp**. Source the microsecond clock from
  `process.hrtime.bigint()` (Node) — `Date.now()` is millisecond-only.
- Assert this port produces a byte-identical **non-random portion** vs the PHP
  fixture — the encoded **microsecond timestamp** and the **version digit at
  position 14** must match exactly. The random bits differ by design; exclude them.
- This gate prevents timestamp-encoding / version-digit-placement drift from
  silently breaking cross-language `id` interop.
