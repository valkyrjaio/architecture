# Python TODO

> Full context: `README_PYTHON.md`, `CONTRACTS_PYTHON.md`

---

## Throwables

- [ ] Implement `ValkyrjaThrowable(BaseException, ABC)` — abstract
- [ ] Implement `ValkyrjaRuntimeException(RuntimeError, ABC)` — abstract
- [ ] Implement `ValkyrjaInvalidArgumentException(ValueError, ABC)` — abstract, parity name, extends `ValueError`
- [ ] Every component ships `ComponentRuntimeException` and `ComponentInvalidArgumentException` — abstract, always present
- [ ] Naming: `ComponentName*`, shared subcomponents `ParentComponentSubComponent*`, unique subcomponents `SubComponent*`
- [ ] Only concrete specific exceptions are thrown — never abstract base exceptions

---

## Container Bindings

- [ ] Add per-component constants files
    - [ ] `container/container_constants.py`
    - [ ] `http/http_constants.py`
    - [ ] `http/routing/http_routing_constants.py`
    - [ ] `cli/cli_constants.py`
    - [ ] `event/event_constants.py`
    - [ ] *(remaining components)*
- [ ] Add `class_()` FQN helper (trailing underscore — `class` is reserved)
- [ ] All bindings use closure-based factories — no reflection-based instantiation

---

## Provider Contracts

- [ ] Implement `ComponentProviderContract(ABC)` with `@staticmethod @abstractmethod` methods
- [ ] Implement `ServiceProviderContract(ABC)` with `publishers()` returning `dict[str, Callable]`
- [ ] Implement `HttpRouteProviderContract(ABC)` with `get_controller_classes() -> list[type]` + `get_routes() -> list[RouteContract]`
- [ ] Implement `CliRouteProviderContract(ABC)` with `get_controller_classes() -> list[type]` + `get_routes() -> list[RouteContract]`
- [ ] Implement `ListenerProviderContract(ABC)` with `get_listener_classes() -> list[type]` + `get_listeners() -> list[ListenerContract]`
- [ ] All provider list methods return simple list/dict literals — no conditional logic

---

## Handler Contracts

- [ ] Implement `@handler` decorator as **metadata marker only** — attaches `_valkyrja_handler` to method, does NOT self-register

```python
def handler(closure):
    def decorator(func):
        func._valkyrja_handler = closure  # metadata only
        return func
    return decorator
```

- [ ] Implement `@parameter` decorator — attaches `_valkyrja_parameters` list to method
- [ ] Define type aliases:
    - [ ] `HttpHandlerFunc = Callable[[ContainerContract, dict[str, Any]], ResponseContract]`
    - [ ] `CliHandlerFunc = Callable[[ContainerContract, dict[str, Any]], OutputContract]`
    - [ ] `ListenerHandlerFunc = Callable[[ContainerContract, dict[str, Any]], Any]`
- [ ] Implement `HttpHandlerContract(ABC)` with `get_handler() -> HttpHandlerFunc`
- [ ] Implement `CliHandlerContract(ABC)` with `get_handler() -> CliHandlerFunc`
- [ ] Implement `ListenerHandlerContract(ABC)` with `get_handler() -> ListenerHandlerFunc`
- [ ] Implement `HttpCacheableHandlerContract` extending `HttpHandlerContract`
- [ ] Implement `CliCacheableHandlerContract` extending `CliHandlerContract`
- [ ] Implement `ListenerCacheableHandlerContract` extending `ListenerHandlerContract`

---

## Bootstrap — Cache vs No Cache

- [ ] Without cache: framework scans controller classes for `_valkyrja_handler` metadata during bootstrap
- [ ] With cache: framework loads cache data files directly — never calls `get_controller_classes()`, never reads `_valkyrja_handler`
- [ ] Implement CGI entry point: `valkyrja.cgi.run(app)`
- [ ] Implement ASGI worker entry point: `valkyrja.worker.run(app)`

---

## Deployment

- [ ] ASGI entrypoint compatible with Uvicorn / Hypercorn / Gunicorn+Uvicorn
- [ ] CGI mode supported — cache optional in dev, required for production
- [ ] Granian (Rust-based) compatibility worth verifying

---

## valkyrja-build Python

- [ ] Create `valkyrja-build` as a separate PyPI package
- [ ] Dev dependency only — never in production
- [ ] Implement `inspect.getfile(ProviderClass)` for class-to-file resolution
- [ ] Implement `ast.parse()` + `ast.walk()` for provider tree walk
- [ ] Implement `collect_imports()` for FQN resolution map
- [ ] Implement `_valkyrja_handler` metadata extraction from AST
- [ ] Implement `_valkyrja_parameters` metadata extraction from AST
- [ ] Implement FQN rewriting via import map
- [ ] Implement `ProcessorContract` invocation for regex pre-compilation
- [ ] Generate `AppContainerData`
- [ ] Generate `AppEventData`
- [ ] Generate `AppHttpRoutingData`
- [ ] Generate `AppCliRoutingData`
- [ ] Move all file generation / scaffolding / `make:*` commands here
