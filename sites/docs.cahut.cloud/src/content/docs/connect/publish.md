---
title: Publish a Window
description: What an AI sends to publish a Window, what comes back, and the rules that its files follow.
---

An AI publishes a Window with the `publish_window` tool. It sends the title, the kind and the files of the Window, and gets back the address where the Window opens. This page is for people who write prompts, skills or MCP clients. The [MCP tools reference](/reference/mcp-tools/#publish_window) gives the full schema.

## A publication

The arguments of `publish_window`:

```json
{
  "title": "Q3 review",
  "kind": "deck",
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
  "updated_at": "2026-10-21T12:02:31.418Z"
}
```

The AI gives the `url` to the user.

## Files

- A path is relative, with `/` as separator, such as `img/logo.svg`, and at most 255 bytes long. No part of it starts with a dot.
- Two paths cannot differ by case alone, and a path cannot be both a file and the folder of another file.
- Text goes as it is, in UTF-8. Binary files, such as images and fonts, go in base64, with `"encoding": "base64"`.
- A Window holds at most 500 files and 32 MiB, once decoded.
- Pages link their files with relative paths, such as `css/deck.css`. A path that starts with `/` points to the root of the Cahut, not to the Window.
- A folder opens on its own `index.html`, so a site can have pages such as `about/index.html`. Cahut lists no folder.

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

`get_window` describes a Window, given its address or `room/name`, and lists its files, with their contents when asked. `list_windows` lists the Windows of the person's Rooms, or of one of them with `room`, most recently updated first.

A Window cannot be updated or deleted yet: publishing again creates another Window, such as `q3-review-2`. Versions, with a rollback to the previous one, come later.

## When a call fails

When the arguments are wrong, the tool returns an error that says what to change:

```text
invalid input: index.html is missing: it is the page that opens
```

The AI can fix its call and try again. Failed calls go to the Logbook too.
