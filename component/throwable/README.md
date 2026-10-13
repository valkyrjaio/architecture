# Throwable

Implements
[`THROWABLE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/throwable/THROWABLE.md).

## Introduction

The Throwable component is the root every failure the framework raises comes
from, and the handler that reports one a unit of work did not catch.

Catching the root contract means "a failure from this framework". Catching the
language's own type means "any failure at all". The two are different questions,
and the root exists so a caller can ask the first.

Every other component declares its own failures under this root, so this
component depends on nothing and everything depends on it.

## Catching

```text
try {
    orders.place(basket)
} catch (ThrowableContract throwable) {
    // A failure from the framework.
}
```

### Catch the narrowest contract that covers what you can handle

Each component declares a throwable contract of its own, and each concrete
failure extends one of two bases: one for a failure the caller caused, one for a
failure at run time.

```text
// Everything this component can raise.
catch (OrmThrowable throwable)

// Only what the caller got wrong.
catch (OrmInvalidArgumentException throwable)
```

Catching the root where you can only handle one component's failure means
handling failures you cannot handle.

## Declaring a throwable

A component declares a contract, the bases it needs, and one concrete failure per
thing that can go wrong.

```text
// The component's own contract, extending the root.
interface CacheThrowable extends ThrowableContract { }

// Its bases.
abstract class CacheInvalidArgumentException
    extends InvalidArgumentException implements CacheThrowable { }

// One concrete failure.
class CacheUnreachableException extends CacheRuntimeException { }
```

**One concrete throwable per failure**, named for what failed rather than for
where it was raised. A caller catches a name that says what went wrong.

## The trace code

The component derives a **trace code** from a throwable: a hash of its type and
its trace, rendered as hex.

```text
code = ThrowableFactory.getTraceCode(throwable)
```

The same failure gets the same code every time it happens, and two different
failures get different codes. That makes a code safe to show a person and safe to
search a log for, which a message is not, because a message carries values that
differ per occurrence.

**A code is comparable within a port, not across ports.** A trace renders as the
language prints it, so the same logical failure hashes differently in two ports.

## The handler

The handler reports a throwable nothing caught. It has one operation.

| Operation | Does                                                  |
| --------- | ----------------------------------------------------- |
| `enable`  | installs the handler, at a reporting level, and says whether to display |

```text
handler = container.getSingleton(ThrowableHandlerContract)

handler.enable(displayErrors: true)
```

### Install it once, at the entry point

The entry point installs the handler before the pipeline starts, so a failure
during boot is reported rather than lost. Installing it later leaves exactly the
window that most needs it uncovered.

An adapter that renders a failure for a person is for development. An adapter
that records one is for production, and what it records goes through
[Log](../log/README.md).

## Service registration

| Binding                      | Is                                   |
| ---------------------------- | ------------------------------------ |
| `ThrowableHandlerContract`   | the handler the entry point installed |

There is no component config and no component provider binding a default. The
entry point registers the handler it installed, because a failure raised while
the container is still being assembled needs a handler that already exists.
