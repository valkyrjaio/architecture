# Http

Implements
[`HTTP.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/http/HTTP.md).

## Introduction

The Http component takes a request and produces a response. It matches the
request to a route, calls that route's handler, and runs middleware at each
stage of the work.

The component owns no server. A runtime receives the bytes and hands the
component a request. The component never binds a port, and it never reads a
socket.

Http is a **protocol component**. It runs a seven-stage pipeline, and it shares
that pipeline's shape with Cli, Grpc and Queue. Every stage carries a name, and
application code attaches middleware to a stage by that name
([`LIFECYCLE.md`](https://github.com/valkyrjaio/architecture/blob/26.x/convention/LIFECYCLE.md)).

The component is six areas:

| Area         | Holds                                                           |
| ------------ | --------------------------------------------------------------- |
| `Message`    | the value types a request and a response are made of            |
| `Routing`    | the routes, and everything that turns a request into one        |
| `Server`     | the handler that runs a unit of work end to end                 |
| `Middleware` | one contract per pipeline stage, and the handlers               |
| `Struct`     | the declared shape of a request's input and a response's output |
| `Client`     | the outbound side, for a request this application sends         |

Every message value type is **immutable**. A reader returns the value, and a
`with` form returns a copy that carries the change. So nothing changes a request
in place. A stage that wants a change returns the copy, and the stage that
follows receives that copy.

This document assumes the framework, where the container resolves each piece and
a provider builds it. [Container](../container/README.md) covers contracts,
bindings and the resolve operations. The message value types need none of that:
construct a request or a response and call it.

## Pointing a runtime at the component

`RequestHandlerContract` is what a runtime calls. The runtime receives the bytes,
builds a `ServerRequestContract`, and hands that request to `run`. Everything
inward belongs to the framework.

```text
requestHandler = container.getSingleton(RequestHandlerContract)

requestHandler.run(request)
```

| Method   | Does                                                                |
| -------- | ------------------------------------------------------------------- |
| `run`    | runs every stage, and sends the response                            |
| `handle` | produces the response from the first five stages, and sends nothing |

An entry point for the runtime you deploy on resolves the handler and calls
`run`.

Warning: `handle` stops after stage 5. A caller that takes `handle` owes stage 6
and stage 7. That caller calls `sendingResponse` first, then **writes the
response that `sendingResponse` returned**, then passes that same response to
`responseSent`. A caller that writes the response it already held discards every
change stage 6 made. Call `run` everywhere else.

## Declaring routes

A route states its path, the methods it answers, its name, and its handler. An
application declares its routes in a **route provider**, which implements
`HttpRouteProviderContract` and returns two lists.

| Method                 | Returns                                      |
| ---------------------- | -------------------------------------------- |
| `getControllerClasses` | the classes a collector reads a marker from  |
| `getRoutes`            | the routes the provider built itself as data |

A provider returns one list, or both. Return an empty list for the way the
provider does not use.

### On a controller method, with a marker

A port whose language has a marker on a method declares the route there, beside
the code the route serves. The provider lists the class, and a collector reads
each marker.

```text
class UserRouteProvider implements HttpRouteProviderContract {

    getControllerClasses() {
        return [UserController]
    }

    getRoutes() {
        return []
    }
}
```

```text
class UserController {

    @Route(path: "/users", name: "users.index")
    @RouteHandler([UserRouteProvider, "indexHandler"])
    index() { }

    @Route(path: "/users", name: "users.store", requestMethods: [RequestMethod.POST])
    @RouteHandler([UserRouteProvider, "storeHandler"])
    store() { }
}
```

The marker declares the route. The marker does not make the framework call the
method it sits on. The route's handler makes that call, and `RouteHandler` names
the handler. A route with no handler answers with a response that has no body.

The marker repeats, so one method serves several paths. Each marker declares its
own route with its own name.

```text
@Route(path: "/", name: "home")
@Route(path: "/welcome", name: "welcome")
@RouteHandler([HomeRouteProvider, "homeHandler"])
home() { }
```

The default request methods are `HEAD` and `GET`. Name `requestMethods` to serve
anything else.

### As data, in the provider

A route is also a data object the provider constructs. Use this where the route
has no controller class, or where another system generates the routes.

```text
class UserRouteProvider implements HttpRouteProviderContract {

    getControllerClasses() {
        return []
    }

    getRoutes() {
        return [
            new Route(
                path: "/users",
                name: "users.index",
                handler: [self, "indexHandler"],
            ),
            new DynamicRoute(
                path: "/users/{id}",
                name: "users.show",
                regex: "",
                parameters: [new Parameter(name: "id", regex: Regex.ID)],
                handler: [self, "showHandler"],
            ),
        ]
    }
}
```

Pass an empty `regex`. The **processor** builds the full match pattern from the
path and the parameters, so a declaration states the path and leaves the pattern
to the framework.

A `Route` and a `DynamicRoute` take the same arguments the marker takes:
`requestMethods`, the five per-route middleware lists, `requestStruct` and
`responseStruct`.

### Listing the providers

The application's component provider returns the route providers from
`getHttpProviders`, and the application's config lists the component provider.

```text
class AppComponentProvider implements ComponentProviderContract {

    getHttpProviders(app) {
        return [new UserRouteProvider()]
    }

    getContainerProviders(app) {
        return [new UserServiceProvider()]
    }

    // ...one method per kind of provider, each returning a list
}
```

`getContainerProviders` returns the service providers that bind the controllers
and the middleware. Every class a handler or a stage resolves needs a binding.

### Know when routes are collected

**Debug mode decides where the collection comes from.** With debug mode on, the
framework builds the collection from the route providers. With it off, the
framework reads the routing data. A generated class supplies that data where one
exists. Where none exists, the framework builds the data from the same
providers at boot.

Either way the collection is built **once**, on the first resolution of it. A
persistent runtime builds it once per process rather than once per request. So a
route that appears in development and not in production is a generated class
that was not regenerated.

A built-in command lists every registered route, which is how you confirm what
the collection holds.

## Writing a handler

A handler is a typed callable. It receives the container and the matched route,
and it returns a response. The framework does not find a handler by name, so a
handler is named in the declaration and nowhere else.

```text
class UserRouteProvider implements HttpRouteProviderContract {

    // ...the two list methods

    static indexHandler(container, route) {
        return container.getSingleton(UserController).index()
    }

    static showHandler(container, route) {
        id = route.getParameter("id").getValue()

        return container.getSingleton(UserController).show(id)
    }
}
```

A handler resolves the controller rather than constructing it, so the controller
takes its own dependencies. Register the controller in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class UserServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            UserController: [self, "publishUserController"],
        }
    }

    static publishUserController(container) {
        container.setSingleton(UserController, new UserController())
    }
}
```

A handler receives the route as a `RouteContract`. **A parameter is read from a
`DynamicRouteContract`** instead. That is the kind a route with a captured
segment carries. So a port with a static type system narrows the route before it
reads a parameter.

The matched route is also on the container under `RouteContract`, so a service
deep in the call reads the route without the handler passing it down.

**Declare the handler as a reference the build tool can read.** A reference to a
class and a method name is written in a generated data class. An anonymous
function is not, so a route that carries one is lost from a cached run.

## Taking parameters from the path

A path segment in braces captures a value, and the route becomes a **dynamic
route**. Declare one parameter per captured segment.

| Parameter declares | Says                                         |
| ------------------ | -------------------------------------------- |
| `name`             | which segment it captures                    |
| `regex`            | the constraint the value satisfies           |
| `cast`             | the type the value is converted to           |
| `isOptional`       | whether the segment may be absent            |
| `default`          | the value to use where the segment is absent |
| `shouldCapture`    | whether the segment's value is stored at all |
| `value`            | the matched value, which the matcher sets    |

A declaration does not set `value`. The matcher sets it, and the handler reads
it.

```text
@Route(path: "/articles/{slug}", name: "articles.show")
@Parameter(name: "slug", regex: Regex.SLUG)
@RouteHandler([ArticleRouteProvider, "showHandler"])
show() { }
```

A port may place the marker on a method parameter instead. Both places collect
the same way.

Warning: the parameter marker does not repeat on a method. A path that captures
more than one segment places one marker per method parameter.

```text
@Route(path: "/{year}/{month}", name: "archive.show")
@RouteHandler([ArchiveRouteProvider, "showHandler"])
show(
    @Parameter(name: "year", regex: Regex.NUM) year,
    @Parameter(name: "month", regex: Regex.NUM) month,
) { }
```

The framework ships a set of named patterns, so a declaration names a shape
rather than spelling a pattern.

| Pattern                                  | Matches                            |
| ---------------------------------------- | ---------------------------------- |
| `Regex.NUM`, `Regex.ID`                  | digits                             |
| `Regex.SLUG`                             | letters, digits and dashes         |
| `Regex.ALPHA`                            | letters                            |
| `Regex.ALPHA_NUM`                        | letters and digits                 |
| `Regex.ALPHA_NUM_UNDERSCORE`             | letters, digits and the underscore |
| `Regex.UUID`, `Regex.ULID`, `Regex.VLID` | an identifier of that format       |
| `Regex.ANY`                              | anything                           |

### Cast the matched value

A matched value arrives as a string. **A cast converts the value before the
handler runs**, so the handler reads a typed value and never parses a string.

```text
@Route(path: "/users/{id}", name: "users.show")
@Parameter(name: "id", regex: Regex.ID, cast: new Cast(CastType.int))
@RouteHandler([UserRouteProvider, "showHandler"])
show() { }
```

```text
static showHandler(container, route) {
    // An int, because of the cast. A raw string without one.
    id = route.getParameter("id").getValue()

    return container.getSingleton(UserController).show(id)
}
```

`Cast` and `CastType` belong to the Type component, which declares the named
types. A cast also takes a type of your own. A cast with `convert` set to false
hands the handler the framework's type object rather than the raw value.

### An optional parameter

Mark the segment optional, and give the parameter a default. The matcher uses
the default where the segment is absent.

```text
@Route(path: "/articles/{page?}", name: "articles.list")
@Parameter(name: "page", regex: Regex.NUM, isOptional: true, default: "1")
@RouteHandler([ArticleRouteProvider, "listHandler"])
list() { }
```

## Reading the request

An inbound request is a `ServerRequestContract`: a request, plus the data the
server itself supplied. Read a parameter from the collection that matches its
**source**.

| Reader             | Returns                     | Holds                        |
| ------------------ | --------------------------- | ---------------------------- |
| `getQueryParams`   | a query collection          | the uri's query string       |
| `getParsedBody`    | a parsed body collection    | the submitted form           |
| `getCookieParams`  | a cookie collection         | what the client sent back    |
| `getServerParams`  | a server collection         | what the runtime supplied    |
| `getAttributes`    | an attribute collection     | what this process set itself |
| `getUploadedFiles` | an uploaded file collection | what the client uploaded     |
| `getHeaders`       | a header collection         | the request's headers        |
| `getBody`          | a stream                    | the request's body           |

**The collections are separate because the sources have different trust
levels.** Reading from the collection that matches the source is how an
application states which source it trusts. A port does not merge them into one
bag.

Each parameter collection answers the same questions:

| Method         | Returns                             |
| -------------- | ----------------------------------- |
| `has`          | whether a key holds a value         |
| `get`          | the value under a key               |
| `getAll`       | every value                         |
| `getOnly`      | the values under the keys you name  |
| `getAllExcept` | every value but the keys you name   |
| `with`         | a copy holding the values you pass  |
| `withAdded`    | a copy holding those values as well |

```text
page  = request.getQueryParams().get("page")
title = request.getParsedBody().get("title")
theme = request.getCookieParams().get("theme")

method = request.getMethod()      // a RequestMethod value
path   = request.getUri().getPath()
```

`get` returns a nested collection where the value is itself a set of values, so
a reader never reaches into a bare map.

### The request kinds

| Kind                        | Is                                              |
| --------------------------- | ----------------------------------------------- |
| `RequestContract`           | an outbound request: method, uri, headers, body |
| `ServerRequestContract`     | an inbound request, with the server's own data  |
| `JsonServerRequestContract` | an inbound request whose body is parsed JSON    |

A JSON request parses its body into a collection of its own, which
`getParsedJson` returns. An entry point chooses which kind it builds, so an
application that serves JSON builds the JSON kind.

### Headers, files and the uri

A header is a `HeaderContract` with a parsed value, and `getHeaderLine` returns
one header's values joined into a string. The framework names every standard
header as a constant, so a reader names the constant rather than a string.

```text
auth = request.getHeaders().getHeaderLine(HeaderName.AUTHORIZATION)
```

An uploaded file reports its own error state. Check the state before you read
the file, because a failed upload has no contents.

```text
file = request.getUploadedFiles().get("avatar")

if (file.getError() == UploadError.OK) {
    file.moveTo("/var/uploads/avatar")
}
```

`getStream` reads the file instead of moving it. `getSize` and
`getClientMediaType` report what the client sent.

The uri is the address in parts, and each part has its own reader:
`getScheme`, `getHost`, `getPort`, `getPath`, `getQuery`, `getFragment` and
`getAuthority`. `getHostPort` and `getSchemeHostPort` return the combined
forms, and `isSecure` answers whether the scheme is secure.

### The body is a stream

A message body is a stream in every port, and the operations are the same
everywhere:

- `read` and `write` move bytes.
- `seek`, `tell` and `rewind` move the position and report it.
- `close` and `detach` end the stream.
- `getSize`, `getMetadata`, `isReadable`, `isWritable` and `isSeekable` report
  what the stream is.

Warning: a read starts at the stream's current position, and a write leaves the
position at the end. Rewind the stream after a write before you read it back.

```text
body = new Stream()
body.write("Welcome")
body.rewind()

contents = body.getContents()
```

Two backings satisfy the contract. A port whose language has seekable
synchronous I/O wraps a native handle. A port whose streams are asynchronous
holds an in-memory buffer instead. A buffer-backed stream cannot wrap a file
descriptor, so a port on that backing delivers a large file through a response
that writes the file directly.

## Validating at the boundary with a struct

A **struct** turns untyped request data into a typed value. It declares the
fields it accepts, the rules each field satisfies, and whether unexpected data
is a failure. Use one wherever a handler would otherwise read raw input.

A struct is a closed set of fields, so a port declares it as whatever its
language uses for one.

### A request struct

A request struct implements `RequestStructContract` and declares one field per
value it accepts. `getValidationRules` returns the rules, which the Validation
component evaluates ([Validation](../validation/README.md)).

```text
struct CreateUserRequestStruct implements RequestStructContract {

    field username

    static getValidationRules(request) {
        username = request.getParsedBody().get("username")

        return {
            "username": [
                new Required(username, "The username is required"),
                new NotEmpty(username, "The username must not be empty"),
            ],
        }
    }
}
```

| Method                                | Answers                                         |
| ------------------------------------- | ----------------------------------------------- |
| `getValidationRules`                  | the rules each field satisfies                  |
| `validate`                            | the validator for this request                  |
| `getDataFromRequest`                  | the declared fields, and nothing else           |
| `determineIfRequestContainsExtraData` | whether the request carries an undeclared field |

Name the struct on the route, and the framework's own middleware enforces it.

```text
@Route(path: "/users", name: "users.store", requestMethods: [RequestMethod.POST])
@RequestStruct(CreateUserRequestStruct.username)
@RouteHandler([UserRouteProvider, "storeHandler"])
store() { }
```

`RequestStructMiddleware` runs at the stage that follows the match. It answers a
request that fails validation with `StatusCode.BAD_REQUEST`, and it answers a
request that carries an undeclared field with `StatusCode.PAYLOAD_TOO_LARGE`. The
handler then reads only the declared fields.

```text
data = CreateUserRequestStruct.getDataFromRequest(request)
```

Warning: the middleware enforces the struct, so a route that names a struct does
nothing until the middleware is scheduled. Schedule
`RequestStructMiddleware` globally, or on that route.

### A response struct

A response struct implements `ResponseStructContract` and maps an internal field
name to the name a client sees.

```text
struct UserResponseStruct implements ResponseStructContract {

    field id       = "userId"
    field username = "userName"
}
```

```text
@Route(path: "/users/{id}", name: "users.show")
@Parameter(name: "id", regex: Regex.ID)
@ResponseStruct(UserResponseStruct.id)
@RouteHandler([UserRouteProvider, "showHandler"])
show() { }
```

`ResponseStructMiddleware` runs at the stage that follows the dispatch. It reads
the response body, keys the data by the declared names, and writes the body
back. A field absent from the data becomes a null in the output, so the shape
stays constant across responses.

Warning: the middleware acts only on a json response, the kind named below. A
handler that returns any other kind keeps the body it built, and the struct
changes nothing.

`getStructuredData` performs the mapping where a caller wants it without the
middleware.

## Building a response

**There is one response kind per kind of body.** Each one sets the status and
the content type its kind requires, so a caller states intent rather than
assembling headers.

| Kind               | Carries                   | Sets                  |
| ------------------ | ------------------------- | --------------------- |
| `Response`         | a body, as a stream       | the status you pass   |
| `EmptyResponse`    | no body                   | the no-content status |
| `TextResponse`     | plain text                | the text content type |
| `HtmlResponse`     | markup                    | the html content type |
| `JsonResponse`     | data, encoded             | the json content type |
| `RedirectResponse` | no body, and a target uri | a redirect status     |

```text
response = Response.create("raw body", StatusCode.OK)
json     = new JsonResponse({ user: "melech" })
html     = new HtmlResponse("<h1>Welcome</h1>")
text     = new TextResponse("Welcome")
empty    = new EmptyResponse()
created  = new JsonResponse({ id: 42 }, StatusCode.CREATED)
```

`Response.create` builds the body stream from a string, so most code reaches for
it rather than for the stream-based construction.

Every status is a value rather than a number. The value reports its own code,
its phrase, and whether it is an error or a redirect.

```text
if (response.getStatusCode().isError()) {
    logger.warning("Upstream failed", { status: response.getStatusCode().code() })
}
```

A json response reads and replaces its own data with `getBodyAsJson` and
`withJsonAsBody`. `withCallback` wraps the body in a named callback and sets the
javascript content type. `withoutCallback` writes the data back as json and sets
the json content type.

### A redirect

A redirect takes a uri rather than a string. A status that is not a redirect
status is a failure rather than a redirect to nowhere.

```text
redirect  = new RedirectResponse(new Uri(path: "/dashboard"))
permanent = new RedirectResponse(new Uri(path: "/dashboard"), StatusCode.MOVED_PERMANENTLY)
```

Two forms rewrite the target from the current request. `back` returns to the
referring address, or to the root where the referrer is missing or belongs to
another host. `secure` targets the secure scheme on the current host.

```text
redirect = new RedirectResponse(new Uri(path: "/")).back(request)
redirect = new RedirectResponse(new Uri(path: "/")).secure("/account", request)
```

To redirect to a route by its name, use `RoutingResponseFactoryContract`. It
builds the path from the route rather than from a literal.

```text
redirect = routingResponseFactory.createRouteRedirectResponse("users.show", { id: 42 })
```

### Headers and cookies

A response is immutable, and its header collection is too. Build the collection
you want, then build the response that carries it.

```text
headers  = response.getHeaders().withAddedHeaders(
    new Header(HeaderName.CACHE_CONTROL, "no-store"),
    new Header("X-Request-Id", "a1b2c3"),
)
response = response.withHeaders(headers)
```

| Method             | Returns a copy that                     |
| ------------------ | --------------------------------------- |
| `withHeader`       | holds one header, replacing a repeat    |
| `withHeaders`      | holds the headers you pass, and no more |
| `withAddedHeaders` | holds those headers as well             |
| `withoutHeader`    | holds every header but the one named    |
| `withoutHeaders`   | holds every header but the ones named   |

A cookie is a header, and the response sets it for you.

```text
cookie   = new Cookie(name: "theme", value: "dark", expire: 1735689600)
response = response.withCookie(cookie)
```

Pass the same cookie to `withoutCookie` to expire a cookie the client holds. The
framework names the common headers as classes of their own, which parse their
own value: a content type, a location, a referrer and a set-cookie.

### Build a response through the factory

`ResponseFactoryContract` builds every kind behind one resolvable contract, so a
service that returns a response does not name a response class.

| Method                   | Builds                   | Takes                               |
| ------------------------ | ------------------------ | ----------------------------------- |
| `createResponse`         | a response               | content, a status, headers          |
| `createTextResponse`     | a text response          | content, a status, headers          |
| `createJsonResponse`     | a json response          | data, a status, headers             |
| `createJsonpResponse`    | a json response, wrapped | a callback, data, a status, headers |
| `createRedirectResponse` | a redirect               | a uri string, a status, headers     |

Every argument is optional, except the callback the wrapped form needs.

```text
response = responseFactory.createJsonResponse({ user: "melech" }, StatusCode.CREATED)
```

### Fail from inside application code

A handler returns the response it means. Code deep inside the application
raises `HttpResponseException` instead, which carries a status, headers, and a
response where one is built already.

```text
throw new HttpResponseException(StatusCode.NOT_FOUND, "Resource not found.")
```

The request handler catches every throwable. For an `HttpResponseException` it
uses the attached response, or it builds a generic response carrying that
throwable's status. **Every other throwable produces a generic server error**,
whatever it carries. The stage that catches a throwable then runs, and middleware
there replaces the response.

## Attaching middleware

The pipeline has seven stages. Each stage has one contract, and one class
implements several contracts to run at several stages. A middleware receives the
stage's handler as its last argument. It calls that handler to continue, or it
returns a response to stop the work.

| Stage             | Contract                            | Receives                               | Returns                  |
| ----------------- | ----------------------------------- | -------------------------------------- | ------------------------ |
| `RequestReceived` | `RequestReceivedMiddlewareContract` | the request                            | a request, or a response |
| `RouteMatched`    | `RouteMatchedMiddlewareContract`    | the request, and the route             | a route, or a response   |
| `RouteNotMatched` | `RouteNotMatchedMiddlewareContract` | the request, and a default response    | a response               |
| `RouteDispatched` | `RouteDispatchedMiddlewareContract` | the request, the response, the route   | a response               |
| `ThrowableCaught` | `ThrowableCaughtMiddlewareContract` | the request, a response, the throwable | a response               |
| `SendingResponse` | `SendingResponseMiddlewareContract` | the request, and the final response    | a response               |
| `ResponseSent`    | `ResponseSentMiddlewareContract`    | the request, and the sent response     | nothing                  |

### What each stage is for

**`RequestReceived` is the only stage that sees every request**, including one
that matches no route. It holds what applies to all of them: a maintenance-mode
check, a rate limit, and a lookup of a cached response. Code here answers at
once and skips the rest.

```text
class MaintenanceModeMiddleware implements RequestReceivedMiddlewareContract {

    construct(underMaintenance: bool) { }

    requestReceived(request, handler) {
        if (this.underMaintenance) {
            return new HtmlResponse(
                "<h1>Service unavailable</h1>",
                StatusCode.SERVICE_UNAVAILABLE,
            )
        }

        return handler.requestReceived(request)
    }
}
```

**`RouteMatched` knows the route, and the handler has not run.** A decision
about this route belongs here: authentication, authorization, resolving a
tenant, validating the input. Code here answers at once instead of letting the
handler run.

```text
class AuthMiddleware implements RouteMatchedMiddlewareContract {

    routeMatched(request, route, handler) {
        if (request.getHeaders().getHeaderLine(HeaderName.AUTHORIZATION) == "") {
            return new RedirectResponse(new Uri(path: "/login"))
        }

        return handler.routeMatched(request, route)
    }
}
```

**`RouteNotMatched` replaces a default response.** The framework supplies a
not-found response, and a not-allowed response where the path matches a route
under another method. Return a response of your own to present the failure the
way the application presents every other one.

**`RouteDispatched` transforms what the handler produced.** It receives the
route as well as the response, so it reads the declaration while it shapes the
output.

```text
class RouteNameHeaderMiddleware implements RouteDispatchedMiddlewareContract {

    routeDispatched(request, response, route, handler) {
        headers = response.getHeaders().withAddedHeaders(
            new Header("X-Route-Name", route.getName()),
        )

        return handler.routeDispatched(request, response.withHeaders(headers), route)
    }
}
```

**`ThrowableCaught` turns a failure into a reply.** It receives the throwable
and a response built from it, so it holds the reporting and the application's
own error presentation.

```text
class ReportThrowableMiddleware implements ThrowableCaughtMiddlewareContract {

    construct(logger: LoggerContract) { }

    throwableCaught(request, response, throwable, handler) {
        this.logger.throwable(throwable, "Http dispatch failed")

        return handler.throwableCaught(request, response, throwable)
    }
}
```

The throwable and its trace belong in a record rather than in the response. Log
records a throwable with its type, its message and its trace
([Log](../log/README.md)).

**`SendingResponse` is the last chance to change the response.** The response is
final, and nothing is written yet. Compression belongs here, and so does a
header that depends on the finished body.

```text
class SecurityHeadersMiddleware implements SendingResponseMiddlewareContract {

    sendingResponse(request, response, handler) {
        headers = response.getHeaders().withAddedHeaders(
            new Header("X-Frame-Options", "DENY"),
            new Header("X-Content-Type-Options", "nosniff"),
        )

        return handler.sendingResponse(request, response.withHeaders(headers))
    }
}
```

**`ResponseSent` runs after the client has the response.** It returns nothing,
and the client sees no effect, so work the caller never waits for belongs here.

```text
class AuditLogMiddleware implements ResponseSentMiddlewareContract {

    construct(logger: LoggerContract) { }

    responseSent(request, response, handler) {
        this.logger.info("Answered " + request.getUri().getPath(), {
            status: response.getStatusCode().code(),
        })

        handler.responseSent(request, response)
    }
}
```

Warning: `SendingResponse` and `ResponseSent` **always run**, including after an
answer at `RequestReceived` and after a throwable. Code at either stage cannot
assume a route matched or a handler ran.

### Globally, per stage

The application's config declares one list per stage. Each list names classes,
and the stage handler resolves each class from the container when the stage
runs.

```text
class Config implements ConfigContract, HttpConfigContract {
    requestReceivedMiddleware = [MaintenanceModeMiddleware]
    routeMatchedMiddleware    = [AuthMiddleware, RequestStructMiddleware]
    routeNotMatchedMiddleware = [ViewRouteNotMatchedMiddleware]
    routeDispatchedMiddleware = [ResponseStructMiddleware]
    throwableCaughtMiddleware = [ReportThrowableMiddleware]
    sendingResponseMiddleware = [SecurityHeadersMiddleware]
    responseSentMiddleware    = [AuditLogMiddleware]
}
```

Every class in a list needs a binding, because the stage handler resolves the
class by its id. Bind your own middleware in a service provider. The framework
publishes its own.

### Per route

Five stages are also scheduled on the route itself. The two global-only stages
are `RequestReceived`, which runs before a route exists, and
`RouteNotMatched`, which runs when no route matched.

A port with a marker for it names the class once, and the collector reads which
stage contracts the class implements.

```text
@Route(path: "/admin", name: "admin.dashboard")
@Middleware(AuthMiddleware)
@Middleware(AuditLogMiddleware)
@RouteHandler([AdminRouteProvider, "dashboardHandler"])
dashboard() { }
```

The stage-named arguments do the same with the stage spelled out, which is also
what a route built as data uses.

```text
@Route(
    path: "/admin",
    name: "admin.dashboard",
    routeMatchedMiddleware: [AuthMiddleware],
    responseSentMiddleware: [AuditLogMiddleware],
)
@RouteHandler([AdminRouteProvider, "dashboardHandler"])
dashboard() { }
```

Global middleware runs first, then the route's own. Both run at the stage they
are scheduled for.

Warning: **middleware is appended and never deduplicated.** Scheduling one class
twice at one stage runs it twice. That is the application's defect to fix, and
the framework does not hide it.

### Cache a whole response

A response cache is middleware, and it needs **both halves scheduled**. The
`RequestReceived` half reads the cache and answers at once. The `ResponseSent`
half writes it.

```text
class Config implements ConfigContract, HttpConfigContract {
    requestReceivedMiddleware = [CacheResponseMiddleware]
    responseSentMiddleware    = [CacheResponseMiddleware]
}
```

Warning: one half alone does nothing useful. A read with no write never hits,
and a write with no read never serves.

The cache key is the request's path and its method. The key covers no query
string, no header and no identity, so the cache replays one entry to every later
caller of that path and method. Keep a response that varies per caller out of
this cache.

Four conditions decide what the cache does:

- **With debug mode on, the read half never serves.** The write half still
  writes, so the cache fills once per key and never hits.
- **An existing entry is never rewritten.** The write half skips a key that
  already holds an entry.
- **An entry expires, and the lifetime is fixed.** The read deletes an expired
  entry and lets the request run. No config property changes the lifetime.
- **A server-error response is never cached**, so a failure does not replay.

`NoCacheResponseMiddleware` serves a different purpose. It sets the headers that
tell the **client** not to store a response. It does not change what this cache
holds.

```text
@Route(
    path: "/account",
    name: "account.show",
    sendingResponseMiddleware: [NoCacheResponseMiddleware],
)
@RouteHandler([AccountRouteProvider, "accountHandler"])
account() { }
```

### The middleware the component ships

| Middleware                        | Stage                                | Does                                   |
| --------------------------------- | ------------------------------------ | -------------------------------------- |
| `CacheResponseMiddleware`         | `RequestReceived` and `ResponseSent` | reads a cached response, and writes it |
| `RedirectTrailingSlashMiddleware` | `RequestReceived`                    | redirects a path with a trailing slash |
| `RequestStructMiddleware`         | `RouteMatched`                       | enforces the route's request struct    |
| `ResponseStructMiddleware`        | `RouteDispatched`                    | shapes the body by the response struct |
| `ViewRouteNotMatchedMiddleware`   | `RouteNotMatched`                    | renders the not-found view             |
| `LogThrowableCaughtMiddleware`    | `ThrowableCaught`                    | records the throwable                  |
| `ViewThrowableCaughtMiddleware`   | `ThrowableCaught`                    | renders the error view                 |
| `NoCacheResponseMiddleware`       | `SendingResponse`                    | tells the client not to store          |

The two that render a view reach the View component
([View](../view/README.md)). The one that records a throwable reaches Log
([Log](../log/README.md)).

## Refining a declared route

A declaration is refined rather than repeated. **Which modifier reaches the
whole class is part of the contract.**

| Modifier   | Declared on                        | Applies to                                        |
| ---------- | ---------------------------------- | ------------------------------------------------- |
| path       | the class or a method              | every route in the class, or that method's routes |
| name       | the class or a method              | the same                                          |
| middleware | a method                           | that method's routes, and it repeats              |
| parameter  | a method, or one of its parameters | that method's routes                              |

A path or a name on the class is how a group of related routes shares its prefix
without each route restating it. On a class the modifier prefixes. On a method
the modifier appends.

```text
@Path("/admin")
@Name("admin")
class AdminController {

    // The path is /admin/dashboard, and the name is admin.dashboard.
    @Route(path: "/dashboard", name: "dashboard")
    @RouteHandler([AdminRouteProvider, "dashboardHandler"])
    dashboard() { }

    // The method modifier appends: /admin/reports/export.
    @Path("/export")
    @Route(path: "/reports", name: "admin.reports.export")
    @RouteHandler([AdminRouteProvider, "exportHandler"])
    export() { }
}
```

A request-method modifier **adds** to the route's methods rather than replacing
them. To serve one method only, name `requestMethods` on the declaration.

```text
// The route answers the default HEAD and GET, and PATCH as well.
@Patch
@Route(path: "/users/{id}", name: "users.update")
@Parameter(name: "id", regex: Regex.ID)
@RouteHandler([UserRouteProvider, "updateHandler"])
update() { }
```

Middleware and a parameter do not reach the class. A group that shares
middleware schedules it per route, or globally in the config.

To share a whole set of defaults across a group, declare a route marker of your
own that carries them. A collector collects a marker by what it extends, so a
marker of your own collects like any other.

```text
class UserRoute extends Route {

    construct(path, name) {
        super(
            path: "/users" + path,
            name: "users." + name,
            routeMatchedMiddleware: [AuthMiddleware],
        )
    }
}
```

```text
class UserController {

    @UserRoute(path: "/{id}", name: "show")
    @Parameter(name: "id", regex: Regex.ID)
    @RouteHandler([UserRouteProvider, "showHandler"])
    show() { }
}
```

Every route declared that way carries the path prefix, the name prefix and the
middleware. The same works for a route built as data: extend the data object,
and return your own from the provider.

## Generating a url

A named route generates its own url. `UrlContract` takes the name and the
parameters, and it returns the path.

```text
path = url.getUrl("users.show", { id: 42 })  // /users/42
path = url.getUrl("users.index", {})         // /users
```

**Generate a url rather than writing a path literal.** A generated url follows a
path change, and a literal becomes a stale link that nothing reports.

A static route supplies an empty parameter set rather than omitting the
argument, so one call shape serves both kinds of route.

## The outbound client

`ClientContract` sends a request this application makes. It takes a request and
returns a response, so the outbound side uses the same message value types the
inbound side uses.

```text
response = client.sendRequest(new Request(
    uri: new Uri(scheme: Scheme.HTTPS, host: "api.example.com", path: "/status"),
    method: RequestMethod.GET,
))

ok   = ! response.getStatusCode().isError()
body = response.getBody().getContents()
```

An **adapter** is the thing that sends. Each one implements `ClientContract`,
and each is registered under its own name, so an application resolves whichever
one it wants. `ClientContract` itself resolves to the adapter the client config
names.

| Adapter       | Does                                      |
| ------------- | ----------------------------------------- |
| `<Lib>Client` | sends over one library the port speaks to |
| `LogClient`   | records the request instead of sending it |
| `NullClient`  | nothing; every response is an empty one   |

Every port ships the log and null adapters. Which library adapter ships is the
port's own choice.

The log and null adapters stand in during development and in a test, so code
that depends on `ClientContract` runs with no network.

## Configuration

The component reads the application's HTTP config for the pipeline, plus one
config contract per concern that configures separately.

**The pipeline config is the application's own.** An HTTP application implements
`HttpConfigContract` on its config class, and the application's boot binds that
class. The framework ships a config class that implements the contract, and the
defaults below are that class's own. Extend that class to take those defaults. A
class that implements the contract from nothing supplies all seven properties
itself.

Warning: nothing falls back. The boot binds the contract only where the
application's own config implements it. An application that implements nothing
has no pipeline config, and the stage handlers then fail to build.

Each **per-concern** contract does have a default. Where your config class
implements one, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves that concern's defaults in place.

```text
// In the component's provider. Your class, or the component's default.
static publishConfig(container) {
    config = container.getSingleton(ConfigContract)

    if (config implements HttpServerConfigContract) {
        container.setSingleton(HttpServerConfigContract, config)

        return
    }

    container.setSingleton(HttpServerConfigContract, new HttpServerConfig())
}
```

### The pipeline config

`HttpConfigContract` declares one list per stage. Each list names the middleware
classes scheduled globally at that stage.

| Property                    | Holds                            | Default                                                         |
| --------------------------- | -------------------------------- | --------------------------------------------------------------- |
| `requestReceivedMiddleware` | the `RequestReceived` middleware | an empty set                                                    |
| `routeMatchedMiddleware`    | the `RouteMatched` middleware    | an empty set                                                    |
| `routeNotMatchedMiddleware` | the `RouteNotMatched` middleware | `ViewRouteNotMatchedMiddleware`                                 |
| `routeDispatchedMiddleware` | the `RouteDispatched` middleware | an empty set                                                    |
| `throwableCaughtMiddleware` | the `ThrowableCaught` middleware | `LogThrowableCaughtMiddleware`, `ViewThrowableCaughtMiddleware` |
| `sendingResponseMiddleware` | the `SendingResponse` middleware | an empty set                                                    |
| `responseSentMiddleware`    | the `ResponseSent` middleware    | an empty set                                                    |

Warning: a list you pass **replaces** the default. Re-list a built-in you keep,
or the list you pass takes that built-in away.

### The server config

| Property                | Holds                              | Default                        |
| ----------------------- | ---------------------------------- | ------------------------------ |
| `responseCacheFilePath` | where a cached response is written | the framework's cache location |

A default that is only known at run time is not a value the config holds. The
framework resolves its own cache location after it sets the base path, so the
default here is nothing and the provider resolves the location.

```text
class Config implements ConfigContract, HttpConfigContract, HttpServerConfigContract {
    requestReceivedMiddleware = [CacheResponseMiddleware]
    responseSentMiddleware    = [CacheResponseMiddleware]

    responseCacheFilePath = "/app/storage/cache/response"
}
```

### The client config

| Property        | Holds                                    | Default                        |
| --------------- | ---------------------------------------- | ------------------------------ |
| `defaultClient` | the adapter `ClientContract` resolves to | the port's own library adapter |

Name a different adapter, and every call that resolves `ClientContract` gets it.
No call site changes, and a call that names an adapter directly is unaffected.

```text
class Config implements ConfigContract, HttpClientConfigContract {
    defaultClient = NullClient
}
```

### The log adapter's config

| Property              | Holds                          | Default            |
| --------------------- | ------------------------------ | ------------------ |
| `httpClientLogLogger` | which logger the records go to | the default logger |

The log adapter names a logger rather than holding one. So the records go
wherever that logger goes. Log covers the adapters that record
([Log](../log/README.md)).

Each property carries the component name as well as the adapter name, because
`Client` names a subcomponent of more than one component. So one config class
sets a different logger per component.

## Service registration

The component has one provider per area, and each one binds what that area
offers.

| Provider's area | Binds                                                                                                                                      |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `Message`       | the response factory                                                                                                                       |
| `Routing`       | the router, the collection, the matcher, the collector, the processor, the url builder, the routing response factory, and the routing data |
| `Middleware`    | one stage handler per stage, each holding the config's middleware for that stage                                                           |
| `Server`        | the server config, the request handler, and each middleware the component ships                                                            |
| `Client`        | the client config, the log adapter's config, `ClientContract`, and each adapter under its own name                                         |

```text
class HttpServerServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            HttpServerConfigContract: [self, "publishConfig"],
            RequestHandlerContract: [self, "publishRequestHandler"],
            CacheResponseMiddleware: [self, "publishCacheResponseMiddleware"],
            // ...one publisher per middleware the component ships
        }
    }

    static publishRequestHandler(container) {
        app = container.getSingleton(ApplicationContract)

        container.setSingleton(RequestHandlerContract, new RequestHandler(
            container: container,
            router: container.getSingleton(RouterContract),
            requestReceivedHandler: container.getSingleton(RequestReceivedHandlerContract),
            throwableCaughtHandler: container.getSingleton(ThrowableCaughtHandlerContract),
            sendingResponseHandler: container.getSingleton(SendingResponseHandlerContract),
            responseSentHandler: container.getSingleton(ResponseSentHandlerContract),
            debug: app.getDebugMode(),
        ))
    }
}
```

The **request handler** owns the pipeline. It declares `run` as the entry point
a runtime calls, and `handle` as the method that produces the response from the
first five stages. The two terminal methods each carry their stage's name, so the
stage-6 method is `sendingResponse` and the stage-7 method is `responseSent`.

A stage handler holds the middleware scheduled at its stage, and `add` is what
registers one. The middleware provider reads the config and adds each class as
it builds the handler.

```text
class HttpMiddlewareServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            RequestReceivedHandlerContract: [self, "publishRequestReceivedHandler"],
            // ...one publisher per stage
        }
    }

    static publishRequestReceivedHandler(container) {
        config = container.getSingleton(HttpConfigContract)

        handler = new RequestReceivedHandler(container)

        container.setSingleton(RequestReceivedHandlerContract, handler)

        handler.add(config.requestReceivedMiddleware)
    }
}
```

The client's provider follows the pattern every service component follows.
`ClientContract` resolves to the adapter the config names. The provider also
binds each adapter under its own name, for a caller that asks for one
directly.

```text
class HttpClientServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            HttpClientConfigContract: [self, "publishConfig"],
            HttpClientLogConfigContract: [self, "publishLogConfig"],
            ClientContract: [self, "publishClient"],
            LogClient: [self, "publishLogClient"],
            NullClient: [self, "publishNullClient"],
        }
    }

    static publishClient(container) {
        config = container.getSingleton(HttpClientConfigContract)

        container.setSingleton(
            ClientContract,
            container.getSingleton(config.defaultClient),
        )
    }

    static publishLogClient(container) {
        config = container.getSingleton(HttpClientLogConfigContract)

        container.setSingleton(LogClient, new LogClient(
            container.getSingleton(config.httpClientLogLogger),
        ))
    }
}
```
