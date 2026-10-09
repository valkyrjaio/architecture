# Handlers

The **cross-language** definition of a handler: the typed callable that a router
or a dispatcher invokes to do an application's own work.

This document holds no code example, for the reason in
[`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md). Each port's component
`README.md` holds the examples.

---

## What a handler is

A route holds a handler. A listener holds a handler. The handler is an explicit
typed callable, and the language enforces its signature, so a wrong signature
fails to compile or fails static analysis rather than failing at run time.

A handler is **not** a framework base class to extend, and not a method the
framework finds by name. It is a value the route or the listener carries.

---

## The signature

| Handler     | Parameters                          | Returns                       |
| ----------- | ----------------------------------- | ----------------------------- |
| Http route  | container, route                    | the response contract         |
| Cli route   | container, route                    | the output contract           |
| Grpc route  | container, route                    | the service response contract |
| Queue route | container, route                    | the job result                |
| Listener    | container, a map of named arguments | the language's any type       |

**A route handler's second parameter is the route itself.** It is the route as it
stands after the middleware for the matching stage ran, so a handler sees the
route the middleware produced rather than the route the collection holds.

**A listener's second parameter is a map**, because a listener has no route. The
map carries the event under a known key, plus whatever arguments the dispatch
passed.

The route is **also set on the container** immediately before the handler runs.
So a handler reaches the route two ways, and a class the handler constructs can
resolve the route without the handler passing it down. The request, the input and
the call are reachable from the container only; they are never handler
parameters.

Warning: resolving an id the container does not hold **throws**. So reading the
route from the container is a claim that the framework registered it, and that
claim holds only inside a dispatch.

---

## The marker is metadata only

A port whose language has attributes, annotations or decorators lets an
application declare a route on the method that handles it. That declaration is
**metadata and nothing else**. It records what to register. It never registers
anything itself, and nothing reads it at run time on the hot path.

The build tool reads the declaration statically. A port with no cache reads the
same declaration through reflection at boot.

A port whose language has no such facility declares routes only in a provider
list, and that is a language limit rather than a gap.

---

## Registration

A route provider declares its routes two ways, and a port supports both it can:

- a **literal list** of routes, which the build tool reads statically
- a **list of controller classes** to scan, for a port whose language declares a
  route on the class

The literal list is the portable form. Every port supports it, and a port whose
language cannot scan supports only it.

**A handler is named, not inlined, wherever a declaration is read statically.** A
route in a provider list and a handler on an attribute both name a method, because
the build tool has to carry the handler into generated output and cannot carry a
function body. Two languages also reject an inline function in an attribute
argument outright, since the argument has to be a constant expression.

An inline function is fine where nothing reads the declaration — a route built and
registered at run time, for instance. That route cannot reach the generated cache
either way ([`DATA_CACHE.md`](DATA_CACHE.md)).

Each list is a literal with no conditional logic, because the build tool reads
the declaration rather than running it ([`PROVIDERS.md`](PROVIDERS.md)).

---

## Why the build tool does not generate the handler

The build tool generates route data. It does **not** generate the handler, and
the reason is a correctness rule rather than a limit.

**Every port runs correctly with no cache** ([`DATA_CACHE.md`](DATA_CACHE.md)). The
uncached path and the cached path must behave identically, so the handler the
cache carries has to be the same handler the uncached path uses. A generated
handler would be a second implementation, and the two would drift.

Generating one would also need the tool to infer every dependency a method
takes, from the method body, in five languages. The application already states
the handler it wants, so inferring it buys nothing.

---

## Why a route is not declared the way a binding is

A container binding is a map from an id to a callback that registers it, and the
build tool reads that map. A natural question is whether a route could be
declared the same way, which would remove the route list entirely.

It cannot, and the reason is what the two declarations are **keyed by**. A
binding is keyed by the thing it produces, so the key is enough to find the
callback. A route is keyed by a path and a method, which the route data holds
rather than the declaration, and a listener is keyed by an event id the listener
holds. So the key is not available until the route or the listener is built, and
a map keyed by something the value has to produce is not a map the tool can read.

A route list states the routes. That is why it exists.

---

## Permitted variation

| Variation                                  | Reason                                        |
| ------------------------------------------ | --------------------------------------------- |
| how a callable is spelled                  | each language names its own function type     |
| whether a route may be declared on a class | only a language with attributes or decorators |
| the map type for a listener's arguments    | each language names its own map               |

The parameters, their order and what each one holds do not vary.
