# Container

The **cross-language** definition of the Container component. It states the
hierarchy, the names and the behavior that every port implements.

A port implements this document. A port does not redefine it. The reference
implementation is PHP ([`AGENTS.md`](../AGENTS.md) §1).

This document holds no code example. The component's `README.md` in each port
holds the examples and the per-language spelling.

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

A port adds `Constant` when it needs a binding-key constants file. PHP and Java
name a class natively, so neither needs one. The rule is in
[`CONTAINER_BINDINGS.md`](../convention/CONTAINER_BINDINGS.md).

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
with no conditional logic ([`PROVIDERS.md`](../convention/PROVIDERS.md)).

**Deferred registration is the default.** The container records which provider
publishes which id, and it runs that provider's callback on the first resolution
of that id. A provider whose services nobody resolves never runs. The container
tracks this through `register`, `isDeferred`, `isPublished` and `publish`.

**A binding made outside a provider is invisible to the build tool.** The tool
reads the provider tree, so a direct registration cannot reach the generated
cache. An application that registers directly works without the cache and breaks
with it. See [`BUILD_TOOL.md`](../convention/BUILD_TOOL.md).

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
[`THROWABLES.md`](../convention/THROWABLES.md).

---

## Permitted variation

| Variation                     | Reason                                               |
| ----------------------------- | ---------------------------------------------------- |
| the binding key type          | a language names a class natively or it does not     |
| a `Constant` subcomponent     | only a port whose keys are strings needs one         |
| the child container's backing | a port may offer a native variant beside the default |

Nothing else varies. A port that omits a canonical operation has a gap, not a
deviation.
