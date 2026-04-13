# Python Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `DISPATCH.md`, `DATA_CACHE.md`, `BUILD_TOOL.md`,
`CONTRACTS_PYTHON.md`
> Port order: Container → Dispatch → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Module namespace:** `valkyrja`
- **ABC** enforces abstract contracts — `TypeError` on direct instantiation
- **`@staticmethod @abstractmethod`** throughout — providers are stateless
- **`inspect.getfile()`** for class-to-file resolution (equivalent of PHP's `ReflectionClass::getFileName()`)
- **`ast` module** for build tool AST parsing
- **Decorators are runtime-executable** — `@handler` self-registers at import time
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
imported at key definition time — the class object cannot exist without its module loading. This defeats Python 3.14's
lazy import mechanism which Valkyrja relies on for cold start performance.

```python
# container_constants.py — required per component
class ContainerConstants:
    CONTAINER = "io.valkyrja.container.ContainerContract"
    ROUTER = "io.valkyrja.http.routing.RouterContract"
    USER_REPOSITORY = "app.repositories.UserRepositoryContract"
    DATABASE = "app.services.DatabaseContract"
```

```python
# bind and resolve via string constant
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.make(ContainerConstants.DATABASE))
)

repo = container.make(ContainerConstants.USER_REPOSITORY)
# with Python 3.14 lazy imports, UserRepository loads here — not at boot
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

**Reference:** `CONTRACTS_PYTHON.md`, `DATA_CACHE.md`

### Type hints

```python
get_container_providers() -> list[type]  # class objects
get_event_providers()     -> list[type]
get_cli_providers()       -> list[type]
get_http_providers()      -> list[type]
get_controller_classes()  -> list[type]  # classes with @handler decorators
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

Publisher methods carry `@handler` decorator — build tool reads decorator argument from AST:

```python
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
    container.set_singleton(UserRepositoryClass, UserRepository(container.make(DatabaseClass)))
```

### HttpRouteProviderContract

```python
class HttpRouteProviderContract(ABC):
    @staticmethod @ abstractmethod
    def get_controller_classes() -> list[type]: ...  # classes with @handler decorators

    @staticmethod @ abstractmethod
    def get_routes() -> list[RouteContract]: ...
```

All provider methods must return simple list/dict literals — no conditional logic.

---

## 4. Handler Contracts — Typed Callable Aliases

**Reference:** `DISPATCH.md`

### Three Callable type aliases

```python
from typing import Callable, Any

HttpHandlerFunc = Callable[[ContainerContract, dict[str, Any]], ResponseContract]
CliHandlerFunc = Callable[[ContainerContract, dict[str, Any]], OutputContract]
ListenerHandlerFunc = Callable[[ContainerContract, dict[str, Any]], Any]
```

### Handler contracts per concern

```python
class HttpHandlerContract(ABC):
    @abstractmethod
    def get_handler(self) -> HttpHandlerFunc: ...

    @abstractmethod
    def set_handler(self, handler: HttpHandlerFunc) -> 'HttpHandlerContract': ...
```

### @handler decorator on controller methods

```python
@handler(lambda c, args: c.get_singleton(UserControllerClass).show(args['id']))
@parameter('id', pattern='[0-9]+')
def show(self, id: int) -> ResponseContract:
    pass
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from container if needed.

---

## 5. Python Imports and Cold Start — Python 3.14 Is the Answer

Eager imports are a well-known Python ecosystem problem. Every module referenced in the import chain loads at boot —
every controller, listener, and service class regardless of whether the current request needs them. FastAPI, Django, and
Flask all have this problem. Real-world applications on AWS Lambda have reported 10–30 second cold starts for large
applications.

**This is a language-level problem and Python 3.14 solves it natively.**

Valkyrja's Python port requires Python 3.14 minimum. This is a deliberate decision — building on 3.14 means lazy imports
are available as a first-class language feature with no framework workarounds needed.

### Python 3.14 Lazy Imports — What They Actually Do

Python 3.14 introduces transparent lazy imports — module execution is deferred until the imported name is **first
accessed**. The key word is "accessed". Constructing a tuple, dict, or any other data structure that references an
imported name counts as accessing it.

```python
from app.services import UserServiceProvider  # lazy proxy — not yet loaded

# tuple construction accesses UserServiceProvider — triggers the import HERE
(UserServiceProvider, 'publish_user')  # ← module loads at this line
```

This means a plain tuple callable in the cache data file loads the provider module at cache file load time — even with
Python 3.14 lazy imports. The lambda wrapper is still required for Python to achieve true deferred loading:

```python
# without lambda — UserServiceProvider loads when cache file loads
APP_CONTAINER_DATA = {
    'app.repositories.UserRepositoryContract': (UserServiceProvider, 'publish_user_repository'),
}

# with lambda — UserServiceProvider loads only when lambda is called
# lambda body is not evaluated until called — name access deferred
APP_CONTAINER_DATA = {
    'app.repositories.UserRepositoryContract': lambda: (UserServiceProvider, 'publish_user_repository'),
}
```

The lambda wrapper is **Python-only**. PHP's `[SomeClass::class, 'method']` uses `::class` which resolves to a plain
string at compile time — no class loading. Compiled languages (Java, Go, TypeScript) have no equivalent concept.

### String Keys for Container Bindings

Because class object keys require the class to be imported (the class object cannot exist without its module loading),
Python uses **string keys** for container bindings — same as Go and TypeScript:

```python
# class object key — forces UserRepositoryContract import at cache load time
APP_CONTAINER_DATA = {
    UserRepositoryContract: lambda: (UserServiceProvider, 'publish_user_repository'),
    # UserRepositoryContract accessed here — module loads
}

# string key — no import of contract class at cache load time
APP_CONTAINER_DATA = {
    'app.repositories.UserRepositoryContract': lambda: (UserServiceProvider, 'publish_user_repository'),
    # string literal — no import triggered
}
```

With string keys and lambda wrappers:

- Boot: cache file loads — no provider or contract modules imported
- First resolution of a binding: lambda called — that provider module loads only then
- Each request: only the providers needed for that request's bindings load

### What the Cache Provides

Even in pre-3.14 Python, the cache eliminates meaningful boot cost beyond imports:

```
Without cache — every boot:
  ✗ Traverse provider tree
  ✗ Scan @handler decorators across all controllers  
  ✗ Build route dispatcher index (regex compilation, path indexing)
  ✗ Register all container bindings
  + Module imports (unavoidable in pre-3.14)

With cache — every boot:
  ✓ Load four pre-built data classes
  ✓ Skip provider tree traversal entirely
  ✓ Skip annotation scanning entirely
  ✓ Skip route index construction entirely
  + Module imports (deferred per-request with Python 3.14)
```

### The Multi-Language Escape Valve

For Lambda-heavy workloads where Python cold starts remain a concern even with 3.14, Valkyrja's multi-language nature is
the answer — switch the Lambda function to the Go or TypeScript port. Same framework architecture, same patterns, same
team knowledge, compiled binary cold start times. This is a genuine differentiator of the Valkyrja ecosystem over
single-language frameworks.

---

## 6. Container — Uniform Lambda Format

The container's internal bindings map always stores lambdas — whether populated from a service provider at runtime or
loaded from a cache data file. This makes the internal format identical in both paths, and resolution always calls the
lambda with no conditional logic.

### Three Parties, One Job Each

**Service provider** — returns a plain method reference map. No lambda, no framework concerns:

```python
class UserServiceProvider(ServiceProviderContract):
    @staticmethod
    def publishers() -> dict:
        return {
            'app.repositories.UserRepositoryContract': UserServiceProvider.publish_user_repository,
        }
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

    def make(self, key: str):
        # always call the lambda — uniform, no conditional check needed
        callable_ref = self._bindings[key]()
        return callable_ref(self)

    def singleton(self, key: str):
        if key not in self._singletons:
            self._singletons[key] = self.make(key)
        return self._singletons[key]
```

**Forge** — reads the plain method reference from `publishers()` AST and writes it as a lambda into the generated cache
file. Cache matches the container's internal format exactly:

```python
# generated AppContainerData — lambda format, matches container internal map
APP_CONTAINER_DATA = {
    'app.repositories.UserRepositoryContract': lambda: UserServiceProvider.publish_user_repository,
}
```

This is the only Python-specific behaviour in the container registration path. The resolution path (`make()`) is uniform
with no conditionals. The service provider stays clean with no framework-specific lambda syntax.

---

## 7. Decorators — Metadata Markers, Not Self-Registrars

Python decorators execute at import time — but `@handler` must **not** self-register routes. It must be a metadata
marker only:

```python
def handler(closure):
    def decorator(func):
        func._valkyrja_handler = closure  # metadata only — no registration
        return func

    return decorator
```

**Why not self-registration:** The cache data file imports controller classes to reference them in route objects. Those
imports cannot be avoided. If `@handler` self-registered, loading from cache would cause double registration or
conflicting state — routes registered from cache AND from decorator execution on import.

**How it works without cache:** The framework scans controller classes for `_valkyrja_handler` metadata during
bootstrap. It reads the metadata and registers routes from it.

**How it works with cache:** The framework loads cache data files directly and never calls `get_controller_classes()` or
scans for `_valkyrja_handler`. Decorator metadata is never read.

The `@handler` decorator carries the closure for build tool extraction. The build tool reads `_valkyrja_handler`
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

        handler_closure = method._valkyrja_handler

        # @parameter decorator attaches parameter list similarly
        parameters = getattr(method, '_valkyrja_parameters', [])

        handlers.append({
            'method': name,
            'handler': handler_closure,
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
- `valkyrja-forge` Python implementation uses `ast` module + `inspect.getfile()`

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

## 9. Build Tool — valkyrja-forge Python — valkyrja-forge Python

**Reference:** `BUILD_TOOL.md`

- Separate PyPI package: `valkyrja-forge`
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
extract @handler decorators + @parameter decorators
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
6. Callable type aliases — HttpHandlerFunc, CliHandlerFunc, ListenerHandlerFunc
7. Handler contracts per concern
8. `@handler` and `@parameter` decorators
9. Route and listener data classes
10. CGI and ASGI entry points
11. Dispatch component
12. valkyrja-forge Python implementation
