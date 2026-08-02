# Version Support

The support policy for each Valkyrja version line, and the runtime that each
language port needs.

[`VERSIONING.md`](VERSIONING.md) tells you how a version number is built and how
a release is cut. This document tells you how long a version line receives
patches, and which runtime version it runs on.

## The policy

Two rules set every date in this document:

- **Bug fixes.** The project patches a defect until three months after it
  releases the next major version. The window gives an application time to
  migrate before the previous version becomes unsupported.
- **Security fixes.** The project patches a vulnerability for two years from the
  initial release date of the major version. After that date, the project
  patches nothing.

Each version line therefore has a security-only tail. Version 27 releases in Q1
2027, so version 26 stops receiving bug fixes in Q2 2027. Version 26 keeps
receiving security fixes until Q1 2028. The tail is nine months.

Warning: an unsupported version receives no patch for a defect and no patch for a
vulnerability. Upgrade an application that runs an end-of-life version.

## Release and support schedule

Every language shares one version line, so every language shares these dates. A
Valkyrja 26 release is a Valkyrja 26 release in PHP, in Java, and in every other
port.

| Version | Release             | Bug Fixes Until | Security Fixes Until |
|:--------|:--------------------|:----------------|:---------------------|
| 25 (*)  | December 11th, 2025 | March 31, 2026  | March 31, 2026       |
| 26      | March 31, 2026      | Q2 2027         | Q1 2028              |
| 27      | Q1 2027             | Q2 2028         | Q1 2029              |
| 28      | Q1 2028             | Q2 2029         | Q1 2030              |

(*) Version 25 was a pre-release line, and PHP is the only port that has one. It
is no longer supported, because version 26 has shipped.

## Runtime requirements

Each port names the runtime versions that it supports. The runtime is the only
column that differs between the ports.

| Version | PHP       | Java    | Node (TypeScript) | Go    | Python |
|:--------|:----------|:--------|:------------------|:------|:-------|
| 25 (*)  | 8.4 – 8.6 | —       | —                 | —     | —      |
| 26      | 8.4 – 8.6 | 21 – 25 | 22+               | 1.26+ | 3.14+  |
| 27      | 8.5 – 8.6 | 23 – 25 | (**)              | (**)  | (**)   |
| 28      | 8.6+      | 25+     | (**)              | (**)  | (**)   |

(*) PHP is the only port with a version 25 line. Every other port starts at
version 26.

(**) The port has not declared a runtime range for this version yet. Version 26
is the first line for TypeScript, Go, and Python, so a later version gets its
range when the port declares one. Do not read the empty cell as "unsupported" —
read it as "not yet decided".

Kotlin has no row. The Kotlin port is planned and it has no repository, so it
has no release and no declared runtime. See [`AGENTS.md`](AGENTS.md) §1 for the
port list and the status of each port.

## Where each port publishes this

Each framework repository repeats the schedule in its own `README.md`, because a
consumer reads the repository, not this document. PHP also carries it in
`src/Valkyrja/VERSIONING_AND_RELEASE_PROCESS.md`, and Java carries it in
`src/main/java/io/valkyrja/VERSIONING_AND_RELEASE_PROCESS.md`.

Warning: a date that changes here must change in every one of those files in the
same pull request. A stale copy tells a consumer that a version still receives
patches when it does not.
