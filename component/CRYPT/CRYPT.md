# Crypt

The Crypt component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Crypt encrypts a value so only this application can read it back, and decrypts
what it encrypted. It is a **service component** on the shape in
[`LOG.md`](../LOG/LOG.md): one contract, several adapters behind it, and the application
names the one it wants in config.

**Crypt is two-way. Hashing is not Crypt's job.** A password is hashed by
[`AUTH.md`](../AUTH/AUTH.md), which is one-way by design: nothing should be able to read
a password back, including this application.

---

## Hierarchy

| Subcomponent | Holds                                                      |
| ------------ | ---------------------------------------------------------- |
| `Manager`    | the component contract, and one class per adapter          |
| `Data`       | the component config, and one config per adapter           |
| `Provider`   | the service provider that binds the contract to an adapter |
| `Throwable`  | the component's throwable contract and its exceptions      |

---

## Canonical operations

| Operation                 | Does                                             |
| ------------------------- | ------------------------------------------------ |
| `encrypt`                 | encrypts a string                                |
| `decrypt`                 | decrypts a string this component encrypted       |
| `encryptArray`            | encrypts a list or map                           |
| `decryptArray`            | decrypts one back to a list or map               |
| `encryptObject`           | encrypts an object                               |
| `decryptObject`           | decrypts one back to an object                   |
| `isValidEncryptedMessage` | reports whether a string is one this can decrypt |

**The typed forms exist so a caller never serializes before encrypting.** A
caller that encodes a structure itself and encrypts the text picks the encoding,
and a second caller picks a different one, so the two cannot read each other.
These operations put the encoding inside the component.

**`isValidEncryptedMessage` exists so a decrypt failure is not the test.** Code
that reads a value which may or may not be encrypted — during a migration, for
instance — asks first rather than catching a throwable to decide.

A decrypt that cannot succeed **throws**. It does not return an empty value,
because an empty value flows onward and is read as data.

---

## Using it

### Configure it

The component config names the **default crypt**: the adapter the component
contract resolves to. Each adapter gets its own config contract
([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Encrypt the structure, not your encoding of it

Pass the list, the map or the object to the typed operation. The component owns
the encoding, so the value survives a change of adapter and another caller reads
it back the same way.

### Never encrypt what should be hashed

Anything the application should be unable to read — a password above all — is
hashed, not encrypted. Encrypting it means the application can read every one of
them, which is the property an attacker wants.

### Supply a key per call only to rotate

Each operation accepts a key, so a caller can decrypt with the old key while
writing with the new one. Leave it unset everywhere else, so one configured key
is the single answer and no call site carries a secret.

### Treat the key as the secret the whole component rests on

Every value the component ever wrote is readable by whoever holds the key, and
unreadable once it is lost. The key comes from the application's config, which is
where a deployment supplies it from its own secret store.

### Expect a rotation to need a migration

A new key cannot read what the old one wrote. Rotating means reading with the old
and writing with the new, which is why `isValidEncryptedMessage` exists.

---

## Permitted variation

| Variation                  | Reason                                                       |
| -------------------------- | ------------------------------------------------------------ |
| which integrations ship    | each ecosystem has its own cryptography libraries            |
| the cipher an adapter uses | an adapter is named for its library, and the library chooses |

The contract, the operations and their names do not vary. Warning: a port does
**not** add an operation that exposes the cipher or an initialization vector to a
caller — those are the adapter's, and a caller that can set them can set them
wrongly. The key is the one exception, and only as the per-call override above.
