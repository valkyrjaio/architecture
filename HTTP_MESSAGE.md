# HTTP Message — Stream

The `StreamContract` interface is the cross-language abstraction for HTTP message body streams. It is modeled on PSR-7's
`StreamInterface` but simplified for cross-language compatibility. PHP's implementation is a PSR-7 compatible wrapper on
top of the Valkyrja stream — all other languages implement `StreamContract` directly.

---

## PHP Source Contract (Reference)

The PHP `StreamContract` is the canonical definition. All other language contracts map from this:

```php
interface StreamContract extends Stringable
{
    public function __toString(): string;
    public function close(): void;
    public function detach();                          // returns resource|null
    public function getSize(): int;
    public function tell(): int;                       // throws RuntimeException
    public function eof(): bool;
    public function isSeekable(): bool;
    public function seek(int $offset, int $whence = SEEK_SET): void;
    public function rewind(): void;
    public function isWritable(): bool;
    public function write(string $string): int;
    public function isReadable(): bool;
    public function read(int $length): string;
    public function getContents(): string;
    public function getMetadata(): array;
    public function getMetadataItem(string $key): mixed;
}
```

---

## Two Implementation Tiers

### Tier 1 — Native Resource Backing

PHP, Java, Go, and Python all have a native I/O abstraction that maps cleanly to the `StreamContract` interface.
Implementations delegate to the language's native stream — file descriptors, sockets, and memory buffers are all
interchangeable from the contract's perspective.

| Language | Native backing                                         | In-memory equivalent                             |
|----------|--------------------------------------------------------|--------------------------------------------------|
| PHP      | `resource` (stream wrappers)                           | `php://memory`, `php://temp`                     |
| Java     | `SeekableByteChannel` / `InputStream` + `OutputStream` | `ByteArrayInputStream` / `ByteArrayOutputStream` |
| Go       | `io.ReadWriteSeeker` + `io.Closer`                     | custom `[]byte` wrapper with position pointer    |
| Python   | `io.RawIOBase` / `io.BufferedIOBase`                   | `io.BytesIO`                                     |

### Tier 2 — Buffer Backing

TypeScript has no native seekable synchronous I/O abstraction. Node.js streams are async and chunk-based — they do not
support `seek()` or `tell()`. TypeScript implements `StreamContract` using `Uint8Array` with a manual position pointer.

All methods work correctly for HTTP body handling. The difference is that TypeScript cannot wrap file descriptors or
sockets directly. Large file delivery uses `FileResponse` which bypasses the stream entirely.

---

## Method Mapping Across Languages

| Method              | PHP                      | Java                                     | Go                             | Python                  | TypeScript                                  |
|---------------------|--------------------------|------------------------------------------|--------------------------------|-------------------------|---------------------------------------------|
| `toString()`        | `stream_get_contents()`  | read all bytes to String                 | `io.ReadAll()`                 | `.getvalue().decode()`  | `buffer.toString('utf8')`                   |
| `close()`           | `fclose()`               | `.close()`                               | `.Close()`                     | `.close()`              | no-op (GC managed)                          |
| `detach()`          | returns raw `resource`   | returns underlying channel/stream        | returns `io.ReadWriteSeeker`   | returns `io.BytesIO`    | returns `Uint8Array`, stream unusable after |
| `getSize()`         | `fstat()['size']`        | channel size via position                | `Seek(0, SeekEnd)`             | `len(buf.getvalue())`   | `buffer.length`                             |
| `tell()`            | `ftell()`                | channel `.position()`                    | `Seek(0, SeekCurrent)`         | `.tell()`               | manual position field                       |
| `eof()`             | `feof()`                 | position >= size                         | position >= size               | position >= size        | position >= buffer.length                   |
| `isSeekable()`      | stream metadata          | true for ByteArray, true for FileChannel | true if implements `io.Seeker` | `.seekable()`           | always true (Buffer only)                   |
| `seek()`            | `fseek()`                | channel `.position(offset)`              | `.Seek(offset, whence)`        | `.seek(offset, whence)` | manual: position = offset                   |
| `rewind()`          | `rewind()`               | channel `.position(0)`                   | `.Seek(0, SeekStart)`          | `.seek(0)`              | position = 0                                |
| `isWritable()`      | stream metadata          | check channel/stream type                | check interface                | `.writable()`           | flag set at construction                    |
| `write()`           | `fwrite()`               | channel `.write()`                       | `.Write(bytes)`                | `.write(bytes)`         | new `Uint8Array` concat                     |
| `isReadable()`      | stream metadata          | check channel/stream type                | check interface                | `.readable()`           | flag set at construction                    |
| `read()`            | `fread()`                | channel `.read(length)`                  | `.Read(buf[:length])`          | `.read(length)`         | slice from position                         |
| `getContents()`     | `stream_get_contents()`  | read remaining bytes                     | `io.ReadAll()` from position   | `.read()` from position | slice from position to end                  |
| `getMetadata()`     | `stream_get_meta_data()` | constructed manually                     | constructed manually           | mode/name attrs         | constructed manually                        |
| `getMetadataItem()` | key lookup on above      | key lookup on above                      | key lookup on above            | key lookup on above     | key lookup on above                         |

---

## `getMetadata()` — Simplified Cross-Language Map

PSR-7's metadata map contains PHP-specific keys (`timed_out`, `blocked`, `unread_bytes`, `wrapper_type`, `stream_type`)
that have no cross-language equivalent. The cross-language metadata map is simplified:

```
{
    seekable: bool,    // can seek/tell be called
    readable: bool,    // can read/getContents be called
    writable: bool,    // can write be called
    uri:      string,  // file path if file-backed, "php://memory" etc., null for buffer
    mode:     string,  // 'r', 'w', 'r+', 'rb', 'wb', 'r+b' etc.
}
```

PHP's `getMetadata()` returns the full PSR-7 map (for compatibility) with these cross-language keys always present.
Other languages return the simplified map only.

---

## Go Implementation Note — Read-Write-Seek

Go's `bytes.Buffer` supports write and read but does not implement `io.Seeker`. `bytes.Reader` supports `io.Seeker` but
is read-only. For a fully readable, writable, and seekable in-memory stream Go requires a custom implementation. Two
approaches are provided — both are valid, the developer chooses based on their use case.

---

### Approach A — Pure `[]byte` with Position

Simple, self-contained, one source of truth. Seek works perfectly with no synchronization concerns. Slightly less
efficient for many small incremental writes due to `append()` allocation behavior.

**Best for:** typical HTTP body handling — write once or a few times, read once. The default stream type.

```go
// ValkyrjaStream — pure []byte backing with position pointer
// Default stream implementation. Simple, seekable, fully self-contained.
type ValkyrjaStream struct {
data     []byte
position int64
readable bool
writable bool
}

// NewStream creates a read-write stream from a string body.
func NewStream(body string) *ValkyrjaStream {
return &ValkyrjaStream{
data:     []byte(body),
readable: true,
writable: true,
}
}

// NewReadOnlyStream creates a read-only stream — e.g. for request bodies.
func NewReadOnlyStream(body string) *ValkyrjaStream {
return &ValkyrjaStream{
data:     []byte(body),
readable: true,
writable: false,
}
}

// NewWriteOnlyStream creates a write-only stream — e.g. for response bodies being built.
func NewWriteOnlyStream() *ValkyrjaStream {
return &ValkyrjaStream{
data:     []byte{},
readable: false,
writable: true,
}
}

// NewStreamFromBytes creates a read-write stream from a byte slice.
func NewStreamFromBytes(data []byte) *ValkyrjaStream {
return &ValkyrjaStream{
data:     data,
readable: true,
writable: true,
}
}

func (s *ValkyrjaStream) Read(p []byte) (n int, err error) {
if !s.readable {
return 0, errors.New("stream is not readable")
}
if s.position >= int64(len(s.data)) {
return 0, io.EOF
}
n = copy(p, s.data[s.position:])
s.position += int64(n)
return n, nil
}

func (s *ValkyrjaStream) Write(p []byte) (n int, err error) {
if !s.writable {
return 0, errors.New("stream is not writable")
}
s.data = append(s.data[:s.position], p...)
s.position += int64(len(p))
return len(p), nil
}

func (s *ValkyrjaStream) Seek(offset int64, whence int) (int64, error) {
var abs int64
switch whence {
case io.SeekStart:
abs = offset
case io.SeekCurrent:
abs = s.position + offset
case io.SeekEnd:
abs = int64(len(s.data)) + offset
default:
return 0, errors.New("invalid whence value")
}
if abs < 0 {
return 0, errors.New("negative position")
}
s.position = abs
return abs, nil
}

func (s *ValkyrjaStream) Tell() (int64, error) {
return s.position, nil
}

func (s *ValkyrjaStream) Eof() bool {
return s.position >= int64(len(s.data))
}

func (s *ValkyrjaStream) IsSeekable() bool { return true }
func (s *ValkyrjaStream) IsReadable() bool  { return s.readable }
func (s *ValkyrjaStream) IsWritable() bool  { return s.writable }

func (s *ValkyrjaStream) Rewind() error {
s.position = 0
return nil
}

func (s *ValkyrjaStream) GetSize() int64 {
return int64(len(s.data))
}

func (s *ValkyrjaStream) GetContents() (string, error) {
if !s.readable {
return "", errors.New("stream is not readable")
}
result := string(s.data[s.position:])
s.position = int64(len(s.data))
return result, nil
}

func (s *ValkyrjaStream) String() string {
return string(s.data)
}

func (s *ValkyrjaStream) Detach() io.ReadWriteSeeker {
data := s.data
s.data = nil                 // stream now unusable
return bytes.NewReader(data) // read-only seeker over detached data
}

func (s *ValkyrjaStream) Close() error {
s.data = nil
return nil
}

func (s *ValkyrjaStream) GetMetadata() map[string]any {
return map[string]any{
"seekable": true,
"readable": s.readable,
"writable": s.writable,
"uri":      nil,
"mode":     modeString(s.readable, s.writable),
}
}

func (s *ValkyrjaStream) GetMetadataItem(key string) any {
return s.GetMetadata()[key]
}
```

---

### Approach B — Hybrid `bytes.Buffer` + `[]byte`

Uses `bytes.Buffer` as a write accumulator — it uses an efficient internal growth strategy that avoids repeated
allocations on many small writes. The `[]byte` slice remains the seekable source of truth. The buffer is flushed to
`data` before any read or seek operation.

**Best for:** write-heavy streams where the body is built incrementally from many small chunks (e.g. template rendering,
chunked encoding, large response assembly).

**Tradeoff:** adds complexity — two sources of truth must be kept synchronized. The flush step on every read/seek has a
cost. For typical HTTP body sizes this complexity is not worth it.

```go
// ValkyrjaBufferedStream — bytes.Buffer write accumulator + []byte seek source
// Use when building a response body from many small writes.
type ValkyrjaBufferedStream struct {
data     []byte       // seekable source of truth
buf      bytes.Buffer // write accumulator — flushed to data before read/seek
position int64
readable bool
writable bool
}

// NewBufferedStream creates an empty read-write buffered stream.
func NewBufferedStream() *ValkyrjaBufferedStream {
return &ValkyrjaBufferedStream{
data:     []byte{},
readable: true,
writable: true,
}
}

// NewBufferedStreamFromString creates a buffered stream with initial content.
func NewBufferedStreamFromString(body string) *ValkyrjaBufferedStream {
return &ValkyrjaBufferedStream{
data:     []byte(body),
readable: true,
writable: true,
}
}

// NewWriteOnlyBufferedStream creates a write-only buffered stream.
// Best choice for building response bodies from many small chunk writes.
func NewWriteOnlyBufferedStream() *ValkyrjaBufferedStream {
return &ValkyrjaBufferedStream{
data:     []byte{},
readable: false,
writable: true,
}
}

// flush moves any pending buf writes into data before read or seek
func (s *ValkyrjaBufferedStream) flush() {
if s.buf.Len() > 0 {
s.data = append(s.data, s.buf.Bytes()...)
s.buf.Reset()
}
}

func (s *ValkyrjaBufferedStream) Write(p []byte) (n int, err error) {
if !s.writable {
return 0, errors.New("stream is not writable")
}
n, err = s.buf.Write(p)
s.position += int64(n)
return
}

func (s *ValkyrjaBufferedStream) Read(p []byte) (n int, err error) {
if !s.readable {
return 0, errors.New("stream is not readable")
}
s.flush()
if s.position >= int64(len(s.data)) {
return 0, io.EOF
}
n = copy(p, s.data[s.position:])
s.position += int64(n)
return n, nil
}

func (s *ValkyrjaBufferedStream) Seek(offset int64, whence int) (int64, error) {
s.flush()
var abs int64
switch whence {
case io.SeekStart:
abs = offset
case io.SeekCurrent:
abs = s.position + offset
case io.SeekEnd:
abs = int64(len(s.data)) + offset
default:
return 0, errors.New("invalid whence value")
}
if abs < 0 {
return 0, errors.New("negative position")
}
s.position = abs
return abs, nil
}

func (s *ValkyrjaBufferedStream) Tell() (int64, error) {
return s.position, nil
}

func (s *ValkyrjaBufferedStream) Eof() bool {
s.flush()
return s.position >= int64(len(s.data))
}

func (s *ValkyrjaBufferedStream) IsSeekable() bool { return true }
func (s *ValkyrjaBufferedStream) IsReadable() bool  { return s.readable }
func (s *ValkyrjaBufferedStream) IsWritable() bool  { return s.writable }

func (s *ValkyrjaBufferedStream) Rewind() error {
s.flush()
s.position = 0
return nil
}

func (s *ValkyrjaBufferedStream) GetSize() int64 {
s.flush()
return int64(len(s.data))
}

func (s *ValkyrjaBufferedStream) GetContents() (string, error) {
if !s.readable {
return "", errors.New("stream is not readable")
}
s.flush()
result := string(s.data[s.position:])
s.position = int64(len(s.data))
return result, nil
}

func (s *ValkyrjaBufferedStream) String() string {
s.flush()
return string(s.data)
}

func (s *ValkyrjaBufferedStream) Detach() io.ReadWriteSeeker {
s.flush()
data := s.data
s.data = nil
s.buf.Reset()
return bytes.NewReader(data)
}

func (s *ValkyrjaBufferedStream) Close() error {
s.data = nil
s.buf.Reset()
return nil
}

func (s *ValkyrjaBufferedStream) GetMetadata() map[string]any {
return map[string]any{
"seekable": true,
"readable": s.readable,
"writable": s.writable,
"uri":      nil,
"mode":     modeString(s.readable, s.writable),
}
}

func (s *ValkyrjaBufferedStream) GetMetadataItem(key string) any {
return s.GetMetadata()[key]
}
```

---

### Approach C — File-Backed (`os.File`)

`os.File` implements `io.ReadWriteSeeker` and `io.Closer` natively — all operations delegate directly with no custom
position tracking needed. Used for file serving, temp file upload handling, and any stream backed by a file descriptor.

**Best for:** file uploads (multipart temp files), large file serving, streaming reads from disk.

```go
// ValkyrjaFileStream — os.File backing
// All read/write/seek operations delegate to os.File natively.
// Used for uploads, file serving, and any file descriptor-backed stream.
type ValkyrjaFileStream struct {
file     *os.File
readable bool
writable bool
}

// NewFileStream wraps an existing *os.File with explicit read/write flags.
func NewFileStream(file *os.File, readable bool, writable bool) *ValkyrjaFileStream {
return &ValkyrjaFileStream{file: file, readable: readable, writable: writable}
}

// OpenReadOnly opens a file at path for reading only.
func OpenReadOnly(path string) (*ValkyrjaFileStream, error) {
f, err := os.Open(path) // O_RDONLY
if err != nil {
return nil, err
}
return NewFileStream(f, true, false), nil
}

// OpenReadWrite opens or creates a file at path for reading and writing.
func OpenReadWrite(path string) (*ValkyrjaFileStream, error) {
f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
if err != nil {
return nil, err
}
return NewFileStream(f, true, true), nil
}

// OpenWriteOnly opens or creates a file at path for writing only (truncates existing).
func OpenWriteOnly(path string) (*ValkyrjaFileStream, error) {
f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
if err != nil {
return nil, err
}
return NewFileStream(f, false, true), nil
}

// NewTempStream creates a writable temporary file stream.
// Used by the framework for multipart upload handling.
func NewTempStream() (*ValkyrjaFileStream, error) {
f, err := os.CreateTemp("", "valkyrja-stream-*")
if err != nil {
return nil, err
}
return NewFileStream(f, true, true), nil
}

func (s *ValkyrjaFileStream) Read(p []byte) (int, error) {
if !s.readable {
return 0, errors.New("stream is not readable")
}
return s.file.Read(p)
}

func (s *ValkyrjaFileStream) Write(p []byte) (int, error) {
if !s.writable {
return 0, errors.New("stream is not writable")
}
return s.file.Write(p)
}

func (s *ValkyrjaFileStream) Seek(offset int64, whence int) (int64, error) {
return s.file.Seek(offset, whence)
}

func (s *ValkyrjaFileStream) Tell() (int64, error) {
return s.file.Seek(0, io.SeekCurrent)
}

func (s *ValkyrjaFileStream) Eof() bool {
pos, _ := s.Tell()
info, err := s.file.Stat()
if err != nil {
return true
}
return pos >= info.Size()
}

func (s *ValkyrjaFileStream) IsSeekable() bool { return true }
func (s *ValkyrjaFileStream) IsReadable() bool  { return s.readable }
func (s *ValkyrjaFileStream) IsWritable() bool  { return s.writable }

func (s *ValkyrjaFileStream) Rewind() error {
_, err := s.file.Seek(0, io.SeekStart)
return err
}

func (s *ValkyrjaFileStream) GetSize() int64 {
info, err := s.file.Stat()
if err != nil {
return -1
}
return info.Size()
}

func (s *ValkyrjaFileStream) GetContents() (string, error) {
if !s.readable {
return "", errors.New("stream is not readable")
}
data, err := io.ReadAll(s.file)
if err != nil {
return "", err
}
return string(data), nil
}

func (s *ValkyrjaFileStream) String() string {
pos, _ := s.Tell()
_ = s.Rewind()
data, _ := io.ReadAll(s.file)
_, _ = s.file.Seek(pos, io.SeekStart) // restore position
return string(data)
}

func (s *ValkyrjaFileStream) Detach() io.ReadWriteSeeker {
f := s.file
s.file = nil // stream now unusable
return f
}

func (s *ValkyrjaFileStream) Close() error {
return s.file.Close()
}

func (s *ValkyrjaFileStream) GetMetadata() map[string]any {
info, _ := s.file.Stat()
mode := ""
if info != nil {
mode = info.Mode().String()
}
return map[string]any{
"seekable": true,
"readable": s.readable,
"writable": s.writable,
"uri":      s.file.Name(),
"mode":     mode,
}
}

func (s *ValkyrjaFileStream) GetMetadataItem(key string) any {
return s.GetMetadata()[key]
}
```

---

### Shared Helper

```go
// modeString returns the mode string for metadata based on readable/writable flags.
func modeString(readable, writable bool) string {
switch {
case readable && writable:
return "r+"
case readable:
return "r"
case writable:
return "w"
default:
return ""
}
}
```

---

### Which to Use

|                      | `ValkyrjaStream`         | `ValkyrjaBufferedStream`  | `ValkyrjaFileStream`              |
|----------------------|--------------------------|---------------------------|-----------------------------------|
| **Backing**          | `[]byte`                 | `bytes.Buffer` + `[]byte` | `os.File`                         |
| **Default**          | ✅ yes                    | —                         | —                                 |
| **Seekable**         | ✅ always                 | ✅ after flush             | ✅ always (native)                 |
| **In-memory**        | ✅                        | ✅                         | ❌                                 |
| **File-backed**      | ❌                        | ❌                         | ✅                                 |
| **Write efficiency** | standard `append()`      | `bytes.Buffer` growth     | OS-level buffering                |
| **`close()` cost**   | nil the slice            | nil + reset               | `file.Close()` syscall            |
| **Best for**         | HTTP bodies, typical use | Incremental write-heavy   | File serving, uploads, temp files |

All three implement `StreamContract` and are interchangeable from the caller's perspective. The framework uses
`ValkyrjaStream` as the default. `ValkyrjaFileStream` with `NewTempStream()` is used internally for multipart upload
handling.

---

## Per-Language StreamContract

### PHP

PHP implements `StreamContract` directly and also implements PSR-7's `StreamInterface` for compatibility. The underlying
resource may be any PHP stream — `php://memory`, `php://temp`, `php://input`, a file path, a socket, or any registered
stream wrapper.

```php
interface StreamContract extends Stringable
{
    public function __toString(): string;
    public function close(): void;
    public function detach(): mixed;                   // resource|null
    public function getSize(): int;
    public function tell(): int;
    public function eof(): bool;
    public function isSeekable(): bool;
    public function seek(int $offset, int $whence = SEEK_SET): void;
    public function rewind(): void;
    public function isWritable(): bool;
    public function write(string $string): int;
    public function isReadable(): bool;
    public function read(int $length): string;
    public function getContents(): string;
    public function getMetadata(): array;
    public function getMetadataItem(string $key): mixed;
}
```

PSR-7 compatibility:

```php
// Valkyrja's Stream implements both contracts
class Stream implements StreamContract, \Psr\Http\Message\StreamInterface
{
    public function __construct(private mixed $resource) {}
    // delegates all methods to fread/fwrite/fseek/ftell/feof etc.
    // getMetadata() returns full PSR-7 map including cross-language keys
}
```

---

### Java

```java
package io.valkyrja.http.message.stream.contract;

public interface StreamContract {

    /** Get all contents as a string from position 0. */
    String toString();

    /** Close the stream and release underlying resources. */
    void close();

    /**
     * Detach the underlying resource.
     * After detach the stream is unusable.
     * Returns the underlying SeekableByteChannel, InputStream, or OutputStream.
     */
    Object detach();

    /** Get the size of the stream in bytes. */
    long getSize();

    /** Get the current read/write position. */
    long tell() throws RuntimeException;

    /** Whether the current position is at the end of the stream. */
    boolean eof();

    /** Whether the stream supports seeking. */
    boolean isSeekable();

    /**
     * Seek to a position.
     * @param offset byte offset
     * @param whence SEEK_SET=0, SEEK_CUR=1, SEEK_END=2
     */
    void seek(long offset, int whence) throws RuntimeException;

    /** Seek to the beginning of the stream. */
    void rewind() throws RuntimeException;

    /** Whether the stream is writable. */
    boolean isWritable();

    /**
     * Write data to the stream.
     * @return number of bytes written
     */
    int write(String data) throws RuntimeException;

    /** Whether the stream is readable. */
    boolean isReadable();

    /**
     * Read up to length bytes from the stream.
     */
    String read(int length) throws RuntimeException;

    /** Get remaining contents from current position to end. */
    String getContents() throws RuntimeException;

    /** Get stream metadata as a map. */
    Map<String, Object> getMetadata();

    /** Get a single metadata value by key. */
    Object getMetadataItem(String key);
}
```

Java implementation uses `SeekableByteChannel` (NIO) for file-backed streams and a `ByteArrayInputStream`/
`ByteArrayOutputStream` pair for in-memory. A `ValkyrjaStream` wrapper class handles the channel and exposes the
`StreamContract` interface.

---

### Go

```go
package stream

import "io"

// StreamContract is the cross-language stream interface.
// Go's implementation uses io.ReadWriteSeeker + io.Closer as the backing type.
// For in-memory streams, ValkyrjaStream provides the backing implementation.
// For file-backed streams, os.File satisfies io.ReadWriteSeeker + io.Closer natively.
type StreamContract interface {

	// String returns all stream contents as a string from position 0.
	String() string

	// Close closes the stream and releases any underlying resources.
	Close() error

	// Detach separates the underlying resource from the stream.
	// After detach the stream is in an unusable state.
	// Returns the underlying io.ReadWriteSeeker.
	Detach() io.ReadWriteSeeker

	// GetSize returns the size of the stream in bytes.
	GetSize() int64

	// Tell returns the current read/write position.
	Tell() (int64, error)

	// Eof returns true if the position is at the end of the stream.
	Eof() bool

	// IsSeekable returns whether the stream supports seeking.
	IsSeekable() bool

	// Seek seeks to a position. Whence: 0=start, 1=current, 2=end.
	Seek(offset int64, whence int) error

	// Rewind seeks to the beginning of the stream.
	Rewind() error

	// IsWritable returns whether the stream can be written to.
	IsWritable() bool

	// Write writes data to the stream. Returns bytes written.
	Write(data string) (int, error)

	// IsReadable returns whether the stream can be read from.
	IsReadable() bool

	// Read reads up to length bytes from the current position.
	Read(length int) (string, error)

	// GetContents returns all remaining contents from current position.
	GetContents() (string, error)

	// GetMetadata returns the stream metadata map.
	GetMetadata() map[string]any

	// GetMetadataItem returns a single metadata value by key.
	GetMetadataItem(key string) any
}
```

---

### Python

Python's `io.BytesIO` covers the full interface natively. The `StreamContract` is defined as an ABC:

```python
# valkyrja/http/message/stream/contract.py
from abc import ABC, abstractmethod
from typing import Any


class StreamContract(ABC):

    @abstractmethod
    def __str__(self) -> str:
        """Get all contents as a string from position 0."""

    @abstractmethod
    def close(self) -> None:
        """Close the stream and release underlying resources."""

    @abstractmethod
    def detach(self) -> Any:
        """
        Detach the underlying resource (io.BytesIO or file object).
        After detach the stream is unusable.
        """

    @abstractmethod
    def get_size(self) -> int:
        """Get the size of the stream in bytes."""

    @abstractmethod
    def tell(self) -> int:
        """Get the current read/write position."""

    @abstractmethod
    def eof(self) -> bool:
        """Whether the current position is at the end of the stream."""

    @abstractmethod
    def is_seekable(self) -> bool:
        """Whether the stream supports seeking."""

    @abstractmethod
    def seek(self, offset: int, whence: int = 0) -> None:
        """Seek to a position. whence: 0=start, 1=current, 2=end."""

    @abstractmethod
    def rewind(self) -> None:
        """Seek to the beginning of the stream."""

    @abstractmethod
    def is_writable(self) -> bool:
        """Whether the stream can be written to."""

    @abstractmethod
    def write(self, data: str) -> int:
        """Write data to the stream. Returns bytes written."""

    @abstractmethod
    def is_readable(self) -> bool:
        """Whether the stream can be read from."""

    @abstractmethod
    def read(self, length: int) -> str:
        """Read up to length bytes from current position."""

    @abstractmethod
    def get_contents(self) -> str:
        """Get remaining contents from current position to end."""

    @abstractmethod
    def get_metadata(self) -> dict[str, Any]:
        """Get stream metadata."""

    @abstractmethod
    def get_metadata_item(self, key: str) -> Any:
        """Get a single metadata value by key."""
```

Implementation wraps `io.BytesIO` for in-memory and any `io.RawIOBase` / `io.BufferedIOBase` for file-backed:

```python
class Stream(StreamContract):
    def __init__(self, source: io.IOBase | str | bytes = b'',
                 readable: bool = True, writable: bool = True) -> None:
        if isinstance(source, (str, bytes)):
            data = source.encode() if isinstance(source, str) else source
            self._stream = io.BytesIO(data)
        else:
            self._stream = source  # file or socket
        self._readable = readable
        self._writable = writable
```

---

### TypeScript

TypeScript uses `Uint8Array` with a manual position pointer. All `StreamContract` methods work correctly for in-memory
HTTP body handling. `isSeekable()` always returns `true` — the buffer supports seek. `close()` is a no-op. `detach()`
returns the underlying `Uint8Array` and marks the stream unusable.

```typescript
// valkyrja/http/message/stream/contract.ts

export interface StreamContract {

    /** Get all contents as a string from position 0. */
    toString(): string

    /** Close the stream. No-op for buffer-backed streams — GC managed. */
    close(): void

    /**
     * Detach the underlying Uint8Array.
     * After detach the stream is unusable.
     */
    detach(): Uint8Array | null

    /** Get the size of the stream in bytes. */
    getSize(): number

    /** Get the current read/write position. */
    tell(): number

    /** Whether the current position is at the end of the stream. */
    eof(): boolean

    /**
     * Whether the stream supports seeking.
     * Always true for buffer-backed streams.
     */
    isSeekable(): boolean

    /** Seek to a position. whence: 0=start, 1=current, 2=end. */
    seek(offset: number, whence?: number): void

    /** Seek to the beginning of the stream. */
    rewind(): void

    /** Whether the stream can be written to. */
    isWritable(): boolean

    /**
     * Write data to the stream.
     * @returns bytes written
     */
    write(data: string): number

    /** Whether the stream can be read from. */
    isReadable(): boolean

    /** Read up to length bytes from current position. */
    read(length: number): string

    /** Get remaining contents from current position to end. */
    getContents(): string

    /** Get stream metadata. */
    getMetadata(): Record<string, unknown>

    /** Get a single metadata value by key. */
    getMetadataItem(key: string): unknown
}
```

---

## Tier Summary

```
Tier 1 — Native resource backing:
  PHP        ✅  resource / stream wrappers — files, sockets, memory, custom wrappers
  Java       ✅  SeekableByteChannel (NIO) — files, sockets, ByteArray
  Go         ✅  io.ReadWriteSeeker — os.File, ValkyrjaStream (custom []byte impl)
  Python     ✅  io.BytesIO / io.RawIOBase — memory, files, sockets

Tier 2 — Buffer backing (manual position tracking):
  TypeScript ⚠️  Uint8Array — in-memory only, no file descriptor wrapping
                  close() is no-op, detach() returns Uint8Array
                  FileResponse / StreamResponse bypass stream for large payloads
```

---

## PHP PSR-7 Compatibility

PHP's `Stream` class implements both `StreamContract` and PSR-7's `StreamInterface`. Since PSR-7's `StreamInterface` has
an identical method surface to `StreamContract` (it is the origin this contract is modeled on), the PHP implementation
satisfies both with no additional code. The only difference is `getMetadata()` — PHP returns the full PSR-7 metadata
map, other languages return the simplified cross-language map.

```
PSR-7 StreamInterface methods = StreamContract methods (one-to-one)
PHP Stream implements StreamContract + StreamInterface simultaneously
All other language Stream implementations implement StreamContract only
```

This means any PHP library or middleware expecting a PSR-7 stream can receive a Valkyrja `Stream` instance with no
adapter needed.
