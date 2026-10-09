# TypeScript / Node.js Port — Implementation Notes

> Reference docs, in `convention/`: `THROWABLES.md`, `CONTAINER_BINDINGS.md`,
> `HANDLERS.md`, `DATA_CACHE.md`, `SINDRI.md`, `CONTRACTS.md`
> Port order: Container → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Module namespace:** `@valkyrja/`
- **`abstract class`** enforces contracts at compile time
- **Decorators plus explicit registration** — both supported, with explicit
  registration the portable form
- **No `::class` equivalent** — string constants for all binding keys
- **Constructed providers** in every provider list (`ComponentProviderContract[]`),
  walked directly at runtime
- **The handler on the route contract** (`getHandler()` / `withHandler()`)
  for typed closures
- **TypeScript compiler API** for build tool
- **Node.js worker model** — single bootstrap, routes in memory permanently
- **Result pattern** available as additive opt-in (`tryMake<T>` style) — not
  required
- Types erased at runtime — no `instanceof` checks on type-erased generics
- `getControllerClasses()` and `getListenerClasses()` **absent** — no reliable
  annotations

---

## 1. Throwables

**Reference:** `THROWABLES.md`

### Hierarchy — all branches extend Error

```typescript
// Throwable branch
export abstract class ValkyrjaThrowable extends Error {
}

export abstract class ComponentThrowable extends ValkyrjaThrowable {
}  // always present
export class ComponentSpecificThrowable extends ComponentThrowable {
}  // concrete

// RuntimeException branch
export abstract class ValkyrjaRuntimeException extends Error {
}

export abstract class ComponentRuntimeException extends ValkyrjaRuntimeException {
}

export class ComponentSpecificException extends ComponentRuntimeException {
}

// InvalidArgumentException branch
export abstract class ValkyrjaInvalidArgumentException extends Error {
}

export abstract class ComponentInvalidArgumentException extends ValkyrjaInvalidArgumentException {
}

export class ComponentSpecificInvalidArgumentException extends ComponentInvalidArgumentException {
}
```

All three branches extend `Error` — TypeScript has no distinct `RuntimeError` or
`InvalidArgumentError` built-ins.

### Rules

- `abstract class` prevents instantiation at compile time
- Every component ships both categoricals even if unused
- Naming: `ComponentName*`, shared subcomponents `ParentComponentSubComponent*`
- No typed throws on function signatures — TypeScript cannot express this

### Result pattern (additive opt-in)

```typescript
type Result<T, E extends Error> =
    | { success: true; value: T }
    | { success: false; error: E }

// available alongside standard throw/catch
function tryMake<T>(abstract: string): Result<T, ContainerException> {
}
```

---

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### String constants — required, no ::class equivalent

```typescript
// container-constants.ts
export const ContainerConstants = {
    CONTAINER: 'Valkyrja.Container.Manager.ContainerContract',
    ROUTER: 'Valkyrja.Http.Routing.Dispatcher.RouterContract',
    USER_REPOSITORY: 'App.Repository.UserRepositoryContract',
} as const
```

### Closure-based bindings

```typescript
container.bind(
    ContainerConstants.ROUTER,
    (c: ContainerContract) => new Router(c.getSingleton(ContainerConstants.DISPATCHER))
)

container.bindSingleton(
    ContainerConstants.ROUTER,
    (c: ContainerContract) => new Router(c.getSingleton(ContainerConstants.DISPATCHER))
)
```

---

## 3. Provider Contracts

**Reference:** `PROVIDER_CONTRACTS.md`, and `DATA_CACHE.md` in `convention/`

### ComponentProviderContract

```typescript
export interface ComponentProviderContract {
    // A list of constructed providers, the same shape every port holds
    // The framework walks them directly — no string lookup needed
    getComponentProviders(app: ApplicationContract): ComponentProviderContract[]

    getContainerProviders(app: ApplicationContract): ServiceProviderContract[]

    getEventProviders(app: ApplicationContract): ListenerProviderContract[]

    getCliProviders(app: ApplicationContract): CliRouteProviderContract[]

    getHttpProviders(app: ApplicationContract): HttpRouteProviderContract[]
}
```

### ServiceProviderContract

```typescript
export interface ServiceProviderContract {
    publishers(): Record<string, (c: ContainerContract) => void>
}
```

No annotation on publisher methods — build tool reads method bodies directly
from AST via TypeScript compiler API:

```typescript
publishers()
:
Record < string, (c: ContainerContract) => void > {
    return {
        [UserRepositoryClass]: this.publishUserRepository,
    }
}

// build tool reads this method body from AST
publishUserRepository(c
:
ContainerContract
):
void {
    c
    .setSingleton(UserRepositoryClass, new UserRepository(c.getSingleton(DatabaseClass)))
}
```

### HttpRouteProviderContract / CliRouteProviderContract

```typescript
export interface HttpRouteProviderContract {
    // getControllerClasses() intentionally absent — no reliable annotations in TypeScript
    getRoutes(): RouteContract[]
}
```

### ListenerProviderContract

```typescript
export interface ListenerProviderContract {
    // getListenerClasses() intentionally absent — no reliable annotations in TypeScript
    getListeners(): ListenerContract[]
}
```

All provider methods must return simple array/object literals — no conditional
logic.

---

## 4. Constructed Providers — Works Without Cache

A provider list holds constructed providers, not strings, so the framework walks
them and calls their methods directly at runtime:

```typescript
// framework bootstrap — walk the tree, no cache, no string lookup
for (const provider of component.getHttpProviders(app)) {
    for (const route of provider.getRoutes()) {
        router.register(route)
    }
}
```

`sindri` reads the same construction expressions through the compiler API, so the
cached and uncached paths register from one declaration.

This means TypeScript works without cache exactly as the other ports do.

---

## 5. Handler Contracts — Named Types

**Reference:** `HANDLERS.md`

### Three named types — compiler enforced

A route handler takes the container and the **route**. Only a listener takes a
map, because a listener has no route. See
[`HANDLERS.md`](../../convention/HANDLERS.md).

```typescript
// HTTP route
(container: ContainerContract, route: RouteContract) => ResponseContract

// CLI route
(container: ContainerContract, route: RouteContract) => OutputContract

// Event listener
(container: ContainerContract, args: Record<string, unknown>) => unknown
```

### The handler lives on the route contract

There is no handler contract per concern. The route contract declares
`getHandler()` and `withHandler()`, and `withHandler` returns a copy rather than
modifying the route, per
[`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md).

### Usage

```typescript
// HTTP — the route is the second parameter, and withHandler returns a copy
route = route.withHandler((container, route) =>
    container.getSingleton<UserController>(UserControllerClass).show(route)
)

// CLI
command = command.withHandler((container, route) =>
    container.getSingleton<SendEmailCommand>(SendEmailCommandClass).run(route)
)

// Listener — a listener has no route, so it takes a map
listener = listener.withHandler((container, args) =>
    container.getSingleton<UserCreatedListener>(UserCreatedListenerClass).handle(args['user_id'] as string)
)
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from
container if needed.

---

## 6. Decorators and Explicit Registration

The port supports both. A provider declares routes explicitly through
`getRoutes()`, and it declares the controller classes to scan through
`getControllerClasses()`, whose routing decorators `sindri` reads statically and
the runtime collector reads on the debug path.

Explicit registration remains the portable form, because a port whose language
has no decorator supports only that one. See
[`DECORATORS.md`](DECORATORS.md) for the decorator rules, including why a
decorator argument never names a class directly.

---

## 7. Build Tool — `@valkyrjaio/sindri`

**Reference:** `SINDRI.md`

- Separate npm package: `@valkyrjaio/sindri`
- Dev dependency only — never in production
- Uses TypeScript compiler API (`ts.createProgram`) — full AST with type
  information
- Type checker resolves all type references to FQN via module resolution
- `tsconfig.json` module resolution used to locate source files
- Must ship `.ts` source files (not just `.d.ts`) and no compiled `.js` —
  consumers compile the framework together with their own app

### Build tool flow

```
AppConfig → component providers
        ↓
ts.createProgram → full AST with type information
        ↓
walk register() / getRoutes() / publishers() method bodies
        ↓
extract handler arrow functions + parameter data
        ↓
type checker → resolve all types to fully qualified module paths
        ↓
ProcessorContract for regex compilation
        ↓
generate AppHttpRoutingData.ts, AppContainerData.ts etc.
        ↓
tsc compiles with generated files
```

---

## 8. Deployment

### Worker (Node.js)

- Single bootstrap — routes in memory permanently
- Cache optional but supported
- Primary deployment model

### CGI / Lambda

- Cache required for production cold start optimization
- `sindri` generates cache data files pre-`tsc`
- Single compile pass — no two-pass needed

---

## Priority Order

1. Container component
2. String constants per component
3. Throwable hierarchy — abstract classes, all extend Error, three branches
4. Result pattern as additive opt-in
5. Closure-based bindings
6. Provider contracts — ComponentProvider, ServiceProvider, RouteProvider,
   ListenerProvider
7. The handler on the route contract — getHandler / withHandler
8. Route and listener data classes
9. `@valkyrjaio/sindri` npm package — TypeScript compiler API implementation
10. AppContainerData, AppHttpRoutingData, AppCliRoutingData, AppEventData
    generation
