# Http

The Http component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Http takes a request and produces a response. It is a **protocol component**, so
it runs the seven-stage pipeline in
[`LIFECYCLE.md`](../convention/LIFECYCLE.md) and shares that pipeline's shape
with Cli, Grpc and Queue.

The component owns no server. A runtime receives the bytes and hands the
component a request; the component never binds a port or reads a socket.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Message`    | the request, the response, and every value type they carry |
| `Routing`    | route data, collection, matching, and dispatch             |
| `Middleware` | the stage contracts and the stage handlers                 |
| `Server`     | the request handler that runs the pipeline                 |
| `Client`     | the outbound client, for a request this application sends  |
| `Struct`     | typed request and response data, validated at the boundary |
| `Throwable`  | the component's throwable contract and its exceptions      |

`Message` and `Routing` each divide again. `Message` holds `Request`,
`Response`, `Header`, `Param`, `File`, `Stream` and `Uri`. `Routing` holds
`Data`, `Collection`, `Collector`, `Matcher`, `Dispatcher`, `Processor`,
`Factory` and `Url`.

---

## Using it

### Configure it and point a runtime at it

The application's HTTP config declares the middleware scheduled at each stage,
one property per stage ([`APPLICATION.md`](APPLICATION.md)). An entry point for
the runtime you deploy on receives the request and drives the component; you
point the runtime at that entry and the framework owns everything inward.

### Declare routes

Declare them in a **route provider** listed in the application's config, as a
literal list the build tool can read. A port whose language has attributes or
decorators may instead declare a route on the controller method and list the
controller classes to scan ([`PROVIDERS.md`](../convention/PROVIDERS.md)).

A route states its path, the methods it answers, its name, and its handler.

### Write a handler

A handler receives the container and the route, and returns a response. It is a
typed callable, not a method the framework finds by name
([`HANDLERS.md`](../convention/HANDLERS.md)).

### Take parameters from the path

A dynamic route declares a parameter per path segment it captures: its name,
whether it is required, its default, and the constraint its value satisfies. The
framework casts the matched value to the declared type **before** the handler
runs, so the handler reads a typed value and never parses a string.

### Read the request

Read a parameter from the collection that matches its **source** — query, server,
cookie, attribute, parsed body or parsed JSON. They are separate because each
has a different trust level, and reading from the right one is how an application
states which it trusts.

Headers, uploaded files and the uri are each their own type with their own
collection. The body is a stream.

### Validate at the boundary

A **struct** turns untyped request data into a typed value. It declares the
fields it accepts, the rules each satisfies, and whether unexpected data is a
failure. Use one wherever a handler would otherwise read raw input.

### Build a response

Pick the response type for the body kind: empty, text, html, json or redirect.
Each sets the status and the content type its kind requires, so the caller states
intent instead of assembling headers. A factory builds them where a handler needs
the container to do it.

Headers and cookies are set through the response's own `with` forms, which return
a copy.

### Attach middleware

Middleware is scheduled **globally** in the config, per stage, or **per route**
on the route itself. Both run at the stage they are scheduled for, global first.

Warning: middleware is appended and never deduplicated. Scheduling the same class
twice at one stage runs it twice, and that is the application's bug to fix rather
than the framework's to hide ([`LIFECYCLE.md`](../convention/LIFECYCLE.md)).

### Refine a declared route

A declaration may be refined rather than repeated, and **which modifiers reach
the whole class is part of the contract**:

| Modifier   | Declared on                        | Applies to                                        |
| ---------- | ---------------------------------- | ------------------------------------------------- |
| path       | the class or a method              | every route in the class, or that method's routes |
| name       | the class or a method              | the same                                          |
| middleware | a method                           | that method's routes, and it repeats              |
| parameter  | a method, or one of its parameters | that method's routes                              |

A path or a name on the class is how a group of related routes shares its prefix
without each route restating it. Middleware and parameters do not reach the class,
so a group that shares middleware schedules it per route or globally in the
config.

### Generate a url

A named route generates its own url, with the parameters supplied. Generate a url
rather than writing a path literal, so a path change does not leave a stale link.
A static route supplies an empty parameter set rather than omitting the argument.

### Know when routes are collected

**Debug mode decides where the collection comes from.** With debug mode on, the
framework builds the collection from the route providers. With it off, it reads
the routing data, which a generated class supplies when one exists and which is
otherwise built from those same providers at boot.

So the cache is never required here either: with no generated class, both
settings build from the providers, and debug mode changes only whether the result
is reused.

Either way the collection is built **once**, on the first resolution of it, and
reused after that — a persistent runtime builds it once per process, not once per
unit of work. So a route that appears in development and not in production is a
generated class that was not regenerated.

A built-in command lists every registered route, which is how you confirm what
the collection actually holds.

### Cache a whole response

A response cache is middleware, and it needs **both** halves registered: the
stage-1 side reads the cache and short-circuits, and the stage-7 side writes it.
Registering only one half silently does nothing useful — the read with no write
never hits, and the write with no read never serves.

The cache directory and the debug flag come from a config contract. An
application that sets them implements that contract on its own config class,
which is why a component's config is split per concern rather than held in one
object ([`COMPONENT_CONFIG.md`](../convention/COMPONENT_CONFIG.md)).

---

## The message value types

Every message type is **immutable**. A reader returns the value, and a `with`
form returns a copy that carries the change. The host object never mutates, so a
request handed to two stages cannot be changed by one of them for the other.
The naming rule is in [`METHOD_NAMING.md`](../convention/METHOD_NAMING.md).

| Type            | Is                                                       |
| --------------- | -------------------------------------------------------- |
| `Request`       | an outbound request: method, uri, headers, body          |
| `ServerRequest` | an inbound request: a request plus the server's own data |
| `Response`      | a status, headers and a body                             |
| `Uri`           | the parts of a uri, parsed                               |
| `Header`        | one header, with a parsed value                          |
| `Stream`        | the body, as a readable and writable sequence of bytes   |
| `UploadedFile`  | one uploaded file and its error state                    |

A **collection** type holds many of one thing: headers, uploaded files, and each
kind of parameter. A collection is a first-class immutable type, never a bare map
a caller reaches into.

**A parameter collection is typed by its source.** Query, server, cookie,
attribute, parsed body and parsed JSON each get their own collection, because
each has a different trust level and a different lifetime. A port does not merge
them into one bag.

**A response type exists per body kind**: empty, text, html, json and redirect.
Each one sets the status and the content type its kind requires, so a caller
states the intent rather than assembling headers.

---

## The stream

The body is a stream in every port, and the contract is the same: read, write,
seek, tell, close, detach, report its size, report whether it is readable,
writable and seekable, and report its metadata.

Two backings satisfy it, and which one a port uses is a language fact:

| Backing             | Used where                                              |
| ------------------- | ------------------------------------------------------- |
| a native handle     | the language has a seekable, synchronous I/O type       |
| an in-memory buffer | the language's streams are asynchronous and cannot seek |

A buffer-backed stream satisfies every method for a message body. It cannot wrap
a file descriptor or a socket, so a port on that backing delivers a large file
through a response type that writes the file directly rather than through the
stream.

---

## Routing

A **route** is an immutable data object. It holds its path, its name, the methods
it answers, its handler, its parameters, and the middleware scheduled at each
stage. It holds no policy a caller sets per request.

The route pipeline is the same in every port:

1. A **collector** reads routes from a provider, or from a declaration on a
   controller class.
2. A **processor** turns a declared path into a matchable form, and compiles the
   pattern for a route that carries parameters.
3. A **collection** holds every route, indexed for lookup rather than scanned,
   and holds each one **as a function that returns it** rather than as the route
   itself.
4. A **matcher** resolves one request to one route, static routes first, then
   dynamic.
5. A **dispatcher** calls the matched route's handler and returns its response.

### The collection defers construction

**A collection holds a function that returns the route, not the route.** Reading a
key runs that function and returns what it produced.

A unit of work matches one route and ignores every other, so constructing all of
them to serve one is waste that scales with the route count. Deferring means a
large application pays only for the routes it actually reaches.

The collection does **not** memoize: each read invokes the function again rather
than replacing the entry with what it returned.

This is a cross-component rule: the listener collection in
[`EVENT.md`](EVENT.md) holds its listeners the same way, for the same reason.

**A static match is attempted before a dynamic one.** A static path is a direct
lookup, and a dynamic path costs a pattern match, so the order is a performance
rule rather than a preference.

**The compiled pattern is a port-local form.** A port compiles the pattern for
the engine its language has, and the compiled form never crosses ports.

---

## Parameters

A route parameter declares a name, whether it is required, a default, and the
constraint its value must satisfy. A port casts the matched value to the declared
type before the handler runs, so a handler receives a typed value and never
parses a string.

---

## The server

The **request handler** owns the pipeline. It declares `run` as the entry point a
runtime calls, and `handle` as the method that produces the response from the
first five stages. The terminal methods and the stage dispatch are in
[`LIFECYCLE.md`](../convention/LIFECYCLE.md).

The handler sets the matched route on the container before it calls the route's
handler, so a route is reachable from the container as well as from the handler's
own parameter ([`HANDLERS.md`](../convention/HANDLERS.md)).

---

## Structs

A **struct** is a typed boundary over request or response data. A request struct
declares the fields it accepts, the rules each one satisfies, and whether extra
data is a failure. A response struct declares what the application returns.

A struct is how an application gets a validated, typed value out of an untyped
request. A port that has no struct subcomponent has a gap.

---

## Permitted variation

| Variation                           | Reason                                                                                      |
| ----------------------------------- | ------------------------------------------------------------------------------------------- |
| the stream's backing                | the language has seekable synchronous I/O or not                                            |
| the compiled route pattern's form   | each regular-expression engine takes its own form                                           |
| a standard-interface wrapper        | a port whose ecosystem has an HTTP message standard ships an adapter beside the native type |
| an `Attribute` routing subcomponent | only a language with attributes declares that way                                           |

A port that ships a wrapper for its ecosystem's standard keeps it **beside** the
native type and never in place of it. The native type is what the framework
passes; the wrapper exists for a third-party library.

Nothing else varies.
