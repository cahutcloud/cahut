---
title: Personal tokens
description: How a personal token lets an AI act for a person in one Cahut, and how to create, list and revoke tokens.
---

A personal token lets an AI act on behalf of one person, in one Cahut. The AI sends it with every request to the MCP endpoint, in the header `Authorization: Bearer <token>`. Without a valid token, the endpoint answers `401 Unauthorized`.

## What a token allows

The AI acts as the person: it publishes in their Rooms, and lists and reads the Windows of their Rooms, through the [MCP tools](/reference/mcp-tools/). It cannot create people or tokens, nor read the Logbook. It cannot reach another Cahut either: a token of one Cahut is unknown to the others.

Every call made with a token goes to the Logbook, reads included. The entry names the token, the software of the AI as it introduces itself, and the person it acts for, such as `Claude Code on my laptop (claude-code 2.1.0) for you`.

## Keep it secret

A token starts with `cahut_pat_`, so that people and secret scanners can recognize one. Cahut shows it once, when it creates it, and keeps only its SHA-256 hash: nobody can read it back, the administrator included.

Treat a token as a password. Give each AI and each machine a token of its own, so that you can revoke one without the others. If a token is lost or leaked, revoke it and create another.

## Create, list and revoke

On the server, as the `cahut` user. Add `--domain` when the server holds several Cahuts.

```bash
sudo -u cahut /usr/local/bin/cahut token create --user you --name "Claude Code on my laptop"
sudo -u cahut /usr/local/bin/cahut token list
sudo -u cahut /usr/local/bin/cahut token revoke --id 1
```

`token list` shows the tokens, newest first, without their secret: the person, the name, when each was created and last used, and whether it is revoked. Cahut records the last use at most once a minute. `--user` keeps the tokens of one person.

```text
ID  USER  NAME                      CREATED     LAST USED         STATE
2   you   Claude Code on my laptop  2026-10-21  2026-10-21 14:02  active
1   you   Claude Code               2026-10-20  never             revoked 2026-10-21
```

A revoked token stops working at once, from the next request of the AI. A revocation cannot be undone: create a new token instead.
