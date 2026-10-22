# ATTRIBUTES.md — what a route attribute declares

The **cross-language** convention for the arguments a route attribute declares.
It applies in every Valkyrja repository, and in every protocol component.

It holds no code example. Each port's own `README.md` holds the examples.

This document uses one term for the facility. PHP has an attribute, Java has an
annotation, and Python and TypeScript have a decorator. The segment that holds a
route attribute carries one name already. The segment is `Attribute/` in PHP and
TypeScript, and `attribute/` in Java. So this document writes **attribute**
throughout.

Go has no such facility. A Go application declares each route in a provider
list, which every port supports, and that is a language limit rather than a gap.
[`HANDLERS.md`](HANDLERS.md) holds the rest of what an attribute is: it is
metadata, it registers nothing, each of its arguments is a literal, and a
handler is named rather than inlined. This document holds which arguments the
attribute takes.

---

## 1. The argument set is the route's own field set

**A route attribute declares one argument for each field of the route that the
attribute builds.** The route data class is the one statement of what a route
holds. The attribute adds no argument for a value the route cannot hold.

Two kinds of field do not map one to one:

- **A field the framework derives takes no argument.** Http derives a dynamic
  route's regular expression from the path and the parameters. Grpc derives a
  route's service and its method name back out of the method. No attribute
  declares a derived field, because an argument for one lets an application
  contradict the framework.
- **A composed field takes one argument for each of its parts.** The Grpc
  route's method is the one composed field. The attribute on the class declares
  the service name as a part, and the route attribute declares the remote
  method's name as the other part. Rule 3 holds the order the parts join in.

Warning: for every other field, an argument the attribute leaves out makes that
field undeclarable from the attribute. An application then reaches the field only
through a provider's prebuilt route. Nothing reports the omission, because the
attribute is still valid.

An argument the attribute adds is worse. The argument accepts a value, and the
route drops it.

Each protocol carries its own fields. Http carries the path, the name, the
request methods and the two structs, and a dynamic route carries the parameters
as well. Cli carries the name, the description, the help text, the arguments and
the options. Queue carries the name and the description. Grpc carries the
method, the two streaming flags and the two message types. Every protocol
carries a list for each of its middleware stages, and every protocol carries the
handler.

---

## 2. The handler is one of those arguments, and it is optional

**Every route attribute takes the handler as an argument, and the argument is
optional.** The argument is optional because the convenience attribute in rule 3
declares the same field.

An application must name the handler on the route attribute or on the
convenience attribute. The port reads both spellings into the one field.

---

## 3. A convenience attribute declares one field

**A convenience attribute declares one of the route's fields, apart from the
route attribute.** Http's `Name` declares the name, and Http's `Path` declares
the path.

A convenience attribute is a second spelling, never a second field. How the
spellings combine follows the kind of field, and each kind combines one way:

- A single-valued field takes the convenience attribute's value. The handler and
  the two structs are single-valued.
- A list field adds the convenience attribute's value. The middleware lists, the
  request methods, the parameters, the arguments and the options are lists.
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

**The convenience attribute for middleware declares no stage.** A middleware
class states its own stages through the contracts that the class implements.
The port reads the class into each stage list whose contract the class
implements.

---

## 4. The spelling varies, the argument names do not

A port spells an argument list the way its language does.

| Language   | An argument list is      |
| ---------- | ------------------------ |
| PHP        | named arguments          |
| Java       | annotation elements      |
| Kotlin     | annotation elements      |
| TypeScript | one options object       |
| Python     | keyword arguments        |
| Go         | no attribute, a provider |

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
