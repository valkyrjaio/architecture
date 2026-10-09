# Sindri

The **cross-language** definition of `sindri`, the build tool that reads an
application's declarations and writes its data classes.

This document holds no code example, for the reason in
[`DOCUMENTATION_STYLE.md`](DOCUMENTATION_STYLE.md). What the tool produces is in
[`DATA_CACHE.md`](DATA_CACHE.md).

---

## What it is

Every port ships `sindri`, under that name, and **it is never a production
dependency.** It is installed for development and removed from
a deployed application.

The split is strict: **the framework has zero source-reading dependencies.** Every
parser, every syntax-tree walk and every code generator lives in `sindri`. A
framework repository that takes a parser dependency has broken the rule.

The framework therefore holds **no cache-generation command**. Generating is the
tool's job, and the framework never invokes it.

---

## What it does

| Job                       | Means                                                      |
| ------------------------- | ---------------------------------------------------------- |
| generate the data classes | read the provider tree and write one class per component   |
| scaffold an application   | create a new project from the language's template          |
| scaffold a file           | create a provider, a controller or a config from a pattern |
| list its own commands     | report what the tool can do                                |

It is a CLI application built on the framework's own Cli component, so its
commands are routes and its output follows [`CLI.md`](../component/CLI.md).

---

## The config is the entry point

The tool takes one argument: the path to the application's config class. There is
**no tool configuration file** — no YAML, no JSON. Everything the tool needs is
already declared in the config, because the config already lists the component
providers the application registers.

So the tool and the framework read the **same declaration**. The framework walks
it at run time; the tool reads it statically. One source, two readers.

They are not identical walks, though. The tool keeps a record of the providers it
has visited and passes over a repeat; the framework keeps none
([`APPLICATION.md`](../component/APPLICATION.md)). A provider reachable from two
parents is therefore collected once by the tool and twice by the framework, which
is a divergence between the cached and uncached paths rather than an intended
difference.

---

## How it reads

The tool **reads** a declaration. It never runs one.

It resolves the config class to a file, parses that file into a syntax tree, and
reads the return value of each provider list method. It then resolves each
provider named there to its own file and repeats, depth first, in declaration
order, skipping a provider it has already visited.

**A list must be a plain literal** — no variable, no method call, no conditional,
no loop. This is a hard contract, not a preference: a syntax-tree reader can read
a literal and cannot evaluate an expression.

**A declaration the tool cannot read is skipped silently.** An entry whose key or
value is not in the readable shape is passed over, and a provider whose file
cannot be found is passed over, with nothing written about either.

Warning: this is the tool's most dangerous behavior. A binding that is skipped is
absent from the generated class, so the application runs correctly without the
cache and is missing a service with it, and the failure surfaces far from its
cause. Treat a declaration the rules below forbid as a defect you have to find by
reading, because the tool will not tell you.

**Imperative code inside a provider method is invisible to the tool.** A binding
registered there reaches the uncached run only. See
[`DATA_CACHE.md`](DATA_CACHE.md).

---

## How a port reads source

Each language has its own facility, and the choice is the port's:

| Reads source with      | Available where                              |
| ---------------------- | -------------------------------------------- |
| the standard library   | a syntax-tree module ships with the language |
| the compiler's own API | the compiler exposes its tree                |
| a parser library       | neither of the above exists                  |

**The tool needs the framework's own source available.** A language that
distributes compiled artifacts has to obtain the sources explicitly, where a
language that distributes source always has them. That is a packaging
difference, not a difference in what the tool does.

---

## The tool bootstraps itself

The tool is built on the framework, so it has the same boot cost, and generating
its own cache would seem to need its own cache. It does not.

The framework runs without a cache, so the first run is simply slower. The tool
runs once uncached, generates its own data classes, and every later run is fast.
There is no circular dependency, only one slow first run.

---

## Output

The tool's output follows the same structure as any Valkyrja CLI application: the
identity header, the work, then a summary. The rules in
[`CLI.md`](../component/CLI.md) govern it, and these are the ones specific to
generating:

- **One line per generated class**, naming what is being generated.
- **A status per line**, from a closed set: the class was written, the class was
  unchanged and so skipped, or generating it failed.

**Skipped is not a failure.** A generated class whose content has not changed is
not rewritten, so a repeated run reports skipped.

Two gaps are worth stating, because a reader expects both and neither exists. The
failure status carries **no detail** — not the provider, not the file, just the
status. And the run sets **no exit code**: the output is created successful and
never changed, so a run that fails every class still exits successfully. A caller
that needs to know whether generation worked reads the output rather than the
exit status.

---

## Permitted variation

| Variation                              | Reason                                           |
| -------------------------------------- | ------------------------------------------------ |
| the source-reading facility            | each language has its own                        |
| how the framework's source is obtained | each ecosystem packages differently              |
| when generation runs                   | a compiled language may generate at compile time |
| which data classes it writes           | one per protocol the port holds                  |

What the tool reads, and the literal-only contract it reads under, do not vary.
