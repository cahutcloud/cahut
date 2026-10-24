---
title: The first start
description: The single-use link that the installation prints opens the first start of a Cahut in the browser, where you create your account, set the house style and connect your first AI.
---

The installation ends with a link to the first start of your Cahut, such as `https://cahut.example.com/console/setup#k7n3…`. Open it in your browser: in three steps, it creates your account, sets the house style of your organization and connects your first AI. No command is left to type on the server.

## The link

The link works once, for 24 hours, and only while the Cahut has nobody: it is worth an account of administrator, so it shows in the terminal of the installation and nowhere else. Its code comes after `#`, which browsers never send to the server, and the console takes it out of the address as soon as it reads it.

For a new link, which cancels the previous one:

```bash
sudo -u cahut /usr/local/bin/cahut setup
```

On a server with several Cahuts, name the Cahut with `--domain`.

## 1. You

Your name, a handle and a password of 15 characters at least. The handle is how you sign in, and the name of your personal Room; the console proposes one from your name. Once you set the [e-mail relay](/admin/mail/) of the Cahut, your account takes an address, where a link reaches you when you forget your password.

You become the first person of the Cahut, and its administrator. While the server has no administrator, you also become the administrator of the server, with the same handle and the same password: two accounts, since the people of a Cahut and the administrators of the server are kept apart. Sign in as the administrator of the server at `/console/server/sign-in`, to declare other Cahuts or to update the server. Changing the password of one account leaves the other as it is.

## 2. Your Cahut

The address of the Cahut, where each Room and each Window has its own. For a Cahut at the IP address of the server, the console says how to move it to a domain: see [Without a domain](/install/server/#without-a-domain).

Then the house style, which every Window that your AIs make follows: the name of your organization, its main color and its logo, in SVG, PNG or WebP. *Keep the neutral style* goes on without it. [House style](/concepts/house-style/) covers the rest, such as fonts and guidance for AIs.

## 3. Your first AI

The address of the MCP server of the Cahut, `https://cahut.example.com/mcp`, and the steps for each client:

| Client | Steps |
| --- | --- |
| Claude Code | The [Cahut plugin](/connect/claude-plugin/), which also teaches Claude the house style, or `claude mcp add` then `claude mcp login`. See [Claude Code](/connect/claude-code/) |
| Claude apps | A custom connector with the address, in the settings of claude.ai or of the Claude app, and the skill for Claude to upload beside it |
| ChatGPT | On the web, Plugins, then Add custom MCP server, with the address and OAuth |
| Le Chat | Connectors, then a custom MCP connector with the address |
| Others | The address, with OAuth, or a [personal token](/connect/tokens/) |

The AI sends you to the console to sign in and choose the Rooms it reaches. The first start sees it as soon as it connects. Then ask it to publish something: [Publish a Window](/connect/publish/).

You can skip this step: the overview of the console shows the same steps until you connect an AI, and so does AIs and access.

## When the link does not work

The page says so when the link was used, expired, or belongs to another Cahut. If the first start is done, sign in at `/console/sign-in`. If nobody finished it, run `cahut setup` again on the server.
