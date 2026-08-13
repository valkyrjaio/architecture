# Commit & PR Title Convention

Every change is described twice — once in the commits on your branch, and once in
the pull request title. Those two strings are **not** the same, and the
difference is deliberate.

Merges are squash-only, and every repo pins the squash commit to the pull
request:

```
allow_squash_merge=true
allow_merge_commit=false
allow_rebase_merge=false
squash_merge_commit_title="PR_TITLE"
squash_merge_commit_message="PR_BODY"
```

So your branch commits are working notes that reviewers read and history
discards. The **PR title becomes the permanent subject line** and the PR
description becomes the permanent body. Write each for the job it actually does.

## The shape

```
[Root] type: Message.
[Root] type(#123): Message for an issue.
[Root] type!: Message for a breaking change.
[Root] type(#123)!: Message for a breaking change with an issue.
```

A **root** says what the change is about. A **type** says what kind of change it
is. Neither may restate the other.

|                                       | Ends with | Issue reference          |
| ------------------------------------- | --------- | ------------------------ |
| **Working-branch commit**             | a period  | permitted, not required  |
| **PR title**                          | no period | required when one exists |
| **Direct push to a protected branch** | no period | —                        |

A working-branch commit is a **ledger entry** — a complete sentence recording what
that commit did, so it takes a period. Anything that becomes a **permanent subject
line** is a title rather than a sentence, and takes no period: the squashed PR
title, and any commit pushed straight onto a protected branch, such as the release
run's own bookkeeping.

The distinction is between working notes and the permanent record, not between
commits and PRs.

GitHub appends the PR number to the squashed subject on its own:

```
commits    [Cli] fix: Remove the always-true allowed responses check.
           [Cli] test: Cover the bare double dash operand case.
PR title   [Cli] fix(#123): Remove the always-true allowed responses check
merged     [Cli] fix(#123): Remove the always-true allowed responses check (#456)
```

Two numbers appear on that final line and they mean different things: the one in
parentheses **before** the colon is the issue, the one **at the end** is the pull
request. Position tells them apart.

## Types

| Type        | Use for                                                 |
| ----------- | ------------------------------------------------------- |
| `feat`      | A new capability or an addition to the public API       |
| `fix`       | Corrects behavior that was broken                       |
| `deprecate` | Marks API as deprecated without removing it yet         |
| `docs`      | Documentation only                                      |
| `test`      | Adds or changes tests only                              |
| `refactor`  | Internal restructuring with no behavior change          |
| `perf`      | Performance improvement with no behavior change         |
| `style`     | Formatting only — whitespace, import order, no behavior |
| `build`     | Build scripts, dependency manifests, packaging          |
| `ci`        | CI workflows, tooling configuration, automated runs     |
| `chore`     | Routine maintenance that fits nothing else              |
| `revert`    | Reverts an earlier change                               |

These are the [Conventional Commits](https://www.conventionalcommits.org/) types
plus `deprecate`, which that set has no word for in this scheme. `perf` and `style`
will be rare; they cost nothing, and keeping the standard set intact avoids arguing
about where a change belongs.

There is no type marking a change as automated. Git already records the author, and
the generated release notes already attribute the bot by name — so a commit keeps
the type its change actually deserves, and _who ran it_ comes from
`git log --author`, which is exact rather than a convention someone must maintain.

The type is chosen **relative to the repo you are in**. A new reusable workflow in
the `.github` repo is that repo's product, so it is a `feat` there, while a
workflow change in a framework repo is `ci`.

`feat`, `deprecate`, and anything marked breaking drive the middle version
component; everything else is a patch. Commits under a release-version root never
trigger a release at all. See [`VERSIONING.md`](VERSIONING.md).

## Breaking changes

Append `!` immediately before the colon — after the issue reference if there is
one:

```
[Http] feat!: Drop the deprecated ServerRequest::withoutAttributes method.
[Http] feat(#123)!: Drop the deprecated ServerRequest::withoutAttributes method
```

`!` is **required** on any change that breaks a public contract. It does not
change which version component is bumped — breaking and `feat` both move the
middle one — and that is exactly why it is required: since the version number
cannot distinguish a break from a feature, `!` is the only marker that does.

## Issue references

`(#123)` goes in the **PR title**, and is omitted entirely when no issue tracks
the work. It is **permitted but not required** on commits: the first commit on a
branch is often the one whose message becomes the PR title, so writing it there is
convenient, but repeating it across every commit adds nothing since the issue is
already in the branch name.

To actually close the issue on merge, put a closing keyword in the **PR
description** — `Closes #123`. The description becomes the squash commit body, so
that is both what GitHub acts on and where the link durably lives.

## Roots

A root names **what the change is about**. Two rules decide what may appear in the
brackets, and everything else follows from them.

### Rule 1 — a root names a thing, never a kind of change or an actor

`[Documentation]`, `[Tests]`, `[Refactor]`, and `[Merge]` all named a kind of
change; the type slot carries those now. `[CI]` and `[Volundr]` name whatever
_performed_ the change — but git already records the author, so the root's one slot
is better spent on what nothing else records.

### Rule 2 — a root is never the repo's own identity

Inside the `phpcsfixer` repo, `[PhpCsFixer]` says nothing: every commit there is
about phpcsfixer. Inside the framework, `[Valkyrja]` says the same nothing. Use a
root that names the thing being worked on within that repo — `[Rule]`, `[Ruleset]`,
whatever fits.

The same name is perfectly good **elsewhere**, because outside its own repo it
carries real information:

```
in ci-phpcsfixer-php     [Ruleset] feat: Add the SeparateMultiUseImports rule.
in a framework repo      [PhpCsFixer] build: Update to v26.1.1.
in a framework repo      [PhpCsFixer] ci: Run linting across the source tree.
```

The clearest case is one artifact taking different roots in different repos. A pull
request template inside `valkyrja/.github` cannot be `[GitHub]` — that repo _is_ the
org's GitHub configuration, so the name says nothing there and `[Template]` is the
root. The identical file in a framework repo is `[GitHub]`, because there it is
precisely the GitHub-specific thing that stands out:

```
in valkyrja/.github      [Template] docs: Update the pull request template.
in a framework repo      [GitHub] docs: Add a PR template superseding the org-wide one.
```

This rule is **positional, not permanent**. `Http` and `Cli` are roots in the
framework today because they live there. Were either split into its own repo, it
would retire as a root _within that repo_ while remaining a good root anywhere that
integrates with it.

### The vocabulary is open

There is no fixed list of roots. A root may be anything that sufficiently describes
the thing being worked on, subject to the two rules above. What follows are
examples of the kinds that come up, not an enumeration:

| Kind                | Examples                                                                    |
| ------------------- | --------------------------------------------------------------------------- |
| **Module**          | `[Http]`, `[Cli]`, `[Container]`, `[Orm]`, `[Api]`                          |
| **Concept**         | `[Provider]`, `[Throwable]`, `[Middleware]`, `[Routing]`                    |
| **Dependencies**    | `[Dependency]`                                                              |
| **External tool**   | `[Composer]`, `[npm]`, `[Gradle]`, `[PhpCsFixer]`, `[Sindri]`               |
| **Port**            | `[PHP]`, `[Java]`, `[Go]`, `[TypeScript]`, `[Python]`                       |
| **Project surface** | `[Git]`, `[Workflow]`, `[Template]`, `[Ruleset]`, `[GitHub]`, `[Functions]` |
| **Process**         | `[Process]`                                                                 |
| **Version line**    | `[25.x]`, `[26.x]`                                                          |
| **Release version** | `[v26.6.1]`                                                                 |
| **Exemption**       | `[Initial]`                                                                 |

A **module** root is spelled exactly as its source directory — `[Orm]`, not
`[ORM]`; `[Api]`, not `[API]`. Elsewhere use the name's own correct casing: `[CI]`
never `[Ci]`, `[GitHub]` never `[Github]`, `[npm]` never `[NPM]`.

**Expansion is expected**, and is how the vocabulary stays useful. The `.github`
repo holds workflows, rulesets, and templates, so `[Workflow]` and `[Ruleset]` are
roots there: neither appears on any list, both satisfy the two rules, and each says
considerably more than the `[GitHub]` they replace. When a repo grows a thing worth
naming, name it.

Having to name a root that honestly covers the whole change is a **forcing
function**: if no single root fits, the change is doing too much. Split it rather
than reaching for a broader root.

### Breadth is not a root

A change touching many modules is still _about_ something, and that something is
the root. This holds at real scale — every one of these touched 20+ modules:

```
[Application] refactor: Convert providers to instance-based contracts.
[Throwable] refactor: Give each component throwable a unique name.
[Provider] refactor: Update the provider naming convention.
[Container] refactor: Rename ProviderContract to ServiceProviderContract.
```

None of them needed a root meaning "everywhere," which is why there is no `[All]`
and no product-name root. The same reasoning one level down replaces stacking: a
change to terminal middleware across HTTP and CLI is not `[Http][Cli]` but
`[Middleware]`, because `Middleware`, `Routing`, `Server`, and `Throwable` all exist
under both.

### Process, Git, Workflow, and GitHub

Four roots sit close together. Each names **the thing the change touches**, not the
subject it concerns:

| Root         | Names                                                    |
| ------------ | -------------------------------------------------------- |
| `[Git]`      | Git's own configuration — `.gitignore`, `.gitattributes` |
| `[Workflow]` | Workflow definitions — reusable and per-repo             |
| `[GitHub]`   | GitHub-specific files in a repo that is not `.github`    |
| `[Process]`  | How the project works — conventions, the release process |

```
[Git] chore: Ignore the .worktrees directory.
[Workflow] fix: Stop the reusable release workflow skipping master.
[Workflow] ci: Update .github workflow refs to v26.12.1.
[Template] docs: Update the bug report issue template.
[GitHub] docs: Add a PR template superseding the org-wide one.
[Process] docs: Update release process documentation.
```

So the release _workflow_ is `[Workflow]` while the release _process documentation_
is `[Process]`, and the commit convention is `[Process]` even though what it governs
is git.

`[GitHub]` is a **fallback**, and by Rule 2 it is unavailable inside the `.github`
repo — there the thing itself gets named (`[Workflow]`, `[Ruleset]`, `[Template]`).
Elsewhere it covers GitHub-specific files that have no lower root of their own.

### Dependencies and tools

`[Dependency]` covers **changes to what is depended on** — adding, updating, or
removing a dependency, one or many, automated or by hand. The tool's own root covers
**changes to that tool** — its manifest structure, its scripts, its configuration,
its rules:

```
[Dependency] build: Update composer dependencies.
[Dependency] build: Add spatie/ray to the dev requirements.
[Composer] build: Add a post-root-package-install script.
[npm] build: Move vitest out of the root devDependencies.
[PhpCsFixer] build: Update to v26.1.1.
[PhpCsFixer] ci: Enable the strict rule set.
```

A dependency update is `build` whether a person or a job made it. The type says the
manifest changed; who changed it comes from `git log --author`, for the same reason
there is no automation type.

The last two lines are the boundary worth internalizing. Bumping a CI tool's _pinned
version_ is a dependency change, so `build` — but scoped to one tool, so the tool
root says more than `[Dependency]` would. Changing that tool's _configuration_ is
`ci`, because the checking machinery changed rather than the manifest. Keeping
`build` for version pinning is what makes those two distinguishable at all.

This split matters more than it looks: 91% of every historical `[Composer]`,
`[npm]`, `[Gradle]`, and `[Maven]` commit was an automated bump, so those six roots
were really one job wearing six names, and the genuine manifest work was invisible
among it.

### Release versions

A release commit carries the version it produced, so new versions are visible at a
glance in `git log`, and the several commits of one release are tied together by
their shared root:

```
[v26.6.1] chore: Update version for the release.
[v26.6.1] chore: Update ApplicationInfo for the release.
[v26.6.1] docs: Update CHANGELOG for the release.
```

A release-version root is the signal that a commit is release bookkeeping rather
than releasable work — see [`VERSIONING.md`](VERSIONING.md). Distinguish it from a
**version line** root: `[26.x]` marks a change made _for_ a maintenance line, e.g.
a backport.

### Stacking

Stack roots only when no single root honestly covers the change:

- **Narrowing in a cross-language repo**, where a module name alone is ambiguous
  and the port disambiguates it: `[Java][Http]`. A single-language repo needs no
  port root — the repo already is the language.
- **Genuinely unrelated areas** in one change — which is usually a sign the change
  should be two changes.

Never join roots with a slash. `[Http/Cli]` and `[Http/Cli/Event]` are malformed;
find the shared root, or stack them.

### Retired roots

These named a kind of change rather than a thing, so the type now carries them:

| Retired                     | Now                             |
| --------------------------- | ------------------------------- |
| `[Documentation]`, `[Docs]` | `docs`                          |
| `[Tests]`, `[Test]`         | `test`                          |
| `[Deprecation]`             | `deprecate`                     |
| `[Refactor]`, `[Revert]`    | `refactor`, `revert`            |
| `[Remove]`, `[Merge]`       | a real root plus a type         |
| `[CI]`                      | the specific tool, or `ci`      |
| `[All]`                     | the concept the change is about |
| `[Release]`                 | `[v26.6.1]` or `[Process]`      |

Freeing the root from the change kind is the point of the convention rather than a
side effect of it. A documentation-only repo used to put `[Documentation]` on every
commit regardless of subject, which conveyed nothing; it can now name what the
change is actually about. The same applies to `ci` and `test` work, which
can finally be scoped: `[Http] ci:`, `[Container] test:`.

`[Initial]` is the one exemption — the first commit of a brand-new repo stays
`[Initial] Initial commit.`, with no type, since it is pushed directly and never
sits in a pull request.

## A branch may mix roots and types

A branch commit does not have to carry the root that the pull request title
carries. Two commits on one branch do not have to carry the same root. The
squash merge discards every branch subject, so only the root in the pull request
title reaches the permanent history.

The type behaves the same way. The example under `## The shape` pairs a `fix`
commit and a `test` commit under one `fix` title. The pull request title carries
the root and the type of the whole change. The release bump reads the type on
that title and no other type. A branch that carries a `feat` commit, a
`deprecate` commit, or a breaking commit therefore takes `feat`, `deprecate`, or
`!` on the title. See [`VERSIONING.md`](VERSIONING.md).

```
commits    [Http] refactor: Align the HTTP terminal stage names.
           [Cli] refactor: Align the CLI terminal stage names.
PR title   [Middleware] refactor(#123): Align the terminal stage names across protocols
merged     [Middleware] refactor(#123): Align the terminal stage names across protocols (#456)
```

A force push that corrects a subject discards a reviewer's place in the diff,
and it changes nothing that a reader sees after the merge. Rewrite a branch
subject only when you rebase for another reason.

## Examples

**Good:**

```
[Container] feat(#123): Add support for contextual bindings
[Http] fix: Fix header normalization on HTTP/2 requests
[Http] feat(#88)!: Remove the deprecated request attribute accessors
[Http] deprecate: Deprecate the untyped request attribute accessors
[Cli] test: Cover the bare double dash operand case
[Middleware] refactor: Rename terminal stages to ResponseSent and ProcessExiting
[Throwable] refactor: Give each component throwable a unique name
[Java][Http] docs: Document the request attribute contract
[Http] ci: Scope the routing workflow to the HTTP test suite
[Workflow] ci: Update .github workflow refs to v26.12.1
[Template] docs: Update the bug report issue template
[Dependency] build: Update composer dependencies
[PhpCsFixer] build: Update to v26.1.1
[Process] docs: Update release process documentation
[v26.6.1] docs: Update CHANGELOG for the release
```

**Bad:**

```
fix bug                                    no root, no type, no detail
[http] fix: stuff                          lowercase root, vague description
Add caching.                               missing root and type
[Container] Add contextual bindings        missing type
[Documentation] docs: Fix a typo           retired root, and it restates the type
[Tests] Add coverage for the matcher.      retired root, missing type
[All] refactor: Rename the providers.      breadth is not a root; use [Provider]
[Valkyrja] style: Run the formatter.       the repo's own identity says nothing
[CI] ci: Add a PHPStan rule.               actor as root, and it restates the type
[Http/Cli] fix: Align the stages.          slashed roots; use the shared root
[Http][Cli] fix: Align the stages.         stacked where [Middleware] says it
[Http] fix(#123): Fix normalization.       trailing period on a PR title
[Http] fix: Fix normalization              missing period on a commit
[Http] fix(#123)! Remove the accessors     missing colon
```

## Enforcement

The `Commit Message Check` required status check runs on every pull request. It
validates four things:

- The PR title and every commit in the PR carry a root and a known type.
- A PR title carries no trailing period.
- Each of the branch's own commits carries a trailing period.
- No line of a commit message in the pull request exceeds 120 characters.

It reads the root as a shape, and it accepts any root.

Its scope matches the period rule exactly: it checks commits _in a pull request_ —
which are working-branch commits — and never sees a direct push to a protected
branch, which is why those carry no period.

Two exemptions: **Dependabot** pull requests, and **reverts**. GitHub's revert
button generates `Revert "<original title>"`, which cannot match the pattern, and a
revert is usually being done under time pressure — the check must not stand in the
way. Reverts should be rare enough that the exemption costs nothing.

Only the _shape_ is machine-checked. The root vocabulary is open by design, so no
pattern can validate it — the pattern accepts any bracketed word, and `[http]`
passes. Root choice and casing are review's job, as they are today.

The release run's own commits are pushed directly to a version branch and never
appear in a pull request, so the check never sees them. That makes the locked forms
below the only thing keeping them correct.

## Locked automated forms

Automation emits far more history than people do — the workflow-ref sweep alone
accounts for roughly 500 commits. These forms are **fixed**, and the workflow that
emits each one is named so the two cannot drift apart. Anything generated must
match exactly.

| Form                                                  | Kind     | Emitted by                        |
| ----------------------------------------------------- | -------- | --------------------------------- |
| `[Workflow] ci: Update <repo> workflow refs to <tag>` | PR title | `_update-workflow-refs.yml`       |
| `[Dependency] build: Update composer dependencies`    | PR title | `_php-update-dependencies.yml`    |
| `[Dependency] build: Update npm dependencies`         | PR title | `_ts-update-dependencies.yml`     |
| `[Dependency] build: Update Gradle dependencies`      | PR title | `_java-update-dependencies.yml`   |
| `[Dependency] build: Update Go dependencies`          | PR title | `_go-update-dependencies.yml`     |
| `[Dependency] build: Update Python dependencies`      | PR title | `_python-update-dependencies.yml` |
| `[Git] style: Add missing trailing newlines`          | PR title | `_fix-trailing-newlines.yml`      |
| `[v<X>] chore: Update version for the release`        | direct   | `_update-version-files.yml`       |
| `[v<X>] chore: Update CHANGELOG for the release`      | direct   | `_release.yml`                    |
| `[v<X>] chore: Update <Info> for the release`         | direct   | `_<lang>-update-info-files.yml`   |
| `[v<X>] chore: Update package.json for the release`   | direct   | `_ts-update-package-json.yml`     |
| `[v<X>] chore: Update pyproject.toml for the release` | direct   | `_python-update-pyproject.yml`    |
| `[<line>] chore: Update references for v<X>`          | direct   | `_<lang>-version-branch.yml`      |

**Nothing automation emits takes a trailing period.** Every form above is either a
PR title or a direct push to a protected branch, and both are permanent subject
lines. The period belongs to working-branch commits, which automation never writes.

Two roots that automation must **not** use: `[Dependabot]` and any other name for
the job that ran, because that is the actor (Rule 1); and the tool root for a
workflow-ref bump, because the file changed is a workflow pin, not the tool.
