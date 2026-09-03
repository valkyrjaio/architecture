# Java Port — Implementation Notes

> Reference docs: `THROWABLES.md`, `CONTAINER_BINDINGS.md`, `HANDLERS.md`,
> `DATA_CACHE.md`, `BUILD_TOOL.md`, `CONTRACTS_JAVA.md`
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

`.class` tokens are used as binding keys — compiler verified. Per-component
constants files are recommended alongside:

```java
// ContainerConstants.java
public final class ContainerConstants {
    public static final Class<RouterContract> ROUTER = RouterContract.class;

    private ContainerConstants() {
    }
}
```

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

**Reference:** `CONTRACTS_JAVA.md`, `DATA_CACHE.md`

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
    static List<Class<?>> getControllerClasses();

    static List<RouteContract> getRoutes();
}
```

### ListenerProviderContract

```java
public interface ListenerProviderContract {
    static List<Class<?>> getListenerClasses();

    static List<ListenerContract> getListeners();
}
```

All provider list methods must return simple `List.of()` literals — no
conditional logic.

---

## 4. Handler Contracts — Typed Closures

**Reference:** `HANDLERS.md`

### Three @FunctionalInterface types

```java
// HTTP
@FunctionalInterface
public interface HttpHandlerFunc {
    ResponseContract handle(ContainerContract container, Map<String, Object> arguments);
}

// CLI
@FunctionalInterface
public interface CliHandlerFunc {
    OutputContract handle(ContainerContract container, Map<String, Object> arguments);
}

// Event listener
@FunctionalInterface
public interface ListenerHandlerFunc {
    Object handle(ContainerContract container, Map<String, Object> arguments);
}
```

### Handler contracts per concern

```java
public interface HttpHandlerContract {
    HttpHandlerFunc getHandler();

    HttpHandlerContract setHandler(HttpHandlerFunc handler);
}
```

### @RouteHandler annotation on controller methods

```java
@RouteHandler((ContainerContract c, Map < String, Object > args) ->
        c.

getSingleton(UserController .class).

show(args.get("id")))

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

**Reference:** `BUILD_TOOL.md`

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

## 8. Build Tool — valkyrja-build Java

**Reference:** `BUILD_TOOL.md`

- Separate Maven/Gradle artifact: `io.valkyrja:build`
- Dev/test scope only — never in production
- Must publish `-sources.jar` as required build dependency for the build tool to
  read framework provider source files
- Handles project scaffolding, `make:*` commands, cache generation
- The annotation processor handles cache generation at compile time for
  application code
- Framework ships pre-generated cache files alongside compiled artifacts

---

## Priority Order

1. Container component (first per port order)
2. Throwable hierarchy — abstract, renamed, ComponentName* convention
3. Closure-based bindings + constants files
4. Provider contracts — ComponentProvider, ServiceProvider, RouteProvider,
   ListenerProvider
5. Handler functional interfaces — HttpHandlerFunc, CliHandlerFunc,
   ListenerHandlerFunc
6. Handler contracts per concern
7. @RouteHandler and @Parameter annotations
8. Records for data classes
9. Annotation processor setup + Trees API lambda extraction
10. JavaPoet cache data class generation
11. valkyrja-build Java artifact
