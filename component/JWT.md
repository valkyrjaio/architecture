# Jwt

The Jwt component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Jwt turns a set of claims into a signed token, and a signed token back into
claims. It is a **service component** on the shape in [`LOG.md`](LOG.md): one
contract, several adapters behind it, and the application names the one it wants
in config.

The component owns the signing and the verification. It owns no policy: what the
claims mean, how long a token lives and who may use one are the application's,
and [`AUTH.md`](AUTH.md) is where that policy usually sits.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the component contract, and one class per adapter          |
| `Data`       | the component config, and one config per algorithm family  |
| `Enum`       | the algorithms                                             |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

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

An algorithm family decides what a key is:

| Family     | Signs with        | Verifies with           |
| ---------- | ----------------- | ----------------------- |
| symmetric  | one shared secret | the same shared secret  |
| asymmetric | a private key     | the matching public key |

**The choice is not a detail.** A symmetric token can be forged by anyone who can
verify it, so it is for tokens this application both issues and reads. An
asymmetric token can be verified by a party that cannot issue one, which is what
lets another service trust a token without being able to mint it.

Each family gets its own config contract, because a shared secret and a key pair
are not the same settings
([`COMPONENT_CONFIG.md`](../convention/COMPONENT_CONFIG.md)).

---

## Using it

### Configure it

The component config names the **default jwt**: the adapter the component
contract resolves to. Pick the algorithm family by who needs to verify, then
configure that family.

### Put an expiry in the claims, always

The component signs what it is given and verifies what it is handed. A token with
no expiry is valid until the key changes, which in practice means forever.

### Keep the claims small and public

Everything in a token is readable by whoever holds it — signing proves the claims
were not altered, it does not hide them. Put an identifier in the token and read
the rest from a store.

### Never accept the algorithm from the token

A token states its own algorithm, and trusting that statement is how a verifier
is tricked into checking a signature the attacker chose. The configured family
decides, and the component's adapters are the only place that decision lives.

### Expect a key rotation to invalidate tokens

Changing the key means every token signed with the old one fails to decode. Plan
for the tokens in flight, which is a reason to keep their lifetime short.

---

## Permitted variation

| Variation                            | Reason                                      |
| ------------------------------------ | ------------------------------------------- |
| which integrations ship              | each ecosystem has its own token libraries  |
| which algorithms an adapter supports | a library supports a subset of the families |

The contract, the two operations and their names do not vary.
