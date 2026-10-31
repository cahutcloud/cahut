# 57. The list of tools stays under 40 KB

- Status: Accepted
- Date: 2026-10-31

## Context

[ADR 29](0029-chatgpt-second-ai-client.md) keeps the list of tools that MCP gives an administrator who may write under 32 KB, and a test holds it there. Codex had cut short a list of 87 KB, of which 61 KB were output schemas, which ADR 29 removed.

On 31 October 2026 the list weighs about 31.8 KB, for 38 tools. Descriptions make about 16 KB of it; the rest is structure that every tool carries: its name, its title twice, its four hints, and the types of its input.

The data of Windows ([ADR 56](0056-a-database-for-each-window.md)) adds `data` to the declaration of `publish_window` and `update_window`, about 1.4 KB in each, then three tools: `query_window_data`, `change_window_data` and `rollback_window_data`. The list would weigh about 37 KB. Shortening every description would win a few kilobytes once, at the cost of what tells an AI how to call each tool.

## Decision

- **The list of tools stays under 40 KB** for an administrator who may write, and the test holds it there. It replaces the bound of 32 KB in ADR 29, and nothing else of it.
- A tool still carries an input schema and no output schema, a title and its hints, as ADR 29 has it.
- Before the bound rises again, the list is tried with the clients of ADR 29 (Claude, ChatGPT, Codex and Mistral Vibe), at its new weight.

## Consequences

- The data of Windows fits, with about 3 KB left for what comes after.
- 40 KB stays under half of the list that Codex cut short. The acceptance of batch 10 tries the list with Codex and Mistral Vibe, as batch 8 did.
- A new tool still weighs on every client that loads the list whole: an operation that a person alone needs stays out of MCP (`PeopleOnly`), and a description says what an AI needs to call it, no more.

## Alternatives considered

- **Shorter descriptions everywhere.** They would win about 3 KB once, and the next tool would need the same again, at the cost of what tells an AI how to call each one.
- **The data of a Window in a tool of its own**, beside `publish_window` and `update_window`. One copy of its schema rather than two, but the queries that the pages of a version call would no longer change with that version, nor go back with a rollback, as ADR 56 has them.
- **A higher bound, such as 64 KB.** More room, but closer to a weight that a client has already cut short, without a client tried at it.
