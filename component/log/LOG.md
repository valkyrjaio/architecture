# Log

The Log component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the master copy of this component's README, which every port carries, so that
every port's document for this component reads alike.

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

This is the **service component shape**: a role subcomponent, a `Data`
subcomponent for the configs, a provider, and throwables. The role subcomponent
is named for the role rather than for the component, so here it is `Logger`.

---

## Structure

Every class a port builds, where it sits, and what it does.

| Class                                                      | Is                                                                        |
| ---------------------------------------------------------- | ------------------------------------------------------------------------- |
| `Logger/Contract/LoggerContract`                           | the component contract: one method per severity, `log`, `throwable`       |
| `Logger/Abstract/Logger`                                   | a base the adapters share, where the port wants one                       |
| `Logger/NullLogger`                                        | the adapter that discards every record                                    |
| `Logger/<Integration>Logger`                               | one class per integration the port ships                                  |
| `Data/Contract/LogConfigContract`                          | the component config contract: the default adapter                        |
| `Data/LogConfig`                                           | the component config a port ships, used where the application writes none |
| `Data/Contract/Log<Adapter>ConfigContract`                 | one per adapter that has settings                                         |
| `Enum/LogLevel`                                            | the severities, each carrying its own name as its value                   |
| `Provider/LogComponentProvider`                            | the component aggregator                                                  |
| `Provider/LogServiceProvider`                              | the container bindings                                                    |
| `Throwable/Contract/LogThrowable`                          | the component's throwable contract                                        |
| `Throwable/Exception/Abstract/LogInvalidArgumentException` | the invalid-argument base                                                 |
| `Throwable/Exception/Abstract/LogRuntimeException`         | the runtime base                                                          |
| `Throwable/Exception/LogInvalidLogLevelException`          | raised for a severity outside the closed set                              |

### Order

A port declares these in one order, so the same file reads alike in every port.

| In                     | The order is                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------- |
| the component contract | the severities from lowest to highest, then `log`, then `throwable`                         |
| the severity values    | lowest to highest                                                                           |
| the provider's map     | the config contracts, the component contract, the adapters, then what an adapter depends on |
| the adapters           | each integration, then file, then in-memory, then null                                      |

### `Throwable/` carries two bases and one exception

The two abstract bases exist so that a caller catches the component's failures by
kind: one for an argument the caller got wrong, one for a failure at run time.
Every concrete exception extends one of them, and both implement
`LogThrowable`.

`LogInvalidLogLevelException` is the only concrete exception the contract
requires. A port adds one per failure its own integrations can raise, such as a
write that fails.

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

**`throwable` exists so a caller does not need to format a throwable itself.**
Formatting it in the application loses the trace, or spells it differently in
every place that logs one.

Every operation takes the message and an optional map of context. The context is
data about the record, not part of the message text.

**`log` validates the severity** against the closed set, and raises
`LogInvalidLogLevelException` for a value outside it. Where that validation lives
is the port's own business; that it happens for every adapter is not.

---

## Using it

### Configure it

The component config names the **default logger**: the adapter the component
contract resolves to. Each adapter gets its own config contract for its own
settings, so an application configures the one it uses and constructs nothing for
the others ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

**The provider registers the contract as a singleton holding the adapter it
resolved. It does not alias the contract to the adapter.** It reads the named
adapter from the config, resolves that adapter's own registration, and registers
the result under the contract.

The two forms differ where it is observable. A singleton holds the instance from
the moment the provider runs, so a later registration under the adapter's name
leaves the contract on the instance it already holds. An alias resolves on each
call, so the contract would follow the later registration.

### Resolve the contract for the default, an adapter for a specific one

Resolve the component contract where the configured adapter is the one you want.
The provider has bound it to the adapter the config names, so a config change
moves every such call site at once.

Resolve an adapter by its own name where you want that adapter and not the
default. The provider binds each one for that reason, and it is how an
application uses more than one at a time.

### Record at the right severity

Pick the method for the severity. The severity is the only thing that decides
where a record goes and whether anyone is woken, so it is the decision the caller
is actually making.

### Pass context as context

Put the values a reader needs in the context map rather than interpolating them
into the message. A message that stays constant across records can be grouped and
counted; one with an id in its text cannot.

---

## What a record looks like

**The same call writes the same record in every port.** A record is not a port's
own format with the same parts in it; it is one format, and a tool that parses
one port's output parses every port's.

```text
[2026-10-12T03:15:22+00:00] orders.ERROR: Payment declined {"orderId":4821}
```

The parts, in order:

| Part         | Is                                                        |
| ------------ | --------------------------------------------------------- |
| a timestamp  | in brackets, first, as ISO 8601 with the offset           |
| the channel  | the name the records are written under, then a dot        |
| the severity | the severity's name in upper case, then a colon and space |
| the message  | as the caller passed it                                   |
| the context  | as JSON, after the message, and only where it has entries |

**One record is one line, and the host's own line ending ends it.** A port uses
the line ending the platform it runs on uses, because the file is read on that
platform. A record whose message carries its own line breaks still ends with
one, so the next record starts on a line of its own.

The message comes before the context so that a reader greps the message. A
record with no context ends at the message, so an empty map adds nothing to
read.

**The context is JSON**, not the way the language prints a map. JSON is the one
rendering every port writes alike, so a reader and a parser both see one shape.

## Where a record goes

The file's name has two parts: a **name**, and the extension for a log. The name
defaults to the application's own name and the current date, so a run writes
beside the days around it rather than into one growing file.

The directory defaults to the framework's own logs location. The config sets
either part ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

## What `throwable` records

**`throwable` records at `error`**, and it puts the throwable in the message
rather than in the context. The caller's context passes through as it was given.

The message's parts, in order:

| Part                    | Is                                              |
| ----------------------- | ----------------------------------------------- |
| the trace code          | the identifier a reader quotes when reporting   |
| the throwable's message | what the throwable itself says                  |
| the caller's message    | what the call site said about it                |
| the stack trace         | last, and fenced so a reader sees where it ends |

Each part is on its own line and names itself, so a reader finds the caller's
message without reading the trace. The trace goes last because it is the longest
part, and a reader who wants the first three should not scroll past it.

## What each provider holds

| Provider               | Holds                                                  |
| ---------------------- | ------------------------------------------------------ |
| `LogComponentProvider` | the component's own service provider, and nothing else |
| `LogServiceProvider`   | every container binding the component makes            |

The component provider declares the component's dependencies on other
components, and the leaf providers for each concern. Log depends on no other
component, and it registers no route, no command and no listener, so every list
it returns is empty except the one naming `LogServiceProvider`.

The service provider's map is ordered, and the order is in **Order** above. It
reads from the most general binding to the most specific: the config contracts
the component reads, the component contract a caller resolves, the adapters that
contract chooses between, then anything an adapter needs built for it.

## The adapters every port ships

Three, whatever else the port adds.

| Adapter   | Records                                        | Is for                            |
| --------- | ---------------------------------------------- | --------------------------------- |
| null      | nothing; every record is discarded             | turning logging off               |
| file      | a line per record, to a file on disk           | a port with nothing else wired up |
| in-memory | each record in a list the caller can read back | asserting on records in a test    |

The **file** adapter takes no dependency outside the language, so a port is
useful before an application chooses a logging library. A port whose ecosystem
interface already writes files satisfies this through that adapter, as PHP does
through its standard-interface adapter.

The **in-memory** adapter is what makes a record assertable. A test binds it,
makes the call, and reads the list, rather than reading a file or trusting that
nothing threw.

A port whose ecosystem has a standard logging interface ships an adapter for it
as well. Which further integrations ship is the port's own choice.

Log takes no dependency on another component for any of these. Every component
logs, so an adapter that reached one would invert the dependency, and a failure
in that component would have nowhere to report itself.

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
