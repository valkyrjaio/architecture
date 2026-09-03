# DOCUMENTATION_STYLE.md — Simplified Technical English

The **cross-language** convention for documentation prose. It applies in every
Valkyrja repository, and to every contributor, human or agent.

Write documentation prose in **ASD-STE100 Simplified Technical English (STE)**,
the controlled English of aerospace maintenance manuals. Two groups read these
documents: contributors whose first language is not English, and coding agents.
Both groups fail on the same input — long sentences, passive voice, ambiguous
pronouns, and one idea written two ways. STE removes that input.

STE has two parts: a set of writing rules, and a dictionary of approved words.
The dictionary is a licensed ASD document, so this project does not certify
against it. Apply the writing rules in this document, and apply the dictionary's
core discipline: **one word has one meaning, and one meaning has one word.**
Technical names (`ContainerContract`, `sindri`, `26.x`) and technical verbs
(`serialize`, `cache`) are always permitted — STE calls these Technical Names
and Technical Verbs.

## What the style governs

It governs prose that this project writes for a reader:

- Markdown documents — `AGENTS.md`, `README.md`, and every design document in
  the architecture repository.
- Doc comments in source — docblock, Javadoc, TSDoc, docstring.
- PR descriptions, issue text, and commit message text.
- Strings the framework prints to a person — CLI help and exception messages.

It does not govern:

- Code, identifiers, file paths, commands, and program output.
- Quoted material from another source. **Never edit a quote to make it
  compliant.**
- Fixed third-party text, such as the license header ([`AGENTS.md`](AGENTS.md)
  §5).

## The rules

1. **One instruction per sentence.** Keep an instruction to 20 words or fewer,
   and a description to 25 words or fewer.
2. **Use the active voice.** Name the actor. Write "`sindri` reads the config",
   not "the config is read".
3. **Use simple tenses.** Prefer the simple present. Do not use an `-ing` form as
   the verb of a sentence.
4. **Keep the articles.** Write "the provider", not "provider".
5. **Use three nouns or fewer in a noun cluster.** Break a longer cluster with a
   preposition or a relative clause.
6. **Use one term for one thing.** A contract is a "contract" in every document;
   do not call it an "interface" in the next paragraph.
7. **Repeat the noun instead of a pronoun** when more than one noun could be the
   referent.
8. **One topic per paragraph, six sentences or fewer.** Put three or more
   conditions or steps in a vertical list.
9. **Put the warning before the instruction.** State what breaks, then state what
   to do.
10. **No slang, no idiom, no humor.** They do not translate, and an agent reads
    them as fact.
11. **Write a full sentence. Do not splice clauses with an em dash.** Give a
    dash-joined afterthought its own sentence.

Rules 1, 2, and 7 carry the most weight. This paragraph breaks all three:

> Because the generated cache must mirror what reflection produces exactly, and
> because a duplicate registration is considered to be the developer's error
> rather than something that ought to be silently repaired by the framework,
> middleware is appended by both the runtime collector and `sindri` without any
> deduplication step being applied to it.

The same content in STE:

> The runtime collector and `sindri` append each middleware in order. Neither one
> dedupes. A duplicate registration is the developer's error. The framework does
> not correct the duplicate, because the generated cache must match reflection
> exactly.

**A shorter document is not the goal.** STE often makes a dense paragraph longer,
because one run-on sentence becomes three plain ones. Count re-reads, not words.
A paragraph that a reader understands on the first pass beats a shorter paragraph
that the same reader must parse twice.

This holds for a decision log too, where compression is most tempting. A reader
opens a decision log to answer one question: what did we consider, and why did we
reject it? A run-on sentence hides exactly that answer. **No section of a
document is exempt** — a design record that a reader consults years later has the
most to gain.

## Full sentences instead of the em dash

An em dash joins two clauses into one sentence. The reader holds the first
clause while the second one arrives. Rule 1 removes that load. Write the second
clause as its own sentence. A sentence that wants a dash-joined afterthought
wants two sentences.

The em dash stays where it is the correct tool:

- A table cell, where the dash marks an empty value.
- A label and the gloss after it, in a heading, a list item, or an example
  marker.
- An interruption that grammar requires, and that a comma cannot carry.

> Wrong — one sentence carries three spliced clauses:
>
> The container resolves the binding on the first call — every later call reads
> the cached instance — and a provider registers the binding before the first
> call.

> Right — each clause is a sentence:
>
> The container resolves the binding on the first call. Every later call reads
> the cached instance. A provider registers the binding before the first call.

## Select before you write

STE governs how you write a sentence. This rule governs which sentences you
write. Include a sentence only when it changes what the reader does or decides.
A document is not more complete because it is longer — every sentence the reader
must skip hides the sentence the reader needs.

Cut these before they reach the page:

- **Process narration.** "This was found after the code was written", "every
  example was checked before commit" — state the conclusion, not the journey.
- **Restatement of the artifact.** A PR description does not re-teach the
  section it adds. A comment does not repeat the code below it. A summary does
  not restate the diff.
- **Background the reader can derive.** The diff, the file, and the linked
  document already carry it.
- **The evidence for a rule.** A guide states the rule. The case that produced
  the rule is context, and a reader does not act on it.

Warning: evidence goes stale before the rule does. A rule that carries its own
evidence sends a reader to check the evidence, and the document then states
something that is no longer true. State the rule, and keep the case that
produced it in the pull request description.

**Add context only when all three of these are true:**

- A reader needs the context.
- The context is relevant to the rule.
- Someone asks for the context.

> Wrong — the rule carries its evidence, and the evidence names a port status
> that changes:
>
> A port that carries a whole component in one pull request is too large to
> review. The Go port shows the cost. One Go pull request added every component
> at once, and it closed without a merge.

> Right — the rule states what breaks, and nothing else:
>
> A port that carries a whole component in one pull request is too large to
> review. A reviewer cannot read the full diff with care, so the review reports
> style and misses the design.

"A shorter document is not the goal" (above) and this rule do not conflict. That
rule forbids compressing a needed sentence into a dense one. This rule forbids
writing a sentence nobody needs. Write every needed sentence plainly, and no
others.

## Code examples

Every rule that has a code shape gets a code example. Prose states the rule; the
example shows it. A reader who does not yet know the rule must be able to copy
the example and be correct.

- **State the rule in prose first.** An example never replaces the rule. The
  prose carries the reason, and the reason is what a reader needs to apply the
  rule to a case the example does not cover.
- **One idea per example, 20 lines or fewer.** Split a longer example.
- **Show the wrong form, then the right form**, when a rule is easy to break.
  Mark each form, and say in a comment why it is wrong or right.
- **Use real names from the framework.** Do not invent `Foo` or `Bar`. Verify
  every name in the source tree before you write it. See
  [Verify every name](#verify-every-name).
- **Tag every fence with its language** — `php`, `java`, `ts`, `go`, `python`,
  `bash`. Highlighting and tooling read the tag.
- **Show PHP first**, because PHP is the reference implementation. Add a second
  example only for a language whose spelling differs. A Layer-2 guide shows only
  its own language.
- **Keep every example valid.** An example that does not compile teaches the
  wrong thing.
- **Write an example, never a copy of the source.** See
  [An example, never a copy](#an-example-never-a-copy).

The taxonomy rule in [`STRUCTURE.md`](STRUCTURE.md) — "for `Abstract`, `Enum`, and
`Trait` the segment carries the meaning, so the name must not repeat it" — takes
this example:

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

## An example, never a copy

**A document does not hold a copy of an implementation. A document shows usage,
and it states what the code does.**

A copy of the source does neither. The copy drifts when the source moves, and the next
reader trusts it. Rule 10 in [`AGENTS.md`](AGENTS.md) §3 describes the same
failure for a comment.

This rule holds in every document. A convention document, a design document, and
a component's `README.md` each obey it.

### The test

Read the block, then look for the same lines in the source tree.

- A real file holds the same body. The block is a **copy**. Remove it.
- The block quotes a declaration, and the next section permits the declaration.
  The block is the **API surface**. Keep it.
- No source file holds the block, and the block shows a caller how to use the
  API. The block is an **example**. Keep it.
- A rule about the copy marks the block as the wrong form. The block is a
  **demonstration**. Keep it, and keep it short.

The source tree decides, and the writer's intent does not. A block that a writer
meant as an example is still a copy when a real file holds the same body.

### The API surface is not a body

A document names the real API. A document never reproduces the implementation.

| A document shows this                    | A document never copies this            |
| ---------------------------------------- | --------------------------------------- |
| A class name, a method name, a namespace | A method body                           |
| A contract declaration                   | A constructor body                      |
| A method signature                       | A property list that holds real values  |
| An enum's cases                          | A real class that implements a contract |

A contract declares what a caller programs against, so a document quotes the
contract. A body is how the framework satisfies the contract today, so a
document states what the body guarantees.

### Name the class, do not reproduce it

A reader who needs the real code opens the file. Name the class, state what the
class does, and stop there.

```php
// Wrong — the document copies a real class to say which commands it registers.
class CliRoutingCliRouteProvider implements CliRouteProviderContract
{
    public function getControllerClasses(): array
    {
        return [
            HelpCommand::class,
            // ... the remaining built-in command classes
        ];
    }
}
```

> Right — the prose names the class and states what the class does:
>
> `CliRoutingComponentProvider::getCliProviders()` returns
> `CliRoutingCliRouteProvider`, which lists the built-in command classes and
> carries a static handler for each.

### Keep what the block conveys

Warning: a removal that drops information is worse than the copy. State what the
block conveys before you remove the block. The replacement carries the same
information.

A copied class body usually conveys two things. It shows the shape of a
contract, and it shows which real class satisfies the contract. A caller-facing
example carries the first. Prose that names the class carries the second.

### Verify every name

Warning: an invented name fails the way a copy fails. A name that no source file
holds, and a name that contradicts another document, each assert something that
is false.

- Confirm a real name in the source tree before you write the name.
- Invent a name only for the caller's own code, such as an application class or
  an application contract.
- Reuse the invented names that this project established already.
  [`CONTAINER.md`](CONTAINER.md) holds `NotifierContract`, `SlackNotifier`, and
  `TeamsNotifier`. A second document uses those names, and it does not redefine
  them.
- Never write `Foo` or `Bar`.

## When you edit an existing document

Rewrite the paragraph you touch, not the whole file. A large style-only rewrite
hides the change that the PR is about, and it makes the diff impossible to
review. A `docs:` PR may rewrite a full document for style alone, but then it
changes nothing else.
