# Java Port — Implementation Notes

> Reference docs, in `convention/`: `THROWABLES.md`, `CONTAINER_BINDINGS.md`,
> `HANDLERS.md`, `DATA_CACHE.md`, `SINDRI.md`, `CONTRACTS.md`
> Port order: Container → Event → Application → CLI → HTTP → Bin

---

## Key Language Decisions

- **Package namespace:** `io.valkyrja`
- **Build tool:** Gradle
- **Records** for data classes (cache data, route data, etc.)
- **`Function<Container, ?>`** lambdas for deferred bindings
- **`@Provides` annotation** with `RetentionPolicy.RUNTIME`
- **Annotation processor + JavaPoet** for cache data class generation
- **Java's built-in `HttpServer`** as zero-dependency default
- **Build toolchain:** Spotless, ArchUnit, ErrorProne + NullAway, JUnit 5
- **Project Loom virtual threads** for concurrency
- All Valkyrja exceptions extend `RuntimeException` (unchecked) — no `throws`
  declarations

---

## 1. Throwables

**Reference:** `THROWABLES.md`

### Hierarchy

```
java.lang.Throwable
└── ValkyrjaThrowable (abstract)
    └── ComponentThrowable (abstract · always present)
        └── ComponentSpecificThrowable (concrete)

java.lang.RuntimeException
└── ValkyrjaRuntimeException (abstract)
    └── ComponentRuntimeException (abstract · always present)
        └── ComponentSpecificException (concrete)

java.lang.IllegalArgumentException   ← Java has no InvalidArgumentException
└── ValkyrjaInvalidArgumentException  ← parity name, extends IllegalArgumentException
    └── ComponentInvalidArgumentException (abstract · always present)
        └── ComponentSpecificInvalidArgumentException (concrete)
```

### Rules

- `ValkyrjaInvalidArgumentException` extends
  `java.lang.IllegalArgumentException` for language-level catchability while
  preserving cross-port naming parity
- All base and categorical exceptions are `abstract`
- Every component ships `ComponentRuntimeException` and
  `ComponentInvalidArgumentException` even if unused
- Shared subcomponents: `HttpRoutingRuntimeException`,
  `CliRoutingRuntimeException` etc.
- Unique subcomponents: `RequestRuntimeException`, `ResponseRuntimeException`
  etc.
- Spotless will flag same-named exceptions across packages — `ComponentName*`
  prefix resolves this

---

## 2. Container Bindings

**Reference:** `CONTAINER_BINDINGS.md`

### Class references

`.class` tokens are used as binding keys — compiler verified. A language that
names a class natively needs no constants file, so Java ships none: the token is
already the key, and the compiler already checks it. See
[`CONTAINER_BINDINGS.md`](../../convention/CONTAINER_BINDINGS.md) for the ports
that do need one, and why.

### Closure-based bindings

All bindings use lambda factories — no reflection-based instantiation:

```java
container.bind(
    RouterContract.class,
    c -> new Router(c.getSingleton(MatcherContract.class))
);

container.singleton(
    RouterContract.class,
    c -> new Router(c.getSingleton(MatcherContract.class))
);
```

---

## 3. Provider Contracts

**Reference:** `PROVIDER_CONTRACTS.md`, and `DATA_CACHE.md` in `convention/`

### ComponentProviderContract

```java
public interface ComponentProviderContract {
    List<ServiceProviderContract> getContainerProviders(ApplicationContract app);

    List<ListenerProviderContract> getEventProviders(ApplicationContract app);

    List<CliRouteProviderContract> getCliProviders(ApplicationContract app);

    List<HttpRouteProviderContract> getHttpProviders(ApplicationContract app);
}
```

### ServiceProviderContract

```java
public interface ServiceProviderContract {
    Map<Class<?>, Consumer<ContainerContract>> publishers();
}
```

`publishers()` returns a map of `.class` token to static method reference. No
`@RouteHandler` annotation on publisher methods — build tool reads method bodies
directly from AST via Trees API.

### HttpRouteProviderContract / CliRouteProviderContract

```java
public interface HttpRouteProviderContract {
    List<Class<?>> getControllerClasses();

    List<RouteContract> getRoutes();
}
```

### ListenerProviderContract

```java
public interface ListenerProviderContract {
    List<Class<?>> getListenerClasses();

    List<ListenerContract> getListeners();
}
```

All provider list methods must return simple `List.of()` literals — no
conditional logic.

---

## 4. Handler Contracts — Typed Closures

**Reference:** `HANDLERS.md`

### The handler types

A route handler takes the container and the **route**. Only a listener takes a map,
because a listener has no route. See
[`HANDLERS.md`](../../convention/HANDLERS.md).

```java
// HTTP route
BiFunction<ContainerContract, RouteContract, ResponseContract>

// CLI route
BiFunction<ContainerContract, RouteContract, OutputContract>

// gRPC route
BiFunction<ContainerContract, RouteContract, ServiceResponseContract>

// Event listener
BiFunction<ContainerContract, Map<String, Object>, Object>
```

### The handler lives on the route contract

There is no handler contract per concern. The route contract declares
`getHandler()` and `withHandler()`, and `withHandler` returns a copy rather than
modifying the route, per
[`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md).

### @RouteHandler annotation on controller methods

```java
@RouteHandler((ContainerContract c, RouteContract route) ->
        c.getSingleton(UserController.class).show(route))

@Parameter(name = "id", pattern = "[0-9]+")
public ResponseContract show(String id) {
}
```

`ServerRequestContract` and `RouteContract` are not parameters — fetch from
container if needed.

---

## 5. Records for Data Classes

Two distinct record styles are used depending on context:

### Framework records — Option A (components + compact constructor)

Framework data records carry runtime state and may be instantiated with varying
values. Each record component takes the name of the interface method it
satisfies, so the compiler generates the accessor. A compact constructor copies
each collection, so a caller cannot mutate the record through the reference it
passed in. A no-arg constructor delegates to the canonical constructor and
defaults every component to an empty collection.

```java
public record NotifierData(
        Map<String, Supplier<ChannelContract>> channels,
        Map<String, String> defaults)
        implements NotifierDataContract {

    public NotifierData {
        channels = Map.copyOf(channels);
        defaults = Map.copyOf(defaults);
    }

    public NotifierData() {
        this(Map.of(), Map.of());
    }
}
```

### Generated/app records — Option B (no components, explicit overrides)

App-level data records (generated by Sindri) declare static, compile-time-known
data. They carry no state — data lives in the method bodies. This maps directly
to what Sindri generates: method bodies returning populated literal maps.

```java
public record AppHttpRoutingData() implements HttpRoutingDataContract {

    @Override
    public Map<String, Supplier<RouteContract>> routes() {
        return Map.of(
                "GET/",    RouteProvider::home,
                "GET/about", RouteProvider::about
        );
    }

    @Override
    public Map<String, Map<String, String>> paths() {
        return Map.of();
    }

    @Override
    public Map<String, Map<String, String>> dynamicPaths() {
        return Map.of();
    }

    @Override
    public Map<String, Map<String, String>> regexes() {
        return Map.of();
    }
}
```

The starter app versions of these files use `Map.of()` for all methods; Sindri
replaces the method bodies with populated maps during cache generation.

---

## 6. Annotation Processor — Cache Generation

**Reference:** `SINDRI.md`

The annotation processor runs during `javac` — no separate build step needed.

### Setup

```java

@SupportedAnnotationTypes("io.valkyrja.http.routing.Handler")
@SupportedSourceVersion(SourceVersion.RELEASE_21)
public class ValkyrjaAnnotationProcessor extends AbstractProcessor {
    private Trees trees;

    @Override
    public synchronized void init(ProcessingEnvironment env) {
        super.init(env);
        this.trees = Trees.instance(env);
    }
}
```

### Lambda extraction via Trees API

The Trees API gives access to lambda source text from the AST at compile time.
FQN resolution is automatic via the compilation unit's import list.

### Code generation via JavaPoet

Generated cache data records are written via JavaPoet during annotation
processing — compiled in the same `javac` pass as application source.

### valkyrja.yaml

The annotation processor reads the application config class to discover the full
provider tree, then walks each provider's source file via Trees API.

---

## 7. Exception Handling Notes

- No `catch (Exception e)` — always catch specific Valkyrja exceptions
- Never declare `throws` on methods — all exceptions extend `RuntimeException`
- `errors.As` equivalent is `instanceof` in catch blocks
- `ValkyrjaInvalidArgumentException` catches at `IllegalArgumentException` level

---

## 8. Build Tool — `sindri` Java

**Reference:** `SINDRI.md`

- Separate Maven/Gradle artifact: `io.valkyrja:sindri`
- Dev/test scope only — never in production
- Must publish `-sources.jar` as required build dependency for the build tool to
  read framework provider source files
- Generates the cache data classes
- The annotation processor handles cache generation at compile time for
  application code
- Framework ships pre-generated cache files alongside compiled artifacts

---

## Priority Order

1. Container component (first per port order)
2. Throwable hierarchy — abstract, renamed, ComponentName* convention
3. Explicit binding factories
4. Provider contracts — ComponentProvider, ServiceProvider, RouteProvider,
   ListenerProvider
5. The handler on the route contract — getHandler / withHandler
6. @RouteHandler and @Parameter annotations
7. Records for data classes
8. Annotation processor setup + Trees API lambda extraction
9. JavaPoet cache data class generation
10. `sindri` Java artifact
