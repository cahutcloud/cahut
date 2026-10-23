---
name: cahut
description: Make documents, decks, dashboards and small sites, and publish them on Cahut, the user's own server, in the house style of their organization; then share them by link, update them or take them back. Use it whenever the user asks to publish, put online, share or update a page, a report, a presentation, a slide deck or a dashboard on Cahut, or to make one in their house style, even when they do not name Cahut but its tools are connected.
---

# Cahut

Cahut keeps what you make on the user's own server. Each piece is a Window: a living page at `https://<domain>/<room>/<window>/`, in a Room of the person you act for. You act through the tools of the Cahut's MCP server, never a shell, and every call goes to its Logbook.

## Make a Window

1. Call `get_house_style` with the kind of Window: `document`, `deck`, `dashboard` or `site`. It gives the name of the organization, its colors and fonts, what it asks of AIs (`guidance`), a guide to the classes of its style sheet (`guide`), and the template of that kind.
2. Start from the template. Keep its structure and its classes, replace its sample content, and give the page its own `<title>`. Follow the guidance of the organization, and [the formats](references/formats.md).
3. Keep the links of the template to the house style, from the root of the domain: `<link rel="stylesheet" href="/-/house-style.css">`, and for a deck `<script src="/-/deck.js" defer></script>`. In your own CSS, use its variables, such as `var(--hs-primary)` or `var(--hs-font-heading)`, rather than colors or fonts of your own.
4. Link the Window's own files with relative paths, such as `img/map.webp`, never `/img/map.webp`: a Window lives under the address of its Room. The house style and the deck engine are the only paths under `/-/`.
5. Publish with `publish_window`: a title, the kind, the files with `index.html` at the root, the Room (the personal Room of the person when left out), and the declaration.
6. Give the user the address that it returns.

## Declare what the pages do

Cahut serves each Window in a sandbox that lets it load and call only what its version declares:

- `scripts`: true when a page runs JavaScript, the deck engine included.
- `origins`: every origin that the pages load from or call, such as `https://cdn.jsdelivr.net` for a chart library. The house style needs none.
- `forms`: the origins that forms send to.

Prefer what needs no origin: inline SVG charts, the bars and legends of the house style, its fonts, images among the files.

## The sandbox

Pages run with an opaque origin. `localStorage`, `sessionStorage`, IndexedDB and cookies throw errors: keep state in memory, or in the address after `#`. Service workers do not register. Links, popups and downloads work.

## Files

- Text files in UTF-8. Images, fonts and other binary files in base64, with `"encoding": "base64"`.
- 500 files and 32 MiB at most. Compress images (WebP or JPEG, a few hundred kilobytes each), and size them for the screen.
- To change a Window, `get_window` with its contents returns its files.

## Change, share, undo

- `update_window` publishes a new version at the same address: give all its files again. `rollback_window` goes back to the previous one. `trash_window` and `restore_window` take a Window away and bring it back.
- A Window opens to the members of its Room. `set_window_access` opens it to everyone signed in to the Cahut (`internal`), to anyone with its Key (`key`: Cahut draws one and returns it once, give it to the user), or to anyone (`public`), until an optional date.

## Knock first

Some calls wait for a person: making a Window public, purging a Window from the trash, deleting a Room, adding a person or changing what a person may do, widening the reach of an AI. The tool then returns an error with a link. Give the link to the user, who approves or declines in the console of the Cahut, then call `get_knock` to learn the outcome. Never call again in a loop.

## The house style

The administrators of a Cahut change it with `set_house_style`, giving only what changes: the name of the organization, its colors, its fonts and their files, its logos, its guidance. `rollback_house_style` goes back to the previous one. A change reaches every Window that links the sheet at once, so ask before you change the brand of an organization, and show the user the result: `/-/templates/deck` shows a deck in the house style.
