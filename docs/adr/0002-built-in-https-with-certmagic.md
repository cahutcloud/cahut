# 2. The binary serves HTTPS itself, with CertMagic

- Status: Accepted, superseded in part by [ADR 8](0008-several-cahuts-per-server.md) (one certificate per Cahut) and [ADR 11](0011-console-on-its-own-port.md) (port 8443)
- Date: 2026-10-21

## Context

Cahut serves the console, its API, the MCP endpoint and every Window over HTTPS, with automatic certificates and one subdomain per Window ([ADR 3](0003-one-origin-per-window.md)). It installs with one command and updates itself from GitHub releases, with rollback.

The brief left two options open: Caddy next to Cahut, or a web server built into the binary.

## Decision

The `cahut` binary listens on ports 80 and 443 itself, with `net/http` and [CertMagic](https://github.com/caddyserver/certmagic), the certificate library inside Caddy.

- Certificates come from Let's Encrypt over ACME, with the TLS-ALPN-01 and HTTP-01 challenges. No access to a DNS provider is needed. Port 80 answers HTTP-01 challenges and redirects everything else to HTTPS.
- Window certificates are obtained on demand. The on-demand decision function accepts only the names the home serves. It is always set, because CertMagic accepts every name when neither a decision function nor an allowlist is set.
- Certificates and ACME account keys live in Cahut's data directory, through CertMagic's file storage, not in the service user's home directory.
- CertMagic renews certificates, with ACME Renewal Information (RFC 9773). This matters as Let's Encrypt shortens certificate lifetimes to 64 days in February 2027 and 45 days in February 2028.

## Consequences

- One binary to install, update and roll back, with no second configuration to keep in sync.
- Cahut owns ports 80 and 443 on its machine, and no other web server can share them. This fits one machine per instance. A mode behind an existing reverse proxy can come later if testers ask for it.
- The service needs `CAP_NET_BIND_SERVICE` rather than root (batch 7 of the [V0 plan](../plan-v0.md)).
- CertMagic follows Caddy's release cycle. Dependabot proposes its updates.
- Let's Encrypt rate limits apply to the home's domain: 50 new certificates per registered domain every 7 days. Renewals through ARI do not count. [ADR 3](0003-one-origin-per-window.md) covers what this means for Windows.

## Alternatives considered

- **Caddy next to Cahut**, configured through its admin API and its on-demand `ask` endpoint. Proven, but two programs to install, update, roll back and debug, with a configuration to keep in sync.
- **Caddy embedded as a Go library.** Brings Caddy's module system and configuration model for features Cahut does not use.
- **nginx or another proxy with certbot.** No certificates on demand per subdomain, and more moving parts.
- **`golang.org/x/crypto/acme/autocert`.** Smaller, but less complete than CertMagic for on-demand issuance, renewal and storage.
