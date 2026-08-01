# Valkyrja

## Blog (Wordpress dupe, or Drupal Dupe)

- Similar to Data cache classes, this creates classes in the admin that hold
  custom page data, etc. Everything loads from a data cache file instead of from
  database. So much faster. Maybe blog posts can live in database because so
  many, so there should be a switch because at some point it'll be too much for
  the data file to handle

## Middleware settlement-stage naming — align across entry verticals

Every implemented port renamed the stage classes and their methods. Two documents
still name stage 7 `Terminated`.

- [valkyrjaio/architecture#129](https://github.com/valkyrjaio/architecture/issues/129)

## README alignment and audit — every port

A PHP audit found four groups of statements in the module READMEs that contradict the
implemented ports. Four pull requests corrected PHP. The other ports have not been
checked, and the same drift is likely, because the same documents seeded them.

- [valkyrjaio/architecture#138](https://github.com/valkyrjaio/architecture/issues/138)

The audit of this repository's own documents is tracked in
[valkyrjaio/architecture#111](https://github.com/valkyrjaio/architecture/issues/111).

## Cross-language work tracked as issues

These items appear in several language `TODO.md` files. Each has one parent issue here
and a child issue per repository.

- **VLID parity** — [architecture#134](https://github.com/valkyrjaio/architecture/issues/134)
- **Standalone `sindri` executable** — [architecture#135](https://github.com/valkyrjaio/architecture/issues/135)
- **Cross-language testing-gap audit** — [architecture#136](https://github.com/valkyrjaio/architecture/issues/136)
- **Repo description once gRPC and Queue land** — [architecture#137](https://github.com/valkyrjaio/architecture/issues/137)
