# Python Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `HANDLERS.md`, `DATA_CACHE.md`, `SINDRI.md`,
> `PROVIDER_CONTRACTS.md`
> Port order: the dependency tiers in `convention/PORT_PARITY.md`

---

## Key Language Decisions

- **Module namespace:** `valkyrja`
- **ABC** enforces abstract contracts — `TypeError` on direct instantiation
- **`@staticmethod @abstractmethod`** throughout — providers are stateless
- **`inspect.getfile()`** for class-to-file resolution (equivalent of PHP's `ReflectionClass::getFileName()`)
- **`ast` module** for build tool AST parsing
- **Decorators are runtime-executable**, which is why `@route_handler` must be metadata only and never self-register
- **`class_()` helper** for FQN derivation (`class` is reserved in Python)
- **ASGI (Uvicorn/Hypercorn)** as the worker mode deployment model
- **CGI mode** supported — Python is interpreted, cache optional in dev
- **Granian** (Rust-based) as an emerging alternative for true multi-threaded workers
- GIL limits true thread parallelism — ASGI async is the idiomatic concurrency model

---

## 1. Throwables

**Reference:** `THROWABLES.md`

### Hierarchy

```python
# Throwable branch
class ValkyrjaThrowable(BaseException, ABC): pass


class ComponentThrowable(ValkyrjaThrowable, ABC): pass  # always present


class ComponentSpecificThrowable(ComponentThrowable): pass  # concrete


# RuntimeException branch
class ValkyrjaRuntimeException(RuntimeError, ABC): pass


class ComponentRuntimeException(ValkyrjaRuntimeException, ABC): pass  # always present


class ComponentSpecificException(ComponentRuntimeException): pass  # concrete


# InvalidArgumentException branch — parity name, extends ValueError
class ValkyrjaInvalidArgumentException(ValueError, ABC): pass


class ComponentInvalidArgumentException(ValkyrjaInvalidArgumentException, ABC): pass


class ComponentSpecificInvalidArgumentException(ComponentInvalidArgumentException): pass
```

### Rules

- `ValkyrjaInvalidArgumentException` extends `ValueError` for language-level catchability
- All base and categorical exceptions are abstract via ABC
- Every component ships both categoricals even if unused
- Naming: `ComponentName*`, shared subcomponents `ParentComponentSubComponent*`
- `except ContainerNotFoundException as e:` — Python resolves top-to-bottom, specific first

---

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### String constants as keys — same as Go and TypeScript

Python requires string constants for container binding keys. Using class objects as keys forces the module to be
imported at key definition time — the class object cannot exist without its module loading. String keys avoid that
eager import entirely, independent of any lazy-import language feature (see §5).

```python
# container_constants.py — required per component
class ContainerConstants:
    CONTAINER = "valkyrja.container.manager.ContainerContract"
    ROUTER = "valkyrja.http.routing.dispatcher.RouterContract"
    USER_REPOSITORY = "app.repository.UserRepositoryContract"
    DATABASE = "app.service.DatabaseContract"
```

```python
# bind and resolve via string constant
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.get_singleton(ContainerConstants.DATABASE))
)

repo = container.get(ContainerConstants.USER_REPOSITORY)
# UserRepository is only *used* when the lambda runs; without PEP 810 its top-level import still loads eagerly
```

### FQN helper — for generating string constants

```python
# 'class' is reserved — use class_()
# Useful for generating constant values, logging, debugging
def class_(cls) -> str:
    return f"{cls.__module__}.{cls.__qualname__}"
```

---

## 3. Provider Contracts

**Reference:** `PROVIDER_CONTRACTS.md`, `DATA_CACHE.md`

### Type hints

```python
get_container_providers() -> list[type]  # class objects
get_event_providers()     -> list[type]
get_cli_providers()       -> list[type]
get_http_providers()      -> list[type]
get_controller_classes()  -> list[type]  # classes with @route_handler decorators
get_listener_classes()    -> list[type]
get_routes()              -> list[RouteContract]  # concrete route objects
get_listeners()           -> list[ListenerContract]
publishers()              -> dict[str, Callable[[ContainerContract], None]]
```

### ComponentProviderContract

```python
class ComponentProviderContract(ABC):
    @staticmethod @ abstractmethod
    def get_container_providers(app: ApplicationContract) -> list[type]: ...

    @staticmethod @ abstractmethod
    def get_event_providers(app: ApplicationContract) -> list[type]: ...

    @staticmethod @ abstractmethod
    def get_cli_providers(app: ApplicationContract) -> list[type]: ...

    @staticmethod @ abstractmethod
    def get_http_providers(app: ApplicationContract) -> list[type]: ...
```

### ServiceProviderContract

```python
class ServiceProviderContract(ABC):
    @staticmethod @ abstractmethod
    def publishers() -> dict[str, Callable[[ContainerContract], None]]: ...
```

The `publishers()` map names each publisher method, and the build tool reads that map from AST:

```python
@staticmethod
def publishers() -> dict:
    return {
        UserRepositoryClass: UserServiceProvider.publish_user_repository,
    }


@staticmethod
def publish_user_repository(container: ContainerContract) -> None:
    container.set_singleton(UserRepositoryClass, UserRepository(container.get_singleton(DatabaseClass)))
```

### HttpRouteProviderContract

```python
class HttpRouteProviderContract(ABC):
    @staticmethod @ abstractmethod
    def get_controller_classes() -> list[type]: ...  # classes with @route_handler decorators

    @staticmethod @ abstractmethod
    def get_routes() -> list[RouteContract]: ...
```

All provider methods must return simple list/dict literals — no conditional logic.

---

## 4. Handler Signatures

**Reference:** `HANDLERS.md`

### Three Callable type aliases

A route handler takes the container and the **route**. Only a listener takes a
map, because a listener has no route. See
[`HANDLERS.md`](../../convention/HANDLERS.md).

```python
from typing import Callable, Any

# HTTP route
Callable[[ContainerContract, RouteContract], ResponseContract]

# CLI route
Callable[[ContainerContract, RouteContract], OutputContract]

# Event listener
Callable[[ContainerContract, dict[str, Any]], Any]
```

### The handler lives on the route contract

There is no handler contract per concern. The route contract declares
`get_handler` and `with_handler`, and `with_handler` returns a copy rather than
modifying the route, per
[`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md).

### @route_handler decorator on controller methods

```python
@route_handler(UserHttpRouteProvider.show_user)
@parameter('id', pattern='[0-9]+')
def show(self, route: RouteContract) -> ResponseContract:
    pass
```

The **route is a parameter**, and it is also set on the container before the
handler runs, so it is reachable both ways. `ServerRequestContract` is
container-only — it is never a handler parameter.

---

## 5. Python Imports and Cold Start — an Open Problem (PEP 690 Withdrawn)

Eager imports are a well-known Python cold-start cost. FastAPI, Django, and Flask all pay it; large applications on AWS
Lambda have reported multi-second cold starts. Python imports every referenced module at import time, and **no stable
language feature changes this today.**

### The lazy-import plan changed — PEP 690 was withdrawn

An earlier draft of this port assumed **PEP 690 (implicit lazy imports)** would ship in Python 3.14 and make top-level
imports lazy by default. **That did not happen.** PEP 690 was withdrawn and never merged; Python 3.14 does **not**
lazy-load imports. Its successor, **PEP 810 (explicit lazy imports)** — an opt-in `lazy import ...` form — is still under
discussion and unshipped as of this writing. Treat lazy imports as a **possible future optimization**, not something the
port relies on.

### Why the architecture still holds

The container design does **not** depend on lazy imports for correctness — it only _benefits_ from them if they arrive:

- **String-constant binding keys** avoid importing the bound class at key-definition time. This is plain Python, true in
  every version: a string literal loads nothing, whereas a class-object key forces the import.

  ```python
  # class object key — forces module import when the dict key is evaluated
  {UserRepositoryContract: lambda: ...}  # UserRepositoryContract accessed → loads

  # string key — no import triggered
  {'app.repositories.UserRepositoryContract': lambda: ...}  # string literal — nothing loads
  ```

- **Lambda-wrapped values** defer _when the provider method is referenced_ from cache-load time to first resolution.
  Also version-independent — a name inside a lambda body is not evaluated until the lambda runs:

  ```python
  # NO lambda — UserServiceProvider accessed at module level when the dict is built (loads at cache load)
  {ContainerConstants.USER_REPOSITORY: UserServiceProvider.publish_user_repository}

  # WITH lambda — UserServiceProvider inside the lambda body — loads only when the lambda is called
  {ContainerConstants.USER_REPOSITORY: lambda: UserServiceProvider.publish_user_repository}
  ```

  The thunk is required wherever the language has **no compile-time reference to a type**, which is the permitted
  variation in `convention/CONTAINER_BINDINGS.md`. A port whose reference is resolved by the compiler names the method
  directly and loads nothing.

**Why not a lambda as the key?** The deferral trick that works for values cannot work for keys — Python must evaluate
every key at dict-construction time to know where to store the value, and would have to re-call a key-lambda on every
lookup. **String constants are the correct and final answer for binding keys** (same as Go and TypeScript);
per-component constants files are required. Accepted and closed.

What none of this removes today is the **top-level `import` of a provider's own dependencies**: once a provider module
loads, its module-level imports load eagerly. A future PEP 810 could turn those into `lazy import` and defer them
per-name — and the framework would need no change to benefit — but it works correctly without it.

### What the cache provides (independent of lazy imports)

```
Without cache — every boot:
  ✗ Traverse the provider tree
  ✗ Scan @route_handler decorators across all controllers
  ✗ Build the route dispatcher index (regex compilation, path indexing)
  ✗ Register all container bindings

With cache — every boot:
  ✓ Load one pre-built data class per major component
  ✓ Skip provider-tree traversal, annotation scanning, and route-index construction
```

Cache removes the framework's own boot work. It does **not** eliminate Python's module-import cost — that remains the
language's cold-start weakness until (and unless) explicit lazy imports ship.

### The Multi-Language Escape Valve

For Lambda-heavy workloads where cold start is the binding constraint, the **Go or TypeScript** port gives compiled /
fast-start runtimes within the same framework ecosystem — same architecture, same patterns, different runtime. This is
the recommended answer for cold-start-sensitive deployments, rather than a language feature that has not materialized.

---

## 6. Container — Uniform Lambda Format

The container's internal bindings map always stores lambdas — whether populated from a service provider at runtime or
loaded from a cache data file. This makes the internal format identical in both paths, and resolution always calls the
lambda with no conditional logic.

### Three Parties, One Job Each

**Service provider** — uses string constant key, plain method reference value. No lambda:

```python
class UserServiceProvider(ServiceProviderContract):
    @staticmethod
    def publishers() -> dict[str, Callable[[ContainerContract], None]]:
        return {
            ContainerConstants.USER_REPOSITORY: UserServiceProvider.publish_user_repository,
        }

    @staticmethod
    def publish_user_repository(c: ContainerContract) -> None:
        c.set_singleton(
            ContainerConstants.USER_REPOSITORY,
            UserRepository(c.get_singleton(ContainerConstants.DATABASE))
        )
```

**Container** — wraps the method reference in a lambda when registering from a provider. Internal map always stores
lambdas:

```python
class Container:

    def register_provider(self, provider: ServiceProviderContract) -> None:
        for key, callable_ref in provider.publishers().items():
            # wrap in lambda — internal map always stores lambdas
            self._bindings[key] = lambda c=callable_ref: c

    def load_cache(self, data: dict) -> None:
        # cache data already in lambda format — register as-is
        self._bindings.update(data)

    def get_service(self, key: str):
        # always call the lambda — uniform, no conditional check needed
        callable_ref = self._bindings[key]()
        return callable_ref(self)

    def get_singleton(self, key: str):
        if key not in self._singletons:
            self._singletons[key] = self.get_service(key)
        return self._singletons[key]
```

**sindri** (the build tool) — reads `publishers()` from AST. Writes constant keys as module-level imports (load at boot). Wraps method
reference values in lambdas (load only when binding resolved). Cache matches the container's internal format exactly:

```python
# generated AppContainerData
from app.constants.container_constants import ContainerConstants  # loads at boot

APP_CONTAINER_DATA = {
    ContainerConstants.USER_REPOSITORY: lambda: UserServiceProvider.publish_user_repository,
    ContainerConstants.ORDER_REPOSITORY: lambda: OrderServiceProvider.publish_order_repository,
    # constants load at boot — providers load only when lambda is called
}
```

This is the only Python-specific behaviour in the container registration path. The resolution path (`make()`) is uniform
with no conditionals. The service provider stays clean with no framework-specific lambda syntax.

---

## 7. Decorators — Metadata Markers, Not Self-Registrars

Python decorators execute at import time — but `@route_handler` must **not** self-register routes. It must be a
metadata marker only:

```python
def route_handler(handler_ref):
    def decorator(func):
        func._valkyrja_handler = handler_ref  # metadata only — no registration
        return func

    return decorator
```

**Why not self-registration:** The cache data file imports controller classes to reference them in route objects. Those
imports cannot be avoided. If `@route_handler` self-registered, loading from cache would cause double registration or
conflicting state — routes registered from cache AND from decorator execution on import.

**How it works without cache:** The framework scans controller classes for `_valkyrja_handler` metadata during
bootstrap. It reads the metadata and registers routes from it.

**How it works with cache:** The framework loads cache data files directly and never calls `get_controller_classes()` or
scans for `_valkyrja_handler`. Decorator metadata is never read.

The `@route_handler` decorator carries a reference to the handler method for build tool extraction. The build tool reads `_valkyrja_handler`
metadata from AST via `inspect.getfile()` + `ast.parse()`.

### Accessing _valkyrja_handler at Runtime

The framework reads the metadata from each method during bootstrap (no cache path):

```python
import inspect


def scan_controller_for_handlers(controller_class: type) -> list[dict]:
    """
    Scan a controller class for methods carrying _valkyrja_handler metadata.
    Called by the framework during bootstrap when no cache is loaded.
    Never called when loading from cache.
    """
    handlers = []

    for name, method in inspect.getmembers(controller_class, predicate=inspect.isfunction):
        if not hasattr(method, '_valkyrja_handler'):
            continue

        handler_ref = method._valkyrja_handler

        # @parameter decorator attaches parameter list similarly
        parameters = getattr(method, '_valkyrja_parameters', [])

        handlers.append({
            'method': name,
            'handler': handler_ref,
            'parameters': parameters,
        })

    return handlers
```

The `@parameter` decorator follows the same pattern:

```python
def parameter(name: str, pattern: str = '[^/]+'):
    def decorator(func):
        if not hasattr(func, '_valkyrja_parameters'):
            func._valkyrja_parameters = []
        func._valkyrja_parameters.append({'name': name, 'pattern': pattern})
        return func

    return decorator
```

Both attributes are simple lists/values attached directly to the function object — readable via `hasattr` / `getattr`
anywhere the function is accessible.

---

## 8. Deployment Models

### CGI / Lambda

- Cache optional in dev — full provider tree traversal at import
- Cache required for production — build tool generates `AppHttpRoutingData` etc.
- `sindri` Python implementation uses `ast` module + `inspect.getfile()`

### Worker (ASGI)

- Uvicorn / Hypercorn / Gunicorn+Uvicorn
- Single bootstrap per process — cache optional
- ASGI entrypoint:

```python
async def __call__(self, scope, receive, send):
    await self.dispatch(scope, receive, send)
```

### CGI and Worker mode

Framework supports both via different entry points. Developer writes application once:

```python
# CGI entry
from valkyrja import cgi

cgi.run(app)

# Worker entry
from valkyrja import worker

worker.run(app)
```

---

## 9. Build Tool — sindri (Python)

**Reference:** `SINDRI.md`

- Separate PyPI package: `sindri`
- Dev dependency only — never in production
- Uses `ast` (stdlib) + `inspect` (stdlib) — no external AST library needed
- `inspect.getfile(ClassName)` resolves any importable class to its source file
- `ast.parse()` + `ast.walk()` for provider tree and decorator extraction
- Generated files use `string.Template` for output

### Build tool flow

```
AppConfig → component providers
        ↓
inspect.getfile(ProviderClass) → source file path
        ↓
ast.parse(source) → AST
        ↓
collect_imports() → import map for FQN resolution
        ↓
extract @route_handler decorators + @parameter decorators
        ↓
resolve type references to FQN
        ↓
ProcessorContract for regex compilation
        ↓
string.Template → generate AppHttpRoutingData etc.
        ↓
deploy with generated files
```

---

## Priority Order

1. Container component
2. String constants per component + `class_()` helper
3. Throwable hierarchy — ABC abstract, three branches, ValueError root for InvalidArgument
4. Closure-based bindings
5. Provider contracts with proper type hints
6. The handler on the route contract — get_handler / with_handler
7. `@route_handler` and `@parameter` decorators
8. Route and listener data classes
9. CGI and ASGI entry points
10. sindri Python implementation
