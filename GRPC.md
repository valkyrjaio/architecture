# gRPC

This document describes how gRPC integrates into Valkyrja as a first-class protocol alongside HTTP and CLI. The design
is language-agnostic and applies to all current and planned Valkyrja ports (PHP, Java, TypeScript, Go, Python).

## Design Principles

gRPC follows the architectural pattern already established for HTTP and CLI:

1. **Worker-agnostic.** The framework never depends on a specific server or worker implementation. Adapters bridge
   external workers (RoadRunner, OpenSwoole, grpc-java, grpc-go, @grpc/grpc-js, grpcio) to Valkyrja's internal
   contracts, exactly as they do for HTTP and CLI today.

2. **Framework features are inherited, not reimplemented.** Middleware, the container, event dispatch, exception
   handling, and observability all work the same way in gRPC as they do in HTTP and CLI. Worker implementations do not
   ship their own versions of these concerns — that is what the framework provides.

3. **Response propagation, Go-style.** Unwinding uses `ServiceResponse` objects flowing back up through the pipeline,
   with each layer inspecting what it received and deciding how to proceed. Exceptions are a fallback: when code cannot
   produce a response directly, `ThrowableCaught` middleware converts them back into the response flow.

4. **No routing logic, just map lookup.** gRPC identifies `(service, method)` at the protocol level via the `:path`
   pseudo-header (`/package.Service/Method`). A direct `Map<string, Route>` lookup resolves it — the same shape CLI uses
   for commands, without pattern matching or parsing. The component is still called `Router` for consistency with HTTP
   and CLI, since its role (resolve an inbound call to a `Route` and dispatch) is the same; only the resolution strategy
   differs.

5. **Symmetry across protocols.** The pipeline shape is identical to HTTP and CLI:

```
HTTP:   Server  → RequestHandler → Router (pattern match) → middleware → handler
CLI:    Console → InputHandler   → Router (map lookup)    → middleware → handler
gRPC:   Server  → ServiceHandler → Router (map lookup)    → middleware → handler
```

## The Wire Protocol

gRPC is HTTP/2 with specific conventions. Every call has three wire segments:

**Request:**

- `:method: POST`, `:path: /package.Service/Method`, `content-type: application/grpc`
- Optional `grpc-timeout` (duration with unit suffix: `5S`, `500m`, `1H`, etc.)
- Optional `grpc-encoding` (compression)
- Custom metadata as HTTP/2 headers (keys ending in `-bin` carry binary values)
- Body: one or more length-prefixed framed messages (1 byte compression flag + 4 bytes big-endian length + N bytes
  protobuf)

**Response:**

- Initial headers: `:status: 200`, `content-type`, initial response metadata
- Body: zero or more length-prefixed framed messages
- Trailers (HTTP/2 trailing headers): `grpc-status` (integer 0–16), optional `grpc-message`, optional
  `grpc-status-details-bin` (base64 `google.rpc.Status` protobuf), custom trailing metadata

Two important properties:

- **gRPC errors return HTTP/2 `:status: 200`.** The actual RPC outcome lives in the `grpc-status` trailer.
  `:status: 200` means "transport worked"; `grpc-status: 5` means "the call failed with NOT_FOUND."
- **Trailers are mandatory.** `grpc-status` is always sent as a trailer, even on success. This is why gRPC requires
  HTTP/2 — HTTP/1.1 does not support trailers cleanly.

The library handles all of this framing. The framework and user code work with decoded message objects, metadata as
structured maps, and status as a value type. Bytes never cross into framework territory.

## Core Contracts

The language-agnostic surface area is intentionally small: eight contracts total, plus the pipeline components.

### `ServiceHandler`

The kernel entry point for gRPC, analogous to `RequestHandler` (HTTP) and `InputHandler` (CLI). Worker adapters hand
calls to `ServiceHandler.handle()`; everything downstream is pure Valkyrja.

Responsibilities:

- Orchestrate the middleware pipeline stages (`CallReceived`, `SendingResponse`, `Terminated`).
- Delegate to `Router` for route resolution and handler dispatch.
- Run `ThrowableCaught` middleware when exceptions propagate up.
- Fast-exit on cancellation signals.

### `ServiceCall` (immutable)

What comes in from the worker adapter. Models the inbound side of the wire.

```
ServiceCall
  getMethod(): string                   // "/package.Service/Method" — the map key
  getMetadata(): Metadata               // inbound headers (one bucket inbound)
  getDeadline(): Deadline               // never null; may be Deadline::none()
  getCancellation(): CancellationToken  // never null; may be Token::never()
  getPeer(): Peer                       // never null; auth may be "insecure"
  getMessages(): iterable<Message>      // decoded inbound messages (length 1 for unary/server-stream)
  getRoute(): Route                     // resolved route metadata
```

### `ServiceResponse` (immutable)

What goes out. Models the outbound side of the wire.

```
ServiceResponse
  getStatus(): Status
  withStatus(Status): static

  getInitialMetadata(): Metadata
  withInitialMetadata(Metadata): static

  getTrailingMetadata(): Metadata
  withTrailingMetadata(Metadata): static

  getMessages(): iterable<Message>
  withMessages(iterable<Message>): static

  isCancellation(): bool    // convenience: status.isCancellation()
```

`messages` is typed as `iterable<Message>` so unary responses use `[singleMessage]` and streaming responses use a lazy
generator/async-iterable. The underlying concrete type differs; the contract does not.

Initial metadata locks the moment the first message is written to the wire (wire-level constraint). Trailing metadata
stays mutable until the handler returns and the adapter flushes the call's close.

### `Route` (immutable)

The value stored in the service map, analogous to HTTP's `Route` and CLI's `Command`. Held in a `Map<string, Route>`
keyed by fully-qualified method name.

```
Route
  getMethod(): string                   // "/package.Service/Method"
  getService(): string                  // "package.Service"
  getMethodName(): string               // "Method"
  getHandler(): Handler                 // class+method reference or callable
  getMiddleware(): list<Middleware>     // stack for this route
  getRequestType(): class-string        // generated protobuf message type
  getResponseType(): class-string       // generated protobuf message type
  isClientStreaming(): bool
  isServerStreaming(): bool
```

### `Status` (immutable)

The gRPC call outcome. Mirrors the pattern HTTP uses for status code plus reason phrase, with an additional field for
rich error details.

```
Status
  getCode(): StatusCode        // enum: OK, CANCELLED, ..., UNAUTHENTICATED
  getMessage(): string         // never null; defaults from code (human-readable)
  getDetails(): ?bytes         // optional; google.rpc.Status protobuf bytes

  isOk(): bool
  isCancellation(): bool       // true for CANCELLED or DEADLINE_EXCEEDED

  static ok(): Status
  static cancelled(?string): Status
  static deadlineExceeded(?string): Status
  static notFound(?string): Status
  static unimplemented(?string): Status
  static internal(?string, ?bytes): Status
  // ... factory per code
```

The `StatusCode` enum is gRPC-specific, not reused from HTTP. The two enums have different ranges, different names, and
different semantics; reusing HTTP's would accept values with no meaning on the wire.

### `Metadata`

Multi-map of string keys to lists of string-or-binary values. Case-insensitive keys. Represents both HTTP/2 headers (
request metadata, initial response metadata) and HTTP/2 trailing headers (trailing response metadata).

```
Metadata
  get(string): ?string|bytes            // first value
  getAll(string): list<string|bytes>    // all values
  has(string): bool
  with(string, string|bytes): static
  withAdded(string, string|bytes): static
  without(string): static
  toArray(): array<string, list<string|bytes>>
  // iteration
```

Keys ending in `-bin` carry binary values (base64-encoded on the wire; decoded at the library boundary). The
`string|bytes` union reflects this.

Metadata may share its underlying primitive with HTTP's `Headers` if the shapes align cleanly; if binary-value handling
makes sharing awkward, they stay separate.

### `Deadline`

Represents the absolute time at which the call's budget expires. Computed once at call receipt from the inbound
`grpc-timeout` header; propagated as an absolute time so every downstream layer agrees on the same reference point.

```
Deadline
  getAbsoluteTime(): Instant
  getRemaining(): Duration
  isExpired(): bool
  hasDeadline(): bool

  static fromTimeout(Duration): Deadline
  static fromAbsolute(Instant): Deadline
  static none(): Deadline    // sentinel; always hasDeadline=false, never expired
```

Never null on `ServiceCall`. `Deadline::none()` is the sentinel for "no deadline set by client."

### `CancellationToken`

The signal for "should this work stop?" Unifies two causes: client-initiated cancellation (HTTP/2 RST_STREAM) and
deadline expiry. Deadline expiry is modeled as a cause of cancellation; code only checks cancellation, consulting
`getReason()` if the distinction matters.

```
CancellationToken
  isCancelled(): bool
  getReason(): ?CancellationReason      // CLIENT_CANCELLED | DEADLINE_EXCEEDED | null
  throwIfCancelled(): void              // throws CancelledException if cancelled
  onCancelled(callable): void           // register listener
```

Never null on `ServiceCall`. Adapters wire the token: listen to the library's native cancellation signal, fire the
token; register the deadline timer, fire the token on expiry.

Language-native awaitable/async extensions (Go's `<-ctx.Done()`, JS `AbortSignal`, etc.) may be added per port where
idiomatic, but the base contract is poll + listener, which works in every language.

### `Peer`

Information about the connection's other end. Derived from the transport, not from a single header.

```
Peer
  getAddress(): string                  // "192.168.1.5:54321" or "unix:/var/run/sock"
  getAddressType(): AddressType         // IPV4 | IPV6 | UNIX | UNKNOWN
  getAuthContext(): AuthContext         // always present; type may be "insecure"

AuthContext
  getType(): string                     // "ssl" | "tls" | "insecure" | custom
  getProperties(): array<string, list<string>>
  getPeerCertificates(): list<Certificate>
  getPeerSubject(): ?string
  getTransportSecurityType(): ?string
```

## Cancellation and Deadline Model

Cancellation enforcement in gRPC is **cooperative** in every target language. No gRPC library in any language forcibly
interrupts running handler code. This is a deliberate ecosystem-wide choice — forcible interruption (thread kill,
goroutine stop) is either unavailable or unsafe (leaves locks held, resources leaked). Valkyrja follows the same model.

### What the library handles automatically

- Parses `grpc-timeout`, computes deadline.
- Fires language-native cancellation signal on client cancel or deadline expiry.
- Rejects writes to a closed call (silent drop or error depending on language).
- Sends `DEADLINE_EXCEEDED` / `CANCELLED` status to the client if deadline/cancellation fires before the handler
  produces a response.

### What the framework does

The framework's role is to **surface the library's signals uniformly and check them at orchestration boundaries**,
converting detected cancellation into `ServiceResponse` objects directly (not exceptions) so the normal
response-propagation flow handles them.

#### The two-question pattern

At every orchestrator boundary where control transfers between units of work, the same two questions are asked:

1. **Has cancellation fired, or has the deadline elapsed?** (inspect `call.getCancellation()`)
2. **Does the response we have in hand already carry a cancellation status?** (inspect `response.isCancellation()`, when
   a response exists)

If either answer is yes, fast-exit: return the cancellation response up the stack, skipping remaining request-processing
middleware.

#### Pre-check creates or overlays; post-check inspects only

The two questions are not symmetric in their effect:

- **Pre-check (before delegation).** If cancellation has fired on the call, construct the cancellation response. If a
  response already exists from earlier pipeline work, overlay the cancellation status on it with
  `response.withStatus(Status.cancelled(reason))` — preserving metadata accumulated by middleware that did manage to
  run. If no response exists yet, build fresh with `ServiceResponse.cancelled(reason)`. This situation only occurs at
  `ServiceHandler` entry, before any middleware has run.
- **Post-check (after delegation returns).** If the returned response already has a cancellation status, pass it through
  unchanged. It is already correct — whatever downstream work produced it may have set useful metadata that should be
  preserved.

The shared check logic:

```
checkAndFinalize(call, response?) -> ServiceResponse?:
    if call.getCancellation().isCancelled():
        reason = call.getCancellation().getReason()
        if response exists:
            return response.withStatus(Status.cancelled(reason))
        else:
            return ServiceResponse.cancelled(reason)

    if response exists and response.isCancellation():
        return response    // already cancelled; preserve as-is

    return null             // no cancellation; continue normally
```

Implemented as a single helper on a common base (or as a utility each orchestrator calls), applied identically at every
delegation site.

#### Check locations

All in framework code, no user involvement required:

- `ServiceHandler` at entry (the only location where no response yet exists), and around each delegation to
  `CallReceived` middleware, `Router`, and `SendingResponse` middleware.
- `Router` around delegation to `RouteMatched`/`RouteNotMatched` middleware, the user handler, and `RouteDispatched`
  middleware.
- `MiddlewareHandler` before invoking its wrapped middleware.
- `ServiceResponse.messages.write()` (inside the adapter's write loop) before each outbound message write.

Every orchestrator boundary runs the two-question check. Beyond `ServiceHandler` entry, a response is almost always
already in hand — either produced by a short-circuiting middleware, by the user handler, or by earlier pipeline work —
so the pre-check's "overlay existing response" branch is the common case. The post-check propagates cancellation
fast-exit up the stack without needing any additional mechanism.

This dual mechanism — checking the call's cancellation token and checking the returned response's status — provides
complete coverage with no gaps.

### Fast-exit path

On cancellation detection, the pipeline collapses to:

```
Normal:     CallReceived → Router (RouteMatched → handler → RouteDispatched)
            → SendingResponse → [wire write] → Terminated

Cancelled:  CallReceived → [cancellation detected]
            → SendingResponse → [wire write] → Terminated
```

`SendingResponse` and `Terminated` still run — they are cheap, and observability of cancelled calls is often more
valuable than observability of successful ones. Request-processing middleware (`RouteMatched`, `RouteNotMatched`,
`RouteDispatched`, `ThrowableCaught`) is skipped.

### User handler cooperation

Framework-provided cancellation handling covers everything above the user handler boundary. Inside the handler, three
mechanisms help without requiring explicit checks:

- **Response writes check automatically.** Writing to `ServiceResponse.messages` throws `CancelledException` if the call
  is cancelled. Streaming handlers that write messages iteratively get automatic cancellation on their next write.
- **Cancellable iteration helper.** `call.cancellable(iterable)` yields items from the source while checking
  cancellation between iterations.
- **Deadline-aware clients.** Valkyrja-provided HTTP and gRPC clients propagate the current `Deadline` to outbound calls
  so downstream work inherits the remaining budget.

For pure CPU-bound loops or third-party SDK calls that are not cancellation-aware, handlers must explicitly check
`call.getCancellation().throwIfCancelled()` at appropriate points. This is the irreducible cooperative part.

### Why the framework does not kill handlers

A runaway handler that ignores cancellation runs to completion. The library drops its response (the client has already
seen `DEADLINE_EXCEEDED`), and the worker is occupied for the duration. This is acceptable degradation: correctness is
preserved (the client got the right outcome at the right time), only server capacity is affected. Worker occupancy from
runaway handlers is managed at the worker pool or platform level (worker recycling, pool sizing, Kubernetes limits) —
outside the framework's scope.

## Middleware Pipeline

The gRPC pipeline mirrors HTTP/CLI with gRPC-specific defaults.

```
1. CallReceived         always runs; pre-router
2. Router resolves route from map
3a. RouteMatched        runs if route found; pre-handler
    User handler runs, produces ServiceResponse
3b. RouteDispatched     runs if route was found; post-handler
 OR
3c. RouteNotMatched     runs if route not found
    Default terminal produces ServiceResponse::unimplemented()

[if any above threw]
4. ThrowableCaught      runs if any earlier stage threw

5. SendingResponse      always runs (including error/cancellation paths)
   Adapter writes messages and trailers to wire
6. Terminated           runs after wire write complete
```

All stages except `CallReceived` and `SendingResponse` are optional. Middleware in each stage is resolved from the
container and composed via `MiddlewareHandler`.

### `MiddlewareHandler` as the short-circuit mechanism

`MiddlewareHandler` is the active orchestrator for each stage. Middleware implementations are passive — resolved from
the container, called via `handle(call, next)`, and free to either return a response directly (short-circuit) or
delegate to `next` (continue the chain).

`next` is itself a `MiddlewareHandler` instance. Its `handle()` method is both the entry point from outside the chain
and the continuation point from inside. This single source of truth is where cancellation checks live, following the
two-question pattern:

```
MiddlewareHandler.handle(call, response?, next):
    // Pre-check: cancellation fired on the call, or response already cancelled
    short_circuit = checkAndFinalize(call, response?)
    if short_circuit != null:
        return short_circuit

    // Delegate to the wrapped middleware
    middleware = container.get(this.middlewareClass)
    returnedResponse = middleware.handle(call, response?, next)

    // Post-check: middleware's returned response is cancelled (fast-exit)
    // or cancellation fired during middleware execution
    short_circuit = checkAndFinalize(call, returnedResponse)
    if short_circuit != null:
        return short_circuit

    return returnedResponse
```

The `response?` parameter reflects that a response may already exist by the time a middleware chain is entered (from an
earlier short-circuit) or may not (at the very start of the first stage). Middleware implementations neither check nor
know about cancellation. Every middleware in the system gets cancellation-correct behavior for free.

Short-circuiting is structural: a middleware returning a response without calling `next` simply skips the remainder of
the chain. No special signal, no flag — the absence of the `next` call is the short-circuit.

### `RouteNotMatched` default

When the Router's map lookup returns no entry, `RouteNotMatched` middleware runs, with a framework-provided terminal
that produces `ServiceResponse::unimplemented()` with `grpc-status: UNIMPLEMENTED (12)`. User middleware in this stage
can log unknown method attempts, monitor for scanning, collect metrics on bad-client rates.

### `ThrowableCaught` and cancellation

Cancellation detected by the framework's check points never produces an exception — it produces a `ServiceResponse`
directly with `CANCELLED` or `DEADLINE_EXCEEDED` status. The cancelled response flows through the normal propagation
path (with fast-exit skipping request-processing middleware).

User code can still throw `CancelledException` via explicit `throwIfCancelled()` calls. These exceptions unwind normally
to `ThrowableCaught`, which converts them to cancellation responses and rejoins the normal flow.

The net effect: `ThrowableCaught` handles all thrown exceptions uniformly; cancellation is never a special case in
exception-handling code because the framework's own cancellation handling stays in the response-propagation path.

### `Terminated` stage

Runs after the adapter has written the full response (all messages + trailing metadata + status) to the wire. Used for
cleanup, async logging, metrics emission, and event publication that should not block the client.

Per-worker viability:

- PHP (RoadRunner/Swoole): supported.
- Java: runs after `StreamObserver.onCompleted()`.
- Go: runs after the handler returns, in the same or a spawned goroutine.
- Python async: runs after the handler coroutine yields its response.
- TypeScript: runs after `callback()` or stream end.

## Worker Adapters

Adapters bridge an external gRPC server implementation to `ServiceHandler`. The adapter's responsibilities:

1. Accept the native call representation from the gRPC library.
2. Decode the inbound message(s) (library handles protobuf deserialization).
3. Build a `ServiceCall`: populate `method`, `metadata`, `deadline`, `cancellation`, `peer`, `messages`, `route`.
4. Wire the `CancellationToken` to the library's native cancellation signal and to the deadline timer.
5. Invoke `ServiceHandler.handle(call)`.
6. Translate the returned `ServiceResponse` into the library's native response API (call `.onNext()`, `return response`,
   `callback()`, etc. depending on the library).
7. During streaming writes, route each write through the cancellation check: writes to `ServiceResponse.messages` on a
   cancelled call raise `CancelledException`.

Adapters do **not** forcibly interrupt handler execution — that is not possible in any target language. They are signal
translators, not enforcers.

### Target adapters by language

**PHP**

- RoadRunner (`spiral/roadrunner-grpc`) — primary recommended adapter.
- OpenSwoole (`Swoole\GrpcServer` / `openswoole/grpc`) — coroutine-based alternative, useful for streaming-heavy
  workloads.
- FrankenPHP — deferred until the ecosystem provides native gRPC termination into PHP workers.

**Java** — `grpc-java` with `ServerBuilder`. Generated `BindableService` implementations delegate to
`ServiceHandler.handle()` with `StreamObserver` adapted to the `ServiceCall`/`ServiceResponse` shape.

**TypeScript** — `@grpc/grpc-js`.

**Go** — `google.golang.org/grpc`.

**Python** — `grpcio` (async API).

Each adapter is expected to be thin (roughly 30–60 lines of glue code). All protocol-framework integration — middleware,
container, error mapping, observability — lives above the adapter in Valkyrja code that is unaware of which worker is
running.

### Adapter interface

```
ServiceAdapter
  start(ServiceHandler): void   // begin accepting calls (bind port, TLS, etc.)
  stop(): void                  // graceful shutdown
```

Adapter-specific configuration (TLS, thread pools, plugin registration, port binding) lives on the adapter
implementation, not in the framework-agnostic contract.

## Service Registration

Service registration follows the discovery → map pattern already used by HTTP routes and CLI commands. Each language
uses its idiomatic mechanism:

- **PHP** — `#[GrpcService]` attribute on generated service classes; scan populates the map at boot; result cached via
  the existing data-class generation mechanism (`App\Grpc\Data` namespace, paralleling `App\Http\Data` and
  `App\Cli\Data`).
- **Java** — `@GrpcService` annotation; annotation processor + JavaPoet generates a data class mapping fully-qualified
  method names to `Route` instances at compile time (matching the existing `@Provides` processor pattern).
- **Go** — build-time tag or `go:generate` directive; generated registry file.
- **TypeScript** — decorator or manifest file.
- **Python** — class decorator.

The underlying artifact is always the same: a `Map<string, Route>` available to `Router` at call time.

## Exception → Status Mapping

Exceptions reaching `ThrowableCaught` are translated to `ServiceResponse` objects with appropriate status. Default
mappings (configurable per application):

- `NotFoundException` → `NOT_FOUND`
- `ValidationException` → `INVALID_ARGUMENT`
- `UnauthorizedException` → `UNAUTHENTICATED`
- `ForbiddenException` → `PERMISSION_DENIED`
- `CancelledException` → `CANCELLED` or `DEADLINE_EXCEEDED` (from reason)
- Any uncaught `Throwable` → `INTERNAL`

Language-native cancellation exceptions (`context.Canceled` in Go, `asyncio.CancelledError` in Python, etc.) are
converted to `CancelledException` at the adapter boundary before reaching `ThrowableCaught`, so exception-handling code
sees a uniform type hierarchy.

Rich error details (`grpc-status-details-bin` carrying `google.rpc.Status` protobuf) can be populated by user middleware
via `Status.withDetails()`.

## Scope of What Is Not Portable

The following is unavoidably per-language and per-worker, and is not part of the framework's agnostic surface:

- Server bootstrap and port binding.
- TLS and mTLS configuration.
- Generated stubs from `.proto` files (each language's `protoc` plugin).
- Native request/response message types produced by the generator.
- Worker-specific configuration (thread pools, coroutine settings, plugin registration).
- The underlying cancellation/context primitive (Go `context.Context`, Java `io.grpc.Context`, JS `AbortSignal`, etc.) —
  Valkyrja's `CancellationToken` wraps these.

Everything above the adapter layer — service map, middleware composition, container resolution, error mapping,
cancellation model, context propagation, observability hooks — is standardized across all five languages.

## Implementation Sequence

Recommended order for building out gRPC support across the ecosystem:

1. Finalize this contracts document (language-agnostic).
2. Prototype in Java. The language's gRPC library is the most mature; design tensions surface fastest there.
3. Port to PHP via the RoadRunner adapter.
4. Add Go. Canonical gRPC implementation; trivial once contracts are proven.
5. Add Python and TypeScript. Async quirks are easier to absorb after the shape is settled.

## Summary

gRPC in Valkyrja is architecturally indistinguishable from HTTP and CLI aside from the specific shape of the Router (map
lookup), the `ServiceCall`/`ServiceResponse` contracts (typed messages instead of body bytes), and the addition of a
`CancellationToken`/`Deadline` cooperation model. The framework contributes what it always contributes: middleware,
container, dispatch, error handling, observability. The worker adapter contributes what it always contributes:
translation between an external protocol server and the framework's internal contract. Cancellation is cooperative
everywhere — the framework checks at orchestration boundaries and inside response writes; user handlers opt into deeper
cooperation via helpers or explicit checks. No new cross-cutting concepts are introduced; the existing Valkyrja
architecture extends naturally to a third protocol.
