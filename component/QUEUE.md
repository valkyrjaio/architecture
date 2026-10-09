# Queue

The **cross-language** definition of the Queue component. It states the
hierarchy, the names and the behavior that every port implements.

A port implements this document. A port does not redefine it. The reference
implementation is PHP ([`AGENTS.md`](../AGENTS.md) §1).

This document holds no code example. The component's `README.md` in each port
holds the examples and the per-language spelling.

---

## What the component owns

Queue takes a job and produces a job result. It is a **protocol component**, so
it runs the pipeline in [`LIFECYCLE.md`](../convention/LIFECYCLE.md) and shares
that pipeline's shape with Http, Cli and Grpc.

Queue differs from the other three in one way: nobody waits for the outcome. A
request has a caller to answer. A job is acknowledged, retried or dead-lettered
instead.

**The framework never depends on a specific broker.** The component holds no
broker client. It defines the job, the outcome and the pipeline, and a processor
is bridged in at the edge. "Processor" covers both a message broker and a managed
platform that delivers over HTTP.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Message`    | the job, its payload, its attributes, and the outcome enum |
| `Routing`    | route data, collection, matching, and dispatch             |
| `Middleware` | the stage contracts and the stage handlers                 |
| `Server`     | the job handler that runs the pipeline                     |
| `Client`     | the outbound client, for a job this application publishes  |
| `Throwable`  | the component's throwable contract and its exceptions      |

`Message` holds `Job`, `Payload`, `Attributes` and the outcome `Enum`. The
split mirrors Http: a job is the unit of work, a payload is its body, and
attributes are its headers.

---

## The job envelope is a cross-port contract

A job crosses process boundaries and often crosses ports. A job this application
publishes may be consumed by an application in another language, so the
serialized form is a contract between ports rather than an internal detail.

**A job published by one port runs unchanged on every other port.** That is the
rule the envelope exists to satisfy, and it is why no field may hold a reference
to a class, a type or a function in the publishing language.

| Field                             | Holds                                                   |
| --------------------------------- | ------------------------------------------------------- |
| `id`                              | a producer-generated id, stable across every retry      |
| `name`                            | the routing key, a plain string, never a code reference |
| `producer`                        | provenance, for a trace; nothing branches on it         |
| `attributes`                      | the headers, as a name to values multi-map              |
| `attempts`                        | the delivery count, counting from one                   |
| `max_attempts`                    | the ceiling before the job is dead-lettered             |
| `priority`                        | higher runs sooner, where the processor supports it     |
| `delay_ms`                        | the hold before the job first becomes eligible          |
| `retry_delay_ms`                  | the hold before a retry                                 |
| `retry_delay_multiply_by_attempt` | whether the retry hold ramps with the attempt count     |
| `enqueued_at_ms`                  | when the job was first published                        |
| `modified_at_ms`                  | when the envelope was last rewritten                    |
| `payload`                         | the body, self-contained, with no code reference        |

### Encoding rules

- **A field name is `snake_case` on the wire.** Each port maps it to its own
  casing internally and never changes the wire name.
- **An absolute instant is epoch milliseconds**, in a field suffixed `_ms`, and
  that value is authoritative. A port may write a human-readable twin suffixed
  `_iso` beside it. Code never reads the twin, and the `_ms` value wins on any
  disagreement.
- **A duration is integer milliseconds**, suffixed `_ms`, with no twin.
- **Every field is always present.** There is no omit-when-default. An empty
  multi-map and an empty payload are written as empty rather than dropped, so
  absent and default are never ambiguous.
- **A consumer ignores an unknown field**, and defaults a field an older producer
  did not send. So the envelope gains a field without breaking an older producer.
- **The body is last**, because a payload can be large.

---

## The outcome

What the pipeline hands back is a closed set of four:

| Outcome       | Means                                                  |
| ------------- | ------------------------------------------------------ |
| `ACK`         | the job succeeded, and the processor may forget it     |
| `RETRY`       | deliver the job again, after the retry hold            |
| `FAIL`        | the job failed, and retrying it will not help          |
| `DEAD_LETTER` | the job exhausted its attempts, or must not be retried |

A closed set is what lets one processor-agnostic pipeline drive every processor.
Turning the outcome into a processor's own signal is the whole job of the code at
the edge.

### Mapping a throwable to an outcome

A throwable caught in the pipeline becomes an outcome. The defaults every port
starts from:

- a throwable with no further meaning becomes `RETRY`, unless the attempt count
  reached the ceiling, in which case it becomes `DEAD_LETTER`
- a throwable marked non-retryable becomes `FAIL` at once
- a shutdown becomes `RETRY` with no attempt penalty, because the work never ran

A throwable marked non-retryable carries a marker contract. Warning: catching the
component's own throwable contract does **not** bound the catch to the component,
because the non-retryable marker extends it. An application throwable that
carries the marker is caught too.

---

## Who performs a retry

The processor decides, and the code at the edge hides the difference:

| Model              | The retry is owned by | Because                                    |
| ------------------ | --------------------- | ------------------------------------------ |
| framework re-queue | the framework         | the processor has no native retry          |
| processor-owned    | the processor         | the processor has a native redelivery loop |

Under **framework re-queue**, the framework publishes a modified copy of the job:
the attempt count incremented, the modification time stamped, and the retry hold
applied. The attempt count and the modification time are authoritative on the
envelope. The producer's initial delay is not applied again.

Under **processor-owned** redelivery, the framework translates the outcome into
the processor's native signal and the processor owns the loop, its backoff and
its counter. The attempt count arrives from the processor and is normalized onto
the job. The envelope is not rewritten.

The handler and the middleware are identical under both. They receive a
normalized job and return an outcome, blind to which model is in use.

---

## Who initiates a delivery

This is the real split, and it is not about which server runs:

| Initiator     | Means                                                                   |
| ------------- | ----------------------------------------------------------------------- |
| the framework | the framework polls the processor, in a long-running loop               |
| the processor | the processor sends a request, and reads the response as the settlement |

**Polling needs no server.** It is a plain long-running process, kept alive by a
process manager, and every language runs one. The entry boots the application
once and gives each job its own child container.

**Being sent a job needs a server**, because something must receive the request.
A port ships this on the language's built-in HTTP handling, so no external server
is required, and a persistent runtime gets its own entry.

**Per-processor logic is never duplicated per runtime.** A runtime entry is thin
plumbing that composes two reusable pieces: one that maps an arrival into a job,
and one that settles an outcome back. So the count is the runtimes plus the
processors, never the runtimes multiplied by the processors.

The mapping for a sent job takes the framework's **own** normalized request type,
never a runtime's native request, so it reuses Http's existing runtime mapping
and never repeats that work.

**The dependency runs one way.** Queue never imports Http types. The entry that
receives a sent job is the one place the two meet, so a polling-only deployment
loads no HTTP stack.

---

## Publishing a job

The **client** publishes. It is the produce side, and it lives in the component
beside the consume pipeline, exactly as Http's client does.

A client per processor is cheap, because publishing is serialize-and-send. The
framework also ships clients that never leave the process: one that runs the job
at once, one that defers it to the end of the current unit of work, and one that
holds it in memory for a test.

A publish reports whether the processor accepted the job, and it throws when the
processor refuses it.

---

## Delivery guarantees

**Delivery is at least once.** A processor redelivers on a timeout or a negative
acknowledgement, so the same job can arrive twice. A handler tolerates duplicate
delivery, and the stable job id is what makes that possible.

An ownership window, a visibility timeout or a lease is a **processor concern**
handled at the edge. It is not a field on the job and not a framework contract.

---

## Permitted variation

| Variation                        | Reason                                              |
| -------------------------------- | --------------------------------------------------- |
| which processors a port bridges  | each ecosystem has its own clients available        |
| the runtime entry for a sent job | each language has its own server runtimes           |
| the human-readable time twin     | a port may write it or omit it; code never reads it |

Nothing else varies. The envelope, the outcome set and the pipeline are identical
in every port.
