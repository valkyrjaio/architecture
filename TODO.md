# Valkyrja

## Blog (Wordpress dupe, or Drupal Dupe)

- Similar to Data cache classes, this creates classes in the admin that hold
  custom page data, etc. Everything loads from a data cache file instead of from
  database. So much faster. Maybe blog posts can live in database because so
  many, so there should be a switch because at some point it'll be too much for
  the data file to handle

## Middleware settlement-stage naming — align across entry verticals

Every implemented port renamed the stage classes and their methods. Two documents
still name stage 7 `Terminated`.

- [valkyrjaio/architecture#129](https://github.com/valkyrjaio/architecture/issues/129)

## README alignment and audit — every port

A report about the PHP Container README started this work. The README named a
static `make()` factory as the recommended way to register a service. The
framework does not work that way. A service provider registers a service, and the
service class carries no registration code.

The audit that followed found four groups of statements that contradict the
implemented ports. Four PRs corrected them:

- valkyrjaio/valkyrja-php#966 — the Container README.
- valkyrjaio/valkyrja-php#967 — the Application and Event READMEs.
- valkyrjaio/valkyrja-php#968 — a request accessor that does not exist.
- valkyrjaio/architecture#108 — 22 files here.

The audit only covered PHP and this repository. **Do not assume the other ports
are correct.** The same drift is likely, because the same documents seeded them.

### Two checks find this drift

Run both checks per repository. Each check compares a document against the
source, not against another document.

1. **Undefined type check.** Collect every type the READMEs name through
   `::class`, `new X(`, `implements`, and `extends`. Report any type that the
   source does not define. Exclude the types a README defines inline for
   illustration.
2. **Undefined method check.** Collect every `->method()` and `::method()` call
   in the READMEs. Report any method name that the source defines nowhere. This
   check found `make()`, and it found `getParsedBodyParam()`.

Both checks report app-side names from examples, such as `PostController`. Treat
those names as expected results, not as defects.

### What the PHP audit found

Record these results, because each one is a pattern to look for in the other
ports:

- A documented container API that exists in no port: 61 call sites used `make()`
  or `singleton()`. Both PHP and Java expose the same 16-method contract, and
  neither declares those two methods.
- A build tool named Forge in the documents and Sindri in the code: about 50
  references.
- Provider methods documented as `static` that the source declares as instance
  methods: 60 declarations, plus 29 provider lists that held class-name strings
  where the source holds instances.
- Two code samples that cannot run: a call to `TeamsNotifier::make()` on a class
  with no `make()` method, and `static Map<...> publishers();` in a Java
  interface, which does not compile.

### Pending work

- [ ] Java — run both checks. Java has **no module READMEs at all** today, so
      the result is a documentation gap, not drift. Decide whether Java mirrors
      the PHP module README set.
- [ ] TypeScript — run both checks against the TypeScript port.
- [ ] Go — run both checks against the Go port.
- [ ] Python — run both checks against the Python port.
- [ ] PHP — re-run both checks after the ports land, to catch new drift.
- [ ] Add the two checks to CI, so a README that names a missing type or method
      fails the build. A one-time sweep does not hold the documents in line.

The audit of this repository's own documents is tracked separately in
valkyrjaio/architecture#111.
