# Validation

The Validation component: what it does, how it is used, and the names and
behavior every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Validation checks a value against a rule and reports what failed. It owns the
rules and the reporting, and nothing about where the value came from.

A protocol's struct is what usually drives it ([`HTTP.md`](../HTTP/HTTP.md)), so this
component is the thing a boundary uses rather than something application code
calls directly on every value.

---

## Hierarchy

| Subcomponent | Holds                                                 |
| ------------ | ----------------------------------------------------- |
| `Validator`  | the component contract: run rules, report failures    |
| `Rule`       | the rule contract, and one class per rule             |
| `Constant`   | the shared vocabulary                                 |
| `Throwable`  | the component's throwable contract and its exceptions |

There is no adapter and no config, because a rule is not a service an
application swaps.

---

## A rule is an object, not a string

Every rule implements one contract:

| Operation    | Does                                         |
| ------------ | -------------------------------------------- |
| `getSubject` | returns the value the rule is about          |
| `isValid`    | answers whether the value satisfies the rule |
| `validate`   | fails loudly when it does not                |

**Two forms of the same question, deliberately.** `isValid` answers; `validate`
fails. A validator running many rules wants the answer so it can collect every
failure, and a caller checking one condition wants the failure. The naming rule
makes which is which readable without the documentation
([`METHOD_NAMING.md`](../../convention/METHOD_NAMING.md)).

**A rule carries its own subject.** A rule is constructed with the value it
judges rather than taking it as an argument, which is what lets a set of rules
over different values be run as one list.

A rule expressed as a string — a name, with arguments packed into it — is what
this component deliberately does not do. A string is not checkable by a compiler,
not discoverable by a reader, and its arguments have to be parsed.

---

## The validator

| Operation              | Does                                      |
| ---------------------- | ----------------------------------------- |
| `setRules`             | gives the validator the rules to run      |
| `isValid`              | runs them, and reports whether all passed |
| `getErrorMessages`     | returns the failures, keyed by subject    |
| `hasFirstErrorMessage` | reports whether there is at least one     |
| `getFirstErrorMessage` | returns the first                         |

**The validator declares `isValid`, the same name a rule declares.** A caller
asks the validator whether the data is valid, and the validator consults its
rules to answer. The rules are not what the caller asks about, so a name such as
"validate the rules" would name the wrong subject. It would also promise a
failure rather than an answer.

**The validator runs every rule rather than stopping at the first failure.** A
person filling a form wants every problem at once, not one per submission. That is
the whole reason a validator exists rather than callers running rules themselves.

The failures are keyed by subject, so one subject reports one message: a later
failing rule for the same subject replaces the earlier one. A caller that needs
every rule's complaint for one subject reads them as the rules run.

The first-message operations exist for the caller that shows one thing — a
single-field check, or an API that reports one reason.

---

## Using it

### Validate at the boundary, once

Validate where untrusted data enters, and let everything past that point hold a
value already checked. A rule applied in two places is a rule that will disagree
with itself.

### Prefer a struct over calling the validator directly

A protocol struct declares its fields and the rules each satisfies, so the
boundary is declarative and the validator runs underneath
([`HTTP.md`](../HTTP/HTTP.md)). Reach for the validator directly only where there is no
struct to declare.

### Collect, then report

Run every rule and report every failure. Stopping at the first makes a person
submit a form once per mistake.

### Write a rule as a class

A new check becomes a rule class implementing the contract. That makes it
testable on its own, discoverable by name, and reusable — none of which a
condition written inline at a call site is.

---

## Permitted variation

| Variation                 | Reason                                                       |
| ------------------------- | ------------------------------------------------------------ |
| which rules ship          | a language's own formats and types differ                    |
| how a message is rendered | a port renders its own text, and an application overrides it |

The two contracts, their operations and the collect-then-report behavior do not
vary.
