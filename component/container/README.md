# Container

Implements
[`CONTAINER.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/container/CONTAINER.md).

## Introduction

The Container component holds the application's services and hands them out. A
service is registered under an **id** and resolved by that id, so nothing
constructs its own dependencies and nothing knows which class satisfies a
contract.

Every other component resolves through this one. A component's provider registers
what the component offers, and application code resolves a contract rather than
naming a class.

The container does not guess. It performs no reflection and no autowiring, so a
service with no registration is a failure rather than something the container
assembles. Every service needs a binding.

This document is the reference the other components point at for contracts,
bindings and resolution.

## A contract is an id

A contract is a resolvable set of methods, independent of the class that
satisfies it. Register a class under a contract's id, and resolution by that id
returns it.

```text
container.bindSingleton(LoggerContract, [self, "publishLogger"])

logger = container.getSingleton(LoggerContract)
```

An id is whatever the language can use as a stable key. A language with a class
reference uses the class itself; a language without one uses a constant holding a
string or a symbol, which is why some ports carry a service-id holder per
component and others do not.

**Declare a contract of your own wherever you want a separate binding.** A
contract that extends another is still its own id, so two contracts over the same
methods resolve to different classes.

```text
interface UserLoggerContract extends LoggerContract { }
interface SystemLoggerContract extends LoggerContract { }
```

## The three kinds of registration

The kind decides the lifetime, and the kind you register under decides the
operation you resolve with.

| Kind      | Built                                 | Registered with                    |
| --------- | ------------------------------------- | ---------------------------------- |
| service   | on every resolve                      | `bind`                             |
| singleton | once, then the same object every time | `bindSingleton`, or `setSingleton` |
| alias     | never; it forwards to another id      | `bindAlias`                        |

```text
// A service. Built for each call, and it takes the caller's arguments.
container.bind(ReportContract, [self, "publishReport"])

// A singleton, built on the first resolve.
container.bindSingleton(LoggerContract, [self, "publishLogger"])

// A singleton you already hold.
container.setSingleton(ConfigContract, config)

// An alias. Resolving the alias resolves the id it names.
container.bindAlias(ShortName, LongFullyQualifiedName)
```

### `bindSingleton` defers, `setSingleton` does not

`bindSingleton` registers a callable and runs it on the first resolve, so nothing
is built until something asks. `setSingleton` registers an object you already
have.

Prefer `bindSingleton` for anything with work in its construction, so an
application that never resolves it never pays for it.

## Resolving

| Method         | Returns                                 | Takes                |
| -------------- | --------------------------------------- | -------------------- |
| `getSingleton` | the singleton under an id               | an id                |
| `getService`   | a service, built for this call          | an id, and arguments |
| `getAliased`   | what an alias forwards to               | an id, and arguments |
| `get`          | whichever kind is registered            | an id, and arguments |
| `getAliasedId` | the id an alias forwards to, or nothing | an alias             |

```text
// Name the kind you registered.
logger = container.getSingleton(LoggerContract)
report = container.getService(ReportContract, { month: 10 })
```

### Prefer the specific method

`get` tries the kinds in order — singleton, then service, then alias — and costs
an extra lookup for each one it tries. On a path that runs for every unit of
work, name the kind.

`get` also passes no arguments to the singleton path. A singleton registered
under the id comes back as it was built, and the arguments are discarded without
a report.

```text
// Wrong where the arguments matter and a singleton holds the id.
type = container.get(cast.type, { value: raw })

// Right. The service path takes the arguments and builds for this call.
type = container.getService(cast.type, { value: raw })
```

## Inspecting

| Method                | Answers                                    |
| --------------------- | ------------------------------------------ |
| `has`                 | whether anything is registered under an id |
| `isSingleton`         | whether the id holds a singleton           |
| `isSingletonBinding`  | whether it holds one not yet built         |
| `isSingletonInstance` | whether it holds one already built         |
| `isService`           | whether the id holds a service             |
| `isAlias`             | whether the id holds an alias              |

```text
if (! container.has(ReportContract)) {
    return response.notImplemented()
}
```

**Ask before acting where a binding is optional.** An application that may or may
not have registered something branches on `has` rather than catching a failure.

The two singleton questions exist because a deferred singleton and a built one
are both singletons. A caller that wants to know whether construction has already
happened asks `isSingletonInstance`.

## Providers

A provider is how a component registers what it offers. Application code does not
call `bind` directly; a provider declares the bindings and the container runs it.

A service provider declares `publishers`, returning a map of **service id to a
named method reference**, and one static publisher method per entry.

```text
class LogServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            LogConfigContract: [self, "publishConfig"],
            LoggerContract: [self, "publishLogger"],
            StandardLogger: [self, "publishStandardLogger"],
        }
    }

    static publishLogger(container) {
        config = container.getSingleton(LogConfigContract)

        container.setSingleton(
            LoggerContract,
            container.getSingleton(config.defaultLogger),
        )
    }
}
```

**The reference is named rather than inline** because the build tool reads the map
as written, without running it. An inline closure cannot be read from a syntax
tree, so a cached run would lose the binding.

### Registration is deferred until something resolves

`register` takes a provider and records which ids it publishes. Nothing runs
until something resolves one of those ids, and then the publisher for that id
runs once.

| Method        | Does                                              |
| ------------- | ------------------------------------------------- |
| `register`    | records a provider's publishers                   |
| `isDeferred`  | answers whether an id has a publisher not yet run |
| `isPublished` | answers whether an id's publisher has run         |
| `publish`     | runs an id's publisher now                        |

```text
container.register(new LogServiceProvider())

// Nothing has been built yet.
container.isDeferred(LoggerContract)   // true
container.isPublished(LoggerContract)  // false

// Resolving runs the publisher.
logger = container.getSingleton(LoggerContract)

container.isPublished(LoggerContract)  // true
```

An application pays for the components it uses rather than for every component it
registered.

## The child container

A child container is scoped to one unit of work. It resolves what the parent
holds, and anything registered on the child goes away when the unit of work does.

It is constructed with the parent and the registrations, rather than asked for
from the parent:

```text
child = new ChildContainer(container, data)

child.setSingleton(ServerRequestContract, request)

// Resolves from the child.
request = child.getSingleton(ServerRequestContract)

// Resolves from the parent.
logger = child.getSingleton(LoggerContract)
```

**A worker entry point makes one per request.** A server that stays up between
requests holds one parent container for the process, and gives each request a
child, so a per-request value never reaches the next request.

An entry point that serves one request per process needs none of this: the
process is the scope.

## Exporting and importing

| Method        | Does                              |
| ------------- | --------------------------------- |
| `getData`     | returns the registrations as data |
| `setFromData` | loads registrations from data     |

The build tool reads the providers, writes the result as data, and a cached run
loads it with `setFromData` rather than reading providers again. The data holds
each reference in the shape the provider declared it, which is what makes a
cached run and an uncached run resolve alike.

```text
// What the build tool writes.
data = container.getData()

// What a cached run loads.
container.setFromData(data)
```

## Failures

| Failure                 | Raised when                                 |
| ----------------------- | ------------------------------------------- |
| an unregistered id      | nothing is registered under the id resolved |
| a cyclic alias          | an alias chain returns to itself            |
| an uncallable publisher | a publisher reference cannot be called      |

Each is a failure rather than a fallback, so a missing registration surfaces where
it happened rather than as a null further on.

## Service registration

The component's provider binds the container itself, so application code resolves
the container the same way it resolves anything else.

| Binding                  | Is                                          |
| ------------------------ | ------------------------------------------- |
| `ContainerContract`      | the container                               |
| `ProvidersAwareContract` | the same container, under its other surface |

**Two contracts, one object.** `ContainerContract` is what application code
resolves through. `ProvidersAwareContract` is what the application's boot and the
build tool use to register providers and ask which have run. A caller that only
resolves services never sees the second.
