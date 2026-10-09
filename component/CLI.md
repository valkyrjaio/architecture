# Cli

The Cli component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Cli takes an input and produces an output. It is a **protocol component**, so it
runs the pipeline in [`LIFECYCLE.md`](../convention/LIFECYCLE.md) and shares that
pipeline's shape with Http, Grpc and Queue.

A command is a route. The component routes an input to a command the same way
Http routes a request to a controller, with the same route data, the same
collection and the same dispatch.

---

## Hierarchy

| Subcomponent  | Holds                                                         |
| ------------- | ------------------------------------------------------------- |
| `Interaction` | the input, the output, and everything a person reads or types |
| `Routing`     | route data, collection, matching, and dispatch                |
| `Middleware`  | the stage contracts and the stage handlers                    |
| `Server`      | the input handler, and the built-in commands                  |
| `Throwable`   | the component's throwable contract and its exceptions         |

`Interaction` is Cli's name for what Http calls `Message`. The unit of work is a
conversation rather than a message, so the subcomponent is named for that.

---

## Using it

### Configure it and point a runtime at it

The application's CLI config declares the middleware scheduled at each stage, one
property per stage, and the default command name
([`APPLICATION.md`](APPLICATION.md)). An entry point builds the input from the
runtime's argument vector and drives the component.

The interaction config carries the level as three independent flags —
interactive, quiet and silent — so a run's level is the combination of them
rather than one of four named values.

The option that **selects** each level is configured separately: each has its own
contract holding that option's long and short name, so an application renames one
without constructing the others.

### Declare commands

A command is a route. Declare commands in a **route provider** listed in the
config, as a literal list, or declare one on the controller method where the
language allows it and list the controller classes to scan
([`PROVIDERS.md`](../convention/PROVIDERS.md)).

A command states its name, its description, its handler, and the arguments and
options it accepts. **Group related commands by prefixing the name**, so a
listing shows them together.

### Declare arguments and options

An argument is positional. An option is named, with one long name and any number
of short names. Each declares whether it is required and the type its value casts
to, and an option also declares whether it takes a value or takes none.

Give each one a description, because that description is what help text prints.

**A declared type is validated before the handler runs, and cast when it is
read.** The router checks every value against its declaration first, so an
invalid value fails before any handler code runs. The cast itself happens on
read, and the route's plain readers hand back the raw string — a handler that
wants the typed value asks for the cast values.

### Write a handler

A handler receives the container and the route, and returns an output
([`HANDLERS.md`](../convention/HANDLERS.md)). It reads its arguments and options
from the route.

### Build the output

Add messages to the output and return it. Each message names the **formatter** it
wants rather than carrying rendered text, so the same output renders plainly where
color is unavailable.

The output is immutable: adding a message returns a copy, and so does writing it.
Return the copy.

### Ask a question

A question is a message that waits. Ask one only where the run is interactive;
the framework supplies the default answer at every other level, so a command
written this way also runs unattended in a pipeline.

### Report the outcome

Return the exit code that names what happened, from the enum rather than a bare
integer. A warning exits zero, so a warning does not fail a build; an application
that wants otherwise opts in.

### Attach middleware

Middleware is scheduled **globally** in the config, per stage, or **per command**
on the route. Both run at the stage they are scheduled for, global first, and
neither is deduplicated.

---

## The input

An input holds the caller, the command name, the arguments and the options. A
port builds it from the runtime's own argument vector, and the parsed result is
identical in every port.

**How the vector is shaped is the one legitimate per-port difference**, because
each runtime hands over a differently shaped vector. The rules, and the index each
port reads the command name from, are in
[`PORTS.md`](../convention/PORTS.md).

An **argument** is positional. An **option** is named, and carries one long name
and **any number of short names**, so one option answers to several spellings.

Each declares whether it is required and the type its value casts to. An option
also declares whether it takes a value, and may declare that it takes none; an
argument always takes one, and may declare that it takes many.

---

## The output

An output is **immutable**, like every message type in the framework. A writer
adds to it and returns a copy, and writing the output to a stream returns a copy
that records the write.

An output holds **messages**, and a message holds **formatting** rather than
pre-rendered text. So the same message renders plainly when color is
unavailable.

| Message kind | Is                                                      |
| ------------ | ------------------------------------------------------- |
| plain        | a line of text                                          |
| success      | a line that reports work that completed                 |
| warning      | a line that reports something a person should notice    |
| error        | a line that reports a failure                           |
| banner       | a padded block around a message, in the message's style |
| header       | the application's identity block                        |
| question     | a prompt that waits for an answer                       |
| answer       | what a person typed, and the validation of it           |
| progress     | the state of work that takes long enough to report      |
| new line     | vertical space                                          |

A **formatter** carries the style a message renders with. A **format** carries one
attribute of a style. A message names the formatter it wants; it never writes an
escape sequence itself.

---

## The identity header

Every built-in command prints a header that identifies the application and the
work it is about to do. The header holds the application's name and version, an
icon, the framework version it was built against, the language runtime version,
the project root, a description of the action, and the exact command the person
typed.

**The framework version and the runtime version are separate lines**, because
they describe different relationships and change for different reasons.

The values come from the application's own info constants and from the matched
route's description and name. An application overrides any of them.

---

## Output structure

Output is organized before it is filled. A command produces, in order, the
identity header, then its work, then a summary.

Rules every port follows:

- **A step reports one line.** Detail lines appear under a step only when that
  step has something to say, so a run with nothing to report stays short.
- **A status label is one word**, and the set is closed.
- **Color is decorative.** Every fact that color conveys is also conveyed by the
  text, so an output pipe or a reader with no color loses nothing.
- **The structure that handles success also handles partial failure, total
  failure and a run that did nothing.** A port does not design a second layout
  for a failure.

---

## Exit codes

A command returns an exit code, and the code is an enum rather than a bare
integer. The set covers success, a general error, and the specific failures a
command can report: a usage error, a data error, a missing input, an
unavailable service, a permission failure, a configuration error, and the rest.

**A warning is a message, not an outcome.** There is no warning exit code, so
emitting a warning does not change what the command returns: a run that warns and
succeeds exits successfully, and nothing fails a build on its own. An application
that wants a warning to fail returns a failing code itself.

---

## Interactivity

A run has a level, and the level decides what reaches the person:

| Level           | Means                                                 |
| --------------- | ----------------------------------------------------- |
| interactive     | questions are asked, and answers are read             |
| non-interactive | questions are not asked, and a default answer is used |
| quiet           | nothing is written for a successful run               |
| silent          | nothing is written, ever                              |

**Quiet keys on the outcome, not on the message kind.** A run that succeeds writes
nothing at all, and a run that fails writes everything it would normally write. So
quiet suppresses the output of a run nobody needs to read, rather than filtering
the messages of every run.

The three flags live on one interaction config. What is split per contract is the
**option name** that selects each level, which is why renaming one option does not
mean constructing the rest
([`COMPONENT_CONFIG.md`](../convention/COMPONENT_CONFIG.md)).

---

## Built-in commands

Every port ships the same four: list the commands, show help for one command,
report the version, and emit a machine-readable command list for shell
completion.

**A command whose whole purpose is machine-readable output prints no header.**
Every other command prints one, and an application suppresses it globally if it
wants to.

**The help and version commands carry a configurable name**, as do the built-in
options, so an application renames one it conflicts with. The two listing
commands take their names from a constant and are not renameable.

---

## Why Cli has no stage 6

A command writes output at any point in the pipeline, so there is no single
moment before the write for a stage to occupy. Every other protocol produces its
outcome once, at one place. See [`LIFECYCLE.md`](../convention/LIFECYCLE.md).

---

## Permitted variation

| Variation                           | Reason                                               |
| ----------------------------------- | ---------------------------------------------------- |
| the argument vector's shape         | each runtime hands over a different vector           |
| an `Attribute` routing subcomponent | only a language with attributes declares that way    |
| the writer's underlying stream      | each language names its standard streams its own way |

Nothing else varies.
