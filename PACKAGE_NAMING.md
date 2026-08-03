# Package and Namespace Naming

A repository name and a package name are two different names. This document
defines the package name, the registry namespace, and the source namespace for
every language.

`REPOSITORY_NAMING.md` in the `.github` repository defines the repository name.
Read that document first. This one starts where it stops.

---

## The core rule

**A repository name organizes the GitHub organization. A package name states
what the package is.**

A repository name carries a category prefix and a language suffix, because the
organization holds every category and every language in one list.
`REPOSITORY_NAMING.md` explains why.

A registry does not. Maven Central holds one Java, Packagist holds one PHP, and
npm holds one TypeScript. The language is already the registry, and the category
is already the namespace. Both therefore come off:

```
ci-phpcsfixer-php    →  valkyrja/phpcsfixer
      ^^^^^^^^^          drop `ci-`, drop `-php`
```

```
sindri-java          →  io.valkyrja:sindri
                         drop `-java`
```

Take the repository name, remove the category prefix, and remove the language
suffix. What remains is the package name.

---

## Registry namespace and package name

Each language publishes to one registry, and each registry has one namespace
for the organization.

| Language   | Registry                                  | Namespace      | Example package        |
| ---------- | ----------------------------------------- | -------------- | ---------------------- |
| PHP        | Packagist                                 | `valkyrja/`    | `valkyrja/valkyrja`    |
| Java       | Maven Central                             | `io.valkyrja:` | `io.valkyrja:valkyrja` |
| TypeScript | npm                                       | `@valkyrjaio/` | `@valkyrjaio/valkyrja` |
| Go         | the module path — see the exception below |
| Python     | PyPI                                      | planned        | planned                |

Warning: the PHP namespace is `valkyrja` and the npm scope is `@valkyrjaio`.
The two differ because npm scopes take the organization handle. Do not
"correct" either one.

### The names in use

| Repository                 | Package                   |
| -------------------------- | ------------------------- |
| `valkyrja-php`             | `valkyrja/valkyrja`       |
| `valkyrja-java`            | `io.valkyrja:valkyrja`    |
| `valkyrja-ts`              | `@valkyrjaio/valkyrja`    |
| `sindri-php`               | `valkyrja/sindri`         |
| `sindri-java`              | `io.valkyrja:sindri`      |
| `sindri-ts`                | `@valkyrjaio/sindri`      |
| `valkyrja-starter-app-php` | `valkyrja/application`    |
| `valkyrja-starter-app-ts`  | `@valkyrjaio/application` |
| `ci-phparkitect-php`       | `valkyrja/phparkitect`    |
| `ci-phpcodesniffer-php`    | `valkyrja/phpcodesniffer` |
| `ci-phpcsfixer-php`        | `valkyrja/phpcsfixer`     |
| `ci-phpstan-php`           | `valkyrja/phpstan`        |
| `ci-phpunit-php`           | `valkyrja/phpunit`        |
| `ci-psalm-php`             | `valkyrja/psalm`          |
| `ci-rector-php`            | `valkyrja/rector`         |
| `ci-spotless-java`         | `io.valkyrja:spotless`    |

A starter application publishes as `application`, because that is what the
package is. The repository name records that it is a starting point, and the
package name does not repeat it.

---

## Two exceptions

### Go takes the whole repository URL

A Go module path is a URL that the toolchain fetches. It therefore states the
repository name in full, including the language suffix:

```go
module github.com/valkyrjaio/valkyrja-go
```

This is not a deviation to correct. `go get` resolves the path, so the path must
name the repository. Go has no registry namespace of its own.

### A project template puts the language first

```
project-template-php  →  valkyrja/php-template
project-template-ts   →  @valkyrjaio/ts-template
```

These keep the language and move it to the front. A project template is not a
language-specific build of one project; it is a distinct template per language,
and the name reads that way. This is the one place a package name states a
language.

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

The package name then matches that last segment, so `io.valkyrja.spotless`
publishes as `io.valkyrja:spotless`. Keep the two in step.

Warning: `sindri-java` publishes as `io.valkyrja:sindri` and its source root is
`io.sindri`. The group id names the organization that publishes the package, and
the source root names the project. The two answer different questions, so they
do not have to match.

---

## What not to do

- **Do not put the language in the package name.** `io.valkyrja:spotless-java`
  states the language twice, because Maven Central holds only Java. Go is the
  exception, and only because the path is a URL.

- **Do not put the `ci-` prefix in the package name.** `ci-` marks a category of
  repository. A consumer that adds `valkyrja/phpstan` already knows it is
  adding a PHPStan configuration.

- **Do not name the package after the repository out of habit.** Every name in
  the table above was chosen. `ci-spotless-java` was published once as
  `io.valkyrja:ci-spotless-java`, which took both the prefix and the suffix, and
  it had to be republished as `io.valkyrja:spotless`.

- **Do not assume a rename is available.** Read the next section first.

---

## A published package name cannot be corrected

Maven Central, Packagist, and npm do not let a package change its name, and they
do not let a published version be removed. A wrong name is permanent. The only
remedy is to publish under the right name and leave the wrong one where it sits.

`io.valkyrja:ci-spotless-java:26.0.0` is such a package. It stays on Maven
Central, it receives no further release, and every consumer moved to
`io.valkyrja:spotless`.

**Check the name before the first release**, and check it against the table
above rather than against the repository name.

---

## Where each name is stated

| Language   | File               | Field                                                                             |
| ---------- | ------------------ | --------------------------------------------------------------------------------- |
| PHP        | `composer.json`    | `name`                                                                            |
| Java       | `build.gradle.kts` | `mavenPublishing { coordinates }` and `rootProject.name` in `settings.gradle.kts` |
| TypeScript | `package.json`     | `name`                                                                            |
| Go         | `go.mod`           | `module`                                                                          |

Warning: Java states the name twice. `coordinates(...)` sets what Maven Central
receives, and `rootProject.name` sets what the built jar is called. A change to
one without the other publishes `spotless` as a jar named `ci-spotless-java`.

```kotlin
// settings.gradle.kts
rootProject.name = "spotless"
```

```kotlin
// build.gradle.kts
mavenPublishing {
    coordinates(group.toString(), "spotless", version.toString())
}
```

---

## Related

- `REPOSITORY_NAMING.md` in the `.github` repository — the repository name
- `COPYRIGHT_HEADER.md` in the `.github` repository — the package identifier
  that the license header states, which is a fourth name and is prose
  (`Valkyrja Spotless`) rather than a coordinate
- [`VERSIONING.md`](VERSIONING.md) — the version that follows the name
