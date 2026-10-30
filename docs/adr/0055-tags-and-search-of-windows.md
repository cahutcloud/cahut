# 55. Windows carry tags, and people and AIs search them

- Status: Accepted
- Date: 2026-10-30

## Context

Windows pile up: a person publishes through several AIs, a team Room gathers what its members make, and each AI publishes again rather than update when it does not find a Window. `list_windows` gives the Windows of the Rooms of its caller, the most recently updated first, and the console shows them Room by Room. To find one, a person or an AI reads titles and descriptions ([ADR 50](0050-window-ids-permanent-addresses-and-names.md)).

Olivier asked on 29 October 2026 for tags on Windows and a small search engine. On 30 October he decided:

- The tags are free: each person and each AI writes its own, and Cahut keeps no list of them, for the Cahut or for a Room.
- Windows carry tags, Rooms do not: a person has few Rooms, the console lists them all, and each has a description.
- The search reads the text of the pages, besides their titles, descriptions and tags.
- It finds the Windows of the Rooms of its caller, as `list_windows` does, and no other.

The list of the tools that an AI gets weighs 31 KB, and [ADR 29](0029-chatgpt-second-ai-client.md) keeps it under 32 KB, for the clients that cut a longer one. A new tool weighs about half a kilobyte.

## Decision

### Tags

- **A tag** is 1 to 32 letters, digits and hyphens, in lower case, such as `q3`, `sales` or `board-meeting`. Letters with accents and of every script are letters: `été` and `日本` are tags.
- **People and AIs write tags as they come.** Cahut lowers them, drops a leading `#`, and joins their words with single hyphens: `#Board Meeting` and `board_meeting` both give `board-meeting`. A tag with anything else, such as `a/b`, is refused with the code `tag_invalid`.
- **A Window carries 10 tags at most**, refused beyond with the code `tags_too_many`. A tag given twice counts once. Cahut keeps them in order.
- **They belong to the Window, not to a version.** A new version and a rollback keep them; so do a new title, a new name, the trash and a new access. They go with the Window when it is deleted for good.
- **`publish_window` takes `tags`.** `retitle_window` changes them, besides the title and the description, without a new version: the list it gives replaces the old one, an empty list removes them, and they stay when it gives none. Its title becomes Retitle, describe and tag a Window. `update_window` does not change them.
- **Changing tags is undone** by giving the old ones back, which `get_window` gives: an AI does it alone, without knocking, and the Logbook records the new tags.
- **Who changes them**: whoever changes the title of the Window, the members of its Room and their AIs, or its author alone for a Window that opens to its author alone ([ADR 53](0053-windows-for-their-author-alone.md)).
- **`list_windows` gives the tags of each Window, and takes `tag`**, which lists the Windows with that tag only, in the Rooms of its caller. The MCP instructions and the skill tell AIs to give a few tags to each Window, reusing those that `list_windows` gives, so that one subject keeps one spelling.
- **Only the people of its Room see the tags of a Window**, and their AIs. The page of the Room that readers see, the Window, its permanent address and its card for robots of previews ([ADR 51](0051-sharing-a-window-and-its-preview.md)) show none: a tag is a tool to sort, and it may name a client or say draft.
- **In the console**, the cards of Windows show their tags. The page of a Room shows the tags of its Windows over their cards, each with how many carry it: a click shows only the Windows with that tag. The page of a Window changes them under Title, description and tags, in a field that suggests the tags of the other Windows of the person.
- **Storage**: a table `window_tags` holds a row for each tag of each Window, and goes with its Window.

### Search

- **`list_windows` takes `query`**: words, which a Window matches when its title, its description, its tags or the text of its pages holds each of them, at the start of a word, whatever the case and the accents: `ete bret` finds `Été en Bretagne`. The Windows come the best first, a word in the title weighing most, then in the tags, the description and the text; each comes with an `excerpt`, a few words of its text around the words found, when its text holds them. `query` goes with `room`, `tag` and `trashed`.
- **No tool is added**: `list_windows` lists and searches, and the list of tools stays under 32 KB.
- **It finds what `list_windows` lists**: the Windows of the Rooms of its caller, those of the Rooms granted to an AI, but those that others keep for themselves ([ADR 53](0053-windows-for-their-author-alone.md)). A Window in the trash comes only with `trashed`. A Window of another Room, even one that opens to the whole Cahut or to the public, does not come.
- **The text of a Window** is that of the HTML pages of its current version, `.html` and `.htm` files, without their markup, scripts and styles, up to 100 KB. Cahut reads it when a version is published and keeps it with that version, so that a rollback finds it again. The text that a script writes once the page has loaded is not there: a deck drawn by its scripts is found by its title, its description and its tags.
- **The Windows published before** get their text once, from the files of their versions, when their Cahut first opens with this version of Cahut.
- **The index** is an FTS5 table in the database of each Cahut, which the SQLite driver holds without cgo: no service apart, and nothing leaves the Cahut. It holds a row for each Window, and changes in the transaction of each call that changes what it holds.
- **In the console**, a search field at the head of every page of a person opens a page of results, with the cards of the Windows found and their excerpts, and the tags of the Windows found, which narrow them.
- **The Logbook** records the words of a search with the number of Windows found, as it records a list today.

### What stays out

- Tags on Rooms, and a list of tags kept by the Cahut or by a Room.
- A search of the Windows of other Rooms, and of the pages that Cahut serves to readers.
- A search in files other than HTML pages, such as PDF, CSV or JSON, and in the Logbook.
- Suggesting words, correcting them, or searching in several languages at once beyond accents and case.

## Consequences

- A person and their AIs find a Window among hundreds by a word of its page, or gather the Windows of one subject under a tag, in the console and through the same `list_windows`.
- An AI that looks for the Window it should update finds it by searching, instead of listing every Window and publishing a new one when it misses it.
- Migrations add `window_tags`, the text of each version and the FTS5 index. The database of a Cahut grows by the text of its Windows, up to 100 KB for each of their two versions, and by its index.
- Publishing reads the pages of the new version once more, to take their text; the files of a Window are 32 MiB at most, and its text 100 KB.
- Free tags may drift, `q3` beside `q3-2026` or `sales` beside `ventes`: the console suggests those in use, and the instructions ask AIs to reuse them.
- A Window drawn by its scripts is found by its words only if its title, description or tags hold them.

## Alternatives considered

- **A list of tags for the Cahut, or for each Room**, kept by its administrators or its owners. More order, but a screen to keep it, a refusal for every tag outside it, and a step before a person tags anything. Olivier chose free tags.
- **Tags on Rooms too.** Few Rooms, all in the console, each with a description: no need yet.
- **Tags that hold spaces**, such as `board meeting`. Closer to what people type, but harder to read in a list and to type in a call; hyphens keep a tag one word.
- **New tools: `set_window_tags`, `list_tags` and `search_windows`.** Clearer names, but about 1.5 KB more in the list of tools, over what ADR 29 allows. `retitle_window` already changes what describes a Window without a new version, and `list_windows` already lists it.
- **A search of the titles, descriptions and tags only.** Lighter, but a person remembers a word of a page more often than its title.
- **A search of every Window that its caller may open**, those of other Rooms open to the whole Cahut or to the public included. A search of the whole Cahut, but a new rule on what a person discovers, which their Rooms do not show them today.
- **A search service apart**, such as Meilisearch or Bleve. More features, but a service to run and keep beside each Cahut, or a dependency; FTS5 is already in the driver.
- **Reading the text with each search**, from the files on disk. Nothing to keep, but slow for many Windows, and a search would read files in a transaction.
