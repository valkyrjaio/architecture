# Throwable

The Throwable component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Throwable owns the **root of every failure the framework raises**, and the handler
that catches what nothing else did.

It is a tier-one component: every other component declares its own throwables
under this root, so this one depends on nothing and everything depends on it.

The naming rules for a component's own throwables are in
[`THROWABLES.md`](../convention/THROWABLES.md). This document is the component;
that one is the convention every component follows.

---

## Hierarchy

| Subcomponent | Holds                                                    |
| ------------ | -------------------------------------------------------- |
| `Contract`   | the root throwable contract                              |
| `Exception`  | the abstract bases every concrete throwable extends      |
| `Handler`    | the handler contract, and one class per adapter          |
| `Factory`    | the support that derives the trace code from a throwable |

There is no `Data` and no `Provider` subcomponent. Nothing configures the
component. The entry point installs the handler. See
[Install the handler once, at the entry point](#install-the-handler-once-at-the-entry-point).

---

## The root is a contract, not a class

**Every throwable the framework raises implements one root contract.** A caller
catches that contract to mean "a failure from this framework", and catches the
language's own type to mean "any failure at all".

A root class instead of a contract would force every throwable to inherit from
it, which a language with single inheritance cannot do while also extending the
language's own base.

**Whether the root contract itself extends the language's throwable type depends
on the language.** Where that type is an interface, the contract extends it and a
catch on the contract is checked by the compiler. Where it is a class, a contract
cannot extend it, so the contract stands alone and the abstract bases carry the
inheritance instead.

Beneath the root sit **two abstract bases**, and the split is the one every
language already makes:

| Base                     | Means                                                |
| ------------------------ | ---------------------------------------------------- |
| runtime failure          | something went wrong while running                   |
| invalid argument failure | the caller passed something the method cannot accept |

Each extends the language's own equivalent and implements the root contract, so a
caller catching the language's invalid-argument type still catches the
framework's.

**Both are abstract, and only concrete throwables are raised.** An abstract base
exists to be caught, never to be thrown — throwing one says nothing about what
failed.

---

## The handler

The handler is what catches a throwable nothing else did. It declares one
operation, **`enable`**, which installs it as the handler for the process.

One operation is all it has, because a handler is not called — it is installed,
and the runtime calls it. Anything a caller invokes deliberately belongs in the
`ThrowableCaught` stage instead
([`LIFECYCLE.md`](../convention/LIFECYCLE.md)).

**A handler is the last resort, not the error path.** A failure a protocol can
answer is answered at its `ThrowableCaught` stage, which can turn it into a response. The
handler catches what escaped that, which means a failure before the pipeline
started or inside the machinery that runs it.

---

## The trace code

The component derives a **trace code** from a throwable: a short, stable value
computed from the throwable's type and its trace.

It exists so the same failure gets the same code every time it happens, and two
different failures get different codes. That makes a code safe to show a person
and safe to search a log for — which is the pair of properties a message does not
have, because a message carries values that differ per occurrence.

**A trace code is not a secret and not an identifier.** It does not identify an
occurrence, only a shape, so two reports with one code are the same bug.

---

## Using it

### Declare a throwable per failure, under your component's root

A component declares its own root contract and its own abstract bases beneath
these, then one concrete throwable per distinct failure. The recursive naming
rule that keeps those names unique across the framework is in
[`THROWABLES.md`](../convention/THROWABLES.md).

### Throw a concrete throwable, catch an abstract one

The thrower names the exact failure. The catcher names the category it can handle.
That is the whole reason the hierarchy has levels.

### Catch the narrowest contract that covers what you can handle

Catching the root contract catches every framework failure, including ones this
caller cannot do anything about. Catch a component's own contract to handle that
component's failures.

Warning: a component contract does **not** bound a catch to the framework when a
marker contract extends it. A component whose throwables carry a marker an
application may also apply catches the application's too, which
[`QUEUE.md`](QUEUE.md) describes for its non-retryable marker.

### Install the handler once, at the entry point

The handler is installed by the entry point before the pipeline starts, so a
failure in the bootstrap is caught. Installing it later leaves exactly the window
that most needs it uncovered.

---

## Permitted variation

| Variation                               | Reason                                                    |
| --------------------------------------- | --------------------------------------------------------- |
| the language base each abstract extends | each language names its own runtime and argument types    |
| whether the contract is an interface    | a language without interfaces uses its nearest equivalent |
| which handler integrations ship         | each ecosystem has its own error reporters                |

The root, the two abstract bases, the handler's one operation and the trace code
do not vary.
