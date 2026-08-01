# Java

```
./.github/ci/spotless/gradlew spotlessApply   # auto-format
./.github/ci/spotless/gradlew    # check without modifying

./gradlew spotlessCheck   # check formatting
./gradlew spotlessApply   # apply formatting
./gradlew archunit        # architecture tests
./gradlew errorprone      # static analysis
./gradlew spotbugs        # bug detection
./gradlew junit           # unit testing
./gradlew ci              # run all checks
```

## TODOs

### Missing badges

Only the Scrutinizer badge is absent, and Scrutinizer may not apply to a Java port.

- [valkyrjaio/valkyrja-java#109](https://github.com/valkyrjaio/valkyrja-java/issues/109)

### Update the repo description once gRPC and Queue land

The About text is wrong today in a way this file did not record: it names PHP, not
Java, and it lists three protocols rather than four.

- [valkyrjaio/valkyrja-java#107](https://github.com/valkyrjaio/valkyrja-java/issues/107)

### Cross-language testing-gap audit

The two known `sindri-java` gaps are closed. The rest of the suite has never been
compared against PHP's.

- [valkyrjaio/valkyrja-java#106](https://github.com/valkyrjaio/valkyrja-java/issues/106)

### Root build rewrites the standalone CI builds (starter app)

The filter gap is fixed. The coupling remains: one root-level miss reaches all six
standalone CI builds at once.

- [valkyrjaio/valkyrja-starter-app-java#88](https://github.com/valkyrjaio/valkyrja-starter-app-java/issues/88)

### Response cache rework — port from PHP

PHP serializes the response to JSON. This port still generates and loads a source
file.

- [valkyrjaio/valkyrja-java#102](https://github.com/valkyrjaio/valkyrja-java/issues/102)

### Incomplete ports (PHP → Java)

These exist in the PHP framework but are not yet ported to Java. Test coverage
currently targets only the code that exists in the Java source; finish the port
(with tests) to reach parity.

- **Event** — the annotations, the collector, the service provider, and the concrete
  throwables are missing.
  [valkyrjaio/valkyrja-java#103](https://github.com/valkyrjaio/valkyrja-java/issues/103)

### Test-port status

Every module's unit tests are ported and JaCoCo line coverage is **100%**
(6798/6798) — every reachable source line is covered.

**Done — one dedicated test file per code-bearing class.** PHP keeps a separate
test file per class (e.g. `ParsedBodyParamCollectionTest`, `TextResponseTest`).
All **201 code-bearing classes** that lacked one (concrete, abstract, the 116
exceptions — each in its own file; the grouped `ExceptionsTest` is a PHP
anti-pattern and is not used here — enums, records, default-method interfaces)
now have a dedicated `<Class>Test` covering 100% of that file, and the former
grouped tests (`ConcreteParamCollectionsTest`, `TypedResponsesTest`,
`TypedHeadersTest`, `FormatterVariantsTest`, `MessageVariantsTest`,
`OutputVariantsTest`, `OptionParameterSubclassesTest`, the per-module
`*ExceptionTest`s) were split out and removed. Suite: 1629 tests / 435 files.

**Deferred — no-bytecode classes.** The **146 pure interfaces** and the **21
annotation markers** have no executable code for JaCoCo to measure, so neither group
has a test file.
[valkyrjaio/valkyrja-java#110](https://github.com/valkyrjaio/valkyrja-java/issues/110)

Note: Java static methods are not polymorphic, so PHP's static-override test
fixtures (`WorkerHttpClass`, `CliClass`, `AppExceptionHandlerClass`) do not
translate — cover those paths by calling the static methods directly and mocking
the resolved handler.

### Port bugs found during testing

- **`CliConfig` default providers** — Java defaulted to
  `ApplicationComponentProvider` (only Container+Event+Application), so
  `InputHandlerContract` never resolved. PHP defaults to
  `CliWithHttpApplicationComponentProvider`. Fixed to match PHP.
- **`InputHandler.run` exit** — Java called `System.exit(code)` directly,
  bypassing the `Exiter` freeze/unfreeze seam (which exists precisely so tests
  can suppress the exit). PHP calls `Exiter::exit($code)`. Fixed to call
  `Exiter.exit(code)`; CLI entry tests now `Exiter.freeze()` around `run()`.
- **Route collection never loaded the generated data** —
  `HttpRoutingServiceProvider.publishRouteCollection` never called
  `RouteCollection.setFromData` (nothing in the framework did), and gated attribute
  collection on `isSingleton(RouteCollectorContract)`, which is always false because
  `Container.setFromData` registers generated entries as deferred *callbacks*. Neither
  route source was ever used outside debug mode, so an application booted with **zero
  routes and answered every request with a 404**. Fixed to mirror PHP: debug mode
  collects from the providers, otherwise the routes load from the generated routing
  data, and `publishData` doubles as the `HttpRoutingDataContract` publisher so an app
  with no generated cache still resolves a collected set.
- **Log component never ported** — nothing published `LoggerContract`, yet
  `HttpServerServiceProvider` resolves it to build `LogThrowableCaughtMiddleware`, so
  any request reaching the throwable-caught path died with an unresolved-service
  error *inside* `RequestHandler.handle`'s catch block — replacing the original error
  and closing the connection with no response at all. Ported PHP's
  `LogComponentProvider` / `LogServiceProvider` and wired them into the HTTP and CLI
  application graphs. PHP defaults the logger to Monolog; Java has no Monolog or PSR
  and must not take a required logging dependency, so the default is a
  zero-dependency `FileLogger` writing a dated file to `Directory.logsStoragePath()` —
  the same observable behavior. `NullLogger` stays available for apps that want it.
- **Attribute collector ignored `@RouteHandler`** — `AttributeRouteCollector.buildHandler`
  always built a handler that constructed the controller through
  `clazz.getDeclaredConstructor().newInstance()`, so any controller with constructor
  dependencies failed with `NoSuchMethodException: <init>()`. Swapping the construction
  strategy alone would not have sufficed: an annotated controller method may take no
  arguments at all, so invoking it as `(container, route)` fails on argument count too.
  The annotation is the mechanism — it names a static
  `(ContainerContract, RouteContract)` handler, typically on the route provider, which
  resolves the controller from the container. PHP takes the handler straight off the
  attribute in `updateHandler`, and Sindri already emitted the same handler into the
  generated routing data; only the runtime collector disagreed. Fixed to honor
  `@RouteHandler`, and the fallback now resolves the controller from the container
  rather than reflecting a constructor. Reached in debug mode and as the no-cache
  fallback.

### JaCoCo exclusions (Java-only, non-unit-testable infra)

- `**/benchmark/**` — performance harnesses.
- `application/entry/{exchange,jetty,netty,tomcat}/**` — the worker-runtime entry
  adapters, folded into the framework from the standalone `entry/*` repos. They are
  bootstrap glue whose `run()` starts a persistent, non-daemon server that blocks
  forever (`server.join()` / `awaitTermination()` / `await()`) with JVM shutdown
  hooks, so they cannot reach full branch coverage from a unit test without leaking
  the server / hanging the test JVM. The package globs also drop the anonymous inner
  classes the HTTP adapters generate (handlers / servlets / channel initializers).
  The reusable logic they build on (`WorkerHttp` / `WorkerGrpc` / `GrpcBridge`) is
  covered directly, and each adapter is exercised end to end by a smoke test that
  boots a real application and drives the full convert → dispatch → emit path.

### Known unreachable branches (coverage < 100% by construction)

`valkyrja` is at **100% line coverage (6798/6798)** and **99.893% branch coverage
(1870/1872)**. The 2 remaining branches are counted by JaCoCo but cannot be executed
by any test without terminating the JVM — they are JaCoCo's two unavoidable cases:

- **`log/logger/abstract_/Logger.java` L26** — the implicit `default` of the exhaustive
  `switch (level)` over every `LogLevel`: all enum constants are handled, so the
  compiler's implicit default arm is unreachable. Adding an explicit `default ->` arm
  would be equally uncoverable.
- **`cli/server/support/Exiter.java` L27** — `if (exit) System.exit(code)`: the
  `exit == true` arm would terminate the test JVM, so only the frozen (`false`) arm
  is exercised.

Branches that were previously listed here have been eliminated by refactoring rather
than left uncovered: `Dispatcher` now uses `Objects.requireNonNullElse(cause, e)`,
`UploadedFile.getStream` uses `Objects.requireNonNull` for its file invariant, and the
dead `Collectors.toMap` merge lambdas (`(a, b) -> a`) in `RouteCollection.all`,
`HeaderCollection`/`ParamCollection`/`UploadedFileCollection` `getOnly`/`getAllExcept`
were replaced with explicit ordered-map loops.

## Sindri

### Ship a standalone executable on each release

Java builds a runnable fat jar. No release carries a release asset today.

- [valkyrjaio/sindri-java#90](https://github.com/valkyrjaio/sindri-java/issues/90)

### (Optional) Move Sindri into an isolated `.github/ci/sindri/` build

Low risk in Java, because Sindri parses source syntactically and never needs the app's
classpath. PHP needs verification first.

- [valkyrjaio/valkyrja-starter-app-java#90](https://github.com/valkyrjaio/valkyrja-starter-app-java/issues/90)

### Sindri generation bugs found comparing Java output to PHP (June 2026)

Generating the application's `App*Data` against the published Sindri 26.1.1 produced
output badly diverging from PHP. Fixes are being made in `java/sindri`:

- **[FIXED] HTTP route values were not real suppliers.** `HttpRouteAttributeReader`
  stored `new NameExpr(name)`, so `routes()` emitted `"version", version` (a bare,
  undefined identifier — non-compiling) instead of `() -> new Route(...)`. Now builds
  a `Supplier<RouteContract>` with the handler method-ref and request methods.
- **[FIXED] HEAD method dropped from `paths()`.** `AstHttpDataFileGenerator.buildPathsBody`
  skipped `HEAD`; PHP includes it. Removed the skip.
- **[FIXED] Dynamic routes + `dynamicPaths()`/`regexes()` not generated.** `regexes()` was
  hardcoded to `Map.of()`; dynamic routes (`{param}` paths) weren't detected and their
  `Parameter`s/regex weren't emitted. Fixed: `HttpRouteParameterReader` now reads
  `@Parameter`/`@Parameters` (resolving `Regex.*` constants via reflection); the reader
  detects `{` paths, precomputes the match regex by running the **real framework
  `Processor`** (drift-proof), and emits a `DynamicRoute` supplier; the generator emits
  `regexes()` and the dynamic `paths()`. Unit-verified in `generatesExpectedDynamicRouteContent`
  (asserts the `DynamicRoute` supplier, the `Regex.ALPHA`→`[a-zA-Z]+` resolution, the computed
  `(?<value>…)` regex, populated `dynamicPaths()`/`regexes()`, and that the whole generated
  file parses as valid Java). **App port gap (separate):** the app's `HomeController` still
  lacks the dynamic route PHP has.
  [valkyrjaio/valkyrja-starter-app-java#89](https://github.com/valkyrjaio/valkyrja-starter-app-java/issues/89)
- **[FIXED] `AppContainerData` missing framework providers (~36 of ~40 callbacks).**
  `fqnToFilePath` only resolved app-namespace source, so framework providers
  (`io.valkyrja.*`, reached via `HttpApplicationComponentProvider`) were skipped, and
  `collectProviderData` only recursed one level. Fixed by (1) `resolveSourceFromClasspath`
  — resolving a framework class's `.java` from the valkyrja **sources jar** on the
  classpath and staging it as a temp file (the portable equivalent of PHP's
  `ReflectionClass::getFileName()`), (2) full breadth-first recursion of the
  component-provider graph with a visited set, (3) adding `io.valkyrja:valkyrja:<v>:sources`
  to Sindri's runtime classpath and to the application's `sindri` task configuration.
  Unit-verified in `resolvesFrameworkProvidersFromClasspath` (app source in a temp dir +
  a "framework" provider, two levels deep, resolved from the test classpath).
  **Caveat:** not yet run end-to-end against the *real* valkyrja sources.
  [valkyrjaio/sindri-java#91](https://github.com/valkyrjaio/sindri-java/issues/91)

### Test gaps to strengthen in ALL THREE languages (Java/PHP/TS)

The bugs above slipped through because the end-to-end generate test asserted only that
the four `App*Data` files **exist**, not their content. Java has since strengthened its
assertions. Mirror them in PHP and TypeScript.

- [valkyrjaio/sindri-php#211](https://github.com/valkyrjaio/sindri-php/issues/211)
- [valkyrjaio/sindri-ts#105](https://github.com/valkyrjaio/sindri-ts/issues/105)

## VLID — cross-language parity

VLID is PHP-only today. Port it here, then assert the non-random portion against the
shared PHP fixture.

- [valkyrjaio/valkyrja-java#105](https://github.com/valkyrjaio/valkyrja-java/issues/105)
