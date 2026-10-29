# Sms

Implements
[`SMS.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/sms/SMS.md).

## Introduction

The Sms component sends a text message to a phone number. It sends through a
gateway, records what it would have sent, or discards it, and your application
code reads the same in every case.

An **adapter** is the thing that sends. Each one implements
`MessengerContract`, and each is registered under its own name, so an
application resolves whichever one it wants.

`MessengerContract` itself resolves to the **default** adapter: the one your
config names. It is the convenience, and it is what lets a development
environment record a message while production sends it, with no call site
changing.

The contract declares one operation, `send`. It takes a message and returns
nothing. A send completes, or it throws.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a messenger

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                |
| -------------------------- | ----------------------------------------- | ------------------------ |
| `MessengerContract`        | the adapter the config names              | environment              |
| an adapter by its own name | that adapter, always                      | nothing                  |
| a contract of your own     | the adapter that contract's binding names | purpose, and environment |

### The default, through `MessengerContract`

Use this where the gateway is the application's business rather than this call
site's, which is almost everywhere.

```text
messenger = container.getSingleton(MessengerContract)

messenger.send(message)
```

### A named adapter, where the gateway is the point

Use this where you know this call site means that gateway whatever the
environment is.

```text
messenger = container.getSingleton(VonageMessenger)

messenger.send(message)
```

### A contract of your own, per purpose

Declare a contract that extends `MessengerContract` for a kind of message, and
bind it to whichever adapter that kind wants. Then the adapter behind each
purpose varies on its own, per environment, and a call site names the purpose
rather than a gateway.

```text
interface AlertMessengerContract extends MessengerContract { }
interface MarketingMessengerContract extends MessengerContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppSmsServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            AlertMessengerContract: [self, "publishAlertMessenger"],
            MarketingMessengerContract: [self, "publishMarketingMessenger"],
        }
    }

    static publishAlertMessenger(container) {
        container.setSingleton(
            AlertMessengerContract,
            container.getSingleton(VonageMessenger),
        )
    }

    static publishMarketingMessenger(container) {
        container.setSingleton(
            MarketingMessengerContract,
            container.getSingleton(NullMessenger),
        )
    }
}
```

The publishers are where an environment decides what each purpose sends
through. An environment that sends an alert and discards a promotion is two
bindings.

## Sending a message

`send` takes a `MessageContract`. The component ships `Message`, which carries
these parts:

| Part        | Holds                                         | Read with   | Replace with    |
| ----------- | --------------------------------------------- | ----------- | --------------- |
| `to`        | the number the message goes to                | `getTo`     | `withTo`        |
| `from`      | the number or sender id it comes from         | `getFrom`   | `withFrom`      |
| `text`      | the body                                      | `getText`   | `withText`      |
| `isUnicode` | whether the text needs the wide character set | `isUnicode` | `withIsUnicode` |

There is no recipient object and no attachment. The component sends one text to
one number.

```text
message = new Message(
    to: "+15555550123",
    from: "+15555550100",
    text: "Your order has shipped.",
)

messenger.send(message)
```

### A message does not mutate

A `with` form returns a copy and leaves the message it was called on alone, so
keep the copy from each step.

```text
// Wrong. The copy is discarded, so the message still carries the old text.
message.withText("Your order is ready for collection.")

// Right. Each with form returns the message to carry forward.
message = message.withText("Your order is ready for collection.")
```

A message built once and sent to several numbers is one `with` form per send:

```text
for (customer in customers) {
    messenger.send(message.withTo(customer.phone))
}
```

### The encoding is the message's, not the adapter's

A text outside the basic character set costs more for each message, and it
splits into parts differently. The caller states whether the text is unicode,
and every adapter honors that answer.

Warning: a wrong answer changes what a send costs and where the text breaks. Set
the encoding on every message you build, rather than leaving the default the
constructor carries.

```text
// An emoji, an accent, or any character outside the basic set.
message = message.withIsUnicode(true)
```

### Send from a queue, not from a request

A send reaches a third party and costs money for each message. Push a job and
send from the consumer ([Queue](../queue/README.md)). A failure then retries
once, rather than being retried by a person who presses a button again.

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property           | Holds                                       | Default                |
| ------------------ | ------------------------------------------- | ---------------------- |
| `defaultMessenger` | the adapter `MessengerContract` resolves to | the port's own gateway |

### Changing the default adapter

Name a different adapter in `defaultMessenger`, and every call that resolves
`MessengerContract` gets it. No call site changes, and a call that names an
adapter directly is unaffected.

```text
class Config implements ConfigContract, SmsConfigContract {
    defaultMessenger = LogMessenger
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter              | Sends through                          |
| -------------------- | -------------------------------------- |
| `<Gateway>Messenger` | one gateway the port speaks to         |
| `LogMessenger`       | nothing; it records what it would send |
| `NullMessenger`      | nothing; every message is discarded    |

Every port ships the log and null adapters. Which gateways ship is the port's
own choice.

### Configuring a gateway adapter

Each gateway adapter carries the credentials that gateway needs, prefixed with
the gateway's name. The Vonage adapter takes these:

| Property       | Holds                 | Default |
| -------------- | --------------------- | ------- |
| `vonageKey`    | the gateway's api key | —       |
| `vonageSecret` | the api secret        | —       |

```text
class Config implements ConfigContract, SmsConfigContract, SmsVonageConfigContract {
    defaultMessenger = VonageMessenger

    vonageKey = environment("VONAGE_KEY")
    vonageSecret = environment("VONAGE_SECRET")
}
```

A credential belongs in the environment rather than in the class. A port adds a
subsection for each further gateway it ships.

### Configuring the log adapter

| Property       | Holds                          | Default            |
| -------------- | ------------------------------ | ------------------ |
| `smsLogLogger` | which logger the records go to | the default logger |

The log adapter names a logger rather than holding one, so the records go
wherever that logger goes ([Log](../log/README.md)). It records the sender, the
receiver and the text.

Every port ships this adapter. A real send reaches a real phone and costs money,
so a development environment binds the log adapter and reads what would have
gone out. A test reads the same records to assert that a send was made.

### Configuring the null adapter

The null adapter has no config contract. It discards every message, so there is
nothing to set.

Name it as the default in a test that must send nothing at all.

## Adding an adapter

An adapter implements `MessengerContract`. Where this port ships a base the
adapters share, extend that instead.

```text
class TwilioMessenger implements MessengerContract {

    construct(client: TwilioClient, config: SmsTwilioConfigContract) { }

    send(message) {
        this.client.create(
            to: message.getTo(),
            from: message.getFrom(),
            body: message.getText(),
            encoding: message.isUnicode() ? "unicode" : "text",
        )
    }
}
```

An adapter reads every part of the message. One that drops the encoding sends a
text the gateway then splits its own way.

Give it a config contract of its own, prefixing each property with the gateway's
name so one config class implements several contracts at once:

```text
interface SmsTwilioConfigContract {
    twilioSid: string
    twilioToken: string
}
```

Register the adapter and its config in a service provider, then name the adapter
in `defaultMessenger`, or resolve it directly where you want it:

```text
class AppSmsServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            TwilioMessenger: [self, "publishTwilioMessenger"],
        }
    }

    static publishTwilioMessenger(container) {
        container.setSingleton(TwilioMessenger, new TwilioMessenger(
            container.getSingleton(TwilioClient),
            container.getSingleton(SmsTwilioConfigContract),
        ))
    }
}
```

## Service registration

The component's provider binds:

| Binding               | Is                                            |
| --------------------- | --------------------------------------------- |
| `MessengerContract`   | the adapter the component config names        |
| each adapter          | bound under its own name, for an explicit ask |
| each gateway's client | built from that gateway's config              |
| the component config  | your application's own config class           |
| each adapter config   | your application's own config class           |

```text
class SmsServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            SmsConfigContract: [self, "publishConfig"],
            SmsVonageConfigContract: [self, "publishVonageConfig"],
            SmsLogConfigContract: [self, "publishLogConfig"],
            MessengerContract: [self, "publishMessenger"],
            VonageMessenger: [self, "publishVonageMessenger"],
            VonageClient: [self, "publishVonageClient"],
            LogMessenger: [self, "publishLogMessenger"],
            NullMessenger: [self, "publishNullMessenger"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements SmsConfigContract) {
            container.setSingleton(SmsConfigContract, config)

            return
        }

        container.setSingleton(SmsConfigContract, new SmsConfig())
    }

    static publishMessenger(container) {
        config = container.getSingleton(SmsConfigContract)

        container.setSingleton(
            MessengerContract,
            container.getSingleton(config.defaultMessenger),
        )
    }

    // The adapter takes the gateway's client, not the config.
    static publishVonageMessenger(container) {
        container.setSingleton(VonageMessenger, new VonageMessenger(
            container.getSingleton(VonageClient),
        ))
    }

    // A publisher of its own builds the client from the config.
    static publishVonageClient(container) {
        config = container.getSingleton(SmsVonageConfigContract)

        container.setSingleton(VonageClient, new VonageClient(
            key: config.vonageKey,
            secret: config.vonageSecret,
        ))
    }

    static publishLogMessenger(container) {
        config = container.getSingleton(SmsLogConfigContract)

        container.setSingleton(LogMessenger, new LogMessenger(
            container.getSingleton(config.smsLogLogger),
        ))
    }

    static publishNullMessenger(container) {
        container.setSingleton(NullMessenger, new NullMessenger())
    }
}
```
