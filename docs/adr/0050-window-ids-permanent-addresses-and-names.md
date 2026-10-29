# 50. Each Window has an id that never changes and a permanent address, and its people change its name, its title and its description

- Status: Accepted, extended by [ADR 52](0052-share-links.md) (a share link opens a Window at a second address under `/-/`)
- Date: 2026-10-29

## Context

A Window is at `https://<domain>/<room>/<window>/` ([ADR 9](0009-rooms-and-windows-as-paths.md)). Its name, the last part of that address, comes from its title when it is published, and nothing changes it. ADR 9 left for after the V0 the renaming of a Window and its move to another Room, "with a redirect from the old address".

The address of a Window changes all the same. When a person changes their handle, their personal Room takes the new one, the addresses of its Windows change with it, and the old ones stop working, without a redirect; the old addresses may even lead one day to the Windows of another person who takes the handle ([ADR 39](0039-people-change-their-name-and-handle.md)).

The work that comes next needs a way to name a Window that does not change: the preview of a shared link names its address in `og:url`; a link of sharing, secret and provisional, opens one Window; the Village, a network of the Windows that their authors list, keeps an index of them across Cahuts; a Window may one day move to another Room, or to another server. An AI that updates a Window finds it today by its address, which a rename would break.

Today a Window has an id, an integer of the database of its Cahut, which is never reused ([ADR 21](0021-ais-knock-first.md)). It names the folder of its files, its passes and its frames, but it is small, it counts the Windows of the Cahut, and another Cahut gives the same integers.

People also ask to change the title of a Window without publishing a version, and to give a Window and a Room a few words that say what they are.

The frames of protected Windows and of previews are at `/-/w/<capability>/` ([ADR 19](0019-window-versions-access-and-declaration.md), [ADR 27](0027-previews-of-windows-in-the-console.md)).

## Decision

### The id of a Window

- Each Window gets an id when it is published: 128 bits drawn at random by the operating system, enough to be unique in the world without the Cahuts asking one another. It is written in 26 characters of base32 in lowercase, in the alphabet of Crockford, which has no `i`, `l`, `o` or `u`: `0123456789abcdefghjkmnpqrstvwxyz`. For instance `3f9k2m7q8x4r1t6v0b5n8c2d4g`.
- The id never changes: not with a new title, a new name, a new version, the trash, a change of access, or a new handle of its author. It is never reused: a purged Window takes its id with it, and an id drawn at random does not come back. A unique index refuses a second Window with the same id in a Cahut.
- The migration gives an id to each Window that already exists.
- The id is not a secret and opens nothing. The members of the Room see it, and so do their AIs and the Logbook; the previews and the Village will show it. If it opened its Window, it would be a Key that nothing could revoke.
- The integer id stays inside the Cahut: the folder of the files, the cookies of passes and the capabilities of frames keep it.
- `publish_window`, `get_window`, `list_windows` and the other operations that describe a Window give `id`, and its permanent address as `permanent_url`.

### The permanent address

- Each Window has a permanent address: `https://<domain>/-/w/<id>`. It is under `/-/`, which no Room can take ([ADR 9](0009-rooms-and-windows-as-paths.md)).
- The frames move to `/-/frame/<capability>/`, so that the address that people share is the short one. A frame is never shared: the page of a protected Window, the page of a Room and the console make a new one each time.
- The permanent address sends its reader to the address of the moment, `/<room>/<window>/`, with what follows the id and the query: `/-/w/<id>/img/logo.svg?q=1` goes to `/<room>/<window>/img/logo.svg?q=1`, and browsers keep the fragment. The redirect is temporary, `307 Temporary Redirect`, so that browsers do not keep it when the name changes.
- It checks the access of the Window first, as the address of the Window does, and redirects only a reader who may open it there: anyone for a public Window, a reader whose session opens it, or a reader with a pass for its Key.
- A reader who may not open it gets what the address of the Window would give them, without learning that address: the page of what is not found, with a link to sign in that brings them back to the permanent address; for a Window with a Key, the page of its Key, whose form names the Window by its id. The right Key gives the pass and sends the reader to the address of the Window.
- A Window in the trash, or purged, is not found at its permanent address.
- The permanent address serves no file. The files of a Window stay at its address, in the sandbox and the frames of ADR 10 and ADR 19, which do not change.

### Naming a Window by its id

- Every operation that takes a Window, in `window`, takes its id or its permanent address as well as its address or its Room and its name as `room/name`.
- The MCP instructions ask AIs to keep the id of the Windows they publish, to find them again whatever their name.
- The Logbook keeps the address of each entry as it was. `publish_window` and `rename_window` write the id in their detail, so that an entry leads to its Window after a change of name.

### The name of a Window

- `rename_window` changes the name of a Window, the last part of its address, and nothing else: its id, its title, its versions, its access and its Key, its passes, its invitations and its place as the home of its Room stay.
- The new name has the form of ADR 9, and is free in the Room, Windows in the trash included. A Window in the trash is not renamed.
- `publish_window` takes a name, when the AI chooses one. A name that is taken is refused, with what to do; without a name, the Window takes its name from its title, as before.
- The old address stops working at once, and its name is free in the Room, as the old handle is in ADR 39: there is no redirect. The permanent address is the link that lasts.
- An AI renames a Window alone: `rename_window` with the old name undoes it, as long as no other Window has taken it.
- The console changes the name on the page of the Window. It says, before the change, that the old address will stop working and that the permanent address will not, and asks again on the spot.
- The invitations that the Cahut sends from now on link the permanent address, so that they still open the Window after a change of name. The personal links of invitations already follow the Window ([ADR 31](0031-e-mail-through-a-relay.md)).

### The title and the description

- A Window and a Room have a description: a sentence or two in plain text, on one line, 300 characters at most, empty by default.
- `publish_window` takes the description of a new Window. `retitle_window` changes the title of a Window, its description or both, without a new version, and its address does not change; `update_window` still takes a title, and a description too. `update_room` changes the title and the description of a Room.
- `retitle_window` with the old texts undoes it: an AI does it alone. The Logbook records each change.
- The console changes them on the page of the Window and on the page of the Room. The description shows there, in the cards of the page of a Room for its readers, and in `list_windows` and `list_rooms`, so that an AI finds the Window it is asked to update.

### Later

- The permanent address of a Room, the renaming of a team Room, and the move of a Window to another Room, which changes who opens it, wait for a record of their own.
- Sharing builds on this record: the preview of a link names the permanent address in `og:url`, and a link of sharing opens a Window through it. Their records settle the rest.

## Consequences

- A link that must last is the permanent address; the address of the moment stays the one that people read and type.
- The permanent address survives a new name of the Window and a new handle of its author. A new domain for the Cahut still breaks it, as it breaks every address.
- A permanent address costs a lookup and the check of access before its redirect, and its reader then makes a second request.
- A new name breaks the links to the old address that were given out, public ones included. The person decides it, warned by the console; an AI that renames a Window says it to its person.
- A Knock that waits for its person and names the Window by its old address finds no Window when it is approved, or the Window that took its name since: whoever approves it reads the address on the Knock, as in ADR 39. An AI that names Windows by their id makes Knocks that a rename does not touch.
- The e-mails of invitations sent before this record give the old address, which a change of name breaks; their personal links still open the Window.
- The frames that are open when a server takes this record stop loading their files, until their page is opened again.
- This record extends ADR 9 (a permanent address under `/-/`, and a name that changes without a redirect, where ADR 9 planned one), revisits ADR 39 (the Windows of a person who changes their handle keep their permanent addresses), and supersedes in part ADR 19 and ADR 27 (the frames are at `/-/frame/<capability>/`).

## Alternatives considered

- **The Window served at its permanent address**, without a redirect. Each Window would have two addresses for every file, its relative links would resolve under `/-/w/<id>/`, and its reader would see a code in the address bar rather than the names of the Room and the Window.
- **A redirect before any check of access.** One step less, but whoever holds the link learns the address of the moment, the name of the Room and a name drawn from the title, of a Window they may not open.
- **A UUID**, in 36 characters with hyphens: the same bits, ten more characters. **A ULID**, which writes the time of creation in its first characters: it would tell when the Window was made, which nothing needs. **The integer id**: it counts the Windows of the Cahut, and every Cahut has the same ones.
- **The permanent address at `/-/window/<id>`**, with the frames where they are. Nothing moves, but the address that people share is five characters longer.
- **A redirect from the old name to the new one**, the old name staying taken by the Window. Shared links would keep working, but each name ever given would stay taken, and a table of old names would grow with each change. ADR 39 chose a clean break for handles, and the permanent address now keeps the links that must last.
- **A title changed by `update_window` without files.** One operation fewer, but `rollback_window` undoes `update_window`, and a rollback does not bring back a title.
