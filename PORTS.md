# Valkyrja Ports

## Overview

Valkyrja is being ported to the following languages in priority order based on
developer audience size, architectural fit, and the framework's cross-language
consistency goals:

1. **PHP** — original, production, reference implementation
2. **Java** — in progress, enterprise market
3. **Go** — proof of concept started, cloud-native sweet spot
4. **Python** — large developer audience, clear differentiation from FastAPI
5. **TypeScript / Node.js** — backend framework, explicit registration model

Future ports (lower priority, some nearly free from earlier work):

- **Kotlin** — maps 1:1 from Java port
- **Scala** — JVM overlap with Java/Kotlin
- **Rust** — performance-critical niche, when the time is right
- **Ruby** — existing community, lower growth trajectory

---

## Core Philosophy Across All Ports

Every port shares the same architectural identity:

- **Explicit architecture first** — no magic, no hidden wiring
- **Language idioms respected** — ports feel native, not translated
- **Cross-port legibility** — a Go developer and a PHP developer can read each
  other's Valkyrja code and understand the structure
- **Flexible deployment** — every port supports both CGI/lambda and worker modes
  without requiring the developer to architect differently

---

## PHP

**Status:** Production, reference implementation.

**Key characteristics:**

- `::class` provides compile-time verified FQN string constants
- Attributes (`#[...]`) supported since PHP 8.0
- Interpreted — no compile step, cache is optional
- FrankenPHP and OpenSwoole/Swoole enable true worker mode
- CGI mode with pre-built cache is the original and primary deployment model

**Language-specific notes:**

- All exception roots available natively (`\Throwable`, `\RuntimeException`,
  `\InvalidArgumentException`)
- Reflection via `ReflectionClass::getFileName()` resolves class name to source
  file
- nikic/php-parser provides full AST for the build tool
- `::class` used for container bindings — compiler-verified, not a raw string
- Dynamic method dispatch via reflection available but deprecated in favor of
  closure-based handlers

**Worker mode:** FrankenPHP (Go runtime underneath), OpenSwoole/Swoole (
coroutine-based C extension)

---

## Java

**Status:** In progress.

**Key characteristics:**

- `.class` provides compile-time verified type tokens (`Class<T>`)
- Annotations (`@...`) native, processed at compile time via annotation
  processor
- Compiled — JVM, Project Loom virtual threads for concurrency
- Lambda/serverless cold start is a known pain point — cache data files critical
- JavaPoet used for code generation in annotation processor

**Language-specific notes:**

- `IllegalArgumentException` replaces `InvalidArgumentException` as the language
  root — Valkyrja names `ValkyrjaInvalidArgumentException` for cross-port
  parity, extending `IllegalArgumentException` under the hood
- All Valkyrja exceptions extend `RuntimeException` (unchecked) — no `throws`
  declarations needed
- Trees API in annotation processor can extract lambda source text from AST at
  compile time
- Spotless flags same-named exceptions across packages — `ComponentName*` prefix
  resolves this
- Kotlin maps 1:1 — identical roots, all exceptions unchecked by default

**Worker mode:** JVM stays warm, Project Loom virtual threads, no FrankenPHP
equivalent needed

**Build toolchain:** Spotless, ArchUnit, ErrorProne + NullAway, JUnit 5

**Package namespace:** `io.valkyrja`

**Build tool:** Gradle for internal tooling

---

## Go

**Status:** Proof of concept started.

**Key characteristics:**

- No class hierarchy — errors implement the `error` interface
- No annotations or decorators of any kind
- Compiled to a single static binary — startup is near-instant
- Goroutines provide native concurrency — no worker mode complexity
- Explicit over implicit is a core Go philosophy

**Language-specific notes:**

- `ValkyrjaThrowable` is an unexported interface implementing `error`
- `ValkyrjaRuntimeException` and `ValkyrjaInvalidArgumentException` are exported
  structs implementing `error`
- "Abstract" enforcement via unexported embedded fields — outside packages
  cannot construct component categoricals directly, must use provided
  constructors
- No `.class` / `::class` equivalent — string constants used for container
  bindings
- No annotations — explicit route registration only, no annotated class scanning
- `errors.As` / `errors.Is` are the idiomatic catch boundaries
- Result pattern native — `(T, error)` return is idiomatic Go
- `go/analysis` AST tooling used by the build tool for cache data file
  generation

**Worker mode:** Go binary stays running, goroutines handle concurrency
natively — single bootstrap always

**Port order:** Container → Dispatch → Event → Application → CLI → HTTP → Bin

---

## Python

**Status:** Planned.

**Key characteristics:**

- Interpreted — no compile step, cache optional exactly like PHP
- Decorators are executable at import time — self-registration pattern works
- GIL limits true thread parallelism — async (ASGI/Uvicorn) is the idiomatic
  concurrency model
- Python 3.13+ free-threaded mode (experimental GIL removal) worth watching
- `inspect.getfile()` resolves class to source file — equivalent of PHP's
  `ReflectionClass::getFileName()`

**Language-specific notes:**

- `BaseException` is the Throwable equivalent
- `RuntimeError` is the RuntimeException equivalent
- `ValueError` is the closest to `InvalidArgumentException` —
  `ValkyrjaInvalidArgumentException` extends `ValueError` for language-level
  catchability while preserving cross-port parity name
- Abstract classes via ABC — `abstract class` raises `TypeError` on direct
  instantiation
- No checked exceptions — convention and ABC enforce the hierarchy
- `class_(cls)` helper (note: `class` is a reserved word) constructs FQN from
  `__module__.__qualname__`
- Decorators self-register at import time — route provider imports trigger
  registration automatically
- `ast` module provides full AST for the build tool

**Deployment models:**

- CGI mode: cache optional, works without it, build tool generates cache for
  production
- ASGI worker (Uvicorn/Hypercorn): bootstrapped once per process, cache largely
  irrelevant
- Gunicorn + Uvicorn workers: multi-process, each bootstrapped once
- Granian (Rust-based): newer option, true multi-threaded workers via Rust
  runtime

**Valkyrja differentiation from FastAPI:**

- FastAPI re-resolves dependencies per request by default — Valkyrja
  pre-resolves at bootstrap
- FastAPI has no disk cache of compiled routes — Valkyrja caches to disk,
  skipping bootstrap on cache hit
- FastAPI assumes long-lived workers — Valkyrja supports both CGI and worker
  equally
- Serverless/lambda cold start is where Valkyrja's cache approach has the
  clearest edge

---

## TypeScript / Node.js

**Status:** Planned.

**Key characteristics:**

- Compiled via `tsc` but type information is erased at runtime
- No reliable decorators (experimental, stage 3) — explicit registration only
- `Function.prototype.toString()` unreliable after any build step
- TypeScript compiler API provides full AST and type information pre-compile
- Node.js worker model — single bootstrap, routes in memory

**Language-specific notes:**

- Single root: `Error` — all three exception branches (Throwable,
  RuntimeException, InvalidArgumentException) extend `Error`
- `abstract class` prevents instantiation at compile time — same guarantee as
  other ports, different mechanism
- No typed throws on function signatures — hierarchy is enforced structurally,
  not by the compiler
- Result pattern available as additive opt-in layer (`tryMake<T>` style) — not
  required
- String constants required for container bindings — no `.class` / `::class`
  equivalent
- TypeScript compiler API used by build tool for pre-compile AST extraction and
  cache data file generation
- Explicit route registration only — no annotated class scanning

**Worker mode:** Node.js stays running, single bootstrap, routes in memory
permanently

**Frontend note:** Core container and DI designed to be isomorphic (runnable in
both Node and browser) — opens door to frontend use without requiring it.
Positioning stays: backend framework that happens to work anywhere TypeScript
runs.

---

## Language Comparison Summary

|            | Concurrency               | Annotations            | `class` ref       | Works without cache | Worker mode |
|------------|---------------------------|------------------------|-------------------|---------------------|-------------|
| PHP        | FPM / FrankenPHP / Swoole | ✅ Attributes           | ✅ `::class`       | ✅ always            | ✅ yes       |
| Java       | Virtual threads (Loom)    | ✅ Annotations          | ✅ `.class`        | ✅ always            | ✅ yes       |
| Go         | Goroutines (native)       | ❌ none                 | ❌ string const    | ✅ always            | ✅ always    |
| Python     | asyncio / ASGI            | ✅ Decorators (runtime) | ✅ class ref       | ✅ always            | ✅ yes       |
| TypeScript | Node.js event loop        | ⚠️ experimental        | ✅ constructor ref | ✅ always            | ✅ yes       |

---

## Discussion Summary

The port list was arrived at by evaluating developer audience size,
architectural fit, and the framework's ability to express its core concepts
idiomatically in each language. The original list considered C#, Scala, Play,
C++, and Ruby alongside the five chosen.

C++ was dropped — no meaningful web framework audience. Play was identified as a
Scala/Java framework, not a language. C# and Kotlin were retained as future
ports (C# for enterprise/Azure, Kotlin nearly free from Java). Rust and Ruby
were acknowledged as valid but lower priority.

The key insight driving the final list: Go and Java are being developed
concurrently with Java taking priority given further progress. Python was chosen
over C# primarily due to audience size and a clear architectural differentiation
story against FastAPI — Valkyrja's compile-once cached bootstrap model offers
something FastAPI explicitly doesn't. TypeScript was identified as missing from
the original list and added due to the prevalence of Vue.js and React.js
ecosystems and NestJS being the closest philosophical equivalent.

The cross-port legibility goal — that a developer familiar with one Valkyrja
port can immediately read another — was established as a north star that guides
every language-specific architectural decision.
