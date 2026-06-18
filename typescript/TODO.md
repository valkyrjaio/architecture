# TypeScript

## TODOs

Missing badges for scrutinizer, coverage %, sonarcloud maintainability

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
