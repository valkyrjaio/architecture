# TypeScript

## TODOs

### Missing badges

Coverage, SonarCloud maintainability, and Scrutinizer are all absent. The sibling Java
port carries six badges; this port carries four.

- [valkyrjaio/valkyrja-ts#142](https://github.com/valkyrjaio/valkyrja-ts/issues/142)

### Update the repo description once gRPC and Queue land

- [valkyrjaio/valkyrja-ts#140](https://github.com/valkyrjaio/valkyrja-ts/issues/140)

### Cross-language testing-gap audit

`sindri-ts` was checked against four specific gaps and had none of them. Its suite has
never been compared against PHP's more broadly.

- [valkyrjaio/valkyrja-ts#139](https://github.com/valkyrjaio/valkyrja-ts/issues/139)

### Reset the reused deps branch in the update-dependencies workflow

`_ts-update-dependencies.yml` lives in the shared `.github` repository, so the work is
tracked there.

- [valkyrjaio/.github#221](https://github.com/valkyrjaio/.github/issues/221)

### Centralize contract type guards (the `instanceof` equivalent)

Four dispatch sites still discriminate a contract with an inline structural check.
One shared guard per contract keeps the discriminating property in one place.

- [valkyrjaio/valkyrja-ts#135](https://github.com/valkyrjaio/valkyrja-ts/issues/135)

## Port gaps (found while porting the Application tests)

These are places where the TS port lags PHP. Tests currently assert the **current
TS behavior**; update both the source and the tests when these land, so TS
matches PHP ~1:1.

- **Event module not ported.** `ApplicationComponentProvider` returns `[Container]`;
  PHP returns `[Container, Event]`.
  [valkyrjaio/valkyrja-ts#136](https://github.com/valkyrjaio/valkyrja-ts/issues/136)
- **Log module not ported.** Only `LoggerContract` exists. Nothing implements or
  publishes it.
  [valkyrjaio/valkyrja-ts#143](https://github.com/valkyrjaio/valkyrja-ts/issues/143)
- **View module not ported.**
  [valkyrjaio/valkyrja-ts#144](https://github.com/valkyrjaio/valkyrja-ts/issues/144)
- **`HttpApplicationComponentProvider` omits Log and View.** It returns six providers;
  PHP returns eight. It is no longer the `[Container]`-only stub this file once
  described — the gap is now exactly the Log and View providers, tracked in the two
  issues above.
- **`Valkyrja` kernel does not cache an empty provider list** — the guard is
  `length > 0`, so an empty result is not reference-stable.
  [valkyrjaio/valkyrja-ts#145](https://github.com/valkyrjaio/valkyrja-ts/issues/145)
- **`Config`/`CliConfig` use positional constructors** (11 params on `Config`).
  [valkyrjaio/valkyrja-ts#146](https://github.com/valkyrjaio/valkyrja-ts/issues/146)
- **PHP route/listener providers expose `getControllerClasses()`/`getListenerClasses()`;**
  TS omits them by design (no reliable annotations) — fixtures reflect the TS
  contracts (`getRoutes()` / `getListeners()` only). This one is deliberate, not a gap.
- **No `Env` module; the functional run-loop half is not ported.** `Exiter` does now
  exist, contrary to what this file recorded.
  [valkyrjaio/valkyrja-ts#147](https://github.com/valkyrjaio/valkyrja-ts/issues/147)
- **Response cache (`CacheResponseMiddleware`) not ported.** Build it the JSON way; do
  not port PHP's removed file-generation classes.
  [valkyrjaio/valkyrja-ts#148](https://github.com/valkyrjaio/valkyrja-ts/issues/148)

### Container namespace

- **Three PHP structures have no TS counterpart** — `NativeChildContainer`, a
  standalone `ProvidersAware`, and the `Provides` trait. Each needs a decision: mirror
  PHP, or record the difference as deliberate.
  [valkyrjaio/valkyrja-ts#149](https://github.com/valkyrjaio/valkyrja-ts/issues/149)
- **`ChildContainer` does not inherit singleton _bindings_ from the parent.** A
  parent `bindSingleton` is rebuilt on every `get` through the child.
  [valkyrjaio/valkyrja-ts#133](https://github.com/valkyrjaio/valkyrja-ts/issues/133)
- **`Container.getFallback` ignores `InvalidReferenceMode`** — the `mode` parameter
  is a no-op, and the method always throws.
  [valkyrjaio/valkyrja-ts#134](https://github.com/valkyrjaio/valkyrja-ts/issues/134)

### Event namespace (largely unported)

Six of the ten pieces are missing. PHP has tests for all of them, with no TypeScript
target yet.

- [valkyrjaio/valkyrja-ts#136](https://github.com/valkyrjaio/valkyrja-ts/issues/136)

## Sindri

### Ship a standalone executable on each release

TypeScript bundles `bin/sindri` to one file, then produces a standalone binary. No
release carries a release asset today.

- [valkyrjaio/sindri-ts#104](https://github.com/valkyrjaio/sindri-ts/issues/104)

### Assert the generated content

The end-to-end generate test asserts only that the four `App*Data` files exist.

- [valkyrjaio/sindri-ts#105](https://github.com/valkyrjaio/sindri-ts/issues/105)

## VLID — cross-language parity

VLID is PHP-only today. Port it here, then assert the non-random portion against the
shared PHP fixture. Source the microsecond clock from `process.hrtime.bigint()`.

- [valkyrjaio/valkyrja-ts#138](https://github.com/valkyrjaio/valkyrja-ts/issues/138)
