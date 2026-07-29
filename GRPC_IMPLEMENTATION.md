# gRPC Implementation Notes

This document records the concrete decisions and refinements that emerged while building the gRPC
protocol in the **Java** reference port. It is a companion to [`GRPC.md`](GRPC.md) — where that
document is the language-agnostic contract, this one is the "what we actually learned building it,"
so the PHP, TypeScript, Go, and Python ports can inherit the same shape and avoid re-deriving the same
answers.

Everything here is intended to be portable. Where a detail is genuinely Java-specific it is called
out as such.

## Module anatomy that was built

The worker-agnostic core lives under a single `grpc` module/namespace, mirroring `http` and `cli`:

```
grpc/
  message/          value types (the wire model)
    call/           ServiceCall (+ contract)         — inbound
    response/       ServiceResponse (+ contract)     — outbound
    status/         Status (+ contract)
    metadata/       Metadata (+ contract)
    deadline/       Deadline (+ contract)
    cancellation/   CancellationToken (+ contract)
    peer/           Peer, AuthContext, Certificate (+ contracts)
    enum_/          StatusCode, CancellationReason, AddressType
  routing/
    data/           Route (+ contract)               — the service-map value
    collection/     RouteCollection (+ contract)     — Map<method, Route>
    dispatcher/     Router (+ contract)
    collector/      AttributeRouteCollector (+ contract)
    attribute/      @Service, @Method, @Middleware (Grpc- prefix dropped; see naming note)
    provider/       GrpcRouteProviderContract + provider pair
    throwable/      routing exceptions
  middleware/
    contract/       7 stage middleware contracts
    handler/        7 stage handlers + abstract Handler + handler contracts
    data/           CallReceivedResult, RouteMatchedResult
    provider/       middleware provider pair
  server/
    handler/        ServiceHandler (+ contract)      — the kernel entry point
    adapter/        ServiceAdapterContract
    provider/       server provider pair
  support/          Cancellation (the two-question helper)
  throwable/        GrpcThrowable → GrpcRuntimeException → CancelledException
```

Plus the application-level wiring that is not inside the `grpc` module: a `GrpcConfig`, the
`Grpc`/`WorkerGrpc` application entry points, and the `getGrpcProviders` extension of the shared
component/application provider contracts.

## Settled decisions

### 1. Messages stay agnostic; translation happens only at the adapter

`ServiceCall` and `ServiceResponse` carry messages as the language's "any" type (`Object` in Java,
`mixed` in PHP, `unknown` in TS, `any`/`interface{}` in Go, `Any` in Python). The framework **never**
references generated protobuf types. At the adapter boundary a message is raw `byte[]`; user handlers
decode/encode. Basing the contracts on a native protobuf `Message` type would couple the core to
protobuf and break cleanly in PHP/TS, so it is explicitly avoided.

### 2. Cancellation is pull-based and checked at each step (buffered model)

In the **buffered model** (unary, server- and client-streaming) there is no writable message sink.
Outbound messages are a **pull-based iterable**; the adapter drains them through
`call.cancellable(...)`, which checks cancellation between items and exits early. This maps cleanly
onto Go channels, JS async-iterables, and PHP generators — all pull-based. The **streaming
(bidirectional) model** is the one deliberate exception — it adds a push sink; see decision 9.

### 3. The two-question check lives in the middleware-handler base

`Cancellation.checkAndFinalize(call, response?)` (pre-check before delegating to the wrapped
middleware, post-check on its return) is implemented once on the **abstract `Handler` base** and used
by every *request-processing* stage handler (`CallReceived`, `RouteMatched`, `RouteNotMatched`,
`RouteDispatched`, `ThrowableCaught`). The two **always-run** stages (`SendingResponse`,
`Terminated`) deliberately do **not** apply it — per the fast-exit path they run even for cancelled
calls. Do not leave the check only in the `Router`, or the "every middleware is cancellation-correct
for free" guarantee is silently broken.

### 4. `ServiceHandler` splits into `handle` / `sending` / `terminate` (+ `run`)

Because the wire write is the adapter's job and must happen **between** `SendingResponse` and
`Terminated`, the kernel exposes:

- `handle(call)` — `CallReceived` → `Router`, with the top-level try/catch → throwable→Status mapping
  and the entry-point cancellation pre-check (the one place no response exists yet).
- `sending(call, response)` — runs the `SendingResponse` stage; always runs.
- `terminate(call, response)` — runs the `Terminated` stage after the wire write.
- `run(call)` — convenience: `handle` then `sending`, returning the response for the adapter to write.

The adapter pattern is: `response = handler.run(call)` → write to wire → `handler.terminate(...)`.

### 5. `WorkerGrpc.dispatch(app, data, call, writer)`

The persistent-worker entry base takes a **writer callback** so the adapter's wire write slots between
`sending` and `terminate` while the per-call child container stays alive across the whole lifecycle:

```
dispatch(app, data, call, writer):
    child = childContainer(app, data); bootstrap child
    handler = child.get(ServiceHandler)
    response = handler.handle(call)
    response = handler.sending(call, response)
    writer(response)          # adapter writes to the wire here
    handler.terminate(call, response)
```

A single-shot `Grpc.handle(config, call)` also exists for embedding/tests (bootstraps per call).

### 6. Router and ServiceHandler must share the stage-handler singletons

Per-route `SendingResponse`/`Terminated`/`ThrowableCaught` middleware are registered by the `Router`
onto the stage-handler instances it holds, and later invoked by the `ServiceHandler`. The provider
wiring must publish those handlers as **container singletons** so both resolve the *same* instance —
otherwise per-route middleware in those stages silently never fires. Add a functional test that proves
it.

### 7. Registration: attributes → `Map<method, Route>`

- `@Service(service = "package.Service")` on a controller class.
- `@Method(name, clientStreaming, serverStreaming)` on each RPC handler method.
- Repeatable `@Middleware(name = X.class)`, dispatched to its stage by the middleware contract type
  it implements.

A runtime-reflection `AttributeRouteCollector` builds a `Route` per method, keyed `"/service/name"`,
with the annotated method wired as the reflective handler `(Container, Route) -> ServiceResponse`. This
matches the CLI/HTTP collectors; it is **not** a compile-time processor. The map is populated at boot
from the aggregated `getGrpcProviders()` route providers.

> **Naming.** These attributes drop the `Grpc` prefix (`@Service`/`@Method`/`@Middleware`, not
> `@GrpcService`/…). They are controller-facing and live in the protocol's own attribute namespace,
> and a gRPC controller never imports the HTTP/CLI equivalents, so the short names never collide.
> **Providers keep their prefix** (`GrpcRouteProviderContract`, `GrpcRoutingDataContract`, `GrpcConfig`)
> because those *are* imported alongside their HTTP/CLI siblings (a component provider references all
> three). Apply the same rule per language.

### 7a. Middleware classification in the route-data cache

`@Middleware(name = X.class)` names only the class; its **stage** is the set of stage contracts the
class implements, discovered by an `isAssignableFrom`/`is_a` cascade — added to **every** stage it
implements (independent `if`s, not `else if`; a class can serve multiple stages). At runtime the
collector does this trivially. The **cache generator** must reproduce it, and how depends on whether
the toolchain can inspect a type's hierarchy:

- **Pre-classify at codegen** — generator resolves the middleware's full ancestry and emits typed
  per-stage lists into the cached `Route`. Used by **PHP** (`is_a`, live autoload), **Python**
  (reflection), **TypeScript** (ts-morph `getImplements`), and achievable in **Java** (JavaParser
  symbol solver, `getAllAncestors`), **Kotlin** (KSP), **C#** (Roslyn), **Scala** (macros/reflection).
  Resolve the *full* hierarchy, not the direct `implements` clause, so a middleware extending an
  abstract base or a custom sub-contract still classifies.
- **Runtime `withMiddleware`** — the cache emits the flat, unclassified class list and a
  `Route.withMiddleware(List<Class>)` runs the cascade lazily when the route's supplier
  materializes (once, on match — not per request). This is the fallback for toolchains that cannot
  read the hierarchy at generation. **Go** needs this (structural, implicit interfaces — nothing to
  read), unless middleware is required to *embed* the contract interface, which makes it declared.

The discriminator is **nominal vs. structural typing**, not the language. Java pre-classifies (symbol
solver) to stay consistent with PHP/TS/Python; only Go deviates to `withMiddleware`. Whichever a
language picks, a class implementing several stage contracts must land in **all** their buckets.

### 8. `ServiceAdapter` contract

`ServiceAdapter { start(ServiceHandler); stop() }` is part of the agnostic surface even though every
implementation is per-worker. Keep it in the core.

### 9. The streaming model for bidirectional calls

Decision 5's `dispatch` covers the **buffered model** (unary, server- and client-streaming): buffer
inbound, dispatch once on half-close, drain one `ServiceResponse`. That model cannot serve an
*interactive* bidirectional call — the client waits for a reply before sending more and never
half-closes early — so a **second dispatch path** exists for methods where **both** streaming flags
are set (`isClientStreaming() && isServerStreaming()`); everything else stays buffered. The full
contract is ratified in [`GRPC.md`](GRPC.md) → *Streaming and Call Shapes*; the Java realization:

- **`WorkerGrpc.dispatchStreaming(app, data, callFactory, OutboundStream)`** dispatches the handler
  **immediately** (not on half-close) on a **per-call concurrent execution unit** — a virtual thread
  in Java, a goroutine in Go, an async task in JS/Python (see *Streaming and Call Shapes*).
- **Inbound is a live stream** (`InboundMessageStream`, a bounded blocking queue) the transport feeds
  as messages arrive; iteration blocks until the next message and ends on half-close **or** cancel.
- **Outbound is a push sink** (`ServiceCall.send` → `OutboundStream`). `SendingResponse` fires **once**
  at first emit (or at close if the handler emits nothing) against an OK shell whose initial metadata
  becomes the headers; the handler returns a **terminal** `ServiceResponse` (status + trailing
  metadata, no messages); `ResponseSent` fires **once** at close. So the pipeline still runs once per
  call, exactly as a buffered multi-message response does.
- **Flow control**: request the high-water (`maxInboundMessages`) up front and refill one per drained
  message, so the queue never outgrows the bound (here it is an in-flight window, not a hard cap).
- **Context is captured on the transport thread** (deadline/peer/metadata) before the worker starts —
  the library binds the call context there, not on the worker unit.
- **`send` is single-thread**: the transport is not thread-safe, so a concurrent `send` throws fast
  rather than silently corrupting the wire. A handler that fans out to threads funnels emissions
  back through one.

## Portable gotchas

- **Deferred publishers vs. availability checks.** Provider `publishers()` register as *deferred
  callbacks*; a service is not "a singleton" until first resolved. When gating on an optional
  registered-but-unmaterialized service (e.g. the route collector inside `publishRouteCollection`), use
  the container's `has(...)`-style "is anything registered" check, **not** an "is instantiated"
  check, or controller-scanned routes are silently dropped.
- **`getSingleton` is strict.** Resolving middleware by class throws if unregistered — route middleware
  must be registered (published) by the application, exactly like HTTP/CLI. This is expected; document
  it for users.
- **Adding `getGrpcProviders` is invasive.** It joins `getCli`/`getHttpProviders` on the shared
  component/application provider contracts and the kernel, so **every** existing implementor must be
  updated. Keep it abstract (consistent with the siblings) rather than a defaulted method.
  - Defaulting it on the contract *looks* like the cheap way out, but it privileges one protocol over
    its siblings and encodes "gRPC is optional" into a contract that says no such thing about HTTP or
    CLI. Keep the contract symmetric.
  - The cost of abstract — an identical empty implementation in every component that contributes no
    gRPC routes — is a **class** problem, not a contract problem. Solve it with a base class
    (`application/provider/abstract_/ComponentProvider` in the Java port) that implements *all* the
    provider methods as empty, and have components extend it and override only what they contribute.
    Adding the next protocol then touches one base class instead of every component.
  - Watch the duplication gate: adding one identical method to ~25 components is ~25 identical new
    blocks, which trips new-code duplication analysis (SonarCloud's `new_duplicated_lines_density`).
    The base class avoids the additions entirely rather than suppressing the warning.
- **Reflective dispatch must rethrow the handler's own throwable.** Attribute/annotation routing invokes
  handlers reflectively, and most languages wrap whatever the target threw (Java's
  `InvocationTargetException`, PHP's `ReflectionException` paths, etc.). Wrapping that in a generic
  runtime exception hides `CancelledException` from the status mapping, so every cancellation reports
  `INTERNAL` and cooperative cancellation is dead for the primary handler path — with no test failure
  to show for it, because unit tests that throw `CancelledException` *directly* bypass the collector.
  Unwrap the reflection wrapper and rethrow the cause. Test through the reflective path, not around it.
- **The framework maps only cancellation and the catch-all.** Domain outcomes (`NOT_FOUND`,
  `INVALID_ARGUMENT`, …) are *returned* by the handler on the `ServiceResponse`, exactly as HTTP
  handlers return status codes — there is no domain-exception hierarchy to map. See
  "Exception → Status Mapping" in `GRPC.md`.
- **`StatusCode` values are the wire codes 0–16** and match the native gRPC library's codes 1:1, so the
  adapter maps status with a single `fromCodeValue(status.code)`. Expose one accessor for the numeric
  value, not two.
- **"No-deadline" needs a finite sentinel.** `Deadline.none().getRemaining()` returns a large but
  finite duration (the Java port uses ~100 years) so downstream arithmetic never overflows. Pick the
  same sentinel in every port.

## No "Exchange" (zero-dependency in-core server) for gRPC

HTTP ships an in-core, zero-dependency server (`ExchangeHttp`, on the JDK `com.sun.net.httpserver`
`HttpServer` / equivalent per language). gRPC **cannot**: that built-in server is HTTP/1.1 only, and
gRPC mandates HTTP/2 with trailers. gRPC's in-core entries are therefore `Grpc` (single-shot) and
`WorkerGrpc` (adapter base) only; actual serving always requires an external transport module.

## Adapters

Adapters live in **separate entry modules/repos** that depend on the *published* framework plus a
native gRPC library. Key points from the Java adapters (grpc-netty and grpc-servlet):

- **Generic dispatch via a fallback handler registry.** Rather than per-service generated stubs, the
  adapter registers a fallback registry that resolves *any* `/service/method` to a generic
  `ServerMethodDefinition` with an **identity `byte[]` marshaller** and a handler that buffers inbound
  messages, builds a `ServiceCall`, calls `WorkerGrpc.dispatch`, and writes the `ServiceResponse` back
  (initial metadata → messages (drained through `call.cancellable`) → status + trailers).
- **The bridge is transport-agnostic.** It depends only on the native gRPC *API* (not the transport),
  so the same bridge code serves every transport (Netty, servlet). Only the server bootstrap differs
  (`NettyServerBuilder` vs `ServletServerBuilder` → a `GrpcServlet` in an embedded servlet container).
- **Build/release ordering.** Entry modules pin the *published* framework version, so their CI cannot
  compile the adapter until the framework is released with gRPC and the dependency is bumped. During
  development, verify the adapter compiles against the local framework with a build-tool composite
  build (e.g. Gradle `--include-build`), then release the framework, then bump + green the adapters.

## Testing

100% line **and** branch coverage. Unit tests live in the port's unit tree mirroring the `grpc` source
tree; end-to-end wiring is exercised by functional tests that boot the full stack from a
`@Service` controller and dispatch calls (including "unknown method → UNIMPLEMENTED" and
"per-route SendingResponse/Terminated middleware fires"). Reusable controllers/middleware are fixtures.

Most tests drive the bridge with a mocked native call. That verifies translation and control flow
but not the wire. One **live end-to-end test** closes that gap by standing up a real gRPC transport.

### Live end-to-end verification (real transport, in Docker)

`GrpcNettyEndToEndTest` (in the Java port's functional test tree) boots a real Netty gRPC server
backed by `GrpcBridge.registry(app, data)` on an ephemeral port, connects a real Netty client, and
drives **all four gRPC call types** over actual HTTP/2 — framing, flow control, headers, messages,
and trailers included:

- **Unary (1 → 1, buffered)** — a `Ping` call: send one message, expect one response and `OK`.
- **Server-streaming (1 → N, buffered)** — a `Fanout` call: send one message, expect several
  responses and `OK`.
- **Client-streaming (N → 1, buffered)** — a `Collect` call: send two messages, half-close, expect
  the single handler response and `OK`.
- **Bidirectional (N ↔ M, streaming model)** — an `Echo` call dispatched on a per-call virtual
  thread: send three messages interleaved with `request(n)`, half-close, expect all three echoed
  back and `OK`.

The first three exercise the buffered dispatch path; the last exercises the streaming model. All use
a raw `byte[]` marshaller (`GrpcBridge.ByteMarshaller`), so no generated protobuf is needed —
the fallback registry handles any method. `grpc-netty-shaded` is a **test-only** dependency of the
port's JUnit build; the published framework artifact keeps `io.grpc` `compileOnly`.

Run it in a clean JDK 21 container so the result does not depend on a developer's local toolchain
(replace `<framework-repo>` with the checkout of the Java framework port):

```bash
docker run --rm \
  -v "<framework-repo>":/work -w /work \
  -v "$HOME/.gradle":/root/.gradle \
  eclipse-temurin:21-jdk \
  ./gradlew -p .github/ci/junit test \
    --tests "io.valkyrja.tests.functional.grpc.endtoend.GrpcNettyEndToEndTest" \
    --console=plain --no-daemon
```

Expect `BUILD SUCCESSFUL`. Mounting `~/.gradle` is only a cache speed-up — omit it for a fully cold,
from-scratch run. This was verified on `eclipse-temurin:21-jdk` with Docker 29.x; the framework
targets JDK 21, which is also what makes the streaming model's virtual-thread-per-call realization
available (see [`GRPC.md`](GRPC.md) → *Streaming and Call Shapes*).

For a **new language port**, replicate the same shape: boot the port's native gRPC server against its
bridge, drive one call of each of the four types (unary, server-streaming, client-streaming,
bidirectional) with a raw-bytes codec, and run it in that language's official runtime container so the
check is reproducible.

## Implementation order for the next port

1. `message/*` value types + enums (+ tests).
2. `routing` (`Route`, `RouteCollection`, `Router`) + `ServiceCall` + `support/Cancellation`.
3. `middleware` (stage contracts + abstract `Handler` with the two-question check + 7 handlers).
4. `server/ServiceHandler` + throwable→Status mapping + `ServiceAdapter` contract.
5. Registration (attributes + collector) and container/app wiring (`GrpcConfig`, provider pairs,
   `getGrpcProviders`, `Grpc`/`WorkerGrpc`).
6. One thin adapter against a real gRPC library to prove the `ServiceCall`/`ServiceResponse` ↔ native
   translation end-to-end.
