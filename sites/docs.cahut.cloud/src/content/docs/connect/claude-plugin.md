---
title: The Cahut plugin for Claude
description: Install the Cahut plugin in Claude Code, which connects your Cahut and gives Claude the skill to make Windows there, or upload the skill in the Claude apps.
---

The Cahut plugin gives Claude two things: the connection to your Cahut over MCP, and a skill that teaches it to make Windows there. The skill carries the formats of documents, decks, dashboards and sites, the sandbox of Windows, sharing them and Knock first. Claude makes each Window in the style you ask for, with its CSS among its own files. Other AI clients find the same rules, shorter, in the instructions of the MCP server.

## In Claude Code

Add the marketplace of the Cahut repository, then install the plugin:

```text
/plugin marketplace add getcahut/cahut
/plugin install cahut@cahut
```

Claude Code asks for the domain of your Cahut, such as `cahut.example.com`, without `https://`. The plugin connects to `https://<domain>/mcp`: run `/mcp`, choose the `cahut` server and **Authenticate**. The console of your Cahut asks which Rooms Claude reaches and whether it may publish, as for any [AI client](/connect/ai-clients/#consent).

If you added your Cahut with `claude mcp add` before, as [Connect Claude Code](/connect/claude-code/) shows, remove that server with `claude mcp remove cahut`, or its tools show twice.

The plugin connects to a Cahut over HTTPS. A Cahut on your machine for development, at `http://localhost:8080`, connects with `claude mcp add` instead; the skill still helps there, with `--plugin-dir` pointing to `plugins/cahut` of the repository.

## In claude.ai and Claude Desktop

Each Cahut serves the skill as a zip, at `https://<domain>/-/cahut-skill.zip`. The **AIs and access** screen of the console links it.

1. Download the skill.
2. Upload it in the settings of Claude, where skills are added.
3. Add your Cahut as a custom connector, with the address of its MCP endpoint: see [Connect an AI client](/connect/ai-clients/). On my.cahut.cloud, add My Cahut from the directory of Claude instead: see [Try Cahut on My Cahut](/start/my-cahut/#connect-claude).

## Ask

With the plugin, one request is enough:

```text
Make a deck of six slides on the results of the quarter, in navy and white, and publish it in marketing.
```

Claude makes the deck in navy and white, with its CSS and its script among its files, publishes it with `publish_window`, and gives you its address. The arrows move from slide to slide.

## What the plugin holds

The plugin lives in [`plugins/cahut`](https://github.com/getcahut/cahut/tree/main/plugins/cahut) of the repository, under the Apache License 2.0:

| File | What it is |
| --- | --- |
| `skills/cahut/SKILL.md` | The skill: how to make, publish, change and share a Window |
| `skills/cahut/references/formats.md` | What makes a good document, deck, dashboard and site |
| `.mcp.json` | The MCP server of your Cahut, at the domain you give |
| `.claude-plugin/plugin.json` | The manifest, which asks for the domain |
