# Type

The Type component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Type wraps a value so that its validity is the type's business rather than every
caller's. A type holds a value, proves it on construction, and hands it back.

It is a tier-one component. Routing casts a parameter to a type
([`HTTP.md`](HTTP.md), [`CLI.md`](CLI.md)), a model declares its properties as
types, and an identifier is a type. So this component depends on nothing and much
depends on it.

---

## One contract

Every type in the component implements the same contract, and it declares four
operations:

| Operation     | Does                                                       |
| ------------- | ---------------------------------------------------------- |
| `fromValue`   | builds the type from a raw value, proving it               |
| `asValue`     | returns the underlying value                               |
| `asFlatValue` | returns it reduced to a scalar                             |
| `modify`      | returns a copy whose value a caller's function transformed |

**`fromValue` is a named constructor, and it is the one static method a data
object keeps**, because it returns the type's own kind.

**`asFlatValue` exists because a type may wrap a structure.** A collection or a
model returns its structure from `asValue`, and something that must store or
transmit one value asks for the flat form. For a type already holding a scalar,
the two agree.

**`modify` keeps a type immutable while staying useful.** A caller that wants a
changed value passes a function and receives a new type, so the validity is
re-proved on the way out rather than bypassed by a setter.

That one contract is the component's whole reason for existing: a caller that
works with any type knows these four operations, whatever the type is.

---

## The families

| Family     | Wraps                                                               |
| ---------- | ------------------------------------------------------------------- |
| scalar     | a string, an integer, a float, a boolean, a null                    |
| structured | an array, an object, a collection, a JSON document                  |
| model      | a declared set of named, typed properties                           |
| identifier | an identifier, in several schemes                                   |
| support    | the abstract bases, the enums and the throwables the families share |

**A scalar type is not pointless.** Wrapping a string makes a parameter's
declared type checkable, which is what lets routing cast a matched value before a
handler runs rather than having the handler validate.

---

## Identifiers

The identifier family is the part most worth knowing, because an application
chooses among its schemes:

| Scheme | Is                                                                  |
| ------ | ------------------------------------------------------------------- |
| random | an identifier with no order, unique by size alone                   |
| sorted | an identifier whose text sorts in the order the values were created |
| local  | a sorted identifier this project defines, in several versions       |

**A sorted identifier is the default worth reaching for.** A random one is
unique but arrives in no order, so a store that indexes it writes into the middle
of an index on every insert. A sorted one appends, which is why a database-backed
queue or table wants it.

Each scheme has a **factory**, because an identifier is created rather than
constructed from a value a caller already has. The factory is the one place the
generation rules live, so two call sites cannot generate differently.

A scheme that has versions declares them as an enum, and a caller names the
version it wants rather than taking whatever the default is this release.

---

## Using it

### Declare a type where a value has rules

A value with a rule — a format, a range, a set of permitted values — becomes a
type, and the rule is proved once, on construction. Every later caller then holds
something already valid.

### Construct through the factory where one exists

A type whose value is generated has a factory, and a type parsed from an untrusted
value has one too. The factory is where the failure is raised, so a constructor is
not the place to catch a parse error.

### Keep a type immutable, and use `modify`

A type does not change its value. A caller wanting a changed one calls `modify`
and takes the copy, which re-proves the value.

### Resolve a type through the container when the type is named by data

Where the type to build comes from data rather than from the code — a cast a route
parameter declares, for instance — the type is resolved through the container with
the value as an argument. A port resolves it with the operation that builds each
time and passes the value, never the one that returns a cached instance, and the
registered factory returns the type contract.

### Pick an identifier scheme by how it is stored

Use a sorted identifier for anything a store indexes, and a random one only where
order would leak something. Decide once per application rather than per table.

---

## Permitted variation

| Variation                                                  | Reason                                           |
| ---------------------------------------------------------- | ------------------------------------------------ |
| which scalar types exist                                   | a language's own set of primitives differs       |
| how the flat form represents a structure                   | each language names its own scalar conversions   |
| whether a type is a class or a wrapper over a native value | a language with value types may not need a class |

The contract, its four operations, the families and the identifier schemes do
not vary.
