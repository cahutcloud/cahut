---
title: What Cahut can and cannot do
description: What a Window holds and does, the sandbox it runs in, its sizes, the limits of a Cahut, and what an AI does alone or once a person approves. To read before you try.
---

Cahut hosts what your AIs make, as Windows: pages that open by link, on your server. This page says what a Window can hold and do, and what it cannot, so that you know before you ask your AI for it. Each part links to the page that says more.

## In short

| You can | You cannot |
| --- | --- |
| Publish documents, decks, dashboards and sites: HTML, styles, scripts, images, fonts and data files | Run code on the server |
| Keep what readers write in a database of the Window: polls, forms, petitions, trackers | Let readers run SQL that writes, or read what the Window does not make readable |
| Run scripts in the page: charts, slides, filters, 3D models | Keep data in the browser: `localStorage`, `sessionStorage`, IndexedDB and cookies throw an error |
| Load libraries and fonts, and call APIs, from the origins that the Window declares | Load or call an origin that it did not declare |
| Send a form to a service that the Window declares, or keep its answers in the data of the Window | Read who a reader of a public Window is |
| Publish up to 500 files and 32 MiB in a Window | Publish or edit a Window in the console: Windows come from AIs |
| Let your AI publish, update, share and put in the trash | Let your AI delete for good, make public or change who does what without your approval |

## A Window is HTML and its files

A Window is HTML and its files: styles, scripts, images, fonts, and data files such as JSON or CSV. Cahut keeps them and serves them as they are, each with the media type of its extension. Its kind, `document`, `deck`, `dashboard` or `site`, says what it is; Cahut serves every kind the same way.

- No code of a Window runs on the server: no PHP, Node.js or Python, no accounts of its own. A page computes what it needs in the browser of its reader. What its readers write, it keeps in the [data of its Window](/concepts/data/), a SQLite database that Cahut runs for it, through the queries that its declaration lists.
- A site has as many pages as it needs, such as `about/index.html`, linked by relative paths.
- The data of a dashboard lives in its files, or comes from an API that its page calls. To refresh it, the AI publishes a new version with `update_window`, at the same address.
- Whoever opens a Window can read every one of its files. Never put an API key, a password or a token in it.
- Nothing edits a file of a Window in place. To change a Window, the AI publishes its new version, with all its files. Cahut keeps two versions, the current one and the one before, and goes back to it in one call.

Apps that run on the server are not part of Cahut today: [Where Cahut stands](/start/status/#what-comes-later) lists what comes later. See [Rooms and Windows](/concepts/rooms-and-windows/#windows) and [Publish a Window](/connect/publish/).

## The sandbox

Every page of a Window runs in a sandbox, apart from the other Windows of the Cahut, from the console and from the session of its reader. In it:

- Scripts, forms, dialogs, downloads, and links that open in a new tab work.
- `localStorage`, `sessionStorage`, IndexedDB and `document.cookie` throw a `SecurityError`. A page keeps its state in memory, which a reload empties, in its address, such as `#slide-3`, or in the [data of its Window](/concepts/data/). A library that stores something must catch the error, or the page stops.
- A page does not know who reads it: Cahut tells it nothing of its reader, and its scripts cannot read the reader's session.
- A Window that is not public opens in a frame of Cahut. The address bar does not follow the pages of a multi-page site.

See [The sandbox](/concepts/rooms-and-windows/#the-sandbox).

## What a Window loads and calls

With each version, the AI declares what its pages do: whether they run scripts, every origin they load anything from or call, and the origins their forms send to. Cahut holds the pages to that declaration.

- An origin starts with `https://` or `wss://`, such as `https://cdn.jsdelivr.net`, and `https://*.example.com` covers the subdomains of `example.com`. Each list holds 20 origins at most.
- The files of the Window need no declaration, nor do the `data:` and `blob:` URLs that its pages make.
- The browser blocks what a page loads or calls from an origin that it did not declare: a library from a CDN that the AI forgot to declare does not load.
- A page in the sandbox has no origin of its own, and its calls carry `Origin: null`: an API answers them only if it accepts calls from any origin.
- A form sends only to the origins of the declaration. To keep its answers in Cahut, the page calls the queries of the data of its Window instead: see [The data of Windows](/concepts/data/).
- The declaration bounds what the pages load and call, not where their reader goes: a link that the reader follows still leaves.

See [What a Window declares](/concepts/rooms-and-windows/#what-a-window-declares) and [The declaration](/connect/publish/#the-declaration).

## Sizes and numbers

| What | At most |
| --- | --- |
| Files in a Window | 500 |
| Size of a Window, its files decoded | 32 MiB |
| Path of a file | 255 bytes |
| Title of a Window or a Room | 200 bytes |
| Description of a Window or a Room | 300 characters |
| Tags of a Window | 10, each of 32 characters |
| Text of the pages of a version that the search reads | 100 KB |
| Words in a search | 10 |
| Name of a Window in its address, taken from its title | 40 characters |
| Origins in each list of a declaration | 20 |
| Versions of a Window | 2: the current one and the one before |
| People in one invitation | 20 |
| Invitations a day, for a person and their AIs | 100 |
| Note of an invitation | 500 characters |

An AI publishes a Window in one call, with all its files: text as it is, and binary files, such as images, in base64, which makes them a third larger in the call. The 32 MiB count once decoded. A Cahut also refuses a publication that would leave less than 1 GiB free on the disk of its server, which all its Cahuts share.

## The limits of a Cahut

A Cahut that you install has no limit until the administrators of its server set some. They can bound:

- its people, its administrators, its team Rooms, its Windows, the Windows of each Room, and its disk;
- in an individual Cahut, the Rooms, the Windows and the disk of each person too.

A call that would cross a limit is refused, with a message that says what to do. The Windows in the trash still count on the disk, but not among the Windows: deleting them for good frees the space. `get_cahut` gives the limits and what the Cahut holds, to people and to AIs. See [Limits](/concepts/cahuts/#limits).

[My Cahut](/start/my-cahut/) is an individual Cahut, without team Rooms, and with limits on the disk and the number of Rooms and Windows.

## What an AI does alone, and what waits for you

An AI acts through the tools of the Cahut, in the Rooms you granted it, and changes nothing if you granted it read access only. It gets no shell and no file of the server. Every call it makes goes to the Logbook.

On its own, an AI does what can be undone:

- publish a Window, update it, go back to its previous version, change its title, its description, its tags and its name in its address, put it in the trash and restore it;
- read the data of a Window, change it after a copy, and go back to a copy;
- open a Window to the whole Cahut or to the holders of its Key, and close it again;
- invite people to read a Window by e-mail, and take an invitation back;
- create a team Room, and change its title, its description and what its address shows;
- narrow what it reaches.

What cannot be undone, and what changes what a person may do, knocks first:

- make a Window public, or give a public Window a query that lets anyone change or delete what its readers wrote;
- remove or rename a table, a column, an index, a view or a trigger of the data of a Window;
- delete for good a Window in the trash, a team Room or a person;
- create a person, disable or enable them, make them an administrator or a member, or change their e-mail address;
- add a person to a Room, remove them, or change their role in it;
- widen what it reaches.

The AI gets the link of a Knock, which waits 24 hours for you in the console, where you approve or decline it. Only you answer the Knocks of your AIs. The operations on people are offered only to the AIs of administrators.

An AI never does these, even with your approval:

- approve a Knock, its own or another's;
- sign in, connect another AI, or create a token;
- manage the server: its Cahuts and their domains, the kind and limits of a Cahut, its e-mail relay, its updates.

See [Knock first and the Logbook](/concepts/knock-first-and-logbook/), and the [operations reference](/reference/operations/), which says which operations are offered to AIs.

## Which AIs work

Cahut works with the AI clients that connect over MCP. Which ones, on which plans, and how to connect each: [Connect an AI client](/connect/ai-clients/). A script, or a client that does not sign in with OAuth, uses a [personal token](/connect/tokens/).
