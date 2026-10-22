---
title: Connect an AI client
description: Give an AI client the address of your Cahut, sign in, and choose the Rooms it reaches and whether it may publish. For claude.ai, Claude Desktop, ChatGPT and other MCP clients with OAuth.
---

An AI client connects to a Cahut with its address alone: `https://cahut.example.com/mcp`. The client sends you to the console of the Cahut, where you sign in and choose what it may do. Then it acts for you, within what you granted, and every call it makes goes to the Logbook.

This works with the MCP clients that sign in with OAuth: claude.ai, Claude Desktop, Claude Code, ChatGPT and others. For a client that cannot, such as a script, use a [personal token](/connect/tokens/).

## Add the Cahut to your client

The address to give is on the **AIs and access** screen of [the console](/admin/console/), ready to copy.

| Client | Where |
| --- | --- |
| claude.ai, Claude Desktop | Settings, Connectors, Add custom connector: the address of the Cahut, nothing else |
| Claude Code | `claude mcp add --transport http cahut https://cahut.example.com/mcp`, then `/mcp` and Authenticate. See [Connect Claude Code](/connect/claude-code/) |
| ChatGPT | In developer mode, add a connector with the address of the Cahut, and OAuth as its authentication |
| Others | The address of the Cahut. The client finds the rest by itself |

The client opens a page of the console in your browser, on port 8443 of the Cahut's domain. That port must be open in the firewall of your hosting provider: see [Install on a server](/install/server/). The servers of the client reach the Cahut on port 443 only.

## Consent

If you are not signed in to the console, you sign in first. Then the console shows:

- **Who asks**: the name of the AI client, as it gives it, and the host the browser goes back to once you answer, such as `claude.ai`. Some clients, such as Claude Code, describe themselves in a document on their own site, and the console names that site. A client that registered by itself could have chosen any name: connect it only if you started from it.
- **What it may do**: read and publish, which is everything you can do in your Rooms, or read only, which is your Rooms, their Windows and the Logbook. A client that asks for read access cannot get more.
- **Where**: all your Rooms, including those you join later, or a list of them.

**Connect** sends the browser back to the client, which receives its tokens. **Refuse** sends it back with nothing.

An AI never gets more than its person: if you leave a Room, or an administrator disables you, it loses that too. An AI limited to some Rooms does not see the others, cannot publish in them and cannot create a Room. An AI that may only read sees only the tools that read.

## Connected AIs

**AIs and access** lists your connected AIs: the client, what it may do, its Rooms, when it last acted and its state. **Disconnect** cuts one off: its next call is refused. To change what an AI may do, disconnect it and connect it again.

| State | What it means |
| --- | --- |
| Active | The AI can act |
| Expired | The AI has not been used for 30 days: connect it again |
| Disconnected | You, an administrator, or the client itself ended the connection |

An administrator of the Cahut sees the connected AIs of others too, and can disconnect them.

Each connection is in the Logbook: your consent, every call the AI makes, under its name and the person it acts for, such as `Claude Code (claude-code 2.1.307) for you`, and the end of the connection.

## For the developers of clients

Each Cahut is its own OAuth 2.1 authorization server, as the [MCP authorization specification](https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization) asks:

| What | Where |
| --- | --- |
| Protected resource metadata (RFC 9728) | `https://cahut.example.com/.well-known/oauth-protected-resource/mcp` |
| Issuer | `https://cahut.example.com` |
| Authorization server metadata (RFC 8414) | `https://cahut.example.com/.well-known/oauth-authorization-server` |
| Token, registration, revocation | `https://cahut.example.com/oauth/token`, `/oauth/register`, `/oauth/revoke` |
| Authorization | `https://cahut.example.com:8443/oauth/authorize`, a page of the console |

The servers of a client need port 443 only. The person's browser opens the authorization page, on the console's port.

- Clients register with a Client ID Metadata Document or by dynamic registration (RFC 7591). Both make public clients, which authenticate with `none`: there is no client secret. A metadata document may prefer another method, such as `private_key_jwt` for ChatGPT, as long as it lists `none` in `token_endpoint_auth_methods_supported`.
- Authorization codes need PKCE with S256, live 5 minutes and are used once.
- Access tokens start with `cahut_oat_` and live one hour. Refresh tokens start with `cahut_ort_` and live 30 days; each refresh replaces the refresh token. A refresh token used twice ends the connection.
- The scopes are `read` and `write`, and `offline_access` for the refresh token, which every connection has. A token granted `write` states `read write offline_access`, and one granted `read` states `read offline_access`. The resource is `https://cahut.example.com/mcp`; a token of one Cahut opens nothing in another.
- Redirect URIs match exactly, except loopback addresses in `http`, whose port may change.

[ADR 16](https://github.com/getcahut/cahut/blob/main/docs/adr/0016-oauth-for-ai-clients.md) and [ADR 17](https://github.com/getcahut/cahut/blob/main/docs/adr/0017-oauth-endpoints-on-port-443.md) give the details. The service logs each request to these endpoints, with its status and the error of a refused one: `journalctl -u cahut` shows them while a client connects.
