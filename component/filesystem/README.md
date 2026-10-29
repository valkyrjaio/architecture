# Filesystem

Implements
[`FILESYSTEM.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/filesystem/FILESYSTEM.md).

## Introduction

The Filesystem component reads and writes files, wherever the files live. It
writes to a local disk, to a remote store, to a map in memory, or nowhere at
all, and your application code reads the same in every case.

An **adapter** is the thing that holds the file. Each one implements
`FilesystemContract`, and each is registered under its own name, so an
application resolves whichever one it wants.

`FilesystemContract` itself resolves to the **default** adapter: the one your
config names. It is what lets a development environment write to a local disk
while production writes to a remote store, with no call site changing.

A path is a string, and a caller writes the same path for every adapter. The
adapter decides what the path reaches: a location on a disk, a key in a bucket,
or an entry in a map.

This document assumes the framework, where the container resolves an adapter and
a provider builds it. Used standalone, the component needs none of that:
construct an adapter and call it.

## Resolving a filesystem

A contract is a resolvable set of methods, independent of the adapter that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                      | Varies by                |
| -------------------------- | ----------------------------------------- | ------------------------ |
| `FilesystemContract`       | the adapter the config names              | environment              |
| an adapter by its own name | that adapter, always                      | nothing                  |
| a contract of your own     | the adapter that contract's binding names | purpose, and environment |

### The default, through `FilesystemContract`

Use this where the store is the application's business rather than this call
site's, which is almost everywhere.

```text
filesystem = container.getSingleton(FilesystemContract)

filesystem.write("invoices/4821.pdf", contents)
```

### A named adapter, where the store is the point

Use this where you know this call site means that store whatever the environment
is.

```text
filesystem = container.getSingleton(S3Filesystem)

filesystem.write("invoices/4821.pdf", contents)
```

### A contract of your own, per purpose

Declare a contract per kind of file, and bind each one to whichever adapter that
kind wants. Then an upload a visitor supplied and an export the application
generated live on different disks, and a call site names the purpose rather than
a store.

```text
interface UploadFilesystemContract extends FilesystemContract { }
interface ExportFilesystemContract extends FilesystemContract { }
```

Register each one in a service provider
([Container](../container/README.md) covers providers and `publishers`):

```text
class AppFilesystemServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            UploadFilesystemContract: [self, "publishUploadFilesystem"],
            ExportFilesystemContract: [self, "publishExportFilesystem"],
        }
    }

    static publishUploadFilesystem(container) {
        container.setSingleton(
            UploadFilesystemContract,
            container.getSingleton(S3Filesystem),
        )
    }

    static publishExportFilesystem(container) {
        container.setSingleton(
            ExportFilesystemContract,
            container.getSingleton(LocalFilesystem),
        )
    }
}
```

The publishers are where an environment decides what each purpose writes to. A
test binds both to the in-memory adapter.

## Reading and writing

| Method         | Does                                      | Takes                    | Returns         |
| -------------- | ----------------------------------------- | ------------------------ | --------------- |
| `exists`       | answers whether a path holds a file       | a path                   | a boolean       |
| `read`         | reads the whole file                      | a path                   | the contents    |
| `write`        | writes a file that must not already exist | a path, and the contents | whether it went |
| `update`       | writes a file that must already exist     | a path, and the contents | whether it went |
| `put`          | writes a file whether or not it exists    | a path, and the contents | whether it went |
| `writeStream`  | the stream form of `write`                | a path, and a stream     | whether it went |
| `updateStream` | the stream form of `update`               | a path, and a stream     | whether it went |
| `putStream`    | the stream form of `put`                  | a path, and a stream     | whether it went |

```text
contents = filesystem.read("invoices/4821.pdf")

// Create it. The call fails where something is there already.
filesystem.write("invoices/4822.pdf", contents)

// Replace it. The call fails where nothing is there.
filesystem.update("invoices/4822.pdf", contents)

// Either one.
filesystem.put("invoices/4822.pdf", contents)
```

### Pick the write by what you know about the path

Warning: `put` accepts a path that holds a file already, so an overwrite nobody
intended succeeds. Say `write` where you mean "create this", and say `update`
where you mean "replace this". Keep `put` for the call that does not care.

```text
// Wrong. The second invoice overwrites the first, and nothing reports it.
filesystem.put("invoices/" + order.number + ".pdf", contents)

// Right. A path that holds a file already fails the call.
filesystem.write("invoices/" + order.number + ".pdf", contents)
```

### Stream anything whose size you did not choose

A whole-file read or write holds the file in memory. A file whose size the
application did not choose is streamed, so one request does not exhaust the
memory of a process that serves many.

```text
// The application built this, so its size follows from the data.
filesystem.put("exports/orders.csv", rows)

// A visitor supplied this, so its size is whatever the visitor sent.
filesystem.putStream("uploads/" + upload.name, upload.stream)
```

## Moving, copying and removing

| Method      | Does                                     | Takes                      | Returns         |
| ----------- | ---------------------------------------- | -------------------------- | --------------- |
| `rename`    | moves a file to another path             | the path, and the new path | whether it went |
| `copy`      | duplicates a file to another path        | the path, and the new path | whether it went |
| `delete`    | removes one file                         | a path                     | whether it went |
| `createDir` | creates a directory                      | a path                     | whether it went |
| `deleteDir` | removes a directory and everything in it | a path                     | whether it went |

```text
// Keep the original where it is.
filesystem.copy("invoices/4821.pdf", "archive/2026/4821.pdf")

// Move it, and leave nothing behind.
filesystem.rename("uploads/4821.pdf", "invoices/4821.pdf")

filesystem.delete("uploads/4821.pdf")

filesystem.createDir("archive/2026")
```

Warning: `deleteDir` removes every file under the path it is given. Pass the
directory you mean, and never the one above it.

```text
filesystem.deleteDir("archive/2025")
```

## Asking about a file

| Method         | Reports                                     | Takes                               | Returns        |
| -------------- | ------------------------------------------- | ----------------------------------- | -------------- |
| `metadata`     | everything the adapter knows about the path | a path                              | a map          |
| `mimetype`     | the content type                            | a path                              | the type       |
| `size`         | the size in bytes                           | a path                              | a number       |
| `timestamp`    | when the file last changed                  | a path                              | a timestamp    |
| `visibility`   | who can read the file                       | a path                              | a visibility   |
| `listContents` | what a directory holds                      | a directory, and whether to recurse | one entry each |

```text
if (filesystem.size("uploads/4821.pdf") > maximumUploadSize) {
    return response.badRequest()
}

type = filesystem.mimetype("uploads/4821.pdf")
```

### Ask one question where one answer is enough

An adapter over a network answers one question with one cheap request, and the
whole set with an expensive one. Call `size` where the size is what you need.
Call `metadata` where you need several facts about one path at once.

### Listing a directory

Pass no directory to list the root the adapter is configured with. Pass the
second argument to reach everything below the directory as well.

```text
// What the directory itself holds.
entries = filesystem.listContents("invoices")

// The directory, and everything under it.
entries = filesystem.listContents("invoices", true)

// Everything the adapter holds.
entries = filesystem.listContents()
```

## Visibility

A file is **public** or **private**, and the two are a closed set. Pass the
value the component declares, never a string and never a number. A permission
written as a number is a local-disk idea, and a remote store does not share it.

| Method                 | Does                                            | Takes                    | Returns         |
| ---------------------- | ----------------------------------------------- | ------------------------ | --------------- |
| `setVisibility`        | sets visibility to the value you pass           | a path, and a visibility | whether it went |
| `setVisibilityPublic`  | makes the file readable by anyone               | a path                   | whether it went |
| `setVisibilityPrivate` | makes the file readable only by the application | a path                   | whether it went |

```text
filesystem.setVisibilityPrivate("invoices/4821.pdf")

filesystem.setVisibilityPublic("avatars/4821.png")

// Where another part of the application decided the value.
filesystem.setVisibility("invoices/4821.pdf", visibility)
```

`visibility` reports the value a path carries now.

```text
if (filesystem.visibility("invoices/4821.pdf") == Visibility.PUBLIC) {
    filesystem.setVisibilityPrivate("invoices/4821.pdf")
}
```

### Set the visibility on every write

An adapter gives a new file the default its store gives, and that default
differs by store. State the intent after the write, so the same code gives the
same answer on every adapter.

```text
filesystem.write("invoices/4821.pdf", contents)
filesystem.setVisibilityPrivate("invoices/4821.pdf")
```

## Testing against the in-memory adapter

The in-memory adapter keeps every file in a map that lives as long as the
process. A test writes and reads real files, and it touches no disk, so one case
leaves nothing for the next one to find.

Bind it in place of the default, and the code under test does not change:

```text
container.setSingleton(
    FilesystemContract,
    container.getSingleton(InMemoryFilesystem),
)
```

Construct it with the files the test needs, so a test that reads does not write
first:

```text
filesystem = new InMemoryFilesystem(
    new InMemoryFile("invoices/4821.pdf", contents),
)

contents = filesystem.read("invoices/4821.pdf")
```

A file carries its own metadata, so a test over `mimetype`, `size` or
`visibility` states what the file reports:

```text
filesystem = new InMemoryFilesystem(
    new InMemoryFile(
        "invoices/4821.pdf",
        contents,
        new InMemoryMetadata("application/pdf", 18204, Visibility.PRIVATE),
    ),
)
```

## Configuration

The component reads one config contract of its own, plus one per adapter.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how
you override, and implementing nothing leaves every default in place.

### The component config

| Property            | Holds                                        | Default                           |
| ------------------- | -------------------------------------------- | --------------------------------- |
| `defaultFilesystem` | the adapter `FilesystemContract` resolves to | the port's own local-disk adapter |

### Changing the default adapter

Name a different adapter in `defaultFilesystem`, and every call that resolves
`FilesystemContract` gets it. No call site changes, and a call that names an
adapter directly is unaffected.

```text
class Config implements ConfigContract, FilesystemConfigContract {
    defaultFilesystem = S3Filesystem
}
```

To use an adapter you wrote yourself, name that class instead. It has to be
registered first, which [Adding an adapter](#adding-an-adapter) covers.

### Available adapters

| Adapter                   | Stores in                               |
| ------------------------- | --------------------------------------- |
| `<Integration>Filesystem` | one disk or service the port speaks to  |
| `InMemoryFilesystem`      | a map that lives as long as the process |
| `NullFilesystem`          | nothing; it discards every write        |

Every port ships the in-memory and the null adapters. Which integrations ship is
the port's own choice.

### An integration with variants of its own

A port that reaches several stores through one library ships one adapter per
store, and the library gets a config property that names the variant it resolves
to. The component config then names the library's adapter, and the library's
config names the store.

| Property                         | Holds                                   | Default                           |
| -------------------------------- | --------------------------------------- | --------------------------------- |
| `default<Integration>Filesystem` | the variant the integration resolves to | the port's own local-disk variant |

### Configuring a local-disk adapter

| Property    | Holds                                   | Default                          |
| ----------- | --------------------------------------- | -------------------------------- |
| `localPath` | the directory every path is relative to | the framework's storage location |

A caller never writes this directory into a path. The adapter prepends it, so
the same call works against a remote store that has no such directory.

```text
class Config implements ConfigContract, FilesystemConfigContract, FilesystemLocalConfigContract {
    defaultFilesystem = LocalFilesystem

    localPath = "/storage/app"
}
```

### Configuring a remote-store adapter

Each store adapter carries the settings that store needs, prefixed with the
store's name.

| Property    | Holds                                        | Default    |
| ----------- | -------------------------------------------- | ---------- |
| `s3Key`     | the credential that names the application    | none       |
| `s3Secret`  | the credential that authorizes it            | none       |
| `s3Region`  | where the bucket is                          | none       |
| `s3Version` | which version of the service the client uses | the latest |
| `s3Bucket`  | the bucket every path is written to          | none       |
| `s3Prefix`  | what every path is prefixed with             | nothing    |
| `s3Options` | the further settings the client takes        | nothing    |

The credentials, the region and the bucket have no default. The application sets
each one, or the provider raises rather than binding a config nothing can use.

```text
class Config implements ConfigContract, FilesystemConfigContract, FilesystemS3ConfigContract {
    defaultFilesystem = S3Filesystem

    s3Key = environment("APP_S3_KEY")
    s3Secret = environment("APP_S3_SECRET")
    s3Region = "us-east-1"
    s3Bucket = "orders"
    s3Prefix = "invoices"
}
```

### Configuring the in-memory adapter

The in-memory adapter has no config contract. It takes the files a test gives
it, and it holds nothing else.

### Configuring the null adapter

The null adapter has no config contract. It discards every write, so there is
nothing to set.

Name it as the default in an environment that writes no files at all. A read
against it reports nothing, so a test that asserts on contents binds the
in-memory adapter instead.

## Adding an adapter

An adapter implements `FilesystemContract`. Where this port ships a base the
adapters share, extend that instead.

```text
class SftpFilesystem implements FilesystemContract {

    construct(client: SftpClient, config: FilesystemSftpConfigContract) { }

    exists(path) {
        return this.client.has(this.rooted(path))
    }

    read(path) {
        return this.client.get(this.rooted(path))
    }

    write(path, contents) {
        if (this.exists(path)) {
            return false
        }

        return this.put(path, contents)
    }

    put(path, contents) {
        return this.client.send(this.rooted(path), contents)
    }

    // ...one method per operation

    rooted(path) {
        return this.config.sftpRoot + "/" + path
    }
}
```

Give it a config contract of its own, prefixing each property with the adapter's
name so one config class implements several contracts at once:

```text
interface FilesystemSftpConfigContract {
    sftpHost: string
    sftpRoot: string
}
```

Register the adapter and its config in a service provider, then name the adapter
in `defaultFilesystem`, or resolve it directly where you want it:

```text
class AppFilesystemServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            SftpFilesystem: [self, "publishSftpFilesystem"],
        }
    }

    static publishSftpFilesystem(container) {
        container.setSingleton(SftpFilesystem, new SftpFilesystem(
            container.getSingleton(SftpClient),
            container.getSingleton(FilesystemSftpConfigContract),
        ))
    }
}
```

An adapter whose store knows nothing of a question answers with what it always
has. A store with no notion of visibility reports one value for every path.

## Service registration

The component's provider binds:

| Binding                   | Is                                                     |
| ------------------------- | ------------------------------------------------------ |
| `FilesystemContract`      | the adapter the component config names                 |
| each adapter              | bound under its own name, for an explicit ask          |
| each integration's client | built from that adapter's config, by its own publisher |
| the component config      | your application's own config class                    |
| each adapter config       | your application's own config class                    |

```text
class FilesystemServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            FilesystemConfigContract: [self, "publishConfig"],
            FilesystemLocalConfigContract: [self, "publishLocalConfig"],
            FilesystemS3ConfigContract: [self, "publishS3Config"],
            FilesystemContract: [self, "publishFilesystem"],
            LocalFilesystem: [self, "publishLocalFilesystem"],
            S3Filesystem: [self, "publishS3Filesystem"],
            S3Client: [self, "publishS3Client"],
            InMemoryFilesystem: [self, "publishInMemoryFilesystem"],
            NullFilesystem: [self, "publishNullFilesystem"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements FilesystemConfigContract) {
            container.setSingleton(FilesystemConfigContract, config)

            return
        }

        container.setSingleton(FilesystemConfigContract, new FilesystemConfig())
    }

    static publishFilesystem(container) {
        config = container.getSingleton(FilesystemConfigContract)

        container.setSingleton(
            FilesystemContract,
            container.getSingleton(config.defaultFilesystem),
        )
    }

    // The adapter takes the client, not the config.
    static publishS3Filesystem(container) {
        container.setSingleton(S3Filesystem, new S3Filesystem(
            container.getSingleton(S3Client),
        ))
    }

    // A publisher of its own builds the client from the config.
    static publishS3Client(container) {
        config = container.getSingleton(FilesystemS3ConfigContract)

        container.setSingleton(S3Client, new S3Client(
            key: config.s3Key,
            secret: config.s3Secret,
            region: config.s3Region,
            bucket: config.s3Bucket,
            prefix: config.s3Prefix,
        ))
    }

    static publishInMemoryFilesystem(container) {
        container.setSingleton(InMemoryFilesystem, new InMemoryFilesystem())
    }

    static publishNullFilesystem(container) {
        container.setSingleton(NullFilesystem, new NullFilesystem())
    }
}
```
