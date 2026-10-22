---
title: Where Cahut stands
description: What works in Cahut today, and what comes later.
---

Cahut is built in slices, each one working from the AI to the reader. Two slices work. In the first, an AI publishes a Window over MCP with a personal token, Cahut serves it over HTTPS at its own address, and the call goes to the Logbook. In the second, people sign in to a web console, gather in team Rooms, connect their AIs and read the Logbook, and administrators manage people, Rooms and the server.

:::caution[Every Window opens to anyone with its address]
There is no access level yet: anyone who has the address of a Window can open it, whatever its Room. Publish nothing that you would not share by link.
:::

## What works today

| Area | Today |
| --- | --- |
| Server | One `cahut` service on a Linux server with systemd, installed with `install.sh`, holding one or more Cahuts, each at its own domain |
| HTTPS | A certificate from Let's Encrypt for the domain of each Cahut, obtained and renewed by the service |
| People | Created by an administrator, each with a personal Room named after their handle and a link to choose a password; administrators of the Cahut and of the server, with rights checked on every call |
| Console | At port 8443 of each Cahut's domain: sign-in, Rooms and their people, Windows, the Logbook, personal tokens, people and the server. See [The console](/admin/console/) |
| Rooms | Team Rooms with owners and members, created by any person or AI; their members publish in them |
| AI access | Personal tokens, created and revoked in the console or on the command line, sent by the AI in a header; an AI acts with the rights of its person |
| MCP | The [tools](/reference/mcp-tools/) of Windows, Rooms, people and the Logbook, over Streamable HTTP |
| Windows | Documents, decks, dashboards and static sites, served at `<domain>/<room>/<window>/` in a sandbox |
| Logbook | Every call of an AI, every change and every sign-in, read in the console or on the command line |

## What comes later

These are planned, in about this order, without dates:

- **OAuth for AI clients.** An AI such as Claude on the web connects with a sign-in and a consent screen, instead of a token copied by hand.
- **Versions and access.** Updating a Window keeps the previous version, and one step rolls back. Each Window opens to its Room only, to everyone signed in to its Cahut, to anyone with its Key, or to the public. See [Who can open a Window](/concepts/access/).
- **Knock first.** An AI asks before anything it cannot undo, such as making a Window public, and a person approves or declines in the console.
- **House style.** The organization's colors, fonts, logo and templates, given to every AI.
- **Installation and updates.** A guided first start in the browser, and updates from the releases, with a rollback.
- **A second AI client** besides Claude, tested from end to end.

After that: apps in containers, comments, search, export to PDF and PowerPoint, single sign-on, and backups to S3 or SFTP storage.
