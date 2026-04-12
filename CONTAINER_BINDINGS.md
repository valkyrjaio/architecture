# Valkyrja Container Bindings

## Core Concept

Every class, interface, and contract in Valkyrja needs a stable, unique string
identifier for use as a container binding key. In PHP and Java this is provided
natively by the language. In Go, Python, and TypeScript it must be provided
explicitly via string constants.

The format is consistent across all languages:

```
io.valkyrja.{component}.{ClassName}
io.valkyrja.container.ContainerNotFoundException
io.valkyrja.http.routing.HttpRoutingNotFoundException
```

---

## Why Not Reflection or Dynamic Resolution

The original PHP and Java implementations used `::class` and `.class`
respectively not just as string identifiers but as dynamic dispatch mechanisms —
passing the class reference to the container which would then use reflection or
dynamic method calls to instantiate the class.

This approach has several problems:

- **Reflection is slow** — runtime introspection adds overhead on every
  resolution
- **Assumes a specific method signature** — breaks if a class doesn't conform
- **Implicit** — impossible to trace what gets called without running the code
- **Non-portable** — Go, Python, and TypeScript have no equivalent mechanism

The solution is **closure-based bindings** — the developer explicitly provides a
factory function. The container stores and invokes the closure. No reflection,
no dynamic dispatch, no assumptions about the class.

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

**Python**

```python
container.bind(
    UserRepositoryClass,
    lambda c: UserRepository(
        c.make(DatabaseClass)
    )
)

container.singleton(
    UserRepositoryClass,
    lambda c: UserRepository(
        c.make(DatabaseClass)
    )
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

PHP and Java use the language's native class reference mechanism as the binding
key. This provides compiler-verified type safety — you cannot pass a
non-existent class:

```php
// ::class is compiler-verified — autoloader guarantees this class exists
$container->bind(UserRepositoryContract::class, fn($c) => ...);
```

```java
// .class is compiler-verified — cannot reference a non-existent class
container.bind(UserRepository .class, c ->...);
```

Go, Python, and TypeScript use string constants — manually maintained, no
compiler verification:

```go
// string constant — manually maintained, convention enforced
const UserRepositoryClass = "io.valkyrja.user.UserRepository"
container.Bind(UserRepositoryClass, func (c ContainerContract) any { ... })
```

This is an honest tradeoff rather than a limitation to paper over. The framework
documents it clearly per language.

---

## Per-Component Constants Files

Every component ships a constants file containing the string identifiers for all
classes, interfaces, and contracts in that component. This is required for all
languages and recommended for PHP and Java (where it provides a useful
complement to `::class` / `.class`).

### Why Per-Component, Not A Single Central File

A single central constants file in the container component would:

- Grow unboundedly as the framework expands
- Create hidden coupling between every component and the container
- Become a merge conflict hotspot in open source contributions
- Require every contributor to modify a central file when adding a class
- Violate component isolation — a developer working on Http would need to
  navigate Container, Dispatcher, Event, Routing constants

Per-component constants files mean:

- Each component owns its identifiers — fully isolated
- Adding a new component means adding a new constants file, not modifying a
  central one
- Third-party packages follow the same pattern without touching framework files
- Consistent with how the exception hierarchy is organized — same mental model
  throughout

### Component Provider Constants Class — Not Part of the Framework

Constants files exist for **binding key strings** — the cross-language string
identity problem for container bindings. They are not for provider class
references.

A constants class that aliases component provider class references (e.g.
`HttpConstants::HTTP_COMPONENT_PROVIDER = HttpComponentProvider::class`) must
not be created. It would allow developers to use constant references in the
application config:

```php
// this breaks the build tool — constant reference not resolvable from AST
new AppConfig(providers: [HttpConstants::HTTP_COMPONENT_PROVIDER])
```

The build tool reads the application config class via AST to discover providers.
It cannot follow constant references without executing code. Provider class
lists must always use `::class` / `.class` / class objects directly.

Binding key constants files (`ContainerConstants`, `HttpConstants` for binding
strings etc.) are correct and should exist. The provider class reference
constants class specifically does not.

### Structure

```
valkyrja/
  container/
    ContainerConstants.php     ← all container class identifiers
    ContainerContract.php
    ContainerException.php
  http/
    HttpConstants.php          ← all http class identifiers
    routing/
      HttpRoutingConstants.php ← all http routing class identifiers
```

---

## Language-Specific Implementation

### PHP

`::class` is the primary mechanism — compiler-verified, returns the FQN string.
The constants file is recommended as a complement:

```php
// ContainerConstants.php
final class ContainerConstants
{
    public const CONTAINER         = ContainerContract::class;
    public const USER_REPOSITORY   = UserRepositoryContract::class;
    public const DATABASE          = DatabaseContract::class;
}
```

The constants file provides a single place to look up every identifier in a
component without navigating the class hierarchy. Useful for config files,
serialization, and any context where importing the class itself is undesirable.

### Java

`.class` is the primary mechanism — compile-time type token (`Class<T>`).
Constants file recommended:

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

Note: Java's `IllegalArgumentException` is the language root for
`ValkyrjaInvalidArgumentException` — the naming uses `InvalidArgument` for
cross-port parity while the inheritance uses `IllegalArgument` for language
correctness.

### Go

No `::class` equivalent. String constants are the only mechanism. The constants
file is required:

```go
// container_constants.go
package container

const (
	ContainerClass      = "io.valkyrja.container.ContainerContract"
	UserRepositoryClass = "io.valkyrja.container.UserRepositoryContract"
	DatabaseClass       = "io.valkyrja.container.DatabaseContract"
)
```

Type safety is convention-enforced. The linter and code review are the
enforcement mechanisms. The string format `io.valkyrja.{component}.{ClassName}`
is the cross-port standard.

### Python

The class itself is a first-class object in Python — `ClassName` can be passed
directly. However for cross-language parity and to support CGI/lambda cache
generation, string constants are used:

```python
# container_constants.py
class ContainerConstants:
    CONTAINER = "io.valkyrja.container.ContainerContract"
    USER_REPOSITORY = "io.valkyrja.container.UserRepositoryContract"
    DATABASE = "io.valkyrja.container.DatabaseContract"
```

A FQN helper is available for cases where the string needs to be derived rather
than hardcoded:

```python
# available as a utility — not required
def class_(cls) -> str:
    return f"{cls.__module__}.{cls.__qualname__}"
```

Note: `class_` uses a trailing underscore because `class` is a reserved word in
Python.

### TypeScript

No class reference mechanism at runtime. String constants are required:

```typescript
// container-constants.ts
export const ContainerConstants = {
    CONTAINER: 'io.valkyrja.container.ContainerContract',
    USER_REPOSITORY: 'io.valkyrja.container.UserRepositoryContract',
    DATABASE: 'io.valkyrja.container.DatabaseContract',
} as const
```

TypeScript's `typeof` and `keyof` can be used to derive types from the constants
object for additional type safety:

```typescript
type ContainerKey = typeof ContainerConstants[keyof typeof ContainerConstants]
```

---

## The Container's Perspective

The container never needs to know anything about the class itself — only the key
and how to build it. This is a cleaner contract than reflection ever was:

```
Key (string)  +  Factory (closure)  =  Complete binding
```

The container stores the closure. When `make(key)` is called:

1. Look up the closure by key
2. Invoke the closure with the container as the argument
3. Return the result

No reflection. No dynamic dispatch. No assumptions. Just a function call.

---

## PHP and Java Migration Note

PHP and Java currently use `::class` / `.class` not just for binding keys but
also for dynamic instantiation via reflection. The migration path is:

1. Add constants files per component
2. Migrate bindings to closure-based factories referencing constants
3. Remove dynamic reflection/method resolution from the container
4. Document closure-based binding as the canonical pattern

The `::class` / `.class` syntax is retained as the value passed to the
constants (PHP/Java) — the language guarantees it refers to a real class. The
constants file just organizes those verified values in one place.

This migration makes every port's container architecture identical at the
behavioral level, with only the key type safety mechanism differing per
language.

---

## Discussion Summary

The container binding problem surfaced when analyzing what a Python or
TypeScript port of Valkyrja's container would look like. PHP's `::class` and
Java's `.class` serve a dual purpose in the current implementation: they provide
a binding key AND enable dynamic dispatch via reflection or dynamic method
calls. Neither capability exists in Go, TypeScript, or Python in any reliable
cross-deployment form.

The first insight was separating these two concerns: the binding key (a string
identifier) and the factory (how to build the class). These were conflated in
the original PHP/Java implementation because reflection made it possible to
derive the factory from the key. Separating them makes the architecture
language-agnostic.

The second insight was that closures solve the factory problem completely across
all languages. Every language supports first-class functions. A closure captures
its dependencies explicitly, executes without reflection, and is transparent —
you can read it and know exactly what will be constructed. This is strictly
better architecture than dynamic dispatch even in PHP and Java.

The third insight was the per-component constants file. The alternative — a
single central constants file in the container component — was rejected because
it creates exactly the kind of tight coupling that the component architecture is
designed to avoid. Per-component constants mean each component owns its
identifiers, contributing to the same isolation principles that govern the rest
of the framework.

The decision to recommend (but not require) constants files for PHP and Java was
made to keep the option open for full cross-language consistency without forcing
a breaking change. The practical benefits — a single lookup location, refactor
safety, grep-ability — make the constants file valuable in PHP and Java
independent of the cross-language parity argument.
