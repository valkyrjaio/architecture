# Type

The Type component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the master copy of this component's README, which every port carries, so that
every port's document for this component reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Type wraps a value so that its validity is the type's business rather than every
caller's. A type holds a value, proves it on construction, and hands it back.

It is a tier-one component. Routing casts a parameter to a type
([`HTTP.md`](../http/HTTP.md), [`CLI.md`](../cli/CLI.md)), a model declares its properties as
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

## Structure

The component is a root, and then one directory per kind of value on one
repeating shape.

### The root

| Class                   | Is                                                |
| ----------------------- | ------------------------------------------------- |
| `Contract/TypeContract` | the contract every type implements                |
| `Abstract/Type`         | the base every type extends                       |
| `Data/Cast`             | names a type to build a value into, and how       |
| `Constant/CastArgument` | the names a cast's own arguments are passed under |
| `Provider/TypeProvider` | the service provider                              |

### One directory per kind

Each kind holds the same three things, named for the kind.

| Class                                                        | Is                                  |
| ------------------------------------------------------------ | ----------------------------------- |
| `<Kind>/Contract/<Kind>Contract`                             | the kind's own contract             |
| `<Kind>/<Kind>T`                                             | the type class                      |
| `<Kind>/Throwable/Exception/<Kind>InvalidFromValueException` | raised for a value the kind rejects |

A kind adds a class of its own where the kind has more than one shape, such as a
boolean's two constants or an array that cannot be empty. It adds a `Factory/`
where construction takes work beyond proving the value, and a `Support/` for a
calculation the type delegates.

### The type class carries a `T`

**A type class ends in `T`**, so `IntT`, `StringT`, `BoolT`. The kinds are named
for the language's own types, and the suffix is what keeps a type class from
colliding with the type it wraps. A port that dropped the suffix would collide in
every language that reserves the name.

The kind's contract and its throwable take no suffix. Only the class a caller
constructs does.

### The kinds

| Group      | Kinds                                    |
| ---------- | ---------------------------------------- |
| scalar     | `Int`, `Float`, `String`, `Bool`, `Null` |
| structured | `Array`, `Object`, `Collection`, `Json`  |
| modelled   | `Model`, `Enum`                          |
| identifier | `Id`, `Uid`, `Uuid`, `Ulid`, `Vlid`      |

### Order

| In                  | The order is                                    |
| ------------------- | ----------------------------------------------- |
| a kind's operations | `fromValue`, `asValue`, `asFlatValue`, `modify` |
| the kinds           | the groups above, in that order                 |

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
the value as an argument. A port resolves it with `getService`, never `get`.
`get` tries the singleton first and passes it no argument. The container returns
a singleton registered under that id as it was built, and it discards the cast
value. The registered
factory returns the type contract. A data object never holds the container, and
[`STATIC_METHODS.md`](../../convention/STATIC_METHODS.md) names the actor that
resolves the cast in each protocol. [`CONTAINER.md`](../container/CONTAINER.md) names the
three service types and the operation for each one.

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
