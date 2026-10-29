# Orm

Implements
[`ORM.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/orm/ORM.md).

## Introduction

The Orm component reads and writes an application's own records in a relational
store. It owns the records, the queries and the schema. It owns nothing about
what a record means to the application.

A **manager** is the thing that speaks to a store. Each one implements
`ManagerContract`, and each is registered under its own name, so an application
resolves whichever one it wants.

`ManagerContract` itself resolves to the **default** manager: the one your config
names. It is the convenience, and it is what lets a test read and write a file on
disk while production reads and writes a database server, with no call site
changing.

An **entity** is the application's own record class. The component reads the
metadata about an entity from the `Registry`, keyed by the entity class, and the
entity declares none of that metadata itself.

The component is four layers, and each layer hides the one below it.

| Layer         | Reach for it when                                          |
| ------------- | ---------------------------------------------------------- |
| repository    | the work is about one entity, which is nearly always       |
| query builder | the query is shaped in a way a repository does not express |
| statement     | the statement is written by hand                           |
| manager       | the work spans entities, or it needs a transaction         |

**Reach for the highest layer that expresses the work.** A repository call reads
plainly, and it runs against every store. A statement written by hand does
neither. Each step down trades one of those for the other, and most code never
leaves the first layer.

This document assumes the framework, where the container resolves a manager and a
provider builds it. Used standalone, the component needs none of that: construct
a manager and call it.

## Resolving a manager

A contract is a resolvable set of methods, independent of the manager that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                   | Gets                                      | Varies by                 |
| ------------------------- | ----------------------------------------- | ------------------------- |
| `ManagerContract`         | the manager the config names              | environment               |
| a manager by its own name | that manager, always                      | nothing                   |
| a contract of your own    | the manager that contract's binding names | database, and environment |

### The default, through `ManagerContract`

Use this where the store is the application's business rather than this call
site's, which is almost everywhere.

```text
orm = container.getSingleton(ManagerContract)

posts = orm.createRepository(Post)
```

### A contract of your own, per database

Declare a contract per database the application reads, and bind each one to the
manager for that database. Then a report that reads a warehouse and a request
that reads the application's own records reach two databases, and neither call
site names a store.

```text
interface ReportingOrmContract extends ManagerContract { }
```

Register it in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppOrmServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            ReportingOrmContract: [self, "publishReportingOrm"],
        }
    }

    static publishReportingOrm(container) {
        container.setSingleton(
            ReportingOrmContract,
            container.getSingleton(PgsqlManager),
        )
    }
}
```

## Registering an entity's metadata

Warning: an entity with no registered metadata fails. The `Registry` infers
nothing, so it reports the unregistered entity rather than guessing a table name.
Register the metadata for every entity in a service provider, before anything
resolves a repository for that entity.

The registry maps an entity class to an `EntityMetadata`.

| Property        | Names                                             |
| --------------- | ------------------------------------------------- |
| `table`         | where the record lives                            |
| `idField`       | the field that identifies the record              |
| `repository`    | which repository reads and writes the record      |
| `relationships` | the properties that hold other records            |
| `unstored`      | the properties the store does not hold            |
| `dated`         | the created and modified fields, and their format |
| `softDelete`    | the deleted field, and its format                 |

`dated` and `softDelete` belong to an entity that implements the matching
contract. [Dated records and soft deletion](#dated-records-and-soft-deletion)
covers both.

| Operation    | Does                                              | Returns        |
| ------------ | ------------------------------------------------- | -------------- |
| `has`        | answers whether the registry holds an entity      | a boolean      |
| `get`        | reads the metadata for an entity                  | the metadata   |
| `withEntity` | returns a registry that also holds one more entry | a new registry |

The registry is immutable. `withEntity` returns a new registry, so one provider
builds the whole registry and registers it under the component's own id.

Warning: one provider owns that id. The container keeps one publisher per id, so
the publisher registered last is the one that runs, and the component's own empty
registry is never built. Register every entity in that one provider, and never
read the registry inside the publisher that registers it.

```text
class AppOrmServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            EntityMetadataRegistryContract: [self, "publishEntityMetadataRegistry"],
        }
    }

    static publishEntityMetadataRegistry(container) {
        container.setSingleton(
            EntityMetadataRegistryContract,
            new EntityMetadataRegistry()
                .withEntity(Post, new EntityMetadata(
                    table: "posts",
                    idField: "id",
                    repository: PostRepository,
                    relationships: ["author"],
                    unstored: ["editorNote"],
                ))
                .withEntity(Comment, new EntityMetadata(
                    table: "comments",
                    idField: "id",
                    repository: Repository,
                )),
        )
    }
}
```

A component that declares an entity of its own holds the metadata for that
entity, and the provider that owns the id registers it with the rest. The user is
such an entity ([Auth](../auth/README.md)).

### The relationships and the unstored properties stay out of a write

`relationships` names each property that holds another record, and `unstored`
names each property the store does not hold. The component drops both lists from
every write, so a property that the table has no column for never reaches a
statement.

The component loads no related record. A method on a repository of your own loads
a relationship, because the application chooses between a join and a second
read.

## Declaring an entity

An entity implements `EntityContract`. Where this port ships a base the entities
share, extend that instead.

| Operation                | Returns                                            |
| ------------------------ | -------------------------------------------------- |
| `getIdValue`             | the value that identifies this record              |
| `asStorableArray`        | every property, in the form the store writes       |
| `asStorableChangedArray` | only the properties that changed since it was read |

```text
class Post implements EntityContract {
    id: int
    title: string
    status: string = "draft"
    body: string|null = null
}
```

The class declares its properties, and it declares the contracts it opts into.
The table, the identifier and the repository are in the registry, so the class
names none of them.

`asStorableArray` also takes a list of properties, and it then returns those
properties alone. Each property converts through the cast the entity declares
([Type](../type/README.md)), so the store holds a flat value where the
application holds a type.

Two further contracts add behavior an entity opts into. Each one declares no
method, and the registry names the fields the component reads and writes.

| Contract                   | Adds                                                     |
| -------------------------- | -------------------------------------------------------- |
| `DatedEntityContract`      | the times the record was created and last changed        |
| `SoftDeleteEntityContract` | a deletion that marks the record rather than removing it |

An entity that implements neither is a plain record, and the component assumes
neither behavior.

## Working through a repository

Ask the manager for the repository for your entity, and do the work there.

| Operation     | Does                                                        | Returns             |
| ------------- | ----------------------------------------------------------- | ------------------- |
| `find`        | reads one record by its identifier                          | the record, or none |
| `findBy`      | reads one record by other conditions                        | the record, or none |
| `all`         | reads every record                                          | the records         |
| `allBy`       | reads every record that matches the conditions              | the records         |
| `create`      | writes a new record                                         | nothing             |
| `update`      | writes the changes to an existing record                    | nothing             |
| `delete`      | deletes, and respects soft deletion where the entity has it | nothing             |
| `forceDelete` | removes the record whatever the entity implements           | nothing             |

```text
posts = orm.createRepository(Post)

post = posts.find(1)

published = posts.allBy(new Where(new Value("status", "published")))
```

Each write builds its statement, prepares it, binds the values and runs it. The
component holds no unit of work, so there is no later step that sends the writes.
A unit of work that writes more than once wraps the writes in a
[transaction](#wrapping-a-multi-record-change-in-a-transaction).

```text
post = new Post()
post.title = "Intro to the Orm"

posts.create(post)

// The identifier the store generated.
id = post.id
```

`create` reads that identifier through the manager's `lastInsertId`, and it
writes the value onto the record.

### `delete` marks, `forceDelete` removes

`delete` reads what the entity implements. For an entity with soft deletion it
stamps the deleted field and it keeps the record. For every other entity it
removes the record.

Warning: `forceDelete` removes the record of an entity with soft deletion, and
the soft deletion then protects nothing. Call it where a law or a data policy
requires the removal, such as a request to erase a person's records.

```text
// Marks the record, for an entity with soft deletion.
posts.delete(post)

// Removes the record, whatever the entity implements.
posts.forceDelete(post)
```

### A repository of your own, per entity

The registry names the repository for an entity, so a repository of your own
holds the queries that entity needs. Extend the shipped repository, add the
methods, and name your class in that entity's metadata.

```text
class PostRepository extends Repository {

    published() {
        return this.allBy(new Where(new Value("status", "published")))
    }
}
```

The repository takes the manager, the entity class and the registry, so it needs
a binding of its own. Register it as a service, because the manager builds one per
entity rather than one per application.

```text
class AppOrmServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            PostRepository: [self, "publishPostRepository"],
        }
    }

    static publishPostRepository(container) {
        container.bind(PostRepository, [self, "createPostRepository"])
    }

    static createPostRepository(container, arguments) {
        [manager, entity, registry] = arguments

        return new PostRepository(manager, entity, registry)
    }
}
```

A query that the operations above cannot express drops to the builder, inside the
repository method. The call site still asks the repository.

## Dropping to a query builder

`createQueryBuilder` returns the factory, and the factory returns one builder per
statement kind. Each method takes the table.

| Method   | Returns            |
| -------- | ------------------ |
| `select` | the select builder |
| `insert` | the insert builder |
| `update` | the update builder |
| `delete` | the delete builder |

Every builder shares these operations. Each `with` operation replaces the list,
and each `withAdded` operation appends to it.

| Operation                      | Sets                           |
| ------------------------------ | ------------------------------ |
| `withFrom`                     | the table                      |
| `withAlias`                    | the alias for the table        |
| `withJoin` / `withAddedJoin`   | the joins                      |
| `withWhere` / `withAddedWhere` | the conditions, and the groups |

The select builder adds the parts a read needs: `withColumns`, `withGroupBy`,
`withOrderBy`, `withLimit` and `withOffset`, each with its `withAdded` form where
the part is a list. The insert builder and the update builder add `withSet`,
which takes the values. The delete builder adds nothing.

```text
select = orm.createQueryBuilder()
    .select("posts")
    .withColumns("posts.*")
    .withJoin(new LeftJoin(
        table: "comments",
        column: "posts.id",
        joinColumn: "comments.post_id",
        comparison: Comparison.EQUALS,
        operator: JoinOperator.ON,
    ))
    .withWhere(new Where(new Value("status", "published")))
    .withOrderBy(new OrderBy("created_at", SortOrder.DESC))
    .withLimit(10)
```

Every operation returns a new builder, so a builder never changes and a shared
base query stays safe to branch from. A builder renders the statement, and it
runs nothing.

**A builder gives up what the repository adds.** The repository hydrates a
record, stamps the dates and respects the soft deletion. A builder does none of
that. A read through a builder asks the statement for the entities, and a write
through a builder stamps the dates itself.

## Dropping to a statement

The manager prepares the statement, and the statement binds the values and runs.

| Operation | Does                                          | Returns       |
| --------- | --------------------------------------------- | ------------- |
| `prepare` | prepares a statement, and runs nothing        | the statement |
| `query`   | prepares a statement and runs it, in one step | the statement |

Bind one value per named parameter, then run the statement, then read the result.
Reuse the parts that built the query, so one place holds each name and its value.

```text
where = new Where(new Value("status", "published"))

select = orm.createQueryBuilder()
    .select("posts")
    .withWhere(where)

// The builder renders the statement, and the manager prepares it.
statement = orm.prepare(select.toString())

statement.bindValue(where.value)
statement.execute()

posts = statement.fetchAllEntities(Post)
```

| Operation          | Returns                                 |
| ------------------ | --------------------------------------- |
| `bindValue`        | whether the value bound                 |
| `execute`          | whether the statement ran               |
| `fetch`            | one row, as the store returned it       |
| `fetchEntity`      | one row, as an entity                   |
| `fetchColumn`      | one column of one row                   |
| `fetchAll`         | every row, as the store returned them   |
| `fetchAllEntities` | every row, as entities                  |
| `getCount`         | the count a counting query returned     |
| `getRowCount`      | how many rows the statement returned    |
| `getColumnCount`   | how many columns the statement returned |
| `getColumnMeta`    | what the store says about one column    |
| `hasError`         | whether the statement failed            |
| `getErrorCode`     | the code the store reported             |
| `getErrorMessage`  | the message the store reported          |

`query` shortens the flow for a statement with no value to bind.

```text
posts = orm.query("SELECT * FROM posts").fetchAllEntities(Post)
```

### Keep a statement out of the application's own code

Warning: a statement written by hand names the tables and the columns, so a
schema that moves breaks it, and no other store runs it. Put the statement behind
a method on a repository of your own, so exactly one place knows the statement
exists.

```text
class PostRepository extends Repository {

    // One place knows this statement.
    live() {
        return this.manager
            .query("SELECT * FROM posts WHERE deleted_at IS NULL")
            .fetchAllEntities(Post)
    }
}
```

## The query's parts are data objects

A condition is an object rather than text. A builder composes the objects, and
nothing writes an operator into a string.

| Object       | Is                                                 |
| ------------ | -------------------------------------------------- |
| `Value`      | a name, and the value bound under that name        |
| `Id`         | a value whose name is the identifier field         |
| `Where`      | one condition: a value, a comparison and a type    |
| `WhereGroup` | conditions in parentheses, and how the group joins |
| `Join`       | one join, and the columns it joins on              |
| `OrderBy`    | one field, and its direction                       |

```text
// status = :status
new Where(new Value("status", "published"))

// id IN (:id0, :id1, :id2)
new Where(new Value("id", [1, 2, 3]), Comparison.IN)

// score >= :score
new Where(new Value("score", 50), Comparison.GREATER_THAN_EQUAL)
```

A `Value` that holds a builder renders that builder as a subquery, so a condition
reads the result of another query.

```text
active = orm.createQueryBuilder()
    .select("authors")
    .withColumns("id")
    .withWhere(new Where(new Value("active", true)))

// author_id IN (SELECT id FROM authors WHERE active = :active)
select = orm.createQueryBuilder()
    .select("posts")
    .withWhere(new Where(new Value("author_id", active), Comparison.IN))
```

### One class per joining word

A condition joins the condition before it. One class per joining word sets the
type, so the class name is the only place a condition states how it joins.

| Class         | Joins with |
| ------------- | ---------- |
| `Where`       | nothing    |
| `NotWhere`    | `NOT`      |
| `AndWhere`    | `AND`      |
| `OrWhere`     | `OR`       |
| `AndNotWhere` | `AND NOT`  |
| `OrNotWhere`  | `OR NOT`   |

A group carries the same six words, and each class wraps its conditions in
parentheses: `WhereGroup`, `NotWhereGroup`, `AndWhereGroup`, `OrWhereGroup`,
`AndNotWhereGroup` and `OrNotWhereGroup`.

Warning: a condition in the wrong place renders a statement the store rejects. A
joining class cannot open a list, because the list already opens with the word
the builder renders. Open each list with a `Where`, a `NotWhere`, a `WhereGroup`
or a `NotWhereGroup`, and join every later part with one of the other classes.

```text
// WHERE (title LIKE :title OR body LIKE :body) AND author_id = :author_id
select = orm.createQueryBuilder()
    .select("posts")
    .withWhere(
        new WhereGroup(
            new Where(new Value("title", "%orm%"), Comparison.LIKE),
            new OrWhere(new Value("body", "%orm%"), Comparison.LIKE),
        ),
        new AndWhere(new Value("author_id", 7)),
    )
```

### The shared vocabulary

| Enum           | Holds                                                              |
| -------------- | ------------------------------------------------------------------ |
| `Comparison`   | the comparisons, such as `EQUALS`, `IN`, `LIKE` and `GREATER_THAN` |
| `WhereType`    | `DEFAULT`, `AND`, `OR`, `NOT`, `AND_NOT` and `OR_NOT`              |
| `JoinType`     | `DEFAULT`, `INNER`, `LEFT`, `RIGHT`, `OUTER` and `FULL_OUTER`      |
| `JoinOperator` | `ON` and `WHERE`                                                   |
| `SortOrder`    | `ASC` and `DESC`                                                   |

One class per join type sets the type for you: `InnerJoin`, `LeftJoin`,
`RightJoin`, `OuterJoin` and `FullOuterJoin`. Construct a `Join` itself where the
type is a value rather than a decision in the code.

## Writing changes, not whole records

`update` writes the properties that changed since the read, and it matches the
record on the identifier field. `asStorableChangedArray` is the form it writes.

```text
post = posts.find(1)
post.title = "A better title"

// UPDATE posts SET title = :title WHERE id = :id
posts.update(post)
```

Writing every property overwrites a column that another process changed between
the read and the write. Writing the changed properties alone does not. That is
the difference between two processes that edit one record safely, and one process
that silently undoes the other.

Warning: an entity with no changed property renders a write with nothing to set,
and no store accepts that statement. Call `update` after a property changed.

## Wrapping a multi-record change in a transaction

| Operation           | Does                           |
| ------------------- | ------------------------------ |
| `beginTransaction`  | opens a transaction            |
| `inTransaction`     | reports whether one is open    |
| `ensureTransaction` | opens one only if none is open |
| `commit`            | commits                        |
| `rollback`          | abandons                       |

Any unit of work that writes more than once wraps the writes, so they all land or
none of them land.

```text
orm.beginTransaction()

try {
    posts.create(post)
    comments.create(comment)

    orm.commit()
} catch (throwable) {
    orm.rollback()

    throw throwable
}
```

### Use the ensure form in anything a caller composes

Warning: transactions do not nest. A second open transaction either fails or
commits the first one early, and which one happens is the store's own behavior.
Call `ensureTransaction` in a method that needs a transaction, because the method
cannot know whether its caller opened one already.

```text
class MemberService {

    construct(orm: ManagerContract) { }

    transfer(from, to) {
        members = this.orm.createRepository(Member)

        // Opens one, or joins the one the caller opened.
        this.orm.ensureTransaction()

        members.update(from)
        members.update(to)

        this.orm.commit()
    }
}
```

## Dated records and soft deletion

An entity opts into each behavior with a contract, and the contract declares no
method. The registry names the fields and the format, so the entity carries
neither.

| The entity implements      | The repository then                                                           |
| -------------------------- | ----------------------------------------------------------------------------- |
| `DatedEntityContract`      | stamps the created and modified fields on a create, the modified on an update |
| `SoftDeleteEntityContract` | stamps the deleted field on a delete, and keeps the record                    |

```text
class Post implements DatedEntityContract, SoftDeleteEntityContract {
    id: int
    title: string
    created_at: string
    updated_at: string
    deleted_at: string|null = null
}
```

Warning: the entity must declare a property under each registered name. The write
reads the properties of the object, so a registered name the entity does not
declare reaches no column. Register the names your own table uses.

```text
new EntityMetadata(
    table: "posts",
    idField: "id",
    repository: Repository,
    dated: new DatedMetadata(
        format: DateFormat.MICROSECOND,
        dateCreatedField: "created_at",
        dateModifiedField: "updated_at",
    ),
    softDelete: new SoftDeleteMetadata(
        format: DateFormat.DEFAULT,
        dateDeletedField: "deleted_at",
    ),
)
```

`DateFormat` holds the formats the component writes. Each one puts the largest
unit first, so a text sort of the column is a date sort, and the date functions of
the store read the value.

| Format        | Fits a column that holds |
| ------------- | ------------------------ |
| `DEFAULT`     | a second                 |
| `MILLISECOND` | a millisecond            |
| `MICROSECOND` | a microsecond            |

**The component writes every time in one zone, and no format records that
zone.** Read every stored time in the zone the component writes, and convert the
value where a person reads it.

Warning: a read returns a record that soft deletion marked. The repository adds
no condition of its own, so a read that wants the live records alone states that
condition itself.

## Schema and migrations

The schema describes the store as data. A migration is the unit that changes it.

| Contract             | Describes                                           |
| -------------------- | --------------------------------------------------- |
| `SchemaContract`     | the tables, and the changes to run                  |
| `TableContract`      | one table, and its columns, indexes and constraints |
| `ColumnContract`     | one column                                          |
| `IndexContract`      | one index                                           |
| `ConstraintContract` | one constraint                                      |

`SchemaContract` creates, reads, renames and drops a table, and `TableContract`
creates, changes and drops a column, an index and a constraint. Each operation
returns the thing it names, so a description reads as one chain.

A migration implements `MigrationContract`, which declares `run` and `rollback`.
Two bases cover the common shapes.

| Base                     | Gives                                                           |
| ------------------------ | --------------------------------------------------------------- |
| `Migration`              | the manager, and the two directions to write                    |
| `TransactionalMigration` | each direction in a transaction, with a rollback on a throwable |

```text
class CreatePostsTable extends TransactionalMigration {

    runMigration() {
        this.orm.query("CREATE TABLE posts (id INT PRIMARY KEY, title VARCHAR(255))")
    }

    rollbackMigration() {
        this.orm.query("DROP TABLE posts")
    }
}
```

`TransactionalMigration` also calls a failure hook before it rethrows, so a
migration logs or cleans up without owning the transaction itself. A port adds a
row for each base it ships and the others do not.

Warning: a schema edited by hand in one environment differs from the schema of
every other environment, and nothing records the difference. Change the schema
through a migration, and never through an edit in one place.

## Configuration

The component reads one config contract of its own, plus one per store.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property         | Holds                                     | Default              |
| ---------------- | ----------------------------------------- | -------------------- |
| `defaultManager` | the manager `ManagerContract` resolves to | the port's own store |

### Changing the default manager

```text
class Config implements ConfigContract, OrmConfigContract {
    defaultManager = PgsqlManager
}
```

### Available managers

| Manager          | Speaks to                    |
| ---------------- | ---------------------------- |
| `<Store>Manager` | one store the port speaks to |
| `NullManager`    | nothing; every read is empty |

Every port ships the null manager. Which stores have a manager is the port's own
choice, so a port adds a row for each one it ships and the others do not. The
three stores below are the ones the ports target, and a port carries the
subsection for each store it ships.

### Configuring the mysql store

| Property        | Holds                                | Default         |
| --------------- | ------------------------------------ | --------------- |
| `mysqlDb`       | the database to connect to           | the application |
| `mysqlHost`     | where to reach the store             | the local host  |
| `mysqlPort`     | the port to reach it on              | the store's own |
| `mysqlUser`     | the user to connect as               | the application |
| `mysqlPassword` | the password to connect with         | the port's own  |
| `mysqlCharset`  | the character set to connect with    | the store's own |
| `mysqlStrict`   | whether to connect in strict mode    | the store's own |
| `mysqlEngine`   | the storage engine to connect with   | the store's own |
| `mysqlOptions`  | the settings the driver itself takes | the port's own  |

```text
class Config implements ConfigContract, OrmConfigContract, OrmMysqlConfigContract {
    defaultManager = MysqlManager

    mysqlDb = "orders"
    mysqlHost = "db.internal"
    mysqlPort = 3306
    mysqlUser = "orders"
    mysqlPassword = environment("ORM_MYSQL_PASSWORD")
}
```

### Configuring the pgsql store

| Property        | Holds                                  | Default         |
| --------------- | -------------------------------------- | --------------- |
| `pgsqlDb`       | the database to connect to             | the application |
| `pgsqlHost`     | where to reach the store               | the local host  |
| `pgsqlPort`     | the port to reach it on                | the store's own |
| `pgsqlUser`     | the user to connect as                 | the application |
| `pgsqlPassword` | the password to connect with           | the port's own  |
| `pgsqlCharset`  | the character set to connect with      | the store's own |
| `pgsqlSchema`   | the schema to search in                | the store's own |
| `pgsqlSslMode`  | how far to insist on an encrypted link | the store's own |
| `pgsqlOptions`  | the settings the driver itself takes   | the port's own  |

### Configuring the sqlite store

| Property         | Holds                                | Default         |
| ---------------- | ------------------------------------ | --------------- |
| `sqliteDb`       | the database to connect to           | the application |
| `sqliteHost`     | where to reach the store             | the local host  |
| `sqlitePort`     | the port to reach it on              | the store's own |
| `sqliteUser`     | the user to connect as               | the application |
| `sqlitePassword` | the password to connect with         | the port's own  |
| `sqliteCharset`  | the character set to connect with    | the store's own |
| `sqliteOptions`  | the settings the driver itself takes | the port's own  |

Warning: SQLite reads a file rather than a server, so the host, the port, the
user and the password reach nothing. Name the file in `sqliteDb`, and leave the
other four alone.

### The null manager

The null manager has no config contract. It reaches no store, so there is nothing
to set.

Name it as the default in a test that runs the repository code with no store
behind it. Every read is empty, and every write and every transaction operation
reports success.

## Adding a manager

A manager implements `ManagerContract`. Where this port ships a base the managers
share, extend that instead and override what the store spells its own way.

```text
class MssqlManager implements ManagerContract {

    construct(connection: Connection, container: ContainerContract) { }

    // The store names its own identifier mechanism.
    lastInsertId(table, idField) {
        return this.query("SELECT SCOPE_IDENTITY()").fetchColumn()
    }

    // ...one method per operation
}
```

Give it a config contract of its own, and prefix each property with the store's
name, so one config class implements several contracts at once.

```text
interface OrmMssqlConfigContract {
    mssqlDb: string
    mssqlHost: string
    mssqlPort: int
    mssqlUser: string
    mssqlPassword: string
}
```

Register the manager and its config in a service provider. Then name the manager
in `defaultManager`, or resolve it where you want it.

```text
class AppOrmServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            MssqlManager: [self, "publishMssqlManager"],
        }
    }

    static publishMssqlManager(container) {
        config = container.getSingleton(OrmMssqlConfigContract)

        container.setSingleton(MssqlManager, new MssqlManager(
            connection: container.getService(Connection, {
                database: config.mssqlDb,
                host: config.mssqlHost,
                // ...one entry per setting the store needs
            }),
            container: container,
        ))
    }
}
```

The statement dialect belongs to the store. The builder exists for that reason.
A store whose dialect differs from the shipped builders also ships builders of
its own.

## Service registration

The component's provider binds:

| Binding                          | Is                                                   |
| -------------------------------- | ---------------------------------------------------- |
| `ManagerContract`                | the manager the component config names               |
| each manager                     | bound under its own name, for an explicit ask        |
| the store connection             | a service, built for the manager that asks           |
| `EntityMetadataRegistryContract` | an empty registry, which your own provider overrides |
| the shipped repository           | a service, built for the entity that asks            |
| the component config             | your application's own config class                  |
| each store config                | your application's own config class                  |

```text
class OrmServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            OrmConfigContract: [self, "publishConfig"],
            OrmMysqlConfigContract: [self, "publishMysqlConfig"],
            ManagerContract: [self, "publishManager"],
            MysqlManager: [self, "publishMysqlManager"],
            NullManager: [self, "publishNullManager"],
            Connection: [self, "publishConnection"],
            EntityMetadataRegistryContract: [self, "publishEntityMetadataRegistry"],
            Repository: [self, "publishRepository"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements OrmConfigContract) {
            container.setSingleton(OrmConfigContract, config)

            return
        }

        container.setSingleton(OrmConfigContract, new OrmConfig())
    }

    static publishManager(container) {
        config = container.getSingleton(OrmConfigContract)

        container.setSingleton(
            ManagerContract,
            container.getSingleton(config.defaultManager),
        )
    }

    static publishMysqlManager(container) {
        config = container.getSingleton(OrmMysqlConfigContract)

        container.setSingleton(MysqlManager, new MysqlManager(
            connection: container.getService(Connection, {
                database: config.mysqlDb,
                host: config.mysqlHost,
                port: config.mysqlPort,
                user: config.mysqlUser,
                password: config.mysqlPassword,
                options: config.mysqlOptions,
            }),
            container: container,
        ))
    }

    static publishEntityMetadataRegistry(container) {
        container.setSingleton(
            EntityMetadataRegistryContract,
            new EntityMetadataRegistry(),
        )
    }

    // A service, because the manager builds one repository per entity.
    static publishRepository(container) {
        container.bind(Repository, [self, "createRepository"])
    }

    static createRepository(container, arguments) {
        [manager, entity, registry] = arguments

        return new Repository(manager, entity, registry)
    }
}
```

The publisher reads the config, and it hands the connection service the settings
the store needs to reach it. A port assembles those settings the way its own
driver takes them.

The connection and the repository are services rather than singletons
([Container](../container/README.md) covers the three kinds of registration).
Each manager asks for a connection built from its own config, and each repository
is built for one entity, so neither one comes back from a cache.
