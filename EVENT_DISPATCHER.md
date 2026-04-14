# Event Dispatcher

## Overview

The event dispatcher is responsible for dispatching events to their registered listeners. Valkyrja's
`EventDispatcherContract` extends PSR-14's `EventDispatcherInterface`, adding convenience methods for conditional
dispatch, direct listener invocation, and ID-based dispatch.

The cross-language design has one deliberate asymmetry: **Go, Python, and TypeScript require events to carry
an `eventId()` method** for listener map lookup, while PHP and Java derive the key internally from the language's native
type identity mechanism. This is the same tradeoff as string constants for container binding keys — accepting a small
deviation in three languages to avoid a worse constraint (globally unique class names across the entire event space).

---

## PHP Source Contract (Reference)

```php
interface EventDispatcherContract extends EventDispatcherInterface
{
    /**
     * Dispatch an event to its registered listeners.
     */
    public function dispatch(object $event): object;

    /**
     * Dispatch an event only if it has registered listeners.
     */
    public function dispatchIfHasListeners(object $event): object;

    /**
     * Dispatch an event by its class string identifier.
     * Constructs the event from the class string and optional arguments.
     *
     * @param class-string            $eventId
     * @param array<array-key, mixed> $arguments
     */
    public function dispatchById(string $eventId, array $arguments = []): object;

    /**
     * Dispatch an event by its class string identifier only if it has listeners.
     *
     * @param class-string            $eventId
     * @param array<array-key, mixed> $arguments
     */
    public function dispatchByIdIfHasListeners(string $eventId, array $arguments = []): object;

    /**
     * Dispatch a specific set of listeners for an event.
     */
    public function dispatchListeners(object $event, ListenerContract ...$listeners): object;

    /**
     * Dispatch a single listener for an event.
     */
    public function dispatchListener(object $event, ListenerContract $listener): object;
}
```

---

## Listener Map Key — The Cross-Language Problem

The dispatcher's internal map keys listeners by event type. In PHP and Java this is the FQN class string derived without
any overhead:

- **PHP** — `get_class($event)` returns the FQN string. Unique across the application by virtue of PHP's namespace
  system.
- **Java** — `event.getClass()` returns the `Class<T>` object. Unique by the JVM's class identity.

In Go, Python, and TypeScript, deriving a reliable unique key from a type at runtime is either expensive (Go's
`reflect.TypeOf()`), fragile (TypeScript's `event.constructor.name` breaks under minification), or class-object-based (
Python's `type(event)` forces imports and prevents lazy loading).

Requiring globally unique class names across the entire event space is an unacceptable constraint — it would mean a
`UserCreatedEvent` could never exist in both an `Http` and a `Queue` component.

**Solution:** Events in Go, Python, and TypeScript implement `EventContract` which requires an `eventId()` / `EventId()`
method returning a string constant. The dispatcher calls this method to look up listeners — no reflection, no class
identity, no minification fragility.

---

## EventContract

### PHP and Java

No `eventId()` method required. Events are plain objects. The dispatcher derives the key internally.

```php
// PHP — any object, no interface required for basic dispatch
class UserCreatedEvent
{
    public function __construct(public readonly string $userId) {}
}

// dispatcher uses get_class($event) = 'App\Event\UserCreatedEvent' as key
```

```java
// Java — any object
public class UserCreatedEvent {
    public final String userId;

    public UserCreatedEvent(String userId) {
        this.userId = userId;
    }
}
// dispatcher uses event.getClass() as key
```

### Go

```go
// EventContract — required for all Go events
type EventContract interface {
EventId() string
}

// Event implementation
type UserCreatedEvent struct {
EventContract // embedded — Sindri convention
UserId string
}

func (e *UserCreatedEvent) EventId() string {
return EventConstants.USER_CREATED // string constant from event_constants.go
}
```

### Python

```python
# EventContract — required for all Python events
class EventContract(ABC):
    @abstractmethod
    def event_id(self) -> str: ...


# Event implementation
class UserCreatedEvent(EventContract):
    def __init__(self, user_id: str) -> None:
        self.user_id = user_id

    def event_id(self) -> str:
        return EventConstants.USER_CREATED  # string constant from event_constants.py
```

### TypeScript

```typescript
// EventContract — required for all TypeScript events
interface EventContract {
    eventId(): string
}

// Event implementation
class UserCreatedEvent implements EventContract {
    constructor(readonly userId: string) {
    }

    eventId(): string {
        return EventConstants.USER_CREATED
    }
}
```

---

## Method Availability Per Language

| Method                                   | PHP | Java | Go | Python | TypeScript |
|------------------------------------------|-----|------|----|--------|------------|
| `dispatch(event)`                        | ✅   | ✅    | ✅  | ✅      | ✅          |
| `dispatchIfHasListeners(event)`          | ✅   | ✅    | ✅  | ✅      | ✅          |
| `dispatchById(id, args)`                 | ✅   | ✅    | ❌  | ❌      | ❌          |
| `dispatchByIdIfHasListeners(id, args)`   | ✅   | ✅    | ❌  | ❌      | ❌          |
| `dispatchListeners(event, ...listeners)` | ✅   | ✅    | ✅  | ✅      | ✅          |
| `dispatchListener(event, listener)`      | ✅   | ✅    | ✅  | ✅      | ✅          |

`dispatchById` and `dispatchByIdIfHasListeners` require constructing an event object from a string identifier — possible
in PHP via `new $eventId(...$arguments)` and in Java via reflection, but not expressible without a developer-maintained
factory registry in Go, Python, or TypeScript. The methods are absent from those three language contracts rather than
introducing a factory registration overhead.

---

## Per-Language EventDispatcherContract

### PHP

```php
interface EventDispatcherContract extends EventDispatcherInterface
{
    public function dispatch(object $event): object;
    public function dispatchIfHasListeners(object $event): object;
    public function dispatchById(string $eventId, array $arguments = []): object;
    public function dispatchByIdIfHasListeners(string $eventId, array $arguments = []): object;
    public function dispatchListeners(object $event, ListenerContract ...$listeners): object;
    public function dispatchListener(object $event, ListenerContract $listener): object;
}
```

Dispatcher derives key via `get_class($event)`. Full PSR-14 compliance. No `eventId()` method required on events.

---

### Java

```java
public interface EventDispatcherContract {

    /** Dispatch an event to its registered listeners. */
    Object dispatch(Object event);

    /** Dispatch an event only if it has registered listeners. */
    Object dispatchIfHasListeners(Object event);

    /**
     * Dispatch an event by its class identifier.
     * Constructs the event via reflection from the class and arguments.
     */
    Object dispatchById(Class<?> eventId, Map<String, Object> arguments);

    /** Dispatch by class identifier only if listeners are registered. */
    Object dispatchByIdIfHasListeners(Class<?> eventId, Map<String, Object> arguments);

    /** Dispatch a specific set of listeners for an event. */
    Object dispatchListeners(Object event, ListenerContract... listeners);

    /** Dispatch a single listener for an event. */
    Object dispatchListener(Object event, ListenerContract listener);
}
```

Java's `dispatchById` takes a `Class<?>` object rather than a string — `.class` is the idiomatic Java equivalent of
PHP's `class-string`. Dispatcher derives key via `event.getClass()`. No `eventId()` method required on events.

---

### Go

```go
type EventDispatcherContract interface {
// Dispatch dispatches an event to its registered listeners.
// Uses event.EventId() to look up listeners.
Dispatch(event EventContract) (EventContract, error)

// DispatchIfHasListeners dispatches only if listeners are registered.
DispatchIfHasListeners(event EventContract) (EventContract, error)

// DispatchListeners dispatches a specific set of listeners.
DispatchListeners(event EventContract, listeners ...ListenerContract) (EventContract, error)

// DispatchListener dispatches a single listener.
DispatchListener(event EventContract, listener ListenerContract) (EventContract, error)
}
```

No `DispatchById` — callers construct the event and call `Dispatch()`. Dispatcher uses `event.EventId()` for listener
lookup.

---

### Python

```python
class EventDispatcherContract(ABC):

    @abstractmethod
    def dispatch(self, event: EventContract) -> EventContract:
        """Dispatch an event to its registered listeners."""

    @abstractmethod
    def dispatch_if_has_listeners(self, event: EventContract) -> EventContract:
        """Dispatch only if listeners are registered."""

    @abstractmethod
    def dispatch_listeners(
            self,
            event: EventContract,
            *listeners: ListenerContract,
    ) -> EventContract:
        """Dispatch a specific set of listeners."""

    @abstractmethod
    def dispatch_listener(
            self,
            event: EventContract,
            listener: ListenerContract,
    ) -> EventContract:
        """Dispatch a single listener."""
```

No `dispatch_by_id` — callers construct the event and call `dispatch()`. Dispatcher uses `event.event_id()` for listener
lookup.

---

### TypeScript

```typescript
export interface EventDispatcherContract {

    /** Dispatch an event to its registered listeners. */
    dispatch(event: EventContract): EventContract

    /** Dispatch only if listeners are registered. */
    dispatchIfHasListeners(event: EventContract): EventContract

    /** Dispatch a specific set of listeners. */
    dispatchListeners(event: EventContract, ...listeners: ListenerContract[]): EventContract

    /** Dispatch a single listener. */
    dispatchListener(event: EventContract, listener: ListenerContract): EventContract
}
```

No `dispatchById` — callers construct the event and call `dispatch()`. Dispatcher uses `event.eventId()` for listener
lookup.

---

## PSR-14 Compliance

PHP's `EventDispatcherContract` extends PSR-14's `EventDispatcherInterface`. The `dispatch()` method signature is
identical — any library expecting PSR-14 accepts a Valkyrja dispatcher.

PSR-14 is PHP-specific. Other language ports implement the equivalent concept natively without reference to the PSR.

---

## EventContract vs No EventContract — The Reasoning

Requiring `eventId()` on events in Go, Python, and TypeScript is a conscious deviation from PHP/Java. The alternatives
were worse:

**Global unique class names** — would mean `UserCreatedEvent` could not exist in two different components. Unenforceable
and unreasonable.

**`reflect.TypeOf()` in Go** — real overhead in a hot dispatch path. Non-idiomatic Go.

**`event.constructor.name` in TypeScript** — works in development, breaks silently under minification.

**`type(event)` as map key in Python** — forces class imports, defeats lazy loading.

The `eventId()` method is one line per event class. The string constant comes from the per-component constants file that
already exists for container bindings. The cost is minimal and the developer experience is consistent with the rest of
the framework's cross-language patterns.
