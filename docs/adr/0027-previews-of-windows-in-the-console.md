# 27. The console shows each Window in a live preview

- Status: Accepted
- Date: 2026-10-23

## Context

The console lists Windows as cards: on the overview, the latest Windows of a person's Rooms; on the page of a Room, all of them. Each card shows the icon of its kind, its title, its age and its access. Batch 2 left the thumbnails for later. Four decks in a row look the same, and a person finds a Window by its look before its title.

Cahut has no picture of a Window. It holds files that a browser draws, often with scripts: a deck draws its slides, a dashboard its charts. The server has no browser to draw them ([ADR 4](0004-sqlite-in-pure-go.md) and the V0 plan keep the binary static, and a headless browser on the server is out of the question). The AIs that publish Windows rarely can make a picture of what they made.

A browser, though, is right there: the one that shows the console. It can draw each Window itself, in a frame shrunk to its card. The question is how that frame opens a Window that is not public, safely:

- The page of a protected Window, at its address, refuses to be framed (`frame-ancestors 'none'`, [ADR 19](0019-window-versions-access-and-declaration.md)), and it needs the cookie of a reader session, which a person signed in to the console may not have.
- The console reads no cookie and holds its session in the storage of the origin ([ADR 20](0020-console-under-a-path-of-port-443.md)). A frame cannot send that session in a header.
- The address of a frame is visible to the Window it shows: its scripts read `location`. Whatever opens one Window must open no other, or a Window could read the others of its reader, as [ADR 10](0010-windows-in-a-csp-sandbox.md) forbids.

## Decision

- **Each card shows the Window itself.** The console frames the Window, drawn 1280 pixels wide, and shrinks it to its card, over the icon of its kind, which stays while the frame loads or when there is none. Frames load lazily, as the cards scroll into view.
- **The frame is a capability of the console's session.** The frames of protected Windows under `/-/w/<capability>/` ([ADR 19](0019-window-versions-access-and-declaration.md)) take a third kind of capability, next to the reader session and the pass: the session of the console. It names one Window, that session, and an expiry four hours later, signed with the same key of the Cahut. Each file served under it checks the signature, the expiry, that the session still lasts, and that its person still opens the Window. A capability names one Window, so a Window that reads its own address opens no other.
- **An operation gives the frames.** `preview_windows` takes the Windows of a list, at most 100, and returns a frame for each that is in the person's Rooms and out of the trash, so that each list costs one call. It is for people only, as it serves a session of the console; AIs do not see it. It changes nothing, so a person's calls do not go to the Logbook, as with the other reads.
- **A preview does not keep the session alive.** Serving a file under the capability of a console session does not record a use of the session: a dashboard that keeps loading its data must not keep an idle console open.
- **The frame does less than the Window.** The files keep the CSP sandbox of their version, and the frame of the card adds its own sandbox, which allows scripts only: no forms, pop-ups, dialogs or downloads. The frame takes neither clicks, which go to the card, nor focus, and is hidden from assistive technologies, which read the card. The console's policy allows frames of its own origin only, so a Window that navigates its frame elsewhere is stopped there.

## Consequences

- People recognize their Windows at a glance, as their readers see them, house style included, with no picture to make, store or keep up to date: a new version shows at once.
- Each card runs its Window: a list of thirty dashboards loads thirty pages and their scripts, and the origins they declared. Lazy loading spreads the cost; a slow Window only delays its own card. A Window whose script loops makes the console slow, as it would make its own page.
- A Window shows in its card what it shows at 1280 pixels, its first screen: the first slide of a deck, the top of a document.
- The console now calls the Windows it lists. The capability of a preview is bound to the session of the console, so signing out of the console closes the previews on their next file, like a removed member or a Window in the trash.
- The pages of Cahut keep `frame-ancestors 'none'`; nothing else frames them.

## Alternatives considered

- **Pictures made on the server.** A headless browser on every server, with its updates and its memory, for thumbnails: out of proportion, and against a single static binary.
- **A picture given by the AI** at publication, such as `preview.png`. Few AIs can draw their own pages, the picture would age with each version, and Windows without one would keep the icon.
- **A reader session for the console**, with the page of each protected Window allowed in a frame of the console. It needs a cookie that the console does not read, a reader session opened behind the person's back, and the pages of Cahut framed, with a frame in a frame for each card.
- **One capability for the whole session**, from which the console makes the address of each Window. One call less, but a Window that reads its own address would open every other Window of its reader.
- **Previews in the outputs of `list_windows` and `get_window`.** No extra call, but a field in the tools of every AI that only a browser signed in to the console can use.
