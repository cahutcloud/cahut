---
title: Where Cahut stands
description: What works in Cahut today, and what comes later.
---

Cahut is built in slices, each one working from the AI to the reader. Four slices work. In the first, an AI publishes a Window over MCP with a personal token, Cahut serves it over HTTPS at its own address, and the call goes to the Logbook. In the second, people sign in to a web console, gather in team Rooms, connect their AIs and read the Logbook, and administrators manage people, Rooms and the server. In the third, AI clients such as claude.ai connect with the address of a Cahut alone, and a person chooses on a consent screen the Rooms they reach and whether they may publish. In the fourth, a Window keeps its previous version and goes back to it, goes to a trash, opens to its Room, its Cahut, the holders of its Key or the public, and declares what it loads and calls.

## What works today

| Area | Today |
| --- | --- |
| Server | One `cahut` service on a Linux server with systemd, installed with `install.sh`, holding one or more Cahuts, each at its own domain |
| HTTPS | A certificate from Let's Encrypt for the domain of each Cahut, obtained and renewed by the service |
| People | Created by an administrator, each with a personal Room named after their handle and a link to choose a password; administrators of the Cahut and of the server, with rights checked on every call |
| Console | Under `/console/` of each Cahut's domain: sign-in, Rooms and their people, Windows with their access, versions and declaration, the Logbook, personal tokens, people and the server. See [The console](/admin/console/) |
| Rooms | Team Rooms with owners and members, created by any person or AI; their members publish in them |
| AI access | OAuth for AI clients, with a consent screen where a person grants read or write access to all their Rooms or some of them, and a list of connected AIs to disconnect; personal tokens too, sent in a header. See [Connect an AI client](/connect/ai-clients/) |
| MCP | The [tools](/reference/mcp-tools/) of Windows, Rooms, people and the Logbook, over Streamable HTTP |
| Windows | Documents, decks, dashboards and static sites, served at `<domain>/<room>/<window>/` in a sandbox, with a policy made from what their AI declares they load and call; two versions, a rollback and a trash. See [Rooms and Windows](/concepts/rooms-and-windows/) |
| Access | Room only by default, internal, with a Key, or public, until a date; readers sign in through the console; an AI cannot make a Window public. See [Who can open a Window](/concepts/access/) |
| Logbook | Every call of an AI, every change and every sign-in, read in the console or on the command line |

## What comes later

These are planned, in about this order, without dates:

- **Knock first.** An AI asks before anything it cannot undo, such as making a Window public or deleting one for good, and a person approves or declines in the console.
- **House style.** The organization's colors, fonts, logo and templates, given to every AI, and on the pages that Cahut shows readers, such as the page of a Key.
- **Installation and updates.** A guided first start in the browser, and updates from the releases, with a rollback.
- **A second AI client** besides Claude, tested from end to end.

After that: apps in containers, comments, search, export to PDF and PowerPoint, single sign-on, and backups to S3 or SFTP storage.
