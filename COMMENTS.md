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
hold in every port.

## A comment inside a method body

Clear code takes no comment. A comment explains what is unclear, or why the
code does something this particular way. When an alternative failed, say so, so
the next editor does not retry it.

```php
// Right — the comment states an ordering reason the code cannot show.
protected static function fromBackedEnum(string|int $value): static|null
{
    // Need to check BackedEnum first because all Enums are UnitEnum
    if (is_a(static::class, BackedEnum::class, true)) {
```

## The doc comment on a method

One sentence states what the method does. The annotations complete the
signature. The comment enhances the signature: it does not restate the
signature, and it does not describe the implementation.

```php
// Right — one sentence states what the method does, and the annotations
// complete the signature.
/**
 * Bind an alias to the container.
 *
 * @param class-string $alias The alias
 * @param class-string $id    The service id to alias
 */
public function bindAlias(string $alias, string $id): static;
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
// Wrong — the paragraph describes this one override of getAlias(). The next
// override of getAlias() inherits the paragraph, and it is then false.
/**
 * @inheritDoc
 *
 * A parent-only alias resolves its target through the child's own get(), so
 * a deferred target publishes into the child scope.
 *
 * @param class-string $id The service id
 *
 * @return class-string|null
 */
#[Override]
protected function getAlias(string $id): string|null
{
    return $this->aliases[$id]
        ?? $this->parent->aliases[$id]
        ?? null;
}
```

```php
// Right — NativeChildContainer::getAlias() keeps the inherited block. Only
// the annotations change, because the signature narrows the types.
/**
 * @inheritDoc
 *
 * @param class-string $id The service id
 *
 * @return class-string|null
 */
#[Override]
protected function getAlias(string $id): string|null
{
    return $this->aliases[$id]
        ?? $this->parent->aliases[$id]
        ?? null;
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
 * The native child container.
 *
 * Checks its own maps first, and falls back to the maps of the parent.
 */
class NativeChildContainer extends Container
```

```php
// Right — the declaration is bare. The methods explain the class.
class NativeChildContainer extends Container
```

One exception: a **test fixture** carries a one-line comment block on the
declaration, because the comment says what the fixture is for, and the
fixture's methods cannot say it.

```php
// Right — a fixture is the exception, because the comment says what the
// fixture is for.
/**
 * Attribute child class used for unit testing.
 */
final class AttributeClassChildFixture extends AttributeFixture
```
