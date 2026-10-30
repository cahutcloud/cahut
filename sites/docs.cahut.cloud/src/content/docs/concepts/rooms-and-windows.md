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
| The permanent address of a Window | `https://cahut.example.com/-/w/<id>` |
| The MCP endpoint | `https://cahut.example.com/mcp` |
| The health check | `https://cahut.example.com/-/healthz` |

An address without its final slash, such as `https://cahut.example.com/marketing/q3-review`, redirects to the one with it, so that the relative links of the page resolve under the Window.

Paths under `/-/` belong to Cahut itself: the permanent addresses of Windows and their cards, the sign-in of readers, the page of a Key, the frames of protected Windows. Who may open what is in [Who can open a Window](/concepts/access/).

## The id and the permanent address

Each Window gets an id when it is published: 26 characters drawn at random, such as `3f9k2m7q8x4r1t6v0b5n8c2d4g`, unique in the world. The id never changes, whatever happens to the Window's title, name or versions, and it is never given to another Window, even after a purge.

The id gives the Window a permanent address, `https://cahut.example.com/-/w/<id>`. It sends whoever may open the Window to its address of the moment, with the rest of its path and its query: `/-/w/<id>/img/logo.svg` goes to `/<room>/<window>/img/logo.svg`. It survives a new name of the Window, and a new handle of the person whose Room holds it. A new domain for the Cahut breaks it, as it breaks every address.

The id opens nothing by itself, and the permanent address follows the access of the Window. A reader who may not open the Window there finds nothing, as at an address that does not exist, or the page of its Key, without learning its address. In the trash, a Window is not found at its permanent address either.

The console shows both addresses on the page of each Window. AIs get the id as `id` and the permanent address as `permanent_url`, and name a Window by either, as by its address or `room/name`. The invitations that a Cahut sends by e-mail give the permanent address. It is the link to share: robots of previews read there a preview in the brand of Cahut, which says no more than the address says to a stranger. See [Share a Window](/concepts/sharing/). The decision is [ADR 50](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0050-window-ids-permanent-addresses-and-names.md).

## Names

The names of Rooms and Windows take 1 to 63 lowercase ASCII letters, digits and hyphens, and start and end with a letter or a digit. A Room's name is unique in its Cahut, and a Window's name is unique in its Room.

- A personal Room takes the handle of its person, from 2 to 32 characters, starting with a letter.
- No Room can take the names `mcp`, `api`, `admin`, `console` or `help`.
- A Window takes its name from its title: up to 40 characters, with accents folded, and `-2`, `-3` added when the name is taken. "Q3 review: Été" becomes `q3-review-ete`. An AI can choose it instead, with `name` in `publish_window`.

### A new name for a Window

The members of its Room, and their AIs, change the name of a Window with `rename_window`, or under **Addresses** on its page in the console. The new name has the form above and is free in the Room, Windows in the trash included; a Window in the trash keeps its name. The title, the id, the permanent address, the versions, the access, the Key and its passes, the invitations and the place of the Window as the home of its Room stay.

The old address stops working at once, without a redirect, and its name is free for another Window: the permanent address is the link that lasts. Renaming the Window back undoes the change while no other Window has taken the old name, so an AI does it alone. The console says it before the change, and asks again. A Knock that waits and names the Window by its old address finds it no more: an AI that names Windows by their id avoids it.

A Window's name holds no random part, so its address can be guessed. Access levels are meant to protect Windows, not hard-to-guess addresses: see [Who can open a Window](/concepts/access/).

## Rooms

Each person has a personal Room, created with them, which they own and share with nobody. Their AI publishes there by default, and can publish in any Room the person is in.

A team Room gathers its owners and members. Any person, or their AI, creates one with `create_room`, and becomes its first owner.

| Who | What they do in a team Room |
| --- | --- |
| Its members | Publish in it, open its Windows, and read in the Logbook what happens in it |
| Its owners | The same, and change its title, its description and what its address shows, add and remove people, and make them owners or members |
| The administrators of the Cahut | Change its title, its description and its people, but open its Windows only if they are in it |

A Room keeps at least one owner, and anyone can leave a team Room. Adding or removing a person, and changing their role, [knock first](/concepts/knock-first-and-logbook/#knock-first): when an AI asks, its person approves in the console. A Room's name, and so its address, does not change, but when a person [changes their handle](/admin/console/#name-and-handle): their personal Room, and in an individual Cahut each of their Rooms, takes the new one, and the old addresses stop working.

In an [individual Cahut](/concepts/cahuts/#collaborative-or-individual), each person still creates Rooms of their own, but stays their only member, and the name of each starts with their handle, such as `bob-projects`.

Its owners, and the administrators of the Cahut, delete a team Room for good with `delete_room`, or on its page in the console once its name is typed: its people, its Windows, those in its trash included, and their files go with it, and its name becomes free. The Logbook keeps what happened in it. A personal Room is never deleted. Deleting a Room knocks first when an AI asks.

## The page of a Room

The address of a Room, such as `https://cahut.example.com/marketing/`, shows each reader the Windows they may open, as cards, most recently updated first. Each card shows the Window itself, shrunk to the card, with its kind, when it last changed, and its access. A reader outside the Room sees only the Windows open to them, and a reader who may open none finds nothing, as at an address that does not exist. The page names nobody: neither who published nor who is in the Room.

The owners of a Room, personal Rooms included, can show one of its Windows at its address instead, such as a page that presents the team or the person: `set_room_home`, or **Page of the Room** under **Manage the Room**, on its page in the console. Readers who may open that Window see it at the address of the Room; the others still see the Windows they may open. The Window keeps its own address and its access, and `set_room_home` without a Window goes back to the list. The decision is [ADR 32](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0032-the-page-of-a-room.md).

## Windows

A Window has a kind: `document`, `deck`, `dashboard` or `site`. The kind says what the Window is; Cahut serves every kind the same way. Apps that run on the server, in containers, come later.

A Window is static: HTML, styles, scripts, images, fonts and data files, served as they are, each with the media type of its extension. Readers' browsers check for a newer copy on every visit. A folder opens on its `index.html`; Cahut lists no folder, and a folder without `index.html` answers `404 Not Found`. [What Cahut can and cannot do](/start/can-and-cannot/) sums up what a Window holds and does.


### Title and description

A Window has a title, of 200 bytes at most, and a description: a sentence or two of plain text, 300 characters at most, which says what it is. A Room has them too. The page of a Room shows its description under its title, and the description of each Window on its card; AIs read them in `list_windows` and `list_rooms` to find what they are asked to update.

`publish_window` and `create_room` take a description. `retitle_window` changes the title of a Window, its description or both, without a new version, and `update_room` those of a Room; their addresses stay. Changing them is undone by changing them back, so an AI does it alone, and the Logbook records the new texts. In the console, the page of a Window changes them under **Title, description and tags**, and **Manage the Room** changes those of a Room.

### Tags

A Window carries tags, which help its people and their AIs find it among many: 10 at most, each of 1 to 32 letters, digits and hyphens, such as `q3`, `sales` or `board-meeting`. Anyone may write them as they come: Cahut writes `#Board Meeting` as `board-meeting`, in lower case, without the `#`, its words joined by hyphens. Accents and other scripts stay, as in `été`. A tag given twice counts once, and the tags of a Window are kept in order.

- `publish_window` takes `tags`. `retitle_window` changes them without a new version: the list it gives replaces the old one, an empty list removes them, and they stay when it gives none. Giving the old ones back undoes the change, so an AI does it alone, and the Logbook records the new tags.
- The tags belong to the Window, not to a version: `update_window` and `rollback_window` keep them, and so do the trash and a new name.
- `list_windows` gives the tags of each Window, and `tag` lists the Windows with one tag. An AI reuses the tags that the Windows already carry, so that one subject keeps one spelling.
- Only the people of the Room, and their AIs, see the tags: the page of the Room that readers see, and the Windows themselves, show none.

In the console, the page of a Room shows the tags of its Windows over their cards: a click shows only the Windows with that tag. The page of a Window changes them under **Title, description and tags**, where the field suggests the tags of the other Windows of the person.

### Search

People and their AIs search their Windows by their words: `list_windows` with `query` gives the Windows whose title, description, tags or pages hold each word, the best first.

- A word matches at the start of a word, whatever its case and accents: `ete bret` finds a Window titled *Été en Bretagne*. Punctuation does not count, and a search takes 10 words at most.
- A word weighs most in the title, then in the tags, the description and the pages.
- Each Window found comes with an `excerpt`, a few words of its pages around the words found, when its pages hold them.
- The search finds what `list_windows` lists: the Windows of the Rooms of the person, and of the Rooms granted to their AI, but those that others keep for themselves. It goes with `room`, `tag` and `trashed`, and finds a Window in the trash only with `trashed`.
- The pages are the HTML files of the current version, `.html` and `.htm`, read without their markup, scripts and styles, up to 100 KB of text. Cahut reads them when a version is published, and keeps their text with it, so a rollback finds it again. The text that a script writes once a page has loaded is not there: a deck drawn by its scripts is found by its title, its description and its tags.
- Each Cahut keeps its index in its own database, and the Logbook records the words of each search with the number of Windows found.

In the console, the field at the head of every page searches, and opens the Windows found with their excerpts and their tags, which narrow them; on a narrow screen, the field is at the head of the menu. The Windows published before the search came get their text a minute after the service first starts with it.
## Versions and the trash

A Window keeps two versions: the current one, which readers see, and the one before.

- `update_window` publishes a new version at the same address, with all its files. The current version becomes the previous one, and the version before that is deleted.
- `rollback_window` goes back to the previous version. The version it leaves becomes the previous one, so a second rollback undoes the first. Versions keep their numbers: after a rollback from 3 to 2, the next update is 4.
- `trash_window` puts a Window in the trash: it no longer opens, it leaves the lists, and its name stays taken. `restore_window` brings it back as it was, with its versions and its access.

The members of the Room, and their AIs, do all three, and the console offers them on the page of each Window.

`purge_window` deletes a Window in the trash for good, with its versions and its files, and frees its name. It cannot be undone: a Window goes to the trash first, and an AI that purges one [knocks first](/concepts/knock-first-and-logbook/#knock-first). In the console, the page of a Window in the trash deletes it for good, after a confirmation.

## A withdrawn Window

The [moderators](/admin/moderation/) of a Cahut, its administrators and those of its server, withdraw a Window that someone reported, or a public Window, with a reason. A withdrawn Window keeps everything: its versions, its access, its Key, its share links and its invitations. Until they reinstate it:

- Its address, its permanent address, its share links and the page of its Key answer with a page that says it was withdrawn, with the status `410 Gone`, to whoever would open it otherwise; anyone else meets what is not found. The page of its Room no longer lists it, nor shows it at its address.
- Its members find it in the console, marked **Withdrawn**, with the date, the reason, and whether the moderators of the Cahut or the administrators of the server withdrew it. They preview it there, put it in the trash, restore it, which keeps it withdrawn, and delete it for good.
- Nothing changes it, nor opens it to someone new: a new version, a rollback, its access, its Key, its name, its title and description, a share link, an invitation and the home of its Room are refused with the code `window_withdrawn`. Its AIs read its state and its reason in `get_window` and `list_windows`.

Reinstated, it opens as it did, its share links until their end. Only the administrators of the server reinstate a Window that they withdrew.

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
