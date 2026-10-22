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
| Internal | Everyone signed in to its Cahut. |
| Key | Anyone who types its Key, and the members of its Room. |
| Public | Anyone who has its address. |

An access other than Room only can end on a date. After that date, the Window opens to its Room only, until someone sets its access again; setting it again moves the date or removes it.

A Window that its reader may not open answers like a Window that does not exist, `404 Not Found`, with a link to sign in. The page of a Room, `https://<domain>/<room>/`, lists the Windows that its reader may open, and answers `404 Not Found` to a reader outside the Room who may open none.

The administrators of a Cahut manage people and Rooms, not content: they open the Windows of their own Rooms, and those open to them like to anyone else.

## Keys

A Key is a password that opens one Window. When the access is Key, Cahut draws one, such as `k7mp-x3qa-9rtw`, unless the person or the AI chooses one of 10 characters at least. Cahut shows it once, next to the address, and keeps only a hash of it: nobody can read it again. The Logbook never holds it.

The reader who opens the address of a Key Window gets a page of Cahut that asks for its Key, in the neutral style of the Cahut until it has a house style. A right Key opens the Window, and the browser keeps a pass for it, in a cookie of its own, for 30 days at most and never beyond the end of the access. A new Key, or another access, ends every pass at once.

Tries are limited by address, and wrong Keys by Window: after a burst, the page asks the reader to wait a minute.

## Readers who sign in

To open a Window that is Room only or internal, a reader signs in. The link on the page of a closed Window, or on the home of the Cahut, `https://<domain>/`, goes through the [console](/admin/console/): the reader signs in there if they need to, and comes back to the Window at once. The browser then keeps a reader session on the address of the Windows, which lasts like a session of the console: 7 days without use, 30 days in any case. **Sign out**, on the home of the Cahut, ends it.

A reader session is no session of the console: it opens Windows, and calls no operation. Disabling a person, or a change of their password, ends their reader sessions too. The administrators of the server are not people of a Cahut, and do not sign in as readers.

## Who changes the access

| Who | What they set |
| --- | --- |
| The members of the Room | Every access, its end and its Key, in the console |
| Their AIs | Room only, internal and Key, with `set_window_access`, within the Rooms they were granted |

Opening a Window to its Cahut or to the holders of a Key keeps it among people the organization chose, and can be undone at once: an AI does it alone. Making a Window public cannot be taken back once it is read: an AI that asks gets an error that sends its person to the console, until [Knock first](/concepts/knock-first-and-logbook/#knock-first) lets the AI ask and a person approve. In the console, making a Window public asks to confirm first.

## How a protected Window is served

Whatever its access, every Window runs in a [sandbox](/concepts/rooms-and-windows/#the-sandbox), apart from the other Windows and from the reader's session.

- A public Window is served at its address.
- Any other Window, once its reader may open it, answers at its address with a page of Cahut that shows it in a sandboxed frame, which allows fullscreen. The frame loads the Window from `https://<domain>/-/w/<capability>/`: an address signed by the Cahut, bound to that Window and to the reader's session or pass, valid four hours. Each file it serves checks that the reader may still open the Window, so a change of access, a new Key or a Room left takes effect on the next file.
- The fragment of the address, such as `#slide-3`, passes on to the frame. The address bar does not follow the pages of a multi-page site.
