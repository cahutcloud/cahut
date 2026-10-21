---
title: "Operations"
description: "The operations of a Cahut and of its server: the only way to act on them."
sidebar:
  order: 3
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Every action on a Cahut goes through a registry of operations. The MCP tools and the command line are thin layers over it: the registry checks who calls, runs the operation and writes the Logbook. Each Cahut has its own registry and its own Logbook, and so does the server that holds the Cahuts.

## Operations of a Cahut

| Operation | What it does | Changes data | Offered to AIs |
| --- | --- | --- | --- |
| `create_token` | Create a personal token that lets an AI act on behalf of a person. | yes | no, administrator only |
| `create_user` | Create a person, with their personal Room. | yes | no, administrator only |
| [`get_window`](/reference/mcp-tools/#get_window) | Describe a Window and list its files, optionally with their contents. | no | yes, as an MCP tool |
| `list_logbook` | Read the latest entries of the Logbook, newest first. | no | no, administrator only |
| `list_tokens` | List personal tokens, newest first. | no | no, administrator only |
| [`list_windows`](/reference/mcp-tools/#list_windows) | List the Windows in the Rooms of the person you act for, most recently updated first. | no | yes, as an MCP tool |
| [`publish_window`](/reference/mcp-tools/#publish_window) | Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. | yes | yes, as an MCP tool |
| `revoke_token` | Revoke a personal token. The AI that used it loses access at once. | yes | no, administrator only |

## Operations of the server

| Operation | What it does | Changes data | Offered to AIs |
| --- | --- | --- | --- |
| `add_domain` | Declare a domain on this server, which creates the Cahut at that domain. | yes | no, administrator only |
| `list_domains` | List the Cahuts of this server and their domains. | no | no, administrator only |
| `list_logbook` | Read the latest entries of the Logbook, newest first. | no | no, administrator only |

## Knock first

An operation that cannot be undone needs the approval of a person before an AI can run it. No operation is irreversible yet, and approvals are not available yet.
