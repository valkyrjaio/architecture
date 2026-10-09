# Application

The Application component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

The Application component boots. It reads the application's config, builds the
container, registers every component the config names, and hands control to an
entry point.

It is the only component that knows the application exists. Every other component
is handed what it needs and never reaches upward.

---

## Hierarchy

| Subcomponent | Holds                                                   |
| ------------ | ------------------------------------------------------- |
| `Kernel`     | the application itself, and the child application       |
| `Entry`      | the entry points, one per protocol and per runtime      |
| `Data`       | the config contracts, and their default implementations |
| `Directory`  | the resolver for a path inside the project              |
| `Constant`   | the framework's own version and identity values         |
| `Provider`   | the component provider contract                         |
| `Throwable`  | the component's throwable contract and its exceptions   |

---

## The config is the entry point for everything

An application declares one config class. It is the single place the framework
reads to learn what the application is, and there is **no configuration file
format** — no YAML, no JSON, no INI. The config is a class, so it is typed, and
the build tool reads it statically.

The base config declares the application's identity and environment:

| Setting           | Says                                                |
| ----------------- | --------------------------------------------------- |
| `applicationName` | what the application is called                      |
| `version`         | the application's own version                       |
| `environment`     | which environment this is                           |
| `debugMode`       | whether the application reports detail on a failure |
| `timezone`        | the default timezone                                |
| `key`             | the application secret                              |
| `namespace`       | the application's own source namespace              |
| `dir`             | the application's source directory                  |
| `dataPath`        | where the generated data classes are written        |
| `dataNamespace`   | the namespace those generated classes take          |
| `providers`       | the component providers to register                 |
| `callbacks`       | the publish callbacks to run                        |

**`providers` replaces rather than extends.** A config that declares the list
declares the whole list. A port does not merge an application's list into a
default one, because a silent merge makes the registered set unpredictable.

### One config contract per protocol

Each protocol gets its own config contract, holding the middleware scheduled at
each of its stages — one property per stage, named for the stage. So an
application configures the protocols it uses and constructs nothing for the ones
it does not.

The property name follows the stage name, which makes the pipeline in
[`LIFECYCLE.md`](../convention/LIFECYCLE.md) readable directly from the config. A
protocol with no stage 6 has no property for one.

The wider rule for splitting a config per adapter and per subcomponent is in
[`COMPONENT_CONFIG.md`](../convention/COMPONENT_CONFIG.md).

---

## The provider tree

A **component provider** declares what a component contributes. It names the
component providers it depends on, and then the providers it adds for each kind:
container bindings, event listeners, and the routes for each protocol.

The framework walks that tree from the config, depth first, in declaration order,
and registers everything it finds. A provider already visited is skipped, so a
cycle terminates.

**The order the config declares is the order that is registered.** The walk
imposes no ordering of its own, so an application controls precedence entirely
through its own list.

Each list is a literal with no conditional logic, because the build tool reads it
statically ([`PROVIDERS.md`](../convention/PROVIDERS.md)).

---

## The application surface

Every port declares these on the application, with no addition and no omission:

| Operation                  | Answers                                          |
| -------------------------- | ------------------------------------------------ |
| `getContainer`             | the container this application built             |
| `getProviders`             | every component provider the config named        |
| `getContainerProviders`    | the service providers collected from the tree    |
| `getEventProviders`        | the listener providers collected from the tree   |
| `getHttpProviders`         | the HTTP route providers collected from the tree |
| `getCliProviders`          | the CLI route providers collected from the tree  |
| `publishProviderCallbacks` | run the config's publish callbacks               |
| `getEnvironment`           | which environment this is                        |
| `getDebugMode`             | whether debug mode is on                         |
| `getVersion`               | the application's version                        |

A port adds one reader per protocol it supports, named for that protocol. A port
with gRPC adds the gRPC one; a port without it does not.

---

## Entry points

An entry point is the class a runtime starts. It is named for the runtime that
drives it, and the **framework and an application group them on different axes**.

**In the framework, the adapter is the directory and the protocol is the class
name.** One adapter serves several protocols, so the runtime owns the directory
and each protocol gets a class inside it. The default, in-core entry for each
protocol sits at the top level beside them.

**In an application and in the template, the protocol is the directory and the
runtime is the class name.** The application's axis is the protocol module, which
owns the config, the controllers, the routing data and the providers; only the
server driving them differs. The default entry keeps the bare name.

Warning: never nest an application's entry under the runtime. That yields several
classes with the same name inside one protocol, and the variant axis is not
always an adapter, so it cannot be a directory.

A shared abstract base holds what every entry does, and a per-protocol worker
base holds what a persistent runtime does.

---

## Booting without the cache

**Every port boots with no generated cache.** The provider tree is walkable at
run time, so the framework registers everything by walking it.

The cache is a cold-start optimization. It is required only where a process
handles one unit of work and then exits, because that process pays the whole boot
cost every time. A persistent runtime pays it once.

This is why a provider exposes a class or constructor reference rather than a
string: the framework can walk the tree itself, and the build tool can read the
same declaration statically. See [`DATA_CACHE.md`](../convention/DATA_CACHE.md).

---

## The child application

A persistent runtime keeps one process alive across many units of work. A **child
application** gives each unit of work its own scope over one shared parent, the
same way a child container does.

The boot happens once, in the parent. Each unit of work then gets a child, and
nothing it registers reaches the parent. See
[`CONTAINER.md`](CONTAINER.md).

---

## Permitted variation

| Variation                    | Reason                                              |
| ---------------------------- | --------------------------------------------------- |
| which runtimes have an entry | each ecosystem has its own servers                  |
| which protocol readers exist | a port declares one per protocol it has             |
| how a path is resolved       | each language names its own path separator and root |

Nothing else varies. The config surface and the provider walk are identical in
every port.
