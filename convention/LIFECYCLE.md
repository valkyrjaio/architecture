# Application Lifecycle

How a Valkyrja application handles one unit of work: the stages it passes
through, their order, what is guaranteed at each one, and the names every
language port shares.

This is the definition the ports are built from. Each port's own `LIFECYCLE.md`
carries the same content with that language's entry points, bootstrap sequence
and examples added.

It holds no code example of its own, because an example has to pick one language.

---

## One pipeline, four protocols

A protocol is a way work arrives. Valkyrja has four, and each one runs the same
pipeline over a different unit of work.

| Protocol | Unit of work | Outcome            |
| -------- | ------------ | ------------------ |
| Http     | a request    | a response         |
| Grpc     | a call       | a service response |
| Queue    | a job        | a job result       |
| Cli      | an input     | an output          |

The pipeline is identical in the middle and differs only at the two edges. The
entry stage names the arrival, and the exit stages name the irreversible act
that ends the work.

---

## The stages

| #   | Http              | Grpc              | Queue             | Cli               | Runs                   |
| --- | ----------------- | ----------------- | ----------------- | ----------------- | ---------------------- |
| 1   | `RequestReceived` | `CallReceived`    | `JobReceived`     | `InputReceived`   | always                 |
| 2   | `RouteMatched`    | `RouteMatched`    | `RouteMatched`    | `RouteMatched`    | a route was found      |
| 3   | `RouteNotMatched` | `RouteNotMatched` | `RouteNotMatched` | `RouteNotMatched` | no route was found     |
| 4   | `RouteDispatched` | `RouteDispatched` | `RouteDispatched` | `RouteDispatched` | a route was found      |
| 5   | `ThrowableCaught` | `ThrowableCaught` | `ThrowableCaught` | `ThrowableCaught` | a throwable was caught |
| 6   | `SendingResponse` | `SendingResponse` | `SettlingResult`  | —                 | always                 |
| 7   | `ResponseSent`    | `ResponseSent`    | `ResultSettled`   | `ProcessExiting`  | always                 |

Stages 2 to 5 are the same four names in every protocol. Only stage 1, stage 6
and stage 7 carry a protocol-specific name.

**Cli has no stage 6.** A command writes output at any point in the pipeline, so
there is no single moment before the write for a stage to occupy. Every other
protocol produces its outcome once, at one place, so stage 6 has a moment to
hold.

**Cli has a stage 7.** `ProcessExiting` runs after the output is written and
before the process ends. It is the terminal stage, and it holds the same slot
that `ResponseSent` and `ResultSettled` hold.

---

## What the stage names mean

The name states when the stage runs. A reader who knows the rule can place a new
stage without being told where it goes.

1. **An object and a past participle.** `RequestReceived`, `RouteMatched`,
   `ThrowableCaught`. The object says what the stage is about, and the participle
   says the thing already happened. This is the dominant form, and stages 1 to 5
   all take it.
2. **Stage 6 is a gerund.** `SendingResponse`, `SettlingResult`. The gerund says
   the irreversible act is about to happen. Code at this stage can still change
   the outcome.
3. **Stage 7 is a past participle.** `ResponseSent`, `ResultSettled`. The
   participle says the act happened. Code at this stage cannot change the
   outcome.
4. **Stage 6 and stage 7 are a matched pair over one object.** `SendingResponse`
   pairs with `ResponseSent`, and `SettlingResult` pairs with `ResultSettled`.

`ProcessExiting` is the one gerund in a stage-7 slot, and the rule explains it.
The irreversible act the name precedes is the process exit, not the write. The
write already happened, so nothing at that stage can change the output.

A stage name is shared vocabulary. Every port uses the same one, so the name
means the same thing wherever it is read.

---

## What each stage guarantees

A guarantee is what code at that stage may rely on, and what it may change.

| Stage | May rely on                                      | May change                                |
| ----- | ------------------------------------------------ | ----------------------------------------- |
| 1     | the unit of work, nothing routed yet             | the unit of work, or the outcome directly |
| 2     | a matched route                                  | the route, or the outcome directly        |
| 3     | no route matched, and a default outcome          | the outcome                               |
| 4     | the handler ran, and produced an outcome         | the outcome                               |
| 5     | a throwable, and a default outcome built from it | the outcome                               |
| 6     | a final outcome, not yet acted on                | the outcome                               |
| 7     | the act completed                                | nothing                                   |

Stage 1 and stage 2 can **short-circuit**: code there returns an outcome instead
of the unit of work, and every remaining stage before 6 is skipped. Stage 6 and
stage 7 still run, because they always run.

Stage 7 is where work that the caller never waits for belongs. The outcome is
already delivered, so a log write, a cache write, or a dispatched event costs the
caller nothing.

---

## What each stage is for

The guarantee decides what belongs at a stage. These are the uses each stage
exists to serve, and they are the same in every protocol.

**Stage 1 — the arrival.** Nothing is routed yet, so this is the only stage that
sees every unit of work, including one that matches no route. It holds the
concerns that apply to all of them: a maintenance-mode check, rate limiting, and
a full-outcome cache lookup. Code here can answer immediately and skip the rest.

**Stage 2 — a route matched.** The route is known and the handler has not run, so
this is where a decision about **this** route belongs: authentication,
authorization, resolving a tenant, validating input. Code here can answer
immediately instead of letting the handler run.

**Stage 3 — no route matched.** A default outcome already exists, so this stage
exists to replace it: a custom not-found page, or a fallback that handles the
unmatched work itself.

**Stage 4 — the handler ran.** The outcome exists and is not yet final. This is
where the outcome is transformed: adding a header, reshaping a body, recording
what the handler produced.

**Stage 5 — a throwable was caught.** The stage receives the throwable and a
default outcome built from it, so it holds error reporting and the
application's own error presentation. Returning an outcome here is what turns a
failure into a reply.

**Stage 6 — the act is about to happen.** The outcome is final and nothing has
been committed, so this is the last chance to change it: compression, a header
that depends on the finished body, a cache-control decision.

**Stage 7 — the act happened.** Nothing can change the outcome. This is where
deferred side effects belong, and it is the stage that makes a slow side effect
free to the caller.

Warning: stage 6 and stage 7 **always run**, including after a short-circuit at
stage 1 and after a throwable at stage 5. Code at either one cannot assume a
route was matched or a handler ran.

---

## The names each stage generates

One stage name produces three identifiers, and a port derives all three from it
rather than inventing any of them.

| Identifier              | Shape                       | Holds                                               |
| ----------------------- | --------------------------- | --------------------------------------------------- |
| the middleware contract | `<Stage>MiddlewareContract` | what an application implements to run at that stage |
| the handler contract    | `<Stage>HandlerContract`    | what the framework calls to dispatch that stage     |
| the handler method      | `<stage>`                   | the one method on that handler contract             |

The handler method carries the stage name, so one class can implement several
stages without a method collision. A handler contract declares exactly one
method, and the shared `HandlerContract` declares `add` for registering
middleware.

A port spells the method in its own convention and changes nothing else. The
identifier is the vocabulary; the casing is syntax.

### Middleware is appended, never deduplicated

Middleware registered at a stage runs in registration order, and the framework
**never** removes a repeat. Scheduling the same class twice at one stage runs it
twice, including when the two registrations spell the name differently.

A duplicate is the application's own defect, and the framework does not hide it.
This also binds the generated form: data the build tool writes mirrors what the
run-time walk produces, duplicate included, so the cached path and the uncached
path behave identically ([`DATA_CACHE.md`](DATA_CACHE.md)).

---

## The server handler

Each protocol has one class that owns the pipeline, named for the unit of work it
takes: `RequestHandler`, `ServiceHandler`, `JobHandler`, `InputHandler`. Each
declares `run` as the entry point a runtime calls, and `handle` as the method
that produces the outcome from stages 1 to 5.

The terminal methods are not yet consistent. `JobHandler` names them for their
stages, `settlingResult` and `resultSettled`, which is the convention above. The
other three carry legacy names — `send` and `terminate` on `RequestHandler`,
`sending` and `terminate` on `ServiceHandler`, and `exit` on `InputHandler`. A
port mirrors the handler it ports today. The rename is a cross-language change,
and it is tracked rather than done one port at a time.

---

## The routing stages are a router concern

Stages 2 to 5 exist because routing exists. A protocol resolves a route from the
unit of work, dispatches the route's handler, and reports a failure to match. A
component with no routing has no pipeline, so it has no stages: the Event
component dispatches a listener directly and takes no middleware.

---

## What belongs in a port's own document

This document stops at the shape. A port's `LIFECYCLE.md` holds:

- the entry point classes, and the runtime each one serves
- the bootstrap sequence, and how the port loads its components
- the worker model for a persistent runtime, and how it isolates state
- every code example

A port's document restates the shape above and adds its examples, so the two read
alike. What a port must **not** do is change a statement while restating it: this
file is where a stage, an order or a guarantee is decided, and a port that
disagrees here is reporting its own defect.
