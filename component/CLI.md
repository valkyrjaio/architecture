# Cli

The **cross-language** definition of the Cli component. It states the hierarchy,
the names and the behavior that every port implements.

A port implements this document. A port does not redefine it. The reference
implementation is PHP ([`AGENTS.md`](../AGENTS.md) §1).

This document holds no code example. The component's `README.md` in each port
holds the examples and the per-language spelling.

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

## The input

An input holds the caller, the command name, the arguments and the options. A
port builds it from the runtime's own argument vector, and the parsed result is
identical in every port.

**How the vector is shaped is the one legitimate per-port difference**, because
each runtime hands over a differently shaped vector. The rules, and the index each
port reads the command name from, are in
[`PORTS.md`](../convention/PORTS.md).

An **argument** is positional. An **option** is named, and carries a long name
and an optional short name. Each declares whether it is required, whether it
takes a value, and the type its value casts to. A port casts before the command
runs, so a command receives a typed value and never parses a string.

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

**A warning exits zero.** A warning is informational, so it does not fail a build
or a pipeline. An application that wants a warning to fail opts into that; the
component never decides it.

---

## Interactivity

A run has a level, and the level decides what reaches the person:

| Level           | Means                                                 |
| --------------- | ----------------------------------------------------- |
| interactive     | questions are asked, and answers are read             |
| non-interactive | questions are not asked, and a default answer is used |
| quiet           | only errors are written                               |
| silent          | nothing is written                                    |

Each level is a separate config contract, so an application configures one
without constructing the others
([`COMPONENT_CONFIG.md`](../convention/COMPONENT_CONFIG.md)).

---

## Built-in commands

Every port ships the same four: list the commands, show help for one command,
report the version, and emit a machine-readable command list for shell
completion.

**A command whose whole purpose is machine-readable output prints no header.**
Every other command prints one, and an application suppresses it globally if it
wants to.

The name of each built-in command and of each built-in option is configurable, so
an application can rename one it conflicts with.

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
