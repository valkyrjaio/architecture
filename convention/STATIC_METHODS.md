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

| Language       | Static interface methods | Notes                                                    |
| -------------- | ------------------------ | -------------------------------------------------------- |
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
language, so the data object keeps it. See "What a Data Object Holds" below.

---

## What a Data Object Holds

A data object holds data and the behavior of its own state. A message, a model,
an entity, a type, and a config are all data objects.

A data object does not hold a static method. A static method describes the type,
not the state, so it belongs to another class. Three rules place it:

- A **factory** takes construction. A named constructor such as `create` or
  `fromArray` belongs on the `*Factory` for that type.
- A **support class** takes a calculation or a rendering.
- An **enum** is exempt. Every language gives an enum static members of its own
  (PHP `from`, `cases`), and a method on an enum case reads that case.

One static method stays on the data object: a named constructor that returns its
own type. `Header::fromValue()` returns a `Header`, so `Header` keeps it. The
rule removes a static method that returns something else — a field name, a table
name, a format, or a validation result.

```php
// Wrong — the trait puts a rendering on the entity, and it returns a string.
trait Dateable
{
    public static function getFormattedDate(): string
    {
        return DateFactory::getFormattedDate(static::getDateFormat());
    }
}
```

```php
// Right — the factory renders the date. The entity does not repeat the factory.
$date = DateFactory::getFormattedDate(DateFormat::DEFAULT);
```

```php
// Right — the named constructor returns its own type, so the type keeps it.
class SlugT extends Type implements SlugContract
{
    public static function fromValue(mixed $value): static
    {
        return new static(SlugFactory::fromMixed($value));
    }
}
```

The return type is `static`, so a subclass gets its own type back. The body
delegates the conversion to the `*Factory` for the type.

The rule reaches every segment that puts a static method on a data object. A
`Contract\` interface declares the method, an `Abstract\` base holds a default,
and a `Trait\` trait mixes the method in. Each one obeys the rule.

Static metadata is the hard case. A method such as `getTableName` describes the
type and carries no instance state. The Static Metadata section below records
the replacement: a metadata registry that the developer fills in a service
provider, keyed by class token.

Warning: a call on a _variable_ class does not port, whatever the method
returns. PHP writes `$type::fromValue($value)`, and no other Valkyrja language
can. The exemption above covers the declaration, not that call. The
Construction Through a Variable Class section below owns the replacement for
it.

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
| --------------------------- | --------------------------------------------------- |
| `$class::fromValue($value)` | Container-registered factory                        |
| `$class::getTable()`        | Entity metadata registry keyed by class token       |
| `$class::getX()` generally  | Registry lookup via container, keyed by class token |

The rule: **if PHP would call it statically on a variable class, every other
language needs an explicit registration**. The developer declares how the
framework finds or creates the value; the framework never guesses.

Each row holds `$class`, not a class name. `StringT::fromValue($value)` names the
class, so it needs no registration and it stays.
