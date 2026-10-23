# Api

Implements
[`API.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/api/API.md).

## Introduction

The Api component gives every machine-readable response one envelope. A client
parses one shape, whatever the endpoint is and whatever happened.

The **envelope** is `JsonContract`: the data, a message, the errors, the
warnings, the envelope's own status, and the protocol's status code. The
**data** is `JsonDataContract`, which holds one item, or a list of items with a
total, or an array you pass as it is.

`ApiContract` maps a value into the envelope. Each mapping comes in two forms:
one returns the envelope, and one returns a response carrying it
([Http](../http/README.md)).

The component owns the envelope and the mapping into it. It owns no routing. A
throwable reaches the envelope in middleware rather than at a call site, so
every failure takes one path.

There is no adapter and no config. One envelope shape serves every response, so
there is nothing to swap and nothing to configure.

This document assumes the framework, where the container resolves the component
and a provider builds it. Used standalone, the component needs none of that:
construct the manager with a response factory and call it.

## The envelope

| Part         | Holds                                        | Default          |
| ------------ | -------------------------------------------- | ---------------- |
| `data`       | what the endpoint returned                   | nothing          |
| `message`    | one sentence a person reads                  | nothing          |
| `errors`     | what went wrong, where something did         | an empty set     |
| `warnings`   | what the client notices but need not act on  | an empty set     |
| `status`     | the outcome, as a value a client branches on | `Status.SUCCESS` |
| `statusCode` | the protocol's own status                    | `StatusCode.OK`  |

| Method                            | Does                               |
| --------------------------------- | ---------------------------------- |
| `getData` / `setData`             | reads or replaces the data         |
| `getMessage` / `setMessage`       | reads or replaces the message      |
| `getErrors` / `setErrors`         | reads or replaces the error set    |
| `setError`                        | adds one error                     |
| `getWarnings` / `setWarnings`     | reads or replaces the warnings     |
| `setWarning`                      | adds one warning                   |
| `getStatus` / `setStatus`         | reads or replaces the status       |
| `getStatusCode` / `setStatusCode` | reads or replaces the status code  |
| `asArray`                         | returns every part, for a response |

Every setter returns the envelope, so a caller chains them.

```text
json = api.jsonFromObject(order)

json.setMessage("Order placed.")
    .setWarning("The delivery estimate is provisional.")
    .setStatusCode(StatusCode.CREATED)
```

A success carrying one object reads like this:

```text
{
    "data": { "order": { "id": 4821, "total": 1999 } },
    "message": null,
    "errors": [],
    "warnings": [],
    "status": "success",
    "statusCode": 200
}
```

### The data part

| Part       | Holds                                       | Default |
| ---------- | ------------------------------------------- | ------- |
| `item`     | one object                                  | nothing |
| `itemKey`  | the name of the key the item sits under     | `item`  |
| `items`    | a list of objects                           | nothing |
| `itemsKey` | the name of the key the list sits under     | `items` |
| `total`    | how many items there are in all             | nothing |
| `messages` | the sentences a person reads about the data | nothing |
| `data`     | an array that the mapping passes as it is   | nothing |

`itemKey` and `itemsKey` name a key rather than becoming one, so neither
reaches the client under its own name. The envelope's own parts are always
present. The data's parts are not: the data holds the keys the mapping set, and
nothing else.

### Empty is present, not absent

Warning: the envelope guarantees to a client that every part is present on
every response. A serializer that omits a part when it is empty breaks that
guarantee, and so does one that omits a part altogether. Serialize every part,
including an empty message, an empty error set and an empty warning set.

A client then never checks for a missing part. A part that appears only
sometimes is the part a client forgets to handle.

```text
// Wrong. The client cannot tell "no errors" from "this response omits errors".
{ "data": { "order": { "id": 4821 } }, "status": "success", "statusCode": 200 }

// Right. The empty parts are present and empty.
{
    "data": { "order": { "id": 4821 } },
    "message": null,
    "errors": [],
    "warnings": [],
    "status": "success",
    "statusCode": 200
}
```

### The status and the status code are separate

`statusCode` says what happened to the request. `status` says what happened to
the operation. Collapsing the two into one value loses information. A request
that succeeded can carry an operation that did not.

`Status` holds the values every port ships: `Status.SUCCESS`, `Status.ERROR` and
`Status.FAIL`. An application extends the set for an outcome of its own.

```text
// The request was answered. The operation did not finish.
json.setStatus(Status.FAIL)
    .setStatusCode(StatusCode.OK)
    .setError("The card issuer declined the charge.")
```

Keep the reason in `status` and the sentence in `message`. A client that has to
read the message is a client that parses prose.

## Mapping a value into the envelope

| Method              | Maps         | Into                             |
| ------------------- | ------------ | -------------------------------- |
| `jsonFromObject`    | one object   | the data, under the object's key |
| `jsonFromObjects`   | many objects | the data, as a list              |
| `jsonFromEntity`    | one record   | the data, under the record's key |
| `jsonFromEntities`  | many records | the data, as a list              |
| `jsonFromArray`     | an array     | the data, as given               |
| `jsonFromException` | a throwable  | a failed outcome                 |

Each one has a `jsonResponseFrom...` form that returns the response instead of
the envelope. Ask for the response where your code returns straight away. Ask
for the envelope where your code adds to it first, and builds the response
itself.

```text
// Return straight away.
return api.jsonResponseFromObject(order)

// Add to the envelope, then build the response yourself.
json = api.jsonFromObject(order)

json.setMessage("Order placed.")

return responseFactory.createJsonResponse(json.asArray())
```

### One object

The mapping names the data after the object's own type. It lowercases the short
type name, so an `Order` sits under `order`. A client then reads a key rather
than a position.

```text
json = api.jsonFromObject(order)

// { "data": { "order": { ... } }, ... }
```

### A collection, with its total

`jsonFromObjects` appends `s` to that key, so a list of `Order` objects sits
under `orders`. A port appends `s` rather than pluralizing properly, so that
every port emits the same key.

```text
return api.jsonResponseFromObjects(order, otherOrder)
```

Set the total whenever the data is a page of something larger. Build the data
yourself for that, and hand it to the array mapping.

```text
data = new JsonData()

data.setItems(orders)
    .setItemsKey("orders")
    .setTotal(200)

return api.jsonResponseFromArray(data.asArray())
```

A page reports both numbers. The item count says how many items arrived. The
total says how many items there are in all, and a client cannot page without the
total.

```text
{
    "data": { "orders": [ { "id": 4821 }, { "id": 4822 } ], "total": 200 },
    "message": null,
    "errors": [],
    "warnings": [],
    "status": "success",
    "statusCode": 200
}
```

### A record

Pass the entity to the entity mapping rather than serializing it at the
endpoint. The component then produces one form for that record, and every
endpoint returning it agrees. The record itself belongs to the storage component
([Orm](../orm/README.md)).

```text
// Wrong. Each endpoint decides what a record looks like.
return api.jsonResponseFromArray({ "id": order.id, "total": order.total })

// Right. The component produces the form.
return api.jsonResponseFromEntity(order)
```

`jsonFromEntities` does the same for many, and it keys them the way
`jsonFromObjects` does.

## A throwable becomes an envelope in middleware

The component ships `ApiThrowableCaughtMiddleware` for the stage that catches a
throwable. Name it in your config for that stage, and every failure in the
request takes one path into an envelope. That includes a failure raised before
the handler ran.
[`LIFECYCLE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/convention/LIFECYCLE.md)
lists the stages.

```text
class Config implements ConfigContract, HttpConfigContract {
    throwableCaughtMiddleware = [
        ApiThrowableCaughtMiddleware,
    ]
}
```

The middleware sets the outcome, not the diagnosis: the status, the status code
the response already carries, and a code that finds the log record. The
throwable's own text and trace stay out of that envelope, because an envelope
reaches a client. The component that records a throwable with its trace is Log
([Log](../log/README.md)).

```text
{
    "data": { "traceCode": "9b2c0e4a1f6d83b7c5e2a40d18fe7c36" },
    "message": null,
    "errors": [],
    "warnings": [],
    "status": "error",
    "statusCode": 500
}
```

A failure rendered by whatever caught it is how an API ends up with several
error shapes. One stage, one shape.

`jsonFromException` maps a throwable where a caller catches one itself. The
middleware is the path every other failure takes, so a call site rarely reaches
for the mapping.

## Service registration

The component's provider binds:

| Binding       | Is                          |
| ------------- | --------------------------- |
| `ApiContract` | the component's own manager |

The manager holds a response factory, so it can return a response as well as an
envelope. It also holds the application's debug mode.

```text
class ApiServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            ApiContract: [self, "publishApi"],
        }
    }

    static publishApi(container) {
        app = container.getSingleton(ApplicationContract)

        container.setSingleton(ApiContract, new Api(
            responseFactory: container.getSingleton(ResponseFactoryContract),
            debug: app.getDebugMode(),
        ))
    }
}
```

Resolve `ApiContract` where an endpoint returns a machine-readable response.
[Container](../container/README.md) covers contracts, bindings and the resolve
operations.

```text
class OrderController {

    construct(api: ApiContract, orders: OrderRepository) { }

    show(request) {
        order = this.orders.find(request.params.id)

        return this.api.jsonResponseFromEntity(order)
    }
}
```

A caller constructs the envelope and its data rather than resolving either one.
A caller builds one per response, so neither is a binding.
