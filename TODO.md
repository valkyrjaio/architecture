# Valkyrja

## Blog (Wordpress dupe, or Drupal Dupe)

- Similar to Data cache classes, this creates classes in the admin that hold
  custom page data, etc. Everything loads from a data cache file instead of from
  database. So much faster. Maybe blog posts can live in database because so
  many, so there should be a switch because at some point it'll be too much for
  the data file to handle

## Middleware settlement-stage naming — align across entry verticals

The queue design (`QUEUE.md`) settled a clearer convention for the two
settlement middleware stages; propagate it to the other verticals. The stages
form a matched before/after pair sharing one object:

- **Stage 6 — pre-action gerund ("about to do the irreversible thing"):**
  Http/gRPC `SendingResponse`, Queue `SettlingResult`, Cli `ProcessExiting`.
- **Stage 7 — post-action past participle ("it happened"):**
  Http/gRPC `ResponseSent`, Queue `ResultSettled`. **Cli has no stage 7** — the
  process is dead after exit, so nothing can run afterward; `ProcessExiting` is
  its stage-6 equivalent, not a terminal stage.

This also regularizes the framework's dominant Object+PastParticiple stage
convention (`RequestReceived`, `RouteMatched`, `ThrowableCaught`, …); the old
`Terminated`/`Exited` were the objectless outliers.

**Each middleware method matches its stage type name** (e.g. `ResponseSent` →
`responseSent()`), so one class can implement multiple middleware stages without
method collisions — so the method renames alongside the type.

Pending work (docs + the terminal middleware stage classes **and their methods**
in every port — PHP, Java, TypeScript, Go, Python):

- [ ] gRPC/Http: rename stage 7 type `Terminated` → `ResponseSent` (+ method) in
      `GRPC.md`, `GRPC_IMPLEMENTATION.md`, and each port's stage class. Keep
      `SendingResponse` for stage 6.
- [ ] Cli: rename `Exited` → `ProcessExiting` (+ method); it is the stage-6
      equivalent (there is no post-exit stage — the process is dead).
- [ ] Once gRPC is renamed, update the gRPC `Terminated` reference in `QUEUE.md`
      (the `Deferred` bridge host-hook list) to the new name.

## Framework tests — request & command mapping fidelity

Add framework-level tests (e2e / smoke / regression / integration — whatever
fits) in **every port** (PHP is the reference, then Java, TypeScript, Go,
Python) proving that an incoming request/command maps faithfully onto the
framework's own objects, independent of routing:

- **HTTP:** an incoming request's method (all types — GET, POST, PUT, PATCH,
  DELETE, HEAD, OPTIONS, …), headers, query params, and body — plus a response's
  status code, reason phrase, headers, and body — map correctly onto
  `ServerRequest` / `Response`. Assert message round-trip fidelity, not merely
  that a route matched.
- **CLI:** an incoming command/input maps correctly — command name, arguments,
  options, and every input shape — onto the `Input` / command objects.

**Why:** a mapping defect (a dropped header, a wrong reason phrase, a mis-parsed
option) is invisible to route-matching tests. The Java and TypeScript
starter-app entry/worker end-to-end work surfaced exactly this class of bug —
caught far from the source. Framework-level mapping tests shift that coverage
left: sooner, and closer to where the mapping actually lives.

Per-port progress:

- [x] **PHP** (reference) — [valkyrja-php#932](https://github.com/valkyrjaio/valkyrja-php/pull/932).
      Three functional test classes under `tests/Tests/Functional/`:
      `Http/Message/Request/RequestMappingTest`,
      `Http/Message/Response/ResponseMappingTest`, and
      `Cli/Interaction/Input/InputMappingTest`. Mirror those names in the
      remaining ports.
- [ ] **Java**
- [ ] **TypeScript**
- [ ] **Go**
- [ ] **Python**

The PHP pass pinned several current behaviors that look like defects rather than
intent; assert them deliberately when porting, and raise them separately rather
than silently diverging:

- `InputFactory` treats `--` and a bare `-` as errors, not as an
  end-of-options terminator.
- `--opt value` does not attach the value to the option — the value lands as a
  positional argument, so only `--opt=value` carries one.
- An option spelled before the command name consumes that slot, so the default
  command name stands and the later bare token becomes a positional argument.
- `--opt=a=b` truncates the value to `a`.
- The query and parsed-body param collections accept non-string scalars when
  built from an array, but their narrowed string return type then raises an
  error on read.
