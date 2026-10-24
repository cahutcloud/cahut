---
title: "HTTP API"
description: "The API of the console: the operations of a Cahut and of its server over HTTP, with a session."
sidebar:
  order: 4
---

<!-- Generated from the code by make docs-reference: do not edit. -->

Each Cahut serves the API of its console under `/console/api/` of its domain, next to the console itself at `/console/`. The console uses this API and nothing else, and AIs reach the same operations over [MCP](/reference/mcp-tools/): there is one registry of [operations](/reference/operations/).

## Calls

| Route | What it runs |
| --- | --- |
| `POST /console/api/<operation>` | An operation of the Cahut of the domain |
| `POST /console/api/server/<operation>` | An operation of the server, for its administrators |

The body of the request is the input of the operation, in JSON, and the body of the response its output, with the same names and schemas as the MCP tools. Fields that the operation does not know are refused. The API takes `POST` only, reads included, and answers no CORS preflight.

```bash
curl -X POST https://cahut.example.com/console/api/list_rooms \
  -H "Authorization: Bearer <session token>" -d '{}'
```

## Sessions

The operations that sign in take no session: `sign_in` and `activate_account`, at both routes. They return a session token, which every other call sends in the header `Authorization: Bearer <token>`. The API reads no cookie.

| Token | Whose session | Where it holds |
| --- | --- | --- |
| `cahut_ses_…` | A person of the Cahut | The Cahut that holds the person |
| `cahut_adm_…` | An administrator of the server | The domain where it opened, for the server and for that Cahut |

A session ends 7 days after its last use, and 30 days after it opened in any case. Personal tokens, which start with `cahut_pat_`, are for AIs over MCP: the API refuses them.

## Errors

A failed call answers with a status and a body such as `{"error": "invalid input: title is empty"}`. The message says what to change. An internal error says only that it happened: the Logbook keeps the details.

| Status | When |
| --- | --- |
| 400 Bad Request | The input is invalid |
| 401 Unauthorized | The session is missing, unknown or ended, or holds at another domain |
| 403 Forbidden | The caller may not do this |
| 404 Not Found | The operation, or what it names, does not exist, or the caller may not see it |
| 405 Method Not Allowed | The request is not a `POST` |
| 409 Conflict | The operation needs the approval of a person, which only AIs meet |
| 413 Request Entity Too Large | The body is larger than 1 MiB |
| 421 Misdirected Request | The domain is not that of a Cahut of the server |
| 429 Too Many Requests | Too many sign-ins: from one address, 10 at once, then one every 6 seconds; for one handle, 10 refused from one address or 100 from everywhere in 15 minutes. An IPv6 address counts with its /64 |
| 500 Internal Server Error | Anything else |

## Operations of a Cahut

| Route | Who calls it | Session |
| --- | --- | --- |
| `POST /api/activate_account` | Anyone, to sign in | no |
| `POST /api/add_room_member` | Any person | yes |
| `POST /api/approve_authorization` | Any person | yes |
| `POST /api/approve_knock` | Any person | yes |
| `POST /api/change_password` | Any person | yes |
| `POST /api/create_activation_link` | Administrators of the Cahut | yes |
| `POST /api/create_room` | Any person | yes |
| `POST /api/create_token` | Any person | yes |
| `POST /api/create_user` | Administrators of the Cahut | yes |
| `POST /api/delete_room` | Any person | yes |
| `POST /api/deny_authorization` | Any person | yes |
| `POST /api/deny_knock` | Any person | yes |
| `POST /api/describe_authorization` | Any person | yes |
| `POST /api/disable_user` | Administrators of the Cahut | yes |
| `POST /api/disconnect_ai` | Any person | yes |
| `POST /api/enable_user` | Administrators of the Cahut | yes |
| `POST /api/get_account` | Any person | yes |
| `POST /api/get_cahut` | Any person | yes |
| `POST /api/get_house_style` | Any person | yes |
| `POST /api/get_knock` | Any person | yes |
| `POST /api/get_room` | Any person | yes |
| `POST /api/get_window` | Any person | yes |
| `POST /api/list_connected_ais` | Any person | yes |
| `POST /api/list_knocks` | Any person | yes |
| `POST /api/list_logbook` | Any person | yes |
| `POST /api/list_rooms` | Any person | yes |
| `POST /api/list_tokens` | Any person | yes |
| `POST /api/list_users` | Any person | yes |
| `POST /api/list_windows` | Any person | yes |
| `POST /api/preview_windows` | Any person | yes |
| `POST /api/publish_window` | Any person | yes |
| `POST /api/purge_window` | Any person | yes |
| `POST /api/remove_room_member` | Any person | yes |
| `POST /api/restore_window` | Any person | yes |
| `POST /api/revoke_token` | Any person | yes |
| `POST /api/rollback_house_style` | Administrators of the Cahut | yes |
| `POST /api/rollback_window` | Any person | yes |
| `POST /api/set_ai_reach` | Any person | yes |
| `POST /api/set_cahut_settings` | Administrators of the server | yes |
| `POST /api/set_house_style` | Administrators of the Cahut | yes |
| `POST /api/set_room_member_role` | Any person | yes |
| `POST /api/set_user_role` | Administrators of the Cahut | yes |
| `POST /api/set_window_access` | Any person | yes |
| `POST /api/sign_in` | Anyone, to sign in | no |
| `POST /api/sign_in_to_windows` | Any person | yes |
| `POST /api/sign_out` | Any person | yes |
| `POST /api/trash_window` | Any person | yes |
| `POST /api/update_room` | Any person | yes |
| `POST /api/update_window` | Any person | yes |

## Operations of the server

| Route | Who calls it | Session |
| --- | --- | --- |
| `POST /api/server/activate_account` | Anyone, to sign in | no |
| `POST /api/server/add_domain` | Administrators of the server | yes |
| `POST /api/server/change_password` | Administrators of the server | yes |
| `POST /api/server/create_activation_link` | Administrators of the server | yes |
| `POST /api/server/create_admin` | Administrators of the server | yes |
| `POST /api/server/create_setup_link` | Administrators of the server | yes |
| `POST /api/server/get_setup` | Anyone, to sign in | no |
| `POST /api/server/get_updates` | Administrators of the server | yes |
| `POST /api/server/list_admins` | Administrators of the server | yes |
| `POST /api/server/list_domains` | Administrators of the server | yes |
| `POST /api/server/list_logbook` | Administrators of the server | yes |
| `POST /api/server/rename_domain` | Administrators of the server | yes |
| `POST /api/server/request_update` | Administrators of the server | yes |
| `POST /api/server/set_up` | Anyone, to sign in | no |
| `POST /api/server/sign_in` | Anyone, to sign in | no |
| `POST /api/server/sign_out` | Administrators of the server | yes |
