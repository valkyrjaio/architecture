# Grpc

The Grpc component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the master copy of this component's README, which every port carries, so that
every port's document for this component reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Grpc takes a service call and produces a service response. It is a **protocol
component**, so it runs the pipeline in
[`LIFECYCLE.md`](../../convention/LIFECYCLE.md) and shares that pipeline's shape
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

## Structure

The component is the four areas every protocol has.

| Area          | Holds                                                     |
| ------------- | --------------------------------------------------------- |
| `Message/`    | the call, and everything that travels with it             |
| `Routing/`    | the methods, and what turns a call into one               |
| `Server/`     | the handler that runs a call end to end, and the adapters |
| `Middleware/` | one contract per pipeline stage, and the handlers         |

### `Message/`

| Directory       | Holds                                               |
| --------------- | --------------------------------------------------- |
| `Call/`         | one call, and which of the four shapes it is        |
| `Response/`     | what the method answers with                        |
| `Status/`       | the outcome, as the protocol's own code and message |
| `Metadata/`     | the key and value pairs a call carries each way     |
| `Deadline/`     | when the caller stops waiting                       |
| `Cancellation/` | the caller's withdrawal, and what observes it       |
| `Stream/`       | a sequence of messages, in either direction         |
| `Peer/`         | who the caller is, as the transport reports it      |

**The deadline and the cancellation are separate.** A deadline passes on its
own; a cancellation is something the caller did. A port that treated them as one
could not tell a slow call from an abandoned one.

### `Server/`

`Handler/` holds the call handler. `Adapter/` holds one class per server a port
can run on, because this protocol's server is the transport rather than the
language's own.

### Order

| In                       | The order is                                                    |
| ------------------------ | --------------------------------------------------------------- |
| the areas                | as the table above lists them                                   |
| the message directories  | the call, the response, the status, then what travels with them |
| the middleware contracts | the stages, in the order the pipeline runs them                 |
| the routing directories  | data, collection, collector, then what reads them               |

---

## Using it

### Configure it and point a transport at it

The application's gRPC config declares the middleware scheduled at each stage,
one property per stage ([`APPLICATION.md`](../application/APPLICATION.md)). An adapter for the
gRPC library you deploy on receives the call and drives the component.

### Define a service

**A service is a class, and each remote method is a method of that class.** The
service declares its full service name, in the protocol's own
`package.Service` form, and each method declares the remote method's name and
which side streams.

A method's own middleware is declared on the method, and the declaration is
repeatable, so a method may carry several.

A collector reads each class that declares a service and each method of it that
declares a remote method. It builds the routing key from the **two** names — the
service name and the method name — and wires the method as the handler.

**A route carries an optional request message type and an optional response
message type.** Each one names the generated message class. A `with` form sets
each one.

The framework reads neither type. An adapter is the consumer, because an adapter
translates between the generated message and the port's own any-or-object type. A
route that names no type still dispatches.

### Write a handler

A handler receives the container and the route, and returns a service response
([`HANDLERS.md`](../../convention/HANDLERS.md)). It resolves the call from the
container when it needs the messages, the metadata, the deadline or the peer.

Warning: a port that invokes the handler reflectively **constructs the controller
with no arguments**. A controller therefore takes its dependencies from the
container inside the method, not through a constructor.

### Register the service

A **route provider** returns the controller classes to scan and any prebuilt
route, and the application lists that provider in its config
([`PROVIDERS.md`](../../convention/PROVIDERS.md)). The component discovers nothing
on its own.

### Read the call

The messages arrive as the port's own any-or-object type. A buffered call holds
every message by the time the handler runs. A streaming call yields them as they
arrive, and the handler replies through the call's own send operation rather than
by returning.

### Return a status

Return a response carrying `OK` and the body for success, and a response carrying
the code that names the failure otherwise. Prefer returning a status to throwing,
because a status is the protocol's own vocabulary and a throwable has to be
mapped to it.

Set metadata on the response through its `with` forms. Metadata validates on
write, so an invalid header name fails where it is set.

### Handle a caller that gives up

Check the cancellation token in any loop or long computation, and stop when it is
set. The framework checks between its own steps and **never interrupts handler
code**, so a handler that does not check runs to completion after the caller has
gone.

A cancellation the framework detects produces a response rather than a throwable,
so the `ThrowableCaught` stage never sees it.

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

**Outbound writes respect backpressure.** A send does not hand the transport more
than it can take. Honoring that is the adapter's obligation for both call shapes,
because only the adapter knows the transport's writability.

Warning: most transports deliver the writability signal on the callback path. A
send that waits for that signal on the same path blocks the signal it waits for.
Backpressure therefore depends on a free callback path.

**The handler stays off the transport's callback path for both call shapes.** The
adapter dispatches the handler and returns from the callback at once, so the
callback path stays free while the handler runs.

---

## Cancellation is cooperative

A caller can give up at any time, and the framework never interrupts running
handler code. Forcible interruption is unsafe in every one of these languages, so
the model is the same everywhere.

The framework checks the cancellation token between steps and stops advancing
when it is set. A handler that runs a long computation, or calls a library that
does not know about cancellation, checks the token itself.

**A detected cancellation produces a response, not a throwable.** So cancellation
is never a special case in the `ThrowableCaught` stage, and that stage handles only real
failures.

A port maps each call to its language's own concurrency primitive. An application
that needs to cap concurrent calls layers its own limit; the component does not
impose one.

---

## The adapter

The component declares the **adapter contract** and ships no adapter. The
contract declares `start` to begin serving and `stop` to shut down, and an
implementation owns both ends of a call: taking a native call apart into the
framework's own types, and writing the framework's response back out.

An adapter names a gRPC library, so it is the one piece that cannot live in the
component. An application supplies one, or installs a package that does.

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
  required, or the `ThrowableCaught` stage sees the reflection wrapper instead of the
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
| which gRPC library an adapter bridges         | each ecosystem has its own, and the component ships no adapter |
| the concurrency primitive per call            | each language has its own                                      |
| whether the library is an optional dependency | the bridge is in core; the library is the application's choice |
| an `Attribute` routing subcomponent           | only a language with attributes declares that way              |

Nothing else varies.
