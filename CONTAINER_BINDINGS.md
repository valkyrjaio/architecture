# Valkyrja Container Bindings

## Core Concept

Every class, interface, and contract in Valkyrja needs a stable, unique identifier for use as a container binding key.
Languages differ in what they use as that key:

| Language   | Binding key type            | Notes                                                          |
| ---------- | --------------------------- | -------------------------------------------------------------- |
| PHP        | `::class` — FQN string      | compiler-verified                                              |
| Java       | `.class` — `Class<T>` token | compiler-verified                                              |
| Python     | string constant             | class object forces import, defeating Python 3.14 lazy loading |
| Go         | string constant             | required — no class reference mechanism                        |
| TypeScript | string constant             | required — interfaces erased at runtime                        |

**Go, Python, and TypeScript require string constants.** PHP and Java use native language mechanisms. Python uses string
constants for the same reason as Go and TypeScript — using class objects as keys forces module imports, defeating Python
3.14's lazy import mechanism which is the primary solution to Python's cold start problem.

**The key is modeled on how the port imports the class.** It is the directory path of the source file, plus the class
name. The class name keeps its PascalCase spelling in every port.

The key is therefore language-specific, because each port's directory layout already differs. PHP and TypeScript write
a StudlyCase directory. Go and Python write a lowercase directory. Each port copies the layout that the port itself
has, so no port converts the case of another port's path:

| Language   | Namespaced class                                        | Key                                                     |
| ---------- | ------------------------------------------------------- | ------------------------------------------------------- |
| PHP        | `Valkyrja\Container\Manager\Contract\ContainerContract` | `Valkyrja\Container\Manager\Contract\ContainerContract` |
| TypeScript | `Valkyrja/Container/Manager/Contract/ContainerContract` | `Valkyrja.Container.Manager.ContainerContract`          |
| Python     | `valkyrja.container.manager.contract.ContainerContract` | `valkyrja.container.manager.ContainerContract`          |
| Go         | `valkyrja/container/manager/contract.ContainerContract` | `valkyrja.container.manager.ContainerContract`          |

Two rules shape every key:

- **Copy the directory path of the source file.** The Cli component and the Http component each hold a
  `RouterContract`, so a shorter path is not unique.
- **Remove the `Contract` segment**, because the class name ends in `Contract` already. A contract whose name does not
  end in `Contract` keeps the segment. Python holds `ValkyrjaThrowable` in
  `src/valkyrja/throwable/contract/valkyrja_throwable.py`, so its key is
  `valkyrja.throwable.contract.ValkyrjaThrowable`.

```typescript
export class ContainerServiceId {
    // Valkyrja\Container\Manager\Contract\ContainerContract — the Contract segment goes
    static readonly Contract = 'Valkyrja.Container.Manager.ContainerContract' as const;
    // Valkyrja\Container\Data\ContainerData — a concrete class keeps every segment
    static readonly Data = 'Valkyrja.Container.Data.ContainerData' as const;
}
```

Warning: do not copy a key from one port into another. A Go package name and a Python module name are lowercase, so
their keys are lowercase. A TypeScript directory is StudlyCase, so its key is StudlyCase. A developer reads a key and
looks for that file, so a key that does not match the port's own import path sends the reader to a path that the port
does not have.

Go is the one port where the key is not literal syntax. A Go import path is slash-separated and carries the module
prefix, as in `github.com/valkyrjaio/valkyrja-go/container/manager`, and only the final `package.Symbol` selector uses
a dot. A Go key drops the module prefix and writes a dot between each segment, so it reads like the key in every other
port. Python is literal: `from valkyrja.container.manager import ContainerContract` is the import that its key names.

PHP and Java do not hand-write a key. `::class` and `.class` return the fully qualified name, so the `Contract` segment
stays in those two ports.

An application uses the same rules with its own root. In TypeScript, `App\Repository\Contract\UserRepositoryContract`
becomes `App.Repository.UserRepositoryContract`, and `App\Service\Contract\DatabaseContract` becomes
`App.Service.DatabaseContract`. In Python the same two classes give `app.repository.UserRepositoryContract` and
`app.service.DatabaseContract`.

Each segment is singular, because the key copies the source directory and a Valkyrja segment is singular. The framework
writes `Repository\`, `Manager\`, `Provider\`, and `Middleware\`. An application that writes `Repositories\` gets
`App.Repositories.UserRepositoryContract`, because the key copies the directory that the application has.

---

## Why Not Reflection or Dynamic Resolution

The original PHP and Java implementations used `::class` and `.class` respectively not just as string identifiers but as
dynamic dispatch mechanisms — passing the class reference to the container which would then use reflection or dynamic
method calls to instantiate the class.

This approach has several problems:

- **Reflection is slow** — runtime introspection adds overhead on every resolution
- **Assumes a specific method signature** — breaks if a class doesn't conform
- **Implicit** — impossible to trace what gets called without running the code
- **Non-portable** — Go, Python, and TypeScript have no equivalent mechanism

The solution is **closure-based bindings** — the developer explicitly provides a factory function. The container stores
and invokes the closure. No reflection, no dynamic dispatch, no assumptions about the class.

---

## Closure-Based Bindings

All bindings across all ports use closures as the factory mechanism:

**PHP**

```php
$container->bind(
    UserRepositoryContract::class,
    static fn(ContainerContract $c): UserRepository => new UserRepository(
        $c->getSingleton(DatabaseContract::class)
    )
);

$container->bindSingleton(
    UserRepositoryContract::class,
    static fn(ContainerContract $c): UserRepository => new UserRepository(
        $c->getSingleton(DatabaseContract::class)
    )
);
```

**Java**

```java
container.bind(
    UserRepositoryContract.class,
    c -> new UserRepository(
        c.getSingleton(DatabaseContract.class)
    )
);

container.bindSingleton(
    UserRepositoryContract.class,
    c -> new UserRepository(
        c.getSingleton(DatabaseContract.class)
    )
);
```

**Go**

```go
container.Bind(
    UserRepositoryClass,
    func(c ContainerContract) any {
        return NewUserRepository(
            c.GetSingleton(DatabaseClass),
        )
    },
)

container.BindSingleton(
    UserRepositoryClass,
    func(c ContainerContract) any {
        return NewUserRepository(
            c.GetSingleton(DatabaseClass),
        )
    },
)
```

**Python** — string constant as key:

```python
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.get_singleton(ContainerConstants.DATABASE))
)

container.bind_singleton(
    ContainerConstants.ROUTER,
    lambda c: Router(c.get_singleton(ContainerConstants.DISPATCHER))
)
```

**TypeScript**

```typescript
container.bind(
    UserRepositoryClass,
    (c: ContainerContract) => new UserRepository(
        c.getSingleton(DatabaseClass)
    )
)

container.bindSingleton(
    UserRepositoryClass,
    (c: ContainerContract) => new UserRepository(
        c.getSingleton(DatabaseClass)
    )
)
```

---

## The Key Type Safety Distinction

**PHP and Java** — compiler-verified class references:

```php
// ::class — autoloader guarantees this class exists
$container->bind(UserRepositoryContract::class, fn($c) => ...);
```

```java
// .class — compile-time type token, cannot reference a non-existent class
container.bind(UserRepositoryContract.class, c -> ...);
```

**Python** — string constants, not class objects. A `type` object is hashable and works as a dict key, so the class
object looks like the better key. It is not. Warning: a class object as the key imports the module that holds the
class, and the container holds every binding that the application can resolve. The cache then imports every such
module at load time, which defeats the lazy import that Python 3.14 gives.

```python
# Wrong — the class object as the key imports the module before anything resolves it.
container.bind(UserRepositoryContract, lambda c: UserRepository(c.get_singleton(DatabaseContract)))
```

```python
# Right — the string constant names the class and imports nothing.
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.get_singleton(ContainerConstants.DATABASE)),
)
```

A bare class name still names the type for direct code usage. It is not a container key, and the container resolves on
the full key string only.

**Go, Python, and TypeScript** — string constants required. No language of the three has a usable class reference at
runtime for this purpose:

- Go has no class system at all
- Python has a usable class object, but a class object as the key imports the module that holds the class, which
  defeats the lazy import that Python 3.14 gives
- TypeScript interfaces are erased at runtime — `Map<Interface, Factory>` is not possible since interfaces don't exist
  at runtime. Constructor references work for concrete classes but most Valkyrja bindings are against
  contracts/interfaces

```go
// Go — string constant required
const UserRepositoryClass = "app.repository.UserRepositoryContract"
container.Bind(UserRepositoryClass, func (c ContainerContract) any { ... })
```

```typescript
// TypeScript — string constant required for interface/contract bindings
// (constructor references work for concrete classes but not interfaces)
export const UserRepositoryClass = 'App.Repository.UserRepositoryContract'
container.bind(UserRepositoryClass, (c) => new UserRepository(c.getSingleton(DatabaseClass)))
```

This is an honest reflection of each language's capabilities rather than a limitation to paper over.

---

## Per-Component Constants Files

Constants files exist for **all five languages** at the component level — they ship as part of the framework source. PHP
and Java constants hold `::class` / `.class` values respectively. Go, Python, and TypeScript hold string literals. The
abstraction is consistent across all languages — callers always reference the constant, never the raw value.

```
Framework constants  → shipped with each component, all five languages
Application constants → written by the developer, following the same pattern
                        (Sindri auto-generation is a planned future enhancement)
```

### Why Per-Component, Not A Single Central File

A single central constants file in the container component would:

- Grow unboundedly as the framework expands
- Create hidden coupling between every component and the container
- Become a merge conflict hotspot in open source contributions
- Require every contributor to modify a central file when adding a class
- Violate component isolation — a developer working on Http would need to navigate Container, Dispatcher, Event, Routing
  constants

Per-component constants files mean:

- Each component owns its identifiers — fully isolated
- Adding a new component means adding a new constants file, not modifying a central one
- Third-party packages follow the same pattern without touching framework files
- Consistent with how the exception hierarchy is organized — same mental model throughout

### Component Provider Constants Class — Not Part of the Framework

Constants files exist for **binding key strings** — the cross-language string identity problem for container bindings.
They are not for provider class references.

A constants class that aliases component provider class references (e.g.
`HttpConstants::HTTP_COMPONENT_PROVIDER = HttpComponentProvider::class`) must not be created. It would allow developers
to use constant references in the application config:

```php
// this breaks the build tool — constant reference not resolvable from AST
new AppConfig(providers: [HttpConstants::HTTP_COMPONENT_PROVIDER])
```

The build tool reads the application config class via AST to discover providers. It cannot follow constant references
without executing code. Provider class lists must always use `::class` / `.class` / class objects directly.

Binding key constants files (`ContainerConstants`, `HttpConstants` for binding strings etc.) are correct and should
exist. The provider class reference constants class specifically does not.

### Structure

```
valkyrja/
  container/
    ContainerConstants.php     ← PHP/Java: optional complement to ::class / .class
    container_constants.go     ← Go: required string constants
    container-constants.ts     ← TypeScript: required string constants
    container_constants.py     ← Python: required string constants
    ContainerContract.php
    ContainerException.php
  http/
    HttpConstants.php
    http_constants.go
    http-constants.ts
    routing/
      HttpRoutingConstants.php
      http_routing_constants.go
      http-routing-constants.ts
```

---

## Language-Specific Implementation

### PHP

`::class` is the primary mechanism — compiler-verified, returns the FQN string. The constants file is recommended as a
complement:

```php
// ContainerConstants.php
final class ContainerConstants
{
    public const CONTAINER         = ContainerContract::class;
    public const USER_REPOSITORY   = UserRepositoryContract::class;
    public const DATABASE          = DatabaseContract::class;
}
```

The constants file provides a single place to look up every identifier in a component without navigating the class
hierarchy. Useful for config files, serialization, and any context where importing the class itself is undesirable.

### Java

`.class` is the primary mechanism — compile-time type token (`Class<T>`). Constants file recommended:

```java
// ContainerConstants.java
public final class ContainerConstants {
    public static final Class<ContainerContract> CONTAINER
            = ContainerContract.class;
    public static final Class<UserRepositoryContract> USER_REPOSITORY
            = UserRepositoryContract.class;

    private ContainerConstants() {
    }
}
```

Note: Java's `IllegalArgumentException` is the language root for `ValkyrjaInvalidArgumentException` — the naming uses
`InvalidArgument` for cross-port parity while the inheritance uses `IllegalArgument` for language correctness.

### Go

No `::class` equivalent. String constants are the only mechanism. The constants file is required:

```go
// container_constants.go
package container

const (
	ContainerClass     = "valkyrja.container.manager.ContainerContract"
	ContainerDataClass = "valkyrja.container.data.ContainerData"
	RouterClass        = "valkyrja.http.routing.dispatcher.RouterContract"
)
```

Type safety is convention-enforced. The linter and code review are the enforcement mechanisms.

A Go package name is lowercase, so a Go key is lowercase. The type name keeps its PascalCase spelling, because Go
writes an exported type in PascalCase. TypeScript spells the same binding `Valkyrja.Container.Manager.ContainerContract`,
so do not copy a key between the two ports.

### Python

**Minimum version: Python 3.14.** String constants are used as binding keys — same as Go and TypeScript. Using class
object keys forces module imports which defeats Python 3.14's lazy import mechanism.

```python
# container_constants.py — required, same as Go and TypeScript
class ContainerConstants:
    CONTAINER = "valkyrja.container.manager.ContainerContract"
    ROUTER = "valkyrja.http.routing.dispatcher.RouterContract"
    USER_REPOSITORY = "app.repository.UserRepositoryContract"
    DATABASE = "app.service.DatabaseContract"
```

### The Uniform Lambda Pattern

The container's internal bindings map always stores lambdas — whether populated from a service provider at runtime or
loaded from a cache data file. This makes resolution uniform with no conditional logic.

**Service provider** — plain method reference, no lambda:

```python
class UserServiceProvider(ServiceProviderContract):
    @staticmethod
    def publishers() -> dict:
        return {
            'app.repository.UserRepositoryContract': UserServiceProvider.publish_user_repository,
        }

    @staticmethod
    def publish_user_repository(c: ContainerContract) -> None:
        c.set_singleton(
            'app.repository.UserRepositoryContract',
            UserRepository(c.get_singleton('app.service.DatabaseContract'))
        )
```

**Container** — wraps method reference in lambda on registration, loads cache as-is, resolves by always calling the
lambda:

```python
class Container:

    def register_provider(self, provider: ServiceProviderContract) -> None:
        for key, callable_ref in provider.publishers().items():
            # wrap in lambda — internal map always stores lambdas
            self._bindings[key] = lambda c=callable_ref: c

    def load_cache(self, data: dict) -> None:
        # cache data already in lambda format — register as-is
        self._bindings.update(data)

    def get_service(self, key: str):
        # always call the lambda — uniform, no conditional check needed
        callable_ref = self._bindings[key]()
        return callable_ref(self)

    def get_singleton(self, key: str):
        if key not in self._singletons:
            self._singletons[key] = self.get_service(key)
        return self._singletons[key]
```

**Sindri** — reads the plain method reference from `publishers()` AST and writes it as a lambda in the generated cache
file, matching the container's internal format:

```python
# generated AppContainerData — lambda format, same as container internal map
APP_CONTAINER_DATA = {
    'app.repository.UserRepositoryContract': lambda: UserServiceProvider.publish_user_repository,
    'app.service.DatabaseContract': lambda: DatabaseServiceProvider.publish_database,
    'valkyrja.http.routing.dispatcher.RouterContract': lambda: HttpServiceProvider.publish_router,
}
```

**Resolution is always uniform:**

```python
# always: call lambda → get method ref → call method ref with container
callable_ref = self._bindings[key]()  # lambda() → method ref
return callable_ref(self)  # method ref(container)
```

This is the only Python-specific behaviour in the container. No conditional checks, no dispatch-style indirection. The
service provider stays clean, the cache format matches the internal map exactly.

The `class_()` FQN helper generates string constants from class objects where needed:

```python
# utility for generating string constants — not for use as a binding key directly
def class_(cls) -> str:
    return f"{cls.__module__}.{cls.__qualname__}"
```

Note: `class_` uses a trailing underscore because `class` is a reserved word in Python.

### TypeScript

TypeScript interfaces and types are erased at runtime — they cannot be used as `Map` keys. Constructor references work
for concrete classes but most Valkyrja bindings are against contracts/interfaces. String constants are required:

```typescript
// container-constants.ts
export const ContainerConstants = {
    CONTAINER: 'Valkyrja.Container.Manager.ContainerContract',
    USER_REPOSITORY: 'App.Repository.UserRepositoryContract',
    DATABASE: 'App.Service.DatabaseContract',
} as const
```

TypeScript's `typeof` and `keyof` derive types from the constants for additional type safety:

```typescript
type ContainerKey = typeof ContainerConstants[keyof typeof ContainerConstants]
```

**Why not constructor references?** Constructor references (`new () => T`) work as `Map` keys at runtime for concrete
classes, but cannot represent interface bindings — the primary use case in Valkyrja. A constructor reference to
`UserRepositoryContract` does not exist if `UserRepositoryContract` is an interface. String constants are the only
mechanism that works uniformly for both interface and class bindings.

---

## Container API

### Binding Methods

| Method                        | Description                                                                                     |
| ----------------------------- | ----------------------------------------------------------------------------------------------- |
| `bind(id, callable)`          | Binds a service ID to a callable factory. Returns a fresh instance on every `getService()` call |
| `bindSingleton(id, callable)` | Same as `bind()` but singleton-scoped — callable invoked once, result cached                    |
| `bindAlias(alias, id)`        | Maps one service ID to another already registered                                               |
| `setSingleton(id, instance)`  | Registers an already-constructed object directly — used inside publisher callbacks              |

The recommended callable convention is `[ClassName::class, 'make']` pointing to a static factory — or
`[ServiceProvider::class, 'publishMethod']` from the publishers map.

### Inspection Methods

| Method                    | Description                                          |
| ------------------------- | ---------------------------------------------------- |
| `has(id)`                 | PSR-11 — true if registered in any form              |
| `isSingleton(id)`         | True if binding OR resolved instance exists          |
| `isSingletonBinding(id)`  | True if callable binding exists but not yet resolved |
| `isSingletonInstance(id)` | True if already resolved and cached                  |
| `isService(id)`           | True if registered as a per-call service             |
| `isAlias(id)`             | True if registered as an alias                       |

`isSingleton` is equivalent to `isSingletonBinding || isSingletonInstance`. The two fine-grained methods are used by
child containers to distinguish "registered but not yet built" from "already live and reusable".

### Resolution Methods

| Method              | Description                                                                             |
| ------------------- | --------------------------------------------------------------------------------------- |
| `get(id)`           | PSR-11 — works across all three types, slightly slower due to additional lookup         |
| `getSingleton(id)`  | Resolves a singleton — creates and caches on first access, returns cached on subsequent |
| `getService(id)`    | Resolves a service — always returns a fresh instance                                    |
| `getAliased(alias)` | Resolves the service the alias points to                                                |

When the type is known, prefer the specific method over `get()`. The difference is small per call but meaningful in hot
paths like route dispatch.

---

## The Container's Perspective

The container never needs to know anything about the class itself — only the key and how to build it. This is a cleaner
contract than reflection ever was:

```
Key (string / class object)  +  Factory (closure)  =  Complete binding
```

The key type per language:

- PHP: `UserRepositoryContract::class` (FQN string)
- Java: `UserRepositoryContract.class` (Class<T> token)
- Python: `'app.repository.UserRepositoryContract'` (string constant)
- Go: `"app.repository.UserRepositoryContract"` (string constant)
- TypeScript: `'App.Repository.UserRepositoryContract'` (string constant)

The container stores the closure. When `make(key)` is called:

1. Look up the closure by key
2. Invoke the closure with the container as the argument
3. Return the result

No reflection. No dynamic dispatch. No assumptions. Just a function call.

---

## PHP and Java Migration Note

PHP and Java currently use `::class` / `.class` not just for binding keys but also for dynamic instantiation via
reflection. The migration path is:

1. Add constants files per component
2. Migrate bindings to closure-based factories referencing constants
3. Remove dynamic reflection/method resolution from the container
4. Document closure-based binding as the canonical pattern

The `::class` / `.class` syntax is retained as the value passed to the constants (PHP/Java) — the language guarantees it
refers to a real class. The constants file just organizes those verified values in one place.

This migration makes every port's container architecture identical at the behavioral level, with only the key type
safety mechanism differing per language.

---

## Discussion Summary

The container binding problem surfaced when analyzing what a Python or TypeScript port of Valkyrja's container would
look like. PHP's `::class` and Java's `.class` serve a dual purpose in the current implementation: they provide a
binding key AND enable dynamic dispatch via reflection or dynamic method calls. Neither capability exists in Go,
TypeScript, or Python in any reliable cross-deployment form.

The first insight was separating these two concerns: the binding key (a string identifier) and the factory (how to build
the class). These were conflated in the original PHP/Java implementation because reflection made it possible to derive
the factory from the key. Separating them makes the architecture language-agnostic.

The second insight was that closures solve the factory problem completely across all languages. Every language supports
first-class functions. A closure captures its dependencies explicitly, executes without reflection, and is transparent —
you can read it and know exactly what will be constructed. This is strictly better architecture than dynamic dispatch
even in PHP and Java.

The third insight was the per-component constants file. The alternative — a single central constants file in the
container component — was rejected because it creates exactly the kind of tight coupling that the component architecture
is designed to avoid. Per-component constants mean each component owns its identifiers, contributing to the same
isolation principles that govern the rest of the framework.

The decision to recommend (but not require) constants files for PHP and Java was made to keep the option open for full
cross-language consistency without forcing a breaking change. The practical benefits — a single lookup location,
refactor safety, grep-ability — make the constants file valuable in PHP and Java independent of the cross-language
parity argument.
