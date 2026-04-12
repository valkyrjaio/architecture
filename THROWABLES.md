# Valkyrja Throwables & Exception Hierarchy

## Naming Convention

### The Single Rule

> **Prepend only as many parent names as needed to make the class name unique
across the entire framework.**

This rule is recursive and applies at every level of the hierarchy.

### The Four Levels

| Level                 | Convention                     | Notes                                                                 |
|-----------------------|--------------------------------|-----------------------------------------------------------------------|
| Framework base        | `Valkyrja*`                    | abstract · extends language root                                      |
| Component             | `ComponentName*`               | abstract · always present · extends `Valkyrja*`                       |
| Subcomponent (unique) | `SubComponent*`                | name is unique across the entire framework                            |
| Subcomponent (shared) | `ParentComponentSubComponent*` | name exists in multiple components (e.g. Routing in Http, Cli, Queue) |

### The Recursive Rule Applied

At every depth, ask one question: **does this name already exist somewhere else
in the framework?**

- **No** → use the name as-is, stop
- **Yes** → prepend the immediate parent name, ask again

This terminates naturally as soon as the name is unique.

**Sub-subcomponent examples:**

- `Http\Message\Request` — both `Message` and `Request` are shared across
  components → `HttpMessageRequestRuntimeException`
- `Http\Message\Stream` — `Stream` is unique across the framework →
  `StreamRuntimeException`
- `Http\Message\Request\Body` — `Body` is unique across the framework →
  `BodyRuntimeException`

---

## Abstract Enforcement

Every base and categorical exception is **abstract**. This is enforced at the
language level, not by convention.

- **PHP** — `abstract class` prevents `new ValkyrjaRuntimeException()`
- **Java** — `abstract class` is a compile-time error if instantiated directly
- **Python** — ABC raises `TypeError` on direct instantiation
- **TypeScript** — `abstract class` is a compile-time TypeScript error
- **Go** — unexported interface/struct with unexported embedded field prevents
  external instantiation

No abstract exception should ever appear in a `throw` / `raise` statement. Only
concrete specific exceptions are thrown.

---

## Always-Present Component Categoricals

Every component ships with the following abstract exceptions regardless of
whether they are currently used:

```
ComponentInvalidArgumentException  (abstract)
ComponentRuntimeException          (abstract)
```

**Why always present even if unused:**

- First-party components can add specific exceptions later without a breaking
  change
- Third-party package contributors have a clear, consistent extension point from
  day one
- Ecosystem consistency — every Valkyrja package looks the same regardless of
  who wrote it
- Adding exceptions is always additive, never structural

Component categorical exceptions are only concretely sub-classed when two or
more specific exceptions of that category exist and a developer might reasonably
want to catch them together. A single specific exception extends the framework
base directly.

---

## The Full Hierarchy

```
Language root (e.g. \Throwable)
└── ValkyrjaThrowable                          (abstract)
    └── ComponentThrowable                     (abstract · always present)
        └── ComponentSpecificThrowable         (concrete · as needed)

Language root (e.g. \RuntimeException)
└── ValkyrjaRuntimeException                   (abstract)
    └── ComponentRuntimeException              (abstract · always present)
        └── ComponentSpecificException         (concrete · as needed)

Language root (e.g. \InvalidArgumentException)
└── ValkyrjaInvalidArgumentException           (abstract)
    └── ComponentInvalidArgumentException      (abstract · always present)
        └── ComponentSpecificInvalidArgument   (concrete · as needed)
            Exception
```

---

## Catch Boundaries

The hierarchy provides meaningful catch boundaries at every level:

```php
// language-wide — catch all invalid argument exceptions across everything
catch (\InvalidArgumentException $e) {}

// framework-wide — catch anything Valkyrja throws
catch (ValkyrjaThrowable $e) {}

// framework category-wide — catch all runtime exceptions from Valkyrja
catch (ValkyrjaRuntimeException $e) {}

// component category-wide — catch all runtime exceptions from Container
catch (ContainerRuntimeException $e) {}

// specific — the most common real-world usage
catch (ContainerNotFoundException $e) {}
```

---

## Concrete Examples

### Container Component

```
ValkyrjaRuntimeException (abstract)
└── ContainerRuntimeException (abstract · always present)
    └── ContainerNotFoundException (concrete)
    └── ContainerBindingException (concrete)

ValkyrjaInvalidArgumentException (abstract)
└── ContainerInvalidArgumentException (abstract · always present)
    └── ContainerInvalidBindingArgumentException (concrete)
    └── ContainerInvalidAliasArgumentException (concrete)
```

### Http — Shared Subcomponent (Routing)

```
ValkyrjaRuntimeException (abstract)
└── HttpRoutingRuntimeException (abstract · always present)
    └── HttpRoutingNotFoundException (concrete)

ValkyrjaRuntimeException (abstract)
└── CliRoutingRuntimeException (abstract · always present)
    └── CliRoutingRouteNotMatchedException (concrete)
```

### Http — Unique Subcomponent (Request)

```
ValkyrjaRuntimeException (abstract)
└── RequestRuntimeException (abstract · always present)
    └── RequestInvalidMethodException (concrete)

ValkyrjaInvalidArgumentException (abstract)
└── RequestInvalidArgumentException (abstract · always present)
    └── RequestInvalidUriArgumentException (concrete)
```

---

## Contributor Decision Tree

1. **Is the subcomponent name shared across multiple components?**
    - Yes → `ParentComponentSubComponent*`
    - No → `SubComponent*`

2. **Does the subcomponent have 2+ exceptions of the same category?**
    - Yes → add abstract `*RuntimeException` / `*InvalidArgumentException`
    - No → extend framework base directly

3. **Does each throw site need its own exception?**
    - Always yes → concrete, named for the problem, never throw an abstract

---

## Language Root Mapping

| Concept                  | PHP                         | Java                       | Go                  | Python          | TypeScript |
|--------------------------|-----------------------------|----------------------------|---------------------|-----------------|------------|
| Throwable                | `\Throwable`                | `Throwable`                | `error` (interface) | `BaseException` | `Error`    |
| RuntimeException         | `\RuntimeException`         | `RuntimeException`         | `error` (struct)    | `RuntimeError`  | `Error`    |
| InvalidArgumentException | `\InvalidArgumentException` | `IllegalArgumentException` | `error` (struct)    | `ValueError`    | `Error`    |

### Java / Kotlin / Scala Note

Java has no `InvalidArgumentException` — the idiomatic root is
`IllegalArgumentException`. Valkyrja names the framework base
`ValkyrjaInvalidArgumentException` for cross-port parity while extending
`IllegalArgumentException` under the hood:

```java
// extends java.lang.IllegalArgumentException for language-level catchability
// named ValkyrjaInvalidArgumentException for cross-port parity
public abstract class ValkyrjaInvalidArgumentException
        extends IllegalArgumentException {
}
```

Kotlin and Scala map 1:1 with Java. All exceptions are unchecked by default in
Kotlin — no `throws` declarations needed.

### Go Note

Go has no class hierarchy. Three branches are maintained for parity:

```go
// ValkyrjaThrowable — unexported interface
type valkyrjaThrowable interface {
error
isValkyrjaThrowable()
}

// ValkyrjaRuntimeException — exported struct
type ValkyrjaRuntimeException struct {
valkyrjaThrowable // unexported embedded field prevents external instantiation
message string
}

// ValkyrjaInvalidArgumentException — exported struct
type ValkyrjaInvalidArgumentException struct {
valkyrjaThrowable
message string
}
```

Callers use `errors.As` / `errors.Is` as the catch boundary:

```go
var target *ContainerNotFoundException
if errors.As(err, &target) {
// handle specifically
}
```

### Python Note

Abstract enforcement via ABC:

```python
from abc import ABC


class ValkyrjaThrowable(BaseException, ABC):
    pass


class ValkyrjaRuntimeException(RuntimeError, ABC):
    pass


# parity name — extends ValueError for language-level catchability
class ValkyrjaInvalidArgumentException(ValueError, ABC):
    pass
```

### TypeScript Note

All three branches extend `Error` — TypeScript has no distinct `RuntimeError` or
`InvalidArgumentError` built-ins:

```typescript
export abstract class ValkyrjaThrowable extends Error {
}

export abstract class ValkyrjaRuntimeException extends Error {
}

export abstract class ValkyrjaInvalidArgumentException extends Error {
}
```

No typed throws on function signatures — hierarchy is enforced structurally via
`abstract class`, not by the compiler. Result pattern available as additive
opt-in:

```typescript
type Result<T, E extends Error> =
    | { success: true; value: T }
    | { success: false; error: E }

// opt-in alongside standard throw/catch
function tryMake<T>(abstract: string): Result<T, ContainerException> {
}
```

---

## Discussion Summary

The naming convention emerged from a Spotless static analysis warning in the
Java port flagging same-named exceptions across packages. This surfaced a
broader problem: exception class names like `InvalidArgumentException` and
`RuntimeException` were mirroring language root names, making it impossible to
know which component an exception originated from by name alone — you had to
check the namespace.

The first decision was to prefix framework base exceptions with `Valkyrja*` and
component exceptions with `ComponentName*`. This immediately resolved the
ambiguity — a `ContainerNotFoundException` is unambiguous in any context: stack
traces, catch blocks, IDE autocomplete, static analysis.

The second decision was to make all base and categorical exceptions abstract.
This was driven by the realization that `ContainerException` as a directly
throwable class is too generic to be useful — if you're throwing it, you should
be throwing something more specific that tells you what actually went wrong.
Abstract enforcement at the language level prevents misuse without relying on
convention.

The third decision — always present component categoricals — came from thinking
about third-party package contributors. A package extending the Container
component should always have `ContainerRuntimeException` available to extend,
even if the core framework hasn't needed it yet. Pre-building the scaffold
prevents breaking changes later.

The shared subcomponent naming problem arose from Routing, Middleware, and
Server being shared across Http, Cli, and the future Queue component.
`RoutingRuntimeException` alone doesn't tell you if you're in HTTP or CLI
routing context. The solution — `HttpRoutingRuntimeException` /
`CliRoutingRuntimeException` — was reached by applying the same uniqueness rule
recursively: prepend parent names until the name is unique across the framework.

The recursive uniqueness rule was the final unifying principle that made the
entire convention self-consistent at any depth. Rather than a lookup table of
special cases, every naming decision at every level answers the same single
question: is this name unique across the framework? If no, prepend the immediate
parent and ask again.