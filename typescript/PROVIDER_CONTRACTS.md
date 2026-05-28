# Valkyrja Provider Contracts — TypeScript

## Overview

TypeScript provider contracts differ from PHP/Java in several important ways:

- No reliable decorators — explicit registration only, no annotated class scanning
- `new X()` used in array literals — `NewExpression` nodes carry the class name directly
- Instance methods throughout — the framework receives provider instances and calls methods directly
- Publisher methods have no annotation — build tool reads method bodies directly from AST
- TypeScript compiler API resolves module paths via `tsconfig.json`
- All methods return simple array/object literals — no conditional logic

### TypeScript Works Without Cache

Provider methods return instances directly. The framework traverses the provider tree by direct method calls — no
string lookup, no registry needed:

```typescript
// framework bootstrap — direct method calls, no cache needed
for (const component of componentProviders) {
    for (const provider of component.getContainerProviders(app)) { // already an instance ✅
        for (const [key, callback] of Object.entries(provider.publishers())) {
            container.bind(key, callback) // direct call ✅
        }
    }
    for (const provider of component.getHttpProviders(app)) { // already an instance ✅
        for (const route of provider.getRoutes()) { // direct method call ✅
            router.register(route)
        }
    }
}
```

Cache is a cold-start optimization, not a correctness requirement.

---

## Type Hints

TypeScript has full generic type support for all return types:

| Method                    | Return type                                      | Reasoning                                             |
|---------------------------|--------------------------------------------------|-------------------------------------------------------|
| `getContainerProviders()` | `ServiceProviderContract[]`                      | Provider instances returned directly                  |
| `getEventProviders()`     | `ListenerProviderContract[]`                     | Provider instances returned directly                  |
| `getCliProviders()`       | `CliRouteProviderContract[]`                     | Provider instances returned directly                  |
| `getHttpProviders()`      | `HttpRouteProviderContract[]`                    | Provider instances returned directly                  |
| `getRoutes()`             | `RouteContract[]`                                | Concrete route data objects — fully typed             |
| `getListeners()`          | `ListenerContract[]`                             | Concrete listener data objects — fully typed          |
| `publishers()`            | `Record<string, (c: ContainerContract) => void>` | Maps binding key string to publisher method reference |

`getControllerClasses()` and `getListenerClasses()` are **not present** in the TypeScript contracts. TypeScript has no
reliable decorator/annotation support, so annotated class scanning is not possible. Including these methods would imply
capability that does not exist. Go has the same omission for the same reason.

`getRoutes()` and `getListeners()` are fully typed as `RouteContract[]` and `ListenerContract[]` — the TypeScript
compiler validates that all returned objects implement the correct contract at compile time.

---

## ComponentProviderContract

Top-level aggregator. Returns arrays of sub-provider **instances** by category. Build tool reads return values directly
from AST via TypeScript compiler API — must be simple array literals with no conditional logic. Each array element must
be a `new X()` expression — Sindri reads `NewExpression.expression` to extract the provider class name.

```typescript
// package: @valkyrja/application/provider/contract
import type {ApplicationContract} from '@valkyrja/application/kernel/contract'
import type {ServiceProviderContract} from '@valkyrja/container/provider/contract'
import type {ListenerProviderContract} from '@valkyrja/event/provider/contract'
import type {CliRouteProviderContract} from '@valkyrja/cli/routing/provider/contract'
import type {HttpRouteProviderContract} from '@valkyrja/http/routing/provider/contract'

/**
 * Defines what a component provider must implement.
 * All methods must return simple array literals.
 * No conditional logic permitted — build tool reads these from AST.
 */
export interface ComponentProviderContract {
    /** Get the component providers this component depends on. The framework ensures all listed components are fully registered before this component's providers are registered. */
    getComponentProviders(app: ApplicationContract): ComponentProviderContract[]

    getContainerProviders(app: ApplicationContract): ServiceProviderContract[]

    getEventProviders(app: ApplicationContract): ListenerProviderContract[]

    getCliProviders(app: ApplicationContract): CliRouteProviderContract[]

    getHttpProviders(app: ApplicationContract): HttpRouteProviderContract[]
}
```

### HttpComponentProvider Implementation

```typescript
import type {ApplicationContract} from '@valkyrja/application/kernel/contract'
import type {ComponentProviderContract} from '@valkyrja/application/provider/contract'
import {HttpContainerProvider} from './HttpContainerProvider'
import {HttpMiddlewareProvider} from './HttpMiddlewareProvider'
import {HttpEventProvider} from './HttpEventProvider'
import {HttpRouteProvider} from './HttpRouteProvider'

export class HttpComponentProvider implements ComponentProviderContract {

    getComponentProviders(app: ApplicationContract): ComponentProviderContract[] {
        return [
            new ContainerComponentProvider(),  // HTTP depends on Container
            new EventComponentProvider(),       // HTTP depends on Event
        ]
    }

    getContainerProviders(app: ApplicationContract): ServiceProviderContract[] {
        return [
            new HttpContainerProvider(),
            new HttpMiddlewareProvider(),
        ]
    }

    getEventProviders(app: ApplicationContract): ListenerProviderContract[] {
        return [
            new HttpEventProvider(),
        ]
    }

    getCliProviders(app: ApplicationContract): CliRouteProviderContract[] {
        return []
    }

    getHttpProviders(app: ApplicationContract): HttpRouteProviderContract[] {
        return [
            new HttpRouteProvider(),
        ]
    }
}
```

---

## ServiceProviderContract

Container bindings provider. `publishers()` returns a map of binding key string constant to publisher method reference.
The build tool reads the map from AST via the TypeScript compiler API, resolves each method reference, and reads that
method body directly — no annotation needed.

```typescript
// package: @valkyrja/container/provider/contract
import type {ContainerContract} from '@valkyrja/container/manager/contract'

/**
 * Defines what a container service provider must implement.
 *
 * publishers() returns a map of binding key to publisher method reference.
 * The map must be a simple object literal — no conditional logic permitted.
 * Each value must be a method reference on the same class.
 *
 * The build tool reads the publishers map from AST via the TypeScript
 * compiler API, resolves each method reference to its source location,
 * and reads that method body for cache generation.
 *
 * No annotation is needed on publisher methods — the method body is
 * read directly from AST.
 *
 * Note: TypeScript has no ::class equivalent — string constants are used
 * for all binding keys. See ContainerConstants files per component.
 *
 * @example
 * publishers(): Record<string, (c: ContainerContract) => void> {
 *     return {
 *         [UserRepositoryClass]: this.publishUserRepository,
 *     }
 * }
 *
 * publishUserRepository(c: ContainerContract): void {
 *     c.setSingleton(UserRepositoryClass, new UserRepository(c.make(DatabaseClass)))
 * }
 */
export interface ServiceProviderContract {
    /**
     * Return a map of string binding key to publisher static method reference.
     * Must return a simple object literal — no conditional logic permitted.
     * Each value must be a static method reference on the same class.
     */
    publishers(): Readonly<Record<string, (c: ContainerContract) => void>>
}
```

### UserServiceProvider Implementation

```typescript
import type {ContainerContract} from '@valkyrja/container/manager/contract'
import type {ServiceProviderContract} from '@valkyrja/container/provider/contract'
import {UserRepository} from '../repositories/UserRepository'
import {UserRepositoryClass} from '../repositories/contract/UserRepositoryConstants'
import {DatabaseClass} from '../services/contract/DatabaseConstants'

export class UserServiceProvider implements ServiceProviderContract {

    /**
     * Build tool reads this map from AST via the TypeScript compiler API,
     * resolves each method reference to its source location,
     * then reads each method body for cache generation.
     */
    publishers(): Readonly<Record<string, (c: ContainerContract) => void>> {
        return {
            [UserRepositoryClass]: this.publishUserRepository,
        }
    }

    publishUserRepository(c: ContainerContract): void {
        c.setSingleton(
            UserRepositoryClass,
            new UserRepository(c.make<DatabaseContract>(DatabaseClass))
        )
    }
}
```

---

## HttpRouteProviderContract

HTTP route provider. TypeScript has no reliable decorators — explicit route definitions only. Routes are complete data
structures — they cannot be expressed as a publisher-style map without losing the metadata the router requires.

```typescript
// package: @valkyrja/http/routing/provider/contract
import type {RouteContract} from '@valkyrja/http/routing/data/contract'

/**
 * Defines what an HTTP route provider must implement.
 */
export interface HttpRouteProviderContract {
    /**
     * Get a list of explicit HTTP route definitions.
     * Fully typed as RouteContract[] — compiler validates all returned objects.
     * Routes are complete data structures — they carry HTTP method, path pattern,
     * dynamic segment constraints, middleware chain, and handler together.
     * They cannot be expressed as a publisher-style map without losing
     * the metadata the router needs to build its dispatcher index.
     * Must return a simple array literal — no conditional logic permitted.
     *
     * Note: getControllerClasses() is intentionally absent. TypeScript has no
     * reliable decorator/annotation support so annotated class scanning is not
     * possible. Including the method would imply capability that does not exist.
     */
    getRoutes(): RouteContract[]
}
```

### UserHttpRouteProvider Implementation

```typescript
import type {RouteContract} from '@valkyrja/http/routing/data/contract'
import type {HttpRouteProviderContract} from '@valkyrja/http/routing/provider/contract'
import {HttpRoute} from '@valkyrja/http/routing/data'
import type {ContainerContract} from '@valkyrja/container/manager/contract'
import {UserControllerClass} from '../controllers/contract/UserControllerConstants'
import {OrderControllerClass} from '../controllers/contract/OrderControllerConstants'
import type {UserController} from '../controllers/UserController'
import type {OrderController} from '../controllers/OrderController'

export class UserHttpRouteProvider implements HttpRouteProviderContract {

    /**
     * Handler is a method reference on this same class.
     * Sindri reads handler method bodies from this file only.
     */
    getRoutes(): RouteContract[] {
        return [
            HttpRoute.get('/users', this.indexUsers.bind(this)),
            HttpRoute.post('/users', this.storeUser.bind(this)),
            HttpRoute.get('/orders', this.indexOrders.bind(this)),
        ]
    }

    /** Handler methods live on the same class — all imports self-contained. */
    indexUsers(c: ContainerContract, args: Record<string, unknown>): ResponseContract {
        return (c.getSingleton(UserControllerClass) as UserController).index(args)
    }

    storeUser(c: ContainerContract, args: Record<string, unknown>): ResponseContract {
        return (c.getSingleton(UserControllerClass) as UserController).store(args)
    }

    indexOrders(c: ContainerContract, args: Record<string, unknown>): ResponseContract {
        return (c.getSingleton(OrderControllerClass) as OrderController).index(args)
    }
}
```

---

## CliRouteProviderContract

```typescript
// package: @valkyrja/cli/routing/provider/contract
import type {RouteContract} from '@valkyrja/cli/routing/data/contract'

/**
 * Defines what a CLI route provider must implement.
 */
export interface CliRouteProviderContract {
    /**
     * Get a list of explicit CLI route definitions.
     * Must return a simple array literal — no conditional logic permitted.
     */
    getRoutes(): RouteContract[]
}
```

---

## ListenerProviderContract

```typescript
// package: @valkyrja/event/provider/contract
import type {ListenerContract} from '@valkyrja/event/data/contract'

/**
 * Defines what an event listener provider must implement.
 */
export interface ListenerProviderContract {
    /**
     * Get a list of explicit listener definitions.
     * Fully typed as ListenerContract[] — compiler validates all returned objects.
     * Listeners carry event type, priority, and handler together.
     * Cannot be expressed as a key/body map without losing
     * the metadata the event dispatcher requires.
     * Must return a simple array literal — no conditional logic permitted.
     *
     * Note: getListenerClasses() is intentionally absent. TypeScript has no
     * reliable decorator/annotation support so annotated class scanning is not
     * possible. Including the method would imply capability that does not exist.
     */
    getListeners(): ListenerContract[]
}
```

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic:

```typescript
// ✅ simple array of instances
return [new UserContainerProvider(), new OrderContainerProvider()]

// ✅ simple object literal with method reference
return {[UserRepositoryClass]: this.publishUserRepository}

// ✅ simple array of route objects
return [
    HttpRoute.get('/users', this.indexUsers.bind(this)),
    HttpRoute.post('/users', this.storeUser.bind(this)),
]

// ❌ conditional logic
if (condition) {
    return [...]
}

// ❌ variable accumulation
const routes: RouteContract[] = []
routes.push(...)
return routes

// ❌ method calls other than constructors/static factories
return [...this.getBaseRoutes(), ...this.getExtraRoutes()]
```

---

## Handler Method Pointer Convention

All handler methods must be **methods on the same class** as the provider or controller that defines the route or
listener. This is the same pattern used by `publishers()` in service providers.

**Why:** Sindri reads exactly one file per provider or controller. All imports for handler bodies are in that one file —
no cross-file import aggregation, no conflict detection, no registry needed.

```
✅ Method reference on the same class
✅ All type references imported in the same file

❌ Inline closures or lambdas in route/listener definitions
❌ References to types not imported in the current file
❌ Handler methods on a different class
```

---

## Design Note — Why Routes Cannot Use a Publisher-Style Map

An early consideration was expressing routes the same way as container bindings — a map of route identifier to handler
function, with the build tool reading function bodies directly. This was rejected because routes are multi-dimensional
data structures, not simple key→factory pairs.

A route carries: HTTP method, path pattern, dynamic segment constraints, regex compilation data, middleware chain,
name/alias, parameter defaults, host constraints, and scheme constraints — all in addition to the handler. The
`HttpRoute.get("/users/{id}", handler)` call is what populates all of these fields together. Decomposing this into a
key/function-body map would lose all metadata the router needs to build its dispatcher index and compile route regexes.

The same reasoning applies to listeners — they carry event type binding, priority, and stop-propagation behavior
alongside the handler. These cannot be expressed as a flat key/body map without losing the data the event dispatcher
requires.

Container bindings by contrast are simple key→factory pairs. This is why `publishers()` works as a map but `getRoutes()`
must return complete route objects.

## Note on TypeScript Decorators and Missing Methods

TypeScript decorators are currently at stage 3 of the TC39 proposal process and are considered experimental. Valkyrja's
TypeScript port does not rely on decorators for any core functionality to avoid coupling to an unstable language
feature.

As a direct consequence, `getControllerClasses()` and `getListenerClasses()` are intentionally absent from all
TypeScript provider contracts. Including them would imply that annotated class scanning works in TypeScript, which it
does not. A method that always returns an empty array or is never called adds noise and invites confusion.

If decorators stabilize and become part of the TypeScript standard, `getControllerClasses()` and `getListenerClasses()`
can be added to the contracts at that point as a non-breaking addition. Until then, TypeScript providers define routes
and listeners exclusively via `getRoutes()` and `getListeners()`.
