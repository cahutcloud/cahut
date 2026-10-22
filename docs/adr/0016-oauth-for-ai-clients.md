# 16. AI clients connect over OAuth, to the Rooms and the scope a person grants

- Status: Accepted
- Date: 2026-10-22

## Context

[ADR 6](0006-in-house-oauth-server.md) decided that each Cahut runs its own minimal OAuth authorization server, and [ADR 11](0011-console-on-its-own-port.md) put it on the console's port, with the issuer `https://<domain>:8443`. Batch 3 of the [V0 plan](../plan-v0.md) builds it. Until then, an AI connects with a personal token that a person pastes into its client, and the token acts with every right of its person ([ADR 13](0013-people-sessions-and-rights.md)).

Clients such as claude.ai and ChatGPT cannot take a pasted token: they discover the authorization server from the MCP endpoint, register, and send the person to a consent screen. The V0 plan asks that the person choose there the Rooms and the scope they give the AI, and find every connected AI in the console, with a way to cut it off.

This record settles what ADR 6 left open: the addresses, the scopes, how Rooms are granted, the lifetimes of codes and tokens, and how consent goes through the registry ([ADR 1](0001-operation-registry.md)).

## Decision

### Addresses

| What | Where |
| --- | --- |
| Protected resource | `https://<domain>/mcp` |
| Its metadata (RFC 9728) | `https://<domain>/.well-known/oauth-protected-resource/mcp`, and at `/.well-known/oauth-protected-resource` |
| Issuer | `https://<domain>:8443` |
| Its metadata (RFC 8414) | `https://<domain>:8443/.well-known/oauth-authorization-server` |
| Authorization | `https://<domain>:8443/oauth/authorize`, a page of the console |
| Token, registration, revocation | `/oauth/token`, `/oauth/register`, `/oauth/revoke` on port 8443 |

- A request to `/mcp` without a valid token gets 401 with `WWW-Authenticate: Bearer resource_metadata="…"`.
- The metadata, token, registration and revocation endpoints answer CORS requests from any origin: they read no cookie and no session, so a page gains nothing it could not get from a server. The API of the console keeps refusing CORS ([ADR 14](0014-console-http-api.md)).
- For development, the issuer is the console's address on `localhost:8081` and the resource is on `localhost:8080`, in plain HTTP, which OAuth 2.1 allows on loopback.

### What a person grants

- A consent creates a **connection**: an AI client, the person it acts for, a scope and Rooms. The console calls connections connected AIs.
- Two scopes: `read` allows the operations that change nothing, and `write` allows everything the person may do. The person grants at most what the client asked for, and either when it asked for nothing. The MCP endpoint offers an AI with `read` only the tools that read, and the registry refuses it the others.
- The person grants every Room of theirs, including those they join later, or a list of them. An AI limited to a list sees no other Room: the shared helper of Rooms answers that a Room outside the list does not exist, the lists of Rooms, Windows and Logbook entries leave the others out, and such an AI cannot create a Room.
- The person's role and state still bound every call ([ADR 13](0013-people-sessions-and-rights.md)): a connection never grants more than its person may do, and a disabled person's connections are suspended with them.
- A personal token keeps every right of its person, as before.

### Codes and tokens

- An authorization code lives 5 minutes and is used once. It is bound to the client, the redirect URI, the PKCE challenge, which must use S256, and the resource. Presenting a used code ends its connection.
- Access tokens start with `cahut_oat_` and live one hour. Refresh tokens start with `cahut_ort_` and live 30 days; each use replaces the refresh token with a new one. Both are opaque random values from `crypto/rand`, stored as SHA-256 hashes.
- A refresh token presented after it was replaced ends its connection, unless it was replaced less than 30 seconds earlier: then the request is refused and the connection stays, so that two refreshes sent at once do not cut off a legitimate client.
- The resource is optional in a request, and means `https://<domain>/mcp` when absent. Any other resource is refused: a token of one Cahut opens nothing in another.
- A connection ends when its person or an administrator disconnects it, when its client revokes its refresh token (RFC 7009), when a code or a refresh token is replayed, or after 30 days without a refresh. An ended connection's tokens stop working at once.
- Responses to an authorization request carry `iss` (RFC 9207).

### Clients

- **Client ID Metadata Documents.** A client ID that is an `https` URL with a path is fetched from that URL. The fetch reaches public addresses only, checks the address it connects to rather than the one it resolved, follows no redirect, and stops after 5 seconds or 16 KiB. The document must name the same client ID, its redirect URIs and no client secret. It is kept in memory for one hour.
- **Dynamic registration** (RFC 7591) is open, as MCP clients need, and limited by address. It ignores unknown fields, registers public clients only, whatever method the client asks for, and returns a random client ID. A client that has received no connection 7 days after it registered is deleted.
- **Redirect URIs** must match one of the client's exactly, except loopback addresses in `http` (`127.0.0.1`, `[::1]`, `localhost`), whose port may change (RFC 8252). The schemes allowed are `https`, `http` on loopback, and the private schemes of applications; `javascript`, `data`, `file`, `blob`, `about` and `vbscript` are refused.

### Consent goes through the registry

- The consent screen is a page of the console. It signs the person in if needed, then calls three operations, kept for people (`PeopleOnly`): `describe_authorization` checks the request and describes the client, the host it will return to, the scope it asks for and the Rooms to choose from; `approve_authorization` creates the connection and its code; `deny_authorization` sends the client back with `access_denied`. Each returns the address to send the browser to.
- A request whose client or redirect URI is invalid gets an error on the screen and never goes back to the client. Any other invalid request goes back to the client with an error, as OAuth requires.
- The screen shows the client's name, and the host of its redirect URI, since anyone can register a client under any name.
- The screen sends no cookie, so there is no cross-site request to forge: the operations need the session token that the console keeps. The console's `frame-ancestors 'none'` prevents clickjacking.
- `list_connected_ais` and `disconnect_ai` are for people too: each person for their own connections, administrators for everyone's, as personal tokens are.
- The token, registration and revocation endpoints are protocol, not operations. A revocation by a client, and the end of a connection for a replayed code or token, run an operation of the registry all the same, so that they reach the Logbook.

### Logbook

- Consents, refusals and disconnections are logged under the person concerned.
- Every call of a connected AI records its connection, in a new `connection_id` column next to `token_id`, and the client's name with its software.
- Exchanges of codes and tokens are not logged: an AI refreshes its token every hour.

## Consequences

- An AI client connects to a Cahut with its address alone. The person consents once, in the console, and sees the AI there afterwards.
- A person can give an AI one Room and read access only, which a personal token cannot do.
- The registry learns the reach of an AI: the actor carries its scope and Rooms, and the operations that list or open Rooms read them. A new operation that touches a Room must go through the shared helper, as it already must for the rights of [ADR 13](0013-people-sessions-and-rights.md).
- Widening the reach of a connected AI is not possible yet: a person disconnects it and connects it again. It comes with Knock first in batch 5.
- Registration stays open to anyone who reaches port 8443, which is how MCP clients work. What it creates is harmless until a person consents, and unused clients are purged.
- No signing key to keep or rotate: tokens are looked up, as sessions are.

## Alternatives considered

- **Scopes per operation**, such as `windows:write`. Precise, but no person would read a list of twenty operations on a consent screen, and clients do not ask for them.
- **Rooms as scopes**, such as `room:marketing`. Clients would have to know the Rooms before they connect; the person knows them, on the consent screen.
- **The consent screen served by the Go server, as an HTML form.** One page fewer for the console, but a second sign-in, a cookie to carry the session, and a form to protect against forgery.
- **Long-lived access tokens without refresh.** Simpler for clients, but a stolen token would hold until someone noticed. One hour bounds it, and a disconnection ends it at once anyway.
- **Logging every token exchange.** Complete, but an hourly refresh per AI would bury the actions that matter.
