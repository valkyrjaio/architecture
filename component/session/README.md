# Session

Implements
[`SESSION.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/session/SESSION.md).

## Introduction

The Session component carries state for one visitor across several units of
work. It carries that state in a cookie, in a token, in a store, or nowhere at
all. Your application code reads the same in every case.

An **adapter** is the thing that carries. Each one implements `SessionContract`,
and each is registered under its own name, so an application resolves whichever
one it wants.

`SessionContract` itself resolves to the **default** adapter: the one your config
names. It is the convenience, and it is what lets a website carry a session in a
cookie while a command carries one on an option, with no call site changing.

The component also owns the **cross-site request forgery token**. That token is
session state, so the session holds it.

A session **modifies itself**. `set` and `remove` write into the one session
every holder shares. A caller that expects a copy back gets nothing.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a session

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                |
| -------------------------- | ----------------------------------------- | ------------------------ |
| `SessionContract`          | the adapter the config names              | environment              |
| an adapter by its own name | that adapter, always                      | nothing                  |
| a contract of your own     | the adapter that contract's binding names | purpose, and environment |

### The default, through `SessionContract`

Use this where the carrier is the application's business rather than this call
site's, which is almost everywhere.

```text
session = container.getSingleton(SessionContract)

session.set("cart.id", cart.id)
```

### A named adapter, where the carrier is the point

Use this where you know this call site means that carrier whatever the
environment is.

```text
tokenSession = container.getSingleton(HeaderTokenSession)

userId = tokenSession.get("user.id")
```

### A contract of your own, per purpose

Declare a contract that extends `SessionContract` for a kind of visitor, and
bind it to whichever adapter that kind wants. Then the carrier behind each kind
varies on its own, and a call site names the kind rather than a carrier.

```text
interface WebSessionContract extends SessionContract { }
interface ApiSessionContract extends SessionContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppSessionServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            WebSessionContract: [self, "publishWebSession"],
            ApiSessionContract: [self, "publishApiSession"],
        }
    }

    static publishWebSession(container) {
        container.setSingleton(
            WebSessionContract,
            container.getSingleton(EncryptedCookieSession),
        )
    }

    static publishApiSession(container) {
        container.setSingleton(
            ApiSessionContract,
            container.getSingleton(HeaderJwtSession),
        )
    }
}
```

An encrypted cookie for a browser and a signed token for a client are two
bindings. A test that wants neither binds both to the null adapter.

## Reading and writing

| Method     | Does                                       | Takes                | Returns         |
| ---------- | ------------------------------------------ | -------------------- | --------------- |
| `start`    | reads what was stored, and begins          | nothing              | nothing         |
| `isActive` | answers whether the session has started    | nothing              | a boolean       |
| `getId`    | reads the session identifier               | nothing              | the identifier  |
| `setId`    | replaces the session identifier            | an identifier        | nothing         |
| `getName`  | reads the session name                     | nothing              | the name        |
| `setName`  | replaces the session name                  | a name               | nothing         |
| `has`      | answers whether a key holds a value        | a key                | a boolean       |
| `get`      | reads the value under a key                | a key, and a default | the value       |
| `set`      | writes one value                           | a key, and a value   | nothing         |
| `remove`   | removes one key                            | a key                | whether it went |
| `all`      | reads everything the session holds         | nothing              | the values      |
| `clear`    | removes every value, and keeps the session | nothing              | nothing         |
| `destroy`  | ends the session, and discards it          | nothing              | nothing         |

The default on `get` is optional. Pass one where a missing key has an answer,
and check `has` first where a missing key is a branch of its own.

```text
session.set("cart.id", cart.id)

if (session.has("cart.id")) {
    cartId = session.get("cart.id")
}

// Every value, for a page that shows what the visitor has done.
values = session.all()
```

A longer example, in a controller:

```text
class CheckoutController {

    construct(session: SessionContract, carts: CartRepository) { }

    create(request) {
        cart = this.carts.create(request.body)

        this.session.set("cart.id", cart.id)

        return response.created(cart)
    }
}
```

### A session modifies itself

`set` and `remove` write into the session. There is no copy to assign, and
nothing to hand on to the next holder.

```text
// Wrong. There is no copy, so the assignment holds nothing.
session = session.set("cart.id", cart.id)

// Right. The call writes into the session every holder shares.
session.set("cart.id", cart.id)
```

### The session reads what was stored when it starts

Warning: a read before the session starts answers from an empty session rather
than failing.

The component's own adapters start as part of being built, so a session the
container resolves has read what was stored. `isActive` answers whether a
session has started, and `start` starts one that has not.

### Put identifiers in the session, not the data

The carrier moves the session on every unit of work, so a large session costs
the application on every request. Store what identifies the data, and read the
data itself from a store.

```text
// Wrong. Every unit of work carries the whole cart.
session.set("cart", cart)

// Right. The session carries the id, and the store holds the cart.
session.set("cart.id", cart.id)
```

### Replace the identifier when the trust level changes

An identifier captured before a sign-in names the session after it. Replace the
identifier on sign-in, so the captured one stops naming anything.

```text
session.setId(randomIdentifier())

session.set("user.id", user.id)
```

### `clear` keeps the session, `destroy` ends it

Clearing empties a session the visitor still holds. Destroying ends it, so the
identifier the visitor holds stops naming a session. A sign-out destroys.

```text
// A sign-out ends the session.
session.destroy()

// A checkout that starts again keeps it.
session.clear()
```

### A token carries the session one way

The token and the jwt adapters read the session from the request that carries
it. They write nothing back. `set` on one of them changes the session this unit
of work holds, and the next request carries whatever token the client holds. An
application that changes such a session issues a new token for the client.

## The forgery token

A form the application renders carries a token, and the handler for that form
checks the token the visitor submitted. The token lives in the session, so a
request from another site cannot carry one that matches.

| Method              | Does                                               | Takes                | Returns   |
| ------------------- | -------------------------------------------------- | -------------------- | --------- |
| `generateCsrfToken` | creates a token, and stores it under an id         | a token id           | the token |
| `validateCsrfToken` | checks a submitted token, and raises when it fails | an id, and the token | nothing   |
| `isValidCsrfToken`  | checks a submitted token                           | an id, and the token | a boolean |

Warning: a check that passes consumes the token. The session removes the token,
so one token answers one submission. Generate a token for each form the
application renders, and a new one after each submission.

The token id is a session key, so give each form an id of its own.

```text
// Rendering the form.
token = session.generateCsrfToken("checkout")

// Handling the submission.
session.validateCsrfToken("checkout", request.body.token)
```

Middleware that rejects a bad request uses the form that raises. Code that
chooses between two paths uses the boolean.

```text
if (! session.isValidCsrfToken("checkout", request.body.token)) {
    return response.badRequest()
}
```

## Configuration

The component reads one config contract of its own, plus one per adapter that
has settings.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property         | Holds                                     | Default                      |
| ---------------- | ----------------------------------------- | ---------------------------- |
| `defaultSession` | the adapter `SessionContract` resolves to | the session the runtime owns |
| `sessionId`      | the identifier every adapter starts with  | none                         |
| `sessionName`    | the name every adapter starts with        | none                         |

`sessionId` and `sessionName` sit on the component config rather than on an
adapter config, because every adapter takes both.

Warning: an adapter that stores the session under the identifier keys every
visitor alike without one. The cache adapter stores that way, so set `sessionId`
where the application uses that adapter.

### Changing the default adapter

Name a different adapter in `defaultSession`, and every call that resolves
`SessionContract` gets it. No call site changes, and a call that names an
adapter directly is unaffected.

```text
class Config implements ConfigContract, SessionConfigContract {
    defaultSession = EncryptedCookieSession
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter                  | Carries the session in                         |
| ------------------------ | ---------------------------------------------- |
| `<Runtime>Session`       | the session the language runtime owns          |
| `CookieSession`          | a cookie the visitor returns                   |
| `EncryptedCookieSession` | the same cookie, encrypted                     |
| `CacheSession`           | the cache, under a key the identifier names    |
| `HeaderTokenSession`     | a bearer token on a request header             |
| `OptionTokenSession`     | a token on a command-line option               |
| `HeaderJwtSession`       | a signed token on a request header             |
| `OptionJwtSession`       | a signed token on a command-line option        |
| `LogSession`             | nothing; it records what it would have carried |
| `NullSession`            | nothing; every read reports an empty session   |

Every port ships the log and null adapters. A port ships the adapter for each
carrier its ecosystem has, and it adds a row for a carrier the others do not
carry.

The cookie, the token and the jwt adapters each have an encrypted variant, named
with `Encrypted` first. A variant decrypts what it reads, and the cookie variant
also encrypts what it writes ([Crypt](../crypt/README.md)). The jwt adapters
read a signed token through [Jwt](../jwt/README.md), and `CacheSession` stores
through [Cache](../cache/README.md).

### Configuring the runtime adapter

The runtime adapter uses the session the runtime owns, and it configures the
cookie that session travels in. The adapter and each of its properties carry the
runtime's name. A port spells that name in place of `Runtime` and `runtime`
below.

| Property                  | Holds                                       | Default                  |
| ------------------------- | ------------------------------------------- | ------------------------ |
| `<runtime>CookiePath`     | the path the cookie is valid for            | `/`                      |
| `<runtime>CookieDomain`   | the domain the cookie is valid for          | none                     |
| `<runtime>CookieLifetime` | how long the cookie lives, in seconds       | `0`, one browser session |
| `<runtime>CookieSecure`   | whether the cookie needs a secure transport | `false`                  |
| `<runtime>CookieHttpOnly` | whether to hide the cookie from scripts     | `false`                  |
| `<runtime>CookieSameSite` | the same-site policy of the cookie          | `none`                   |

Warning: the defaults let the cookie travel without a secure transport, and let
a script on the page read it. Set `<runtime>CookieSecure` and
`<runtime>CookieHttpOnly` in production.

```text
class Config implements ConfigContract, SessionConfigContract, SessionRuntimeConfigContract {
    defaultSession = RuntimeSession

    runtimeCookieDomain = "orders.example.com"
    runtimeCookieSecure = true
    runtimeCookieHttpOnly = true
    runtimeCookieSameSite = SameSite.LAX
}
```

### Configuring the token adapters

| Property          | Holds                                     | Default                  |
| ----------------- | ----------------------------------------- | ------------------------ |
| `tokenHeaderName` | the request header that carries the token | the authorization header |
| `tokenOptionName` | the command-line option that carries it   | the token option         |

One contract holds both properties, and each token adapter reads the one it
needs. The header adapter expects a bearer token, and it raises where the header
holds something else.

```text
class Config implements ConfigContract, SessionConfigContract, SessionTokenConfigContract {
    defaultSession = HeaderTokenSession

    tokenHeaderName = "X-Session-Token"
}
```

### Configuring the jwt adapters

| Property        | Holds                                     | Default                  |
| --------------- | ----------------------------------------- | ------------------------ |
| `jwtHeaderName` | the request header that carries the token | the authorization header |
| `jwtOptionName` | the command-line option that carries it   | the token option         |

The jwt adapters hold their own pair, so an application carries a jwt on one
header and a plain token on another.

### Configuring the log adapter

| Property           | Holds                          | Default            |
| ------------------ | ------------------------------ | ------------------ |
| `sessionLogLogger` | which logger the records go to | the default logger |

The log adapter names a logger rather than holding one, so the records go
wherever that logger goes ([Log](../log/README.md)). It records the session
after each write, so a record carries what the session held at that point.

### The adapters with no config of their own

The cookie adapters, the cache adapter and the null adapter have no config
contract. A cookie adapter writes its cookie under the session identifier, the
cache adapter stores through whichever cache the application configures, and the
null adapter carries nothing.

Name the null adapter as the default in a test, or in an environment that keeps
no state between units of work.

## Adding an adapter

An adapter implements `SessionContract`. Where this port ships a base the
adapters share, extend that instead and implement `start`. The base holds the
values, the forgery token and the operations over both, so an adapter supplies
the carrier.

```text
class MemcachedSession extends Session {

    construct(client: MemcachedClient, config: SessionMemcachedConfigContract) { }

    start() {
        stored = this.client.fetch(this.storeKey())

        if (stored == "") {
            return
        }

        this.data = decode(stored)
    }

    // One override per write, so the carrier takes the change.
    set(key, value) {
        super.set(key, value)

        this.client.store(this.storeKey(), encode(this.data))
    }

    // ...remove, clear and destroy also write

    storeKey() {
        return this.config.memcachedPrefix + "." + this.getId()
    }
}
```

Give it a config contract of its own where it needs settings, prefixing each
property with the adapter's name so one config class implements several contracts
at once:

```text
interface SessionMemcachedConfigContract {
    memcachedHost: string
    memcachedPort: int
    memcachedPrefix: string
}
```

Register the adapter and its config in a service provider. Then name the adapter
in `defaultSession`, or resolve it directly where you want it:

```text
class AppSessionServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            MemcachedSession: [self, "publishMemcachedSession"],
        }
    }

    static publishMemcachedSession(container) {
        config = container.getSingleton(SessionConfigContract)

        container.setSingleton(MemcachedSession, new MemcachedSession(
            container.getSingleton(MemcachedClient),
            container.getSingleton(SessionMemcachedConfigContract),
            sessionId: config.sessionId,
            sessionName: config.sessionName,
        ))
    }
}
```

An adapter you write takes the identifier and the name from the component
config, as the component's own adapters do.

## Service registration

The component's provider binds:

| Binding              | Is                                            |
| -------------------- | --------------------------------------------- |
| `SessionContract`    | the adapter the component config names        |
| each adapter         | bound under its own name, for an explicit ask |
| the component config | your application's own config class           |
| each adapter config  | your application's own config class           |

```text
class SessionServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            SessionConfigContract: [self, "publishConfig"],
            SessionRuntimeConfigContract: [self, "publishRuntimeConfig"],
            SessionTokenConfigContract: [self, "publishTokenConfig"],
            SessionJwtConfigContract: [self, "publishJwtConfig"],
            SessionLogConfigContract: [self, "publishLogConfig"],
            SessionContract: [self, "publishSession"],
            RuntimeSession: [self, "publishRuntimeSession"],
            CookieSession: [self, "publishCookieSession"],
            CacheSession: [self, "publishCacheSession"],
            LogSession: [self, "publishLogSession"],
            NullSession: [self, "publishNullSession"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements SessionConfigContract) {
            container.setSingleton(SessionConfigContract, config)

            return
        }

        container.setSingleton(SessionConfigContract, new SessionConfig())
    }

    static publishSession(container) {
        config = container.getSingleton(SessionConfigContract)

        container.setSingleton(
            SessionContract,
            container.getSingleton(config.defaultSession),
        )
    }

    static publishLogSession(container) {
        config = container.getSingleton(SessionConfigContract)
        logConfig = container.getSingleton(SessionLogConfigContract)

        container.setSingleton(LogSession, new LogSession(
            container.getSingleton(logConfig.sessionLogLogger),
            sessionId: config.sessionId,
            sessionName: config.sessionName,
        ))
    }
}
```

Every adapter publisher reads the identifier and the name from the component
config, so one setting reaches whichever adapter the application names.
