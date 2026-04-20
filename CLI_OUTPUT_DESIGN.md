# Valkyrja CLI Output Design

This document specifies the structure, visual grammar, and customization
hooks for Valkyrja's default CLI output. It serves as both an
implementation reference for framework contributors and a public contract
for application developers building on Valkyrja.

Banner Template
---------------

Every Valkyrja CLI invocation produces a banner with the following
structure:

```
╭── {app-name} v{app-version}
│
│   {app-icon}
│
│   Running on Valkyrja v{valkyrja-version} · PHP {php-version}
│   {project-root-path}
╰── {action-description} · {exact-command}
```

```php

echo <<<TEXT

    ╭── Application v1.0.0
    │
    │   ▗▄▄▖     ▗▄▄▖
    │   ▝▜██▄▄▄▄▄██▛▘
    │      ▝▜███▛▘
    │         █
    │
    │   Running on Valkyrja vXX.X.X · PHP X.X.X
    │   ~/app
    ╰── Running whatever command · whatever:command

    TEXT;

```

Each slot is resolved from `CliConfig` with sensible defaults:

| Slot                 | Source                            | Default                              |
|----------------------|-----------------------------------|--------------------------------------|
| `app-name`           | `CliConfig::$appName`             | `Valkyrja Application`               |
| `app-version`        | `CliConfig::$appVersion`          | `1.0.0`                              |
| `app-icon`           | `CliConfig::$appIcon`             | Valkyrie icon (see below)            |
| `valkyrja-version`   | Valkyrja runtime constant         | Current framework version            |
| `php-version`        | `PHP_VERSION` constant            | Current PHP version                  |
| `project-root-path`  | `CliConfig::$projectRoot`         | Detected from application bootstrap  |
| `action-description` | Per-command, set by command class | Command class's declared description |
| `exact-command`      | Preserved from invocation         | Subcommand name and flags as typed   |

Applications override any slot by setting the corresponding property on
their `CliConfig` instance. Properties not set fall through to defaults.

Design Principles
-----------------

Valkyrja's CLI output follows the same five principles as Sindri:

1. **Hierarchy before density** — organize output into clearly delimited
   sections (banner, work, summary) before filling any section with
   detail.
2. **Information over decoration** — every line earns its place by
   telling the user something they need.
3. **Happy paths stay quiet** — steps that succeed without incident
   produce a single line. Detail lines appear only when a step has
   something to say.
4. **Consistency beats cleverness** — the same verb, separator, and
   layout are used across all runs.
5. **Graceful degradation** — the output structure that handles success
   cleanly must also handle partial failure, total failure, and no-op
   runs without redesign.

Default Valkyrie Icon
---------------------

The default `app-icon` is the Valkyrie:

```
▗▄▄▖     ▗▄▄▖
▝▜██▄▄▄▄▄██▛▘
   ▝▜███▛▘
      █
```

Rendered using Unicode block and quadrant characters. The icon depicts
a winged figure descending — wingtips at the upper corners sweep down
and inward to a central body, with a small base anchoring the
composition. The icon is four lines tall and twelve columns wide.

When rendered in the banner, it is indented three spaces from the left
frame:

```
│   ▗▄▄▖     ▗▄▄▖
│   ▝▜██▄▄▄▄▄██▛▘
│      ▝▜███▛▘
│         █
```

Applications may override `CliConfig::$appIcon` with any multi-line
string. Custom icons should be 3–6 lines tall and roughly 8–14 columns
wide to fit cleanly in the banner frame.

Banner Structure
----------------

The banner has three logical parts:

**Top row** — `╭── {app-name} v{app-version}` — identifies the
application by name and version. This is what the developer or operator
sees first and is the slot most worth customizing.

**Body** — contains the application icon, a blank line above and below,
and two context lines:

1. The runtime environment: `Running on Valkyrja v{version} · PHP
   {version}`
2. The project root path

**Bottom row** — `╰── {action-description} · {exact-command}` — names
the action narratively (e.g., "Serving application," "Migrating
database") and preserves the exact subcommand that produced it. The
description is set per-command by the command class; the exact command
is preserved from the user's invocation.

The bottom row serves triple duty: it closes the banner frame visually,
describes the action in human terms, and retains verifiable command
information.

Output Structure
----------------

Every Valkyrja CLI invocation produces output in three sections:

```
[Banner]

[Work Output]

[Summary]
```

A resolved-configuration block (as in Sindri) is optional and may be
included by individual commands when their resolved inputs are
non-obvious. Sections are separated by single blank lines.

### Work Output

The work output section displays each unit of work the command
performs, one per line, using dot-leader formatting with right-aligned
status labels:

```
Running migration 2026_01_15_000001_create_users_table..........Success
Running migration 2026_01_15_000002_add_email_index............Skipped
  ▸ Migration already applied
Running migration 2026_01_15_000003_create_posts_table.........Fail
  ▸ Foreign key constraint failed: users table not found
  ▸ See database/migrations/2026_01_15_000003_create_posts_table.php:24
```

**Rules:**

- Each step line starts with a verb and a subject (e.g., `Running
  migration X`).
- Dot leaders (`.....`) fill the space between the step description
  and the status label. Status labels are aligned to the same column
  across all steps.
- Status labels are single words: `Success`, `Skipped`, `Warning`,
  `Fail`.
- Detail lines are prefixed with `▸` and indented two spaces from the
  step line's starting column.
- A step with no detail lines has no detail output.
- Steps with detail lines are separated from the next step by a blank
  line; consecutive plain-`Success` steps stack without separation.

### Summary

The summary is a single line that reports completion time and per-status
counts of work units:

```
Completed in 0.4s · 2 succeeded
```

```
Completed in 0.4s · 1 succeeded · 1 skipped · 1 failed
```

**Rules:**

- Leading verb is `Completed` across all outcomes.
- Completion time is reported in seconds with one decimal place.
- Counts are separated by middle-dot (`·`).
- Zero-count categories are omitted from the summary.
- Categories always appear in the order: succeeded, skipped, warning,
  failed.
- Category labels: `succeeded` (the work was completed), `skipped`,
  `warning`, `failed`.

Status Semantics
----------------

Each of the four status labels has precise meaning:

| Status    | Meaning                                                 | Detail lines expected?                 |
|-----------|---------------------------------------------------------|----------------------------------------|
| `Success` | The step completed cleanly                              | No                                     |
| `Skipped` | The step was intentionally not performed                | Yes — explain why                      |
| `Warning` | The step completed but encountered a non-blocking issue | Yes — explain the issue                |
| `Fail`    | The step could not complete                             | Yes — explain and point to remediation |

**Summary counting:**

- `Success` → counted as `succeeded`
- `Warning` → counted as `warning` (and also as `succeeded`, since work was done)
- `Skipped` → counted as `skipped`
- `Fail` → counted as `failed`

In the mixed-outcome summary, a step that produced `Warning` is
reported in `warning` count to make non-clean outcomes visible.

Color (When Available)
----------------------

When a color-capable terminal is detected, the following color
assignments apply:

| Element                  | Color                       |
|--------------------------|-----------------------------|
| Banner frame characters  | Default                     |
| Application icon         | Default (or app-specified)  |
| `Success` status label   | Green                       |
| `Skipped` status label   | Gray / dim                  |
| `Warning` status label   | Yellow                      |
| `Fail` status label      | Red                         |
| Detail lines (`▸`)       | Parent status color, dimmed |
| Summary: `failed` count  | Red (when > 0)              |
| Summary: `warning` count | Yellow (when > 0)           |

Color is decorative. All information conveyed by color must also be
conveyed by the status label text, so non-color terminals (CI logs,
pipes, `NO_COLOR=1`) lose no information.

Exit Codes
----------

CLI exit codes are determined by the worst status in the work output:

| Worst status | Exit code |
|--------------|-----------|
| All Success  | `0`       |
| Skipped      | `0`       |
| Warning      | `0`       |
| Fail (any)   | `1`       |

A `Warning` exit code of `0` is deliberate — warnings are informational
and must not fail CI. Commands that treat warnings as errors can opt in
via a `--warnings-as-errors` flag or per-command configuration.

Customization via CliConfig
---------------------------

The `CliConfig` class is the single source of truth for application-level
CLI customization. All banner slots and several behavior toggles are
exposed as properties:

| Property            | Type     | Default                | Description                        |
|---------------------|----------|------------------------|------------------------------------|
| `$appName`          | `string` | `Valkyrja Application` | Top-row application name           |
| `$appVersion`       | `string` | `1.0.0`                | Top-row application version        |
| `$appIcon`          | `string` | Valkyrie (see above)   | Multi-line icon string             |
| `$projectRoot`      | `string` | Auto-detected          | Path shown in banner body          |
| `$showBanner`       | `bool`   | `true`                 | Print banner before command output |
| `$useColor`         | `bool`   | Auto-detected          | Force or suppress color output     |
| `$warningsAsErrors` | `bool`   | `false`                | Exit 1 on any warning              |

Applications instantiate and configure `CliConfig` during their bootstrap.
Properties not explicitly set fall through to the framework defaults
documented above.

**Example:**

```
namespace App;

use Valkyrja\Cli\Config\CliConfig;

return new CliConfig(
    appName: 'Acme Inc. API',
    appVersion: '1.2.3',
    appIcon: <<<ICON
       ▄▄▄
      ▐███▌
       ▀▀▀
    ICON,
);
```

Special Cases
-------------

### Happy path

All steps succeed. Summary is compact.

```
Running migration 2026_01_15_000001_create_users_table..........Success
Running migration 2026_01_15_000002_create_posts_table.........Success

Completed in 0.4s · 2 succeeded
```

### No-op run

Nothing needs doing.

```
Running migration 2026_01_15_000001_create_users_table..........Skipped
  ▸ Migration already applied
Running migration 2026_01_15_000002_create_posts_table.........Skipped
  ▸ Migration already applied

Completed in 0.1s · 2 skipped
```

### Total failure

All steps fail.

```
Running migration 2026_01_15_000001_create_users_table..........Fail
  ▸ Database connection refused
Running migration 2026_01_15_000002_create_posts_table.........Fail
  ▸ Database connection refused

Completed in 0.1s · 2 failed
```

### Mixed outcome

Realistic run with varied per-step results.

```
Running migration 2026_01_15_000001_create_users_table..........Success
Running migration 2026_01_15_000002_add_email_index............Skipped
  ▸ Migration already applied
Running migration 2026_01_15_000003_create_posts_table.........Fail
  ▸ Foreign key constraint failed: users table not found
  ▸ See database/migrations/2026_01_15_000003_create_posts_table.php:24

Completed in 0.4s · 1 succeeded · 1 skipped · 1 failed
```

Alignment Reference
-------------------

Column alignment rules used throughout Valkyrja CLI output:

- **Dot-leader status column:** all step lines pad to the same total
  width so status labels align on the right.
- **Detail line indent:** two spaces from the start of the step line,
  followed by `▸ ` and the detail content.
- **Banner body content:** three spaces from the left frame character
  (`│   `).
- **Custom icons:** preserve their original indentation; the framework
  prepends `│   ` to each line.

Relationship to Sindri Output
-----------------------------

Valkyrja's CLI output structure intentionally mirrors Sindri's design.
The banner template, status semantics, summary format, color
assignments, and exit codes are deliberately consistent across both
tools so that developers using Sindri to build a Valkyrja application
encounter the same visual language at every step of their workflow.

The key visual distinctions:

- **Sindri uses the Mjölnir icon** in its banner; Valkyrja-built
  applications use the Valkyrie icon by default.
- **Sindri's top-line identifies Sindri** (`Sindri v26.0.0`); a
  Valkyrja application's top-line identifies the application
  (`Acme Inc. API v1.2.3`).
- **Sindri's body line says "Running on Valkyrja"** because Sindri is
  built on Valkyrja; an application's body line also says "Running on
  Valkyrja" for the same reason. Valkyrja consistently appears as the
  substrate.
- **Sindri's "succeeded" category is labeled "written"** because
  Sindri's primary output is generated files; Valkyrja's "succeeded"
  category is labeled "succeeded" because applications can perform any
  kind of work and the count reflects task completion rather than file
  output.

Non-Goals
---------

This document does not specify:

- The internal architecture of Valkyrja's CLI output rendering code.
- The exact terminal dimensions Valkyrja CLI should target (the design
  is flexible across standard 80-column and wider terminals).
- Color palette values in specific ANSI codes (implementation detail;
  use standard terminal colors mapped to roles).
- Internationalization of status labels (future concern; currently
  English only).
- The full set of commands a Valkyrja application ships with (covered
  by per-command documentation).
