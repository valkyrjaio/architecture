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
- **Use real names from the framework.** Do not invent `Foo` or `Bar`.
- **Tag every fence with its language** — `php`, `java`, `ts`, `go`, `python`,
  `bash`. Highlighting and tooling read the tag.
- **Show PHP first**, because PHP is the reference implementation. Add a second
  example only for a language whose spelling differs. A Layer-2 guide shows only
  its own language.
- **Keep every example valid.** An example that does not compile teaches the
  wrong thing.
- **Write a generic example in a convention or design document.** Warning: a
  verbatim copy of real source drifts when the source changes. The document then
  asserts something that is false — the failure that rule 10 in
  [`AGENTS.md`](AGENTS.md) §3 describes for a comment. Show the shape in the
  framework's naming style, and copy no real method. A generic example still
  uses real framework names — the rule above forbids `Foo` and `Bar`.
- **Documentation about the code is the exception — it shows the real code.** In
  a component's `README.md`, and in a repo's own documents, the code is the
  subject. There the example must match the source.

The taxonomy rule in [`AGENTS.md`](AGENTS.md) §4 — "for `Abstract`, `Enum`, and
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

## When you edit an existing document

Rewrite the paragraph you touch, not the whole file. A large style-only rewrite
hides the change that the PR is about, and it makes the diff impossible to
review. A `docs:` PR may rewrite a full document for style alone, but then it
changes nothing else.
