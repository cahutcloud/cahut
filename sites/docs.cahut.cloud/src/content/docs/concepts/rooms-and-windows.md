---
title: Rooms and Windows
description: How Rooms and Windows are named and addressed, the kinds of Windows, and the sandbox that their pages run in.
---

A Room is a space for a person or a team. A Window is a living page in a Room: a document, a deck, a dashboard or a site, made of HTML and its files.

## Addresses

| What | Address |
| --- | --- |
| A Cahut | `https://cahut.example.com/` |
| A Room | `https://cahut.example.com/<room>/` |
| A Window | `https://cahut.example.com/<room>/<window>/` |
| A file of a Window | `https://cahut.example.com/<room>/<window>/img/logo.svg` |
| The MCP endpoint | `https://cahut.example.com/mcp` |
| The health check | `https://cahut.example.com/-/healthz` |

An address without its final slash, such as `https://cahut.example.com/marketing/q3-review`, redirects to the one with it, so that the relative links of the page resolve under the Window.

Today the address of a Cahut answers with a line of text, and the address of a Room with `404 Not Found`. Later, the address of a Room will list the Windows that its reader may open. Paths under `/-/` belong to Cahut itself.

## Names

The names of Rooms and Windows take 1 to 63 lowercase ASCII letters, digits and hyphens, and start and end with a letter or a digit. A Room's name is unique in its Cahut, and a Window's name is unique in its Room.

- A personal Room takes the handle of its person, from 2 to 32 characters, starting with a letter.
- No Room can take the names `mcp`, `api`, `admin`, `console` or `help`.
- A Window takes its name from its title: up to 40 characters, with accents folded, and `-2`, `-3` added when the name is taken. "Q3 review: Été" becomes `q3-review-ete`.

A Window's name holds no random part, so its address can be guessed. Access levels are meant to protect Windows, not hard-to-guess addresses: see [Who can open a Window](/concepts/access/).

## Rooms

Each person has a personal Room, created with them, which they own. Their AI publishes there by default, and can publish in any Room the person belongs to. Team Rooms, with their owners and members, come with accounts and the console.

## Windows

A Window has a kind: `document`, `deck`, `dashboard` or `site`. The kind says what the Window is; Cahut serves every kind the same way. Apps that run on the server, in containers, come later.

A Window is static: HTML, styles, scripts, images, fonts and data files, served as they are, each with the media type of its extension. Readers' browsers check for a newer copy on every visit. A folder opens on its `index.html`; Cahut lists no folder, and a folder without `index.html` answers `404 Not Found`.

Today a Window has a single version and cannot change. Updates that keep the previous version, and a rollback to it, come later.

## The sandbox

The pages of every Window come from the domain of their Cahut, and a browser would let them read each other. So every file of a Window is served with these headers:

```http
Content-Security-Policy: sandbox allow-scripts allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads
Access-Control-Allow-Origin: *
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Cache-Control: no-cache
```

The `sandbox` directive, without `allow-same-origin`, gives the page an opaque origin:

- Its scripts cannot read the other Windows of the Cahut, nor the session of its reader.
- `localStorage`, `sessionStorage`, IndexedDB and `document.cookie` throw a `SecurityError`. A page keeps its state in memory or in its address.
- Scripts, forms, dialogs, downloads, and links that open in a new tab keep working.
- Module scripts, `fetch` calls and web fonts become cross-origin requests. `Access-Control-Allow-Origin: *` lets them load the files of the Window, without credentials.
