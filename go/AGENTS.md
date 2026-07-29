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

Go has no exceptions — errors are values, and Go uses the **`*Error` suffix**
throughout (not `*Exception`, which is foreign to Go). The error types are
exported structs — `ValkyrjaRuntimeError`, `ValkyrjaInvalidArgumentError`,
`ContainerNotFoundError`, etc. — with an unexported marker interface
(`valkyrjaThrowable` embedding `error`) standing in for the abstract base and
unexported categoricals like `containerRuntimeError`. Return errors; do not
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
  Reusable doubles live in a `fixtures` package mirroring the source tree, named
  with the `Fixture` suffix.
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

### Pinned tool modules: upgrade with `go get tool`, never `go get -u tool`

golangci-lint is pinned in an isolated tool module (`.github/ci/lint/go.mod`, the
analog of PHP's per-tool `.github/ci/<tool>/composer.json`) and run via
`go tool -modfile=…`, so its dependency graph never mixes with the framework's.
That module is upgraded with **`go get tool`** — the `tool` meta-pattern with no
`-u`.

- **`-u` upgrades the tool's whole transitive graph, not just the tool.** MVS is
  then free to select a dependency the tool's authors never built against, and
  the pinned tool stops compiling. Without `-u`, each declared tool still moves
  to its latest release, but its dependencies are resolved from *that release's*
  own `go.mod` — which is exactly what pinning a tool means.
- **A same-path breaking release is invisible to the toolchain.** Go's import
  compatibility rule is a convention, not an enforced constraint: `go-header`
  v1.0.0 rewrote its API (`Configuration`, `WithTemplate`, `WithValues`, `Target`
  removed; `New` takes `*Settings`) while keeping the same module path, with no
  `/v2` to signal it. Nothing stopped `-u` from pulling it into a golangci-lint
  v2.12.2 that requires v0.5.0, and the lint job died with compile errors before
  running a single linter.
- **The tell: compile errors inside `$GOMODCACHE`, not lint findings.** Errors
  pointing at `…/golangci-lint/v2@vX/pkg/golinters/<name>/…` mean an indirect
  dependency outran the tool. That is not a problem with our code, and there is
  no local lever to fix it — we import no linter's packages, so there is no call
  site of ours to port. Pin the offending dependency back to what the tool
  requires and wait for an upstream release that adopts the new API;
  `go get tool` then picks it up on its own.
- **Non-tool modules keep `go get -u ./...`** — there the module being upgraded
  *is* the code under test, so upgrading its graph is the point.

The reusable workflows enforce this
([`_go-update-dependencies.yml`](https://github.com/valkyrjaio/.github/blob/master/.github/workflows/_go-update-dependencies.yml)).
`_go-check-outdated-dependencies.yml` deliberately performs the very same upgrade
so the two can never disagree — change them in lockstep, or the check reports
every tool module as perpetually outdated for an upgrade the updater will not
make.

---

## Go-specific notes

- **Dynamic route regexes: `(?P<name>…)`, anchored, no delimiters, and RE2's
  limits.** Decided ahead of the HTTP routing port so it is built right the first
  time. Go's `regexp` is **RE2**, not PCRE, which constrains the shape three ways:
  - **Named groups are `(?P<name>…)`.** Go 1.22+ also accepts the `(?<name>…)`
    spelling Java and TypeScript emit, but `(?P<name>…)` is the portable form,
    compiles on every toolchain, and matches Python's only option (see
    [`../python/AGENTS.md`](../python/AGENTS.md)). Read parameters back with
    `re.SubexpIndex(name)` against `FindStringSubmatch`.
  - **No delimiters, and the anchors are load-bearing.** `regexp.Compile` takes a
    bare pattern; PHP's `/^…$/` compiles but matches nothing, since the slashes
    are literal (PHP needs them because `preg_match` requires them — see
    [`../php/AGENTS.md`](../php/AGENTS.md)). Unlike Java's `Matcher.matches()`,
    which implies a full match on its own, Go's `MatchString` *searches* — so the
    `^` / `$` framing is what makes a route match exactly, not decoration.
    `\/` compiles fine, so `Regex.PATH` carries over unchanged.
  - **No lookahead, lookbehind, or backreferences.** RE2 rejects them outright
    (`(?=` fails with "invalid or unsupported Perl syntax"). Any route-regex
    feature built on them in PHP is unportable to Go, so keep parameter regexes to
    plain character classes, quantifiers, and alternation.

---

More: [`README.md`](README.md), [`PROVIDER_CONTRACTS.md`](PROVIDER_CONTRACTS.md),
[`TODO.md`](TODO.md), and the Go section of [`../CI_TOOLS.md`](../CI_TOOLS.md).
