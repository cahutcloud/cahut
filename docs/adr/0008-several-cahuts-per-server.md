# 8. A server holds several Cahuts, each at its own domain

- Status: Accepted, extended by [ADR 26](0026-cahut-at-an-ip-address.md) (a Cahut can live at the IP address of its server) and by [ADR 30](0030-kind-and-limits-of-each-cahut.md) (each Cahut has a kind and limits, set by the administrators of the server)
- Date: 2026-10-21

## Context

[ADR 3](0003-one-origin-per-window.md) gave each Window a subdomain of the home. It needed a wildcard DNS record, which many people do not know how to create, and one certificate per Window, which runs into Let's Encrypt's limit of 50 new certificates per registered domain every 7 days.

The word "home" named both the product's metaphor and the instance on a server. An instance is now called a Cahut, with a capital, like the product.

People also want more than one Cahut on a server: one per brand, per client or per team, or, for a host, one per organization. With one machine per instance, each of them needed a server of its own.

## Decision

- A Cahut lives at exactly one domain name, such as `cahut.example.com` or `example.org`, with one A or AAAA record pointing to its server. It needs no wildcard record and uses no subdomain of its domain.
- A server runs one `cahut` service, which holds one or more Cahuts. It routes each connection and each request to the Cahut of its host name, from the TLS server name and the `Host` header. A name that no Cahut declares gets `421 Misdirected Request`.
- Cahuts share nothing but the server. Each has its own people, Rooms, Windows, house style, tokens, OAuth clients and Logbook. A person of one Cahut has no account in another.
- Administration has two levels. The server's administrators declare Cahuts and their domains, and can act in every Cahut. A Cahut's administrators manage its people and its Rooms.
- Each Cahut keeps its records in a SQLite database of its own ([ADR 4](0004-sqlite-in-pure-go.md)), in a directory of its own with its Window files: `cahuts/<id>/` in the data directory. The server's records live in `server.db`: the Cahuts and their domains, the server's administrators, and the Logbook of server operations. The server finds the Cahut of a request before it reads any record, so a query cannot reach the records of another Cahut.
- CertMagic ([ADR 2](0002-built-in-https-with-certmagic.md)) manages one certificate per Cahut, for its domain only, and obtains those of all known Cahuts at start. A Cahut declared from the console while the server runs gets its certificate at once. One declared from the command line gets it at the first TLS handshake for its domain, through on-demand issuance whose decision function accepts declared domains only.

## Consequences

- The first start asks for one ordinary DNS record. Publishing needs no new certificate: a Cahut has one, renewed through ARI, however many Windows it holds.
- All the Windows of a Cahut share one origin. [ADR 9](0009-rooms-and-windows-as-paths.md) sets their addresses and [ADR 10](0010-windows-in-a-csp-sandbox.md) isolates them.
- A Cahut is backed up, moved or deleted with its directory and its row in `server.db`.
- Cahuts share the disk, the memory, the processor, the updates and the backups. One Cahut can fill the disk for all the others: quotas per Cahut come after the V0.
- Changing the domain of a Cahut changes the address of every Window it holds.
- Migrations run on `server.db` and on the database of every Cahut at start.
- The nip.io fallback of ADR 3 goes away. A trial without a domain is settled in batch 7 of the [V0 plan](../plan-v0.md); an IP address certificate is one option.
- The product rule "one machine per instance" becomes "one server, one or more Cahuts".

## Alternatives considered

- **A subdomain per Window, as in ADR 3.** The strongest isolation, at the cost of the wildcard record and of the certificate limit.
- **One installation per Cahut.** Each would need ports 80 and 443 for itself, so a proxy in front of them, against a binary that serves HTTPS itself.
- **One database for all Cahuts**, with a Cahut column in every table. Simpler migrations, but one condition forgotten in one query leaks records between Cahuts.
- **Accounts shared by the whole server.** One sign-in for a person who works in several Cahuts, but Cahuts would no longer be sealed off from each other, which a host serving several organizations needs.
