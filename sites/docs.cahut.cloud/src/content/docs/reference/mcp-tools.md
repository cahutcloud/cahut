---
title: "MCP tools"
description: "The tools that a Cahut offers to AIs over MCP, with their arguments, results and JSON Schemas."
sidebar:
  order: 1
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Each Cahut serves MCP at `https://<domain>/mcp`, over Streamable HTTP. An AI connects with a personal token of that Cahut, which it sends in the header `Authorization: Bearer <token>`: see [Connect an AI](/connect/claude-code/).

The tools are the operations of the Cahut that are open to AIs ([Operations](/reference/operations/)). An AI sees the tools that the person it acts for may use: some are for the AIs of administrators of the Cahut only. The operations that prove who someone is, such as signing in, are for people only. Every call goes through the registry of operations, which checks it, runs it and writes it to the Logbook, reads included. Arguments with unknown fields are refused.

A tool returns its result as structured content, and as the same JSON in a text block. When a call fails because of its arguments, the error says what to change. Any other error reads only as an internal error: the administrator of the Cahut finds the details in its Logbook.

## Instructions

The Cahut gives these instructions to the AI when it connects:

> This Cahut keeps what AIs make, on the user's own server. Publish a document, deck, dashboard or site with publish_window: give its HTML and its other files, and you get back the address where it opens; give that address to the user. Make index.html a complete page that works on its own. Link its files with relative paths, such as css/deck.css, never /css/deck.css: a Window is served under the address of its Room. Declare what the pages do: whether they run scripts, every origin they load from or call, such as a CDN or a font service, and where their forms send. They can load and call nothing else. Pages run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: keep state in memory or in the address. To change a Window, publish its new version with update_window at the same address; rollback_window goes back to the previous one. A Window opens to the members of its Room until set_window_access opens it to the whole Cahut or to anyone with its Key; making it public waits for a person. Every call is recorded in the Logbook of the Cahut.

## Limits

| Limit | Value |
| --- | --- |
| Files in one Window | 500 |
| Size of the files of one Window, decoded | 32 MiB |
| Size of one request to `/mcp` | about 43.7 MiB, room for the files in base64 |

## Tools

| Tool | What it does |
| --- | --- |
| [`add_room_member`](#add_room_member) | Add a person to a team Room, as a member or an owner. They can then open and publish its Windows. |
| [`create_room`](#create_room) | Create a team Room, at &lt;domain&gt;/&lt;name&gt;/. You become its owner, and can then add members. Windows published in it are for its members. |
| [`create_user`](#create_user) | Create a person, with their personal Room, and the link where they choose their password. |
| [`disable_user`](#disable_user) | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. |
| [`enable_user`](#enable_user) | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. |
| [`get_account`](#get_account) | Describe the person you act for: their handle, name, role in the Cahut and Rooms. |
| [`get_cahut`](#get_cahut) | Describe this Cahut: its domain, the address where its Rooms and Windows open, and that of its MCP endpoint. |
| [`get_room`](#get_room) | Describe a Room: its title, its owners and members, and how many Windows it holds. |
| [`get_window`](#get_window) | Describe a Window and list its files, optionally with their contents. |
| [`list_logbook`](#list_logbook) | Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. |
| [`list_rooms`](#list_rooms) | List your Rooms, your personal Room first, with your role and how many members and Windows each has. |
| [`list_users`](#list_users) | List the people of the Cahut, with their handles, names, roles and states. |
| [`list_windows`](#list_windows) | List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first. |
| [`publish_window`](#publish_window) | Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. |
| [`remove_room_member`](#remove_room_member) | Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner. |
| [`restore_window`](#restore_window) | Bring a Window back from the trash, with its versions. |
| [`rollback_window`](#rollback_window) | Go back to the previous version of a Window. The version you leave becomes the previous one, so calling rollback_window again returns to it. |
| [`set_room_member_role`](#set_room_member_role) | Make a person of a team Room an owner or a member of it. A Room keeps at least one owner. |
| [`set_user_role`](#set_user_role) | Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. |
| [`set_window_access`](#set_window_access) | Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person. |
| [`trash_window`](#trash_window) | Put a Window in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was. |
| [`update_room`](#update_room) | Change the title of a Room. Its name, and so its address, stays. |
| [`update_window`](#update_window) | Publish a new version of a Window: give all its files again, as for publish_window. Its address stays. Its current version becomes the previous one, which rollback_window goes back to; the version before that is deleted. |

## `add_room_member`

Add a person to a team Room, as a member or an owner. They can then open and publish its Windows.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

Create a team Room, at &lt;domain&gt;/&lt;name&gt;/. You become its owner, and can then add members. Windows published in it are for its members.

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

Create a person, with their personal Room, and the link where they choose their password.

Offered to the AIs of administrators of the Cahut only.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

### Arguments

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `handle` | string | yes |  |
| `name` | string | yes |  |
| `admin` | boolean | no |  |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "admin": {
      "type": "boolean"
    },
    "handle": {
      "type": "string"
    },
    "name": {
      "type": "string"
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

## `disable_user`

Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay.

Offered to the AIs of administrators of the Cahut only.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
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

Enable a disabled person again: they can sign in, and their AIs can act with their tokens.

Offered to the AIs of administrators of the Cahut only.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
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

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
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

Describe this Cahut: its domain, the address where its Rooms and Windows open, and that of its MCP endpoint.

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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
| `mcp` | string | no | the address of its MCP endpoint, where AI clients connect |
| `url` | string | no | the address of the Cahut, under which its Rooms and Windows open |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "domain": {
      "type": "string"
    },
    "mcp": {
      "description": "the address of its MCP endpoint, where AI clients connect",
      "type": "string"
    },
    "url": {
      "description": "the address of the Cahut, under which its Rooms and Windows open",
      "type": "string"
    }
  },
  "required": [
    "domain"
  ],
  "type": "object"
}
```

</details>

## `get_room`

Describe a Room: its title, its owners and members, and how many Windows it holds.

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

## `list_logbook`

Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf.

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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
| `entries[].outcome` | string | yes | ok, refused or failed |
| `entries[].via` | string | yes | the adapter the call came through: cli, api, mcp or oauth |
| `entries[].client` | string | no | the AI, as its token or its connection, and its software, name it |
| `entries[].connection_id` | integer | no | the connection of the AI, when it connected over OAuth |
| `entries[].detail` | string | no |  |
| `entries[].error` | string | no |  |
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
          "operation": {
            "type": "string"
          },
          "outcome": {
            "description": "ok, refused or failed",
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

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

List the people of the Cahut, with their handles, names, roles and states.

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: true`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

## `remove_room_member`

Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. A Room keeps at least one owner.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

## `rollback_window`

Go back to the previous version of a Window. The version you leave becomes the previous one, so calling rollback_window again returns to it.

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

## `set_room_member_role`

Make a person of a team Room an owner or a member of it. A Room keeps at least one owner.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

## `set_user_role`

Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator.

Offered to the AIs of administrators of the Cahut only.

Knock first: an AI needs the approval of a person to run it ([Operations](/reference/operations/#knock-first)).

Annotations: `readOnlyHint: false`, `destructiveHint: true`, `openWorldHint: false`.

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

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
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

Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut; key, anyone who types its Key; or public, anyone with its address. The access may end at a date, after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, and returns it once: give it to the person. Making a Window public waits for the approval of a person.

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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

Annotations: `readOnlyHint: false`, `destructiveHint: false`, `openWorldHint: false`.

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
