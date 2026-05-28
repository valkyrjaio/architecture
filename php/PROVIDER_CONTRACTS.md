# Valkyrja Provider Contracts — PHP

## Overview

PHP provider contracts previously used static methods with `::class` string references. They now use instance methods
returning provider instances directly, aligning with all other Valkyrja language ports:

- `new X()` used in array literals — PHP-Parser's `Expr\New_` AST node carries the class name directly
- Instance methods on `ComponentProviderContract` — the framework instantiates providers and calls methods directly
- `publishers()` is an instance method returning a map of binding key to static callable reference
- `#[Handler]` attribute on annotated controller/listener methods — build tool reads attribute argument from AST
- All methods must return simple array literals with no conditional logic

### PHP Works Without Cache

Provider methods return instances directly. The framework traverses the provider tree by direct method calls — no
string lookup, no registry needed:

```php
// framework bootstrap — direct method calls, no cache needed
foreach ($app->getProviders() as $component) {
    foreach ($component->getContainerProviders($app) as $provider) {
        foreach ($provider->publishers() as $key => $callback) {
            $container->bind($key, $callback); // direct call ✅
        }
    }
    foreach ($component->getHttpProviders($app) as $provider) {
        foreach ($provider->getRoutes() as $route) { // direct call ✅
            $router->register($route);
        }
    }
}
```

Cache is a cold-start optimization for CGI and serverless deployments. Direct method calls are the
runtime-correct path; the cache eliminates repeated traversal only.

---

## ComponentProviderContract

Top-level aggregator. Returns arrays of sub-provider **instances** by category. Build tool reads return values directly
from AST — must be simple array literals with no conditional logic. Each array element must be a `new X()` expression —
Sindri reads `Expr\New_.class` (a `Name` node) to extract the provider class name.

```php
// package: Valkyrja\Application\Provider\Contract
namespace Valkyrja\Application\Provider\Contract;

use Valkyrja\Application\Kernel\Contract\ApplicationContract;
use Valkyrja\Cli\Routing\Provider\Contract\CliRouteProviderContract;
use Valkyrja\Container\Provider\Contract\ServiceProviderContract;
use Valkyrja\Event\Provider\Contract\ListenerProviderContract;
use Valkyrja\Http\Routing\Provider\Contract\HttpRouteProviderContract;

interface ComponentProviderContract
{
    /**
     * Get the component providers this component depends on.
     * The framework ensures all listed components are fully registered
     * before this component's providers are registered.
     * Must return a simple array literal — no conditional logic.
     * Each element must be a new X() expression.
     *
     * @return array<ComponentProviderContract>
     */
    public function getComponentProviders(ApplicationContract $app): array;

    /**
     * Get the component's container service providers.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<ServiceProviderContract>
     */
    public function getContainerProviders(ApplicationContract $app): array;

    /**
     * Get the component's event listener providers.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<ListenerProviderContract>
     */
    public function getEventProviders(ApplicationContract $app): array;

    /**
     * Get the component's CLI route providers.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<CliRouteProviderContract>
     */
    public function getCliProviders(ApplicationContract $app): array;

    /**
     * Get the component's HTTP route providers.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<HttpRouteProviderContract>
     */
    public function getHttpProviders(ApplicationContract $app): array;
}
```

### HttpComponentProvider Implementation

```php
namespace Valkyrja\Http\Provider;

use Valkyrja\Application\Kernel\Contract\ApplicationContract;
use Valkyrja\Application\Provider\Contract\ComponentProviderContract;
use Valkyrja\Cli\Routing\Provider\Contract\CliRouteProviderContract;
use Valkyrja\Container\Provider\Contract\ServiceProviderContract;
use Valkyrja\Event\Provider\Contract\ListenerProviderContract;
use Valkyrja\Http\Routing\Provider\Contract\HttpRouteProviderContract;

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
            new HttpContainerProvider(),
            new HttpMiddlewareProvider(),
        ];
    }

    public function getEventProviders(ApplicationContract $app): array
    {
        return [
            new HttpEventProvider(),
        ];
    }

    public function getCliProviders(ApplicationContract $app): array
    {
        return [];
    }

    public function getHttpProviders(ApplicationContract $app): array
    {
        return [
            new HttpRouteProvider(),
        ];
    }
}
```

---

## ServiceProviderContract

Container bindings provider. `publishers()` returns a map of binding key to publisher callable reference. The build tool
reads the map from AST, resolves each callable reference, and reads that method body directly.

```php
namespace Valkyrja\Container\Provider\Contract;

use Valkyrja\Container\Manager\Contract\ContainerContract;

interface ServiceProviderContract
{
    /**
     * Any custom publishers for services provided by this provider.
     *
     * The map must be a simple array literal — no conditional logic.
     * Each value must be a static callable on the same class.
     * The build tool reads the map from AST, resolves each callable,
     * and reads that method body directly for cache generation.
     *
     * @return array<class-string, callable(ContainerContract): void>
     *
     * Example:
     *   return [
     *       UserRepositoryContract::class => [self::class, 'publishUserRepository'],
     *   ];
     *
     *   public static function publishUserRepository(ContainerContract $container): void
     *   {
     *       $container->setSingleton(
     *           UserRepositoryContract::class,
     *           new UserRepository($container->make(DatabaseContract::class))
     *       );
     *   }
     */
    public function publishers(): array;
}
```

### UserServiceProvider Implementation

```php
namespace App\Provider;

use Valkyrja\Container\Manager\Contract\ContainerContract;
use Valkyrja\Container\Provider\Contract\ServiceProviderContract;
use App\Repository\UserRepository;
use App\Repository\Contract\UserRepositoryContract;
use App\Service\Contract\DatabaseContract;

class UserServiceProvider implements ServiceProviderContract
{
    /**
     * Build tool reads this map from AST, resolves each callable,
     * then reads each publisher method body for cache generation.
     */
    public function publishers(): array
    {
        return [
            UserRepositoryContract::class => [self::class, 'publishUserRepository'],
        ];
    }

    /**
     * Build tool reads this method body from AST for cache generation.
     * No annotation needed — method is discovered via publishers() map.
     */
    public static function publishUserRepository(ContainerContract $container): void
    {
        $container->setSingleton(
            UserRepositoryContract::class,
            new UserRepository($container->make(DatabaseContract::class))
        );
    }
}
```

---

## HttpRouteProviderContract

HTTP route provider. Two sources: annotated controller classes (scanned for `#[Handler]`) and explicit route object
definitions.

```php
namespace Valkyrja\Http\Routing\Provider\Contract;

use Valkyrja\Http\Routing\Data\Contract\RouteContract;

interface HttpRouteProviderContract
{
    /**
     * Get a list of attributed controller or action classes.
     * Build tool scans each class for #[Handler] attributes.
     * Returns empty array if using explicit routes only.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<class-string>
     */
    public function getControllerClasses(): array;

    /**
     * Get a list of explicit route definitions.
     * Routes are complete data structures — they carry HTTP method, path pattern,
     * dynamic segment constraints, middleware chain, and handler together.
     * Must return a simple array literal — no conditional logic.
     *
     * @return array<RouteContract>
     */
    public function getRoutes(): array;
}
```

### UserHttpRouteProvider Implementation

```php
namespace App\Http\Provider;

use Valkyrja\Container\Manager\Contract\ContainerContract;
use Valkyrja\Http\Message\Response\Contract\ResponseContract;
use Valkyrja\Http\Routing\Data\HttpRoute;
use Valkyrja\Http\Routing\Data\Contract\RouteContract;
use Valkyrja\Http\Routing\Provider\Contract\HttpRouteProviderContract;
use App\Http\Controller\UserController;
use App\Http\Controller\OrderController;

class UserHttpRouteProvider implements HttpRouteProviderContract
{
    public function getControllerClasses(): array
    {
        return [
            UserController::class,
            OrderController::class,
        ];
    }

    /**
     * Handler is a callable on this same class.
     * Sindri reads the handler method body from this file only — no cross-file imports.
     */
    public function getRoutes(): array
    {
        return [
            HttpRoute::get('/orders', [self::class, 'indexOrders']),
        ];
    }

    /** Handler method lives on the same class — all imports self-contained. */
    public static function indexOrders(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(OrderController::class)->index($args);
    }
}
```

---

## Annotated Controller — PHP

`#[Handler]` lives on the **implementation method** and carries a callable reference — class + method name. The handler
may live on the controller, the route provider, or any other class.

**Handler on the same controller:**

```php
namespace App\Http\Controller;

use Valkyrja\Container\Manager\Contract\ContainerContract;
use Valkyrja\Http\Message\Response\Contract\ResponseContract;
use Valkyrja\Http\Routing\Attribute\Handler;
use Valkyrja\Http\Routing\Attribute\Parameter;
use Valkyrja\Http\Routing\Attribute\Route;

class UserController
{
    #[Route(method: 'GET', path: '/users/{id}')]
    #[Parameter(name: 'id', pattern: '[0-9]+')]
    #[Handler(class: UserController::class, method: 'showHandler')]
    public function show(string $id): ResponseContract
    {
        return $this->userService->findById($id)->toResponse();
    }

    #[Route(method: 'POST', path: '/users')]
    #[Handler(class: UserController::class, method: 'storeHandler')]
    public function store(array $data): ResponseContract
    {
        // actual implementation
    }

    // Sindri resolves Handler → this file, reads this method body using this file's imports
    public static function showHandler(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(self::class)->show($args['id']);
    }

    public static function storeHandler(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(self::class)->store($args);
    }
}
```

**Handler on the route provider:**

```php
class UserController
{
    // #[Handler] points to the route provider — Sindri follows the callable
    #[Route(method: 'GET', path: '/users/{id}')]
    #[Parameter(name: 'id', pattern: '[0-9]+')]
    #[Handler(class: UserHttpRouteProvider::class, method: 'showUser')]
    public function show(string $id): ResponseContract { ... }
}

class UserHttpRouteProvider implements HttpRouteProviderContract
{
    // Sindri resolves callable → this file, reads this method using this file's imports
    public static function showUser(ContainerContract $c, array $args): ResponseContract
    {
        return $c->getSingleton(UserController::class)->show($args['id']);
    }
}
```

---

## CliRouteProviderContract

```php
namespace Valkyrja\Cli\Routing\Provider\Contract;

use Valkyrja\Cli\Routing\Data\Contract\RouteContract;

interface CliRouteProviderContract
{
    /** @return array<class-string> */
    public function getControllerClasses(): array;

    /** @return array<RouteContract> */
    public function getRoutes(): array;
}
```

---

## ListenerProviderContract

```php
namespace Valkyrja\Event\Provider\Contract;

use Valkyrja\Event\Data\Contract\ListenerContract;

interface ListenerProviderContract
{
    /** @return array<class-string> */
    public function getListenerClasses(): array;

    /** @return array<ListenerContract> */
    public function getListeners(): array;
}
```

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic:

```php
// ✅ simple array of instances
return [new HttpContainerProvider(), new HttpMiddlewareProvider()];

// ✅ simple map with callable reference
return [UserRepositoryContract::class => [self::class, 'publishUserRepository']];

// ✅ simple array of route objects
return [HttpRoute::get('/users', [self::class, 'indexUsers'])];

// ❌ conditional logic
if ($condition) { return [...]; }

// ❌ variable accumulation
$routes = []; $routes[] = ...; return $routes;

// ❌ inline closures as route handlers
return [HttpRoute::get('/users', function (ContainerContract $c, array $args) { ... })];
```

---

## Handler Method Pointer Convention

All handler methods must be **static methods on the same class** as the provider or controller that defines the route or
listener. This is the same pattern used by `publishers()` in service providers.

**Why:** Sindri reads exactly one file per provider or controller. All imports for handler bodies are in that one file —
no cross-file import aggregation, no conflict detection, no registry needed.

```
✅ Callable on the same class: [self::class, 'methodName']
✅ All type references imported in the same file

❌ Inline closures or lambdas in route/listener definitions
❌ References to types not imported in the current file
❌ Handler methods on a different class
```

---

## Design Note — Why Routes Cannot Use a Publisher-Style Map

An early consideration was expressing routes the same way as container bindings — a map of route key to handler callable,
with the build tool reading method bodies directly. This was rejected because routes are multi-dimensional data
structures, not simple key→factory pairs.

A route carries: HTTP method, path pattern, dynamic segment constraints, regex compilation data, middleware chain,
name/alias, parameter defaults, host constraints, and scheme constraints — all in addition to the handler. The
`HttpRoute::get("/users/{id}", handler)` call populates all of these fields together. Decomposing this into a key/body
map would lose all metadata the router needs to build its dispatcher index.

The same reasoning applies to listeners — they carry event type binding, priority, and stop-propagation behavior
alongside the handler. These cannot be expressed as a flat key/body map without losing data the event dispatcher
requires.

Container bindings by contrast are simple key→factory pairs. This is why `publishers()` works as a map but `getRoutes()`
must return complete route objects.