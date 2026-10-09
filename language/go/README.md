# Go Port — Implementation Notes

> Reference docs, in `convention/`: `THROWABLES.md`, `CONTAINER_BINDINGS.md`,
> `HANDLERS.md`, `DATA_CACHE.md`, `SINDRI.md`, `CONTRACTS.md`
> Port order: Container → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Module path:** `github.com/valkyrjaio/valkyrja-go/vN`; a component package sits directly under it
- **No annotations** — explicit registration only throughout
- **No `::class` equivalent** — string constants for all binding keys
- **Interfaces** for contracts, **structs** for implementations
- **Unexported embedded fields** for "abstract" enforcement
- **Named function types** for typed handler closures
- **`go/analysis` + `go/ast`** for build tool
- **`go generate`** triggers build tool
- **`(T, error)` return pattern** — idiomatic Go, used throughout
- **`errors.As` / `errors.Is`** for typed error checking
- Go module system downloads full source — build tool has framework source
  available

---

## 1. Throwables / Errors

**Reference:** `THROWABLES.md`

### Three branches maintained for cross-port parity

```go
// Throwable — unexported interface
type valkyrjaThrowable interface {
error
isValkyrjaThrowable()
}

// RuntimeError — exported struct with unexported field
type ValkyrjaRuntimeError struct {
valkyrjaThrowable // unexported embedded — prevents external instantiation
message string
}
func (e *ValkyrjaRuntimeError) Error() string { return e.message }

// InvalidArgumentError — exported struct with unexported field
type ValkyrjaInvalidArgumentError struct {
valkyrjaThrowable
message string
}
func (e *ValkyrjaInvalidArgumentError) Error() string { return e.message }
```

### Component categoricals — always present, unexported interface

```go
// always present per component, unexported
type containerRuntimeError interface {
ValkyrjaRuntimeError
isContainerRuntimeError()
}

// concrete errors — exported, implement the unexported interface
type ContainerNotFoundError struct {
ValkyrjaRuntimeError
}
```

### Naming convention

- Shared subcomponents: `HttpRoutingRuntimeError`,
  `CliRoutingRuntimeError`
- Unique subcomponents: `RequestRuntimeError`, `ResponseRuntimeError`
- Unexported = abstract equivalent (component categoricals)
- Exported = concrete (specific errors only)

### Error checking

```go
var target *ContainerNotFoundError
if errors.As(err, &target) {
// handle specifically
}
```

### Result pattern — idiomatic

```go
// (T, error) return is idiomatic Go — used throughout
user, err := container.Get(UserRepositoryClass)
if err != nil {
return nil, err
}
```

---

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### String constants — required, no ::class equivalent

Every class, interface, and contract needs a string constant:

```go
// container_constants.go
package container

const (
	ContainerClass      = "valkyrja.container.manager.ContainerContract"
	RouterClass         = "valkyrja.http.routing.dispatcher.RouterContract"
	UserRepositoryClass = "app.repository.UserRepositoryContract"
)
```

### Closure-based bindings

```go
container.Bind(
RouterClass,
func (c ContainerContract) any {
return NewRouter(c.GetSingleton(EventDispatcherClass).(EventDispatcherContract))
},
)

container.BindSingleton(
RouterClass,
func(c ContainerContract) any {
return NewRouter(c.GetSingleton(EventDispatcherClass).(EventDispatcherContract))
},
)
```

---

## 3. Provider Contracts

**Reference:** `PROVIDER_CONTRACTS.md`, and `DATA_CACHE.md` in `convention/`

### ComponentProviderContract

```go
type ComponentProviderContract interface {
GetContainerProviders(app ApplicationContract) []ServiceProviderContract
GetEventProviders(app ApplicationContract) []ListenerProviderContract
GetCliProviders(app ApplicationContract) []CliRouteProviderContract
GetHttpProviders(app ApplicationContract) []HttpRouteProviderContract
}
```

### ServiceProviderContract

```go
type ServiceProviderContract interface {
Publishers() map[string]func (ContainerContract)
}
```

Publisher functions can be **struct methods OR package-level functions** — build
tool handles both:

```go
// struct method
func (p *UserServiceProvider) Publishers() map[string]func (ContainerContract) {
return map[string]func (ContainerContract){
UserRepositoryClass: p.PublishUserRepository,
// or package-level:
// UserRepositoryClass: PublishUserRepository,
}
}

func (p *UserServiceProvider) PublishUserRepository(c ContainerContract) {
c.SetSingleton(UserRepositoryClass, NewUserRepository(c.GetSingleton(DatabaseClass)))
}
```

### HttpRouteProviderContract / CliRouteProviderContract

```go
type HttpRouteProviderContract interface {
// GetControllerClasses intentionally absent — Go has no annotations
GetRoutes() []RouteContract
}
```

### ListenerProviderContract

```go
type ListenerProviderContract interface {
// GetListenerClasses intentionally absent — Go has no annotations
GetListeners() []ListenerContract
}
```

All provider methods must return simple slice/map literals — no conditional
logic.

---

## 4. Handler Signatures

**Reference:** `HANDLERS.md`

### The handler types

A route handler takes the container and the **route**. Only a listener takes a
map, because a listener has no route. See
[`HANDLERS.md`](../../convention/HANDLERS.md).

```go
// HTTP route
func (container ContainerContract, route RouteContract) ResponseContract

// CLI route
func (container ContainerContract, route RouteContract) OutputContract

// Event listener
func (container ContainerContract, arguments map[string]any) any
```

### The handler lives on the route contract

There is no handler contract per concern. The route contract declares
`GetHandler` and `WithHandler`, and `WithHandler` returns a copy rather than
modifying the route, per
[`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md).

### Usage

```go
// HTTP — the route is the second parameter, and WithHandler returns a copy
route = route.WithHandler(func (c ContainerContract, r RouteContract) ResponseContract {
return c.GetSingleton(UserControllerClass).(*UserController).Show(r)
})

// CLI
command = command.WithHandler(func (c ContainerContract, r RouteContract) OutputContract {
return c.GetSingleton(SendEmailCommandClass).(*SendEmailCommand).Run(r)
})

// Listener — a listener has no route, so it takes a map
listener = listener.WithHandler(func (c ContainerContract, args map[string]any) any {
return c.GetSingleton(UserCreatedListenerClass).(*UserCreatedListener).Handle(args["user_id"])
})
```

The **route is a parameter**, and it is also set on the container before the
handler runs, so it is reachable both ways. `ServerRequestContract` is
container-only — it is never a handler parameter.

---

## 5. No Annotations — Explicit Registration Only

Go has no annotations. There is no `GetControllerClasses()` — routes and
listeners are always registered explicitly via `GetRoutes()` and
`GetListeners()`.

The build tool (`go generate` + `go/analysis`) scans `GetRoutes()` and
`GetListeners()` method bodies for handler function literals and extracts them
from the AST.

---

## 6. Build Tool — `sindri` Go

**Reference:** `SINDRI.md`

- Separate Go module: `github.com/valkyrjaio/sindri-go/vN`
- Triggered via `go generate`
- Uses `go/packages`, `go/ast`, `go/analysis` — all standard library
- Go module system downloads full source — no special source shipping policy
  needed
- Package paths from the application config class map directly to directory
  paths
- Generated files use `text/template` + `go/format` for clean output
- `go generate` → `go build` — single effective compile pass

### Build tool flow

```
AppConfig → component providers
        ↓
go/packages.Load() → source files
        ↓
go/ast → walk GetRoutes() / Publishers() method bodies
        ↓
Extract handler func literals + parameter data
        ↓
Resolve imports to fully qualified package paths
        ↓
Run ProcessorContract for regex compilation
        ↓
text/template → generate AppHttpRoutingData, AppContainerData etc.
        ↓
go/format → format generated source
        ↓
go build compiles with generated files
```

---

## 7. Concurrency

- Goroutines handle concurrency natively — no worker mode complexity
- Single binary compiled — routes registered once at startup, in memory
  permanently
- Cache data files still supported for CGI/lambda cold start optimization
- Go binary startup is near-instant — cache less critical here than other
  languages

---

## Priority Order

1. Container component
2. String constants per component
3. Throwable / error hierarchy — three branches, unexported interfaces as
   abstract
4. Closure-based bindings
5. Provider contracts — ComponentProvider, ServiceProvider, RouteProvider,
   ListenerProvider
6. The handler on the route contract — GetHandler / WithHandler
7. Route and listener data classes
8. go generate + go/analysis build tool
9. AppContainerData, AppHttpRoutingData, AppCliRoutingData, AppEventData
   generation
