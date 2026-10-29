---
title: Who can open a Window
description: Each Window opens to its Room, to everyone signed in to its Cahut, to anyone with its Key, or to the public, until a date if you want. How readers sign in, how Keys work, and how a protected Window is served.
sidebar:
  label: Window access
---

The access of a Window says who can open it. A new Window opens to its Room only. The members of the Room change its access in the console, on the page of the Window, and their AIs with `set_window_access`.

## Four levels

| Access | Who opens the Window |
| --- | --- |
| Room only | The members of its Room, signed in. The default. |
| Internal | Everyone signed in to its Cahut, except in an [individual Cahut](/concepts/cahuts/#collaborative-or-individual). |
| Key | Anyone who types its Key, and the members of its Room. |
| Public | Anyone who has its address. |

An access other than Room only can end on a date. After that date, the Window opens to its Room only, until someone sets its access again; setting it again moves the date or removes it.

Besides its access, a Window may open to whoever holds one of its [share links](/concepts/sharing/#share-links), without an account or a Key, until the link ends. Changing the access does not end them.

A Window that its reader may not open answers like a Window that does not exist, `404 Not Found`, with a link to sign in. The page of a Room, `https://<domain>/<room>/`, lists the Windows that its reader may open, or shows the Window that its owners chose to the readers who may open it, and answers `404 Not Found` to a reader outside the Room who may open none. See [The page of a Room](/concepts/rooms-and-windows/#the-page-of-a-room).

The administrators of a Cahut manage people and Rooms, not content: they open the Windows of their own Rooms, and those open to them like to anyone else.

## Keys

A Key is a password that opens one Window. When the access is Key, Cahut draws one, such as `k7mp-x3qa-9rtw`, unless the person or the AI chooses one of 10 characters at least. Cahut shows it once, next to the address, and keeps only a hash of it: nobody can read it again. The Logbook never holds it.

The reader who opens the address of a Key Window gets a page that asks for its Key, in the brand of Cahut. A right Key opens the Window, and the browser keeps a pass for it, in a cookie of its own, for 30 days at most and never beyond the end of the access. A new Key, or another access, ends every pass at once.

Tries are limited by address, and wrong Keys by Window: after a burst, the page asks the reader to wait a minute. When many wrong Keys reach a Window, from anywhere, it still opens to a reader who comes with the right Key, and each address that guessed wrong waits a minute between tries.

## Invitations

The members of the Room of a Window, and their AIs, invite people to read it by e-mail, with a note, from the page of the Window in the console or with [`invite_to_window`](/reference/mcp-tools/#invite_to_window). A note holds 500 characters at most. It can set words in `**bold**` or `*italics*`, and list things on lines that start with `-` or `1.`: the editor of the console writes these marks. The Cahut must send e-mail: see [E-mail](/admin/mail/).

| Who | What they get |
| --- | --- |
| A person of the Cahut, named by handle, who opens the Window already | Its address, where they sign in |
| A person of the Cahut who does not open it, for a Window with a Key | A personal link |
| Anyone else, named by address, for a Window with a Key | A personal link |
| Anyone, for a public Window | Its address |

A personal link stands for the Key, for its recipient alone: the Key itself never goes by e-mail. It opens a page of the console that says who invites, to which Window, with what note, and opens the Window on a click, with a pass as the right Key gives. It works again on another device, for 30 days and never beyond the end of the access. **Take back**, in the list of invitations of the Window, ends it and its passes: the Window closes to that recipient alone. A new Key, or another access, ends every personal link of the old Key.

A Window open to its Room or to the Cahut goes to people of the Cahut only, by handle; to invite someone else by address, set its access to Key first. In an [individual Cahut](/concepts/cahuts/), invitations go by address only. An invitation goes to 20 people at most, and a person sends 100 a day at most, their AIs included. An AI invites without knocking, since it could pass the Key on anyway; the message says that the AI sent it, and the Logbook names every recipient.

## Readers who sign in

To open a Window that is Room only or internal, a reader signs in. The link on the page of a closed Window goes through the [console](/admin/console/): the reader signs in there if they need to, and comes back at once. The browser then keeps a reader session on the address of the Windows, which lasts like a session of the console: 7 days without use, 30 days in any case. **Sign out**, at the top of the pages of the Cahut, ends it with the session of the console that opened it in that browser, which would otherwise sign the reader in again at once. Signing out of the console ends it too.

A reader session is no session of the console: it opens Windows, and calls no operation. Disabling a person, or a change of their password, ends their reader sessions too. The administrators of the server are not people of a Cahut, and do not sign in as readers.

## Who changes the access

| Who | What they set |
| --- | --- |
| The members of the Room | Every access, its end and its Key, in the console |
| Their AIs | Room only, internal and Key, with `set_window_access`, within the Rooms they were granted; public, once their person approves |

Opening a Window to its Cahut or to the holders of a Key keeps it among people the organization chose, and can be undone at once: an AI does it alone. Making a Window public cannot be taken back once it is read: an AI that asks [knocks first](/concepts/knock-first-and-logbook/#knock-first), and the Window opens to the public once its person approves in the console. In the console, making a Window public asks to confirm first.

## How a protected Window is served

Whatever its access, every Window runs in a [sandbox](/concepts/rooms-and-windows/#the-sandbox), apart from the other Windows and from the reader's session.

- A public Window is served at its address.
- Any other Window, once its reader may open it, answers at its address with a page of Cahut that shows it in a sandboxed frame, which allows fullscreen. The frame loads the Window from `https://<domain>/-/frame/<capability>/`: an address signed by the Cahut, bound to that Window and to the reader's session or pass, valid four hours. Each file it serves checks that the reader may still open the Window, so a change of access, a new Key or a Room left takes effect on the next file.
- The fragment of the address, such as `#slide-3`, passes on to the frame. The address bar does not follow the pages of a multi-page site.
- In the console, the card of each Window shows a preview of it, in the same kind of frame, bound to the session of the console rather than to a reader. The preview runs the scripts of the Window, but takes no click, sends no form and opens no pop-up, and it closes when you sign out of the console.
