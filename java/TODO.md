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

Missing badges for scrutinizer

### Cross-language testing-gap audit

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

Known starting point: the two known `sindri-java` gaps are closed — the golden now
feeds its dynamic route a `Processor`-computed regex with an end-to-end pin
alongside it, and the gRPC generator has a golden like the other four. The rest of
the suite has never been compared against PHP's.

### Root build rewrites the standalone CI builds (starter app)

In `valkyrja-starter-app-java`, `./gradlew useLatestVersions` at the root text-replaces version
strings into all six standalone `.github/ci/*/build.gradle.kts` files — separate Gradle builds the
root project does not own. The root's own `dependencyUpdates` report aggregates *every* project's
dependencies, not just its own, so a single bad candidate there propagates to all six files at
once.

That is exactly how `io.netty:netty-codec-http:5.0.0.Alpha2` reached every CI build from one
unfiltered report ([starter-app-java#53](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/53)).
The filter gap itself is fixed —
[#54](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/54) hoisted
`rejectVersionIf` out of `subprojects { }` into `allprojects { }`, so the root's aggregate report is
filtered too and prereleases no longer ride in this way. The coupling is what remains, and it is why
a single root-level miss had a six-file blast radius rather than a one-file one.

Decide whether the CI builds should be updated only by their own `useLatestVersions` runs — the
update-dependencies workflow already visits each `.github/ci/*` directory separately, so the root's
incidental rewrite is redundant as well as wide. Note that `app/build.gradle.kts` is *not* caught by
it (`:app` has its own filtered report), so the blast radius is specifically the CI directories.

### Branch coverage in CI

JaCoCo already measures **branch coverage** (its `BRANCH` counter), not just line
coverage. Add/raise the coverage gate to require **100% branch coverage** (every
`if`/ternary/`&&`/`||`/`switch` arm exercised both ways), not only 100% line
coverage — a line can be fully covered while one side of a condition never runs.
PHP is doing the equivalent via Cobertura `branch-rate` (see `architecture/php/TODO.md`).

The branch-coverage pass has been completed for `valkyrja`: **1870/1872 branches
(99.893%)**, with the only remaining gaps being the 2 irreducible branches listed
under "Known unreachable branches" below. Several genuinely-dead branches were removed
during the pass rather than left uncovered (`Answer.isValidResponse`,
`QuestionWriter.writeQuestion`, `Response.sendHttpLine`, `UploadedFile.moveTo`'s
stream-null guard, `MarshalUriFactory`/`UriFactory`/`Header`/`Value`/`Cookie`/
`RedirectResponse`/`JsonServerRequest` conditions, and the redundant `usedA` guard
in `CheckCommandForTypoMiddleware.similarText`).

### Response cache rework — port from PHP (DONE in PHP)

**This has been completed in PHP and needs to be ported to Java.** The response
cache no longer generates/loads a source file; it serializes the response to JSON
and reconstructs it. Apply the same change here:

1. **`http/server/CacheResponseMiddleware`** — on `terminated()`, serialize the
   response to JSON: `class`, `statusCode`, `reasonPhrase`, `headers` (list of
   `{name, value}`), `body`, and `uri` (redirects only). On `requestReceived()`,
   reconstruct from the JSON by instantiating the stored response class with only
   its `headers` argument (the one constructor arg shared by every response
   subclass — all extend `Response`) and applying `withStatusCode` /
   `withReasonPhrase` / `withBody` (+ `withUri` for redirects). No source-file
   generation, no class-loading of a generated file. Keep the TTL/expiry/validity
   logic as-is.
2. **Delete** the now-unused file-generation classes:
   - `http/server/generator/ResponseFileGenerator.java` (+ `contract/ResponseFileGeneratorContract`)
   - `support/generator/abstract_/FileGenerator.java` (+ `contract/FileGeneratorContract`
     and its status enum) — only consumed by `ResponseFileGenerator`
   - their tests, and any README "File Generation" section.

The existing `CacheResponseMiddleware` test is behavioral (round-trips every
response type through the cache) and should pass unchanged once the JSON rework is
in place. PHP reference commit: `[Http] Replace response-cache file generation with
JSON serialization and remove FileGenerator.` (see `architecture/php/TODO.md`).

### Incomplete ports (PHP → Java)

These exist in the PHP framework but are not yet ported to Java. Test coverage
currently targets only the code that exists in the Java source; finish the port
(with tests) to reach parity.

- **Event** — missing vs PHP:
  - `attribute/` — attribute-based listener support (`Listener`, `ListenerHandler`)
  - `collector/` — `AttributesListenerCollector`
  - event `ServiceProvider` (only `EventComponentProvider` exists)
  - concrete event throwables — `EventInvalidArgumentException` /
    `EventRuntimeException` are abstract with no concrete subclass yet

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

**Deferred — no-bytecode classes.** The **146 pure interfaces** (abstract methods
only) and **21 annotation markers** (`@interface`) have no executable code for
JaCoCo to measure. PHP still has a test file per class, so eventually add
structural/contract tests for these too (assert method signatures / annotation
presence) to fully mirror PHP's per-class layout — not yet done.

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

- Ship a standalone, downloadable executable on each release so Sindri can be
  used without adding it as a build dependency.
    - Java: build a runnable **fat/uber jar** from `bin/sindri` (Gradle shadow
      plugin, or the `jar` task with a `Main-Class` manifest and bundled
      dependencies) and attach it to the GitHub release as a release asset so it
      can be downloaded and run directly (`java -jar sindri.jar ...`).
    - This mirrors PHP shipping a **Phar** and TypeScript shipping a standalone
      binary on release — see each language's `TODO.md` for the per-language task.

- **(Optional) Move Sindri into an isolated `.github/ci/sindri/` build in the
  application**, like the other CI tools (`junit`, `errorprone`, …), instead of
  wiring it into the `:app` build. The application currently exposes Sindri via a
  `sindri` dependency configuration + `JavaExec` tasks (`./gradlew sindri` /
  `sindriHttp` / `sindriCli`) in `app/build.gradle.kts`.
    - Java: low risk — Sindri parses source **syntactically** (no symbol solver),
      so it never needs the app's compile/runtime classpath; an isolated build just
      needs its own `io.valkyrja:sindri` dependency and a task whose `workingDir`
      points at the app module so it finds `Config.java` and writes the `App*Data`
      files in place. Verify the config path resolves from the isolated build dir.
    - PHP: **needs verification first.** An isolated `ci/sindri` is a separate
      Composer project, so it would not have the application's autoload / installed
      dependencies on its include path, and `bin/sindri` may fail to locate the app
      config or resolve provider/controller classes referenced from it. Confirm
      Sindri can find and read the right config from outside the app's vendor tree
      before adopting this layout in PHP.

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
  lacks the dynamic route PHP has — add it to the Java app to get matching output.
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
  **Caveat:** not yet run end-to-end against the *real* valkyrja sources — that needs the
  app to regenerate with this Sindri build, which happens after a Sindri release. The real
  sources jar is confirmed on the classpath and contains the provider `.java`.

### Test gaps to strengthen in ALL THREE languages (Java/PHP/TS)

The bugs above slipped through because the end-to-end generate test
(`GenerateDataFromConfigCommandTest`) only asserted the four `App*Data` files **exist**,
not their **content** — so non-compiling/empty output passed. When fixing each bug,
strengthen tests to assert the generated content, and mirror these in PHP and TS:

- **Assert generated `routes()` content**: a real `() -> new Route(...)` supplier with
  the handler method-ref and request methods (not a bare name); the bare-name placeholder
  must never appear. (Done in Java's `generatesExpectedHttpRoutingContent`.)
- **Assert `paths()` includes `HEAD`** for default-method routes. (Done in Java.)
- **Assert `AppContainerData` callbacks include framework-provider publishers**, not just
  app-local ones — requires a fixture whose component provider pulls in a "framework"
  provider resolved from outside the app namespace. (Done in Java's
  `resolvesFrameworkProvidersFromClasspath`, including a two-levels-deep provider to guard
  the recursion depth.)
- **Assert dynamic-route output**: `DynamicRoute` supplier, populated `dynamicPaths()` and
  `regexes()`, and that `Regex.*` parameter constants resolve. (Done in Java's
  `generatesExpectedDynamicRouteContent`.)
- **Parse the generated file** in a test so malformed output / bad escaping is caught
  structurally, not just by substring. (Done in Java via `StaticJavaParser.parse` on the
  generated `AppHttpRoutingData`; ideally extend to a full compile, and to PHP/TS.)

## VLID — cross-language parity

**Cross-language change — mirror in every port (Go, PHP, Python, TypeScript).**
VLID (`Type/Vlid`) is PHP-only today; port it here (code + tests). It is the source
of the queue envelope `id` (a **VLID V1** — the longest, most-random version). Lock
cross-language parity:

- Port `Type/Vlid`, then add a conformance test: generate a VLID for **each version
  V1–V4** from a **fixed input timestamp**.
- Assert this port produces a byte-identical **non-random portion** vs the PHP
  fixture — the encoded **microsecond timestamp** and the **version digit at
  position 14** must match exactly. The random bits differ by design; exclude them.
- This gate prevents timestamp-encoding / version-digit-placement drift from
  silently breaking cross-language `id` interop.
