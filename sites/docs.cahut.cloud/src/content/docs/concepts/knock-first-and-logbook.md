---
title: Knock first and the Logbook
description: AIs act through bounded operations, ask before anything they cannot undo, and every call goes to the Logbook.
sidebar:
  label: Knock first and Logbook
---

An AI is the intended operator of a Cahut, but never without guardrails. It acts through bounded operations, asks before anything it cannot undo, and leaves a trace of everything it does.

## Bounded operations

An AI acts only through typed operations, such as `publish_window`, each with a name, an input and an output. It never gets a shell, a database or the files of the server. The MCP tools and the command line are thin layers over one registry of operations, which:

1. checks who calls: the administrator, a person, or an AI acting on behalf of a person;
2. refuses what the caller may not do, such as an operation reserved to the administrator;
3. runs the operation, in a database transaction;
4. writes the Logbook entry, in the same transaction for a change.

No way in can skip a step, since none has another way to act. The [operations reference](/reference/operations/) lists the operations of a Cahut and of the server.

## Knock first

What can be undone, the AI does on its own. What cannot waits for a person to approve it.

Each operation declares whether it can be undone. When an AI calls one that cannot, the registry refuses the call: the error says that it needs the approval of a person, and the Logbook records the call as refused. No operation is irreversible today, and approvals are not available yet.

As planned, an irreversible call will create a request that waits for a person, with an expiry. The AI gives the user a link, a person approves or declines in the console, and the Logbook records both steps. These will need an approval:

- deleting a Window for good;
- making a Window public;
- adding a person to a Cahut;
- changing the permissions of a person or of an AI;
- erasing backups, or changing where they go.

## The Logbook

Each Cahut has a Logbook, and the server has one for its own operations. An entry says:

| Column | What it says |
| --- | --- |
| `TIME` | When the call started |
| `OPERATION` | The operation, such as `publish_window` |
| `WHO` | `admin`, a person, or an AI: the name of its token and its software, for the person it acts for |
| `VIA` | The way in: `mcp` for AIs, `cli` for the command line |
| `TARGET` | What the operation acted on, such as `you/q3-review` |
| `OUTCOME` | `ok`, `refused` or `failed`, with the error |

An entry also keeps a detail, such as the kind, title, number of files and size of a published Window.

What goes to the Logbook:

- every call of an AI, reads included, whatever its outcome;
- every change made by the administrator, including those refused or failed.

The reads of the administrator, such as `token list`, stay out of it, so that it stays readable. A change and its entry are written in the same transaction: there is no change without its entry. The Logbook never holds a token, a password, or the contents of a file.

### Read the Logbook

On the server:

```bash
sudo -u cahut /usr/local/bin/cahut logbook --domain cahut.example.com
```

```text
TIME              OPERATION       WHO                                                    VIA  TARGET          OUTCOME
2026-10-21 14:05  publish_window  Claude Code on my laptop (claude-code 2.1.0) for you  mcp  Q3 review       failed: invalid input: index.html is missing: it is the page that opens
2026-10-21 14:02  publish_window  Claude Code on my laptop (claude-code 2.1.0) for you  mcp  you/q3-review   ok
2026-10-21 13:58  create_token    admin                                                  cli  token 1 of you  ok
2026-10-21 13:57  create_user     admin                                                  cli  you             ok
```

The latest entries come first, 20 by default: `--limit` asks for up to 1000. Times are in the server's time zone. `--server` prints the Logbook of the server instead, with the domains declared on it. The console will show the Logbook too.
