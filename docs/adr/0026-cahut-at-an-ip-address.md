# 26. A Cahut can live at the IP address of its server until it has a domain

- Status: Accepted
- Date: 2026-10-23

## Context

[ADR 8](0008-several-cahuts-per-server.md) puts each Cahut at its own domain, with one A or AAAA record pointing to its server, and leaves to batch 7 a way to try Cahut without a domain, naming an IP address certificate as one option. The [V0 plan](../plan-v0.md) lists the question as open.

Without a domain, someone who wants to try Cahut on a new server has no name to give to `install.sh`. They need HTTPS all the same: browsers refuse a self-signed certificate with a full-page warning, and AI clients refuse it outright. Let's Encrypt has issued certificates for IP addresses since January 2026, under its `shortlived` profile only, valid for 160 hours, through the HTTP-01 and TLS-ALPN-01 challenges. CertMagic supports both the profile and IP addresses.

## Decision

- **A Cahut can be declared at the public IP address of its server**, IPv4 or IPv6, instead of a domain: `https://203.0.113.7/<room>/<window>/`, `https://[2001:db8::7]/console/`. Only global unicast addresses are accepted: not private, loopback, link-local or shared addresses. One address holds one Cahut, as one domain does.
- **Its certificate is short-lived.** CertMagic obtains it from Let's Encrypt under the `shortlived` profile, and renews it when a third of its life is left. Domains keep the classic profile, valid for 90 days. A client that opens an IP address sends no server name in TLS: Cahut picks the certificate from the local address of the connection.
- **`install.sh --ip <address>`** declares the Cahut at that address. With neither `--domain` nor `--ip`, the script takes the source address of the default route when it is public, and otherwise asks for one of the two options: behind NAT, the machine does not know the address that the world sees.
- **A Cahut moves to a domain with `rename_domain`**, from the command line (`cahut domain rename <old> <new>`) or in the console, for its administrators of the server. Everything stays: people, Rooms, Windows, house style, tokens and Logbook. The old address serves nothing any more, and the links that name it stop working. AI clients connect again at the new address, since their OAuth tokens name the old one. The new domain gets its certificate at its first TLS handshake, as ADR 8 has it. Any Cahut can move this way, from a domain too.
- **The server checks the address of each Cahut at each request**, by one read of `server.db`, so that a move made from the command line takes effect without restarting the service.

## Consequences

- Trying Cahut needs a server with a public address, and nothing else: no DNS record, no account anywhere but Let's Encrypt.
- A six-day certificate depends on Let's Encrypt every few days. An outage of a few days, or port 80 closed after the installation, leaves the Cahut with an expired certificate where a domain would have weeks of margin. `cahut doctor` reports the expiry.
- Some AI clients may refuse an MCP server at an IP address: batch 8 checks it with each client, and the documentation will say which.
- Moving to a domain breaks the links already shared, which is the price of starting without one. The console says so before the move.
- The address of a server can change, as when a VPS is rebuilt: the Cahut then needs a move, like a change of domain.
- Several Cahuts on one server still need domains: the IP address is for one, to try.

## Alternatives considered

- **A wildcard DNS service such as sslip.io**, with names like `203-0-113-7.sslip.io`. A real domain with a 90-day certificate, but it depends on a service run by a third party, and Let's Encrypt's limit of 50 certificates per registered domain each week is shared by every user of that service, unless it is on the Public Suffix List.
- **A name given by the project**, such as `<address>.cahut.cloud`. It would need the project to run DNS for every trial server, and to see each of them: against a self-hosted product with no service behind it.
- **A self-signed certificate.** Nothing to depend on, but browsers warn on every page, and AI clients refuse to connect.
- **No fallback: a domain is required.** The simplest, and the rule for real use, but trying Cahut would start with buying a domain and waiting for DNS.
