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

### Branch coverage in CI

JaCoCo already measures **branch coverage** (its `BRANCH` counter), not just line
coverage. Add/raise the coverage gate to require **100% branch coverage** (every
`if`/ternary/`&&`/`||`/`switch` arm exercised both ways), not only 100% line
coverage — a line can be fully covered while one side of a condition never runs.
PHP is doing the equivalent via Cobertura `branch-rate` (see `architecture/php/TODO.md`).

The branch-coverage pass has been completed for `valkyrja`: **1578/1582 branches
(99.75%)**, with the only remaining gaps being the 4 irreducible branches listed
under "Known unreachable lines" below. Several genuinely-dead branches were removed
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

Every module's unit tests are ported and JaCoCo line coverage is **99.83%**
(5716/5726); the only uncovered lines are the provably-unreachable ones listed
under "Known unreachable lines" below — all reachable source lines are covered.

**Done — one dedicated test file per code-bearing class.** PHP keeps a separate
test file per class (e.g. `ParsedBodyParamCollectionTest`, `TextResponseTest`).
All **201 code-bearing classes** that lacked one (concrete, abstract, the 116
exceptions — each in its own file; the grouped `ExceptionsTest` is a PHP
anti-pattern and is not used here — enums, records, default-method interfaces)
now have a dedicated `<Class>Test` covering 100% of that file, and the former
grouped tests (`ConcreteParamCollectionsTest`, `TypedResponsesTest`,
`TypedHeadersTest`, `FormatterVariantsTest`, `MessageVariantsTest`,
`OutputVariantsTest`, `OptionParameterSubclassesTest`, the per-module
`*ExceptionTest`s) were split out and removed. Suite: 1210 tests / 382 files.

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

### JaCoCo exclusions (Java-only, non-unit-testable infra)

- `**/benchmark/**` — performance harnesses.
- `application/entry/ExchangeHttp`, `application/entry/ExchangeCgiHttp` — thin
  `com.sun.net.httpserver` bootstrap adapters with no PHP equivalent; their
  `run()` starts non-daemon server threads that cannot be exercised from a unit
  test without leaking the server / hanging the test JVM.

### Known unreachable branches (coverage < 100% by construction)

`valkyrja` is at **100% line coverage (5715/5715)** and **99.875% branch coverage
(1596/1598)**. The 2 remaining branches are counted by JaCoCo but cannot be executed
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
