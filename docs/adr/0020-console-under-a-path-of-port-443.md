# 20. The console is served under `/console/` on port 443

- Status: Accepted, extended by [ADR 43](0043-the-console-holds-two-sessions.md) (the console holds two sessions, that of a person and that of an administrator of the server)
- Date: 2026-10-23

## Context

[ADR 11](0011-console-on-its-own-port.md) put the console, its HTTP API and the consent screen of OAuth on port 8443 of each Cahut's domain, so that they would have an origin apart from the Windows. It set aside a path of port 443, because only the sandbox of the Windows would then protect the console.

Since then, more people need port 8443 than its administrators:

- every person who uses their console, to see their Rooms, their Windows and the AIs they connected;
- every person who connects an AI, since the consent screen is a page of the console ([ADR 17](0017-oauth-endpoints-on-port-443.md));
- every reader of a protected Window who must sign in, since a reader signs in through the console ([ADR 19](0019-window-versions-access-and-declaration.md)). Windows open to their Room by default: most readers sign in.

Port 8443 fails them where port 443 works. Company networks, hotel and school Wi-Fi, and egress proxies often let port 443 out and nothing else; the servers of claude.ai did not reach it either (ADR 17). Hosting providers keep their own firewall in front of the server, where port 8443 must be opened by hand at the first start. And an address with `:8443` looks odd, or suspicious, on the very page that asks for a password.

Meanwhile, [ADR 10](0010-windows-in-a-csp-sandbox.md) and ADR 19 made the sandbox the rule for every file of a Window: each one is served with the CSP directive `sandbox` without `allow-same-origin`, so its page has an opaque origin, and the pages of Cahut on port 443 run only Cahut's code under `default-src 'none'`.

## Decision

- The console is served on port 443 of each Cahut's domain, under the path `/console/`: `https://<domain>/console/`. Its screens take the paths that followed the port, such as `/console/logbook` and `/console/rooms/olivier`.
- Its HTTP API moves with it: `POST https://<domain>/console/api/<operation>`, and `/console/api/server/<operation>` for the operations of the server ([ADR 14](0014-console-http-api.md)).
- The authorization endpoint of OAuth stays a page of the console: `https://<domain>/console/oauth/authorize`. The issuer, its metadata and its other endpoints stay where ADR 17 put them.
- The activation links, the sign-in of readers (`/-/sign-in` sends the browser to `/console/read`) and every address that the command line prints name the console under `/console/`.
- `console` and `admin` are already reserved names, which no Room and no handle can take. `/admin` and every path under it redirect to `/console/`.
- The server no longer listens on port 8443. The installer opens ports 80 and 443 only. A server for development serves the console under `/console/` of its single address.
- What ADR 11 settled otherwise holds: one console for every domain, whose content depends on the domain and on who signs in; no cookie, and a session token kept in the storage of the origin and sent in the `Authorization` header; an API that reads no cookie and grants no CORS access; a strict CSP with no inline script and `frame-ancestors 'none'`.
- Every response on port 443 carries the sandbox of ADR 10, except those of Cahut's own handlers, named one by one: the console, the pages of Cahut at the root, at the address of a Room and under `/-/`, the endpoints of MCP and OAuth, and the metadata under `/.well-known/`. A path that no handler names is served sandboxed. A test walks the routes and fails when a response that serves a file of a Window, or an error, lacks the sandbox.

This record replaces, in ADR 11, the port 8443 and the alternative of a path on port 443; in ADR 2, the port 8443 among the ports that the server listens on; in ADR 14, the address of the API; and in ADR 16 and ADR 17, the address of the authorization endpoint.

## Consequences

- A person needs port 443 only, for the console as for the Windows, and an administrator opens no port beyond 80 and 443. The step of the hosting provider's firewall, and the risk the V0 plan names for company networks, go away.
- The console shares its origin with the Windows and with the pages of Cahut on port 443. Windows still cannot reach it: their opaque origin reads none of the domain's storage, their requests carry `Origin: null` and get no CORS access from the API, and they never hold the console's token. A cookie set on the domain gives them nothing, since the console reads none.
- The origin no longer stands between the console and the Windows: the sandbox alone does. A response of a Window served without it would expose the console's token to that Window's scripts. Hence the sandbox by default, and the test.
- A flaw in any page of Cahut on port 443, such as a script injected into the page of a Room, would now reach the console's token too. Those pages stay rendered on the server, with no inline script and under `default-src 'none'`.
- The reader's session cookie of ADR 19 is sent with the requests of the console, which ignores it. Signing in to the console still does not open Windows, and a reader session still does not open the console.
- Links printed before this record name port 8443 and stop working. No Cahut is in use yet beyond the staging server.

## Alternatives considered

- **Keeping port 8443.** An origin of its own, but out of reach on the networks that let port 443 out only, and a firewall to open at every hosting provider.
- **A subdomain for the console**, such as `console.<domain>`. A Cahut is often already at a subdomain, such as `staging.cahut.cloud`: the console would need a name one level deeper, a second DNS record and a second certificate for each Cahut, and a wildcard certificate of the parent domain would not cover it.
- **The path `/admin/`.** The word names a role in Cahut, the administrators of a Cahut and of the server, while the console serves every person and every reader who signs in. It redirects to `/console/` for those who type it.
- **The path `/-/console`.** It keeps every page of Cahut under `/-/`, but makes an address that people type and share harder to read. `/-/` stays for the addresses that Cahut follows on its own.
- **A cookie for the console's session, now that it shares the origin.** It would bring back cross-site request forgery, which a token in a header avoids.
