# Valkyrja Dispatch

## Overview

The Dispatch component is currently central to how routes (CLI and HTTP) and listeners are dispatched from their
respective routers and the event dispatcher. It relies heavily on `::class` (PHP) and `.class` (Java) to dynamically
resolve and call methods on controllers, actions, and listeners via reflection.

This approach is being deprecated as the central dispatch mechanism and replaced with explicit closure-based handlers on
routes and listeners directly. The Dispatch component is retained as an optional power feature for PHP and Java only.

---

## Why Dispatch Cannot Be Central Across All Ports

The Dispatch component's dynamic resolution works by:

1. Receiving a class reference (`::class` / `.class`)
2. Resolving the class from the container
3. Dynamically determining which method to call
4. Calling that method via reflection or dynamic dispatch

Steps 2-4 are fundamentally incompatible with Go, TypeScript, and Python (in lambda/CGI contexts) because:

- **Go** — no reflection-based method dispatch, no `::class` equivalent
- **TypeScript** — types erased at runtime, no reliable class reference mechanism post-compile
- **Python** — technically possible via `getattr` but not reliable across all deployment contexts and fights the
  language's grain for a framework-level concern

A framework component that only works in two of five ports cannot be central to how the framework functions.

---

## The Typed Handler Signature

A major benefit of the closure-based handler approach over dispatch is that the required closure signature can be fully
type-hinted and enforced at the language level. Each handler type has its own specific signature:

| Handler type   | Parameters                                | Return type        |
|----------------|-------------------------------------------|--------------------|
| HTTP route     | `ContainerContract`, `map<string, mixed>` | `ResponseContract` |
| CLI route      | `ContainerContract`, `map<string, mixed>` | `OutputContract`   |
| Event listener | `ContainerContract`, `map<string, mixed>` | `any` / `mixed`    |

The second parameter is `map<string, mixed>` in all three cases — named arguments from the matched route, command, or
event. The return type differs per concern.

`ServerRequestContract` and `RouteContract` are **not** explicit parameters. They are always available via the container
when needed. Keeping them out of the signature:

- Makes the signature uniform and minimal across all handler types
- Avoids passing HTTP-specific objects to CLI handlers where they make no sense
- Lets the developer decide what to resolve — no unnecessary overhead for handlers that don't need them

```php
// HTTP handler — fetch request from container only if needed
static fn(ContainerContract $c, array<string, mixed> $args): ResponseContract => (
    $c->getSingleton(UserController::class)->show(
        $c->getSingleton(ServerRequestContract::class), // available if needed
        $args['id']
    )
)

// CLI handler — same signature shape, different concern
static fn(ContainerContract $c, array<string, mixed> $args): OutputContract => (
    $c->getSingleton(UserCommand::class)->run($args)
)

// Listener — same shape, returns any
static fn(ContainerContract $c, array<string, mixed> $args): mixed => (
    $c->getSingleton(UserCreatedListener::class)->handle($args['user_id'])
)
```

This moves validation from runtime (dispatch discovering the wrong method at request time) to compile time or static
analysis time — a wrong signature is caught before the application ever runs.

---

### Named Handler Types Per Language

Each concern gets its own named handler type. All five languages define three types — one per concern.

**PHP** — enforced by PHPStan/Psalm via docblock typing:

```php
// HTTP
/** Closure(ContainerContract, array<string, mixed>): ResponseContract */

// CLI
/** Closure(ContainerContract, array<string, mixed>): OutputContract */

// Event listener
/** Closure(ContainerContract, array<string, mixed>): mixed */
```

**Java** — three `@FunctionalInterface` types, compiler enforced:

```java
// HTTP
@FunctionalInterface
public interface HttpHandlerFunc {
    ResponseContract handle(ContainerContract container, Map<String, Object> arguments);
}

// CLI
@FunctionalInterface
public interface CliHandlerFunc {
    OutputContract handle(ContainerContract container, Map<String, Object> arguments);
}

// Event listener
@FunctionalInterface
public interface ListenerHandlerFunc {
    Object handle(ContainerContract container, Map<String, Object> arguments);
}
```

**Go** — three named function types, compiler enforced:

```go
// HTTP
type HttpHandlerFunc func (container ContainerContract, arguments map[string]any) ResponseContract

// CLI
type CliHandlerFunc func (container ContainerContract, arguments map[string]any) OutputContract

// Event listener
type ListenerHandlerFunc func (container ContainerContract, arguments map[string]any) any
```

**Python** — three `Callable` type aliases, enforced by mypy/pyright:

```python
from typing import Callable, Any

HttpHandlerFunc = Callable[[ContainerContract, dict[str, Any]], ResponseContract]
CliHandlerFunc = Callable[[ContainerContract, dict[str, Any]], OutputContract]
ListenerHandlerFunc = Callable[[ContainerContract, dict[str, Any]], Any]
```

**TypeScript** — three named types, compiler enforced:

```typescript
// HTTP
type HttpHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => ResponseContract

// CLI
type CliHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => OutputContract

// Event listener
type ListenerHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => unknown
```

---

## The Replacement: Handler and CacheableHandler Contracts

Each concern gets its own `HandlerContract` using its specific named handler type. The base `HandlerContract` defines
the method names. Each concern's contract tightens the type.

### Base HandlerContract

Defines method names only — no return type on the closure. Each concern's contract overrides with the specific typed
closure:

```php
// PHP — base
interface HandlerContract
{
    public function getHandler(): Closure;
    public function setHandler(Closure $handler): static;
}
```

```java
// Java — no base needed, each concern uses its own @FunctionalInterface directly
```

```go
// Go — base interface, each concern embeds and overrides
type HandlerContract interface {
GetHandler() any // tightened by each concern's contract
SetHandler(any) HandlerContract
}
```

```python
# Python — base ABC
class HandlerContract(ABC):
    @abstractmethod
    def get_handler(self) -> Callable: ...

    @abstractmethod
    def set_handler(self, handler: Callable) -> 'HandlerContract': ...
```

```typescript
// TypeScript — base interface
interface HandlerContract {
    getHandler(): (...args: any[]) => unknown

    setHandler(handler: (...args: any[]) => unknown): this
}
```

---

### HTTP Handler Contract

```php
// PHP
interface HttpHandlerContract extends HandlerContract
{
    /**
     * @return Closure(ContainerContract, array<string, mixed>): ResponseContract
     */
    public function getHandler(): Closure;

    /**
     * @param Closure(ContainerContract, array<string, mixed>): ResponseContract $handler
     */
    public function setHandler(Closure $handler): static;
}

// usage — PHPStan enforces signature
$route->setHandler(
    static fn(ContainerContract $c, array<string, mixed> $args): ResponseContract
        => $c->getSingleton(UserController::class)->show($args['id'])
);
```

```java
// Java
public interface HttpHandlerContract {
    HttpHandlerFunc getHandler();

    HttpHandlerContract setHandler(HttpHandlerFunc handler);
}

// usage — compiler enforces HttpHandlerFunc
route.

setHandler((container, arguments) ->
        container.

getSingleton(UserController .class).

show(arguments.get("id"))
        );
// wrong return type? compile error
```

```go
// Go
type HttpHandlerContract interface {
GetHandler() HttpHandlerFunc
SetHandler(HttpHandlerFunc) HttpHandlerContract
}

// usage — compiler enforces HttpHandlerFunc
route.SetHandler(func (c ContainerContract, args map[string]any) ResponseContract {
return c.GetSingleton(UserControllerClass).(*UserController).Show(args["id"])
})
```

```python
# Python
class HttpHandlerContract(HandlerContract, ABC):
    @abstractmethod
    def get_handler(self) -> HttpHandlerFunc: ...

    @abstractmethod
    def set_handler(self, handler: HttpHandlerFunc) -> 'HttpHandlerContract': ...


# usage
route.set_handler(
    lambda c, args: c.get_singleton(UserControllerClass).show(args['id'])
)
```

```typescript
// TypeScript
interface HttpHandlerContract extends HandlerContract {
    getHandler(): HttpHandlerFunc

    setHandler(handler: HttpHandlerFunc): this
}

// usage — tsc enforces HttpHandlerFunc
route.setHandler((container, args) =>
    container.getSingleton<UserController>(UserControllerClass).show(args['id'] as string)
)
```

---

### CLI Handler Contract

```php
// PHP
interface CliHandlerContract extends HandlerContract
{
    /**
     * @return Closure(ContainerContract, array<string, mixed>): OutputContract
     */
    public function getHandler(): Closure;

    /**
     * @param Closure(ContainerContract, array<string, mixed>): OutputContract $handler
     */
    public function setHandler(Closure $handler): static;
}

// usage
$command->setHandler(
    static fn(ContainerContract $c, array<string, mixed> $args): OutputContract
        => $c->getSingleton(SendEmailCommand::class)->run($args)
);
```

```java
// Java
public interface CliHandlerContract {
    CliHandlerFunc getHandler();

    CliHandlerContract setHandler(CliHandlerFunc handler);
}

// usage
command.

setHandler((container, arguments) ->
        container.

getSingleton(SendEmailCommand .class).

run(arguments)
);
```

```go
// Go
type CliHandlerContract interface {
GetHandler() CliHandlerFunc
SetHandler(CliHandlerFunc) CliHandlerContract
}

// usage
command.SetHandler(func (c ContainerContract, args map[string]any) OutputContract {
return c.GetSingleton(SendEmailCommandClass).(*SendEmailCommand).Run(args)
})
```

```python
# Python
class CliHandlerContract(HandlerContract, ABC):
    @abstractmethod
    def get_handler(self) -> CliHandlerFunc: ...

    @abstractmethod
    def set_handler(self, handler: CliHandlerFunc) -> 'CliHandlerContract': ...


# usage
command.set_handler(
    lambda c, args: c.get_singleton(SendEmailCommandClass).run(args)
)
```

```typescript
// TypeScript
interface CliHandlerContract extends HandlerContract {
    getHandler(): CliHandlerFunc

    setHandler(handler: CliHandlerFunc): this
}

// usage
command.setHandler((container, args) =>
    container.getSingleton<SendEmailCommand>(SendEmailCommandClass).run(args)
)
```

---

### Listener Handler Contract

```php
// PHP
interface ListenerHandlerContract extends HandlerContract
{
    /**
     * @return Closure(ContainerContract, array<string, mixed>): mixed
     */
    public function getHandler(): Closure;

    /**
     * @param Closure(ContainerContract, array<string, mixed>): mixed $handler
     */
    public function setHandler(Closure $handler): static;
}

// usage
$listener->setHandler(
    static fn(ContainerContract $c, array<string, mixed> $args): mixed
        => $c->getSingleton(UserCreatedListener::class)->handle($args['user_id'])
);
```

```java
// Java
public interface ListenerHandlerContract {
    ListenerHandlerFunc getHandler();

    ListenerHandlerContract setHandler(ListenerHandlerFunc handler);
}

// usage
listener.

setHandler((container, arguments) ->
        container.

getSingleton(UserCreatedListener .class).

handle(arguments.get("user_id"))
        );
```

```go
// Go
type ListenerHandlerContract interface {
GetHandler() ListenerHandlerFunc
SetHandler(ListenerHandlerFunc) ListenerHandlerContract
}

// usage
listener.SetHandler(func (c ContainerContract, args map[string]any) any {
return c.GetSingleton(UserCreatedListenerClass).(*UserCreatedListener).Handle(args["user_id"])
})
```

```python
# Python
class ListenerHandlerContract(HandlerContract, ABC):
    @abstractmethod
    def get_handler(self) -> ListenerHandlerFunc: ...

    @abstractmethod
    def set_handler(self, handler: ListenerHandlerFunc) -> 'ListenerHandlerContract': ...


# usage
listener.set_handler(
    lambda c, args: c.get_singleton(UserCreatedListenerClass).handle(args['user_id'])
)
```

```typescript
// TypeScript
interface ListenerHandlerContract extends HandlerContract {
    getHandler(): ListenerHandlerFunc

    setHandler(handler: ListenerHandlerFunc): this
}

// usage
listener.setHandler((container, args) =>
    container.getSingleton<UserCreatedListener>(UserCreatedListenerClass).handle(args['user_id'] as string)
)
```

---

### The Type Safety Advantage Over Dispatch

```php
// old dispatch — no type safety, validated at runtime only
$route->setHandler(UserController::class);
$route->setHandlerMethod('index');
// wrong method name?       RuntimeException at request time — in production
// wrong return type?       RuntimeException at request time — in production
// method doesn't exist?    RuntimeException at request time — in production

// new handler — method pointer on the route provider, type enforced before shipping
// handler lives on the same class as the route definition — forge reads one file
HttpRoute::get('/users/{id}', [self::class, 'showUser'])

public static function showUser(ContainerContract $c, array<string, mixed> $args): ResponseContract
{
    return $c->getSingleton(UserController::class)->show($args['id']);
}
// wrong return type?           PHPStan catches it at CI time
// missing parameter?           PHPStan catches it at CI time
// method doesn't exist?        PHPStan catches it at CI time
// controller not in container? ContainerException at bootstrap, not request time
```

In Java, Go, and TypeScript the enforcement is even stronger — these are compile errors, not static analysis warnings. A
wrong handler signature never reaches a running binary.

|                      | PHP (dispatch) | PHP (handler)      | Java               | Go                 | Python             | TypeScript         |
|----------------------|----------------|--------------------|--------------------|--------------------|--------------------|--------------------|
| Enforcement          | ❌ none         | ⚠️ PHPStan         | ✅ compiler         | ✅ compiler         | ⚠️ mypy            | ✅ compiler         |
| When caught          | Runtime        | CI                 | Compile            | Compile            | CI                 | Compile            |
| HTTP return type     | ❌              | `ResponseContract` | `ResponseContract` | `ResponseContract` | `ResponseContract` | `ResponseContract` |
| CLI return type      | ❌              | `OutputContract`   | `OutputContract`   | `OutputContract`   | `OutputContract`   | `OutputContract`   |
| Listener return type | ❌              | `mixed`            | `Object`           | `any`              | `Any`              | `unknown`          |

---

### CacheableHandler Contract

The `CacheableHandler` contract extends `Handler` with a string representation of the closure for use during cache data
file generation. The string is written by the same developer who writes the closure — it is never used at runtime, only
at cache generation time. The same typed `HandlerFunc` signature applies to the cacheable form:

```php
// PHP
interface CacheableHandlerContract extends HandlerContract
{
    public function getCacheableHandler(): string;
    public function setCacheableHandler(string $handler): static;
}
```

```java
// Java
public interface CacheableHandlerContract extends HandlerContract {
    String getCacheableHandler();

    CacheableHandlerContract setCacheableHandler(String handler);
}
```

```go
// Go
type CacheableHandlerContract interface {
HandlerContract
GetCacheableHandler() string
SetCacheableHandler(string) CacheableHandlerContract
}
```

```python
# Python
class CacheableHandlerContract(HandlerContract, ABC):
    @abstractmethod
    def get_cacheable_handler(self) -> str:
        pass

    @abstractmethod
    def set_cacheable_handler(self, handler: str) -> 'CacheableHandlerContract':
        pass
```

```typescript
// TypeScript
interface CacheableHandlerContract extends HandlerContract {
    getCacheableHandler(): string

    setCacheableHandler(handler: string): this
}
```

---

## The Annotation / Attribute Approach

For PHP, Java, and Python — where annotations/attributes/decorators are available — the developer annotates the action
method rather than manually constructing route objects with handlers:

**PHP**

```php
#[Handler(static fn(ContainerContract $c, array $args): Response
    => $c->getSingleton(UserController::class)->index($args[0]))]
public function index(Request $request): Response
{
    // actual implementation
}
```

**Java**

```java
@Handler((ContainerContract c, List < Object > args) ->
        c.

getSingleton(UserController .class).

index((Request) args.

get(0)))

public Response index(Request request) {
    // actual implementation
}
```

**Python**

```python
@handler(lambda c, args: c.get_singleton(UserController).index(args[0]))
def index(request: Request) -> Response:
    # actual implementation
    pass
```

For **Go** and **TypeScript** — where no annotations exist — explicit registration is used:

**Go**

```go
router.Get("/users",
valkyrja.Handler(func(c ContainerContract, args []any) any {
return c.GetSingleton(UserControllerClass).(*UserController).Index(args[0])
}),
)
```

**TypeScript**

```typescript
router.get('/users',
    handler((c: ContainerContract, args: any[]) =>
        c.getSingleton(UserController).index(args[0]))
)
```

---

## The `CacheableHandler` String — When It Is And Isn't Needed

The `CacheableHandler` string representation is only needed for CGI and lambda deployments where cache data files are
required. It is **never used at runtime** — the closure is always used at runtime.

For **PHP, Java, and Python** the build tool (valkyrja-build) extracts the handler closure source text automatically via
AST and generates the cache data files. The developer never writes a `CacheableHandler` string.

For **Go and TypeScript** the build tool reads the route provider source files via AST (go/analysis and TypeScript
compiler API respectively), extracts the handler closure source text, and generates cache data files. The developer also
never writes a `CacheableHandler` string.

The `CacheableHandler` contract exists as an escape hatch for edge cases where automatic extraction is not possible or
the developer wants explicit control over the cached form.

---

## Dispatch Component Retention

The Dispatch component is retained for PHP and Java as an opt-in power feature. It is not removed — it simply loses its
status as a required central dependency of the routing and event pipeline.

**What Dispatch retains:**

- Dynamic method resolution via `::class` / `.class`
- Reflection-based handler calling
- All existing PHP/Java dispatch behavior

**What changes:**

- Routes and listeners no longer require Dispatch to function
- The router and event dispatcher invoke the handler closure directly if present
- Dispatch is only invoked as a fallback if no closure handler is set (backwards compatibility during migration)
- New routes and listeners should use closure handlers

**Migration path:**

1. Introduce `Handler` and `CacheableHandler` contracts on routes and listeners
2. New routes use closure handlers — Dispatch not involved
3. Existing routes continue to work via Dispatch (backwards compatible)
4. Deprecation warnings added to Dispatch-based route definitions
5. Dispatch removed from core pipeline in a future major version
6. Dispatch component remains available as an optional package for PHP and Java

---

## Per-Language Summary

### PHP

Dispatch retained as opt-in. Closure handlers are the new canonical approach. The `#[Handler]` attribute drives both
runtime dispatch and cache generation (via build tool AST extraction).

```php
// old — dispatch-based (deprecated)
$route->setHandler(UserController::class);
$route->setHandlerMethod('index');

// new — closure-based
$route->setHandler(
    static fn(ContainerContract $c, array $args): Response
        => $c->getSingleton(UserController::class)->index($args[0])
);
```

### Java

Dispatch retained as opt-in. Annotation processor extracts `@Handler` lambda via Trees API at compile time, generates
cache data classes via JavaPoet. No developer-written `CacheableHandler` string needed.

```java
// old — dispatch-based (deprecated)
route.setHandler(UserController .class);
route.

setHandlerMethod("index");

// new — closure-based
route.

setHandler(
    (ContainerContract c, List<Object> args) ->
        c.

getSingleton(UserController .class).

index((Request) args.

get(0))
        );
```

### Go

Dispatch never existed in Go — not applicable. Explicit closure registration is the only mechanism. Build tool uses
go/analysis to extract handler closures from route provider source files.

```go
// go — always explicit
router.Get("/users",
valkyrja.Handler(func(c ContainerContract, args []any) any {
return c.GetSingleton(UserControllerClass).(*UserController).Index(args[0])
}),
)
```

### Python

Decorators self-register at import time. Build tool uses `ast` module + `inspect.getfile()` to extract handler closures
for cache generation. Dispatch not applicable.

```python
# python — decorator-based registration
@handler(lambda c, args: c.get_singleton(UserController).index(args[0]))
def index(request: Request) -> Response:
    pass
```

### TypeScript

No decorators. Explicit registration only. Build tool uses TypeScript compiler API to extract handler closures from
route provider source files.

```typescript
// typescript — explicit registration
router.get('/users',
    handler((c: ContainerContract, args: any[]) =>
        c.getSingleton(UserController).index(args[0]))
)
```

---

## Annotated Controllers — PHP, Java, Python

For annotated controllers, annotations live on the **implementation method**. The forge tool reads the annotations and
constructs a route object — exactly the same shape as a route returned from `getRoutes()`. **No method body extraction.
No import resolution of the callable.** The callable from `#[Handler]` is written directly into the generated cache data
class as a literal, just as it appears in the source.

This is identical to how service bindings work:

```php
// service binding — callable written as a literal
SomeServiceId::class => [SomeServiceProvider::class, 'publishSomeClass']

// route — callable written as a literal  
new Route('/users/{id}', 'user.show', [SomeClass::class, 'theHandlerMethod'])
```

The forge tool reads literals, writes literals. No execution, no body extraction, no cross-file resolution.

Go and TypeScript have no annotation support — routes are always registered explicitly via `getRoutes()`.

---

### Annotation Structure

```
#[Route]      — HTTP method + path — lives on the implementation method
#[Parameter]  — dynamic segment constraints — lives on the implementation method
#[Handler]    — callable reference — lives on the implementation method
```

The callable in `#[Handler]` is the value written into the generated route object unchanged.

---

### PHP

```php
class UserController
{
    // Forge reads these annotations and constructs a Route object.
    // The callable [SomeClass::class, 'theHandlerMethod'] is written
    // directly into the generated cache as-is — no body extraction.
    #[Route('GET', '/users/{id}')]
    #[Parameter('id', pattern: '[0-9]+')]
    #[Handler([self::class, 'showHandler'])]
    public function show(string $id): ResponseContract
    {
        // actual implementation — irrelevant to forge
    }

    // The handler — may be on this class or any other class
    public static function showHandler(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(self::class)->show($args['id']);
    }
}
```

Generated output — identical shape to an explicit `getRoutes()` route:

```php
new \Valkyrja\Http\Routing\Data\HttpRoute(
    path:       '/users/{id}',
    name:       'user.show',
    method:     'GET',
    parameters: [new \Valkyrja\Http\Routing\Data\Parameter('id', '[0-9]+')],
    handler:    [self::class, 'showHandler'],  // written as-is from the annotation
)
```

**Forge reads — PHP:**

```
1. Find #[Route], #[Parameter], #[Handler] on the implementation method
2. Extract path, HTTP method, parameter name/pattern, callable — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### Java

```java
public class UserController {

    // Forge reads annotations and constructs a Route object.
    // Callable written directly into generated cache — no body extraction.
    @Route(method = "GET", path = "/users/{id}")
    @Parameter(name = "id", pattern = "[0-9]+")
    @Handler(clazz = UserController.class, method = "showHandler")
    public ResponseContract show(String id) {
        // actual implementation — irrelevant to forge
    }

    public static ResponseContract showHandler(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(UserController.class).show((String) args.get("id"));
    }
}
```

Generated output:

```java
new HttpRoute(
    "/users/{id}","user.show","GET",
    List.of(new Parameter("id", "[0-9]+")),
        new

HandlerRef(UserController .class, "showHandler")  // written as-is
)
```

**Forge reads — Java:**

```
1. Find @Route, @Parameter, @Handler on the implementation method
2. Extract path, HTTP method, parameter name/pattern, clazz + method — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### Python

```python
class UserController:

    # Forge reads these decorators and constructs a Route object.
    # The callable tuple is written directly into the generated cache — no body extraction.
    @route('GET', '/users/{id}')
    @parameter('id', pattern='[0-9]+')
    @handler((UserController, 'show_handler'))  # callable tuple — written as-is
    def show(self, id: str) -> ResponseContract:
        pass  # actual implementation — irrelevant to forge

    @staticmethod
    def show_handler(c: ContainerContract, args: dict) -> ResponseContract:
        return c.get_singleton(UserController).show(args['id'])
```

Generated output:

```python
HttpRoute(
    path='/users/{id}',
    name='user.show',
    method='GET',
    parameters=[Parameter('id', '[0-9]+')],
    handler=(UserController, 'show_handler'),  # written as-is from decorator
)
```

**Forge reads — Python:**

```
1. Find @route, @parameter, @handler decorators on the implementation method
2. Extract path, HTTP method, parameter name/pattern, callable tuple — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### The Forge Pattern (All Languages)

```
Annotations / decorators carry literals.
Forge reads literals.
Forge writes literals into the generated cache data class.
No method body extraction. No import resolution of the callable itself.

Same as service bindings:
  SomeServiceId::class => [SomeProvider::class, 'publishMethod']  ← literal, written as-is

Same as explicit routes:
  new Route('/path', 'name', [SomeClass::class, 'theHandlerMethod'])  ← literal, written as-is
```

## Design Note — Why Routes and Listeners Cannot Use a Publisher-Style Map

When designing the handler pattern for service providers, a natural question arose: could routes and listeners be
expressed the same way as container bindings — a map of identifier to handler method, with the build tool reading method
bodies directly from AST, eliminating the need for `getRoutes()` / `getListeners()` lists entirely?

This was considered and rejected for a fundamental architectural reason.

**Container bindings are simple key→factory pairs.** The binding key IS the complete identity of the binding. The
factory closure is the only additional data needed. The build tool can generate a complete, self-contained cache entry
from just the method body.

**Routes are multi-dimensional data structures.** A route carries:

- HTTP method (GET, POST, PUT, DELETE, PATCH)
- Path pattern (`/users/{id}`)
- Dynamic segment definitions and constraints (`{id}` → `[0-9]+`)
- Regex compilation from the path pattern
- Middleware chain
- Name / alias
- Parameter defaults
- Host constraints
- Scheme constraints

All of this metadata, in addition to the handler, makes up a route. The `HttpRoute::get('/users/{id}', handler)` call is
what populates all of these fields together as a complete data object. Decomposing this into a key/method-body map would
lose all the metadata the router needs to build its dispatcher trie, compile route regexes, and resolve middleware
chains. The router cannot function with just a path string and a handler — it needs the full route object.

**Listeners have the same problem.** A listener carries event type binding, priority, and stop-propagation behavior
alongside the handler. These cannot be expressed as a flat key/body map without losing the data the event dispatcher
requires.

This is why `getRoutes()` and `getListeners()` return complete object lists while `publishers()` uses a map — the
difference reflects a genuine architectural distinction, not an inconsistency.

---

## Discussion Summary

The Dispatch component's architecture was examined when planning the Go and Python ports. The component works by
receiving a class reference and using reflection or dynamic method calls to invoke the appropriate handler — a pattern
that works elegantly in PHP and Java but has no equivalent in Go (no reflection-based method dispatch), TypeScript (
types erased at runtime), or Python in a reliable cross-deployment form.

The first realization was that Dispatch conflates two concerns: knowing what to call (the class and method reference)
and actually calling it (the dynamic invocation). Closure-based handlers collapse these into one: the closure IS the
invocation, explicitly written by the developer.

The second realization was that closure-based dispatch is strictly better architecture even in PHP and Java. Closures
are faster (no reflection overhead), more transparent (you can read exactly what will be called), and naturally
testable (trivially replaceable in tests). The dynamic dispatch was a convenience that came at a real cost.

The Handler and CacheableHandler contracts were designed to be the cross-language solution. Handler carries the
executable closure used at runtime. CacheableHandler carries the string representation used only at cache generation
time — never at runtime. The developer writes both for CGI/lambda deployments, but the build tool's AST extraction
capabilities mean this double-write burden is eliminated for PHP, Java, Python, Go, and TypeScript in practice.

The annotation/attribute approach for PHP, Java, and Python allows the framework to provide a clean developer
experience — the developer annotates the action method and the framework handles the rest. Go and TypeScript use
explicit registration which is honest to their philosophy of explicit-over-implicit.

The decision to retain Dispatch as an optional component rather than removing it entirely was driven by backwards
compatibility and the genuine usefulness of dynamic dispatch for PHP and Java developers who want it. Removing it from
the core pipeline while keeping it available as an opt-in respects existing users while establishing the correct
architecture for all ports going forward.

A further benefit of the closure-based handler approach — identified after the initial design — is typed closure
signatures. The dispatch approach had no type enforcement on what method was called or what it returned. Errors were
discovered at request time in production. With explicit closures, each language can enforce the handler signature at the
level the language supports — compile time for Java, Go, and TypeScript; static analysis time for PHP and Python. This
moves an entire class of runtime errors to before the application ships.

Each handler concern has its own specific return type: HTTP handlers return `ResponseContract`, CLI handlers return
`OutputContract`, and event listeners return `any` / `mixed`. The second parameter — `map<string, mixed>` of named
arguments — is consistent across all three. `ServerRequestContract` and `RouteContract` are intentionally absent from
the handler signature. They are always available via the container when needed, keeping the signature minimal and
avoiding coupling HTTP-specific objects to CLI and listener handlers.
