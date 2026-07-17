# GO

## TODOs

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
