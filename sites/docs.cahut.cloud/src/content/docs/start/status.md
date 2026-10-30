---
title: Where Cahut stands
description: What works in Cahut today, and what comes later.
---

Cahut is built in slices, each one working from the AI to the reader. Nine slices work. The first eight make the first version, which met the success criteria of the project on a fresh server, with Claude and ChatGPT. In the first, an AI publishes a Window over MCP with a personal token, Cahut serves it over HTTPS at its own address, and the call goes to the Logbook. In the second, people sign in to a web console, gather in team Rooms, connect their AIs and read the Logbook, and administrators manage people, Rooms and the server. In the third, AI clients such as claude.ai connect with the address of a Cahut alone, and a person chooses on a consent screen the Rooms they reach and whether they may publish. In the fourth, a Window keeps its previous version and goes back to it, goes to a trash, opens to its Room, its Cahut, the holders of its Key or the public, and declares what it loads and calls. In the fifth, an AI that asks for something it cannot undo, or that changes what a person may do, knocks first: its person approves or declines in the console. In the sixth, a plugin gives Claude the skill to make Windows, and each Cahut serves the same skill as a zip for the Claude apps. In the seventh, one command installs a server, a link opens the first start in the browser, and the server updates itself from signed releases, going back on its own when a new version fails. In the eighth, ChatGPT connects as a second AI client, through the same tools as Claude, and Le Chat too; a security review hardened sign-in, sessions, OAuth and the limits of the server. In the ninth, each Cahut sends e-mail through the SMTP relay that its administrators set: links to choose a password, invitations to read a Window, and notices.

## What works today

| Area | Today |
| --- | --- |
| Server | One `cahut` service on a Linux server with systemd, installed in one command that checks the signature of the release, holding one or more Cahuts, each at its own domain, or one at the IPv4 address of the server until it has a domain. See [Install on a server](/install/server/) |
| First start | A single-use link that opens the first start in the browser: your account, your first AI. See [The first start](/install/first-start/) |
| Updates | Every day from the signed releases, with a health check and an automatic rollback, from the command line or the console. See [Updates](/admin/updates/) |
| Diagnosis | `cahut doctor` checks the server, the firewall, SELinux, and the address and certificate of each Cahut. See [cahut doctor](/admin/doctor/) |
| HTTPS | A certificate from Let's Encrypt for the domain of each Cahut, or for its IP address, obtained and renewed by the service |
| Kind and limits | Each Cahut collaborative, or individual, where people work alone and do not see one another; limits on its people, administrators, Rooms, Windows and disk; both set by the administrators of the server. See [Cahuts](/concepts/cahuts/#collaborative-or-individual) |
| Legal pages | The legal notice, terms, privacy policy and support of the host, linked at the foot of the console and of the pages that readers see, in English and French, set by the administrators of the server. See [Legal pages](/concepts/cahuts/#legal-pages) |
| People | Created by an administrator, each with a personal Room named after their handle and a link to choose a password; administrators of the Cahut and of the server, with rights checked on every call |
| Console | Under `/console/` of each Cahut's domain: sign-in, Knocks, Rooms and their people, Windows with their access, versions and declaration, the Logbook, personal tokens, people, e-mail, the kind and limits of the Cahut, and the server. See [The console](/admin/console/) |
| Rooms | Team Rooms with owners and members, created by any person or AI, and deleted for good by their owners; their members publish in them. The page of each Room lists the Windows its reader may open, or shows the Window its owners choose. See [The page of a Room](/concepts/rooms-and-windows/#the-page-of-a-room) |
| AI access | OAuth for AI clients such as claude.ai, Claude Code, Mistral Vibe and ChatGPT, with a consent screen where a person grants read or write access to all their Rooms or some of them, and a list of connected AIs to change or disconnect; an AI asks for more by Knock first; personal tokens too, sent in a header. See [Connect an AI client](/connect/ai-clients/) |
| MCP | The [tools](/reference/mcp-tools/) of Windows, Rooms, people and the Logbook, over Streamable HTTP |
| Windows | Documents, decks, dashboards and static sites, served at `<domain>/<room>/<window>/` in a sandbox, with a policy made from what their AI declares they load and call; a title, a description and tags; two versions, a rollback, a trash and deletion for good. See [Rooms and Windows](/concepts/rooms-and-windows/) |
| Access | Room only by default, internal, with a Key, or public, until a date; readers sign in through the console; an AI makes a Window public once its person approves. See [Who can open a Window](/concepts/access/) |
| E-mail | Through the SMTP relay of each Cahut, set by its administrators: activation links, forgotten passwords, invitations to read a Window with a personal link, and notices of Knocks, of the account and of limits. Without a relay, a Cahut sends none and shows no feature of e-mail. See [E-mail](/admin/mail/) |
| Claude | A plugin for Claude Code that connects a Cahut and gives Claude the skill to make Windows, and the skill as a zip for the Claude apps. See [The Cahut plugin for Claude](/connect/claude-plugin/) |
| Moderation | Anyone reports a Window, without an account, on a page that the foot of the pages for readers links; the administrators of the Cahut and of its server read the reports in the console, open the Window whatever its access, and withdraw it with a reason until they reinstate it. See [Moderation](/admin/moderation/) |
| Knock first | What an AI cannot undo, or what changes what a person may do, waits 24 hours for its person, who approves or declines it in the console. See [Knock first and the Logbook](/concepts/knock-first-and-logbook/#knock-first) |
| Logbook | Every call of an AI, every change and every sign-in, read in the console or on the command line |

## What comes later

In about this order and without dates: comments, search, single sign-on, and backups to S3 or SFTP storage; apps in containers; export to PDF and PowerPoint, and SSH access for terminal AIs.
