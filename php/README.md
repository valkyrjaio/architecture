# PHP Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `HANDLERS.md`, `DATA_CACHE.md`, `BUILD_TOOL.md`

PHP is the reference implementation. All other ports are measured against it. The following changes are required to
bring the existing implementation into alignment with the decisions made during cross-port planning.

---

## Status

**Existing implementation requires the following changes.** Nothing here is net-new architecture — it is alignment work
to make PHP consistent with the cross-port decisions.

---

## 1. Throwables — Rename and Abstract

**Reference:** `THROWABLES.md`

### Rename all exceptions and throwables

Every exception and throwable across every component must be renamed to follow the convention:

- Framework base → `Valkyrja*` (e.g. `ValkyrjaThrowable`, `ValkyrjaRuntimeException`,
  `ValkyrjaInvalidArgumentException`)
- Component → `ComponentName*` (e.g. `ContainerRuntimeException`, `HttpRuntimeException`)
- Shared subcomponent → `ParentComponentSubComponent*` (e.g. `HttpRoutingRuntimeException`,
  `CliRoutingRuntimeException`)
- Unique subcomponent → `SubComponent*` (e.g. `RequestRuntimeException`, `ResponseRuntimeException`)
- Sub-subcomponent → prepend only as many parent names as needed to make the name unique across the framework

### Make all base and categorical exceptions abstract

- `ValkyrjaThrowable` → abstract
- `ValkyrjaRuntimeException` → abstract
- `ValkyrjaInvalidArgumentException` → abstract
- Every `Component*RuntimeException` → abstract
- Every `Component*InvalidArgumentException` → abstract

### Ensure every component has categorical abstracts

Every component must ship `ComponentRuntimeException` and `ComponentInvalidArgumentException` even if currently unused.
Add where missing.

### Create specific concrete exceptions per throw site

Audit every `throw` statement in the codebase. Every throw must use a specific concrete exception named for the problem.
No throwing abstract base exceptions.

---

## 2. Container Bindings — Constants and Closures

**Reference:** `CONTAINER_BINDINGS.md`

### Add per-component constants files

Every component needs a constants file containing FQN string identifiers for all classes, interfaces, and contracts in
that component:

```php
// Http/HttpConstants.php
final class HttpConstants
{
    public const ROUTER           = RouterContract::class;
    public const REQUEST          = ServerRequestContract::class;
    public const RESPONSE_FACTORY = ResponseFactoryContract::class;
}
```

### Migrate container bindings to closure-based factories

All container bindings must use explicit closure factories. Remove all dynamic reflection-based instantiation:

```php
// Wrong — the container instantiates the class through reflection.
$container->bind(RouterContract::class);

// Right — an explicit closure factory names every dependency.
$container->bind(
    RouterContract::class,
    static fn(ContainerContract $c): RouterContract => new Router(
        $c->getSingleton(MatcherContract::class)
    )
);
```

---

## 3. Service Providers — publishers() map

**Reference:** `DATA_CACHE.md`, `CONTAINER_BINDINGS.md`, `PROVIDERS.md`

### publishers() map — the sole registration mechanism

Service providers must return a `publishers()` map of service IDs to static method references. The `provides()` method
from earlier versions is removed — the publishers map is the sole source of truth. Sindri reads this map via AST:

```php
public function publishers(): array
{
    return [
        RouterContract::class => [self::class, 'publishRouter'],
    ];
}

public static function publishRouter(ContainerContract $container): void
{
    $container->setSingleton(
        RouterContract::class,
        new Router($container->getSingleton(MatcherContract::class))
    );
}
```

### Static `make()` factory — an optional alternative

The publisher callback above constructs the service inline. This is what the framework does everywhere. A service class
implements its contract and carries no registration code.

A service class may instead expose a static `make()` factory that the publisher delegates to:

```php
class Router implements RouterContract
{
    public static function make(ContainerContract $container, array $arguments = []): static
    {
        return new static($container->getSingleton(MatcherContract::class));
    }
}

public static function publishRouter(ContainerContract $container): void
{
    $container->setSingleton(RouterContract::class, Router::make($container));
}
```

The signature matches the `callable` that `bind()` and `bindSingleton()` accept, so `[Router::class, 'make']` also works
as a direct binding. Use this when a class owns a construction step that more than one caller must reuse. Otherwise
construct the service in the publisher. Neither form uses reflection or autowiring.

### Binding methods available in publisher callbacks

Publisher callbacks have access to the full container binding API:

| Method                        | Use                                                                   |
| ----------------------------- | --------------------------------------------------------------------- |
| `setSingleton(id, instance)`  | Register an already-constructed singleton — most common in publishers |
| `bindSingleton(id, callable)` | Register a deferred singleton with a callable factory                 |
| `bind(id, callable)`          | Register a per-call service (fresh instance every resolution)         |
| `bindAlias(alias, id)`        | Map one service ID to another                                         |

### Provider list methods must return simple list literals

All `getComponentProviders()`, `getContainerProviders()`, `getEventProviders()`, `getCliProviders()`,
`getHttpProviders()`, `getControllerClasses()`, `getRoutes()`, `getListeners()` methods must return simple array
literals with no conditional logic, variables, or method calls other than constructors and static factories.

---

## 4. Handler Contracts — Typed Closures

**Reference:** `HANDLERS.md`

### Add typed handler function types

Define the three handler function types as docblock-enforced closure signatures:

```php
// HTTP routes
/** Closure(ContainerContract, array<string, mixed>): ResponseContract */

// CLI routes
/** Closure(ContainerContract, array<string, mixed>): OutputContract */

// Event listeners
/** Closure(ContainerContract, array<string, mixed>): mixed */
```

### Add HttpHandlerContract, CliHandlerContract, ListenerHandlerContract

Each concern gets its own handler contract extending the base `HandlerContract` with the typed closure signature.

### Add #[Handler] attribute to route/listener data classes

Routes and listeners need `#[Handler]` attribute support on controller/action methods. The attribute carries the typed
closure:

```php
#[Handler(static fn(ContainerContract $c, array<string, mixed> $args): ResponseContract
    => $c->getSingleton(UserController::class)->show($args['id']))]
#[Parameter('id', pattern: '[0-9]+')]
public function show(int $id): ResponseContract {}
```

### Add #[Parameter] attribute

Routes with dynamic segments need `#[Parameter]` attribute support on controller/action methods carrying the parameter
name and pattern.

---

## 5. Bin → sindri

**Reference:** `BUILD_TOOL.md`

### Extract Bin component to separate repository

- Create `sindri` as a separate Composer package
- Move all file generation, scaffolding, and `make:*` commands to the new package
- Add `nikic/php-parser` as a dependency of `sindri`, not the framework
- The framework must have zero AST or build tooling dependencies after this change
- `sindri` is a `require-dev` dependency only — never in production

### Cache generation in sindri — complete

`sindri` walks the provider tree with nikic/php-parser and generates `AppContainerData`, `AppEventData`,
`AppHttpRoutingData`, and `AppCliRoutingData`. The framework holds no cache command.

---

## 6. Provider Contracts

**Reference:** `DATA_CACHE.md`

### Implement ComponentProviderContract

```php
interface ComponentProviderContract
{
    /**
     * Get the component providers this component depends on.
     * The framework ensures all listed components are fully registered
     * before this component's providers are registered.
     * Sindri uses this during the dependency resolution pass (Step 1a) to build
     * the full ordered, deduplicated component list before walking any providers.
     */
    public function getComponentProviders(ApplicationContract $app): array;
    public function getContainerProviders(ApplicationContract $app): array;
    public function getEventProviders(ApplicationContract $app): array;
    public function getCliProviders(ApplicationContract $app): array;
    public function getHttpProviders(ApplicationContract $app): array;
}
```

Example implementation:

```php
class HttpComponentProvider implements ComponentProviderContract
{
    public function getComponentProviders(ApplicationContract $app): array
    {
        return [
            new ContainerComponentProvider(),  // HTTP depends on Container
            new EventComponentProvider(),       // HTTP depends on Event
        ];
    }

    public function getContainerProviders(ApplicationContract $app): array
    {
        return [
            new HttpServiceProvider(),
            new HttpMiddlewareProvider(),
        ];
    }

    public function getEventProviders(ApplicationContract $app): array
    {
        return [new HttpListenersProvider()];
    }

    public function getCliProviders(ApplicationContract $app): array
    {
        return [];
    }

    public function getHttpProviders(ApplicationContract $app): array
    {
        return [new HttpRoutesProvider()];
    }
}
```

### Implement HttpRouteProviderContract and CliRouteProviderContract

```php
interface HttpRouteProviderContract
{
    public static function getControllerClasses(): array;
    public static function getRoutes(): array;
}
```

### Implement ListenerProviderContract

```php
interface ListenerProviderContract
{
    public static function getListenerClasses(): array;
    public static function getListeners(): array;
}
```

---

## 7. Application Config as Build Tool Entry Point

**Reference:** `BUILD_TOOL.md`, `DATA_CACHE.md`

### No valkyrja.yaml needed

The application config class is the build tool entry point — it already lists all component providers. No separate yaml
file required.

```php
// AppConfig — this IS the build tool entry point
new AppConfig(
    providers: [
        new HttpComponentProvider(),
        new ContainerComponentProvider(),
        new EventComponentProvider(),
        new CliComponentProvider(),
        App\Providers\AppProvider::class,
    ]
)
```

### Drop the component provider constants class

A constants class that provides string aliases for component provider class references must not be created. If it
exists, remove it. It would allow developers to write `HttpConstants::HTTP_COMPONENT_PROVIDER` in the config which the
build tool cannot resolve from AST.

Binding key constants files (for container bindings) are unaffected — they are correct and should remain.

### Ensure all provider list methods use ::class directly

Audit all provider list methods (`getComponentProviders`, `getContainerProviders`, `getHttpProviders` etc.) to ensure
they return `::class` references directly — never constant references.

---

## Priority Order

1. **Throwable renaming and abstraction** — foundational, everything else builds on stable exception types
2. **Provider contract interfaces** — needed before build tool work
3. **publishers() map migration** — needed before build tool work
4. **Handler contracts and #[Handler] attribute**
5. **#[Parameter] attribute**
6. **Bin extraction to sindri**
7. **Container constants files** — additive, can happen incrementally per component
8. **Closure-based container bindings** — additive, can happen incrementally per component
