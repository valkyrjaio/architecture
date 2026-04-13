# Valkyrja Provider Contracts — Python

## Overview

Python provider contracts differ from PHP/Java in several ways:

- Decorators are **metadata markers** — they attach closure metadata to methods at import time but do NOT self-register
  routes. The framework reads metadata during bootstrap; skips it when loading from cache.
- `inspect.getfile(ClassName)` resolves class to source file — equivalent of PHP's `ReflectionClass::getFileName()`
- No `::class` needed — class objects are first-class values in Python, passed directly
- ABC enforces abstract contracts — `TypeError` raised on direct instantiation
- `@staticmethod @abstractmethod` throughout — providers are stateless
- Publisher methods have a `@handler` decorator carrying the closure — build tool reads the decorator argument from AST
- `class_` helper available (trailing underscore because `class` is reserved) for FQN string derivation

---

## Type Hints

Python classes are first-class `type` objects — passing a class directly is the idiomatic equivalent of PHP's `::class`
and Java's `.class`. The type hints reflect this accurately:

| Method                      | Return type                                      | Reasoning                                                      |
|-----------------------------|--------------------------------------------------|----------------------------------------------------------------|
| `get_container_providers()` | `list[type]`                                     | Returns class objects implementing `ServiceProviderContract`   |
| `get_event_providers()`     | `list[type]`                                     | Returns class objects implementing `ListenerProviderContract`  |
| `get_cli_providers()`       | `list[type]`                                     | Returns class objects implementing `CliRouteProviderContract`  |
| `get_http_providers()`      | `list[type]`                                     | Returns class objects implementing `HttpRouteProviderContract` |
| `get_controller_classes()`  | `list[type]`                                     | Returns class objects carrying `@handler` decorated methods    |
| `get_listener_classes()`    | `list[type]`                                     | Returns class objects carrying `@handler` decorated methods    |
| `get_routes()`              | `list[RouteContract]`                            | Returns concrete route data objects                            |
| `get_listeners()`           | `list[ListenerContract]`                         | Returns concrete listener data objects                         |
| `publishers()`              | `dict[str, Callable[[ContainerContract], None]]` | Maps binding key to publisher function reference               |

`list[type]` is used for class lists rather than `list[Type[SomeContract]]` because controller and listener classes do
not implement a provider contract — they carry `@handler` decorators. `list[type]` is the honest and accurate type for
any list of Python class objects.

For stricter typing on provider class lists, `list[Type[ServiceProviderContract]]` etc. can be used and mypy/pyright
will validate that the listed classes implement the correct contract.

---

## ComponentProviderContract

Top-level aggregator. Returns lists of sub-provider classes. Build tool reads return values directly from AST — must be
simple list literals with no conditional logic.

```python
# package: valkyrja.application.provider.contract
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from valkyrja.application.kernel.contract import ApplicationContract
    from valkyrja.container.provider.contract import ServiceProviderContract
    from valkyrja.event.provider.contract import ListenerProviderContract
    from valkyrja.cli.routing.provider.contract import CliRouteProviderContract
    from valkyrja.http.routing.provider.contract import HttpRouteProviderContract


class ComponentProviderContract(ABC):
    """
    Defines what a component provider must implement.

    All methods must return simple list literals.
    No conditional logic permitted — build tool reads these from AST.
    """

    @staticmethod
    @abstractmethod
    def get_container_providers(app: 'ApplicationContract') -> list:
        """
        Get the component's container service providers.
        Must return a simple list literal — no conditional logic.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_event_providers(app: 'ApplicationContract') -> list:
        """
        Get the component's event listener providers.
        Must return a simple list literal — no conditional logic.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_cli_providers(app: 'ApplicationContract') -> list:
        """
        Get the component's CLI route providers.
        Must return a simple list literal — no conditional logic.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_http_providers(app: 'ApplicationContract') -> list:
        """
        Get the component's HTTP route providers.
        Must return a simple list literal — no conditional logic.
        """
        return []
```

### HttpComponentProvider Implementation

```python
from valkyrja.application.provider.contract import ComponentProviderContract
from valkyrja.http.provider import (
    HttpContainerProvider,
    HttpMiddlewareProvider,
    HttpEventProvider,
    HttpRouteProvider,
)


class HttpComponentProvider(ComponentProviderContract):

    @staticmethod
    def get_container_providers(app) -> list[type]:
        return [
            HttpContainerProvider,
            HttpMiddlewareProvider,
        ]

    @staticmethod
    def get_event_providers(app) -> list[type]:
        return [
            HttpEventProvider,
        ]

    @staticmethod
    def get_cli_providers(app) -> list[type]:
        return []

    @staticmethod
    def get_http_providers(app) -> list[type]:
        return [
            HttpRouteProvider,
        ]
```

---

## ServiceProviderContract

Container bindings provider. `publishers()` returns a map of binding key to publisher method reference. The build tool
reads the map from AST, resolves each method reference via `inspect.getfile()`, and reads that method body. Publisher
methods carry a `@handler` decorator — the build tool reads the decorator argument from AST for cache generation.

```python
# package: valkyrja.container.provider.contract
from abc import ABC, abstractmethod
from typing import Callable, TYPE_CHECKING

if TYPE_CHECKING:
    from valkyrja.container.manager.contract import ContainerContract


class ServiceProviderContract(ABC):
    """
    Defines what a container service provider must implement.

    publishers() returns a map of binding key to publisher method reference.
    The map must be a simple dict literal — no conditional logic permitted.
    Each value must be a static method reference on the same class.

    The build tool reads the publishers map from AST, resolves each method
    reference via inspect.getfile(), and reads the _valkyrja_handler metadata
    on that method for cache generation. The @handler decorator on publisher
    methods is a metadata marker only — it does not execute at import time.

    Note: 'class_' helper available for FQN derivation since 'class' is reserved:
        def class_(cls) -> str:
            return f"{cls.__module__}.{cls.__qualname__}"

    Example:
        @staticmethod
        def publishers() -> dict:
            return {
                UserRepositoryClass: UserServiceProvider.publish_user_repository,
            }

        @handler(lambda c, args: c.set_singleton(
            UserRepositoryClass, UserRepository(c.make(DatabaseClass))
        ))
        @staticmethod
        def publish_user_repository(container: ContainerContract) -> None:
            container.set_singleton(
                UserRepositoryClass,
                UserRepository(container.make(DatabaseClass))
            )
    """

    @staticmethod
    @abstractmethod
    def publishers() -> dict[str, Callable[['ContainerContract'], None]]:
        """
        Return a map of binding key to publisher method reference.
        Must return a simple dict literal — no conditional logic permitted.
        """
        return {}
```

### UserServiceProvider Implementation

```python
from valkyrja.container.provider.contract import ServiceProviderContract
from valkyrja.container.manager.contract import ContainerContract
from app.repositories import UserRepository
from app.repositories.contract import UserRepositoryClass
from app.services.contract import DatabaseClass


class UserServiceProvider(ServiceProviderContract):

    @staticmethod
    def publishers() -> dict:
        """
        Build tool reads this map from AST, resolves each method reference
        via inspect.getfile(), then reads each method's @handler decorator
        for cache generation.
        """
        return {
            UserRepositoryClass: UserServiceProvider.publish_user_repository,
        }

    @handler(lambda c, args: c.set_singleton(
        UserRepositoryClass,
        UserRepository(c.make(DatabaseClass))
    ))
    @staticmethod
    def publish_user_repository(container: ContainerContract) -> None:
        """
        Build tool reads the @handler decorator argument from AST.
        The decorator carries the closure used for cache generation.
        The method body is the runtime implementation.
        """
        container.set_singleton(
            UserRepositoryClass,
            UserRepository(container.make(DatabaseClass))
        )
```

---

## HttpRouteProviderContract

HTTP route provider. Two sources: annotated controller classes (scanned for `@handler` decorated methods) and explicit
route object definitions. Routes are complete data structures — they cannot be expressed as a publisher-style map
without losing the metadata the router requires.

```python
# package: valkyrja.http.routing.provider.contract
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from valkyrja.http.routing.data.contract import RouteContract


class HttpRouteProviderContract(ABC):
    """Defines what an HTTP route provider must implement."""

    @staticmethod
    @abstractmethod
    def get_controller_classes() -> list:
        """
        Get a list of attributed controller or action classes.
        Build tool uses inspect.getfile() to locate each class source file,
        then scans for @handler decorated methods.
        Returns empty list if using explicit routes only.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_routes() -> list:
        """
        Get a list of explicit route definitions.
        Routes are complete data structures — they carry HTTP method, path pattern,
        dynamic segment constraints, middleware chain, and handler together.
        They cannot be expressed as a publisher-style map without losing
        the metadata the router needs to build its dispatcher index.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []
```

### UserHttpRouteProvider Implementation

```python
from valkyrja.http.routing.provider.contract import HttpRouteProviderContract
from valkyrja.http.routing.data import HttpRoute
from app.http.controllers import UserController, OrderController
from app.http.controllers.contract import OrderControllerClass


class UserHttpRouteProvider(HttpRouteProviderContract):

    @staticmethod
    def get_controller_classes() -> list[type]:
        """
        Build tool calls inspect.getfile(UserController) to locate source,
        then scans for @handler decorated methods.
        Python classes are first-class type objects — list[type] is accurate.
        """
        return [
            UserController,
            OrderController,
        ]

    @staticmethod
    def get_routes() -> list[RouteContract]:
        """
        Handler is a method pointer on this same class.
        Forge reads handler method bodies from this file only.
        """
        return [
            HttpRoute.get('/orders', UserHttpRouteProvider.index_orders),
            HttpRoute.get('/users', UserHttpRouteProvider.index_users),
        ]

    @staticmethod
    def index_orders(c: ContainerContract, args: dict) -> ResponseContract:
        """Handler method lives on the same class — all imports self-contained."""
        return c.get_singleton(OrderControllerClass).index(args)

    @staticmethod
    def index_users(c: ContainerContract, args: dict) -> ResponseContract:
        return c.get_singleton(UserControllerClass).index(args)
```

### Controller with @handler Decorator

The `@handler` decorator is a **metadata marker only** — it does not self-register routes at import time. It attaches
the closure as metadata on the method. The framework reads this metadata during bootstrap (no cache) and skips it
entirely when loading from cache.

This is intentional and consistent with PHP's `#[Handler]` attribute — both are inert metadata that the framework reads
when needed, not active registrars.

```python
from valkyrja.http.routing.handler import handler
from valkyrja.container.manager.contract import ContainerContract
from app.http.controllers.contract import UserControllerClass


def handler(closure):
    """
    Metadata marker — attaches closure to method as _valkyrja_handler.
    Does NOT register the route at import time.
    Framework reads _valkyrja_handler during bootstrap (no cache).
    Framework skips entirely when loading from cache.
    """

    def decorator(func):
        func._valkyrja_handler = closure  # metadata only — no registration
        return func

    return decorator


class UserController:

    @handler(lambda c, args: c.get_singleton(UserControllerClass).index(args[0]))
    def index(self, request) -> Response:
        """
        Build tool reads _valkyrja_handler metadata from AST
        when scanning this class for route handlers.
        The decorator carries the closure used in cache generation.
        The method body is the actual runtime implementation.
        """
        pass

    @handler(lambda c, args: c.get_singleton(UserControllerClass).store(args[0]))
    def store(self, request) -> Response:
        pass
```

### Why Not Self-Registration

Python decorators execute at module import time. If `@handler` self-registered routes, importing a controller module
would immediately register its routes — even when loading from cache where those routes are already pre-built. The cache
data file imports the same controller classes anyway (to reference them in route objects), so the imports cannot be
avoided. Self-registration would cause double registration or conflicting state.

Metadata-first solves this cleanly: the decorator is always inert, the framework decides whether to read the metadata
based on whether cache is loaded.

---

## CliRouteProviderContract

```python
# package: valkyrja.cli.routing.provider.contract
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from valkyrja.cli.routing.data.contract import RouteContract


class CliRouteProviderContract(ABC):
    """Defines what a CLI route provider must implement."""

    @staticmethod
    @abstractmethod
    def get_controller_classes() -> list:
        """
        Get a list of attributed controller or action classes.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_routes() -> list:
        """
        Get a list of explicit CLI route definitions.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []
```

---

## ListenerProviderContract

```python
# package: valkyrja.event.provider.contract
from abc import ABC, abstractmethod
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from valkyrja.event.data.contract import ListenerContract


class ListenerProviderContract(ABC):
    """Defines what an event listener provider must implement."""

    @staticmethod
    @abstractmethod
    def get_listener_classes() -> list:
        """
        Get a list of attributed listener classes.
        Build tool uses inspect.getfile() to locate each class source file,
        then scans for @handler decorated methods.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []

    @staticmethod
    @abstractmethod
    def get_listeners() -> list:
        """
        Get a list of explicit listener definitions.
        Listeners carry event type, priority, and handler together.
        Cannot be expressed as a key/body map without losing
        the metadata the event dispatcher requires.
        Must return a simple list literal — no conditional logic permitted.
        """
        return []
```

---

## Build Tool Contract

Any method the build tool reads must return a single flat literal with no logic:

```python
# ✅ simple list literal
return [UserController, OrderController]

# ✅ simple dict literal with method reference
return {UserRepositoryClass: UserServiceProvider.publish_user_repository}

# ✅ simple list of route objects
return [HttpRoute.get('/users', lambda c, args: ...)]

# ❌ conditional logic
if condition:
    return [...]
return [...]

# ❌ variable accumulation
routes = []
routes.append(...)
return routes

# ❌ method calls other than constructors/static factories
return get_extra_routes()
```

---

## Handler Method Pointer Convention

All handler methods must be **static methods on the same class** as the provider or controller that defines the route or
listener. This is the same pattern used by `publishers()` in service providers.

**Why:** The forge tool reads exactly one file per provider or controller. All imports for handler bodies are in that
one file — no cross-file import aggregation, no conflict detection, no registry needed.

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

Container bindings by contrast are simple key→factory pairs. This is why `publishers()` works as a map but
`get_routes()` must return complete route objects.
