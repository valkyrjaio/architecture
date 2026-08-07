# PR_DESCRIPTION.md — what a pull request description holds

The **cross-language** convention for the pull request description. It applies
in every Valkyrja repository, and to every contributor, human or agent.

The description is permanent. Merges are squash-only, and the squash merge
writes the pull request title as the commit subject and the entire description
as the commit body ([`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md)). The
description is what `git log` shows for the change, forever. Write it for the
reviewer today, and for the reader years later.

Warning: too little and too much fail the same way — the reader cannot tell
what the pull request is about. A short description names nothing. A long
description buries the sentence the reader needs. The rules below remove both
failures.

## The description holds the what and the why

A pull request answers five questions. GitHub records the when and the where,
and the diff records the how. The description holds the two the reader cannot
derive: what changed, and why the change is right.

Warning: a description that restates a specific edit binds the description to
the code. A later push then makes the description false until someone corrects
both together. Write "Update the Python binding keys, because the old keys were
wrong", not "Change the key from `valkyrja.container.ContainerContract` to
`valkyrja.container.manager.ContainerContract`". The first form stays true
through every revision of the pull request.

## Budget the description

State what changed, why, and any trap — most pull requests fit in three to six
sentences. Add a table or a verification note only when it carries something a
reviewer would otherwise miss. Do not restate the diff, and do not narrate the
process that produced it. A reviewer skips a wall of text, so every extra
sentence hides the one the reviewer needs.

## A stable name sets the level of detail

"Update a method for naming consistency" names no method, and a restated
signature buries the answer in detail. A class name and a contract method name
survive every revision, so the description names them. Write "Rename
`checkRoute` to `isValidRoute` to follow the method-naming families". The
signature stays in the diff.

## A decision lands in the description, not in a comment

A comment in code must stay true indefinitely, so a temporary explanation never
goes in a comment — a version pinned pending a release, a workaround awaiting a
fix ([`AGENTS.md`](AGENTS.md) §3, rule 10). Put the explanation in the
description instead. Nothing is lost: the description becomes the commit body,
so the explanation is pinned to when it was true, and `git log` and `git blame`
reach it.

State the decision in a sentence or two. The description takes the decision,
not the essay around it.

## The template

The description follows the
[PR template](https://github.com/valkyrjaio/.github/blob/26.x/.github/PULL_REQUEST_TEMPLATE.md),
which has three sections:

- **Description** — the prose that the rules above govern.
- **Types of changes** — check every box that applies.
- **Changes** — one bullet per file or per logical change: the bold file or
  component name, an em dash, and what changed. The list is optional for a
  small pull request whose Description already covers everything.

When an issue tracks the work, put `Closes #123` in the Description. The
description becomes the squash commit body, so that line is what closes the
issue on merge, and it is where the link durably lives.
