# Filesystem

The Filesystem component: what it does, how it is used, and the names and
behavior every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the template a port's own component `README.md` starts from: it carries these
headings in this order, so every port's document reads alike and anyone moving
between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Filesystem reads and writes files, wherever they live. It is a **service
component** on the shape in [`LOG.md`](../LOG/LOG.md): one contract, several adapters
behind it, and the application names the one it wants in config.

**The point of the component is that a local disk and a remote store are the same
contract.** Application code that writes a file never learns which it got, so
moving from one to the other is a config change.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the component contract, and one class per adapter          |
| `Data`       | the component config, and one config per adapter           |
| `Enum`       | the visibility values                                      |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

---

## Canonical operations

### Reading and writing

| Operation      | Does                                      |
| -------------- | ----------------------------------------- |
| `exists`       | reports whether a path holds a file       |
| `read`         | reads the whole file                      |
| `write`        | writes a file that must not already exist |
| `update`       | writes a file that must already exist     |
| `put`          | writes a file whether or not it exists    |
| `writeStream`  | the stream form of `write`                |
| `updateStream` | the stream form of `update`               |
| `putStream`    | the stream form of `put`                  |

**The three write operations differ in what they assert about the path.** `write`
requires that nothing is there, `update` requires that something is, and `put`
accepts either. A caller that means "create this" says `write`, so the overwrite
nobody intended fails instead of succeeding.

Warning: an adapter that implements all three as one unconditional write gives a
caller no way to express the difference, and the failure it is meant to catch
passes silently.

**Every write has a stream form.** A whole-file write holds the file in memory,
so anything that might be large uses the stream form. That is why they are
separate operations rather than one that guesses.

### Moving and removing

| Operation   | Does                                 |
| ----------- | ------------------------------------ |
| `rename`    | moves a file to another path         |
| `copy`      | duplicates it to another path        |
| `delete`    | removes one file                     |
| `createDir` | creates a directory                  |
| `deleteDir` | removes a directory and its contents |

### Asking about a file

| Operation      | Reports                                     |
| -------------- | ------------------------------------------- |
| `metadata`     | everything the adapter knows about the path |
| `mimetype`     | the content type                            |
| `size`         | the size in bytes                           |
| `timestamp`    | when it last changed                        |
| `visibility`   | whether it is readable by anyone            |
| `listContents` | what a directory holds                      |

The single-fact operations exist beside `metadata` because an adapter over a
network answers one question with one cheap request and the whole set with an
expensive one.

### Visibility

| Operation              | Does                                         |
| ---------------------- | -------------------------------------------- |
| `setVisibility`        | sets visibility to a value the caller passes |
| `setVisibilityPublic`  | makes it readable by anyone                  |
| `setVisibilityPrivate` | makes it readable only by the application    |

**Visibility is an enum, not a string or a number.** A permission expressed as a
number is a local-disk idea that no remote store shares, so the component states
the intent and each adapter maps it.

The two named forms exist because those are the only two intents, and naming them
removes the chance of passing the wrong value.

---

## Using it

### Configure it

The component config names the **default filesystem**: the adapter the component
contract resolves to. Each adapter gets its own config contract, and an adapter
over a remote store usually has a further config per bucket or location
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Say what you mean about existence

Pick `write`, `update` or `put` by what the caller knows. Reaching for `put`
everywhere turns an overwrite bug into silent data loss.

### Stream anything a visitor supplied

A file whose size the application did not choose is streamed. A whole-file read
of an uploaded file is how one request exhausts the memory of a process serving
many.

### Set visibility explicitly

An adapter's default differs by service. State the intent on write, so the same
code gives the same answer on every adapter.

---

## Permitted variation

| Variation                       | Reason                                               |
| ------------------------------- | ---------------------------------------------------- |
| which integrations ship         | each ecosystem has its own storage clients           |
| how a stream is represented     | each language names its own stream type              |
| which metadata an adapter knows | a remote store answers some questions and not others |

The contract, the operations and their names do not vary.
