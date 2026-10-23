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

The Cahuts of a server share its disk, memory and processor, its updates and its administrator. One Cahut can fill the disk for all the others: quotas per Cahut come later.

## Who administers

The server's command line acts as the administrator of the server and of every Cahut: whoever runs `cahut` as the service user can read its data folder anyway. A person created with `--admin` is recorded as an administrator of their Cahut. That role takes effect with the console, where the administrators of a Cahut will manage its people.
