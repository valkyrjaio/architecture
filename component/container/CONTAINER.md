# Container

The Container component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the master copy of this component's README, which every port carries, so that
every port's document for this component reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

The container registers every service in an application and resolves it on
demand. Every other component depends on it, and no component depends on an
application.

The container is the only place a service is constructed. A class never
constructs its own dependency, and a class never reaches for a global.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the container itself, and the child container              |
| `Data`       | the serialized registration set the build tool generates   |
| `Provider`   | the service provider contract, and the component providers |
| `Throwable`  | the component's throwable contract and its exceptions      |

A port adds `Constant` when it needs a binding-key constants file; a language that
names a class natively needs none. A port whose language has attributes or
annotations may also add a segment for the marker that declares a provider
method. The binding-key rule is in
[`CONTAINER_BINDINGS.md`](../../convention/CONTAINER_BINDINGS.md).

---

## Structure

Every class a port builds, where it sits, and what it does.

| Class                                                            | Is                                                     |
| ---------------------------------------------------------------- | ------------------------------------------------------ |
| `Manager/Contract/ContainerContract`                             | the component contract a caller resolves through       |
| `Manager/Contract/ProvidersAwareContract`                        | the surface for registering and asking after providers |
| `Manager/Container`                                              | the container                                          |
| `Manager/ChildContainer`                                         | a container scoped to one unit of work                 |
| `Manager/Trait/ProvidersAware`                                   | the provider bookkeeping both containers share         |
| `Data/ContainerData`                                             | the generated bindings a cached run loads              |
| `Provider/Contract/ServiceProviderContract`                      | the contract every service provider implements         |
| `Provider/ContainerComponentProvider`                            | the component aggregator                               |
| `Provider/ContainerServiceProvider`                              | the component's own bindings                           |
| `Throwable/Contract/ContainerThrowable`                          | the component's throwable contract                     |
| `Throwable/Exception/Abstract/ContainerInvalidArgumentException` | the caller's-failure base                              |
| `Throwable/Exception/Abstract/ContainerRuntimeException`         | the run-time-failure base                              |
| `Throwable/Exception/ContainerCyclicAliasException`              | raised for an alias chain that returns to itself       |
| `Throwable/Exception/ContainerInvalidPublishCallbackException`   | raised for a publisher that cannot be called           |
| `Throwable/Exception/ContainerInvalidReferenceException`         | raised for a service id nothing registered             |

**Two contracts, not one.** `ContainerContract` is what application code
resolves through. `ProvidersAwareContract` is what the application's boot and the
build tool use to register providers and ask which have run. A caller that only
resolves services never sees the second.

**The provider bookkeeping is shared, not inherited.** Both containers register
providers the same way, and neither is the other's parent, so the behavior lives
in something both include rather than in a base one extends.

### Order

| In                        | The order is                                                |
| ------------------------- | ----------------------------------------------------------- |
| the contract's operations | register, then resolve, then inspect                        |
| within register           | the singleton, the service, the alias                       |
| within resolve            | the one for each kind, then the any-kind one                |
| the throwables            | the contract, the two bases, then the concrete ones by name |

The registering operations come before the resolving ones because nothing
resolves before something registers, and a reader of the contract meets the two
in that order.

---

## The three service types

A registration is one of three kinds, and the kind decides the lifetime.

| Kind      | Lifetime                                    | Registered with                 |
| --------- | ------------------------------------------- | ------------------------------- |
| singleton | built once, then the same object every time | `bindSingleton`, `setSingleton` |
| service   | built again for every resolution            | `bind`                          |
| alias     | no object of its own; names another id      | `bindAlias`                     |

A singleton holds two states, and the container tells them apart: a **binding**
that is registered and not yet built, and an **instance** that is built and
cached. The distinction exists so a child container can reuse a built instance
and still defer an unbuilt one.

---

## Using it

### Register a service

Pick the method by the lifetime the service needs. `bindSingleton` for a service
the application shares, `bind` for one the caller needs fresh, `setSingleton` for
an object that is already built, and `bindAlias` to give an existing id a second
name.

A binding takes a **factory**, and the factory receives the container and the
caller's arguments. It constructs the service and returns it, resolving each
dependency from the container it was handed. A factory never reaches for a
global, and it never constructs the container.

### Every service needs a binding

**The container builds nothing that a binding does not describe.** There is no
autowiring, and the container never constructs the class an id happens to name.

So a resolution of an unregistered id **throws**. This is the rule that catches
people out, because a config names a class and the framework resolves that class
through the container: naming a middleware, a listener or a replacement by class
is not registering it. One explicit place states how each service is built.

### Resolve a service

Call the method for the kind you registered: `getSingleton`, `getService`, or
`getAliased`. `get` resolves any kind and costs an extra lookup, so prefer the
specific method on a path that runs for every unit of work.

Resolving a singleton builds it on the first call and returns the cached object
afterward. Resolving a service builds it every time.

### Inspect before acting

The `is*` and `has` operations let an application branch on what is registered.
`has` answers whether an id resolves at all. `isSingletonInstance` answers
whether a singleton is built **without building it**, which is how code reads
live state without forcing construction. `getAliasedId` reports what an alias
names without resolving it.

### Declare a provider rather than registering directly

An application registers through a **service provider**, and the provider
declares one map from binding key to the callback that registers it. The
application lists the provider in its config, and the framework walks the tree
and registers everything it finds
([`APPLICATION.md`](../application/APPLICATION.md)).

A publish callback usually builds the instance and sets it as a singleton. It may
instead bind a factory, when the service has to stay fresh per resolution.

### Know when a callback runs

**Registration is deferred by default.** The container stores the map and runs the
callback for an id on the **first resolution of that id**, through any resolve
operation, and at most once. A provider whose services nobody resolves never
runs.

So the cost of a provider an application does not use is the storage of its map,
and nothing else.

### Export and import

`getData` exports the whole registration set and `setFromData` imports one. This
is how a generated data class replaces the provider walk
([`DATA_CACHE.md`](../../convention/DATA_CACHE.md)), and how a child container
receives its parent's registrations.

---

## Canonical operations

Every port declares these on `ContainerContract`, with no addition and no
omission. The name is the vocabulary. A port spells it in its own convention and
changes nothing else.

| Operation             | Answers                                      |
| --------------------- | -------------------------------------------- |
| `has`                 | is this id registered in any form            |
| `bind`                | register a service                           |
| `bindSingleton`       | register a singleton                         |
| `bindAlias`           | register an alias to another id              |
| `setSingleton`        | register an object that is already built     |
| `isService`           | is this id a service                         |
| `isSingleton`         | is this id a singleton, built or not         |
| `isSingletonBinding`  | is this id a singleton that is not built yet |
| `isSingletonInstance` | is this id a singleton that is built         |
| `isAlias`             | is this id an alias                          |
| `get`                 | resolve an id of any kind                    |
| `getService`          | resolve a service                            |
| `getSingleton`        | resolve a singleton                          |
| `getAliased`          | resolve what an alias names                  |
| `getAliasedId`        | report the id an alias names                 |
| `getData`             | export the registration set                  |
| `setFromData`         | import a registration set                    |

**`get` tries the kinds in order: singleton, then service, then alias.** It passes
the caller's arguments to the service and the alias. It passes none to the
singleton, because a singleton takes none.

So `get` is the wrong operation when the arguments matter. A singleton registered
under that id is returned as it was built, and the arguments are discarded
without a report.

`get` works for every kind and costs an extra lookup. A caller that knows the
kind calls the specific method, because route dispatch runs it on every unit of
work.

**A resolution of an unregistered id fails.** It throws, and it never returns an
empty value. So a documented container read is a claim that the binding exists,
and the claim is checkable.

---

## Provider registration

A **service provider** declares what it registers and nothing else. It carries
no construction logic of its own, and it never registers itself.

A provider declares one map, `publishers`, from binding key to the callback that
registers it. The build tool reads that map statically, so the map is a literal
with no conditional logic ([`PROVIDERS.md`](../../convention/PROVIDERS.md)).

**Deferred registration is the default.** The container records which provider
publishes which id, and it runs that provider's callback on the first resolution
of that id. A provider whose services nobody resolves never runs.

The container tracks this through a second canonical surface, separate from the
operations above because it is about providers rather than about resolving an id:
`register` records a provider's publishers, `isDeferred` answers whether an id is
registered and not yet published, `isPublished` answers whether its callback has
run, and `publish` runs it. Every port declares these four.

**A binding made outside a provider is invisible to the build tool.** The tool
reads the provider tree, so a direct registration cannot reach the generated
cache. An application that registers directly works without the cache and breaks
with it. See [`SINDRI.md`](../../convention/SINDRI.md).

---

## The child container

A persistent runtime keeps one process alive across many units of work, so state
accumulates. A **child container** gives each unit of work its own scope over one
shared parent.

The invariant: a child reads everything the parent holds, and the parent never
sees anything the child registers. So one unit of work cannot leak a service into
the next.

Resolution order is the child first, then the parent. A singleton the child
resolves caches in the child, not the parent, even when the parent holds the
binding. That is what keeps the parent clean, and it means the same unbuilt
binding is built again for each unit of work.

An alias resolves where the id it names resolves, not where the alias is
registered.

---

## Failure modes

The component declares a throwable for each failure a caller can provoke:

- an alias chain that returns to itself
- a resolution of an id that nothing registered
- a publish callback that does not register what it promised

The naming rule and the hierarchy are in
[`THROWABLES.md`](../../convention/THROWABLES.md).

---

## Permitted variation

| Variation                        | Reason                                                |
| -------------------------------- | ----------------------------------------------------- |
| the binding key type             | a language names a class natively or it does not      |
| a `Constant` subcomponent        | only a port whose keys are strings needs one          |
| a provider-method marker segment | only a language with attributes declares one that way |
| the child container's backing    | a port may offer a native variant beside the default  |

Nothing else varies. A port that omits a canonical operation has a gap, not a
deviation.
