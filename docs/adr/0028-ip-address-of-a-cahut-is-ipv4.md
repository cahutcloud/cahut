# 28. The IP address of a Cahut is an IPv4 address, for now

- Status: Accepted
- Date: 2026-10-23

## Context

[ADR 26](0026-cahut-at-an-ip-address.md) lets a Cahut live at the public IP address of its server, IPv4 or IPv6, until it has a domain.

[ADR 10](0010-windows-in-a-csp-sandbox.md) and [ADR 19](0019-window-versions-access-and-declaration.md) serve every file of a Window in a CSP sandbox: its origin is opaque, so `'self'` matches nothing, and the policy of each file names the origin of its Cahut, such as `https://cahut.example.com`, for the house style, the deck engine and the Window's own files. The grammar of a host source in Content Security Policy Level 3 takes host names and IPv4 addresses, made of letters, digits, hyphens and dots, but no IPv6 address: `https://[2001:db8::7]` is not a source that a policy can hold. The Windows of a Cahut at an IPv6 address could load nothing from their Cahut.

## Decision

- **A Cahut without a domain takes the public IPv4 address of its server.** `add_domain`, `rename_domain` and `install.sh --ip` refuse an IPv6 address, and say why.
- Everything else in ADR 26 holds: short-lived certificates, the certificate chosen by the local address of a connection, the move to a domain, the detection of the address by `install.sh`.
- A server that has only an IPv6 address needs a domain, with an AAAA record.

## Consequences

- The fallback covers the servers that have a public IPv4 address, which most VPS have.
- IPv6 can come back if the policy of a Window stops naming its Cahut, for instance with a scheme source or with `'self'` once browsers agree on what it matches in a sandbox.

## Alternatives considered

- **`https:` as the source of the Cahut** in the policy of a Window at an IPv6 address. It would let the Window load from any secure origin, which the declaration of a Window is there to prevent.
- **IPv6 without the house style.** A Window that loads nothing from its Cahut works only if it holds all its files, which the templates and the deck engine do not.
