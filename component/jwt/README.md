# Jwt

Implements
[`JWT.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/jwt/JWT.md).

## Introduction

The Jwt component turns a map of claims into a signed token, and a signed token
back into the claims. The signature is what makes the token worth anything: a
token whose signature does not verify is rejected before its claims are read.

An **adapter** is the thing that signs and verifies. Each one implements
`JwtContract`, and each is registered under its own name.

`JwtContract` itself resolves to the **default** adapter: the one your config
names.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a jwt

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

```text
jwt = container.getSingleton(JwtContract)

token = jwt.encode({ "sub": user.id, "exp": expiry })
```

## Encoding and decoding

| Method   | Does                                | Takes      | Returns    |
| -------- | ----------------------------------- | ---------- | ---------- |
| `encode` | signs the claims into a token       | the claims | the token  |
| `decode` | verifies a token and returns claims | the token  | the claims |

```text
// Sign the claims.
token = jwt.encode({ "sub": user.id, "exp": now + 3600 })

// Verify and read them back.
claims = jwt.decode(token)
```

**`decode` verifies before it returns.** A token whose signature does not verify
raises rather than returning claims, so a caller that holds claims holds verified
ones. The null adapter is the exception, which is why it is bound in a test and
nowhere else.

```text
try {
    claims = jwt.decode(fromHeader)
} catch (throwable) {
    return response.unauthorized()
}
```

### The claims are yours

The component signs what you give it and returns what it verified. It does not
add claims, does not check an expiry, and does not know what a subject is. An
application that wants an expiry puts one in the claims and checks it after
decoding.

## Configuration

The component reads one config contract of its own, one for the adapter, and one
per algorithm family.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default.

### The component config

| Property     | Holds                                 | Default                |
| ------------ | ------------------------------------- | ---------------------- |
| `defaultJwt` | the adapter `JwtContract` resolves to | the port's own adapter |

### Available adapters

| Adapter        | Signs with                             |
| -------------- | -------------------------------------- |
| `<Library>Jwt` | one library the port speaks to         |
| `NullJwt`      | nothing; it signs and verifies nothing |

**There is no log adapter here.** A token is the value, so recording what would
have been signed records the claims and produces nothing a caller can use.

### The algorithm, and a config per family

The adapter's own config names the algorithm, and the algorithm decides which
family the adapter reads its keys from. A shared secret and a key pair are not
the same settings, so each family holds its keys in a config of its own. The
algorithm is a value in the component's closed set.

| Config                 | Property             | Holds                      |
| ---------------------- | -------------------- | -------------------------- |
| the adapter's own      | `<library>Algorithm` | which algorithm signs      |
| a shared-secret family | `<family>Key`        | the secret both sides hold |
| a key-pair family      | `<family>PrivateKey` | the key that signs         |
|                        | `<family>PublicKey`  | the key that verifies      |

**Name the algorithm as well as the keys.** A config that sets a family's keys
and leaves the algorithm alone keeps the adapter's default algorithm, so the
adapter reads a different family and signs with that family's default key.

```text
class Config implements
    ConfigContract,
    JwtConfigContract,
    Jwt<Library>ConfigContract,
    JwtRsConfigContract {

    defaultJwt = <Library>Jwt

    <library>Algorithm = Algorithm.RS256

    rsPrivateKey = file("/etc/keys/jwt.key")
    rsPublicKey = file("/etc/keys/jwt.pub")
}
```

**Choose the family by who verifies.** A shared secret is simpler and means
anything that can verify can also sign. A key pair lets something verify a token
without being able to issue one, which is what you want once more than one
service reads your tokens.

## Adding an adapter

An adapter implements `JwtContract`.

```text
class PasetoJwt implements JwtContract {

    construct(config: JwtPasetoConfigContract) { }

    encode(claims) {
        return paseto.sign(claims, this.config.pasetoPrivateKey)
    }

    decode(token) {
        return paseto.verify(token, this.config.pasetoPublicKey)
    }
}
```

Register the adapter and its config in a service provider, then name it in
`defaultJwt`:

```text
class AppJwtServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            PasetoJwt: [self, "publishPasetoJwt"],
        }
    }

    static publishPasetoJwt(container) {
        container.setSingleton(PasetoJwt, new PasetoJwt(
            container.getSingleton(JwtPasetoConfigContract),
        ))
    }
}
```

## Service registration

| Binding              | Is                                            |
| -------------------- | --------------------------------------------- |
| `JwtContract`        | the adapter the component config names        |
| each adapter         | bound under its own name, for an explicit ask |
| the component config | your application's own config class           |
| the adapter's config | your application's own config class           |
| each family's config | your application's own config class           |
