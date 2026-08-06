# COMMENTS.md — what a comment may state

The **cross-language** convention for comments and doc comments. It applies in
every Valkyrja port — PHP, Java, Go, Python, TypeScript, and Kotlin — and to
every contributor, human or agent.

Code speaks for itself. A comment adds what the code cannot show, so most code
carries no comment. A file where every block carries a comment is a wall of
text, and a reader skips walls. Comment the one line in ten that needs one, and
that warning stands alone. ([`AGENTS.md`](AGENTS.md) §3 states this as golden
rules 10 and 13, together with the rule that a comment never states a transient
condition.)

PHP examples are shown, because PHP is the reference implementation. The rules
hold in every port. The examples are generic by design: they show the shape in
the framework's naming style, and they copy no real method, so this document
does not drift when the source changes.

## A comment inside a method body

Clear code takes no comment. A comment explains what is unclear, or why the
code does something this particular way. When an alternative failed, say so, so
the next editor does not retry it.

```php
// Right — the comment states an ordering reason the code cannot show.
protected function matchRoute(string $path): Route|null
{
    // Match the static routes first, because a dynamic pattern can also match a static path.
    return $this->matchStatic($path)
        ?? $this->matchDynamic($path);
}
```

## The doc comment on a method

One sentence states what the method does. The annotations complete the
signature. The comment enhances the signature: it does not restate the
signature, and it does not describe the implementation.

```php
// Right — one sentence states what the method does, and the annotations
// complete the signature.
/**
 * Bind a service to the container.
 *
 * @template T of object
 *
 * @param class-string<T>                           $id       The service id
 * @param callable(self, array<array-key, mixed>):T $callable The callable
 */
public function bind(string $id, callable $callable): static;
```

A doc comment describes the method, not one implementation of the method. Every
override inherits the comment, so a sentence about one implementation becomes
false in the next override, and the next reader trusts the false sentence.

## The doc comment on an override

Keep the inherited block as it is. Change only what the new signature requires —
the parameter and return annotations. Never add a paragraph that says what this
implementation does. Every language inherits a doc comment: PHP `@inheritDoc`,
Java `{@inheritDoc}`, TypeScript `@inheritdoc`, Python's inherited docstring,
and Go's doc comment on the interface method.

```php
// Wrong — the paragraph describes this one override of get(). The next
// override of get() inherits the paragraph, and it is then false.
/**
 * @inheritDoc
 *
 * A miss reads through to the parent store, and the parent's value publishes
 * into this store.
 *
 * @param non-empty-string $key The cache key
 */
#[Override]
public function get(string $key): string|null
{
    return $this->store[$key]
        ?? $this->parent->get($key);
}
```

```php
// Right — the inherited block stays as it is. Only the annotations change,
// because the signature narrows the types.
/**
 * @inheritDoc
 *
 * @param non-empty-string $key The cache key
 */
#[Override]
public function get(string $key): string|null
{
    return $this->store[$key]
        ?? $this->parent->get($key);
}
```

When an override genuinely needs an explanation, put the explanation in the
pull request description. The squash merge writes the description as the commit
body, so the explanation stays attached to the commit that introduced the
override, reachable by `git log` and `git blame`.

## The doc comment on a type declaration

A class, a contract, a trait, an enum, and a struct carry no doc comment. The
declaration explains itself: the name and the segment say what the type is
([`AGENTS.md`](AGENTS.md) §4 encodes the kind in both), and each method's doc
comment says what the type does. When a type needs more than that, the
component's documentation takes it.

```php
// Wrong — the block re-teaches what the name and the methods already say.
/**
 * The redis cache.
 *
 * Stores each value in redis, and reads through to the parent store on a
 * miss.
 */
final class RedisCache implements CacheContract
```

```php
// Right — the declaration is bare. The methods explain the class.
final class RedisCache implements CacheContract
```

One exception: a **test fixture** carries a one-line comment block on the
declaration, because the comment says what the fixture is for, and the
fixture's methods cannot say it.

```php
// Right — a fixture is the exception, because the comment says what the
// fixture is for.
/**
 * Pass-through middleware used for unit testing.
 */
final class PassThroughMiddlewareFixture implements RouteMatchedMiddlewareContract
```
