---
name: cahut
description: Make documents, decks, dashboards and small sites, and publish them on Cahut, the user's own server; then share them by link, update them or take them back. Use it whenever the user asks to publish, put online, share or update a page, a report, a presentation, a slide deck or a dashboard on Cahut, even when they do not name Cahut but its tools are connected.
---

# Cahut

Cahut keeps what you make on the user's own server. Each piece is a Window: a living page at `https://<domain>/<room>/<window>/`, in a Room of the person you act for. You act through the tools of the Cahut's MCP server, never a shell, and every call goes to its Logbook.

## Make a Window

1. Make the Window in the style that the user asks for, or that the skills of their organization give, and follow [the formats](references/formats.md). Cahut serves no style sheet, template or script for Windows: the Window carries its own.
2. Write a complete `index.html` with its own `<title>`. Put its CSS, its fonts, its scripts and its images among its files, such as `css/style.css`, or inline in the page.
3. Link the Window's own files with relative paths, such as `img/map.webp`, never `/img/map.webp`: a Window lives under the address of its Room.
4. Publish with `publish_window`: a title, the kind (`document`, `deck`, `dashboard` or `site`), the files with `index.html` at the root, the Room (the personal Room of the person when left out), and the declaration.
5. Give the user the address that it returns.

## Declare what the pages do

Cahut serves each Window in a sandbox that lets it load and call only what its version declares:

- `scripts`: true when a page runs JavaScript, such as the script that moves between the slides of a deck.
- `origins`: every origin that the pages load from or call, such as `https://cdn.jsdelivr.net` for a chart library. The Window's own files need none.
- `forms`: the origins that forms send to.

Prefer what needs no origin: inline SVG charts, CSS bars, the fonts of the system or fonts among the files, images among the files.

## The sandbox

Pages run with an opaque origin. `localStorage`, `sessionStorage`, IndexedDB and cookies throw errors: keep state in memory, or in the address after `#`. Service workers do not register. Links, popups and downloads work.

## Files

- Text files in UTF-8. Images, fonts and other binary files in base64, with `"encoding": "base64"`.
- 500 files and 32 MiB at most. Compress images (WebP or JPEG, a few hundred kilobytes each), and size them for the screen.
- To change a Window, `get_window` with its contents returns its files.

## Change, share, undo

- Each Window has an `id` that never changes. Name the Window by its id in later calls, as `window`, to find it whatever its name. Its `permanent_url`, `https://<domain>/-/w/<id>`, sends whoever may open it to its address: give it for a link that lasts.
- `update_window` publishes a new version at the same address: give all its files again. `rollback_window` goes back to the previous one. `trash_window` and `restore_window` take a Window away and bring it back.
- A Window opens to the members of its Room. `set_window_access` opens it to everyone signed in to the Cahut (`internal`), to anyone with its Key (`key`: Cahut draws one and returns it once, give it to the user), or to anyone (`public`), until an optional date.
- The address of a Room, `https://<domain>/<room>/`, shows each reader the Windows they may open. `set_room_home` shows one of its Windows there instead, such as a page that presents the Room, to the readers who may open it; without a Window, it goes back to the list.

## Knock first

Some calls wait for a person: making a Window public, purging a Window from the trash, deleting a Room, adding a person or changing what a person may do, widening the reach of an AI. The tool then returns an error with a link. Give the link to the user, who approves or declines in the console of the Cahut, then call `get_knock` to learn the outcome. Never call again in a loop.
