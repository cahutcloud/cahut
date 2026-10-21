# 6. Cahut runs its own minimal OAuth authorization server

- Status: Accepted
- Date: 2026-10-21

## Context

AI clients authorize with OAuth before they call Cahut over MCP. Cahut is both the authorization server and the resource server, in one binary, with SQLite, public clients only, and a single protected resource: the home's MCP endpoint.

The [MCP authorization specification](https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization) (revision 2026-07-28) requires:

- OAuth 2.1, protected resource metadata (RFC 9728) and authorization server metadata (RFC 8414);
- PKCE with S256, advertised in the metadata;
- resource indicators (RFC 8707), with the audience checked on every token;
- exact redirect URI matching, and refresh token rotation for public clients.

It recommends Client ID Metadata Documents (CIMD) for client registration. Dynamic Client Registration (DCR, RFC 7591) is still allowed but deprecated. The `iss` response parameter (RFC 9207) is recommended.

The clients behave differently:

| Client | Registration | Notes |
| --- | --- | --- |
| claude.ai, Claude Desktop | CIMD when the server advertises it, otherwise DCR | CIMD needs `none` among the token endpoint auth methods |
| Claude Code | CIMD, DCR or a given client ID | Loopback redirect on any port |
| ChatGPT | CIMD preferred, DCR | S256 must be advertised; `iss` must match exactly |
| Le Chat | DCR | Unknown registration fields must be ignored; CIMD support not verified |

The Go libraries, as of October 2026:

| Library | State | What Cahut would miss |
| --- | --- | --- |
| ory/fosite | Last release v0.49.0 in December 2024; the code now lives inside ory/hydra | DCR, CIMD, RFC 8707 |
| zitadel/oidc (`op`) | Active; an OpenID provider first | DCR, RFC 8707 |
| luikyv/go-oidc | v0.25, one maintainer, OpenID certified | CIMD |
| go-oauth2/oauth2 | Active, small | Metadata, revocation endpoint, DCR, CIMD, RFC 8707 |
| MCP Go SDK | v1.8, active | Resource server side only: bearer token middleware and RFC 9728 metadata, no authorization server |

## Decision

Cahut implements its own authorization server, limited to what MCP clients use:

- authorization server metadata (RFC 8414);
- client registration by CIMD and by DCR;
- authorization code with PKCE S256 only;
- refresh token rotation with reuse detection;
- resource binding (RFC 8707), revocation (RFC 7009) and `iss` in responses (RFC 9207).

It serves public clients only. There is no implicit grant, no password grant, no client credentials and no OpenID Connect.

- The resource server side uses the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk): its bearer token middleware and its protected resource metadata handler.
- Tokens are opaque random values from `crypto/rand`, stored hashed in SQLite. No JWT, so a revocation takes effect at once.
- Granting and revoking access are operations of the registry ([ADR 1](0001-operation-registry.md)). Each consent and each revocation goes to the Logbook.
- The server is tested against the [MCP conformance suite](https://github.com/modelcontextprotocol/conformance) and against the real clients: Claude and ChatGPT in batch 3 of the [V0 plan](../plan-v0.md), the second client in batch 8.

## Consequences

- About 1,500 to 2,500 lines of Go plus their tests, which we write and maintain. This code is security critical; the security review of batch 8 covers it first.
- The known pitfalls each get tests:
  - authorization codes are single use, bound to the client, the redirect URI, the PKCE verifier and the resource, and redeemed in one transaction;
  - a reused refresh token revokes its whole family, without breaking legitimate concurrent refreshes;
  - CIMD fetching is guarded against SSRF, DNS rebinding, redirects and oversized responses;
  - DCR is rate limited, ignores unknown fields, and unused clients are purged;
  - the consent screen resists CSRF and clickjacking, and shows the redirect host;
  - one exact `iss` value everywhere, the loopback exception for terminal clients, CORS on the metadata endpoints.
- The consent screen and the list of connected AIs fit the product directly: the person picks the Rooms and the scope, and revokes an AI in one click.
- We follow revisions of the MCP specification ourselves.

## Alternatives considered

- **ory/fosite.** A mature design, but frozen outside Hydra, and missing DCR, CIMD and RFC 8707.
- **luikyv/go-oidc.** The most complete library, but without CIMD, at version 0.x with a single maintainer. Worth another look if it gains CIMD.
- **zitadel/oidc.** An OpenID provider, larger than the need, and still missing DCR and RFC 8707.
- **An identity provider next to Cahut** (Keycloak, Authentik, Ory Hydra). A second service to install and update, against one machine and one command.
