# Support

Implements
[`SUPPORT.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/support/SUPPORT.md).

## Introduction

The Support component holds the clock. Reading the time through it rather than
from the language is what makes a test over anything time-dependent possible: a
test freezes the clock, runs the code, and gets the same answer every run.

There are two clocks, one per resolution, and a renderer that turns an instant
into text.

Nothing resolves these from the container. A clock is a value a caller reads, so
there is no contract, no adapter and no config.

## Reading the clock

| Clock                 | Answers in    |
| --------------------- | ------------- |
| the second clock      | whole seconds |
| the millisecond clock | milliseconds  |

Each declares the same three operations.

| Operation  | Does                                                         | Returns     |
| ---------- | ------------------------------------------------------------ | ----------- |
| `get`      | returns the frozen instant where frozen, else reads the host | the instant |
| `freeze`   | holds the clock at the instant passed                        | nothing     |
| `unfreeze` | returns the clock to the host                                | nothing     |

```text
// What application code calls.
now = Time.get()

// The millisecond clock, for a value that needs the resolution.
startedAt = Microtime.get()
```

The millisecond clock also answers in whole milliseconds, for a caller that wants
an integer rather than the clock's own fractional value.

```text
elapsed = Microtime.getMilliseconds()
```

### Read the clock through the component

Code that reads the host's clock directly cannot be frozen, so a test over it
either sleeps or asserts a range. Reading through the component makes the same
test exact.

```text
// Wrong. Nothing can freeze this.
expiresAt = hostClock() + 3600

// Right.
expiresAt = Time.get() + 3600
```

## Freezing in a test

`freeze` takes the instant to hold. It is not optional, so a test states the
instant it means rather than depending on when it ran.

```text
Time.freeze(1760000000)

token = issueToken()

assert token.expiresAt == 1760003600

Time.unfreeze()
```

**Unfreeze after.** A frozen clock outlives the test that froze it, so the next
test reads an instant it did not choose. A test suite that freezes in a setup
step unfreezes in the matching teardown.

## Rendering an instant

The renderer turns a count of milliseconds into text, in **RFC 3339** — the one
rendering a person reads and a machine parses.

```text
text = Rfc3339.fromMilliseconds(Microtime.getMilliseconds())
```

It reads no clock. A caller passes the instant, so rendering a frozen instant in
a test gives the same text every run.
