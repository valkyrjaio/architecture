# Orm

The Orm component: what it does, how it is used, and the names and behavior every
language port shares.

This is the definition the ports are built from. The template beside this file,
`README.md`, is where a port's own component document starts. It carries these
headings in this order, so every port's document reads alike.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

Orm reads and writes an application's own records in a relational store. It is a
**service component** on the shape in [`LOG.md`](../log/LOG.md), with the adapters named
for the stores they speak to.

It is the largest component, and the one where a port has the most room to invent
its own shape. The layers below are the shape, and they exist in this order
because each one hides the one beneath it.

---

## Hierarchy

| Subcomponent   | Holds                                                        |
| -------------- | ------------------------------------------------------------ |
| `Entity`       | the record contracts an application's own records implement  |
| `Repository`   | the per-entity read and write surface                        |
| `Manager`      | the component contract: connections, transactions, factories |
| `QueryBuilder` | one builder per statement kind                               |
| `Statement`    | a prepared statement and its result                          |
| `Schema`       | tables, columns, indexes, constraints, migrations            |
| `Registry`     | the entity metadata the component reads                      |
| `Data`         | the component config, and one config per store               |
| `Middleware`   | the stages a persistence operation passes through            |
| `Enum`         | the shared vocabulary: orders, joins, operators              |
| `Constant`     | the names the component shares across its own layers         |
| `Factory`      | construction for the pieces a caller does not build itself   |
| `Provider`     | the service provider                                         |
| `Throwable`    | the component's throwable contract and its exceptions        |

---

## Four layers, each hiding the next

| Layer         | A caller reaches for it when                               |
| ------------- | ---------------------------------------------------------- |
| repository    | the work is about one entity — nearly always               |
| query builder | the query is shaped in a way a repository does not express |
| statement     | the statement is written by hand                           |
| manager       | the work spans entities, or needs a transaction            |

**Reach for the highest layer that expresses the work.** A repository call is
readable and portable across stores; a hand-written statement is neither. Each
step down trades one for the other, and most code never leaves the first.

---

## The entity

An entity is the application's own record. It implements a contract rather than
extending a base, so the application owns the class.

| Operation                | Returns                                            |
| ------------------------ | -------------------------------------------------- |
| `getIdValue`             | the value that identifies this record              |
| `asStorableArray`        | every property, in the form the store writes       |
| `asStorableChangedArray` | only the properties that changed since it was read |

**The changed-only form is what makes an update cheap and safe.** Writing every
property overwrites a column another process changed between the read and the
write; writing only what this caller changed does not.

The component also reads metadata about an entity:

| Metadata          | Names                                         |
| ----------------- | --------------------------------------------- |
| the table         | where the record lives                        |
| the identifier    | the field that identifies the record          |
| the repository    | which repository reads and writes the record  |
| the relationships | which properties hold other records           |
| the unstored      | the properties the store does not hold        |
| the dated fields  | where the created and last changed times live |
| the deleted field | where a soft deletion records its time        |
| the date formats  | how the component writes a stored time        |

**A service provider declares that metadata in the `Registry`, keyed by the
entity class.** The entity does not declare it.

**A port declares the metadata rather than inferring it.** A port that infers it
needs reflection, and the cached path must not require reflection.

A component that defines its own entity contract registers its own metadata in
the same `Registry`.

Two further contracts add behavior an entity opts into:

| Contract      | Adds                                                     |
| ------------- | -------------------------------------------------------- |
| dated         | the times the record was created and last changed        |
| soft deletion | a deletion that marks the record rather than removing it |

An entity that implements neither is a plain record, and the component does not
assume either. A contract states that the behavior applies. The `Registry` names
the fields the component reads and writes, so an entity carries none of those
field names.

---

## The repository

A repository is the surface for one entity:

| Operation     | Does                                                      |
| ------------- | --------------------------------------------------------- |
| `find`        | reads one record by its identifier                        |
| `findBy`      | reads one record by other criteria                        |
| `all`         | reads every record                                        |
| `allBy`       | reads every record matching criteria                      |
| `create`      | writes a new record                                       |
| `update`      | writes the changes to an existing record                  |
| `delete`      | deletes, respecting soft deletion where the entity has it |
| `forceDelete` | removes the record whatever the entity implements         |

**`delete` and `forceDelete` are separate because the intent differs.** An
application that marks records deleted still occasionally needs one gone, and a
caller that means that says so rather than passing a flag.

---

## The manager

The manager is the component contract. It owns what is not about one entity:

| Operation            | Does                                       |
| -------------------- | ------------------------------------------ |
| `createRepository`   | returns the repository for an entity       |
| `createQueryBuilder` | returns a builder                          |
| `prepare`            | prepares a statement                       |
| `query`              | runs one                                   |
| `beginTransaction`   | opens a transaction                        |
| `inTransaction`      | reports whether one is open                |
| `ensureTransaction`  | opens one only if none is open             |
| `commit`             | commits                                    |
| `rollback`           | abandons                                   |
| `lastInsertId`       | reports the identifier the store generated |

**`ensureTransaction` exists because transactions do not nest.** A method that
needs a transaction cannot know whether its caller already opened one, and
opening a second either fails or silently commits early depending on the store.
Asking to be _in_ one rather than to _start_ one is the operation that composes.

---

## Using it

### Configure it

The component config names the **default store**, and each store gets its own
config contract ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Register each entity's metadata

A service provider registers an entity's metadata in the `Registry` before
anything resolves a repository for that entity. The `Registry` infers nothing, so an
unregistered entity fails rather than falling back.

### Work through a repository

Resolve the manager, ask it for the repository for your entity, and work there.
Dropping to a builder or a statement is a decision to make per query, not a
default.

### Write changes, not whole records

Update through the changed-only form, so a write touches the columns this caller
changed. That is the difference between two processes editing one record safely
and one silently undoing the other.

### Wrap a multi-record change in a transaction

Any unit of work that writes more than once wraps it, and uses the ensure form so
it composes with a caller that already opened one.

### Keep statements out of application code

A hand-written statement names tables and columns, so it breaks when the schema
moves and it does not port between stores. Where one is genuinely needed, put it
behind a repository method so exactly one place knows it exists.

### Migrate the schema, do not edit it

The schema subcomponent describes tables, columns, indexes and constraints as
data, and a migration is the unit that changes them. A schema edited by hand in
one environment is a schema that differs between environments.

---

## Permitted variation

| Variation                             | Reason                                                |
| ------------------------------------- | ----------------------------------------------------- |
| which stores have an adapter          | each ecosystem has its own drivers                    |
| the statement dialect                 | each store has its own, which is why a builder exists |
| how a prepared value is bound         | each driver names its own binding mechanism           |
| whether a store generates identifiers | some do, and some take one the application supplies   |

The layers, the entity contracts, the repository surface and the manager's
operations do not vary.
