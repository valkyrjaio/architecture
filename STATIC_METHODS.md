# Static Interface Methods — Cross-Language Design

## The Problem

PHP allows interface methods to be declared `static`, and implementations can be
dispatched dynamically:

```php
interface EntityContract {
    public static function getTable(): string;
    public static function fromValue(mixed $value): static;
}

// called dynamically — $class is a string or ::class constant
$table = $class::getTable();
$entity = $class::fromValue($raw);
```

No other Valkyrja target language supports this:

| Language       | Static interface methods | Notes                                                   |
|----------------|--------------------------|----------------------------------------------------------|
| **PHP**        | Yes                      | Reference implementation — being changed for consistency |
| **Java**       | No                       | Static interface methods exist but cannot be overridden  |
| **TypeScript** | No                       | Same limitation as Java                                  |
| **Go**         | No                       | No static methods at all; interfaces are instance-only   |
| **Python**     | Partial                  | `typing.Protocol` can express it, but runtime-only check |

No other language can enforce or dynamically dispatch a static interface method.
**PHP therefore removes the static method that the framework calls on a variable
class.** This document covers that call, and it covers nothing else.

A named constructor is not that call. A named constructor returns its own type,
and the caller names the class. `StringT::fromValue($value)` ports to every
language, so the data object keeps it. See "What a data object holds" in
[`AGENTS.md`](AGENTS.md) §4.

---

## Two Categories of Static PHP Patterns

### 1. Construction Through a Variable Class

PHP pattern:
```php
// $type holds a class name. Only PHP can make this call.
$instance = $type::fromValue($raw);
```

The framework holds the class in a variable, and it must create an instance of
that class. The cross-language solution is **container-registered factories**:
the developer explicitly registers a callable that creates the type, and the
framework calls it.

A direct call such as `StringT::fromValue($raw)` is not this pattern. The caller
names the class, so every language can write it.

```java
// Registered in a service provider
container.bind(MyEntity.class, (c, args) -> new MyEntity(args.get("raw")));

// Framework resolves via container — no reflection, no convention
MyEntity entity = container.getService(MyEntity.class, Map.of("raw", raw));
```

The developer owns the creation logic. If no factory is registered, the
container throws a resolution error — explicit failure, not silent fallback.

### 2. Static Metadata (`getTable`, `getPrimaryKey`, etc.)

PHP pattern:
```php
$table = MyEntity::getTable();
$key   = MyEntity::getPrimaryKey();
```

These return read-only metadata that describes the type — table names, primary
keys, column mappings, and similar. They carry no instance state and are
consulted by the framework at query-build time.

The cross-language solution is an **entity metadata registry** — a map of class
token to metadata object, registered by the developer in a service provider,
similar to how routes are registered via `getRoutes()`.

```java
// EntityMetadata carries all static facts about the entity
EntityMetadata metadata = new EntityMetadata(
    "users",       // table
    "id",          // primary key
    // column map, soft-delete column, timestamps, etc.
);

// Registered in a service provider
container.setSingleton(EntityMetadata.class, MyEntity.class, metadata);

// Framework looks up at query time — no static dispatch
EntityMetadata meta = container.getSingleton(EntityMetadata.class, MyEntity.class);
String table = meta.getTable();
```

This is the same pattern used for routes: the developer provides a list of data
objects; the framework stores and queries them by key. Nothing is inferred by
convention.

---

## Why Not Annotations / Decorators?

Annotations (`@Entity(table = "users")` in Java, `@entity` in Python, struct tags
in Go) can express static metadata compactly. However, they do not generalize
cleanly across all five languages, and they push metadata discovery into the
framework's reflection layer — the same magic the registry approach avoids.

The registry approach is consistent across every port: the developer writes a
service provider, calls `setSingleton`, done. No reflection, no annotation
scanning, no per-language metadata API.

Annotations and decorators remain available for ergonomic sugar in language-
specific implementations, but the canonical registration path is always the
registry.

---

## Summary

| PHP pattern                 | Cross-language equivalent                           |
|-----------------------------|------------------------------------------------------|
| `$class::fromValue($value)` | Container-registered factory                         |
| `$class::getTable()`        | Entity metadata registry keyed by class token        |
| `$class::getX()` generally  | Registry lookup via container, keyed by class token  |

The rule: **if PHP would call it statically on a variable class, every other
language needs an explicit registration**. The developer declares how the
framework finds or creates the value; the framework never guesses.

Each row holds `$class`, not a class name. `StringT::fromValue($value)` names the
class, so it needs no registration and it stays.
