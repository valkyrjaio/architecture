# Container Bindings

The **cross-language** rule for the key a container binding is registered under,
and for the factory that builds the service.

This document holds no code example, for the reason in
[`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md). The container's own behavior
is in [`CONTAINER.md`](../component/CONTAINER.md).

---

## Two things, not one

A binding is a **key** and a **factory**. The key identifies the service. The
factory builds it.

PHP and Java originally conflated them, because a class reference can serve as
both an identifier and the instruction for how to construct the thing. Separating
them is what makes the container portable, because three of the five languages
cannot use a type as a key at all.

---

## The key

| Port       | Key is            | Verified by  |
| ---------- | ----------------- | ------------ |
| PHP        | a class reference | the compiler |
| Java       | a class token     | the compiler |
| TypeScript | a string constant | nothing      |
| Go         | a string constant | nothing      |
| Python     | a string constant | nothing      |

A language that erases its types, or that has no reference to a type at all,
cannot key by type. Python could, and deliberately does not: using a type object
as a key forces the module to import before anything resolves it, which defeats
the lazy import that is Python's answer to a slow cold start.

### How a string key is spelled

**The key is modeled on how the port imports the class.** It is the source file's
directory path, written the way that language writes a namespace, plus the class
name. The class name keeps its PascalCase spelling in every port.

The `Contract` directory segment is **removed**, because the class name already
ends in `Contract` and repeating it says nothing.

**The key is therefore per port**, because each port's directory layout already
differs. A port writes the casing its own layout has, and no port converts
another port's path. A port with a PascalCase layout writes PascalCase segments;
a port with a lowercase layout writes lowercase segments.

Warning: a string key is checked by nothing. A typo survives the whole gate and
fails when something resolves it. That is why a string key is never written
inline.

---

## Where a string key lives

A port whose keys are strings holds them in a **constants class per component**,
in that component's `Constant` segment, named for the component and suffixed
`ServiceId`. A caller references the constant and never the literal.

A port whose language names a class natively holds no such file, because the
class reference is already the key and the compiler already checks it.

### Why per component, and never one central file

A single file in the container component would list every key in the framework.
It would grow without bound, and it would make every component depend on the
container for its own identifiers. That is the coupling the component
architecture exists to prevent.

Per-component constants mean a component owns its own identifiers, the same way
it owns its own throwables and its own config.

---

## Never write a class name as a literal

A port whose language can reference a class **references it**. It never writes the
name of a class as a string.

A reference resolves against the file's own imports, so it is correct whenever the
file compiles. A string is text: no tool renames it, no static analysis checks it,
and no architecture linter sees the dependency it creates.

A configuration format is the one exception, because a config file has no class
reference. The rule there is to keep the authoritative list in code and assert
that the config matches it.

---

## The factory

**A factory is a closure, in every port.** Every one of the five languages has
first-class functions, so a closure is the one construction mechanism that is
available everywhere.

A closure is also better than the alternative in the two languages that have
one. It names its dependencies explicitly, it constructs without reflection, and
a reader can see exactly what it builds. Reflection-based construction is neither
checkable nor readable.

The container wraps a registration so that resolution is uniform: a resolution
always invokes a callable, and no path needs a check for which kind of
registration it found. The build tool writes the generated form in the same shape
the container holds at run time, so the cached path and the uncached path behave
identically.

A service class carries **no registration code**. A provider registers it. A
static factory on the service class is an optional convenience that a port may
offer and the framework does not use.

---

## Permitted variation

| Variation                        | Reason                                           |
| -------------------------------- | ------------------------------------------------ |
| the key's type                   | the language has a type reference or it does not |
| the key string's casing          | each port writes its own directory layout        |
| whether a constants class exists | only a port whose keys are strings needs one     |

The container's own operations do not vary. They are in
[`CONTAINER.md`](../component/CONTAINER.md).
