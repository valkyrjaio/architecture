# Adding a New Language Port

A practical, end-to-end playbook for bringing Valkyrja to a new language —
covering everything from studying the existing ports, to the architecture
documents, to the `template` repo, to the reusable CI workflows and rulesets in
`.github`, to the framework/build-tool repos themselves.

> **How to read this.** The body (§0–§8) is written to be **language-agnostic** —
> it describes the process and the invariants that hold for *any* port. Concrete,
> language-specific quirks live in the **Findings log** (§9). Wherever the body
> says "see Findings," expect that a real language hit a wrinkle there and yours
> might too. When you finish a port, append your own findings — this is a living
> document.

---

## 0. Mental model — what actually gets created

A language port is not one repo; it is a **set of repos plus org-level wiring**.
Before touching anything, understand the moving parts:

| Piece | Where it lives | Purpose |
|-------|----------------|---------|
| **Architecture docs** | `architecture/<lang>/` | Layer-2 decisions, provider contracts, agent guide, TODO |
| **Template repo** | `valkyrjaio/project-template-<lang>` | The scaffold every new repo of that language is cloned from |
| **Reusable workflows** | `valkyrjaio/.github/.github/workflows/_*-<lang>.yml` | The CI + release machinery every repo calls |
| **Ruleset** | `valkyrjaio/.github/rulesets/<lang>/` | Required-status-check branch protection |
| **Framework repo** | `valkyrjaio/valkyrja-<lang>` | The runtime (zero AST/build deps) |
| **Build tool repo** | `valkyrjaio/sindri-<lang>` (naming varies) | Dev-only cache generator (AST / compiler API) |
| **Entry adapters** | `entry/*` in the framework repo | Server adapters for the language's ecosystem |
| **Org config** | `SUPPORTED_LANGUAGES` var, org secrets | Publishing credentials, language enablement |

**PHP is the reference implementation.** When any port disagrees with PHP on
structure, naming, or tests, PHP wins unless an architecture doc says otherwise.
For a *new* language, first pick the **closest existing port** to crib from — a
statically-typed/compiled language mirrors the compiled ports; a
dynamic/interpreted one mirrors the interpreted ports. The `template` repos of all
languages are near-identical in shape, so **diffing two of them is the fastest way
to learn the skeleton**.

---

## 1. Phase 0 — Study and decide (before writing any code)

Read these in the architecture repo, in order: `README.md`,
[`AGENTS.md`](AGENTS.md) (cross-language canon), [`PORTS.md`](PORTS.md),
[`THROWABLES.md`](THROWABLES.md), [`CONTAINER_BINDINGS.md`](CONTAINER_BINDINGS.md),
[`DISPATCH.md`](DISPATCH.md), [`DATA_CACHE.md`](DATA_CACHE.md),
[`BUILD_TOOL.md`](BUILD_TOOL.md), [`STATIC_METHODS.md`](STATIC_METHODS.md),
[`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md), then the closest language's
`<lang>/` folder.

Then **decide and write down** (these become the Layer-2 docs in Phase 1):

- **Package root / namespace** and the **casing convention** for the whole tree.
- **How code is shared across repos.** Every component repo contributes into one
  shared namespace; decide the language's mechanism for that (module/package
  system, namespace declaration, etc.). Get it right from repo #1 or you will
  refactor every repo later. *(Some ecosystems have a subtle rule here — see
  Findings.)*
- **Contracts mechanism** (native interface vs an abstract-base equivalent). Keep
  the `*Contract` name suffix regardless.
- **Binding keys — string constants vs class/type references.** Decided by the
  language's *import semantics*: **any language where merely naming a class forces
  its module to load must use string-constant keys**; only languages with a
  compile-time constant reference (no load) may use references. Get this wrong and
  the cold-start / cache design breaks.
- **Throwable hierarchy → native roots.** Map the abstract branches onto the
  language's native exception/error roots. Keep the parity name suffix — *though
  the suffix word itself can differ per language (see Findings).*
- **Structure-taxonomy segment spelling.** The taxonomy (`Contract`, `Provider`,
  `Factory`, `Constant`, `Exception`/error, `Throwable`, `Abstract`, `Enum`,
  `Type`, …) is defined in [`AGENTS.md`](AGENTS.md) §4. Spell each segment in the
  language's idiom, and **escape any segment that collides with a reserved word or
  a standard-library name** (a trailing underscore is the established convention).
  Drop segments for constructs the language lacks.
- **Handler/attribute mechanism** — how the handler marker is expressed as **inert
  metadata**, never a self-registrar.
- **Build tool AST strategy** — which parser/analyzer the build tool uses.
- **Deployment model / entry adapters** — the server story and whether a
  CGI/serverless mode exists.

> ⚠️ **Verify any "load-bearing" language feature actually exists and is stable
> today.** A port was once designed around an *assumed-upcoming* language feature
> that was later withdrawn and never shipped, forcing a rewrite of multiple docs
> and the whole cold-start design. Before you lean on a feature, confirm it's real
> and shipped; prefer designs that stay correct *without* the optimistic feature.
> (See Findings for the specific case.)

---

## 2. Phase 1 — Architecture docs (`architecture/<lang>/`)

These are **Layer-2** guides — the per-language deltas from the cross-language
canon. Create the folder with:

- **`README.md`** — port implementation notes: key language decisions, the
  component port order, deployment models.
- **`PROVIDER_CONTRACTS.md`** — the provider contract interfaces with real code
  examples (component/service/route/listener providers, handler markers, data
  classes).
- **`AGENTS.md`** — the Layer-2 agent guide. Follow the sibling structure exactly:
  *Layout & naming* → *Exceptions* → *Structure taxonomy* (segment spelling +
  nuances) → *Tests* (framework + PHPUnit→target mapping) → *Build & CI tools*
  (tool list, isolation, run commands) → *CI gate* → *language-specific notes*.
- **`TODO.md`** — the port checklist.

Then **update the shared docs** so the language is discoverable: add the language
row + doc links to the top-level `README.md` and [`AGENTS.md`](AGENTS.md) tables,
and add its characteristics to [`PORTS.md`](PORTS.md).

> **Keep the docs internally consistent.** Names drift over time and per-language
> docs fall behind. When you add a language, reconcile stale names and broken
> doc-links across *all* of that language's files, and make sure the
> cross-language canon and the Layer-2 doc agree — **the canon wins; fix the
> Layer-2 doc.**

---

## 3. Phase 2 — The `template` repo (`project-template-<lang>`)

This is the first concrete artifact and the source of truth for repo/file/class
structure. Clone a sibling `template` and translate it.

**Top-level metadata** (near-identical across languages): `AGENTS.md` (a thin
pointer to the two canonical guides), `README.md`, `VERSION.md` (`vNN.0.0`),
`CHANGELOG.md`, `LICENSE.md`, `.editorconfig`, `.gitattributes`, `.gitignore`,
plus the language's version pin.

**Root manifest as a facade.** The root build/package file exposes a **shortcut
per CI tool** that delegates into that tool's directory. The mechanism is
whatever the ecosystem offers — a scripts block, a dedicated task runner, or a
`Makefile`. **The reusable workflows call these shortcuts by name, so the names
are a contract.**

**Per-tool CI isolation under `.github/ci/<tool>/`.** The cornerstone convention:
**each tool gets its own dependency manifest and committed lockfile, with any
installed dependencies gitignored** — so tools never share dependency versions.
Some ecosystems have no per-repo installed-deps directory (a global cache
instead), and some collapse *all* tooling into a single binary so the whole
`.github/ci/` reduces to one directory — see Findings.

**Choose the tool set** to cover these roles, mapping from the reference
implementation's `.github/ci/` and the closest existing port:

- **Lint / format**
- **Type / static analysis**
- **Architecture** (segment/naming enforcement)
- **Security**
- **Tests + coverage**
- **Migration / codemods** (optional)

Not every role has a strong tool in every language — **note any gap explicitly
rather than pretending it's covered.** Conversely, some languages cover several
roles with one tool, collapsing multiple CI jobs into one (see Findings).

**Placeholder source + info file.** A placeholder source namespace with an info
class carrying `VERSION` and `VERSION_BUILD_DATE_TIME` (the release workflow
rewrites these). Use `.gitkeep` to hold otherwise-empty dirs.

**Test skeleton.** Mirror the cross-language test taxonomy (unit / functional /
fixtures / abstract) plus an abstract base test case — *adapted to the language's
idiom (some ecosystems co-locate tests rather than using a parallel tree; see
Findings)*. **If the coverage gate can't pass an empty suite, add one tiny smoke
test** so a freshly-cloned template is green.

**Workflows** (`.github/workflows/`): `ci.yml` (calls the `_*-<lang>.yml`
reusables), `update-dependencies.yml`, `release-new-version.yml`,
`create-version-branch.yml`, `rebase-from-master.yml`, `rebase-to-master.yml`,
`cherry-pick-commits.yml`, `restore-branch-from-backup.yml`. The
language-agnostic reusables already exist and are pinned to a SHA; the
language-specific `_*-<lang>.yml` may not exist yet — reference them and let
pinning follow (see Phase 3 ordering note, and §7 on automatic ref-pinning).

**VALIDATE THE GATE LOCALLY before committing.** Run the full CI gate through the
root facade on the actual template. This is the one part you *can* verify without
GitHub Actions, and it catches the real bugs (lint rules that fight idiomatic
code, empty-suite coverage failures, task-runner env/path wiring). Don't skip it.

---

## 4. Phase 3 — Reusable workflows in `.github`

> **Base branch:** the active mainline of `.github` (and most repos) is **`NN.x`
> (e.g. `26.x`)**, *not* `master` — `master` is frozen/stale. Branch from and PR
> into `NN.x`. Confirm with `git log origin/NN.x` vs `origin/master`.

**Tool workflows** — one `_<tool>-<lang>.yml` per CI tool. Mirror a sibling.
Every one follows the same shape:

1. `workflow_call` with a toolchain-version input, a `ci-directory`, and a
   required `paths` (dorny/paths-filter `ci`/`files` keys).
2. checkout → generate app token (for PR comments) → `dorny/paths-filter` →
   "restore CI code from base if only source changed" → set up the toolchain
   **with caching keyed on the lockfile** → run the tool **via the root facade
   shortcut** → post/clear a PR comment on failure/success.
3. `name: Z Reusable <Tool>` — the **`Z ` prefix is mandatory** (sorts reusables
   to the bottom of the Actions list; a workflow enforces the convention).

**The release chain** — reuse the agnostic building blocks, only port the
language-specific pieces:

- Reuse as-is: `_get-version`, `_get-version-for-release`, `_update-version-files`,
  `_release`.
- Port: `_create-<lang>-release.yml`, `_create-<lang>-version-branch.yml`,
  `_version-branch-<lang>.yml`, `_check-outdated-<lang>-dependencies.yml`,
  `_update-<lang>-info-files.yml`, `_update-<lang>-dependencies.yml`, and (if the
  ecosystem publishes — see below) `_<lang>-release-<registry>-publish.yml`.
- The info-file and version-branch workflows rewrite the version into the info
  file **and**, if the manifest carries a version field, the manifest. *(Some
  ecosystems encode the major version in the module path rather than a manifest
  field — see Findings.)*

**Publishing varies wildly by ecosystem** — plan for one of:

- **Token-based** — store a registry token as an **org secret**; the publish
  workflow reads it. Note the tool's expected env var often differs from the
  secret name.
- **OIDC / trusted publishing** — no stored secret; short-lived credential via
  `id-token: write`.
- **No publish step at all** — some ecosystems serve packages straight from the
  git tag, so there is no publish reusable, no credential, and no publish job.

When a publish workflow exists it has **no caller in `.github`** — it is wired as
a second **`publish` job (`needs: release`)** in each *framework/package* repo's
`release-new-version.yml`, **not** in the template (templates omit it). See the
existing framework repos for the pattern, and Findings for per-ecosystem specifics.

**Concurrency discipline (when multiple agents work in `.github` at once).** Keep
each language branch **additive**: add `_*-<lang>.yml` and `rulesets/<lang>/`. The
real rule for shared files (e.g. `_enforce-repo-settings.yml`): **don't co-edit a
shared file that has *live* concurrent work touching it.** If a parallel port is
still in flight on that file, do your shared-file edit as a separate later PR; if
that port has already merged and never touched the file, bundling the edit into
your PR is fine. When the local clone is shared with another agent, **work in a
dedicated `git worktree`** off `NN.x` so your branch/index can't collide.

**Ordering note.** The template (Phase 2) references reusables that don't exist
until this phase. Either order works: build the template first and accept its
GitHub CI is red until the reusables merge (fine — you validated the gate
locally), or build the reusables first.

---

## 5. Phase 4 — Rulesets and enforcement

**Create `rulesets/<lang>/Required <Lang> PR Checks.json`** by mirroring a
sibling. Key fields:

- `name: "Required <Lang> PR Checks"`, `enforcement: active`, conditions include
  `~DEFAULT_BRANCH` and `refs/heads/??.x`.
- `required_status_checks` contexts use the format **`"<Job Name> / <Job Name>"`**
  — the caller's job `name` and the reusable's job `name` must match. Include one
  context per CI job (however many the language's tooling produces).
- **Omit `id`** for a new ruleset. On apply, tooling strips to
  `{name, target, enforcement, conditions, rules, bypass_actors}` and matches by
  `name`, so `id`/`source` are cosmetic.

**Two wiring points — you need both:**

1. **`_create-repo.yml`** applies `rulesets/<lang>/` to **new** repos
   *automatically*, gated on the repo-name suffix being in the org
   **`SUPPORTED_LANGUAGES`** variable. Confirm the language suffix is in that var.
2. **`_enforce-repo-settings.yml`** applies rulesets to **existing** repos via
   **hardcoded per-language blocks**. Add a `[[ "$REPO_NAME" =~ -<lang>$ ]]` block
   mirroring the existing ones. This is a **shared-file edit** — apply the
   concurrency rule from §4.

**Document the checks for contributors.** `.github/CONTRIBUTING.md` has a
*Running CI Locally* section with a per-language subsection — fill in your
language's (they start as *Coming soon.*) with a table of each CI check and the
local command that runs it via the root facade. This is the human-facing twin of
the ruleset's required status checks; **the two lists must stay in sync** (same
checks, one for machines, one for humans).

The repo-name **suffix** (e.g. `-python`, `-go`) is load-bearing for both
detection points — pick it in Phase 0 and use it consistently in every repo name.

---

## 6. Phase 5 — Framework, build tool, and adapter repos

With the scaffolding in place, port the framework itself into `valkyrja-<lang>`
(and the build tool into `sindri-<lang>`), created from the `template`.

- **Component port order:** Container → Dispatch → Event → Application → CLI →
  HTTP → Bin.
- **Port code and tests together**, mirroring the source repo's test layout and
  mapping the framework (PHPUnit → the target's runner). Target **100% coverage**
  (branch where the language/tooling supports it — see Findings).
- The framework repo is **runtime-only, zero AST/build deps**; all code-gen lives
  in the dev-only build tool.
- If the ecosystem publishes, the framework repo's `release-new-version.yml` gets
  the two-job structure (`release` + `publish`), unlike the template.
- Add `entry/*` server adapters per the deployment decision.

---

## 7. Cross-cutting conventions (apply everywhere)

- **Definition of done:** every code branch tested, all tests pass, the *full* CI
  gate green, coverage 100% (line and branch where supported) and never dropping.
- **Every source file** carries the license header; every file ends with a
  trailing newline; American English throughout.
- **Commits:** `[Component] Imperative description.` (trailing period). For adding
  a whole language's workflows the sibling history uses the **language tag**
  (e.g. `[Go]`); the very first commit of a brand-new repo is
  `[Initial] Initial commit.`.
- **PR titles:** same tag, **no** trailing period. Fill the PR template
  (Description, Types of changes, Changes — bold path — em dash — what changed).
- **Branch targeting:** improvements/fixes → lowest affected `NN.x`; features →
  `master` *in principle*, but note `master` may be frozen and `NN.x` is the live
  line — check. Branch prefixes: `feature/`, `improvement/`, `fix/`, `docs/`.
- **Ask before each write action** — before committing, before pushing, before
  opening a PR (creating a branch needs no prompt).
- **Template workflow-ref pinning is automatic.** Pin the template's `_*-<lang>.yml`
  refs to the current `.github` release SHA as the siblings do; regardless of what
  you pin to, `_update-workflow-refs.yml` rewrites every `valkyrjaio/.github/...@<ref>`
  (any ref, including `@master`) to the latest release SHA and opens a PR in each
  consuming repo on every `.github` release. So you never hand-maintain these.

---

## 8. Master checklist

**Phase 0 — decide**
- [ ] Read the architecture canon + the closest existing language's `<lang>/`
- [ ] Namespace root + casing; cross-repo namespace-sharing mechanism
- [ ] Contracts mechanism; binding-key strategy (string vs reference)
- [ ] Throwable→native-root mapping; taxonomy segment spelling + escapes
- [ ] Handler/metadata mechanism; build-tool AST strategy; deployment model
- [ ] **Verify every "load-bearing" language feature actually exists today**

**Phase 1 — architecture docs**
- [ ] `architecture/<lang>/{README,PROVIDER_CONTRACTS,AGENTS,TODO}.md`
- [ ] Update top-level `README.md`, `AGENTS.md`, `PORTS.md` tables
- [ ] Reconcile stale names / broken links; canon vs Layer-2 agreement

**Phase 2 — template repo**
- [ ] Metadata + editor/git config + version pin
- [ ] Root facade with per-tool shortcut tasks
- [ ] `.github/ci/<tool>/` isolation (manifest + committed lockfile, deps ignored)
- [ ] Placeholder source + info file; test skeleton (+ smoke test if needed)
- [ ] All 8 workflows
- [ ] **Run the full gate locally and make it green**

**Phase 3 — .github reusables** *(branch off `NN.x`; worktree if shared)*
- [ ] `_<tool>-<lang>.yml` per tool (`Z Reusable` names, lockfile-cached)
- [ ] Release chain + `_update-<lang>-dependencies.yml` + publish workflow (if any)
- [ ] Registry credential decided; org secret named (if publishing)
- [ ] Additive branch; shared-file edits per the concurrency rule; PR into `NN.x`

**Phase 4 — rulesets & enforcement**
- [ ] `rulesets/<lang>/Required <Lang> PR Checks.json` (contexts match job names)
- [ ] Language suffix in `SUPPORTED_LANGUAGES`
- [ ] `-<lang>$` block in `_enforce-repo-settings.yml`
- [ ] Fill in the `#### <Language>` section of `.github/CONTRIBUTING.md` (checks + local commands)

**Phase 5 — framework/build-tool/adapters**
- [ ] `valkyrja-<lang>` + `sindri-<lang>` from the template
- [ ] Ports in component order, code+tests together, 100% coverage
- [ ] `publish` job wired into the framework repo's `release-new-version.yml` (if publishing)
- [ ] `entry/*` adapters

---

## 9. Findings log

Concrete, language-specific wrinkles that arose per port. Treat each as a
*possibility to check for* in your language — the analogous issue may or may not
apply. Append a section when you finish a port.

### Python
- **The withdrawn-feature trap (origin of §1's warning).** The port was designed
  around **PEP 690 implicit lazy imports**, which was **withdrawn and never
  shipped**. The container/cold-start design was reframed around string-constant
  keys + lambda-wrapped values (correct in any Python version); PEP 810 is the
  unshipped successor. Rewrote `README.md`, `PROVIDER_CONTRACTS.md`, `AGENTS.md`.
- **PEP 420 namespace package:** **no `__init__.py`** at the namespace root
  (`src/valkyrja/`), so every repo shares the `valkyrja` namespace. Verify the
  built wheel resolves it.
- **Casing:** idiomatic Python is lowercase (`valkyrja/template/`), unlike the
  PascalCase (`Template/`) of the compiled/TS ports. Don't fight the linters.
- **Facade needs an extra tool:** `uv` matches the per-tool isolation 1:1 but has
  no scripts block, so **poethepoet** (`[tool.poe.tasks]`) provides the root
  facade. `uv run -q` silences the nested-venv warning.
- **Empty-suite coverage:** pytest can't pass under `--cov-fail-under=100` with no
  tests → add a smoke test to keep a fresh template green.
- **Segment collisions:** `enum`/`type`/`id` shadow stdlib/builtins (Go hit `type`
  too). Stay on absolute imports; be aware.
- **Enforcement gap:** `rulesets/python/` needed a `-python$` block in
  `_enforce-repo-settings.yml`, shipped as a **separate follow-up PR** because the
  concurrent Go work was still live on shared files at the time.
- **Template refs left at `@master`** initially (the reusables didn't exist yet);
  the other ports instead pin the current `.github` SHA. Either way
  `_update-workflow-refs.yml` repins on release.

### Go
- **Semantic Import Versioning (SIV) is the dominant Go-only concern.** Any major
  ≥ 2 must live in the module path (`.../valkyrja-go/v26`), so the annual `NN.0.0`
  scheme forces a `/vNN` suffix in `go.mod` and every internal import; a major bump
  rewrites the path, not a manifest field. (The PoC `go.mod` lacked `/v26` — a `v26`
  tag on it is invalid; catch this early.)
- **No publish workflow.** The module proxy serves from the git tag — releasing is
  tag-only, so there is no publish reusable, no registry credential/OIDC, and no
  `publish` job in the framework repo's `release-new-version.yml`.
- **One tool covers five roles.** golangci-lint bundles lint, format
  (gofmt/goimports), static analysis (staticcheck/govet), security (gosec),
  architecture (go-cleanarch), and dead-code (unused); `go test` does
  tests+coverage. So `.github/ci/` is a single `.github/ci/lint/` module and CI is
  2 jobs (`golangci-lint`, `Test`) — hence 2 ruleset contexts, not 4–7.
- **Tool pinned via a `tool` directive (Go 1.24+), not a vendored dir.**
  golangci-lint lives in `.github/ci/lint/go.mod` + committed `go.sum`, run from the
  repo root as `go tool -modfile=.github/ci/lint/go.mod golangci-lint`. Root facade
  is a **`Makefile`** (`make ci|lint|fmt|test|coverage|tidy-check`). No per-repo
  installed-deps dir to gitignore (global module cache).
- **Idiomatic-Go naming fights the linters — disable, don't rename.** The
  cross-language `GetX`/`SetX` accessors and `contract.FooContract` package stutter
  trip revive's `get-return`/`exported` rules; disable those two rules in
  `.golangci.yml` (with a comment on why) rather than break parity. Strategy:
  enable-almost-all, disable conflicts. In golangci-lint v2, formatters (`gofmt`,
  `gci`, …) are a separate section — listing one under `linters` errors.
- **No `src/`, co-located tests.** Packages sit at the repo root (lowercase, like
  Python but flatter); tests are `*_test.go` beside the source, not a parallel
  `tests/` tree; reusable doubles are `*Fixture` in a `fixtures` package.
- **`*Error`, not `*Exception`.** Go is the only port using the `*Error` suffix
  (`THROWABLES.md`); the first-draft Go docs used `*Exception` and had to be swept.
- **Coverage is statement-level, not branch**, and `go test -cover` passes an
  all-covered/empty package fine (no `--cov-fail-under` equivalent) — treat untested
  branches as manual-review gaps.
- **Segment `type` is a Go keyword** → `type_`, mirroring Java's `abstract_`/`enum_`.
- **`-go` suffix worked identically** in `_create-repo.yml` and
  `_enforce-repo-settings.yml`, and the `-go$` enforce block shipped in the **same**
  PR as the reusables/ruleset — safe because the concurrent Python work had already
  merged and never touched that shared file (this is the origin of §4's refined
  concurrency rule).

### <next language>
- _(to be added)_
