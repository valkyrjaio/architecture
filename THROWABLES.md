# Valkyrja Throwables & Exception Hierarchy

## Naming Convention

### The Single Rule

> **Prepend only as many parent names as needed to make the class name unique across the entire framework.**

This rule is recursive and applies at every level of the hierarchy.

### The Suffix Convention

The `*` in `ComponentName*` represents a language-native suffix. The suffix is chosen per language to match what major
frameworks in that ecosystem use:

| Language   | Suffix       | Major frameworks                                 |
|------------|--------------|--------------------------------------------------|
| PHP        | `*Exception` | Laravel, Symfony                                 |
| Java       | `*Exception` | Spring                                           |
| Kotlin     | `*Exception` | same roots as Java                               |
| Scala      | `*Exception` | same roots as Java                               |
| Python     | `*Exception` | FastAPI, Django, NestJS-equivalent               |
| TypeScript | `*Exception` | NestJS                                           |
| Go         | `*Error`     | idiomatic Go — `*Exception` does not exist in Go |

Go is the only port that uses `*Error`. The Go standard library itself uses no suffix at all (`io.EOF`,
`os.ErrNotExist`) but `*Error` is more conventional for custom framework error types than `*Exception` which is entirely
foreign to Go.

The **stem** is always identical across all ports — `ContainerNotFound`, `HttpRoutingNotFound`, `RequestInvalidMethod`.
Only the suffix differs for Go. Cross-port legibility is preserved at the stem level which is the meaningful part.

**Examples across ports:**

| PHP / Java / Python / TypeScript | Go                          |
|----------------------------------|-----------------------------|
| `ContainerNotFoundException`     | `ContainerNotFoundError`    |
| `HttpRoutingNotFoundException`   | `HttpRoutingNotFoundError`  |
| `RequestInvalidMethodException`  | `RequestInvalidMethodError` |
| `ValkyrjaRuntimeException`       | `ValkyrjaRuntimeError`      |

### The Four Levels

| Level                 | Convention                     | Notes                                                                 |
|-----------------------|--------------------------------|-----------------------------------------------------------------------|
| Framework base        | `Valkyrja*`                    | abstract · extends language root                                      |
| Component             | `ComponentName*`               | abstract · always present · extends `Valkyrja*`                       |
| Subcomponent (unique) | `SubComponent*`                | name is unique across the entire framework                            |
| Subcomponent (shared) | `ParentComponentSubComponent*` | name exists in multiple components (e.g. Routing in Http, Cli, Queue) |

### The Recursive Rule Applied

At every depth, ask one question: **does this name already exist somewhere else in the framework?**

- **No** → use the name as-is, stop
- **Yes** → prepend the immediate parent name, ask again

This terminates naturally as soon as the name is unique.

**Sub-subcomponent examples** (using PHP/Java/Python/TypeScript suffix):

- `Http\Message` — `Message` is shared with gRPC → `HttpMessageRuntimeException`
- `Http\Message\Request` — `Request` is unique (Cli uses `Input`) → `RequestRuntimeException`
- `Http\Message\Response` — `Response` is unique (Cli uses `Output`) → `ResponseRuntimeException`
- `Http\Message\Stream` — `Stream` is not unique (future filesystem/queue/gRPC) → `HttpStreamRuntimeException`
- `Http\Message\Header` — `Header` is not unique (email/queue/gRPC) → `HttpHeaderRuntimeException`
- `Http\Message\File` — renamed to `UploadedFile` (unique, removes ambiguity with filesystem) →
  `UploadedFileRuntimeException`
- `Http\Routing` — `Routing` is shared with Cli and Queue → `HttpRoutingRuntimeException`
- `Cli\Routing` — shared, different parent → `CliRoutingRuntimeException`

Go uses `*Error` suffix throughout: `HttpMessageRuntimeError`, `RequestRuntimeError`, `HttpStreamRuntimeError` etc.

---

## Abstract Enforcement

Every base and categorical exception is **abstract**. This is enforced at the language level, not by convention.

- **PHP** — `abstract class` prevents `new ValkyrjaRuntimeException()`
- **Java** — `abstract class` is a compile-time error if instantiated directly
- **Python** — ABC raises `TypeError` on direct instantiation
- **TypeScript** — `abstract class` is a compile-time TypeScript error
- **Go** — unexported interface/struct with unexported embedded field prevents external instantiation

No abstract exception should ever appear in a `throw` / `raise` statement. Only concrete specific exceptions are thrown.

---

## Always-Present Component Categoricals

Every component ships with the following abstract exceptions regardless of whether they are currently used:

```
ComponentInvalidArgumentException  (abstract)
ComponentRuntimeException          (abstract)
```

**Why always present even if unused:**

- First-party components can add specific exceptions later without a breaking change
- Third-party package contributors have a clear, consistent extension point from day one
- Ecosystem consistency — every Valkyrja package looks the same regardless of who wrote it
- Adding exceptions is always additive, never structural

Component categorical exceptions are only concretely sub-classed when two or more specific exceptions of that category
exist and a developer might reasonably want to catch them together. A single specific exception extends the framework
base directly.

---

## The Full Hierarchy

The `*Exception` suffix applies to PHP, Java, Python, and TypeScript. Go uses `*Error` throughout.

```
Language root (e.g. \Throwable)
└── ValkyrjaThrowable                              (abstract)
    └── ComponentThrowable                         (abstract · always present)
        └── ComponentSpecificThrowable             (concrete · as needed)

Language root (e.g. \RuntimeException)
└── ValkyrjaRuntimeException                       (abstract)
    └── ComponentRuntimeException                  (abstract · always present)
        └── ComponentSpecificException             (concrete · as needed)

Language root (e.g. \InvalidArgumentException)
└── ValkyrjaInvalidArgumentException               (abstract)
    └── ComponentInvalidArgumentException          (abstract · always present)
        └── ComponentSpecificInvalidArgument       (concrete · as needed)
            Exception / Error
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

Go equivalents: `ContainerNotFoundError`, `ContainerBindingError`, `ContainerInvalidBindingArgumentError` etc.

### Http — Shared Subcomponent (Routing)

`Routing` exists in Http, Cli, and future Queue — prefix required.

```
ValkyrjaRuntimeException (abstract)
└── HttpRoutingRuntimeException (abstract · always present)
    └── HttpRoutingNotFoundException (concrete)

ValkyrjaRuntimeException (abstract)
└── CliRoutingRuntimeException (abstract · always present)
    └── CliRoutingRouteNotMatchedException (concrete)
```

### Http\Message Subcomponents

`Message` is shared with gRPC — prefix required.
`Stream` and `Header` are not unique — prefix required.
`Request` and `Response` are unique (Cli uses Input/Output) — stand alone.
`File` renamed to `UploadedFile` — unique, disambiguates from filesystem.

```
ValkyrjaRuntimeException (abstract)
└── HttpMessageRuntimeException (abstract · always present)
    └── HttpMessageNotFoundException (concrete)

└── HttpStreamRuntimeException (abstract · always present)
    └── HttpStreamReadException (concrete)

└── HttpHeaderRuntimeException (abstract · always present)
    └── HttpHeaderInvalidException (concrete)

└── RequestRuntimeException (abstract · always present)
    └── RequestInvalidMethodException (concrete)

└── ResponseRuntimeException (abstract · always present)
    └── ResponseInvalidStatusException (concrete)

└── UploadedFileRuntimeException (abstract · always present)
    └── UploadedFileInvalidMimeException (concrete)
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

Java has no `InvalidArgumentException` — the idiomatic root is `IllegalArgumentException`. Valkyrja names the framework
base `ValkyrjaInvalidArgumentException` for cross-port parity while extending `IllegalArgumentException` under the hood:

```java
// extends java.lang.IllegalArgumentException for language-level catchability
// named ValkyrjaInvalidArgumentException for cross-port parity
public abstract class ValkyrjaInvalidArgumentException
        extends IllegalArgumentException {
}
```

Kotlin and Scala map 1:1 with Java. All exceptions are unchecked by default in Kotlin — no `throws` declarations needed.

### Go Note

Go has no class hierarchy. Three branches are maintained for parity. Go uses `*Error` suffix throughout — `*Exception`
is foreign to the Go ecosystem. The Go standard library uses no suffix (`io.EOF`, `os.ErrNotExist`) but `*Error` is the
most conventional choice for custom framework error types.

```go
// ValkyrjaThrowable — unexported interface (no suffix needed for interface)
type valkyrjaThrowable interface {
error
isValkyrjaThrowable()
}

// ValkyrjaRuntimeError — exported struct (*Error suffix in Go)
type ValkyrjaRuntimeError struct {
valkyrjaThrowable // unexported embedded field prevents external instantiation
message string
}

// ValkyrjaInvalidArgumentError — exported struct
type ValkyrjaInvalidArgumentError struct {
valkyrjaThrowable
message string
}
```

Component categoricals follow the same `*Error` suffix:

```go
// always present per component — unexported interface
type containerRuntimeError interface {
error
isContainerRuntimeError()
}

// concrete — exported struct
type ContainerNotFoundError struct {
ValkyrjaRuntimeError
}
```

Callers use `errors.As` / `errors.Is` as the catch boundary:

```go
var target *ContainerNotFoundError
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

All three branches extend `Error` — TypeScript has no distinct `RuntimeError` or `InvalidArgumentError` built-ins:

```typescript
export abstract class ValkyrjaThrowable extends Error {
}

export abstract class ValkyrjaRuntimeException extends Error {
}

export abstract class ValkyrjaInvalidArgumentException extends Error {
}
```

No typed throws on function signatures — hierarchy is enforced structurally via `abstract class`, not by the compiler.
Result pattern available as additive opt-in:

```typescript
type Result<T, E extends Error> =
    | { success: true; value: T }
    | { success: false; error: E }

// opt-in alongside standard throw/catch
function tryMake<T>(abstract: string): Result<T, ContainerException> {
}
```

---

## Suffix Decision — *Exception vs *Error

The question of whether to use `*Exception` or `*Error` as the class suffix across ports was investigated by examining
what major frameworks in each ecosystem actually use:

- **Go** major frameworks (Gin, Echo, Fiber) — mixed conventions, no `*Exception` anywhere. Go's standard library itself
  uses no suffix or `Err*` prefix for sentinels. `*Error` is the most idiomatic choice for custom framework error types.
- **Python** major frameworks (FastAPI, Django) — predominantly `*Exception`. FastAPI uses `HTTPException`,
  `RequestValidationError`. Python's own library convention (requests, SQLAlchemy) uses `*Error` suffix, creating a
  community split.
- **TypeScript / NestJS** — `*Exception` throughout. NestJS uses `HttpException`, `NotFoundException`,
  `BadRequestException` as its core naming.
- **PHP / Java** — `*Exception` is the universal convention in both ecosystems.

**Decision:** Language-native suffix per port. Four of five language ecosystems (PHP, Java, Python, TypeScript) use
`*Exception` as their dominant convention in major frameworks. Go is the exception — `*Exception` is entirely foreign
there, and `*Error` is the idiomatic Go choice.

The stem is always identical across all ports (`ContainerNotFound`, `HttpRoutingNotFound`) — only the suffix differs for
Go. Cross-port legibility is preserved at the stem level which is the meaningful part.

## Discussion Summary

The naming convention emerged from a Spotless static analysis warning in the Java port flagging same-named exceptions
across packages. This surfaced a broader problem: exception class names like `InvalidArgumentException` and
`RuntimeException` were mirroring language root names, making it impossible to know which component an exception
originated from by name alone — you had to check the namespace.

The first decision was to prefix framework base exceptions with `Valkyrja*` and component exceptions with
`ComponentName*`. This immediately resolved the ambiguity — a `ContainerNotFoundException` is unambiguous in any
context: stack traces, catch blocks, IDE autocomplete, static analysis.

The second decision was to make all base and categorical exceptions abstract. This was driven by the realization that
`ContainerException` as a directly throwable class is too generic to be useful — if you're throwing it, you should be
throwing something more specific that tells you what actually went wrong. Abstract enforcement at the language level
prevents misuse without relying on convention.

The third decision — always present component categoricals — came from thinking about third-party package contributors.
A package extending the Container component should always have `ContainerRuntimeException` available to extend, even if
the core framework hasn't needed it yet. Pre-building the scaffold prevents breaking changes later.

The shared subcomponent naming problem arose from Routing, Middleware, and Server being shared across Http, Cli, and the
future Queue component. `RoutingRuntimeException` alone doesn't tell you if you're in HTTP or CLI routing context. The
solution — `HttpRoutingRuntimeException` / `CliRoutingRuntimeException` — was reached by applying the same uniqueness
rule recursively: prepend parent names until the name is unique across the framework.

The recursive uniqueness rule was the final unifying principle that made the entire convention self-consistent at any
depth. Rather than a lookup table of special cases, every naming decision at every level answers the same single
question: is this name unique across the framework? If no, prepend the immediate parent and ask again.
