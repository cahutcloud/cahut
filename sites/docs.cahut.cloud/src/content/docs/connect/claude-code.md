---
title: Connect Claude Code
description: Create a personal token in the console or on the server, add your Cahut to Claude Code over MCP, and publish a first Window.
---

Claude Code connects to a Cahut over MCP, with a personal token in a header. Other MCP clients that can send a header connect the same way.

Your account comes from an administrator, or from the command line on the server, as the `cahut` user. The token then comes from [the console](/admin/console/), or from the command line too. With a single Cahut on the server, `--domain` may be left out.

## 1. Create your account

A token acts for a person of the Cahut. Create yours, if it does not exist yet:

```bash
sudo -u cahut /usr/local/bin/cahut user create --domain cahut.example.com --handle you --name "Your Name" --admin
```

```text
Created you (Your Name), admin of cahut.example.com, with a personal Room at cahut.example.com/you.

Send you this link, where they choose their password. It works once, until 2026-10-29, and is shown only now.

https://cahut.example.com:8443/activate#...
```

Open the link: you choose your password, and you are signed in to the console. `cahut user link` gives you a new one if it expires.

Your handle also names your personal Room, where your AI publishes by default. It takes 2 to 32 lowercase letters, digits and hyphens, starts with a letter and ends with a letter or a digit. `mcp`, `api`, `admin`, `console` and `help` are reserved.

## 2. Create a token

Name the token after the AI and the machine it runs on: the Logbook shows that name.

In the console, open **AIs and access**, give the token its name, and choose **Create a token**. The console shows the token once, with the command of step 3, ready to copy.

On the server, the command line does the same:

```bash
sudo -u cahut /usr/local/bin/cahut token create --domain cahut.example.com --user you --name "Claude Code on my laptop"
```

```text
Token 1 for you (Claude Code on my laptop) in cahut.example.com. It is shown once: copy it now.

cahut_pat_...

The AI sends it in the Authorization header: Bearer <token>.
```

Copy the token now: Cahut keeps only a hash of it. See [Personal tokens](/connect/tokens/).

## 3. Add the Cahut to Claude Code

On the machine where Claude Code runs:

```bash
claude mcp add --transport http cahut https://cahut.example.com/mcp --header "Authorization: Bearer <token>"
```

Add `--scope user` to reach the Cahut from all your projects. Avoid `--scope project`: it writes the token into the project's `.mcp.json`, which may end up in a repository.

In Claude Code, `/mcp` shows the server and its tools, such as `publish_window`, `get_window` and `list_windows`. The [MCP tools](/reference/mcp-tools/) lists them all: the AI of an administrator sees a few more.

## 4. Publish

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
TIME              OPERATION       WHO                                                    VIA  TARGET               OUTCOME
2026-10-21 14:02  publish_window  Claude Code on my laptop (claude-code 2.1.0) for you  mcp  you/project-summary  ok
```

[Publish a Window](/connect/publish/) says what the AI sends and the rules its files follow.

## Other AI clients

Any MCP client that speaks Streamable HTTP and can send an `Authorization` header connects the same way, with a token of its own. Clients that connect only through OAuth, such as connectors added in a web app, need the OAuth server of Cahut, which comes later.
