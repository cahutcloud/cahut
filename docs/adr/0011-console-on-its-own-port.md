# 11. The console has its own port

- Status: Accepted, superseded in part by [ADR 13](0013-people-sessions-and-rights.md) (server administrators sign in to each Cahut) and [ADR 17](0017-oauth-endpoints-on-port-443.md) (the OAuth endpoints that clients' servers call are on port 443)
- Date: 2026-10-21

## Context

The console was to sit at the main origin of the home ([ADR 5](0005-preact-console-embedded.md)). Windows now take the paths of the Cahut's domain on port 443 ([ADR 9](0009-rooms-and-windows-as-paths.md)), and one server holds several Cahuts ([ADR 8](0008-several-cahuts-per-server.md)). Browsers treat another port as another origin, but send the same cookies to every port of a host.

## Decision

- The console, its HTTP API and the OAuth authorization server ([ADR 6](0006-in-house-oauth-server.md)) listen on port 8443 of every Cahut's domain, with the domain's certificate. The port can be changed.
- There is one console for every domain. What it shows depends on the domain and on who signs in. The people of a Cahut sign in at their Cahut's domain only, and see only that Cahut; the sign-in page names no other domain. A server administrator can sign in at any domain and reaches every Cahut.
- Port 443 keeps what readers and AIs need: Windows, Room pages, Key pages, the reader's sign-in, and the MCP endpoint at `https://<domain>/mcp`. AI clients reach MCP on the usual port.
- Each Cahut is its own OAuth authorization server, with the issuer `https://<domain>:8443`. The protected resource metadata of its MCP endpoint, on port 443, names that issuer. A token is bound to the MCP endpoint of one Cahut (RFC 8707).
- The console uses no cookie. After sign-in, it keeps a session token in the storage of its own origin and sends it in the `Authorization` header. The OAuth consent screen is a page of the console and uses the same token. The API reads no cookie and grants no CORS access.
- A reader who must sign in to open a Window on port 443 goes through the console's sign-in. The console then hands a single-use code to `https://<domain>/-/session`, which sets the reader's session cookie ([ADR 10](0010-windows-in-a-csp-sandbox.md)).

## Consequences

- Windows can read neither the console's pages nor the API's answers, and a cookie set on the domain gives them nothing, since the console reads none. Cross-site request forgery has no cookie to ride on.
- A script injected into the console could read its token: the console runs under a strict CSP, with no inline script.
- The installer opens port 8443 in the server's firewall, but the hosting provider's firewall must allow it too: one more step at the first start. Some company networks only let port 443 out: the console is out of reach there, the Windows are not.
- Port 8443 is open to the people of every Cahut, so it cannot be restricted to a few addresses. Sign-in is rate limited, and two-factor authentication comes later.
- This record replaces the line of ADR 5 that served the console at the main origin, and the server of [ADR 2](0002-built-in-https-with-certmagic.md) listens on ports 80, 443 and 8443.

## Alternatives considered

- **The console under a path of port 443**, such as `/-/console`. It would share its origin with the Windows, and only the sandbox would protect it.
- **A subdomain for the console**, such as `console.<domain>`. A second name and a second certificate per Cahut, against one domain per Cahut.
- **Session cookies with the `__Host-` prefix.** Cookies ignore ports, so a page on port 443 could plant one that the console would receive.
- **MCP on port 8443 as well.** Port 443 would only serve readers, but every AI client would need a port other than the usual one.
