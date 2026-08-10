# SHELL_SCRIPTS.md — the rules for shell

The **cross-language** convention for shell. It applies in every Valkyrja
repository, and to every contributor, human or agent.

A repository in every language holds shell. A CI helper, a pre-commit hook, and
a script under `.github/ci/scripts/` are all shell. Every script runs under
`bash`, and its first line is `#!/usr/bin/env bash`. The license header
([`AGENTS.md`](AGENTS.md) §5) follows the shebang. The rules below use bash
constructs that plain `sh` does not have.

SonarCloud reads a `.sh` file in each repository, and `shellcheck` reads one
on a developer's machine. Most rules below are a rule of one of those two
tools, so a script that breaks one reports a finding.

Warning: shell inside a GitHub Actions `run:` block is invisible to both tools.
Neither one reads a workflow file, so a rule below goes unenforced until the
shell moves into a `.sh` file. That is a reason to put the work in a script, not
a reason to relax the rule.

## Use `[[ ]]` for a test, never `[ ]`

`[[ ]]` is a shell keyword rather than a command, so it parses the test rather
than expanding it first. An empty variable and a variable holding a space are
then safe without a quote.

```bash
# Wrong — `[ ]` is a command, and its arguments expand before it runs.
if [ -n "$COVERAGE_REPORT" ]; then
    echo "The coverage report is there."
fi
```

```bash
# Right — `[[ ]]` parses the test.
if [[ -n "$COVERAGE_REPORT" ]]; then
    echo "The coverage report is there."
fi
```

Warning: quote the right side of a `=` inside `[[ ]]`. An unquoted right side is
a pattern, so `[[ "$a" = $b ]]` matches a glob where `[ "$a" = "$b" ]` compared
two strings.

## Give every `case` a default branch

A `case` with no `*)` says nothing about the value it does not name, and a
reader cannot tell a deliberate skip from a missing branch.

```bash
# Right — the default states that every other value is ignored.
case "$CHECK_OUTCOME" in
    success) passed=$((passed + 1)) ;;
    failure) failed=$((failed + 1)) ;;
    *) ;;
esac
```

## A local variable takes `lower_case`

A script mixes a local with a global that a function reports through, and case
is what separates them.

```bash
# Right — `branch` belongs to the function, `DEFAULT_BRANCH` to the caller.
read_default_branch() {
    local branch
    branch=$(git remote show origin | sed -n 's/.*HEAD branch: //p')

    DEFAULT_BRANCH="$branch"
}
```

## Put a list of arguments in an array, never in a string

A string of arguments needs word splitting to work, and word splitting breaks
on a value that holds a space.

```bash
# Wrong — the string splits on every space, including one inside a value.
PHPUNIT_FLAGS="--colors=always --filter $TEST_FILTER"
composer phpunit -- $PHPUNIT_FLAGS
```

```bash
# Right — the array carries each argument, whatever it holds.
PHPUNIT_FLAGS=(--colors=always --filter "$TEST_FILTER")
composer phpunit -- "${PHPUNIT_FLAGS[@]}"
```

## A suppression states its reason

A linter is wrong sometimes, and the next reader must be able to tell a
considered suppression from a silenced one.

```bash
# shellcheck disable=SC2001 # `^` anchors each line, and `${var//}` anchors nothing.
SUMMARY=$(echo "$REPORT" | sed 's/^/- /')
```

## The `set` line follows the workflow

Warning: the `set` line of a script that a GitHub Actions workflow runs depends
on how the workflow runs it. A `run:` step that names no shell gives `bash -e`
alone, and an explicit `shell: bash` gives `-eo pipefail`. A `set` line that
does not match the step changes what the step did.

Match the `set` line of a script to the shell options of the `run:` step that
runs the script. `.github/workflows/README.md` states one exception: a script
that a `shell: bash` step runs also sets `-u`.

```bash
# Wrong — a bare `run:` step gives `-e` alone; the script adds two options.
set -euo pipefail
```

```bash
# Right — a bare `run:` step runs the script, so the script sets `-e` alone.
set -e
```

```bash
# Right — the step names `shell: bash`, and the script also sets `-u`.
set -euo pipefail
```

The rule and the table are in
[`.github/workflows/README.md`](https://github.com/valkyrjaio/.github/blob/26.x/.github/workflows/README.md).
