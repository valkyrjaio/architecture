# Event

Implements
[`EVENT.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/event/EVENT.md).

## Introduction

The Event component lets one part of an application announce that something
happened, and lets other parts act on it without the announcing part knowing they
exist.

An **event** is an object. A **listener** is something to call when that event is
dispatched. The component holds the listeners, and dispatching an event calls the
ones registered for it.

Use it where the thing that happened matters to code that the code causing it
should not have to know about. An order being placed is an event; sending the
confirmation, updating the ledger and warming a cache are listeners.

This document assumes the framework, where the container resolves the dispatcher
and a provider registers the listeners.

## Resolving the dispatcher

A contract is a resolvable set of methods, independent of the class that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

```text
dispatcher = container.getSingleton(EventDispatcherContract)

dispatcher.dispatch(new OrderPlaced(order))
```

## Defining an event

An event is a plain object carrying what a listener needs. Nothing to extend and
nothing to register.

```text
class OrderPlaced {
    construct(order: Order) { }
}
```

### An event may carry its own arguments

Implement the arguments contract where the event is built from values rather than
constructed directly, which is what dispatching by id does.

```text
class OrderPlaced implements ArgumentsCapableEventContract {

    setArguments(arguments) {
        this.order = arguments["order"]

        return this
    }
}
```

### An event may collect what its listeners return

Implement the collecting contract where the caller wants the listeners' results.
Each listener's return value is added, and the caller reads them after
dispatching.

```text
class PriceRequested implements DispatchCollectableEventContract {

    addDispatch(dispatch) {
        this.quotes.push(dispatch)
    }

    getDispatches() {
        return this.quotes
    }
}
```

```text
event = dispatcher.dispatch(new PriceRequested(basket))

cheapest = min(event.getDispatches())
```

An event that implements neither contract is a plain event, and dispatching it
returns it unchanged.

## Dispatching

| Method                       | Does                                                       | Returns   |
| ---------------------------- | ---------------------------------------------------------- | --------- |
| `dispatch`                   | calls every listener for the event                         | the event |
| `dispatchIfHasListeners`     | the same, and does nothing where none are registered       | the event |
| `dispatchById`               | builds the event from an id and arguments, then dispatches | the event |
| `dispatchByIdIfHasListeners` | the same, and does nothing where none are registered       | the event |

`dispatch` returns the event, so a collecting event is read from the return
value.

```text
// Announce it. Nothing here knows who listens.
dispatcher.dispatch(new OrderPlaced(order))

// Build from an id and arguments, for a caller holding the id as a value.
dispatcher.dispatchById(OrderPlaced, { order: order })
```

### Use the conditional form where building the event costs something

`dispatch` builds the event before it knows whether anything listens.
`dispatchIfHasListeners` checks first, so an event whose construction reads a
database is not built for nothing.

```text
// The event is built either way.
dispatcher.dispatch(new ReportReady(expensiveSummary()))

// Checked first.
dispatcher.dispatchByIdIfHasListeners(ReportReady, { id: report.id })
```

## Listeners

A listener is a declaration: which event, a name, and what to call. It is data
rather than a closure, so the build tool can write it out and a cached run can
read it back.

```text
listener = new Listener(OrderPlaced, "send-confirmation", [Mailer, "sendConfirmation"])
```

| Part        | Is                                         |
| ----------- | ------------------------------------------ |
| the event   | which event this listener is for           |
| the name    | what identifies this listener among others |
| the handler | what to call, as a named reference         |

**The name matters.** It is how a listener is removed or replaced, and how a
second registration of the same handler is told from the first.

### Declaring listeners in a provider

A provider of listeners declares them, and the component registers them.

```text
class AppListenerProvider implements ListenerProviderContract {

    getListeners() {
        return [
            new Listener(OrderPlaced, "send-confirmation", [Mailer, "sendConfirmation"]),
            new Listener(OrderPlaced, "update-ledger", [Ledger, "record"]),
        ]
    }
}
```

The list is a literal with no conditional logic, because the build tool reads it
as written rather than running it.

A provider also declares `getListenerClasses`, returning the classes to scan for
a declared listener where the port declares one with a marker. A port whose
language has no marker returns an empty list from it.

### Adding and removing at run time

| Method                    | Does                                 |
| ------------------------- | ------------------------------------ |
| `addListener`             | adds one                             |
| `removeListener`          | removes one                          |
| `removeListenerById`      | removes one by its name              |
| `hasListener`             | answers whether one is registered    |
| `setListenersForEvent`    | replaces every listener for an event |
| `removeListenersForEvent` | removes every listener for an event  |

```text
collection = container.getSingleton(ListenerCollectionContract)

collection.addListener(
    new Listener(OrderPlaced, "audit", [Audit, "record"])
)
```

**Every reading and writing method has a by-id form.** A caller holding the event
object uses the object form; a caller holding only its id uses the id form. The
two reach the same listeners.

| By object                 | By id                         |
| ------------------------- | ----------------------------- |
| `hasListenersForEvent`    | `hasListenersForEventById`    |
| `getListenersForEvent`    | `getListenersForEventById`    |
| `setListenersForEvent`    | `setListenersForEventById`    |
| `removeListenersForEvent` | `removeListenersForEventById` |

### Asking what is registered

| Method                   | Returns                                |
| ------------------------ | -------------------------------------- |
| `getListeners`           | every listener                         |
| `getEvents`              | every event anything is registered for |
| `getEventsWithListeners` | every event, with its listeners        |

## Exporting and importing

| Method        | Does                          |
| ------------- | ----------------------------- |
| `getData`     | returns the listeners as data |
| `setFromData` | loads listeners from data     |

The build tool reads the providers and the markers, writes the result as data,
and a cached run loads it rather than collecting again. The collection and the
collector are separate for this reason: a cached run loads the collection and
runs no collector.

## Service registration

The component's provider binds:

| Binding                      | Is                                                      |
| ---------------------------- | ------------------------------------------------------- |
| `EventDispatcherContract`    | the dispatcher                                          |
| `ListenerCollectionContract` | the listeners                                           |
| `ListenerCollectorContract`  | the collector for the way this port declares a listener |

```text
class EventServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            ListenerCollectionContract: [self, "publishCollection"],
            EventDispatcherContract: [self, "publishDispatcher"],
        }
    }

    static publishDispatcher(container) {
        container.setSingleton(EventDispatcherContract, new EventDispatcher(
            container.getSingleton(ContainerContract),
            container.getSingleton(ListenerCollectionContract),
        ))
    }
}
```
