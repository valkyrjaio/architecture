# Attribute

The Attribute component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Attribute reads the markers a language lets an application attach to its own
code — an attribute, an annotation, a decorator — and returns them as objects.

It owns the reading. It owns no meaning: what a marker causes is the business of
the component that defines it, which is why routing defines its own markers and
this component only finds them.

**This component exists only in a port whose language has such markers.** A port
without them has no equivalent and loses nothing, because every declaration a
marker can express is also expressible as a literal list
([`PROVIDERS.md`](../convention/PROVIDERS.md)).

---

## Hierarchy

| Subcomponent | Holds                                                 |
| ------------ | ----------------------------------------------------- |
| `Collector`  | the component contract                                |
| `Contract`   | the contract a marker implements to be given context  |
| `Trait`      | the shared behavior a marker reuses                   |
| `Provider`   | the service provider                                  |
| `Throwable`  | the component's throwable contract and its exceptions |

---

## Canonical operations

One operation per place a marker can be attached, each named for the place:

| Operation               | Reads markers on            |
| ----------------------- | --------------------------- |
| `forClass`              | the class itself            |
| `forClassMembers`       | every member of a class     |
| `forClassAndMembers`    | both, in one call           |
| `forConstant`           | one constant                |
| `forConstants`          | every constant              |
| `forProperty`           | one property                |
| `forProperties`         | every property              |
| `forMethod`             | one method                  |
| `forMethods`            | every method                |
| `forMethodParameter`    | one parameter of a method   |
| `forMethodParameters`   | every parameter of a method |
| `forFunction`           | a free function             |
| `forFunctionParameters` | its parameters              |
| `forClosure`            | an inline function          |
| `forClosureParameters`  | its parameters              |

The `for*` naming matches [`REFLECTION.md`](REFLECTION.md), because the two
answer the same kind of question about the same kinds of declaration.

**The combined operations are not conveniences.** Reading a class and its members
separately reflects the class twice, and reflection is the expensive part
([`REFLECTION.md`](REFLECTION.md)), so the combined form is what a collector
scanning many classes uses.

---

## A marker is metadata, never a registrar

**A marker records what to register. It never registers anything itself.** This is
the component's one hard rule, and it is stated in
[`HANDLERS.md`](../convention/HANDLERS.md) as well because that is where a reader
meets it first.

The reason is that a marker in some languages **executes** when the code
containing it loads. A marker that registered on execution would register every
time a class loads, would register things the application never uses, and would
make the cached path and the uncached path disagree — the cached path never
executes the marker at all.

So a marker is read, in one of two ways:

- the **build tool** reads it statically, and writes the result into generated data
- the **runtime collector** reads it at boot, where there is no generated data

Both read the same marker, which is what keeps the two paths identical
([`DATA_CACHE.md`](../convention/DATA_CACHE.md)).

---

## A marker may be given its context

A marker often needs to know what it is attached to — a route marker on a method
needs that method in order to name a handler. A marker that needs it implements a
contract through which the collector supplies the declaration it was found on.

A marker that does not need it implements nothing, and stays a plain value.

---

## Using it

### Define markers in the component they belong to

Routing's markers live in routing. This component is how they are found, not
where they live.

### Keep a marker's arguments literal

A marker's arguments are read statically by the build tool, so they are literals.
An argument computed at load time is invisible to the tool, which silently drops
it ([`SINDRI.md`](../convention/SINDRI.md)).

### Never name a class in a marker argument where the language would load it

A marker argument that names a class may force that class to load as the marker's
own file loads. Where a language does that, the argument is deferred instead — a
port that has this problem records how in its own guide.

### Scan classes an application lists, never the whole tree

A provider lists the classes to scan ([`PROVIDERS.md`](../convention/PROVIDERS.md)).
Scanning everything is slow, and it finds markers in code the application did not
mean to register.

---

## Permitted variation

| Variation                                | Reason                                           |
| ---------------------------------------- | ------------------------------------------------ |
| whether the component exists             | only a language with markers has one             |
| which places a marker may attach to      | each language permits its own set                |
| whether a marker is readable at run time | a language may expose markers only at build time |

The operation names, the metadata-only rule and the two reading paths do not vary
in a port that has the component.
