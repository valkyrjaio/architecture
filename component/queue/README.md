# Queue

Implements
[`QUEUE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/queue/QUEUE.md).

## Introduction

The Queue component runs work outside the unit of work that asked for it. One
part of the application publishes a **job**, and a consumer runs that job later,
in its own process.

A job is a name and a body. The name routes the job to a handler, exactly as a
path routes to a controller, and the body carries what the handler reads. The
component takes a job in and hands an **outcome** back.

Nobody waits for the outcome. A job is acknowledged, retried or dead-lettered
instead of answered, so a caller gets no reply and the handler reports what
happened to the processor rather than to a person.

The framework depends on no specific processor. The component defines the job,
the outcome and the pipeline, and a **client** publishes to one processor.
"Processor" covers a message broker and a managed platform that delivers over
HTTP. A client's own library is the application's dependency rather than the
framework's.

This document assumes the framework, where the container resolves the client and
a provider builds it.

## Declaring a job

A job route is a name, a description and a handler. Declare it on the method
that handles it, where the language has a marker for that, and the collector
reads each class a route provider lists.

```text
@Name("orders")
class OrderJobController {

    @Route(name: "charge-card", description: "Charge the card for an order")
    @RouteHandler([self, "chargeCard"])
    static chargeCard(container, route) {
        job = container.getSingleton(JobContract)

        container.getSingleton(PaymentGateway).charge(job.getPayload().get("orderId"))

        return JobResult.ACK
    }
}
```

| Part            | Is                                               |
| --------------- | ------------------------------------------------ |
| the name        | the routing key, which is the job's `name` field |
| the description | what the job does, for a reader                  |
| the handler     | what to call, as a named reference               |
| the middleware  | what runs around this job, per stage             |

**The handler is named rather than taken from the method the marker sits on.**
The marker records what to register, and `@RouteHandler` names the reference the
route carries. A route that names none runs nothing and acknowledges the job.

**The name is a plain string.** It is the key a consumer matches, so renaming a
handler class does not strand a job already in the queue.

A `@Name` on the class prefixes every route the class declares, so one class
holds the jobs of one area.

### Declaring routes in a provider

A route provider returns the classes to scan and any prebuilt route, and the
application lists that provider in its config
([`PROVIDERS.md`](https://github.com/valkyrjaio/architecture/blob/26.x/convention/PROVIDERS.md)).
The component discovers nothing on its own.

```text
class AppQueueRouteProvider implements QueueRouteProviderContract {

    getControllerClasses() {
        return [OrderJobController]
    }

    getRoutes() {
        return [
            new Route(
                name: "orders.email-receipt",
                description: "Email the receipt for an order",
                handler: [ReceiptMailer, "send"],
            ),
        ]
    }
}
```

Each list is a literal with no conditional logic, because the build tool reads
the declaration rather than running it. A port whose language has no marker
returns an empty list from `getControllerClasses`.

## Writing a handler

A handler takes the container and the route, and returns an outcome. It reads
the job from the container, because the job is not a parameter.

```text
static chargeCard(container, route) {
    job = container.getSingleton(JobContract)

    if (container.getSingleton(OrderRepository).isCharged(job.getId())) {
        return JobResult.ACK
    }

    container.getSingleton(PaymentGateway).charge(job.getPayload().get("orderId"))

    return JobResult.ACK
}
```

**Write every handler to tolerate a duplicate delivery.** Delivery is at least
once, so the same job can arrive twice. The job's id is stable across every
retry, so it is what an idempotency check keys on.

Return the outcome where the handler knows what happened, and throw only for a
genuine failure. [The outcome](#the-outcome) covers what each one settles as.

## Publishing a job

Resolve a client, build the job, and push it.

```text
client = container.getSingleton(ClientContract)
factory = container.getSingleton(JobFactoryContract)

client.push(factory.create("orders.charge-card", { orderId: order.id }))
```

| Method        | Does                                                   |
| ------------- | ------------------------------------------------------ |
| `push`        | stamps a fresh job and enqueues it                     |
| `retry`       | enqueues a job that is already incremented             |
| `requeue`     | increments the job as dispatched, then retries it      |
| `getPushed`   | returns the jobs this client took in this unit of work |
| `clearPushed` | drops that record, ending the unit of work it covers   |

`push` throws where the processor refuses the job. The framework calls `retry`
and `requeue`, and application code calls `push`.

A publish reports whether the processor accepted the job. It reports nothing
about the run, because the run happens in another process.

### The factory builds the job

The job is a data object, so construction lives on a factory. `create` takes a
name and a body, and the client stamps the rest on the way out.

| Method      | Does                                  |
| ----------- | ------------------------------------- |
| `create`    | builds a job from a name and a body   |
| `fromArray` | builds a job from a decoded envelope  |
| `fromJson`  | builds a job from an encoded envelope |
| `toJson`    | encodes a job as the wire envelope    |

### The producer sets the retry policy

The ceiling, the holds and the ramp ride on the job, so the side that knows what
the work costs decides how often it runs again.

```text
job = factory.create("orders.charge-card", { orderId: order.id })
    .withMaxAttempts(3)
    .withDelayMs(30000)
    .withRetryDelayMs(5000)
    .withRetryDelayMultiplyByAttempt(true)
    .withPriority(10)

client.push(job)
```

| Changes                           | Holds                                               | Default |
| --------------------------------- | --------------------------------------------------- | ------- |
| `withMaxAttempts`                 | the ceiling before the job is dead-lettered         | 5       |
| `withDelayMs`                     | the hold before the job first becomes eligible      | 0       |
| `withRetryDelayMs`                | the hold before a retry                             | 1000    |
| `withRetryDelayMultiplyByAttempt` | whether the retry hold ramps with the attempt count | false   |
| `withPriority`                    | higher runs sooner, where the processor supports it | 0       |

`getRetryDelayForAttemptMs` returns the hold the next retry of this job waits.
With the ramp set it is the retry hold times the attempt count, and it is the
retry hold itself otherwise.

### The envelope

A job crosses process boundaries, and it often crosses ports. A job this
application publishes may run on an application in another language, so the
serialized form is a contract between the ports.

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

A field name is `snake_case` on the wire, and each port maps it to its own
casing internally. An absolute instant and a duration are both integer
milliseconds, in a field suffixed `_ms`, and that value is authoritative.

**No field holds a reference to a class, a type or a function.** A name is a
string and a body is data, so a consumer in another language reads the same job.

**Every field is always present.** An empty multi-map and an empty body are
written as empty rather than left out, so a consumer never distinguishes absent
from default. A consumer ignores a field it does not know, and it defaults a
field an older producer did not send.

### The body and the headers

The body is the data the handler reads. The headers are what a processor and a
middleware read.

```text
payload = job.getPayload()

orderId = payload.get("orderId")
```

| On the body | Does                                           |
| ----------- | ---------------------------------------------- |
| `has`       | answers whether a key holds a value            |
| `get`       | reads the value under a key                    |
| `getAll`    | reads every value                              |
| `with`      | returns a body with the given values           |
| `withAdded` | returns a body with the given values merged in |

```text
job = job.withAttributes(
    job.getAttributes().withAttribute("x-tenant", tenant.id),
)
```

| On the headers       | Does                                 |
| -------------------- | ------------------------------------ |
| `has`                | answers whether a name holds a value |
| `get`                | reads a name's values                |
| `getFirst`           | reads a name's first value           |
| `getLine`            | reads a name's values as one line    |
| `withAttribute`      | replaces a name's values             |
| `withAddedAttribute` | appends to a name's values           |
| `withoutAttributes`  | removes a name                       |

A job, a body and a set of headers are each immutable. Every change returns a
new one, so a middleware that rewrites a job hands the rewritten job on rather
than mutating what the caller holds.

## Running a consumer

Running a consumer is the deployment's job. The framework ships the entry
points, and it ships neither a server nor a long-running command.

| Deployment     | Means                                                  |
| -------------- | ------------------------------------------------------ |
| polling        | the framework asks the processor for the next job      |
| being sent one | the processor sends the job, and reads the answer back |

**Polling needs no server.** Point a process manager at the polling entry for
your processor. The entry boots the application once, and it gives each job its
own child container ([Container](../container/README.md) covers the child
container).

**Being sent a job needs something that receives the delivery.** Point your
deployment at the entry that receives a sent job. A deployment that only polls
loads none of that.

The handler and the middleware are identical either way. They take a job and
return an outcome, blind to how the job arrived.

## The outcome

What the pipeline hands back is a closed set of four.

| Outcome       | Means                                                  |
| ------------- | ------------------------------------------------------ |
| `ACK`         | the job succeeded, and the processor may forget it     |
| `RETRY`       | deliver the job again, after the retry hold            |
| `FAIL`        | the job failed, and retrying it will not help          |
| `DEAD_LETTER` | the job exhausted its attempts, or must not be retried |

| Asks             | Answers                                             |
| ---------------- | --------------------------------------------------- |
| `isTerminal`     | whether nothing further happens to this job         |
| `isDeadLettered` | whether the job goes to the dead-letter destination |

A closed set is what lets one pipeline drive every processor. Turning the
outcome into a processor's own signal is the whole job of the entry.

### Mapping a throwable to an outcome

A throwable caught in the pipeline becomes an outcome, and the retry-policy
middleware is what maps it.

| The throwable            | Becomes                                  |
| ------------------------ | ---------------------------------------- |
| a shutdown               | `RETRY`, with no attempt spent           |
| one marked non-retryable | `FAIL`, at once                          |
| anything else            | `RETRY`, or `DEAD_LETTER` at the ceiling |

A `RETRY` that reaches the ceiling becomes `DEAD_LETTER`, whichever stage
produced it.

### Mark a throwable that must not retry

A throwable carries the **non-retryable marker** where retrying cannot help. A
malformed body and a rejected card are both final, so they fail at once instead
of consuming the attempt budget.

```text
class CardDeclined extends Exception implements QueueNonRetryableThrowable { }
```

```text
static chargeCard(container, route) {
    job = container.getSingleton(JobContract)

    charge = container.getSingleton(PaymentGateway).charge(job.getPayload().get("orderId"))

    if (charge.declined) {
        throw new CardDeclined("The card was declined.")
    }

    return JobResult.ACK
}
```

Warning: catching the component's throwable contract does **not** bound the
catch to the component. The non-retryable marker extends that contract, so a
catch of it also catches an application throwable that carries the marker.

## Middleware

Seven stages run around a job, and each one has its own contract. An
application implements the contract for the stage it wants, and one class may
implement several.

| Stage             | Runs                   | May change                |
| ----------------- | ---------------------- | ------------------------- |
| `JobReceived`     | always                 | the job, or the outcome   |
| `RouteMatched`    | a route was found      | the route, or the outcome |
| `RouteNotMatched` | no route was found     | the outcome               |
| `RouteDispatched` | a route was found      | the outcome               |
| `ThrowableCaught` | a throwable was caught | the outcome               |
| `SettlingResult`  | always                 | the outcome               |
| `ResultSettled`   | always                 | nothing                   |

Each stage's contract declares one method, named for the stage, and that method
takes the stage's own handler as its last parameter. Calling the handler passes
the work to the next middleware; returning an outcome instead ends the pipeline
there.

```text
class RejectStaleJobMiddleware implements JobReceivedMiddlewareContract {

    jobReceived(job, handler) {
        if (job.getEnqueuedAtMs() < now() - 86400000) {
            return JobResult.DEAD_LETTER
        }

        return handler.jobReceived(job)
    }
}
```

`JobReceived` and `RouteMatched` can **short-circuit**: each returns either the
thing it received or an outcome, and an outcome skips every remaining stage
before `SettlingResult`.

`SettlingResult` is the last place the outcome changes, because the entry
settles it next. `ResultSettled` runs after the settlement, so it holds work the
processor never waits for.

### Scheduling middleware

The application's Queue config holds one list per stage, and the provider gives
each stage handler the list for that stage.

```text
class Config implements QueueConfigContract {
    jobReceivedMiddleware = [RejectStaleJobMiddleware]
    throwableCaughtMiddleware = [LogThrowableCaughtMiddleware, RetryPolicyThrowableCaughtMiddleware]
}
```

Middleware runs in registration order, and the framework removes no repeat.
Scheduling one class twice at a stage runs it twice.

### Middleware on one job

Declare it on the route, and the router appends it to the stage handlers for
that job alone.

```text
@Route(name: "charge-card", description: "Charge the card for an order")
@RouteHandler([self, "chargeCard"])
@Middleware(name: RecordChargeAttemptMiddleware)
static chargeCard(container, route) {
    // ...
}
```

A route carries a list per stage, and the stage a class lands in follows from
the contracts the class implements. `RouteNotMatched` takes no route-level
middleware, because no route matched.

## Clients

A client is the produce side. Each one implements `ClientContract`, and each is
registered under its own name, so an application resolves whichever one it
wants. `ClientContract` itself resolves to the client the config names.

| Client              | Publishes to                                               |
| ------------------- | ---------------------------------------------------------- |
| `<Processor>Client` | one processor the port speaks to                           |
| `SyncClient`        | nothing; it runs the job at once, and blocks               |
| `DeferredClient`    | nothing; it holds the job for the end of this unit of work |
| `InMemoryClient`    | nothing; it records the job for a test                     |

Every port ships the three in-process clients. Which processors ship is the
port's own choice, because the component bridges none itself.

The in-process clients are interchangeable with a durable one, so a test and a
local run need no broker. `SyncClient` runs the whole retry chain inline, with
no hold between the attempts.

**Who performs a retry depends on the processor.** A client whose processor has
no native redelivery publishes a modified copy of the job: the attempt count
incremented, the modification time stamped, and the retry hold applied. A client
whose processor owns redelivery hands over the processor's own signal and writes
no copy, and the attempt count arrives from the processor.

### Adding a client

A client implements `ClientContract`. Where this port ships a base the clients
share, extend that instead and implement how the job is published.

```text
class NatsClient extends Client {

    construct(
        connection: NatsConnection,
        factory: JobFactoryContract,
        config: QueueNatsClientConfigContract,
    ) { }

    publish(job) {
        this.connection.publish(
            this.config.queueNatsClientSubject,
            this.factory.toJson(job),
        )
    }
}
```

Give it a config contract of its own, prefixing each property with the client's
name so one config class implements several contracts at once:

```text
interface QueueNatsClientConfigContract {
    queueNatsClientHost: string
    queueNatsClientPort: int
    queueNatsClientSubject: string
}
```

Register the client and its config in a service provider
([Container](../container/README.md) covers providers and `publishers`), then
name the client in `defaultQueueClient`, or resolve it directly where you want
it:

```text
class AppQueueServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            NatsClient: [self, "publishNatsClient"],
        }
    }

    static publishNatsClient(container) {
        container.setSingleton(NatsClient, new NatsClient(
            container.getSingleton(NatsConnection),
            container.getSingleton(JobFactoryContract),
            container.getSingleton(QueueNatsClientConfigContract),
        ))
    }
}
```

An application that publishes through a second client declares a contract of its
own over `ClientContract`, and binds that contract to the second client.

## Configuration

The component reads the application's Queue config for the pipeline, one config
contract of its own for the client, and one per client.

Your application has one config class. Where it implements one of these
contracts, the provider binds your class. Where it does not, the provider binds
the component's own default. So implementing a contract is how you override, and
implementing nothing leaves every default in place.

### The pipeline config

| Property                    | Holds                               | Default                             |
| --------------------------- | ----------------------------------- | ----------------------------------- |
| `jobReceivedMiddleware`     | the middleware at `JobReceived`     | none                                |
| `routeMatchedMiddleware`    | the middleware at `RouteMatched`    | none                                |
| `routeNotMatchedMiddleware` | the middleware at `RouteNotMatched` | none                                |
| `routeDispatchedMiddleware` | the middleware at `RouteDispatched` | none                                |
| `throwableCaughtMiddleware` | the middleware at `ThrowableCaught` | the log and retry-policy middleware |
| `settlingResultMiddleware`  | the middleware at `SettlingResult`  | none                                |
| `resultSettledMiddleware`   | the middleware at `ResultSettled`   | none                                |

Replacing `throwableCaughtMiddleware` replaces the defaults, so a config that
sets it names the retry-policy middleware again. Without it a throwable never
becomes a retry.

Which processor a deployment runs on is the **entry's** own configuration. The
pipeline config holds nothing about a broker, which is what keeps the component
free of any processor.

### The client config

| Property             | Holds                                   | Default                       |
| -------------------- | --------------------------------------- | ----------------------------- |
| `defaultQueueClient` | the client `ClientContract` resolves to | the port's own durable client |

```text
class Config implements ConfigContract, QueueClientConfigContract {
    defaultQueueClient = SyncClient
}
```

### Configuring a processor client

Each processor client carries the settings that processor needs, prefixed with
the client's name.

| Property                      | Holds                            | Default             |
| ----------------------------- | -------------------------------- | ------------------- |
| `queue<Processor>ClientHost`  | where to reach the processor     | the local host      |
| `queue<Processor>ClientPort`  | the port to reach it on          | the processor's own |
| `queue<Processor>ClientQueue` | the queue the jobs are pushed to | a default queue     |

```text
class Config implements ConfigContract, QueueClientConfigContract, QueueRedisClientConfigContract {
    defaultQueueClient = RedisClient

    queueRedisClientHost = "queue.internal"
    queueRedisClientPort = 6379
    queueRedisClientQueue = "queues:orders"
}
```

### Configuring an in-process client

| Property        | Holds                                 | Default |
| --------------- | ------------------------------------- | ------- |
| `syncEntry`     | the entry that runs each job inline   | none    |
| `deferredEntry` | the entry that runs each buffered job | none    |

Each names an entry of your own application, so the framework has no default to
fall back to. The provider throws where your config implements neither contract
and something resolves one of the two clients.

The in-memory client has no config contract. It records each job and publishes
nowhere, so there is nothing to set. Name it as the default in a test that
asserts on what was published.

## Service registration

The component ships one provider per area, and they bind:

| Binding                     | Is                                                 |
| --------------------------- | -------------------------------------------------- |
| `JobFactoryContract`        | the factory that builds a job                      |
| `RouteCollectionContract`   | the jobs, and the lookup the router uses           |
| `RouteCollectorContract`    | the collector for the way this port declares a job |
| `RouterContract`            | the router                                         |
| each stage handler contract | that stage's handler, as a singleton               |
| `JobHandlerContract`        | the handler that runs a job end to end             |
| each shipped middleware     | bound under its own name                           |
| `ClientContract`            | the client the client config names                 |
| each client                 | bound under its own name, for an explicit ask      |
| each client config          | your application's own config class                |

```text
class QueueServerServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            JobHandlerContract: [self, "publishJobHandler"],
            LogThrowableCaughtMiddleware: [self, "publishLogThrowableCaughtMiddleware"],
            RetryPolicyThrowableCaughtMiddleware: [self, "publishRetryPolicyThrowableCaughtMiddleware"],
        }
    }

    static publishJobHandler(container) {
        container.setSingleton(JobHandlerContract, new JobHandler(
            container: container,
            router: container.getSingleton(RouterContract),
            jobReceivedHandler: container.getSingleton(JobReceivedHandlerContract),
            throwableCaughtHandler: container.getSingleton(ThrowableCaughtHandlerContract),
            settlingResultHandler: container.getSingleton(SettlingResultHandlerContract),
            resultSettledHandler: container.getSingleton(ResultSettledHandlerContract),
        ))
    }
}
```

**The router and the job handler resolve the same stage handlers.** Each stage
handler is a singleton, so the middleware a route appends at a stage reaches the
handler that runs it. Two instances means that middleware never runs, and
nothing reports it.

The job handler declares `run` as the entry point a consumer calls, and `handle`
as the method that produces the outcome from the first five stages. The terminal
methods carry their stage names, `settlingResult` and `resultSettled`, and the
entry calls each one around its own settlement.
