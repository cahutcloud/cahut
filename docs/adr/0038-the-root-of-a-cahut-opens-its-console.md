# 38. The root of a Cahut opens its console

- Status: Accepted
- Date: 2026-10-25

## Context

[ADR 9](0009-rooms-and-windows-as-paths.md) makes `https://<domain>/` a page of the Cahut. Batch 2 sent it to the console. Batch 4 made it a home for readers: the Rooms of a reader signed in to read Windows, and the address of MCP; whoever had not signed in went to `/console/read`, which signs them in and comes back to the root.

The people of a Cahut type its domain to work there, and their work is in the console: their Rooms, their Windows in previews, their Knocks and the Logbook. The home showed a thin part of the overview of the console, under another session. The pages of the Cahut know a reader by a cookie; the console keeps its session in the storage of the Cahut's origin, which the server never sees ([ADR 20](0020-console-under-a-path-of-port-443.md)). A person signed in to the console who opened the root was asked to sign in again, to read Windows, and then found a list of Rooms rather than the console.

## Decision

- `https://<domain>/` answers `303 See Other` to the console of its Cahut, `https://<domain>/console/`, whoever asks and whatever cookie the request carries.
- The console decides what opens, as at any of its addresses: its overview for whoever has signed in to it in that browser, and its page of sign-in, `/console/sign-in`, for the others, which then opens the overview.
- The redirect is not permanent: browsers do not keep it, and the root may become a page again.
- The page that listed the Rooms of a reader is gone: the console lists them, with their Windows.
- The other pages of the Cahut do not change. A Room, a Key, what is not found and a link that has expired still sign a reader in through `/console/read`, which brings them back where they were. The logo at the top of these pages, and Sign out, lead to the root, and so to the console.

## Consequences

- The domain of a Cahut is the way in for its people: whoever types it lands in the console, signed in, or on its page of sign-in.
- Whoever has not signed in learns nothing more at the root than before: they reached the same page of sign-in through `/console/read`.
- The server still knows nothing of the session of the console, which stays in the browser.
- A reader who signed in to read Windows, and whose session of the console has ended, signs in to the console from the root. On the pages of a Room or a Window, their reader session lasts as before.
- This record supersedes in part ADR 9, whose root was a page of the Cahut, and [ADR 32](0032-the-page-of-a-room.md), which counted the home of the Cahut among the pages that readers see.

## Alternatives considered

- **A redirect by the reader session**: to `/console/` with the cookie of a reader, to `/console/sign-in` without. The two sessions are not the same: a person signed in to the console who has opened no Window has no reader session, and would meet the page of sign-in while signed in. Only the console knows its session.
- **The home for readers, with a link to the console.** A page between the domain and the console, for a list that the console shows with previews.
- **The console at the root.** Its pages would take paths that Rooms may take, and ADR 20 serves it under `/console/` so that it shares nothing with them.
