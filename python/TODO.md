# Python TODO

> Full context: `README_PYTHON.md`, `CONTRACTS_PYTHON.md`

---

## Cross-language testing-gap audit

Compare this port's test suite against the other languages' and either close each
difference or record it in `AGENTS.md` as deliberate. Out of scope for the work
that prompted it; tracked here so it is not lost.

Prompted by a concrete miss: `sindri-java`'s golden snapshot never exercised the
dynamic-route regex path, so a framework regex-format change rode through a
dependency bump silently, while `sindri-ts` caught the equivalent change at once
([sindri-java#54](https://github.com/valkyrjaio/sindri-java/pull/54)). Only that
single gap was checked across ports — nobody has compared the suites broadly.

What to look for:

- Behavior a sibling port asserts that this one does not.
- An assertion pinned to a fragment where a sibling pins the whole value — a
  fragment survives the framing around it changing, which is exactly how the miss
  above happened.
- A generator, adapter, or component carrying snapshot/branch coverage on one
  side and none here.
- Test tooling that differs: what the static analyzers and formatters actually
  cover, and whether the test tree is inside or outside that scope.

Not every difference is a defect — some are forced by the language. Read the
per-language notes in `AGENTS.md` before "aligning" anything; the dynamic route
regex framing is the worked example (PHP must keep its PCRE delimiters, every
other port must not).

Known starting point: the port does not exist yet beyond `template`. Do this as the
port is written — per the canonical rule that code and tests land together — rather
than as a later pass over a finished suite.

## Reset the reused deps branch in the update-dependencies workflow

`_python-update-dependencies.yml` reuses an open `deps/update-dependencies-*` branch across runs,
committing each run's updates on top of the previous run's tree. Nothing a run writes to that
branch can be walked back by a later run, so a bad version pinned once stays pinned until
somebody deletes the branch by hand.

Java hit this concretely. An unfiltered root `dependencyUpdates` report bumped
`io.netty:netty-codec-http` to the 2015-era `5.0.0.Alpha2` prerelease and turned CI red
([starter-app-java#53](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/53)). Fixing
the filter ([#54](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/54)) did **not**
repair the branch: `useLatestVersions` only ever upgrades, so the re-run reported the alpha as
merely "exceeding" the latest and changed nothing. The PR had to be closed and the branch deleted.
The Java workflow now resets the reused branch to the base commit before running the updates
([.github#151](https://github.com/valkyrjaio/.github/pull/151)), so each run recomputes from
scratch and the branch always holds exactly `base + today's updates`.

Confirm first that this port's updater can actually strand a bad version the way
`useLatestVersions` does — a tool that rewrites every constraint to "latest" on each run may
already self-correct, in which case record that in `AGENTS.md` instead of adding the reset.

Tradeoff to weigh: the reset discards any commit pushed onto the deps branch by hand.

## High priority — name test fixtures `fixtures`, not `classes`

**Cross-language change — mirror this in every port (Go, Java, PHP, TypeScript)
so the test trees stay 1:1.** Reusable test doubles/sample classes live under a
`fixtures` package/dir, **not** `classes`. "Fixtures" is the widely-understood
term; "classes" is generic and reads oddly next to `unit`/`functional`. Python is
not ported yet, so build it this way from the start (no rename needed):

- [ ] Put reusable doubles under `tests/fixtures/` (mirroring PHP's `Fixtures`
      subdivisions: `provider`, `contract`, …), never `tests/classes/`.
- [ ] Decide whether the `*Class` suffix convention also becomes `*Fixture`
      (preferred for full parity) — pick one and apply it everywhere.

---

## Throwables

- [ ] Implement `ValkyrjaThrowable(BaseException, ABC)` — abstract
- [ ] Implement `ValkyrjaRuntimeException(RuntimeError, ABC)` — abstract
- [ ] Implement `ValkyrjaInvalidArgumentException(ValueError, ABC)` — abstract, parity name, extends `ValueError`
- [ ] Every component ships `ComponentRuntimeException` and `ComponentInvalidArgumentException` — abstract, always
  present
- [ ] Naming: `ComponentName*`, shared subcomponents `ParentComponentSubComponent*`, unique subcomponents
  `SubComponent*`
- [ ] Only concrete specific exceptions are thrown — never abstract base exceptions

---

## Container Bindings

- [ ] Add per-component string constants files (required — same as Go and TypeScript)
    - [ ] `container/container_constants.py`
    - [ ] `http/http_constants.py`
    - [ ] `http/routing/http_routing_constants.py`
    - [ ] `cli/cli_constants.py`
    - [ ] `event/event_constants.py`
    - [ ] *(remaining components)*
- [ ] Add `class_()` FQN helper (trailing underscore — `class` is reserved)
- [ ] All bindings use string constant keys and closure-based factories

```python
# correct — string constant as key, no class object import forced
container.bind(
    ContainerConstants.USER_REPOSITORY,
    lambda c: UserRepository(c.make(ContainerConstants.DATABASE))
)
```

---

## Provider Contracts

- [ ] Implement `ComponentProviderContract(ABC)` with `@staticmethod @abstractmethod` methods
- [ ] Implement `ServiceProviderContract(ABC)` with `publishers()` returning `dict[str, Callable]`
- [ ] Implement `HttpRouteProviderContract(ABC)` with `get_controller_classes() -> list[type]` +
  `get_routes() -> list[RouteContract]`
- [ ] Implement `CliRouteProviderContract(ABC)` with `get_controller_classes() -> list[type]` +
  `get_routes() -> list[RouteContract]`
- [ ] Implement `ListenerProviderContract(ABC)` with `get_listener_classes() -> list[type]` +
  `get_listeners() -> list[ListenerContract]`
- [ ] All provider list methods return simple list/dict literals — no conditional logic

---

## Handler Contracts

- [ ] Implement `@handler` decorator as **metadata marker only** — attaches `_valkyrja_handler` to method, does NOT
  self-register

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

## Python 3.14 Lazy Imports — Track for Future Optimisation

Python eagerly imports everything — this is a language characteristic, not a framework problem.
No action at the framework level. Track the following:

- [ ] Monitor Python 3.14 lazy imports feature for stable release
- [ ] Test Valkyrja Python port compatibility with Python 3.14 lazy imports when available
- [ ] If compatible — document as an optional cold start optimisation for Python 3.14+ deployments
- [ ] No framework changes needed — lazy imports would be a Python runtime feature

---

## Bootstrap — Cache vs No Cache

- [ ] Without cache: framework scans controller classes for `_valkyrja_handler` metadata during bootstrap

```python
# framework bootstrap — reads metadata from each method
for name, method in inspect.getmembers(controller_class, predicate=inspect.isfunction):
    if hasattr(method, '_valkyrja_handler'):
        closure = method._valkyrja_handler
        parameters = getattr(method, '_valkyrja_parameters', [])
        # register route from closure + parameters
```

- [ ] With cache: framework loads cache data files directly — never calls `get_controller_classes()`, never scans
  `_valkyrja_handler`
- [ ] Implement CGI entry point: `valkyrja.cgi.run(app)`
- [ ] Implement ASGI worker entry point: `valkyrja.worker.run(app)`

---

## Deployment

- [ ] ASGI entrypoint compatible with Uvicorn / Hypercorn / Gunicorn+Uvicorn
- [ ] CGI mode supported — cache optional in dev, required for production
- [ ] Granian (Rust-based) compatibility worth verifying

---

## valkyrja-forge Python

- [ ] Create `valkyrja-forge` as a separate PyPI package
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

---

## VLID — cross-language parity

**Cross-language change — mirror in every port (Go, Java, PHP, TypeScript).** VLID
(`Type/Vlid`) is PHP-only today; port it here (code + tests). It is the source of the
queue envelope `id` (a **VLID V1** — the longest, most-random version). Lock
cross-language parity:

- [ ] Port `Type/Vlid`, then add a conformance test: generate a VLID for **each
      version V1–V4** from a **fixed input timestamp**.
- [ ] Assert this port produces a byte-identical **non-random portion** vs the PHP
      fixture — the encoded **microsecond timestamp** and the **version digit at
      position 14** must match exactly. The random bits differ by design; exclude them.
- [ ] This gate prevents timestamp-encoding / version-digit-placement drift from
      silently breaking cross-language `id` interop.
