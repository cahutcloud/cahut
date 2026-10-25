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

> Cahut keeps what you make on the user's own server, as Windows. To publish a document, deck, dashboard or site: call get_house_style with its kind, then publish_window with a complete index.html, its files linked by relative paths (css/deck.css, never /css/deck.css), and a declaration of its scripts and of every origin it loads from or calls. Give the user the address it returns. A call that cannot be undone makes a Knock: give the user its link to approve it.
>
> House style: when it is active, start from its template, link the style sheet of the organization, /-/house-style.css, use its classes and variables, and follow its guidance; when it is inactive, make the Window in the style the user asks for. Only the house style and the deck engine are under /-/: a Window is served under the address of its Room. Pages can load and call nothing that their declaration does not list, and send forms only where it says. They run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: keep state in memory or in the address. To change a Window, publish its new version with update_window at the same address; rollback_window goes back to the previous one. A Window opens to the members of its Room until set_window_access opens it to the whole Cahut or to anyone with its Key; making it public waits for a person. invite_to_window sends its address to people by e-mail, with a personal link for a Window with a Key. Knock first: a call that cannot be undone, or that changes what a person may do, does not run at once; once the person has answered its Knock, get_knock gives the outcome. Every call is recorded in the Logbook of the Cahut.

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
| [`create_user`](#create_user) | Create a person, with their personal Room, and the link where they choose their password, which goes to their e-mail address when you give one and the Cahut sends e-mail. Knock first: a person approves the call before it runs. |
| [`delete_room`](#delete_room) | Delete a team Room for good, with its members and its Windows, those in its trash included, and their files: this cannot be undone. Its address becomes free. Its owners and the administrators of the Cahut delete it; a personal Room is never deleted. Knock first: a person approves the call before it runs. |
| [`disable_user`](#disable_user) | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. Knock first: a person approves the call before it runs. |
| [`enable_user`](#enable_user) | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. Knock first: a person approves the call before it runs. |
| [`get_account`](#get_account) | Describe the person you act for: their handle, name, role in the Cahut and Rooms. |
| [`get_cahut`](#get_cahut) | Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused. |
| [`get_house_style`](#get_house_style) | Get the house style of this Cahut before making a Window: the name of the organization, its colors, fonts and logo, what it asks of AIs, the path of its style sheet, which a Window links from the root of the domain, a guide to its classes, and the template of the kind of Window you make. When it is inactive, it says so, and the Window takes the style the user asks for. |
| [`get_knock`](#get_knock) | Describe a Knock: whether it waits, was approved and run, failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked. |
| [`get_room`](#get_room) | Describe a Room: its title, its owners and members, and how many Windows it holds. |
| [`get_window`](#get_window) | Describe a Window and list its files, optionally with their contents. |
| [`invite_to_window`](#invite_to_window) | Send the address of a Window by e-mail, with a note: to people of the Cahut, named by handle, who sign in to read it, or to anyone else, named by address, when the Window opens with its Key or to the public. For a Window with a Key, each address gets a personal link that opens it for them alone, until revoke_invitation takes it back; the Key itself never goes. 20 people a call, and 100 invitations a day. get_cahut says whether the Cahut sends e-mail. |
| [`list_invitations`](#list_invitations) | List the invitations to a Window, the latest first: who got one, who invited, and whether it still stands. |
| [`list_knocks`](#list_knocks) | List Knocks, those that wait first: the calls that cannot be undone, or that change what a person may do, which wait for a person to approve them. An AI lists the Knocks it made; a person, those of their AIs. |
| [`list_logbook`](#list_logbook) | Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. |
| [`list_rooms`](#list_rooms) | List your Rooms, your personal Room first, with your role and how many members and Windows each has. |
| [`list_users`](#list_users) | List the people of the Cahut, with their handles, names, roles and states. In an individual Cahut, for its administrators only. |
| [`list_windows`](#list_windows) | List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first. |
| [`publish_window`](#publish_window) | Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. |
| [`purge_window`](#purge_window) | Delete a Window in the trash for good, with its versions and its files: this cannot be undone. Its address becomes free. Put the Window in the trash first, with trash_window. Knock first: a person approves the call before it runs. |
| [`remove_room_member`](#remove_room_member) | Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. Knock first: a person approves the call before it runs. |
| [`restore_window`](#restore_window) | Bring a Window back from the trash, with its versions. |
| [`revoke_invitation`](#revoke_invitation) | Take back an invitation to a Window: its personal link, and the passes it gave, no longer open the Window. The e-mail stays sent. |
| [`rollback_house_style`](#rollback_house_style) | Go back to the previous house style of this Cahut. The house style you leave becomes the previous one, so calling rollback_house_style again returns to it. For the administrators of the Cahut. |
| [`rollback_window`](#rollback_window) | Go back to the previous version of a Window. The version you leave becomes the previous one, so calling rollback_window again returns to it. |
| [`set_ai_reach`](#set_ai_reach) | Set what an AI connected over OAuth reaches: read or write, and every Room of its person or some of them. A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, and widening it, such as asking for write or for another Room, waits for its person to approve. Knock first when it widens the reach of an AI: a person approves the call before it runs. |
| [`set_house_style`](#set_house_style) | Change the house style of this Cahut: the name of the organization, its colors, its fonts and their files, its logo, what it asks of AIs, and whether AIs use it. Give only what changes: the rest stays. Every Window that links the style sheet takes the change at once. The house style before becomes the previous one, which rollback_house_style restores. For the administrators of the Cahut. |
| [`set_room_home`](#set_room_home) | Show a Window of a Room at the Room's address, &lt;domain&gt;/&lt;room&gt;/, in place of the page that lists its Windows, such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. An empty window goes back to the list. The Window keeps its own address and its access. |
| [`set_room_member_role`](#set_room_member_role) | Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. Knock first: a person approves the call before it runs. |
| [`set_user_email`](#set_user_email) | Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. Their old address gets a notice. For the administrators of the Cahut. Knock first: a person approves the call before it runs. |
| [`set_user_role`](#set_user_role) | Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. Knock first: a person approves the call before it runs. |
| [`set_window_access`](#set_window_access) | Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. Knock first when it makes a Window public: a person approves the call before it runs. |
| [`trash_window`](#trash_window) | Put a Window in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was. |
| [`update_room`](#update_room) | Change the title of a Room. Its name, and so its address, stays. |
| [`update_window`](#update_window) | Publish a new version of a Window: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. |

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
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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
| `name` | string | no | name of the Room, the first part of its address; from the title when empty |
| `owner` | string | no | handle of its first owner; you when empty. Only administrators of the Cahut name someone else |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
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
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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

Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused.

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
| `limits` | object | yes | the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only |
| `limits.admins` | integer or null | no | how many of them may be administrators of the Cahut |
| `limits.disk` | integer or null | no | how many bytes the Cahut may keep: every version of every Window, those in the trash included, and the files of the house style |
| `limits.disk_per_person` | integer or null | no | in an individual Cahut, how many bytes the Windows of each person may keep, every version included |
| `limits.people` | integer or null | no | how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count |
| `limits.rooms` | integer or null | no | how many team Rooms the Cahut may have, or in an individual Cahut how many Rooms beside the personal ones; personal Rooms come with their person and do not count |
| `limits.rooms_per_person` | integer or null | no | in an individual Cahut, how many Rooms each person may have beside their personal Room |
| `limits.windows` | integer or null | no | how many Windows out of the trash the Cahut may hold |
| `limits.windows_per_person` | integer or null | no | in an individual Cahut, how many Windows out of the trash each person may hold, over their Rooms |
| `limits.windows_per_room` | integer or null | no | how many Windows out of the trash each Room may hold, personal Rooms included |
| `sends_mail` | boolean | yes | whether this Cahut sends e-mail, through the relay that its administrators set; without one, nothing that would send e-mail works, such as invitations |
| `mcp` | string | no | the address of its MCP endpoint, where AI clients connect |
| `url` | string | no | the address of the Cahut, under which its Rooms and Windows open |
| `usage` | object or null | no | what the Cahut holds, as its limits count it; absent for a member of an individual Cahut |
| `usage.admins` | integer | yes | administrators who are not disabled |
| `usage.disk` | integer | yes | bytes of every version of every Window, those in the trash included, and of the files of the house style |
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
          "description": "how many bytes the Cahut may keep: every version of every Window, those in the trash included, and the files of the house style",
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
    "sends_mail": {
      "description": "whether this Cahut sends e-mail, through the relay that its administrators set; without one, nothing that would send e-mail works, such as invitations",
      "type": "boolean"
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
          "description": "bytes of every version of every Window, those in the trash included, and of the files of the house style",
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
    "limits"
  ],
  "type": "object"
}
```

</details>

## `get_house_style`

Get the house style of this Cahut before making a Window: the name of the organization, its colors, fonts and logo, what it asks of AIs, the path of its style sheet, which a Window links from the root of the domain, a guide to its classes, and the template of the kind of Window you make. When it is inactive, it says so, and the Window takes the style the user asks for.

Title: Get the house style. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `kind` | string | no | the kind of Window you make, to get its template: document, deck, dashboard or site |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "kind": {
      "description": "the kind of Window you make, to get its template: document, deck, dashboard or site",
      "type": "string"
    }
  },
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `active` | boolean | yes | whether AIs use the house style; when it is inactive, make the Window in the style the user asks for |
| `chart_colors` | array of strings or null | yes | the colors of the series of charts, in their order, as `--hs-chart-1` to `--hs-chart-8` in the style sheet: give them to series in this order, never cycled |
| `colors` | object | yes |  |
| `colors.background` | string | yes | the background of pages, `--hs-background` |
| `colors.line` | string | yes | borders and rules, `--hs-line` |
| `colors.muted` | string | yes | secondary text, such as captions, `--hs-muted` |
| `colors.on_primary` | string | yes | text on the primary color, `--hs-on-primary` |
| `colors.on_secondary` | string | yes | text on the secondary color, `--hs-on-secondary` |
| `colors.primary` | string | yes | the main color of the brand, for headings, links, buttons and slides, `--hs-primary` |
| `colors.secondary` | string | yes | the second color of the brand, for accents, `--hs-secondary` |
| `colors.surface` | string | yes | the background of cards and panels, `--hs-surface` |
| `colors.text` | string | yes | body text, `--hs-text` |
| `deck_engine` | string | yes | the path of the deck engine, from the root of the domain: &lt;script src="/-/deck.js" defer&gt;&lt;/script&gt; in the head of a deck, which then declares that it runs scripts |
| `fonts` | object | yes |  |
| `fonts.body` | object | yes | the font of body text, `--hs-font-body` |
| `fonts.body.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.body.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.body.family` | string | no | its family; absent for the font of the system |
| `fonts.body.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.body.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.body.files[].size` | integer | yes | in bytes |
| `fonts.body.files[].style` | string | yes | normal or italic |
| `fonts.body.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.body.files[].weight` | integer | yes |  |
| `fonts.body.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.heading` | object | yes | the font of headings, `--hs-font-heading` |
| `fonts.heading.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.heading.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.heading.family` | string | no | its family; absent for the font of the system |
| `fonts.heading.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.heading.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.heading.files[].size` | integer | yes | in bytes |
| `fonts.heading.files[].style` | string | yes | normal or italic |
| `fonts.heading.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.heading.files[].weight` | integer | yes |  |
| `fonts.heading.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.mono` | string | yes | the CSS stack of code and figures in columns, `--hs-font-mono` |
| `has_logo` | boolean | yes | whether the house style has a logo |
| `has_logo_inverse` | boolean | yes | whether it has a version of its logo for dark backgrounds |
| `logo` | string | yes | the path of the logo, from the root of the domain; without a logo, it shows the name of the organization |
| `logo_inverse` | string | yes | the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters |
| `name` | string | yes | the name of the organization; the domain of the Cahut until an administrator sets one |
| `neutral` | boolean | yes | whether the Cahut has the neutral house style, which no administrator has changed |
| `own_chart_colors` | boolean | yes | whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors |
| `stylesheet` | string | yes | the style sheet of the house style, at this path from the root of the domain: link it with &lt;link rel="stylesheet" href="/-/house-style.css"&gt;. The Window declares nothing for it |
| `guidance` | string | no | what the organization asks of AIs, such as its tone of voice or rules for charts: follow it |
| `guide` | string | no | how to use the style sheet, its variables and its classes, in Markdown |
| `message` | string | no | what to do when the house style is inactive |
| `previous` | object or null | no | the previous house style, which rollback_house_style restores; absent when there is none |
| `previous.neutral` | boolean | yes | whether it is the neutral house style |
| `previous.set_at` | string | yes |  |
| `previous.set_by` | string | no |  |
| `set_at` | string or null | no | when the house style was set; absent for the neutral house style |
| `set_by` | string | no | the handle of the person who set it, or for whom an AI set it |
| `template` | object or null | no | the template of the kind asked for |
| `template.html` | string | yes | index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title |
| `template.kind` | string | yes | the kind of Window it makes |
| `template.scripts` | boolean | yes | whether it runs scripts, which the Window then declares |
| `warnings` | array of strings or null | no | what may not read well, and what to change |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "active": {
      "description": "whether AIs use the house style; when it is inactive, make the Window in the style the user asks for",
      "type": "boolean"
    },
    "chart_colors": {
      "description": "the colors of the series of charts, in their order, as --hs-chart-1 to --hs-chart-8 in the style sheet: give them to series in this order, never cycled",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "colors": {
      "additionalProperties": false,
      "properties": {
        "background": {
          "description": "the background of pages, --hs-background",
          "type": "string"
        },
        "line": {
          "description": "borders and rules, --hs-line",
          "type": "string"
        },
        "muted": {
          "description": "secondary text, such as captions, --hs-muted",
          "type": "string"
        },
        "on_primary": {
          "description": "text on the primary color, --hs-on-primary",
          "type": "string"
        },
        "on_secondary": {
          "description": "text on the secondary color, --hs-on-secondary",
          "type": "string"
        },
        "primary": {
          "description": "the main color of the brand, for headings, links, buttons and slides, --hs-primary",
          "type": "string"
        },
        "secondary": {
          "description": "the second color of the brand, for accents, --hs-secondary",
          "type": "string"
        },
        "surface": {
          "description": "the background of cards and panels, --hs-surface",
          "type": "string"
        },
        "text": {
          "description": "body text, --hs-text",
          "type": "string"
        }
      },
      "required": [
        "text",
        "background",
        "primary",
        "on_primary",
        "secondary",
        "on_secondary",
        "muted",
        "line",
        "surface"
      ],
      "type": "object"
    },
    "deck_engine": {
      "description": "the path of the deck engine, from the root of the domain: \u003cscript src=\"/-/deck.js\" defer\u003e\u003c/script\u003e in the head of a deck, which then declares that it runs scripts",
      "type": "string"
    },
    "fonts": {
      "additionalProperties": false,
      "properties": {
        "body": {
          "additionalProperties": false,
          "description": "the font of body text, --hs-font-body",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "heading": {
          "additionalProperties": false,
          "description": "the font of headings, --hs-font-heading",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "mono": {
          "description": "the CSS stack of code and figures in columns, --hs-font-mono",
          "type": "string"
        }
      },
      "required": [
        "heading",
        "body",
        "mono"
      ],
      "type": "object"
    },
    "guidance": {
      "description": "what the organization asks of AIs, such as its tone of voice or rules for charts: follow it",
      "type": "string"
    },
    "guide": {
      "description": "how to use the style sheet, its variables and its classes, in Markdown",
      "type": "string"
    },
    "has_logo": {
      "description": "whether the house style has a logo",
      "type": "boolean"
    },
    "has_logo_inverse": {
      "description": "whether it has a version of its logo for dark backgrounds",
      "type": "boolean"
    },
    "logo": {
      "description": "the path of the logo, from the root of the domain; without a logo, it shows the name of the organization",
      "type": "string"
    },
    "logo_inverse": {
      "description": "the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters",
      "type": "string"
    },
    "message": {
      "description": "what to do when the house style is inactive",
      "type": "string"
    },
    "name": {
      "description": "the name of the organization; the domain of the Cahut until an administrator sets one",
      "type": "string"
    },
    "neutral": {
      "description": "whether the Cahut has the neutral house style, which no administrator has changed",
      "type": "boolean"
    },
    "own_chart_colors": {
      "description": "whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors",
      "type": "boolean"
    },
    "previous": {
      "additionalProperties": false,
      "description": "the previous house style, which rollback_house_style restores; absent when there is none",
      "properties": {
        "neutral": {
          "description": "whether it is the neutral house style",
          "type": "boolean"
        },
        "set_at": {
          "type": "string"
        },
        "set_by": {
          "type": "string"
        }
      },
      "required": [
        "neutral",
        "set_at"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "set_at": {
      "description": "when the house style was set; absent for the neutral house style",
      "type": [
        "null",
        "string"
      ]
    },
    "set_by": {
      "description": "the handle of the person who set it, or for whom an AI set it",
      "type": "string"
    },
    "stylesheet": {
      "description": "the style sheet of the house style, at this path from the root of the domain: link it with \u003clink rel=\"stylesheet\" href=\"/-/house-style.css\"\u003e. The Window declares nothing for it",
      "type": "string"
    },
    "template": {
      "additionalProperties": false,
      "description": "the template of the kind asked for",
      "properties": {
        "html": {
          "description": "index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title",
          "type": "string"
        },
        "kind": {
          "description": "the kind of Window it makes",
          "type": "string"
        },
        "scripts": {
          "description": "whether it runs scripts, which the Window then declares",
          "type": "boolean"
        }
      },
      "required": [
        "kind",
        "html",
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "warnings": {
      "description": "what may not read well, and what to change",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "name",
    "active",
    "neutral",
    "stylesheet",
    "logo",
    "logo_inverse",
    "has_logo",
    "has_logo_inverse",
    "deck_engine",
    "colors",
    "own_chart_colors",
    "chart_colors",
    "fonts"
  ],
  "type": "object"
}
```

</details>

## `get_knock`

Describe a Knock: whether it waits, was approved and run, failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked.

Title: Describe a Knock. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

Describe a Room: its title, its owners and members, and how many Windows it holds.

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
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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

Describe a Window and list its files, optionally with their contents.

Title: Describe a Window. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |
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
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `files` | array of objects or null | yes | the files of the current version |
| `files[].path` | string | yes |  |
| `files[].size` | integer | yes |  |
| `files[].type` | string | yes | media type |
| `files[].content` | string | no |  |
| `files[].encoding` | string | no | utf-8 or base64 |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `versions` | array of objects or null | yes | the current version, then the previous one if there is one |
| `versions[].current` | boolean | yes | whether readers see this version |
| `versions[].file_count` | integer | yes |  |
| `versions[].number` | integer | yes |  |
| `versions[].published_at` | string | yes |  |
| `versions[].published_by` | string | yes | handle of the person it was published for |
| `versions[].size` | integer | yes | total size of its files, in bytes |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
            "description": "handle of the person it was published for",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
    "room",
    "version",
    "size",
    "file_count",
    "created_by",
    "created_at",
    "updated_at",
    "access",
    "versions",
    "files"
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
| `window` | string | yes | the Window's address, or its Room and its name as room/name |
| `note` | string | no | a few words for them, in plain text, 500 characters at most |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "note": {
      "description": "a few words for them, in plain text, 500 characters at most",
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
      "description": "the Window's address, or its Room and its name as room/name",
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

List the invitations to a Window, the latest first: who got one, who invited, and whether it still stands.

Title: List the invitations to a Window. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's address, or its Room and its name as room/name",
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

List Knocks, those that wait first: the calls that cannot be undone, or that change what a person may do, which wait for a person to approve them. An AI lists the Knocks it made; a person, those of their AIs.

Title: List Knocks. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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
| `entries[].via` | string | yes | the adapter the call came through: cli, api, mcp or oauth |
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
            "description": "the adapter the call came through: cli, api, mcp or oauth",
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

List your Rooms, your personal Room first, with your role and how many members and Windows each has.

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
| `rooms[].windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
            "description": "how many Windows the Room holds, out of the trash",
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

List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first.

Title: List Windows. Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | no | only the Windows of this Room, by its name |
| `trashed` | boolean | no | list the Windows in the trash instead |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "room": {
      "description": "only the Windows of this Room, by its name",
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
| `windows[].access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `windows[].created_at` | string | yes |  |
| `windows[].created_by` | string | yes | handle of the person it was first published for |
| `windows[].file_count` | integer | yes |  |
| `windows[].kind` | string | yes |  |
| `windows[].name` | string | yes | the Window's name, last part of its address |
| `windows[].room` | string | yes | the name of its Room, the part of its address before its name |
| `windows[].size` | integer | yes | total size of the files of the current version, in bytes |
| `windows[].title` | string | yes |  |
| `windows[].updated_at` | string | yes |  |
| `windows[].url` | string | yes | where the Window opens |
| `windows[].version` | integer | yes | the number of the version that readers see |
| `windows[].access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `windows[].declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `windows[].declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `windows[].declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `windows[].declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `windows[].previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `windows[].trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

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
            "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
            "description": "handle of the person it was first published for",
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
                "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
          "file_count": {
            "type": "integer"
          },
          "kind": {
            "type": "string"
          },
          "name": {
            "description": "the Window's name, last part of its address",
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
          }
        },
        "required": [
          "name",
          "title",
          "kind",
          "url",
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

Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens.

Title: Publish a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `declaration` | object or null | yes | what the pages do: Cahut lets them load and call the origins listed here, and nothing else |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `files` | array of objects or null | yes | the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css |
| `files[].content` | string | yes |  |
| `files[].path` | string | yes | relative path with / as separator, such as index.html or img/logo.svg |
| `files[].encoding` | string | no | utf-8 (the default) or base64, for binary files |
| `kind` | string | yes | one of document, deck, dashboard or site |
| `title` | string | yes | title of the Window, also used in its address |
| `room` | string | no | name of the Room to publish in; the personal Room of the person you act for when empty |

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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "room": {
      "description": "name of the Room to publish in; the personal Room of the person you act for when empty",
      "type": "string"
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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

Delete a Window in the trash for good, with its versions and its files: this cannot be undone. Its address becomes free. Put the Window in the trash first, with trash_window. Knock first: a person approves the call before it runs.

Knock first: the call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Delete a Window for good. Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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

## `restore_window`

Bring a Window back from the trash, with its versions.

Title: Restore a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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

Take back an invitation to a Window: its personal link, and the passes it gave, no longer open the Window. The e-mail stays sent.

Title: Take back an invitation. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

## `rollback_house_style`

Go back to the previous house style of this Cahut. The house style you leave becomes the previous one, so calling rollback_house_style again returns to it. For the administrators of the Cahut.

Offered to the AIs of administrators of the Cahut only.

Title: Roll back the house style. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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
| `active` | boolean | yes | whether AIs use the house style; when it is inactive, make the Window in the style the user asks for |
| `chart_colors` | array of strings or null | yes | the colors of the series of charts, in their order, as `--hs-chart-1` to `--hs-chart-8` in the style sheet: give them to series in this order, never cycled |
| `colors` | object | yes |  |
| `colors.background` | string | yes | the background of pages, `--hs-background` |
| `colors.line` | string | yes | borders and rules, `--hs-line` |
| `colors.muted` | string | yes | secondary text, such as captions, `--hs-muted` |
| `colors.on_primary` | string | yes | text on the primary color, `--hs-on-primary` |
| `colors.on_secondary` | string | yes | text on the secondary color, `--hs-on-secondary` |
| `colors.primary` | string | yes | the main color of the brand, for headings, links, buttons and slides, `--hs-primary` |
| `colors.secondary` | string | yes | the second color of the brand, for accents, `--hs-secondary` |
| `colors.surface` | string | yes | the background of cards and panels, `--hs-surface` |
| `colors.text` | string | yes | body text, `--hs-text` |
| `deck_engine` | string | yes | the path of the deck engine, from the root of the domain: &lt;script src="/-/deck.js" defer&gt;&lt;/script&gt; in the head of a deck, which then declares that it runs scripts |
| `fonts` | object | yes |  |
| `fonts.body` | object | yes | the font of body text, `--hs-font-body` |
| `fonts.body.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.body.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.body.family` | string | no | its family; absent for the font of the system |
| `fonts.body.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.body.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.body.files[].size` | integer | yes | in bytes |
| `fonts.body.files[].style` | string | yes | normal or italic |
| `fonts.body.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.body.files[].weight` | integer | yes |  |
| `fonts.body.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.heading` | object | yes | the font of headings, `--hs-font-heading` |
| `fonts.heading.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.heading.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.heading.family` | string | no | its family; absent for the font of the system |
| `fonts.heading.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.heading.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.heading.files[].size` | integer | yes | in bytes |
| `fonts.heading.files[].style` | string | yes | normal or italic |
| `fonts.heading.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.heading.files[].weight` | integer | yes |  |
| `fonts.heading.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.mono` | string | yes | the CSS stack of code and figures in columns, `--hs-font-mono` |
| `has_logo` | boolean | yes | whether the house style has a logo |
| `has_logo_inverse` | boolean | yes | whether it has a version of its logo for dark backgrounds |
| `logo` | string | yes | the path of the logo, from the root of the domain; without a logo, it shows the name of the organization |
| `logo_inverse` | string | yes | the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters |
| `name` | string | yes | the name of the organization; the domain of the Cahut until an administrator sets one |
| `neutral` | boolean | yes | whether the Cahut has the neutral house style, which no administrator has changed |
| `own_chart_colors` | boolean | yes | whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors |
| `stylesheet` | string | yes | the style sheet of the house style, at this path from the root of the domain: link it with &lt;link rel="stylesheet" href="/-/house-style.css"&gt;. The Window declares nothing for it |
| `guidance` | string | no | what the organization asks of AIs, such as its tone of voice or rules for charts: follow it |
| `guide` | string | no | how to use the style sheet, its variables and its classes, in Markdown |
| `message` | string | no | what to do when the house style is inactive |
| `previous` | object or null | no | the previous house style, which rollback_house_style restores; absent when there is none |
| `previous.neutral` | boolean | yes | whether it is the neutral house style |
| `previous.set_at` | string | yes |  |
| `previous.set_by` | string | no |  |
| `set_at` | string or null | no | when the house style was set; absent for the neutral house style |
| `set_by` | string | no | the handle of the person who set it, or for whom an AI set it |
| `template` | object or null | no | the template of the kind asked for |
| `template.html` | string | yes | index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title |
| `template.kind` | string | yes | the kind of Window it makes |
| `template.scripts` | boolean | yes | whether it runs scripts, which the Window then declares |
| `warnings` | array of strings or null | no | what may not read well, and what to change |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "active": {
      "description": "whether AIs use the house style; when it is inactive, make the Window in the style the user asks for",
      "type": "boolean"
    },
    "chart_colors": {
      "description": "the colors of the series of charts, in their order, as --hs-chart-1 to --hs-chart-8 in the style sheet: give them to series in this order, never cycled",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "colors": {
      "additionalProperties": false,
      "properties": {
        "background": {
          "description": "the background of pages, --hs-background",
          "type": "string"
        },
        "line": {
          "description": "borders and rules, --hs-line",
          "type": "string"
        },
        "muted": {
          "description": "secondary text, such as captions, --hs-muted",
          "type": "string"
        },
        "on_primary": {
          "description": "text on the primary color, --hs-on-primary",
          "type": "string"
        },
        "on_secondary": {
          "description": "text on the secondary color, --hs-on-secondary",
          "type": "string"
        },
        "primary": {
          "description": "the main color of the brand, for headings, links, buttons and slides, --hs-primary",
          "type": "string"
        },
        "secondary": {
          "description": "the second color of the brand, for accents, --hs-secondary",
          "type": "string"
        },
        "surface": {
          "description": "the background of cards and panels, --hs-surface",
          "type": "string"
        },
        "text": {
          "description": "body text, --hs-text",
          "type": "string"
        }
      },
      "required": [
        "text",
        "background",
        "primary",
        "on_primary",
        "secondary",
        "on_secondary",
        "muted",
        "line",
        "surface"
      ],
      "type": "object"
    },
    "deck_engine": {
      "description": "the path of the deck engine, from the root of the domain: \u003cscript src=\"/-/deck.js\" defer\u003e\u003c/script\u003e in the head of a deck, which then declares that it runs scripts",
      "type": "string"
    },
    "fonts": {
      "additionalProperties": false,
      "properties": {
        "body": {
          "additionalProperties": false,
          "description": "the font of body text, --hs-font-body",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "heading": {
          "additionalProperties": false,
          "description": "the font of headings, --hs-font-heading",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "mono": {
          "description": "the CSS stack of code and figures in columns, --hs-font-mono",
          "type": "string"
        }
      },
      "required": [
        "heading",
        "body",
        "mono"
      ],
      "type": "object"
    },
    "guidance": {
      "description": "what the organization asks of AIs, such as its tone of voice or rules for charts: follow it",
      "type": "string"
    },
    "guide": {
      "description": "how to use the style sheet, its variables and its classes, in Markdown",
      "type": "string"
    },
    "has_logo": {
      "description": "whether the house style has a logo",
      "type": "boolean"
    },
    "has_logo_inverse": {
      "description": "whether it has a version of its logo for dark backgrounds",
      "type": "boolean"
    },
    "logo": {
      "description": "the path of the logo, from the root of the domain; without a logo, it shows the name of the organization",
      "type": "string"
    },
    "logo_inverse": {
      "description": "the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters",
      "type": "string"
    },
    "message": {
      "description": "what to do when the house style is inactive",
      "type": "string"
    },
    "name": {
      "description": "the name of the organization; the domain of the Cahut until an administrator sets one",
      "type": "string"
    },
    "neutral": {
      "description": "whether the Cahut has the neutral house style, which no administrator has changed",
      "type": "boolean"
    },
    "own_chart_colors": {
      "description": "whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors",
      "type": "boolean"
    },
    "previous": {
      "additionalProperties": false,
      "description": "the previous house style, which rollback_house_style restores; absent when there is none",
      "properties": {
        "neutral": {
          "description": "whether it is the neutral house style",
          "type": "boolean"
        },
        "set_at": {
          "type": "string"
        },
        "set_by": {
          "type": "string"
        }
      },
      "required": [
        "neutral",
        "set_at"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "set_at": {
      "description": "when the house style was set; absent for the neutral house style",
      "type": [
        "null",
        "string"
      ]
    },
    "set_by": {
      "description": "the handle of the person who set it, or for whom an AI set it",
      "type": "string"
    },
    "stylesheet": {
      "description": "the style sheet of the house style, at this path from the root of the domain: link it with \u003clink rel=\"stylesheet\" href=\"/-/house-style.css\"\u003e. The Window declares nothing for it",
      "type": "string"
    },
    "template": {
      "additionalProperties": false,
      "description": "the template of the kind asked for",
      "properties": {
        "html": {
          "description": "index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title",
          "type": "string"
        },
        "kind": {
          "description": "the kind of Window it makes",
          "type": "string"
        },
        "scripts": {
          "description": "whether it runs scripts, which the Window then declares",
          "type": "boolean"
        }
      },
      "required": [
        "kind",
        "html",
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "warnings": {
      "description": "what may not read well, and what to change",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "name",
    "active",
    "neutral",
    "stylesheet",
    "logo",
    "logo_inverse",
    "has_logo",
    "has_logo_inverse",
    "deck_engine",
    "colors",
    "own_chart_colors",
    "chart_colors",
    "fonts"
  ],
  "type": "object"
}
```

</details>

## `rollback_window`

Go back to the previous version of a Window. The version you leave becomes the previous one, so calling rollback_window again returns to it.

Title: Roll back a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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

Set what an AI connected over OAuth reaches: read or write, and every Room of its person or some of them. A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, and widening it, such as asking for write or for another Room, waits for its person to approve. Knock first when it widens the reach of an AI: a person approves the call before it runs.

Offered also to the AIs that may only read, though it changes things.

Knock first when it widens the reach of an AI: that call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Change what an AI reaches. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

## `set_house_style`

Change the house style of this Cahut: the name of the organization, its colors, its fonts and their files, its logo, what it asks of AIs, and whether AIs use it. Give only what changes: the rest stays. Every Window that links the style sheet takes the change at once. The house style before becomes the previous one, which rollback_house_style restores. For the administrators of the Cahut.

Offered to the AIs of administrators of the Cahut only.

Title: Change the house style. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `active` | boolean or null | no | whether AIs use the house style: false makes it inactive, and get_house_style then gives them nothing of it; the pages of Cahut and the Windows that link its sheet keep it |
| `body_font` | object or null | no | the font of body text, which replaces the current one with its files; with no family, the font of the system |
| `body_font.fallback` | string | no | the generic family to fall back on: sans-serif (the default), serif or monospace |
| `body_font.family` | string | no | the family, such as Inter or Georgia; empty for the default |
| `body_font.files` | array of objects or null | no | the files of the family, which readers' browsers load; none for a family that their systems have, such as Georgia |
| `body_font.files[].content` | string | no | the file in base64: WOFF2, WOFF, TrueType or OpenType |
| `body_font.files[].keep` | string | no | instead of content, the name of a file of the current house style to keep, as get_house_style lists it |
| `body_font.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `body_font.files[].style` | string | no | normal (the default) or italic |
| `body_font.files[].weight` | integer | no | its weight, from 100 to 900, 400 by default; for a variable font, the lightest weight it covers |
| `chart_colors` | array of strings or null | no | up to 8 colors for the series of charts, in order, each as #rrggbb; an empty list goes back to those that Cahut derives from the primary and secondary colors |
| `colors` | object or null | no | the colors to change; those left out stay |
| `colors.background` | string | no | the background of pages, as #rrggbb; body text must contrast 4.5:1 at least on it |
| `colors.primary` | string | no | the main color of the brand, as #rrggbb |
| `colors.secondary` | string | no | the second color of the brand, as #rrggbb |
| `colors.text` | string | no | body text, as #rrggbb |
| `guidance` | string or null | no | what the organization asks of AIs, at most 4000 characters: its tone of voice, its rules for titles or charts. Readers never see it |
| `heading_font` | object or null | no | the font of headings, which replaces the current one with its files; with no family, headings take the body font |
| `heading_font.fallback` | string | no | the generic family to fall back on: sans-serif (the default), serif or monospace |
| `heading_font.family` | string | no | the family, such as Inter or Georgia; empty for the default |
| `heading_font.files` | array of objects or null | no | the files of the family, which readers' browsers load; none for a family that their systems have, such as Georgia |
| `heading_font.files[].content` | string | no | the file in base64: WOFF2, WOFF, TrueType or OpenType |
| `heading_font.files[].keep` | string | no | instead of content, the name of a file of the current house style to keep, as get_house_style lists it |
| `heading_font.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `heading_font.files[].style` | string | no | normal (the default) or italic |
| `heading_font.files[].weight` | integer | no | its weight, from 100 to 900, 400 by default; for a variable font, the lightest weight it covers |
| `logo` | object or null | no | the logo of the organization, for light backgrounds |
| `logo.content` | string | yes | the image: SVG as text, or PNG or WebP in base64; empty to remove the logo |
| `logo.encoding` | string | no | utf-8 (the default, for SVG) or base64 |
| `logo_inverse` | object or null | no | the version of the logo for dark backgrounds, such as slides in the primary color |
| `logo_inverse.content` | string | yes | the image: SVG as text, or PNG or WebP in base64; empty to remove the logo |
| `logo_inverse.encoding` | string | no | utf-8 (the default, for SVG) or base64 |
| `name` | string or null | no | the name of the organization; empty for the domain of the Cahut |
| `neutral` | boolean | no | start again from the neutral house style, then apply the other fields |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "active": {
      "description": "whether AIs use the house style: false makes it inactive, and get_house_style then gives them nothing of it; the pages of Cahut and the Windows that link its sheet keep it",
      "type": [
        "null",
        "boolean"
      ]
    },
    "body_font": {
      "additionalProperties": false,
      "description": "the font of body text, which replaces the current one with its files; with no family, the font of the system",
      "properties": {
        "fallback": {
          "description": "the generic family to fall back on: sans-serif (the default), serif or monospace",
          "type": "string"
        },
        "family": {
          "description": "the family, such as Inter or Georgia; empty for the default",
          "type": "string"
        },
        "files": {
          "description": "the files of the family, which readers' browsers load; none for a family that their systems have, such as Georgia",
          "items": {
            "additionalProperties": false,
            "properties": {
              "content": {
                "description": "the file in base64: WOFF2, WOFF, TrueType or OpenType",
                "type": "string"
              },
              "keep": {
                "description": "instead of content, the name of a file of the current house style to keep, as get_house_style lists it",
                "type": "string"
              },
              "max_weight": {
                "description": "for a variable font, the heaviest weight it covers",
                "type": "integer"
              },
              "style": {
                "description": "normal (the default) or italic",
                "type": "string"
              },
              "weight": {
                "description": "its weight, from 100 to 900, 400 by default; for a variable font, the lightest weight it covers",
                "type": "integer"
              }
            },
            "type": "object"
          },
          "type": [
            "null",
            "array"
          ]
        }
      },
      "type": [
        "null",
        "object"
      ]
    },
    "chart_colors": {
      "description": "up to 8 colors for the series of charts, in order, each as #rrggbb; an empty list goes back to those that Cahut derives from the primary and secondary colors",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "colors": {
      "additionalProperties": false,
      "description": "the colors to change; those left out stay",
      "properties": {
        "background": {
          "description": "the background of pages, as #rrggbb; body text must contrast 4.5:1 at least on it",
          "type": "string"
        },
        "primary": {
          "description": "the main color of the brand, as #rrggbb",
          "type": "string"
        },
        "secondary": {
          "description": "the second color of the brand, as #rrggbb",
          "type": "string"
        },
        "text": {
          "description": "body text, as #rrggbb",
          "type": "string"
        }
      },
      "type": [
        "null",
        "object"
      ]
    },
    "guidance": {
      "description": "what the organization asks of AIs, at most 4000 characters: its tone of voice, its rules for titles or charts. Readers never see it",
      "type": [
        "null",
        "string"
      ]
    },
    "heading_font": {
      "additionalProperties": false,
      "description": "the font of headings, which replaces the current one with its files; with no family, headings take the body font",
      "properties": {
        "fallback": {
          "description": "the generic family to fall back on: sans-serif (the default), serif or monospace",
          "type": "string"
        },
        "family": {
          "description": "the family, such as Inter or Georgia; empty for the default",
          "type": "string"
        },
        "files": {
          "description": "the files of the family, which readers' browsers load; none for a family that their systems have, such as Georgia",
          "items": {
            "additionalProperties": false,
            "properties": {
              "content": {
                "description": "the file in base64: WOFF2, WOFF, TrueType or OpenType",
                "type": "string"
              },
              "keep": {
                "description": "instead of content, the name of a file of the current house style to keep, as get_house_style lists it",
                "type": "string"
              },
              "max_weight": {
                "description": "for a variable font, the heaviest weight it covers",
                "type": "integer"
              },
              "style": {
                "description": "normal (the default) or italic",
                "type": "string"
              },
              "weight": {
                "description": "its weight, from 100 to 900, 400 by default; for a variable font, the lightest weight it covers",
                "type": "integer"
              }
            },
            "type": "object"
          },
          "type": [
            "null",
            "array"
          ]
        }
      },
      "type": [
        "null",
        "object"
      ]
    },
    "logo": {
      "additionalProperties": false,
      "description": "the logo of the organization, for light backgrounds",
      "properties": {
        "content": {
          "description": "the image: SVG as text, or PNG or WebP in base64; empty to remove the logo",
          "type": "string"
        },
        "encoding": {
          "description": "utf-8 (the default, for SVG) or base64",
          "type": "string"
        }
      },
      "required": [
        "content"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "logo_inverse": {
      "additionalProperties": false,
      "description": "the version of the logo for dark backgrounds, such as slides in the primary color",
      "properties": {
        "content": {
          "description": "the image: SVG as text, or PNG or WebP in base64; empty to remove the logo",
          "type": "string"
        },
        "encoding": {
          "description": "utf-8 (the default, for SVG) or base64",
          "type": "string"
        }
      },
      "required": [
        "content"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "name": {
      "description": "the name of the organization; empty for the domain of the Cahut",
      "type": [
        "null",
        "string"
      ]
    },
    "neutral": {
      "description": "start again from the neutral house style, then apply the other fields",
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
| `active` | boolean | yes | whether AIs use the house style; when it is inactive, make the Window in the style the user asks for |
| `chart_colors` | array of strings or null | yes | the colors of the series of charts, in their order, as `--hs-chart-1` to `--hs-chart-8` in the style sheet: give them to series in this order, never cycled |
| `colors` | object | yes |  |
| `colors.background` | string | yes | the background of pages, `--hs-background` |
| `colors.line` | string | yes | borders and rules, `--hs-line` |
| `colors.muted` | string | yes | secondary text, such as captions, `--hs-muted` |
| `colors.on_primary` | string | yes | text on the primary color, `--hs-on-primary` |
| `colors.on_secondary` | string | yes | text on the secondary color, `--hs-on-secondary` |
| `colors.primary` | string | yes | the main color of the brand, for headings, links, buttons and slides, `--hs-primary` |
| `colors.secondary` | string | yes | the second color of the brand, for accents, `--hs-secondary` |
| `colors.surface` | string | yes | the background of cards and panels, `--hs-surface` |
| `colors.text` | string | yes | body text, `--hs-text` |
| `deck_engine` | string | yes | the path of the deck engine, from the root of the domain: &lt;script src="/-/deck.js" defer&gt;&lt;/script&gt; in the head of a deck, which then declares that it runs scripts |
| `fonts` | object | yes |  |
| `fonts.body` | object | yes | the font of body text, `--hs-font-body` |
| `fonts.body.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.body.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.body.family` | string | no | its family; absent for the font of the system |
| `fonts.body.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.body.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.body.files[].size` | integer | yes | in bytes |
| `fonts.body.files[].style` | string | yes | normal or italic |
| `fonts.body.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.body.files[].weight` | integer | yes |  |
| `fonts.body.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.heading` | object | yes | the font of headings, `--hs-font-heading` |
| `fonts.heading.fallback` | string | yes | the generic family it falls back on: sans-serif, serif or monospace |
| `fonts.heading.stack` | string | yes | the CSS font-family of the font, its family first |
| `fonts.heading.family` | string | no | its family; absent for the font of the system |
| `fonts.heading.files` | array of objects or null | no | the files that readers' browsers load; none for a family that their systems have |
| `fonts.heading.files[].name` | string | yes | the name of the file, which set_house_style keeps when given in keep |
| `fonts.heading.files[].size` | integer | yes | in bytes |
| `fonts.heading.files[].style` | string | yes | normal or italic |
| `fonts.heading.files[].url` | string | yes | its path, from the root of the domain |
| `fonts.heading.files[].weight` | integer | yes |  |
| `fonts.heading.files[].max_weight` | integer | no | for a variable font, the heaviest weight it covers |
| `fonts.mono` | string | yes | the CSS stack of code and figures in columns, `--hs-font-mono` |
| `has_logo` | boolean | yes | whether the house style has a logo |
| `has_logo_inverse` | boolean | yes | whether it has a version of its logo for dark backgrounds |
| `logo` | string | yes | the path of the logo, from the root of the domain; without a logo, it shows the name of the organization |
| `logo_inverse` | string | yes | the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters |
| `name` | string | yes | the name of the organization; the domain of the Cahut until an administrator sets one |
| `neutral` | boolean | yes | whether the Cahut has the neutral house style, which no administrator has changed |
| `own_chart_colors` | boolean | yes | whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors |
| `stylesheet` | string | yes | the style sheet of the house style, at this path from the root of the domain: link it with &lt;link rel="stylesheet" href="/-/house-style.css"&gt;. The Window declares nothing for it |
| `guidance` | string | no | what the organization asks of AIs, such as its tone of voice or rules for charts: follow it |
| `guide` | string | no | how to use the style sheet, its variables and its classes, in Markdown |
| `message` | string | no | what to do when the house style is inactive |
| `previous` | object or null | no | the previous house style, which rollback_house_style restores; absent when there is none |
| `previous.neutral` | boolean | yes | whether it is the neutral house style |
| `previous.set_at` | string | yes |  |
| `previous.set_by` | string | no |  |
| `set_at` | string or null | no | when the house style was set; absent for the neutral house style |
| `set_by` | string | no | the handle of the person who set it, or for whom an AI set it |
| `template` | object or null | no | the template of the kind asked for |
| `template.html` | string | yes | index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title |
| `template.kind` | string | yes | the kind of Window it makes |
| `template.scripts` | boolean | yes | whether it runs scripts, which the Window then declares |
| `warnings` | array of strings or null | no | what may not read well, and what to change |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "active": {
      "description": "whether AIs use the house style; when it is inactive, make the Window in the style the user asks for",
      "type": "boolean"
    },
    "chart_colors": {
      "description": "the colors of the series of charts, in their order, as --hs-chart-1 to --hs-chart-8 in the style sheet: give them to series in this order, never cycled",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    },
    "colors": {
      "additionalProperties": false,
      "properties": {
        "background": {
          "description": "the background of pages, --hs-background",
          "type": "string"
        },
        "line": {
          "description": "borders and rules, --hs-line",
          "type": "string"
        },
        "muted": {
          "description": "secondary text, such as captions, --hs-muted",
          "type": "string"
        },
        "on_primary": {
          "description": "text on the primary color, --hs-on-primary",
          "type": "string"
        },
        "on_secondary": {
          "description": "text on the secondary color, --hs-on-secondary",
          "type": "string"
        },
        "primary": {
          "description": "the main color of the brand, for headings, links, buttons and slides, --hs-primary",
          "type": "string"
        },
        "secondary": {
          "description": "the second color of the brand, for accents, --hs-secondary",
          "type": "string"
        },
        "surface": {
          "description": "the background of cards and panels, --hs-surface",
          "type": "string"
        },
        "text": {
          "description": "body text, --hs-text",
          "type": "string"
        }
      },
      "required": [
        "text",
        "background",
        "primary",
        "on_primary",
        "secondary",
        "on_secondary",
        "muted",
        "line",
        "surface"
      ],
      "type": "object"
    },
    "deck_engine": {
      "description": "the path of the deck engine, from the root of the domain: \u003cscript src=\"/-/deck.js\" defer\u003e\u003c/script\u003e in the head of a deck, which then declares that it runs scripts",
      "type": "string"
    },
    "fonts": {
      "additionalProperties": false,
      "properties": {
        "body": {
          "additionalProperties": false,
          "description": "the font of body text, --hs-font-body",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "heading": {
          "additionalProperties": false,
          "description": "the font of headings, --hs-font-heading",
          "properties": {
            "fallback": {
              "description": "the generic family it falls back on: sans-serif, serif or monospace",
              "type": "string"
            },
            "family": {
              "description": "its family; absent for the font of the system",
              "type": "string"
            },
            "files": {
              "description": "the files that readers' browsers load; none for a family that their systems have",
              "items": {
                "additionalProperties": false,
                "properties": {
                  "max_weight": {
                    "description": "for a variable font, the heaviest weight it covers",
                    "type": "integer"
                  },
                  "name": {
                    "description": "the name of the file, which set_house_style keeps when given in keep",
                    "type": "string"
                  },
                  "size": {
                    "description": "in bytes",
                    "type": "integer"
                  },
                  "style": {
                    "description": "normal or italic",
                    "type": "string"
                  },
                  "url": {
                    "description": "its path, from the root of the domain",
                    "type": "string"
                  },
                  "weight": {
                    "type": "integer"
                  }
                },
                "required": [
                  "name",
                  "url",
                  "weight",
                  "style",
                  "size"
                ],
                "type": "object"
              },
              "type": [
                "null",
                "array"
              ]
            },
            "stack": {
              "description": "the CSS font-family of the font, its family first",
              "type": "string"
            }
          },
          "required": [
            "fallback",
            "stack"
          ],
          "type": "object"
        },
        "mono": {
          "description": "the CSS stack of code and figures in columns, --hs-font-mono",
          "type": "string"
        }
      },
      "required": [
        "heading",
        "body",
        "mono"
      ],
      "type": "object"
    },
    "guidance": {
      "description": "what the organization asks of AIs, such as its tone of voice or rules for charts: follow it",
      "type": "string"
    },
    "guide": {
      "description": "how to use the style sheet, its variables and its classes, in Markdown",
      "type": "string"
    },
    "has_logo": {
      "description": "whether the house style has a logo",
      "type": "boolean"
    },
    "has_logo_inverse": {
      "description": "whether it has a version of its logo for dark backgrounds",
      "type": "boolean"
    },
    "logo": {
      "description": "the path of the logo, from the root of the domain; without a logo, it shows the name of the organization",
      "type": "string"
    },
    "logo_inverse": {
      "description": "the path of the logo for dark backgrounds, such as the primary color; without such a version, it shows the name of the organization in light letters",
      "type": "string"
    },
    "message": {
      "description": "what to do when the house style is inactive",
      "type": "string"
    },
    "name": {
      "description": "the name of the organization; the domain of the Cahut until an administrator sets one",
      "type": "string"
    },
    "neutral": {
      "description": "whether the Cahut has the neutral house style, which no administrator has changed",
      "type": "boolean"
    },
    "own_chart_colors": {
      "description": "whether the house style sets the colors of its charts, rather than Cahut deriving them from its primary and secondary colors",
      "type": "boolean"
    },
    "previous": {
      "additionalProperties": false,
      "description": "the previous house style, which rollback_house_style restores; absent when there is none",
      "properties": {
        "neutral": {
          "description": "whether it is the neutral house style",
          "type": "boolean"
        },
        "set_at": {
          "type": "string"
        },
        "set_by": {
          "type": "string"
        }
      },
      "required": [
        "neutral",
        "set_at"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "set_at": {
      "description": "when the house style was set; absent for the neutral house style",
      "type": [
        "null",
        "string"
      ]
    },
    "set_by": {
      "description": "the handle of the person who set it, or for whom an AI set it",
      "type": "string"
    },
    "stylesheet": {
      "description": "the style sheet of the house style, at this path from the root of the domain: link it with \u003clink rel=\"stylesheet\" href=\"/-/house-style.css\"\u003e. The Window declares nothing for it",
      "type": "string"
    },
    "template": {
      "additionalProperties": false,
      "description": "the template of the kind asked for",
      "properties": {
        "html": {
          "description": "index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title",
          "type": "string"
        },
        "kind": {
          "description": "the kind of Window it makes",
          "type": "string"
        },
        "scripts": {
          "description": "whether it runs scripts, which the Window then declares",
          "type": "boolean"
        }
      },
      "required": [
        "kind",
        "html",
        "scripts"
      ],
      "type": [
        "null",
        "object"
      ]
    },
    "warnings": {
      "description": "what may not read well, and what to change",
      "items": {
        "type": "string"
      },
      "type": [
        "null",
        "array"
      ]
    }
  },
  "required": [
    "name",
    "active",
    "neutral",
    "stylesheet",
    "logo",
    "logo_inverse",
    "has_logo",
    "has_logo_inverse",
    "deck_engine",
    "colors",
    "own_chart_colors",
    "chart_colors",
    "fonts"
  ],
  "type": "object"
}
```

</details>

## `set_room_home`

Show a Window of a Room at the Room's address, &lt;domain&gt;/&lt;room&gt;/, in place of the page that lists its Windows, such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. An empty window goes back to the list. The Window keeps its own address and its access.

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
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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

Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. Knock first when it makes a Window public: a person approves the call before it runs.

Knock first when it makes a Window public: that call of an AI makes a Knock, which the person it acts for approves ([Operations](/reference/operations/#knock-first)).

Title: Set who opens a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `access` | string | yes | room, internal, key or public |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |
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
      "description": "room, internal, key or public",
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
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `key` | string | no | the new Key, returned this once: give it to the person, who passes it on with the address |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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

Put a Window in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was.

Title: Put a Window in the trash. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "window": {
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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

Change the title of a Room. Its name, and so its address, stays.

Title: Retitle a Room. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `room` | string | yes | the Room's name |
| `title` | string | yes | the new title |

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
    "title": {
      "description": "the new title",
      "type": "string"
    }
  },
  "required": [
    "room",
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
| `people` | array of objects or null | yes | the people in the Room, owners first |
| `people[].handle` | string | yes |  |
| `people[].name` | string | yes |  |
| `people[].role` | string | yes | owner or member of the Room |
| `title` | string | yes |  |
| `windows` | integer | yes | how many Windows the Room holds, out of the trash |
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
      "description": "how many Windows the Room holds, out of the trash",
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

Publish a new version of a Window: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted.

Title: Update a Window. Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `declaration` | object or null | yes | what the pages do: Cahut lets them load and call the origins listed here, and nothing else |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `files` | array of objects or null | yes | every file of the new version, as for publish_window; files left out are not in it |
| `files[].content` | string | yes |  |
| `files[].path` | string | yes | relative path with / as separator, such as index.html or img/logo.svg |
| `files[].encoding` | string | no | utf-8 (the default) or base64, for binary files |
| `window` | string | yes | the Window's address, or its Room and its name as room/name |
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
      "description": "the Window's address, or its Room and its name as room/name",
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
| `access` | string | yes | who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was first published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files of the current version, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes | the number of the version that readers see |
| `access_ends_at` | string or null | no | when the access ends, after which the Window opens to its Room only; absent when it does not end |
| `declaration` | object or null | no | what the pages of the current version do, as declared when it was published; absent for a Window published before declarations |
| `declaration.scripts` | boolean | yes | whether the pages run scripts, inline or in files |
| `declaration.forms` | array of strings or null | no | the origins that the forms of the pages send to, if any |
| `declaration.origins` | array of strings or null | no | every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://\*.example.com for the subdomains of example.com. The files of the Window need none |
| `previous_version` | integer | no | the number of the version that rollback_window goes back to; absent when there is none |
| `trashed_at` | string or null | no | when the Window went to the trash; absent when it is not in it |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "access": {
      "description": "who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public",
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
      "description": "handle of the person it was first published for",
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
          "description": "every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none",
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
    "file_count": {
      "type": "integer"
    },
    "kind": {
      "type": "string"
    },
    "name": {
      "description": "the Window's name, last part of its address",
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
    }
  },
  "required": [
    "name",
    "title",
    "kind",
    "url",
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
