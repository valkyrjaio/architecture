```php
/* +-----------------------------------------------------------------------+ *
 * | This file is part of the Valkyrja Framework package.                  | *
 * +-----------------------------------------------------------------------+ *
 * | Copyright (c) 2025-present Melech Mizrachi <melechmizrachi@gmail.com> | *
 * +-----------------------------------------------------------------------+ *
 * | This source file is subject to the MIT license that is bundled with   | *
 * | this package in the file LICENSE.md, and is also available through    | *
 * | the world-wide-web at the following url:                              | *
 * +-----------------------------------------------------------------------+ *
 * | https://github.com/valkyrjaio/valkyrja/blob/master/LICENSE.md         | *
 * +-----------------------------------------------------------------------+ */
```

# Versioning
------------

- Major – Most Breaking Changes, etc. Yearly
- Minor – New features or fixes that are breaking changes but must be
  implemented before a major release
- Patch - Fixes that are not breaking changes

# Changes for v26
-----------------

## First for Parity with Java

- dedicated Readme for ServiceProvider, ComponentProvider, etc.
- ServiceProvider should not be final to allow for extending. This is a
  framework after all.
- ServiceProvider should have an attribute for the method
- Change service and singletons to lambdas instead of what they are now. Add
  Service and Singleton annotations/Attributes
- Bypass dispatch by adding handler, and making a closure for these using the
  dispatch object.
- Rename Event\Dispatcher to something else. EventDispatcher?
  ListenerDispatcher?
- Rename collectors to what they're collecting. EventCollectorContract,
  CliRouteCollectorContract, HttpRouteCollectorContract, etc.
- !!! Remove the ComponentProvider constant class

## Container

- bindService and bindSingleton to take closures/callables

## Http/Cli (and Listener)

- Need to deprecate dispatch entirely. We cannot have it be the crux that
  handles calling route and listener actions magically. This only works in PHP
  and Java, will fail to have a solution for this in the other languages.
    - Only allow handler for routes and listener.
    - CacheableHandlerRoute|CacheableHandlerListener for cache writing of said
      closure.

## Http

- Move File/Throwable/Exception/Constant to just File/Constant.

## Bin

- Becomes build tool. Own repo.

## Auth

- Add a login via email retriever. Code gets stored to cache, email sent, user
  enters code, code is verified against cache.
    - Cache TTL is a configurable value
        - IE to the user this looks like code is valid for 5 minutes. After 5
          minutes the cache is cleared automatically.
            - Also cleared after successful retrieval for security reasons

## Contract and Class name constants

- Other languages (Go, Python, TypeScript) do not have support for class names
  like PHP and Java.
- We would need a constant class that has all the defined class names and
  contracts for the component

## Validation

- Rename getException to throwException or actually return the exception and
  throw where required.

## Github

- Remove the validate composer individual checks?

## ALL

- Change `array<array-key, blah>` to `array<blah>`
- Check all strings and see if they should be non-empty-string
- Rethink optional parameters, but maybe we can do this in v26

## Env

- Continue deprecating this.

## Http

- RateLimiterMiddleware
    - The lesson from the Ebay interview :)
- Should config for middleware be baked into http and cli since they're so
  integral, or should they be their own config. I am leaning toward the latter.
    - If we keep as is then we have the ability to make a single handler later
      and use the config data class to house the middlewares and add to the list
      via the matched route in Router.
    - Keep as is and make a single handler later.

## PHP Code Sniffer

```
  <!-- Require docblocks for functions/methods -->
  <rule ref="Squiz.Commenting.FunctionComment.Missing"/>
```

## Tests

- Add a full app example to tests and test it in Functional tests.
- Orm
    - Update ServiceProviderTest to check the DSN and Options passed in for each
      Pdo type
        - In the callable function do assertions
        - Check default options as well as configured through env
- Update all data tests to make sure all get methods return expected data after
  a with/set is called, not just the corresponding get method for the with/set
- Test all Exception classes for proper inheritance
    - Do this in a separate test class for each exception instead of in one
      massive file that's hard to read and keep track of. So break apart any
      existing exceptions class tests
- Every Constant value
- Every Enum case value
- Test all contracts to ensure they have all the expected methods!!!!
    - See ModelTest.php for example
- Test ALL manner of Http Controller Route types including all parameterized
  options!!!!!!!!!!!!!!!!!!!!!!!
- Update tests to not rely on classes made in classes namespace if a mock with
  proper expectations can be used instead.
    - Update middleware tests to use mocks

## CI

- Debug commit message checker workflow

## Writer

- Add a writer class for Ooutput buffers, and echo's etc. This way these things
  can be mocked

## Filesystem

- Add a default filesystem so we can mock this in tests
- Add contract for FlysystemFilesystemContract with getFlysystem pub method

## Http, Cli, Event

- Add handler (closure) to routes and listeners.
- Either a handler or a dispatch object must exist
- If handler have the preferred way of doing things be to call that handler as
  the action then dispatch from the handler to a service handling that actions
  logic, or if within an action class it would be a pubsf method of name
  dispatch
- Update data collection logic to check if routes/listeners, is empty
    - if empty continue to next block

## Http

- Add RequestReceivedMiddleware to Route?
    - This way certain middleware that only have to do for certain routes will
      run, and not on every single request (when other routes wouldn't need that
      particular request received middleware)
- Message Collections, make flattened variants.
- #[MapRequestFormParams(Someclass::class)]
    - Auto maps the request params to the Someclass properties
- Add Uri TODO filtering
    - https://github.com/laminas/laminas-diactoros/blob/3.9.x/src/Uri.php
- CacheResponse instead of creating a class that gets loaded, do a json file or
  some file so we can have parity between languages.
    - headers
    - status code
    - status phrase
    - contents
    - etc
    - The we build the response object with those details within the middleware
      class
- Factories to actual classes
    - No longer abstract
        - No more private methods
        - Use static vs self
    - RequestFactory
    - HeaderFactory
    - ServerRequestFactory
    - UploadedFileFactory
    - UriFactory

## Cli

- Option in config to hide certain commands from the output list
    - Hidden commands?
    - Secret commands?
- No formatting cli global option
- Hide help and version from list command output
- Add OutputThrowableHandler
    - What?!
- See HelpCommand global options, make this configurable. Env key for
  APP_CLI_GLOBAL_OPTIONS
- `-v` date should maybe use local system date/time?
    - Change CI release to UTC then
        - No because we still want the Changelog to be in MST, imo
- VersionCommand
    - Add Env keys for title, etc
        - Move out of Application.php
    - ValkyrjaVersion (bin/valkyrja -v)
    - AppVersion (cli -v)
    - Should be one class, that takes env keys for the output.
        - Same for ListCommand
- Update helpText to take a (callable():Message)|Message
    - `callable():Message` doesn't exactly work. May need to find a different
      workaround for this
- Progress bar
    - https://stackoverflow.com/questions/2124195/command-line-progress-bar-in-php

## Bin

- .valkyrja.apps
    - created by create:application command in bin
    - HashMap of name => [namespace => capitalized-string, ...]

## ThrowableHandler

- Add a default ThrowableHandler that logs the exception displays the error on
  screen
    - One for Cli and one for Http

## Application

- Expand ApplicationTest

## Container

- Add an EventCapableContainer that extends the Container and adds event
  dispatching to it

## Type

- Add Resource type with methods
- Add StreamResource type
    - Can use this for mocking anywhere fopen etc are used
- Undo the UuidV1 int cast change me thinks
- Add fromMixed(mixed $value) to each support helper class of each type.

## Rector

- Need a test for the new rule
    - FindThis
        - FindThisNot
        - DontFindThis
        - etc; find all variations and make a test class that can be parsed by
          PhpParser then ran through the rule and output checked
- Move this to a separate repo!!!

## Psalm Type Fixing

- Analyzer::incrementMixedCount
- AssignmentAnalyzer
    - Move $codebase->analyzer->incrementMixedCount($statements_analyzer->
      getFilePath()) and all codeblock related to it inside of the other if
      statement
- /Valkyrja/Cli/Routing/Data/ArgumentParameter.php:152
- /Valkyrja/Cli/Routing/Data/OptionParameter.php:294
- /Valkyrja/Dispatch/Dispatcher/Dispatcher.php:76
- /Valkyrja/Dispatch/Dispatcher/Dispatcher.php:186
- /Valkyrja/Http/Message/Factory/RequestFactory.php:

## Auth

- Cli Commands
    - Use new APP_NAMESPACE env to fill in template
    - auth:make:authenticator
- Http Middleware
- Http Controllers
    - `register(RegistrationAttempt): User`
    - `forgotPassword(ForgotPasswordAttempt): User`
    - `resetPassword(ResetPasswordAttempt): User`
    - `lock(LockAttempt): User`
    - `unlock(UnlockAttempt): User`
    - Authentication
    - Registration
    - ForgotPassword
    - ResetPassword
    - Data
- SessionTokens in DB so you can dismiss certain login sessions or log all out

## Queue

- Message
    - `getQueueName`
    - `getQueueData`
    - Enum
        - QueueResult
            - Retry
            - Complete
            - Error
- Middleware
- Routing
- Server

## Cache

- Add Psr Cache support via new PsrCache manager

## Event

- Update Dispatcher with better template docblocks for each method
- Add Priority capability
- Subscriber for event
    - Is same concept as Controller for Cli/Http
    - I don't think so now that I think of it. Event dispatching should be super
      simplistic, especially since it's done usually on the fly when needed
- Stop propagation abstract event

## Application Skeleton Repository

- I want to deprecate this whole fucking thing and keep everything within the
  main framework.
    - Have cli commands to generate the necessary files to get started.
        - Really just need something to make `cli`, `public/index.php`
        - Need a bin cli app for the framework as a whole that is an aside to
          the one that a person would build

## View

- New Syn or something templating system that can be on parity with other
  languages as well, so we have one templating engine for all languages with
  templates able to be used across all the language frameworks for valkyrja
- Change Orka to Syn? Sýn (f.): View, sight, vision.
- Change Orka to Syni? Sýni (n.): Sight, look.

## Http

- Controller Namespace
    - Similar to Command in Cli with default controllers and/or actions
- Add the following content types:
    - https://stackoverflow.com/questions/23714383/what-are-all-the-possible-values-for-http-content-type-header
- Middleware docs inspiration:
    - https://botman.io/2.0/middleware

## ORM

- Statement::fetch() should return null when no data found
- `$this->getXValue()` for any `static::getXField()`
- Use defaultable service for Entity instead of Entity Matchers??
- QueryBuilderFactory::fromQuery(string $query): QueryBuilder
    - SelectQueryBuilder::fromQuery(string $query): SelectQueryBuilder
    - UpdateQueryBuilder::fromQuery(string $query): UpdateQueryBuilder
    - InsertQueryBuilder::fromQuery(string $query): InsertQueryBuilder
    - DeleteQueryBuilder::fromQuery(string $query): DeleteQueryBuilder
- docs inspiration
    - https://symfony.com/doc/current/doctrine.html#learn-more

## Validation

- Valid values
    - Allow UnitEnum or BackedEnum

## Debugging/Profiling/Debug Bar/Profiler

- `ProfilerCapableContract`
- Add ability within ServiceProviders to log what was called for later usage
    - Add this ability within a lot of classes - Should be based on a debug flag
    - Use the event dispatcher perhaps?
        - ```php
              if (debug) {
                $this->event->dispatch(new DebugMethodCalled($this, __METHOD__));
              }
          ```
    - It's either that or there's a class/trait that's used by all
      classes/methods to log to
    - Got one better:
        - `DebugContainer`
            - `get()`
                - gets the instance from the service container like normal, but
                  then creates a new instance of the `DebugClass` class that
                  takes as an arg the class returned
                    - This won't work because it'll break contracts and
                      typehinting
                    - It will have to be an anonymous class that extends the
                      instance returned I think
            - `DebugClass`
                - `call`
    - This needs to start with whatever needs debugging capabilities will have
      to implement it themselves, perhaps. We'll have to really research this
      one. But for the time being let's not focus attention here and work on
      finishing the application, and the tests

## Client

- Add PSR

## Log

- Add file logger?

## Modules Concept

- Index.php file for each module
- Composer.json for each module
- Shared composer import from lib folder

## Cli and Http and soon to be Queue

- Rename current command to Route
    - This makes sense since a Router routes Route objects to a handler
    - in Http a Route takes a Request and goes to a handler for an Action
    - in Cli a Route takes an Input and goes to a handler for a Command
    - in Queue a Route takes an Payload and goes to a handler for a Queue
    - *Action is a name for a class that handles one specific Http Route as a
      whole
    - *Controller is a name for a class that handles multiple Http Routes
    - *Command is a name for a class that handles one specific Cli Route as a
      whole
    - *Controller is a name for a class that handles multiple Cli Routes
    - *Queue is a name for a class that handles one specific Queue Route as a
      whole
    - *Controller is a name for a class that handles multiple Queue Routes
    - CommandRoute? and ActionRoute?
        - So to have a distinct name for each between Cli/Http and soon to be
          Queue?
        - You don't have an ActionRouter or CommandRouter, these are already
          distinguished via the namespace
        - I think this level of granularity is unnecessary
    - Controller houses Route definitions tied to single methods
        - For simple actions or commands this is fine.
        - For complex actions or commands a separate Action/Command file would
          be used
        - This allows for multiple methods to be used for a single
          Action/Command without muddying the others
        - And has a singular place to find all defined routes.
        - Separation of concerns

## Blogging system built on Valkyrja

## Test

- Move output from run()
- Move to tests directory

## Add callable capability to Reflector

- https://stackoverflow.com/questions/13071186/how-to-get-the-number-of-parameters-of-a-run-time-determined-callable

## Lifecycle

- App created with Env::class, and Config|Data object
    - Create Container
        - bootstrap stuff from Entry\App
    - If Config, we do normal setup
        - Config Class is slimmed down to only needing
            - Container
            - Event
            - Cli
            - Http
        - App Components are iterated through
            - Here we also ensure we call the core ones
                - ApplicationComponent
                - ContainerComponent
                - EventComponent
                - CliComponent
                - HttpComponent
                - DispatcherComponent
                - AttributeComponent
                - ReflectionComponent
            - Iterate over components
            - This is exactly what is done now, except that now no config
              classes in any component
            - $container->setSingleton(Config::class, Config)
    - If Data passed
        - Data is split apart and
            - $container->setSingleton(ContainerData::class, Data->
              ContainerData)
            - $container->setSingleton(EventData::class, Data->EventData)
            - $container->setSingleton(CliData::class, Data->CliData)
            - $container->setSingleton(HttpData::class, Data->HttpData)
    - Container->setData($container->getSingleton(ContainerData))
        - Also add a getData() Same as getCacheable now
        - Data
            - Same as Cache
        - If ContainerData was set in Data passed step
            - Then it's just used
            - Otherwise defaults to ServiceProvider set
                - Collector is called
                - Code from setupNotCached
    - Event, Http, Cli in ServiceProvider
        - Get Data from Container
            - Pass into __construct
        - ServiceProvider has an entry to get the Data (which was either passed
          with the CachedData object)
        - ORRRRRRRR a default entry is in ServiceProvider and
            - Collector is called
            - Code from setupNotCached

## Psalm issues

### Env

- https://psalm.dev/r/36fd31ac0e

### Container

- https://psalm.dev/r/4431cf022b

## Github

- Composite action
    - https://docs.github.com/en/actions/tutorials/creating-a-composite-action

## All

- Is null return cheating?
    - Should we redefine all methods to not allow null returns but throw an
      exception instead? This will give parity between PHP and other languages
      (Java/Go)
    - What about parameters that are optional?
    - " Similar to my convention: Single item queries - Create... returns a new
      instance, or throws; Get... returns an expected existing instance, or
      throws; GetOrCreate... returns an existing instance, or new instance if
      none exists, or throws; Find... returns an existing instance, if it
      exists, or null. For collection queries - Get... always returns a
      collection, which is empty if no matching items are found."
        - For single item queries:
            - `Create`... returns a new instance, or throws
            - `Get`... returns an expected existing instance, or throws
            - `Retrieve`... returns an expected existing instance, or throws
            - `GetOrCreate`... returns an existing instance, or new instance if
              none exists, or throws
            - `Find`... returns an existing instance, if it exists, or null
        - For collection queries:
            - `Get`... always returns a collection, which is empty if no
              matching[1] items are found
                - [1] This is the only one that doesn't throw an exception.
    - https://www.yegor256.com/2014/05/13/why-null-is-bad.html I agree, but
      disagree at the same time. You should probably throw exceptions instead of
      return null in most cases where a value is expected then handle said
      exception specifically
        - Take an example where a user calls into a system and asks for a phone
          operator by name
            - If the operator exists, we transfer the call
            - If the operator doesn't exist, an exception is thrown
                - The exception is handled in the middleware for the application
                  and specifically this route, and we handle that case
                  specifically; context is preserved
        - When a nullobject is used instead we just continue like normal, which
          is problematic to say the least. Context is lost, and anyway we have
          to handle an exception or null instance case, but now we have a
          nullObject floating around that didn't need to exist in the first
          place
            - Another example is a cli program where a specific command is
              called
                - if the command exists it gets called
                    - if the command requires certain parameters either those
                      params exist or an exception is thrown
                        - if an exception is thrown then the middleware should
                          catch that use case
                - if not then an exception is thrown and the middleware catches
                  it
    - Only allow null on parameters that are object.
        - array -> []
        - string -> ''
        - bool -> false
        - int -> 0
        - float -> 0.0
        - etc
- Data objects for methods/functions with more than 3 params?
    - Repeated params across codebase should be a data object

## Naming

- `validate{Something}($value): void` = ensure `$value` is valid or throw
  exception
- `invalidate{Something}($value): void` = ensure `$value` is not valid or throw
  exception
- `isValid{Something}($value)` = check if `$value` is not valid
- `modify{Something}($value): void` = modify `$value` by reference
- `get{Something}($value)` = get a modified version of `$value` without
  modifying `$value`
- `set{Something}($value)` = set `$value` without modifying host
- `with{Something}($value): self` = set `$value` on cloned host
- `process{Something}(array $value): void` = process array `$value` by reference
- `filter{Something}($value): void` = filter `$value` by reference
- `getFiltered{Something}($value)` = get a filtered version of `$value` without
  modifying `$value`
- `parse{Something}($value): void` = parse `$value` by reference
- `getParsed{Something}($value)` = get a parsed version of `$value` without
  modifying `$value`

## Benchmarking

- fc07889: 5519.52, 4957.28
- d414d51: 5541.12, 4938.62

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
