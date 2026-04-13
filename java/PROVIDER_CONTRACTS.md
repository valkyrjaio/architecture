# Valkyrja Provider Contracts — Java

## Overview

Java provider contracts mirror the PHP reference implementation. Key differences:

- `.class` used instead of `::class` for class references — compiler-verified type tokens
- `List.of(...)` for immutable list returns
- `Map.of(...)` for immutable map returns
- All Valkyrja exceptions extend `RuntimeException` (unchecked)
- Publisher methods have no `@Handler` annotation — the build tool reads method bodies directly from AST
- Static interface methods with default empty returns

---

## ComponentProviderContract

Top-level aggregator. Returns lists of sub-providers by category. Build tool reads return values directly from AST —
must be simple list literals with no conditional logic.

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
     * Get the component's container service providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    static List<Class<? extends ServiceProviderContract>> getContainerProviders(
            ApplicationContract app
    ) {
        return List.of();
    }

    /**
     * Get the component's event listener providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    static List<Class<? extends ListenerProviderContract>> getEventProviders(
            ApplicationContract app
    ) {
        return List.of();
    }

    /**
     * Get the component's CLI route providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    static List<Class<? extends CliRouteProviderContract>> getCliProviders(
            ApplicationContract app
    ) {
        return List.of();
    }

    /**
     * Get the component's HTTP route providers.
     * Must return a simple List.of() literal — no conditional logic.
     */
    static List<Class<? extends HttpRouteProviderContract>> getHttpProviders(
            ApplicationContract app
    ) {
        return List.of();
    }
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

    public static List<Class<? extends ServiceProviderContract>> getContainerProviders(
            ApplicationContract app
    ) {
        return List.of(
                HttpContainerProvider.class,
                HttpMiddlewareProvider.class
        );
    }

    public static List<Class<? extends ListenerProviderContract>> getEventProviders(
            ApplicationContract app
    ) {
        return List.of(
                HttpEventProvider.class
        );
    }

    public static List<Class<? extends CliRouteProviderContract>> getCliProviders(
            ApplicationContract app
    ) {
        return List.of();
    }

    public static List<Class<? extends HttpRouteProviderContract>> getHttpProviders(
            ApplicationContract app
    ) {
        return List.of(
                HttpRouteProvider.class
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
    static Map<Class<?>, Consumer<ContainerContract>> publishers() {
        return Map.of();
    }
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
    public static Map<Class<?>, Consumer<ContainerContract>> publishers() {
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
    static List<Class<?>> getControllerClasses() {
        return List.of();
    }

    /**
     * Get a list of explicit route definitions.
     * Routes are complete data structures — they carry method, path,
     * middleware, constraints, and handler together.
     * Must be a simple List.of() literal — no conditional logic.
     */
    static List<RouteContract> getRoutes() {
        return List.of();
    }
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

    public static List<Class<?>> getControllerClasses() {
        return List.of(
                UserController.class,
                OrderController.class
        );
    }

    /**
     * Handler is a method pointer on this same class.
     * Forge reads the handler method body from this file only — no cross-file imports.
     */
    public static List<RouteContract> getRoutes() {
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
        // actual implementation — not read by Sindri
        return userService.findById(id).toResponse();
    }

    @Route(method = "POST", path = "/users")
    @Handler(clazz = UserController.class, method = "storeHandler")
    public ResponseContract store(Map<String, Object> data) {
        // actual implementation
    }

    // Forge resolves clazz=UserController.class, method="showHandler" → this file
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

    // Forge resolves callable → this file, reads this method using this file's imports
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
    static List<Class<?>> getControllerClasses() {
        return List.of();
    }

    /**
     * Get a list of explicit CLI route definitions.
     * Must be a simple List.of() literal — no conditional logic.
     */
    static List<RouteContract> getRoutes() {
        return List.of();
    }
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
    static List<Class<?>> getListenerClasses() {
        return List.of();
    }

    /**
     * Get a list of explicit listener definitions.
     * Listeners are complete data structures — they carry event type,
     * priority, and handler together.
     * Must be a simple List.of() literal — no conditional logic.
     */
    static List<ListenerContract> getListeners() {
        return List.of();
    }
}
```

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic:

```
✅ return List.of(UserController.class, OrderController.class);
✅ return Map.of(UserRepositoryContract.class, UserServiceProvider::publishUserRepository);
✅ return List.of(HttpRoute.get("/users", (c, args) -> ...));

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
