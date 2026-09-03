# Valkyrja Handlers

## Overview

A route (CLI and HTTP) and an event listener carry an explicit closure handler. The router calls the route's handler,
and the event dispatcher calls the listener's handler.

Each handler has a typed signature, and the language enforces that signature. This document records the handler design
for every port.

---

## The Typed Handler Signature

The language enforces the closure signature. Each handler type has its own signature:

| Handler type   | Parameters                                | Return type        |
| -------------- | ----------------------------------------- | ------------------ |
| HTTP route     | `ContainerContract`, `RouteContract`      | `ResponseContract` |
| CLI route      | `ContainerContract`, `RouteContract`      | `OutputContract`   |
| Event listener | `ContainerContract`, `map<string, mixed>` | `any` / `mixed`    |

A route handler takes the matched route, so the handler reads the route's own data. Each concern declares its own
route contract in its own namespace, so the HTTP `RouteContract` and the CLI `RouteContract` are two distinct types. A
listener takes a `map<string, mixed>`, because an event carries named arguments and no route.

`ServerRequestContract` is **not** an explicit parameter. The container holds the request, and a handler that needs the
request resolves the request. Keeping the request out of the signature:

- Avoids passing an HTTP-specific object to a CLI handler, where the object makes no sense
- Lets the developer decide what to resolve, so a handler pays for nothing it does not use

```php
// The handler reads the route's own data, and it resolves the request when it needs the request.
static fn(ContainerContract $c, RouteContract $route): ResponseContract => (
    $c->getSingleton(AuditController::class)->record(
        $route->getName(),
        $c->getSingleton(ServerRequestContract::class),
    )
)
```

```php
// HTTP handler — fetch the request from the container only if needed
static fn(ContainerContract $c, RouteContract $route): ResponseContract => (
    $c->getSingleton(UserController::class)->show($route)
)

// CLI handler — same signature shape, different concern
static fn(ContainerContract $c, RouteContract $route): OutputContract => (
    $c->getSingleton(UserCommand::class)->run($route)
)

// Listener — takes the event's named arguments, returns any
static fn(ContainerContract $c, array<string, mixed> $args): mixed => (
    $c->getSingleton(UserCreatedListener::class)->handle($args['user_id'])
)
```

The compiler or the static analyzer checks the signature. A wrong signature never reaches a running application.

---

### Named Handler Types Per Language

Each concern gets its own named handler type. All five languages define three types — one per concern.

**PHP** — enforced by PHPStan/Psalm via docblock typing:

```php
// HTTP
/** Closure(ContainerContract, RouteContract): ResponseContract */

// CLI
/** Closure(ContainerContract, RouteContract): OutputContract */

// Event listener
/** Closure(ContainerContract, array<string, mixed>): mixed */
```

**Java** — three `@FunctionalInterface` types, compiler enforced:

```java
// HTTP
@FunctionalInterface
public interface HttpHandlerFunc {
    ResponseContract handle(ContainerContract container, RouteContract route);
}

// CLI
@FunctionalInterface
public interface CliHandlerFunc {
    OutputContract handle(ContainerContract container, RouteContract route);
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
type HttpHandlerFunc func (container ContainerContract, route RouteContract) ResponseContract

// CLI
type CliHandlerFunc func (container ContainerContract, route RouteContract) OutputContract

// Event listener
type ListenerHandlerFunc func (container ContainerContract, arguments map[string]any) any
```

**Python** — three `Callable` type aliases, enforced by mypy/pyright:

```python
from typing import Callable, Any

HttpHandlerFunc = Callable[[ContainerContract, RouteContract], ResponseContract]
CliHandlerFunc = Callable[[ContainerContract, RouteContract], OutputContract]
ListenerHandlerFunc = Callable[[ContainerContract, dict[str, Any]], Any]
```

**TypeScript** — three named types, compiler enforced:

```typescript
// HTTP
type HttpHandlerFunc = (
    container: ContainerContract,
    route: RouteContract
) => ResponseContract

// CLI
type CliHandlerFunc = (
    container: ContainerContract,
    route: RouteContract
) => OutputContract

// Event listener
type ListenerHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => unknown
```

---

## Handler and CacheableHandler Contracts

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
     * @return Closure(ContainerContract, RouteContract): ResponseContract
     */
    public function getHandler(): Closure;

    /**
     * @param Closure(ContainerContract, RouteContract): ResponseContract $handler
     */
    public function setHandler(Closure $handler): static;
}

// usage — PHPStan enforces signature
$httpRoute->setHandler(
    static fn(ContainerContract $c, RouteContract $route): ResponseContract
        => $c->getSingleton(UserController::class)->show($route)
);
```

```java
// Java
public interface HttpHandlerContract {
    HttpHandlerFunc getHandler();

    HttpHandlerContract setHandler(HttpHandlerFunc handler);
}

// usage — compiler enforces HttpHandlerFunc
httpRoute.

setHandler((container, route) ->
        container.

getSingleton(UserController .class).

show(route)
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
httpRoute.SetHandler(func (c ContainerContract, route RouteContract) ResponseContract {
return c.GetSingleton(UserControllerClass).(*UserController).Show(route)
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
http_route.set_handler(
    lambda c, route: c.get_singleton(UserControllerClass).show(route)
)
```

```typescript
// TypeScript
interface HttpHandlerContract extends HandlerContract {
    getHandler(): HttpHandlerFunc

    setHandler(handler: HttpHandlerFunc): this
}

// usage — tsc enforces HttpHandlerFunc
httpRoute.setHandler((container, route) =>
    container.getSingleton<UserController>(UserControllerClass).show(route)
)
```

---

### CLI Handler Contract

```php
// PHP
interface CliHandlerContract extends HandlerContract
{
    /**
     * @return Closure(ContainerContract, RouteContract): OutputContract
     */
    public function getHandler(): Closure;

    /**
     * @param Closure(ContainerContract, RouteContract): OutputContract $handler
     */
    public function setHandler(Closure $handler): static;
}

// usage
$cliCommand->setHandler(
    static fn(ContainerContract $c, RouteContract $route): OutputContract
        => $c->getSingleton(SendEmailCommand::class)->run($route)
);
```

```java
// Java
public interface CliHandlerContract {
    CliHandlerFunc getHandler();

    CliHandlerContract setHandler(CliHandlerFunc handler);
}

// usage
cliCommand.

setHandler((container, route) ->
        container.

getSingleton(SendEmailCommand .class).

run(route)
);
```

```go
// Go
type CliHandlerContract interface {
GetHandler() CliHandlerFunc
SetHandler(CliHandlerFunc) CliHandlerContract
}

// usage
cliCommand.SetHandler(func (c ContainerContract, route RouteContract) OutputContract {
return c.GetSingleton(SendEmailCommandClass).(*SendEmailCommand).Run(route)
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
cli_command.set_handler(
    lambda c, route: c.get_singleton(SendEmailCommandClass).run(route)
)
```

```typescript
// TypeScript
interface CliHandlerContract extends HandlerContract {
    getHandler(): CliHandlerFunc

    setHandler(handler: CliHandlerFunc): this
}

// usage
cliCommand.setHandler((container, route) =>
    container.getSingleton<SendEmailCommand>(SendEmailCommandClass).run(route)
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

### Where the Language Catches a Wrong Handler

The handler is a method pointer on the route provider. The handler and the route definition are on the same class, so
`sindri` reads one file.

```php
HttpRoute::get('/users/{id}', [self::class, 'showUser'])

public static function showUser(ContainerContract $c, RouteContract $route): ResponseContract
{
    return $c->getSingleton(UserController::class)->show($route);
}
// wrong return type?           PHPStan catches it at CI time
// missing parameter?           PHPStan catches it at CI time
// method doesn't exist?        PHPStan catches it at CI time
// controller not in container? ContainerException at bootstrap, not request time
```

Java, Go, and TypeScript enforce the signature more strictly than PHP. A wrong signature is a compile error, not a
static analysis warning, so it never reaches a running binary.

|                      | PHP                | Java               | Go                 | Python             | TypeScript         |
| -------------------- | ------------------ | ------------------ | ------------------ | ------------------ | ------------------ |
| Enforcement          | ⚠️ PHPStan         | ✅ compiler        | ✅ compiler        | ⚠️ mypy            | ✅ compiler        |
| When caught          | CI                 | Compile            | Compile            | CI                 | Compile            |
| HTTP return type     | `ResponseContract` | `ResponseContract` | `ResponseContract` | `ResponseContract` | `ResponseContract` |
| CLI return type      | `OutputContract`   | `OutputContract`   | `OutputContract`   | `OutputContract`   | `OutputContract`   |
| Listener return type | `mixed`            | `Object`           | `any`              | `Any`              | `unknown`          |

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
#[Handler(static fn(ContainerContract $c, RouteContract $route): ResponseContract
    => $c->getSingleton(UserController::class)->index($route))]
public function index(RouteContract $route): ResponseContract
{
    // actual implementation
}
```

**Java**

```java
@RouteHandler((ContainerContract c, RouteContract route) ->
        c.

getSingleton(UserController .class).

index(route))

public ResponseContract index(RouteContract route) {
    // actual implementation
}
```

**Python**

```python
@route_handler(lambda c, route: c.get_singleton(UserController).index(route))
def index(route: RouteContract) -> ResponseContract:
    # actual implementation
    pass
```

For **Go** and **TypeScript** — where no annotations exist — explicit registration is used:

**Go**

```go
router.Get("/users",
valkyrja.Handler(func(c ContainerContract, route RouteContract) any {
return c.GetSingleton(UserControllerClass).(*UserController).Index(route)
}),
)
```

**TypeScript**

```typescript
router.get('/users',
    handler((c: ContainerContract, route: RouteContract) =>
        c.getSingleton(UserController).index(route))
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

## Internal Route and Listener Collection — Supplier Pattern

The router and event dispatcher store routes and listeners internally as a map of name/key to a callable that returns
the route or listener object. This applies to both runtime registration and cache loading.

### Why a Callable Rather Than the Object Directly

A route or listener loaded from cache may never be needed for a given request. Storing a callable defers construction
until the route is actually matched — and once resolved, the callable is replaced in the map with one that returns the
already-constructed instance. The map self-optimizes on first access.

This is the same pattern as the container's `Supplier` approach for service bindings — uniform across all concerns.

### The Pattern Across All Languages

```
Map<name, callable>
  where callable returns Route or Listener on invocation

First access:
  callable invoked → Route constructed
  map entry replaced with trivial callable returning the cached instance

Subsequent access:
  trivial callable invoked → cached instance returned immediately
```

**PHP**

```php
// internal map — Closure returning RouteContract
private array $routes = []; // string → Closure(): RouteContract

// registration
$this->routes[$name] = fn() => $route;

// first resolution — replace with cached instance
$resolved = ($this->routes[$name])();
$this->routes[$name] = fn() => $resolved;

// subsequent resolution — trivial closure
$route = ($this->routes[$name])();
```

**Java**

```java
// internal map — Supplier<RouteContract>
private Map<String, Supplier<RouteContract>> routes = new HashMap<>();

// registration
routes.

put(name, () ->route);

// first resolution — replace with cached instance
RouteContract resolved = routes.get(name).get();
routes.

put(name, () ->resolved);

// subsequent resolution — trivial supplier
RouteContract route = routes.get(name).get();
```

**Go**

```go
// internal map — func() RouteContract
var routes map[string]func () RouteContract

// registration
routes[name] = func () RouteContract { return route }

// first resolution — replace with cached instance
resolved := routes[name]()
routes[name] = func () RouteContract { return resolved }

// subsequent resolution — trivial func
route := routes[name]()
```

**Python**

```python
# internal map — lambda returning RouteContract
_routes: dict[str, Callable[[], RouteContract]] = {}

# registration
_routes[name] = lambda: route

# first resolution — replace with cached instance
resolved = _routes[name]()
_routes[name] = lambda: resolved

# subsequent resolution — trivial lambda
route = _routes[name]()
```

**TypeScript**

```typescript
// internal map — () => RouteContract
private
routes: Map<string, () => RouteContract> = new Map()

// registration
this.routes.set(name, () => route)

// first resolution — replace with cached instance
const resolved = this.routes.get(name)!()
this.routes.set(name, () => resolved)

// subsequent resolution — trivial function
const route = this.routes.get(name)!()
```

### Uniform Resolution

Resolution is always the same call regardless of whether the entry is a lazy supplier or a trivial cached-instance
supplier:

| Language   | Resolution call          |
| ---------- | ------------------------ |
| PHP        | `($routes[$name])()`     |
| Java       | `routes.get(name).get()` |
| Go         | `routes[name]()`         |
| Python     | `routes[name]()`         |
| TypeScript | `routes.get(name)!()`    |

No branching, no `instanceof` check, no `if lazy else cached`. The map self-manages. This applies equally to the
listener collection — same pattern, `ListenerContract` instead of `RouteContract`.

---

## Per-Language Summary

### PHP

The closure handler is the only mechanism. The `#[RouteHandler]` attribute drives the runtime call and the cache
generation. The build tool extracts the closure through AST.

```php
$httpRoute->setHandler(
    static fn(ContainerContract $c, RouteContract $route): ResponseContract
        => $c->getSingleton(UserController::class)->index($route)
);
```

### Java

The closure handler is the only mechanism. The annotation processor extracts the `@RouteHandler` lambda through the
Trees API at compile time, then generates the cache data classes through JavaPoet. The developer writes no
`CacheableHandler`
string.

```java
httpRoute.setHandler(
    (ContainerContract c, RouteContract route) ->
        c.getSingleton(UserController.class).index(route)
);
```

### Go

Explicit closure registration is the only mechanism. The build tool uses go/analysis to extract the handler closure from
the route provider source files.

```go
// go — always explicit
router.Get("/users",
valkyrja.Handler(func(c ContainerContract, route RouteContract) any {
return c.GetSingleton(UserControllerClass).(*UserController).Index(route)
}),
)
```

### Python

Decorators self-register at import time. The build tool uses the `ast` module and `inspect.getfile()` to extract the
handler closure for cache generation.

```python
# python — decorator-based registration
@route_handler(lambda c, route: c.get_singleton(UserController).index(route))
def index(route: RouteContract) -> ResponseContract:
    pass
```

### TypeScript

No decorators. Explicit registration only. Build tool uses TypeScript compiler API to extract handler closures from
route provider source files.

```typescript
// typescript — explicit registration
router.get('/users',
    handler((c: ContainerContract, route: RouteContract) =>
        c.getSingleton(UserController).index(route))
)
```

---

## Annotated Controllers — PHP, Java, Python

For annotated controllers, annotations live on the **implementation method**. Sindri reads the annotations and
constructs a route object — exactly the same shape as a route returned from `getRoutes()`. **No method body extraction.
No import resolution of the callable.** The callable from `#[RouteHandler]` is written directly into the generated
cache data class as a literal, just as it appears in the source.

This is identical to how service bindings work:

```php
// service binding — callable written as a literal
SomeServiceId::class => [SomeServiceProvider::class, 'publishSomeClass']

// route — callable written as a literal
new Route('/users/{id}', 'user.show', [SomeClass::class, 'theHandlerMethod'])
```

Sindri reads literals, writes literals. No execution, no body extraction, no cross-file resolution.

Go and TypeScript have no annotation support — routes are always registered explicitly via `getRoutes()`.

---

### Annotation Structure

```
#[Route]      — HTTP method + path — lives on the implementation method
#[Parameter]  — dynamic segment constraints — lives on the implementation method
#[RouteHandler]    — callable reference — lives on the implementation method
```

The callable in `#[RouteHandler]` is the value written into the generated route object unchanged.

---

### PHP

```php
class UserController
{
    // Sindri reads these annotations and constructs a Route object.
    // The callable [SomeClass::class, 'theHandlerMethod'] is written
    // directly into the generated cache as-is — no body extraction.
    #[Route('GET', '/users/{id}')]
    #[Parameter('id', pattern: '[0-9]+')]
    #[Handler([self::class, 'showHandler'])]
    public function show(RouteContract $route): ResponseContract
    {
        // actual implementation — irrelevant to Sindri
    }

    // The handler — may be on this class or any other class
    public static function showHandler(ContainerContract $c, RouteContract $route): ResponseContract
    {
        return $c->getSingleton(self::class)->show($route);
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

**Sindri reads — PHP:**

```
1. Find #[Route], #[Parameter], #[RouteHandler] on the implementation method
2. Extract path, HTTP method, parameter name/pattern, callable — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### Java

```java
public class UserController {

    // Sindri reads annotations and constructs a Route object.
    // Callable written directly into generated cache — no body extraction.
    @Route(method = "GET", path = "/users/{id}")
    @Parameter(name = "id", pattern = "[0-9]+")
    @RouteHandler(clazz = UserController.class, method = "showHandler")
    public ResponseContract show(RouteContract route) {
        // actual implementation — irrelevant to Sindri
    }

    public static ResponseContract showHandler(ContainerContract c, RouteContract route) {
        return c.getSingleton(UserController.class).show(route);
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

**Sindri reads — Java:**

```
1. Find @Route, @Parameter, @RouteHandler on the implementation method
2. Extract path, HTTP method, parameter name/pattern, clazz + method — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### Python

```python
class UserController:

    # Sindri reads these decorators and constructs a Route object.
    # The callable tuple is written directly into the generated cache — no body extraction.
    @route('GET', '/users/{id}')
    @parameter('id', pattern='[0-9]+')
    @route_handler((UserController, 'show_handler'))  # callable tuple — written as-is
    def show(self, route: RouteContract) -> ResponseContract:
        pass  # actual implementation — irrelevant to Sindri

    @staticmethod
    def show_handler(c: ContainerContract, route: RouteContract) -> ResponseContract:
        return c.get_singleton(UserController).show(route)
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

**Sindri reads — Python:**

```
1. Find @route, @parameter, @route_handler decorators on the implementation method
2. Extract path, HTTP method, parameter name/pattern, callable tuple — all literals
3. Construct route data from extracted literals
4. Write into generated AppHttpRoutingData — callable written as-is
```

---

### The Sindri Pattern (All Languages)

```
Annotations / decorators carry literals.
Sindri reads literals.
Sindri writes literals into the generated cache data class.
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

## Why `sindri` Does Not Generate the Handler

The developer writes the handler. `sindri` extracts the handler that the developer wrote, and generates none of its own.
Three reasons hold:

**A generated handler breaks the no-cache runtime path.** The framework must run correctly before anyone runs the build
tool. A route whose handler exists only in a generated cache file has no handler on the no-cache path. The no-cache path
must behave the same as the cached path.

**A generated handler needs dependency inference.** To write a handler for a method, `sindri` must read the method
signature, resolve each parameter type to a container binding, and write the matching `getSingleton()` call. That
inference fails on a complex or an ambiguous signature. The developer knows which binding the method needs.

**A generated handler blocks custom logic.** A developer who logs a call, transforms a value, or branches before the
call to the method has no place for that code. A generated handler is a fixed template.

The developer writes a handler method next to the implementation method, and that cost is real. In exchange the handler
is explicit, a reader sees what it calls, the compiler or the analyzer checks it, a test replaces it in isolation, and
it behaves the same with a cache and without one.

**Future consideration:** `sindri` could generate handler scaffolding from an annotated method, for a developer who opts
in. It would write a handler stub for the developer to review and change. It would inject no logic at build time, and it
would not change the runtime path. The current version does not support this.
