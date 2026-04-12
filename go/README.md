# Go Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `DISPATCH.md`,
`DATA_CACHE.md`, `BUILD_TOOL.md`, `CONTRACTS_GO.md`
> Port order: Container → Dispatch → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Package namespace:** `io/valkyrja`
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

// RuntimeException — exported struct with unexported field
type ValkyrjaRuntimeException struct {
valkyrjaThrowable // unexported embedded — prevents external instantiation
message string
}
func (e *ValkyrjaRuntimeException) Error() string { return e.message }

// InvalidArgumentException — exported struct with unexported field
type ValkyrjaInvalidArgumentException struct {
valkyrjaThrowable
message string
}
func (e *ValkyrjaInvalidArgumentException) Error() string { return e.message }
```

### Component categoricals — always present, unexported interface

```go
// always present per component, unexported
type containerRuntimeError interface {
ValkyrjaRuntimeException
isContainerRuntimeError()
}

// concrete errors — exported, implement the unexported interface
type ContainerNotFoundException struct {
ValkyrjaRuntimeException
}
```

### Naming convention

- Shared subcomponents: `HttpRoutingRuntimeException`,
  `CliRoutingRuntimeException`
- Unique subcomponents: `RequestRuntimeException`, `ResponseRuntimeException`
- Unexported = abstract equivalent (component categoricals)
- Exported = concrete (specific errors only)

### Error checking

```go
var target *ContainerNotFoundException
if errors.As(err, &target) {
// handle specifically
}
```

### Result pattern — idiomatic

```go
// (T, error) return is idiomatic Go — used throughout
user, err := container.Make(UserRepositoryClass)
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
	ContainerClass      = "io.valkyrja.container.ContainerContract"
	RouterClass         = "io.valkyrja.http.routing.RouterContract"
	UserRepositoryClass = "io.valkyrja.app.repositories.UserRepositoryContract"
)
```

### Closure-based bindings

```go
container.Bind(
RouterClass,
func (c ContainerContract) any {
return NewRouter(c.Make(DispatcherClass).(DispatcherContract))
},
)

container.Singleton(
RouterClass,
func(c ContainerContract) any {
return NewRouter(c.Make(DispatcherClass).(DispatcherContract))
},
)
```

---

## 3. Provider Contracts

**Reference:** `CONTRACTS_GO.md`, `DATA_CACHE.md`

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
c.SetSingleton(UserRepositoryClass, NewUserRepository(c.Make(DatabaseClass)))
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

## 4. Handler Contracts — Typed Function Types

**Reference:** `DISPATCH.md`

### Three named function types

```go
// HTTP
type HttpHandlerFunc func (container ContainerContract, arguments map[string]any) ResponseContract

// CLI
type CliHandlerFunc func (container ContainerContract, arguments map[string]any) OutputContract

// Event listener
type ListenerHandlerFunc func (container ContainerContract, arguments map[string]any) any
```

### Handler contracts per concern

```go
type HttpHandlerContract interface {
GetHandler() HttpHandlerFunc
SetHandler(HttpHandlerFunc) HttpHandlerContract
}

type CliHandlerContract interface {
GetHandler() CliHandlerFunc
SetHandler(CliHandlerFunc) CliHandlerContract
}

type ListenerHandlerContract interface {
GetHandler() ListenerHandlerFunc
SetHandler(ListenerHandlerFunc) ListenerHandlerContract
}
```

### Usage

```go
// HTTP
route.SetHandler(func (c ContainerContract, args map[string]any) ResponseContract {
return c.GetSingleton(UserControllerClass).(*UserController).Show(args["id"])
})

// CLI
command.SetHandler(func (c ContainerContract, args map[string]any) OutputContract {
return c.GetSingleton(SendEmailCommandClass).(*SendEmailCommand).Run(args)
})

// Listener
listener.SetHandler(func (c ContainerContract, args map[string]any) any {
return c.GetSingleton(UserCreatedListenerClass).(*UserCreatedListener).Handle(args["user_id"])
})
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from
container if needed.

---

## 5. No Annotations — Explicit Registration Only

Go has no annotations. There is no `GetControllerClasses()` — routes and
listeners are always registered explicitly via `GetRoutes()` and
`GetListeners()`.

The build tool (`go generate` + `go/analysis`) scans `GetRoutes()` and
`GetListeners()` method bodies for handler function literals and extracts them
from the AST.

---

## 6. Build Tool — valkyrja-build Go

**Reference:** `BUILD_TOOL.md`

- Separate Go module: `io/valkyrja/build`
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
6. Named handler function types — HttpHandlerFunc, CliHandlerFunc,
   ListenerHandlerFunc
7. Handler contracts per concern
8. Route and listener data classes
9. Dispatch component
10. go generate + go/analysis build tool
11. AppContainerData, AppHttpRoutingData, AppCliRoutingData, AppEventData
    generation
