# Support

The Support component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the template a port's own component `README.md` starts from: it carries these
headings in this order, so every port's document reads alike and anyone moving
between ports recognizes both.

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
Queue envelope uses ([`QUEUE.md`](../QUEUE/QUEUE.md)).

A format helper renders a machine-readable instant from milliseconds, so two
components do not each write their own formatting.

---

## Using it

### Never read the clock directly

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
