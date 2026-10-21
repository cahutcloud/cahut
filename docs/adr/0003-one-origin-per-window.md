# 3. Each Window gets its own web origin

- Status: Superseded by [ADR 8](0008-several-cahuts-per-server.md), [ADR 9](0009-rooms-and-windows-as-paths.md) and [ADR 10](0010-windows-in-a-csp-sandbox.md)
- Date: 2026-10-21

## Context

A Window holds HTML and scripts written by an AI, which nobody reviews line by line. If a Window ran on the console's origin, its scripts could read the console's pages and call the API with the reader's session. Browsers isolate pages by origin: scheme, host and port.

Two ways to isolate a Window:

- **A separate origin**, here a subdomain per Window. With a wildcard DNS record, CertMagic obtains a certificate for each subdomain at the first TLS handshake, without DNS provider access ([ADR 2](0002-built-in-https-with-certmagic.md)).
- **The CSP `sandbox` directive**, which gives the page an opaque origin. `localStorage`, `sessionStorage`, IndexedDB and `document.cookie` then throw `SecurityError`, requests carry `Origin: null`, and forms and popups are blocked unless allowed. Many dashboards and small apps made by AIs keep their state in `localStorage`, so they would break.

## Decision

Each Window is served at its own subdomain of the home, for example `q3-review.home.example.com`. The console, the API and the MCP endpoint stay at `home.example.com`.

- The home needs a wildcard DNS record, `*.home.example.com`. The guided first start asks for it.
- A Window's certificate is requested when the Window is published, so a failure shows up at once and the AI can report it. On-demand issuance at the first handshake remains as a fallback.
- The address of a Window stays the same across its versions.
- The console's cookies are host-only, with the `__Host-` prefix, and Cahut reads no other cookie. A Window can neither read them nor plant a cookie that the console would accept.
- Windows and the console share a registrable domain, so they are same-site, and `SameSite` cookies do not stop a Window from sending requests to the API with the reader's cookies. The API therefore rejects cross-origin requests that change state, based on `Sec-Fetch-Site` and `Origin` (`http.CrossOriginProtection` in the standard library), and grants no CORS access to Windows.
- **Without a domain**, the home falls back to an address under nip.io, such as `q3-review.203-0-113-7.nip.io`, which resolves to the IP written in the name. Windows keep their own origins. This fallback is for trying Cahut, not for production use: it depends on a third-party DNS service, and every user of nip.io and sslip.io shares one registered domain for Let's Encrypt's rate limits, since neither is on the Public Suffix List.

## Consequences

- Windows are isolated from the console and from each other, and AI-made pages keep storage, cookies and requests to their own origin.
- Each new Window means a new certificate. Let's Encrypt allows 50 new certificates per registered domain every 7 days, so a home that publishes more than about 50 new Windows a week hits the limit. Updates do not count, since a Window keeps its address. Options, to settle when the limit is measured in batch 1:
  - a wildcard certificate through the DNS-01 challenge, when the admin gives access to a supported DNS provider;
  - a second ACME issuer;
  - telling the AI, at publish time, when the next certificate will be available.
- The first start asks the admin for one more DNS record. The nip.io fallback keeps the path without a domain open for a trial.

## Alternatives considered

- **The console's origin with a CSP sandbox for every Window.** No wildcard needed, but AI-made pages lose storage, cookies and same-origin requests. It remains the last resort for a home reachable only by IP address, now that Let's Encrypt issues IP address certificates, which last 6 days. Not planned for the V0.
- **One origin per Room.** Fewer certificates, but a Window could read the other Windows of its Room with the reader's access.
- **A separate registrable domain for Windows**, as GitHub does with githubusercontent.com. Windows become cross-site from the console, so `SameSite` cookies protect the console as well. The strongest option, but it asks the admin for a second domain. It can become an option later.
