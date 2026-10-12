# Cache

The Cache component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Cache stores a value under a key so a later read does not recompute it. It is a
**service component** on the shape in [`LOG.md`](../LOG/LOG.md): one contract, several
adapters behind it, and the application names the one it wants in config.

A cached value is **disposable by definition**. Anything the application cannot
lose belongs in a store that is not a cache.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the component contract, and one class per adapter          |
| `Tagger`     | the tagged view of the same store                          |
| `Data`       | the component config, and one config per adapter           |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

---

## Canonical operations

| Operation   | Does                                               |
| ----------- | -------------------------------------------------- |
| `has`       | reports whether a key holds a value                |
| `get`       | reads one value                                    |
| `many`      | reads several values in one call                   |
| `put`       | writes one value, with a lifetime                  |
| `putMany`   | writes several values in one call                  |
| `forever`   | writes a value with no lifetime                    |
| `increment` | raises a stored number, atomically                 |
| `decrement` | lowers a stored number, atomically                 |
| `forget`    | removes one key                                    |
| `flush`     | removes everything                                 |
| `getPrefix` | reports the prefix every key in this store carries |
| `getTagger` | returns the tagged view for a set of tags          |

**The many-forms exist because a round trip costs more than a read.** A caller
reading ten keys calls `many` once rather than `get` ten times, and an adapter
over a network turns that into one request.

**`increment` and `decrement` are atomic.** A read, an addition and a write are
three operations and two callers interleave between them, so a counter built that
way loses increments. These two are the reason the component declares them at all.

**`forever` is a separate operation, not a lifetime value.** A caller that means
"no expiry" says so, rather than passing a number that happens to be large.

---

## The prefix

Every key a store writes carries a prefix, and the store reports it. The prefix
is what lets two applications, or two environments, share one cache service
without reading each other's values.

A caller never applies the prefix itself. It passes the bare key and the store
applies it, so a key is spelled one way everywhere.

---

## Tags

A **tagger** is the same store seen through a set of tags. It declares the same
read and write operations, so code that holds one needs to know nothing else, and
it adds the operations that make a tag worth having:

| Operation             | Does                               |
| --------------------- | ---------------------------------- |
| `getTags`             | reports the tags this view carries |
| `tag` / `tagMany`     | adds a key to this view's tags     |
| `untag` / `untagMany` | removes a key from them            |

**A tag exists so a group of keys can be dropped together.** Flushing a tagged
view removes the keys that carry its tags and leaves the rest, which is what a
caller wants when one entity changes and twenty derived values go stale.

---

## Using it

### Configure it

The component config names the **default store**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Read through the cache, do not read around it

Ask the cache, and compute on a miss. Code that computes first and writes second
pays the cost every time, which is the mistake the component exists to prevent.

### Give every value a lifetime, unless it cannot change

A lifetime is the only thing that bounds how wrong a cached value can be. Use
`forever` where the value is derived from something immutable, and a lifetime
everywhere else.

### Never cache what you cannot lose

Every adapter may evict under pressure and a flush removes everything. A value
the application needs after a restart belongs in a real store.

---

## The adapter every port ships

Every port ships a **null adapter** that stores nothing and reports every read as
a miss. It is what a test binds, and what an application binds to measure the
cost of running uncached.

It also ships a **log adapter**, which records each operation instead of storing
anything ([`LOG.md`](../LOG/LOG.md)), so a developer can see what an application caches
without a cache service running.

---

## Permitted variation

| Variation               | Reason                                                |
| ----------------------- | ----------------------------------------------------- |
| which integrations ship | each ecosystem has its own cache clients              |
| whether tags are native | an adapter without them emulates tags over plain keys |

The contract, the operations and their names do not vary.
