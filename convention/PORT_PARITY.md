# Port Parity

The **baseline every port reaches**, and the order it reaches it in. This
document replaces the per-language `TODO.md` files, which tracked each port's
gaps separately and so let a port fall behind without anything saying so.

A gap here is work. A gap is not a deviation, and it is never closed by changing
this document.

---

## What parity means

Two ports are at parity for a component when all of these hold:

1. The component's hierarchy matches its `component/` document.
2. Every canonical contract exists, with every canonical operation on it.
3. Every canonical name matches, transformed only as
   [`PORTS.md`](PORTS.md) permits.
4. The component's tests mirror the reference port's tests.
5. Coverage is 100%, line and branch ([`AGENTS.md`](../AGENTS.md) §3).
6. The component has a `README.md` holding that port's examples.

Parity is per component, not per repository. A port is "at parity for Http" or it
is not.

---

## The component baseline

Every port reaches this set. The order is the port order in
[`AGENTS.md`](../AGENTS.md) §4, and a port does not start a component before the
ones it depends on.

| Tier | Components                                | Why this tier                           |
| ---- | ----------------------------------------- | --------------------------------------- |
| 1    | Container, Event, Application             | nothing runs without them               |
| 2    | Cli, Http                                 | the first two protocols                 |
| 3    | Throwable, Type, Support, Reflection, Log | every other component depends on these  |
| 4    | Validation, Queue, Grpc                   | the remaining first-class protocols     |
| 5    | Cache, Session, Filesystem, Crypt, Jwt    | the service components                  |
| 6    | Auth, Orm, View, Mail, Sms, Broadcast, Api, Attribute | the application components |

A tier is a dependency order, not a priority. A port finishes a tier before it
starts the next one, because a component in a later tier depends on an earlier
one.

---

## Where each port stands

| Component   | PHP | Java | TypeScript | Go  | Python |
| ----------- | --- | ---- | ---------- | --- | ------ |
| Container   | yes | yes  | yes        | no  | no     |
| Event       | yes | yes  | yes        | no  | no     |
| Application | yes | yes  | yes        | no  | no     |
| Cli         | yes | yes  | yes        | no  | no     |
| Http        | yes | yes  | yes        | no  | no     |
| Throwable   | yes | yes  | yes        | no  | no     |
| Type        | yes | yes  | yes        | no  | no     |
| Support     | yes | yes  | no         | no  | no     |
| Reflection  | yes | yes  | no         | no  | no     |
| Log         | yes | yes  | yes        | no  | no     |
| Validation  | yes | yes  | yes        | no  | no     |
| Queue       | partial | no | no        | no  | no     |
| Grpc        | no  | yes  | in progress | no | no     |
| Cache       | yes | no   | no         | no  | no     |
| Session     | yes | no   | no         | no  | no     |
| Filesystem  | yes | no   | no         | no  | no     |
| Crypt       | yes | no   | no         | no  | no     |
| Jwt         | yes | no   | no         | no  | no     |
| Auth        | yes | no   | no         | no  | no     |
| Orm         | yes | no   | no         | no  | no     |
| View        | yes | no   | no         | no  | no     |
| Mail        | yes | no   | no         | no  | no     |
| Sms         | yes | no   | no         | no  | no     |
| Broadcast   | yes | no   | no         | no  | no     |
| Api         | yes | no   | no         | no  | no     |
| Attribute   | yes | no   | no         | no  | no     |

Go and Python hold a repository, a license and a release workflow, and no
framework source. Queue holds its contracts in PHP and no implementation yet.

Warning: this table goes stale the moment a component lands. Read it as a
starting point and confirm against the port before you rely on a row.

---

## The build tool reaches parity too

Each port ships `sindri`, and the tool is at parity when it generates every data
class the port's components need, from the same provider declarations, with the
same output shape ([`BUILD_TOOL.md`](BUILD_TOOL.md)).

| Port       | Boots without cache | Generates the cache |
| ---------- | ------------------- | ------------------- |
| PHP        | yes                 | yes                 |
| Java       | yes                 | yes                 |
| TypeScript | yes                 | yes                 |
| Go         | —                   | no                  |
| Python     | —                   | no                  |

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
