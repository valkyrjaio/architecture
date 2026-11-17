# Type

Implements
[`TYPE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/type/TYPE.md).

## Introduction

The Type component wraps a value in something that proves the value is what it
claims to be. A string that must be a slug becomes a slug type, and nothing
downstream re-checks it.

Use it where a value has rules. A value with no rules stays what the language
calls it.

Every type implements one contract, so code that takes any type knows four
operations whatever the type is.

## The one contract

| Operation     | Does                                              | Returns   |
| ------------- | ------------------------------------------------- | --------- |
| `fromValue`   | builds the type from a raw value, proving it      | the type  |
| `asValue`     | returns the underlying value                      | the value |
| `asFlatValue` | returns it reduced to a scalar                    | a scalar  |
| `modify`      | returns a copy whose value a function transformed | the type  |

```text
slug = SlugT.fromValue("my-first-post")

text = slug.asValue()
```

**`fromValue` is a named constructor**, so the caller names the type it wants and
gets that type or a failure. A value that does not satisfy the type raises rather
than producing a type that lies.

```text
try {
    slug = SlugT.fromValue(fromRequest)
} catch (throwable) {
    return response.badRequest()
}
```

### `asFlatValue` exists because a type may wrap a structure

A collection or a model returns its structure from `asValue`. Something that must
store or transmit one value asks for the flat form. For a type already holding a
scalar the two agree.

```text
collection.asValue()      // the structure
collection.asFlatValue()  // one scalar
```

### `modify` keeps a type immutable

A caller that wants a changed value passes a function and receives a new type, so
the validity is proved again on the way out.

```text
upper = slug.modify((value) => value.toUpperCase())
```

The original is unchanged, and `upper` is a slug or the modification raised.

## The kinds

| Group      | Kinds                                 |
| ---------- | ------------------------------------- |
| scalar     | integer, float, string, boolean, null |
| structured | array, object, collection, json       |
| modelled   | model, enumeration                    |
| identifier | id, uid, uuid, ulid, vlid             |

A kind's class carries a suffix in a language whose own type names collide with
the kind, which is most of them.

### Pick an identifier scheme by how it is stored

| Scheme | Suits                                             |
| ------ | ------------------------------------------------- |
| uuid   | an identifier generated anywhere, ordered nowhere |
| ulid   | an identifier that sorts by when it was made      |
| vlid   | an identifier with the project's own layout       |

An identifier that sorts by creation keeps an index from fragmenting, which an
identifier generated at random does not.

## Casts

A **cast** names a type to build a value into, and how. It is data, so a route
parameter or a command argument declares one and something else resolves it.

| Part     | Is                                                              |
| -------- | --------------------------------------------------------------- |
| the type | which type to build                                             |
| convert  | whether to convert the value, or require it already be the type |
| is array | whether the value is a list of that type                        |

```text
cast = new Cast(SlugT, convert: true, isArray: false)
```

### Resolving a type a cast names

Where the type comes from data rather than from the code, it is resolved through
the container with the value as an argument.

```text
type = container.getService(cast.type, { value: raw })
```

Resolve with the operation that builds for this call, never the any-kind one: the
any-kind operation tries the singleton path first and passes it no argument, so
the value is discarded. [Container](../container/README.md) covers the resolve
operations.

**A data object never holds the container.** A route parameter declares the cast;
the thing that resolves it holds the container.

## Service registration

| Binding                | Is                             |
| ---------------------- | ------------------------------ |
| each type a cast names | bound so a cast can resolve it |

A type a caller constructs by name needs no binding. A type a cast names does,
because the cast holds the type as data and the container is what turns data into
an object.
