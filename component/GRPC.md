# Grpc

The **cross-language** definition of the Grpc component. It states the
hierarchy, the names and the behavior that every port implements.

A port implements this document. A port does not redefine it.

**The reference implementation for this component is Java**, not PHP. Grpc is the
one component where the reference is not the reference port
([`AGENTS.md`](../AGENTS.md) §1). When a port disagrees with Java on this
component, Java is right.

This document holds no code example. The component's `README.md` in each port
holds the examples and the per-language spelling.

---

## What the component owns

Grpc takes a service call and produces a service response. It is a **protocol
component**, so it runs the pipeline in
[`LIFECYCLE.md`](../convention/LIFECYCLE.md) and shares that pipeline's shape
with Http, Cli and Queue.

The component owns no transport. A gRPC library receives the call and an adapter
normalizes it; the component never speaks the wire protocol.

**The framework never depends on a specific server or worker.** The adapter is
the only place a library is named.

---

## Hierarchy

| Subcomponent | Holds                                                   |
| ------------ | ------------------------------------------------------- |
| `Message`    | the call, the response, and every value type they carry |
| `Routing`    | route data, collection, matching, and dispatch          |
| `Middleware` | the stage contracts and the stage handlers              |
| `Server`     | the service handler, and the transport adapter          |
| `Support`    | helpers shared across the component                     |
| `Throwable`  | the component's throwable contract and its exceptions   |

`Message` divides into `Call`, `Response`, `Status`, `Metadata`, `Deadline`,
`Cancellation`, `Peer` and `Stream`. Each is a value type with its own contract.

---

## The message is never the library's message

**A payload is the port's own any-or-object type.** The component does not adopt
the message type of any gRPC library, and bytes never reach the framework.

This is a portability rule, not a preference. A library's generated message shape
differs per ecosystem, so a component built on one library's shape cannot be
ported. Translation happens at the adapter and nowhere else.

---

## The call

A call carries the method it names, its metadata, its deadline, its cancellation
token, its peer, and its messages. It reports whether it is streaming, and it
carries the route once the router resolves one.

| Value type          | Holds                                          |
| ------------------- | ---------------------------------------------- |
| `ServiceCall`       | one inbound call and everything known about it |
| `ServiceResponse`   | the outbound messages and the status           |
| `Status`            | a code, a message, and optional details        |
| `Metadata`          | the call's headers, as a validated multi-map   |
| `Deadline`          | when the call expires                          |
| `CancellationToken` | whether the caller gave up                     |
| `Peer`              | who called, and how they authenticated         |
| `OutboundStream`    | the sink a streaming response writes to        |

Every one is immutable, with a `with` form for a change.

**A deadline and a cancellation token are never absent.** Each has a sentinel for
"none", so no caller checks for an empty value. A sentinel deadline is a finite
far-future instant rather than an unbounded one, so arithmetic on it cannot
overflow.

**Metadata validates on write.** A name must be a valid header name and a value
must match its name's kind, so an invalid header fails where the caller set it
rather than when the response is written.

---

## Status

A status carries a code from the gRPC standard set: `OK`, `CANCELLED`, `UNKNOWN`,
`INVALID_ARGUMENT`, `DEADLINE_EXCEEDED`, `NOT_FOUND`, `ALREADY_EXISTS`,
`PERMISSION_DENIED`, `RESOURCE_EXHAUSTED`, `FAILED_PRECONDITION`, `ABORTED`,
`OUT_OF_RANGE`, `UNIMPLEMENTED`, `INTERNAL`, `UNAVAILABLE`, `DATA_LOSS` and
`UNAUTHENTICATED`.

The set is the protocol's, so a port adds nothing to it and renames nothing in
it.

A status reports whether it is `OK` and whether it represents a cancellation, so
a caller tests the meaning rather than comparing codes.

---

## The two call shapes

| Shape     | Means                                                          |
| --------- | -------------------------------------------------------------- |
| buffered  | the messages are collected, then the handler runs once         |
| streaming | the handler runs while messages arrive, and replies as it goes |

**Buffered is the default**, and it is the shape the pipeline is designed around.
Every stage sees one call and one response.

Streaming is the deliberate exception. A handler that must reply before the caller
finishes sending needs a sink to push to, so a streaming call carries an outbound
stream. Without it, an interactive caller that waits for a reply before sending
more would deadlock.

**Middleware runs once per call, not once per message.** A stage is about the call,
so a streaming call does not multiply the pipeline.

**Outbound writes respect backpressure.** A send does not hand the sink more than
it can take. Honoring that is the adapter's obligation, because only the adapter
knows the transport's writability.

---

## Cancellation is cooperative

A caller can give up at any time, and the framework never interrupts running
handler code. Forcible interruption is unsafe in every one of these languages, so
the model is the same everywhere.

The framework checks the cancellation token between steps and stops advancing
when it is set. A handler that runs a long computation, or calls a library that
does not know about cancellation, checks the token itself.

**A detected cancellation produces a response, not a throwable.** So cancellation
is never a special case in the throwable stage, and that stage handles only real
failures.

A port maps each call to its language's own concurrency primitive. An application
that needs to cap concurrent calls layers its own limit; the component does not
impose one.

---

## The adapter

The adapter bridges one gRPC library to the component. It declares `start` to
begin serving and `stop` to shut down, and it owns both ends of a call: taking a
native call apart into the framework's own types, and writing the framework's
response back out.

The write happens **between** stage 6 and stage 7, which is what makes those two
stages meaningful: stage 6 can still change the response, and stage 7 runs after
it is on the wire.

---

## Wiring rules a port must follow

These are the mistakes that are silent, so each one is stated rather than left to
be rediscovered.

- **The router and the stage handlers resolve the same instances.** The
  application publishes each stage handler as a singleton. Two instances means
  per-route middleware at those stages never runs, and nothing reports it.
- **Reflective dispatch rethrows the handler's own throwable.** A language that
  invokes a method reflectively wraps whatever the target threw. Unwrapping it is
  required, or the throwable stage sees the reflection wrapper instead of the
  failure.
- **A class that implements several stage contracts is registered at every one of
  them.** How a port decides which stages a class implements depends on whether
  its type system is nominal or structural, not on the language. Either way the
  class lands in all its buckets.
- **A service is registered by the application**, exactly as a route is in Http
  and Cli. The component discovers nothing on its own.

---

## Permitted variation

| Variation                                     | Reason                                                         |
| --------------------------------------------- | -------------------------------------------------------------- |
| which gRPC library an adapter bridges         | each ecosystem has its own                                     |
| the concurrency primitive per call            | each language has its own                                      |
| whether the library is an optional dependency | the bridge is in core; the library is the application's choice |
| an `Attribute` routing subcomponent           | only a language with attributes declares that way              |

Nothing else varies.
