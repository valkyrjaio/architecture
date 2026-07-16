# AGENTS.md — Go (Layer 2)

Per-language guide for the **Go** Valkyrja repos. Read the cross-language
canonical first: [`../AGENTS.md`](../AGENTS.md). This file records the Go
**deltas**. PHP is the reference implementation; mirror its behavior, adapting to
Go idiom. Authoritative port detail: [`README.md`](README.md),
[`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md).

---

## Layout & naming

- **New repos** are scaffolded from the language's `template` repo
  (`project-template-go`, in progress) — the source of truth for repo/file/class
  structure (canonical rule: §3.9).
- **Module path:** `github.com/valkyrjaio/valkyrja-go`. Components map to
  lowercase packages (`container`, `http`, `cli`, `event`, `application`,
  `dispatcher`), with a `contract` sub-package for interfaces and a `data`
  sub-package for data structs.
- **Contracts:** Go `interface` types, structural (no `implements`). The
  **name keeps the `*Contract` suffix** (`ContainerContract`, `RouterContract`,
  `ServiceProviderContract`) but the type lives in a `contract` package rather
  than a suffixed namespace.
- **Providers:** exported structs with receiver methods implementing the provider
  contracts; the cache-optional design calls those interface methods directly on
  the provider structs at bootstrap.
- **Binding keys:** string constants (no `::class` equivalent), format
  `io.valkyrja.{component}.{Name}`, in `const` blocks.
- Every file starts with the license header.

### Errors (not exceptions)

Go has no exceptions — errors are values. The throwable naming parity is kept on
the error types (`ValkyrjaRuntimeException`, `ValkyrjaInvalidArgumentException` as
exported structs) with an unexported marker interface (`valkyrjaThrowable`
embedding `error`) standing in for the abstract base. Return errors; do not
`panic` for normal control flow. Detail: [`../THROWABLES.md`](../THROWABLES.md).

---

## Structure taxonomy

The cross-language taxonomy ([`../AGENTS.md`](../AGENTS.md) §4) applies loosely —
Go's model diverges most of the five. Segments are **lowercase** packages:
`contract`, `provider`, `data`, `factory`, `constant`, `security`, `command`.

Go nuances:

- **Reserved words can't be package names.** `type`, `const`, `func`, `map`,
  `range`, `interface`, `return` are Go keywords — never use them as a package
  segment (so the `Type\` segment is spelled differently; confirm the port's
  choice in [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md)).
- **No abstract classes, enums, or traits.** Enums are `const` blocks; shared
  behavior is struct embedding; there are no `abstract`/`enum`/`trait` segments.
- **Contracts by structure, not suffix-in-namespace** — interfaces live in
  `contract` packages and keep the `*Contract` name.
- **No architecture linter beyond `go-cleanarch`** (bundled in golangci-lint) —
  enforce the taxonomy in review.

---

## Tests

- **Layout:** Go convention — `*_test.go` files co-located with source, in a
  `package <name>_test` (external) or same-package (internal) test package.
  Reusable doubles live in a `fixtures` package mirroring the source tree.
- **Framework:** built-in `go test`.
- **Coverage:** `go test -coverprofile` — **100% (statement and branch), never
  dropping**; every code branch has a test. (Go reports statement coverage
  natively; treat untested branches as gaps.)

---

## Build & CI tools

- **Build tool (`sindri-go`):** uses `go/packages` + `go/ast` + `go/analysis` to
  walk the provider tree and generate the four cache data structs; triggered via
  `go generate`. Dev-only; the framework has zero AST deps.
- **CI (`golangci-lint` — one meta-linter for everything except tests):** bundles
  staticcheck, `go vet`, errcheck, gosec (security), revive, go-cleanarch
  (architecture), unused (dead code), gofmt/goimports (formatting).

### CI gate (run before done)

**Every check green, all tests pass, coverage 100%.** Run the full gate:
`gofmt`/`goimports` (clean) → `golangci-lint run` → `go test -cover ./...`.

---

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md), and the Go section of [`../CI_TOOLS.md`](../CI_TOOLS.md).
