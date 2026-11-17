# Components

One directory per component, lowercase, holding two files.

| File        | Is                                                                        |
| ----------- | ------------------------------------------------------------------------- |
| `<NAME>.md` | the contract: the hierarchy, the names and the behavior every port builds |
| `README.md` | the master copy of that component's README, which every port carries      |

## The contract

`<NAME>.md` states what a port must build. It carries no code example, because an
example has to pick one language and the spelling then travels further than the
rule.

A change to what a component **is** lands in the contract before it lands in any
port.

## The master README

`README.md` is the document a port ships as that component's README. A port
carries it so that every port's README for one component reads alike, and a
reader who knows one port recognizes another.

A port copies the file and changes four things:

1. Tags each `text` fence with the language, and fills it with that port's
   example.
2. Adds a row for an adapter, an integration, or a config property this port
   ships and the others do not.
3. Names the port's own classes where the master names a role.
4. Repoints the contract link at the branch this port targets.

A port changes nothing else. The prose is the same in every port, so an edit to
it belongs in the master rather than in one port's copy. A port that needs the
prose changed changes it here, and every port takes the change.

**A `text` fence holds pseudo code, not a description of an example.** A fence
carries the statements, in their order, with the names, the messages and the
comments written out. A port translates the syntax and changes nothing else, so
every port's example is a translation of one example.

A fence that described the example instead — "resolving the contract and
recording a severity" — would let each port invent its own, which is the drift
the master exists to prevent. A fence that held one language's real code would
make that language the thing the other ports copy.

The pseudo code uses `//` for a comment, `=` for assignment, `name(argument)` for
a call, and `{ key: value }` for a map. A port spells each of those its own way.

**A fence uses the framework's own shapes, never an invented one.** A provider
declares `publishers`, returning a map of service id to a named method reference,
and a static publisher method per entry. A publisher registers with
`setSingleton`, and a caller resolves with the operation for the kind that was
registered. A fence that invented a shorter shape would teach every port the
wrong one.

## What goes in which file

| Statement                                | Belongs in  |
| ---------------------------------------- | ----------- |
| the subcomponents and what each holds    | `<NAME>.md` |
| the canonical operation names            | both        |
| why an operation exists                  | `<NAME>.md` |
| what a reader does to use the component  | `README.md` |
| the config properties and their defaults | `README.md` |
| what the provider binds                  | `README.md` |
| what a port may legitimately vary        | `<NAME>.md` |

A port's reader reads the README. A port's author reads the contract.
