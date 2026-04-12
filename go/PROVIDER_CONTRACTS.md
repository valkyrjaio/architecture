# Valkyrja Provider Contracts — Go

## Overview

Go provider contracts differ from PHP/Java in several important ways:

- No annotations — build tool reads method bodies directly from AST in all cases
- No `::class` / `.class` — string constants used for all class references
- Publisher methods can be struct methods OR package-level functions — build
  tool handles both
- No abstract classes — interfaces enforce the contract, unexported types
  enforce instantiation restrictions
- All methods return simple slice/map literals — no conditional logic
- `GetControllerClasses()` and `GetListenerClasses()` are **absent** — Go has no
  annotations, so annotated class scanning is not possible. Including these
  methods would imply capability that does not exist.

### Go Works Without Cache

Go interface methods are called directly on provider structs at runtime — no
reflection, no string lookup, no registry needed. The framework traverses the
provider tree by direct method calls:

```go
// framework bootstrap — direct method calls, no cache needed
for _, componentProvider := range componentProviders {
for _, HttpProvider := range componentProvider.GetHttpProviders(app) {
for _, route := range HttpProvider.GetRoutes() {  // direct method call ✅
router.Register(route)
}
}
for _, ServiceProvider := range componentProvider.GetContainerProviders(app) {
for key, publisher := range ServiceProvider.Publishers() {
container.Bind(key, publisher) // direct function call ✅
}
}
}
```

Cache is a cold-start optimization for CGI and lambda deployments. Go's compiled
binary and fast startup mean cache is less critical here than in other
languages, but it is fully supported via the valkyrja-build tool when needed.

---

## ComponentProviderContract

Top-level aggregator. Returns slices of sub-provider interface values. Build
tool reads return values directly from AST — must be simple slice literals with
no conditional logic.

```go
// package: io/valkyrja/application/provider/contract
package contract

import appContract "io/valkyrja/application/kernel/contract"

// ComponentProviderContract defines what a component provider must implement.
// All methods must return simple slice literals — no conditional logic permitted.
// The build tool reads these return values directly from AST.
type ComponentProviderContract interface {
	// GetContainerProviders returns the component's container service providers.
	GetContainerProviders(app appContract.ApplicationContract) []ServiceProviderContract

	// GetEventProviders returns the component's event listener providers.
	GetEventProviders(app appContract.ApplicationContract) []ListenerProviderContract

	// GetCliProviders returns the component's CLI route providers.
	GetCliProviders(app appContract.ApplicationContract) []CliRouteProviderContract

	// GetHttpProviders returns the component's HTTP route providers.
	GetHttpProviders(app appContract.ApplicationContract) []HttpRouteProviderContract
}
```

### HttpComponentProvider Implementation

```go
package provider

import (
	appContract "io/valkyrja/application/kernel/contract"
	cliContract "io/valkyrja/cli/routing/provider/contract"
	ctnContract "io/valkyrja/container/provider/contract"
	evtContract "io/valkyrja/event/provider/contract"
	httpContract "io/valkyrja/http/routing/provider/contract"
)

type HttpComponentProvider struct{}

func (p *HttpComponentProvider) GetContainerProviders(
	app appContract.ApplicationContract,
) []ctnContract.ServiceProviderContract {
	return []ctnContract.ServiceProviderContract{
		&HttpContainerProvider{},
		&HttpMiddlewareProvider{},
	}
}

func (p *HttpComponentProvider) GetEventProviders(
	app appContract.ApplicationContract,
) []evtContract.ListenerProviderContract {
	return []evtContract.ListenerProviderContract{
		&HttpEventProvider{},
	}
}

func (p *HttpComponentProvider) GetCliProviders(
	app appContract.ApplicationContract,
) []cliContract.CliRouteProviderContract {
	return []cliContract.CliRouteProviderContract{}
}

func (p *HttpComponentProvider) GetHttpProviders(
	app appContract.ApplicationContract,
) []httpContract.HttpRouteProviderContract {
	return []httpContract.HttpRouteProviderContract{
		&HttpRouteProvider{},
	}
}
```

---

## ServiceProviderContract

Container bindings provider. `Publishers()` returns a map of binding key string
to publisher function reference. The build tool reads the map from AST, resolves
each function reference, and reads that function body directly. Publisher
functions can be struct methods OR package-level functions — both are valid.

```go
// package: io/valkyrja/container/provider/contract
package contract

import ctnContract "io/valkyrja/container/manager/contract"

// ServiceProviderContract defines what a container service provider must implement.
//
// Publishers() returns a map of binding key to publisher function reference.
// The map must be a simple map literal — no conditional logic permitted.
// Each value must be a method reference on the same struct OR a package-level
// function reference. The build tool reads the map from AST, resolves each
// function reference, and reads that function body for cache generation.
//
// No annotations are needed — method bodies are read directly from AST.
//
// Example:
//   func (p *UserServiceProvider) Publishers() map[string]func(ctnContract.ContainerContract) {
//       return map[string]func(ctnContract.ContainerContract){
//           repoContract.UserRepositoryClass: p.PublishUserRepository,
//           // or a package-level function:
//           // repoContract.UserRepositoryClass: PublishUserRepository,
//       }
//   }
//
//   func (p *UserServiceProvider) PublishUserRepository(c ctnContract.ContainerContract) {
//       c.SetSingleton(repoContract.UserRepositoryClass, repositories.NewUserRepository(...))
//   }
type ServiceProviderContract interface {
	// Publishers returns a map of binding key to publisher function reference.
	// Must be a simple map literal — no conditional logic permitted.
	Publishers() map[string]func(ctnContract.ContainerContract)
}
```

### UserServiceProvider Implementation

```go
package provider

import (
	ctnContract "io/valkyrja/container/manager/contract"
	"app/repositories"
	repoContract "app/repositories/contract"
	svcContract "app/services/contract"
)

type UserServiceProvider struct{}

// Publishers returns the map of binding key to publisher function reference.
// Build tool reads this map from AST, resolves each function reference,
// then reads each function body for cache generation.
func (p *UserServiceProvider) Publishers() map[string]func(ctnContract.ContainerContract) {
	return map[string]func(ctnContract.ContainerContract){
		repoContract.UserRepositoryClass: p.PublishUserRepository,
	}
}

// PublishUserRepository publishes the UserRepository binding.
// Build tool reads this method body from AST for cache generation.
// No annotation needed — discovered via Publishers() map.
func (p *UserServiceProvider) PublishUserRepository(c ctnContract.ContainerContract) {
	c.SetSingleton(
		repoContract.UserRepositoryClass,
		repositories.NewUserRepository(
			c.Make(svcContract.DatabaseClass).(svcContract.DatabaseContract),
		),
	)
}

// Package-level function alternative — also valid and readable by build tool.
// Useful for complex logic that doesn't need receiver state.
func PublishUserRepository(c ctnContract.ContainerContract) {
	c.SetSingleton(
		repoContract.UserRepositoryClass,
		repositories.NewUserRepository(
			c.Make(svcContract.DatabaseClass).(svcContract.DatabaseContract),
		),
	)
}
```

---

## HttpRouteProviderContract

HTTP route provider. Go has no annotations — explicit route definitions only.
`GetControllerClasses()` returns string constants (no `::class` equivalent).
Routes are complete data structures carrying method, path, constraints,
middleware, and handler together.

```go
// package: io/valkyrja/http/routing/provider/contract
package contract

import dataContract "io/valkyrja/http/routing/data/contract"

// HttpRouteProviderContract defines what an HTTP route provider must implement.
type HttpRouteProviderContract interface {
	// GetControllerClasses returns a list of controller class string constants.
	// Go has no ::class equivalent — string constants from the constants file are used.
	// Returns empty slice if using explicit routes only (most common in Go).
	// Must be a simple slice literal — no conditional logic permitted.
	GetControllerClasses() []string

	// GetRoutes returns a list of explicit route definitions.
	// Routes are complete data structures — they carry HTTP method, path pattern,
	// dynamic segment constraints, middleware chain, and handler together.
	// They cannot be expressed as a publisher-style map without losing
	// the metadata the router needs to build its dispatcher index.
	// Must be a simple slice literal — no conditional logic permitted.
	GetRoutes() []dataContract.RouteContract
}
```

### UserHttpRouteProvider Implementation

```go
package provider

import (
	ctnContract "io/valkyrja/container/manager/contract"
	"io/valkyrja/http/routing/data"
	dataContract "io/valkyrja/http/routing/data/contract"
	"app/controllers"
)

type UserHttpRouteProvider struct{}

// GetControllerClasses returns string constants — Go has no ::class equivalent.
// Returns empty slice since Go has no annotations to scan.
func (p *UserHttpRouteProvider) GetControllerClasses() []string {
	return []string{}
}

func (p *UserHttpRouteProvider) GetRoutes() []dataContract.RouteContract {
	return []dataContract.RouteContract{
		data.Get("/users", func(c ctnContract.ContainerContract, args []any) any {
			return c.GetSingleton(controllers.UserControllerClass).(*controllers.UserController).Index(args[0])
		}),
		data.Post("/users", func(c ctnContract.ContainerContract, args []any) any {
			return c.GetSingleton(controllers.UserControllerClass).(*controllers.UserController).Store(args[0])
		}),
		data.Get("/orders", func(c ctnContract.ContainerContract, args []any) any {
			return c.GetSingleton(controllers.OrderControllerClass).(*controllers.OrderController).Index(args[0])
		}),
	}
}
```

---

## CliRouteProviderContract

```go
// package: io/valkyrja/cli/routing/provider/contract
package contract

import dataContract "io/valkyrja/cli/routing/data/contract"

// CliRouteProviderContract defines what a CLI route provider must implement.
type CliRouteProviderContract interface {
	// GetControllerClasses returns a list of controller class string constants.
	// Returns empty slice (Go has no annotations to scan).
	GetControllerClasses() []string

	// GetRoutes returns a list of explicit CLI route definitions.
	// Must be a simple slice literal — no conditional logic permitted.
	GetRoutes() []dataContract.RouteContract
}
```

---

## ListenerProviderContract

```go
// package: io/valkyrja/event/provider/contract
package contract

import dataContract "io/valkyrja/event/data/contract"

// ListenerProviderContract defines what an event listener provider must implement.
type ListenerProviderContract interface {
	// GetListenerClasses returns a list of listener class string constants.
	// Returns empty slice (Go has no annotations to scan).
	GetListenerClasses() []string

	// GetListeners returns a list of explicit listener definitions.
	// Listeners are complete data structures — they carry event type, priority,
	// and handler together. Cannot be expressed as a key/body map without
	// losing the metadata the event dispatcher requires.
	// Must be a simple slice literal — no conditional logic permitted.
	GetListeners() []dataContract.ListenerContract
}
```

---

## Build Tool Contract

Any method or function the build tool reads must return a single flat literal
with no logic:

```go
// ✅ simple slice of route objects
return []dataContract.RouteContract{
data.Get("/users", func (c ContainerContract, args []any) any { ... }),
}

// ✅ simple map with method reference
return map[string]func (ContainerContract){
UserRepositoryClass: p.PublishUserRepository,
}

// ✅ simple map with package-level function reference
return map[string]func (ContainerContract){
UserRepositoryClass: PublishUserRepository,
}

// ❌ conditional logic
if condition {
return []dataContract.RouteContract{...}
}

// ❌ variable accumulation
routes := []dataContract.RouteContract{}
routes = append(routes, ...)
return routes
```

---

## Design Note — Why Routes Cannot Use a Publisher-Style Map

An early consideration was expressing routes the same way as container
bindings — a map of route identifier to handler function, with the build tool
reading function bodies directly, eliminating the need for `GetRoutes()`
entirely. This was rejected because routes are multi-dimensional data
structures, not simple key→factory pairs.

A route carries: HTTP method, path pattern, dynamic segment constraints, regex
compilation data, middleware chain, name/alias, parameter defaults, host
constraints, and scheme constraints — all in addition to the handler. The
`data.Get("/users/{id}", handler)` call is what populates all of these fields
together. Decomposing this into a key/function-body map would lose all metadata
the router needs to build its dispatcher index and compile route regexes.

The same reasoning applies to listeners — they carry event type binding,
priority, and stop-propagation behavior alongside the handler. These cannot be
expressed as a flat key/body map without losing the data the event dispatcher
requires.

Container bindings by contrast are simple key→factory pairs. The binding key IS
the complete identity. This is why `Publishers()` works as a map but
`GetRoutes()` must return complete route objects.
