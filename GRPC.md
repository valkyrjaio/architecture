# gRPC

This document describes how gRPC integrates into the Valkyrja framework as a first-class protocol alongside HTTP and
CLI. The design is language-agnostic and applies to all current and planned Valkyrja ports (PHP, Java, TypeScript, Go,
Python).

## Design Principles

gRPC in Valkyrja follows the same architectural pattern already established for HTTP and CLI:

1. **Worker-agnostic.** The framework never depends on a specific server or worker implementation. Adapters bridge
   external workers to Valkyrja's internal contracts, exactly as they do for HTTP (CGI, FrankenPHP, RoadRunner, Swoole,
   etc.) and CLI.
2. **Framework features are inherited, not reimplemented.** Middleware, the container, event dispatch, exception
   handling, and observability all work the same way in gRPC as they do in HTTP and CLI. Worker implementations do not
   ship their own versions of these concerns.
3. **No router needed.** gRPC already identifies `(service, method)` at the protocol level via the `:path`
   pseudo-header (`/package.Service/Method`). Valkyrja uses a direct `Map<string, ServiceRoute>` lookup, the same shape
   CLI uses for commands, without any pattern matching or parsing.
4. **Symmetry across protocols.** The conceptual pipeline is identical for all three protocols:

```
HTTP:   Server  → RequestHandler → [Router → Route]    → Middleware → Dispatch
CLI:    Console → InputHandler   → [Parser → Command]  → Middleware → Dispatch
gRPC:   Server  → ServiceHandler → [Map → ServiceRoute] → Middleware → Dispatch
```

The bracketed segment is the only per-protocol difference. In gRPC, it is the cheapest of the three because the work is
done at the protocol layer before the framework is invoked.

## Core Contracts

The language-agnostic surface area is intentionally small. Five contracts are sufficient.

### `ServiceHandler`

The kernel entry point for gRPC, analogous to `RequestHandler` (HTTP) and `InputHandler` (CLI). Worker adapters hand
calls to `ServiceHandler::handle()`; everything downstream is pure Valkyrja.

Responsibilities:

- Look up the `ServiceRoute` in the service map by fully-qualified method name.
- Compose and execute the middleware pipeline for that route.
- Resolve the handler via the container and dispatch.
- Map exceptions to gRPC status codes on the way out.

### `ServiceRoute`

The value stored in the service map, analogous to `Route` (HTTP) and `Command` (CLI).

Carries:

- Handler reference (class + method, or callable).
- Middleware stack for the service/method.
- Request and response message type information (from the `.proto`).
- Streaming flags (`isClientStreaming`, `isServerStreaming`) as metadata.
- Any per-route configuration (deadlines, auth requirements, etc.).

### `ServiceCall`

What comes in from the worker adapter. Wraps:

- Fully-qualified method name.
- Incoming messages as an iterable/stream (length 1 for unary and server-streaming calls).
- Metadata (inbound headers).
- Deadline and cancellation signal.
- Peer information.
- Correlation/trace identifiers.

### `ServiceResponse`

What goes out. Wraps:

- Outgoing messages as a sink/writer (accepts one write for unary and client-streaming calls).
- Outgoing metadata (trailers and initial headers).
- Status code and message (on error paths).

### `Interceptor`

The gRPC-specific name for middleware, aligned with terminology already used by gRPC libraries in every target language.
Operates on `ServiceCall` and produces a `ServiceResponse`, composing identically to HTTP and CLI middleware. Existing
framework middleware contracts are reused; the only difference is the shape of the call/response objects being passed
through.

## Streaming

Streaming is treated as the primitive, with unary as a degenerate case. This avoids the lock-in that would occur if
unary were the primitive and streaming were bolted on later.

gRPC's four call patterns reduce to two boolean dimensions:

|                       | client streaming: no | client streaming: yes |
|-----------------------|----------------------|-----------------------|
| server streaming: no  | Unary                | Client-streaming      |
| server streaming: yes | Server-streaming     | Bidirectional         |

The `ServiceCall.messages` field is always an iterable/stream; for unary and server-streaming calls it yields exactly
one message. The `ServiceResponse` always exposes a sink; for unary and client-streaming responses it accepts exactly
one write. This keeps the contract stable whether only unary is implemented initially or all four patterns are supported
from day one.

A `@Unary` convenience wrapper (name adapted per language's conventions) adapts a plain `(Request) -> Response` handler
into the streaming contract, giving developers the simpler ergonomics for the common case without altering the
underlying architecture.

Each target language has a native iterable + sink primitive to implement these:

| Language   | Incoming stream                     | Outgoing sink             |
|------------|-------------------------------------|---------------------------|
| PHP        | `Generator` / RR channels           | `Generator` / RR channels |
| Java       | `Flow.Publisher` / `StreamObserver` | `StreamObserver`          |
| TypeScript | `AsyncIterable<T>`                  | writable stream           |
| Go         | channel `<-chan T`                  | channel `chan<- T`        |
| Python     | `AsyncIterator[T]`                  | `AsyncGenerator[T, None]` |

The contract stays identical; the primitive used to satisfy it is whichever is idiomatic in each language.

## Worker Adapters

Worker adapters bridge an external gRPC server implementation to `ServiceHandler`. The adapter's entire job is:

1. Receive the native call representation from the worker.
2. Wrap it in a `ServiceCall`.
3. Invoke `ServiceHandler::handle()`.
4. Unwrap the `ServiceResponse` back into the worker's native form.

This mirrors the existing adapter pattern used for HTTP across CGI, worker, and runtime-embedded modes, and for CLI
across different console implementations.

### Target Adapters by Language

**PHP**

- RoadRunner (`spiral/roadrunner-grpc`) — primary recommended adapter; most mature PHP gRPC option.
- OpenSwoole (`Swoole\GrpcServer` / `openswoole/grpc`) — coroutine-based alternative; useful for streaming-heavy
  workloads.
- FrankenPHP — deferred until the ecosystem provides a native gRPC termination path into PHP workers.

**Java**

- `grpc-java` with `ServerBuilder` — canonical. Generated `BindableService` implementations delegate to
  `ServiceHandler::handle()` with `StreamObserver` adapted to `ServiceCall`/`ServiceResponse`.

**TypeScript**

- `@grpc/grpc-js` — the standard Node.js gRPC library.

**Go**

- `google.golang.org/grpc` — generated service interface methods wrap parameters into `ServiceCall` and forward.

**Python**

- `grpcio` (async API) — handlers adapt `ServicerContext` and request iterables into `ServiceCall`.

The adapter in each language is expected to be small (roughly fifty lines of glue code). All protocol-framework
integration — middleware, container resolution, error mapping, observability — lives above the adapter in Valkyrja code
that is unaware of which worker is running.

## Service Registration

Service registration follows the discovery → map pattern already used by HTTP routes and CLI commands. A `@GrpcService`
annotation (or language-idiomatic equivalent: PHP attribute, TS decorator, Go build-time tag, Python class decorator)
marks generated service implementations. A build-time or boot-time scan populates the service map keyed by
fully-qualified method name.

In Java, this aligns with the annotation-processor-plus-data-class-generation approach already adopted for `@Provides`.
The same `@GrpcService` processor generates the service map as a data class at compile time.

In PHP, attribute scanning populates the map at boot, with the result cached via the existing data class generation
mechanism (`App\Grpc\Data` namespace following the pattern established by `App\Http\Data` and `App\Cli\Data`).

Equivalent mechanisms apply in the remaining languages, always producing the same underlying artifact: a
`Map<string, ServiceRoute>` available to `ServiceHandler` at call time.

## Error Handling

gRPC status codes are produced by the standard Valkyrja exception handling pipeline, augmented with a gRPC-specific
output mapping. The exception handler already responsible for translating exceptions to HTTP responses gains an
equivalent path that maps exceptions to `(status_code, status_message, trailers)` triples. No separate status mapper
contract is required; this is an output format of the existing exception handling system.

## Context Propagation

Per-call context (deadline, cancellation, metadata) is carried on the `ServiceCall` object and passed explicitly through
the middleware pipeline, rather than through ambient or thread-local storage. This choice:

- Aligns with Go's idiomatic explicit-context style.
- Works uniformly in PHP, which lacks ambient context primitives.
- Is consistent with how HTTP and CLI already pass request/input objects.
- Keeps middleware and handlers trivially testable.

In Java and Python, where ambient context is idiomatic in some codebases, the explicit `ServiceCall` remains the
contract; adapters may bridge to `io.grpc.Context` or `contextvars` at the boundary if interop with other libraries
requires it.

## Scope of What Is Not Portable

The following is unavoidably per-language and per-worker, and is not part of the framework's agnostic surface:

- Server bootstrap and port binding.
- TLS configuration.
- Generated stubs from `.proto` files (each language's `protoc` plugin).
- Native request/response message types produced by the generator.
- Worker-specific configuration (thread pools, coroutine settings, plugin registration).

Everything above the adapter layer — service map, middleware composition, container resolution, error mapping, context
propagation, observability hooks — is standardized across all five languages.

## Implementation Sequence

The recommended order for building out gRPC support across the ecosystem:

1. Finalize this contracts document (language-agnostic).
2. Prototype in Java. The language's gRPC library is the most mature; design tensions surface fastest there.
3. Port to PHP via the RoadRunner adapter.
4. Add Go. Canonical gRPC implementation; trivial once contracts are proven.
5. Add Python and TypeScript. Async quirks are easier to absorb after the shape is settled.

## Summary

gRPC in Valkyrja is architecturally indistinguishable from HTTP and CLI aside from the bracketed discovery step in the
request pipeline. The framework contributes what it always contributes: middleware, container, dispatch, error handling,
observability. The worker adapter contributes what it always contributes: translation between an external protocol
server and the framework's internal contract. No new cross-cutting concepts are introduced; the existing Valkyrja
architecture extends naturally to a third protocol.
