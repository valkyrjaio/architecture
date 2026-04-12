# Python Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `DISPATCH.md`,
`DATA_CACHE.md`, `BUILD_TOOL.md`, `CONTRACTS_PYTHON.md`
> Port order: Container → Dispatch → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Module namespace:** `valkyrja`
- **ABC** enforces abstract contracts — `TypeError` on direct instantiation
- **`@staticmethod @abstractmethod`** throughout — providers are stateless
- **`inspect.getfile()`** for class-to-file resolution (equivalent of PHP's
  `ReflectionClass::getFileName()`)
- **`ast` module** for build tool AST parsing
- **Decorators are runtime-executable** — `@handler` self-registers at import
  time
- **`class_()` helper** for FQN derivation (`class` is reserved in Python)
- **ASGI (Uvicorn/Hypercorn)** as the worker mode deployment model
- **CGI mode** supported — Python is interpreted, cache optional in dev
- **Granian** (Rust-based) as an emerging alternative for true multi-threaded
  workers
- GIL limits true thread parallelism — ASGI async is the idiomatic concurrency
  model
- Python 3.13+ free-threaded mode worth watching

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


class ComponentRuntimeException(ValkyrjaRuntimeException,
                                ABC): pass  # always present


class ComponentSpecificException(ComponentRuntimeException): pass  # concrete


# InvalidArgumentException branch — parity name, extends ValueError
class ValkyrjaInvalidArgumentException(ValueError, ABC): pass


class ComponentInvalidArgumentException(ValkyrjaInvalidArgumentException,
                                        ABC): pass


class ComponentSpecificInvalidArgumentException(
    ComponentInvalidArgumentException): pass
```

### Rules

- `ValkyrjaInvalidArgumentException` extends `ValueError` for language-level
  catchability
- All base and categorical exceptions are abstract via ABC
- Every component ships both categoricals even if unused
- Naming: `ComponentName*`, shared subcomponents `ParentComponentSubComponent*`
- `except ContainerNotFoundException as e:` — Python resolves top-to-bottom,
  specific first

---

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### String constants

```python
# container_constants.py
class ContainerConstants:
    CONTAINER = "io.valkyrja.container.ContainerContract"
    ROUTER = "io.valkyrja.http.routing.RouterContract"
    USER_REPOSITORY = "io.valkyrja.app.repositories.UserRepositoryContract"
```

### FQN helper

```python
# 'class' is reserved — use class_()
def class_(cls) -> str:
    return f"{cls.__module__}.{cls.__qualname__}"
```

### Closure-based bindings

```python
container.bind(
    RouterClass,
    lambda c: Router(c.make(DispatcherClass))
)

container.singleton(
    RouterClass,
    lambda c: Router(c.make(DispatcherClass))
)
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

Publisher methods carry `@handler` decorator — build tool reads decorator
argument from AST:

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
    container.set_singleton(UserRepositoryClass,
                            UserRepository(container.make(DatabaseClass)))
```

### HttpRouteProviderContract

```python
class HttpRouteProviderContract(ABC):
    @staticmethod @ abstractmethod
    def get_controller_classes() -> list[
        type]: ...  # classes with @handler decorators

    @staticmethod @ abstractmethod
    def get_routes() -> list[RouteContract]: ...
```

All provider methods must return simple list/dict literals — no conditional
logic.

---

## 4. Handler Contracts — Typed Callable Aliases

**Reference:** `DISPATCH.md`

### Three Callable type aliases

```python
from typing import Callable, Any

HttpHandlerFunc = Callable[
    [ContainerContract, dict[str, Any]], ResponseContract]
CliHandlerFunc = Callable[[ContainerContract, dict[str, Any]], OutputContract]
ListenerHandlerFunc = Callable[[ContainerContract, dict[str, Any]], Any]
```

### Handler contracts per concern

```python
class HttpHandlerContract(ABC):
    @abstractmethod
    def get_handler(self) -> HttpHandlerFunc: ...

    @abstractmethod
    def set_handler(self,
                    handler: HttpHandlerFunc) -> 'HttpHandlerContract': ...
```

### @handler decorator on controller methods

```python
@handler(lambda c, args: c.get_singleton(UserControllerClass).show(args['id']))
@parameter('id', pattern='[0-9]+')
def show(self, id: int) -> ResponseContract:
    pass
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from
container if needed.

---

## 5. Decorators — Metadata Markers, Not Self-Registrars

Python decorators execute at import time — but `@handler` must **not**
self-register routes. It must be a metadata marker only:

```python
def handler(closure):
    def decorator(func):
        func._valkyrja_handler = closure  # metadata only — no registration
        return func

    return decorator
```

**Why not self-registration:** The cache data file imports controller classes to
reference them in route objects. Those imports cannot be avoided. If `@handler`
self-registered, loading from cache would cause double registration or
conflicting state — routes registered from cache AND from decorator execution on
import.

**How it works without cache:** The framework scans controller classes for
`_valkyrja_handler` metadata during bootstrap. It reads the metadata and
registers routes from it.

**How it works with cache:** The framework loads cache data files directly and
never calls `get_controller_classes()` or scans for `_valkyrja_handler`.
Decorator metadata is never read.

The `@handler` decorator carries the closure for build tool extraction. The
build tool reads `_valkyrja_handler` metadata from AST via `inspect.getfile()` +
`ast.parse()`.

---

## 6. Deployment Models

### CGI / Lambda

- Cache optional in dev — full provider tree traversal at import
- Cache required for production — build tool generates `AppHttpRoutingData` etc.
- `valkyrja-build` Python implementation uses `ast` module + `inspect.getfile()`

### Worker (ASGI)

- Uvicorn / Hypercorn / Gunicorn+Uvicorn
- Single bootstrap per process — cache optional
- ASGI entrypoint:

```python
async def __call__(self, scope, receive, send):
    await self.dispatch(scope, receive, send)
```

### CGI and Worker mode

Framework supports both via different entry points. Developer writes application
once:

```python
# CGI entry
from valkyrja import cgi

cgi.run(app)

# Worker entry
from valkyrja import worker

worker.run(app)
```

---

## 7. Build Tool — valkyrja-build Python

**Reference:** `BUILD_TOOL.md`

- Separate PyPI package: `valkyrja-build`
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
3. Throwable hierarchy — ABC abstract, three branches, ValueError root for
   InvalidArgument
4. Closure-based bindings
5. Provider contracts with proper type hints
6. Callable type aliases — HttpHandlerFunc, CliHandlerFunc, ListenerHandlerFunc
7. Handler contracts per concern
8. `@handler` and `@parameter` decorators
9. Route and listener data classes
10. CGI and ASGI entry points
11. Dispatch component
12. valkyrja-build Python implementation
