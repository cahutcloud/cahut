---
title: Try Cahut on My Cahut
description: My Cahut, at my.cahut.cloud, is a Cahut you sign up to, without a server. Claude finds it in its directory of connectors; other AI clients connect with its address.
---

My Cahut is a Cahut online at `my.cahut.cloud`, to try Cahut without a server. You sign up, connect your AI, and it publishes your Windows at `https://my.cahut.cloud/<handle>/<window>/`.

My Cahut is offered by [Tetra7](https://www.tetra7.io/), which hosts it in France to support the project. It runs the same Cahut as a server you install, with limits. The open source project stays a personal project of Olivier Durand. The terms, the privacy policy and the support of My Cahut are at `account.cahut.cloud`: [terms](https://account.cahut.cloud/en/legal/terms), [privacy](https://account.cahut.cloud/en/legal/privacy), [support](https://account.cahut.cloud/en/support).

## Sign up

Create your account at [account.cahut.cloud](https://account.cahut.cloud/en/signup), with an e-mail address and a handle. It is free during the discovery phase, without a credit card, and seats are limited. Your handle names your personal Room, the start of the address of your Windows. You sign in to the console at [my.cahut.cloud/console/](https://my.cahut.cloud/console/).

## Connect Claude

In claude.ai and Claude Desktop, My Cahut is in the directory of connectors of Claude: there is no address to copy.

1. Open [My Cahut in the directory of Claude](https://claude.ai/directory/my-cahut), or search for My Cahut among the connectors of Claude.
2. Select **Connect**. Claude sends you to the console of my.cahut.cloud: sign in, then choose the Rooms that Claude reaches and whether it may publish, as [Consent](/connect/ai-clients/#consent) shows.
3. Ask Claude to make something and publish it on My Cahut.

Claude shows My Cahut with the Community label, which it gives to the connectors that passed its automated review.

To give Claude the skill that teaches it to make Windows, upload the skill beside the connector: `https://my.cahut.cloud/-/cahut-skill.zip`, as [The Cahut plugin for Claude](/connect/claude-plugin/#in-claudeai-and-claude-desktop) shows.

## Other AI clients

Other AI clients connect with the address of My Cahut, `https://my.cahut.cloud/mcp`, as [Connect an AI client](/connect/ai-clients/) shows:

- **Claude Code**: `claude mcp add --transport http cahut https://my.cahut.cloud/mcp`, or [the Cahut plugin](/connect/claude-plugin/) with the domain `my.cahut.cloud`.
- **ChatGPT**: a custom MCP server with that address and OAuth, on a Business, Enterprise or Edu plan.
- **Le Chat**: a custom MCP connector with that address.

## What My Cahut holds

My Cahut is an individual Cahut: its people share the Cahut, not their work, and do not see one another.

- Your Rooms are yours alone. Beside your personal Room, the Rooms you create take your handle and a hyphen, such as `helen-projects`. Nobody else joins them.
- A Window opens to its Room, to anyone with its [Key](/concepts/access/), or to the public. An AI that makes it public [knocks first](/concepts/knock-first-and-logbook/#knock-first).
- There are no team Rooms, and the disk space and the number of Rooms and Windows are limited. `get_cahut` gives your limits and your use, so an AI that is refused knows why.

For team Rooms, your own domain and no limits, [install Cahut on your server](/install/server/).
