# Jwt

The Jwt component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the master copy of this component's README, which every port carries, so that
every port's document for this component reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Jwt turns a set of claims into a signed token, and a signed token back into
claims. It is a **service component** on the shape in [`LOG.md`](../log/LOG.md): one
contract, several adapters behind it, and the application names the one it wants
in config.

The component owns the signing and the verification. It owns no policy: what the
claims mean, how long a token lives and who may use one are the application's,
and [`AUTH.md`](../auth/AUTH.md) is where that policy usually sits.

---

## Hierarchy

| Subcomponent | Holds                                                                    |
| ------------ | ------------------------------------------------------------------------ |
| `Manager`    | the component contract, and one class per adapter                        |
| `Data`       | the component config, one per adapter with settings, and one per key set |
| `Enum`       | the algorithms                                                           |
| `Provider`   | the service provider that binds the contract to an adapter               |
| `Throwable`  | the component's throwable contract and its exceptions                    |

---

## Two operations

| Operation | Does                                                          |
| --------- | ------------------------------------------------------------- |
| `encode`  | takes a map of claims and returns a signed token              |
| `decode`  | takes a token, verifies the signature, and returns the claims |

**`decode` verifies.** It is not a parse: a token whose signature does not check
out throws rather than returning its claims. So a caller that holds decoded
claims holds claims this application signed, and no call site needs to remember
to verify separately.

That is the component's whole reason for existing as a contract rather than a
helper — the one operation that must never be skipped cannot be.

---

## Signing is asymmetric or symmetric, and the config says which

An algorithm signs with one of two kinds of key:

| Kind       | Signs with        | Verifies with           |
| ---------- | ----------------- | ----------------------- |
| symmetric  | one shared secret | the same shared secret  |
| asymmetric | a private key     | the matching public key |

**The choice is not a detail.** A symmetric token can be forged by anyone who can
verify it, so it is for tokens this application both issues and reads. An
asymmetric token can be verified by a party that cannot issue one, which is what
lets another service trust a token without being able to mint it.

Each adapter gets its own config contract for whatever its library needs, and
each key set gets one for its keys. Each contract's shape is in
[`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md).

Every port also ships a **null adapter**, which encodes and decodes without
signing or verifying. It is what a test binds, and it is the one place the
verification rule above does not hold — which is why it is never bound outside a
test.

---

## Structure

| Subcomponent | Holds                                                                    |
| ------------ | ------------------------------------------------------------------------ |
| `Manager/`   | the component contract, and one class per adapter                        |
| `Data/`      | the component config, one per adapter with settings, and one per key set |
| `Enum/`      | the algorithms, as a closed set                                          |
| `Provider/`  | the component aggregator, and the component's bindings                   |
| `Throwable/` | the component's throwable contract and its exceptions                    |

### An adapter is named for what it wraps, then for the role

| Adapter               | Wraps                                     |
| --------------------- | ----------------------------------------- |
| `<Integration><Role>` | one library or service the port speaks to |
| `Null<Role>`          | nothing; every call is discarded          |

The role comes last, so a reader scanning a directory sees what the classes are
before seeing what each one wraps. A name stacks where an integration has
variants of its own, with the narrower part first.

Every port ships the null adapter. Which integrations ship is the port's own
choice.

### A config per key set, not per algorithm

A shared secret and a key pair are not the same settings, and the algorithms that
sign with the same keys want the same settings. Those keys are a **key set**, and
each key set gets its own config contract. The algorithm itself is a value in
`Enum/`.

**The adapter's config names the algorithm, not the key set's.** The algorithm is
what decides which key set the adapter reads, so it cannot sit in the key set it
selects. An adapter whose config leaves the algorithm unset keeps the adapter's
default, reads the key set that default belongs to, and signs with that key set's
default key.

**An algorithm with no key set signs with the application secret.** A port ships a
key set for some of the algorithms in `Enum/`, not for all of them, so an
algorithm outside those has no keys of its own to read. It uses the application
config's `key` for both signing and verifying
([`APPLICATION.md`](../application/APPLICATION.md)).

Warning: a config that names such an algorithm and sets no keys signs, and
reports nothing. The application secret ships as a placeholder that every
installation carries, so the token is signed with a value an attacker knows.

### Order

| In                   | The order is                                             |
| -------------------- | -------------------------------------------------------- |
| the adapters         | each integration, then the shared kinds, then null       |
| the config contracts | the component's own, each key set's, then each adapter's |
| the provider's map   | the configs, the component contract, the adapters        |
| the throwables       | the contract, the two bases, then the concrete ones      |

---

## Using it

### Configure it

The component config names the **default jwt**: the adapter the component
contract resolves to. Pick the algorithm by who needs to verify, then configure
its key set.

### Put an expiry in the claims, always

The component signs what it is given and verifies what it is handed. A token with
no expiry is valid until the key changes, which in practice means forever.

### Keep the claims small and public

Everything in a token is readable by whoever holds it — signing proves the claims
were not altered, it does not hide them. Put an identifier in the token and read
the rest from a store.

### Never accept the algorithm from the token

A token states its own algorithm, and trusting that statement is how a verifier
is tricked into checking a signature the attacker chose. The configured key set
decides, and the component's adapters are the only place that decision lives.

### Expect a key rotation to invalidate tokens

Changing the key means every token signed with the old one fails to decode. Plan
for the tokens in flight, which is a reason to keep their lifetime short.

---

## Permitted variation

| Variation                            | Reason                                        |
| ------------------------------------ | --------------------------------------------- |
| which integrations ship              | each ecosystem has its own token libraries    |
| which algorithms an adapter supports | a library supports a subset of the algorithms |

The contract, the two operations and their names do not vary.
