# METHOD_NAMING.md — method name prefixes

The **cross-language** convention for naming a method. It applies in every Valkyrja
port — PHP, Java, Go, Python, TypeScript, and Kotlin.

A reader of a method name must be able to answer two questions without opening the
method:

1. **What does the method do?**
2. **Does the caller's own value change?**

The prefix answers both. This document defines each prefix, and gives the spelling for
each language.

---

## 1. The four families

### Validation

A validation method examines a value and reports the result. It never changes the value.

| Prefix | The method does | On failure |
|---|---|---|
| `validate{Something}` | Makes sure the value **is** valid | Throws |
| `invalidate{Something}` | Makes sure the value **is not** valid | Throws |
| `isValid{Something}` | Checks whether the value is valid | Returns `false` |

`validate` and `invalidate` return nothing. `isValid` returns a boolean.

**Warning: `invalidate` does not mean "make invalid".** In many codebases `invalidate`
clears a cache or expires a token. It does not do that here. It asserts that a value is
*not* valid, and throws when the value is valid. Do not use the name for anything else.

### Transformation that returns a copy

A `get` method returns a new value. The argument is unchanged.

| Prefix | The method returns |
|---|---|
| `get{Something}` | A modified copy of the value |
| `getFiltered{Something}` | A filtered copy of the value |
| `getParsed{Something}` | A parsed copy of the value |

### Transformation in place

These methods change the value the caller passed in. They return nothing.

| Prefix | The method does |
|---|---|
| `modify{Something}` | Modifies the value |
| `filter{Something}` | Filters the value |
| `parse{Something}` | Parses the value |
| `process{Something}` | Processes a collection |

**The pattern is regular.** A bare verb changes the argument and returns nothing. The
same verb behind `get` returns a copy and leaves the argument alone. `parse` and
`getParsed` are the same operation with opposite effects on the caller's value.

Read [§2](#2-in-place-transformation-does-not-exist-in-every-language) before you use
this family. It does not translate to every port.

### State on the host

These two change an object's own state. They are the pair a reader confuses most often,
so the difference matters.

| Prefix | The host | Returns |
|---|---|---|
| `set{Something}` | **Is modified** | Nothing, or the host itself for chaining |
| `with{Something}` | **Is not modified** | A **new** instance carrying the value |

`with` is the immutable form. It clones the host, sets the value on the clone, and
returns the clone. The original is untouched, so a caller that holds a reference to it
sees no change.

---

## 2. In-place transformation does not exist in every language

**Warning: the in-place family in [§1](#transformation-in-place) assumes the language can
modify the caller's variable. Most Valkyrja languages cannot do this for every type.**

PHP takes a parameter by reference with `&`. No other Valkyrja language has that.

| Language | Can modify the caller's value | How |
|---|---|---|
| PHP | Any type | `&$value` |
| Java | A mutable object only | Mutate the object the reference points to |
| TypeScript | A mutable object only | Mutate the object the reference points to |
| Python | A mutable object only | Mutate the object the reference points to |
| Go | Any type | A pointer parameter, `*T` |

A string, a number, and a boolean are immutable in Java, TypeScript, and Python. A method
there cannot change the caller's string.

**The rule that follows: use the in-place family only when the language can honor it.
Otherwise use the `get` form, which every language can express.** A `parseValue(String
value)` in Java that claims to parse in place is a lie the type system permits, and the
caller finds out at runtime.

---

## 3. The spelling per language

The convention is the same in every port. Only the casing and the parameter form change.

| Language | Case | `isValid` | `getParsed` | `with` | In place |
|---|---|---|---|---|---|
| PHP | camelCase | `isValidPath` | `getParsedPath` | `withPath(): static` | `parsePath(string &$path): void` |
| Java | camelCase | `isValidPath` | `getParsedPath` | `withPath()` returns its own type | Mutable argument only |
| TypeScript | camelCase | `isValidPath` | `getParsedPath` | `withPath(): this` | Mutable argument only |
| Kotlin | camelCase | `isValidPath` | `getParsedPath` | `withPath()` returns its own type | Mutable argument only |
| Python | snake_case | `is_valid_path` | `get_parsed_path` | `with_path()` | Mutable argument only |
| Go | PascalCase when exported | `IsValidPath` | `GetParsedPath` | `WithPath()` returns a new value | `ParsePath(path *string)` |

Two notes on the edges of the table.

- **Go exports by capitalization.** An exported method starts with an upper-case letter,
  so every prefix above is capitalized. Go's own idiom usually drops a `Get` prefix
  (`x.Name()` rather than `x.GetName()`). Valkyrja does not follow that idiom, because
  the prefix carries meaning here — `GetParsedPath` and `ParsePath` are different
  operations, and dropping the prefix erases the difference.
- **Python uses snake_case throughout,** so each prefix is separated by an underscore.

### `with` returns the late-bound type

`with` clones the host, so it must return the **type of the object it was called on**,
not the type that declares the method. A subclass that inherits `withPath` must get its
own type back.

PHP spells this `static`, not `self`. `self` resolves to the declaring class and breaks
the subclass:

```php
// Wrong — `self` returns the declaring class, so a subclass loses its own type.
public function withPath(string $path): self
```

```php
// Right — `static` is late-bound, so a subclass gets its own type back.
public function withPath(string $path): static
{
    $new = clone $this;

    $new->path = $path;

    return $new;
}
```

**Warning: a `@return static` docblock on a templated contract is load-bearing.** Psalm
needs it for subtype inference, and Rector's `RemoveUselessReturnTagRector` deletes a tag
it reads as redundant. Give the tag a description so the rule keeps it.

---

## 4. Examples

Show the wrong form first, then the right form. PHP is the reference implementation, so
every example is PHP unless a language spells it differently.

The parse pair — the bare verb changes the argument, the `get` form does not:

```php
// Wrong — the name says "get", so the caller does not expect $path to change.
public function getParsedPath(string &$path): void
```

```php
// Right — "parse" changes the argument; "getParsed" returns a copy.
public function parsePath(string &$path): void

public function getParsedPath(string $path): string
```

The `set` and `with` pair — one modifies the host, the other does not:

```php
// Wrong — the name says "with", so the caller does not expect $this to change.
public function withPath(string $path): static
{
    $this->path = $path;

    return $this;
}
```

```php
// Right — "with" clones; "set" modifies the host.
public function withPath(string $path): static
{
    $new = clone $this;

    $new->path = $path;

    return $new;
}

public function setPath(string $path): void
{
    $this->path = $path;
}
```

Validation — `validate` throws, `isValid` reports:

```php
// Right — the prefix says which one the caller gets.
public function validatePath(string $path): void;   // throws UriInvalidPathException

public function isValidPath(string $path): bool;    // returns false
```

---

## 5. What this document does not cover

Three naming rules live elsewhere. They govern a different axis and do not repeat here.

- **The class name, the segment, and the modifier** — the structure taxonomy in
  [`AGENTS.md`](AGENTS.md) §4.
- **A static method on a data object** — a named constructor stays on its own type;
  anything else moves to a factory or a support class. See
  [`STATIC_METHODS.md`](STATIC_METHODS.md).
- **An exception name** — [`THROWABLES.md`](THROWABLES.md).

### An open question, recorded rather than decided

`php/TODO.md` carries a worked-through position on a second naming family, for a method
that retrieves one item: `Create…` and `Get…` and `Retrieve…` throw when the item is
absent, `GetOrCreate…` creates it, and `Find…` returns null. A collection query always
returns a collection, empty when nothing matches.

That family is a natural fit for this document, and the reasoning behind it is explicitly
cross-language — it exists to give parity with Java and Go. **It is not adopted here,
because it has not been decided.** Read the record in `php/TODO.md`, under "Is returning
null cheating?", before you rely on it.
