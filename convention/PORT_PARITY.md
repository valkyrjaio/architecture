# Port Parity

The **baseline every port reaches**, and the order it reaches it in. This
document replaces the per-language `TODO.md` files, which tracked each port's
gaps separately and so let a port fall behind without anything saying so.

A gap between a port and this baseline is missing work, not a variation. The
baseline does not move to accommodate a port.

---

## What parity means

Two ports are at parity for a component when all of these hold:

1. The component's hierarchy matches its `component/` document.
2. Every canonical contract exists, with every canonical operation on it.
3. Every canonical name matches, transformed only as
   [`PORTS.md`](PORTS.md) permits.
4. The component's tests cover the same behavior the other ports' tests cover.
5. Coverage is 100%, line and branch
   ([`TESTING_METHODOLOGY.md`](TESTING_METHODOLOGY.md)).
6. The component has a `README.md` holding that port's examples.

Parity is per component, not per repository. A port is "at parity for Http" or it
is not.

---

## The component baseline

Every port reaches this set, and a port does not start a component before the
ones it depends on.

| Tier | Components                                            | Why this tier                          |
| ---- | ----------------------------------------------------- | -------------------------------------- |
| 1    | Throwable, Type, Support, Reflection                  | every other component depends on these |
| 2    | Container, Event, Application                         | nothing runs without them              |
| 3    | Cli, Http                                             | the first two protocols                |
| 4    | Log, Validation, Queue, Grpc                          | the remaining first-class protocols    |
| 5    | Cache, Session, Filesystem, Crypt, Jwt                | the service components                 |
| 6    | Auth, Orm, View, Mail, Sms, Broadcast, Api, Attribute | the application components             |

A tier is a dependency order, not a priority. A port finishes a tier before it
starts the next one, because a component in a later tier depends on an earlier
one.

---

## Status is not recorded here

**This document does not say which port has which component.** A status table
goes stale the moment a component lands, because the port that lands it has no
reason to come back and edit this repository. A stale table is worse than no
table: a reader trusts it and plans against it.

Read the port's own source tree for what it has. The component's document under
`component/` says what that component **must** be; the port says what it **is**
today.

---

## The build tool reaches parity too

Each port ships `sindri`, and the tool is at parity when it generates every data
class the port's components need, from the same provider declarations, with the
same output shape ([`SINDRI.md`](SINDRI.md)).

Booting without the cache is not optional, and it is not a parity milestone. It
is a correctness requirement every port satisfies from its first component
([`DATA_CACHE.md`](DATA_CACHE.md)).

---

## What every port ships beside the source

A repository is not at parity on its source alone. Each port also carries:

- the full CI gate its language has, with every check in it
- 100% coverage, line and branch, measured per file
- a `README.md` per component
- a `LIFECYCLE.md` describing that port's entry points and bootstrap
- the license header on every source file
- a `template` repository its new repos are scaffolded from

---

## A deviation is recorded, not discovered

A port that cannot match a name or a shape records the reason in its own
`language/<name>/AGENTS.md`, and the permitted transforms are in
[`PORTS.md`](PORTS.md). A difference that is not recorded is a defect by default.

Warning: "the language cannot do this" is a claim, and it is checked before it is
written. Most apparent limits are a missing abstraction rather than a language
limit.
