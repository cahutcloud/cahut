---
title: "Operations"
description: "The operations of a Cahut and of its server: the only way to act on them."
sidebar:
  order: 3
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Every action on a Cahut goes through a registry of operations. The MCP tools and the command line are thin layers over it: the registry checks who calls, runs the operation and writes the Logbook. Each Cahut has its own registry and its own Logbook, and so does the server that holds the Cahuts.

Each operation states who may call it: any person of the Cahut, its administrators, or the administrators of the server, who use the command line. An AI acts with the rights of the person it acts for.

## Operations of a Cahut

| Operation | What it does | Who calls it | Changes data | Offered to AIs |
| --- | --- | --- | --- | --- |
| `activate_account` | Choose a password from a single-use activation link, and sign in. | Anyone, to sign in | yes | no, people only |
| [`add_room_member`](/reference/mcp-tools/#add_room_member) | Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. | Any person | yes | yes |
| `approve_authorization` | Connect an AI client to this Cahut, for the person who consents, with the scope and the Rooms they grant. Returns the address that sends the browser back to the client, with its authorization code. | Any person | yes | no, people only |
| `change_password` | Change your password. Your other sessions end. | Any person | yes | no, people only |
| `create_activation_link` | Give a person a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. Their password works until they use the link. | Administrators of the Cahut | yes | no, people only |
| [`create_room`](/reference/mcp-tools/#create_room) | Create a team Room, at &lt;domain&gt;/&lt;name&gt;/. You become its owner, and can then add members. Windows published in it are for its members. | Any person | yes | yes |
| `create_token` | Create a personal token that lets an AI act on behalf of a person. | Any person | yes | no, people only |
| [`create_user`](/reference/mcp-tools/#create_user) | Create a person, with their personal Room, and the link where they choose their password. | Administrators of the Cahut | yes | to the AIs of administrators |
| `deny_authorization` | Refuse the request of an AI client to connect to this Cahut. Returns the address that tells the client so. | Any person | yes | no, people only |
| `describe_authorization` | Check the request of an AI client to connect to this Cahut, and describe it to the person who consents: the client, where it will be sent back, the scope it asks for, and the Rooms the person may grant. | Any person | no | no, people only |
| [`disable_user`](/reference/mcp-tools/#disable_user) | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. | Administrators of the Cahut | yes | to the AIs of administrators |
| `disconnect_ai` | Disconnect an AI connected over OAuth. Its tokens stop working at once; to use it again, the person connects it again. | Any person | yes | no, people only |
| [`enable_user`](/reference/mcp-tools/#enable_user) | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. | Administrators of the Cahut | yes | to the AIs of administrators |
| [`get_account`](/reference/mcp-tools/#get_account) | Describe the person you act for: their handle, name, role in the Cahut and Rooms. | Any person | no | yes |
| [`get_cahut`](/reference/mcp-tools/#get_cahut) | Describe this Cahut: its domain, the address where its Rooms and Windows open, and that of its MCP endpoint. | Any person | no | yes |
| [`get_room`](/reference/mcp-tools/#get_room) | Describe a Room: its title, its owners and members, and how many Windows it holds. | Any person | no | yes |
| [`get_window`](/reference/mcp-tools/#get_window) | Describe a Window and list its files, optionally with their contents. | Any person | no | yes |
| `list_connected_ais` | List the AIs connected over OAuth, newest first, with the person each acts for, its scope and its Rooms. | Any person | no | no, people only |
| [`list_logbook`](/reference/mcp-tools/#list_logbook) | Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. | Any person | no | yes |
| [`list_rooms`](/reference/mcp-tools/#list_rooms) | List your Rooms, your personal Room first, with your role and how many members and Windows each has. | Any person | no | yes |
| `list_tokens` | List personal tokens, newest first. | Any person | no | no, people only |
| [`list_users`](/reference/mcp-tools/#list_users) | List the people of the Cahut, with their handles, names, roles and states. | Any person | no | yes |
| [`list_windows`](/reference/mcp-tools/#list_windows) | List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first. | Any person | no | yes |
| [`publish_window`](/reference/mcp-tools/#publish_window) | Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. | Any person | yes | yes |
| [`remove_room_member`](/reference/mcp-tools/#remove_room_member) | Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. | Any person | yes | yes |
| `revoke_token` | Revoke a personal token. The AI that used it loses access at once. | Any person | yes | no, people only |
| [`set_room_member_role`](/reference/mcp-tools/#set_room_member_role) | Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. | Any person | yes | yes |
| [`set_user_role`](/reference/mcp-tools/#set_user_role) | Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. | Administrators of the Cahut | yes | to the AIs of administrators |
| `sign_in` | Sign in with a handle and a password, and open a session. | Anyone, to sign in | yes | no, people only |
| `sign_out` | End the session you signed in with. | Any person | yes | no, people only |
| [`update_room`](/reference/mcp-tools/#update_room) | Change the title of a Room. Its name, and so its address, stays. | Any person | yes | yes |

## Operations of the server

They are not offered over MCP.

| Operation | What it does | Who calls it | Changes data | Offered to AIs |
| --- | --- | --- | --- | --- |
| `activate_account` | Choose a password from a single-use activation link, as an administrator of the server, and sign in. | Anyone, to sign in | yes | no, people only |
| `add_domain` | Declare a domain on this server, which creates the Cahut at that domain, optionally with its first administrator and the link where they choose their password. | Administrators of the server | yes | no |
| `change_password` | Change your password, as an administrator of the server. Your other sessions end. | Administrators of the server | yes | no, people only |
| `create_activation_link` | Give an administrator of the server a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. Their password works until they use the link. | Administrators of the server | yes | no, people only |
| `create_admin` | Create an administrator of the server, and the link where they choose their password. | Administrators of the server | yes | no |
| `list_admins` | List the administrators of the server. | Administrators of the server | no | no |
| `list_domains` | List the Cahuts of this server and their domains. | Administrators of the server | no | no |
| `list_logbook` | Read the latest entries of the server's Logbook, newest first. | Administrators of the server | no | no |
| `sign_in` | Sign in as an administrator of the server, with a handle and a password, and open a session at this domain. | Anyone, to sign in | yes | no, people only |
| `sign_out` | End the session you signed in with, as an administrator of the server. | Administrators of the server | yes | no, people only |

## Knock first

An operation that cannot be undone, or that changes what a person may do, needs the approval of a person before an AI can run it. A person who calls it in the console acts at once. Approvals are not available yet, so AIs cannot run these operations: `add_room_member`, `create_user`, `disable_user`, `enable_user`, `remove_room_member`, `set_room_member_role`, `set_user_role`.
