# ATTRIBUTES.md — what a route attribute declares

The **cross-language** convention for the arguments a route attribute declares.
It applies in every Valkyrja repository, and in every protocol component.

It holds no code example.

This document uses one term for the facility. PHP has an attribute, Java has an
annotation, and Python and TypeScript have a decorator. One term for one thing is
the rule, so this document writes **attribute** throughout.

Go has no such facility. A Go application declares each route in a provider
list, which every port supports. That is a language limit rather than a gap.

[`HANDLERS.md`](HANDLERS.md) holds the rest of what an attribute is:

- It is metadata.
- It registers nothing.
- Each of its arguments is a literal.
- A handler is named rather than inlined.

This document holds which arguments the attribute takes.

---

## 1. The argument set is the route's own field set

**A route attribute declares one argument for each field of the route that the
attribute builds.** The route data class is the one statement of what a route
holds. The attribute adds no argument for a value the route cannot hold.

Two kinds of field do not map one to one:

- **A field the framework derives from other fields takes no argument.** An
  argument for a derived field lets an application contradict the framework.
- **A composed field takes one argument for each of its parts.** The Grpc
  route's key is built from the service name and the remote method's name. The
  attribute on the class declares the service name, and the route attribute
  declares the remote method's name. Section 3 holds the order the parts join in.

Warning: for every other field, an argument the attribute leaves out makes that
field undeclarable from the attribute. An application then reaches the field only
through a provider's prebuilt route. Nothing reports the omission, because the
attribute is still valid.

An argument the attribute adds is worse. The argument accepts a value, and the
route drops it.

Each protocol's own route holds its own fields. Every protocol's route holds a
handler, and a list for each of the protocol's middleware stages.

---

## 2. The handler is one of those arguments, and it is optional

**Every route attribute takes the handler as an argument, and the argument is
optional.** The argument is optional because a convenience attribute declares the
same field, and that attribute is named `RouteHandler`. The port reads both
spellings into the one field.

---

## 3. A convenience attribute declares one field

**A convenience attribute declares one of the route's fields, apart from the
route attribute.** Http's `Name` declares the name, and Http's `Path` declares
the path.

A convenience attribute is a second spelling, never a second field. How the
spellings combine follows the kind of field, and each kind combines one way:

- A single-valued field takes the convenience attribute's value. The handler is
  single-valued.
- A list field adds the convenience attribute's value. A middleware list is a
  list.
- A composed field joins its parts in one order, and the attribute on the class
  gives the leading part. A path or a name takes three parts: the class prefix,
  the route attribute's value, then the convenience attribute on the method.

**A path and a name on the class reach every route in the class.** A controller
therefore declares a path prefix and a name prefix once, rather than once for
each route. An attribute on the class also declares a part of a composed field,
which is how Grpc declares the service name once for every method of a service.

Warning: middleware and a parameter do not reach the class. A group of routes
that shares middleware schedules it for each route, or globally in the config
([`HTTP.md`](../component/HTTP.md)).

**The convenience attribute for middleware declares no stage.** It names one
middleware class, and the stages come from the class rather than from the
attribute.

---

## 4. The spelling varies, the argument names do not

A port spells an argument list the way its language does, and each
`language/<name>/AGENTS.md` holds that port's spelling. PHP passes named
arguments and Java declares annotation elements.

The **names** are the same in every port, because they are the route's own field
names.

**A language that cannot hold the value holds a reference to it.** A Java
annotation element takes a constant only, so it holds no callable. Java spells
the handler as the class and the method name instead.

**A language that cannot hold an absent value marks one with a sentinel.** A
Java annotation element takes no null, so an optional element carries a default
that means absent. A default therefore differs per port, although the argument
names do not. The Java spelling is in
[`language/java/AGENTS.md`](../language/java/AGENTS.md).

---

## 5. An attribute is named after the data class it builds

**A route attribute takes the name of the data class that it builds.** Http, Cli
and Queue each name the attribute `Route`, because each one builds a `Route`. A
finer attribute follows the same rule, so an attribute that builds a `Parameter`
is named `Parameter`.

**An attribute that declares a part of a composed field is named for that part.**
It builds no data class of its own, so Grpc's class attribute is `Service`, for
the service name it declares.

Warning: an attribute named for the protocol's own vocabulary hides the pair. A
reader cannot tell which class the attribute declares, and each port then names
the pair its own way.
