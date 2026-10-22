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

> This Cahut keeps what AIs make, on the user's own server. Publish a document, deck, dashboard or site with publish_window: give its HTML and its other files, and you get back the address where it opens; give that address to the user. Make index.html a complete page that works on its own. Link its files with relative paths, such as css/deck.css, never /css/deck.css: a Window is served under the address of its Room. Pages run in a sandbox where localStorage, sessionStorage, IndexedDB and cookies throw errors: keep state in memory or in the address. Every call is recorded in the Logbook of the Cahut.

## Limits

| Limit | Value |
| --- | --- |
| Files in one Window | 500 |
| Size of the files of one Window, decoded | 32 MiB |
| Size of one request to `/mcp` | about 43.7 MiB, room for the files in base64 |

## Tools

| Tool | What it does |
| --- | --- |
| [`create_user`](#create_user) | Create a person, with their personal Room, and the link where they choose their password. |
| [`disable_user`](#disable_user) | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. |
| [`enable_user`](#enable_user) | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. |
| [`get_account`](#get_account) | Describe the person you act for: their handle, name, role in the Cahut and Rooms. |
| [`get_window`](#get_window) | Describe a Window and list its files, optionally with their contents. |
| [`list_logbook`](#list_logbook) | Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, and for anyone else what they did and what AIs did on their behalf. |
| [`list_users`](#list_users) | List the people of the Cahut, with their handles, names, roles and states. |
| [`list_windows`](#list_windows) | List the Windows in the Rooms of the person you act for, most recently updated first. |
| [`publish_window`](#publish_window) | Publish a new Window: a document, deck, dashboard or site made of HTML and its files, in a Room of the person you act for, by default their personal Room. Returns the address where it opens. |
| [`set_user_role`](#set_user_role) | Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator. |

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
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was published for |
| `file_count` | integer | yes |  |
| `files` | array of objects or null | yes |  |
| `files[].path` | string | yes |  |
| `files[].size` | integer | yes |  |
| `files[].type` | string | yes | media type |
| `files[].content` | string | no |  |
| `files[].encoding` | string | no | utf-8 or base64 |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes |  |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was published for",
      "type": "string"
    },
    "file_count": {
      "type": "integer"
    },
    "files": {
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
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files, in bytes",
      "type": "integer"
    },
    "title": {
      "type": "string"
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
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
| `entries[].via` | string | yes | the adapter the call came through: cli, mcp |
| `entries[].client` | string | no | the AI, as its token and its software name it |
| `entries[].detail` | string | no |  |
| `entries[].error` | string | no |  |
| `entries[].target` | string | no | what the operation acted on |
| `entries[].token_id` | integer | no |  |
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
            "description": "the AI, as its token and its software name it",
            "type": "string"
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
            "description": "the adapter the call came through: cli, mcp",
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

List the Windows in the Rooms of the person you act for, most recently updated first.

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
| `windows` | array of objects or null | yes |  |
| `windows[].created_at` | string | yes |  |
| `windows[].created_by` | string | yes | handle of the person it was published for |
| `windows[].file_count` | integer | yes |  |
| `windows[].kind` | string | yes |  |
| `windows[].name` | string | yes | the Window's name, last part of its address |
| `windows[].room` | string | yes | the name of its Room, the part of its address before its name |
| `windows[].size` | integer | yes | total size of the files, in bytes |
| `windows[].title` | string | yes |  |
| `windows[].updated_at` | string | yes |  |
| `windows[].url` | string | yes | where the Window opens |
| `windows[].version` | integer | yes |  |

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
          "created_at": {
            "type": "string"
          },
          "created_by": {
            "description": "handle of the person it was published for",
            "type": "string"
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
          "room": {
            "description": "the name of its Room, the part of its address before its name",
            "type": "string"
          },
          "size": {
            "description": "total size of the files, in bytes",
            "type": "integer"
          },
          "title": {
            "type": "string"
          },
          "updated_at": {
            "type": "string"
          },
          "url": {
            "description": "where the Window opens",
            "type": "string"
          },
          "version": {
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
          "updated_at"
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
    "files"
  ],
  "type": "object"
}
```

</details>

### Result

| Field | Type | Always present | Description |
| --- | --- | --- | --- |
| `created_at` | string | yes |  |
| `created_by` | string | yes | handle of the person it was published for |
| `file_count` | integer | yes |  |
| `kind` | string | yes |  |
| `name` | string | yes | the Window's name, last part of its address |
| `room` | string | yes | the name of its Room, the part of its address before its name |
| `size` | integer | yes | total size of the files, in bytes |
| `title` | string | yes |  |
| `updated_at` | string | yes |  |
| `url` | string | yes | where the Window opens |
| `version` | integer | yes |  |

<details>
<summary>JSON Schema</summary>

```json
{
  "additionalProperties": false,
  "properties": {
    "created_at": {
      "type": "string"
    },
    "created_by": {
      "description": "handle of the person it was published for",
      "type": "string"
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
    "room": {
      "description": "the name of its Room, the part of its address before its name",
      "type": "string"
    },
    "size": {
      "description": "total size of the files, in bytes",
      "type": "integer"
    },
    "title": {
      "type": "string"
    },
    "updated_at": {
      "type": "string"
    },
    "url": {
      "description": "where the Window opens",
      "type": "string"
    },
    "version": {
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
    "updated_at"
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
