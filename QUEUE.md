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

The adapter handles broker-specific framing (deletion, visibility extension, backoff, dead-letter routing). The
framework works with decoded envelopes, attributes as structured maps, and the outcome as a value type. Broker specifics
never cross into framework territory.

## Wire Envelope

The **cross-language interop contract**: the one JSON document any port serializes when it enqueues and deserializes
when it consumes. A `Job` published by the PHP port must run unchanged on the Go, Java, TypeScript, or Python port, and
vice versa. This is the one place in the design where the exact bytes matter — treat it as a versioned contract, not an
implementation detail.

It is **HTTP-shaped**, and that mental model governs the whole envelope:

| Envelope     | HTTP analog         | Role                                                       |
|--------------|---------------------|------------------------------------------------------------|
| `name`       | request line (path) | the routing key                                            |
| `attributes` | headers             | cross-cutting metadata a producer stamps on every job      |
| `payload`    | body                | the job-specific data                                      |
| `producer`   | `User-Agent`        | provenance (promoted to a first-class, auto-stamped field) |

Two rules make it portable, and everything else follows from them:

1. **`name` is the only routing key, and it is a plain string.** No class names, no fully-qualified types, no
   language-specific references anywhere in the envelope. It resolves to a handler through each port's own `Router` map.
   It *must* travel in the envelope: the broker hands over an opaque blob with no request line, so the routing key has
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

| Field                             | Type                   | Default             | Description                                                                                                                                                                                           |
|-----------------------------------|------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `id`                              | string                 | generated (VLID V1) | A **VLID V1** (`Type/Vlid`). Producer-generated, **stable across retries** — the dedup/idempotency key and trace-correlation id; also gives DB-backed queues clustered-index locality.                |
| `name`                             | string                 | — (caller-supplied) | Routing key — the `Router` map key, read as `Job.getName()`. Plain string; never a code reference.                                                                                                    |
| `producer`                        | string                 | auto-stamped        | Provenance `AppName lang/version` (AppName from config, `lang` hardcoded per port, `version` from `ApplicationInfo`). Trace-only — no consumer branches on it.                                        |
| `attributes`                      | object (`str → [str]`) | `{}`                | The headers multi-map. Empty = `{}`.                                                                                                                                                                  |
| `attempts`                        | int                    | `1`                 | 1-based delivery count. Framework-incremented on re-queue redelivery; normalized to `Job.getAttempts()` at consume.                                                                                   |
| `max_attempts`                    | int                    | `5`                 | Ceiling before dead-lettering. Producer-set; defaults from `QueueConfig`.                                                                                                                             |
| `priority`                        | int                    | `0`                 | Higher runs sooner where the processor supports it.                                                                                                                                                   |
| `delay_ms`                        | int                    | `0`                 | Initial hold before the job is eligible; `0` = immediate. Producer-authored intent, applied on first enqueue only.                                                                                    |
| `retry_delay_ms`                  | int                    | config default      | Hold before a *retry* re-enqueue. Producer-set; defaults to a non-zero from `QueueConfig` (`0` allowed but BAD — immediate retry). Honored by durable adapters; internal adapters retry immediately. |
| `retry_delay_multiply_by_attempt` | bool                   | `false`             | When `true`, the retry hold is `retry_delay_ms × (attempts − 1)` (linear ramp, self-bounding via `max_attempts`); `false` = fixed. No jitter, no policy object.                                       |
| `enqueued_at_ms`                  | int                    | stamped at enqueue  | Epoch **milliseconds** first enqueued. Authoritative.                                                                                                                                                 |
| `enqueued_at_iso`                 | string                 | stamped at enqueue  | RFC 3339 UTC rendering of `enqueued_at_ms`. Informational only.                                                                                                                                       |
| `modified_at_ms`                  | int                    | `= enqueued_at_ms`  | Epoch **milliseconds** the envelope was last re-written; initialized to the enqueue time, bumped on the re-queue redelivery path. Authoritative.                                                      |
| `modified_at_iso`                 | string                 | `= enqueued_at_iso` | RFC 3339 UTC rendering of `modified_at_ms`. Informational only.                                                                                                                                       |
| `payload`                         | object                 | `{}`                | The body. Self-contained JSON; empty = `{}`, never `null`. No code/type references.                                                                                                                   |

**Every field is always present — on the object and the wire.** There is no omit-when-default:
variability lives only in the *values* (which `attributes` keys exist, what `payload` holds, the numbers and times).
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
"response" with headers, a URL, or a status the way HTTP lets you *build* a `Response`: **all transport is the
entry/adapter's** (delivery, settlement, redelivery, dead-lettering). The envelope is data; the outcome is an enum;
everything in between belongs to the adapter. And because `attributes` is the headers equivalent, it gets a first-class
data class exactly as HTTP headers do (see `Attributes`
under [Core Contracts](#core-contracts)) — not a raw map a handler pokes at.

### What is *not* in the envelope, and why

- **`queue`** — addressing, not body. The consumer is bound to its queue by config and the producer targets it through
  the connection, exactly as you don't name the destination server inside an HTTP request body. (Contrast `name`, which
  must ride inside — there is no request line.)
- **`version`** (a schema discriminator) — without upcaster logic a consumer facing an unknown version can only nack,
  which recovers nothing, and breaking envelope changes are coordinated events anyway. The one useful thing a
  version-like field could give — *who produced this* — is served by `producer`.
- **`payload_type` / any class-string** — a PHP class name is meaningless to a Go consumer. `name` resolves the handler;
  `payload` carries the data. There is no decode hint anywhere — the payload is self-describing JSON.
- **Broker delivery metadata** — the native message id / receive handle and the **visibility-timeout deadline** are
  supplied by the adapter into the `Job` at receipt, not carried on the wire. A consumed
  `Job` = **deserialized envelope + broker delivery metadata**.

### Rejected alternatives (decision log)

- **`available_at` (absolute instant) → `delay_ms` (relative).** `enqueued_at_ms` already anchors a relative delay, so
  the absolute form is redundant; scheduling is expressed as durations (like
  `retry_delay_ms`), and absolute wall-clock scheduling is a *scheduler* concern that enqueues with no delay when it
  fires.
- **Epoch-only or ISO-only timestamps → both.** Epoch for code (unambiguous, no s-vs-ms trap), ISO for humans (readable
  on a dead-letter queue). The extra bytes are meaningless next to broker I/O.
- **Bare-default timestamp (unsuffixed) → always suffixed.** A bare `enqueued_at` integer reintroduces the unit
  ambiguity `_ms` exists to kill; every millisecond value carries `_ms`.
- **`_utc` → `_iso` for the string half.** `_ms` and `_iso` are both *format* labels (same axis); `_utc`
  would name the zone instead, and both fields are UTC anyway.
- **`date_`/`ms_` prefixes → `_ms`/`_iso` suffixes.** Suffixes match the duration convention and keep an instant's two
  views adjacent.
- **Routing key named `queue`, `type`, or `job` → `name`.** `queue` is the ingest point (the server/console analog),
  not the discriminator; the wire key is the `Job`'s `name` (read via `Job.getName()`), which keeps it consistent with
  the single message class.
- **Attributes folded into `payload` → kept separate.** Cross-cutting metadata a producer stamps on every job (tenant,
  trace id, region) is headers, not body; burying it in `payload` forces every handler to dig it out.
- **Retry fields (`attempts`, `max_attempts`, `delay_ms`, `modified_at`) moved to a processor-only header → kept
  first-class on `Job`.** Splitting them out would make the envelope shape *conditional on the processor*, the exact
  thing the cross-processor contract exists to prevent — and framework-requeue processors need them in the body anyway
  (the entry rewrites the whole `Job`). Instead the shape stays uniform and only the **sourcing** varies: the
  entry/adapter reads the value from the wire body (framework-requeue) or from the processor's native counter/headers
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
  Message      // Job, JobResult, Attributes
  Middleware   // the pipeline stage handlers
  Routing      // Route, Router, RouteCollection, the @Route attribute + collector
  Server       // JobHandler + the QueueAdapter (consume) contract
```

`Message` is the analog of `Http/Message` and `gRPC/message` — the category housing the message and its parts, with`Job`
as the class inside (just as `Http/Message` houses `Request`, not a class literally named `Message`).

### Where the concrete adapters and entry points live

Producing and consuming are organized asymmetrically, for the same reason Http is:

- **Producer (`Client`) adapters live *in* the module** (`Queue/Client`) — one lightweight class per processor (`Sync`,
  `Deferred`, `InMemory`, a Guzzle/HTTP push, SQS, Redis, …), exactly like
  `Http/Client`'s adapters. Pushing is cheap (serialize + send) and you push from anywhere, so the framework bundles
  support for any and all external pushes.
- **Consumer *entry points* live in `Application/Entry`** — the bootable classes that select the config and drive
  `JobHandler`. The **internal** ones (`Sync`/`Deferred`/`InMemory` consumption) and the default **`PullQueue`** (a plain
  long-running loop, no server) are an easy map and sit right in `Application/Entry`. Only the **push** entries
  (`PushQueue` / `PushWorkerQueue`) are per-web-server-runtime and live in that server's repo, exactly as the Http and
  gRPC entries do — but they stay **thin**: they *compose* the reusable, per-processor **mapper** and **re-queuer**
  (which live in the Queue module) rather than reimplementing them (see
  [Push vs. pull](#push-vs-pull--who-initiates)).

**Running a consumer is the dev's to wire — the framework ships entries, not a server.** It never ships an HTTP server
or a `queue:work` command; it ships the entries and you point a runtime at the bootstrap, exactly as Http ships CGI +
worker entries and you point nginx/php-fpm or a Swoole/RoadRunner runtime at `index.php`:

- **Pull** (`PullQueue`, the default) is a plain long-running loop run under the dev's process manager (systemd /
  supervisor / a Docker `CMD` / a k8s `Deployment`) — **no server**, works in every language. Graceful shutdown (stop →
  in-flight `→ RETRY`) and an optional bounded lifetime (`--max-jobs`/`--max-time`-style self-exit) let the supervisor
  cycle the process for memory hygiene.
- **Push** (`PushQueue` / `PushWorkerQueue`) rides on your **existing HTTP server** — CGI out of the box, or your worker
  runtime. The processor POSTs; the entry maps the body → `Job`. No queue-specific server (see
  [Push vs. pull](#push-vs-pull--who-initiates)).

The dev hooks a runtime to the bootstrap; the framework owns everything from the entry inward.

## Core Contracts

The language-agnostic surface mirrors Http/Cli/gRPC, with queue vocabulary.

### `JobHandler`

The kernel entry point, analogous to `ServiceHandler` (gRPC) / `RequestHandler` (HTTP). Worker adapters hand messages to
`JobHandler.handle()`.

Responsibilities:

- Orchestrate the middleware stages (`JobReceived`, `Acking`, `Terminated`).
- Delegate to `Router` for resolution and dispatch.
- Run `ThrowableCaught` middleware when exceptions propagate.

As in Http/Cli/gRPC, split the kernel so the **broker settlement** (ack/nack/extend) can happen between the
"acking" stage and "terminated": `handle` (through `ThrowableCaught`) → `acking` (always-run) →
[adapter settles with the broker] → `terminate` (always-run). A `run` convenience bundles
`handle`+`acking`.

### `Router`

Resolves a `Job` to its `Route` via the flat map and dispatches it through the per-route middleware — the same shape as
the Cli/Http/gRPC routers, with `Job`/`JobResult` swapped in. `JobHandler` hands the
`Job` to the `Router`, which figures out how to handle and route it.

```
Router
  dispatch(Job): JobResult              // resolve from the map, then dispatch
  dispatchRoute(Job, Route): JobResult  // dispatch a pre-resolved route
```

A missing map entry routes to `JobNotMatched` (default terminal: `FAIL` → dead-letter) — the analog of gRPC's
`UNIMPLEMENTED`.

### `Job` (immutable)

The **single message class for both directions** — no separate request/response split
(see [One class, produced and consumed](#one-class-produced-and-consumed)). A producer builds a `Job` and dispatches it;
the consumer receives the same `Job`. It is **immutable**, exactly like Http `Request` / Cli `Input`: a `Job` is known
at ingest and never mutated in place — the framework only ever produces a *new* one via `with*` (attempts incremented,
etc.) until, at the very end, an entry decides whether to re-queue it based on the processor. On **produce** the
framework stamps `id`, `producer`, `enqueued_at` (and `attempts` = 1); on **consume** the adapter normalizes `attempts`
from the processor. It is the in-memory form of the [Wire Envelope](#wire-envelope).

```
Job   // immutable — every with* returns a new Job, like Http Request / Cli Input
  getName(): string                         // wire `name` — the Router map key
  getPayload(): ParsedJsonParamCollection   // the JSON body, matching Http's getParsedJson()
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

The "response" — the settlement decision and nothing else. Like Cli's `ExitCode`: a closed set the adapter reads and
acts on, carrying no payload. Not every processor can pass detail back (a push processor answers with an HTTP status),
so a result never carries any.

The contrast with Cli is instructive: Cli returns an `Output` *object* (with an `ExitCode` inside) because output is
legal at **every** lifecycle stage. A queue job's outcome isn't — after a `RETRY`/`FAIL` there is nothing more to do
*to the job itself*, though later stages still receive the `Job`. So the result stays a bare enum; the `Job`, not the
result, carries the detail.

```
JobResult   // ACK | RETRY | FAIL | DEAD_LETTER
```

- **`ACK`** — success; remove from the queue.
- **`RETRY`** — put back for redelivery after the `Job`'s `retry_delay_ms` (× the attempt if
  `retry_delay_multiply_by_attempt` is set). Handler-returned; the framework converts it to
  `DEAD_LETTER` once `attempts` reaches `max_attempts`.
- **`FAIL`** — the handler gives up *on purpose* (non-retryable: bad payload, validation) → dead-letter now, no retries.
  Handler-returned.
- **`DEAD_LETTER`** — the framework exhausted `max_attempts` on a retry chain → dead-letter. Framework-produced, not
  handler-returned; distinct from `FAIL` so the two ways a job dies are told apart.

Failure detail (the throwable, a reason) is logged by `ThrowableCaught` when it happens. Distinguishing the four
outcomes *after the fact* is a **testing concern only** — in production the outcome just drives settlement directly
(re-enqueue / throw / record) and is never read back. For tests, a **fixture** over the Queue/WorkerQueue keeps an
in-memory `Job.id → [JobResult…]` map so a job's whole life reads back as
`[Ack]`, `[Fail]`, or `[Retry, Retry, DeadLetter]`. That fixture exists **specifically to test the middleware, `Client`
s, and entry classes** — it is not a production mechanism, and it is separate from settlement (the re-enqueue via the
seam).

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
visibility-timeout / message-ownership window is a **broker concern the adapter manages** — it is not a framework
contract on the `Job`.)

## Middleware Pipeline

```
1. JobReceived      always runs; pre-router
2. Router resolves job from map
3a. JobMatched          runs if job found; pre-handler
    User handler runs, produces JobResult
3b. JobDispatched       runs if job was found; post-handler
 OR
3c. JobNotMatched       runs if job not found
    Default terminal produces JobResult::fail() (unknown job → dead-letter)

[if any above threw]
4. ThrowableCaught      converts throwable → JobResult (default: RETRY within maxAttempts, else DEAD_LETTER)

5. Acking               always runs (including error paths)
   Adapter settles (delete / release + retry_delay / dead-letter)
6. Terminated           runs after settlement (metrics, events, cleanup)
```

All stages except `JobReceived` and `Acking` are optional; `Acking` and `Terminated` always run.

### Exception → outcome mapping

`ThrowableCaught` translates exceptions to a `JobResult`. Sensible defaults (configurable per application and
overridable per route):

- A **retryable** exception (or any uncaught throwable) → `RETRY` (re-enqueued after `retry_delay_ms`), **unless**
  `attempts >= max_attempts`, in which case → `DEAD_LETTER`.
- A **non-retryable** exception (bad message, validation) → `FAIL` immediately.
- A worker shutdown → `RETRY` with no penalty (the message returns for another worker), since the work was not
  completed.

## Worker Adapters

The adapter is the queue protocol's **entry module** — the direct analog of the entry classes in the other protocols:
Http's server + `RequestHandler`, Cli's console + `InputHandler`, gRPC's
`ServiceAdapter` + `ServiceHandler`. It owns both ends of a delivery and nothing in between: the **entry** (accept a
native delivery from the processor and normalize it into a `Job`) and the **response** (take the `JobResult` the kernel
returns and settle it back with the processor). Routing, middleware, and the handler are all processor-agnostic; only
the adapter knows what a Cloud Tasks POST or an SQS receipt looks like. "Processor" is the umbrella term here — a
message broker (SQS, AMQP, Redis) or a managed platform (Cloud Tasks, Lambda, Pub/Sub push).

The entry and exit stay **clean** — plain `Job` in, `JobResult` out — up to the point where they must be mapped onto a
specific processor's runtime (e.g. OpenSwoole in PHP); that translation is the adapter's only real work, and it is the
same idea on both sides of a delivery.

Adapters bridge an external processor to `JobHandler`. Responsibilities:

1. Poll/subscribe for messages from the broker (long-poll, blocking pop, push subscription, …).
2. Decode the message; build a `Job` (name, payload, attributes, id, attempts).
3. Invoke `JobHandler.handle(job)` (via the worker base `dispatch`).
4. **Settle** with the processor based on the `JobResult` outcome (see [The outcome is an enum](#the-outcome-is-an-enum)
   and [Redelivery](#redelivery-re-queue-vs-processor-owned)). It slots between `acking` and `terminate` via the worker
   base's settlement callback.

Adapters may consume in **batches** and dispatch each message independently (each in its own child container), settling
per message.

### The outcome is an enum

What the kernel hands back for settlement is a small, closed set — a `JobResult`: `ACK | RETRY | FAIL | DEAD_LETTER`,
exactly like Cli's `ExitCode`. The adapter reads the enum and acts, nothing more. That closed outcome is what lets one
processor-agnostic kernel drive every processor — turning `ACK`/`RETRY`/`FAIL` into processor-specific action is the
adapter's whole job on the response side.

### Redelivery: re-queue vs. processor-owned

Who actually performs a `RETRY` depends on the processor, and the adapter encapsulates the difference:

- **Re-queue adapters** — framework-owned redelivery, for processors with no native retry (database, Redis, …). The
  adapter still holds the `Job` it dispatched, so on `RETRY` it builds a modified copy via
  `Job.with*()` (the `Job` is immutable) — `attempts` incremented, `modified_at` stamped — and re-enqueues it with the
  hold from `retry_delay_ms` (× the attempt if `retry_delay_multiply_by_attempt`
  is set; the producer's original `delay_ms` is not re-applied). `ACK` deletes; `FAIL` and `DEAD_LETTER`
  (the latter when `attempts >= max_attempts`) route to the dead-letter destination. Here `attempts` and
  `modified_at` are envelope-authoritative.
- **Processor-owned adapters** — native redelivery, where the processor owns the loop (SQS, AMQP, Cloud Tasks, Pub/Sub
  push). The adapter translates the outcome into the processor's native signal (nack/redeliver, return a failure status,
  extend visibility, …) and the processor owns the retry, its backoff, and its counter. `attempts` comes back through
  the processor's header/receive-count, which the adapter normalizes into `Job.getAttempts()`; the envelope is not
  rewritten, so
  `modified_at` is not authored on this path.

Either way the handler and middleware are unchanged — a normalized `Job` in, a `JobResult`
out, blind to which redelivery model the adapter chose.

### Adapter interface

```
QueueAdapter
  start(JobHandler): void   // begin consuming (connect, subscribe, poll loop)
  stop(): void                  // graceful shutdown (stop polling, drain in-flight)
```

### Push vs. pull — who initiates

The real distinction is **who initiates the delivery**, not what server runs. The kernel is identical for both — a
`Job` in, a `JobResult` out.

- **Pull** — the framework *polls* the processor (SQS long-poll, AMQP consumer, Redis `BLPOP`, database poll). This is
  just a **long-running loop** — **`PullQueue`**, the out-of-the-box default. It boots the app + container **once** and
  reuses them (child container per job), which *is* the persistent model already, so there's no separate
  `PullWorkerQueue`. It needs **no server** — a plain process kept alive by the loop, run under a supervisor exactly as
  Laravel's `queue:work` runs (a `while` loop in a plain CLI process, minus the CLI-command wrapper). This works in
  **every language** (Go/Node are built for long-running loops; Java/Python/PHP run one trivially). "Settle" acts on the
  held connection (delete / release-with-delay / dead-letter).
- **Push** — the processor *sends* an **HTTP request** (Cloud Tasks, Pub/Sub push, SQS→HTTPS, any webhook broker) and
  reads the **response status** as the settlement (2xx = `ACK`/delete, non-2xx = redeliver). It's a *normal* HTTP
  request; the entry maps its **body** → `Job` (ignoring the headers). Because push needs a web server to *receive*
  those requests, it comes in **CGI** mode (**`PushQueue`** — one job per invocation) or **worker** mode
  (**`PushWorkerQueue`** — a live server receiving pushes), on the *same* server you already run for HTTP.
  `getAttempts()` comes from a broker-set retry-count header.

**Worker mode means a *web server*, and only push needs one.** Pull actively polls, so it is inherently persistent and
never needs a server; push receives requests, so it does. That is the whole difference — and it's why `PullQueue`
stands alone while push has both a CGI and a worker form.

**For push, the server runtime and the processor are orthogonal — no server-per-processor explosion, but there *is* a
per-runtime entry.** You reuse the *runtime*, never the *entry*: each web-server-runtime repo needs its own `PushQueue`
/ `PushWorkerQueue`, exactly as gRPC added CGI/worker entries to Tomcat / Netty / Jetty (Java) and OpenSwoole /
FrankenPHP (PHP) — because the entry maps and dispatches differently than the HTTP one. `PullQueue` has no such
multiplication: it is a single plain loop, the same everywhere.

The **per-processor** logic is *not* baked into those entries — it is extracted into **reusable, server-agnostic
classes** selected by `QueueConfig.processor`: a **mapper** (push: `ServerRequest → Job`; pull: the connect-and-poll
client) and a **re-queuer** (the settlement / re-enqueue — `JobResult → Response` status for push, broker re-enqueue for
pull). A push entry is thin runtime plumbing that *composes* them; it never reimplements mapping or re-queueing. So the
push totals are **M web-server entries + N mappers + N re-queuers, never M×N** — otherwise that per-processor logic
would be copy-pasted across every runtime (exchange, Tomcat, Netty, Jetty for Java; CGI, FrankenPHP, RoadRunner,
OpenSwoole for PHP).

For **push**, the mapper takes a *normalized* Valkyrja **`ServerRequest`**, never a native runtime request — it
**reuses Http's existing runtime→`ServerRequest` mapping** (Tomcat/Netty/OpenSwoole/… already produce one). So the push
mapper is purely `ServerRequest → Job`, the runtime→request work is never re-done, and the push side leans almost
entirely on the reused Http layer.

This is the one wrinkle over Http and gRPC: each of *them* is a **single** "processor", so their entry never switches;
Queue has many, so the entry maps on `QueueConfig.processor`.

**Decoupling:** the Queue core never imports HTTP types. The push entry's mapper is the one place they meet (`Request`
body → `Job`); the dependency is one-way (the push adapter depends on HTTP, never the reverse), so a pull-only
deployment loads no HTTP stack.

### Target adapters

Database, Redis, SQS, RabbitMQ/AMQP, Beanstalkd (pull); GCP Cloud Tasks / Pub/Sub push, SQS→HTTPS (push). The in-process
**internal adapters** (`Sync`, `Deferred`, `InMemory`) are produce-side `Client`
adapters, covered under *Producing* below. Broker-specific config (connection, prefetch, visibility, dead-letter
destination, push endpoint path) lives on the adapter, not in the agnostic contract.

## Producing (enqueuing) — the `Client`

Consuming is the pipeline above; producing is the other half, and it is the **one place the queue has no natural
analog** in the sibling protocols. To *make* a request elsewhere you reach for a client: Http uses `Http/Client`; Cli
execs a script (or invokes the command class directly); gRPC uses the generated stub. A queue has none of these, so
producing is modeled on the closest fit — **`Http/Client`**.

`Queue/Client` is the producer: a container service with a **per-processor adapter** (one adapter per processor type,
mirroring the consume-side entry adapters). Its only job is to hand a `Job` to the processor.

```
Client
  push(Job): void          // fresh enqueue — stamps id/producer/enqueued_at, attempts = 1
  retry(Job): void         // re-enqueue an existing Job for retry (id preserved, attempts already ++)
  getPushed(): Job[]        // the Jobs handed to this client this lifecycle
```

`retry(Job)` is the settlement seam — the consumer hands it the `attempts++` `Job` on a `RETRY` outcome. Its behavior is
the per-processor **re-queuer**: for a **framework-owned** processor it is essentially `push()` of the updated `Job`;
for a **processor-owned** one it hands the retry signal (a retry-count header, nack, visibility extension, or the push
response status) to the processor instead. Unlike `push`, it does **not** re-stamp `id` (stable across retries) or reset
`attempts`.

- **Build with `Job::create`.** The caller builds the `Job` — `Job::create(name, payload)`, where the object or array
  becomes the JSON `payload` via the `Payload` type. There is deliberately **no**
  `push(name, payload)` convenience on the `Client`: ergonomic construction lives on `Job`, so the
  `Client` stays single-purpose — ship a `Job`.
- **Fire-and-forget.** `push` does **not** await a `JobResult` — that is strictly the consume side. It returns nothing
  meaningful; it succeeds once the processor acknowledges the item was enqueued, and throws on an enqueue error. The two
  sides are asymmetric by design: the `Client` publishes (void / enqueue-ack), the entry + `Router` consume (`Job` →
  `JobResult`).
- **The framework stamps the rest.** At `push` the framework sets `id` (VLID V1), `producer`
  (`AppName lang/version`), `enqueued_at`, `modified_at` (= `enqueued_at`), and ensures `attempts` (`1`); the producer
  supplies only the authorable fields (`name`, `payload`, `attributes`, `priority`,
  `delay`, `max_attempts`, target queue/connection — all already on `Job`, so no options object).
- **`getPushed()` records every push, lifecycle-scoped.** The `Client` keeps the (stamped) `Job`s handed to it during
  this unit of work, returned as `Job[]`. One primitive, three payoffs: it is the
  `Deferred` adapter's buffer (drained on terminate), the test surface (`assertPushed` with no fake), and per-request
  observability. **It must be scoped to the request/command/rpc** (resolved from the child container, discarded each
  cycle) — a process-global record would leak in a long-running server and bleed one request's `Deferred` jobs into the
  next.
- **No middleware on produce.** Producing is a thin service straight over the adapter's publish; the entire middleware
  pipeline runs on **consume**. Cross-cutting `attributes` (trace id, tenant) are stamped as producer-service defaults,
  not via a produce-side middleware stage.

### Internal adapters (no broker)

Three `Client` adapters run jobs **in-process**, no broker required — produce and consume fuse in one process. All obey
the invariant that **app code only ever calls `Client.push`**; only these adapters reach the **Queue entry point** —
`Queue.run(config, job)`, which builds the child container from the
`QueueConfig` and then drives the same `JobHandler` → `Router` every real adapter uses (never
`JobHandler` directly). Swapping between them and a real broker is a **config change, zero code change**
— the caller cannot tell where a job ran.

| Adapter    | `push` does (besides record)  | when it runs              |
|------------|-------------------------------|---------------------------|
| `Sync`     | runs it inline                | **now**, blocking         |
| `Deferred` | buffers it (into `getPushed`) | on host **terminate**     |
| `InMemory` | buffers it                    | when a test **drains** it |

- **`Sync`** — the zero-config default. `push` runs the full pipeline inline and blocks, and it **runs the job to
  completion, retries and all**: on `RETRY` it re-runs the `attempts++` `Job` **immediately**
  (there's no durable place to hold `retry_delay_ms`, so the delay is skipped) until it `ACK`s or hits
  `max_attempts`, at which point the terminal `FAIL`/`DEAD_LETTER` **surfaces at the call site as a throw**. So a `Sync`
  `push` *can* throw on a job's ultimate failure, unlike an async `push`, which throws only on an *enqueue* error — the
  one deliberate behavioral difference. Only the *timing* differs from prod (immediate vs. `retry_delay_ms`); the retry
  *count* is identical.
- **`Deferred`** — the latency upgrade (Laravel's `dispatchAfterResponse`). `push` only buffers; a thin **per-host
  terminate bridge middleware** (Http terminate / Cli after-run / gRPC `Terminated`) drains
  `getPushed()` → the Queue entry point (`Queue.run`) after the response. Opt-in: register the bridge to use it, else
  fall back to `Sync`. Two caveats: **not durable** (in-process; a crash after the response loses the jobs), and
  **runtime-dependent** (true "after the client has the response" needs the host to finish the response then keep
  working — PHP-FPM `fastcgi_finish_request`, Swoole/RoadRunner, Node; where unavailable it degrades to "batched at end
  of request, client still waits").
- **`InMemory`** — the test adapter. `push` records; a test drains/asserts over `getPushed()`. Distinct from `Sync`
  (which runs now) — `InMemory` holds the jobs until you process them.

**The consume mechanics (how these retry across the isolation boundary).** The entry is
`run(config, job, client): void` — it returns nothing, exactly like `Http`/`Cli`/`Grpc.run` (their output is already
emitted by the time `run` returns). The isolated consumer runs the pipeline, and on
`RETRY` it mints the `attempts++` `Job` (immutable `with*`) and calls the injected **`Client`**'s `retry(job)` — the
`Client` is the *single* thing shared across the isolation boundary. The job **handler** never sees the `Client`
(it's a `run` parameter, framework plumbing, not in the isolated container), so job code stays isolated; only the
framework's settlement uses it. `Sync` loops those re-runs immediately; `InMemory` re-buffers for the test to re-drain;
a real/broker adapter re-enqueues with `retry_delay_ms`. The **outcome** is never returned — it's read off the per-job
result log (`Job.id → [JobResult…]`), which is exactly why
`[Ack]`, `[Fail]`, and `[Retry, Retry, DeadLetter]` are all distinguishable in a test without a return value. (This is
why the producer can't reconstruct the retry `Job` from `getPushed` — the incremented
`Job` is minted *inside* the consumer; it must ride out via the injected `Client`.)

**For posterity — the internal adapters can also be served over the wire.** Nothing stops a `Sync`/`Deferred`/`InMemory`
adapter from being fronted by HTTP like any other processor: the framework simply *becomes the processor* on the
producing end, with a connection sent over the wire to it. `Sync` would then run the job immediately on receipt (an
added complexity, not a v1 goal). Noted so the option isn't lost.

## Registration

Same discovery → map pattern as the other modules:

- An attribute/annotation/decorator (e.g. `@Route(name, queue, maxAttempts, retryDelayMs,
  retryDelayMultiplyByAttempt)`) on handler classes/methods, plus a repeatable middleware attribute dispatched to its
  stage.
- A collector reflects (or generates) these into `Route`s keyed by job name.
- A job route-provider contract (`getControllerClasses()` + `getRoutes()`) aggregated at boot.

## Application Wiring

Mirror Http/Cli/gRPC (none is unique — they are all the same concept), with one queue-specific addition — embedding the
queue into a host app.

- **`QueueConfig` is a consume-side config.** Connections/queues, a **`processor`** property (which processor an entry
  maps for — the one wrinkle Http/gRPC don't have, since they're single-processor), default per-stage middleware, and
  worker options (prefetch, max-attempts, retry-delay defaults). It carries its own providers (as every Valkyrja config
  does), so handing it over brings the whole queue wiring — routes, middleware, data-cache classes. The produce side
  only *borrows* it, through the internal adapters.
- **`Queue.run(config, job)` is the one consume entry — and it runs the job in an isolated Queue application +
  container**, its own instance (a "process within the process"), never the host's. Both external delivery and internal
  `push` funnel through it; it drives `JobHandler` → `Router`. Internal adapters and the `Deferred` bridge call
  **this**, never `JobHandler` directly, so the same routes/middleware/config apply no matter how a job arrived.
    - **The isolation is the point, not a side effect.** A job cannot reach the host's request-scoped state (the live
      request/response, request singletons, host container bindings), so an embedded-dev run behaves **identically** to
      a standalone-prod worker — and to a test run. The "works in dev, breaks in prod" class of bug (a job accidentally
      leaning on shared host state) simply cannot occur. This is the dividend of routing through the entry rather than
      `JobHandler`, which would have shared the host container.
    - **`Queue` (single-shot) vs. `WorkerQueue` (boot-once) — and why it matters for cost.**
      `Queue.run(config, job)` is **single-shot**: it makes a new application + container, handles that one job, and
      **exits** — nothing persists for a next job, because it isn't running as a server. Right for one-off dispatch and
      tests, but a host pushing repeatedly through it pays a full app + container boot *per push*. To amortize, use **
      `WorkerQueue`**: it boots the application + container **once**, then takes jobs one at a time via a dedicated
      method (the same shape a real broker worker loops over), each in a fresh **child container**, the adapter settling
      via the callback. So "bootstrap once, child container per job" is a property of `WorkerQueue`, not something
      `Queue` does on its own — a repeatedly-pushing internal adapter bootstraps a `WorkerQueue` once per host lifecycle
      and feeds each pushed job to it. Mirrors Http's single-shot handler vs. `WorkerHttp`.
- **Embedding is opt-in, via a contract on the host config.** `HttpConfig` / `CliConfig` / `GrpcConfig`
  optionally implement a `QueueConfigProvidedContract` (`getQueueConfig(): QueueConfig`). Present → that host app can
  run jobs in-process (`sync`/`deferred`/`inmemory`) against that config, and its entry point selects it, so a whole app
  (Http + its Queue) lives in one config. Absent → no embedding; use external processors or a dedicated worker app.
  Because it's a config-level choice, it's naturally **per-environment**: a dev config wires the contract (embed the
  queue, run `sync`/`deferred` — no broker infra), while a prod config omits it and points at an external processor —
  same job code, environment swapped by config alone. **Base `HttpConfig` has zero knowledge of Queue** — opt-in
  coupling only, so the modules stay independent by default (the same property that keeps Http and Cli split).
- **Same routes, any entry model.** Because internal and external consumption share `Queue.run` and the one
  `RouteCollection`, a `@Route` handler defined once runs identically via an external broker, an in-app `sync` push, a
  `deferred` drain, or an `inmemory` test. Define once, run anywhere.
- **Provider wiring.** Middleware/routing/server provider pairs — stage handlers published as **shared singletons** so
  the `Router` and `JobHandler` register/invoke the same instances; `getQueueProviders`
  added across `ComponentProviderContract`, `ApplicationContract`, the kernel, the child application, and every
  implementor.

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
payload, and the poll/subscribe loop. Everything above the adapter — job map, middleware composition, container
resolution, outcome mapping, observability — is standardized across all ports.

## Implementation Sequence

1. Finalize this contract document.
2. Prototype in the reference port (PHP) or the most-mature secondary (Java), with a sync/in-memory adapter to prove the
   pipeline and the outcome model end-to-end.
3. Add a real broker adapter (Redis or SQS) to prove settlement, backoff, and dead-lettering.
4. Port to the remaining languages once the shape is settled.
