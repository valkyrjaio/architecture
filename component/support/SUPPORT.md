# Support

The Support component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Support holds what the framework needs from outside itself, behind something a
test can control. Today that is **time**.

It is deliberately small, and it is **not a place to put things that do not fit
elsewhere.** A support class earns its place by making something untestable
testable, or by holding a calculation that has no owner. Anything with a domain
belongs in the component that owns that domain.

---

## Hierarchy

| Subcomponent | Holds                      |
| ------------ | -------------------------- |
| `Time`       | the clock, and its formats |

A port may hold one more here where it has something that fits the test above and
belongs to no component.

There is no config, no provider and no adapter, because nothing here is chosen
by an application.

---

## Structure

Every class a port builds, where it sits, and what it does.

| Class            | Is                                                  |
| ---------------- | --------------------------------------------------- |
| `Time/Time`      | the clock, in whole seconds                         |
| `Time/Microtime` | the clock, in milliseconds                          |
| `Time/Rfc3339`   | renders an instant as text a person and a wire read |

There is no contract, no adapter, no config and no provider. A clock is a value
a caller reads, so nothing resolves it and nothing configures it.

### Each clock declares the same three operations

| Operation  | Does                                                         |
| ---------- | ------------------------------------------------------------ |
| `get`      | returns the frozen instant where frozen, else reads the host |
| `freeze`   | holds the clock at an instant the caller passes, or at now   |
| `unfreeze` | returns the clock to the host                                |

`get` is the only operation a caller outside a test uses. The pair around it
exists for a test, and a port declares the pair on each clock rather than on one
of them.

The millisecond clock also answers in milliseconds, for a caller that needs the
resolution rather than the unit the clock is named for.

### `Rfc3339` renders, and reads nothing

It takes an instant and returns text. It does not read a clock, so a caller
renders a frozen instant in a test and gets the same text every run.

The format is **RFC 3339**, which is the one rendering a person reads and a
machine parses, and the one every port writes alike.

### Order

| In                   | The order is                     |
| -------------------- | -------------------------------- |
| a clock's operations | `get`, then `freeze`, `unfreeze` |
| the clocks           | seconds, then milliseconds       |

---

## Why time is a component at all

**A reading of the clock is the one input a test cannot control from outside.**
Code that asks the language for the current time cannot be tested for what it
does at a boundary — an expiry, a schedule, a window — without waiting for that
moment to arrive.

So the framework never reads the clock directly. It asks this component, and the
component can be frozen.

| Operation  | Does                                                          |
| ---------- | ------------------------------------------------------------- |
| `get`      | returns the current time                                      |
| `freeze`   | fixes the time that `get` returns, at a value a caller passes |
| `unfreeze` | releases it, so `get` reads the clock again                   |

**The clock reads wall time, not elapsed time.** A monotonic reading cannot be
compared with a stored instant, so a component that reads one cannot tell whether
a stored expiry has passed. `freeze` takes the value to freeze at, so a test can
place the clock at a boundary rather than only stopping it where it is.

Two resolutions are available, because the choice has consequences: whole seconds
for a human-facing value, and sub-second precision for anything that measures or
orders. A sub-second reading also reports its milliseconds, which is the unit the
Queue envelope uses ([`QUEUE.md`](../queue/QUEUE.md)).

A format helper renders a machine-readable instant from milliseconds, so two
components do not each write their own formatting.

---

## Using it

### Read the clock through the component

Ask this component. A direct read is untestable, and it is the reason a test ends
up sleeping or asserting a range instead of a value.

### Freeze in a test, and unfreeze after

Freeze, assert the behavior at the boundary, then unfreeze. A frozen clock left
frozen leaks into the next test, which is why the release operation exists rather
than the freeze simply expiring.

### Pick the resolution by what the value is for

Use whole seconds for something a person reads, and sub-second precision for
anything compared, ordered or measured. Two events in the same second are
indistinguishable at the coarser resolution.

### Store an instant as a number, render it for a person

Keep the millisecond value as the authoritative one and render the readable form
from it. The Queue envelope states this as a cross-port rule, and it holds
wherever an instant is stored.

---

## Permitted variation

| Variation                          | Reason                                              |
| ---------------------------------- | --------------------------------------------------- |
| how freezing is implemented        | each language has its own way to hold process state |
| which resolutions the clock offers | a language's clock has its own precision            |

What the component is **for** does not vary, and neither does the rule that the
framework never reads the clock directly.
