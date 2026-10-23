# 17. The OAuth endpoints that clients' servers call are on port 443

- Status: Accepted, superseded in part by [ADR 20](0020-console-under-a-path-of-port-443.md) (the authorization endpoint is under `/console/` on port 443)
- Date: 2026-10-22

## Context

[ADR 11](0011-console-on-its-own-port.md) put the console and the OAuth authorization server on port 8443 of each Cahut's domain, and [ADR 16](0016-oauth-for-ai-clients.md) made the issuer `https://<domain>:8443`, with its metadata, token, registration and revocation endpoints on that port.

Two kinds of callers reach these endpoints:

- a person's browser opens the authorization endpoint, the consent screen, which is a page of the console and reads the console's session;
- the servers of an AI client fetch the metadata, register, exchange codes and refresh tokens, on their own, without a browser.

When the staging Cahut was added as a custom connector in claude.ai, the connection failed with "Couldn't register with … sign-in service". A capture on the server showed that Anthropic's servers opened connections to port 443 only, and never to port 8443: they read the protected resource metadata, then could not reach the issuer it named. The same holds wherever a client runs behind an egress proxy or a company firewall that lets port 443 out and nothing else, which ADR 11 already noted for people.

## Decision

- The issuer of each Cahut is the address of the Cahut: `https://<domain>`.
- Its metadata (RFC 8414), and its token, registration and revocation endpoints, are on port 443: `https://<domain>/.well-known/oauth-authorization-server`, `https://<domain>/oauth/token`, `/oauth/register` and `/oauth/revoke`. They answer CORS requests from any origin, read no cookie and no session, as ADR 16 settles.
- The authorization endpoint stays a page of the console: `https://<domain>:8443/oauth/authorize`. Only the person's browser opens it, and the console's session lives at that origin.
- `oauth` becomes a reserved name, which no Room takes, since `/oauth/` is now a path of the Cahut on port 443.
- `/.well-known/openid-configuration` answers 404 in JSON on both ports: Cahut is no OpenID provider.
- The protocol's endpoints log each request, with its status, the client's address and user agent and the error of a refused request, without a query or a token, so that an administrator can follow a connection that fails out of sight.

This record replaces, in ADR 11, the OAuth authorization server on port 8443 and its issuer `https://<domain>:8443`, and in ADR 16, the rows of the table of addresses for the issuer, its metadata, and the token, registration and revocation endpoints.

## Consequences

- An AI client whose servers reach the MCP endpoint reach the authorization server too: one port to open for them, the usual one.
- The person still needs port 8443 for the consent screen, as for the console.
- The token endpoint shares its origin with the Windows. Windows run in a sandbox with an opaque origin ([ADR 10](0010-windows-in-a-csp-sandbox.md)), and the endpoints read no cookie: a Window gains nothing it could not get from any other site.
- Tokens issued with the issuer of ADR 16 stay valid, since tokens are looked up, not signed. Clients connected before this change read the new metadata at their next connection.

## Alternatives considered

- **Everything on port 8443, and asking vendors to reach it.** Out of our hands, and company firewalls would still block it.
- **The consent screen on port 443 too.** It needs the person's session, which lives at the console's origin by ADR 11; a second sign-in on port 443 would need a cookie there, next to the Windows.
- **A second issuer on port 443 next to the one on 8443.** Clients use the first authorization server that the resource names: one issuer is enough, and two would only differ in where they are.
