---
title: "MCP tools"
description: "The tools that a Cahut offers to AIs over MCP, with their arguments, results and JSON Schemas."
sidebar:
  order: 1
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Each Cahut serves MCP at `https://<domain>/mcp`, over Streamable HTTP. An AI connects with a personal token of that Cahut, which it sends in the header `Authorization: Bearer <token>`: see [Connect an AI](/connect/claude-code/).

A Cahut speaks the versions of the protocol from 2024-11-05 to 2026-07-28. A client of 2025-11-25 or earlier opens a session with `initialize`. A client of 2026-07-28, such as ChatGPT, calls `server/discover` and opens no session: each request names the version in the header `Mcp-Protocol-Version`.

The tools are the operations of the Cahut that are open to AIs ([Operations](/reference/operations/)). An AI sees the tools that the person it acts for may use: some are for the AIs of administrators of the Cahut only. The operations that prove who someone is, such as signing in, and those that serve the console are for people only. Every call goes through the registry of operations, which checks it, runs it and writes it to the Logbook, reads included. Arguments with unknown fields are refused.

A tool returns its result as structured content, and as the same JSON in a text block. When a call fails because of its arguments, the error says what to change. Any other error reads only as an internal error: the administrator of the Cahut finds the details in its Logbook.

## Instructions

The Cahut gives these instructions to the AI when it connects:

> Cahut keeps what you make on the user's own server, as Windows. To publish a document, deck, dashboard or site: call publish_window with a complete index.html and all its files, linked by relative paths (css/deck.css, never /css/deck.css), and a declaration of its scripts and of every origin it loads from or calls. Give the user the address it returns. A call that cannot be undone makes a Knock: give the user its link to approve it.
>
> Make each Window in the style the user asks for, with its own CSS, fonts and scripts among its files: Cahut serves no style sheet for Windows, and a Window is served under the address of its Room. Pages can load and call nothing that their declaration does not list, and send forms only where it says. They run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: keep state in memory or in the address. To change a Window, publish its new version with update_window at the same address; rollback_window goes back to the previous one. Each Window has an id that never changes: name the Window by its id to find it again, and give its permanent_url for a link that lasts. Give each Window a description of a sentence or two, and a few tags, reusing those of the Windows that list_windows gives; retitle_window changes its title, description and tags without a new version. list_windows finds the Windows with a tag, and with query those whose titles, descriptions, tags or pages hold some words: search before you publish anew a Window that may be there already. A Window opens to the members of its Room until set_window_access opens it to the whole Cahut or to anyone with its Key; making it public waits for a person. publish_window with access author keeps a Window for the person you act for alone, such as a draft in a team Room; set_window_access to room shares it with the Room. invite_to_window sends its address to people by e-mail, with a personal link for a Window with a Key. create_share_link gives a link that opens a Window to whoever holds it, without an account or a Key, for 30 days at most: give it to the user, who passes it on; revoke_share_link takes it back. For a Window that one reader opens once, publish it with access author, then make a share link with max_openings 1. Knock first: a call that cannot be undone, or that changes what a person may do, does not run at once; once the person has answered its Knock, get_knock gives the outcome. Every call is recorded in the Logbook of the Cahut.

## Limits

| Limit | Value |
| --- | --- |
| Files in one Window | 500 |
| Size of the files of one Window, decoded | 32 MiB |
| Size of one request to `/mcp` | about 43.7 MiB, room for the files in base64 |
| Requests larger than 4 MiB served at once | 2, the others wait their turn |
| Sessions opened with one token | 20 at once, then one every 3 seconds |
| Space left free on the disk of the server | 1024 MiB, below which a publication is refused |

## Tools

| Tool | What it does |
| --- | --- |
| [`add_room_member`](#add_room_member) | Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. Knock first: a person approves the call before it runs. |
| [`create_room`](#create_room) | Create a team Room, at &lt;domain&gt;/&lt;name&gt;/. You become its owner, and can then add members. Windows published in it are for its members. In an individual Cahut, you stay its only member, and its name starts with your handle and a hyphen. |
| [`create_share_link`](#create_share_link) | Make a share link: whoever holds it opens the Window, without an account or a Key, until it ends, besides its access. It ends in 7 days unless you give an end, 30 days at most, and may open to a number of readers only. It is returned this once: give it to the person. revoke_share_link takes it back. Not for a public Window. |
| [`create_user`](#create_user) | Create a person, with their personal Room, and the link where they choose their password, which goes to their e-mail address when you give one and the Cahut sends e-mail. Knock first: a person approves the call before it runs. |
| [`delete_room`](#delete_room) | Delete a team Room for good, with its members and its Windows, those in its trash included, and their files: this cannot be undone. Its address becomes free. Its owners and the administrators of the Cahut delete it; a personal Room is never deleted. Knock first: a person approves the call before it runs. |
| [`delete_user`](#delete_user) | Delete a person for good: their account, their tokens and the AIs they connected, their personal Room and the team Rooms they are the only owner of, with the Windows of these Rooms and their files. This cannot be undone. The Windows they published in other Rooms stay, without an author. Their handle becomes free. No one deletes themself, nor the last administrator of the Cahut; disable_user stops a person and keeps what is theirs. Knock first: a person approves the call before it runs. |
| [`disable_user`](#disable_user) | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. Knock first: a person approves the call before it runs. |
| [`enable_user`](#enable_user) | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. Knock first: a person approves the call before it runs. |
| [`get_account`](#get_account) | Describe the person you act for: their handle, name, role in the Cahut and Rooms. |
| [`get_cahut`](#get_cahut) | Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, the links to its legal pages, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused. |
| [`get_knock`](#get_knock) | Check a Knock of this Cahut, a call of an AI that waits for a person to approve it: whether it waits, was approved and run, failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked. |
| [`get_room`](#get_room) | Describe a Room of this Cahut: its title, its owners and members, and how many Windows it holds. |
| [`get_window`](#get_window) | Describe a Window of this Cahut and list its files, optionally with their contents. |
| [`invite_to_window`](#invite_to_window) | Send the address of a Window by e-mail, with a note: to people of the Cahut, named by handle, who sign in to read it, or to anyone else, named by address, when the Window opens with its Key or to the public. For a Window with a Key, each address gets a personal link that opens it for them alone, until revoke_invitation takes it back; the Key itself never goes. 20 people a call, and 100 invitations a day. get_cahut says whether the Cahut sends e-mail. |
| [`list_invitations`](#list_invitations) | List the invitations to a Window of this Cahut, the latest first: who got one, who invited, and whether it still stands. |
| [`list_knocks`](#list_knocks) | List the Knocks of this Cahut, those that wait first. A Knock is a call of an AI that waits for a person to approve it, since it cannot be undone or changes what a person may do. An AI lists the Knocks it made; a person, those of their AIs. |
| [`list_logbook`](#list_logbook) | Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. |
| [`list_rooms`](#list_rooms) | List your Rooms in this Cahut, your personal Room first, with your role and how many members and Windows each has. |
| [`list_users`](#list_users) | List the people of the Cahut, with their handles, names, roles and states. In an individual Cahut, for its administrators only. |
| [`list_windows`](#list_windows) | List the Windows in the Rooms of the person you act for in this Cahut, or in one of them, or those with a tag, most recently updated first; with query, those whose title, description, tags or pages hold its words, the best first. |
| [`publish_window`](#publish_window) | Publish a new Window in this Cahut: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. |
| [`purge_window`](#purge_window) | Delete a Window of this Cahut that is in the trash for good, with its versions and its files: this cannot be undone. Its address becomes free. Put the Window in the trash first, with trash_window. Knock first: a person approves the call before it runs. |
| [`remove_room_member`](#remove_room_member) | Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. Knock first: a person approves the call before it runs. |
| [`rename_window`](#rename_window) | Change the name of a Window of this Cahut, the last part of its address, &lt;domain&gt;/&lt;room&gt;/&lt;name&gt;/. Its title, its id and its permanent address stay. The old address stops working at once, and its name is free in the Room: rename_window back to it undoes the change, while no other Window takes it. |
| [`restore_window`](#restore_window) | Bring a Window of this Cahut back from the trash, with its versions. |
| [`retitle_window`](#retitle_window) | Change the title of a Window of this Cahut, its description, its tags, or several, without publishing a new version. Its address stays. Give the old ones again to undo it. |
| [`revoke_invitation`](#revoke_invitation) | Take back an invitation to a Window of this Cahut: its personal link, and the passes it gave, no longer open the Window. The e-mail stays sent. |
| [`revoke_share_link`](#revoke_share_link) | Take back a share link to a Window of this Cahut: it no longer opens the Window, to those who opened it too. |
| [`rollback_window`](#rollback_window) | Go back to the previous version of a Window of this Cahut. The version you leave becomes the previous one, so calling rollback_window again returns to it. |
| [`set_ai_reach`](#set_ai_reach) | Set what an AI connected over OAuth reaches in this Cahut: read or write, and every Room of its person or some of them. A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, and widening it, such as asking for write or for another Room, waits for its person to approve. Knock first when it widens the reach of an AI: a person approves the call before it runs. |
| [`set_room_home`](#set_room_home) | Show a Window of a Room of this Cahut at the Room's address, &lt;domain&gt;/&lt;room&gt;/, in place of the page that lists its Windows, such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. An empty window goes back to the list. The Window keeps its own address and its access. |
| [`set_room_member_role`](#set_room_member_role) | Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. Knock first: a person approves the call before it runs. |
| [`set_user_email`](#set_user_email) | Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. Their old address gets a notice. For the administrators of the Cahut. Knock first: a person approves the call before it runs. |
| [`set_user_role`](#set_user_role) | Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. Knock first: a person approves the call before it runs. |
| [`set_window_access`](#set_window_access) | Set who opens a Window: author, its author alone, for themself; room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. Knock first when it makes a Window public: a person approves the call before it runs. |
| [`trash_window`](#trash_window) | Put a Window of this Cahut in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was. |
| [`update_room`](#update_room) | Change the title of a Room of this Cahut, its description, or both. Its name, and so its address, stays. Give the old title and description again to undo it. |
| [`update_window`](#update_window) | Publish a new version of a Window of this Cahut: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. Cahut calls no other system. |

## `add_room_member`

Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. Knock first: a person approves the call before it runs.

Not offered in an individual Cahut, where each Room keeps one member, its owner.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Add a member to a Room. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `user` | string | yes | handle of the person |
| `role` | string | no | owner or member; member when empty |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "role": {
      "description": "owner or member; member when empty",
      "type": "string"
    },
    "room": {
      "description": "the Room's name",
      "type": "string"
    },
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "room",
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `create_room`

Create a team Room, at &lt;domain&gt;/&lt;name&gt;/. You become its owner, and can then add members. Windows published in it are for its members. In an individual Cahut, you stay its only member, and its name starts with your handle and a hyphen.

Title: Create a team Room. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `title` | string | yes | title of the Room, such as Marketing |
| `description` | string | no | what the Room is, in a sentence or two of plain text, 300 characters at most |
| `name` | string | no | name of the Room, the first part of its address; from the title when empty |
| `owner` | string | no | handle of its first owner; you when empty. Only administrators of the Cahut name someone else |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "description": {
      "description": "what the Room is, in a sentence or two of plain text, 300 characters at most",
      "type": "string"
    },
    "name": {
      "description": "name of the Room, the first part of its address; from the title when empty",
      "type": "string"
    },
    "owner": {
      "description": "handle of its first owner; you when empty. Only administrators of the Cahut name someone else",
      "type": "string"
    },
    "title": {
      "description": "title of the Room, such as Marketing",
      "type": "string"
    }
  },
  "required": [
    "title"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at"
  ],
  "type": "object"
}
```

</details>

## `create_share_link`

Make a share link: whoever holds it opens the Window, without an account or a Key, until it ends, besides its access. It ends in 7 days unless you give an end, 30 days at most, and may open to a number of readers only. It is returned this once: give it to the person. revoke_share_link takes it back. Not for a public Window.

Title: Share a Window with a link. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: true`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `ends_at` | string or null | no | when the link stops opening the Window, within 30 days, such as 2026-11-05T18:00:00Z; 7 days from now when absent |
| `label` | string | no | a few words that tell the links of the Window apart, such as For Acme, 100 characters at most, which only the members of its Room see |
| `max_openings` | integer | no | how many readers may open it, from 1 to 1000; any number when absent |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "ends_at": {
      "description": "when the link stops opening the Window, within 30 days, such as 2026-11-05T18:00:00Z; 7 days from now when absent",
      "type": [
        "null",
        "string"
      ]
    },
    "label": {
      "description": "a few words that tell the links of the Window apart, such as For Acme, 100 characters at most, which only the members of its Room see",
      "type": "string"
    },
    "max_openings": {
      "description": "how many readers may open it, from 1 to 1000; any number when absent",
      "type": "integer"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `expires_at` | string | yes | when it stops opening the Window, to those who opened it too |
| `id` | integer | yes |  |
| `link` | string | yes | the share link, returned this once: give it to the person, who passes it on. Whoever holds it opens the Window |
| `openings` | integer | yes | how many readers opened it |
| `stands` | boolean | yes | whether it opens the Window to someone new: neither ended nor taken back, with openings left |
| `window` | string | yes | the Window, as room/name |
| `ai` | string | no | the AI that made it for them |
| `by` | string | no | the handle of the person who made it |
| `label` | string | no |  |
| `max_openings` | integer | no | how many readers may open it; absent when any number may |
| `opened_at` | string or null | no | when it was last opened |
| `revoked_at` | string or null | no | when it was taken back |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "ai": {
      "description": "the AI that made it for them",
      "type": "string"
    },
    "by": {
      "description": "the handle of the person who made it",
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "expires_at": {
      "description": "when it stops opening the Window, to those who opened it too",
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "label": {
      "type": "string"
    },
    "link": {
      "description": "the share link, returned this once: give it to the person, who passes it on. Whoever holds it opens the Window",
      "type": "string"
    },
    "max_openings": {
      "description": "how many readers may open it; absent when any number may",
      "type": "integer"
    },
    "opened_at": {
      "description": "when it was last opened",
      "type": [
        "null",
        "string"
      ]
    },
    "openings": {
      "description": "how many readers opened it",
      "type": "integer"
    },
    "revoked_at": {
      "description": "when it was taken back",
      "type": [
        "null",
        "string"
      ]
    },
    "stands": {
      "description": "whether it opens the Window to someone new: neither ended nor taken back, with openings left",
      "type": "boolean"
    },
    "window": {
      "description": "the Window, as room/name",
      "type": "string"
    }
  },
  "required": [
    "id",
    "created_at",
    "expires_at",
    "openings",
    "stands",
    "window",
    "link"
  ],
  "type": "object"
}
```

</details>

## `create_user`

Create a person, with their personal Room, and the link where they choose their password, which goes to their e-mail address when you give one and the Cahut sends e-mail. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Add a person. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `name` | string | yes |  |
| `admin` | boolean | no |  |
| `email` | string | no | the e-mail address of the person, where their activation link goes |
| `send_activation_email` | boolean or null | no | false keeps the activation link from the e-mail address of the person, when you give it to them yourself; true when absent |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "admin": {
      "type": "boolean"
    },
    "email": {
      "description": "the e-mail address of the person, where their activation link goes",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "name": {
      "type": "string"
    },
    "send_activation_email": {
      "description": "false keeps the activation link from the e-mail address of the person, when you give it to them yourself; true when absent",
      "type": [
        "null",
        "boolean"
      ]
    }
  },
  "required": [
    "handle",
    "name"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `activation` | object or null | no | the link where the person chooses their password; an AI never gets it |
| `activation.code` | string | yes | the person opens the console of the Cahut at /activate#&lt;code&gt; to choose their password |
| `activation.expires_at` | string | yes |  |
| `activation.user` | string | yes | handle of the person |
| `activation.emailed` | boolean | no | the link went to the e-mail address of the person too |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "activation": {
      "additionalProperties": false,
      "description": "the link where the person chooses their password; an AI never gets it",
      "properties": {
        "code": {
          "description": "the person opens the console of the Cahut at /activate#\u003ccode\u003e to choose their password",
          "type": "string"
        },
        "emailed": {
          "description": "the link went to the e-mail address of the person too",
          "type": "boolean"
        },
        "expires_at": {
          "type": "string"
        },
        "user": {
          "description": "handle of the person",
          "type": "string"
        }
      },
      "required": [
        "user",
        "code",
        "expires_at"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room"
  ],
  "type": "object"
}
```

</details>

## `delete_room`

Delete a team Room for good, with its members and its Windows, those in its trash included, and their files: this cannot be undone. Its address becomes free. Its owners and the administrators of the Cahut delete it; a personal Room is never deleted. Knock first: a person approves the call before it runs.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Delete a team Room. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the name of the team Room |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "the name of the team Room",
      "type": "string"
    }
  },
  "required": [
    "room"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the name of the Room, now free |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows were deleted with it, those in its trash included |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "the name of the Room, now free",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows were deleted with it, those in its trash included",
      "type": "integer"
    }
  },
  "required": [
    "room",
    "title",
    "windows"
  ],
  "type": "object"
}
```

</details>

## `delete_user`

Delete a person for good: their account, their tokens and the AIs they connected, their personal Room and the team Rooms they are the only owner of, with the Windows of these Rooms and their files. This cannot be undone. The Windows they published in other Rooms stay, without an author. Their handle becomes free. No one deletes themself, nor the last administrator of the Cahut; disable_user stops a person and keeps what is theirs. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Delete a person for good. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `user` | string | yes | handle of the person |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `name` | string | yes |  |
| `rooms` | array of strings or null | yes | the Rooms deleted with the person: their personal Room, and the team Rooms they were the only owner of |
| `user` | string | yes | the handle of the person, now free |
| `windows` | integer | yes | how many Windows were deleted with these Rooms, those in their trash included |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "name": {
      "type": "string"
    },
    "rooms": {
      "description": "the Rooms deleted with the person: their personal Room, and the team Rooms they were the only owner of",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "user": {
      "description": "the handle of the person, now free",
      "type": "string"
    },
    "windows": {
      "description": "how many Windows were deleted with these Rooms, those in their trash included",
      "type": "integer"
    }
  },
  "required": [
    "user",
    "name",
    "rooms",
    "windows"
  ],
  "type": "object"
}
```

</details>

## `disable_user`

Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Disable a person. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `user` | string | yes | handle of the person |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room"
  ],
  "type": "object"
}
```

</details>

## `enable_user`

Enable a disabled person again: they can sign in, and their AIs can act with their tokens. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Enable a person. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `user` | string | yes | handle of the person |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room"
  ],
  "type": "object"
}
```

</details>

## `get_account`

Describe the person you act for: their handle, name, role in the Cahut and Rooms.

Title: Describe your account. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

This tool takes no arguments.

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `rooms` | array of objects or null | yes | the Rooms of the person, their personal Room first |
| `rooms[].kind` | string | yes | personal or team |
| `rooms[].role` | string | yes | owner or member of the Room |
| `rooms[].room` | string | yes | the Room's name, first part of its address |
| `rooms[].title` | string | yes |  |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |
| `email_change` | object or null | no | the e-mail address that waits for the person to confirm it, from the link sent to it |
| `email_change.email` | string | yes |  |
| `email_change.expires_at` | string | yes | when the link sent to it stops working |
| `notices_off` | boolean | no | whether the e-mail notices that are neither about the account nor about a Knock are off |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "email_change": {
      "additionalProperties": false,
      "description": "the e-mail address that waits for the person to confirm it, from the link sent to it",
      "properties": {
        "email": {
          "type": "string"
        },
        "expires_at": {
          "description": "when the link sent to it stops working",
          "type": "string"
        }
      },
      "required": [
        "email",
        "expires_at"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "notices_off": {
      "description": "whether the e-mail notices that are neither about the account nor about a Knock are off",
      "type": "boolean"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "rooms": {
      "description": "the Rooms of the person, their personal Room first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "kind": {
            "description": "personal or team",
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          },
          "room": {
            "description": "the Room's name, first part of its address",
            "type": "string"
          },
          "title": {
            "type": "string"
          }
        },
        "required": [
          "room",
          "title",
          "kind",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room",
    "rooms"
  ],
  "type": "object"
}
```

</details>

## `get_cahut`

Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, the links to its legal pages, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused.

Title: Describe the Cahut. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

This tool takes no arguments.

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `domain` | string | yes |  |
| `kind` | string | yes | collaborative, where people share team Rooms, or individual, where each person works alone and does not see the others |
| `legal` | object | yes | the links to the legal pages of the Cahut, which its host publishes elsewhere, and which the console and the pages for readers show at their foot; empty when it links none |
| `legal.fr` | object or null | no | the addresses of the same pages in French, where they differ: the console links them when it speaks French. Each needs the address of its page for every other language |
| `legal.fr.legal_notice` | string | no | the address of the legal notice: who publishes the service and who hosts it |
| `legal.fr.privacy` | string | no | the address of the privacy policy |
| `legal.fr.support` | string | no | the address of the page where people get help |
| `legal.fr.terms` | string | no | the address of the terms of use |
| `legal.legal_notice` | string | no | the address of the legal notice: who publishes the service and who hosts it |
| `legal.privacy` | string | no | the address of the privacy policy |
| `legal.support` | string | no | the address of the page where people get help |
| `legal.terms` | string | no | the address of the terms of use |
| `limits` | object | yes | the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only |
| `limits.admins` | integer or null | no | how many of them may be administrators of the Cahut |
| `limits.disk` | integer or null | no | how many bytes the Cahut may keep: every version of every Window, those in the trash included |
| `limits.disk_per_person` | integer or null | no | in an individual Cahut, how many bytes the Windows of each person may keep, every version included |
| `limits.people` | integer or null | no | how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count |
| `limits.rooms` | integer or null | no | how many team Rooms the Cahut may have, or in an individual Cahut how many Rooms beside the personal ones; personal Rooms come with their person and do not count |
| `limits.rooms_per_person` | integer or null | no | in an individual Cahut, how many Rooms each person may have beside their personal Room |
| `limits.windows` | integer or null | no | how many Windows out of the trash the Cahut may hold |
| `limits.windows_per_person` | integer or null | no | in an individual Cahut, how many Windows out of the trash each person may hold, over their Rooms |
| `limits.windows_per_room` | integer or null | no | how many Windows out of the trash each Room may hold, personal Rooms included |
| `sends_mail` | boolean | yes | whether this Cahut sends e-mail, through the relay that its administrators set; without one, nothing that would send e-mail works, such as invitations |
| `mcp` | string | no | the address of its MCP endpoint, where AI clients connect |
| `moderation_email` | string | no | the address that receives each report about the Windows of the Cahut, beside its administrators, which the administrators of the server set; absent when it has none, and for anyone but its administrators |
| `suspension` | object or null | no | since when and why the administrators of the server suspended the Cahut; absent while it is active |
| `suspension.reason` | string | yes | why the administrators of the server suspended the Cahut; its administrators read it, and readers, people and AIs never do |
| `suspension.since` | string | yes | when the administrators of the server suspended the Cahut |
| `url` | string | no | the address of the Cahut, under which its Rooms and Windows open |
| `usage` | object or null | no | what the Cahut holds, as its limits count it; absent for a member of an individual Cahut |
| `usage.admins` | integer | yes | administrators who are not disabled |
| `usage.disk` | integer | yes | bytes of every version of every Window, those in the trash included |
| `usage.people` | integer | yes | people who are not disabled, administrators included |
| `usage.rooms` | integer | yes | team Rooms, or in an individual Cahut the Rooms beside the personal ones |
| `usage.windows` | integer | yes | Windows out of the trash |
| `yours` | object or null | no | in an individual Cahut, what the person you act for holds, as the limits per person count it |
| `yours.disk` | integer | yes | bytes of every version of the Windows of their Rooms, those in the trash included |
| `yours.rooms` | integer | yes | Rooms beside the personal one |
| `yours.windows` | integer | yes | Windows out of the trash, over their Rooms |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "domain": {
      "type": "string"
    },
    "kind": {
      "description": "collaborative, where people share team Rooms, or individual, where each person works alone and does not see the others",
      "type": "string"
    },
    "legal": {
      "additionalProperties": false,
      "description": "the links to the legal pages of the Cahut, which its host publishes elsewhere, and which the console and the pages for readers show at their foot; empty when it links none",
      "properties": {
        "fr": {
          "additionalProperties": false,
          "description": "the addresses of the same pages in French, where they differ: the console links them when it speaks French. Each needs the address of its page for every other language",
          "properties": {
            "legal_notice": {
              "description": "the address of the legal notice: who publishes the service and who hosts it",
              "type": "string"
            },
            "privacy": {
              "description": "the address of the privacy policy",
              "type": "string"
            },
            "support": {
              "description": "the address of the page where people get help",
              "type": "string"
            },
            "terms": {
              "description": "the address of the terms of use",
              "type": "string"
            }
          },
          "type": [
            "null",
            "object"
          ]
        },
        "legal_notice": {
          "description": "the address of the legal notice: who publishes the service and who hosts it",
          "type": "string"
        },
        "privacy": {
          "description": "the address of the privacy policy",
          "type": "string"
        },
        "support": {
          "description": "the address of the page where people get help",
          "type": "string"
        },
        "terms": {
          "description": "the address of the terms of use",
          "type": "string"
        }
      },
      "type": "object"
    },
    "limits": {
      "additionalProperties": false,
      "description": "the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only",
      "properties": {
        "admins": {
          "description": "how many of them may be administrators of the Cahut",
          "type": [
            "null",
            "integer"
          ]
        },
        "disk": {
          "description": "how many bytes the Cahut may keep: every version of every Window, those in the trash included",
          "type": [
            "null",
            "integer"
          ]
        },
        "disk_per_person": {
          "description": "in an individual Cahut, how many bytes the Windows of each person may keep, every version included",
          "type": [
            "null",
            "integer"
          ]
        },
        "people": {
          "description": "how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count",
          "type": [
            "null",
            "integer"
          ]
        },
        "rooms": {
          "description": "how many team Rooms the Cahut may have, or in an individual Cahut how many Rooms beside the personal ones; personal Rooms come with their person and do not count",
          "type": [
            "null",
            "integer"
          ]
        },
        "rooms_per_person": {
          "description": "in an individual Cahut, how many Rooms each person may have beside their personal Room",
          "type": [
            "null",
            "integer"
          ]
        },
        "windows": {
          "description": "how many Windows out of the trash the Cahut may hold",
          "type": [
            "null",
            "integer"
          ]
        },
        "windows_per_person": {
          "description": "in an individual Cahut, how many Windows out of the trash each person may hold, over their Rooms",
          "type": [
            "null",
            "integer"
          ]
        },
        "windows_per_room": {
          "description": "how many Windows out of the trash each Room may hold, personal Rooms included",
          "type": [
            "null",
            "integer"
          ]
        }
      },
      "type": "object"
    },
    "mcp": {
      "description": "the address of its MCP endpoint, where AI clients connect",
      "type": "string"
    },
    "moderation_email": {
      "description": "the address that receives each report about the Windows of the Cahut, beside its administrators, which the administrators of the server set; absent when it has none, and for anyone but its administrators",
      "type": "string"
    },
    "sends_mail": {
      "description": "whether this Cahut sends e-mail, through the relay that its administrators set; without one, nothing that would send e-mail works, such as invitations",
      "type": "boolean"
    },
    "suspension": {
      "additionalProperties": false,
      "description": "since when and why the administrators of the server suspended the Cahut; absent while it is active",
      "properties": {
        "reason": {
          "description": "why the administrators of the server suspended the Cahut; its administrators read it, and readers, people and AIs never do",
          "type": "string"
        },
        "since": {
          "description": "when the administrators of the server suspended the Cahut",
          "type": "string"
        }
      },
      "required": [
        "since",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "url": {
      "description": "the address of the Cahut, under which its Rooms and Windows open",
      "type": "string"
    },
    "usage": {
      "additionalProperties": false,
      "description": "what the Cahut holds, as its limits count it; absent for a member of an individual Cahut",
      "properties": {
        "admins": {
          "description": "administrators who are not disabled",
          "type": "integer"
        },
        "disk": {
          "description": "bytes of every version of every Window, those in the trash included",
          "type": "integer"
        },
        "people": {
          "description": "people who are not disabled, administrators included",
          "type": "integer"
        },
        "rooms": {
          "description": "team Rooms, or in an individual Cahut the Rooms beside the personal ones",
          "type": "integer"
        },
        "windows": {
          "description": "Windows out of the trash",
          "type": "integer"
        }
      },
      "required": [
        "people",
        "admins",
        "rooms",
        "windows",
        "disk"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "yours": {
      "additionalProperties": false,
      "description": "in an individual Cahut, what the person you act for holds, as the limits per person count it",
      "properties": {
        "disk": {
          "description": "bytes of every version of the Windows of their Rooms, those in the trash included",
          "type": "integer"
        },
        "rooms": {
          "description": "Rooms beside the personal one",
          "type": "integer"
        },
        "windows": {
          "description": "Windows out of the trash, over their Rooms",
          "type": "integer"
        }
      },
      "required": [
        "rooms",
        "windows",
        "disk"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "domain",
    "sends_mail",
    "kind",
    "limits",
    "legal"
  ],
  "type": "object"
}
```

</details>

## `get_knock`

Check a Knock of this Cahut, a call of an AI that waits for a person to approve it: whether it waits, was approved and run, failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked.

Title: Check an approval request (Knock). Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `knock` | integer | yes | the number of the Knock |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "knock": {
      "description": "the number of the Knock",
      "type": "integer"
    }
  },
  "required": [
    "knock"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `client` | string | yes | the AI that called, as its token or its connection, and its software, name it |
| `created_at` | string | yes |  |
| `expires_at` | string | yes | when the Knock expires, if no one answers it |
| `id` | integer | yes |  |
| `input` | object | yes | the arguments of the call |
| `operation` | string | yes | the operation that the AI called |
| `state` | string | yes | waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access) |
| `user` | string | yes | handle of the person the AI acts for, who answers the Knock |
| `answered_at` | string or null | no |  |
| `detail` | string | no |  |
| `error` | string | no | why the operation failed, once approved |
| `room` | string | no | the name of the Room the call concerns, if any |
| `target` | string | no | what the call acts on |
| `url` | string | no | where the person approves or denies it, in the console |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "answered_at": {
      "type": [
        "null",
        "string"
      ]
    },
    "client": {
      "description": "the AI that called, as its token or its connection, and its software, name it",
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "detail": {
      "type": "string"
    },
    "error": {
      "description": "why the operation failed, once approved",
      "type": "string"
    },
    "expires_at": {
      "description": "when the Knock expires, if no one answers it",
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "input": {
      "additionalProperties": true,
      "description": "the arguments of the call",
      "type": "object"
    },
    "operation": {
      "description": "the operation that the AI called",
      "type": "string"
    },
    "room": {
      "description": "the name of the Room the call concerns, if any",
      "type": "string"
    },
    "state": {
      "description": "waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access)",
      "type": "string"
    },
    "target": {
      "description": "what the call acts on",
      "type": "string"
    },
    "url": {
      "description": "where the person approves or denies it, in the console",
      "type": "string"
    },
    "user": {
      "description": "handle of the person the AI acts for, who answers the Knock",
      "type": "string"
    }
  },
  "required": [
    "id",
    "operation",
    "input",
    "user",
    "client",
    "state",
    "created_at",
    "expires_at"
  ],
  "type": "object"
}
```

</details>

## `get_room`

Describe a Room of this Cahut: its title, its owners and members, and how many Windows it holds.

Title: Describe a Room. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name, first part of its address |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    }
  },
  "required": [
    "room"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `get_window`

Describe a Window of this Cahut and list its files, optionally with their contents.

Title: Describe a Window. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `include_content` | boolean | no | also return the content of each file |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "include_content": {
      "description": "also return the content of each file",
      "type": "boolean"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `files` | array of objects or null | yes | the files of the current version |
| `files[].path` | string | yes |  |
| `files[].size` | integer | yes |  |
| `files[].type` | string | yes | media type |
| `files[].content` | string | no |  |
| `files[].encoding` | string | no | utf-8 or base64 |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `links` | array of objects or null | yes | its share links, the latest first, without the links themselves |
| `links[].created_at` | string | yes |  |
| `links[].expires_at` | string | yes | when it stops opening the Window, to those who opened it too |
| `links[].id` | integer | yes |  |
| `links[].openings` | integer | yes | how many readers opened it |
| `links[].stands` | boolean | yes | whether it opens the Window to someone new: neither ended nor taken back, with openings left |
| `links[].ai` | string | no | the AI that made it for them |
| `links[].by` | string | no | the handle of the person who made it |
| `links[].label` | string | no |  |
| `links[].max_openings` | integer | no | how many readers may open it; absent when any number may |
| `links[].opened_at` | string or null | no | when it was last opened |
| `links[].revoked_at` | string or null | no | when it was taken back |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `share_links` | integer | yes | how many share links open the Window to someone new, besides its access |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `versions` | array of objects or null | yes | the current version, then the previous one if there is one |
| `versions[].current` | boolean | yes | whether readers see this version |
| `versions[].file_count` | integer | yes |  |
| `versions[].number` | integer | yes |  |
| `versions[].published_at` | string | yes |  |
| `versions[].published_by` | string | yes | handle of the person it was published for, empty once that person is deleted |
| `versions[].size` | integer | yes | total size of its files, in bytes |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "files": {
      "description": "the files of the current version",
      "items": {
        "additionalProperties": false,
        "properties": {
          "content": {
            "type": "string"
          },
          "encoding": {
            "description": "utf-8 or base64",
            "type": "string"
          },
          "path": {
            "type": "string"
          },
          "size": {
            "type": "integer"
          },
          "type": {
            "description": "media type",
            "type": "string"
          }
        },
        "required": [
          "path",
          "size",
          "type"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "links": {
      "description": "its share links, the latest first, without the links themselves",
      "items": {
        "additionalProperties": false,
        "properties": {
          "ai": {
            "description": "the AI that made it for them",
            "type": "string"
          },
          "by": {
            "description": "the handle of the person who made it",
            "type": "string"
          },
          "created_at": {
            "type": "string"
          },
          "expires_at": {
            "description": "when it stops opening the Window, to those who opened it too",
            "type": "string"
          },
          "id": {
            "type": "integer"
          },
          "label": {
            "type": "string"
          },
          "max_openings": {
            "description": "how many readers may open it; absent when any number may",
            "type": "integer"
          },
          "opened_at": {
            "description": "when it was last opened",
            "type": [
              "null",
              "string"
            ]
          },
          "openings": {
            "description": "how many readers opened it",
            "type": "integer"
          },
          "revoked_at": {
            "description": "when it was taken back",
            "type": [
              "null",
              "string"
            ]
          },
          "stands": {
            "description": "whether it opens the Window to someone new: neither ended nor taken back, with openings left",
            "type": "boolean"
          }
        },
        "required": [
          "id",
          "created_at",
          "expires_at",
          "openings",
          "stands"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "share_links": {
      "description": "how many share links open the Window to someone new, besides its access",
      "type": "integer"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "versions": {
      "description": "the current version, then the previous one if there is one",
      "items": {
        "additionalProperties": false,
        "properties": {
          "current": {
            "description": "whether readers see this version",
            "type": "boolean"
          },
          "file_count": {
            "type": "integer"
          },
          "number": {
            "type": "integer"
          },
          "published_at": {
            "type": "string"
          },
          "published_by": {
            "description": "handle of the person it was published for, empty once that person is deleted",
            "type": "string"
          },
          "size": {
            "description": "total size of its files, in bytes",
            "type": "integer"
          }
        },
        "required": [
          "number",
          "current",
          "size",
          "file_count",
          "published_by",
          "published_at"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access",
    "versions",
    "files",
    "share_links",
    "links"
  ],
  "type": "object"
}
```

</details>

## `invite_to_window`

Send the address of a Window by e-mail, with a note: to people of the Cahut, named by handle, who sign in to read it, or to anyone else, named by address, when the Window opens with its Key or to the public. For a Window with a Key, each address gets a personal link that opens it for them alone, until revoke_invitation takes it back; the Key itself never goes. 20 people a call, and 100 invitations a day. get_cahut says whether the Cahut sends e-mail.

Title: Invite people to read a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: true`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `to` | array of strings or null | yes | who to invite, 20 at most: handles of people of the Cahut, who sign in to read it, or e-mail addresses of anyone else, for a Window with a Key or public |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `note` | string | no | a few words for them, 500 characters at most, in plain text with \*\*bold\*\*, \*italics\*, and lists whose lines start with - or 1. |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "note": {
      "description": "a few words for them, 500 characters at most, in plain text with **bold**, *italics*, and lists whose lines start with - or 1.",
      "type": "string"
    },
    "to": {
      "description": "who to invite, 20 at most: handles of people of the Cahut, who sign in to read it, or e-mail addresses of anyone else, for a Window with a Key or public",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window",
    "to"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `invitations` | array of objects or null | yes |  |
| `invitations[].created_at` | string | yes |  |
| `invitations[].id` | integer | yes |  |
| `invitations[].link` | boolean | yes | whether it carries a personal link, which opens the Window for this recipient alone |
| `invitations[].to` | string | yes | the handle of the person invited, or the address |
| `invitations[].ai` | string | no | the AI that sent the invitation for them |
| `invitations[].by` | string | no | the handle of the person who invited |
| `invitations[].ended_at` | string or null | no | when it was revoked, or ended with the Key or the access it stood for |
| `invitations[].expires_at` | string or null | no | when its personal link stops working |
| `window` | string | yes |  |
| `unreachable` | array of strings or null | no | people of the Cahut who have no e-mail address, to whom nothing went |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "invitations": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "ai": {
            "description": "the AI that sent the invitation for them",
            "type": "string"
          },
          "by": {
            "description": "the handle of the person who invited",
            "type": "string"
          },
          "created_at": {
            "type": "string"
          },
          "ended_at": {
            "description": "when it was revoked, or ended with the Key or the access it stood for",
            "type": [
              "null",
              "string"
            ]
          },
          "expires_at": {
            "description": "when its personal link stops working",
            "type": [
              "null",
              "string"
            ]
          },
          "id": {
            "type": "integer"
          },
          "link": {
            "description": "whether it carries a personal link, which opens the Window for this recipient alone",
            "type": "boolean"
          },
          "to": {
            "description": "the handle of the person invited, or the address",
            "type": "string"
          }
        },
        "required": [
          "id",
          "to",
          "link",
          "created_at"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "unreachable": {
      "description": "people of the Cahut who have no e-mail address, to whom nothing went",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "window": {
      "type": "string"
    }
  },
  "required": [
    "window",
    "invitations"
  ],
  "type": "object"
}
```

</details>

## `list_invitations`

List the invitations to a Window of this Cahut, the latest first: who got one, who invited, and whether it still stands.

Title: List the invitations to a Window. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `invitations` | array of objects or null | yes | the latest first |
| `invitations[].created_at` | string | yes |  |
| `invitations[].id` | integer | yes |  |
| `invitations[].link` | boolean | yes | whether it carries a personal link, which opens the Window for this recipient alone |
| `invitations[].to` | string | yes | the handle of the person invited, or the address |
| `invitations[].ai` | string | no | the AI that sent the invitation for them |
| `invitations[].by` | string | no | the handle of the person who invited |
| `invitations[].ended_at` | string or null | no | when it was revoked, or ended with the Key or the access it stood for |
| `invitations[].expires_at` | string or null | no | when its personal link stops working |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "invitations": {
      "description": "the latest first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "ai": {
            "description": "the AI that sent the invitation for them",
            "type": "string"
          },
          "by": {
            "description": "the handle of the person who invited",
            "type": "string"
          },
          "created_at": {
            "type": "string"
          },
          "ended_at": {
            "description": "when it was revoked, or ended with the Key or the access it stood for",
            "type": [
              "null",
              "string"
            ]
          },
          "expires_at": {
            "description": "when its personal link stops working",
            "type": [
              "null",
              "string"
            ]
          },
          "id": {
            "type": "integer"
          },
          "link": {
            "description": "whether it carries a personal link, which opens the Window for this recipient alone",
            "type": "boolean"
          },
          "to": {
            "description": "the handle of the person invited, or the address",
            "type": "string"
          }
        },
        "required": [
          "id",
          "to",
          "link",
          "created_at"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "invitations"
  ],
  "type": "object"
}
```

</details>

## `list_knocks`

List the Knocks of this Cahut, those that wait first. A Knock is a call of an AI that waits for a person to approve it, since it cannot be undone or changes what a person may do. An AI lists the Knocks it made; a person, those of their AIs.

Title: List approval requests (Knocks). Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `limit` | integer | no | how many Knocks, 50 by default, at most 500 |
| `waiting` | boolean | no | only the Knocks that wait |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "limit": {
      "description": "how many Knocks, 50 by default, at most 500",
      "type": "integer"
    },
    "waiting": {
      "description": "only the Knocks that wait",
      "type": "boolean"
    }
  },
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `knocks` | array of objects or null | yes |  |
| `knocks[].client` | string | yes | the AI that called, as its token or its connection, and its software, name it |
| `knocks[].created_at` | string | yes |  |
| `knocks[].expires_at` | string | yes | when the Knock expires, if no one answers it |
| `knocks[].id` | integer | yes |  |
| `knocks[].input` | object | yes | the arguments of the call |
| `knocks[].operation` | string | yes | the operation that the AI called |
| `knocks[].state` | string | yes | waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access) |
| `knocks[].user` | string | yes | handle of the person the AI acts for, who answers the Knock |
| `knocks[].answered_at` | string or null | no |  |
| `knocks[].detail` | string | no |  |
| `knocks[].error` | string | no | why the operation failed, once approved |
| `knocks[].room` | string | no | the name of the Room the call concerns, if any |
| `knocks[].target` | string | no | what the call acts on |
| `knocks[].url` | string | no | where the person approves or denies it, in the console |
| `waiting` | integer | yes | how many Knocks wait, in all |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "knocks": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "answered_at": {
            "type": [
              "null",
              "string"
            ]
          },
          "client": {
            "description": "the AI that called, as its token or its connection, and its software, name it",
            "type": "string"
          },
          "created_at": {
            "type": "string"
          },
          "detail": {
            "type": "string"
          },
          "error": {
            "description": "why the operation failed, once approved",
            "type": "string"
          },
          "expires_at": {
            "description": "when the Knock expires, if no one answers it",
            "type": "string"
          },
          "id": {
            "type": "integer"
          },
          "input": {
            "additionalProperties": true,
            "description": "the arguments of the call",
            "type": "object"
          },
          "operation": {
            "description": "the operation that the AI called",
            "type": "string"
          },
          "room": {
            "description": "the name of the Room the call concerns, if any",
            "type": "string"
          },
          "state": {
            "description": "waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access)",
            "type": "string"
          },
          "target": {
            "description": "what the call acts on",
            "type": "string"
          },
          "url": {
            "description": "where the person approves or denies it, in the console",
            "type": "string"
          },
          "user": {
            "description": "handle of the person the AI acts for, who answers the Knock",
            "type": "string"
          }
        },
        "required": [
          "id",
          "operation",
          "input",
          "user",
          "client",
          "state",
          "created_at",
          "expires_at"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "waiting": {
      "description": "how many Knocks wait, in all",
      "type": "integer"
    }
  },
  "required": [
    "knocks",
    "waiting"
  ],
  "type": "object"
}
```

</details>

## `list_logbook`

Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf.

Title: Read the Logbook. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `limit` | integer | no | how many entries, 50 by default, at most 1000 |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "limit": {
      "description": "how many entries, 50 by default, at most 1000",
      "type": "integer"
    }
  },
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `entries` | array of objects or null | yes |  |
| `entries[].actor` | string | yes | admin, person or ai |
| `entries[].at` | string | yes |  |
| `entries[].id` | integer | yes |  |
| `entries[].operation` | string | yes |  |
| `entries[].outcome` | string | yes | ok, knocked (waiting for a person's approval), refused or failed |
| `entries[].via` | string | yes | the adapter the call came through: cli, api, mcp, oauth, or web for the page of a share link |
| `entries[].address` | string | no | where a call that signs in came from: an IPv4 address, or the /64 of an IPv6 one |
| `entries[].client` | string | no | the AI, as its token or its connection, and its software, name it |
| `entries[].connection_id` | integer | no | the connection of the AI, when it connected over OAuth |
| `entries[].detail` | string | no |  |
| `entries[].error` | string | no |  |
| `entries[].knock_id` | integer | no | the Knock that the call made, answered or ran for |
| `entries[].target` | string | no | what the operation acted on |
| `entries[].token_id` | integer | no | the personal token the AI presented |
| `entries[].user` | string | no |  |
| `entries[].user_id` | integer | no | the person, or the person an AI acts for |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "entries": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "actor": {
            "description": "admin, person or ai",
            "type": "string"
          },
          "address": {
            "description": "where a call that signs in came from: an IPv4 address, or the /64 of an IPv6 one",
            "type": "string"
          },
          "at": {
            "type": "string"
          },
          "client": {
            "description": "the AI, as its token or its connection, and its software, name it",
            "type": "string"
          },
          "connection_id": {
            "description": "the connection of the AI, when it connected over OAuth",
            "type": "integer"
          },
          "detail": {
            "type": "string"
          },
          "error": {
            "type": "string"
          },
          "id": {
            "type": "integer"
          },
          "knock_id": {
            "description": "the Knock that the call made, answered or ran for",
            "type": "integer"
          },
          "operation": {
            "type": "string"
          },
          "outcome": {
            "description": "ok, knocked (waiting for a person's approval), refused or failed",
            "type": "string"
          },
          "target": {
            "description": "what the operation acted on",
            "type": "string"
          },
          "token_id": {
            "description": "the personal token the AI presented",
            "type": "integer"
          },
          "user": {
            "type": "string"
          },
          "user_id": {
            "description": "the person, or the person an AI acts for",
            "type": "integer"
          },
          "via": {
            "description": "the adapter the call came through: cli, api, mcp, oauth, or web for the page of a share link",
            "type": "string"
          }
        },
        "required": [
          "id",
          "at",
          "operation",
          "actor",
          "via",
          "outcome"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "entries"
  ],
  "type": "object"
}
```

</details>

## `list_rooms`

List your Rooms in this Cahut, your personal Room first, with your role and how many members and Windows each has.

Title: List Rooms. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `all` | boolean | no | every Room of the Cahut, for an administrator; your Rooms when false |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "all": {
      "description": "every Room of the Cahut, for an administrator; your Rooms when false",
      "type": "boolean"
    }
  },
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `rooms` | array of objects or null | yes |  |
| `rooms[].created_at` | string | yes |  |
| `rooms[].kind` | string | yes | personal or team |
| `rooms[].members` | integer | yes | how many people are in the Room |
| `rooms[].name` | string | yes | the Room's name, first part of its address |
| `rooms[].title` | string | yes |  |
| `rooms[].windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `rooms[].description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `rooms[].home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `rooms[].role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "rooms": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "created_at": {
            "type": "string"
          },
          "description": {
            "description": "what the Room is, in a sentence or two; absent when it has none",
            "type": "string"
          },
          "home": {
            "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
            "type": "string"
          },
          "kind": {
            "description": "personal or team",
            "type": "string"
          },
          "members": {
            "description": "how many people are in the Room",
            "type": "integer"
          },
          "name": {
            "description": "the Room's name, first part of its address",
            "type": "string"
          },
          "role": {
            "description": "your role in the Room, owner or member; empty when you are not in it",
            "type": "string"
          },
          "title": {
            "type": "string"
          },
          "windows": {
            "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
            "type": "integer"
          }
        },
        "required": [
          "name",
          "title",
          "kind",
          "members",
          "windows",
          "created_at"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "rooms"
  ],
  "type": "object"
}
```

</details>

## `list_users`

List the people of the Cahut, with their handles, names, roles and states. In an individual Cahut, for its administrators only.

In an individual Cahut, offered to the AIs of its administrators only.

Title: List people. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

This tool takes no arguments.

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `users` | array of objects or null | yes |  |
| `users[].handle` | string | yes |  |
| `users[].id` | integer | yes |  |
| `users[].name` | string | yes |  |
| `users[].personal_room` | string | yes | name of the person's personal Room, the handle |
| `users[].role` | string | yes | admin or member of the Cahut |
| `users[].status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `users[].email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "users": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "email": {
            "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
            "type": "string"
          },
          "handle": {
            "type": "string"
          },
          "id": {
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "personal_room": {
            "description": "name of the person's personal Room, the handle",
            "type": "string"
          },
          "role": {
            "description": "admin or member of the Cahut",
            "type": "string"
          },
          "status": {
            "description": "pending until the person chooses a password, then active, or disabled",
            "type": "string"
          }
        },
        "required": [
          "id",
          "handle",
          "name",
          "role",
          "status",
          "personal_room"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "users"
  ],
  "type": "object"
}
```

</details>

## `list_windows`

List the Windows in the Rooms of the person you act for in this Cahut, or in one of them, or those with a tag, most recently updated first; with query, those whose title, description, tags or pages hold its words, the best first.

Title: List and search Windows. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `query` | string | no | words to search, such as q3 sales; accents and case do not count |
| `room` | string | no | only the Windows of this Room, by its name |
| `tag` | string | no | only the Windows with this tag |
| `trashed` | boolean | no | list the Windows in the trash instead |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "query": {
      "description": "words to search, such as q3 sales; accents and case do not count",
      "type": "string"
    },
    "room": {
      "description": "only the Windows of this Room, by its name",
      "type": "string"
    },
    "tag": {
      "description": "only the Windows with this tag",
      "type": "string"
    },
    "trashed": {
      "description": "list the Windows in the trash instead",
      "type": "boolean"
    }
  },
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `windows` | array of objects or null | yes |  |
| `windows[].access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `windows[].created_at` | string | yes |  |
| `windows[].created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `windows[].file_count` | integer | yes |  |
| `windows[].id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `windows[].kind` | string | yes |  |
| `windows[].name` | string | yes | the Window's name, last part of its address |
| `windows[].permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `windows[].room` | string | yes | the name of its Room, the part of its address before its name |
| `windows[].size` | integer | yes | total size of the files of the current version, in bytes |
| `windows[].tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `windows[].title` | string | yes |  |
| `windows[].updated_at` | string | yes |  |
| `windows[].url` | string | yes | where the Window opens |
| `windows[].version` | integer | yes | the number of the version that readers see |
| `windows[].access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `windows[].declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `windows[].declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `windows[].declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `windows[].declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `windows[].description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `windows[].excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `windows[].previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `windows[].trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `windows[].withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `windows[].withdrawn.at` | string | yes |  |
| `windows[].withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `windows[].withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "windows": {
      "items": {
        "additionalProperties": false,
        "properties": {
          "access": {
            "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
            "type": "string"
          },
          "access_ends_at": {
            "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
            "type": [
              "null",
              "string"
            ]
          },
          "created_at": {
            "type": "string"
          },
          "created_by": {
            "description": "handle of the person it was first published for, empty once that person is deleted",
            "type": "string"
          },
          "declaration": {
            "additionalProperties": false,
            "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
            "properties": {
              "forms": {
                "description": "the origins that the forms of the pages send to, if any",
                "items": {
                  "type": "string"
                },
                "type": [
                  "null",
                  "array"
                ]
              },
              "origins": {
                "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
                "items": {
                  "type": "string"
                },
                "type": [
                  "null",
                  "array"
                ]
              },
              "scripts": {
                "description": "whether the pages run scripts, inline or in files",
                "type": "boolean"
              }
            },
            "required": [
              "scripts"
            ],
            "type": [
              "null",
              "object"
            ]
          },
          "description": {
            "description": "what the Window is, in a sentence or two; absent when it has none",
            "type": "string"
          },
          "excerpt": {
            "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
            "type": "string"
          },
          "file_count": {
            "type": "integer"
          },
          "id": {
            "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
            "type": "string"
          },
          "kind": {
            "type": "string"
          },
          "name": {
            "description": "the Window's name, last part of its address",
            "type": "string"
          },
          "permanent_url": {
            "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
            "type": "string"
          },
          "previous_version": {
            "description": "the number of the version that rollback_window goes back to; absent when there is none",
            "type": "integer"
          },
          "room": {
            "description": "the name of its Room, the part of its address before its name",
            "type": "string"
          },
          "size": {
            "description": "total size of the files of the current version, in bytes",
            "type": "integer"
          },
          "tags": {
            "description": "its tags, in order, which help find it among many",
            "items": {
              "type": "string"
            },
            "type": [
              "null",
              "array"
            ]
          },
          "title": {
            "type": "string"
          },
          "trashed_at": {
            "description": "when the Window went to the trash; absent when it is not in it",
            "type": [
              "null",
              "string"
            ]
          },
          "updated_at": {
            "type": "string"
          },
          "url": {
            "description": "where the Window opens",
            "type": "string"
          },
          "version": {
            "description": "the number of the version that readers see",
            "type": "integer"
          },
          "withdrawn": {
            "additionalProperties": false,
            "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
            "properties": {
              "at": {
                "type": "string"
              },
              "by_server": {
                "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
                "type": "boolean"
              },
              "reason": {
                "description": "why the moderators withdrew it, for the members of its Room",
                "type": "string"
              }
            },
            "required": [
              "at",
              "reason"
            ],
            "type": [
              "null",
              "object"
            ]
          }
        },
        "required": [
          "id",
          "name",
          "title",
          "tags",
          "kind",
          "url",
          "permanent_url",
          "room",
          "version",
          "size",
          "file_count",
          "created_by",
          "created_at",
          "updated_at",
          "access"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "windows"
  ],
  "type": "object"
}
```

</details>

## `publish_window`

Publish a new Window in this Cahut: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens.

Title: Publish a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `declaration` | object or null | yes | what the pages do: Cahut lets them load and call the origins listed here, and nothing else |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `files` | array of objects or null | yes | the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css |
| `files[].content` | string | yes |  |
| `files[].path` | string | yes | relative path with / as separator, such as index.html or img/logo.svg |
| `files[].encoding` | string | no | utf-8 (the default) or base64, for binary files |
| `kind` | string | yes | one of document, deck, dashboard or site |
| `title` | string | yes | title of the Window, also used in its address |
| `access` | string | no | room, the members of its Room (the default), or author, the person you act for alone: a draft, or a Window that one reader opens once with create_share_link and max_openings 1 |
| `description` | string | no | what the Window is, in a sentence or two of plain text, 300 characters at most: readers see it on the page of its Room, and AIs read it to find the Window |
| `name` | string | no | its name in its address, 1 to 63 lowercase letters, digits and hyphens; from the title when empty |
| `room` | string | no | name of the Room to publish in; the personal Room of the person you act for when empty |
| `tags` | array of strings or null | no | 10 at most, such as q3 or sales, which help find it among many: reuse those that list_windows gives |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "room, the members of its Room (the default), or author, the person you act for alone: a draft, or a Window that one reader opens once with create_share_link and max_openings 1",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages do: Cahut lets them load and call the origins listed here, and nothing else",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two of plain text, 300 characters at most: readers see it on the page of its Room, and AIs read it to find the Window",
      "type": "string"
    },
    "files": {
      "description": "the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css",
      "items": {
        "additionalProperties": false,
        "properties": {
          "content": {
            "type": "string"
          },
          "encoding": {
            "description": "utf-8 (the default) or base64, for binary files",
            "type": "string"
          },
          "path": {
            "description": "relative path with / as separator, such as index.html or img/logo.svg",
            "type": "string"
          }
        },
        "required": [
          "path",
          "content"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "kind": {
      "description": "one of document, deck, dashboard or site",
      "type": "string"
    },
    "name": {
      "description": "its name in its address, 1 to 63 lowercase letters, digits and hyphens; from the title when empty",
      "type": "string"
    },
    "room": {
      "description": "name of the Room to publish in; the personal Room of the person you act for when empty",
      "type": "string"
    },
    "tags": {
      "description": "10 at most, such as q3 or sales, which help find it among many: reuse those that list_windows gives",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "description": "title of the Window, also used in its address",
      "type": "string"
    }
  },
  "required": [
    "title",
    "kind",
    "files",
    "declaration"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `purge_window`

Delete a Window of this Cahut that is in the trash for good, with its versions and its files: this cannot be undone. Its address becomes free. Put the Window in the trash first, with trash_window. Knock first: a person approves the call before it runs.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Delete a Window for good. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `remove_room_member`

Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. Knock first: a person approves the call before it runs.

Not offered in an individual Cahut, where each Room keeps one member, its owner.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Remove a member from a Room. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `user` | string | yes | handle of the person; yourself to leave the Room |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "the Room's name",
      "type": "string"
    },
    "user": {
      "description": "handle of the person; yourself to leave the Room",
      "type": "string"
    }
  },
  "required": [
    "room",
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `rename_window`

Change the name of a Window of this Cahut, the last part of its address, &lt;domain&gt;/&lt;room&gt;/&lt;name&gt;/. Its title, its id and its permanent address stay. The old address stops working at once, and its name is free in the Room: rename_window back to it undoes the change, while no other Window takes it.

Title: Rename a Window in its address. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | yes | the new name, 1 to 63 lowercase letters, digits and hyphens |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "name": {
      "description": "the new name, 1 to 63 lowercase letters, digits and hyphens",
      "type": "string"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window",
    "name"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `restore_window`

Bring a Window of this Cahut back from the trash, with its versions.

Title: Restore a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `retitle_window`

Change the title of a Window of this Cahut, its description, its tags, or several, without publishing a new version. Its address stays. Give the old ones again to undo it.

Title: Retitle, describe and tag a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `description` | string or null | no | the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent |
| `tags` | array of strings or null | no | the new tags, in place of the old ones, 10 at most; an empty list removes them, and they stay when absent |
| `title` | string | no | the new title; the title stays when empty. The address does not change |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "description": {
      "description": "the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent",
      "type": [
        "null",
        "string"
      ]
    },
    "tags": {
      "description": "the new tags, in place of the old ones, 10 at most; an empty list removes them, and they stay when absent",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "description": "the new title; the title stays when empty. The address does not change",
      "type": "string"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `revoke_invitation`

Take back an invitation to a Window of this Cahut: its personal link, and the passes it gave, no longer open the Window. The e-mail stays sent.

Title: Take back an invitation. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `invitation` | integer | yes | the id of the invitation, which list_invitations gives |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "invitation": {
      "description": "the id of the invitation, which list_invitations gives",
      "type": "integer"
    }
  },
  "required": [
    "invitation"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `id` | integer | yes |  |
| `link` | boolean | yes | whether it carries a personal link, which opens the Window for this recipient alone |
| `to` | string | yes | the handle of the person invited, or the address |
| `ai` | string | no | the AI that sent the invitation for them |
| `by` | string | no | the handle of the person who invited |
| `ended_at` | string or null | no | when it was revoked, or ended with the Key or the access it stood for |
| `expires_at` | string or null | no | when its personal link stops working |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "ai": {
      "description": "the AI that sent the invitation for them",
      "type": "string"
    },
    "by": {
      "description": "the handle of the person who invited",
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "ended_at": {
      "description": "when it was revoked, or ended with the Key or the access it stood for",
      "type": [
        "null",
        "string"
      ]
    },
    "expires_at": {
      "description": "when its personal link stops working",
      "type": [
        "null",
        "string"
      ]
    },
    "id": {
      "type": "integer"
    },
    "link": {
      "description": "whether it carries a personal link, which opens the Window for this recipient alone",
      "type": "boolean"
    },
    "to": {
      "description": "the handle of the person invited, or the address",
      "type": "string"
    }
  },
  "required": [
    "id",
    "to",
    "link",
    "created_at"
  ],
  "type": "object"
}
```

</details>

## `revoke_share_link`

Take back a share link to a Window of this Cahut: it no longer opens the Window, to those who opened it too.

Title: Take back a share link. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `link` | integer | yes | the id of the share link, which get_window gives |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "link": {
      "description": "the id of the share link, which get_window gives",
      "type": "integer"
    }
  },
  "required": [
    "link"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `expires_at` | string | yes | when it stops opening the Window, to those who opened it too |
| `id` | integer | yes |  |
| `openings` | integer | yes | how many readers opened it |
| `stands` | boolean | yes | whether it opens the Window to someone new: neither ended nor taken back, with openings left |
| `ai` | string | no | the AI that made it for them |
| `by` | string | no | the handle of the person who made it |
| `label` | string | no |  |
| `max_openings` | integer | no | how many readers may open it; absent when any number may |
| `opened_at` | string or null | no | when it was last opened |
| `revoked_at` | string or null | no | when it was taken back |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "ai": {
      "description": "the AI that made it for them",
      "type": "string"
    },
    "by": {
      "description": "the handle of the person who made it",
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "expires_at": {
      "description": "when it stops opening the Window, to those who opened it too",
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "label": {
      "type": "string"
    },
    "max_openings": {
      "description": "how many readers may open it; absent when any number may",
      "type": "integer"
    },
    "opened_at": {
      "description": "when it was last opened",
      "type": [
        "null",
        "string"
      ]
    },
    "openings": {
      "description": "how many readers opened it",
      "type": "integer"
    },
    "revoked_at": {
      "description": "when it was taken back",
      "type": [
        "null",
        "string"
      ]
    },
    "stands": {
      "description": "whether it opens the Window to someone new: neither ended nor taken back, with openings left",
      "type": "boolean"
    }
  },
  "required": [
    "id",
    "created_at",
    "expires_at",
    "openings",
    "stands"
  ],
  "type": "object"
}
```

</details>

## `rollback_window`

Go back to the previous version of a Window of this Cahut. The version you leave becomes the previous one, so calling rollback_window again returns to it.

Title: Roll back a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `set_ai_reach`

Set what an AI connected over OAuth reaches in this Cahut: read or write, and every Room of its person or some of them. A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, and widening it, such as asking for write or for another Room, waits for its person to approve. Knock first when it widens the reach of an AI: a person approves the call before it runs.

Offered also to the AIs that may only read, though it changes things.

Knock first when it widens the reach of an AI: that call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Change what an AI reaches. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `scope` | string | yes | read, or write |
| `all_rooms` | boolean | no | every Room of the person, including those they join later |
| `id` | integer | no | the connection, from list_connected_ais; an AI leaves it out, to set its own |
| `rooms` | array of strings or null | no | the names of the Rooms the AI reaches, when all_rooms is false |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "all_rooms": {
      "description": "every Room of the person, including those they join later",
      "type": "boolean"
    },
    "id": {
      "description": "the connection, from list_connected_ais; an AI leaves it out, to set its own",
      "type": "integer"
    },
    "rooms": {
      "description": "the names of the Rooms the AI reaches, when all_rooms is false",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "scope": {
      "description": "read, or write",
      "type": "string"
    }
  },
  "required": [
    "scope"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `all_rooms` | boolean | yes | every Room of the person, including those they join later |
| `client` | string | yes | the name of the AI client |
| `client_id` | string | yes |  |
| `created_at` | string | yes |  |
| `id` | integer | yes |  |
| `redirect_host` | string | yes | where its authorization went back to |
| `rooms` | array of strings or null | yes | the Rooms granted, when all_rooms is false |
| `scope` | string | yes | read, or write |
| `status` | string | yes | active, expired after 30 days without use, or ended |
| `user` | string | yes | the person the AI acts for |
| `end_reason` | string | no |  |
| `ended_at` | string or null | no |  |
| `last_used_at` | string or null | no |  |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "all_rooms": {
      "description": "every Room of the person, including those they join later",
      "type": "boolean"
    },
    "client": {
      "description": "the name of the AI client",
      "type": "string"
    },
    "client_id": {
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "end_reason": {
      "type": "string"
    },
    "ended_at": {
      "type": [
        "null",
        "string"
      ]
    },
    "id": {
      "type": "integer"
    },
    "last_used_at": {
      "type": [
        "null",
        "string"
      ]
    },
    "redirect_host": {
      "description": "where its authorization went back to",
      "type": "string"
    },
    "rooms": {
      "description": "the Rooms granted, when all_rooms is false",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "scope": {
      "description": "read, or write",
      "type": "string"
    },
    "status": {
      "description": "active, expired after 30 days without use, or ended",
      "type": "string"
    },
    "user": {
      "description": "the person the AI acts for",
      "type": "string"
    }
  },
  "required": [
    "id",
    "user",
    "client",
    "client_id",
    "redirect_host",
    "scope",
    "all_rooms",
    "rooms",
    "status",
    "created_at"
  ],
  "type": "object"
}
```

</details>

## `set_room_home`

Show a Window of a Room of this Cahut at the Room's address, &lt;domain&gt;/&lt;room&gt;/, in place of the page that lists its Windows, such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. An empty window goes back to the list. The Window keeps its own address and its access.

Title: Choose the home page of a Room. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `window` | string | no | name of a Window of the Room, shown at the Room's address; empty to go back to the page that lists the Room's Windows |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "the Room's name",
      "type": "string"
    },
    "window": {
      "description": "name of a Window of the Room, shown at the Room's address; empty to go back to the page that lists the Room's Windows",
      "type": "string"
    }
  },
  "required": [
    "room"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `set_room_member_role`

Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. Knock first: a person approves the call before it runs.

Not offered in an individual Cahut, where each Room keeps one member, its owner.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Change a role in a Room. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `user` | string | yes | handle of the person |
| `role` | string | no | owner or member; member when empty |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "role": {
      "description": "owner or member; member when empty",
      "type": "string"
    },
    "room": {
      "description": "the Room's name",
      "type": "string"
    },
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "room",
    "user"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `set_user_email`

Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. Their old address gets a notice. For the administrators of the Cahut. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Set the e-mail address of a person. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `email` | string | yes | their e-mail address; empty removes it |
| `user` | string | yes | handle of the person |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "their e-mail address; empty removes it",
      "type": "string"
    },
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "user",
    "email"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room"
  ],
  "type": "object"
}
```

</details>

## `set_user_role`

Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. Knock first: a person approves the call before it runs.

Offered to the AIs of administrators of the Cahut only.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Change the role of a person. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `role` | string | yes | admin or member |
| `user` | string | yes | handle of the person |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "role": {
      "description": "admin or member",
      "type": "string"
    },
    "user": {
      "description": "handle of the person",
      "type": "string"
    }
  },
  "required": [
    "user",
    "role"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `id` | integer | yes |  |
| `name` | string | yes |  |
| `personal_room` | string | yes | name of the person's personal Room, the handle |
| `role` | string | yes | admin or member of the Cahut |
| `status` | string | yes | pending until the person chooses a password, then active, or disabled |
| `email` | string | no | the e-mail address of the person, which only the person and the administrators of the Cahut read |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "email": {
      "description": "the e-mail address of the person, which only the person and the administrators of the Cahut read",
      "type": "string"
    },
    "handle": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "personal_room": {
      "description": "name of the person's personal Room, the handle",
      "type": "string"
    },
    "role": {
      "description": "admin or member of the Cahut",
      "type": "string"
    },
    "status": {
      "description": "pending until the person chooses a password, then active, or disabled",
      "type": "string"
    }
  },
  "required": [
    "id",
    "handle",
    "name",
    "role",
    "status",
    "personal_room"
  ],
  "type": "object"
}
```

</details>

## `set_window_access`

Set who opens a Window: author, its author alone, for themself; room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. Knock first when it makes a Window public: a person approves the call before it runs.

Knock first when it makes a Window public: that call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Set who opens a Window. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: true`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `access` | string | yes | author, room, internal, key or public |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `ends_at` | string or null | no | when the access ends, such as 2026-12-31T18:00:00Z; after it, the Window opens to its Room only. It does not end when absent |
| `key` | string | no | for key: the Key to set, of 10 to 256 characters. When empty, the Window keeps its Key, or Cahut draws one |
| `new_key` | boolean | no | for key: draw a new Key; the old one no longer opens the Window |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "author, room, internal, key or public",
      "type": "string"
    },
    "ends_at": {
      "description": "when the access ends, such as 2026-12-31T18:00:00Z; after it, the Window opens to its Room only. It does not end when absent",
      "type": [
        "null",
        "string"
      ]
    },
    "key": {
      "description": "for key: the Key to set, of 10 to 256 characters. When empty, the Window keeps its Key, or Cahut draws one",
      "type": "string"
    },
    "new_key": {
      "description": "for key: draw a new Key; the old one no longer opens the Window",
      "type": "boolean"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window",
    "access"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `key` | string | no | the new Key, returned this once: give it to the person, who passes it on with the address |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "key": {
      "description": "the new Key, returned this once: give it to the person, who passes it on with the address",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `trash_window`

Put a Window of this Cahut in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was.

Title: Put a Window in the trash. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>

## `update_room`

Change the title of a Room of this Cahut, its description, or both. Its name, and so its address, stays. Give the old title and description again to undo it.

Title: Retitle and describe a Room. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `description` | string or null | no | the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent |
| `title` | string | no | the new title; the title stays when empty |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "description": {
      "description": "the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent",
      "type": [
        "null",
        "string"
      ]
    },
    "room": {
      "description": "the Room's name",
      "type": "string"
    },
    "title": {
      "description": "the new title; the title stays when empty",
      "type": "string"
    }
  },
  "required": [
    "room"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `kind` | string | yes | personal or team |
| `members` | integer | yes | how many people are in the Room |
| `name` | string | yes | the Room's name, first part of its address |
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `people[].kept` | integer | no | how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows of the Room you see, out of the trash: all but those that others keep for themselves |
| `description` | string | no | what the Room is, in a sentence or two; absent when it has none |
| `home` | string | no | the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them |
| `role` | string | no | your role in the Room, owner or member; empty when you are not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "description": {
      "description": "what the Room is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "home": {
      "description": "the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them",
      "type": "string"
    },
    "kind": {
      "description": "personal or team",
      "type": "string"
    },
    "members": {
      "description": "how many people are in the Room",
      "type": "integer"
    },
    "name": {
      "description": "the Room's name, first part of its address",
      "type": "string"
    },
    "people": {
      "description": "the people in the Room, owners first",
      "items": {
        "additionalProperties": false,
        "properties": {
          "handle": {
            "type": "string"
          },
          "kept": {
            "description": "how many Windows of the Room they keep for themselves, which open to the Room if they leave it; given to them, to the owners of the Room and to the administrators",
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "role": {
            "description": "owner or member of the Room",
            "type": "string"
          }
        },
        "required": [
          "handle",
          "name",
          "role"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "role": {
      "description": "your role in the Room, owner or member; empty when you are not in it",
      "type": "string"
    },
    "title": {
      "type": "string"
    },
    "windows": {
      "description": "how many Windows of the Room you see, out of the trash: all but those that others keep for themselves",
      "type": "integer"
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "members",
    "windows",
    "created_at",
    "people"
  ],
  "type": "object"
}
```

</details>

## `update_window`

Publish a new version of a Window of this Cahut: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. Cahut calls no other system.

Title: Update a Window. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `declaration` | object or null | yes | what the pages do: Cahut lets them load and call the origins listed here, and nothing else |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `files` | array of objects or null | yes | every file of the new version, as for publish_window; files left out are not in it |
| `files[].content` | string | yes |  |
| `files[].path` | string | yes | relative path with / as separator, such as index.html or img/logo.svg |
| `files[].encoding` | string | no | utf-8 (the default) or base64, for binary files |
| `window` | string | yes | the Window's id or address, or its Room and its name as room/name |
| `description` | string or null | no | a new description, a sentence or two of plain text; empty to remove it, and it stays when absent |
| `title` | string | no | a new title; the title stays when empty. The address does not change |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages do: Cahut lets them load and call the origins listed here, and nothing else",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "a new description, a sentence or two of plain text; empty to remove it, and it stays when absent",
      "type": [
        "null",
        "string"
      ]
    },
    "files": {
      "description": "every file of the new version, as for publish_window; files left out are not in it",
      "items": {
        "additionalProperties": false,
        "properties": {
          "content": {
            "type": "string"
          },
          "encoding": {
            "description": "utf-8 (the default) or base64, for binary files",
            "type": "string"
          },
          "path": {
            "description": "relative path with / as separator, such as index.html or img/logo.svg",
            "type": "string"
          }
        },
        "required": [
          "path",
          "content"
        ],
        "type": "object"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "description": "a new title; the title stays when empty. The address does not change",
      "type": "string"
    },
    "window": {
      "description": "the Window's id or address, or its Room and its name as room/name",
      "type": "string"
    }
  },
  "required": [
    "window",
    "files",
    "declaration"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `access` | string | yes | who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for, empty once that person is deleted |
| `file_count` | integer | yes |  |
| `id` | string | yes | the Window's id, which never changes: name the Window by it to find it whatever its name |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `permanent_url` | string | yes | the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `tags` | array of strings or null | yes | its tags, in order, which help find it among many |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make |
| `description` | string | no | what the Window is, in a sentence or two; absent when it has none |
| `excerpt` | string | no | for a search, a few words of its pages around the words found; absent when its pages hold none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |
| `withdrawn` | object or null | no | when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn |
| `withdrawn.at` | string | yes |  |
| `withdrawn.reason` | string | yes | why the moderators withdrew it, for the members of its Room |
| `withdrawn.by_server` | boolean | no | the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: author (its author alone), room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
      "type": "string"
    },
    "access_ends_at": {
      "description": "when the access ends, after which the Window opens to its Room only; absent when it does not end",
      "type": [
        "null",
        "string"
      ]
    },
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was first published for, empty once that person is deleted",
      "type": "string"
    },
    "declaration": {
      "additionalProperties": false,
      "description": "what the pages of the current version do, as declared when it was published; absent for a Window published before declarations",
      "properties": {
        "forms": {
          "description": "the origins that the forms of the pages send to, if any",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "origins": {
          "description": "every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make",
          "items": {
            "type": "string"
          },
          "type": [
            "null",
            "array"
          ]
        },
        "scripts": {
          "description": "whether the pages run scripts, inline or in files",
          "type": "boolean"
        }
      },
      "required": [
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "description": {
      "description": "what the Window is, in a sentence or two; absent when it has none",
      "type": "string"
    },
    "excerpt": {
      "description": "for a search, a few words of its pages around the words found; absent when its pages hold none",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "id": {
      "description": "the Window's id, which never changes: name the Window by it to find it whatever its name",
      "type": "string"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
      "type": "string"
    },
    "permanent_url": {
      "description": "the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts",
      "type": "string"
    },
    "previous_version": {
      "description": "the number of the version that rollback_window goes back to; absent when there is none",
      "type": "integer"
    },
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files of the current version, in bytes",
      "type": "integer"
    },
    "tags": {
      "description": "its tags, in order, which help find it among many",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "title": {
      "type": "string"
    },
    "trashed_at": {
      "description": "when the Window went to the trash; absent when it is not in it",
      "type": [
        "null",
        "string"
      ]
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
      "description": "the number of the version that readers see",
      "type": "integer"
    },
    "withdrawn": {
      "additionalProperties": false,
      "description": "when and why the moderators of this Cahut withdrew the Window, which serves nothing to its readers until they reinstate it; absent when it is not withdrawn",
      "properties": {
        "at": {
          "type": "string"
        },
        "by_server": {
          "description": "the administrators of the server withdrew it, and only they reinstate it; otherwise the administrators of the Cahut did",
          "type": "boolean"
        },
        "reason": {
          "description": "why the moderators withdrew it, for the members of its Room",
          "type": "string"
        }
      },
      "required": [
        "at",
        "reason"
      ],
      "type": [
        "null",
        "object"
      ]
    }
  },
  "required": [
    "id",
    "name",
    "title",
    "tags",
    "kind",
    "url",
    "permanent_url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access"
  ],
  "type": "object"
}
```

</details>
