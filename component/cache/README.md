# Cache

Implements
[`CACHE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/cache/CACHE.md).

## Introduction

The Cache component stores a value under a key so a later read does not
recompute it. It keeps values in a store the application configures, and your
code reads the same whichever store that is.

An **adapter** is the thing that stores. Each one implements `CacheContract`, and
each is registered under its own name, so an application resolves whichever one
it wants.

`CacheContract` itself resolves to the **default** adapter: the one your config
names. It is the convenience, and it is what lets a development environment keep
values in memory while production keeps them in a shared store, with no call site
changing.

A cached value is **disposable**. The component gives no promise that a value put
in is still there, so anything the application cannot lose belongs in a store
rather than here.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a cache

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                |
| -------------------------- | ----------------------------------------- | ------------------------ |
| `CacheContract`            | the adapter the config names              | environment              |
| an adapter by its own name | that adapter, always                      | nothing                  |
| a contract of your own     | the adapter that contract's binding names | purpose, and environment |

### The default, through `CacheContract`

```text
cache = container.getSingleton(CacheContract)

cache.put("orders.4821", payload, 300)
```

### A contract of your own, per cached thing

Declare a contract per kind of cached value, and bind each to whichever adapter
that kind wants. Then a short-lived fragment and a long-lived lookup table live
in different stores without a call site knowing.

```text
interface FragmentCacheContract extends CacheContract { }
interface LookupCacheContract extends CacheContract { }
```

```text
class AppCacheServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            FragmentCacheContract: [self, "publishFragmentCache"],
            LookupCacheContract: [self, "publishLookupCache"],
        }
    }

    static publishFragmentCache(container) {
        container.setSingleton(
            FragmentCacheContract,
            container.getSingleton(RedisCache),
        )
    }

    static publishLookupCache(container) {
        container.setSingleton(
            LookupCacheContract,
            container.getSingleton(NullCache),
        )
    }
}
```

## Reading and writing

| Method      | Does                                          | Takes                   | Returns         |
| ----------- | --------------------------------------------- | ----------------------- | --------------- |
| `has`       | answers whether a key holds a value           | a key                   | a boolean       |
| `get`       | reads the value under a key                   | a key                   | the value       |
| `many`      | reads the values under several keys           | the keys                | the values      |
| `put`       | stores a value for a number of seconds        | a key, a value, seconds | nothing         |
| `putMany`   | stores several values for a number of seconds | the pairs, seconds      | nothing         |
| `forever`   | stores a value with no expiry                 | a key, a value          | nothing         |
| `increment` | raises a stored number                        | a key, and how much     | the new value   |
| `decrement` | lowers a stored number                        | a key, and how much     | the new value   |
| `forget`    | removes one key                               | a key                   | whether it went |
| `flush`     | removes everything                            | nothing                 | whether it went |
| `getPrefix` | returns what every key is prefixed with       | nothing                 | the prefix      |
| `getTagger` | returns a tagger for a set of tags            | the tags                | a tagger        |

```text
// Store for five minutes.
cache.put("orders.4821", payload, 300)

// Read it back. Check first where a miss matters.
if (cache.has("orders.4821")) {
    payload = cache.get("orders.4821")
}

// Store with no expiry, for something that changes only when you change it.
cache.forever("countries", payload)

// Count something, without reading and writing it yourself.
cache.increment("orders.placed.today")
```

### Read and write in batches where you can

`many` and `putMany` make one call to the store where the single-key methods make
one per key. A page that reads twenty fragments reads them in one round trip.

```text
payloads = cache.many("orders.4821", "orders.4822", "orders.4823")

cache.putMany({ "a": first, "b": second }, 300)
```

### Every key carries a prefix

The adapter prefixes every key it stores, so two applications sharing one store
do not collide. `getPrefix` returns it, and your keys do not include it.

## Tags

A tag groups keys so an application clears a set of them together. Ask the cache
for a tagger over the tags you mean, and use the tagger as you would the cache.

```text
tagger = cache.getTagger("orders", "customer.4821")

tagger.put("orders.4821", payload, 300)

// Later, clear everything tagged for that customer.
cache.getTagger("customer.4821").flush()
```

| Method                          | Does                                  |
| ------------------------------- | ------------------------------------- |
| `getTags`                       | returns the tags this tagger holds    |
| `tag` / `untag`                 | adds or removes one key from the tags |
| `tagMany` / `untagMany`         | adds or removes several               |
| the reading and writing methods | as on the cache, scoped to the tags   |

A tagger carries the same reading and writing methods as the cache, so code that
takes a cache takes a tagger.

**An adapter whose store cannot tag has no tagger.** The capability is the
store's, so a store without it reports that rather than pretending.

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property       | Holds                                   | Default              |
| -------------- | --------------------------------------- | -------------------- |
| `defaultCache` | the adapter `CacheContract` resolves to | the port's own store |

### Changing the default adapter

```text
class Config implements ConfigContract, CacheConfigContract {
    defaultCache = NullCache
}
```

### Available adapters

| Adapter        | Stores in                               |
| -------------- | --------------------------------------- |
| `<Store>Cache` | one store the port speaks to            |
| `LogCache`     | nothing; it records what it would store |
| `NullCache`    | nothing; every read is a miss           |

Every port ships the log and null adapters. Which stores ship is the port's own
choice.

### Configuring a store adapter

Each store adapter carries the settings that store needs, prefixed with the
store's name.

| Property        | Holds                           | Default         |
| --------------- | ------------------------------- | --------------- |
| `<store>Host`   | where to reach the store        | the local host  |
| `<store>Port`   | the port to reach it on         | the store's own |
| `<store>Prefix` | what every key is prefixed with | the application |

```text
class Config implements ConfigContract, CacheConfigContract, CacheRedisConfigContract {
    defaultCache = RedisCache

    redisHost = "cache.internal"
    redisPort = 6379
    redisPrefix = "orders"
}
```

### Configuring the log adapter

| Property    | Holds                           | Default            |
| ----------- | ------------------------------- | ------------------ |
| `logLogger` | which logger the records go to  | the default logger |
| `logPrefix` | what every key is prefixed with | the application    |

The log adapter names a logger rather than holding one, so the records go
wherever that logger goes ([Log](../log/README.md)).

### Configuring the null adapter

| Property     | Holds                           | Default         |
| ------------ | ------------------------------- | --------------- |
| `nullPrefix` | what every key is prefixed with | the application |

It stores nothing, so the prefix is all there is to set. Name it as the default
in a test that should not share state between cases.

## Adding an adapter

An adapter implements `CacheContract`. Where this port ships a base the adapters
share, extend that instead.

```text
class MemcachedCache implements CacheContract {

    construct(client: MemcachedClient, config: CacheMemcachedConfigContract) { }

    get(key) {
        return this.client.fetch(this.prefixed(key))
    }

    put(key, value, seconds) {
        this.client.store(this.prefixed(key), value, seconds)
    }

    // ...one method per operation

    prefixed(key) {
        return this.config.memcachedPrefix + "." + key
    }
}
```

Give it a config contract of its own, prefixing each property with the adapter's
name so one config class implements several contracts at once:

```text
interface CacheMemcachedConfigContract {
    memcachedHost: string
    memcachedPort: int
    memcachedPrefix: string
}
```

Register the adapter and its config in a service provider, then name the adapter
in `defaultCache`, or resolve it directly where you want it:

```text
class AppCacheServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            MemcachedCache: [self, "publishMemcachedCache"],
        }
    }

    static publishMemcachedCache(container) {
        container.setSingleton(MemcachedCache, new MemcachedCache(
            container.getSingleton(MemcachedClient),
            container.getSingleton(CacheMemcachedConfigContract),
        ))
    }
}
```

A store that cannot tag returns no tagger. Implement the tagger only where the
store supports it.

## Service registration

The component's provider binds:

| Binding              | Is                                            |
| -------------------- | --------------------------------------------- |
| `CacheContract`      | the adapter the component config names        |
| each adapter         | bound under its own name, for an explicit ask |
| the component config | your application's own config class           |
| each adapter config  | your application's own config class           |

```text
class CacheServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            CacheConfigContract: [self, "publishConfig"],
            CacheRedisConfigContract: [self, "publishRedisConfig"],
            CacheContract: [self, "publishCache"],
            RedisCache: [self, "publishRedisCache"],
            LogCache: [self, "publishLogCache"],
            NullCache: [self, "publishNullCache"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements CacheConfigContract) {
            container.setSingleton(CacheConfigContract, config)

            return
        }

        container.setSingleton(CacheConfigContract, new CacheConfig())
    }

    static publishCache(container) {
        config = container.getSingleton(CacheConfigContract)

        container.setSingleton(
            CacheContract,
            container.getSingleton(config.defaultCache),
        )
    }

    static publishRedisCache(container) {
        container.setSingleton(RedisCache, new RedisCache(
            container.getSingleton(RedisClient),
            container.getSingleton(CacheRedisConfigContract),
        ))
    }
}
```
