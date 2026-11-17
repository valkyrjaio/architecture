# Mail

Implements
[`MAIL.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/mail/MAIL.md).

## Introduction

The Mail component sends a message to people. It sends through an API, through a
mail server, to a log, or nowhere at all, and your application code reads the
same in every case.

An **adapter** is the thing that sends. Each one implements `MailerContract`,
and each is registered under its own name, so an application resolves whichever
one it wants.

`MailerContract` itself resolves to the **default** adapter: the one your config
names. It is the convenience, and it is what lets a development environment
record every message while production sends it, with no call site changing.

The component owns the message and the sending. It owns no templating. A body is
text the caller supplies, and an application that wants a template renders the
template first ([View](../view/README.md)).

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a mailer

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                             |
| -------------------------- | ----------------------------------------- | ------------------------------------- |
| `MailerContract`           | the adapter the config names              | environment                           |
| an adapter by its own name | that adapter, always                      | nothing                               |
| a contract of your own     | the adapter that contract's binding names | the kind of mail, and the environment |

### The default, through `MailerContract`

Use this where the integration is the application's business rather than this
call site's, which is almost everywhere.

```text
mailer = container.getSingleton(MailerContract)

mailer.send(message)
```

### A named adapter, where the integration is the point

Use this where you know this call site means that integration whatever the
environment is.

```text
mailer = container.getSingleton(MailgunMailer)

mailer.send(message)
```

### A contract of your own, per kind of mail

Declare a contract that extends `MailerContract` for a kind of mail, and bind it
to whichever adapter that kind wants. Then a password reset leaves through the
transactional integration, and a campaign leaves through the marketing one. Each
call site names the kind rather than the integration.

```text
interface TransactionalMailerContract extends MailerContract { }
interface MarketingMailerContract extends MailerContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`).
`SesMailer` below is an adapter you add yourself, which
[Adding an adapter](#adding-an-adapter) covers:

```text
class AppMailServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            TransactionalMailerContract: [self, "publishTransactionalMailer"],
            MarketingMailerContract: [self, "publishMarketingMailer"],
        }
    }

    static publishTransactionalMailer(container) {
        container.setSingleton(
            TransactionalMailerContract,
            container.getSingleton(MailgunMailer),
        )
    }

    static publishMarketingMailer(container) {
        container.setSingleton(
            MarketingMailerContract,
            container.getSingleton(SesMailer),
        )
    }
}
```

The publishers are where an environment decides what each kind of mail sends
through. A test binds both to the null adapter.

## Building a message

A message carries everything a send needs. Construct it with the sender, the
subject and the body, then add the rest.

A message is **immutable**. A reader returns the value, and a `with` form
returns a copy.

| Reads                    | Returns a copy through        | Holds                                               |
| ------------------------ | ----------------------------- | --------------------------------------------------- |
| `getFrom`                | `withFrom`                    | the sender, as one recipient                        |
| `getRecipients`          | `withAddedRecipient`          | who receives the message                            |
| `getReplyToRecipients`   | `withAddedReplyToRecipient`   | where a reply goes                                  |
| `getCopyRecipients`      | `withAddedCopyRecipient`      | who receives a visible copy                         |
| `getBlindCopyRecipients` | `withAddedBlindCopyRecipient` | who receives a copy the other recipients cannot see |
| `getAttachments`         | `withAddedAttachment`         | the files that travel with the message              |
| `getSubject`             | `withSubject`                 | the subject line                                    |
| `getBody`                | `withBody`                    | the body                                            |
| `isHtml`                 | `withIsHtml`                  | whether the body is markup                          |
| `getPlainBody`           | `withPlainBody`               | the text alternative to a markup body               |

`hasPlainBody` answers whether a text alternative is set.

Each set has an **added** form rather than a setter. `withAddedRecipient`
appends one recipient and keeps the ones already there.

```text
message = new Message(
    from: new Recipient("orders@example.com", "Orders"),
    subject: "Your order is on its way",
    body: "<p>Your order is on its way.</p>",
)

message = message
    .withAddedRecipient(new Recipient(order.email, order.name))
    .withAddedCopyRecipient(new Recipient("warehouse@example.com"))
    .withAddedBlindCopyRecipient(new Recipient("audit@example.com"))
    .withAddedReplyToRecipient(new Recipient("support@example.com", "Support"))
    .withIsHtml()
    .withPlainBody("Your order is on its way.")
    .withAddedAttachment(new Attachment("/var/invoices/4821.pdf", "invoice.pdf"))

mailer.send(message)
```

Warning: a `with` form leaves the message it was called on unchanged. Keep the
copy each call returns.

```text
// Wrong. The copy is discarded, so the message that is sent does not carry
// the new recipient.
message.withAddedRecipient(new Recipient(order.email, order.name))

mailer.send(message)
```

```text
// Right. The copy is what you send.
message = message.withAddedRecipient(new Recipient(order.email, order.name))

mailer.send(message)
```

### A recipient is an address and a name

Every recipient set holds recipients rather than addresses. A recipient is one
address and an optional name, so a call site passes one value rather than
formatting two.

| Method                   | Does                                 |
| ------------------------ | ------------------------------------ |
| `getEmail` / `withEmail` | reads or replaces the address        |
| `hasName`                | answers whether a name is set        |
| `getName` / `withName`   | reads or replaces the name beside it |

The name is optional. Construct a recipient with the address alone where you
have no name.

```text
named = new Recipient("ada@example.com", "Ada Lovelace")
unnamed = new Recipient("ada@example.com")
```

A loop builds a set one recipient at a time.

```text
for (customer in customers) {
    message = message.withAddedRecipient(
        new Recipient(customer.email, customer.name),
    )
}
```

### Give a markup body a plain alternative

`withBody` takes the body, and `withIsHtml` says that the body is markup. A
message whose body is markup carries the text alternative beside it, through
`withPlainBody`. An adapter sends both, so a client without markup support reads
the plain one.

Warning: a markup body with no plain alternative reaches a client that cannot
read it. Set the plain body wherever you set `withIsHtml`.

```text
message = message
    .withBody("<p>Your order is on its way.</p>")
    .withIsHtml()
    .withPlainBody("Your order is on its way.")
```

A plain-text message sets neither. `withIsHtml` means markup where you pass no
argument, and `withIsHtml(false)` returns a copy whose body is text again.

### An attachment is a path and a name

| Method                 | Does                                          |
| ---------------------- | --------------------------------------------- |
| `getPath` / `withPath` | reads or replaces the path to the file        |
| `hasName`              | answers whether a name is set                 |
| `getName` / `withName` | reads or replaces the name the recipient sees |

The path reaches the file on the machine that sends. The name is what the
recipient sees. Each port names its own path type.

```text
message = message.withAddedAttachment(
    new Attachment("/var/invoices/4821.pdf", "invoice.pdf"),
)
```

## Sending a message

`MailerContract` declares one method.

| Method | Does              | Takes     | Returns |
| ------ | ----------------- | --------- | ------- |
| `send` | sends the message | a message | nothing |

A send reports nothing. It either completes or throws.

```text
mailer = container.getSingleton(MailerContract)

mailer.send(message)
```

### Send from a queue, not from a request

Warning: a send reaches a third party, so it is slow and it fails in ways the
caller cannot answer. Push a job, and send from the consumer
([Queue](../queue/README.md)). A failure in the consumer retries, and the
request that pushed the job already answered.

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how
you override, and implementing nothing leaves every default in place.

### The component config

| Property        | Holds                                    | Default                    |
| --------------- | ---------------------------------------- | -------------------------- |
| `defaultMailer` | the adapter `MailerContract` resolves to | the port's own integration |

### Changing the default adapter

Name a different adapter in `defaultMailer`, and every call that resolves
`MailerContract` gets it. No call site changes, and a call that names an adapter
directly is unaffected.

```text
class Config implements ConfigContract, MailConfigContract {
    defaultMailer = LogMailer
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter               | Sends through                               |
| --------------------- | ------------------------------------------- |
| `<Integration>Mailer` | one third party the port sends through      |
| `LogMailer`           | nothing; it records what it would have sent |
| `NullMailer`          | nothing; the message is discarded           |

Every port ships the log and the null adapter. Which integrations ship is the
port's own choice.

### Configuring an integration adapter

Each integration carries the settings the third party needs, prefixed with the
integration's name. An integration that reaches a mail server carries a host and
credentials.

| Property                  | Holds                             | Default               |
| ------------------------- | --------------------------------- | --------------------- |
| `<integration>Host`       | where to reach the third party    | a placeholder         |
| `<integration>Port`       | which port to reach it on         | the third party's own |
| `<integration>Username`   | who to authenticate as            | a placeholder         |
| `<integration>Password`   | the secret to authenticate with   | a placeholder         |
| `<integration>Encryption` | which transport encryption to use | a usable default      |

An integration that reaches an API carries a key instead, and the domain it
sends from.

| Property              | Holds                                      | Default       |
| --------------------- | ------------------------------------------ | ------------- |
| `<integration>ApiKey` | the key the third party authenticates with | a placeholder |
| `<integration>Domain` | the domain the message is sent from        | a placeholder |

Warning: a setting left at a placeholder fails every send. Set each setting the
third party needs.

```text
class Config implements ConfigContract, MailConfigContract, MailMailgunConfigContract {
    defaultMailer = MailgunMailer

    mailgunDomain = "mail.example.com"
    mailgunApiKey = environment("MAILGUN_API_KEY")
}
```

### Configuring the log adapter

| Property        | Holds                          | Default            |
| --------------- | ------------------------------ | ------------------ |
| `mailLogLogger` | which logger the records go to | the default logger |

The log adapter names a logger rather than holding one, so the records go
wherever that logger goes ([Log](../log/README.md)). It records the sender,
every recipient set, the attachments, the subject, and both bodies.

A development environment names it as the default, so a developer reads what
would have left the machine.

### Configuring the null adapter

The null adapter has no config contract. It discards the message, so there is
nothing to set.

Name it as the default in a test that must send nothing.

## Adding an adapter

An adapter implements `MailerContract`. It reads the message and shapes it the
way the third party takes it.

```text
class SesMailer implements MailerContract {

    construct(client: SesClient, config: MailSesConfigContract) { }

    send(message) {
        request = this.client.newEmail(this.config.sesRegion)

        request.setFrom(this.address(message.getFrom()))
        request.setSubject(message.getSubject())

        for (recipient in message.getRecipients()) {
            request.addTo(this.address(recipient))
        }

        // ...one loop per further recipient set, and one for the attachments

        if (message.isHtml()) {
            request.setHtml(message.getBody())
            request.setText(message.getPlainBody())
        } else {
            request.setText(message.getBody())
        }

        request.send()
    }
}
```

An adapter formats an address itself, because each third party takes its own
form. Read `hasName` to choose the format.

```text
address(recipient) {
    if (! recipient.hasName()) {
        return recipient.getEmail()
    }

    return recipient.getName() + " <" + recipient.getEmail() + ">"
}
```

Give the adapter a config contract of its own, prefixing each property with the
integration's name so one config class implements several contracts at once:

```text
interface MailSesConfigContract {
    sesRegion: string
    sesApiKey: string
    sesSecret: string
}
```

Register the adapter and its config in a service provider, then name the adapter
in `defaultMailer`, or resolve it directly where you want it:

```text
class AppMailServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            SesMailer: [self, "publishSesMailer"],
        }
    }

    static publishSesMailer(container) {
        container.setSingleton(SesMailer, new SesMailer(
            container.getSingleton(SesClient),
            container.getSingleton(MailSesConfigContract),
        ))
    }
}
```

## Service registration

The component's provider binds:

| Binding                 | Is                                            |
| ----------------------- | --------------------------------------------- |
| `MailerContract`        | the adapter the component config names        |
| each adapter            | bound under its own name, for an explicit ask |
| each integration client | bound under its own name, for the adapter     |
| the component config    | your application's own config class           |
| each adapter config     | your application's own config class           |

```text
class MailServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            MailConfigContract: [self, "publishConfig"],
            MailMailgunConfigContract: [self, "publishMailgunConfig"],
            MailLogConfigContract: [self, "publishLogConfig"],
            MailerContract: [self, "publishMailer"],
            MailgunMailer: [self, "publishMailgunMailer"],
            MailgunClient: [self, "publishMailgunClient"],
            LogMailer: [self, "publishLogMailer"],
            NullMailer: [self, "publishNullMailer"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements MailConfigContract) {
            container.setSingleton(MailConfigContract, config)

            return
        }

        container.setSingleton(MailConfigContract, new MailConfig())
    }

    static publishMailer(container) {
        config = container.getSingleton(MailConfigContract)

        container.setSingleton(
            MailerContract,
            container.getSingleton(config.defaultMailer),
        )
    }

    // The adapter takes the integration client, and the setting it sends with.
    static publishMailgunMailer(container) {
        config = container.getSingleton(MailMailgunConfigContract)

        container.setSingleton(MailgunMailer, new MailgunMailer(
            container.getSingleton(MailgunClient),
            config.mailgunDomain,
        ))
    }

    // A publisher of its own builds the client from the config.
    static publishMailgunClient(container) {
        config = container.getSingleton(MailMailgunConfigContract)

        container.setSingleton(MailgunClient, new MailgunClient(
            apiKey: config.mailgunApiKey,
        ))
    }

    static publishLogMailer(container) {
        config = container.getSingleton(MailLogConfigContract)

        container.setSingleton(LogMailer, new LogMailer(
            container.getSingleton(config.mailLogLogger),
        ))
    }

    static publishNullMailer(container) {
        container.setSingleton(NullMailer, new NullMailer())
    }
}
```
