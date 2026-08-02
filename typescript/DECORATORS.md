# Routing decorators — design rationale (TypeScript, and what it means for Python)

Why Valkyrja's TypeScript routing decorators look the way they do, which
alternatives were tried and rejected, and which parts of this carry over to the
future Python port.

This is the **design record**. For the implementation task list see
[`DECORATORS_IMPLEMENTATION.md`](DECORATORS_IMPLEMENTATION.md).

---

## 1. The core insight

The difficulty with decorators in TypeScript is **not** that decorators are a
weaker feature than PHP attributes or Java annotations. It is that the four
languages _encode a class reference differently_:

| Language       | Reference inside the attribute / decorator                                                 | Evaluated eagerly? | Safe in a cycle |
| -------------- | ------------------------------------------------------------------------------------------ | ------------------ | --------------- |
| **PHP**        | `HttpRouteProvider::class` — a **compile-time string literal**                             | No                 | ✅              |
| **Java**       | `HttpRouteProvider.class` — a **constant-pool entry**, resolved lazily by the classloader  | No                 | ✅              |
| **TypeScript** | `[HttpRouteProvider, 'method']` — a **live binding dereferenced at class-definition time** | **Yes**            | ❌              |
| **Python**     | `[HomeController, 'method']` — a **live attribute read at import time**                    | **Yes**            | ❌              |

PHP and Java store _inert data_. TypeScript and Python evaluate a _real value_.
Everything below follows from that single difference.

`#[Route]` in PHP is inert until `ReflectionAttribute::newInstance()` is called;
a Java annotation is inert until reflected. A TypeScript decorator is an
**ordinary function call** that the language runs while defining the class.

### The precise rule

> **Never name a class in a decorator argument.**

It is not about "same class vs. different class", and not about _where_ handlers
live. It is only about whether the argument expression dereferences a class
binding that has not finished initializing.

---

## 2. Evidence (all verified, Node 24.15.0)

| Observation                                                                                 | Result                                                                                                                     |
| ------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Decorator syntax under `node --experimental-strip-types` / `--experimental-transform-types` | ❌ `SyntaxError` — Node strips types but does not lower decorators                                                         |
| Decorator syntax under the repos' Vitest 4 (Rolldown/oxc)                                   | ❌ `SyntaxError`                                                                                                           |
| Decorator syntax under `tsx` / esbuild                                                      | ✅ executes                                                                                                                |
| Stage-3 `context.metadata` → `Class[Symbol.metadata]`                                       | ✅ round-trips (this is the NestJS mechanism)                                                                              |
| `tsc` output for a decorated class                                                          | **Lowered to `__esDecorate(...)` calls — NOT stripped.** Production runs them too                                          |
| Decorator _argument_ evaluation timing                                                      | At class definition, **before** the class exists and without any instantiation                                             |
| `@RouteHandler([HomeController, 'x'])` _inside_ `HomeController`                            | ❌ `ReferenceError: Cannot access '_HomeController' before initialization`                                                 |
| Cross-module cycle (controller names provider, provider imports controller)                 | ❌ TDZ **if the provider module evaluates first**; works if the controller does — i.e. order-dependent, therefore unusable |
| Wrapping the class in a closure _inside_ `getControllerClasses()`                           | ❌ No effect — a method body already defers; the eager dereference is in the decorator                                     |
| `[() => Class, 'method']` (thunk)                                                           | ✅ Works in **every** case above, including self-reference                                                                 |
| Distinguishing a thunk from a class at runtime                                              | ✅ Arrow functions have `prototype === undefined`; classes have an object prototype                                        |

### Production is not exempt

`tsc` lowers decorators into `__esDecorate` calls that still run at class
definition, so a compiled production build carries the identical hazard. What it
does _not_ do is duplicate work: decorators only write to a per-class
`Symbol.metadata` object, and the only reader — `AttributeRouteCollector` — runs
solely on the debug path. In cached mode the generated data stays authoritative
and the metadata is simply never read.

That residual cost (a few object writes per decorated class at import) is
**unavoidable** for decorators in any runtime-evaluated language. Eliminating it
would require Angular-style AOT stripping via a custom build transform, which we
rejected: it is non-standard and would make development and production behave
differently — the exact class of problem this effort exists to remove.

---

## 3. Options considered

### ✅ Chosen — thunk the class reference

```ts
@RouteHandler([() => HttpRouteProvider, 'versionHandler'])
```

Creating a closure **captures** a binding without dereferencing it; TDZ fires
only on access. `() => HomeController` is JavaScript's native inert class
reference — the true analogue of PHP's `::class`.

- Works for cross-class references **and** self-reference, in any module order.
- Revives CLI `helpText`, which had to be dropped as `[TestCommand, 'help']`
  (a class cannot name itself while defining) and works as
  `[() => TestCommand, 'help']`.
- Handlers stay **on the route provider**; routes stay **on the controller
  action**; controller signatures stay clean; **no extra class**.
- Sindri unwraps the arrow body to the identifier, so generated output is
  byte-identical to the non-thunk form.
- **Thunk-only** — the bare `[Class, 'method']` form is deliberately _not_
  accepted. It appears to work whenever module order happens to favour it, then
  fails with an opaque TDZ error when an unrelated import reorders the graph.
  One shape that always works beats two shapes where one is a trap.

### ❌ Rejected

| Option                                                                 | Why rejected                                                                                                                                                                                                                    |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Handlers moved into the controller**                                 | Impossible — the decorator would name the class it lives in (self-TDZ). Also loses the separation the handlers were extracted for.                                                                                              |
| **Plain string** (`['App\Http\Provider\HttpRouteProvider', 'method']`) | Semantically closest to PHP, but a handler is a _callable_; a string forces a name→class registry plus a call convention. Large machinery for no gain over the thunk.                                                           |
| **Closure around the whole tuple** (`() => [Class, 'method']`)         | A closure of a closure. The handler is already a callable, so this reads as double-wrapping, and any closure taking the handler's parameters would duplicate the signature.                                                     |
| **Separate `RouteHandlerProvider` class**                              | Works, but adds a third class per component and makes the route provider inconsistent with service/event providers unless they all gain one. Rejected as structural noise.                                                      |
| **`@Route` on the provider's handler methods**                         | Works and removes the reference entirely — but forfeits class-wide annotations (`@Path`, `@Name`) and the principle of _annotating the thing actually being used_. Routes belong next to the controller action, as in PHP/Java. |
| **Controller method _is_ the handler**                                 | Works, but forces every controller method to take `(container, route)`, defeating the reason the handler middleman exists — to keep controller methods expressed in terms of what they actually need.                           |
| **AOT-strip decorators in production**                                 | Non-standard, and creates dev/prod divergence.                                                                                                                                                                                  |

---

## 4. Handler typing — an orthogonal fix

Do not conflate these. **The thunk solves the runtime TDZ. It does nothing for
types.** The type safety comes from a mapped type that is equally valid with or
without a thunk:

```ts
type HttpHandlerKeys<T> = { [K in keyof T]: T[K] extends HttpHandler ? K : never }[keyof T];
export type HttpHandlerReference<T> = [() => T, HttpHandlerKeys<T> & string];
```

`T` is inferred from the thunk's return type; `HttpHandlerKeys<T>` narrows the
second element to only those keys whose value matches the handler signature. A
misspelled method **and** a correctly-named method with the wrong signature both
become compile errors — neither was caught by the previous
`[new (...args: unknown[]) => unknown, string]`, which erased statics and
accepted any `string`.

> Keep a comment at both sites recording which fix does what. A future reader who
> assumes the thunk is load-bearing for _typing_ may "simplify" it away and
> silently reintroduce the TDZ.

---

## 5. What this means for Python

Python **does** have decorators (`@decorator`, since 2.4) and they behave like
TypeScript's for the purposes above: the decorator expression and its arguments
are evaluated **at class-body execution / import time**, not lazily. Python also
has the same partial-initialization hazard for circular imports, surfacing as
`ImportError: cannot import name … (most likely due to a circular import)` or an
`AttributeError` on a half-initialized module — the same failure, different
message.

So the Python port should inherit these conclusions directly:

- **Thunk the class reference.** `lambda: HomeController` is the exact analogue of
  `() => HomeController`.
- **Detect the thunk at runtime.** Prefer an explicit convention (always a thunk);
  if both forms must be supported, `inspect.isclass(ref[0])` distinguishes them
  cleanly — the counterpart of the `prototype === undefined` check.
- **Expect decorators to run in production too.** Python has no AOT-stripping
  step either; the cached data remains authoritative and the decorator metadata
  simply goes unread.
- **Static reading is unaffected.** The build tool parses the decorator with `ast`
  and unwraps the lambda body to the class name, exactly as Sindri unwraps the
  arrow body.

PHP and Java need none of this: their references are inert by construction, so
they keep the plain `::class` / `.class` form.

---

## 6. Practical rules

1. Never name a class directly in a decorator argument — always
   `[() => Class, 'method']`.
2. Keep handlers on the route provider; keep `@Route` on the controller action.
3. `getControllerClasses()` returns plain class references — it is a method body,
   so it is already deferred and needs no thunk.
4. Decorated source cannot run under `node --experimental-strip-types`; a
   decorator-capable loader (`tsx`, or `node --import` with an esbuild/swc hook)
   is required in development, and a compiled build in production.
5. Test runners that cannot lower decorators (Vitest 4 / Rolldown-oxc) need an
   esbuild transform plugin for files containing decorators; otherwise test the
   decorators by invoking them directly against a synthetic context.

---

## 7. Related finding

The decorator path was the first consumer of `Processor`-computed route regexes
(imperative routes hand-supply their own and skip the computation). It exposed a
latent PHP-ism: `Regex.START`/`END` were `'/^'` and `'$/'`, i.e. PHP-style
delimiters, which JavaScript's `RegExp` constructor treats as **literal
characters** — so a computed dynamic-route regex could never match. Corrected to
bare anchors `'^'` / `'$'`. Ports carrying regex constants over from PHP should
check the same thing.
