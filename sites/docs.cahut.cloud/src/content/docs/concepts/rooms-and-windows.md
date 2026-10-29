---
title: Rooms and Windows
description: How Rooms and Windows are named and addressed, the kinds of Windows, their versions and trash, and the sandbox that their pages run in.
---

A Room is a space for a person or a team. A Window is a living page in a Room: a document, a deck, a dashboard or a site, made of HTML and its files.

## Addresses

| What | Address |
| --- | --- |
| A Cahut, with the Rooms of its reader | `https://cahut.example.com/` |
| A Room, with the Windows its reader may open, or the Window it shows | `https://cahut.example.com/<room>/` |
| A Window | `https://cahut.example.com/<room>/<window>/` |
| A file of a Window | `https://cahut.example.com/<room>/<window>/img/logo.svg` |
| The MCP endpoint | `https://cahut.example.com/mcp` |
| The health check | `https://cahut.example.com/-/healthz` |

An address without its final slash, such as `https://cahut.example.com/marketing/q3-review`, redirects to the one with it, so that the relative links of the page resolve under the Window.

Paths under `/-/` belong to Cahut itself: the sign-in of readers, the page of a Key, the frames of protected Windows. Who may open what is in [Who can open a Window](/concepts/access/).

## Names

The names of Rooms and Windows take 1 to 63 lowercase ASCII letters, digits and hyphens, and start and end with a letter or a digit. A Room's name is unique in its Cahut, and a Window's name is unique in its Room.

- A personal Room takes the handle of its person, from 2 to 32 characters, starting with a letter.
- No Room can take the names `mcp`, `api`, `admin`, `console` or `help`.
- A Window takes its name from its title: up to 40 characters, with accents folded, and `-2`, `-3` added when the name is taken. "Q3 review: Été" becomes `q3-review-ete`.

A Window's name holds no random part, so its address can be guessed. Access levels are meant to protect Windows, not hard-to-guess addresses: see [Who can open a Window](/concepts/access/).

## Rooms

Each person has a personal Room, created with them, which they own and share with nobody. Their AI publishes there by default, and can publish in any Room the person is in.

A team Room gathers its owners and members. Any person, or their AI, creates one with `create_room`, and becomes its first owner.

| Who | What they do in a team Room |
| --- | --- |
| Its members | Publish in it, open its Windows, and read in the Logbook what happens in it |
| Its owners | The same, and change its title and what its address shows, add and remove people, and make them owners or members |
| The administrators of the Cahut | Change its title and its people, but open its Windows only if they are in it |

A Room keeps at least one owner, and anyone can leave a team Room. Adding or removing a person, and changing their role, [knock first](/concepts/knock-first-and-logbook/#knock-first): when an AI asks, its person approves in the console. A Room's name, and so its address, does not change, but when a person [changes their handle](/admin/console/#name-and-handle): their personal Room, and in an individual Cahut each of their Rooms, takes the new one, and the old addresses stop working.

In an [individual Cahut](/concepts/cahuts/#collaborative-or-individual), each person still creates Rooms of their own, but stays their only member, and the name of each starts with their handle, such as `bob-projects`.

Its owners, and the administrators of the Cahut, delete a team Room for good with `delete_room`, or on its page in the console once its name is typed: its people, its Windows, those in its trash included, and their files go with it, and its name becomes free. The Logbook keeps what happened in it. A personal Room is never deleted. Deleting a Room knocks first when an AI asks.

## The page of a Room

The address of a Room, such as `https://cahut.example.com/marketing/`, shows each reader the Windows they may open, as cards, most recently updated first. Each card shows the Window itself, shrunk to the card, with its kind, when it last changed, and its access. A reader outside the Room sees only the Windows open to them, and a reader who may open none finds nothing, as at an address that does not exist. The page names nobody: neither who published nor who is in the Room.

The owners of a Room, personal Rooms included, can show one of its Windows at its address instead, such as a page that presents the team or the person: `set_room_home`, or **Page of the Room** under **Manage the Room**, on its page in the console. Readers who may open that Window see it at the address of the Room; the others still see the Windows they may open. The Window keeps its own address and its access, and `set_room_home` without a Window goes back to the list. The decision is [ADR 32](https://github.com/getcahut/cahut/blob/main/docs/adr/0032-the-page-of-a-room.md).

## Windows

A Window has a kind: `document`, `deck`, `dashboard` or `site`. The kind says what the Window is; Cahut serves every kind the same way. Apps that run on the server, in containers, come later.

A Window is static: HTML, styles, scripts, images, fonts and data files, served as they are, each with the media type of its extension. Readers' browsers check for a newer copy on every visit. A folder opens on its `index.html`; Cahut lists no folder, and a folder without `index.html` answers `404 Not Found`.

## Versions and the trash

A Window keeps two versions: the current one, which readers see, and the one before.

- `update_window` publishes a new version at the same address, with all its files. The current version becomes the previous one, and the version before that is deleted.
- `rollback_window` goes back to the previous version. The version it leaves becomes the previous one, so a second rollback undoes the first. Versions keep their numbers: after a rollback from 3 to 2, the next update is 4.
- `trash_window` puts a Window in the trash: it no longer opens, it leaves the lists, and its name stays taken. `restore_window` brings it back as it was, with its versions and its access.

The members of the Room, and their AIs, do all three, and the console offers them on the page of each Window.

`purge_window` deletes a Window in the trash for good, with its versions and its files, and frees its name. It cannot be undone: a Window goes to the trash first, and an AI that purges one [knocks first](/concepts/knock-first-and-logbook/#knock-first). In the console, the page of a Window in the trash deletes it for good, after a confirmation.

## The sandbox

The pages of every Window come from the domain of their Cahut, and a browser would let them read each other. So every file of a Window is served with these headers:

```http
Content-Security-Policy: sandbox allow-scripts allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads; default-src 'none'; ...
Access-Control-Allow-Origin: *
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Cache-Control: no-cache
```

The files of a protected Window, in their frame, carry `Referrer-Policy: no-referrer` instead. The `sandbox` directive, without `allow-same-origin`, gives the page an opaque origin:

- Its scripts cannot read the other Windows of the Cahut, nor the session of its reader.
- `localStorage`, `sessionStorage`, IndexedDB and `document.cookie` throw a `SecurityError`. A page keeps its state in memory or in its address.
- Scripts, forms, dialogs, downloads, and links that open in a new tab keep working.
- Module scripts, `fetch` calls and web fonts become cross-origin requests. `Access-Control-Allow-Origin: *` lets them load the files of the Window, without credentials.

## What a Window declares

The AI that publishes a version declares what its pages do: whether they run scripts, every origin they load anything from or call, and the origins their forms send to. Cahut writes the rest of the policy from it, so that a Window that declares no call can make none:

```text
sandbox allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads [allow-scripts];
default-src 'none';
script-src <cahut> 'unsafe-inline' 'unsafe-eval' <origins>     (only with scripts)
style-src <cahut> 'unsafe-inline' <origins>;
img-src <cahut> data: blob: <origins>;
font-src <cahut> data: <origins>;
media-src <cahut> data: blob: <origins>;
connect-src <cahut> data: blob: <origins>;
frame-src <cahut> <origins>;
worker-src <cahut> blob:;
form-action <forms>, or 'none';
base-uri 'none'
```

`<cahut>` is the address of the Cahut, such as `https://cahut.example.com`, where the files of the Window come from. `data:` and `blob:` URLs hold what the page already has, and make no request: its scripts may read them with `fetch`, as three.js does for the textures of a 3D model. A Window without scripts loses `allow-scripts` from its sandbox as well. The console shows the declaration on the page of each Window, and the Logbook records it with each publication.

The declaration limits what pages load and call, not where their reader goes: a link that the reader follows still leaves. The versions published before declarations keep the sandbox alone, and the console says they are not declared. See [Publish a Window](/connect/publish/#the-declaration).
