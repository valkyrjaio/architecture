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

A language with a first-class reference to a type can conflate them, because that
reference serves as both an identifier and the instruction for how to construct
the thing. Separating them is what makes the container portable, because a
language without such a reference cannot key by type at all.

---

## The key

| Key is                  | Verified by  | Used where                                 |
| ----------------------- | ------------ | ------------------------------------------ |
| a reference to the type | the compiler | the language has one and importing is free |
| a string constant       | nothing      | everywhere else                            |

Three cases force a string. A language that erases its types at run time has no
reference to use. A language with no type reference at all has nothing to use.
And a language where naming a type forces its module to load **could** use one
and must not: the import would run before anything resolves the binding, which
defeats the lazy loading that answers a slow cold start.

So the string key is the portable case, not the exception.

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
in that component's `Constant` segment, named for the component and for what it
holds. A caller references the constant and never the literal.

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

A factory is anything the language can call. **Which form is permitted depends on
who reads the declaration**, and there are two readers.

**A provider's map is read statically by the build tool, so its values are
references to named methods.** An inline function body is not a value the tool
can carry into generated output, so a provider that holds one cannot be cached.
The same rule governs every list a provider declares
([`PROVIDERS.md`](PROVIDERS.md)).

**A binding registered directly is read by nothing, so any callable is fine
there**, an inline function included. Such a binding is already outside what the
build tool can see, so the restriction that exists for the tool's sake does not
apply.

A named method is still better for the reader wherever there is a choice. It has
a name that says what it builds, it resolves each dependency explicitly, it
constructs without reflection, and it is testable on its own. Reflection-based
construction is neither checkable nor readable.

**The container wraps every registration in a function of its own**, so that
resolution is uniform: a resolution always invokes something callable, and no
path needs a check for which kind of registration it found.

**A generated cache may add one more wrapper, where the language forces it.** A
language with a compile-time reference to a type names the method in the
generated file and loads nothing. A language without one evaluates the name as
the file loads, which would import every provider module at cache load and undo
the saving the cache exists for. Such a port wraps each generated value in a
thunk, so the name is looked up when the binding first resolves rather than when
the cache loads. The declared shape is unchanged; only the generated file differs. That wrapping is the
container's, not the author's — it is why a resolution path has no branch, and it
is not an invitation to write the function by hand. The build tool writes the generated form in the same shape
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
