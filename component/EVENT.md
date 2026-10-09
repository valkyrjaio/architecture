# Event

The Event component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

The component dispatches an event to the listeners registered for it. An event
says something happened. A listener does the work that the event causes, and the
code that dispatches the event does not know what that work is.

The component has **no middleware pipeline**. It resolves no route, so it has no
stage to run middleware at. The pipeline belongs to the protocol components
([`LIFECYCLE.md`](../convention/LIFECYCLE.md)).

---

## Hierarchy

| Subcomponent | Holds                                                    |
| ------------ | -------------------------------------------------------- |
| `Dispatcher` | the dispatcher, the one entry point a caller uses        |
| `Collection` | the listener collection, keyed by event id               |
| `Collector`  | the readers that build listeners from a declaration      |
| `Contract`   | the contracts an event implements to opt into a behavior |
| `Data`       | the listener data object, and the generated listener set |
| `Provider`   | the listener provider contract                           |
| `Throwable`  | the component's throwable contract and its exceptions    |

A port adds `Attribute` when the language declares a listener with an attribute
or an annotation, and `Constant` when it needs string binding keys.

---

## The event id

A listener is registered against an **event id**, and the id is what the
collection is keyed by. The id is a string, and it names one event type.

How the id is derived is the component's one real asymmetry, and the language
decides it:

| Where the id comes from      | So an event                                       |
| ---------------------------- | ------------------------------------------------- |
| the language's type identity | needs no method, and any object dispatches        |
| a method on the event        | implements `EventContract` and returns its own id |

A language that keeps its types at run time and can name one cheaply derives the
id itself. A language that erases its types, or that pays an import to name one,
cannot. Those languages declare `EventContract` with one method that returns the
id, and an event implements it.

This is the same trade that
[`CONTAINER_BINDINGS.md`](../convention/CONTAINER_BINDINGS.md) makes for a
binding key, and it is accepted for the same reason: the alternative is an event
name that has to be globally unique in every port.

Warning: the id is shared vocabulary. Two components must not register a listener
under the same id for different events.

---

## Using it

### Define an event

An event is a plain object the application owns. It carries the data the
listeners need, as typed properties with a reader for each one. It extends no
framework class.

A port whose language cannot derive the event id from the type implements the
event contract and returns the id. Every other port needs nothing.

### Dispatch it

Dispatch the event object when you hold one. Dispatch by **id** when you do not,
and the dispatcher resolves the event from the container — so dispatching by id
is a claim that the id is a registered binding
([`CONTAINER.md`](CONTAINER.md)).

Use the `IfHasListeners` form when building the event is expensive and nothing
may be listening. It checks the collection first and skips the work.

**Dispatch returns the event.** A listener may change it, so the caller reads the
result from the object the dispatch returned rather than the one it passed in.

### Write a listener

A listener's handler receives the container and a map. The event is in the map
under a known key, and the handler resolves whatever else it needs from the
container ([`HANDLERS.md`](../convention/HANDLERS.md)).

### Pass data from the call site

A dispatch by id has no event object to carry the caller's data. An event that
needs it implements the arguments-capable contract, which declares one method
that receives the call site's arguments. The dispatcher calls it after it
resolves the event and before it invokes any listener. The event stores them as
typed properties and exposes a reader for each.

### Collect what listeners return

**The dispatcher discards a handler's return value by default.** An event that
needs the values implements the dispatch-collectable contract, which declares one
method to receive each value and one to read them all back.

The dispatcher passes every handler's return value, in invocation order,
including the empty value from a handler that returns nothing. This is the shape
for a pipeline where each listener contributes one part of a result.

### Stop the remaining listeners

An event that may be stopped implements the stoppable contract. The dispatcher
checks after **every** listener, so a stopped event runs no further listener. A
port whose language has a standard interface for this uses the standard one.

### Register the listeners

Declare a **listener provider** and list it in the application's config. The
provider declares its listeners as a literal list, and a port whose language
declares a listener on the class itself may instead list the classes to scan
([`PROVIDERS.md`](../convention/PROVIDERS.md)).

A listener may also be added to or removed from the collection at run time. The
collection is not frozen once the application boots.

---

## Canonical operations

Every port declares these on the dispatcher, with no addition and no omission.

| Operation                    | Does                                               |
| ---------------------------- | -------------------------------------------------- |
| `dispatch`                   | dispatch an event to its listeners                 |
| `dispatchIfHasListeners`     | dispatch only when a listener is registered        |
| `dispatchById`               | dispatch by event id, and resolve the event itself |
| `dispatchByIdIfHasListeners` | the same, and only when a listener is registered   |
| `dispatchListeners`          | invoke a given set of listeners                    |
| `dispatchListener`           | invoke one listener                                |

`dispatch` returns the event. A listener receives the event and may change it, so
the caller reads the result from the returned object.

`dispatchById` resolves the event **from the container**, so dispatching by id is
a claim that the id is a registered binding.

---

## What a listener is

A listener is a data object, not a class to subclass. It holds three things, and
each one has a reader and a `with` form that returns a copy:

- the **event id** it listens for
- its own **name**, unique within the collection
- its **handler**, the typed callable the dispatcher invokes

The handler signature is in [`HANDLERS.md`](../convention/HANDLERS.md). A
listener's handler is the one handler whose second parameter is a map rather than
a route, because a listener has no route.

---

## The collection

**The collection holds a function that returns each listener, not the listener.**
The function runs on the first read for that key, and the collection keeps what
it returned, so an application pays only for the listeners it dispatches. The
route collections in [`HTTP.md`](HTTP.md) and the other protocols hold their
routes the same way, for the same reason.

The collection is addressable two ways for every operation: by the listener or
the event itself, and by its id. So each read and write appears twice, once bare
and once with an `ById` suffix.

It answers whether a listener is registered, adds and removes one, reports every
listener for an event, replaces the whole set for an event, and exports and
imports its own data for the cache.

A listener may be added and removed at run time. The collection is not frozen
once the application boots.

---

## Event behaviors a port must support

An event opts into a behavior by implementing a contract. Each one is optional,
and an event that implements none still dispatches.

| Behavior             | Means                                                       |
| -------------------- | ----------------------------------------------------------- |
| stoppable            | a listener can stop the remaining listeners from running    |
| arguments capable    | the dispatcher passes the caller's arguments into the event |
| dispatch collectable | the event records what each listener returned               |

**Stop means stop.** The dispatcher checks after every listener, so a stopped
event runs no further listener. A port that holds a language-standard interface
for this uses the standard one rather than declaring its own.

---

## Listener registration

A **listener provider** declares its listeners, the same way a service provider
declares its bindings ([`PROVIDERS.md`](../convention/PROVIDERS.md)). It declares
them two ways, and a port supports both:

- a literal list of listeners, which the build tool reads statically
- a list of classes to scan, for a port whose language declares a listener on the
  class itself

A port whose language has no attribute or decorator supports only the literal
list, and that is a language limit rather than a gap.

---

## Permitted variation

| Variation                        | Reason                                               |
| -------------------------------- | ---------------------------------------------------- |
| whether an event declares its id | the language can or cannot name a type cheaply       |
| an `Attribute` subcomponent      | only a language with attributes declares that way    |
| a standard stoppable interface   | a port uses its language's standard where one exists |

Nothing else varies. The dispatcher surface is the same in every port.
