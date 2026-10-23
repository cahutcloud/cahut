# 29. ChatGPT is the second AI client of the V0, through the same tools as Claude

- Status: Accepted
- Date: 2026-10-23

## Context

The [V0 plan](../plan-v0.md) asks that the same Cahut work with Claude and a second AI client, ChatGPT or Le Chat, and leaves the choice open until batch 8.

Claude connects already: claude.ai and Claude Desktop through a custom connector, Claude Code with `claude mcp add` or the `cahut` plugin and its skill. [ADR 16](0016-oauth-for-ai-clients.md) and [ADR 17](0017-oauth-endpoints-on-port-443.md) built the authorization server of each Cahut to what ChatGPT asks as well: its Client ID Metadata Document, `https://chatgpt.com/oauth/client.json`, which authenticates with `private_key_jwt`; `S256`; the `iss` parameter in every authorization response; the `resource` parameter carried into the token. Codex, the coding agent of OpenAI, passed a smoke test in batch 3 through the same document. Le Chat registers by dynamic client registration, which Cahut supports, and has never been tried.

The clients differ in what they do with a server. ChatGPT in developer mode asks the person to confirm every call to a tool that is not marked read-only, and shows the tool's title when it asks. Codex guarantees an AI only the first 512 characters of a server's instructions. Le Chat reads neither MCP resources nor prompts. Only Claude reads a skill. At the opening of batch 8, Codex with a model of OpenAI and Mistral Vibe with a model of Mistral each published a deck in the house style and opened it with a Key, at the first try, against a local Cahut; Codex cut the list of tools short when it read it, 87 KB of which 61 KB were output schemas.

## Decision

- **The second client is ChatGPT**, in developer mode, on the plans of OpenAI that offer it (Plus, Pro, Business, Enterprise and Edu at the time of writing). It is the most used assistant after Claude, its requirements are those the authorization server already meets, and its acceptance in batch 8 closes the V0. Le Chat is tried in the same batch and documented if it works, without being a criterion of the V0.
- **The same tools for every client.** No tool, schema or instruction is made for one client. What a client needs, every client gets, and the registry stays the only way in ([ADR 1](0001-operation-registry.md)).
- **The instructions of the server** say first, in under 512 characters, what an AI needs to publish a Window: call `get_house_style`, publish with relative links and a declaration, give the address to the person, knock first for what cannot be undone. The rest follows.
- **Every tool has a title** in plain English, such as "Publish a Window", from a `Title` that every operation states, and its four hints: read-only, destructive, idempotent, open world.
- **Tools carry an input schema and no output schema.** Their results stay structured, with the same JSON as text for the clients that read only text. The reference of the documentation still describes every output. The list of tools stays under 32 KB for an administrator who may write, and a test holds it there.
- **What a client cannot read** (a skill, resources, prompts) the instructions and the descriptions of the tools carry: the skill for Claude adds to them, it does not replace them.
- **The console and the documentation give the steps for each client**: Claude Code, the Claude apps, ChatGPT, Le Chat, and other MCP clients.

## Consequences

- A change to a tool is checked once, for every client. A client that needs something new raises it for all.
- An AI that reads only the start of the instructions knows enough to publish correctly.
- ChatGPT asks the person before every call that changes something, on top of Knock first; the hints let it run read-only tools without asking.
- An AI that wants the shape of a result learns it from the result itself or from the documentation, not from the list of tools.
- Clients that require an output schema, if any appear, would need one back: the registry still holds the Go type of every output, so it can return without a change of the operations.

## Alternatives considered

- **Le Chat as the second client.** A European assistant fits the self-hosted audience of Cahut, but it is less used, and ChatGPT's requirements, already met, are the stricter of the two. Le Chat is tried anyway.
- **Both as criteria of the V0.** It would hold the milestone to the slowest of two clients that Cahut does not control.
- **Tools tailored to each client**, such as the `search` and `fetch` tools that ChatGPT's deep research once required. Developer mode reads any tool, and two sets of tools would split the registry's single way in.
- **Keeping output schemas.** They make the list of tools four times heavier for the clients that load it whole, and no client tried needs them.
