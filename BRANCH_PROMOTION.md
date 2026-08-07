# Branch Promotion

This document records the design for the branch promotion automation. The
automation moves each merged change from a lower version branch to every higher
branch that needs the change. It replaces the rebase-to-master workflows and the
manual cherry-pick dispatch in the `.github` repository.

---

## 1. The problem

A fix merges to the lowest affected `??.x` branch (see
[`AGENTS.md`](AGENTS.md) §7). Every higher branch needs the same fix. Two
mechanisms carry fixes upward today, and each one has a limit:

- **`rebase-to-master`** force-pushes `master` onto the latest `??.x`. The
  workflow refuses when `master` holds commits of its own, so it stops working
  the moment `master` starts the next major's work.
- **`cherry-pick-commits`** is a manual dispatch. A person picks the hash, picks
  the destination, and runs it. Nothing records which commits moved, and nothing
  notices the commits that did not.

The automation below promotes every eligible commit on a schedule, records what
moved, and asks a person only when a commit cannot move on its own.

---

## 2. What a promoted commit preserves

Warning: this section is the reason the design lands commits by fast-forward.
Read it before the mechanics.

The squash merge on the source branch produces a commit that identifies itself:

- The **subject** is the pull request title plus the pull request number.
  GitHub links the number to the origin pull request from any branch.
- The **body** is the pull request description, which holds the durable
  explanation (see [`AGENTS.md`](AGENTS.md) §3, rule 10).
- The **author** is the person who wrote the change.

A promotion preserves all three, and adds one line: the
`(cherry picked from commit …)` trailer that `git cherry-pick -x` writes.
GitHub links that hash to the origin commit. From `git blame` on any branch, the
origin pull request is one click through the subject, and the origin commit is
one click through the trailer.

Warning: a squash merge of a promotion pull request destroys all three. The
squash writes a new commit whose subject is the promotion pull request's title,
whose body is the promotion pull request's description, and whose author is the
bot. Every repository enforces squash-only merges, so a promotion pull request
must not land through the merge button.

---

## 3. The design

One scheduled sweep runs in the `.github` repository and fans out per
repository. For each repository, for each rung of the branch ladder, the sweep:

1. Enumerates the merged commits on the lower branch that the ledger (§6) does
   not show on the higher branch.
2. Filters them through the promotion signal (§4).
3. Cherry-picks each eligible commit with `git cherry-pick -x` onto the higher
   branch's tip, in commit order.
4. Opens a **promotion pull request** whose branch holds exactly that one
   commit. CI runs on the pull request.
5. Lands the pull request by fast-forward (§5) when the checks pass.
6. Escalates to an agent when the cherry-pick conflicts (§7).

The sweep promotes one rung at a time: `26.x` → `27.x`, then `27.x` → `master`.
Each rung picks from the commit the previous rung landed, so a conflict
resolution carries forward instead of resolving twice. The trailer chain links
every rung back to the origin, and the subject keeps the origin pull request
number on every rung.

The sweep runs from day one, while `master` holds no commits of its own. Every
pick is then clean, and `master` stays current commit by commit — with no force
push, and no backup branch.

---

## 4. The promotion signal

The sweep decides where a commit goes from two sources, in priority order.

**The pull request template's `Ships to` section** is the explicit signal. The
author checks the branches the change applies to, or checks "this branch only".
The squash merge writes the section into the commit body, so the sweep reads
the signal from the commit itself. This is the override channel: a fix that
only applies to one version states so here, and the sweep obeys.

**The commit type** supplies the default when the section is absent:

| Type on a version branch                       | Default             |
| ---------------------------------------------- | ------------------- |
| `fix`, `perf`, `refactor`, `style`, `docs`     | Promote upward      |
| `feat`, `deprecate`, any `!`                   | Promote upward      |
| `build`, `ci`, `chore` from a bot              | Never promote       |
| `revert`                                       | Promote upward      |

Warning: a bot's dependency and workflow commits never promote, whatever their
type. Each branch's own automation maintains its own lockfiles and workflow
pins. A promoted lockfile commit fights that automation and corrupts the
destination's `content-hash`.

---

## 5. Landing by fast-forward

The sweep lands a promotion pull request by moving the destination ref to the
pull request's head:

```bash
gh api --method PATCH "repos/valkyrjaio/valkyrja/git/refs/heads/master" \
  -f sha="$PROMOTION_HEAD_SHA"
```

The commit that lands is byte-identical to the commit CI tested: same author,
same subject, same body, same trailer. GitHub marks the promotion pull request
as merged when its head commit reaches the base branch, so the pull request
closes as merged and the commit page shows both pull requests — the origin
through the subject, and the promotion through the merge association.

The fast-forward requires the promotion branch to sit on the destination's
current tip. When the tip has moved, the sweep re-picks onto the new tip and
lets CI run again. Promotions are serialized per repository per rung, so a
stale tip is rare.

Warning: the GitHub App bypasses required checks, so the ref update succeeds
whatever the checks say. The sweep must read the check conclusions itself and
fast-forward only when every required check has passed.

The squash-only repository settings stay as they are. No merge method changes,
and the merge button stays squash-only for every human pull request.

---

## 6. The ledger

The `-x` trailer is the ledger. The fast-forward lands the trailer verbatim, so
the destination branch's history records every promotion:

```bash
git log origin/master --grep "cherry picked from commit $ORIGIN_SHA"
```

A commit is promoted when the grep finds it, and unpromoted when the grep does
not. No side state exists, so the ledger cannot drift from the branches.

A deliberate skip needs a record too, or the sweep reconsiders the commit on
every run. The sweep skips a commit when the origin pull request carries the
`promotion:skip` label, and the label is the durable record of the decision.

---

## 7. Conflicts and the agent

A cherry-pick that conflicts escalates to an agent. The agent resolves the
conflict, amends the resolution into the cherry-pick commit, and pushes the
promotion branch. The promotion branch holds exactly one commit at every point.
That contract is the deliberate exception to the no-amend rule: the amend
preserves the single pristine commit that the fast-forward lands, and the pull
request description — which never becomes a commit — carries the resolution
notes.

The label on the promotion pull request states who must act next:

- **`promotion:agent-resolved`** — the agent resolved the conflict. The sweep
  never fast-forwards this pull request on its own. A person reviews the
  resolution and approves; the sweep lands it on the next run after the
  approval.
- **`promotion:blocked`** — the agent could not resolve the conflict, or was
  not confident. The pull request description states what blocked it. A person
  either pushes the resolution or leaves a comment that tells the agent how to
  proceed.

The next sweep run reads each open promotion pull request. New human commits
mean the sweep re-validates and proceeds. New human comments mean the sweep
dispatches the agent with the thread as instructions. The pull request is the
whole state; the sweep keeps no memory between runs.

---

## 8. Decisions considered and rejected

**Squash-merge the promotion pull request.** Rejected. The squash rewrites the
author, the subject, and the body (§2). The trace would be text the tooling
asserts, not something git records.

**Rebase-merge the promotion pull request.** Rejected. A rebase merge preserves
the author and the message, but it requires `allow_rebase_merge` on every
repository. That gives every human pull request a second merge button, and a
multi-commit feature pull request that lands through it pollutes history. The
fast-forward gives strictly more — the identical commit — without changing any
repository setting.

**Direct push without a pull request.** Rejected as the default. The current
`cherry-pick-commits` workflow pushes straight to the destination, which skips
CI and offers no surface for review or for the conflict loop. A cherry-pick
that applies cleanly can still be wrong on the destination, and CI is the
backstop. The manual dispatch stays as the attended escape hatch.

**A central state file for the ledger.** Rejected. A file drifts from the
branches it describes, and every write needs its own commit. The `-x` trailer
travels with the commit it records and cannot drift (§6).

---

## 9. What the automation replaces

- **`rebase-to-master`** and **`rebase-all-to-master`** — retired. The sweep
  keeps `master` current without force pushes, before and after `master`
  diverges.
- **`cherry-pick-commits`** — kept as the attended escape hatch for a one-off,
  human-driven pick.
