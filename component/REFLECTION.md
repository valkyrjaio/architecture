# Reflection

The Reflection component: what it does, how it is used, and the names and
behavior every language port shares.

This is the definition the ports are built from. Each port's `README.md` for this
component carries the same content with that language's examples added, so the
two read alike and anyone moving between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Reflection asks the language about its own code: what a class declares, what a
method takes, what a parameter's type is.

It exists so that **one component depends on the language's reflection and the
rest do not.** Reflection is slow, it differs sharply between languages, and some
languages have very little of it. Confining it to one place means a port with
weak reflection has one component to work around rather than many.

---

## Hierarchy

| Subcomponent | Holds                                                 |
| ------------ | ----------------------------------------------------- |
| `Reflector`  | the component contract                                |
| `Provider`   | the service provider                                  |
| `Throwable`  | the component's throwable contract and its exceptions |

There is no config and no adapter: there is one way to ask a language about
itself.

---

## Canonical operations

### Asking about a declaration

| Operation          | Describes             |
| ------------------ | --------------------- |
| `forClass`         | a class               |
| `forClassConstant` | a constant on a class |
| `forClassProperty` | a property on a class |
| `forClassMethod`   | a method on a class   |
| `forFunction`      | a free function       |
| `forClosure`       | an inline function    |

One operation per kind of declaration, each named for the kind. The same
`for*` naming is used by the attribute collector, so the two read alike
([`ATTRIBUTE.md`](ATTRIBUTE.md)).

### Asking what something needs

| Operation                       | Returns                                        |
| ------------------------------- | ---------------------------------------------- |
| `getDependencies`               | what a declaration needs in order to be called |
| `getDependenciesFromParameters` | the same, from parameters already in hand      |

**This is the operation the framework actually uses.** It is how the container
can be asked to build something without a caller listing its dependencies, and
how the build tool learns what a provider's method requires.

The second form exists because the caller often already holds the parameters,
having obtained them from somewhere else, and reflecting a second time to get
them back would be the expensive part repeated.

---

## Using it

### Resolve the contract, do not reach for the language

Ask this component. A direct use of the language's reflection in another component
is the thing this component exists to prevent, and it is what makes a port with
weaker reflection hard.

### Do not reflect on a hot path

Reflection is the slowest way to learn anything. Where the answer is needed per
unit of work, the build tool should have computed it already
([`DATA_CACHE.md`](../convention/DATA_CACHE.md)).

**That is the real rule about this component: the cached path should not need
it.** Generated data exists so a deployed application answers these questions
without asking.

### Expect a port to have less of it

A compiled language may be unable to answer some of these at run time at all. A
port in that position answers what it can and relies on the build tool for the
rest, which is why the generated data is not optional there.

---

## Permitted variation

| Variation                                           | Reason                                                        |
| --------------------------------------------------- | ------------------------------------------------------------- |
| what a port can answer at run time                  | each language exposes a different amount about itself         |
| whether an answer comes from generated data instead | a port with little reflection reads what the build tool wrote |

The contract and its operation names do not vary. A port that cannot answer an
operation at run time still declares it, and fails loudly rather than returning
an empty answer.
