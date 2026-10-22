---
title: Connect Claude Code
description: Add your Cahut to Claude Code over MCP, sign in to choose what it may do, and publish a first Window. Or connect it with a personal token.
---

Claude Code connects to a Cahut over MCP. It signs in with OAuth, like the other [AI clients](/connect/ai-clients/): you give it the address of the Cahut, and you choose in the console the Rooms it reaches and whether it may publish. A personal token in a header works too, for a machine without a browser.

Your account comes from an administrator, or from the command line on the server, as the `cahut` user. With a single Cahut on the server, `--domain` may be left out.

## 1. Create your account

An AI acts for a person of the Cahut. Create yours, if it does not exist yet:

```bash
sudo -u cahut /usr/local/bin/cahut user create --domain cahut.example.com --handle you --name "Your Name" --admin
```

```text
Created you (Your Name), admin of cahut.example.com, with a personal Room at cahut.example.com/you.

Send you this link, where they choose their password. It works once, until 2026-10-29, and is shown only now.

https://cahut.example.com:8443/activate#...
```

Open the link: you choose your password, and you are signed in to the console. `cahut user link` gives you a new one if it expires.

Your handle also names your personal Room, where your AI publishes by default. It takes 2 to 32 lowercase letters, digits and hyphens, starts with a letter and ends with a letter or a digit. `mcp`, `api`, `admin`, `console`, `help` and `oauth` are reserved.

## 2. Add the Cahut to Claude Code

On the machine where Claude Code runs:

```bash
claude mcp add --transport http cahut https://cahut.example.com/mcp
```

Add `--scope user` to reach the Cahut from all your projects. Then sign in, from Claude Code with `/mcp`, choosing the server and **Authenticate**, or from the shell:

```bash
claude mcp login cahut
```

Your browser opens the console of the Cahut. Sign in, choose what Claude Code may do and in which Rooms, and choose **Connect**: see [Consent](/connect/ai-clients/#consent). Back in Claude Code, `/mcp` shows the server and its tools, such as `publish_window`, `get_window` and `list_windows`. The [MCP tools](/reference/mcp-tools/) lists them all: the AI of an administrator sees a few more, and an AI that may only read sees fewer.

On a machine without a browser, such as a server over SSH, `claude mcp login --no-browser cahut` prints the address to open elsewhere.

## 3. Publish

Ask Claude Code, for example:

```text
Make a one-page summary of this project as a document, and publish it on Cahut.
```

It calls `publish_window` and answers with the address of the Window, such as `https://cahut.example.com/you/project-summary/`. Open it in your browser, and share the link.

The call is in the Logbook of the Cahut:

```bash
sudo -u cahut /usr/local/bin/cahut logbook --domain cahut.example.com
```

```text
TIME              OPERATION       WHO                                        VIA  TARGET               OUTCOME
2026-10-21 14:02  publish_window  Claude Code (claude-code 2.1.307) for you  mcp  you/project-summary  ok
```

[Publish a Window](/connect/publish/) says what the AI sends and the rules its files follow.

## With a personal token

Without OAuth, Claude Code sends a personal token in a header. Create one in the console, under **AIs and access**: it shows once, with this command ready to copy. Or create it on the server:

```bash
sudo -u cahut /usr/local/bin/cahut token create --domain cahut.example.com --user you --name "Claude Code on my laptop"
```

Then:

```bash
claude mcp add --transport http cahut https://cahut.example.com/mcp --header "Authorization: Bearer <token>"
```

A token has all your rights, in every Room. Avoid `--scope project` with it: it writes the token into the project's `.mcp.json`, which may end up in a repository. See [Personal tokens](/connect/tokens/).
