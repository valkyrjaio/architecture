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

Since no other language can enforce or dynamically dispatch static interface
methods, **PHP is being updated to use instance methods throughout** — removing
the PHP-specific static dispatch entirely so all ports look and behave the same.

---

## Two Categories of Static PHP Patterns

### 1. Static Factory Methods (`from*`, `create*`)

PHP pattern:
```php
$entity = MyEntity::fromValue($raw);
```

These create new instances of a type from a raw value. The cross-language
solution is **container-registered factories**: the developer explicitly
registers a callable that creates the type, and the framework calls it.

```java
// Registered in a service provider
container.bind(MyEntity.class, raw -> new MyEntity(raw));

// Framework resolves via container — no reflection, no convention
MyEntity entity = container.make(MyEntity.class, raw);
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
