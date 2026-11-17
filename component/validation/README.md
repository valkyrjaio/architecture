# Validation

Implements
[`VALIDATION.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/validation/VALIDATION.md).

## Introduction

The Validation component checks a value against a rule, and reports what failed.
It owns the rules and the reporting. It owns nothing about where the value came
from.

A **rule** is an object. You construct it with the value it judges and the
message a person reads when that value fails. Every rule implements
`RuleContract`, so a set of rules over different values is one list.

A **validator** takes that list and runs every rule in it. The validator
collects the failures rather than stopping at the first one, so a person filling
a form sees every problem at once.

There is no adapter and no config. A rule is not a service an application swaps,
so nothing configures this component and nothing binds a default.

A protocol struct is what usually drives the validator. A struct declares its
fields and the rules each field satisfies, and the boundary runs those rules for
you. [Http](../http/README.md) covers the struct.

## A rule is an object, not a string

| Method       | Does                                   | Takes   | Returns   |
| ------------ | -------------------------------------- | ------- | --------- |
| `getSubject` | returns the value the rule is about    | nothing | the value |
| `isValid`    | answers whether the value satisfies it | nothing | a boolean |
| `validate`   | fails loudly when the value does not   | nothing | nothing   |

No method takes the subject, because the rule already holds it. A rule that
needs no other component takes the subject first, then whatever it compares
against, and the message last.

```text
// The subject, the length to compare against, and the message.
rule = new Min(fullName, 2, "A full name is too short")

// The rule hands back the value it judges.
subject = rule.getSubject()
```

A rule written as a string is what this component does not do. Such a rule packs
its arguments into a name. A compiler cannot check that name, a reader cannot
discover it, and something has to parse the arguments back out.

### Two forms of the same question

`isValid` answers the question. `validate` throws the component's rule-failure
throwable, carrying the message you gave the rule.
[Throwable](../throwable/README.md) covers the throwable contracts.

Use `isValid` where your code branches on the answer. Use `validate` where there
is nothing to branch on, and a failure ends the unit of work.

```text
// Branch on the answer.
if (! new Email(address, ErrorMessage.IS_EMAIL).isValid()) {
    return response.badRequest()
}

// Fail loudly. The caller up the stack handles the throwable.
new Email(address, ErrorMessage.IS_EMAIL).validate()
```

The validator wants the answer, so the validator runs `validate` and catches the
failure. That is how it reaches the message and keeps going.

### The shipped messages

`ErrorMessage` holds one message per shipped rule, so an application that wants
no wording of its own passes a constant.

```text
new Required(fullName, ErrorMessage.REQUIRED)

// Pass your own text where the person needs more than the general wording.
new Required(fullName, "A full name is required")
```

## The rules every port starts from

Each group holds the rules for one kind of value, so a reader looking for a rule
about text looks in one place.

Which rules ship is the port's own choice, because a language brings its own
formats and types. A port adds a row for each rule it ships and the others do
not.

Warning: a rule for one kind of value fails a subject of another kind, rather
than converting it. A number rule fails the text `"25"`, and a text rule fails
the number `25`. Convert the subject before you hand it to the rule.

### The rules that answer a question about a value

| Rule        | Passes when                      | Takes besides the subject |
| ----------- | -------------------------------- | ------------------------- |
| `Required`  | the subject is truthy            | nothing                   |
| `NotEmpty`  | the subject is not empty         | nothing                   |
| `IsEmpty`   | the subject is empty             | nothing                   |
| `Equal`     | the subject equals a value       | the value                 |
| `NotEqual`  | the subject differs from a value | the value                 |
| `IsString`  | the subject is text              | nothing                   |
| `IsNumeric` | the subject is a number          | nothing                   |
| `IsBool`    | the subject is a boolean         | nothing                   |
| `Email`     | the subject is an email address  | nothing                   |

Among the rules that check a type, `IsNumeric` is the one that accepts both
forms. It passes the text `"25"` as well as the number `25`, so it reports
whether a value will convert.

### The rules for text

| Rule         | Passes when                            | Takes besides the subject |
| ------------ | -------------------------------------- | ------------------------- |
| `Min`        | the subject is at least that long      | a length                  |
| `Max`        | the subject is at most that long       | a length                  |
| `Contains`   | the subject holds the other text       | the text to find          |
| `StartsWith` | the subject starts with the other text | the text to find          |
| `EndsWith`   | the subject ends with the other text   | the text to find          |
| `Alpha`      | the subject holds letters only         | nothing                   |
| `Lowercase`  | the subject is lowercase               | nothing                   |
| `Uppercase`  | the subject is uppercase               | nothing                   |
| `Regex`      | the subject matches the pattern        | the pattern               |

### The rules for a number

| Rule          | Passes when                  | Takes besides the subject |
| ------------- | ---------------------------- | ------------------------- |
| `GreaterThan` | the subject is above a bound | the bound                 |
| `LessThan`    | the subject is below a bound | the bound                 |

### The rules that read a record

| Rule              | Passes when               | Takes besides the subject    |
| ----------------- | ------------------------- | ---------------------------- |
| `EntityExists`    | a matching record exists  | the entity, and a field name |
| `EntityNotExists` | no matching record exists | the entity, and a field name |

## Writing a rule of your own

A rule implements `RuleContract`. Extend the base `Rule` where this port ships
one, and implement `isValid` alone. The base holds the subject, holds the
message, and turns a `false` answer into the failure.

```text
class StrongPassword extends Rule {

    isValid() {
        return isString(this.subject)
            && length(this.subject) >= 12
            && matches(this.subject, "/[0-9]/")
    }
}
```

A rule that compares against something and needs no other component declares a
constructor of its own. That constructor takes the subject first, the value the
rule compares against next, and the message last. It hands the subject and the
message to the base, and holds the rest itself.

```text
class DivisibleBy extends Rule {

    construct(subject, divisor: int, errorMessage: string) {
        super(subject, errorMessage)

        this.divisor = divisor
    }

    isValid() {
        return isInt(this.subject)
            && this.subject % this.divisor == 0
    }
}
```

A check written inline at a call site is none of the things a rule class is. A
rule is testable on its own, discoverable by name, and reusable.

### A rule that reaches another component

A rule that needs another component takes that component's contract first, ahead
of the subject. The message comes third, ahead of the values the rule compares
against. The rule lives in a group named for that component, so the dependency
is visible in the path.

```text
sponsorId = request.getParsedBody().get("sponsorId")

sponsorExists = new EntityExists(
    // The contract the Orm component binds, which the caller holds.
    orm,
    sponsorId,
    ErrorMessage.ENTITY_EXISTS,
    Member,
    "id",
)
```

`EntityExists` reads the record through the component that stores records
([Orm](../orm/README.md)). The field name is optional, and the rule uses the
entity's own id field where you leave it out.

## Running a set of rules

Group the rules by the subject they judge, and give the map to a validator.

| Method                 | Does                                      | Returns               |
| ---------------------- | ----------------------------------------- | --------------------- |
| `setRules`             | gives the validator the rules to run      | nothing               |
| `isValid`              | runs them, and reports whether all passed | a boolean             |
| `getErrorMessages`     | returns the failures, keyed by subject    | a map                 |
| `hasFirstErrorMessage` | reports whether there is at least one     | a boolean             |
| `getFirstErrorMessage` | returns the first failure's text          | a message, or nothing |

```text
validator = new Validator({
    "fullName": [
        new Required(fullName, ErrorMessage.REQUIRED),
        new Max(fullName, 60, "A full name is too long"),
    ],
    "email": [
        new Required(email, ErrorMessage.REQUIRED),
        new Email(email, ErrorMessage.IS_EMAIL),
    ],
})

if (! validator.isValid()) {
    return response.badRequest(validator.getErrorMessages())
}
```

You construct a validator where the values are, because it holds the rules for
one unit of work. `setRules` replaces the set on a validator you already hold.

Warning: `getErrorMessages`, `hasFirstErrorMessage` and `getFirstErrorMessage`
report nothing until `isValid` runs. A caller that reads the map first finds it
empty, and reads that as valid data. Run `isValid`, then read.

Each entry in the map names its subject, then carries the failing rule's
message. An entry under `fullName` reads as that subject, a colon, a space, and
the text you gave the rule.

### Read every failure, not the first

`isValid` runs every rule before it answers. Report the whole map, so a person
corrects every mistake in one pass.

```text
// Wrong. The person submits the form once per mistake.
return response.badRequest(validator.getFirstErrorMessage())

// Right. Every problem reaches the person at once.
return response.badRequest(validator.getErrorMessages())
```

The first-message methods are for the caller that shows one thing. A check over
a single field is one, and so is an endpoint that reports one reason. Ask
`hasFirstErrorMessage` before you read the message.

### One message per subject

The validator keys the failures by subject, so one subject reports one message.
A later failing rule for one subject replaces the message the earlier rule
stored.

```text
// Both rules fail. The map holds one entry for "fullName", from the last failure.
validator = new Validator({
    "fullName": [
        new Max(fullName, 60, "A full name is too long"),
        new Regex(fullName, "/^[A-Za-z ]+$/", "A full name holds letters and spaces only"),
    ],
})
```

The last failure for a subject is the one that survives. So put the rule whose
message helps a person most at the end of that subject's list. A caller that
needs every rule's complaint for one subject runs those rules itself, and reads
each failure as it happens.

## Prefer a struct over the validator

A struct declares its fields, and returns the rules each field satisfies. The
boundary then reads as a declaration, and the validator runs underneath the
struct. [Http](../http/README.md) covers the struct.

```text
enum CreateMemberStruct implements RequestStructContract {

    // The mixin for the request part this struct reads.
    use ParsedBodyRequestStruct

    case fullName
    case email

    static getValidationRules(request) {
        parsedBody = request.getParsedBody()

        fullName = parsedBody.get(self.fullName.name)
        email = parsedBody.get(self.email.name)

        return {
            self.fullName.name: [
                new Required(fullName, ErrorMessage.REQUIRED),
                new Max(fullName, 60, "A full name is too long"),
            ],
            self.email.name: [
                new Required(email, ErrorMessage.REQUIRED),
                new Email(email, ErrorMessage.IS_EMAIL),
            ],
        }
    }
}
```

The stage that matched the route builds the validator from the struct, runs it,
and answers with a failed response where a rule failed. Nothing past that stage
sees an unchecked value.

Validate once, where untrusted data enters. A rule applied in two places is a
rule that will disagree with itself.

## Service registration

The component ships no service provider, and it binds nothing. You construct a
validator rather than resolve it, because it holds the rules for one unit of
work. [Container](../container/README.md) covers contracts, bindings and
providers.

```text
validator = new Validator(rules)
```

A rule that reaches another component takes that component's contract. The
caller resolves that contract, and passes it to the rule.

```text
class MemberController {

    // The Orm component's own manager contract.
    construct(orm: ManagerContract) { }

    store(request) {
        sponsorId = request.getParsedBody().get("sponsorId")

        validator = new Validator({
            "sponsorId": [
                new EntityExists(this.orm, sponsorId, ErrorMessage.ENTITY_EXISTS, Member),
            ],
        })

        if (! validator.isValid()) {
            return response.badRequest(validator.getErrorMessages())
        }
    }
}
```
