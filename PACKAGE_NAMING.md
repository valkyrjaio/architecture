# Package and Namespace Naming

A repository name and a package name are two different names. This document
defines the package name, the registry namespace, and the source namespace for
every language.

`REPOSITORY_NAMING.md` in the `.github` repository defines the repository name.
Read that document first. This one starts where it stops.

---

## The core rule

**Drop the `-{lang}` suffix. Keep everything else.**

A repository name carries a language suffix, because the organization holds
every language in one list. `REPOSITORY_NAMING.md` explains why.

A registry does not. Maven Central holds one Java, Packagist holds one PHP, and
npm holds one TypeScript, so the language is already the registry. The suffix
therefore comes off, and nothing else does:

```
ci-spotless-java    →  io.valkyrja:ci-spotless
sindri-java         →  io.valkyrja:sindri
valkyrja-java       →  io.valkyrja:valkyrja
ci-phpcsfixer-php   →  valkyrja/ci-phpcsfixer
```

The `ci-` prefix stays. It names what the package is — a CI tool
configuration — rather than how the organization files the repository. A
developer reading a dependency list learns something from it.

The seven PHP CI packages read `valkyrja/phpcsfixer`, `valkyrja/phpstan`, and so
on until August of 2026. Those names predate this convention. Each one now
publishes with the prefix, and Packagist marks the old name as abandoned and
names the new one as its replacement.

---

## Registry namespace and package name

Each language publishes to one registry, and each registry has one namespace
for the organization.

| Language   | Registry      | Namespace       | Example package                     |
| ---------- | ------------- | --------------- | ----------------------------------- |
| PHP        | Packagist     | `valkyrja/`     | `valkyrja/valkyrja`                 |
| Java       | Maven Central | `io.valkyrja:`  | `io.valkyrja:valkyrja`              |
| TypeScript | npm           | `@valkyrjaio/`  | `@valkyrjaio/valkyrja`              |
| Python     | PyPI          | none — flat     | `valkyrja`                          |
| Go         | none          | the module path | `github.com/valkyrjaio/valkyrja-go` |

Warning: the PHP namespace is `valkyrja` and the npm scope is `@valkyrjaio`.
The two differ because npm scopes take the organization handle. Do not
"correct" either one.

### PyPI has no namespace

PyPI holds one flat list of names, so a package cannot sit under the
organization. The organization name goes into the package name instead:

| Repository        | PyPI package       |
| ----------------- | ------------------ |
| `valkyrja-python` | `valkyrja`         |
| `sindri-python`   | `valkyrja-sindri`  |
| `ci-ruff-python`  | `valkyrja-ci-ruff` |

Warning: `sindri` on PyPI is taken. "Sindri Python SDK" by Sindri Labs owns it,
and PyPI does not release a held name. Sindri therefore publishes as
`valkyrja-sindri` on PyPI alone.

This cuts against `COPYRIGHT_HEADER.md`, which classes Sindri as a standalone
project and gives it no Valkyrja prefix. That intent holds for the copyright
identifier and for every other registry. PyPI's flat namespace does not respect
it, and the name was already gone.

### The names in use

| Repository                 | Package                      |
| -------------------------- | ---------------------------- |
| `valkyrja-php`             | `valkyrja/valkyrja`          |
| `valkyrja-java`            | `io.valkyrja:valkyrja`       |
| `valkyrja-ts`              | `@valkyrjaio/valkyrja`       |
| `sindri-php`               | `valkyrja/sindri`            |
| `sindri-java`              | `io.valkyrja:sindri`         |
| `sindri-ts`                | `@valkyrjaio/sindri`         |
| `sindri-python`            | `valkyrja-sindri`            |
| `valkyrja-starter-app-php` | `valkyrja/application`       |
| `valkyrja-starter-app-ts`  | `@valkyrjaio/application`    |
| `ci-phparkitect-php`       | `valkyrja/ci-phparkitect`    |
| `ci-phpcodesniffer-php`    | `valkyrja/ci-phpcodesniffer` |
| `ci-phpcsfixer-php`        | `valkyrja/ci-phpcsfixer`     |
| `ci-phpstan-php`           | `valkyrja/ci-phpstan`        |
| `ci-phpunit-php`           | `valkyrja/ci-phpunit`        |
| `ci-psalm-php`             | `valkyrja/ci-psalm`          |
| `ci-rector-php`            | `valkyrja/ci-rector`         |
| `ci-spotless-java`         | `io.valkyrja:ci-spotless`    |
| `ci-ruff-python`           | `valkyrja-ci-ruff`           |

A starter application publishes as `application`, not `valkyrja-starter-app`.
The repository name records that the application is a starting point, and the
package is simply the application. This is a name, not a suffix the rule
removes.

---

## Go is the outlier

Go has no package registry. `go get` fetches the module from its source, so the
module path **is** the GitHub repository URL, and it therefore states the
repository name in full — language suffix included:

```go
module github.com/valkyrjaio/valkyrja-go
```

This is not a deviation to correct. The path has to resolve, and the rule above
does not apply to a name the toolchain uses as an address.

---

## Source namespace

The source namespace is a third name, and it does not always match the package
name.

| Language   | Framework root         | Build tool root      |
| ---------- | ---------------------- | -------------------- |
| PHP        | `Valkyrja\`            | `Sindri\`            |
| Java       | `io.valkyrja`          | `io.sindri`          |
| Kotlin     | `io.valkyrja`          | `io.sindri`          |
| Go         | `io/valkyrja`          | `io/sindri`          |
| Python     | `valkyrja`             | `sindri`             |
| TypeScript | `@valkyrjaio/valkyrja` | `@valkyrjaio/sindri` |

A Java package that is not the framework takes a segment under the framework
root that names what it is:

```java
package io.valkyrja.spotless;
```

The source namespace drops the `ci-` prefix, because a Java package segment
names the tool rather than the repository category. `io.valkyrja.spotless`
therefore publishes as `io.valkyrja:ci-spotless`: the namespace and the package
name are related but not identical.

Warning: `sindri-java` publishes as `io.valkyrja:sindri` and its source root is
`io.sindri`. The group id names the organization that publishes the package, and
the source root names the project. The two answer different questions, so they
do not have to match.

---

## What not to do

- **Do not put the language in the package name.** `io.valkyrja:ci-spotless-java`
  states the language twice, because Maven Central holds only Java. Go is the
  outlier, and only because its module path is an address.

- **Do not remove the `ci-` prefix.** Only the language suffix comes off. The
  prefix states what the package is, and a dependency list is where that reading
  matters most.

- **Do not name the package after the repository out of habit.** `ci-spotless-java`
  was published once as `io.valkyrja:ci-spotless-java`, which kept the language
  suffix, and it had to be republished as `io.valkyrja:ci-spotless`.

- **Do not assume a rename is available.** Read the next section first.

---

## A published package name cannot be corrected

Maven Central, Packagist, and npm do not let a package change its name, and they
do not let a published version be removed. A wrong name is permanent. The only
remedy is to publish under the right name and leave the wrong one where it sits.

`io.valkyrja:ci-spotless-java:26.0.0` is such a package. It stays on Maven
Central, it receives no further release, and every consumer moved to
`io.valkyrja:ci-spotless`.

**Check the name before the first release**, and check it against the table
above rather than against the repository name.

---

## Where each name is stated

| Language   | File               | Field                                                                             |
| ---------- | ------------------ | --------------------------------------------------------------------------------- |
| PHP        | `composer.json`    | `name`                                                                            |
| Java       | `build.gradle.kts` | `mavenPublishing { coordinates }` and `rootProject.name` in `settings.gradle.kts` |
| TypeScript | `package.json`     | `name`                                                                            |
| Go         | `go.mod`           | `module` — the full repository URL                                                |
| Python     | `pyproject.toml`   | `project.name`                                                                    |

Warning: Java states the name twice. `coordinates(...)` sets what Maven Central
receives, and `rootProject.name` sets what the built jar is called. A change to
one without the other publishes `ci-spotless` as a jar named `spotless`.

```kotlin
// settings.gradle.kts
rootProject.name = "ci-spotless"
```

```kotlin
// build.gradle.kts
mavenPublishing {
    coordinates(group.toString(), "ci-spotless", version.toString())
}
```

---

## Related

- `REPOSITORY_NAMING.md` in the `.github` repository — the repository name
- `COPYRIGHT_HEADER.md` in the `.github` repository — the package identifier
  that the license header states, which is a fourth name and is prose
  (`Valkyrja Spotless`) rather than a coordinate
- [`VERSIONING.md`](VERSIONING.md) — the version that follows the name
