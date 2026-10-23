# Log

Implements
[`LOG.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/log/LOG.md).

## Introduction

The Log component records what happened in your application, at a severity. It
writes to a file, to the system log, to a third-party service, or nowhere at all,
and your application code reads the same in every case.

An **adapter** is the thing that records. Each one implements `LoggerContract`,
and each is registered under its own name, so an application resolves whichever
one it wants.

`LoggerContract` itself resolves to the **default** adapter: the one your config
names, or the one the framework's own provider names where your config does not.
It is a convenience, and it is what lets an environment change which adapter
records without any call site changing — a file in development, a hosted service
in production, nothing at all in a test.

`LoggerContract` declares the shape a standard logging interface has: one method
per severity, a general method that takes the severity as a value, and a map of
context on each. Where the language has such an interface, the contract extends
it directly, so the component's logger works anywhere that interface is expected.

Every other component in the framework logs through this component. Nothing
writes to a file or a stream of its own.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a logger

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

The component ships `LoggerContract`, bound to the adapter your config names.
You can declare contracts of your own over the same methods, and you can resolve
an adapter class directly.

| Resolve                    | Gets                                      | Varies by                 |
| -------------------------- | ----------------------------------------- | ------------------------- |
| `LoggerContract`           | the adapter the config names              | environment               |
| an adapter by its own name | that adapter, always                      | nothing                   |
| a contract of your own     | the adapter that contract's binding names | activity, and environment |

### The default, through `LoggerContract`

Use this where the destination is the application's business rather than this
call site's, which is almost everywhere.

```text
logger = container.getSingleton(LoggerContract)

logger.info("Order placed")
```

### A named adapter, where the destination is the point

Use this where you know this call site means that destination whatever the
environment is — an audit trail that has to land in the same place everywhere.

```text
auditLogger = container.getSingleton(StandardLogger)

auditLogger.info("Order placed", { orderId: order.id })
```

### A contract of your own, per logging activity

Declare a contract that extends `LoggerContract` for a kind of logging activity,
and bind it to whichever adapter that activity wants. Then the adapter behind
each activity varies on its own, per environment, and a call site names the
activity rather than a destination.

```text
interface UserLoggerContract extends LoggerContract { }
interface ProductLoggerContract extends LoggerContract { }
interface SystemLoggerContract extends LoggerContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppLogServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            UserLoggerContract: [self, "publishUserLogger"],
            ProductLoggerContract: [self, "publishProductLogger"],
            SystemLoggerContract: [self, "publishSystemLogger"],
        }
    }

    static publishUserLogger(container) {
        container.setSingleton(
            UserLoggerContract,
            container.getSingleton(StandardLogger),
        )
    }

    static publishSystemLogger(container) {
        container.setSingleton(
            SystemLoggerContract,
            container.getSingleton(SyslogLogger),
        )
    }

    // ...one publisher per contract
}
```

The publishers are where an environment decides what each activity writes to.

### The same adapter, twice, with different config

A publisher constructs what it registers, so two contracts can share one adapter
class and give it different settings. Put the settings on your own config, and
read them in the publisher:

```text
interface AppLogConfigContract {
    userLogName: string
    userLogPath: string
    productLogName: string
    productLogPath: string
}
```

```text
class AppLogServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            UserLoggerContract: [self, "publishUserLogger"],
            ProductLoggerContract: [self, "publishProductLogger"],
        }
    }

    static publishUserLogger(container) {
        config = container.getSingleton(AppLogConfigContract)

        container.setSingleton(UserLoggerContract, new StandardLogger(
            name: config.userLogName,
            path: config.userLogPath,
            level: Severity.INFO,
        ))
    }

    static publishProductLogger(container) {
        config = container.getSingleton(AppLogConfigContract)

        container.setSingleton(ProductLoggerContract, new StandardLogger(
            name: config.productLogName,
            path: config.productLogPath,
            level: Severity.WARNING,
        ))
    }
}
```

One adapter class, two destinations, and each contract varies on its own.

Resolve the contract for the activity where that activity happens. It declares
the same methods as `LoggerContract`:

```text
class UserController {

    construct(logger: UserLoggerContract) { }

    show(request) {
        this.logger.info("Profile viewed", { userId: request.params.id })
    }
}
```

Sending user records to a hosted service and system records to the system log is
two bindings. A test that wants neither binds both to the null writer.

## Writing log messages

Call the method for the severity you mean, on whichever logger you resolved.

| Method      | Records                                        | Takes                           | Returns |
| ----------- | ---------------------------------------------- | ------------------------------- | ------- |
| `debug`     | detail useful only while diagnosing            | a message, and context          | nothing |
| `info`      | something that happened and is worth a record  | a message, and context          | nothing |
| `notice`    | something unusual that is not a problem        | a message, and context          | nothing |
| `warning`   | something that will become a problem           | a message, and context          | nothing |
| `error`     | a failure that did not stop the application    | a message, and context          | nothing |
| `critical`  | a failure that stopped part of the application | a message, and context          | nothing |
| `alert`     | a failure that needs someone now               | a message, and context          | nothing |
| `emergency` | the application is unusable                    | a message, and context          | nothing |
| `log`       | a record at the severity you pass              | a severity, a message, context  | nothing |
| `throwable` | a throwable, with its type, message and trace  | a throwable, a message, context | nothing |

Every method returns nothing. An adapter that cannot write does not stop the
unit of work.

Context is optional on every method. The severity decides where a record goes,
and whether it wakes anyone.

```text
logger.info("Order placed")
logger.warning("Payment gateway slow to answer")
logger.error("Payment declined")
```

A longer example, in a controller:

```text
class OrderController {

    construct(logger: LoggerContract, orders: OrderRepository) { }

    place(request) {
        order = this.orders.create(request.body)

        this.logger.info("Order placed", { orderId: order.id })

        return response.created(order)
    }
}
```

### Record a throwable with `throwable`

Use `throwable` wherever the application catches one. It records the throwable's
type, its message and its trace, and it puts the throwable in the context itself.

```text
try {
    gateway.charge(order)
} catch (throwable) {
    logger.throwable(throwable, "Charge failed")
}
```

A throwable the application formats itself loses the trace, or reads differently
in every place that logs one.

### Choose the severity at run time with `log`

Where the severity is a value rather than a decision in the code, pass it:

```text
severity = response.ok ? Severity.INFO : Severity.ERROR

logger.log(severity, "Webhook delivered", { status: response.status })
```

## Context

Context is a map of the values a reader of the record needs. It travels with the
record, and the adapter writes it beside the message rather than inside it.

```text
// Wrong. The id is in the message, so every record reads differently.
logger.info("User 4821 failed to sign in")

// Right. The message is constant, and the id is context.
logger.info("User failed to sign in", { userId: user.id })
```

The difference matters once there is more than one record. A message that stays
constant can be grouped, counted and alerted on, so a thousand failed sign-ins
become one line in a report with a count beside it. A message carrying the id in
its text is a thousand distinct messages, and nothing groups them.

What an adapter does with the map depends on the adapter. One writing a line to a
file appends the map to the line. One sending to a hosted service sends the map
as fields that service can query. Either way, a value a reader needs goes in the
context and not in the message.

An adapter may also interpolate a context value into the message where the
message names it. A port documents the form its own adapters accept.

```text
logger.info("User {userId} failed to sign in", { userId: user.id })
```

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

```text
// In the component's provider. Your class, or the component's default.
static publishConfig(container) {
    config = container.getSingleton(ConfigContract)

    if (config implements LogConfigContract) {
        container.setSingleton(LogConfigContract, config)

        return
    }

    container.setSingleton(LogConfigContract, new LogConfig())
}
```

### The component config

| Property        | Holds                                    | Default                       |
| --------------- | ---------------------------------------- | ----------------------------- |
| `defaultLogger` | the adapter `LoggerContract` resolves to | the standard-interface writer |

### Changing the default adapter

Name a different adapter in `defaultLogger`, and every call that resolves
`LoggerContract` gets it. No call site changes, and a call that names an adapter
directly is unaffected.

```text
class Config implements ConfigContract, LogConfigContract {
    defaultLogger = NullLogger
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter         | Writes to                             |
| --------------- | ------------------------------------- |
| standard writer | the ecosystem's own logging interface |
| null writer     | nowhere; every record is discarded    |

Every port ships both. A port adds a row for each further integration it ships.

### Configuring the standard writer

| Property        | Holds                         | Default                      |
| --------------- | ----------------------------- | ---------------------------- |
| `standardName`  | the channel the records go to | the application, and today   |
| `standardPath`  | where the records are written | the framework's log location |
| `standardLevel` | the lowest severity it writes | `debug`                      |

A default that is only known at run time, such as today's date in a channel name,
is not a value the config holds. The adapter derives it when it builds.

`standardLevel` drops every record below it. Setting it to `error` on a hosted
service keeps the noise out while a file adapter still takes everything.

```text
class Config implements ConfigContract, LogConfigContract, LogStandardConfigContract {
    defaultLogger = StandardLogger

    standardName = "orders"
    standardPath = "/var/log/orders"
    standardLevel = Severity.WARNING
}
```

### Configuring the null writer

The null writer has no config contract. It discards every record, so there is
nothing to set.

Name it as the default in a test, or in an environment that writes no records at
all.

## Adding an adapter

An adapter implements `LoggerContract`. Where this port ships a base the adapters
share, extend that instead and implement one method per severity.

```text
class SlackLogger implements LoggerContract {

    construct(config: LogSlackConfigContract) { }

    info(message, context = {}) {
        this.send(Severity.INFO, message, context)
    }

    // ...one method per severity, plus log and throwable

    send(severity, message, context) {
        if (severity.isBelow(this.config.slackLevel)) {
            return
        }

        http.post(this.config.slackWebhook, { text: message, fields: context })
    }
}
```

Give it a config contract of its own where it needs settings, prefixing each
property with the adapter's name so one config class implements several contracts
at once:

```text
interface LogSlackConfigContract {
    slackWebhook: string
    slackLevel: Severity
}
```

Register the adapter and its config in a service provider. Then name the adapter
in `defaultLogger`, or resolve it directly where you want it:

```text
class AppLogServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            SlackLogger: [self, "publishSlackLogger"],
        }
    }

    static publishSlackLogger(container) {
        container.setSingleton(SlackLogger, new SlackLogger(
            container.getSingleton(LogSlackConfigContract),
        ))
    }
}
```

An adapter you write takes whatever it needs. The component's own adapters take a
writer because a separate publisher builds that writer from config.

A component that reaches a third party also ships a **log writer** of its own,
which records what it would have done instead of doing it. That writer resolves
this component's contract, so the dependency runs from that component to this
one.

## Service registration

The component's provider binds:

| Binding              | Is                                            |
| -------------------- | --------------------------------------------- |
| `LoggerContract`     | the adapter the component config names        |
| each adapter         | bound under its own name, for an explicit ask |
| the component config | your application's own config class           |
| each adapter config  | your application's own config class           |

```text
class LogServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            LogConfigContract: [self, "publishConfig"],
            LogStandardConfigContract: [self, "publishStandardConfig"],
            LoggerContract: [self, "publishLogger"],
            StandardLogger: [self, "publishStandardLogger"],
            StandardWriter: [self, "publishStandardWriter"],
            NullLogger: [self, "publishNullLogger"],
        }
    }

    static publishLogger(container) {
        config = container.getSingleton(LogConfigContract)

        container.setSingleton(
            LoggerContract,
            container.getSingleton(config.defaultLogger),
        )
    }

    // The adapter takes the writer, not the config.
    static publishStandardLogger(container) {
        container.setSingleton(StandardLogger, new StandardLogger(
            container.getSingleton(StandardWriter),
        ))
    }

    // A publisher of its own builds the writer from the config.
    static publishStandardWriter(container) {
        config = container.getSingleton(LogStandardConfigContract)

        container.setSingleton(StandardWriter, new StandardWriter(
            path: config.standardPath ?? Directory.logs(),
            name: config.standardName ?? applicationName() + today(),
            level: config.standardLevel,
        ))
    }

    static publishNullLogger(container) {
        container.setSingleton(NullLogger, new NullLogger())
    }
}
```
