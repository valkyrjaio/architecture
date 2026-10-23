# Cli

Implements
[`CLI.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/cli/CLI.md).

## Introduction

The Cli component takes a command-line invocation and writes a result back to
the person who ran it. It parses the runtime's argument vector, matches a
**command**, runs that command's code, and writes what the command produced.

A command is a route. The component holds every command the application
declares, looks one up by the name in the invocation, binds the invocation's
values to that command's declared parameters, and calls the command's handler.
Http routes a request to a controller the same way, with the same route data,
the same collection and the same dispatch.

Cli is a **protocol component**, so it runs the framework's pipeline over one
unit of work. The unit of work is an **input**, and the outcome is an
**output**.

| Http             | Cli            |
| ---------------- | -------------- |
| `Request`        | `Input`        |
| `Response`       | `Output`       |
| `RequestHandler` | `InputHandler` |

Everything a person reads or types is a value. An input holds the caller, the
command name, the arguments and the options. An output holds messages, and each
message names the formatter it wants rather than carrying rendered text. So the
same command renders in color on a terminal and plainly in a pipe.

This document assumes the framework, where the container resolves the handler
and a provider registers the commands.
[Container](../container/README.md) covers contracts, bindings and the resolve
operations.

## Declaring a command

A command states its name, its description, its handler, and the arguments and
options it accepts.

A **route provider** is where an application declares its commands. The
provider declares them two ways, and one provider uses both:

| Method                 | Returns                                   |
| ---------------------- | ----------------------------------------- |
| `getRoutes`            | commands built as data                    |
| `getControllerClasses` | the classes to scan for a declared marker |

```text
class AppCliRouteProvider implements CliRouteProviderContract {

    getControllerClasses() {
        return [GreetCommand, UserPurgeCommand]
    }

    getRoutes() {
        return [
            new Route(
                name: "cache:clear",
                description: "Clear the application cache",
                handler: [self, "cacheClearHandler"],
                arguments: [
                    new ArgumentParameter(
                        name: "store",
                        description: "The cache store to clear",
                        mode: ArgumentMode.OPTIONAL,
                    ),
                ],
                options: [
                    new OptionParameter(
                        name: "force",
                        description: "Clear without confirmation",
                        shortNames: ["f"],
                    ),
                ],
            ),
        ]
    }
}
```

The list is a literal with no conditional logic. The build tool reads it as
written rather than running it.

A component provider returns the route providers from `getCliProviders`, and
the application's CLI config lists the component providers
([`PROVIDERS.md`](../../convention/PROVIDERS.md)).

```text
class AppComponentProvider implements ComponentProviderContract {

    getCliProviders(app) {
        return [new AppCliRouteProvider()]
    }
}
```

### Declaring on the method that runs the command

A port whose language carries a marker declares the command beside the code
that runs it. The collector reads each class `getControllerClasses` returns and
turns every marker on it into a command. A port spells the marker its own way.

```text
class UserPurgeCommand {

    // Reached as `user:purge` and as `user:prune`.
    @Route(name: "user:purge", description: "Purge deleted users")
    @Route(name: "user:prune", description: "Purge deleted users")
    @RouteHandler([AppCliRouteProvider, "userPurgeHandler"])
    run() { }
}
```

The marker is repeatable, so one method serves several command names.

**A marker holds a constant in most languages, so it cannot hold a closure.**
Name the handler with the handler marker, which takes a named reference. A
command built as data takes a closure where the language has one.

### Group related commands by prefixing the name

`list` prints the commands in name order, so a shared prefix shows a group
together. Name each one with its group, as `user:greet` and `user:purge`.

A port that carries a name marker prefixes every command a class declares,
joined with a dot. On the class the segment is a prefix, and on a method it is a
suffix.

```text
@Name("user")
class UserCommands {

    // Reached as `user.greet`.
    @Route(name: "greet", description: "Greet a user")
    greet() { }
}
```

## Writing a handler

A handler takes the container and the matched route, and it returns an output
([`HANDLERS.md`](../../convention/HANDLERS.md)). It is the one thing the
component calls, so it resolves the command class and calls it.

```text
static greetHandler(container, route) {
    return container.getSingleton(GreetCommand).run()
}
```

The matched route is registered in the container before the handler runs. So a
command class takes the route through its constructor, and it reads its values
from there.

```text
class GreetCommand {

    construct(route: RouteContract, outputFactory: OutputFactoryContract) { }

    run() {
        name = this.route.getArgumentValue("name")
        shout = this.route.hasProvidedOption("shout")

        greeting = "Hello, " + name + "!"

        return this.outputFactory
            .createOutput()
            .withMessages(
                new Message(shout ? upperCase(greeting) : greeting),
                new NewLine(),
            )
    }
}
```

Register the command class in a service provider, the same as any other
service. A port ships a `Controller` base whose constructor takes the input and
the output factory. Extend it, or take what the command needs directly.

The input is registered in the container as well, and so is the final output
after dispatch.

## Arguments and options

An **argument** is positional. An **option** is named, and it carries one long
name and any number of short names, so one option answers to several spellings.

| Value              | On an argument              | On an option                                |
| ------------------ | --------------------------- | ------------------------------------------- |
| `name`             | required                    | required, and it is the long name           |
| `description`      | required                    | required                                    |
| `cast`             | the type the values cast to | the type the values cast to                 |
| `mode`             | `REQUIRED` or `OPTIONAL`    | `REQUIRED` or `OPTIONAL`                    |
| `valueMode`        | `DEFAULT` or `ARRAY`        | `NONE`, `DEFAULT` or `ARRAY`                |
| `shortNames`       | —                           | the one-letter spellings of the same option |
| `defaultValue`     | —                           | what a read falls back to                   |
| `validValues`      | —                           | the values the invocation may give          |
| `valueDisplayName` | —                           | the placeholder the help page prints        |

Give each one a description. The help page prints that description, and it is
the only place a person reads what the parameter is for.

The component ships a ready-made option parameter for each global option, so a
command that wants one adds it rather than declaring it again.

### An argument takes one value, or every value left

A `DEFAULT` argument takes at most one value, and the router binds the
invocation's positional values to the declared arguments in declaration order.
An `ARRAY` argument takes every positional value left, so declare it last.

```text
@Route(name: "files:process", description: "Process files")
@ArgumentParameter(
    name: "paths",
    description: "The file paths to process",
    mode: ArgumentMode.REQUIRED,
    valueMode: ArgumentValueMode.ARRAY,
)
run() {
    // myapp files:process a.txt b.txt c.txt
    for (argument in this.route.getArgument("paths").getArguments()) {
        path = argument.getValue()
    }
}
```

### An option takes a value, takes none, or repeats

| Value mode | The option                                     |
| ---------- | ---------------------------------------------- |
| `DEFAULT`  | takes one value                                |
| `NONE`     | takes no value, and is a flag                  |
| `ARRAY`    | repeats, and collects every value it was given |

```text
@Route(name: "user:export", description: "Export users")
@OptionParameter(
    name: "format",
    description: "The export format",
    valueDisplayName: "format",
    defaultValue: "json",
    shortNames: ["f"],
    validValues: ["json", "csv"],
)
run() {
    // myapp user:export --format=csv
    // myapp user:export -f=csv
    format = this.route.getOptionValue("format")
}
```

**Read a flag with the provided question, not with a value.** A `NONE` option
carries no value at all, so a read of its value answers nothing.

```text
// Wrong. A `NONE` option never holds a value.
isDryRun = this.route.getOptionValue("dry-run") != ""

// Right. The question is whether the invocation gave the option.
isDryRun = this.route.hasProvidedOption("dry-run")
```

### Reading the values off the route

| Method                                      | Answers                                    |
| ------------------------------------------- | ------------------------------------------ |
| `hasArgument` / `hasOption`                 | whether the command declares the parameter |
| `hasProvidedArgument` / `hasProvidedOption` | whether the invocation gave it             |
| `getArgumentValue` / `getOptionValue`       | the first value, or a default              |
| `getArgument` / `getOption`                 | the parameter itself, with every value     |

The router keeps every declared parameter on the route, whatever the invocation
gave. So the first pair answers what the command declares, and the second pair
answers what the person typed.

A value read takes the parameter name first and a default second. An option
reads three sources in order: the first value the invocation gave, the default
the call site passes, then the `defaultValue` the option declares. An argument
declares no default, so it reads the first two.

```text
// The `format` option declares a default value of "json".
format = route.getOptionValue("format")

namespace = route.getArgumentValue("namespace", "all")
```

Match an option by its long name. A short name is a spelling of the invocation,
not a name the route is keyed by.

### A declared type is validated before the handler runs, and cast on read

The router checks every value against its declaration before it calls the
handler. A missing required parameter, a repeated single-value option, a value
outside the valid values, and a value on an option that takes none each fail
there. So no handler code runs on an invalid invocation, and the failure
reaches the throwable stage as an error output.

The cast itself happens on read. A parameter holds the cast and converts
nothing. The **caster** takes the parameter and returns its values cast, and
the caster holds the container
([`STATIC_METHODS.md`](../../convention/STATIC_METHODS.md)).

```text
class WorkCommand {

    construct(route: RouteContract, caster: CasterContract) { }

    @Route(name: "queue:work", description: "Work the queue")
    @ArgumentParameter(
        name: "workers",
        description: "The worker count",
        cast: new Cast(CastType.int),
        mode: ArgumentMode.REQUIRED,
    )
    run() {
        // myapp queue:work 4
        [workers] = this.caster.getCastValues(this.route.getArgument("workers"))
    }
}
```

The route's own readers hand back the raw string, so a handler that wants the
typed value asks the caster. A parameter that declares no cast gives the caster
nothing to apply, and its values come back as they arrived.

**Register a cast type as a service.** The caster resolves the cast's type with
the operation for a service, and that operation is the one that takes the value
as an argument ([Container](../container/README.md)). The framework registers
every type its own cast types name, so a framework cast needs nothing further.

An application that casts to a type of its own registers that type under the id
the cast names, and reads the raw value from the argument the caster passes.

## Reading the input

An input is the parsed invocation.

| Method           | Returns                                     |
| ---------------- | ------------------------------------------- |
| `getCaller`      | the binary the person ran                   |
| `getCommandName` | the command name, or the configured default |
| `getArguments`   | every positional value, in order            |
| `getOptions`     | every option the invocation gave            |
| `getOption`      | every value given for one name              |
| `hasOption`      | whether the invocation gave the name        |

```text
input = container.getSingleton(InputContract)

caller = input.getCaller()           // "myapp"
command = input.getCommandName()     // "user:greet"
formats = input.getOption("format")
```

`getOption` returns a list, because an option repeats. An argument carries its
value, and an option carries its name, its value, and whether it was spelled
long or short.

**Prefer the route's parameters inside a command.** The route's values are
validated, and the caster casts them. Read the input for an option no command
declares, or for a positional value outside the declared parameters.

An input is immutable. Each writing method returns a copy.

### What the parser does with each token

The parsed result is identical in every port. How the runtime shapes the vector
is the one legitimate per-port difference. The rules, and the index each port
reads the command name from, are in
[`PORTS.md`](../../convention/PORTS.md).

- The first token is the **caller**.
- The second token is the **command name**, where it is neither an option nor
  the end-of-options marker. Where it is one of those, the configured default
  command name applies.
- Every later token that does not start with a dash is an **argument** value.
- Every token that starts with a dash is an **option**, wherever it sits on the
  line.

An option token takes these forms:

- `--name=value` sets a long option's value. The parser splits on the first
  `=`, so `--filter=a=b` keeps the value `a=b`.
- `--name` sets a long option with no value, which is a flag.
- `-n=value` sets a short option's value, and `-n` sets a short flag.
- `-abc` is the three short flags `-a`, `-b` and `-c`. A combined group carries
  no value.
- `--` ends option parsing. Every later token is an argument, including one
  that starts with a dash.
- A lone `-` is an argument. It names standard input by convention. At the
  second position it sets the command name, the same as any other token there.

```text
myapp user:greet Melech --shout -n=3 -- --literal-argument
```

The caller is `myapp` and the command is `user:greet`. The arguments are
`Melech` and `--literal-argument`. The options are `shout`, with no value, and
`n`, with the value `3`.

## Writing output

An output carries messages and an exit code. Build one through the **output
factory**, add messages, and return it.

```text
return this.outputFactory
    .createOutput(exitCode: ExitCode.SUCCESS)
    .withMessages(
        new SuccessMessage("Done."),
        new NewLine(),
    )
```

An output is immutable, like every message type in the framework.

| Method                                   | Does                                  |
| ---------------------------------------- | ------------------------------------- |
| `withMessages`                           | replaces the messages not yet written |
| `withAddedMessage` / `withAddedMessages` | appends                               |
| `withExitCode`                           | sets the exit code                    |
| `writeMessages`                          | writes, and records what it wrote     |

Every one of them returns a copy. Return the copy.

The input handler writes the messages after the command returns, so a command
queues messages rather than writing them. A command that must write early calls
the write itself and keeps the copy the call returns.

### Where the output goes

The factory builds one output per destination.

| Factory method       | Writes to                                   |
| -------------------- | ------------------------------------------- |
| `createOutput`       | the standard output stream, with formatting |
| `createPlainOutput`  | the standard output stream, with none       |
| `createFileOutput`   | a file, which it appends to                 |
| `createStreamOutput` | a stream the caller holds                   |
| `createEmptyOutput`  | nowhere                                     |

Each one takes the exit code and any number of messages, and each copies the
run's three interaction flags from the interaction config.

Warning: a flag suppresses a file write and a stream write as well as a
terminal write. Construct the output directly where the destination must take
the messages whatever the flags say.

A file output never truncates, so it keeps what each earlier run wrote. The
caller owns the truncation.

### A message, a formatter, a format and a writer

Four pieces produce one line, and each one holds a separate decision.

| Piece       | Holds                                                  |
| ----------- | ------------------------------------------------------ |
| a message   | the text, and the formatter it wants                   |
| a formatter | the formats one style is made of                       |
| a format    | one attribute of a style, and the codes that switch it |
| a writer    | the handling that one kind of message needs            |

A message names the formatter it wants. It never writes an escape sequence
itself, so the output decides whether the style reaches the destination at all.

| Message          | Is                                                      |
| ---------------- | ------------------------------------------------------- |
| `Message`        | a line of text, with an optional formatter              |
| `SuccessMessage` | a line that reports work that completed                 |
| `WarningMessage` | a line that reports something a person should notice    |
| `ErrorMessage`   | a line that reports a failure                           |
| `Banner`         | a padded block around a message, in the message's style |
| `NewLine`        | vertical space                                          |
| `Messages`       | several messages, concatenated                          |
| `Header`         | the application's identity block                        |
| `Question`       | a prompt that waits for an answer                       |
| `Answer`         | what a person typed, and the validation of it           |
| `Progress`       | the state of work that takes long enough to report      |

A message carries no line break of its own. Compose one with a new line.

```text
output = output.withAddedMessages(
    new Banner(new ErrorMessage("Deployment failed.")),
    new NewLine(),
    new Message("Check the log for details."),
    new NewLine(),
)
```

A formatter holds the formats a style is made of. There is one format per
attribute: the text color, the background color, and the style itself.

```text
message = new Message(
    "Heads up!",
    new Formatter(
        new TextColorFormat(TextColor.CYAN),
        new StyleFormat(Style.BOLD),
    ),
)
```

The component ships a formatter for each message kind it ships, plus one for a
highlighted run of text. Extend the base formatter for a style of your own.

A writer handles one kind of message. It answers whether a message is its own,
and it does the handling for the one it claims. Every output carries the
question writer, because a question needs a read rather than a write.

```text
class ProgressWriter implements WriterContract {

    shouldWriteMessage(message) {
        return message implements ProgressContract
    }

    write(output, message) {
        // ...
    }
}

output = output.withWriters(new QuestionWriter(), new ProgressWriter())
```

Warning: a writer list replaces the list the output holds. Include the question
writer, or the output writes a question's text and reads no answer.

### The identity header

A command prints a header that identifies the application and names the work it
is about to do. The header holds the application's name and version, an icon,
the framework version the application was built against, the language runtime
version, the project root, a description of the action, and the command name.

The framework version and the runtime version are separate lines. They describe
different relationships, and they change for different reasons.

```text
output = this.outputFactory
    .createOutput()
    .withMessages(new Header(appName, appVersion, this.route))
```

The header reads the action's description and the command name off the matched
route. It reads the application's name and version from the application's
config, and it reads the icon, the framework version and the runtime version
from the framework's own info constants. Each value has a writing method, so a
command overrides any one of them.

### Structure the output before you fill it

A command produces the header, then its work, then a summary. Four rules hold
in every port:

- **A step reports one line.** A detail line appears under a step only where
  that step has something to say, so a run with nothing to report stays short.
- **A status label is one word**, and the set of labels is closed.
- **Color is decorative.** The text carries every fact that color carries, so a
  pipe and a reader with no color lose nothing.
- **One structure handles every outcome.** The layout that reports success
  reports partial failure, total failure and a run that did nothing. A port
  does not design a second layout for a failure.

### Asking a question

A question is a message that waits. It pairs with an **answer**, which holds
the default response, the allowed responses, and an optional validation
callable. The question's own callable receives the output and the final answer,
and it returns the output to continue with.

```text
question = new Question(
    "Drop all tables?",
    function (output, answer) {
        if (answer.getUserResponse() != "yes") {
            return output
        }

        return output.withAddedMessages(
            new SuccessMessage("Tables dropped."),
            new NewLine(),
        )
    },
    new Answer(defaultResponse: "no", allowedResponses: ["yes", "no"]),
)

return this.outputFactory
    .createOutput()
    .withMessages(question)
```

The rendered prompt lists the allowed responses and the default. An empty
response leaves the answer as it was, so the default stands. A response outside
the allowed list asks the question again, unless the validation callable
accepts it.

**Write a question as though the run is unattended.** The framework supplies
the default answer at every level that reads no answer, and a closed or empty
standard input behaves the same way. So a command written this way also runs in
a pipeline.

### The run's level

A run carries three independent flags, and the combination of them decides what
reaches the person.

| Level           | Means                                                 |
| --------------- | ----------------------------------------------------- |
| interactive     | questions are asked, and answers are read             |
| non-interactive | questions are not asked, and a default answer is used |
| quiet           | nothing is written for a successful run               |
| silent          | nothing is written, ever                              |

**Quiet keys on the outcome, not on the message kind.** A run that succeeds
writes nothing at all, and a run that fails writes everything it would normally
write. So quiet suppresses the output of a run nobody needs to read, rather
than filtering the messages of every run.

A question reads an answer only on an interactive run that is neither quiet nor
silent. Every other run keeps the answer the question already holds.

The three flags live on one interaction config, and the option that selects
each level carries its own config contract.
[Renaming a built-in command or option](#renaming-a-built-in-command-or-option)
covers those contracts.

A command overrides any flag on an output of its own.

### Report the outcome with an exit code

Return the code that names what happened, from the enum rather than as a bare
integer. The process exits with the code the output reports.

| Case             | Value | Reports                            |
| ---------------- | ----- | ---------------------------------- |
| `SUCCESS`        | 0     | a normal exit                      |
| `ERROR`          | 1     | a general error                    |
| `USAGE_ERROR`    | 64    | a command used incorrectly         |
| `DATA_ERROR`     | 65    | bad input data                     |
| `NO_INPUT`       | 67    | an input that was not found        |
| `NO_USER`        | 68    | a user that does not exist         |
| `UNAVAILABLE`    | 69    | a service that is unavailable      |
| `SOFTWARE_ERROR` | 70    | an internal software error         |
| `OS_ERROR`       | 71    | an operating system error          |
| `OS_FILE_ERROR`  | 72    | an operating system file error     |
| `CANT_CREATE`    | 73    | an output file it cannot create    |
| `IO_ERROR`       | 74    | a read or a write that failed      |
| `TEMP_FAIL`      | 75    | a temporary failure, worth a retry |
| `PROTOCOL_ERROR` | 76    | a remote error in a protocol       |
| `NO_PERMISSION`  | 77    | permission denied                  |
| `CONFIG_ERROR`   | 78    | a configuration error              |
| `AUTO_EXIT`      | 255   | a reserved outcome                 |

**A warning is a message, not an outcome.** There is no warning exit code, so a
run that warns and succeeds exits successfully and fails no build. An
application that wants a warning to fail returns a failing code itself.

The exit code also takes a plain integer, for a code the enum does not name.

## The commands the component ships

Every port ships the same four.

| Command     | Does                                           |
| ----------- | ---------------------------------------------- |
| `list`      | lists every command, with its description      |
| `help`      | prints the help page for one command           |
| `version`   | reports the application version                |
| `list:bash` | prints the command names, for shell completion |

```text
myapp list                     // every command
myapp list -n=user             // the commands whose name starts with `user`
myapp help --command=list      // the help page for one command
myapp list --help              // the same page
myapp version                  // the version
myapp list:bash myapp          // the command names, for shell completion
```

The help page prints the command's name, its description and its usage line,
then each declared parameter with its description, its mode, its value
placeholder, its valid values and its default value. The command's own help
text comes last. A command supplies that help text as a named reference that
returns a message, and a command that supplies none keeps the rest of the page.

**A command whose whole purpose is machine-readable output prints no header.**
`list:bash` is the one, because a shell reads its output. Every other command
prints one.

A global option works on every command.

| Option             | Short | Does                                |
| ------------------ | ----- | ----------------------------------- |
| `--help`           | `-h`  | prints the command's help page      |
| `--version`        | `-v`  | reports the application version     |
| `--quiet`          | `-q`  | writes nothing for a successful run |
| `--silent`         | `-s`  | writes nothing, and reads no answer |
| `--no-interaction` | `-N`  | reads no answer                     |

The first two rewrite the invocation. An input that carries `--help` becomes a
`help` invocation for the command that was named, so the help page and the
`help` command print the same thing. The other three set the run's flags.

Another component ships commands of its own through the same route-provider
contract. Http routing ships one that lists the application's routes.

## Middleware

Middleware runs at a **stage** of the pipeline. Each stage has one contract and
one handler, and a middleware class implements the contract of every stage it
runs at. A middleware forwards the call through the handler chain, or it
answers with an output and stops the chain.

| Stage             | Contract                            | Runs                                     | Per command |
| ----------------- | ----------------------------------- | ---------------------------------------- | ----------- |
| `InputReceived`   | `InputReceivedMiddlewareContract`   | before the command is matched            | no          |
| `RouteMatched`    | `RouteMatchedMiddlewareContract`    | after the match, before the handler      | yes         |
| `RouteNotMatched` | `RouteNotMatchedMiddlewareContract` | where no command matched                 | no          |
| `RouteDispatched` | `RouteDispatchedMiddlewareContract` | after the handler produced an output     | yes         |
| `ThrowableCaught` | `ThrowableCaughtMiddlewareContract` | where a throwable was caught             | yes         |
| `ProcessExiting`  | `ProcessExitingMiddlewareContract`  | after the write, before the process ends | yes         |

Cli has no stage between the final output and the write. A command writes
output at any point, so there is no single moment for such a stage to occupy.
The stage numbering and what each stage guarantees are in
[`LIFECYCLE.md`](../../convention/LIFECYCLE.md).

### What each stage takes and returns

| Stage             | Takes                             | Returns                |
| ----------------- | --------------------------------- | ---------------------- |
| `InputReceived`   | the input                         | an input, or an output |
| `RouteMatched`    | the input, and the matched route  | a route, or an output  |
| `RouteNotMatched` | the input, and an error output    | an output              |
| `RouteDispatched` | the input, the output, the route  | an output              |
| `ThrowableCaught` | the input, an output, a throwable | an output              |
| `ProcessExiting`  | the input, and the written output | nothing                |

Stage 1 returns an input, so a middleware there rewrites the invocation. The
built-in help and version middleware do that. Stage 2 returns the route, so a
middleware there rewrites the matched command. Each of the two also answers
with an output, which skips every stage up to the write.

```text
class MaintenanceModeMiddleware implements InputReceivedMiddlewareContract {

    construct(outputFactory: OutputFactoryContract, isDown: bool) { }

    inputReceived(input, handler) {
        if (this.isDown) {
            return this.outputFactory
                .createOutput(exitCode: ExitCode.UNAVAILABLE)
                .withMessages(new ErrorMessage("The application is down."))
        }

        return handler.inputReceived(input)
    }
}
```

The throwable stage receives a default error output built from the throwable,
and returning an output there is what turns a failure into a report. The stage
also receives a throwable that no command raised, because the write of the
output runs after the command returned.

The exiting stage returns nothing. The output is written already, so nothing
there changes what the person read. Deferred work belongs here, and it costs
the person nothing.

```text
class FlushLogsMiddleware implements ProcessExitingMiddlewareContract {

    construct(logger: LoggerContract) { }

    processExiting(input, output, handler) {
        this.logger.info("Process exiting.")

        handler.processExiting(input, output)
    }
}
```

Warning: a throwable in the exiting stage reaches no throwable middleware, so
the framework writes no log entry for it. Report a failure from inside the
middleware where a run must record one.

### Scheduling middleware

Middleware is scheduled **globally** in the config, one property per stage, or
**per command** on the route. Four of the six stages take a per-command list.

```text
@Route(
    name: "db:migrate",
    description: "Run database migrations",
    routeMatchedMiddleware: [ProductionGuardMiddleware],
)
@Middleware(CommandAuditMiddleware)
run() { }
```

The route's lists name a stage each. The middleware marker names no stage, and
the collector sorts the class into every stage whose contract it implements.

Both lists run at the stage they are scheduled for, global first. **Neither one
is deduplicated**, so a class scheduled twice at one stage runs twice. A
duplicate is the application's own defect, and the framework does not hide it.

The stage handlers resolve each middleware class from the container, so
register your middleware the same way you register a command class.

## Configuration

The component reads the application's CLI config, the interaction config, and
one config contract per renameable built-in.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how
you override, and implementing nothing leaves every default in place.

### The CLI config

| Property                    | Holds                                            | Default                                      |
| --------------------------- | ------------------------------------------------ | -------------------------------------------- |
| `defaultCommandName`        | the command that runs where the input names none | `list`                                       |
| `inputReceivedMiddleware`   | the middleware scheduled at stage 1              | the help, version and interaction middleware |
| `routeMatchedMiddleware`    | the middleware scheduled at stage 2              | none                                         |
| `routeNotMatchedMiddleware` | the middleware scheduled at stage 3              | the typo middleware                          |
| `routeDispatchedMiddleware` | the middleware scheduled at stage 4              | none                                         |
| `throwableCaughtMiddleware` | the middleware scheduled at stage 5              | the log and output middleware                |
| `processExitingMiddleware`  | the middleware scheduled at stage 7              | none                                         |

A list you pass replaces the default, so name the defaults you keep.

```text
class Config implements ConfigContract, CliConfigContract {
    defaultCommandName = "list"

    inputReceivedMiddleware = [
        CheckForHelpOptionsMiddleware,
        CheckForVersionOptionsMiddleware,
        CheckGlobalInteractionOptionsMiddleware,
        MaintenanceModeMiddleware,
    ]
}
```

### The interaction config

| Property        | Holds                              | Default |
| --------------- | ---------------------------------- | ------- |
| `isInteractive` | whether a question reads an answer | `true`  |
| `isQuiet`       | whether a successful run writes    | `false` |
| `isSilent`      | whether any run writes             | `false` |

The output factory copies the three flags onto every output it builds. The
interaction middleware sets them from the global options, so every command
honors those options with no code of its own.

### Renaming a built-in command or option

Each renameable built-in carries its own config contract, so an application
renames one without constructing the others. Each property holds a name with no
leading dash.

| Contract                             | Properties                                                          | Renames                                |
| ------------------------------------ | ------------------------------------------------------------------- | -------------------------------------- |
| `CliHelpCommandConfigContract`       | `helpCommandName`, `helpOptionName`, `helpOptionShortName`          | the `help` command, and `--help`       |
| `CliVersionCommandConfigContract`    | `versionCommandName`, `versionOptionName`, `versionOptionShortName` | the `version` command, and `--version` |
| `CliNoInteractionConfigContract`     | `noInteractionOptionName`, `noInteractionOptionShortName`           | `--no-interaction`                     |
| `CliQuietInteractionConfigContract`  | `quietOptionName`, `quietOptionShortName`                           | `--quiet`                              |
| `CliSilentInteractionConfigContract` | `silentOptionName`, `silentOptionShortName`                         | `--silent`                             |

```text
class Config implements ConfigContract, CliQuietInteractionConfigContract {
    quietOptionName = "hush"
    quietOptionShortName = "H"
}
```

With that config, `--hush` and `-H` set the quiet flag, and `--quiet` no longer
does.

Warning: a command name in one of these contracts moves where the option
routes, and it does not rename the command itself. The shipped command keeps
the name it declares. Declare a command under the new name as well, or the
option routes to a command that does not exist.

The two listing commands take their names from a constant and are not
renameable.

## Service registration

Each of the component's four areas carries a provider of its own.

### Interaction

| Binding                        | Is                                  |
| ------------------------------ | ----------------------------------- |
| `CliInteractionConfigContract` | your application's own config class |
| `OutputFactoryContract`        | the output factory                  |

### Routing

| Binding                   | Is                                                     |
| ------------------------- | ------------------------------------------------------ |
| `RouteCollectionContract` | the commands, and the lookup the dispatcher uses       |
| `RouteCollectorContract`  | the collector for the way this port declares a command |
| `CasterContract`          | the caster                                             |
| `RouterContract`          | the dispatcher                                         |
| the routing data          | the collected commands, as data                        |

The collection and the collector are separate. A cached run loads the
collection from data and runs no collector.

### Middleware

| Binding               | Is                             |
| --------------------- | ------------------------------ |
| one handler per stage | the chain the config scheduled |

Each handler is built, then the config's list for that stage is added to it.

### Server

| Binding                   | Is                                                  |
| ------------------------- | --------------------------------------------------- |
| `InputHandlerContract`    | the handler that runs an invocation end to end      |
| each built-in command     | bound under its own name                            |
| each server middleware    | bound under its own name                            |
| the five config contracts | your application's own config class, or the default |

```text
class CliServerServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            CliHelpCommandConfigContract: [self, "publishHelpCommandConfig"],
            InputHandlerContract: [self, "publishInputHandler"],
            ListCommand: [self, "publishListCommand"],
            CheckForHelpOptionsMiddleware: [self, "publishCheckForHelpOptionsMiddleware"],
        }
    }

    static publishHelpCommandConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements CliHelpCommandConfigContract) {
            container.setSingleton(CliHelpCommandConfigContract, config)

            return
        }

        container.setSingleton(
            CliHelpCommandConfigContract,
            new CliHelpCommandConfig(),
        )
    }

    static publishInputHandler(container) {
        container.setSingleton(InputHandlerContract, new InputHandler(
            container: container,
            router: container.getSingleton(RouterContract),
            inputReceivedHandler: container.getSingleton(InputReceivedHandlerContract),
            throwableCaughtHandler: container.getSingleton(ThrowableCaughtHandlerContract),
            processExitingHandler: container.getSingleton(ProcessExitingHandlerContract),
            interactionConfig: container.getSingleton(CliInteractionConfigContract),
        ))
    }
}
```

**The input handler binding is the extension point.** An application that wants
a different handler binds its own class under the same contract in a provider
of its own. The entry point resolves the contract, so it takes whatever the
application bound.
