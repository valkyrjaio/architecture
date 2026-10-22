# ATTRIBUTES.md — what a route attribute declares

The **cross-language** convention for the arguments a route attribute declares.
It applies in every Valkyrja repository, and in every protocol component.

It holds no code example, and it names no component's own attribute or data
class. A rule that carries a real name goes false when that name changes, and
the reader then trusts the document. So every rule below states what generates a
name, and each port reads its own names back out of the rule.

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
- **A composed field takes one argument for each of its parts.** An attribute on
  the class declares the leading part, and the route attribute declares the rest.
  Section 3 holds the order the parts join in.

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
same field. The port reads both spellings into the one field.

---

## 3. A convenience attribute declares one field

**A convenience attribute declares one of the route's fields, apart from the
route attribute.** It is a second spelling, never a second field.

How the two spellings combine follows the kind of field, and each kind combines
one way:

- A single-valued field takes the convenience attribute's value.
- A list field adds the convenience attribute's value.
- A composed field joins its parts in one order. The attribute on the class gives
  the leading part, the route attribute gives the next, and a convenience
  attribute on the method gives the last.

**A convenience attribute for a composed field reaches every route in the
class.** A controller therefore declares a shared leading part once, rather than
once for each route.

Warning: a convenience attribute for a list field does not reach the class. A
group of routes that shares a list value declares it for each route, or globally
in the config.

**A convenience attribute for a middleware list declares no stage.** It names one
middleware class, and the stages come from that class rather than from the
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

## 5. A name comes from what the attribute declares

Two rules generate every attribute name, so no port chooses one.

- **A route attribute takes the name of the data class that it builds.** An
  attribute that builds a route is named for the route class, and an attribute
  that builds a finer data class is named for that class.
- **A convenience attribute takes the name of the field that it declares.** It
  builds no data class, so the field is what names it. An attribute that declares
  a part of a composed field is named for that part.

A name that repeats its own segment breaks the second rule, because the segment
carries that meaning already ([`STRUCTURE.md`](STRUCTURE.md)).

Warning: an attribute named for the protocol's own vocabulary hides the pair. A
reader cannot tell which class or field the attribute declares, and each port
then names the pair its own way.
