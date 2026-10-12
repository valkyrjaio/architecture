# Crypt

Implements
[`CRYPT.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/crypt/CRYPT.md).

## Introduction

The Crypt component encrypts a value so it can be stored or sent, and decrypts
it back into what it was. It is **two-way**: everything it encrypts, it can
return. A value that should never come back is hashed rather than encrypted, and
hashing belongs to whatever hashes passwords.

An **adapter** is the thing that encrypts. Each one implements `CryptContract`,
and each is registered under its own name.

`CryptContract` itself resolves to the **default** adapter: the one your config
names.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a crypt

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

```text
crypt = container.getSingleton(CryptContract)

encrypted = crypt.encrypt(cardNumber)
```

Resolve an adapter by its own name where you want that one, and declare a
contract of your own where a kind of value should use a different key or a
different adapter.

## Encrypting and decrypting

| Method                    | Does                                                    | Takes                 | Returns    |
| ------------------------- | ------------------------------------------------------- | --------------------- | ---------- |
| `encrypt`                 | encrypts text                                           | the text, and a key   | the cipher |
| `decrypt`                 | returns the text                                        | the cipher, and a key | the text   |
| `encryptArray`            | encrypts a map                                          | the map, and a key    | the cipher |
| `decryptArray`            | returns the map                                         | the cipher, and a key | the map    |
| `encryptObject`           | encrypts an object                                      | the object, and a key | the cipher |
| `decryptObject`           | returns the object                                      | the cipher, and a key | the object |
| `isValidEncryptedMessage` | answers whether a value is a cipher this component made | the value             | a boolean  |

**The key is optional on every method.** Pass none and the adapter uses the key
its config names, which is what application code does. Pass one where a value
needs a key of its own, such as a per-tenant secret.

```text
// The configured key.
encrypted = crypt.encrypt(cardNumber)
cardNumber = crypt.decrypt(encrypted)

// A key of your own.
encrypted = crypt.encrypt(cardNumber, tenant.key)
```

### The structured forms exist so you do not serialize yourself

`encryptArray` and `encryptObject` take the value and return a cipher, and the
matching methods return the value. A caller that serialized first would choose a
format, and two call sites would choose differently.

```text
encrypted = crypt.encryptArray({ "card": number, "expiry": expiry })

payload = crypt.decryptArray(encrypted)
```

### Check before decrypting a value you did not store

`isValidEncryptedMessage` answers whether a value is a cipher this component
produced. Use it where the value came from somewhere you do not control, so a
malformed value is a branch rather than a throwable.

```text
if (! crypt.isValidEncryptedMessage(fromCookie)) {
    return response.badRequest()
}

payload = crypt.decrypt(fromCookie)
```

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default.

### The component config

| Property       | Holds                                   | Default               |
| -------------- | --------------------------------------- | --------------------- |
| `defaultCrypt` | the adapter `CryptContract` resolves to | the port's own cipher |

### Available adapters

| Adapter         | Encrypts with                         |
| --------------- | ------------------------------------- |
| `<Cipher>Crypt` | one cipher the port speaks to         |
| `NullCrypt`     | nothing; it returns what it was given |

**There is no log adapter here.** Recording what a value would have encrypted
writes that value in the clear.

Name the null adapter as the default only in a test that does not care about the
cipher. It returns the input, so a test asserting that a value was encrypted
would pass against an unencrypted value.

### Configuring a cipher adapter

| Property      | Holds                    | Default |
| ------------- | ------------------------ | ------- |
| `<cipher>Key` | the key the adapter uses | none    |

The key has no default. A cipher with a default key is a cipher with a published
key, so the application sets it or the adapter raises.

```text
class Config implements ConfigContract, CryptConfigContract, CryptSodiumConfigContract {
    defaultCrypt = SodiumCrypt

    sodiumKey = environment("APP_CRYPT_KEY")
}
```

## Adding an adapter

An adapter implements `CryptContract`.

```text
class AesCrypt implements CryptContract {

    construct(config: CryptAesConfigContract) { }

    encrypt(message, key = null) {
        return aes.encrypt(message, key ?? this.config.aesKey)
    }

    decrypt(encrypted, key = null) {
        return aes.decrypt(encrypted, key ?? this.config.aesKey)
    }

    // ...the structured forms, and the validity check
}
```

Register the adapter and its config in a service provider, then name it in
`defaultCrypt`:

```text
class AppCryptServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            AesCrypt: [self, "publishAesCrypt"],
        }
    }

    static publishAesCrypt(container) {
        container.setSingleton(AesCrypt, new AesCrypt(
            container.getSingleton(CryptAesConfigContract),
        ))
    }
}
```

## Service registration

| Binding              | Is                                            |
| -------------------- | --------------------------------------------- |
| `CryptContract`      | the adapter the component config names        |
| each adapter         | bound under its own name, for an explicit ask |
| the component config | your application's own config class           |
| each adapter config  | your application's own config class           |
