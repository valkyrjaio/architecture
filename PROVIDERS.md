# Providers

Providers are the mechanism by which Valkyrja's framework and application code is wired together. They register
container bindings, routes, CLI commands, and event listeners. This document covers the provider type hierarchy, the
naming convention, and the contracts each provider type must satisfy.

---

## Provider Type Hierarchy

There are five provider contract types, each with a distinct responsibility:

```
ComponentProviderContract       — top-level aggregator per component
  ├── ServiceProviderContract   — container bindings (publishers map)
  ├── HttpRouteProviderContract — HTTP routes and controllers
  ├── CliRouteProviderContract  — CLI commands and controllers
  └── ListenerProviderContract  — event listeners
```

A `ComponentProvider` groups the other four types for a given component. It is the entry point the application config
references. The other four are the leaf providers that actually register things.

---

## ComponentProviderContract

The top-level aggregator. One per component. Returns lists of the four leaf provider types, categorized by concern.

```php
// PHP
interface ComponentProviderContract
{
    public static function getContainerProviders(ApplicationContract $app): array;
    public static function getEventProviders(ApplicationContract $app): array;
    public static function getHttpProviders(ApplicationContract $app): array;
    public static function getCliProviders(ApplicationContract $app): array;
}
```

All four methods must return simple list literals — no conditional logic, no variables, no loops. The build tool reads
these from AST.

**One ComponentProvider per component.** The application config lists only ComponentProviders:

```php
class AppConfig implements ConfigContract
{
    public static function getProviders(): array
    {
        return [
            HttpComponentProvider::class,
            ContainerComponentProvider::class,
            EventComponentProvider::class,
            CliComponentProvider::class,
            AppComponentProvider::class,    // application-defined
        ];
    }
}
```

---

## ServiceProviderContract

Registers container bindings. Returns a map of binding key to publisher method reference.

```php
// PHP
interface ServiceProviderContract
{
    public static function publishers(): array;
}
```

```java
// Java
public interface ServiceProviderContract {
    static Map<Class<?>, Runnable> publishers();
}
```

```go
// Go
type ServiceProviderContract interface {
Publishers() map[string]func (ContainerContract)
}
```

```python
# Python
class ServiceProviderContract(ABC):
    @staticmethod
    @abstractmethod
    def publishers() -> dict[str, Callable[[ContainerContract], None]]: ...
```

```typescript
// TypeScript
interface ServiceProviderContract {
    publishers(): Record<string, (c: ContainerContract) => void>
}
```

**The `publishers()` map** — keys are binding identifiers, values are method references on the same class. The build
tool reads this from AST and writes each value as a lambda in the generated `AppContainerData`.

```php
class HttpServiceProvider implements ServiceProviderContract
{
    public static function publishers(): array
    {
        return [
            RouterContract::class            => [self::class, 'publishRouter'],
            RouteDispatcherContract::class   => [self::class, 'publishDispatcher'],
            MiddlewareContract::class        => [self::class, 'publishMiddleware'],
        ];
    }

    public static function publishRouter(ContainerContract $c): void
    {
        $c->setSingleton(RouterContract::class, new Router($c->make(DispatcherContract::class)));
    }

    // ... other publisher methods
}
```

**Container lambda wrapping (Python only):** The Python container wraps each publisher method reference in a lambda when
registering from a provider at runtime. The generated cache already stores lambdas. Resolution is always `binding()()` —
uniform, no conditional check. See `CONTAINER_BINDINGS.md`.

---

## HttpRouteProviderContract

Registers HTTP routes, either explicitly via `getRoutes()` or via annotated controller classes via
`getControllerClasses()`.

```php
// PHP
interface HttpRouteProviderContract
{
    public static function getControllerClasses(): array;  // PHP, Java, Python only
    public static function getRoutes(): array;
}
```

```go
// Go — no getControllerClasses(), annotations not supported
type HttpRouteProviderContract interface {
GetRoutes() []RouteContract
}
```

```typescript
// TypeScript — no getControllerClasses(), annotations not supported
interface HttpRouteProviderContract {
    getRoutes(): RouteContract[]
}
```

**Handler method pointer convention** — route handlers must be static methods on the same provider class (or pointed to
via `#[Handler]` callable on annotated controllers). No inline closures or lambdas in route definitions.

```php
class UserHttpRouteProvider implements HttpRouteProviderContract
{
    public static function getControllerClasses(): array
    {
        return [UserController::class, OrderController::class];
    }

    public static function getRoutes(): array
    {
        return [
            HttpRoute::get('/users/{id}', [self::class, 'showUser']),
            HttpRoute::post('/users',     [self::class, 'createUser']),
        ];
    }

    public static function showUser(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(UserController::class)->show($args['id']);
    }

    public static function createUser(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(UserController::class)->create($args);
    }
}
```

---

## CliRouteProviderContract

Identical structure to `HttpRouteProviderContract`. CLI commands instead of HTTP routes. Handler returns
`OutputContract` instead of `ResponseContract`.

```php
interface CliRouteProviderContract
{
    public static function getControllerClasses(): array;  // PHP, Java, Python only
    public static function getRoutes(): array;
}
```

---

## ListenerProviderContract

Registers event listeners, either explicitly via `getListeners()` or via annotated listener classes via
`getListenerClasses()`.

```php
// PHP
interface ListenerProviderContract
{
    public static function getListenerClasses(): array;  // PHP, Java, Python only
    public static function getListeners(): array;
}
```

```go
// Go — no getListenerClasses()
type ListenerProviderContract interface {
GetListeners() []ListenerContract
}
```

```php
class UserEventListenerProvider implements ListenerProviderContract
{
    public static function getListenerClasses(): array
    {
        return [UserCreatedListener::class];
    }

    public static function getListeners(): array
    {
        return [
            Listener::on(UserCreatedEvent::class, [self::class, 'onUserCreated']),
        ];
    }

    public static function onUserCreated(ContainerContract $c, array $args): mixed
    {
        return $c->getSingleton(UserCreatedListener::class)->handle($args['user_id']);
    }
}
```

---

## Naming Convention

All provider implementations — framework and application-defined — must be **uniquely named across the entire framework
**. The naming rule is identical to the throwable naming rule: prepend parent component (and subcomponent if needed)
names until the name is unique.

### The Forcing Function

The generated `AppContainerData`, `AppEventData`, `AppHttpRoutingData`, and `AppCliRoutingData` files each reference
providers from multiple components in a single generated file. Identical class names across components produce namespace
collisions that prevent compilation. Unique names are a hard requirement, not a style preference.

**Why this wasn't a problem in PHP:** PHP callables in the cache used fully qualified class names —
`\Valkyrja\Http\Provider\HttpServiceProvider::publishRouter`. No import statement required, no collision possible. Two
classes named `ServiceProvider` in different namespaces coexist without conflict because the FQN distinguishes them.

Other languages do not support this:

- **Java** — `import` statements at the top of the file; two classes with the same simple name require aliasing or FQN
  usage throughout, which is non-idiomatic
- **Go** — package-qualified names (`http.ServiceProvider` vs `container.ServiceProvider`) would work but Go's
  convention is to use the unqualified name after import, which collides
- **Python** — `from app.http.provider import ServiceProvider` and `from app.container.provider import ServiceProvider`
  in the same generated file is a straight name collision — one overwrites the other
- **TypeScript** — same as Python; named imports collide on the simple name

Unique class names across the framework eliminate the collision entirely — no aliasing, no FQN workarounds, no
language-specific hacks. The convention that PHP could get away without now becomes a hard requirement for the
multi-language ports.

**Cross-application and package collisions:** The naming convention solves collisions within the framework but does not
prevent collisions between the framework and application code, or between third-party packages. If a developer's
application class name collides with a package class name in the generated file, that is the developer's responsibility
to resolve — the same way any import conflict in their codebase is theirs to fix. Sindri generates the best output it
can from what it reads and aggregates. The developer is responsible for ensuring the generated file is valid.

### The Pattern

```
ComponentName + Type + Provider
```

Where `Type` is one of:

- `Component` — top-level aggregator (one per component)
- `Service` — container bindings
- `HttpRoutes` — HTTP route definitions
- `CliRoutes` — CLI route definitions
- `Listeners` — event listener definitions

For shared subcomponents, prepend until unique — same recursive rule as throwables:

```
HttpRouting + HttpRoutes + Provider = HttpRoutingHttpRoutesProvider
```

### The Uniqueness Rule

> Is this name unique across the entire framework? If no, prepend the immediate parent name and ask again.

```
ServiceProvider           — not unique (every component has one)
HttpServiceProvider       — unique ✅

RoutesProvider            — not unique (HTTP and CLI both have routes)
HttpRoutesProvider        — unique ✅

MiddlewareProvider        — not unique (HTTP and CLI could both have middleware)
HttpMiddlewareProvider    — unique ✅
```

### Framework Provider Names

```
HttpComponentProvider               top-level HTTP aggregator
HttpServiceProvider                 HTTP container bindings
HttpRoutesProvider                  HTTP route definitions
HttpListenersProvider               HTTP-related event listeners

ContainerComponentProvider          top-level Container aggregator
ContainerServiceProvider            Container bindings

EventComponentProvider              top-level Event aggregator
EventServiceProvider                Event container bindings
EventListenersProvider              Event listeners

CliComponentProvider                top-level CLI aggregator
CliServiceProvider                  CLI container bindings
CliRoutesProvider                   CLI route definitions
```

For shared subcomponents where the subcomponent name alone is not unique, both parent and subcomponent prefix are
required:

```
HttpRoutingHttpRoutesProvider       HTTP routing subcomponent routes
HttpRoutingListenersProvider        HTTP routing subcomponent listeners
```

### Application-Defined Providers

Application providers follow the same rule. A developer extending the HTTP service provider prefixes with their
application or feature name:

```
AppHttpServiceProvider              application-level HTTP service override
UserHttpRoutesProvider              user feature HTTP routes
OrderHttpRoutesProvider             order feature HTTP routes
UserEventListenersProvider          user feature event listeners
```

Never:

```
HttpServiceProvider     — conflicts with the framework class
ServiceProvider         — ambiguous across the entire codebase
RoutesProvider          — ambiguous (HTTP? CLI?)
```

---

## Build Tool Requirements

All provider list methods must satisfy the build tool contract — simple literals, no conditional logic:

```
✅ Simple list/array literal
✅ Class references (::class / .class / ClassName / string constants)
✅ Method pointer references ([self::class, 'method'])
✅ Route/Listener constructor calls with literal arguments

❌ Conditional logic (if / switch / ternary)
❌ Variable references
❌ Loops or variable accumulation
❌ Inline closures or lambdas as route handlers
```

If any provider method violates this contract Sindri emits an error and aborts cache generation. The application still
runs without cache — the provider tree is traversed at runtime instead.

---

## Provider Registration Flow

```
AppConfig.getProviders()
  → [HttpComponentProvider, ContainerComponentProvider, ...]

HttpComponentProvider.getContainerProviders()
  → [HttpServiceProvider, HttpMiddlewareProvider]

HttpComponentProvider.getHttpProviders()
  → [UserHttpRoutesProvider, OrderHttpRoutesProvider]

HttpComponentProvider.getEventProviders()
  → [HttpRoutingListenersProvider]

HttpServiceProvider.publishers()
  → { RouterContract::class => [self::class, 'publishRouter'], ... }

UserHttpRoutesProvider.getRoutes()
  → [HttpRoute::get('/users/{id}', [self::class, 'showUser']), ...]

UserHttpRoutesProvider.getControllerClasses()   // PHP, Java, Python only
  → [UserController::class, ...]
```

At runtime (no cache) — the framework traverses this tree on every boot.
With cache — Sindri traverses this tree once at build time and writes the four data classes. The framework loads the
data classes directly, skipping the tree entirely.
