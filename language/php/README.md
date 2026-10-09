# PHP Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `HANDLERS.md`, `DATA_CACHE.md`, `SINDRI.md`

PHP is the most complete port. The documents in `component/` and `convention/` are what every port is measured
against, PHP included.

> **Warning: this document is a plan, and parts of it predate the current contracts.** It was written before the
> component and convention documents existed, so an item here may describe a shape those documents have since settled
> differently. **They govern.** Read the component's own document before you act on an item below, and treat a
> disagreement as this plan being out of date.

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

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### No per-component constants files

**Retired.** PHP names a class natively and the compiler checks it, so a constants file of FQN strings adds a second
place to get the name wrong. Only a port whose binding keys are strings holds one. See
[`CONTAINER_BINDINGS.md`](../../convention/CONTAINER_BINDINGS.md).

Binding-key constants that a port genuinely needs are a different thing, and this does not retire those.

### Container bindings declare their factory

A provider's `publishers()` map holds a reference to a named method, because the build tool reads that map statically. A
binding registered directly takes any callable. Either way, remove dynamic reflection-based instantiation:

```php
// Right — in a provider's publishers() map, a reference to a named method.
// The build tool reads this map, so the value has to be a value it can carry.
public function publishers(): array
{
    return [RouterContract::class => [self::class, 'publishRouter']];
}
```

```php
// Right — registered directly, where nothing reads the declaration statically.
// Any callable is fine here, an inline closure included.
$container->bind(
    RouterContract::class,
    static fn(ContainerContract $c): RouterContract => new Router(
        $c->getSingleton(MatcherContract::class)
    )
);
```

A binding made outside a provider cannot reach the generated cache, whichever form it takes.

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
        NotifierContract::class => [self::class, 'publishNotifier'],
    ];
}

public static function publishNotifier(ContainerContract $container): void
{
    $container->setSingleton(
        NotifierContract::class,
        new TeamsNotifier()
    );
}
```

Each publisher is a static method that takes the container and returns nothing. The publisher constructs the service
inline and registers the result with `setSingleton()`. A publisher that needs a dependency reads it from the container
by its contract.

### Static `make()` factory — an optional alternative

The publisher constructs the service inline. This is what the framework does everywhere. A service class implements its
contract and carries no registration code.

A service class may instead expose a static `make()` factory that the publisher delegates to:

```php
class SlackNotifier implements NotifierContract
{
    public function __construct(private string $webhookUrl) {}

    public static function make(ContainerContract $container, array $arguments = []): static
    {
        return new static($container->getSingleton(HttpConfig::class)->key);
    }
}

public static function publishNotifier(ContainerContract $container): void
{
    $container->setSingleton(NotifierContract::class, SlackNotifier::make($container));
}
```

The signature matches the `callable` that `bind()` and `bindSingleton()` accept, so `[SlackNotifier::class, 'make']`
also works as a direct binding. Use this when a class owns a construction step that more than one caller must reuse.
Otherwise construct the service in the publisher. Neither form uses reflection or autowiring.

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
literals with no conditional logic, variables, or method calls other than constructors.

---

## 4. Handler Signatures

**Reference:** `HANDLERS.md`

### The typed handler signatures

These are settled and already implemented. A route handler takes the container and the **route**; only a listener takes
a map. See [`HANDLERS.md`](../../convention/HANDLERS.md).

```php
// HTTP routes
/** callable(ContainerContract, RouteContract): ResponseContract */

// CLI routes
/** callable(ContainerContract, RouteContract): OutputContract */

// Event listeners
/** callable(ContainerContract, array<string, mixed>): mixed */
```

### Add the handler attributes to route/listener data classes

Routes need `#[RouteHandler]` attribute support on controller/action methods, and listeners need `#[ListenerHandler]`.
Each attribute carries a reference to the handler method, not an inline closure — an attribute argument has to be a
constant expression, so a closure does not compile there:

```php
#[RouteHandler([UserController::class, 'showHandler'])]
#[Parameter('id', pattern: '[0-9]+')]
public function show(int $id): ResponseContract {}
```

### Add #[Parameter] attribute

Routes with dynamic segments need `#[Parameter]` attribute support on controller/action methods carrying the parameter
name and pattern.

---

## 5. File Generation → sindri

**Reference:** `SINDRI.md`

### Move the file generation and the `make:*` commands to sindri

- Move all file generation, scaffolding, and `make:*` commands to `sindri`
- The framework must have zero AST or build tooling dependencies after this change
- `sindri` is a `require-dev` dependency only — never in production

### Cache generation in sindri — complete

`sindri` walks the provider tree with nikic/php-parser and generates `AppContainerData`, `AppEventData`,
`AppHttpRoutingData`, and `AppCliRoutingData`. The framework holds no cache generation command.

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
    public function getControllerClasses(): array;
    public function getRoutes(): array;
}
```

### Implement ListenerProviderContract

```php
interface ListenerProviderContract
{
    public function getListenerClasses(): array;
    public function getListeners(): array;
}
```

---

## 7. Application Config as Build Tool Entry Point

**Reference:** `SINDRI.md`, `DATA_CACHE.md`

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
        new App\Providers\AppProvider(),
    ]
)
```

### Drop the component provider constants class

A constants class that provides string aliases for component provider class references must not be created. If it
exists, remove it. It would allow developers to write `HttpConstants::HTTP_COMPONENT_PROVIDER` in the config which the
build tool cannot resolve from AST.

A binding-key constants file that a port genuinely needs is a different thing, and is unaffected. PHP needs none.

### Ensure all provider list methods return constructed providers

Audit all provider list methods (`getComponentProviders`, `getContainerProviders`, `getHttpProviders` etc.) to ensure
they return a literal list of constructed providers — never a constant reference, and never a `::class` string, which
the declared provider-contract array type does not accept. A publish callback is the separate case that takes a
method reference.

---

## Priority Order

1. **Throwable renaming and abstraction** — foundational, everything else builds on stable exception types
2. **Provider contract interfaces**
3. **publishers() map migration**
4. **Handler contracts and #[RouteHandler] attribute**
5. **#[Parameter] attribute**
6. **File generation and `make:*` commands to sindri**
7. **Retired** — per-component container constants files are not built; see §2
8. **Explicit binding factories** — additive, can happen incrementally per component
