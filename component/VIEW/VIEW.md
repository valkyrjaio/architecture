# View

The View component: what it does, how it is used, and the names and behavior
every language port shares.

This is the definition the ports are built from. `README.md` beside this file is
the template a port's own component `README.md` starts from: it carries these
headings in this order, so every port's document reads alike and anyone moving
between ports recognizes both.

It holds no code example of its own, because an example has to pick one language,
and the spelling then travels further than the rule.

---

## What the component owns

View renders a template with data into text. It is a **service component** on the
shape in [`LOG.md`](../LOG/LOG.md), with the adapters named for the template engines they
drive.

The component owns the rendering. It owns no response: turning rendered text into
a response is one factory's job, and the response itself belongs to
[`HTTP.md`](../HTTP/HTTP.md).

---

## Hierarchy

| Subcomponent | Holds                                                       |
| ------------ | ----------------------------------------------------------- |
| `Renderer`   | the component contract, and one class per engine            |
| `Template`   | one template, its data, and its structure                   |
| `Factory`    | the factory that turns a rendered view into a response      |
| `Orka`       | the engine this project ships, for a port with none to wrap |
| `Data`       | the component config, and one config per engine             |
| `Provider`   | the service provider                                        |
| `Throwable`  | the component's throwable contract and its exceptions       |

---

## The renderer

| Operation        | Does                                             |
| ---------------- | ------------------------------------------------ |
| `render`         | renders a named template with data               |
| `renderFile`     | renders a template at a path                     |
| `createTemplate` | returns a template object to build up and render |
| `startRender`    | begins capturing output                          |
| `endRender`      | ends capturing, and returns what was captured    |

**The capture pair exists for a language whose templates write to output rather
than returning text.** Where an engine returns a string, the pair wraps it; where
the engine writes, the pair is how the component gets a string back. It is what
lets one contract cover both kinds of engine.

---

## The template

A template is an object, not a path and a map. It holds the data and the
structure:

| Holds         | Is                                              |
| ------------- | ----------------------------------------------- |
| the name      | which template to render                        |
| the variables | the data available while rendering              |
| the layout    | the template this one renders inside            |
| the blocks    | named regions this template fills in its layout |
| the partials  | templates rendered inside this one              |

**A layout and a block are the pair that makes composition work.** A page
declares which layout it lives in and fills the layout's named regions, so the
page describes only what differs. Without blocks, a layout can only wrap, and
every page repeats the parts the layout cannot reach.

**The component's own engine escapes by default.** It offers an explicit form to
render a value unescaped, so the unsafe path is the one a caller has to name. The
reverse default — escaping on request — is how an injection reaches a page.

An engine the component wraps brings its own default, and an engine whose
template is a plain source file escapes nothing. Where an application uses one of
those, escaping is the application's to apply at every output.

A template reports whether a block exists, so a layout can render a region only
where the page filled it.

---

## Using it

### Configure it

The component config names the **default renderer**, and each engine gets its own
config contract ([`COMPONENT_CONFIG.md`](../../convention/COMPONENT_CONFIG.md)).

### Render through the contract

Resolve the component contract. Application code never names an engine, so
changing the engine is a config change.

### Pass data in, do not reach out

Give the template its variables. A template that resolves a service or reads
global state becomes untestable and renders differently depending on when it ran.

### Compose with a layout and blocks

Declare the layout and fill its blocks, rather than including a header and a
footer. A page then states what is different about it and nothing else.

### Let escaping stay on

Render unescaped only where the value is known to be markup the application
produced. Each use of that operation is a place to check.

### Return a response through the factory

Build the response from the rendered view through the factory, so the status and
the content type are set once rather than at each call site.

---

## Permitted variation

| Variation                        | Reason                                                                      |
| -------------------------------- | --------------------------------------------------------------------------- |
| which engines have an adapter    | each ecosystem has its own template engines                                 |
| whether the capture pair is used | an engine returns text, or writes it                                        |
| the template syntax              | the engine owns it, which is why the component wraps rather than defines it |

The contracts, the template's parts and the escape-by-default rule do not vary.
