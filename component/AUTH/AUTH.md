# Auth

The Auth component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Auth decides **who the caller is**, and keeps that decision for the length of a
session. It is a **service component** on the shape in [`LOG.md`](../LOG/LOG.md).

It owns authentication. It does not own authorization: whether a user may do a
thing is the application's rule, and the component's part is only to say which
user is asking and to carry the permissions the application attached.

**Hashing a password lives here, not in [`CRYPT.md`](../CRYPT/CRYPT.md).** A password is
hashed one way, so that nothing — this application included — can read it back.
Encryption is two-way, and a password that can be decrypted is a password an
attacker can read.

---

## Hierarchy

| Subcomponent    | Holds                                                 |
| --------------- | ----------------------------------------------------- |
| `Authenticator` | the component contract: who is authenticated          |
| `Hasher`        | password hashing and confirmation                     |
| `Store`         | retrieval and persistence of the user record          |
| `Entity`        | the user contracts an application's user implements   |
| `Data`          | the configs, the attempts, and the retrieval criteria |
| `Constant`      | the shared vocabulary                                 |
| `Provider`      | the service provider                                  |
| `Throwable`     | the component's throwable contract and its exceptions |

---

## The user is a set of opted-into contracts

**The application owns its user class.** It implements a base contract, and then
one further contract per capability it wants. A capability may build on another,
so implementing one can bring its prerequisite. Being verifiable requires being
mailable, because a verification has to be sent. The component reads only what the
user implements, so a feature the application did not opt into does not exist
for it.

The base contract declares the values authentication reads: the value that
identifies a user, and the value a check compares a password against.

A user is an entity, so the field each value lives in belongs to the `Registry`
rather than to the contract. The `Registry` also names the field a reset token is
stored under, and the metadata a capability below reads. Its entries are in
[`ORM.md`](../ORM/ORM.md).

Capabilities an application adds by implementing one more contract:

| Capability           | Lets the application                                     |
| -------------------- | -------------------------------------------------------- |
| verifiable           | require a confirmed address before access                |
| mailable             | send to the user through [`MAIL.md`](../MAIL/MAIL.md)    |
| lockable             | lock an account after failed attempts                    |
| two-factor           | require a second factor                                  |
| pin                  | authenticate with a short code                           |
| anti-phish code      | show the user a value only the real site knows           |
| device-authenticated | recognize a known device, and record the ones it knows   |
| last-online          | record when the user was last seen                       |
| permissible          | carry the permissions the application authorizes against |

A recovery code and a known device are **separate records** rather than
capabilities: each is its own entity the user relates to, so they are declared as
records and not as contracts the user class implements.

**This is the component's central design.** A fixed user class would force every
application to carry every feature's columns; a set of contracts means the
component's behavior follows what the application declared.

---

## Canonical operations

### The authenticator

| Operation               | Does                                                     |
| ----------------------- | -------------------------------------------------------- |
| `authenticate`          | establishes who the caller is, from an attempt           |
| `unauthenticate`        | discards that                                            |
| `isAuthenticated`       | reports whether anyone is                                |
| `getAuthenticated`      | returns the user                                         |
| `getImpersonated`       | returns the user being acted as, where that differs      |
| `getAuthenticatedUsers` | returns the set, for an application with several at once |
| `setAuthenticatedUsers` | replaces that set                                        |

**Impersonation is in the contract, not bolted on.** An administrator acting as a
user needs both identities available — the acting one for permissions and the
real one for the audit record — so a single "current user" cannot express it, and
an application that fakes it loses the audit trail.

### The hasher

| Operation         | Does                                             |
| ----------------- | ------------------------------------------------ |
| `hashPassword`    | turns a password into the value that is stored   |
| `confirmPassword` | checks a supplied password against a stored hash |

Two operations, and **no operation that returns a password.** The component
cannot read one back, and that is the property being protected.

### The store

| Operation        | Does                                            |
| ---------------- | ----------------------------------------------- |
| `hasRetrievable` | reports whether a user matching criteria exists |
| `retrieve`       | reads the user                                  |
| `create`         | writes a new user                               |
| `update`         | writes changes to one                           |

---

## Attempts

An **attempt** is a typed request to do one thing, rather than a bag of
parameters: authenticating, resetting a password, requesting a reset, locking and
unlocking each have their own type.

**The attempt is what makes the operation auditable and validatable.** A typed
attempt states what was asked, so the thing that logs it and the thing that rate
limits it read the same object rather than guessing from arguments.

---

## Using it

### Configure it

The component config names the adapters, and the session-backed form has its own
config ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Declare only the capabilities you want

Implement the base user contract, then add a capability contract per feature. Each
one you add is behavior the component starts applying, so adding them all is
adding work the application does not need.

### Always hash, never encrypt

Store the hash. A password the application can read back is the failure this
component is shaped to prevent.

### Authenticate through an attempt

Build the attempt and pass it. Code that reaches past the attempt loses the record
of what was asked.

### Keep authorization in the application

The component says who the caller is and carries the permissions attached. What
those permissions mean is the application's, enforced in middleware at the
`RouteMatched` stage ([`LIFECYCLE.md`](../../convention/LIFECYCLE.md)).

### Regenerate the session identifier on sign-in

Authentication changes the trust level, so the session identifier changes with it
([`SESSION.md`](../SESSION/SESSION.md)).

---

## Permitted variation

| Variation                           | Reason                                                          |
| ----------------------------------- | --------------------------------------------------------------- |
| which hashing algorithm ships       | each ecosystem has its own, and the strongest changes over time |
| where the authenticated set is kept | a port may hold it in a session or in a token                   |
| which store adapters ship           | a user record may live in any store                             |

The contracts, the capability set and the operation names do not vary.
