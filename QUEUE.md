# Queues

This document describes how queue/job processing integrates into Valkyrja as a first-class protocol alongside HTTP, CLI,
and gRPC. The design is language-agnostic and applies to all current and planned ports (PHP, Java, TypeScript, Go,
Python).

Queues reuse the same shape as the other three protocol modules — **Http**, **Cli**, and **gRPC**: a worker-agnostic
core, external adapters, and a flat map lookup by name. The main new idea is the **outcome model**: a consumed message
is not answered to a waiting client — it is **acknowledged, retried, or dead-lettered**. Read this alongside
[`ADDING_A_MODULE.md`](ADDING_A_MODULE.md) and, for the adapter patterns it reuses, [`GRPC.md`](GRPC.md) and
[`GRPC_IMPLEMENTATION.md`](GRPC_IMPLEMENTATION.md).

## Design Principles

1. **Worker-agnostic.** The framework never depends on a specific broker. Adapters bridge external brokers (SQS, Redis,
   RabbitMQ/AMQP, Beanstalkd, database, in-memory/sync) to the framework's internal contracts, exactly as HTTP/CLI/gRPC
   adapters do.

2. **Framework features are inherited, not reimplemented.** Middleware, the container, event dispatch, exception
   handling, and observability all work the same in queues as everywhere else.

3. **Response propagation, Go-style.** Unwinding uses a `JobResult` flowing back up the pipeline; each layer inspects it
   and decides how to proceed. Exceptions are a fallback: `ThrowableCaught`
   middleware converts them into a `JobResult` (typically a retry or a failure).

4. **No routing logic, just map lookup.** A message carries a **job name/type**; a direct
   `Map<name, Route>` lookup resolves it — the same shape CLI and gRPC use. The component is still called `Router`.

5. **Symmetry across protocols.** The pipeline shape is identical to the others:

```
HTTP:   Server  → RequestHandler → Router (pattern match) → middleware → handler
CLI:    Console → InputHandler   → Router (map lookup)    → middleware → handler
gRPC:   Server  → ServiceHandler → Router (map lookup)    → middleware → handler
Queue:  Worker  → JobHandler → Router (map lookup)    → middleware → handler
```

## The Broker Model

A broker delivers a **message envelope** and expects an **acknowledgement decision** back. The exact fields vary by
broker; the framework models the common subset:

**Inbound (delivered):**

- A **job name/type** (the map key) — usually a message attribute or a field in the body.
- A **payload/body** (opaque bytes or a decoded structure; agnostic like gRPC messages).
- **Attributes/headers** (a metadata multi-map).
- **Delivery metadata**: message id, receive/attempt count, enqueue time, and a **visibility timeout**
  (how long this consumer "owns" the message before the broker redelivers it).
- Optionally a **priority** and a **delay/available-at**.

**Outcome (returned):**

- **Ack** — processed successfully; remove from the queue.
- **Retry / release** — put back for redelivery, optionally after a **backoff delay**; increments the attempt count.
- **Fail / dead-letter** — give up; route to a dead-letter queue (or drop, per policy).

Two properties shape everything:

- **At-least-once delivery.** Brokers redeliver on timeout or nack, so handlers must tolerate **duplicate delivery**
  (idempotency is a user concern the framework surfaces but cannot enforce).
- **The "response" is a decision, not a payload.** There is no client awaiting bytes. The pipeline's outbound value is
  the ack/retry/fail decision plus observability metadata.

The entry handles broker-specific framing (deletion, visibility extension, backoff, dead-letter routing). The
framework works with decoded envelopes, attributes as structured maps, and the outcome as a value type. Broker specifics
never cross into framework territory.

## Wire Envelope

The **cross-language interop contract**: the one JSON document any port serializes when it enqueues and deserializes
when it consumes. A `Job` published by the PHP port must run unchanged on the Go, Java, TypeScript, or Python port, and
vice versa. This is the one place in the design where the exact bytes matter — treat it as a versioned contract, not an
implementation detail.

It is **HTTP-shaped**, and that mental model governs the whole envelope:

| Envelope     | HTTP analog         | Role                                                       |
| ------------ | ------------------- | ---------------------------------------------------------- |
| `name`       | request line (path) | the routing key                                            |
| `attributes` | headers             | cross-cutting metadata a producer stamps on every job      |
| `payload`    | body                | the job-specific data                                      |
| `producer`   | `User-Agent`        | provenance (promoted to a first-class, auto-stamped field) |

Two rules make it portable, and everything else follows from them:

1. **`name` is the only routing key, and it is a plain string.** No class names, no fully-qualified types, no
   language-specific references anywhere in the envelope. It resolves to a handler through each port's own `Router` map.
   It _must_ travel in the envelope: the broker hands over an opaque blob with no request line, so the routing key has
   to ride inside.
2. **`payload` is a self-contained, language-agnostic JSON object** carrying everything the job needs. Binary data is
   base64-encoded inside a field the job itself defines (e.g. `{"image_b64": "…"}`); the envelope never carries opaque
   bytes, an encoding tag, or a decode hint.

### Schema

```json
{
  "id"                              : "01JABCDEF0123456789ABCDEFG",
  "name"                            : "SendWelcomeEmail",
  "producer"                        : "AuthService php/26.2.3",
  "attributes"                      : {
    "tenant" : [
      "acme"
    ]
  },
  "attempts"                        : 1,
  "max_attempts"                    : 5,
  "priority"                        : 0,
  "delay_ms"                        : 0,
  "retry_delay_ms"                  : 1000,
  "retry_delay_multiply_by_attempt" : false,
  "enqueued_at_ms"                  : 1768564798000,
  "enqueued_at_iso"                 : "2026-07-16T11:59:58.000Z",
  "modified_at_ms"                  : 1768564798000,
  "modified_at_iso"                 : "2026-07-16T11:59:58.000Z",
  "payload"                         : {
    "user_id" : 42
  }
}
```

| Field                             | Type                   | Default             | Description                                                                                                                                                                                 |
| --------------------------------- | ---------------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`                              | string                 | generated (VLID V1) | A **VLID V1** (`Type/Vlid`). Producer-generated, **stable across retries** — the dedup/idempotency key and trace-correlation id; also gives DB-backed queues clustered-index locality.      |
| `name`                            | string                 | — (caller-supplied) | Routing key — the `Router` map key, read as `Job.getName()`. Plain string; never a code reference.                                                                                          |
| `producer`                        | string                 | auto-stamped        | Provenance `AppName lang/version`. `AppName` comes from the config of the pushing application, `lang` is fixed per port, and `version` comes from `ApplicationInfo`. Trace-only.            |
| `attributes`                      | object (`str → [str]`) | `{}`                | The headers multi-map. Empty = `{}`.                                                                                                                                                        |
| `attempts`                        | int                    | `1`                 | 1-based delivery count. Framework-incremented on re-queue redelivery; normalized to `Job.getAttempts()` at consume. The retry ramp multiplies by the count before that increment.           |
| `max_attempts`                    | int                    | `5`                 | Ceiling before dead-lettering. The producer sets it.                                                                                                                                        |
| `priority`                        | int                    | `0`                 | Higher runs sooner where the processor supports it.                                                                                                                                         |
| `delay_ms`                        | int                    | `0`                 | Initial hold before the job is eligible; `0` = immediate. Producer-authored intent, applied on first enqueue only.                                                                          |
| `retry_delay_ms`                  | int                    | `1000`              | Hold before a _retry_ re-enqueue. The producer sets it. A `0` retries at once, so a failing dependency gets no time to recover. Durable adapters honor it; internal adapters retry at once. |
| `retry_delay_multiply_by_attempt` | bool                   | `false`             | When `true`, the retry hold is `retry_delay_ms × attempts` (linear ramp, self-bounding via `max_attempts`); `false` = fixed. No jitter, no policy object.                                   |
| `enqueued_at_ms`                  | int                    | stamped at enqueue  | Epoch **milliseconds** first enqueued. Authoritative.                                                                                                                                       |
| `enqueued_at_iso`                 | string                 | stamped at enqueue  | RFC 3339 UTC rendering of `enqueued_at_ms`. Informational only.                                                                                                                             |
| `modified_at_ms`                  | int                    | `= enqueued_at_ms`  | Epoch **milliseconds** the envelope was last re-written; initialized to the enqueue time, bumped on the re-queue redelivery path. Authoritative.                                            |
| `modified_at_iso`                 | string                 | `= enqueued_at_iso` | RFC 3339 UTC rendering of `modified_at_ms`. Informational only.                                                                                                                             |
| `payload`                         | object                 | `{}`                | The body. Self-contained JSON; empty = `{}`, never `null`. No code/type references.                                                                                                         |

**Every field is always present — on the object and the wire.** There is no omit-when-default:
variability lives only in the _values_ (which `attributes` keys exist, what `payload` holds, the numbers and times).
**Empty ≠ absent** — `attributes` and `payload` may be `{}` but are never dropped. This mirrors an HTTP message, whose
top-level structure is fixed while the headers and body vary.

Field order is identity → routing → provenance → headers → scheduling/retry → timestamps → **body last**
(`payload` can be large, so it trails, HTTP-style).

### Encoding rules

- **Naming:** `snake_case`, always. Each port maps it to native casing internally.
- **Time:** every absolute instant is a pair — `<name>_at_ms` (epoch milliseconds, UTC, **authoritative**, the only
  value code reads) plus an optional `<name>_at_iso` (RFC 3339, `Z`, millisecond precision
  `.SSS`, **informational**). Consumers **must not** parse `_iso` for logic; on any conflict `_ms` wins. All
  **durations** are integer milliseconds with a `_ms` suffix — no `_iso` twin.
- **Payload:** a JSON object, self-contained, zero code/type references. Binary → base64 in a job-defined field.
- **Always-present:** every first-class field is written on every envelope, defaults included (`0`,
  `{}`, `modified_at = enqueued_at`) — no omit-when-default, no absent-vs-default ambiguity.
- **Forward compatibility:** consumers **ignore unknown top-level fields** and **default any field a (possibly older)
  producer didn't send**, so the contract can still gain fields over time without breaking older producers.

### One class, produced and consumed

A queue uses a **single message class — `Job`** — for both directions. This **matches gRPC**, whose `Message` is the
same type inbound and outbound, and **differs from Http** (`Request`/`Response`) and **Cli** (`Input`/`Output`), which
carry a distinct type each way. The producer builds a `Job` and dispatches it; the consumer receives the same `Job`.
There is no separate response envelope: the handler returns a **`JobResult`** (the `ACK | RETRY | FAIL | DEAD_LETTER`
outcome enum), not another message. So the whole pipeline is **`Job` in → `JobResult` out**.

A producer can therefore ship **only the fields above** — the data envelope, nothing else. There is no settable
"response" with headers, a URL, or a status the way HTTP lets you _build_ a `Response`: **all transport is the
framework's** (delivery, settlement, redelivery, dead-lettering), split between the client that pushes and the entry
that consumes. The envelope is data; the outcome is an enum; everything in between belongs to those two. And because
`attributes` is the headers equivalent, it gets a first-class data class exactly as HTTP headers do (see `Attributes`
under [Core Contracts](#core-contracts)) — not a raw map a handler pokes at.

### What is _not_ in the envelope, and why

- **`queue`** — addressing, not body. The consumer is bound to its queue by config and the producer targets it through
  the connection, exactly as you don't name the destination server inside an HTTP request body. (Contrast `name`, which
  must ride inside — there is no request line.)
- **`version`** (a schema discriminator) — without upcaster logic a consumer facing an unknown version can only nack,
  which recovers nothing, and breaking envelope changes are coordinated events anyway. The one useful thing a
  version-like field could give — _who produced this_ — is served by `producer`.
- **`payload_type` / any class-string** — a PHP class name is meaningless to a Go consumer. `name` resolves the handler;
  `payload` carries the data. There is no decode hint anywhere — the payload is self-describing JSON.
- **Broker delivery metadata** — the native message id, receive handle and **visibility-timeout deadline** never cross
  the wire and never reach the `Job`. The entry holds them for the one delivery it is settling, so a consumed `Job` is
  the deserialized envelope and nothing more.

### Rejected alternatives (decision log)

- **`available_at` (absolute instant) → `delay_ms` (relative).** `enqueued_at_ms` already anchors a relative delay, so
  the absolute form is redundant; scheduling is expressed as durations (like
  `retry_delay_ms`), and absolute wall-clock scheduling is a _scheduler_ concern that enqueues with no delay when it
  fires.
- **Epoch-only or ISO-only timestamps → both.** Epoch for code (unambiguous, no s-vs-ms trap), ISO for humans (readable
  on a dead-letter queue). The extra bytes are meaningless next to broker I/O.
- **Bare-default timestamp (unsuffixed) → always suffixed.** A bare `enqueued_at` integer reintroduces the unit
  ambiguity `_ms` exists to kill; every millisecond value carries `_ms`.
- **`_utc` → `_iso` for the string half.** `_ms` and `_iso` are both _format_ labels (same axis); `_utc`
  would name the zone instead, and both fields are UTC anyway.
- **`date_`/`ms_` prefixes → `_ms`/`_iso` suffixes.** Suffixes match the duration convention and keep an instant's two
  views adjacent.
- **Routing key named `queue`, `type`, or `job` → `name`.** `queue` is the ingest point (the server/console analog),
  not the discriminator; the wire key is the `Job`'s `name` (read via `Job.getName()`), which keeps it consistent with
  the single message class.
- **Attributes folded into `payload` → kept separate.** Cross-cutting metadata a producer stamps on every job (tenant,
  trace id, region) is headers, not body; burying it in `payload` forces every handler to dig it out.
- **Retry fields (`attempts`, `max_attempts`, `delay_ms`, `modified_at`) moved to a processor-only header → kept
  first-class on `Job`.** Splitting them out would make the envelope shape _conditional on the processor_, the exact
  thing the cross-processor contract exists to prevent — and framework-requeue processors need them in the body anyway
  (the client rewrites the whole `Job`). Instead the shape stays uniform and only the **sourcing** varies: the
  entry reads the value from the wire body (framework-requeue) or from the processor's native counter/headers
  (processor-owned) and normalizes it into `Job.getAttempts()`. Same field everywhere; sourced correctly per redelivery
  model.
- **Re-applying `delay_ms` on every retry → applied on first publish only.** `delay_ms` is producer-authored intent; the
  `Client` applies it once, to the processor's native delay, at enqueue. Retries are timed elsewhere — processor-owned
  retries by the processor's own backoff/visibility, framework-requeue retries by `retry_delay_ms` — so `delay_ms` never
  re-fires. It stays on the envelope as a record of intent (like `enqueued_at`), and is simply inert when the processor
  controls attempts; that harmlessness is why we did **not** move the retry fields to a processor-only header.

## Module Structure

`Queue` mirrors the Http module one-to-one:

```
Queue/
  Client       // produce side — push(Job) + every publish adapter (Sync, Deferred, InMemory, Guzzle, SQS, …)
  Message      // Job, JobResult, Attributes, Payload, JobFactory
  Middleware   // the pipeline stage handlers
  Routing      // Route, Router, RouteCollection, the @Route attribute + collector
  Server       // JobHandler + the ThrowableCaught middleware it ships
```

`Message` is the analog of `Http/Message` and `gRPC/message` — the category housing the message and its parts, with`Job`
as the class inside (just as `Http/Message` houses `Request`, not a class literally named `Message`).

### Where the concrete adapters and entry points live

Producing and consuming are organized asymmetrically, for the same reason Http is:

- **Producer (`Client`) adapters live _in_ the module** (`Queue/Client`) — one lightweight class per processor (`Sync`,
  `Deferred`, `InMemory`, a Guzzle/HTTP push, SQS, Redis, …), exactly like `Http/Client`'s adapters. Pushing is cheap
  (serialize + send) and you push from anywhere, so the framework bundles support for any and all external pushes.
- **Consumer _entry points_ live in `Application/Entry`** — the bootable classes that select the config and drive
  `JobHandler`. **`Queue`** runs one job and exits. **`WorkerQueue`** is the base that boots once and gives each job a
  fresh child container. **`InternalQueue`** runs every job an internal client pushes, **`PullQueue`** holds the poll
  loop that a per-processor entry extends, and **`PushQueue`** (CGI, on the language's built-in HTTP handler) answers a
  pushing processor. All of them ship out of the box, in three places. `Queue` and `PushQueue` are concrete, and sit
  directly in `Application/Entry`. `WorkerQueue`, `InternalQueue` and `PullQueue` are abstract bases, and sit in
  `Application/Entry/Abstract`. A per-processor pull entry such as `Redis/RedisQueue` sits in a segment of its own. An
  application extends `InternalQueue` itself and implements its one abstract member, `getConfig()`, which returns the
  config of the queue application that entry boots. A sync or deferred client config names that subclass. Only
  **`PushWorkerQueue`** is per-web-server-runtime and lives in that server's repo, exactly as the Http and gRPC worker
  entries do. It stays **thin**: it _composes_ the reusable, per-processor **mapper** and **response mapping** (which
  live in the Queue module) rather than reimplementing them (see [Push vs. pull](#push-vs-pull--who-initiates)).

**Running a consumer is the dev's to wire — the framework ships entries, not a server.** It never ships an HTTP server
or a `queue:work` command; it ships the entries and you point a runtime at the bootstrap, exactly as Http ships CGI +
worker entries and you point nginx/php-fpm or a Swoole/RoadRunner runtime at `index.php`:

- **Pull** (a `PullQueue` entry, the default) is a plain long-running loop run under the dev's process manager (systemd
  / supervisor / a Docker `CMD` / a k8s `Deployment`) — **no server**, works in every language. Graceful shutdown (stop
  → in-flight `→ RETRY`) and an optional bounded lifetime (`--max-jobs`/`--max-time`-style self-exit) let the supervisor
  cycle the process for memory hygiene.
- **Push** (`PushQueue` / `PushWorkerQueue`) rides on your **existing HTTP server** — CGI out of the box, or your worker
  runtime. The processor POSTs; the entry maps the body → `Job`. No queue-specific server (see
  [Push vs. pull](#push-vs-pull--who-initiates)).

The dev hooks a runtime to the bootstrap; the framework owns everything from the entry inward.

## Core Contracts

The language-agnostic surface mirrors Http/Cli/gRPC, with queue vocabulary.

### `JobHandler`

The kernel entry point, analogous to `ServiceHandler` (gRPC) / `RequestHandler` (HTTP). A worker entry hands each job to
`JobHandler.run()`, or to `JobHandler.handle()` when it has to act before settlement.

Responsibilities:

- Orchestrate the middleware stages (`JobReceived`, `SettlingResult`, `ResultSettled`).
- Delegate to `Router` for resolution and dispatch.
- Run `ThrowableCaught` middleware when exceptions propagate.

As in Http/Cli/gRPC, split the kernel so the **broker settlement** (ack/nack/extend) can happen between the
`SettlingResult` stage and `ResultSettled`: `handle` (through `ThrowableCaught`) → `settlingResult` (always-run) →
[the entry settles with the broker] → `resultSettled` (always-run). A `run` convenience bundles
`handle`+`settlingResult`. Each middleware method matches its stage type name (`settlingResult`, `resultSettled`, …) so
a single class can implement multiple middleware stages without method collisions.

### `Router`

Resolves a `Job` to its `Route` via the flat map and dispatches it through the per-route middleware — the same shape as
the Cli/Http/gRPC routers, with `Job`/`JobResult` swapped in. `JobHandler` hands the
`Job` to the `Router`, which figures out how to handle and route it.

```
Router
  dispatch(Job): JobResult              // resolve from the map, then dispatch
  dispatchRoute(Job, Route): JobResult  // dispatch a pre-resolved route
```

A missing map entry routes to `RouteNotMatched` (default terminal: `FAIL` → dead-letter) — the analog of gRPC's
`UNIMPLEMENTED`.

### `Job` (immutable)

The **single message class for both directions** — no separate request/response split
(see [One class, produced and consumed](#one-class-produced-and-consumed)). A producer builds a `Job` and dispatches it;
the consumer receives the same `Job`. It is **immutable**, exactly like Http `Request` / Cli `Input`: a `Job` is known
at ingest and never mutated in place — the framework only ever produces a _new_ one via `with*` (attempts incremented,
etc.) until, at the very end, an entry decides whether to re-queue it based on the processor. On **produce** the
framework stamps `id`, `producer`, `enqueued_at` (and `attempts` = 1); on **consume** the entry normalizes `attempts`
from the processor. It is the in-memory form of the [Wire Envelope](#wire-envelope).

```
Job   // immutable — every with* returns a new Job, like Http Request / Cli Input
  getName(): string                         // wire `name` — the Router map key
  getPayload(): Payload                     // the JSON body, matching Http's getParsedJson()
  getAttributes(): Attributes               // the headers data class (like Http headers)
  getProducer(): string                     // provenance, "AppName lang/version"
  getId(): string                           // the VLID V1 — stable across retries
  getAttempts(): int, getMaxAttempts(): int, getPriority(): int
  getDelayMs(): int, getRetryDelayMs(): int, getRetryDelayMultiplyByAttempt(): bool
  getEnqueuedAtMs(): int, getEnqueuedAtIso(): string
  getModifiedAtMs(): int, getModifiedAtIso(): string
  with*(…): Job                             // one immutable setter per field above
```

### `JobResult` (enum)

The "response" — the settlement decision and nothing else. Like Cli's `ExitCode`: a closed set the entry reads and
acts on, carrying no payload. Not every processor can pass detail back (a push processor answers with an HTTP status),
so a result never carries any.

The contrast with Cli is instructive: Cli returns an `Output` _object_ (with an `ExitCode` inside) because output is
legal at **every** lifecycle stage. A queue job's outcome isn't — after a `RETRY`/`FAIL` there is nothing more to do
_to the job itself_, though later stages still receive the `Job`. So the result stays a bare enum; the `Job`, not the
result, carries the detail.

```
JobResult   // ACK | RETRY | FAIL | DEAD_LETTER
```

- **`ACK`** — success; remove from the queue.
- **`RETRY`** — put the job back for redelivery after a hold. **The hold is the `Job`'s `retry_delay_ms`, fixed.** A
  producer can opt into a linear ramp with `retry_delay_multiply_by_attempt`, which is `false` by default; the hold is
  then `retry_delay_ms × attempts`, where `attempts` is the count on the delivery that failed. A handler returns this
  outcome. The framework converts it to `DEAD_LETTER` once `attempts` reaches `max_attempts`.
- **`FAIL`** — the handler gives up _on purpose_ (non-retryable: bad payload, validation) → dead-letter now, no retries.
  Handler-returned.
- **`DEAD_LETTER`** — the framework exhausted `max_attempts` on a retry chain → dead-letter. Framework-produced, not
  handler-returned; distinct from `FAIL` so the two ways a job dies are told apart.

Failure detail (the throwable, a reason) is logged by `ThrowableCaught` when it happens. Distinguishing the four
outcomes _after the fact_ is a **testing concern only** — in production the outcome just drives settlement directly
(re-enqueue / throw / record) and is never read back. For tests, a **middleware fixture** at the `ResultSettled` stage
keeps an in-memory `Job.id → [JobResult…]` map, so a job's whole life reads back as `[Ack]`, `[Fail]`, or
`[Retry, Retry, DeadLetter]`. That fixture exists **specifically to test the middleware, `Client`s, and entry classes**,
and is not a production mechanism. The stage runs on every job that produces an outcome, so the map records what the
pipeline settled rather than what any one client did with the outcome.

### `Route` (immutable)

The value stored in the job map, keyed by job name. It answers only **"for this name, here is the handler"** — it holds
**no** retry/attempts policy. Those ride on the `Job`, because the **producer** decides them, not the consumer.

```
Route
  getName(): string          // "SendWelcomeEmail" — the map key
  getHandler(): Handler      // class+method reference or callable
  getMiddleware(): per-stage lists
```

### `Attributes`

The **headers data class**: a first-class, immutable, case-insensitive multi-map, housed and passed exactly as HTTP
houses request/response headers (not a raw map a handler pokes at). It is the envelope's `attributes` field. (Any
visibility-timeout / message-ownership window is a **broker concern the entry manages** — it is not a framework
contract on the `Job`.)

## Middleware Pipeline

```
1. JobReceived      always runs; pre-router
2. Router resolves job from map
3a. RouteMatched        runs if job found; pre-handler
    User handler runs, produces JobResult
3b. RouteDispatched     runs if job was found; post-handler
 OR
3c. RouteNotMatched     runs if job not found
    Default terminal produces JobResult::fail() (unknown job → dead-letter)

[if any above threw]
4. ThrowableCaught      converts throwable → JobResult (default: RETRY within maxAttempts, else DEAD_LETTER)

5. SettlingResult       always runs (including error paths)
   Entry settles (delete / release + retry_delay / dead-letter)
6. ResultSettled        runs after settlement (metrics, events, cleanup)
```

`JobReceived`, `SettlingResult`, and `ResultSettled` are unconditional — they run on every job that produces an outcome,
including error paths. Debug is the one exception: a throwable then leaves `handle` instead of becoming an outcome. No
stage runs after that throw, which costs the job its `ThrowableCaught` stage as well as both settlement stages. The
remaining stages are conditional, running whenever their case applies: `RouteMatched`/`RouteDispatched` when the job
resolves, `RouteNotMatched` when it doesn't, and `ThrowableCaught` when an earlier stage throws.

### Exception → outcome mapping

`ThrowableCaught` translates exceptions to a `JobResult`. Sensible defaults (configurable per application and
overridable per route):

- A **retryable** exception (or any uncaught throwable) → `RETRY` (re-enqueued after `retry_delay_ms`), **unless**
  `attempts >= max_attempts`, in which case → `DEAD_LETTER`.
- A **non-retryable** exception (bad message, validation) → `FAIL` immediately.
- A worker shutdown → `RETRY` with no penalty (the message returns for another worker), since the work was not
  completed.

## Worker Entries

The entry of a processor is the queue protocol's direct analog of the entry classes in the other protocols: Http's
server + `RequestHandler`, Cli's console + `InputHandler`, gRPC's `ServiceAdapter` + `ServiceHandler`. It owns both ends
of a delivery and nothing in between: the **receipt** (accept a native delivery from the processor and normalize it into
a `Job`) and the **response** (take the `JobResult` the kernel returns and settle it back with the processor). Routing,
middleware, and the handler are all processor-agnostic; only the entry knows what a Cloud Tasks POST or an SQS receipt
looks like. "Processor" is the umbrella term here — a message broker (SQS, AMQP, Redis) or a managed platform (Cloud
Tasks, Lambda, Pub/Sub push).

Receipt and response stay **clean** — plain `Job` in, `JobResult` out — up to the point where they must be mapped onto
a specific processor's runtime (e.g. OpenSwoole in PHP); that translation is the entry's only real work, and it is the
same idea on both sides of a delivery.

An entry bridges an external processor to `JobHandler`. Responsibilities:

1. Poll/subscribe for messages from the broker (long-poll, blocking pop, push subscription, …).
2. Decode the message; build a `Job` (name, payload, attributes, id, attempts).
3. Run the job through `JobHandler` (`JobHandler.run()`, or `handle` and `settlingResult` apart).
4. **Settle** with the processor based on the `JobResult` outcome (see [The outcome is an enum](#the-outcome-is-an-enum)
   and [Redelivery](#redelivery-who-performs-a-retry)). It slots between `settlingResult` and `resultSettled`.

An entry may consume in **batches** and dispatch each message independently (each in its own child container), settling
per message.

### The outcome is an enum

What the kernel hands back for settlement is a small, closed set — a `JobResult`: `ACK | RETRY | FAIL | DEAD_LETTER`,
exactly like Cli's `ExitCode`. The entry reads the enum and acts, nothing more. That closed outcome is what lets one
processor-agnostic kernel drive every processor — turning `ACK`/`RETRY`/`FAIL` into processor-specific action is the
entry's whole job on the response side.

### Redelivery: who performs a retry

Who actually performs a `RETRY` depends on the processor, and the entry of that processor encapsulates the difference:

- **Internal clients** — the client settles every outcome, for `Sync` and `Deferred`. The client is the processor, so
  the entry calls `Client.settle` with the outcome, and the client's own `settle` routes a `RETRY` to its `requeue`. The
  entry never calls `requeue` itself on this path.
- **Re-queue entries** — the entry names the retry itself, for a broker with no retry of its own (database, Redis, …).
  The entry still holds the `Job` it dispatched, so on `RETRY` it hands that `Job` to `Client.requeue`. The client
  builds a modified copy via `Job.with*()` (the `Job` is immutable) — `attempts` incremented, `modified_at` stamped —
  and re-enqueues it with the hold from `retry_delay_ms`. **That hold is fixed.** A producer can opt into a linear ramp
  with `retry_delay_multiply_by_attempt`, which is `false` by default; the hold is then `retry_delay_ms × attempts`,
  read from the dispatched `Job` and not from the incremented copy. The producer's original `delay_ms` is not
  re-applied. `ACK` deletes; `FAIL` and `DEAD_LETTER` (the latter when `attempts >= max_attempts`) route to the
  dead-letter destination. Here `attempts` and `modified_at` are envelope-authoritative.
- **The processor** — it redelivers, owning the loop (SQS, AMQP, Beanstalkd, Pub/Sub, Cloud Tasks, …). The entry
  translates the outcome into the processor's native signal (nack/redeliver, return a failure status, extend visibility,
  …) and the processor owns the retry, its backoff, and its counter. `attempts` comes back through the processor's
  header/receive-count, which the entry normalizes into `Job.getAttempts()`; the envelope is not rewritten, so
  `modified_at` is not authored on this path.

Either way the handler and middleware are unchanged — a normalized `Job` in, a `JobResult` out, blind to which
redelivery model the entry chose.

### The pull entry's interface

A pull entry extends `PullQueue` and implements the four steps marked below. `PullQueue` brings the rest.

```
PullQueue
  run(QueueConfig, maxJobs, maxSeconds): void          // inherited: boot, then loop
  loop(Application, maxJobs, maxSeconds): void         // inherited: the loop itself
  connect(Application): void                           // implement: open the connection
  receive(): Job|null                                  // implement: one delivery, or nothing
  settle(Job, JobResult, Client): void                 // implement: abstract on WorkerQueue
  disconnect(): void                                   // implement: graceful shutdown
```

### Push vs. pull — who initiates

The real distinction is **who initiates the delivery**, not what server runs. The kernel is identical for both — a
`Job` in, a `JobResult` out.

- **Pull** — the framework _polls_ the processor (SQS long-poll, AMQP consumer, Redis `BLPOP`, database poll). This is
  just a **long-running loop** — the **`PullQueue`** base that each processor entry extends. It boots the app and
  container **once** and reuses them (child container per job), which _is_ the persistent model already, so there's no
  separate `PullWorkerQueue`. It needs **no server** — a plain process kept alive by the loop, run under a supervisor
  exactly as Laravel's `queue:work` runs (a `while` loop in a plain CLI process, minus the CLI-command wrapper). This
  works in **every language** (Go/Node are built for long-running loops; Java/Python/PHP run one trivially). "Settle"
  acts on the held connection (delete / release-with-delay / dead-letter).
- **Push** — the processor _sends_ an **HTTP request** (Cloud Tasks, Pub/Sub push, SQS→HTTPS, any webhook broker) and
  reads the **response status** as the settlement (2xx = `ACK`/delete, non-2xx = redeliver). It's a _normal_ HTTP
  request; the entry maps its **body** → `Job` (ignoring the headers). Because push needs a web server to _receive_
  those requests, it comes in **CGI** mode (**`PushQueue`** — one job per invocation) or **worker** mode
  (**`PushWorkerQueue`** — a live server receiving pushes). `PushQueue` (CGI) is **also an out-of-the-box default**: it
  uses each language's built-in HTTP handler (Java `com.sun.net.httpserver` exchange, Go `net/http`, Node `http`, Python
  WSGI, PHP CGI/FPM), so no external server is needed. Only `PushWorkerQueue` requires a real worker runtime.
  `getAttempts()` comes from a broker-set retry-count header.

**Worker mode means a _web server_, and only push needs one.** Pull actively polls, so it is inherently persistent and
never needs a server; push receives requests, so it does. That is the whole difference — and it's why `PullQueue`
stands alone while push has both a CGI and a worker form.

**For push-worker, the server runtime and the processor are orthogonal — no server-per-processor explosion, but there
_is_ a per-runtime entry.** The built-in `PushQueue` (CGI) and `PullQueue` (loop) ship as defaults. Only
**`PushWorkerQueue`** is per-web-server-runtime, so each such repo needs its own. gRPC added a worker entry to Tomcat /
Netty / Jetty (Java) and OpenSwoole / FrankenPHP (PHP) for the same reason: the entry maps and dispatches differently
than the HTTP one. The `PullQueue` loop and the CGI `PushQueue` have no such multiplication: each one is built in and
the same everywhere, and a pull entry adds only its own four steps over that one loop.

The **per-processor** logic is _not_ baked into the agnostic entries. **Which entry a bin script boots is the
selection** for a pull processor, so no config property names a processor. A pull entry extends `PullQueue` and
implements the four steps [its interface](#the-pull-entrys-interface) marks, and inherits the loop.

A push processor has no bin script, so the default client of the application is the selection. The config names that
client and the application's service provider wires it, so an application that consumes from one processor sets its
default client to the client of that processor. The processor then supplies a **mapper** (`ServerRequest → Job`) and a
**response mapping** (`JobResult → Response` status), both overridable on `PushQueue`. So the push totals are **M
web-server entries + N mappers and response mappings, never M×N** — otherwise that per-processor logic
would be copy-pasted across every runtime (exchange, Tomcat, Netty, Jetty for Java; CGI, FrankenPHP, RoadRunner,
OpenSwoole for PHP).

For **push**, the mapper takes a _normalized_ Valkyrja **`ServerRequest`**, never a native runtime request — it
**reuses Http's existing runtime→`ServerRequest` mapping** (Tomcat/Netty/OpenSwoole/… already produce one). So the push
mapper is purely `ServerRequest → Job`, the runtime→request work is never re-done, and the push side leans almost
entirely on the reused Http layer.

This is the one wrinkle over Http and gRPC: each of _them_ is a **single** "processor", so one entry serves it; Queue
has many, so each processor gets its own entry class over the shared loop.

**Decoupling:** the Queue core never imports HTTP types. The push entry's mapper is the one place they meet (`Request`
body → `Job`); the dependency is one-way (the push entry depends on HTTP, never the reverse), so a pull-only
deployment loads no HTTP stack.

### Target adapters

Database, Redis, SQS, RabbitMQ/AMQP, Beanstalkd (pull); GCP Cloud Tasks / Pub/Sub push, SQS→HTTPS (push). The in-process
**internal adapters** (`Sync`, `Deferred`, `InMemory`) are produce-side `Client`
adapters, covered under _Producing_ below. Broker-specific config (connection, prefetch, visibility, dead-letter
destination, push endpoint path) lives on the adapter, not in the agnostic contract.

## Producing (enqueuing) — the `Client`

Consuming is the pipeline above; producing is the other half, and it is the **one place the queue has no natural
analog** in the sibling protocols. To _make_ a request elsewhere you reach for a client: Http uses `Http/Client`; Cli
execs a script (or invokes the command class directly); gRPC uses the generated stub. A queue has none of these, so
producing is modeled on the closest fit — **`Http/Client`**.

`Queue/Client` is the producer: a container service with a **per-processor adapter** (one adapter per processor type,
mirroring the consume-side entries). Its only job is to hand a `Job` to the processor.

```
Client
  push(Job): void              // fresh enqueue — stamps id/producer/enqueued_at, attempts = 1
  requeue(Job): void           // settle a RETRY — bumps attempts and derives the hold
  retry(Job, delayMs): void    // re-enqueue an already incremented Job for an explicit hold
  getPushed(): Job[]           // the Jobs handed to this client this lifecycle
  clearPushed(): void          // end the unit of work the record belongs to
```

An internal client carries one more method, because the in-process client is the processor:

```
InternalClient — an abstract base, not a contract (Sync and Deferred extend it)
  settle(Job, JobResult): void // every outcome, not only a RETRY
```

The entry calls `settle` on a client that is an `InternalClient`, and on no other, so `Sync` and `Deferred` each see
every `ACK`, `FAIL` and `DEAD_LETTER` as well as every `RETRY`. `InMemory` extends `Client` and records pushes only.
Only `Sync` acts on a failed outcome, by recording the first `FAIL` or `DEAD_LETTER` and throwing it at the call site.
`Deferred` runs after the response, so it has no call site left to throw at and records nothing. That one call is the
whole settlement: `settle` routes a `RETRY` to its own `requeue`, so the entry never calls `requeue` itself for an
internal client. The `requeue` seam below is what the entry of a re-queue processor calls instead, because such a
processor has no client-side settlement of its own.

`requeue(Job)` is the settlement seam — the entry of a processor with no native redelivery hands it the `Job` **as
dispatched**, and it bumps `attempts` and derives the hold from the ramp of the attempt that just failed. `retry` is the
lower seam it calls, and it takes the **already incremented** copy with the hold supplied. Neither re-stamps `id`, which
stays stable across retries.

A processor that redelivers on its own never calls either one: its entry hands the retry signal to the processor instead
— a nack, a visibility change, a release, or the push response status.

- **Build with the job factory.** The caller builds the `Job` through `JobFactory.create(name, payload)`, where the
  object or array becomes the JSON `payload` via the `Payload` type. Construction sits on a factory rather than on
  `Job`, because a data object holds no static method ([`STATIC_METHODS.md`](STATIC_METHODS.md)). There is deliberately
  **no** `push(name, payload)` convenience on the `Client` either, so the `Client` stays single-purpose — ship a `Job`.
- **Fire-and-forget.** `push` does **not** await a `JobResult` — that is strictly the consume side. It returns nothing
  meaningful; it succeeds once the processor acknowledges the item was enqueued, and throws on an enqueue error. The two
  sides are asymmetric by design: the `Client` publishes (void / enqueue-ack), the entry + `Router` consume (`Job` →
  `JobResult`).
- **The framework stamps the rest.** At `push` the framework sets `id` (VLID V1), `producer` (`AppName lang/version`),
  `enqueued_at`, `modified_at` (= `enqueued_at`), and ensures `attempts` (`1`); the producer supplies only the
  authorable fields (`name`, `payload`, `attributes`, `priority`, `delay_ms`, `max_attempts`, `retry_delay_ms`,
  `retry_delay_multiply_by_attempt` — all already on `Job`, so no options object). The target queue and connection are
  not among them, because addressing is config rather than envelope, as the envelope section says.
- **`getPushed()` records every push, lifecycle-scoped.** The `Client` keeps the (stamped) `Job`s handed to it during
  this unit of work, returned as `Job[]`. One primitive, two payoffs: the test surface (a test asserts over it directly,
  with no fake), and per-request observability. `Sync`, `Deferred` and `InMemory` each hold a buffer of their own, which
  `getPushed` does not read. **The record ends with the unit of work, not with the process.** A client is a container
  singleton, and a long-running host's client outlives every job it runs, so `clearPushed` is what bounds the record.
  `PullQueue.loop` is its only caller in the framework, and it calls it before each job, which leaves the last job's
  record readable once the loop exits. Every other host calls `clearPushed` itself, at the end of whatever that host
  treats as its unit of work. For a `Deferred` client that is the drain, and for a `Sync` client the request it pushed
  from. Nothing else bounds the record, so a long-running host that never calls it accumulates every push it ever made.
- **No middleware on produce.** Producing is a thin service straight over the adapter's publish; the entire middleware
  pipeline runs on **consume**. Cross-cutting `attributes` (trace id, tenant) are stamped as producer-service defaults,
  not via a produce-side middleware stage.

### Internal adapters (no broker)

Three `Client` adapters need no broker, and two of them run jobs. Application code only ever calls `Client.push`. A
swap between one of these adapters and a broker is therefore a config change, with no code change.

`Sync` and `Deferred` hand each job to the `InternalQueue` entry of the application. The entry runs a separate queue
application, the same as a worker that a broker delivers to. The entry drives the same `JobHandler` → `Router` pipeline
as every other job, so no client calls `JobHandler` directly. `InMemory` holds its jobs until a test hands each job to
an entry.

| Adapter    | `push` does (besides record)  | when it runs              |
| ---------- | ----------------------------- | ------------------------- |
| `Sync`     | appends, the outermost drains | **now**, blocking         |
| `Deferred` | buffers it                    | on host **terminate**     |
| `InMemory` | buffers it                    | when a test **drains** it |

- **`Sync`** runs the full pipeline inline and blocks, and it runs the whole retry chain. A push appends the job to a
  buffer, and an outermost push drains that buffer to completion, so a job that pushes another job still runs it before
  the outer push returns. On `RETRY` it appends the `attempts++` `Job` and the same drain runs it again at once, because
  no durable place holds `retry_delay_ms`. The chain ends when the job acknowledges or reaches `max_attempts`. Only the
  timing differs from production, and the retry count is identical.
- **`Deferred`** buffers each job, and a per-host terminate bridge middleware drains the buffer after the response. The
  bridge runs at the Http terminate stage, the Cli after-run stage, or the gRPC `Terminated` stage. To use `Deferred`,
  an application registers the bridge. A deferred job is not durable, because a crash after the response loses the job.
  The drain needs a host that keeps working after the response, such as PHP-FPM with `fastcgi_finish_request`, Swoole,
  RoadRunner, or Node. A host without that capability drains at the end of the request, so the client waits for the
  drain and the caller pays its cost.
- **`InMemory`** — the test adapter. `push` buffers the job, and a test reads the buffer with `getBuffered()` or takes
  it with `drain()`. Distinct from `Sync` (which runs now) — `InMemory` holds the jobs until you process them.

Warning: a `Sync` push throws when the chain ends in `FAIL` or `DEAD_LETTER`. An asynchronous `push` throws only on an
enqueue error. This throw is the one deliberate difference in behavior between the adapters.

**How a job crosses into the queue application.** A client boots the queue application of the `InternalQueue` entry on
the first job, then calls the per-job method of the entry, `handle(app, data, job, client)`. The entry runs each job
in a fresh child container, the same way a worker runs each job that a broker delivers. The client is the one thing
that the two applications share, and that signature is how it crosses: the entry passes it to the settlement step and
never binds it in the queue container. Job code therefore cannot reach the client that pushed the job.

A pull worker takes the other route. `PullQueue.loop` resolves the client from the container of the queue application it
booted, because a worker has no caller to hand one in. The providers of a queue application bring the client services,
and a worker's config implements `QueueClientConfigContract` as well as `QueueConfigContract`, so the container can
resolve the client the worker settles a `RETRY` through.

**How the outcome comes back.** The entry returns nothing, the same as `Http.run` and `Cli.run`. A `RETRY` reaches the
processor three ways, along the split [Redelivery](#redelivery-who-performs-a-retry) draws. The entry of an internal
client calls `settle`, which routes the `RETRY` to `requeue`. The entry of a re-queue processor calls `requeue` itself,
handing it the `Job` **as dispatched**. The entry of a processor that redelivers on its own calls neither. `Sync` runs
the retry at once, `Deferred` re-buffers it so the drain it is already inside runs it, and a re-queue client re-enqueues
it with `retry_delay_ms`. `Sync` also passes itself to the entry as the client, and the entry calls its `settle`, so
every settled outcome reaches it. `Sync` records the first **failed** outcome of the drain, a `FAIL` or a `DEAD_LETTER`,
and throws it at the call site once the buffer empties. An `ACK` records nothing, so an acknowledged job cannot mask a
failure from a job that a later push added.

A test reads each outcome from the per-job result log that [`JobResult`](#jobresult-enum) describes. The pipeline writes
that log, and the client's `settle` does not. The log tells `[Ack]`, `[Fail]`, and `[Retry, Retry, DeadLetter]` apart
without a return value. The client mints the incremented `Job` and records it, so `getPushed` carries each redelivery as
well as each fresh push.

**The entry isolates the process state.** An application sets the base path and the default timezone of the process when
it boots. The `InternalQueue` entry keeps those values of the host intact:

- It restores the values of the host after the boot.
- It applies the values of the queue application before each job.
- It restores the values of the host again after each job.

A job therefore reads the configuration of the queue application, and the host never sees the change. The entry leaves
the exception handler of the host in place, because the host owns the process.

**For posterity — the internal adapters can also be served over the wire.** Nothing stops a `Sync`/`Deferred`/`InMemory`
adapter from being fronted by HTTP like any other processor: the framework simply _becomes the processor_ on the
producing end, with a connection sent over the wire to it. `Sync` would then run the job immediately on receipt (an
added complexity, not a v1 goal). Noted so the option isn't lost.

## Registration

Same discovery → map pattern as the other modules:

- An attribute/annotation/decorator (e.g. `@Route(name, description, handler)`) on handler classes/methods, plus a
  repeatable middleware attribute dispatched to its stage. The attribute carries no retry or attempts policy, because
  the producer decides those and the envelope carries them.
- A collector reflects (or generates) these into `Route`s keyed by job name.
- A job route-provider contract (`getControllerClasses()` + `getRoutes()`) aggregated at boot.

## Application Wiring

Queue wiring mirrors Http, Cli, and gRPC. Each queue application is a separate application, and a host application
reaches it only through a client.

- **`QueueConfigContract` is the config of a queue application.** It holds the default middleware for each stage, and it
  carries its own providers, as every Valkyrja config does. The providers bring the whole queue wiring, which includes
  the routes, the middleware, and the data-cache classes. A host application never holds the config of a queue
  application.
- **Every job runs through a queue entry, in an isolated queue application.** The queue application has its own
  container, and it never shares the container of the host. The entry drives `JobHandler` → `Router` for a job from a
  broker and for a job from an internal client alike. The same routes, middleware, and config therefore apply to every
  job.
  - **The isolation is the point.** A job cannot reach the request-scoped state of the host. That state includes the
    live request, the request singletons, and the container bindings of the host. A development run with an internal
    client therefore behaves the same as a production worker and a test run. A job that uses host state fails in
    development too, so it cannot pass in development and fail in production.
  - **`Queue` boots per job, and `WorkerQueue` boots once.** `Queue.run(config, job)` builds a new application and
    container, handles one job, and exits. It settles nothing, because the signal that settles an outcome belongs to a
    processor, so a caller that needs a `RETRY` settled boots the entry of that processor instead. `Queue` suits a
    one-off dispatch and a test, but a host that pushes often pays a full boot on every push. `WorkerQueue` boots the
    application and container once, and then runs each job in a fresh child container. `PullQueue` and `InternalQueue`
    both extend `WorkerQueue`, so a broker worker and an internal client share one boot model. This split mirrors `Http`
    and `WorkerHttp`.
- **A host application selects a client in its config.** `HttpConfig`, `CliConfig`, and `GrpcConfig` hold no queue
  config. `QueueClientConfigContract` names the default client, and a client with settings of its own has a config
  contract for them. A host holds that client config, and never the config of a queue application. The sync and deferred
  client configs name the `InternalQueue` entry of the application, and that entry returns the queue config. A
  development config selects the sync client, and a production config selects a broker client. The job code stays the
  same, and only the config changes.
- **Same routes, any entry model.** Internal and external consumption share the queue entries and the one
  `RouteCollection`. A `@Route` handler that an application defines once therefore runs the same way from a broker, a
  `sync` push, a `deferred` drain, or an `inmemory` test.
- **Provider wiring.** The middleware, routing, and server providers publish each stage handler as a shared singleton,
  so the `Router` and `JobHandler` register and invoke the same instance. The client provider publishes each client
  config and each client, and it binds `ClientContract` to the default client. `getQueueProviders` exists on
  `ComponentProviderContract`, `ApplicationContract`, the kernel, the child application, and every implementor.

## What differs from CLI and gRPC

- **No synchronous client response.** The outbound value is an **ack/retry/fail decision**, not a payload to a waiting
  caller.
- **Retries, backoff, dead-letter, max-attempts** are first-class — the retry loop is the queue's defining behavior,
  driven by the attempt count carried on the message.
- **At-least-once + idempotency.** Duplicate delivery is expected; the framework exposes attempt count and message id,
  but idempotency is the handler's responsibility.
- **Producing is part of the module** (enqueue side), unlike CLI/gRPC which only consume.
- **Batch consumption and delayed/scheduled jobs** have no analog in the request/response modules.

## Scope of What Is Not Portable

Per-broker and per-language: connection/pool setup, visibility/prefetch/dead-letter configuration, serialization of the
payload, and the poll/subscribe loop. Everything above the entry — job map, middleware composition, container
resolution, outcome mapping, observability — is standardized across all ports.

## Implementation Sequence

1. Finalize this contract document.
2. Prototype in the reference port (PHP) or the most-mature secondary (Java), with a sync/in-memory adapter to prove the
   pipeline and the outcome model end-to-end.
3. Add a real broker adapter (Redis or SQS) to prove settlement, backoff, and dead-lettering.
4. Port to the remaining languages once the shape is settled.
