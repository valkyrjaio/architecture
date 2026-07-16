# Queues

This document describes how queue/job processing integrates into Valkyrja as a first-class protocol
alongside HTTP, CLI, and gRPC. The design is language-agnostic and applies to all current and planned
ports (PHP, Java, TypeScript, Go, Python).

Queues are, architecturally, close to **CLI** (a direct map lookup by a name) and **gRPC** (a
worker-agnostic core, external worker adapters, and a cooperative timeout/cancellation model). The
main new idea is the **outcome model**: a consumed message is not answered to a waiting client — it is
**acknowledged, retried, or failed**. Read this alongside [`ADDING_A_MODULE.md`](ADDING_A_MODULE.md)
and, for the cancellation/adapter patterns it reuses, [`GRPC.md`](GRPC.md) and
[`GRPC_IMPLEMENTATION.md`](GRPC_IMPLEMENTATION.md).

## Design Principles

1. **Worker-agnostic.** The framework never depends on a specific broker. Adapters bridge external
   brokers (SQS, Redis, RabbitMQ/AMQP, Beanstalkd, database, in-memory/sync) to the framework's
   internal contracts, exactly as HTTP/CLI/gRPC adapters do.

2. **Framework features are inherited, not reimplemented.** Middleware, the container, event dispatch,
   exception handling, and observability all work the same in queues as everywhere else.

3. **Response propagation, Go-style.** Unwinding uses a `JobResult` flowing back up the pipeline; each
   layer inspects it and decides how to proceed. Exceptions are a fallback: `ThrowableCaught`
   middleware converts them into a `JobResult` (typically a retry or a failure).

4. **No routing logic, just map lookup.** A message carries a **job name/type**; a direct
   `Map<name, Route>` lookup resolves it — the same shape CLI and gRPC use. The component is still
   called `Router`.

5. **Symmetry across protocols.** The pipeline shape is identical to the others:

```
HTTP:   Server  → RequestHandler → Router (pattern match) → middleware → handler
CLI:    Console → InputHandler   → Router (map lookup)    → middleware → handler
gRPC:   Server  → ServiceHandler → Router (map lookup)    → middleware → handler
Queue:  Worker  → MessageHandler → Router (map lookup)    → middleware → handler
```

## The Broker Model

A broker delivers a **message envelope** and expects an **acknowledgement decision** back. The exact
fields vary by broker; the framework models the common subset:

**Inbound (delivered):**

- A **job name/type** (the map key) — usually a message attribute or a field in the body.
- A **payload/body** (opaque bytes or a decoded structure; agnostic like gRPC messages).
- **Attributes/headers** (a metadata multi-map).
- **Delivery metadata**: message id, receive/attempt count, enqueue time, and a **visibility timeout**
  (how long this consumer "owns" the message before the broker redelivers it).
- Optionally a **priority** and a **delay/available-at**.

**Outcome (returned):**

- **Ack** — processed successfully; remove from the queue.
- **Retry / release** — put back for redelivery, optionally after a **backoff delay**; increments the
  attempt count.
- **Fail / dead-letter** — give up; route to a dead-letter queue (or drop, per policy).

Two properties shape everything:

- **At-least-once delivery.** Brokers redeliver on timeout or nack, so handlers must tolerate
  **duplicate delivery** (idempotency is a user concern the framework surfaces but cannot enforce).
- **The "response" is a decision, not a payload.** There is no client awaiting bytes. The pipeline's
  outbound value is the ack/retry/fail decision plus observability metadata.

The adapter handles broker-specific framing (deletion, visibility extension, backoff, dead-letter
routing). The framework works with decoded envelopes, attributes as structured maps, and the outcome
as a value type. Broker specifics never cross into framework territory.

## Core Contracts

The language-agnostic surface mirrors gRPC's, with queue vocabulary.

### `MessageHandler`

The kernel entry point, analogous to `ServiceHandler` (gRPC) / `RequestHandler` (HTTP). Worker
adapters hand messages to `MessageHandler.handle()`.

Responsibilities:

- Orchestrate the middleware stages (`MessageReceived`, `Acking`, `Terminated`).
- Delegate to `Router` for resolution and dispatch.
- Run `ThrowableCaught` middleware when exceptions propagate.
- Fast-exit on cancellation / visibility-timeout expiry.

As with gRPC, split the kernel so the **broker settlement** (ack/nack/extend) can happen between the
"acking" stage and "terminated": `handle` (through `ThrowableCaught`) → `acking` (always-run) →
[adapter settles with the broker] → `terminate` (always-run). A `run` convenience bundles
`handle`+`acking`.

### `QueueMessage` (immutable)

What comes in from the adapter. Models the inbound side.

```
QueueMessage
  getJob(): string                      // "SendWelcomeEmail" — the map key
  getPayload(): iterable<Message>       // decoded body/messages (agnostic type)
  getAttributes(): Attributes           // headers/attributes multi-map
  getId(): string                       // broker message id
  getAttempts(): int                    // delivery/attempt count (1-based)
  getDeadline(): Deadline               // from the visibility timeout; never null
  getCancellation(): CancellationToken  // never null
  getRoute(): Route                     // resolved route metadata
```

### `JobResult` (immutable)

What goes out — the settlement decision.

```
JobResult
  getOutcome(): Outcome                 // ACK | RETRY | FAIL
  getStatus(): Status                   // rich outcome detail (message, error)
  getRetryDelay(): Duration             // for RETRY; the backoff before redelivery
  getAttributes(): Attributes           // metadata to attach on redelivery / for observability

  isAck(): bool
  isRetry(): bool
  isFail(): bool

  static ack(): JobResult
  static retry(?Duration delay): JobResult
  static fail(?string reason): JobResult
```

### `Route` (immutable)

The value stored in the job map, keyed by job name.

```
Route
  getName(): string                     // "SendWelcomeEmail" — the map key
  getHandler(): Handler                 // class+method reference or callable
  getMiddleware(): per-stage lists
  getMaxAttempts(): int                 // before dead-lettering
  getBackoff(): BackoffPolicy           // delay strategy between retries
  getPayloadType(): class-string        // optional decode hint
```

### `Status`, `Attributes`, `Deadline`, `CancellationToken`

Reused conceptually from gRPC:

- **`Status`** — an outcome code + human-readable message + optional error detail. A queue-specific
  enum (e.g. `OK`, `RETRYABLE`, `FAILED`), distinct from HTTP/gRPC codes.
- **`Attributes`** — a case-insensitive multi-map (message attributes / headers).
- **`Deadline`** — the absolute time the **visibility timeout** expires; computed once at receipt.
  `getRemaining()` tells a long handler how much ownership time is left; the adapter may extend it.
- **`CancellationToken`** — fires on worker shutdown or visibility-timeout expiry. Same cooperative,
  poll + listener model as gRPC; deadline expiry is modeled as a cause of cancellation.

## Middleware Pipeline

```
1. MessageReceived      always runs; pre-router
2. Router resolves job from map
3a. JobMatched          runs if job found; pre-handler
    User handler runs, produces JobResult
3b. JobDispatched       runs if job was found; post-handler
 OR
3c. JobNotMatched       runs if job not found
    Default terminal produces JobResult::fail() (unknown job → dead-letter)

[if any above threw]
4. ThrowableCaught      converts throwable → JobResult (default: retry within maxAttempts, else fail)

5. Acking               always runs (including error/cancellation paths)
   Adapter settles with the broker (delete / release+backoff / dead-letter)
6. Terminated           runs after settlement (metrics, events, cleanup)
```

All stages except `MessageReceived` and `Acking` are optional. The abstract middleware `Handler` base
carries the two-question cancellation check; request-processing stages inherit it, while `Acking` and
`Terminated` always run.

### Exception → outcome mapping

`ThrowableCaught` translates exceptions to a `JobResult`. Sensible defaults (configurable per
application and overridable per route):

- A **retryable** exception (or any uncaught throwable) → `RETRY` with the route's backoff, **unless**
  `attempts >= maxAttempts`, in which case → `FAIL` (dead-letter).
- A **non-retryable** exception (bad message, validation) → `FAIL` immediately.
- A cancellation/shutdown → `RETRY` with no penalty (the message returns for another worker), since the
  work was not completed.

## Cancellation, Timeout, and Retries

- **Visibility timeout is the deadline.** Computed once at receipt; propagated as an absolute time so
  every layer agrees. If it elapses mid-handler, the broker will redeliver — the framework surfaces
  this as cancellation so cooperative handlers can stop. The adapter may **extend** visibility for
  long jobs.
- **Cooperative cancellation.** Identical to gRPC: the framework checks at orchestration boundaries and
  converts detected cancellation into a `JobResult` (a no-penalty `RETRY`); handlers opt into deeper
  cooperation via `message.cancellable(iterable)` / explicit `throwIfCancelled()`.
- **Backoff & dead-letter.** On `RETRY`, the adapter releases the message with the route's backoff
  delay (fixed, exponential, …). When `attempts` exceeds `maxAttempts`, the outcome becomes `FAIL` and
  the adapter routes to the dead-letter destination.
- **Graceful shutdown.** On worker stop, in-flight messages are cancelled → `RETRY` so no work is lost.

## Worker Adapters

Adapters bridge an external broker to `MessageHandler`. Responsibilities:

1. Poll/subscribe for messages from the broker (long-poll, blocking pop, push subscription, …).
2. Decode the message; build a `QueueMessage` (job, payload, attributes, id, attempts, deadline from
   the visibility timeout, cancellation, route).
3. Wire the `CancellationToken` to worker-shutdown and the visibility-timeout timer.
4. Invoke `MessageHandler.handle(message)` (via the worker base `dispatch`).
5. **Settle** with the broker based on the `JobResult`: delete on `ACK`, release-with-backoff on
   `RETRY`, dead-letter on `FAIL`. This is the queue analog of gRPC's "write to the wire", and slots
   between `acking` and `terminate` via the worker base's settlement callback.

Adapters may consume in **batches** and dispatch each message independently (each in its own child
container), settling per message.

### Adapter interface

```
QueueAdapter
  start(MessageHandler): void   // begin consuming (connect, subscribe, poll loop)
  stop(): void                  // graceful shutdown (stop polling, drain in-flight)
```

### Target adapters

Sync/in-memory (tests, local), database, Redis, SQS, RabbitMQ/AMQP, Beanstalkd. Broker-specific config
(connection, prefetch, visibility, dead-letter destination) lives on the adapter, not in the agnostic
contract.

## Producing (enqueuing)

Consuming is the pipeline above; producing is the symmetric other half. A framework-provided
`Dispatcher`/`Queuer` enqueues jobs: `dispatch(name, payload, options)` where options include
**delay/available-at**, **priority**, target **queue/connection**, and attributes. Producing is a thin
container service over the same adapter; it does not run the middleware pipeline (that happens on
consume). Deadline-aware producers propagate remaining budget where a job is dispatched from within
another unit of work.

## Registration

Same discovery → map pattern as the other modules:

- An attribute/annotation/decorator (e.g. `@Job(name, queue, maxAttempts, backoff)`) on handler
  classes/methods, plus a repeatable middleware attribute dispatched to its stage.
- A collector reflects (or generates) these into `Route`s keyed by job name.
- A job route-provider contract (`getControllerClasses()` + `getRoutes()`) aggregated at boot.

## Application Wiring

Mirror gRPC exactly:

- `QueueConfig` + contract — connections/queues, default per-stage middleware, worker options (prefetch,
  max attempts, backoff defaults). Bound in the application bootstrap.
- Middleware/routing/server provider pairs — stage handlers published as **shared singletons** so the
  `Router` and `MessageHandler` register/invoke the same instances.
- `getQueueProviders` added across `ComponentProviderContract`, `ApplicationContract`, the kernel, the
  child application, and every implementor.
- Application entry — a single-shot `Queue.handle(config, message)` for tests/embedding, and a
  `WorkerQueue` base: `bootstrap(config)` once, then a consume loop that `dispatch(app, data, message,
  settler)` per message into an isolated child container, with the adapter settling via the callback.

## What differs from CLI and gRPC

- **No synchronous client response.** The outbound value is an **ack/retry/fail decision**, not a
  payload to a waiting caller.
- **Retries, backoff, dead-letter, max-attempts** are first-class — the retry loop is the queue's
  defining behavior, driven by the attempt count carried on the message.
- **At-least-once + idempotency.** Duplicate delivery is expected; the framework exposes attempt count
  and message id, but idempotency is the handler's responsibility.
- **Producing is part of the module** (enqueue side), unlike CLI/gRPC which only consume.
- **Batch consumption and delayed/scheduled jobs** have no analog in the request/response modules.

## Scope of What Is Not Portable

Per-broker and per-language: connection/pool setup, visibility/prefetch/dead-letter configuration,
serialization of the payload, the underlying cancellation/context primitive, and the poll/subscribe
loop. Everything above the adapter — job map, middleware composition, container resolution, outcome
mapping, cancellation model, observability — is standardized across all ports.

## Implementation Sequence

1. Finalize this contract document.
2. Prototype in the reference port (PHP) or the most-mature secondary (Java), with a sync/in-memory
   adapter to prove the pipeline and the outcome model end-to-end.
3. Add a real broker adapter (Redis or SQS) to prove settlement, backoff, and dead-lettering.
4. Port to the remaining languages once the shape is settled.
