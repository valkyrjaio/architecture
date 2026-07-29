# GO

## TODOs

### Cross-language testing-gap audit

Compare this port's test suite against the other languages' and either close each
difference or record it in `AGENTS.md` as deliberate. Out of scope for the work
that prompted it; tracked here so it is not lost.

Prompted by a concrete miss: `sindri-java`'s golden snapshot never exercised the
dynamic-route regex path, so a framework regex-format change rode through a
dependency bump silently, while `sindri-ts` caught the equivalent change at once
([sindri-java#54](https://github.com/valkyrjaio/sindri-java/pull/54)). Only that
single gap was checked across ports — nobody has compared the suites broadly.

What to look for:

- Behavior a sibling port asserts that this one does not.
- An assertion pinned to a fragment where a sibling pins the whole value — a
  fragment survives the framing around it changing, which is exactly how the miss
  above happened.
- A generator, adapter, or component carrying snapshot/branch coverage on one
  side and none here.
- Test tooling that differs: what the static analyzers and formatters actually
  cover, and whether the test tree is inside or outside that scope.

Not every difference is a defect — some are forced by the language. Read the
per-language notes in `AGENTS.md` before "aligning" anything; the dynamic route
regex framing is the worked example (PHP must keep its PCRE delimiters, every
other port must not).

Known starting point: the port is a proof of concept and `http/routing` is still
empty, so this is mostly forward-looking — the goal is that routing lands already
carrying the tests the other ports have, rather than needing an audit afterwards.

### Reset the reused deps branch in the update-dependencies workflow

`_go-update-dependencies.yml` reuses an open `deps/update-dependencies-*` branch across runs,
committing each run's updates on top of the previous run's tree. Nothing a run writes to that
branch can be walked back by a later run, so a bad version pinned once stays pinned until
somebody deletes the branch by hand.

Java hit this concretely. An unfiltered root `dependencyUpdates` report bumped
`io.netty:netty-codec-http` to the 2015-era `5.0.0.Alpha2` prerelease and turned CI red
([starter-app-java#53](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/53)). Fixing
the filter ([#54](https://github.com/valkyrjaio/valkyrja-starter-app-java/pull/54)) did **not**
repair the branch: `useLatestVersions` only ever upgrades, so the re-run reported the alpha as
merely "exceeding" the latest and changed nothing. The PR had to be closed and the branch deleted.
The Java workflow now resets the reused branch to the base commit before running the updates
([.github#151](https://github.com/valkyrjaio/.github/pull/151)), so each run recomputes from
scratch and the branch always holds exactly `base + today's updates`.

Confirm first that this port's updater can actually strand a bad version the way
`useLatestVersions` does — a tool that rewrites every constraint to "latest" on each run may
already self-correct, in which case record that in `AGENTS.md` instead of adding the reset.

Tradeoff to weigh: the reset discards any commit pushed onto the deps branch by hand.

### High priority — name test fixtures `fixtures`, not `classes`

**Cross-language change — mirror this in every port (Java, PHP, Python,
TypeScript) so the test trees stay 1:1.** Reusable test doubles/sample types live
under a `fixtures` package/dir, **not** `classes`. "Fixtures" is the widely-
understood term; "classes" is generic and reads oddly next to `unit`/`functional`
(and "classes" is an odd word in Go anyway). Go is not ported yet, so build it
this way from the start (no rename needed): put reusable doubles under a
`fixtures` test package, never `classes`.

### VLID — cross-language parity

**Cross-language change — mirror in every port (Java, PHP, Python, TypeScript).**
VLID (`Type/Vlid`) is PHP-only today; port it here (code + tests). It is the source
of the queue envelope `id` (a **VLID V1** — the longest, most-random version). Lock
cross-language parity:

- Port `Type/Vlid`, then add a conformance test: generate a VLID for **each version
  V1–V4** from a **fixed input timestamp**.
- Assert this port produces a byte-identical **non-random portion** vs the PHP
  fixture — the encoded **microsecond timestamp** and the **version digit at
  position 14** must match exactly. The random bits differ by design; exclude them.
- This gate prevents timestamp-encoding / version-digit-placement drift from
  silently breaking cross-language `id` interop.
