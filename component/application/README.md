# Application

Implements
[`APPLICATION.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/application/APPLICATION.md).

## Introduction

The Application component boots. It reads the application's config, builds the
container, registers every component the config names, and hands control to an
entry point.

It is the component the application object lives in. Every other component
receives the application or resolves it, and holds no reference of its own.

The framework requires two classes of the application. It requires one config
class, and one component provider that the config names. The framework
assembles everything else from those two.

There is **no configuration file format**. No YAML, no JSON, no INI. The config
is a class, so the language checks it and the build tool reads it without
running it.

`ApplicationContract` is what an application resolves to reach the boot's
result. The boot registers it on the container.
[Container](../container/README.md) covers contracts, bindings and the resolve
operations.

## Writing the application config

An application declares one config class, and that class carries the protocol's
config contract. Where the port ships a class holding the defaults, extend it
and set what the application changes. Where the port's config is a value type,
implement the contract instead.

```text
class Config extends HttpConfig {

    construct() {
        super(
            namespace: "App",
            dir: "/srv/orders",
            version: "1.0.0",
            environment: "production",
            debugMode: false,
            timezone: "UTC",
            key: "a secret this environment holds",
            dataPath: "src/App/Http/Data",
            dataNamespace: "App.Http.Data",
            providers: [
                new AppHttpComponentProvider(),
            ],
            requestReceivedMiddleware: [
                CacheResponseMiddleware,
            ],
        )
    }
}
```

The config passes each base setting in rather than redeclaring it as a static
default. So the config is plain code, and each setting takes an environment
variable, a deployment secret or a computed value where the application needs
one.

### The base config

`ConfigContract` declares what every application config carries.

| Property        | Holds                                               | Default                        |
| --------------- | --------------------------------------------------- | ------------------------------ |
| `namespace`     | the application's own source namespace              | `App`                          |
| `dir`           | the application's source directory                  | —                              |
| `version`       | the application's own version                       | the framework's own version    |
| `environment`   | which environment this is                           | `production`                   |
| `debugMode`     | whether the application reports detail on a failure | off                            |
| `timezone`      | the default timezone                                | `UTC`                          |
| `key`           | the application secret                              | a placeholder                  |
| `dataPath`      | where the generated data classes are written        | a path under the application   |
| `dataNamespace` | the namespace matching that path                    | the namespace for that path    |
| `providers`     | the component providers to register                 | the framework's own aggregator |
| `callbacks`     | the publish callbacks to run                        | none                           |

`ApplicationInfo` holds the framework's own version, its build time and the art
it prints. `version` defaults to the version there, and an application sets its
own.

Warning: the shipped `dir` is never the project root. Set `dir` to the project
root in every application config.

Warning: the default `key` is a placeholder, and every installation of the
framework carries the same one. Set `key` to a secret of the application's own
before the application serves anything.

The contract that carries `applicationName` varies. A port holds it on the base
config, or on the config of the protocol that prints it. A port that shows no
identity declares it nowhere. The name is read where an identity is shown, and
nothing else reads it.

### One config contract per protocol

Each protocol gets its own config contract. Every one of them extends
`ConfigContract`, so one class carries the base settings and the protocol's own.

| Contract              | Adds                                                       |
| --------------------- | ---------------------------------------------------------- |
| `HttpConfigContract`  | the middleware at each HTTP stage                          |
| `CliConfigContract`   | the default command name, and the middleware at each stage |
| `QueueConfigContract` | the middleware at each queue stage                         |

Each contract declares one property per stage, named for the stage. The
pipeline is therefore readable directly from the config. A protocol that has no
stage of a kind declares no property for that stage.

| Protocol | Stage properties                                                                                                                                                                                    |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP     | `requestReceivedMiddleware`, `routeMatchedMiddleware`, `routeNotMatchedMiddleware`, `routeDispatchedMiddleware`, `throwableCaughtMiddleware`, `sendingResponseMiddleware`, `responseSentMiddleware` |
| CLI      | `inputReceivedMiddleware`, `routeMatchedMiddleware`, `routeNotMatchedMiddleware`, `routeDispatchedMiddleware`, `throwableCaughtMiddleware`, `processExitingMiddleware`                              |
| Queue    | `jobReceivedMiddleware`, `routeMatchedMiddleware`, `routeNotMatchedMiddleware`, `routeDispatchedMiddleware`, `throwableCaughtMiddleware`, `settlingResultMiddleware`, `resultSettledMiddleware`     |

A port adds a contract per protocol it supports. A port with gRPC declares a
gRPC config contract; a port without it does not.

Each property holds the middleware by name. The handler resolves each one from
the container when the stage runs.

An application serves each protocol from an entry point of its own, with a
config class of its own. So a project that serves HTTP and the CLI writes two
config classes, and each one names the aggregator for its protocol.

### Overriding a component's config

Your application has one config class. Where it implements a component's config
contract, that component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how
you override, and implementing nothing leaves every default in place.

```text
class Config extends HttpConfig implements LogConfigContract, CacheConfigContract {
    defaultLogger = NullLogger
    defaultCache = NullCache
}
```

Each component's provider resolves `ConfigContract` and asks whether that
object also satisfies the component's own config contract. It binds your class
on a yes, and the component's default on a no. Each component's README holds the publisher that does it, and the
properties that component reads.

One property per setting, prefixed with the component or the adapter it
configures, so one class implements as many contracts as the application needs.
[`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md) holds the prefix
rules and the order a config declares its properties.

## Listing the component providers

`providers` is the whole list the framework registers. An application names one
provider of its own, as the config above does, and that provider names what it
depends on.

**`providers` replaces rather than extends.** A config that declares the list
declares the whole list. The framework merges nothing into it, because a silent
merge makes the registered set unpredictable.

### A component provider declares its own dependencies

A component provider says what one component contributes. It names the component
providers it depends on, and then the providers it adds for each kind.

| Operation               | Returns                                           |
| ----------------------- | ------------------------------------------------- |
| `getComponentProviders` | the component providers this component depends on |
| `getContainerProviders` | the component's service providers                 |
| `getEventProviders`     | the component's listener providers                |
| `getHttpProviders`      | the component's HTTP route providers              |
| `getCliProviders`       | the component's CLI route providers               |
| `getQueueProviders`     | the component's queue route providers             |

Each one takes the application, so a provider that varies by environment reads
the environment from it.

```text
class AppHttpComponentProvider implements ComponentProviderContract {

    getComponentProviders(app) {
        return [
            new HttpApplicationComponentProvider(),
        ]
    }

    getContainerProviders(app) {
        return [
            new AppHttpServiceProvider(),
        ]
    }

    getHttpProviders(app) {
        return [
            new AppHttpRouteProvider(),
        ]
    }

    // ...the event, cli and queue operations return an empty list
}
```

The framework ships an aggregator per shape of application. Each one names the
components that shape needs, so an application names one provider rather than
listing components itself.

| Aggregator                                | Names                                                             |
| ----------------------------------------- | ----------------------------------------------------------------- |
| `ApplicationComponentProvider`            | Container, and Event                                              |
| `HttpApplicationComponentProvider`        | the application aggregator, and the HTTP, Log and View components |
| `CliApplicationComponentProvider`         | the application aggregator, and the CLI and Log components        |
| `CliWithHttpApplicationComponentProvider` | the CLI aggregator, and the HTTP components                       |

A port adds one aggregator per protocol it supports, and one per combination an
application runs together.

### The framework walks the tree

The walk is depth first and post-order. A provider's whole dependency subtree
registers before the provider itself, so anything a provider depends on is
already registered when it runs.

The config's order decides precedence among siblings. A later provider's binding
for the same id replaces an earlier one's, so order the list by what depends on
what.

A provider list holds **constructed providers**, and it carries no conditional
logic. The build tool reads the construction expression without running it.
[`PROVIDERS.md`](../../convention/PROVIDERS.md) holds the shape each list
takes.

## Pointing a runtime at an entry point

An entry point is the class a runtime starts. Pick the one for the protocol and
the runtime being deployed, and point the runtime at it:

```text
// The project's HTTP entry file, which the web server serves.
Http.run(new Config())
```

```text
// The project's CLI entry file, which the shell runs.
Cli.run(new Config())
```

`run` takes the config and returns nothing. It boots, resolves the protocol's
handler, builds the unit of work from the runtime, and runs it.

### The two axes

An entry point is named for a **protocol**, and, where a runtime needs its own
server, for a **runtime**.

| Entry           | Runs                                            |
| --------------- | ----------------------------------------------- |
| `App`           | nothing; it is the boot every entry shares      |
| `Http`          | the protocol on the language's own server       |
| `Cli`           | the protocol from a command                     |
| `WorkerHttp`    | nothing; it is what a persistent runtime shares |
| `<Runtime>Http` | the protocol on that runtime's server           |

The framework groups by **runtime**, with the protocol in the class name. One
runtime directory holds the class for that runtime, and the protocol entry at
the top level holds the one the language serves itself.

An application groups by **protocol**, with the runtime in the class name. The
application's axis is the protocol module, which owns the config, the
controllers, the routing data and the providers. Only the server driving them
differs.

Warning: an application that nests its entry under the runtime gives several
classes the same name inside one protocol. Keep the two axes apart. A runtime
directory holds the class for that runtime, and nothing else.

### Extending an entry point

Each step of the boot is an operation on the entry point. So an application
subclasses the entry for its protocol, replaces the step it needs, and points
the runtime at its own class.

An application replaces two steps most often. `defaultExceptionHandler` is the
handler the boot installs at step 1, before the container holds anything.
`bootstrapThrowableHandler` is the handler the entry installs afterward, from
the config.

```text
class OrdersHttp extends Http {

    static defaultExceptionHandler() {
        new OrderThrowableHandler().enable()
    }
}
```

```text
OrdersHttp.run(new Config())
```

Every step is a static operation, so a runtime that cannot subclass the entry
calls the steps in order itself. A port whose language dispatches a static call
on the declaring class rather than the subclass takes that second route.

## Booting

`start` is what every entry point shares. It does this, in this order:

1. Installs a default throwable handler, where `debugMode` is on.
2. Records the time the application started.
3. Sets the project's base path from `dir`.
4. Constructs the container.
5. Constructs the application, which sets the default timezone from `timezone`.
6. Registers the config, the container and the application on the container.
7. Registers the protocol's config contract, where the config implements it.
8. Runs each callback the config names.
9. Loads the container data.

The protocol's own entry point then takes over. It installs the configured
throwable handler where `debugMode` is on, resolves the handler for the
protocol, builds the unit of work, and runs it.

An application that wants the boot's result without running a unit of work calls
`start` itself:

```text
app = App.start(new Config())

container = app.getContainer()
```

### What the application answers

| Operation                  | Answers                                           |
| -------------------------- | ------------------------------------------------- |
| `getContainer`             | the container this application built              |
| `getProviders`             | every component provider the tree collected       |
| `getContainerProviders`    | the service providers collected from the tree     |
| `getEventProviders`        | the listener providers collected from the tree    |
| `getHttpProviders`         | the HTTP route providers collected from the tree  |
| `getCliProviders`          | the CLI route providers collected from the tree   |
| `getQueueProviders`        | the queue route providers collected from the tree |
| `publishProviderCallbacks` | runs the config's publish callbacks               |
| `getEnvironment`           | which environment this is                         |
| `getDebugMode`             | whether debug mode is on                          |
| `getVersion`               | the application's version                         |

A port declares one reader per protocol it supports, named for that protocol.

`getProviders` walks the tree on the first call and keeps the result. The other
collecting operations read that result, and each one keeps its own collection.
So a second call reads the collection rather than building it again.

```text
app = container.getSingleton(ApplicationContract)

// Every component provider the config's list reached.
providers = app.getProviders()
```

## Booting without the cache, and then with it

**Every port boots with no generated data at all.** The provider tree is
walkable at run time, so build and test the application that way first.

With nothing generated, the boot finds no container data. It registers each
service provider the tree collected, and then stores the result as the container
data. The boot ends with a container that holds every binding.

The cache is a cold-start optimization. A process that handles one unit of work
and then exits pays the whole boot cost every time. That process is where the
generated data earns its place. A persistent runtime pays the cost once.

To boot with the cache, register the generated data in a callback the config
names. The boot runs every callback before it loads the container data, so the
data is already there and the boot walks nothing:

```text
class Config extends HttpConfig {

    construct() {
        super(
            // ...the rest of the config
            callbacks: [
                [AppHttpComponentProvider, "publishData"],
            ],
        )
    }
}
```

```text
class AppHttpComponentProvider implements ComponentProviderContract {

    static publishData(app) {
        app.getContainer().setSingleton(ContainerData, new AppContainerData())
    }

    // ...the component provider operations
}
```

An application that wants the generated data in one environment and the walk in
another branches on `getDebugMode` inside that callback. The build tool writes
one generated class per component, and
[`DATA_CACHE.md`](../../convention/DATA_CACHE.md) names each one.

Warning: a binding registered outside a provider reaches the uncached run only.
The build tool reads providers, so it never sees that binding. Register every
binding in a provider.

## Resolving a path inside the project

`Directory` resolves a path inside the project. The boot sets the base path from
the config's `dir`, and every operation resolves from there.

| Operation                   | Returns                               |
| --------------------------- | ------------------------------------- |
| `basePath`                  | the project root                      |
| `appPath`                   | the application directory             |
| `srcPath`                   | the source directory                  |
| `dataPath`                  | the data directory                    |
| `envPath`                   | the environment directory             |
| `publicPath`                | the directory the web server serves   |
| `resourcesPath`             | the resources directory               |
| `storagePath`               | the storage directory                 |
| `frameworkStoragePath`      | the framework's own storage directory |
| `frameworkStorageCachePath` | the framework's cache directory       |
| `logsStoragePath`           | the log directory                     |
| `testsPath`                 | the test directory                    |
| `vendorPath`                | the dependency directory              |
| `path`                      | a path with the separator prepended   |

Every operation takes an optional path to append, and returns the resolved path.

A port declares an operation for each directory its own layout holds, and none
for a directory it does not. The dependency directory, the environment
directory and the test directory differ most between ports.

```text
root = Directory.basePath()

template = Directory.resourcesPath("views")

logs = Directory.logsStoragePath()
```

Each segment is a value the application sets, so a project laid out differently
sets the segment rather than every path that uses it:

```text
Directory.srcPath = "lib"
```

Warning: a relative path that a call site builds itself reads the current
working directory, which the runtime decides. Resolve every path inside the
project through `Directory`.

## The child application

A persistent runtime keeps one process alive across many units of work. A
**child application** gives each unit of work its own scope over one shared
parent.

The boot happens once, in the parent. Each unit of work then gets a child, and
nothing the child registers reaches the parent. A child container gives the
container the same scope ([Container](../container/README.md)).

`ChildApplication` is constructed with the parent and the child container. It
answers `getContainer` with the child container, and it forwards every other
operation to the parent. So the boot's result is read once and shared, and only
the container is per-request.

`WorkerHttp` is what a persistent HTTP runtime shares.

| Operation                 | Does                                                     |
| ------------------------- | -------------------------------------------------------- |
| `bootstrap`               | boots once, and returns the application                  |
| `handle`                  | serves one unit of work on a child                       |
| `bootstrapParentServices` | resolves the services the parent holds                   |
| `getChildApplication`     | constructs the child application for a unit of work      |
| `getChildContainer`       | constructs the child container for a unit of work        |
| `bootstrapChildContainer` | registers the application and the container on the child |
| `handleRequest`           | resolves the protocol's handler and runs it              |
| `getRequest`              | builds the unit of work from the runtime                 |

A runtime entry point extends it, boots once, and calls `handle` per unit of
work:

```text
class CustomHttp extends WorkerHttp {

    static run(config) {
        app = self.bootstrap(config)
        data = app.getContainer().getData()

        // The runtime's own loop.
        while (runtime.hasRequest()) {
            self.handle(app, data, self.getRequest())
        }
    }
}
```

`bootstrap` resolves the services that belong in the parent before it returns,
so each unit of work reads them rather than building them again. Override
`bootstrapParentServices` to add one of the application's own:

```text
class CustomHttp extends WorkerHttp {

    static bootstrapParentServices(app) {
        app.getContainer().getSingleton(RouteCollectionContract)
        app.getContainer().getSingleton(OrderRepository)
    }
}
```

An override replaces the base's own list. Repeat what the base resolves where
the application still needs it. A service that no override resolves is built on
the child instead, once for each unit of work.

Warning: a value written to the parent container reaches every later unit of
work. Do not write to the parent container after `bootstrap` returns. Register
a per-request value on the child.

An entry point that serves one unit of work per process needs none of this. The
process is the scope.

## Service registration

The component ships no service provider. The boot registers what the rest of the
framework resolves, before any provider runs:

| Binding               | Is                                         |
| --------------------- | ------------------------------------------ |
| `ConfigContract`      | the application's own config class         |
| the config's class    | the same object, under its own name        |
| `ContainerContract`   | the container the boot built               |
| `ApplicationContract` | the application the boot built             |
| `HttpConfigContract`  | the config, where the config implements it |
| `CliConfigContract`   | the config, where the config implements it |

A port adds a row per protocol it supports. The boot registers that protocol's
config contract where the application's config implements it.

**One config object, under several ids.** `ConfigContract` is what a component's
provider resolves to ask whether the application overrides that component's
settings. The protocol's contract is what the protocol's handler resolves to
read the middleware for each stage. The config's own class is what application
code resolves to read a property that no contract declares.

The component's own providers are **component** providers. They name components
rather than bindings, and
[Listing the component providers](#listing-the-component-providers) holds the
table of the aggregators the framework ships.

Each aggregator declares its component list and nothing else. It adds no
binding, no listener and no route of its own, so every other operation returns
an empty list.

The container's own service provider is what turns the collected service
providers into bindings. It registers each one on the container and stores the
result as the container data, which is the uncached boot's last step.
