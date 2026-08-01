# PHP

Work items are tracked as issues in the repository that owns the work. This file is the
index. A section with no issue link holds a decision already made, or a design record —
not a task.

---

## Tracked as issues

### Testing

- **Every contract declares its expected method signatures** — [valkyrja-php#1035](https://github.com/valkyrjaio/valkyrja-php/issues/1035)
- **One test per exception, replacing the grouped `ExceptionsTest` files** — [valkyrja-php#1036](https://github.com/valkyrjaio/valkyrja-php/issues/1036)
- **Assert every Constant value** — [valkyrja-php#1037](https://github.com/valkyrjaio/valkyrja-php/issues/1037)
- **Assert every Enum case value** — [valkyrja-php#1038](https://github.com/valkyrjaio/valkyrja-php/issues/1038)
- **A full app example, exercised in the Functional tests** — [valkyrja-php#1039](https://github.com/valkyrjaio/valkyrja-php/issues/1039)
- **Orm `ServiceProviderTest` — assert the DSN and options per PDO type** — [valkyrja-php#1040](https://github.com/valkyrjaio/valkyrja-php/issues/1040)
- **Data tests — assert every getter after a `with`/`set`** — [valkyrja-php#1041](https://github.com/valkyrjaio/valkyrja-php/issues/1041)
- **Use mocks instead of fixture classes where expectations suffice** — [valkyrja-php#1042](https://github.com/valkyrjaio/valkyrja-php/issues/1042)
- **Test component — move the output out of `run()`, move it to `tests`** — [valkyrja-php#1043](https://github.com/valkyrjaio/valkyrja-php/issues/1043)

### Http

- **`RateLimiterMiddleware`** — [valkyrja-php#1024](https://github.com/valkyrjaio/valkyrja-php/issues/1024)
- **Flattened Message collection variants** — [valkyrja-php#1025](https://github.com/valkyrjaio/valkyrja-php/issues/1025)
- **`#[MapRequestFormParams]`** — [valkyrja-php#1026](https://github.com/valkyrjaio/valkyrja-php/issues/1026)
- **The five factories become concrete classes** — [valkyrja-php#1027](https://github.com/valkyrjaio/valkyrja-php/issues/1027)
- **Controller namespace with default controllers and actions** — [valkyrja-php#1028](https://github.com/valkyrjaio/valkyrja-php/issues/1028)
- **Add the missing content types** — [valkyrja-php#1029](https://github.com/valkyrjaio/valkyrja-php/issues/1029)
- **A handler closure on routes and listeners** (Http, Cli, Event) — [valkyrja-php#1030](https://github.com/valkyrjaio/valkyrja-php/issues/1030)
- **Skip an empty routes or listeners block during data collection** — [valkyrja-php#1031](https://github.com/valkyrjaio/valkyrja-php/issues/1031)
- **A default `ThrowableHandler` for Http and Cli** — [valkyrja-php#1032](https://github.com/valkyrjaio/valkyrja-php/issues/1032)
- **`EventCapableContainer`** — [valkyrja-php#1033](https://github.com/valkyrjaio/valkyrja-php/issues/1033)
- **A Writer class for output buffers and echo** — [valkyrja-php#1034](https://github.com/valkyrjaio/valkyrja-php/issues/1034)

### Cli

- **Hide commands from the list output** — [valkyrja-php#1044](https://github.com/valkyrjaio/valkyrja-php/issues/1044)
- **A no-formatting global option** — [valkyrja-php#1045](https://github.com/valkyrjaio/valkyrja-php/issues/1045)
- **Hide help and version from the list output** — [valkyrja-php#1046](https://github.com/valkyrjaio/valkyrja-php/issues/1046)
- **Configurable global options (`APP_CLI_GLOBAL_OPTIONS`)** — [valkyrja-php#1048](https://github.com/valkyrjaio/valkyrja-php/issues/1048)
- **`-v` date in local system time** — [valkyrja-php#1049](https://github.com/valkyrjaio/valkyrja-php/issues/1049)
- **Consolidate `VersionCommand` and `ListCommand`** — [valkyrja-php#1050](https://github.com/valkyrjaio/valkyrja-php/issues/1050)
- **`helpText` takes a callable returning a `Message`** — [valkyrja-php#1051](https://github.com/valkyrjaio/valkyrja-php/issues/1051)
- **Progress bar** — [valkyrja-php#1052](https://github.com/valkyrjaio/valkyrja-php/issues/1052)

### Type

- **`Resource` type** — [valkyrja-php#1053](https://github.com/valkyrjaio/valkyrja-php/issues/1053)
- **`StreamResource` type** — [valkyrja-php#1054](https://github.com/valkyrjaio/valkyrja-php/issues/1054)
- **`fromMixed` on each support helper** — [valkyrja-php#1077](https://github.com/valkyrjaio/valkyrja-php/issues/1077)
- **A binding-key constant class per component** — [valkyrja-php#1055](https://github.com/valkyrjaio/valkyrja-php/issues/1055)
- **Port the Type module to every language** — [architecture#143](https://github.com/valkyrjaio/architecture/issues/143)

### Orm

- **Model casting becomes closures, for cross-language compatibility** — [valkyrja-php#1056](https://github.com/valkyrjaio/valkyrja-php/issues/1056)
- **`Statement::fetch()` returns null when no row is found** — [valkyrja-php#1057](https://github.com/valkyrjaio/valkyrja-php/issues/1057)
- **`getXValue()` for every `getXField()`** — [valkyrja-php#1058](https://github.com/valkyrjaio/valkyrja-php/issues/1058)
- **A defaultable service for Entity instead of entity matchers** — [valkyrja-php#1059](https://github.com/valkyrjaio/valkyrja-php/issues/1059)
- **`QueryBuilderFactory::fromQuery` and the four variants** — [valkyrja-php#1060](https://github.com/valkyrjaio/valkyrja-php/issues/1060)

### Auth

- **Login by email code, with a configurable cache TTL** — [valkyrja-php#1061](https://github.com/valkyrjaio/valkyrja-php/issues/1061)
- **Cli commands** — [valkyrja-php#1062](https://github.com/valkyrjaio/valkyrja-php/issues/1062)
- **Http middleware** — [valkyrja-php#1063](https://github.com/valkyrjaio/valkyrja-php/issues/1063)
- **Http controllers and their data objects** — [valkyrja-php#1064](https://github.com/valkyrjaio/valkyrja-php/issues/1064)

### Queue

Read [`../QUEUE.md`](../QUEUE.md) first. It carries the settled design.

- **Message, and the `QueueResult` enum** — [valkyrja-php#1065](https://github.com/valkyrjaio/valkyrja-php/issues/1065)
- **Middleware** — [valkyrja-php#1066](https://github.com/valkyrjaio/valkyrja-php/issues/1066)
- **Routing** — [valkyrja-php#1067](https://github.com/valkyrjaio/valkyrja-php/issues/1067)
- **Server** — [valkyrja-php#1068](https://github.com/valkyrjaio/valkyrja-php/issues/1068)

### Event

- **Better template docblocks on `Dispatcher`** — [valkyrja-php#1069](https://github.com/valkyrjaio/valkyrja-php/issues/1069)
- **Listener priority** — [valkyrja-php#1070](https://github.com/valkyrjaio/valkyrja-php/issues/1070)
- **A stop-propagation abstract event** — [valkyrja-php#1071](https://github.com/valkyrjaio/valkyrja-php/issues/1071)

### Types and static analysis

- **`array<array-key, X>` becomes `array<X>`** — [valkyrja-php#1072](https://github.com/valkyrjaio/valkyrja-php/issues/1072)
- **Audit every string for `non-empty-string`** — [valkyrja-php#1073](https://github.com/valkyrjaio/valkyrja-php/issues/1073)
- **The five recorded Psalm findings** — [valkyrja-php#1074](https://github.com/valkyrjaio/valkyrja-php/issues/1074)
- **The two upstream Psalm issues** — [valkyrja-php#1075](https://github.com/valkyrjaio/valkyrja-php/issues/1075)
- **A data object for a method with more than three parameters** — [valkyrja-php#1076](https://github.com/valkyrjaio/valkyrja-php/issues/1076)
- **Analyze the tests with PHPStan and Psalm** — [valkyrja-php#1006](https://github.com/valkyrjaio/valkyrja-php/issues/1006)

### Tooling

- **Characterize the Xdebug branch-map inflation, then gate on branch coverage** — [valkyrja-php#1078](https://github.com/valkyrjaio/valkyrja-php/issues/1078)
- **PHP Code Sniffer requires docblocks on functions and methods** — [valkyrja-php#1079](https://github.com/valkyrjaio/valkyrja-php/issues/1079)
- **A default filesystem, and `FlysystemFilesystemContract`** — [valkyrja-php#1080](https://github.com/valkyrjaio/valkyrja-php/issues/1080)
- **Reset the reused deps branch in the update-dependencies workflow** — [.github#221](https://github.com/valkyrjaio/.github/issues/221)

### Documentation and scaffolding

- **Dedicated READMEs for `ServiceProvider`, `ComponentProvider`, and the rest** — [valkyrja-php#1081](https://github.com/valkyrjaio/valkyrja-php/issues/1081)
- **Deprecate the application skeleton repo in favor of generator commands** — [valkyrja-php#1082](https://github.com/valkyrjaio/valkyrja-php/issues/1082)
- **`.valkyrja.apps`, written by a `create:application` command** — [valkyrja-php#1083](https://github.com/valkyrjaio/valkyrja-php/issues/1083)
- **Re-run the README drift checks, and add them to CI** — [valkyrja-php#1018](https://github.com/valkyrjaio/valkyrja-php/issues/1018)

### Validation

- **Allow a `UnitEnum` or a `BackedEnum` in valid values** — [valkyrja-php#1084](https://github.com/valkyrjaio/valkyrja-php/issues/1084)

### View

- **A shared templating engine, so one template works in every port** — [architecture#142](https://github.com/valkyrjaio/architecture/issues/142). This needs a design document before any port starts. The proposed rename of Orka to Syn or Syni is part of that decision.

### Cross-language

- **VLID conformance fixture** — [valkyrja-php#1015](https://github.com/valkyrjaio/valkyrja-php/issues/1015)
- **Compare this port's test suite against the other ports** — [valkyrja-php#1016](https://github.com/valkyrjaio/valkyrja-php/issues/1016)
- **Update the repo description once gRPC and Queue land** — [valkyrja-php#1017](https://github.com/valkyrjaio/valkyrja-php/issues/1017)
- **Ship a standalone `sindri` Phar on each release** — [sindri-php#210](https://github.com/valkyrjaio/sindri-php/issues/210)

---

## Decisions and design records

These are not tasks. Each one records a decision already made, or a design worked
through. They stay here so the reasoning is not lost.

### Split the framework out of the main repo

Eventually split the framework out and keep the main (`valkyrja`) repo lighter. **Not a
task for now** — captured so it is not lost. The `--path-coverage` CI cost is one symptom
of how large the main repo has become.

### Branch coverage in CI — what was solved, and how

**100% branch and path coverage IS achievable in PHP.** Earlier notes claimed Xdebug's
per-internal-call exception edges made it impossible. That was wrong.

The real cause of a phantom uncovered branch is namespace function resolution. An
unqualified builtin call such as `strpos()` inside a namespace compiles to a runtime
lookup — does `Ns\strpos` exist, and otherwise fall back to `\strpos`. Xdebug counts the
never-taken namespace-local edge as an uncovered branch, and it multiplies path counts.
Importing the function with `use function strpos;` resolves it at compile time, and the
phantom branch disappears. The `@auto` php-cs-fixer ruleset now does this automatically,
through `native_function_invocation` and `global_namespace_import`.

Three genuinely unreachable categories exist. Each one is handled, not excused.

- **An exhaustive `match($enum)` with no `default`.** PHP compiles in an
  `UnhandledMatchError` throw that no input can reach. Fold the last arm into `default`.
- **A call a test cannot make.** An I/O syscall intercepted by namespace-function
  shadowing (`header`, `flush`, `ob_flush`, `ob_get_level` in `Response`, where
  qualifying them for coverage breaks the shadow), and a call that ends the process
  (`exit()` in `Exiter`). Wrap each one in a protected seam method, override the seam in
  a `Tests\Fixtures` fixture for the behavior assertions, and add one real-call test to
  cover the seam body. `ResponseSendRecorderFixture` is the pattern. Only where no real
  call is possible at all, which is `exit()`, does the seam body carry
  `@codeCoverageIgnore`.
- **A conditional inside a trait.** Xdebug emits one function entry per file for a trait
  method, keyed by the *trait*, so every using class shares it and class load order
  decides whose hits survive. Move the branch out of the trait into a support class that
  the trait delegates to. `Trait\Arrayable` to `Support\Enumerable` is the pattern, and
  `Trait\JsonSerializable` now follows it.

**Speed is solved.** `--path-coverage` is slow in CI on `valkyrja`, the largest repo, so
it is split into its own `path-coverage` Composer script there. The other repos keep
`--path-coverage` in their `coverage` script. `phpunit-path-coverage-parallel` runs the
suite as one shard per component and merges with `phpcov`: **190s against 1942s
serially**, a 10x gain whose floor is the slowest shard.

**The source-side blockers are gone.** Every branch that no *test* could reach has been
closed by a source change, one pull request per behavior decision:
`Answer::isValidResponse()`'s dead `allowedResponses === []` clause (#948),
`DispatchFactory::fromReflection()`'s exhaustive `match (true)` (#943),
`RequestHandler::getOutputDufferFlags()`'s dead HHVM 3.3 `-1` fallback (#944),
`UploadedFileFactory::isValidSapiEnvironmentForUploads()`'s `PHP_SAPI` seam (#946),
`Exiter::exit()`'s process-ending arm (#947), and `Trait\JsonSerializable`'s
trait-hosted conditional (#949). The reachable gaps closed earlier, in #936 and #939.

**Java and TypeScript** both reach 100% branch, and both now gate on it.

The one open action — characterize the measurement inflation, then gate — is
[valkyrja-php#1078](https://github.com/valkyrjaio/valkyrja-php/issues/1078).

### Container — a `Provides` attribute for provider methods

A `ServiceProvider` could carry an attribute on the method it provides. This can be added
later. At this time it is not required.

### Http — where middleware config lives

Should middleware config be baked into Http and Cli, because they are so integral, or be
its own config?

**Decided: keep it as it is, and make a single handler later.** The current shape
preserves the ability to build one handler later and use the config data class to house
the middleware, adding to the list through the matched route in `Router`.

### Http — `RequestReceivedMiddleware` on a Route

An open question, not a decision. Request-received middleware on a route would let
middleware that matters to certain routes run only for those routes, rather than on every
request.

### Event — a Subscriber for events

Considered and rejected in the same breath. A subscriber is the same concept as a
Controller for Cli and Http. But event dispatching should be very simple, because a
developer usually does it on the fly when it is needed.

### Cli, Http, and Queue — renaming Command to Route

Considered at length, and rejected.

A `Router` routes `Route` objects to a handler. In Http a route takes a request and goes
to a handler for an action. In Cli it takes an input and goes to a handler for a command.
In Queue it would take a payload and go to a handler for a queue.

The vocabulary that follows from it:

- **Action** — a class that handles one specific Http route as a whole.
- **Command** — a class that handles one specific Cli route as a whole.
- **Controller** — a class that handles several routes, for any protocol.

A distinct per-protocol name such as `CommandRoute` and `ActionRoute` was considered and
rejected. There is no `ActionRouter` or `CommandRouter`, because the namespace already
distinguishes them. **This level of granularity is unnecessary.**

What stands regardless of the naming: a Controller houses route definitions tied to
single methods, which is right for a simple action or command. A complex one gets its own
Action or Command file, so several methods can serve it without muddying the others.

### Is returning null cheating?

A worked-through position, not an open task.

Every method could throw rather than return null. That gives parity with Java and Go. The
convention that follows, for a single-item query:

- `Create…` returns a new instance, or throws.
- `Get…` returns an expected existing instance, or throws.
- `Retrieve…` returns an expected existing instance, or throws.
- `GetOrCreate…` returns an existing instance, or a new one if none exists, or throws.
- `Find…` returns an existing instance if it exists, or null.

For a collection query, `Get…` always returns a collection, empty when nothing matches.
That is the only form that does not throw.

The reasoning: take a caller who asks a phone system for an operator by name. If the
operator exists, the system transfers the call. If the operator does not exist, the
system throws, and the middleware for the application and that route handles the case
specifically. The context is preserved. A null object instead continues as though nothing
happened, which loses the context and still forces the caller to handle absence — now
with an object that did not need to exist.

A separate rule falls out of the same thinking. **Allow null only on a parameter that is
an object.** An array defaults to `[]`, a string to `''`, a bool to `false`, an int to
`0`, and a float to `0.0`.

### Lifecycle

The bootstrap design, recorded in full.

An app is created with `Env::class` and a `Config` or `Data` object. It creates the
container, and bootstraps from `Entry\App`.

**With a `Config`,** setup runs normally. The config class is slimmed to Container,
Event, Cli, and Http. The app iterates its components, and makes sure it calls the core
ones — `ApplicationComponent`, `ContainerComponent`, `EventComponent`, `CliComponent`,
`HttpComponent`, `DispatcherComponent`, `AttributeComponent`, `ReflectionComponent`. That
is what happens now, except that no component holds a config class. The container sets
`Config::class` as a singleton.

**With a `Data` object,** the app splits the data apart and sets each part as a
singleton: `ContainerData`, `EventData`, `CliData`, `HttpData`.

The container then takes its data from the `ContainerData` singleton, and gains a
`getData()` that mirrors today's `getCacheable()`. When `ContainerData` came from the
passed `Data`, the container uses it directly. Otherwise the container falls back to the
service-provider set, where the collector runs the code from `setupNotCached`.

Event, Http, and Cli each get their data from the container and pass it into the
constructor. The service provider has an entry to get the data, whether the data arrived
with the cached data object or defaults to the collector path.

### Debugging, profiling, and a debug bar

Explored, then deliberately parked: **for the time being let us not focus attention here,
and work on finishing the application and the tests.**

What was considered. A `ProfilerCapableContract`. The ability for a service provider to
log what it called, behind a debug flag, perhaps through the event dispatcher:

```php
if (debug) {
    $this->event->dispatch(new DebugMethodCalled($this, __METHOD__));
}
```

A `DebugContainer` whose `get()` returns the instance wrapped in a `DebugClass` was
considered and found unworkable. The wrapper breaks the contract and the type hints. An
anonymous class that extends the returned instance would be needed instead.

The likely conclusion is that anything which needs debug capability has to implement it
itself. This needs real research before anyone starts.

### Documentation inspiration

Links kept for whoever writes the corresponding documentation.

- **Orm** — <https://symfony.com/doc/current/doctrine.html#learn-more>
- **Http middleware** — <https://botman.io/2.0/middleware>
- **Content types** — <https://stackoverflow.com/questions/23714383/what-are-all-the-possible-values-for-http-content-type-header>
- **Uri filtering** — <https://github.com/laminas/laminas-diactoros/blob/3.9.x/src/Uri.php>
- **Reflector, callable parameter counts** — <https://stackoverflow.com/questions/13071186/how-to-get-the-number-of-parameters-of-a-run-time-determined-callable>
- **A GitHub composite action** — <https://docs.github.com/en/actions/tutorials/creating-a-composite-action>

---

## Open questions

Each one needs a decision before it can become an issue.

- **Rethink optional parameters.** Toward what?
- **A `sindri` "init" command.** The name and the scope are both open. "Interactive —
  stays open, and lets a developer keep running commands in one session" is the only
  spec.
- **Debug the commit message checker workflow.** Debug what?
- **`OutputThrowableHandler`.** Recorded with no description.
- **A file logger.** Java built a zero-dependency one. Does PHP want one, when Monolog is
  the default?
- **A modules concept** — an `index.php` and a `composer.json` for each module, with a
  shared composer import from a lib folder. The goal is not stated.
- **Continue deprecating Env.** The scope is unknown until the `sindri` work settles.
- **Auth session tokens in the database**, so a developer can dismiss one login session
  or log every session out.
- **PSR Cache support through a new `PsrCache` manager**, and **PSR support in Client.**
  Both raise the same question: does this framework want a PSR dependency?
- **Undo the UuidV1 int cast change.** No commit is named.
- **Expand `ApplicationTest`.** To cover what?
- **Implement gRPC fully.** [`../GRPC.md`](../GRPC.md) and
  [`../GRPC_IMPLEMENTATION.md`](../GRPC_IMPLEMENTATION.md) describe a very large body of
  work. Break it into pieces before any of it becomes an issue.

---

## PHP version support

**Warning: the two tables below disagree.** They differ on the bug-fix dates for versions
26, 27, and 28, and on the PHP range for version 27. Do not use either one until somebody
confirms which is correct. Both are kept here, unchanged, so the conflict is visible.

<table width="100">
    <thead>
        <tr>
            <th>Version</th>
            <th>PHP (*)</th>
            <th>Release</th>
            <th>Bug Fixes Until</th>
            <th>Security Fixes Until</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>25</td>
            <td>8.4 - 8.6</td>
            <td>December 11th, 2025</td>
            <td>Q1 2026</td>
            <td>Q1 2026</td>
        </tr>
        <tr>
            <td>26</td>
            <td>8.4 - 8.6</td>
            <td>Q1 2026</td>
            <td>Q3 2027</td>
            <td>Q1 2028</td>
        </tr>
        <tr>
            <td>27</td>
            <td>8.5 - 8.7</td>
            <td>Q1 2027</td>
            <td>Q3 2028</td>
            <td>Q1 2029</td>
        </tr>
        <tr>
            <td>28</td>
            <td>8.6+</td>
            <td>Q1 2028</td>
            <td>Q3 2029</td>
            <td>Q1 2030</td>
        </tr>
    </tbody>
</table>

| Version | PHP (*)   | Release             | Bug Fixes Until | Security Fixes Until |
|---------|-----------|---------------------|-----------------|----------------------|
| 25      | 8.4 - 8.6 | December 11th, 2025 | Q1 2026         | Q1 2026              |
| 26      | 8.4 - 8.6 | Q1 2026             | Q1 2027         | Q1 2028              |
| 27      | 8.5 - 8.6 | Q1 2027             | Q1 2028         | Q1 2029              |
| 28      | 8.6+      | Q1 2028             | Q1 2029         | Q1 2030              |
