# Contracts

Valkyrja uses contracts — interfaces or abstract classes — to define the shape every implementation must satisfy. This
document covers how contracts are defined, enforced, and extended in each language, and how they map to each other.

---

## What a Contract Is

A contract defines:

- The methods a class must implement
- The parameter types and return types of those methods
- The intent and behaviour expected (documented in the contract, not enforced by the type system)

Contracts are never instantiated directly. They exist solely to define the shape of something, not to provide
implementation. Implementations satisfy the contract — the framework and application code depend on the contract, never
the implementation.

---

## Language Mechanisms

### PHP — `interface`

PHP's `interface` keyword defines a contract. A class implements one or more interfaces with `implements`. The PHP
engine enforces at runtime that all interface methods are implemented — a missing method raises a fatal error on class
load.

```php
interface StreamContract extends Stringable
{
    public function read(int $length): string;
    public function write(string $data): int;
    public function close(): void;
}

class Stream implements StreamContract
{
    public function read(int $length): string { /* ... */ }
    public function write(string $data): int { /* ... */ }
    public function close(): void { /* ... */ }
}
```

**Abstract classes** are used in Valkyrja for base implementations that share logic across implementations — they extend
an interface and provide partial implementation:

```php
abstract class AbstractStream implements StreamContract
{
    // shared logic — e.g. getMetadata() derived from isReadable()/isWritable()
    public function getMetadata(): array
    {
        return [
            'readable' => $this->isReadable(),
            'writable' => $this->isWritable(),
        ];
    }

    // still abstract — implementations must provide these
    abstract public function read(int $length): string;
    abstract public function write(string $data): int;
    abstract public function close(): void;
}
```

**Multiple interfaces** — PHP classes can implement any number of interfaces:

```php
class Stream implements StreamContract, \Psr\Http\Message\StreamInterface
{
    // satisfies both contracts simultaneously
}
```

**Type hints** — PHP 8.0+ supports union types (`int|string`), intersection types (`StreamContract&Stringable`), and
nullable types (`?string`). PHPStan and Psalm enforce contract compliance statically.

---

### Java — `interface`

Java's `interface` keyword defines a contract. A class implements one or more interfaces with `implements`. The Java
compiler (`javac`) enforces at compile time that all interface methods are implemented — a missing method is a compile
error.

```java
package io.valkyrja.http.message.stream.contract;

public interface StreamContract {
    String read(int length) throws RuntimeException;

    int write(String data) throws RuntimeException;

    void close();
}

public class Stream implements StreamContract {
    public String read(int length) { /* ... */
        return "";
    }

    public int write(String data) { /* ... */
        return 0;
    }

    public void close() { /* ... */ }
}
```

**Abstract classes** — same role as PHP: partial implementation shared across concrete classes:

```java
public abstract class AbstractStream implements StreamContract {

    // shared implementation
    public Map<String, Object> getMetadata() {
        return Map.of(
                "readable", isReadable(),
                "writable", isWritable()
        );
    }

    // still abstract — subclasses must implement
    public abstract String read(int length);

    public abstract int write(String data);

    public abstract void close();
}
```

**Default methods** — Java 8+ interfaces support `default` method implementations, allowing interface evolution without
breaking existing implementations:

```java
public interface StreamContract {
    String read(int length) throws RuntimeException;

    // default implementation — implementing classes inherit this
    // but can override if needed
    default String getContents() throws RuntimeException {
        return read(Integer.MAX_VALUE);
    }
}
```

**Multiple interfaces** — Java classes can implement any number of interfaces:

```java
public class Stream implements StreamContract, AutoCloseable {
    // satisfies both
}
```

**Generics** — Java's type system supports generics in interfaces:

```java
public interface ContainerContract {
    <T> T make(Class<T> key);

    <T> void bind(Class<T> key, Supplier<T> factory);
}
```

---

### Go — `interface`

Go's `interface` keyword defines a contract. Go uses **structural typing** (duck typing) — a type satisfies an interface
simply by implementing all its methods. No `implements` declaration is needed. The Go compiler enforces satisfaction at
compile time when an interface type is used.

```go
package stream

type StreamContract interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	Seek(offset int64, whence int) (int64, error)
	Tell() (int64, error)
}

// ValkyrjaStream satisfies StreamContract implicitly
// — no 'implements StreamContract' declaration needed
type ValkyrjaStream struct {
	data     []byte
	position int64
}

func (s *ValkyrjaStream) Read(p []byte) (int, error)                   { /* ... */ return 0, nil }
func (s *ValkyrjaStream) Write(p []byte) (int, error)                  { /* ... */ return 0, nil }
func (s *ValkyrjaStream) Close() error                                 { return nil }
func (s *ValkyrjaStream) Seek(offset int64, whence int) (int64, error) { /* ... */ return 0, nil }
func (s *ValkyrjaStream) Tell() (int64, error)                         { return s.position, nil }
```

**Compile-time satisfaction check** — Go's convention for explicitly asserting a type satisfies an interface (useful as
a compile-time guard):

```go
// compile-time assertion — fails to compile if ValkyrjaStream
// does not satisfy StreamContract
var _ StreamContract = (*ValkyrjaStream)(nil)
```

**Interface composition** — Go interfaces are composed by embedding:

```go
// stdlib interfaces composed into a Valkyrja interface
type ReadWriteSeekCloser interface {
io.Reader
io.Writer
io.Seeker
io.Closer
}
```

**No abstract classes** — Go has no abstract class mechanism. Shared logic is provided via embedding structs:

```go
// shared base — embedded in concrete types
type baseStream struct {
readable bool
writable bool
}

func (b *baseStream) IsReadable() bool { return b.readable }
func (b *baseStream) IsWritable() bool { return b.writable }

// ValkyrjaStream embeds baseStream — inherits IsReadable/IsWritable
type ValkyrjaStream struct {
baseStream // embedded — methods promoted
data     []byte
position int64
}
```

**Multiple interface satisfaction** — a Go type satisfies as many interfaces as it has methods for, automatically:

```go
// ValkyrjaFileStream satisfies both StreamContract and io.ReadWriteSeeker
// because os.File (embedded) satisfies io.ReadWriteSeeker natively
type ValkyrjaFileStream struct {
*os.File
readable bool
writable bool
}
```

---

### Python — `ABC` (Abstract Base Class)

Python has no native `interface` keyword. The `abc` module (`ABC` and `abstractmethod`) provides the equivalent. `ABC`
marks a class as abstract — it cannot be instantiated. `@abstractmethod` marks individual methods as required. A
subclass that fails to implement all abstract methods raises `TypeError` when instantiated.

```python
from abc import ABC, abstractmethod


class StreamContract(ABC):

    @abstractmethod
    def read(self, length: int) -> str:
        """Read up to length bytes from current position."""

    @abstractmethod
    def write(self, data: str) -> int:
        """Write data to the stream. Returns bytes written."""

    @abstractmethod
    def close(self) -> None:
        """Close the stream."""


class Stream(StreamContract):

    def read(self, length: int) -> str:
        return ""  # actual implementation

    def write(self, data: str) -> int:
        return 0

    def close(self) -> None:
        pass
```

**Enforcement:**

```python
class BadStream(StreamContract):
    pass  # missing read(), write(), close()


BadStream()
# TypeError: Can't instantiate abstract class BadStream
# with abstract methods close, read, write
```

**`@abstractmethod` with body** — abstract methods can have a body in Python, which subclasses can call via `super()`:

```python
class StreamContract(ABC):

    @abstractmethod
    def get_metadata(self) -> dict:
        # default implementation — subclasses call super().get_metadata()
        # to get the base keys and extend from there
        return {
            'readable': self.is_readable(),
            'writable': self.is_writable(),
        }
```

**Multiple inheritance** — Python supports multiple inheritance; a class can satisfy multiple contracts:

```python
class Stream(StreamContract, AnotherContract):
    # must implement all abstract methods from both
    pass
```

**Type checking** — `@abstractmethod` alone does not enforce types at runtime. mypy and pyright validate type hints
statically. For runtime enforcement use `beartype` or `typeguard`.

**`ABC` vs `ABCMeta`** — `ABC` is a convenience class equivalent to `class Foo(metaclass=ABCMeta)`. Always use `ABC` —
it is cleaner and idiomatic:

```python
# these are equivalent — always use the first form
class StreamContract(ABC): ...


class StreamContract(metaclass=ABCMeta): ...
```

---

### TypeScript — `interface`

TypeScript's `interface` keyword defines a contract. Like Go, TypeScript uses structural typing — a class satisfies an
interface if it has all the required members, regardless of whether it explicitly declares `implements`. The TypeScript
compiler (`tsc`) enforces satisfaction at compile time.

```typescript
interface StreamContract {
    read(length: number): string

    write(data: string): number

    close(): void

    seek(offset: number, whence?: number): void

    tell(): number
}

class Stream implements StreamContract {
    read(length: number): string {
        return ''
    }

    write(data: string): number {
        return 0
    }

    close(): void {
    }

    seek(offset: number, whence = 0): void {
    }

    tell(): number {
        return 0
    }
}
```

**`implements` is optional but recommended** — TypeScript satisfies interfaces structurally (duck typing), but
explicitly declaring `implements` gives a clearer error message when a method is missing:

```typescript
// without implements — error points to usage site
// with implements — error points to the class definition
class BadStream implements StreamContract {
    // missing read(), write() etc.
    // error: Class 'BadStream' incorrectly implements interface 'StreamContract'
}
```

**`abstract class`** — TypeScript supports abstract classes for shared implementation, same role as PHP/Java abstract
classes:

```typescript
abstract class AbstractStream implements StreamContract {

    // shared implementation
    getMetadata(): Record<string, unknown> {
        return {
            readable: this.isReadable(),
            writable: this.isWritable(),
        }
    }

    // abstract — subclasses must implement
    abstract read(length: number): string

    abstract write(data: string): number

    abstract close(): void

    abstract isReadable(): boolean

    abstract isWritable(): boolean
}
```

**Interface extension** — TypeScript interfaces can extend other interfaces:

```typescript
interface ReadableStreamContract extends StreamContract {
    getContents(): string
}
```

**`readonly` properties** — TypeScript interfaces support `readonly` for immutable data contracts:

```typescript
interface HttpRouteContract {
    readonly path: string
    readonly method: string
    readonly name?: string
}
```

**Type erasure** — TypeScript interfaces exist only at compile time. At runtime there is no `StreamContract` object, no
reflection, no `instanceof` check against an interface. `instanceof` works only for classes:

```typescript
// compile time — fine
const stream: StreamContract = new Stream()

// runtime — StreamContract does not exist
stream instanceof StreamContract  // ReferenceError: StreamContract is not defined

// runtime — works (class exists at runtime)
stream instanceof Stream  // true
```

This is why TypeScript container binding keys must be string constants — interfaces cannot be used as keys at runtime
because they do not exist.

---

## Cross-Language Summary

|                | Keyword                   | Typing                        | Enforcement             | Abstract class   | Multiple contracts      | Runtime reflection                    |
|----------------|---------------------------|-------------------------------|-------------------------|------------------|-------------------------|---------------------------------------|
| **PHP**        | `interface`               | Nominal                       | Runtime (fatal error)   | `abstract class` | ✅ multiple `implements` | ✅ `instanceof`, `ReflectionClass`     |
| **Java**       | `interface`               | Nominal                       | Compile time            | `abstract class` | ✅ multiple `implements` | ✅ `instanceof`, reflection API        |
| **Go**         | `interface`               | Structural                    | Compile time (at usage) | embedded struct  | ✅ implicit, automatic   | ⚠️ `reflect.TypeOf()` only via values |
| **Python**     | `ABC` + `@abstractmethod` | Structural (+ optional hints) | Runtime (`TypeError`)   | same `ABC` class | ✅ multiple inheritance  | ✅ `isinstance()`, `__mro__`           |
| **TypeScript** | `interface`               | Structural                    | Compile time            | `abstract class` | ✅ multiple `implements` | ❌ erased at runtime                   |

---

## Naming Convention

All Valkyrja contracts follow the `*Contract` suffix convention across all languages:

```
StreamContract
ContainerContract
RouterContract
HttpRouteProviderContract
ServiceProviderContract
ComponentProviderContract
ApplicationContract
```

This applies in all five languages. The suffix is the signal to any developer that this type is a contract — never
instantiated, always implemented.
