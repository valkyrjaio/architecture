# Sindri CLI Output Design

This document specifies the structure, visual grammar, and semantic rules for
Sindri's terminal output. It serves as both an implementation reference for
contributors and a public contract for users and tooling that consume
Sindri's output.

Design Principles
-----------------

Sindri's output follows five principles:

1. **Hierarchy before density** — organize output into clearly delimited
   sections (banner, config, work, summary) before filling any section with
   detail.
2. **Information over decoration** — every line earns its place by telling
   the user something they need. Thematic flourishes (Norse verbs, hammer
   icon) appear once per invocation, not throughout.
3. **Happy paths stay quiet** — steps that succeed without incident produce
   a single line. Detail lines appear only when a step has something to say.
4. **Consistency beats cleverness** — the same verb, separator, and layout
   are used across all runs. A user who has read Sindri's output once knows
   how to read it everywhere.
5. **Graceful degradation** — the output structure that handles success
   cleanly must also handle partial failure, total failure, and no-op runs
   without redesign.

Output Structure
----------------

Every Sindri invocation produces output in four sections:

```
[Banner]

[Resolved Configuration]

[Work Output]

[Summary]
```

Sections are separated by single blank lines. Within each section, content
follows its own internal rhythm.

### Banner

The banner identifies the tool, its runtime environment, and the action
about to be performed. It is drawn using Unicode box-drawing characters to
form a visually distinct frame.

```
╭── Sindri v26.0.0
│
│   ▗▄█████▄▖
│   ▝▀█████▀▘
│       █
│       █
│
│   Running on Valkyrja v26.1.0 · PHP 8.4.17
│   ~/app
╰── Forging application cache · cache:generate
```

```php
echo <<<TEXT
    
    ╭── Sindri v26.0.0
    │
    │   ▗▄█████▄▖
    │   ▝▀█████▀▘
    │       █
    │       █
    │
    │   Running on Valkyrja v26.1.0 · PHP 8.4.17
    │   ~/app
    ╰── Forging application cache · cache:generate
    
    TEXT;
```

The banner has three logical parts:

**Top row** — `╭── Sindri v26.0.0` — identifies the tool and its version.

**Body** — contains the Mjölnir icon, a blank line, and two context lines:

1. The runtime environment: framework version and PHP version, separated by
   `·`
2. The project's working directory

**Bottom row** — `╰── Forging application cache · cache:generate` — names
the action narratively ("Forging application cache") and preserves the
exact subcommand that produced it (`cache:generate`). Additional flags are
appended after the subcommand name.

The bottom row serves triple duty: it closes the banner frame visually,
describes the action in human terms, and retains verifiable command
information.

#### Mjölnir Icon

The hammer icon is rendered using Unicode block and quadrant characters:

```
▗▄█████▄▖
▝▀█████▀▘
    █
    █
```

It is indented three spaces from the left frame (`│   ▗▄█████▄▖`), aligning
it visually with the text content below.

### Resolved Configuration

The configuration block shows what Sindri *resolved* from the user's
invocation, independent of what they typed. This block answers "what inputs
is Sindri actually working with?"

```
Config File:    src/App/Config.php
Data Directory: src/App/Data
```

**Rules:**

- Keys are aligned with padding so that values begin at the same column.
- Paths are shown relative to the project root when they fall within it;
  fall back to `~`-prefixed absolute paths otherwise.
- Flags at default values are displayed alongside flags that were
  explicitly set — the user sees the full resolved configuration, not just
  their overrides.
- The block is omitted when the subcommand has no resolved configuration
  (e.g., simple commands with no arguments or flags).

### Work Output

The work output section displays each unit of work Sindri performs, one
per line, using dot-leader formatting with right-aligned status labels.

```
Generating Container Data......................Success
Generating Event Data..........................Skipped
  ▸ No data change detected, file contents identical
Generating Cli Routes Data.....................Warning
  ▸ Route handler Valkyrja\Cli\RouteX::handle has no return type
Generating Http Routes Data....................Fail
  ▸ Missing required constraint on route parameter {id}
  ▸ See src/Http/Routing/UserRoutes.php:42
```

**Rules:**

- Each step line starts with a verb and a subject (e.g., `Generating
  Container Data`).
- Dot leaders (`.....`) fill the space between the step description and
  the status label. Status labels are aligned to the same column across
  all steps.
- Status labels are single words: `Success`, `Skipped`, `Warning`, `Fail`.
- Detail lines are prefixed with `▸` and indented two spaces from the
  step line's starting column.
- A step with no detail lines has no detail output.
- Steps with detail lines are separated from the next step by a blank
  line; consecutive plain-`Success` steps stack without separation.

### Summary

The summary is a single line that reports the invocation's completion
time and per-status counts of work units.

```
Completed in 2.3s · 4 written
```

```
Completed in 2.3s · 1 written · 1 skipped · 1 warning · 1 failed
```

**Rules:**

- Leading verb is `Completed` across all outcomes — including partial
  failure and total failure. The verb does not editorialize about whether
  the run succeeded; the counts do that job.
- Completion time is reported in seconds with one decimal place.
- Counts are separated by middle-dot (`·`).
- Zero-count categories are omitted from the summary.
- Categories always appear in the order: written, skipped, warning,
  failed.
- Category labels: `written` (not `success`), `skipped`, `warning`,
  `failed`. "Written" reflects what was actually produced; "success"
  would describe the operation, which is different.

Status Semantics
----------------

Each of the four status labels has precise meaning:

| Status    | Meaning                                                              | Detail lines expected?                 |
|-----------|----------------------------------------------------------------------|----------------------------------------|
| `Success` | The step completed cleanly and produced its output                   | No                                     |
| `Skipped` | The step was intentionally not performed; existing output is current | Yes — explain why                      |
| `Warning` | The step produced output but encountered a non-blocking issue        | Yes — explain the issue                |
| `Fail`    | The step could not produce its output                                | Yes — explain and point to remediation |

**Summary counting:**

- `Success` → counted as `written`
- `Warning` → counted as `warning` (and also as `written`, since output was
  produced)
- `Skipped` → counted as `skipped`
- `Fail` → counted as `failed`

In the mixed-outcome summary, a step that produced `Warning` is reported in
`warning` count, not `written` count, to make non-clean outcomes visible.

Color (When Available)
----------------------

When Sindri detects a color-capable terminal, the following color
assignments apply:

| Element                    | Color                       |
|----------------------------|-----------------------------|
| Banner frame characters    | Default                     |
| Mjölnir icon               | Orange / warm (forge fire)  |
| `Success` status label     | Green                       |
| `Skipped` status label     | Gray / dim                  |
| `Warning` status label     | Yellow                      |
| `Fail` status label        | Red                         |
| Detail lines (`▸`)         | Parent status color, dimmed |
| Summary: `failed` count    | Red (when > 0)              |
| Summary: `warning` count   | Yellow (when > 0)           |
| Configuration block keys   | Gray / dim                  |
| Configuration block values | Default                     |

Color is decorative. All information conveyed by color must also be
conveyed by the status label text, so non-color terminals (CI logs, pipes,
`NO_COLOR=1`) lose no information.

Exit Codes
----------

Sindri exit codes are determined by the worst status in the work output:

| Worst status | Exit code |
|--------------|-----------|
| All Success  | `0`       |
| Skipped      | `0`       |
| Warning      | `0`       |
| Fail (any)   | `1`       |

A `Warning` exit code of `0` is deliberate — warnings are informational and
must not fail CI. Builds that treat warnings as errors can opt in via a
`--warnings-as-errors` or equivalent flag.

Special Cases
-------------

### Happy path

All four steps succeed. Summary is compact.

```
Generating Container Data......................Success
Generating Event Data..........................Success
Generating Cli Routes Data.....................Success
Generating Http Routes Data....................Success

Completed in 2.3s · 4 written
```

### No-op run

All files are current, nothing needs regenerating.

```
Generating Container Data......................Skipped
  ▸ No data change detected, file contents identical
Generating Event Data..........................Skipped
  ▸ No data change detected, file contents identical
Generating Cli Routes Data.....................Skipped
  ▸ No data change detected, file contents identical
Generating Http Routes Data....................Skipped
  ▸ No data change detected, file contents identical

Completed in 0.3s · 4 skipped
```

### Total failure

All steps fail. Summary is blunt.

```
Generating Container Data......................Fail
  ▸ Config file not found at src/App/Config.php
Generating Event Data..........................Fail
  ▸ Config file not found at src/App/Config.php
Generating Cli Routes Data.....................Fail
  ▸ Config file not found at src/App/Config.php
Generating Http Routes Data....................Fail
  ▸ Config file not found at src/App/Config.php

Completed in 0.1s · 4 failed
```

### Mixed outcome

Realistic build with varied per-step results.

```
Generating Container Data......................Success
Generating Event Data..........................Skipped
  ▸ No data change detected, file contents identical
Generating Cli Routes Data.....................Warning
  ▸ Route handler Valkyrja\Cli\RouteX::handle has no return type
Generating Http Routes Data....................Fail
  ▸ Missing required constraint on route parameter {id}
  ▸ See src/Http/Routing/UserRoutes.php:42

Completed in 2.3s · 1 written · 1 skipped · 1 warning · 1 failed
```

Alignment Reference
-------------------

Column alignment rules used throughout Sindri output:

- **Configuration block keys:** padded to align the longest key's colon
  column.
- **Dot-leader status column:** all step lines pad to the same total
  width so status labels align on the right.
- **Detail line indent:** two spaces from the start of the step line,
  followed by `▸ ` and the detail content.
- **Banner body content:** three spaces from the left frame character
  (`│   `).

Non-Goals
---------

This document does not specify:

- The internal architecture of Sindri's output rendering code.
- The exact terminal dimensions Sindri should target (the design is
  flexible across standard 80-column and wider terminals).
- Color palette values in specific ANSI codes (implementation detail;
  use standard terminal colors mapped to roles).
- Internationalization of status labels (future concern; currently
  English only).
