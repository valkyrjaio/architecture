# STRUCTURE.md — the structure taxonomy

The **cross-language** convention for the structure taxonomy. It applies in
every Valkyrja repository, and to every contributor, human or agent.

A class's _kind_ is encoded three ways at once — its **name suffix**, the
**segment** (namespace/package/directory) it lives in, and its **modifier** —
and all three must agree. This is the machine-verified spec (PHP's PHPArkitect
`Rules` is the reference; Java ArchUnit and Kotlin Konsist mirror it; where a
language has no architecture linter — Go, Python, TypeScript — it is enforced
in review). PHP segment spellings are shown; **each Layer-2 guide gives the
per-language spelling** (case + reserved-word handling + constructs a language
lacks).

## The taxonomy

| Kind                                                   | Identified by                                | Name                               | Segment                        | Modifier  |
| ------------------------------------------------------ | -------------------------------------------- | ---------------------------------- | ------------------------------ | --------- |
| Contract                                               | is an interface                              | `*Contract`                        | `Contract\`                    | interface |
| Service provider                                       | implements `ServiceProviderContract`         | `*ServiceProvider`                 | `Provider\`                    | —         |
| Component provider                                     | implements `ComponentProviderContract`       | `*ComponentProvider`               | `Provider\`                    | —         |
| Route provider                                         | implements `Http`/`CliRouteProviderContract` | `*RouteProvider`                   | `Provider\`                    | —         |
| Listener provider                                      | implements `ListenerProviderContract`        | `*ListenerProvider`                | `Provider\`                    | —         |
| Factory                                                | —                                            | `*Factory`                         | `Factory\`                     | —         |
| Constant                                               | —                                            | `*Constant`                        | `Constant\`                    | final     |
| Attribute / annotation                                 | has the attribute marker                     | —                                  | `Attribute\`                   | —         |
| CLI command                                            | —                                            | `*Command`                         | `Cli\Command\`                 | —         |
| Security                                               | —                                            | `*Security`                        | `Security\`                    | final     |
| Concrete exception                                     | implements Throwable                         | `*Exception` (Go: `*Error`)        | `Exception\`                   | —         |
| Any throwable                                          | extends/implements Throwable                 | —                                  | `Throwable\`                   | —         |
| Base `*RuntimeException` / `*InvalidArgumentException` | —                                            | as-is                              | `Abstract\`                    | abstract  |
| Type / Model / Entity                                  | extends the base                             | —                                  | `Type\` / `Model\` / `Entity\` | —         |
| Abstract class                                         | is abstract                                  | must **not** contain `Abstract`    | `Abstract\`                    | abstract  |
| Enum                                                   | is an enum                                   | must **not** contain `Enum`        | `Enum\`                        | enum      |
| Trait                                                  | is a trait                                   | must **not** contain `Trait` (src) | `Trait\`                       | trait     |
| Test fixture                                           | reusable double in `Fixtures\`               | `*Fixture`                         | `Fixtures\`                    | final     |

## The relationships are bidirectional

Everything in `Contract\` must be an interface _and_ named `*Contract`;
everything in `Enum\` must be an enum; every final constant lives in
`Constant\`; and so on. For `Abstract`, `Enum`, and `Trait` the _segment_
carries the meaning, so the **name must not repeat it** — an abstract `Stream`
is `Abstract\Stream`, never `AbstractStream`.

```php
// Wrong — the class name repeats the segment.
namespace Valkyrja\Log\Logger\Abstract;

abstract class AbstractLogger implements LoggerContract {}
```

```php
// Right — the segment says "abstract", so the name does not.
namespace Valkyrja\Log\Logger\Abstract;

abstract class Logger implements LoggerContract {}
```

## Tests

Concrete test classes are `final`, live in `Unit\`/`Functional\`, and are named
`*Test`; reusable doubles live in `Fixtures\`, named `*Fixture`, never `*Test`.
A fixture that is itself an enum, trait, or contract keeps that type's naming
(`*Enum` / `*Trait` / `*Contract`) — the type rule supersedes the fixture
marker, just as the segment does for `Abstract`/`Enum`/`Trait` above.

## No author docblock

No class carries an `@author` docblock. The rule governs every class in every
segment, in `src` and in the tests alike.
