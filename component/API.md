# Api

The Api component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Api gives every machine-readable response one envelope, so a client parses one
shape whatever the endpoint and whatever happened.

The component owns the envelope and the mapping into it. It owns no routing and
no response: it produces the envelope, or a response carrying it, through
[`HTTP.md`](HTTP.md).

---

## Hierarchy

| Subcomponent | Holds                                                  |
| ------------ | ------------------------------------------------------ |
| `Manager`    | the component contract: the mappings into the envelope |
| `Model`      | the envelope, and the data it wraps                    |
| `Middleware` | the stage that applies the envelope to a response      |
| `Constant`   | the shared vocabulary                                  |
| `Provider`   | the service provider                                   |
| `Throwable`  | the component's throwable contract and its exceptions  |

---

## The envelope

Every response carries the same top-level parts, and a part that has nothing to
say is present and empty rather than absent:

| Part            | Holds                                             |
| --------------- | ------------------------------------------------- |
| the data        | what the endpoint returned                        |
| the message     | one human-readable sentence about the outcome     |
| the errors      | what went wrong, where something did              |
| the warnings    | what the client should notice but need not act on |
| the status      | the outcome, as a value a client branches on      |
| the status code | the protocol's own status                         |

**Empty is not absent**, the same rule the Queue envelope states
([`QUEUE.md`](QUEUE.md)). A client that can rely on every part being present
never checks for a missing one, and a part that appears only sometimes is the
part a client forgets to handle.

**The status and the status code are separate.** The protocol's status says what
happened to the request; the envelope's status says what happened to the
operation. They differ often enough that collapsing them loses information — a
request that succeeded carrying an operation that did not.

The data part has its own shape for a collection: the items, the key they sit
under, and the total. **The total is separate from the item count**, because a
page of ten out of two hundred has to report both.

---

## Canonical operations

The contract maps each thing an endpoint might return into the envelope, and each
mapping comes in two forms — the envelope, or a response carrying it:

| Maps from     | Into                                     |
| ------------- | ---------------------------------------- |
| an object     | the envelope's data                      |
| many objects  | the envelope's collection data           |
| an entity     | the storable form of a record            |
| many entities | the storable form of many                |
| an array      | the data as given                        |
| a throwable   | the errors, and a status that matches it |

**The two forms exist so a caller composes or returns.** Code that returns
immediately asks for the response; code that adds to the envelope first asks for
the envelope and builds the response itself.

**Mapping a throwable is the operation that matters most.** It is what makes a
failure arrive in the same shape as a success, so a client has one parser. A
failure rendered by whatever caught it is how an API ends up with several error
shapes.

---

## Using it

### Return the envelope from every machine-readable endpoint

One shape, every endpoint, success and failure alike. The value of an envelope is
entirely in its being universal — a client that has to handle a second shape gets
no benefit from the first.

### Map entities rather than serializing them

Pass the entity to the entity mapping, so the storable form is produced once by
the component rather than by each endpoint
([`ORM.md`](ORM.md)).

### Map the throwable at the throwable stage

Catch at the pipeline's throwable stage and map there
([`LIFECYCLE.md`](../convention/LIFECYCLE.md)), so every failure in the request
takes the same path, including one raised before the handler ran.

### Report the total for a collection

Set the total whenever the data is a page of something larger. A client cannot
page without it, and it cannot tell a short page from the last one.

### Keep the message for a person and the errors for the client

The message is one sentence someone reads. The errors are what code branches on.
Putting the reason only in the message means the client parses prose.

---

## Permitted variation

| Variation                 | Reason                                                            |
| ------------------------- | ----------------------------------------------------------------- |
| the serialization format  | a port may emit a format other than JSON behind the same envelope |
| which status values exist | an application extends the set for its own outcomes               |

The envelope's parts, the empty-not-absent rule, the status split and the
mappings do not vary.
