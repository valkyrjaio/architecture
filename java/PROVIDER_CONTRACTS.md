# Valkyrja Provider Contracts — Java

## Overview

Java provider contracts differ from PHP in several important ways:

- No annotations needed on `ComponentProviderContract` methods — the build tool reads method bodies directly from AST
- `new X()` used in list literals — `ObjectCreationExpr` nodes carry the class name directly; no reflection needed
- `List.of(...)` for immutable list returns
- `Map.of(...)` for immutable map returns
- All Valkyrja exceptions extend `RuntimeException` (unchecked)
- Publisher methods are static — build tool reads method bodies via `ClassName::methodName` references
- Instance methods throughout — the framework instantiates providers and calls methods directly

### Java Works Without Cache

Provider methods return instances directly. The framework traverses the provider tree by direct method calls — no
reflection, no string lookup, no registry needed:

```java
// framework bootstrap — direct method calls, no cache needed
for (ComponentProviderContract component : componentProviders) {
    for (ServiceProviderContract provider : component.getContainerProviders(app)) {
        for (var entry : provider.publishers().entrySet()) {
            container.bind(entry.getKey(), entry.getValue()); // direct call ✅
        }
    }
    for (HttpRouteProviderContract provider : component.getHttpProviders(app)) {
        for (RouteContract route : provider.getRoutes()) { // direct call ✅
            router.register(route);
        }
    }
}
```

Cache is a cold-start optimization for environments where startup time is critical.

---

## ComponentProviderContract

Top-level aggregator. Returns lists of sub-provider **instances** by category. Build tool reads return values directly
from AST — must be simple `List.of()` literals with no conditional logic. Each list element must be a `new X()`
expression — Sindri reads `ObjectCreationExpr.type` to extract the provider class name.

```java
// package: io.valkyrja.application.provider.contract
package io.valkyrja.application.provider.contract;

import io.valkyrja.application.kernel.contract.ApplicationContract;
import io.valkyrja.cli.routing.provider.contract.CliRouteProviderContract;
import io.valkyrja.container.provider.contract.ServiceProviderContract;
import io.valkyrja.event.provider.contract.ListenerProviderContract;
import io.valkyrja.http.routing.provider.contract.HttpRouteProviderContract;

import java.util.List;

public interface ComponentProviderContract {

    /**
     * Get the component providers this component depends on.
     * The framework ensures all listed components are fully registered
     * before this component's providers are registered.
     * Must return a simple List.of() literal — no conditional logic.
     * Each element must be a new X() expression.
     */
    List<ComponentProviderContract> getComponentProviders(ApplicationContract app);

    /**
     * Get the component's container service providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    List<ServiceProviderContract> getContainerProviders(ApplicationContract app);

    /**
     * Get the component's event listener providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    List<ListenerProviderContract> getEventProviders(ApplicationContract app);

    /**
     * Get the component's CLI route providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    List<CliRouteProviderContract> getCliProviders(ApplicationContract app);

    /**
     * Get the component's HTTP route providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    List<HttpRouteProviderContract> getHttpProviders(ApplicationContract app);
}
```

### HttpComponentProvider Implementation

```java
package io.valkyrja.http.provider;

import io.valkyrja.application.kernel.contract.ApplicationContract;
import io.valkyrja.application.provider.contract.ComponentProviderContract;
import io.valkyrja.cli.routing.provider.contract.CliRouteProviderContract;
import io.valkyrja.container.provider.contract.ServiceProviderContract;
import io.valkyrja.event.provider.contract.ListenerProviderContract;
import io.valkyrja.http.routing.provider.contract.HttpRouteProviderContract;

import java.util.List;

public class HttpComponentProvider implements ComponentProviderContract {

    @Override
    public List<ComponentProviderContract> getComponentProviders(ApplicationContract app) {
        return List.of(
                new ContainerComponentProvider(),  // HTTP depends on Container
                new EventComponentProvider()        // HTTP depends on Event
        );
    }

    @Override
    public List<ServiceProviderContract> getContainerProviders(ApplicationContract app) {
        return List.of(
                new HttpContainerProvider(),
                new HttpMiddlewareProvider()
        );
    }

    @Override
    public List<ListenerProviderContract> getEventProviders(ApplicationContract app) {
        return List.of(
                new HttpEventProvider()
        );
    }

    @Override
    public List<CliRouteProviderContract> getCliProviders(ApplicationContract app) {
        return List.of();
    }

    @Override
    public List<HttpRouteProviderContract> getHttpProviders(ApplicationContract app) {
        return List.of(
                new HttpRouteProvider()
        );
    }
}
```

---

## ServiceProviderContract

Container bindings provider. `publishers()` returns a map of class token to publisher method reference. The build tool
reads the map from AST, resolves each method reference, and reads that method body directly — no `@Handler` annotation
needed on publisher methods.

```java
package io.valkyrja.container.provider.contract;

import io.valkyrja.container.manager.contract.ContainerContract;

import java.util.Map;
import java.util.function.Consumer;

public interface ServiceProviderContract {

    /**
     * Any custom publishers for services provided by this provider.
     *
     * The map must be a simple Map.of() literal — no conditional logic.
     * Each value must be a static method reference on the same class.
     * The build tool reads the map from AST, resolves each method reference,
     * and reads that method body directly for cache generation.
     *
     * No @Handler annotation is needed on publisher methods.
     *
     * Example:
     *   Map.of(
     *       UserRepositoryContract.class, UserServiceProvider::publishUserRepository
     *   )
     *
     *   public static void publishUserRepository(ContainerContract container) {
     *       container.setSingleton(
     *           UserRepositoryContract.class,
     *           new UserRepository(container.make(DatabaseContract.class))
     *       );
     *   }
     */
    Map<Class<?>, Consumer<ContainerContract>> publishers();
}
```

### UserServiceProvider Implementation

```java
package app.providers;

import io.valkyrja.container.manager.contract.ContainerContract;
import io.valkyrja.container.provider.contract.ServiceProviderContract;
import app.repositories.UserRepository;
import app.repositories.contract.UserRepositoryContract;
import app.services.contract.DatabaseContract;

import java.util.Map;
import java.util.function.Consumer;

public class UserServiceProvider implements ServiceProviderContract {

    /**
     * Build tool reads this map from AST, resolves each method reference,
     * then reads each publisher method body for cache generation.
     */
    @Override
    public Map<Class<?>, Consumer<ContainerContract>> publishers() {
        return Map.of(
                UserRepositoryContract.class, UserServiceProvider::publishUserRepository
        );
    }

    /**
     * Build tool reads this method body from AST for cache generation.
     * No @Handler annotation needed — method is discovered via publishers() map.
     */
    public static void publishUserRepository(ContainerContract container) {
        container.setSingleton(
                UserRepositoryContract.class,
                new UserRepository(container.make(DatabaseContract.class))
        );
    }
}
```

---

## HttpRouteProviderContract

HTTP route provider. Two sources of routes: annotated controller classes (scanned for `@Handler`) and explicit route
object definitions. Routes are data structures — they carry method, path, middleware, constraints, and the handler
together. They cannot be expressed as a publisher-style map.

```java
package io.valkyrja.http.routing.provider.contract;

import io.valkyrja.http.routing.data.contract.RouteContract;

import java.util.List;

public interface HttpRouteProviderContract {

    /**
     * Get a list of attributed controller or action classes.
     * Build tool scans each class for @Handler annotations.
     * Returns empty list if using explicit routes only.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<Class<?>> getControllerClasses();

    /**
     * Get a list of explicit route definitions.
     * Routes are complete data structures — they carry method, path,
     * middleware, constraints, and handler together.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<RouteContract> getRoutes();
}
```

### UserHttpRouteProvider Implementation

```java
package app.http.providers;

import io.valkyrja.container.manager.contract.ContainerContract;
import io.valkyrja.http.routing.data.HttpRoute;
import io.valkyrja.http.routing.data.contract.RouteContract;
import io.valkyrja.http.routing.provider.contract.HttpRouteProviderContract;
import app.http.controllers.UserController;
import app.http.controllers.OrderController;

import java.util.List;
import java.util.Map;

public class UserHttpRouteProvider implements HttpRouteProviderContract {

    @Override
    public List<Class<?>> getControllerClasses() {
        return List.of(
                UserController.class,
                OrderController.class
        );
    }

    /**
     * Handler is a method pointer on this same class.
     * Sindri reads the handler method body from this file only — no cross-file imports.
     */
    @Override
    public List<RouteContract> getRoutes() {
        return List.of(
                HttpRoute.get("/orders", UserHttpRouteProvider::indexOrders)
        );
    }

    /** Handler method lives on the same class — all imports self-contained. */
    public static ResponseContract indexOrders(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(OrderController.class).index(args);
    }
}
```

---

## Annotated Controller — Java

`@Handler` lives on the **implementation method** and carries a callable reference — class + method name. The handler
may live on the controller, the route provider, or any other class.

**Handler on the same controller:**

```java
package app.http.controllers;

import io.valkyrja.container.manager.contract.ContainerContract;
import io.valkyrja.http.routing.data.contract.ResponseContract;
import io.valkyrja.http.routing.annotation.Handler;
import io.valkyrja.http.routing.annotation.Parameter;
import io.valkyrja.http.routing.annotation.Route;

import java.util.Map;

public class UserController {

    // Annotations on the implementation method.
    // @Handler carries (class, method) — Sindri follows it to wherever the handler lives.
    @Route(method = "GET", path = "/users/{id}")
    @Parameter(name = "id", pattern = "[0-9]+")
    @Handler(clazz = UserController.class, method = "showHandler")
    public ResponseContract show(String id) {
        return userService.findById(id).toResponse();
    }

    @Route(method = "POST", path = "/users")
    @Handler(clazz = UserController.class, method = "storeHandler")
    public ResponseContract store(Map<String, Object> data) {
        // actual implementation
    }

    // Sindri resolves clazz=UserController.class, method="showHandler" → this file
    // reads this method body using this file's imports
    public static ResponseContract showHandler(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(UserController.class).show((String) args.get("id"));
    }

    public static ResponseContract storeHandler(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(UserController.class).store(args);
    }
}
```

**Handler on the route provider:**

```java
public class UserController {

    // @Handler points to the route provider — Sindri follows the callable
    @Route(method = "GET", path = "/users/{id}")
    @Parameter(name = "id", pattern = "[0-9]+")
    @Handler(clazz = UserHttpRouteProvider.class, method = "showUser")
    public ResponseContract show(String id) {
        // actual implementation
    }
}

public class UserHttpRouteProvider implements HttpRouteProviderContract {

    // Sindri resolves callable → this file, reads this method using this file's imports
    public static ResponseContract showUser(ContainerContract c, Map<String, Object> args) {
        return c.getSingleton(UserController.class).show((String) args.get("id"));
    }
}
```

---

## CliRouteProviderContract

```java
package io.valkyrja.cli.routing.provider.contract;

import io.valkyrja.cli.routing.data.contract.RouteContract;

import java.util.List;

public interface CliRouteProviderContract {

    /**
     * Get a list of attributed controller or action classes.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<Class<?>> getControllerClasses();

    /**
     * Get a list of explicit CLI route definitions.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<RouteContract> getRoutes();
}
```

---

## ListenerProviderContract

```java
package io.valkyrja.event.provider.contract;

import io.valkyrja.event.data.contract.ListenerContract;

import java.util.List;

public interface ListenerProviderContract {

    /**
     * Get a list of attributed listener classes.
     * Build tool scans each class for @Handler annotations.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<Class<?>> getListenerClasses();

    /**
     * Get a list of explicit listener definitions.
     * Listeners are complete data structures — they carry event type,
     * priority, and handler together.
     * Must be a simple List.of() literal — no conditional logic.
     */
    List<ListenerContract> getListeners();
}
```

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic:

```
✅ return List.of(new UserController(), new OrderController());
✅ return Map.of(UserRepositoryContract.class, UserServiceProvider::publishUserRepository);
✅ return List.of(HttpRoute.get("/users", UserHttpRouteProvider::indexUsers));

❌ if (condition) { return List.of(...); }
❌ return List.of(getExtra());
❌ var list = new ArrayList<>(); list.add(...); return list;
```

---

## Handler Method Pointer Convention

All handler methods must be **static methods on the same class** as the provider or controller that defines the route or
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

An early consideration was expressing routes the same way as container bindings — a map of route key to handler method
reference, with the build tool reading method bodies directly. This was rejected because routes are multi-dimensional
data structures, not simple key→factory pairs.

A route carries: HTTP method, path pattern, dynamic segment constraints, regex compilation data, middleware chain,
name/alias, parameter defaults, host constraints, and scheme constraints — all in addition to the handler. The
`HttpRoute.get("/users/{id}", handler)` call is what populates all of these fields together. Decomposing this into a
key/method-body map would lose all metadata the router needs to build its dispatcher index and compile route regexes.

The same reasoning applies to listeners — they carry event type binding, priority, and stop-propagation behavior
alongside the handler. These cannot be expressed as a flat key/body map without losing the data the event dispatcher
requires.

Container bindings by contrast are simple key→factory pairs. The binding key IS the complete identity. This is why
`publishers()` works as a map but `getRoutes()` must return complete route objects.
