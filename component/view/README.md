# View

Implements
[`VIEW.md`](https://github.com/valkyrjaio/architecture/blob/26.x/component/view/VIEW.md).

## Introduction

The View component renders a template with data into text. It returns the text to
the caller, and it sends nothing.

A **renderer** is the component's adapter: the thing that renders. Each one
implements `RendererContract`, each is named for the template engine it drives,
and each is registered under its own name, so an application resolves whichever
one it wants.

`RendererContract` itself resolves to the **default** renderer: the one your
config names. It is the convenience, and it is what lets one engine render in
development while another renders in production, with no call site changing.

The component owns the rendering, and it owns no response. Turning rendered text
into a response is one factory's job, and the response itself belongs to
[Http](../http/README.md).

This document assumes the framework, where the container resolves a renderer and
a provider builds it. Used standalone, the component needs none of that:
construct a renderer and call it.

## Resolving a renderer

A contract is a resolvable set of methods, independent of the renderer that
satisfies it. [Container](../container/README.md) covers contracts, bindings and
the resolve operations.

| Resolve                    | Gets                                       | Varies by                |
| -------------------------- | ------------------------------------------ | ------------------------ |
| `RendererContract`         | the renderer the config names              | environment              |
| a renderer by its own name | that renderer, always                      | nothing                  |
| a contract of your own     | the renderer that contract's binding names | purpose, and environment |

### The default, through `RendererContract`

Use this where the engine is the application's business rather than this call
site's, which is almost everywhere.

```text
renderer = container.getSingleton(RendererContract)

html = renderer.render("order/show", { order: order })
```

### A named renderer, where the engine is the point

Use this where this call site means that engine whatever the environment is. A
template written in one engine's syntax only renders in that engine.

```text
renderer = container.getSingleton(OrkaRenderer)

html = renderer.render("order/show", { order: order })
```

### A contract of your own, per purpose

Declare a contract per kind of rendered thing, and bind each to whichever
renderer that kind wants. Then a page and a mail body render through different
renderers without a call site knowing.

```text
interface PageRendererContract extends RendererContract { }
interface MailBodyRendererContract extends RendererContract { }
```

```text
class AppViewServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            PageRendererContract: [self, "publishPageRenderer"],
            MailBodyRendererContract: [self, "publishMailBodyRenderer"],
        }
    }

    static publishPageRenderer(container) {
        container.setSingleton(
            PageRendererContract,
            container.getSingleton(OrkaRenderer),
        )
    }

    static publishMailBodyRenderer(container) {
        container.setSingleton(
            MailBodyRendererContract,
            container.getSingleton(NullRenderer),
        )
    }
}
```

The publishers are where an environment decides what each kind renders through.
A mail body bound to the null renderer renders nothing, while every page still
renders.

## Rendering

| Operation        | Does                                      | Takes                      | Returns           |
| ---------------- | ----------------------------------------- | -------------------------- | ----------------- |
| `render`         | renders a named template with data        | a name, and variables      | the text          |
| `renderFile`     | renders one template file by itself       | a file name, and variables | the text          |
| `createTemplate` | returns a template to build up and render | a name, and variables      | a template        |
| `startRender`    | begins capturing output                   | nothing                    | nothing           |
| `endRender`      | ends capturing                            | nothing                    | the captured text |

```text
// Render a page with its data.
html = renderer.render("order/show", { order: order })

// Build the template up first, where the page has structure to set.
template = renderer.createTemplate("order/show", { order: order })

template.setLayout("layout/app")

html = template.render()
```

`render` is the short form. `createTemplate` is what you call where the page
declares a layout, fills a block, or sets a variable after construction.

`renderFile` renders one template and builds no template object, so no layout
wraps what it returns. The template object calls it for the template itself, for
the layout, and for each partial.

### The capture pair

`startRender` and `endRender` exist for an engine whose templates write to output
rather than returning text. Such a renderer starts the language's output buffer
in `startRender`, and returns the buffer's contents in `endRender`.

**A caller needs the pair only where the caller writes output instead of building
a string.** Call `startRender`, write, then call `endRender` for the text. The
template uses the pair this way for a block, so a block is the output between one
call and the other.

**An engine that returns text needs no capture.** That renderer starts nothing,
and `endRender` returns empty text. So a block captured through such a renderer
holds empty text. Such an engine composes a page with a layout and a block syntax
of its own, and a page built on it uses that syntax rather than the blocks below.

## The template

A template is an object, not a path and a map. It holds the data and the
structure.

| Method                          | Does                                       |
| ------------------------------- | ------------------------------------------ |
| `getName` / `setName`           | reads or sets which template renders       |
| `getVariables` / `setVariables` | reads the variables, or merges more in     |
| `getVariable` / `setVariable`   | reads or sets one variable                 |
| `setLayout`                     | names the template this one renders inside |
| `withoutLayout`                 | clears the layout this template was given  |
| `startBlock` / `endBlock`       | opens and closes one named block           |
| `getBlock`                      | returns what a block holds                 |
| `hasBlock`                      | answers whether a block was filled         |
| `getPartial`                    | renders another template inside this one   |
| `escape`                        | escapes a value for output                 |
| `render`                        | renders the template, and then its layout  |

`setVariables` merges into the variables the template holds. `getVariable`
returns nothing for a key the template does not hold.

### Compose with a layout and blocks

A page declares the layout it lives in and fills the layout's named regions, so
the page describes only what differs. Without blocks a layout can only wrap, and
every page repeats the parts the layout cannot reach.

The page names its layout and renders:

```text
template = renderer.createTemplate("order/show", { order: order })

template.setLayout("layout/app")

html = template.render()
```

Inside the page, each block fills one region of the layout:

```text
template.startBlock("title")

// The page writes the block's content between the two calls.
"Order " + template.escape(order.reference)

template.endBlock()

template.startBlock("content")

template.getPartial("order/lines", { lines: order.lines })

template.endBlock()
```

Inside the layout, each region the page filled is rendered. The layout writes
what each call returns:

```text
template.getBlock("title")

if (template.hasBlock("sidebar")) {
    template.getBlock("sidebar")
}

template.getBlock("content")
```

**The page renders before the layout does.** So every block the page fills is
there when the layout asks for it, and a block the page did not fill makes
`hasBlock` answer false.

Each engine spells these operations as directives in its own syntax. This
project's engine writes them like this, and each directive rewrites to the call
beside it:

| Directive                              | Calls                   |
| -------------------------------------- | ----------------------- |
| `@layout("layout/app")`                | `setLayout`             |
| `@startblock("content")` … `@endblock` | `startBlock` `endBlock` |
| `@block("content")`                    | `getBlock`              |
| `@hasblock("sidebar")` … `@endif`      | `hasBlock`              |
| `@partial("order/lines")`              | `getPartial`            |

### Pass data in, do not reach out

Give the template its variables. A template that resolves a service or reads
global state becomes untestable, and it renders differently depending on when it
ran.

## Escaping

**The component's own engine escapes by default.** It offers an explicit form to
render a value unescaped, so the unsafe path is the one a caller has to name.
Escaping on request is how an injection reaches a page.

| In the component's own engine | Renders the value |
| ----------------------------- | ----------------- |
| `{{ value }}`                 | escaped           |
| `{{{ value }}}`               | unescaped         |

The template's `escape` operation is what the escaped form calls. It raises where
the value cannot be encoded.

Which side of escaping belongs to the engine depends on the engine:

| Engine                                    | Escapes                                    |
| ----------------------------------------- | ------------------------------------------ |
| the component's own engine                | every value, until a caller asks it not to |
| an engine the component wraps             | whatever that engine's own default is      |
| an engine whose template is a source file | nothing                                    |

A plain source file is code, so no engine reads it for a value to escape. Where
an application renders one, escaping is the application's to apply at every
output.

```text
// In a source-file template. Escape each value the application outputs.
template.escape(order.reference)
```

## Returning a response

| Operation                | Does                                   | Takes                                         | Returns    |
| ------------------------ | -------------------------------------- | --------------------------------------------- | ---------- |
| `createResponseFromView` | renders a template, and wraps the text | a name, data, a status code, and a header set | a response |

Build the response through `ViewResponseFactoryContract`, so the status and the
content type are set once rather than at each call site. What the factory returns
is an [Http](../http/README.md) response.

```text
factory = container.getSingleton(ViewResponseFactoryContract)

return factory.createResponseFromView("order/show", { order: order })
```

The status code and the header set are optional. Pass a status code where the
page is not a success:

```text
return factory.createResponseFromView("error/not-found", {}, StatusCode.NOT_FOUND)
```

The factory renders through the renderer it was built with, which is the one
`RendererContract` resolves to.

## Configuration

The component reads one config contract of its own, plus one per engine.

Your application has one config class. Where it implements one of these
contracts, the component's provider binds your class. Where it does not, the
provider binds the component's own default. So implementing a contract is how you
override, and implementing nothing leaves every default in place.

### The component config

| Property          | Holds                                       | Default                           |
| ----------------- | ------------------------------------------- | --------------------------------- |
| `defaultRenderer` | the renderer `RendererContract` resolves to | the port's own source-file engine |

### Changing the default renderer

Name a different renderer in `defaultRenderer`, and every call that resolves
`RendererContract` gets it. No call site changes, and a call that names a
renderer directly is unaffected.

```text
class Config implements ConfigContract, ViewConfigContract {
    defaultRenderer = NullRenderer
}
```

To use a renderer you wrote yourself, name that class instead. It has to be
registered first, which [Adding a renderer](#adding-a-renderer) covers.

### Available renderers

| Renderer           | Renders with                       |
| ------------------ | ---------------------------------- |
| `<Engine>Renderer` | one template engine the port wraps |
| `OrkaRenderer`     | the engine this project ships      |
| `NullRenderer`     | nothing; it renders empty text     |

Every port ships the null renderer. Which engines have a renderer is the port's
own choice, and a port with no engine to wrap ships this project's engine.

A language that reads its own source file as a template counts as an engine here.
The renderer takes the language's name, and so does each of its config
properties.

### Configuring a renderer that reads a source file

Each property carries the engine's name as its prefix, so one config class
implements several of these contracts at once.

| Property                | Holds                                | Default             |
| ----------------------- | ------------------------------------ | ------------------- |
| `<engine>Path`          | the directory that holds a template  | the views directory |
| `<engine>FileExtension` | the extension of a template file     | the engine's own    |
| `<engine>Paths`         | the extra named template directories | nothing             |

### Configuring this project's engine

This project's engine understands a set of **replacements**. Each one matches a
directive in a template, such as `@layout`, and rewrites the directive into the
call it stands for.

| Property               | Holds                                    | Default                |
| ---------------------- | ---------------------------------------- | ---------------------- |
| `orkaPath`             | the directory that holds a template      | the views directory    |
| `orkaFileExtension`    | the extension of a template file         | the engine's own       |
| `orkaPaths`            | the extra named template directories     | nothing                |
| `orkaCoreReplacements` | the engine's own set of replacements     | every core replacement |
| `orkaReplacements`     | the replacements you add to the core set | the debug replacement  |

A template name that starts with `@` names one of the extra directories rather
than the main one. `@admin/order/show` reads `order/show` from whatever
`orkaPaths` maps `@admin` to.

```text
class Config implements ConfigContract, ViewConfigContract, ViewOrkaConfigContract {
    defaultRenderer = OrkaRenderer

    orkaPath = "/resources/templates"
    orkaPaths = { "@admin": "/resources/admin/templates" }
}
```

The engine compiles each template, and it writes the compiled template to the
storage directory. It compiles again on each render while the application runs in
debug mode.

### Configuring a wrapped engine

Each wrapped engine carries the settings that engine needs, prefixed with the
engine's name.

| Property               | Holds                                       | Default               |
| ---------------------- | ------------------------------------------- | --------------------- |
| `<engine>Paths`        | the named template directories              | nothing               |
| `<engine>Extensions`   | the engine extensions to add                | nothing               |
| `<engine>CompiledPath` | where the engine writes a compiled template | the storage directory |

### Configuring the null renderer

The null renderer has no config contract. It renders empty text, so there is
nothing to set.

Name it as the default in a test, or in an environment that renders nothing at
all.

## Adding a renderer

A renderer implements `RendererContract`. Where this port ships a base the
renderers share, extend that instead.

```text
class MustacheRenderer implements RendererContract {

    construct(engine: MustacheEngine) { }

    render(name, variables = {}) {
        return this.renderFile(name, variables)
    }

    renderFile(name, variables = {}) {
        return this.engine.render(name, variables)
    }

    createTemplate(name, variables = {}) {
        return new Template(this, name, variables)
    }

    // The engine returns text, so there is nothing to capture.
    startRender() { }

    endRender() {
        return ""
    }
}
```

Give it a config contract of its own, prefixing each property with the engine's
name so one config class implements several contracts at once:

```text
interface ViewMustacheConfigContract {
    mustachePaths: map<string, string>
    mustacheExtensions: list<string>
    mustacheCompiledPath: string
}
```

Register the renderer, the engine and the config in a service provider. Then name
the renderer in `defaultRenderer`, or resolve it directly where you want it:

```text
class AppViewServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            MustacheRenderer: [self, "publishMustacheRenderer"],
            MustacheEngine: [self, "publishMustacheEngine"],
        }
    }

    static publishMustacheRenderer(container) {
        container.setSingleton(MustacheRenderer, new MustacheRenderer(
            container.getSingleton(MustacheEngine),
        ))
    }

    static publishMustacheEngine(container) {
        config = container.getSingleton(ViewMustacheConfigContract)

        container.setSingleton(MustacheEngine, new MustacheEngine(
            paths: config.mustachePaths,
            compiledPath: config.mustacheCompiledPath,
        ))
    }
}
```

A renderer that wraps an engine takes the engine, not the config. A publisher of
its own builds the engine from the engine's config.

## Service registration

The component's provider binds:

| Binding                       | Is                                            |
| ----------------------------- | --------------------------------------------- |
| `RendererContract`            | the renderer the component config names       |
| each renderer                 | bound under its own name, for an explicit ask |
| each wrapped engine           | built from that engine's config               |
| `ViewResponseFactoryContract` | the factory, over the default renderer        |
| the component config          | your application's own config class           |
| each engine config            | your application's own config class           |

```text
class ViewServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            ViewConfigContract: [self, "publishConfig"],
            ViewOrkaConfigContract: [self, "publishOrkaConfig"],
            RendererContract: [self, "publishRenderer"],
            OrkaRenderer: [self, "publishOrkaRenderer"],
            NullRenderer: [self, "publishNullRenderer"],
            ViewResponseFactoryContract: [self, "publishResponseFactory"],
        }
    }

    static publishConfig(container) {
        config = container.getSingleton(ConfigContract)

        if (config implements ViewConfigContract) {
            container.setSingleton(ViewConfigContract, config)

            return
        }

        container.setSingleton(ViewConfigContract, new ViewConfig())
    }

    static publishRenderer(container) {
        config = container.getSingleton(ViewConfigContract)

        container.setSingleton(
            RendererContract,
            container.getSingleton(config.defaultRenderer),
        )
    }

    static publishOrkaRenderer(container) {
        app = container.getSingleton(ApplicationContract)
        config = container.getSingleton(ViewOrkaConfigContract)

        replacements = []

        for (replacement in concat(config.orkaCoreReplacements, config.orkaReplacements)) {
            replacements.push(container.getSingleton(replacement))
        }

        container.setSingleton(OrkaRenderer, new OrkaRenderer(
            Directory.basePath(config.orkaPath),
            config.orkaFileExtension,
            config.orkaPaths,
            Directory.storagePath("views"),
            app.getDebugMode(),
            ...replacements,
        ))
    }

    static publishResponseFactory(container) {
        container.setSingleton(ViewResponseFactoryContract, new ViewResponseFactory(
            container.getSingleton(ResponseFactoryContract),
            container.getSingleton(RendererContract),
        ))
    }
}
```

### An engine the port ships gets a provider of its own

An engine the component wraps needs a config and a renderer, which the
component's own provider binds. An engine the component **ships** has classes of
its own to register, so it gets a service provider beside the component's, named
for the engine.

That provider binds each replacement under its own name. The renderer asks for
every replacement together, so one publisher binds the set:

```text
class ViewOrkaServiceProvider implements ServiceProviderContract {

    publishers() {
        return {
            Layout: [self, "publishReplacements"],
            StartBlock: [self, "publishReplacements"],
            EndBlock: [self, "publishReplacements"],
            Escaped: [self, "publishReplacements"],
            Unescaped: [self, "publishReplacements"],

            // ...one entry per replacement
        }
    }

    static publishReplacements(container) {
        container.setSingleton(Layout, new Layout())
        container.setSingleton(StartBlock, new StartBlock())
        container.setSingleton(EndBlock, new EndBlock())
        container.setSingleton(Escaped, new Escaped())
        container.setSingleton(Unescaped, new Unescaped())

        // ...one binding per replacement
    }
}
```
