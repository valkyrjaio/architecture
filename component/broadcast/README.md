# Broadcast

Implements
[`BROADCAST.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/broadcast/BROADCAST.md).

## Introduction

The Broadcast component pushes a message to listeners connected over a live
channel, so a browser or a device learns that something happened without asking
for it. It pushes through a service, records what it would have pushed, or
discards it, and your application code reads the same in every case.

An **adapter** is the thing that pushes. Each one implements
`BroadcasterContract`, and each is registered under its own name, so an
application resolves whichever one it wants.

`BroadcasterContract` itself resolves to the **default** adapter: the one your
config names. It is the convenience, and it is what lets a development
environment record a push while production makes it, with no call site changing.

The contract declares one operation, `send`. It takes a message and returns
nothing. A push completes, or it throws.

The component owns the message and the push. It owns no connection: holding the
socket is the service the adapter speaks to. The component authorizes nothing
either, so whether a listener may join a channel is that service's to enforce.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a broadcaster

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                |
| -------------------------- | ----------------------------------------- | ------------------------ |
| `BroadcasterContract`      | the adapter the config names              | environment              |
| an adapter by its own name | that adapter, always                      | nothing                  |
| a contract of your own     | the adapter that contract's binding names | purpose, and environment |

### The default, through `BroadcasterContract`

Use this where the service is the application's business rather than this call
site's, which is almost everywhere.

```text
broadcaster = container.getSingleton(BroadcasterContract)

broadcaster.send(message)
```

### A named adapter, where the service is the point

Use this where you know this call site means that adapter whatever the
environment is.

```text
broadcaster = container.getSingleton(PusherBroadcaster)

broadcaster.send(message)
```

### A contract of your own, per purpose

Declare a contract that extends `BroadcasterContract` for a kind of broadcast,
and bind it to whichever adapter that kind wants. Then a public announcement and
a message for one person push through different adapters without a call site
knowing.

```text
interface PublicBroadcasterContract extends BroadcasterContract { }
interface PrivateBroadcasterContract extends BroadcasterContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppBroadcastServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            PublicBroadcasterContract: [self, "publishPublicBroadcaster"],
            PrivateBroadcasterContract: [self, "publishPrivateBroadcaster"],
        }
    }

    static publishPublicBroadcaster(container) {
        container.setSingleton(
            PublicBroadcasterContract,
            container.getSingleton(PusherBroadcaster),
        )
    }

    static publishPrivateBroadcaster(container) {
        container.setSingleton(
            PrivateBroadcasterContract,
            container.getSingleton(CryptPusherBroadcaster),
        )
    }
}
```

The publishers are where an environment decides what each purpose pushes
through. A call site then names the purpose, and the encryption is a binding
rather than a decision in the code.

## Sending a message

`send` takes a `MessageContract`. The component ships `Message`, which carries
these parts:

| Part      | Holds                                                     | Read with    | Replace with  |
| --------- | --------------------------------------------------------- | ------------ | ------------- |
| `channel` | which listeners receive it                                | `getChannel` | `withChannel` |
| `event`   | what happened, as a name the listener matches on          | `getEvent`   | `withEvent`   |
| `data`    | the payload, as structured data rather than text          | `getData`    | `withData`    |
| `message` | a human-readable form, for a listener that shows it as-is | `getMessage` | `withMessage` |

```text
message = new Message(
    channel: "orders.4821",
    event: "order.shipped",
    message: "Your order has shipped.",
    data: { orderId: 4821, carrier: "post" },
)

broadcaster.send(message)
```

A listener subscribes to the channel and matches on the event. It reads the data
for the values it needs, and it shows the message as-is.

### A message does not mutate

A `with` form returns a copy and leaves the message it was called on alone, so
keep the copy from each step.

```text
// Wrong. The copy is discarded, so the message still carries the old channel.
message.withChannel("orders.4822")

// Right. Each with form returns the message to carry forward.
message = message.withChannel("orders.4822")
```

One message built once and pushed to several channels is one `with` form per
push:

```text
for (watcher in watchers) {
    broadcaster.send(message.withChannel("user." + watcher.id))
}
```

### The channel decides who hears it

The channel decides who receives the message, and the event decides what the
listener does with it. One channel carries many events, so a listener subscribes
once and matches on the event each time.

Warning: a listener reads everything on a channel it joined. Put anything that
only one person may see on a channel only that person can subscribe to.

```text
// Wrong. Every listener on the application's channel reads the balance.
message = new Message(
    channel: "app",
    event: "balance.updated",
    message: "Your balance changed.",
    data: { balance: account.balance },
)

// Right. The channel names the one person who may read it.
message = new Message(
    channel: "user." + account.userId,
    event: "balance.updated",
    message: "Your balance changed.",
    data: { balance: account.balance },
)
```

### Broadcast from a listener, not from the code that acted

Dispatch an event when something happens, and broadcast from a listener for that
event ([Event](../event/README.md)). The code that changed the state then knows
nothing about who is watching, and a push that fails cannot fail the change.

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property             | Holds                                         | Default                |
| -------------------- | --------------------------------------------- | ---------------------- |
| `defaultBroadcaster` | the adapter `BroadcasterContract` resolves to | the port's own service |

### Changing the default adapter

Name a different adapter in `defaultBroadcaster`, and every call that resolves
`BroadcasterContract` gets it. No call site changes, and a call that names an
adapter directly is unaffected.

```text
class Config implements ConfigContract, BroadcastConfigContract {
    defaultBroadcaster = LogBroadcaster
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter                     | Pushes through                               |
| --------------------------- | -------------------------------------------- |
| `<Service>Broadcaster`      | one service the port speaks to               |
| `Crypt<Service>Broadcaster` | the same service, with the message encrypted |
| `LogBroadcaster`            | nothing; it records what it would push       |
| `NullBroadcaster`           | nothing; every push is discarded             |

Every port ships the log and null adapters. Which services ship is the port's
own choice.

### Configuring a service adapter

Each service adapter carries the settings that service needs, prefixed with the
service's name. The Pusher adapter takes these:

| Property        | Holds                       | Default |
| --------------- | --------------------------- | ------- |
| `pusherKey`     | the service's app key       | —       |
| `pusherSecret`  | the app secret              | —       |
| `pusherId`      | the app id                  | —       |
| `pusherCluster` | which cluster to connect to | `us1`   |
| `pusherUseTls`  | whether to connect over TLS | on      |

```text
class Config implements ConfigContract, BroadcastConfigContract, BroadcastPusherConfigContract {
    defaultBroadcaster = PusherBroadcaster

    pusherKey = environment("PUSHER_KEY")
    pusherSecret = environment("PUSHER_SECRET")
    pusherId = environment("PUSHER_ID")
    pusherCluster = "us1"
    pusherUseTls = true
}
```

A credential belongs in the environment rather than in the class. A port adds a
subsection for each further service it ships.

### The adapter that encrypts before it pushes

A pushed message reaches a browser, which is a machine you do not control. The
component ships an adapter that encrypts the message and then hands it to the
adapter that pushes it. Its name carries both parts, with the encrypting part
first: `CryptPusherBroadcaster` encrypts, then pushes through Pusher.

Name it where the payload is the reader's alone. The listener holds the key and
decrypts what it receives.

```text
class Config implements ConfigContract, BroadcastConfigContract, BroadcastPusherConfigContract {
    defaultBroadcaster = CryptPusherBroadcaster

    pusherKey = environment("PUSHER_KEY")
    pusherSecret = environment("PUSHER_SECRET")
    pusherId = environment("PUSHER_ID")
}
```

It has no config contract of its own. It reads the config of the adapter that
pushes, and it resolves [Crypt](../crypt/README.md) for the encryption. That
component's config names the key.

The channel is still the unit of access. Encryption keeps a listener on the
channel from reading the payload, and it does not keep the listener off the
channel.

### Configuring the log adapter

| Property             | Holds                          | Default            |
| -------------------- | ------------------------------ | ------------------ |
| `broadcastLogLogger` | which logger the records go to | the default logger |

The log adapter names a logger rather than holding one, so the records go
wherever that logger goes ([Log](../log/README.md)). It records the channel, the
event, the data and the message.

Every port ships this adapter. A real push needs a live service and connected
listeners, so a development environment binds the log adapter and reads what
would have gone out. A test reads the same records to assert that a push was
made.

### Configuring the null adapter

The null adapter has no config contract. It discards every message, so there is
nothing to set.

Name it as the default in a test that must push nothing at all.

## Adding an adapter

An adapter implements `BroadcasterContract`. Where this port ships a base the
adapters share, extend that instead.

```text
class AblyBroadcaster implements BroadcasterContract {

    construct(client: AblyClient, config: BroadcastAblyConfigContract) { }

    send(message) {
        this.client.publish(
            channel: message.getChannel(),
            event: message.getEvent(),
            payload: message.getData(),
            text: message.getMessage(),
        )
    }
}
```

Give it a config contract of its own, prefixing each property with the service's
name so one config class implements several contracts at once:

```text
interface BroadcastAblyConfigContract {
    ablyKey: string
}
```

Register the adapter and its config in a service provider, then name the adapter
in `defaultBroadcaster`, or resolve it directly where you want it:

```text
class AppBroadcastServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            AblyBroadcaster: [self, "publishAblyBroadcaster"],
        }
    }

    static publishAblyBroadcaster(container) {
        container.setSingleton(AblyBroadcaster, new AblyBroadcaster(
            container.getSingleton(AblyClient),
            container.getSingleton(BroadcastAblyConfigContract),
        ))
    }
}
```

To encrypt before an adapter of your own pushes, declare a second adapter that
takes the first one. It encrypts the message, then hands the message on. Name the
class for both parts, with the encrypting part first.

## Service registration

The component's provider binds:

| Binding               | Is                                            |
| --------------------- | --------------------------------------------- |
| `BroadcasterContract` | the adapter the component config names        |
| each adapter          | bound under its own name, for an explicit ask |
| each service's client | built from that service's config              |
| the component config  | your application's own config class           |
| each adapter config   | your application's own config class           |

```text
class BroadcastServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            BroadcastConfigContract: [self, "publishConfig"],
            BroadcastPusherConfigContract: [self, "publishPusherConfig"],
            BroadcastLogConfigContract: [self, "publishLogConfig"],
            BroadcasterContract: [self, "publishBroadcaster"],
            PusherBroadcaster: [self, "publishPusherBroadcaster"],
            CryptPusherBroadcaster: [self, "publishCryptPusherBroadcaster"],
            PusherClient: [self, "publishPusherClient"],
            LogBroadcaster: [self, "publishLogBroadcaster"],
            NullBroadcaster: [self, "publishNullBroadcaster"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements BroadcastConfigContract) {
            container.setSingleton(BroadcastConfigContract, config)

            return
        }

        container.setSingleton(BroadcastConfigContract, new BroadcastConfig())
    }

    static publishBroadcaster(container) {
        config = container.getSingleton(BroadcastConfigContract)

        container.setSingleton(
            BroadcasterContract,
            container.getSingleton(config.defaultBroadcaster),
        )
    }

    // The adapter takes the service's client, not the config.
    static publishPusherBroadcaster(container) {
        container.setSingleton(PusherBroadcaster, new PusherBroadcaster(
            container.getSingleton(PusherClient),
        ))
    }

    // The encrypting adapter takes the adapter that pushes, plus the crypt.
    static publishCryptPusherBroadcaster(container) {
        container.setSingleton(CryptPusherBroadcaster, new CryptPusherBroadcaster(
            container.getSingleton(PusherBroadcaster),
            container.getSingleton(CryptContract),
        ))
    }

    // A publisher of its own builds the client from the config.
    static publishPusherClient(container) {
        config = container.getSingleton(BroadcastPusherConfigContract)

        container.setSingleton(PusherClient, new PusherClient(
            key: config.pusherKey,
            secret: config.pusherSecret,
            id: config.pusherId,
            cluster: config.pusherCluster,
            useTls: config.pusherUseTls,
        ))
    }

    static publishLogBroadcaster(container) {
        config = container.getSingleton(BroadcastLogConfigContract)

        container.setSingleton(LogBroadcaster, new LogBroadcaster(
            container.getSingleton(config.broadcastLogLogger),
        ))
    }

    static publishNullBroadcaster(container) {
        container.setSingleton(NullBroadcaster, new NullBroadcaster())
    }
}
```
