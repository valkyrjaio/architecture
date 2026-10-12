# Session

The Session component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

A session carries state for one visitor across several units of work. It is a
**service component** on the shape in [`LOG.md`](../LOG/LOG.md): one contract, several
adapters behind it, and the application names the one it wants in config.

The component also owns the **cross-site request forgery token**, because that
token is session state and nothing else has the session to put it in.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the component contract, and one class per adapter          |
| `Data`       | the component config, and one config per adapter           |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

---

## The session is mutable

**A session is the one piece of request state that is not immutable.** Every
message type in the framework returns a copy from a `with` form; a session
modifies itself through `set` and `remove`.

That is deliberate. `set` modifies the host and `with` returns a copy, which is
what each name promises. A session is a single thing shared by everything
handling this unit of work, so a copy would be the wrong shape. Two holders
would write into two sessions, and one would lose.

---

## Canonical operations

| Operation             | Does                                            |
| --------------------- | ----------------------------------------------- |
| `start`               | begins the session, reading whatever was stored |
| `isActive`            | reports whether it has started                  |
| `getId` / `setId`     | reads and replaces the session identifier       |
| `getName` / `setName` | reads and replaces the session's name           |
| `has`                 | reports whether a key holds a value             |
| `get`                 | reads one value                                 |
| `set`                 | writes one value                                |
| `remove`              | removes one key                                 |
| `all`                 | reads everything the session holds              |
| `clear`               | removes every value, and keeps the session      |
| `destroy`             | ends the session and discards it                |

**`clear` and `destroy` differ, and the difference matters.** Clearing empties a
session the visitor still has; destroying ends it. A sign-out destroys, so the
identifier the visitor holds stops naming anything.

### The forgery token

| Operation           | Does                                                        |
| ------------------- | ----------------------------------------------------------- |
| `generateCsrfToken` | creates a token for a form, and stores it in the session    |
| `validateCsrfToken` | checks a submitted token, and fails loudly when it is wrong |
| `isValidCsrfToken`  | checks a submitted token, and answers with a boolean        |

**Two forms exist because the caller's intent differs.** A middleware that
rejects a bad request wants the failing form; code that chooses between two paths
wants the boolean. The naming rule makes which is which readable without the
documentation ([`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md)).

---

## Using it

### Configure it

The component config names the **default session**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Start before reading

A session reads its stored state when it starts. Reading before that answers from
an empty session rather than failing, which is the harder bug of the two.

### Put identifiers in the session, not the data

Store what identifies the visitor and read the rest from a real store. A session
is copied to wherever it is stored on every unit of work, so a large session is a
cost paid on every request.

### Regenerate the identifier when the trust level changes

Replace the identifier on sign-in, so a token captured before it cannot be used
after. This is the one reason the component exposes a setter for the identifier
at all.

### Destroy on sign-out

Destroy rather than clear, so the identifier the visitor holds stops naming a
session.

---

## The adapters every port ships

| Adapter | Does                                                      |
| ------- | --------------------------------------------------------- |
| null    | holds nothing, so every read reports an empty session     |
| log     | records each operation, through [`LOG.md`](../LOG/LOG.md) |

---

## Permitted variation

| Variation                       | Reason                                                     |
| ------------------------------- | ---------------------------------------------------------- |
| which integrations ship         | each ecosystem has its own session stores                  |
| whether a runtime owns sessions | a language with a built-in session gets an adapter over it |

The contract, the operations and their names do not vary.
