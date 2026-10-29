# 52. A share link opens a Window to whoever holds it, without an account or a Key, until it ends

- Status: Accepted
- Date: 2026-10-29

## Context

A Window opens to its Room, to the people of its Cahut, to anyone with its Key, or to the public ([ADR 19](0019-window-versions-access-and-declaration.md)). To show a Window to someone outside the Cahut, its people today give it a Key and pass the Key on apart, invite the reader by e-mail, which gives them a personal link that stands for the Key ([ADR 31](0031-e-mail-through-a-relay.md)), or make it public, which knocks first when an AI asks ([ADR 21](0021-ais-knock-first.md)). Each of these changes the access of the Window for everyone, or needs a relay and an address.

People ask for a link that is long, cannot be guessed, belongs to one sharing, opens a Window without an account or a Key, and ends: a link to paste in a message, in Slack or in a document, with an end date, a number of openings if they want one, a way to take it back, and each opening in the Logbook. It is the personal link of an invitation without a recipient.

[ADR 50](0050-window-ids-permanent-addresses-and-names.md) gave each Window an id that never changes and a permanent address, `/-/w/<id>`, which is not a secret and opens nothing. [ADR 51](0051-sharing-a-window-and-its-preview.md) gave each Window a menu Share in the console, and a page of Cahut that answers the robots of previews at the permanent address and opens nothing; it left to this record what the preview of a link shows.

A Window that opens once, and an access for its author alone, come next, as a limit of the share link. This record keeps what they need.

## Decision

### A link beside the access

- **A share link is a grant of its own, added to the access of its Window**, not an access. A Window keeps its access, `room`, `internal` or `key`, for those it opens to, and each link opens it, besides, to whoever holds that link.
- **A Window has several links**, one for each sharing, each with its own end, count and label, taken back one by one. At most 20 links stand at once on a Window.
- **A public Window takes no link**: its permanent address already opens it to anyone. The call is refused, with what to do.
- A link names its Window inside the Cahut, by its integer id. It opens the Window, not its Room: the page of the Room does not list the Window for its reader.

### Making a link

- `create_share_link` makes a link to a Window, with an end, an optional number of openings and an optional label. The members of its Room make one, as they set its access and its Key; administrators do not, outside their own Rooms ([ADR 13](0013-people-sessions-and-rights.md)).
- **An AI makes a link without knocking first**, as it sets a Key without knocking and reads it ([ADR 19](0019-window-versions-access-and-declaration.md)): a link gives no more than a Key that the AI could pass on anywhere, and less, since it ends within 30 days and counts its openings. The output gives the AI the link, once, to give to its person. `revoke_share_link` undoes it, for the test of governance, and `create_share_link` undoes `revoke_share_link` with a new link, as for invitations.
- **The link is shown once.** It holds a random token of 256 bits, written in 52 characters of base32 in lowercase, and Cahut keeps only its SHA-256 hash, as it keeps the codes of invitations. Neither the console nor `get_window` gives it again: a person who loses it makes another one.
- **A label**, of 100 characters at most, on one line, tells the links of a Window apart, such as `For Acme` or `Slack #design`. Only the members of the Room see it, in the console and in `get_window`.

### The address of a link

- **A link is `https://<domain>/-/s/<token>`**, in the path, under `/-/`, which no Room can take ([ADR 9](0009-rooms-and-windows-as-paths.md)). The server knows the link from the first request: it answers with a page of Cahut, rendered on the server and without a script, as the page of a Key is, and it gives the robots of previews their tags at the same address.
- The token never reaches the Window. The page of the link sends its reader to the address of the Window after the click, and is served with `Referrer-Policy: no-referrer`, so that no request it makes, and no page that follows, carries the link. Cahut logs the requests of OAuth and of MCP only, never the address of a page.
- The link is not the permanent address, and does not carry the id of its Window.

### The page of a link, and what an opening is

- **Every visit to the address of a link gets its page**, `200 OK`, which opens nothing and counts nothing: a person before the click, a robot of previews, the filter of a mail service that follows links, a browser that loads pages ahead. The page names the Window by its title, gives the domain of the Cahut and the end of the link, and how many openings are left when it has a number of them, with a button, Open `<title>`, in a form that posts to the same address. It names neither who made the link nor its label.
- **An opening is the click.** The form posts to the address of the link; `open_share_link` spends one opening and gives the reader a pass. That opening is a session of reading, not a request: the page of the Window, its frame and each of its files, its images, its styles and its scripts, count nothing more.
- **A reader who opens the Window already goes straight to it**, without spending an opening: a member of its Room signed in, a reader with a pass of its Key, or a reader who opened the same link before and still holds its pass. The visit to the link sends them to the address of the Window.
- **A link that no longer opens**, because it ended, was taken back, has spent its openings or names nothing, gives one page, `404 Not Found`, that says the link opens nothing and to ask whoever gave it for a new one, without saying which. A Window in the trash gives the same page.
- Visits to the addresses of links are limited by address, as sign-in is.

### Reading after the click

- **The pass of a link** is a pass of [ADR 19](0019-window-versions-access-and-declaration.md), in the cookie of the passes of its Window, `__Host-cahut-key-<id>`, kept hashed, which names the link that gave it. It opens the Window whatever its access, until the link ends: a link that is taken back ends its passes, and a link that spends its openings keeps those it gave.
- The page of the Window then frames it as it frames a Window with a Key: the capability of the frame names the pass, and each file checks that the pass still stands. A link taken back closes the Window at the next file.
- A new Key, or an access other than `key`, ends the passes of the Key and its invitations, and not the passes of links.

### How long a link lasts

- **A link always ends.** Its end is 7 days after it is made unless it is given, and 30 days at most, as an invitation. A link does not change once made: a person makes a new one and takes the old one back.
- **A number of openings is optional**, from 1 to 1000. Once they are spent, the link opens nobody new; those who opened it keep reading until it ends, unless it is taken back.
- The end of the access of the Window does not end its links, which do not depend on its access.

### The Logbook

- `create_share_link` is logged with the number of the link, its end and its number of openings; `revoke_share_link` with the number of the link. Neither gives the token or the label.
- **Each opening is logged** as `open_share_link`, in the Room of the Window, so that its members read it, with the number of the link and of the opening, such as `link 3, opening 2 of 5`. Its actor is someone with a link, unnamed, even when the reader is signed in to the Cahut: the Logbook keeps neither their account, nor their address on the network, nor their browser.
- A visit to the page of a link, and a reader who goes straight to the Window, are not logged.

### The preview of a link

- The page of a link carries the tags of [ADR 51](0051-sharing-a-window-and-its-preview.md), for every robot, those that are not on its list included: a robot needs no other page, since the page opens nothing.
  - `og:title` is the title of the Window: the link opens the Window to whoever holds it, and its page shows the title, as the page of a Key does.
  - `og:description` is `Opens with this link until <date>.`
  - `og:url` is the link itself, so that a platform that follows `og:url` does not send its readers to the permanent address, which does not open for them.
  - `og:image` is the card of the Window at its id, as ADR 51 draws it: with its title for a Window with a Key, without it for a Window open to its Room or to its Cahut. The token lives at one address only.
- A link that no longer opens gives its page of `404 Not Found`, and no preview. The previews that platforms drew stay in their messages, with the title.

### The life of the Window

| When | The links |
| --- | --- |
| The Window changes its name, its title, its description or its version | Open it as it is now, at its new address |
| It goes to the trash | Open nothing; once restored, they open it again until their end |
| It is purged | Go with it, and their passes |
| Its access changes, or ends | Stand, whatever the access: a public Window opens to all, and its links send their readers straight to it |
| Its Cahut is suspended ([ADR 48](0048-suspending-a-cahut.md)) | Give the page of a suspended Cahut, with every address of the Cahut; their time goes on |
| The person who made a link is disabled, leaves the Room, or is deleted | Stand until their end: a link belongs to its Window, as its Key does, and the members of the Room take it back |

### Operations and bounds

| Operation | Who calls it | What it does |
| --- | --- | --- |
| `create_share_link` | The members of the Room, and their AIs | Makes a link, and gives it once |
| `revoke_share_link` | The members of the Room, and their AIs | Ends a link and the passes it gave |
| `open_share_link` | The page of a link, for its visitors only (`PeopleOnly`, `Visitors`) | Spends an opening and gives a pass |

- `get_window` lists the links of the Window, the latest first, without their token: their number, label, who made them and through which AI, when, their end, their openings and the last of them, and when they were taken back. There is no `list_share_links`: the list of tools of AIs stays under 32 KB ([ADR 29](0029-chatgpt-second-ai-client.md)). `get_window` also gives the number of links that stand.
- `create_share_link` is marked open world: it opens a Window to people outside the Cahut. `revoke_share_link` is marked destructive.
- A person makes 100 links a day at most, their AIs included, as for invitations, and a Window holds 20 links that stand.
- No setting of [ADR 30](0030-kind-and-limits-of-each-cahut.md) counts links: they take no disk, and open Windows that the limits count already. An individual Cahut takes links as a collaborative one: they name nobody.
- The refusals that people meet in the console carry a code ([ADR 37](0037-refusals-with-a-code-for-the-console.md)): a public Window, an end past or beyond 30 days, a number of openings out of bounds, too many links on a Window or in a day.

### In the console

- **The menu Share** of a Window that is not public gets an entry, Create a share link, which asks for a label, an end, 1 day, 7 days, 30 days or a date, 7 days by default, and a number of openings, none by default. Once made, the link shows with a button to copy it, and the menu says that Cahut shows it this once.
- **The page of the Window** lists its links under its access: the links that stand, then those that ended, with their label, who made them, their end, their openings, and a button to take back each link that stands. The access says how many links stand, so that a Window open to its Room does not look closed when links open it.
- The words go in English and French, as the rest of the console ([ADR 36](0036-console-in-english-and-french.md)). The page of a link and its preview speak English, as the other pages that readers see.

### Later

- **A Window that opens once** is a link with one opening whose opening moves the Window to a new access, for its author alone. Its record builds on this one: the page of the link is the step it needs, where the click spends the opening and the visit does not; its pass may last a session of reading rather than the life of the link.
- A notice to the person who made a link when it is first opened, or has spent its openings.
- A setting that turns links off in a Cahut, if an organization asks for it.
- Links to a Room, once Rooms have a permanent address.

## Consequences

- A Window open to its Room can be read by someone outside the Cahut, for a few days, without an account, a relay or a Key to pass on apart, and without changing who else opens it.
- The access of a Window no longer says alone who opens it: the console and `get_window` add its links.
- Whoever holds a link opens the Window: a link that is forwarded opens it to the next person, until its end or its count. The count, the end and the revocation bound it, as its secret bounds a Key.
- An AI makes links without its person, as it sets Keys. The Logbook names it, and the person takes the links back in the console. The vendor of the AI holds the link, as it holds a Key that the AI draws.
- The preview of a link shows the title of its Window wherever it is pasted, and keeps it after the link ends.
- A filter of mail that clicks the button of the page spends an opening, and the person sees it in the Logbook. A filter that only follows the link spends nothing.
- The links of a person who is disabled or leaves a Room stand until they end, 30 days at most, unless the members of the Room take them back.
- A migration adds the links and their counts to the database of each Cahut, and to each pass the link that gave it.
- This record extends [ADR 19](0019-window-versions-access-and-declaration.md) (a pass may come from a link, and opens its Window whatever its access), [ADR 50](0050-window-ids-permanent-addresses-and-names.md) (a second address under `/-/` that opens a Window) and [ADR 51](0051-sharing-a-window-and-its-preview.md) (the menu Share makes links, and the page of a link gives its preview).

## Alternatives considered

- **An access `link`**, beside `room`, `internal`, `key` and `public`. One link per Window, which would replace its access: the Window would stop opening to its Room or to its Key for the time of the link, and a link given to two people could not be taken back from one of them.
- **A link that knocks first** when it opens a Window that its Room alone reads. A wait without a guard, since the AI can set a Key without knocking and read it; and the AI could not give the link, a secret, to its person, who would find it in the console.
- **The token in the fragment of the permanent address**, `/-/w/<id>#<token>`. The server would never receive it, but could not tell a link from the permanent address: robots would preview the Window as a stranger sees it, a script would have to read the fragment, and the redirect of the permanent address would carry the fragment into the Window.
- **The page of a link in the console**, as for invitations, at `/console/share#<token>`. In English and in French, but with a script, a code to exchange, and robots that get the console without tags.
- **The link shown again** in the console, kept in clear: a leak of the database would open every Window that has a link.
- **An opening at the first visit.** No click, but the robots that are not on the list and the filters of mail would spend it before the reader.
- **An opening for each request, or for each address on the network.** A page loads many files, and the address of a phone changes as it moves.
- **The address and the browser of readers in the Logbook.** They would tell a link that spread, but give the members of the Room personal data of people who are not in the Cahut.
- **A card with the title at the address of the link.** The card would show what the tags show, but the token would live at a second address, which platforms fetch and keep.
- **No title in the preview of a link** to a Window open to its Room, as ADR 51 has it for its permanent address. Nothing in the servers of the platforms, but a link meant for someone outside the Cahut would show bare, while it opens the Window to whoever holds it.
- **Links that end with the person who made them**, disabled or out of the Room. Safer after an account is taken, but a link given to a client would stop with the departure of a colleague; links last 30 days at most, and the members of the Room see and take them back.
- **Links that end when the access of the Window changes.** A person who closes a Window would close its links too, but a link is meant for a Window whose access stays closed.
