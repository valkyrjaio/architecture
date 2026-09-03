# PR_DESCRIPTION.md — what a pull request description holds

The **cross-language** convention for the pull request description. It applies
in every Valkyrja repository, and to every contributor.

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

## Test every sentence

Keep a sentence only when it states what changed or why the change is right,
and the diff cannot show it. Discard every other sentence. Add a table or a
verification note only when it carries something a reviewer would otherwise
miss.

Four kinds of sentence always fail the test:

- **The origin of the change.** Where a rule lived before, who could see it,
  how many times a review flagged the problem. The reader acts on the change,
  not on its history.
- **The case for an uncontested decision.** A decision is one sentence: the
  choice, then the reason. Discard the argument for it — the cost of the
  alternative, the harm avoided. Make that case in the review thread, when a
  reviewer contests the decision.
- **A description of the added text.** The diff shows what the pull request
  adds. Do not describe the added text, and do not explain a reason the added
  text already states.
- **A defense of the edit's shape.** Why the change is a paragraph and not a
  table row, why a section keeps its place. The shape is the how, and the diff
  records the how.

> Wrong — six sentences, and every one fails the test:
>
> The requirement lived only in a code comment, which no reader of the document
> can see. A review flagged the missing key twice. This pull request records
> the requirement. The alternative was a silent default, but a default doubles
> the code paths to test and hides a wrong configuration from the developer.
> The new section states the error the framework throws, so a reader knows what
> a missing key does. The section sits below the configuration table, so a
> revert removes one paragraph instead of rebuilding the table.

> Right — the same change, the what and the why:
>
> Record in the component's `README.md` that the `app.env` key is required. The
> framework throws on a missing key instead of guessing a default, and the
> document did not say so.

## The description's prose holds only sentences

The prose in a description carries no heading of its own, no code block, and no
walkthrough of a failure. The diff holds the code, and one clause carries the
failure that the change answers. A table and a verification note stay permitted
under the test above, and the `Closes #123` line stays required. The template's
Types of changes and Changes sections stand outside this shape.

A promotion pull request is the exception, and its description never becomes a
commit. [`BRANCH_PROMOTION.md`](BRANCH_PROMOTION.md) states what that
description carries when a cherry-pick conflicts.

There is no sentence count, because a count sets a target and an author writes
to a target. Cut a sentence that states neither the what nor the why. Cut a
word that adds nothing to the sentence around it.

> Wrong — a title line and a walkthrough of the mechanism stand in for the
> sentence that says what changed:
>
> **Why the exiter sets the exit code**
>
> `Exiter::exit()` ends the process before the buffer drains, so a write the
> buffer still holds never reaches the operating system. The command then
> reports success for output that never arrived, and the listing below shows
> the body that replaces it.

> Right — one sentence says what changed, and one clause says why:
>
> Set the exit code instead of ending the process, because a process that ends
> at once drops a write a stream has buffered.

## A stable name sets the level of detail

"Update a method for naming consistency" names no method, and a restated
signature buries the answer in detail. A class name and a contract method name
survive every revision, so the description names them. Write "Rename
`checkRoute` to `isValidRoute` to follow the method-naming conventions". The
signature stays in the diff.

## Name the file, not the place inside it

Name the file and what changed in it. Do not name where inside the file the
change went — after which section, below which table, next to which method.
The diff shows the position. This rule governs the Description prose and the
Changes list alike.

> Wrong — the bullet names a position the diff already shows:
>
> - **`README.md`** — adds the queue section below the configuration table.

> Right — the bullet names the file and the change:
>
> - **`README.md`** — documents the queue component.

A position rarely matters. When the position is the change, name it, and state
why it matters:

> - **`ci.yml`** — runs the lint job before the build jobs, because every
>   build job reuses the lint cache.

## A temporary explanation lands in the description, not in a comment

A comment in code or config must stay true indefinitely, so a temporary
explanation never goes in a comment ([`COMMENTS.md`](COMMENTS.md)). Put the
explanation in the description instead.

State the explanation in a sentence or two. The description takes the
explanation, not the essay around it.

## The template

The description follows the
[PR template](https://github.com/valkyrjaio/.github/blob/26.x/.github/PULL_REQUEST_TEMPLATE.md),
which has three sections:

- **Description** — the prose that the rules above govern.
- **Types of changes** — check every box that applies.
- **Changes** — one bullet per file or per logical change: the bold file or
  component name, an em dash, and what changed — never the place inside the
  file. The list is optional for a small pull request whose Description
  already covers everything.

When an issue tracks the work, put `Closes #123` in the Description. The
description becomes the squash commit body, so that line is what closes the
issue on merge, and it is where the link durably lives.
