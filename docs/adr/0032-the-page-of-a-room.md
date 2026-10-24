# 32. The page of a Room shows its Windows in previews, or a Window that its owners choose, in the brand of Cahut until the house style

- Status: Accepted, extended by [ADR 33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md) (the pages carry the brand of Cahut while the house style keeps the look of the neutral one, whatever it says to AIs)
- Date: 2026-10-24

## Context

[ADR 9](0009-rooms-and-windows-as-paths.md) makes `https://<domain>/<room>/` the page of a Room, which lists the Windows its reader may open. Batch 4 made it a bare list: a title, a kind and an access on each line. It is the address that people give for a Room, and the first page that the readers of a personal Room see. A person who gives the address of their Room wants it to look like them; a team wants it to present its work.

The console already shows each Window in a live preview, a frame shrunk to its card ([ADR 27](0027-previews-of-windows-in-the-console.md)). The readers of a Room get nothing of the kind.

[ADR 22](0022-house-style-of-each-cahut.md) gives the pages that readers see, such as the page of a Key, the house style of the Cahut: neutral until an administrator sets one, and never the brand of Cahut. A Cahut without a house style thus shows its readers pages that look like neither the organization nor Cahut, in the fonts of the system.

## Decision

### A Window as the home of a Room

- The owners of a Room, personal Rooms included, and the administrators of the Cahut, choose a Window of the Room that the address of the Room shows, in place of the list: `set_room_home` takes the name of a Window of the Room, or none to go back to the list. `get_room` and `list_rooms` give it as `home`; the page of a Room in the console sets it.
- The address of the Room frames the Window in a page of Cahut, as the address of a protected Window does ([ADR 19](0019-window-versions-access-and-declaration.md)): the frame opens the Window at its own address when it is public, and otherwise under a capability that names it alone. The query of the address goes to the frame; the fragment follows it, as on the page of a protected Window.
- The Window shows to the readers who may open it, as at its own address. The others get the list of the Windows they may open, or `404 Not Found` when they are outside the Room and open none. The page of a Room does not ask for the Key of its home: that would tell whoever knows the name of the Room that it holds a Window with a Key.
- The Window keeps its address, its access and its versions. In the trash, it leaves the list at the address of the Room until it is restored; purged, it leaves the Room without a home.
- Choosing the home changes what readers see, not who may open what: no access widens. Choosing again undoes it, so `set_room_home` does not knock first, and undoes itself.

### The list of the Windows

- Each Window that the reader may open is a card, most recently updated first: its preview, its kind with its icon, when it last changed, and its access as a badge with an icon and a word. The count is that of the reader, not of the Room.
- The preview is the Window itself, as in the console: drawn 1280 pixels wide and shrunk to its card by a script of Cahut, `/-/pages.js`, over the icon of its kind. Its frame opens the Window as its own address would for this reader: at that address when it is public, and otherwise under a capability of the reader's session or of their pass for its Key, which names that Window alone. The frame adds the sandbox of a preview, scripts only, loads lazily, and takes neither clicks nor focus.
- The page names nobody: no author and no members. A reader learns the title of the Room, its address, and the Windows open to them.
- The members of a Room without Windows read that it is empty, and how their AI publishes there.

### The brand of Cahut until the house style

- The pages that readers see, the home of the Cahut, the page of a Room, the page of a Key, and the pages of what is not found and of a link that has expired, take the tokens and the components of the brand of Cahut, the `c-` classes of its interface kit.
- While the Cahut has the neutral house style, they carry the brand of Cahut: its logo beside the domain, its fonts and its colors, light or dark as the reader's system is, the pictogram of a Room and the clouds. Once an administrator sets a house style, they carry it, as ADR 22 has it: its variables take the roles of the tokens, its logo or its name is at the top, its single palette applies, and the logo of Cahut appears nowhere. Back to neutral, the pages carry the brand of Cahut again.
- The neutral house style itself does not change: the sheet that Windows link and the templates keep it, so that no Window carries the brand of Cahut.
- The binary embeds the files of the brand that these pages use, from `docs/charte-graphique/` as they are, and serves them under `/-/brand/`, at their paths in the brand guidelines, so that the sheet of the fonts finds them. Its folders list nothing.

## Consequences

- A personal Room can be its person's page: a portfolio or a résumé made by their AI, at a short address. A team Room opens on what presents its work.
- A reader who opens the page of a Room runs the Windows open to them, as the console does for its people: one frame per card, loaded as the cards scroll into view, each with the origins that its Window declared. An anonymous reader loads the public Windows of the Room, as opening each would.
- A Cahut without a house style shows the brand of Cahut to its readers; an organization that sets its house style takes it off every page at once.
- The policy of the pages of Cahut names a second script.
- `/-/brand/` is a new path of Cahut, and the binary grows by the fonts, the sheets and the pictures of the brand, about 140 KB.
- This record extends ADR 9 (what the address of a Room shows), ADR 19 (the frames of protected Windows open at the address of their Room and in its cards), ADR 22 (the pages of Cahut carry the brand of Cahut while the house style is neutral) and ADR 27 (the previews of Windows on the page of a Room).

## Alternatives considered

- **A redirect from the Room to its home Window.** No frame, but the address in the bar becomes that of the Window, and whoever shares the Room shares the Window.
- **The files of the home Window served at the address of the Room.** Its relative links would resolve under the Room, among its other Windows: `css/deck.css` would name a Window called `css`.
- **A description and a cover for each Room**, set in the console. They do less than a Window, which an AI already makes in the house style.
- **The brand of Cahut on every page of Cahut**, house style or not. The readers of an organization would see the brand of Cahut beside its own, which ADR 22 rules out.
- **The brand of Cahut as the neutral house style**, for Windows too. Every Window made before a house style would carry it, and keep it when it copies the sheet into its files.
- **Copies of the files of the brand in the web server.** Two copies drift apart, and the brand guidelines forbid copying a token, a font or a logo.
