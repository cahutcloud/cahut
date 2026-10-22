# 1. One registry of typed operations is the only way in

- Status: Accepted, superseded in part by [ADR 13](0013-people-sessions-and-rights.md) (operations that knock first)
- Date: 2026-10-21

## Context

Four product rules shape every line of Cahut: bounded operations, knock first, everything is logged, one API. AIs act over MCP today and over forced-command SSH later. People act through the web console.

If each entry point carried its own logic, each rule would hold only as long as every handler remembered it. The first MCP tool that forgot to write to the Logbook, or the first console screen that skipped a permission check, would break the promise for everyone.

## Decision

Every action on a home is an operation in a single registry.

- An operation declares a stable name (`publish_window`), a typed input and output (Go structs, from which the JSON Schemas of the MCP tools are generated), the permission it needs, and its class: reversible or irreversible.
- The registry runs every call the same way:
  1. Identify the actor: a person, or an AI acting on behalf of a person.
  2. Check the person's permissions and the AI's scope.
  3. Turn an irreversible operation requested by an AI into a Knock, which waits for a person to approve it.
  4. Execute.
  5. Write the Logbook entry.
- The Logbook entry of a change is written in the same database transaction as the change. Refused and failed calls are logged too.
- Every call made by an AI is logged, reads included. Changes made by a person are logged. The console's own reads for display are not, so the Logbook stays readable.
- The MCP tools, the HTTP API used by the console, and later the SSH access are thin adapters. They decode the request, call the registry and encode the result. They hold no business logic and never touch storage.
- People use the same operations from the console. A person who runs an irreversible operation from the console makes the decision themselves, after a confirmation. A Knock is only for AIs.

## Consequences

- The product rules hold by construction. An adapter cannot skip a permission check, the Logbook or Knock first, because it has no other way to act.
- A new capability is one new operation. It shows up in MCP, in the API and in the console at once, and a capability reserved for the console cannot exist.
- Tests target the registry. Adapter tests stay small.
- Operation names and schemas become a contract with AI clients and their skills. Renaming an operation breaks them, so changes go through deprecation first.
- Every feature has to be modeled as operations, reads included. That costs some ceremony for simple screens.

## Alternatives considered

- **Logic in each adapter.** Faster to start, but every rule must be written and tested once per entry point.
- **A console that calls the internals directly.** Breaks the one-API rule without anyone noticing. See [ADR 5](0005-preact-console-embedded.md).
- **A generic tool that takes free-form commands.** Contrary to bounded operations.
