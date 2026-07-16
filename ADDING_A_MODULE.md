# Adding a Protocol Module

This document describes, independent of any specific protocol, how to add a new **entry/protocol
module** to Valkyrja — the way HTTP, CLI, and gRPC are structured, and the way Queues (and future
protocols) will be. It is the generalization of those ports: follow it and a new module drops into the
existing container, middleware, dispatch, error-handling, and observability machinery without
reinventing any of it.

If you are building a concrete module, read this alongside the design doc for that module (e.g.
[`GRPC.md`](GRPC.md)) and the implementation notes for the closest existing one.

## What a "module" is

A module is a **protocol surface**: a way work enters the application (an HTTP request, a CLI
invocation, a gRPC call, a queued job) and a response leaves it. Every module shares the same spine:

```
Server/Worker → Handler (kernel) → Router → middleware → user handler → response propagation
```

Two hard rules define the boundary:

1. **Worker-agnostic core.** The module's core never depends on a specific server/worker/driver.
   Adapters bridge external workers to the core's contracts. The core is pure framework code.
2. **Framework features are inherited, not reimplemented.** The container, middleware composition,
   event dispatch, exception handling, and observability all work the same as in every other module.
   Adapters translate; they do not re-implement these concerns.

Responses propagate **Go-style**: a response value flows back up the pipeline, each layer inspecting
what it received. Exceptions are the fallback — `ThrowableCaught` middleware converts them back into
the response flow.

## The standard anatomy

Mirror the package/namespace layout of the existing modules. Names below are generic; substitute the
protocol's vocabulary (HTTP "Request/Response", CLI "Input/Output", gRPC "ServiceCall/ServiceResponse",
Queue "Message/Result").

### 1. Message / value types — the model

The immutable inbound and outbound value types plus their supporting values and enums:

- **Inbound** (`Request`/`Input`/`ServiceCall`/`Message`): what the adapter hands in. Carries the map
  key used for routing, the payload/messages, metadata, and any per-invocation context (deadline,
  cancellation, peer, attempt count, …). Exposes the resolved `Route` once matched.
- **Outbound** (`Response`/`Output`/`ServiceResponse`/`Result`): what goes back. Carries a status/
  outcome, metadata, and the payload/messages.
- **Supporting values**: status/outcome types, a metadata multi-map, timeouts/deadlines, cancellation,
  connection/peer info — whatever the protocol needs.

All value types are **immutable** with `with*` copy methods and static factories, matching the sibling
modules exactly.

### 2. Routing

- **`Route`** — the immutable value stored in the map: the key, the handler reference, per-stage
  middleware lists, and any protocol metadata (streaming flags, request/response types, retry policy,
  …). Analogous to HTTP `Route` / CLI `Command`.
- **`RouteCollection`** — the map. HTTP pattern-matches; CLI, gRPC, and Queue use a **direct map
  lookup** by key. Keep it a plain `Map<key, Route>` unless the protocol genuinely needs matching.
- **`Router`** — resolves the inbound to a `Route` and dispatches: register the route's per-stage
  middleware onto the shared stage handlers, run the pre-handler middleware, invoke the user handler,
  run the post-handler middleware. A missing entry routes to the "not matched" stage with a protocol
  default (HTTP 404, gRPC `UNIMPLEMENTED`, …). Keep the name `Router` even for map lookup — the role
  (resolve + dispatch) is identical; only the resolution strategy differs.

### 3. Middleware pipeline

- **Stage contracts** — one interface per pipeline stage. The common shape is: an always-run
  pre-router stage, matched/not-matched/dispatched stages around the handler, a throwable-caught
  stage, an always-run "sending" stage, and an always-run "terminated" stage. Each stage method takes
  the inbound (and response, where one exists) plus the stage handler as `next`.
- **Handler contracts + result records** — the per-stage handler interface, and small result records
  for stages that can either continue *or* short-circuit with a response.
- **Abstract `Handler` base** — holds the ordered middleware, resolves each from the container,
  advances the chain, and (where the protocol has a cooperative cancellation model) centralizes the
  cancellation check so every request-processing stage inherits it. A middleware that returns without
  calling `next` structurally short-circuits.
- **Concrete stage handlers** — one per stage, extending the base.

Middleware is passive: resolved from the container, called via `handle(inbound, [response], next)`,
free to return a response directly (short-circuit) or delegate to `next`.

### 4. Server / kernel handler

- **Kernel `Handler`** (`RequestHandler`/`InputHandler`/`ServiceHandler`/…) — the entry point the
  adapter calls. Orchestrates the stages, runs the top-level try/catch that maps a throwable to a
  response (then through `ThrowableCaught`), and performs any entry-point checks. If the wire write is
  the adapter's job and must sit between "sending" and "terminated", split the kernel so the adapter
  can interleave (see gRPC's `handle`/`sending`/`terminate`).
- **Adapter contract** — a tiny `start(handler)/stop()` (or equivalent) interface. Portable even though
  every implementation is per-worker.

### 5. Throwable hierarchy

A `Throwable` contract for the module, a `RuntimeException` base, and specific exceptions — mirroring
the sibling modules' `*/throwable/` trees.

### 6. Registration → the map

Each language uses its idiomatic discovery mechanism to populate `Map<key, Route>`:

- An attribute/annotation/decorator marking controllers and their handler methods (e.g.
  `@GrpcService` + `@GrpcMethod`), plus a repeatable middleware attribute dispatched to its stage by
  the contract type it implements.
- A **collector** that reflects (or, where preferred, code-generates) those into `Route`s.
- A **route-provider contract** (`getControllerClasses()` + `getRoutes()`) that components supply, so
  the collection is built at boot from the aggregated providers.

## Application wiring

### Config

Add a `Config` value + contract carrying the module's port/options and its per-stage middleware lists,
with sensible defaults. Bind it in the application bootstrap (`App.bootstrapServices`) so the config is
a container singleton keyed by its contract — add the `instanceof <YourConfig>` binding next to the
existing ones.

### Providers

Two provider kinds, following the existing pairs:

- **Service providers** publish the module's services into the container via a `publishers()` map
  (kernel handler, `Router`, `RouteCollection`, collector, and the stage handlers). Publish the stage
  handlers as **singletons** so the `Router` and the kernel handler resolve the *same* instances —
  otherwise per-route middleware in the "sending"/"terminated" stages silently never fires.
- **Component providers** group the service providers and declare the module's route providers.

### Provider aggregation (the invasive part)

Route providers are aggregated by the application. Adding a new module means adding a `get<Module>
Providers` method to the shared `ComponentProviderContract` and `ApplicationContract`, a backing field
+ method in the kernel, a delegating override in the child application, and — because these are
abstract to keep every provider explicit — a `return []` in **every** existing implementor. It is
mechanical but wide; do it in one pass and compile to confirm.

### Application entry

Provide the entry points, mirroring HTTP's `Http`/`WorkerHttp` and gRPC's `Grpc`/`WorkerGrpc`:

- A **single-shot** entry for embedding/tests (bootstrap + handle one unit of work).
- A **worker base** for persistent runtimes: `bootstrap(config)` once, then `dispatch(app, data,
  inbound, …)` per unit of work, creating an isolated child container each time so state never bleeds
  between units. If the adapter must write between "sending" and "terminated", pass it a **writer
  callback** so the write slots into the middle while the child container stays alive.
- If (and only if) the protocol can run on a zero-dependency in-core server, add one (HTTP's
  `Exchange*`). Most protocols cannot and rely entirely on external adapter modules.

## Container mechanics to remember

- **Publishers are deferred.** `publishers()` register lazy callbacks; a service is not "instantiated"
  until first resolved. When gating on an *optional* registered-but-unmaterialized service, use the
  container's `has(...)` availability check, **not** an "is instantiated" check.
- **`getSingleton` is strict.** Resolving a class that was never published throws. User-supplied
  middleware must be registered (published) by the application, exactly as in the existing modules.

## Adapters

Adapters live in **separate entry modules/repos** depending on the *published* framework plus the
native driver/server library. They translate the native call into the module's inbound type, invoke
the kernel handler, and translate the outbound type back. They are thin (translation only) and never
re-implement framework concerns. During development, verify an adapter compiles against the local
framework with a build-tool composite build, then release the framework, then bump + green the adapter.

## Testing

- Mirror the source tree in the test tree; **100% line and branch coverage**.
- Unit-test each contract/impl through its public API; use container-backed fixtures for middleware
  chains.
- Add **functional** tests that boot the full stack from a registered controller and drive a unit of
  work end-to-end — including the "not matched" default and "per-route sending/terminated middleware
  actually fires".
- Reusable controllers/middleware are fixtures.

## Checklist

- [ ] Inbound + outbound value types, supporting values, enums (immutable, `with*`, factories).
- [ ] `Route`, `RouteCollection`, `Router`.
- [ ] Stage contracts + handler contracts + result records + abstract `Handler` + concrete handlers.
- [ ] Kernel handler (+ contract) with try/catch → status mapping; adapter contract.
- [ ] Throwable contract + base + specific exceptions.
- [ ] Registration: attributes + collector + route-provider contract.
- [ ] Config + contract; bound in the application bootstrap.
- [ ] Service + component provider pairs; stage handlers published as shared singletons.
- [ ] `get<Module>Providers` across the shared contracts, kernel, child app, and every implementor.
- [ ] Single-shot + worker-base application entry.
- [ ] Adapter(s) in separate entry module(s).
- [ ] Tests mirroring the source tree at 100% coverage, including end-to-end wiring.

## Reference implementations

- **CLI** — map lookup, no client transport; closest to Queue.
- **HTTP** — pattern matching, request/response bytes, in-core `Exchange*` server.
- **gRPC** — map lookup, typed messages, cooperative cancellation, external adapters. See
  [`GRPC.md`](GRPC.md) and [`GRPC_IMPLEMENTATION.md`](GRPC_IMPLEMENTATION.md).
