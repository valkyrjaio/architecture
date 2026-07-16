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
