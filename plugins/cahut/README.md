# The Cahut plugin for Claude

It connects Claude to your Cahut over MCP, and gives it the skill to make and publish Windows there: documents, decks, dashboards and small sites, each in the style you ask for. See [ADR 22](../../docs/adr/0022-house-style-of-each-cahut.md) and [ADR 49](../../docs/adr/0049-no-house-style.md).

## Install it

In Claude Code, add the marketplace of this repository, then install the plugin:

```text
/plugin marketplace add getcahut/cahut
/plugin install cahut@cahut
```

Claude Code asks for the domain of your Cahut, such as `cahut.example.com`, and connects to `https://<domain>/mcp`. Run `/mcp` to sign in: the console of your Cahut asks which Rooms Claude reaches and whether it may publish. If you added your Cahut with `claude mcp add` before, remove that server, or the tools show twice.

In the Claude apps, where skills are uploaded, each Cahut serves the skill alone at `https://<domain>/-/cahut-skill.zip`, and its console links it. Add the Cahut as a custom connector beside it.

## What it holds

- `skills/cahut/`: the skill, which Claude loads when you ask it to make or publish a Window: the formats, the declaration, the sandbox and Knock first.
- `.mcp.json`: the MCP server of your Cahut, at the domain you give.

## License

These files are copied into your AI client, so they are under the [Apache License 2.0](LICENSE), unlike the rest of the repository ([ADR 15](../../docs/adr/0015-license.md)).
