# Auth

Implements
[`AUTH.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/auth/AUTH.md).

## Introduction

The Auth component decides who the caller is, and keeps that decision for the
length of a session. It authenticates. It does not authorize: whether a user may
do a thing is the application's rule, and this component's part is to say which
user is asking and to carry the permissions the application attached.

The component also hashes a password. Hashing is **one-way**. Nothing reads the
password back, this application included. Encryption is two-way, and the Crypt
component owns it ([Crypt](../crypt/README.md)).

Four things are swappable, each on its own. An **authenticator** decides what
authenticating means. A **store** reads and writes the user record. A **hasher**
turns a secret into the value that is stored. An **entity** is what the
application's user is.

`AuthenticatorContract` resolves to the **default** authenticator: the one your
config names. It is the convenience, and it is what lets a test authenticate
against users held in memory while production reads them from a database, with
no call site changing.

This document assumes the framework, where the container resolves an
authenticator and a provider builds it. Used standalone, the component needs
none of that: construct an authenticator and call it.

## Resolving an authenticator

A contract is a resolvable set of methods, independent of the class that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                          | Gets                                            | Varies by                 |
| -------------------------------- | ----------------------------------------------- | ------------------------- |
| `AuthenticatorContract`          | the authenticator the config names              | environment               |
| an authenticator by its own name | that authenticator, always                      | nothing                   |
| a contract of your own           | the authenticator that contract's binding names | audience, and environment |

### The default, through `AuthenticatorContract`

Use this where the application has one kind of user, which is almost everywhere.

```text
authenticator = container.getSingleton(AuthenticatorContract)

user = authenticator.authenticate(attempt)
```

### A named authenticator, where the kind of authentication is the point

Use this where this call site means that authenticator whatever the environment
is — an endpoint that reads the decision from the session, in an application
whose default reads it from a token.

```text
authenticator = container.getSingleton(SessionAuthenticator)
```

### A contract of your own, per audience

Declare a contract per audience, and bind each to an authenticator built for
that audience. An administrator and a customer then authenticate against
different user classes, and a call site names the audience.

```text
interface AdminAuthenticatorContract extends AuthenticatorContract { }
interface CustomerAuthenticatorContract extends AuthenticatorContract { }
```

```text
class AppAuthServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            AdminAuthenticatorContract: [self, "publishAdminAuthenticator"],
            CustomerAuthenticatorContract: [self, "publishCustomerAuthenticator"],
        }
    }

    static publishAdminAuthenticator(container) {
        container.setSingleton(AdminAuthenticatorContract, new SessionAuthenticator(
            session: container.getSingleton(SessionContract),
            store: container.getSingleton(OrmStore),
            hasher: container.getSingleton(PasswordHasherContract),
            entity: AdminUser,
            sessionItemId: "auth.admins",
        ))
    }

    // ...one publisher per contract
}
```

Each audience keeps its decision under a session item id of its own, so signing
in as a customer does not sign anyone in as an administrator.

## The four surfaces

| Surface           | Decides                                                   | Named by               |
| ----------------- | --------------------------------------------------------- | ---------------------- |
| the authenticator | what authenticating means, and where the decision is kept | `defaultAuthenticator` |
| the store         | where the user record is read and written                 | `defaultStore`         |
| the hasher        | how a secret is turned into something stored              | its own binding        |
| the user entity   | what the application's user is                            | `defaultUserEntity`    |

They are separate because an application changes one without the others. Moving
the decision from the session to a token changes the authenticator alone.
Reading the user from a directory rather than a database changes the store
alone. Changing the hash changes the hasher alone. Adding a capability to the
user changes the entity alone.

The authenticator takes the other three. It holds the store it reads users from,
the hasher it checks passwords with, and the class of the user it retrieves.

## The operations

| Operation               | On                | Does                                               |
| ----------------------- | ----------------- | -------------------------------------------------- |
| `authenticate`          | the authenticator | establishes who the caller is, from an attempt     |
| `unauthenticate`        | the authenticator | discards the decision for one user id              |
| `isAuthenticated`       | the authenticator | answers whether anyone is authenticated            |
| `getAuthenticated`      | the authenticator | returns the user                                   |
| `getImpersonated`       | the authenticator | returns the user being acted as                    |
| `getAuthenticatedUsers` | the authenticator | returns the authenticated set                      |
| `setAuthenticatedUsers` | the authenticator | replaces the authenticated set                     |
| `hasRetrievable`        | the store         | answers whether a user matching a retrieval exists |
| `retrieve`              | the store         | reads the user a retrieval names                   |
| `create`                | the store         | writes a new user                                  |
| `update`                | the store         | writes changes to one                              |
| `hashPassword`          | the hasher        | turns a password into the value that is stored     |
| `confirmPassword`       | the hasher        | checks a supplied password against a stored hash   |

**No operation returns a password.** The component cannot read one back.

### Authenticate through an attempt

Build an attempt and pass it. The authenticator reads the user the attempt
names, checks the password against the stored hash, and records the decision.

```text
attempt = new AuthenticationAttempt(
    new RetrievalByUsername(request.body.username),
    request.body.password,
)

user = authenticator.authenticate(attempt)
```

A retrieval and an attempt are separate objects. A **retrieval** says which user
to look for. An **attempt** pairs that retrieval with the secret the caller
offered, and with nothing else. The store takes the retrieval, so the secret
never travels to the store beside the record it is checked against.

| Retrieval                  | Finds the user by |
| -------------------------- | ----------------- |
| `RetrievalByUsername`      | the username      |
| `RetrievalById`            | the identifier    |
| `RetrievalByIdAndUsername` | both of those     |
| `RetrievalByResetToken`    | the reset token   |

An attempt is typed per operation rather than shaped as a bag of parameters, so
the thing that logs it and the thing that rate limits it read one object.

| Attempt                 | Asks to                                               |
| ----------------------- | ----------------------------------------------------- |
| `AuthenticationAttempt` | authenticate, with a retrieval and a password         |
| `ForgotPasswordAttempt` | start a password reset, with a retrieval              |
| `ResetPasswordAttempt`  | finish a reset, with a retrieval and the new password |
| `LockAttempt`           | lock the account a retrieval names                    |
| `UnlockAttempt`         | unlock that account                                   |

Warning: an answer that separates an unknown user from a wrong password tells an
attacker which names exist. A failed attempt throws. Catch it, and report one
answer for both.

```text
try {
    user = authenticator.authenticate(attempt)
} catch (throwable) {
    return response.unauthorized()
}
```

### Keep the decision between requests

The authenticator holds an `AuthenticatedUsersContract`: the set of user ids
that are authenticated, and which of them is current. It holds ids rather than
users, so a later request reads the record from the store and sees what the
record says now.

| Method                                                  | Does                                      |
| ------------------------------------------------------- | ----------------------------------------- |
| `hasCurrent`, `getCurrent`, `setCurrent`                | reads and sets the current user's id      |
| `isImpersonating`, `getImpersonated`, `setImpersonated` | reads and sets the impersonated user's id |
| `isUserAuthenticated`                                   | answers whether one id is in the set      |
| `add`, `remove`                                         | adds one id to the set, or takes one out  |
| `all`                                                   | returns every id in the set               |

`SessionAuthenticator` keeps the set in the session, under the item id its
config names ([Session](../session/README.md)). An application may authenticate
several users at once, such as a support tool with two accounts open. It reads
the set, changes it, and sets it back.

```text
users = authenticator.getAuthenticatedUsers()

users.add(secondUser.getIdValue())

authenticator.setAuthenticatedUsers(users)
```

`unauthenticate` takes an id rather than acting on the current user, so an
application with several authenticated users signs one of them out.

```text
authenticator.unauthenticate(user.getIdValue())
```

### Regenerate the session identifier on sign-in

Warning: a session identifier that survives the sign-in lets anyone who knew it
before act as the user after. Authentication changes the trust level, so
regenerate the identifier once `authenticate` returns
([Session](../session/README.md)).

### Read who the caller is

```text
if (authenticator.isAuthenticated()) {
    user = authenticator.getAuthenticated()
}
```

Impersonation is in the contract rather than in the application. An
administrator acting as a user needs both identities: the acting one for what
the screen shows, and the real one for the audit record.

```text
authenticator.getAuthenticatedUsers().setImpersonated(customer.getIdValue())

// What the screen shows.
actingAs = authenticator.getImpersonated()

// Who to record in the audit trail.
administrator = authenticator.getAuthenticated()
```

### Hash a password before you store it

The hasher turns a password into the value the store writes. Call it wherever
the application takes a password, and store what it returns.

```text
user = new AppUser()
user.username = request.body.username
user.password = hasher.hashPassword(request.body.password)

store.create(user)
```

`confirmPassword` is the only check against a stored hash. It answers whether a
supplied password matches, and `authenticate` calls it for you.

```text
// Wrong — hash the supplied password, and compare the two hashes. Each hash
// carries a salt of its own, so two hashes of one password differ.
if (hasher.hashPassword(request.body.password) === user.getPasswordValue()) {
    // ...
}

// Right — compare through the hasher.
if (hasher.confirmPassword(request.body.password, user.getPasswordValue())) {
    // ...
}
```

### Authorization stays in the application

The component says which user is asking, and `PermissibleUserContract` carries
the permissions the application attached. What a permission means is the
application's rule. Enforce it in middleware, once the route is matched, so one
decision covers every route that needs it.

```text
if (user.isDenied("orders.refund")) {
    return response.forbidden()
}
```

## The capabilities a user opts into

**The application owns its user class.** The class implements the base contract,
`UserContract`, and then one further contract per capability the application
wants. The component reads only what the user implements, so a feature the
application did not opt into does not exist for it.

`UserContract` declares the two values authentication reads: the value that
identifies the user, through `getUsernameValue`, and the value a check compares
a password against, through `getPasswordValue`.

A user is an entity. The Orm entity registry names the field each value lives
in, keyed by the user class ([Orm](../orm/README.md)). The user class carries no
field name. The registry also names the field a reset token is stored under, and
the fields each capability below reads.

| Capability           | Contract                          | Lets the application                                     |
| -------------------- | --------------------------------- | -------------------------------------------------------- |
| verifiable           | `VerifiableUserContract`          | require a confirmed address before access                |
| mailable             | `MailableUserContract`            | send to the user ([Mail](../mail/README.md))             |
| lockable             | `LockableUserContract`            | lock an account after failed attempts                    |
| two-factor           | `TwoFactorUserContract`           | require a second factor                                  |
| pin                  | `PinUserContract`                 | authenticate with a short code                           |
| anti-phish code      | `AntiPhishCodeUserContract`       | show the user a value only the real site knows           |
| device-authenticated | `DeviceAuthenticatedUserContract` | recognize a known device                                 |
| last-online          | `LastOnlineUserContract`          | record when the user was last seen                       |
| permissible          | `PermissibleUserContract`         | carry the permissions the application authorizes against |

**A capability may require another.** Being verifiable requires being mailable,
because a verification has to be sent. Implementing the verifiable contract
brings the mailable one with it.

```text
// Verifiable requires mailable, so the address comes with the capability.
class AppUser extends User implements VerifiableUserContract, LockableUserContract {

    email: string
    isVerified: boolean
    loginAttempts: int
    isLocked: boolean
}
```

Extend one of the component's own user classes where it already carries what you
want, and implement the contracts on a class of your own where it does not.

| Class            | Implements                                 |
| ---------------- | ------------------------------------------ |
| `User`           | the base contract, and nothing else        |
| `MailableUser`   | the base contract, and mailable            |
| `VerifiableUser` | the base contract, mailable and verifiable |
| `LockableUser`   | the base contract, and lockable            |

**A recovery code and a known device are separate records.** Each is an entity
of its own that the user relates to, declared through `UserRecoveryCodeContract`
and `UserDeviceContract`, rather than a contract the user class implements. One
user holds many of each, and the user class declares a relationship to them.

Adding every capability is adding work the application does not need. Declare
the ones the application uses, and add one when a feature needs it.

## Configuration

The component reads one config contract of its own, plus one per authenticator
that has settings.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how
you override, and implementing nothing leaves every default in place.

```text
// In the component's provider. Your class, or the component's default.
static publishConfig(container) {
    config = container.getSingleton(ConfigContract)

    if (config implements AuthConfigContract) {
        container.setSingleton(AuthConfigContract, config)

        return
    }

    container.setSingleton(AuthConfigContract, new AuthConfig())
}
```

### The component config

| Property               | Holds                                                 | Default                   |
| ---------------------- | ----------------------------------------------------- | ------------------------- |
| `defaultAuthenticator` | the authenticator `AuthenticatorContract` resolves to | the session authenticator |
| `defaultStore`         | the store `StoreContract` resolves to                 | the Orm store             |
| `defaultUserEntity`    | the user class the authenticator retrieves            | the component's own user  |

Three of the four surfaces are named here. The hasher is the fourth, and
[Changing the hasher](#changing-the-hasher) covers it.

### Changing the default authenticator

Name a different authenticator in `defaultAuthenticator`, and every call that
resolves `AuthenticatorContract` gets it. No call site changes, and a call that
names an authenticator directly is unaffected.

```text
class Config implements ConfigContract, AuthConfigContract {
    defaultAuthenticator = TokenAuthenticator
}
```

To use an authenticator you wrote yourself, name that class instead. It has to
be registered first, which [Adding an authenticator](#adding-an-authenticator)
covers.

### Changing the default store

```text
class Config implements ConfigContract, AuthConfigContract {
    defaultStore = InMemoryStore
}
```

### Changing the default user entity

Name your own user class, and the authenticator retrieves that class. Every
capability it implements is a capability the component applies.

```text
class Config implements ConfigContract, AuthConfigContract {
    defaultUserEntity = AppUser
}
```

### Changing the hasher

The hasher has no config property. The component's provider binds
`PasswordHasherContract` to the port's own hasher, and a provider of your own
rebinds it.

```text
class AppAuthServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            PasswordHasherContract: [self, "publishPasswordHasher"],
        }
    }

    static publishPasswordHasher(container) {
        container.setSingleton(
            PasswordHasherContract,
            container.getSingleton(Argon2PasswordHasher),
        )
    }
}
```

Warning: the new hasher does not confirm a password against a hash the old one
made. Confirm with the old hasher, then hash again with the new one, until every
stored hash is the new form.

### Available authenticators

| Authenticator          | Keeps the decision in |
| ---------------------- | --------------------- |
| `SessionAuthenticator` | the session           |

Every port ships the session authenticator. Which other ones ship is the port's
own choice.

### Configuring the session authenticator

| Property                | Holds                                         | Default                       |
| ----------------------- | --------------------------------------------- | ----------------------------- |
| `sessionItemId`         | the session item the set is kept under        | the component's own id        |
| `sessionAllowedClasses` | the classes the session may read the set into | the component's own set class |

Warning: a session value a caller modified names the class the session reads it
back into. Keep `sessionAllowedClasses` to the classes that hold the
authenticated set.

```text
class Config implements ConfigContract, AuthConfigContract, AuthSessionConfigContract {
    defaultAuthenticator = SessionAuthenticator

    sessionItemId = "orders.auth.users"
    sessionAllowedClasses = [AuthenticatedUsers]
}
```

### Available stores

| Store           | Reads the user from                         |
| --------------- | ------------------------------------------- |
| `OrmStore`      | the Orm component ([Orm](../orm/README.md)) |
| `InMemoryStore` | a set of users it is constructed with       |
| `NullStore`     | nothing; no user is ever retrievable        |

Every port ships all three. The in-memory store takes its users as a
constructor argument, which is what a test authenticates against.

### Available hashers

| Hasher                      | Hashes with                          |
| --------------------------- | ------------------------------------ |
| `<Ecosystem>PasswordHasher` | the ecosystem's own password hashing |

Every port ships one hasher over whatever its ecosystem provides. Which
algorithm that is belongs to the ecosystem, and the strongest one changes over
time.

## Adding an authenticator

An authenticator implements `AuthenticatorContract`. Where this port ships a
base the authenticators share, extend that instead and keep the decision your
own way.

```text
class TokenAuthenticator extends Authenticator {

    construct(
        request: ServerRequestContract,
        store: StoreContract,
        hasher: PasswordHasherContract,
        entity: class,
        config: AuthTokenConfigContract,
    ) { }

    // The base reads the authenticated set from where the authenticator keeps
    // it. This one reads it from the token the request carries.
    getAuthenticatedUsers() {
        return this.decode(this.request.headers[this.config.tokenHeaderName])
    }
}
```

Give it a config contract of its own where it needs settings, prefixing each
property with the authenticator's name so one config class implements several
contracts at once:

```text
interface AuthTokenConfigContract {
    tokenHeaderName: string
    tokenSecondsToLive: int
}
```

Register the authenticator and its config in a service provider. Then name the
authenticator in `defaultAuthenticator`, or resolve it directly where you want
it:

```text
class AppAuthServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            TokenAuthenticator: [self, "publishTokenAuthenticator"],
        }
    }

    static publishTokenAuthenticator(container) {
        config = container.getSingleton(AuthConfigContract)

        container.setSingleton(TokenAuthenticator, new TokenAuthenticator(
            request: container.getSingleton(ServerRequestContract),
            store: container.getSingleton(StoreContract),
            hasher: container.getSingleton(PasswordHasherContract),
            entity: config.defaultUserEntity,
            config: container.getSingleton(AuthTokenConfigContract),
        ))
    }
}
```

## Adding a store

A store implements `StoreContract`. It takes a retrieval and the user class, and
it returns the user that matches.

```text
class LdapStore implements StoreContract {

    construct(client: LdapClient, config: AuthLdapConfigContract) { }

    hasRetrievable(retrieval, user) {
        return this.find(retrieval, user) !== null
    }

    retrieve(retrieval, user) {
        return this.find(retrieval, user)
            ?? throw new AuthInvalidRetrievableUserException("No user for the criteria")
    }

    // ...one method per operation

    find(retrieval, user) {
        return this.client.search(
            this.config.ldapBaseDn,
            retrieval.getRetrievalFields(user),
        )
    }
}
```

`getRetrievalFields` returns the values to match, keyed by the field the Orm
registry names for that user class. A store reads those values and asks its own
backing in its own language, so a new retrieval needs no change to a store.

Give it a config contract of its own, and register both in a service provider:

```text
interface AuthLdapConfigContract {
    ldapHost: string
    ldapBaseDn: string
}
```

A store that cannot write a user throws rather than reporting success. Name the
null store where an environment authenticates nobody.

## Adding a hasher

A hasher implements `PasswordHasherContract`. Two methods, and neither one
returns a password.

```text
class Argon2PasswordHasher implements PasswordHasherContract {

    construct(config: AuthArgon2ConfigContract) { }

    hashPassword(password) {
        return argon2.hash(password, this.config.argon2Memory, this.config.argon2Time)
    }

    confirmPassword(password, hashedPassword) {
        return argon2.verify(password, hashedPassword)
    }
}
```

Give it a config contract of its own where the algorithm takes settings:

```text
interface AuthArgon2ConfigContract {
    argon2Memory: int
    argon2Time: int
}
```

Warning: a comparison that stops at the first byte that differs tells an
attacker how much of a hash was right. Compare in constant time, which the
ecosystem's own verification does.

Register the hasher and its config in a service provider, then rebind
`PasswordHasherContract` to it, which
[Changing the hasher](#changing-the-hasher) covers.

## Service registration

The component's provider binds:

| Binding                   | Is                                            |
| ------------------------- | --------------------------------------------- |
| `AuthenticatorContract`   | the authenticator the component config names  |
| each authenticator        | bound under its own name, for an explicit ask |
| `StoreContract`           | the store the component config names          |
| each store                | bound under its own name, for an explicit ask |
| `PasswordHasherContract`  | the port's own hasher                         |
| the component config      | your application's own config class           |
| each authenticator config | your application's own config class           |

```text
class AuthServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            AuthConfigContract: [self, "publishConfig"],
            AuthSessionConfigContract: [self, "publishSessionConfig"],
            AuthenticatorContract: [self, "publishAuthenticator"],
            SessionAuthenticator: [self, "publishSessionAuthenticator"],
            StoreContract: [self, "publishStore"],
            OrmStore: [self, "publishOrmStore"],
            InMemoryStore: [self, "publishInMemoryStore"],
            NullStore: [self, "publishNullStore"],
            PasswordHasherContract: [self, "publishPasswordHasher"],
        }
    }

    static publishAuthenticator(container) {
        config = container.getSingleton(AuthConfigContract)

        container.setSingleton(
            AuthenticatorContract,
            container.getSingleton(config.defaultAuthenticator),
        )
    }

    // The authenticator takes the default store, the hasher, and the user
    // class, so a change to any of those reaches it here.
    static publishSessionAuthenticator(container) {
        config = container.getSingleton(AuthConfigContract)
        sessionConfig = container.getSingleton(AuthSessionConfigContract)

        container.setSingleton(SessionAuthenticator, new SessionAuthenticator(
            session: container.getSingleton(SessionContract),
            store: container.getSingleton(StoreContract),
            hasher: container.getSingleton(PasswordHasherContract),
            entity: config.defaultUserEntity,
            sessionItemId: sessionConfig.sessionItemId,
            allowedClasses: sessionConfig.sessionAllowedClasses,
        ))
    }

    static publishStore(container) {
        config = container.getSingleton(AuthConfigContract)

        container.setSingleton(
            StoreContract,
            container.getSingleton(config.defaultStore),
        )
    }

    static publishOrmStore(container) {
        container.setSingleton(OrmStore, new OrmStore(
            container.getSingleton(ManagerContract),
        ))
    }

    // The port's own hasher, over whatever its ecosystem provides.
    static publishPasswordHasher(container) {
        container.setSingleton(
            PasswordHasherContract,
            new <Ecosystem>PasswordHasher(),
        )
    }
}
```

The user entity gets no binding. The config names the class, and the
authenticator retrieves it through the store.
