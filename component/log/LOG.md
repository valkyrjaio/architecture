# Log

The Log component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Log records what happened, at a severity. It is a **service component**: one
contract, several adapters behind it, and the application names the one it wants
in config.

Every other component logs through this one. No component writes to a file or a
stream of its own.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Logger`     | the component contract, and one class per adapter          |
| `Data`       | the component config, and one config per adapter           |
| `Enum`       | the severity levels                                        |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

This is the **service component shape**, shared with Mail, Sms, Broadcast, Cache,
Crypt, Session, Filesystem and Jwt: a role subcomponent named for what the
component does, a `Data` subcomponent for the configs, a provider, and
throwables. The role subcomponent is named for the role, not for the component —
here `Logger`.

---

## Canonical operations

One method per severity, named for the severity, plus two general forms:

| Operation   | Records                                        |
| ----------- | ---------------------------------------------- |
| `debug`     | detail useful only while diagnosing            |
| `info`      | something that happened and is worth a record  |
| `notice`    | something unusual that is not a problem        |
| `warning`   | something that will become a problem           |
| `error`     | a failure that did not stop the application    |
| `critical`  | a failure that stopped part of the application |
| `alert`     | a failure that needs someone now               |
| `emergency` | the application is unusable                    |
| `log`       | a record at a severity the caller passes       |
| `throwable` | a throwable, with its type, message and trace  |

**`throwable` exists so a caller never formats a throwable itself.** Formatting it
in the application loses the trace, or spells it differently in every place that
logs one.

Every operation takes the message and an optional map of context. The context is
data about the record, not part of the message text.

---

## Using it

### Configure it

The component config names the **default logger**: the adapter the component
contract resolves to. Each adapter gets its own config contract for its own
settings, so an application configures the one it uses and constructs nothing for
the others ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Resolve the contract, never an adapter

Resolve the component contract from the container. The provider has already bound
it to the adapter the config names, so application code never names an adapter
and never changes when the adapter does.

### Record at the right severity

Pick the method for the severity. The severity is the only thing that decides
where a record goes and whether anyone is woken, so it is the decision the caller
is actually making.

### Pass context as context

Put the values a reader needs in the context map rather than interpolating them
into the message. A message that stays constant across records can be grouped and
counted; one with an id in its text cannot.

---

## The adapter every port ships

Beyond its integrations, every port ships a **null adapter** that discards every
record. It is what a test binds, and what an application binds to turn logging
off without touching a call site.

A service component that reaches a third party also ships a **log adapter**,
which records what it would have done instead of doing it. That adapter resolves
this component's contract, so the dependency runs from that component to this
one.

---

## Permitted variation

| Variation                    | Reason                                                 |
| ---------------------------- | ------------------------------------------------------ |
| which integrations ship      | each ecosystem has its own logging libraries           |
| a standard logging interface | a port whose ecosystem has one ships an adapter for it |

The contract, the severities and their names do not vary.
