# Sms

The Sms component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Sms sends a text message to a phone number. It is a **service component** on the
shape in [`LOG.md`](../log/LOG.md): one contract, several adapters behind it, and the
application names the one it wants in config.

---

## Hierarchy

| Subcomponent | Holds                                                         |
| ------------ | ------------------------------------------------------------- |
| `Messenger`  | the component contract, and one class per adapter             |
| `Data`       | the message, the component config, and one config per adapter |
| `Provider`   | the service provider that binds the contract to an adapter    |
| `Throwable`  | the component's throwable contract and its exceptions         |

The role subcomponent is `Messenger`, named for the role rather than the
component, the same way Mail's is `Mailer` and Log's is `Logger`.

---

## One operation

The contract declares **`send`**, and nothing else. It takes a message and
reports nothing: a send either completes or throws. Mail and Broadcast declare
the same one operation, for the same reason.

---

## The message

A message is **immutable**: a reader returns the value, and a `with` form returns
a copy ([`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md)).

| Holds        | Is                                             |
| ------------ | ---------------------------------------------- |
| the sender   | the number or sender id the message comes from |
| the receiver | the number it goes to                          |
| the text     | the body                                       |
| the encoding | whether the text needs the wide character set  |

**The encoding is a property of the message, not of the adapter.** A text outside
the basic character set costs more per message and splits differently, so the
caller states it and every adapter honors the same answer.

---

## Using it

### Configure it

The component config names the **default messenger**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Build the message, then send it

Build the message through its `with` forms, resolve the component contract, and
send. Keep the returned copy from each step — the message does not mutate.

### Send from a queue, not from a request

A send reaches a third party and costs money per message. Push a job and send
from the consumer ([`QUEUE.md`](../queue/QUEUE.md)), so a failure retries once rather than
being retried by a person pressing a button again.

---

## The adapters every port ships

| Adapter | Does                                                               |
| ------- | ------------------------------------------------------------------ |
| null    | discards the message                                               |
| log     | records what it would have sent, through [`LOG.md`](../log/LOG.md) |

The log adapter is what a development environment binds, because a real send
reaches a real phone and costs money. The null adapter is what a test binds.

---

## Permitted variation

| Variation               | Reason                                     |
| ----------------------- | ------------------------------------------ |
| which integrations ship | each ecosystem has its own gateway clients |

The contract, the message's parts and their names do not vary.
