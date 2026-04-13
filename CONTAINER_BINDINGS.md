# Valkyrja Container Bindings

## Core Concept

Every class, interface, and contract in Valkyrja needs a stable, unique identifier for use as a container binding key.
Languages differ in what they use as that key:

| Language   | Binding key type            | Notes                                                          |
|------------|-----------------------------|----------------------------------------------------------------|
| PHP        | `::class` — FQN string      | compiler-verified                                              |
| Java       | `.class` — `Class<T>` token | compiler-verified                                              |
| Python     | string constant             | class object forces import, defeating Python 3.14 lazy loading |
| Go         | string constant             | required — no class reference mechanism                        |
| TypeScript | string constant             | required — interfaces erased at runtime                        |

**Go, Python, and TypeScript require string constants.** PHP and Java use native language mechanisms. Python uses string
constants for the same reason as Go and TypeScript — using class objects as keys forces module imports, defeating Python
3.14's lazy import mechanism which is the primary solution to Python's cold start problem.

The string constant format for Go and TypeScript:

```
io.valkyrja.{component}.{ClassName}
io.valkyrja.container.ContainerContract
io.valkyrja.http.routing.HttpRoutingContract
```

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
        $c->make(DatabaseContract::class)
    )
);

$container->singleton(
    UserRepositoryContract::class,
    static fn(ContainerContract $c): UserRepository => new UserRepository(
        $c->make(DatabaseContract::class)
    )
);
```

**Java**

```java
container.bind(
        UserRepositoryContract .class,
        c ->new

UserRepository(
        c.make(DatabaseContract.class)
    )
            );

            container.

singleton(
        UserRepositoryContract .class,
        c ->new

UserRepository(
        c.make(DatabaseContract.class)
    )
            );
```

**Go**

```go
container.Bind(
UserRepositoryClass,
func (c ContainerContract) any {
return NewUserRepository(
c.Make(DatabaseClass),
)
},
)

container.Singleton(
UserRepositoryClass,
func(c ContainerContract) any {
return NewUserRepository(
c.Make(DatabaseClass),
)
},
)
```

**Python** — string constant as key:

```python
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.make(ContainerConstants.DATABASE))
)

container.singleton(
    ContainerConstants.ROUTER,
    lambda c: Router(c.make(ContainerConstants.DISPATCHER))
)
```

**TypeScript**

```typescript
container.bind(
    UserRepositoryClass,
    (c: ContainerContract) => new UserRepository(
        c.make(DatabaseClass)
    )
)

container.singleton(
    UserRepositoryClass,
    (c: ContainerContract) => new UserRepository(
        c.make(DatabaseClass)
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
container.bind(UserRepositoryContract .class, c ->...);
```

**Python** — class objects as keys. `type` objects are hashable in Python and work natively as dict keys. This is
idiomatic, IDE-supported, and eliminates the need for string constants entirely:

```python
# class object as key — hashable, IDE autocomplete works, cannot mistype
container.bind(UserRepositoryContract, lambda c: UserRepository(c.make(Database)))
container.make(UserRepositoryContract)  # same key, type-checked by mypy/pyright
```

The key must be the exact class object — subclasses are different keys, which is correct for a DI container (bind
against the contract, resolve against the contract):

```python
container.bind(UserRepositoryContract, lambda c: UserRepository(...))  # contract as key
container.make(UserRepositoryContract)  # ✅ resolves correctly
container.make(UserRepository)  # ❌ KeyError — different object, intentional
```

**Go and TypeScript** — string constants required. Neither language has a usable class reference at runtime for this
purpose:

- Go has no class system at all
- TypeScript interfaces are erased at runtime — `Map<Interface, Factory>` is not possible since interfaces don't exist
  at runtime. Constructor references work for concrete classes but most Valkyrja bindings are against
  contracts/interfaces

```go
// Go — string constant required
const UserRepositoryClass = "io.valkyrja.user.UserRepositoryContract"
container.Bind(UserRepositoryClass, func (c ContainerContract) any { ... })
```

```typescript
// TypeScript — string constant required for interface/contract bindings
// (constructor references work for concrete classes but not interfaces)
export const UserRepositoryClass = 'io.valkyrja.user.UserRepositoryContract'
container.bind(UserRepositoryClass, (c) => new UserRepository(c.make(DatabaseClass)))
```

This is an honest reflection of each language's capabilities rather than a limitation to paper over.

---

## Per-Component Constants Files

Constants files are **required for Go, Python, and TypeScript** where string constants are the only binding key
mechanism. They are **optional but recommended for PHP and Java** as a complement to `::class` / `.class`.

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
	ContainerClass      = "io.valkyrja.container.ContainerContract"
	UserRepositoryClass = "io.valkyrja.container.UserRepositoryContract"
	DatabaseClass       = "io.valkyrja.container.DatabaseContract"
)
```

Type safety is convention-enforced. The linter and code review are the enforcement mechanisms. The string format
`io.valkyrja.{component}.{ClassName}` is the cross-port standard.

### Python

**Minimum version: Python 3.14.** String constants are used as binding keys — same as Go and TypeScript. Using class
object keys forces module imports which defeats Python 3.14's lazy import mechanism.

```python
# container_constants.py — required, same as Go and TypeScript
class ContainerConstants:
    CONTAINER = "io.valkyrja.container.ContainerContract"
    ROUTER = "io.valkyrja.http.routing.RouterContract"
    USER_REPOSITORY = "app.repositories.UserRepositoryContract"
    DATABASE = "app.services.DatabaseContract"
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
            'app.repositories.UserRepositoryContract': UserServiceProvider.publish_user_repository,
        }

    @staticmethod
    def publish_user_repository(c: ContainerContract) -> None:
        c.set_singleton(
            'app.repositories.UserRepositoryContract',
            UserRepository(c.make('app.services.DatabaseContract'))
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

    def make(self, key: str):
        # always call the lambda — uniform, no conditional check needed
        callable_ref = self._bindings[key]()
        return callable_ref(self)

    def singleton(self, key: str):
        if key not in self._singletons:
            self._singletons[key] = self.make(key)
        return self._singletons[key]
```

**Forge** — reads the plain method reference from `publishers()` AST and writes it as a lambda in the generated cache
file, matching the container's internal format:

```python
# generated AppContainerData — lambda format, same as container internal map
APP_CONTAINER_DATA = {
    'app.repositories.UserRepositoryContract': lambda: UserServiceProvider.publish_user_repository,
    'app.services.DatabaseContract': lambda: DatabaseServiceProvider.publish_database,
    'io.valkyrja.http.RouterContract': lambda: HttpServiceProvider.publish_router,
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
    CONTAINER: 'io.valkyrja.container.ContainerContract',
    USER_REPOSITORY: 'io.valkyrja.container.UserRepositoryContract',
    DATABASE: 'io.valkyrja.container.DatabaseContract',
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

## The Container's Perspective

The container never needs to know anything about the class itself — only the key and how to build it. This is a cleaner
contract than reflection ever was:

```
Key (string / class object)  +  Factory (closure)  =  Complete binding
```

The key type per language:

- PHP: `UserRepositoryContract::class` (FQN string)
- Java: `UserRepositoryContract.class` (Class<T> token)
- Python: `'app.repositories.UserRepositoryContract'` (string constant)
- Go: `"io.valkyrja.user.UserRepositoryContract"` (string constant)
- TypeScript: `'io.valkyrja.user.UserRepositoryContract'` (string constant)

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
