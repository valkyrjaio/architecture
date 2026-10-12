# Mail

The Mail component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Mail sends a message to people. It is a **service component** on the shape in
[`LOG.md`](../LOG/LOG.md): one contract, several adapters behind it, and the application
names the one it wants in config.

The component owns the message and the sending. It owns no templating — a body
is text the caller supplies, rendered by [`VIEW.md`](../VIEW/VIEW.md) where the
application wants a template.

---

## Hierarchy

| Subcomponent | Holds                                                                             |
| ------------ | --------------------------------------------------------------------------------- |
| `Mailer`     | the component contract, and one class per adapter                                 |
| `Data`       | the message and its value types, the component config, and one config per adapter |
| `Provider`   | the service provider that binds the contract to an adapter                        |
| `Throwable`  | the component's throwable contract and its exceptions                             |

---

## One operation

The contract declares **`send`**, and nothing else. It takes a message and
reports nothing: a send either completes or throws.

Mail, Sms and Broadcast all declare exactly `send`, because each does one thing
to one message. A component that does one thing names that thing and stops.

---

## The message

A message is **immutable**, like every value type in the framework: a reader
returns the value, and a `with` form returns a copy
([`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md)).

| Holds              | Is                                                    |
| ------------------ | ----------------------------------------------------- |
| the sender         | one recipient, who the message is from                |
| the recipients     | who receives it                                       |
| the reply-to set   | where a reply goes, when that differs from the sender |
| the copy set       | who receives a visible copy                           |
| the blind copy set | who receives a copy the other recipients cannot see   |
| the subject        | the subject line                                      |
| the body           | the body, and whether it is markup                    |
| the plain body     | the text alternative, when the body is markup         |
| the attachments    | the files that ride with it                           |

Each set has an **added** form rather than a setter, so building a message adds
one recipient at a time and never replaces a set by accident.

A **recipient** is an address and an optional display name. An **attachment** is
a path and an optional name. Both are value types, not strings, so an address
with a name is one thing a caller passes rather than two it formats.

**The plain body is separate from the body.** A message whose body is markup
carries the text alternative beside it rather than having a reader strip the
markup, which loses the structure the text wants.

---

## Using it

### Configure it

The component config names the **default mailer**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Build the message, then send it

Build the message through its `with` and added forms, resolve the component
contract, and send. Keep the returned copy from each step — the message does not
mutate.

### Send from a queue, not from a request

A send reaches a third party, so it is slow and it can fail in ways the caller
cannot answer. Push a job and send from the consumer
([`QUEUE.md`](../QUEUE/QUEUE.md)), so a failure retries and a caller never waits.

---

## The adapters every port ships

Two adapters are not integrations, and every port ships both:

| Adapter | Does                                                               |
| ------- | ------------------------------------------------------------------ |
| null    | discards the message                                               |
| log     | records what it would have sent, through [`LOG.md`](../LOG/LOG.md) |

The log adapter is what a development environment binds: a developer reads the
message without anyone receiving it. The null adapter is what a test binds.

A service component that reaches a third party ships both, for the same two
reasons.

---

## Permitted variation

| Variation               | Reason                                      |
| ----------------------- | ------------------------------------------- |
| which integrations ship | each ecosystem has its own mail libraries   |
| how a file is attached  | each language names its own path and stream |

The contract, the message's parts and their names do not vary.
