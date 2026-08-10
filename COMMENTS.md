# COMMENTS.md — what a comment may state

The **cross-language** convention for comments and doc comments. It applies in
every Valkyrja port — PHP, Java, Go, Python, TypeScript, and Kotlin — and to
every contributor, human or agent.

Code speaks for itself. A comment adds what the code cannot show, so most code
carries no comment. A file where every block carries a comment is a wall of
text, and a reader skips walls. Comment the one line in ten that needs one.
That comment then stands alone. ([`AGENTS.md`](AGENTS.md) §3 summarizes this
document's rules as golden rules 10, 13, 14, and 15.)

PHP examples are shown, because PHP is the reference implementation; a rule
about config shows YAML. The rules hold in every port. The examples are generic
by design: they show the shape in the framework's naming style, and they copy
no real method, so this document does not drift when the source changes.

## What a comment states

A comment states what the code cannot show. Keep a comment to one or two
lines: the constraint, the invariant, or the reason the obvious approach
fails. Do not write a comment that narrates what the next line does — the code
shows it.

```php
// Wrong — the comment narrates what the next line does.
// Get the config from the container.
$config = $container->getSingleton(ConfigContract::class);
```

```php
// Right — the code shows what it does, so the line takes no comment.
$config = $container->getSingleton(ConfigContract::class);
```

When an explanation needs a paragraph, put the paragraph in the pull request
description or in a document. Shrink the comment to one sentence that states
the conclusion.

```php
// Wrong — the explanation grew into a paragraph, so the paragraph moved into
// the comment.

// The cache key includes the locale. Two locales can render one route
// differently, and a shared key would publish one locale's render to the
// other. Do not remove the locale from the key.
$key = $this->getCacheKey($route, $locale);
```

```php
// Right — one sentence states the conclusion. The paragraph lives in the pull
// request description.

// The key includes the locale, because two locales render one route differently.
$key = $this->getCacheKey($route, $locale);
```

A wall of comments hides the one warning that matters. Comment the one line
in ten that needs a comment:

```php
// Wrong — every line carries a true comment, so the one warning drowns.

// The boot registers the config before any handler runs.
$config = $container->getSingleton(ConfigContract::class);
// The path is absolute.
$path = $config->cachePath;
// The cache file is generated. A missing file means the build has not run.
$cache = CacheFactory::fromPath($path);
```

```php
// Right — the one comment that matters stands alone.

$config = $container->getSingleton(ConfigContract::class);
$path = $config->cachePath;

// The cache file is generated. A missing file means the build has not run.
$cache = CacheFactory::fromPath($path);
```

## A comment never states a transient condition

A comment in code or config must stay true indefinitely. Do not write a
comment to explain something temporary, or something automation will later
rewrite:

- a version pinned pending a release
- a workaround awaiting a fix
- a value automation regenerates

Automation rewrites values, not the prose around them, so the comment outlives
what it described and becomes an assertion that is now false. That is worse
than no comment, because the next reader trusts it.

Put the explanation in the pull request description instead
([`PR_DESCRIPTION.md`](PR_DESCRIPTION.md)). Nothing is lost. The squash merge
writes the description as the commit body, so the explanation lives in git
history permanently. The explanation stays attached to the commit that
introduced it, and `git log` and `git blame` reach it. This is also why the
commits you write carry no body — the squash commit's body comes from the
pull request.

The test is whether the comment states a _decision or invariant_ or a _current
condition_. A decision stays in the comment. A condition goes to the pull
request description.

```php
// Wrong — the comment states a condition, and the fix makes it false.

// Retry twice as a workaround until the redis client reconnects on its own.
$response = $this->retry($request, 2);
```

```php
// Right — the comment states a decision, and the decision stays true.

// Retry twice, because the redis client drops one connection on failover.
$response = $this->retry($request, 2);
```

The same rule governs config:

```yaml
# Wrong — the comment states a condition. Automation bumps the version and
# leaves the sentence, so the comment is then false.
sindri-version: "26.5.0" # Pinned ahead of the others until the next release.
```

```yaml
# Right — the comment states a decision, and the decision stays true.

# This job asserts only generated code, so its coverage report is meaningless.
coverage-report: false
```

## A comment inside a method body

A comment inside a body explains what is unclear, or why the code does
something this particular way. When an alternative failed, say so, so the next
editor does not retry it.

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
pull request description. The transient rule above says why nothing is lost.

## The doc comment on a type declaration

A class, a contract, a trait, an enum, and a struct carry no doc comment. The
declaration explains itself: the name and the segment say what the type is
([`STRUCTURE.md`](STRUCTURE.md) encodes the kind in both), and each method's doc
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
