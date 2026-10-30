# 53. A Window can open to its author alone, and a share link that opens once gives it a single reader

- Status: Proposed
- Date: 2026-10-30

## Context

A Window opens to its Room, to the people of its Cahut, to anyone with its Key, or to the public ([ADR 19](0019-window-versions-access-and-declaration.md)). The narrowest of these is the Room: every member of a team Room, and the AIs of each, list the Window, open it, change it and put it in the trash, and the Logbook of the Room tells them what happens to it.

[ADR 52](0052-share-links.md) added share links, beside the access: a link opens a Window to whoever holds it, until it ends, to a number of readers if one is given. Its page, at `/-/s/<token>`, opens nothing and counts nothing; the click on its button spends an opening, so that robots of previews and filters of mail spend none. Its pass opens the Window until the link ends. A link may already open once.

People ask for a Window that opens once: its link no longer works after the first opening. ADR 52 left it for this record, as a limit of the share link. Olivier decided on 29 October 2026 that after the opening only the link dies: the Window stays, for its author alone, and does not go to the trash; in a team Room too, the other members no longer see it.

That needs an access narrower than the Room, which Cahut does not have. People also ask to keep a draft to themselves in a team Room, until they share it with the Room.

Each Window already names the person it was first published for, `created_by`, whether they published it themself or through their AI. [ADR 47](0047-deleting-a-person.md) empties it when that person is deleted.

## Decision

### An access for its author alone

- **A new access, `author`**, opens a Window to its author alone, signed in, while they are in its Room. It joins the table of ADR 19:

  | Access | Who opens it |
  | --- | --- |
  | `author` | Its author, signed in, while they are in its Room. |
  | `room` | The members of its Room, signed in. The default. |
  | `internal` | Everyone signed in to its Cahut. |
  | `key` | Anyone who types its Key, and the members of its Room. |
  | `public` | Anyone who has its address. |

- **The author of a Window is the person it was first published for**, `created_by`, whether they published it themself or through their AI. It never changes: not with a version that another member publishes, nor a new title, name or access, nor the trash. Every Window that exists has its author already, but those whose person was deleted, which have none and take no `author`.
- **Only its author keeps a Window for themself.** `set_window_access` to `author` is refused to anyone else, with a code. Once a Window opens to its author alone, only they act on it: they update it, retitle and rename it, roll it back, set its access, make and take back its links, put it in the trash, restore and purge it.
- `author` has no end, as `room` has none: it lasts until its author sets another access.
- The passes of its Key end, as with any access other than `key`. Its share links stand, as they stand whatever the access ([ADR 52](0052-share-links.md)): its author sees them all, those that other members made included, and takes them back.
- **`publish_window` takes an access**: `room`, the default, or `author`. A Window published for its author alone never shows to the Room, not even for the time between two calls. The other accesses stay with `set_window_access`.
- **A draft** is a Window for its author alone: an AI publishes it with `author`, and `set_window_access` to `room` shares it with the Room when it is ready.
- In a personal Room, `author` and `room` open the Window to the same person. `author` is allowed there all the same, so that an AI makes a Window that opens once in the same way in every Room.

### A Window that opens once

- **It is not a thing of its own**: it is a Window for its author alone, with a share link that opens once. Cahut stores nothing more, and nothing moves at the opening: the Window opened to its author alone before the click, and still does after; the link has spent its opening, and opens nobody new.
- **For an AI**, it takes two calls: `publish_window` with the access `author`, then `create_share_link` with `max_openings` 1, which gives the link once. The MCP instructions and the skill give this recipe. No tool is added, and the list of tools stays under 32 KB ([ADR 29](0029-chatgpt-second-ai-client.md)).
- **In the console**, the menu Share of a Window that is not public gets an entry for its author, Share once…, beside Create a share link…: the same dialog, with a label and an end, which makes a link that opens once, and a box, Keep it for me alone, checked, which sets the access to `author` first. When the Window opened to its Room, the dialog says that the Room will no longer see it.
- A link that opens once to a Window that opens to its Room, or to its Cahut, stays what ADR 52 has: the Room keeps its Window, and one reader outside it opens the Window once.

### The page of a link that opens once, and its pass

- **Before the click**, only the author and the link open the Window. A visit to the page of the link spends nothing, whoever makes it: a person, a robot of previews, a filter of mail, a browser that loads pages ahead ([ADR 52](0052-share-links.md)).
- The page of a link that opens once says it, under its button: the link opens once; whoever opens it reads the Window in this browser, for 4 hours; nobody else can open it after. Its `og:description` is `Opens once, until <date>.`
- **The pass of a link that opens once lasts a session of reading**: 4 hours after the click, as long as a frame of the Window lasts ([ADR 19](0019-window-versions-access-and-declaration.md)), and never beyond the end of the link, in a cookie without an end date, which the browser drops when it closes. The reader reloads the Window, follows its pages and comes back to the link during that time, and goes straight to the Window. This holds for every link that opens once, whatever the access of its Window: one opening is one reading. The other links keep their passes until they end.
- **After the click**, the link has spent its opening: whoever comes next gets the page of a link that opens nothing, `404 Not Found`. Once the pass ends, its reader gets the same page.
- **Its author** then finds the Window in their lists and on its page in the console, for them alone, with the link opened, when, and the opening in the Logbook. They make a new link, open the Window to the Room, or put it in the trash.

### A new link

- The author makes another link to the Window for another reader, as to any Window: each link that opens once has its own reader, and the 20 links that stand on a Window bound them ([ADR 52](0052-share-links.md)).
- A link taken back before the click opens nothing, and the author makes another one.

### What the others see

| Who | What they see of a Window for its author alone |
| --- | --- |
| Its author, and their AIs | The Window, as any Window of theirs: in the lists, the previews and the page of its Room, with its links and the entries of the Logbook about it |
| The other members of its Room, its owners included, and their AIs | Nothing: neither in `list_windows`, the console and its previews, nor on the page of the Room. A call that names it, by its id, its address or its name, gets no Window, as for a Window that does not exist |
| The administrators of the Cahut | Of the Windows outside their Rooms, the totals and the Logbook, which they read whole ([ADR 13](0013-people-sessions-and-rights.md), [ADR 30](0030-kind-and-limits-of-each-cahut.md)); in their own Rooms, what members see |
| Whoever holds a share link to it | The Window, through the link, as [ADR 52](0052-share-links.md) has it |
| A robot of previews | At the permanent address, what a Window open to its Room shows: `A Window at <domain>`, and the card without a title ([ADR 51](0051-sharing-a-window-and-its-preview.md)). At a link, the tags of ADR 52 |
| The command line | Every Window, as before |

- `get_room` and `list_rooms` count the Windows that their caller sees.
- **The home of a Room** ([ADR 32](0032-the-page-of-a-room.md)): its owners choose among the Windows they see. A home that its author then keeps for themself shows to them alone, and the other readers get the list, as for any reader who does not open the home; `get_room` names it to its author only.
- **Usage and limits** ([ADR 30](0030-kind-and-limits-of-each-cahut.md)): a Window for its author alone counts as any other, in the Windows of its Room and of its Cahut, and in the disk. A member at the limit of a Room may thus see fewer Windows than the limit counts.
- **Names**: its name stays taken in its Room. A member who chooses it hears that it is taken, by a Window or one in its trash, as today, and nothing more.
- **Invitations** ([ADR 31](0031-e-mail-through-a-relay.md)): `invite_to_window` refuses a Window for its author alone, with a code, and says to make a share link: nobody else signs in to read it, and it has no Key.
- **The Logbook**: the entries about a Window for its author alone go to its author, not to the members of its Room: its publication, its new versions, its title and its name, its links and each opening. The entry that keeps it for its author goes to the Room, which loses it, and so does the entry that opens it to the Room again; the entries that follow go to the Room. Entries that went to the Room before stay there. The administrators read every entry, as before.

### When the author leaves

| When | Their Windows for themselves alone |
| --- | --- |
| The author is disabled | Stay as they are, with their links, until the person is enabled again: disabling keeps what is theirs ([ADR 47](0047-deleting-a-person.md)) |
| The author leaves the Room, or is removed from it | Open to the Room again, with the access `room`: what is published in a team Room belongs to the Room, and its owners find it. The console says how many, before; `remove_room_member` knocks first for an AI already. The Logbook entry of the removal gives their number |
| The author is deleted | Open to their Room again, without an author, as ADR 47 keeps the other Windows of the person in team Rooms; those of the personal Room go with it |
| The Room is deleted | Go with it, as every Window of the Room: the console says so before |

### Knock first

- **Neither keeping a Window for its author nor opening it again knocks.** `set_window_access` to `author` is undone by `set_window_access` to the access it had. Opening it to the Room, to the Cahut or to a Key from `author` stays among people that the organization chose, as ADR 19 lets an AI widen `room` to `internal` or `key`, and is undone by `set_window_access` too. Making it public still knocks first.
- `publish_window` with the access `author` is undone by `trash_window`, which closes its links ([ADR 52](0052-share-links.md)); a link that opens once is undone by `revoke_share_link`, as any link.
- No operation is added, and the map of the test of governance does not change.
- An AI hides from the Room only the Windows of its person, as its person could, and the Logbook of the Room says it.

### Refusals

The refusals that people meet in the console carry a code ([ADR 37](0037-refusals-with-a-code-for-the-console.md)):

| Code | When |
| --- | --- |
| `window_author_only` | Someone else than its author asks for `author` |
| `window_without_author` | `author` for a Window whose person was deleted |
| `invitation_author_only` | An invitation to a Window for its author alone |

### In the console

- The access of a Window gets a first choice for its author, Only you, with an icon of its own. The other members do not see it.
- The cards of the Window, its page and the Share menu say Only you, with the same icon. Its share links show under its access, as ADR 52 has it, so that its author sees which link stands and which was opened.
- Leaving a Room, or removing a member, says how many Windows that the person keeps for themself will open to the Room.
- Deleting a Room says that the Windows that its people keep for themselves go with it.
- The words go in English and French ([ADR 36](0036-console-in-english-and-french.md)). The page of a link speaks English, as the other pages that readers see.

### Later

- Moving a Window to another Room ([ADR 50](0050-window-ids-permanent-addresses-and-names.md)), so that an author who leaves a team Room takes their Windows to their personal Room.
- Handing a Window to another author.
- A notice to the author when a link that opens once is opened ([ADR 52](0052-share-links.md)).

## Consequences

- A person keeps a Window for themself in a team Room, as a draft or for a reader outside the Cahut, and sends it with a link that opens once: the Room never sees it, and its author keeps it after.
- Once applies to the link, not to what the reader does: during the 4 hours they may save, print or copy what they read. Cahut bounds who opens a Window, not what its readers keep.
- A filter of mail that clicks the button of a link spends its only opening, and its reader then finds a link that opens nothing. Its author sees the opening in the Logbook, and makes a new link.
- A reader who closes their browser, or comes back after 4 hours, needs a new link.
- A Room no longer sees all its Windows. Its owners run it without seeing what its members keep for themselves, which counts in its limits and its disk, and goes with it when it is deleted.
- An owner who removes a member opens to the Room the Windows that the member kept for themself. The console says it before, and the Logbook keeps it.
- A team Room may hold a Window that nobody opens while its author is disabled. Deleting the person, or enabling them, settles it.
- Each check of who opens a Window, and each list of Windows, compares the author with the person, for a column that the database already reads.
- A migration makes the table of Windows again, to admit the access `author`, and adds to the Logbook the person who reads an entry in place of the members of its Room.
- This record extends [ADR 19](0019-window-versions-access-and-declaration.md) (an access for the author alone, which `publish_window` takes too) and [ADR 52](0052-share-links.md) (the pass of a link that opens once lasts a session of reading, and its page says so).

## Alternatives considered

- **A Window that opens once, as a thing of its own**, with a flag of its own and a tool. Clearer to name, but a third way to open a Window beside its access and its links, and a tool more in the list that AIs load.
- **`once` on `publish_window`**, which publishes for the author and makes the link in one call. One call less for the AI, but a field of the list of tools for what two calls do, and a link without its label or its end.
- **A link that moves its Window to `author` at its opening**, as the first idea had it. The Room would see the Window until the click, and lose it at the click of a stranger: the access of a Window would change at the hand of a visitor.
- **The Window in the trash after the opening.** Olivier ruled it out: its author keeps what they sent.
- **The pass until the end of the link**, as ADR 52 has it for the other links. The reader of a Window that opens once could come back to it for 30 days.
- **A pass for the page alone**, without a session. The Window would close at the first reload, and the files of a page that load late would fail.
- **The owners of the Room see a Window for its author alone**, as an administrator sees everything in many tools. The draft of a person would not be theirs, and ADR 13 keeps administrators out of the content of others.
- **Windows for their author alone that stay hidden when the author leaves the Room**, until they come back. The Room would hold, for good, Windows that nobody opens nor manages, in its limits and its disk.
- **Windows for their author alone deleted when the author leaves the Room.** What is theirs alone would go with them, but an owner who removes a member would delete the work of the member, which cannot be undone.
- **The entries of a Window for its author alone in the Logbook of the Room.** The members would read the title of a Window that they do not see.
- **The access named `private` or `me`.** `private` says nothing of who opens it, and `me` names a different person to each reader of the record; `author` names the one person.
