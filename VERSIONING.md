# Versioning & Releases

Every repo in the project shares one version scheme:

```
YY.FEATURE.PATCH        e.g. 26.6.0
```

| Component | Meaning                                         | Bumped by     |
| --------- | ----------------------------------------------- | ------------- |
| `YY`      | Two-digit year — the major line, one per year   | manually only |
| `FEATURE` | New capability, deprecation, or breaking change | automatically |
| `PATCH`   | Everything else                                 | automatically |

Each year gets a maintenance branch named after it — `26.x` for 2026 — and every
stable release is cut from one of those branches, never from `master`.

This document covers how a version number is built and how a release is cut. For
how long each version line takes patches, and for the runtime that each language
port needs, see [`VERSION_SUPPORT.md`](VERSION_SUPPORT.md).

## Why the middle component serves two roles

This is calendar versioning with a semantic tail. `YY` is the major, and it only
moves when a new year opens, so there is no component left to distinguish a new
feature from a breaking change. Both move `FEATURE`.

That is a deliberate trade, not an oversight. Under strict SemVer an urgent fix
that happens to break a public contract cannot ship until someone cuts a major
release — so either the fix waits on ceremony that helps nobody, or the break gets
smuggled into a patch. Collapsing the two lets the fix ship the day it is ready.

The consequence to understand: a consumer on `^26.1` will pick up a breaking
`26.2.0` automatically. Two things keep that manageable.

1. **Planned breaks wait for the year boundary.** Deprecate first, remove at the
   next `YY`. That is what the `deprecate` type and the `deprecation/` branch
   prefix are for.
2. **Unplanned breaks are marked.** Any change that breaks a public contract
   carries `!` in its PR title, so the release notes name it even though the
   version number cannot. See [`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md).

### Until adoption, the year branch takes every break

Valkyrja has no users yet, so the current-year branch takes a breaking change
whenever the framework needs one. A planned break does not wait for the year
boundary, and a new feature, a deprecation and a breaking change land on that
branch rather than on `master`.

The rules above return when adoption grows and the framework settles.

## How the bump is computed

Releases read the commit subjects on the version branch since the last tag.
Because squash merges take their subject from the PR title, those subjects _are_
the conventional PR titles — no API calls, no commit-body parsing.

| Found in the window                     | Bump       |
| --------------------------------------- | ---------- |
| any `feat`, any `deprecate`, or any `!` | `FEATURE`  |
| any other type                          | `PATCH`    |
| nothing, or only release-version roots  | no release |

A release run commits its own bookkeeping — the version file, the application
info, the CHANGELOG — and then tags the last of those commits, so they ship
_inside_ the release they describe rather than trailing it:

```
v26.6.1 → [v26.6.1] chore: Update version for the release.
          [v26.6.1] chore: Update ApplicationInfo for the release.
          [v26.6.1] docs: Update CHANGELOG for the release.
```

Between one release and the next merged pull request there are therefore no
commits at all, and "nothing pending" is the ordinary reason a scheduled run does
nothing.

**The non-triggering signal is the root, not the type.** Release bookkeeping is
identified by its release-version root (`[v26.6.1]`), which is why those commits
take whatever type their change deserves. Keying it to the type instead would not
work: `chore` and `docs` also carry ordinary merged work, which must still cut a
patch.

That rule is a safety net rather than the main mechanism. It covers a run that
pushed its bookkeeping but failed before tagging: the next run then sees leftover
release commits and must not mistake that paperwork for releasable work.

## The CHANGELOG is not the release notes

Two different artifacts, and conflating them leads to changing the wrong one.

`CHANGELOG.md` is the **in-repo ledger**: one flat, chronological line per merged
pull request, per version, forever. Flat is the point — it is a record to search, so
every entry has the same shape and nothing is hidden inside a category.

The **GitHub release notes** are the presentation layer, generated per release and
read once by people deciding whether to upgrade. Grouping and prose belong here, not
in the ledger.

So a release-notes improvement is never a reason to restructure `CHANGELOG.md`. And
because the type now sits in every PR title, a hand-written release page — the kind
that walks through what is new and why — is far easier to assemble from the ledger
than it used to be.

## Cutting a release

`release-new-version.yml` takes a single `bump` input:

| Option    | Effect                                                     |
| --------- | ---------------------------------------------------------- |
| `auto`    | Compute from commit types since the last tag (**default**) |
| `patch`   | Force a patch bump                                         |
| `feature` | Force a feature bump                                       |
| `yearly`  | First release on a new `YY.x` branch — always `YY.0.0`     |
| `rc`      | Release candidate for the next year, from `master` only    |

`auto` is the normal path, and the explicit `patch` / `feature` options remain as
overrides for when the computed answer is wrong.

Stable releases must be dispatched from a `YY.x` branch. Dispatching one from
`master` fails by design — `master` is where the next year is prepared, and `rc`
is the only release type that comes from it.

## Automation owns the year, never the year boundary

`yearly` and `rc` are **manual only, always**, and this is the one rule in the
release design worth defending hardest.

The split follows the version format exactly. `FEATURE` and `PATCH` move on
_evidence_ — the commit log says what merged, so a machine can compute the answer.
`YY` moves on a _decision_: that a year of accumulated breaking work hangs together
well enough to ship. An RC is that same decision announced early. Nothing in a
commit type expresses readiness, so nothing in the log can be read to produce it.

Three further reasons the boundary stays human:

- **RC numbers are a communication device.** `RC1`, `RC2`, `RC3` publish to the
  language registries and their sequence tells consumers something. Automation
  minting them on a schedule turns the numbers into noise.
- **`master` is deliberately the least stable branch** — features, deprecations,
  and breaking changes all target it. Giving the riskiest branch the most casual
  release path inverts the risk ordering.
- **A yearly code path cannot be trusted.** Anything that fires once every twelve
  months has not been exercised since the last time, and it would fire during
  precisely the window someone is hand-preparing the rollover.

So the sweep below **never dispatches to `master`**, which leaves the RC path
unreachable from automation by construction rather than by conditionals. That
matters: a design that needs several conditions to all hold in order to _avoid_
publishing is a design where being wrong publishes.

## Automatic releases

A scheduled sweep dispatches `auto` releases across every supported version
branch, so merged work ships on a daily cadence instead of waiting for someone to
remember. Nothing accumulates and nothing gets stuck behind an unrelated change.

Each dispatched run reads its own branch's log for commits pending release, so a
branch with nothing merged since its last tag skips that day instead of cutting an
empty release. That is what makes a daily sweep safe across every supported
branch: quiet branches cost a workflow run and produce nothing.

Which branches it covers comes from the `SUPPORTED_VERSIONS` repository variable
— a **regex** matched against the major with bash `=~`, e.g. `2[6-9]`. Not a
space-separated list; `SUPPORTED_LANGUAGES` is the space-separated one, and the two
are easy to confuse. A value like `26 27` matches nothing, which aborts releases
loudly and makes several other sweeps silently process zero branches.

The sweep enumerates `??.x` branches, keeps those whose major matches, and
dispatches each one's own `release-new-version.yml` with `bump: auto`.

Two mechanical constraints shape that design, and both are worth knowing before
changing it:

- **Scheduled workflows only run from the default branch**, which for these repos
  is the current-year `YY.x` branch, not `master`.
- **The release workflow derives its target from the ref it runs on.** So the sweep
  cannot fan out in-process; it dispatches each branch with
  `gh workflow run … --ref 26.x`, the same idiom the existing cross-repo sweeps use.

A `workflow_dispatch` triggered with `GITHUB_TOKEN` does not create a run, so the
sweep authenticates as the project's GitHub App.

Before creating a new year's branch, widen `SUPPORTED_VERSIONS` first — the
new-version workflow validates the _new_ major against it and aborts otherwise.
