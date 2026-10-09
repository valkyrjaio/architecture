# Data Cache

The **cross-language** definition of the generated data classes: what they hold,
what they are named, and the guarantee that an application never needs them.

This document holds no code example, for the reason in
[`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md). The tool that writes these
classes is in [`BUILD_TOOL.md`](BUILD_TOOL.md).

---

## What they are

A data class is a generated source file holding fully resolved registration data:
every container binding, every listener, and every route, with nothing left to
discover.

Loading one replaces the whole boot. Instead of walking the provider tree,
reading every declaration and building each collection, the framework constructs
one object per component and is ready.

---

## One class per component

| Class                | Holds                                            |
| -------------------- | ------------------------------------------------ |
| `AppContainerData`   | every binding, from every service provider       |
| `AppEventData`       | every listener, from every listener provider     |
| `AppHttpRoutingData` | every HTTP route, from every HTTP route provider |
| `AppCliRoutingData`  | every CLI route, from every CLI route provider   |

A port adds one class per protocol it holds, named the same way: the `App`
prefix, the component, and `Data`. A port with gRPC generates a gRPC routing data
class; a port with Queue generates a Queue routing data class.

The prefix is `App` because the class aggregates the **application's** whole
tree, framework providers included. It is not one component's data.

Where the classes are written, and the namespace they take, come from the
application's config ([`APPLICATION.md`](../component/APPLICATION.md)).

---

## Every port boots without them

**The cache is an optimization, never a correctness requirement.** Every port
runs with no generated class at all, by walking the provider tree at run time.

This is the rule that shapes the provider declarations. A provider exposes a
class or constructor reference rather than a string, so the framework can walk
the tree itself and the tool can read the same declaration statically. One
declaration, two readers.

The cached path and the uncached path **behave identically**. A difference between
them is a defect in the tool, not a trade-off.

Where the cache matters is a process that handles one unit of work and exits,
because that process pays the whole boot cost every time. A persistent runtime
pays it once, so the cache is optional there.

---

## What the generated form preserves

Two properties are not negotiable, because they are what make the generated form
equivalent to the live one.

**Resolution is uniform.** A registration is held in the same shape whether it
came from a provider at run time or from a generated class, so no resolution path
needs to know which it found.

**Middleware is appended, never deduplicated.** The generated data mirrors what
the run-time walk produces, including a middleware registered twice. A duplicate
is the application's own bug, and the cache must not quietly differ from the
uncached run by fixing it ([`AGENTS.md`](../AGENTS.md) §2).

---

## Load order

A generated class is loaded before the component that uses it, and the container
data is loaded first, because every other component resolves through the
container.

A layer may ship its own generated data: the framework, a third-party package
built on the framework, and the application each generate their own. A later
layer's data is loaded after an earlier one's, so an application can override
what a package registered.

---

## What a declaration must look like

The tool reads a declaration; it does not run one. So every provider list is a
plain literal, with no variable, no call and no conditional.

A list the tool cannot read is a failure the tool reports, with the provider and
the method named. It is never a silent omission, because a silently missing
binding fails much later and far away.

The full contract, and what the tool does with each list, are in
[`PROVIDERS.md`](PROVIDERS.md) and [`BUILD_TOOL.md`](BUILD_TOOL.md).

---

## A binding made outside a provider cannot be cached

The tool reads the provider tree. A binding registered anywhere else is invisible
to it, so that binding reaches the uncached run and not the cached one.

An application that registers directly therefore works until it generates a
cache, and then breaks. Register through a provider, or accept that the
application cannot use the cache at all.

---

## Permitted variation

| Variation                       | Reason                                                                  |
| ------------------------------- | ----------------------------------------------------------------------- |
| which protocol classes exist    | a port generates one per protocol it holds                              |
| the file extension and layout   | each language writes its own source form                                |
| whether generation is built yet | a port reaches it on the schedule in [`PORT_PARITY.md`](PORT_PARITY.md) |

The class names, what each one aggregates, and the load order do not vary.
