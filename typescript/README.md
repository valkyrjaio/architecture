# TypeScript / Node.js Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `DISPATCH.md`,
`DATA_CACHE.md`, `BUILD_TOOL.md`, `CONTRACTS_TYPESCRIPT.md`
> Port order: Container → Dispatch → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Module namespace:** `@valkyrja/`
- **`abstract class`** enforces contracts at compile time
- **No reliable decorators** — explicit registration only, no annotated class
  scanning
- **No `::class` equivalent** — string constants for all binding keys
- **Constructor references** (`Array<new () => Contract>`) allow direct
  instantiation at runtime
- **Named types** (`HttpHandlerFunc`, `CliHandlerFunc`, `ListenerHandlerFunc`)
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
    CONTAINER: 'io.valkyrja.container.ContainerContract',
    ROUTER: 'io.valkyrja.http.routing.RouterContract',
    USER_REPOSITORY: 'io.valkyrja.app.repositories.UserRepositoryContract',
} as const
```

### Closure-based bindings

```typescript
container.bind(
    ContainerConstants.ROUTER,
    (c: ContainerContract) => new Router(c.make(ContainerConstants.DISPATCHER))
)

container.singleton(
    ContainerConstants.ROUTER,
    (c: ContainerContract) => new Router(c.make(ContainerConstants.DISPATCHER))
)
```

---

## 3. Provider Contracts

**Reference:** `CONTRACTS_TYPESCRIPT.md`, `DATA_CACHE.md`

### ComponentProviderContract

```typescript
export interface ComponentProviderContract {
    // Array<new () => T> is TypeScript's equivalent of PHP's ::class list
    // Allows direct instantiation at runtime — no string lookup needed
    getContainerProviders(app: ApplicationContract): Array<new () => ServiceProviderContract>

    getEventProviders(app: ApplicationContract): Array<new () => ListenerProviderContract>

    getCliProviders(app: ApplicationContract): Array<new () => CliRouteProviderContract>

    getHttpProviders(app: ApplicationContract): Array<new () => HttpRouteProviderContract>
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
    .setSingleton(UserRepositoryClass, new UserRepository(c.make(DatabaseClass)))
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

## 4. Constructor References — Works Without Cache

The `Array<new () => Contract>` return type is the key insight. The framework
receives actual class constructors — not strings — so it can instantiate
providers and call methods directly at runtime:

```typescript
// framework bootstrap — direct instantiation, no cache, no string lookup
for (const ProviderClass of component.getHttpProviders(app)) {
    const provider = new ProviderClass()
    for (const route of provider.getRoutes()) {
        router.register(route)
    }
}
```

This means TypeScript works without cache exactly like PHP and Python.

---

## 5. Handler Contracts — Named Types

**Reference:** `DISPATCH.md`

### Three named types — compiler enforced

```typescript
type HttpHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => ResponseContract

type CliHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => OutputContract

type ListenerHandlerFunc = (
    container: ContainerContract,
    arguments: Record<string, unknown>
) => unknown
```

### Handler contracts per concern

```typescript
interface HttpHandlerContract {
    getHandler(): HttpHandlerFunc

    setHandler(handler: HttpHandlerFunc): this
}

interface CliHandlerContract {
    getHandler(): CliHandlerFunc

    setHandler(handler: CliHandlerFunc): this
}

interface ListenerHandlerContract {
    getHandler(): ListenerHandlerFunc

    setHandler(handler: ListenerHandlerFunc): this
}
```

### Usage

```typescript
// HTTP
route.setHandler((container, args) =>
    container.getSingleton<UserController>(UserControllerClass).show(args['id'] as string)
)

// CLI
command.setHandler((container, args) =>
    container.getSingleton<SendEmailCommand>(SendEmailCommandClass).run(args)
)

// Listener
listener.setHandler((container, args) =>
    container.getSingleton<UserCreatedListener>(UserCreatedListenerClass).handle(args['user_id'] as string)
)
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from
container if needed.

---

## 6. No Annotations — Explicit Registration Only

TypeScript decorators are experimental (stage 3). Valkyrja's TypeScript port
does not rely on them. All route and listener registration is explicit via
`getRoutes()` and `getListeners()`.

If decorators stabilize, `getControllerClasses()` and `getListenerClasses()` can
be added as non-breaking additions.

---

## 7. Build Tool — @valkyrja/build

**Reference:** `BUILD_TOOL.md`

- Separate npm package: `@valkyrja/build`
- Dev dependency only — never in production
- Uses TypeScript compiler API (`ts.createProgram`) — full AST with type
  information
- Type checker resolves all type references to FQN via module resolution
- `tsconfig.json` module resolution used to locate source files
- Must ship `.ts` source files (not just `.d.ts`) alongside compiled `.js`

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
- `valkyrja-build` generates cache data files pre-`tsc`
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
7. Named handler types — HttpHandlerFunc, CliHandlerFunc, ListenerHandlerFunc
8. Handler contracts per concern
9. Route and listener data classes
10. Dispatch component
11. @valkyrja/build npm package — TypeScript compiler API implementation
12. AppContainerData, AppHttpRoutingData, AppCliRoutingData, AppEventData
    generation
