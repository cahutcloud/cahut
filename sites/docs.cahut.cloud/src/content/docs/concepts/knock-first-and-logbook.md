---
title: Knock first and the Logbook
description: AIs act through bounded operations, ask before anything they cannot undo, and every call goes to the Logbook.
sidebar:
  label: Knock first and Logbook
---

An AI is the intended operator of a Cahut, but never without guardrails. It acts through bounded operations, asks before anything it cannot undo, and leaves a trace of everything it does.

## Bounded operations

An AI acts only through typed operations, such as `publish_window`, each with a name, an input and an output. It never gets a shell, a database or the files of the server. The MCP tools and the command line are thin layers over one registry of operations, which:

1. checks who calls: the administrator, a person, or an AI acting on behalf of a person, and whether that person administers the Cahut or is disabled;
2. refuses what the caller may not do, such as an operation for administrators called by a member, or signing in called by an AI;
3. runs the operation, in a database transaction;
4. writes the Logbook entry, in the same transaction for a change.

No way in can skip a step, since none has another way to act. The [operations reference](/reference/operations/) lists the operations of a Cahut and of the server.

## Knock first

What can be undone, the AI does on its own. What cannot, and what changes what a person may do, waits for the person the AI acts for.

When an AI calls such an operation, the call does not run. Cahut checks it as if it ran, so that a mistake comes back to the AI at once, then makes a Knock: a request that waits 24 hours. The AI gets the link of the Knock, such as `https://cahut.example.com/console/knocks/12`, and gives it to you.

The link opens the Knock in the console, where you sign in if you need to. It says which AI asks, what it asks in plain words, the details of the call, and when it expires. **Approve** runs the call, as the AI asked it, and shows you what it gave back that you need, such as the link where a new person chooses their password. **Decline** drops it. The AI learns the outcome with [`get_knock`](/reference/mcp-tools/#get_knock).

- Only the person the AI acts for answers its Knocks: not another person, not an administrator, and never an AI.
- Cahut checks the call again when you approve it, as things stand then. A Knock may still fail, if the Window is gone or the AI lost a Room: it says why, and nothing changes.
- A Knock that nobody answers expires after 24 hours. A Knock whose AI lost its access before an answer is canceled: its token was revoked, its connection ended, or its person was disabled.
- The same call, made again while its Knock waits, gets the same Knock. An AI has at most 20 Knocks waiting.
- **Knocks**, in the console's navigation, counts those that wait for you, and so does the title of the console's tab.

These operations knock first when an AI calls them:

| Operation | Why |
| --- | --- |
| `set_window_access`, to make a Window public | Once a public Window is read, that cannot be taken back |
| `purge_window` | It deletes a Window in the trash for good |
| `delete_room` | It deletes a team Room for good, with its Windows |
| `create_user`, `set_user_role`, `disable_user`, `enable_user` | They change who acts in the Cahut |
| `add_room_member`, `set_room_member_role`, `remove_room_member` | They change who opens and publishes in a Room |
| `set_ai_reach`, to widen what an AI reaches | It gives an AI more than its person granted |

A person who calls one of these operations in the console acts at once, after a confirmation. The [operations reference](/reference/operations/#knock-first) lists them too. Erasing backups, or changing where they go, will knock first when backups come.

## The Logbook

Each Cahut has a Logbook, and the server has one for its own operations. An entry says:

| Column | What it says |
| --- | --- |
| `TIME` | When the call started |
| `OPERATION` | The operation, such as `publish_window` |
| `WHO` | `admin`, a person, or an AI: the name of its token and its software, for the person it acts for |
| `VIA` | The way in: `mcp` for AIs, `api` for the console, `cli` for the command line |
| `TARGET` | What the operation acted on, such as `you/q3-review` |
| `OUTCOME` | `ok`, `knocked`, `refused` or `failed`, with the error |

An entry also keeps a detail, such as the kind, title, number of files and size of a published Window.

What goes to the Logbook:

- every call of an AI, reads included, whatever its outcome;
- the call of an AI that knocks, as `knocked`, then, under the number of its Knock, its approval, as the call of the AI through the console (`api`), or its decline, as `deny_knock`;
- every change made by the administrator or by a person, including those refused or failed;
- every sign-in, under the person who signs in, refused ones included.

The reads of the administrator and of people, such as `token list`, stay out of it, so that it stays readable. An administrator of the Cahut reads every entry; anyone else reads what they did and what their AIs did. A change and its entry are written in the same transaction: there is no change without its entry. The Logbook never holds a token, a password, or the contents of a file.

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

The latest entries come first, 20 by default: `--limit` asks for up to 1000. Times are in the server's time zone. `--server` prints the Logbook of the server instead, with the domains declared on it. The console shows it too, under **Logbook**, by day and with filters.
