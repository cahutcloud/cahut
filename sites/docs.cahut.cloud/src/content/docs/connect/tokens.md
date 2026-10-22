---
title: Personal tokens
description: How a personal token lets an AI act for a person in one Cahut, and how to create, list and revoke tokens.
---

A personal token lets an AI act on behalf of one person, in one Cahut, without signing in. Clients that sign in with OAuth, such as claude.ai and Claude Code, [connect with the address of the Cahut](/connect/ai-clients/) instead, to the Rooms and the scope you choose: prefer them. A token suits a script, or a client that cannot sign in. The AI sends it with every request to the MCP endpoint, in the header `Authorization: Bearer <token>`. Without a valid token, the endpoint answers `401 Unauthorized`.

## What a token allows

The AI acts as the person, with all their rights, in every Room of theirs, through the [MCP tools](/reference/mcp-tools/): it publishes in their Rooms, lists and reads the Windows of their Rooms, lists the people of the Cahut, and reads in the Logbook what its person and their AIs did. The AI of an administrator also sees the tools that manage people, but they [knock first](/concepts/knock-first-and-logbook/#knock-first), so it cannot run them yet. No AI creates tokens, signs in or changes a password: these operations are for people. An AI cannot reach another Cahut either: a token of one Cahut is unknown to the others.

Every call made with a token goes to the Logbook, reads included. The entry names the token, the software of the AI as it introduces itself, and the person it acts for, such as `Claude Code on my laptop (claude-code 2.1.0) for you`.

## Keep it secret

A token starts with `cahut_pat_`, so that people and secret scanners can recognize one. Cahut shows it once, when it creates it, and keeps only its SHA-256 hash: nobody can read it back, the administrator included.

Treat a token as a password. Give each AI and each machine a token of its own, so that you can revoke one without the others. If a token is lost or leaked, revoke it and create another.

## Create, list and revoke

In [the console](/admin/console/), **AIs and access** lists your tokens, creates one, shown once with the command that connects Claude Code, and revokes one in one click. An administrator of the Cahut sees and revokes the tokens of others too.

On the server, as the `cahut` user, the command line does the same. Add `--domain` when the server holds several Cahuts.

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

A revoked token stops working at once, from the next request of the AI. A revocation cannot be undone: create a new token instead. The tokens of a disabled person stop working too, until the person is enabled again.
