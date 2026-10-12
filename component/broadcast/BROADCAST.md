# Broadcast

The Broadcast component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Broadcast pushes a message to listeners connected over a live channel, so a
browser or a device learns that something happened without asking. It is a
**service component** on the shape in [`LOG.md`](../log/LOG.md): one contract, several
adapters behind it, and the application names the one it wants in config.

The component owns the message and the push. It owns no connection: holding the
socket is the adapter's service, which is why an adapter is a third-party client
rather than a server this framework runs.

---

## Hierarchy

| Subcomponent  | Holds                                                         |
| ------------- | ------------------------------------------------------------- |
| `Broadcaster` | the component contract, and one class per adapter             |
| `Data`        | the message, the component config, and one config per adapter |
| `Provider`    | the service provider that binds the contract to an adapter    |
| `Throwable`   | the component's throwable contract and its exceptions         |

---

## One operation

The contract declares **`send`**, and nothing else. Mail and Sms declare the same
one operation, for the same reason: the component does one thing to one message.

---

## The message

A message is **immutable**: a reader returns the value, and a `with` form returns
a copy ([`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md)).

| Holds       | Is                                                        |
| ----------- | --------------------------------------------------------- |
| the channel | which listeners receive it                                |
| the event   | what happened, as a name the listener matches on          |
| the data    | the payload, as structured data rather than text          |
| the message | a human-readable form, for a listener that shows it as-is |

**The channel and the event are separate.** The channel decides who receives the
message and the event decides what they do with it, so one channel carries many
events and a listener subscribes once.

**The data is structured.** A listener parses no text, which is what keeps a
payload readable by a listener written in another language.

---

## Structure

| Subcomponent   | Holds                                                  |
| -------------- | ------------------------------------------------------ |
| `Broadcaster/` | the component contract, and one class per adapter      |
| `Data/`        | the message, and the configs                           |
| `Provider/`    | the component aggregator, and the component's bindings |
| `Throwable/`   | the component's throwable contract and its exceptions  |

### An adapter is named for what it wraps, then for the role

| Adapter               | Wraps                                       |
| --------------------- | ------------------------------------------- |
| `<Integration><Role>` | one service the port sends through          |
| `Log<Role>`           | nothing; it records what it would have sent |
| `Null<Role>`          | nothing; every call is discarded            |

The role comes last, so a reader scanning a directory sees what the classes are
before seeing what each one wraps. A name stacks where one adapter wraps
another, with the outer behavior first.

**The log adapter is required here.** This component reaches a third party, so a
developer needs the calls recorded without them leaving the machine, and a test
needs to assert that one was made.

### An adapter may wrap another adapter

A message pushed to a live channel reaches a browser, so a port ships an adapter
that encrypts the message and then hands it to the adapter that pushes it. The
name carries both, with the encrypting part first.

That adapter resolves the component that encrypts, so the dependency runs from
this component to that one.

### Order

| In                   | The order is                                        |
| -------------------- | --------------------------------------------------- |
| the adapters         | each integration, then the log adapter, then null   |
| the config contracts | the component's own, then one per adapter           |
| the provider's map   | the configs, the component contract, the adapters   |
| the throwables       | the contract, the two bases, then the concrete ones |

---

## Using it

### Configure it

The component config names the **default broadcaster**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Name a channel for who may hear it

A channel is the unit of access. Put anything that only one person may see on a
channel only that person can subscribe to, because the component authorizes
nothing: whether a listener may join a channel is the adapter's service to
enforce.

### Broadcast from a listener, not from the code that acted

Dispatch an event when something happens, and broadcast from a listener for that
event ([`EVENT.md`](../event/EVENT.md)). The code that changed the state then knows
nothing about who is watching, and a broadcast failure cannot fail the change.

---

## The adapters every port ships

| Adapter | Does                                                               |
| ------- | ------------------------------------------------------------------ |
| null    | discards the message                                               |
| log     | records what it would have sent, through [`LOG.md`](../log/LOG.md) |

The log adapter is what a development environment binds, since a real push needs
a live service and connected listeners.

---

## Permitted variation

| Variation               | Reason                                          |
| ----------------------- | ----------------------------------------------- |
| which integrations ship | each ecosystem has its own broadcasting clients |

The contract, the message's parts and their names do not vary.
