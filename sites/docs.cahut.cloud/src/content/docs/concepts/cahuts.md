---
title: Cahuts and servers
description: One server holds one or more Cahuts, each at its own domain, with its own people, tokens, Windows and Logbook.
---

A server runs one `cahut` service, which holds one or more Cahuts. Each Cahut lives at exactly one domain, such as `cahut.example.com` or `example.org`. A server can hold a Cahut per brand, per client or per team, or, for a host, one per organization.

## A domain per Cahut

A Cahut needs one DNS record, A or AAAA, that points its domain to the server. It needs no wildcard record and uses no subdomain: its Rooms and Windows are paths of its domain, such as `cahut.example.com/marketing/q3-review/`.

The service routes each request to the Cahut of its host name. A name that no Cahut declares gets `421 Misdirected Request`. The service obtains one certificate per Cahut from Let's Encrypt, for its domain only, and renews it. It never asks for a certificate for a name that no Cahut declares.

## Without a domain

To try Cahut before it has a domain, a Cahut can live at the public IPv4 address of its server, such as `https://203.0.113.7/`, with a certificate of Let's Encrypt for that address. A server holds one such Cahut. It moves to a domain later, with everything in it: its links at the address stop working, and AIs connect again. See [Without a domain](/install/server/#without-a-domain).

## Sealed off from each other

Cahuts share nothing but the server. Each has:

- its own people and Rooms: a person of one Cahut has no account in another;
- its own tokens: a token of one Cahut is invalid in the others;
- its own Windows and their files;
- its own Logbook;
- its own SQLite database, in a folder of its own.

The service finds the Cahut of a request before it reads any record, and then opens the database of that Cahut only. No request can reach the records of another Cahut.

The server keeps records of its own: the Cahuts and their domains, and the Logbook of the operations on the server, such as declaring a domain.

## On disk

```text
/var/lib/cahut/
├── server.db              the Cahuts, their domains and the Logbook of the server
├── certificates/          certificates and ACME account keys
└── cahuts/
    ├── 1/
    │   ├── cahut.db       people, Rooms, tokens, Windows and Logbook of the first Cahut
    │   └── windows/
    │       └── 7/1/       the files of version 1 of Window 7
    └── 2/
```

The folders of the Cahuts and of the Windows are named by their ids, not by their domains or names. [The cahut service](/admin/service/#back-up) says how to back them up.

## What Cahuts share

The Cahuts of a server share its disk, memory and processor, its updates and its administrator. Limits keep one Cahut from taking what the others need.

## Collaborative or individual

The administrators of the server choose the kind of each Cahut, in its console under **Kind and limits**, or with [`cahut settings set`](/admin/commands/#set-the-kind-and-limits-of-a-cahut). A Cahut is collaborative until they say otherwise.

- **Collaborative**: people create team Rooms, add one another to them, and open Windows to the whole Cahut.
- **Individual**: people share the Cahut, not their work, and do not see one another, such as the clients of one consultant or the students of one course.

In an individual Cahut:

- every Room has one member, its owner: a person creates Rooms of their own beside their personal Room, and nobody can be added to a Room;
- the name of a Room starts with the handle of its owner, such as `bob-projects`, so that nobody meets the names of the others;
- only the administrators of the Cahut list its people;
- a Window opens to its Room, to anyone with its Key, or to the public, never to everyone signed in to the Cahut.

The administrators of the Cahut see everyone, every Room and the whole Logbook, and open only the Windows of their own Rooms, as in any Cahut. A collaborative Cahut becomes individual once no Room has more than one member and no Window opens to the whole Cahut; an individual Cahut can always become collaborative.

## Limits

The administrators of the server can bound what each Cahut holds, such as when a host gives each organization its share of a server. A Cahut has no limit until they set some, in its console under **Kind and limits**, or with [`cahut settings set`](/admin/commands/#set-the-kind-and-limits-of-a-cahut). Each limit is optional:

| Limit | What counts |
| --- | --- |
| People | The people who are active or wait for their activation, administrators included. Disabled people do not count. |
| Administrators | The administrators of the Cahut, counted the same way. |
| Team Rooms | Team Rooms. A personal Room comes with its person, and the limit of people already bounds them. |
| Windows | The Windows out of the trash, in the whole Cahut. |
| Windows per Room | The Windows out of the trash in one Room, personal Rooms included. |
| Disk | Every version of every Window, those in the trash included, and the files of the house style, in units of 1000 bytes: 10 GB is 10 billion bytes. |

In an individual Cahut, three more limits bound each person, the same for everyone: their Rooms beside the personal one, their Windows out of the trash, and the disk that their Windows keep. No person can then take what the Cahut gives to all, and a refusal speaks of the person's own use. A member of an individual Cahut reads the limits per person and their own use, not what the whole Cahut holds.

A call that would cross a limit is refused, with a message that says what to do: put Windows in the trash, purge some, or ask for a higher limit. Lowering a limit below what the Cahut holds deletes nothing: the Cahut keeps it all, and refuses what would add to it.

Every change of the limits goes to the Logbook of the Cahut, and `get_cahut` gives the limits with what the Cahut holds, to people and to AIs.

## Legal pages

A host who runs Cahuts for the public publishes a legal notice, terms of use and a privacy policy on a site of its own. The administrators of the server link them from each Cahut, in its console under **Kind and limits**, or with [`cahut settings links`](/admin/commands/#link-the-legal-pages-of-a-cahut). A Cahut links nothing until they do, and its pages then show nothing more.

Each of four pages is optional: the legal notice, the terms, the privacy policy, and where people get help. Each is the address of a page that starts with `https://`; Cahut links these pages and serves none of them. A page may have another address in French, which the console links when it speaks French; it takes the place of the other address in French only, and needs it.

The pages that the Cahut serves itself link them in a discreet line at their foot, in the language of the page:

- in the console, under the sign-in, the activation links, the confirmation of an address, invitations and the consent screen of AI clients, and at the foot of every screen once signed in;
- on the pages that readers see: the page of a Room, the page that asks for a Key, what is not found, and the links that have expired.

Windows show none of them: a Window is the content of its Room, and a Room that shows a Window at its address leaves the page to it. Every change of the links goes to the Logbook of the Cahut, and `get_cahut` gives them to people and to AIs.

## Who administers

The server's command line acts as the administrator of the server and of every Cahut: whoever runs `cahut` as the service user can read its data folder anyway. A person created with `--admin` is recorded as an administrator of their Cahut. That role takes effect with the console, where the administrators of a Cahut will manage its people.
