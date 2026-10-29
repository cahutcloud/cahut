---
title: Publish a Window
description: What an AI sends to publish a Window, what it declares, what comes back, and how it updates, shares and rolls back a Window.
---

An AI publishes a Window with the `publish_window` tool. It sends the title, the kind and the files of the Window, with a declaration of what its pages do, and gets back the address where the Window opens. This page is for people who write prompts, skills or MCP clients. The [MCP tools reference](/reference/mcp-tools/#publish_window) gives the full schema, and [What Cahut can and cannot do](/start/can-and-cannot/) sums up what a Window holds and does.

## The look of a Window

An AI makes each Window in the style the person asks for. Its CSS, fonts, images and scripts are files of the Window, linked by relative paths; a deck carries the script that moves from slide to slide, or loads a library from an origin it declares. Cahut serves no style sheet or script for Windows: a Window published with links to `/-/house-style.css` or `/-/deck.js` loses its look until it is updated with CSS of its own.

## A publication

The arguments of `publish_window`:

```json
{
  "title": "Q3 review",
  "kind": "deck",
  "declaration": { "scripts": true, "origins": ["https://cdn.jsdelivr.net"] },
  "files": [
    { "path": "index.html", "content": "<!doctype html><html lang=\"en\">...</html>" },
    { "path": "css/deck.css", "content": "body { margin: 0; }" },
    { "path": "img/chart.png", "content": "iVBORw0KGgo...", "encoding": "base64" }
  ]
}
```

| Field | What it holds |
| --- | --- |
| `title` | The title of the Window, at most 200 bytes. The name in its address comes from it. |
| `kind` | `document`, `deck`, `dashboard` or `site`. |
| `room` | Optional: the name of a Room of the person the AI acts for. Without it, the Window goes to their personal Room. |
| `files` | The files, with `index.html` at the root: it is the page that opens. |
| `declaration` | What the pages do: see [The declaration](#the-declaration). |

Cahut answers with the Window:

```json
{
  "name": "q3-review",
  "title": "Q3 review",
  "kind": "deck",
  "url": "https://cahut.example.com/you/q3-review/",
  "room": "you",
  "version": 1,
  "size": 48213,
  "file_count": 3,
  "created_by": "you",
  "created_at": "2026-10-21T12:02:31.418Z",
  "updated_at": "2026-10-21T12:02:31.418Z",
  "access": "room",
  "declaration": { "scripts": true, "origins": ["https://cdn.jsdelivr.net"] }
}
```

The AI gives the `url` to the user. The new Window opens to the members of its Room: see [Share a Window](#share-a-window).

## Files

- A path is relative, with `/` as separator, such as `img/logo.svg`, and at most 255 bytes long. No part of it starts with a dot.
- Two paths cannot differ by case alone, and a path cannot be both a file and the folder of another file.
- Text goes as it is, in UTF-8. Binary files, such as images and fonts, go in base64, with `"encoding": "base64"`.
- A Window holds at most 500 files and 32 MiB, once decoded.
- Pages link their files with relative paths, such as `css/deck.css`. A path that starts with `/` points to the root of the Cahut, not to the Window, and finds none of its files there.
- A folder opens on its own `index.html`, so a site can have pages such as `about/index.html`. Cahut lists no folder.

## The declaration

Every publication declares what its pages do, and Cahut lets them do that and nothing more:

| Field | What it holds |
| --- | --- |
| `scripts` | `true` when the pages run scripts, inline or in files. |
| `origins` | Every origin the pages load anything from or call: scripts, styles, fonts, images, `fetch` calls, WebSockets. Such as `https://cdn.jsdelivr.net`, `https://fonts.googleapis.com` and `https://fonts.gstatic.com`. `https://*.example.com` covers the subdomains of `example.com`, not `example.com` itself. |
| `forms` | The origins that the forms of the pages send to, if any. |

- An origin is `https://` or `wss://`, a host and an optional port, without a path. Each list holds 20 origins at most.
- The files of the Window need no origin: they come from the Cahut, which the policy always allows. A deck that moves from slide to slide with a script declares `scripts`.
- The `data:` and `blob:` URLs that the pages make need no origin either: a 3D model whose textures three.js reads from `blob:` URLs, as it does for a `.glb` file, loads as it is.
- A page that loads from an origin it did not declare does not get it: the browser blocks the request, and the console of the browser says so. Declare every origin, or put the files in the Window.
- A deck that only shows text and images declares `{"scripts": false}`: it can then call nothing at all.

[Rooms and Windows](/concepts/rooms-and-windows/#what-a-window-declares) gives the policy that Cahut writes from it.

## The sandbox

Every file of a Window is served with the header `Content-Security-Policy: sandbox`, which gives its page an origin of its own. In that sandbox:

- Scripts, forms, dialogs, downloads and links that open in a new tab work.
- `localStorage`, `sessionStorage`, IndexedDB and cookies throw an error. Keep state in memory or in the address, and check that the libraries of the page do not stop on that error.
- The page cannot read the other Windows of the Cahut, nor the session of its reader.
- Module scripts, `fetch` calls and web fonts load as cross-origin requests, which the files of the Window allow.

[Rooms and Windows](/concepts/rooms-and-windows/#the-sandbox) gives the full header.

## Names and addresses

Cahut names the Window after its title: up to 40 lowercase ASCII letters, digits and hyphens, with accents folded. "Q3 review: Été" becomes `q3-review-ete`. When the Room already holds a Window of that name, Cahut adds `-2`, then `-3`, and so on. The Window opens at `https://<domain>/<room>/<name>/`.

## After publishing

`get_window` describes a Window, given its address or `room/name`: its access, its two versions, its declaration, and its files, with their contents when asked. `list_windows` lists the Windows of the person's Rooms, or of one of them with `room`, most recently updated first; with `trashed`, it lists the trash.

### Update a Window

`update_window` publishes a new version at the same address. It takes the Window, all the files of the new version and their declaration, and optionally a new title:

```json
{
  "window": "you/q3-review",
  "declaration": { "scripts": false },
  "files": [{ "path": "index.html", "content": "<!doctype html>..." }]
}
```

The version it replaces is kept. `rollback_window`, with the Window alone, goes back to it, and a second rollback returns. Publishing again with `publish_window` creates another Window, such as `q3-review-2`: to change a Window, update it.

### Share a Window

`set_window_access` sets who opens a Window:

```json
{ "window": "you/q3-review", "access": "key", "ends_at": "2026-12-31T18:00:00Z" }
```

| Field | What it holds |
| --- | --- |
| `access` | `room`, `internal`, `key` or `public`. `public` knocks first: the person the AI acts for approves it in the console. |
| `ends_at` | Optional: when the access ends, after which the Window opens to its Room only. |
| `key` | Optional, for `key`: the Key to set, of 10 characters at least. Without it, the Window keeps its Key, or Cahut draws one. |
| `new_key` | Optional, for `key`: draw a new Key; the old one stops working. |

When a Key is set, the answer holds it in `key`, this once. The AI gives the address and the Key to the user. See [Who can open a Window](/concepts/access/).

### Put a Window in the trash

`trash_window` takes a Window out of the lists, and it no longer opens. `restore_window` brings it back as it was.

## When a call fails

When the arguments are wrong, the tool returns an error that says what to change:

```text
invalid input: index.html is missing: it is the page that opens
```

The AI can fix its call and try again. Failed calls go to the Logbook too.
