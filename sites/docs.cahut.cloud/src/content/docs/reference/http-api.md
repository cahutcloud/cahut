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

A refusal that people meet in the console gives its code too, which never changes, and the values of its message, as in `{"error": "invalid input: the password has 9 characters: use at least 15", "code": "password_too_short", "values": {"length": 9, "limit": 15}}`. The console says it in the language of its person. A program tells these refusals apart by their code: the wording of a message may change.

| Code | Values | Message |
| --- | --- | --- |
| `account_disabled` |  | This account is disabled: an administrator of the Cahut can enable it |
| `activation_link_gone` |  | This activation link is unknown, used or expired: ask an administrator of the Cahut for a new one |
| `cahut_needs_an_admin` |  | The Cahut needs an administrator: make another person one first |
| `cahut_suspended` |  | This Cahut is suspended by the administrators of its server: it keeps everything it holds, and nothing in it can be read or changed until they resume it. Ask the administrators of the Cahut, or, for an AI, tell the person you act for |
| `description_control` |  | The description holds a control character: write it in plain text |
| `description_too_long` | `length`, `limit` | The description has {length} characters: use at most {limit}, a sentence or two |
| `email_already_yours` | `email` | {email} is already your e-mail address |
| `email_link_gone` |  | This link is unknown, used or expired: change your address again from your account |
| `email_taken` | `email` | {email} is already the e-mail address of another person of this Cahut |
| `handle_invalid` | `handle` | Handle {handle}: use 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit |
| `handle_knocks_waiting` |  | Knocks wait for you: answer them before you change your handle, since they name the addresses of your personal Room |
| `handle_reserved` | `handle` | Handle {handle} is reserved: it names a page of every Cahut |
| `handle_room_cannot_move` | `room`, `name` | Your Room {room} would be named {name}, which another Room holds or which is too long: choose another handle |
| `handle_taken` | `handle` | Handle {handle} is taken, by a person or a Room |
| `invitation_gone` |  | This invitation does not open the Window any more: it was taken back, it expired, or the Window changed its Key or its access. Ask the person who invited you for a new one |
| `invite_address_internal` | `room`, `window` | {room}/{window} opens to the people of the Cahut only: invite people of the Cahut by handle, or set its access to key first, so that each address gets a personal link |
| `invite_address_room_only` | `room`, `window` | {room}/{window} opens to its Room only: invite people of the Cahut by handle, or set its access to key first, so that each address gets a personal link |
| `invite_bad_address` | `to` | To: {to} is not an e-mail address: give one address alone, such as jane@example.com |
| `invite_by_address` | `to` | {to}: in this Cahut, which is individual, invite by e-mail address, to a Window with a Key or public |
| `invite_disabled` | `handle` | {handle} is disabled, and cannot read the Window |
| `invite_no_person` | `to` | No person {to} in this Cahut: invite someone else by e-mail address |
| `invite_not_in_room` | `handle`, `room`, `window` | {handle} is not in the Room {room}, and {room}/{window} opens to its Room only: add them to the Room, or set its access to internal or key first |
| `key_has_spaces` |  | The Key starts or ends with a space |
| `key_too_long` | `limit` | The Key has more than {limit} characters |
| `key_too_short` | `length`, `limit` | The Key has {length} characters: give it {limit} at least, or let Cahut draw one |
| `legal_link_in_french_only` | `link`, `page` | {link}: give {page} too, the address of the page for every other language; the address in French takes its place in French only |
| `legal_link_invalid` | `link` | {link}: give the address of a page, which starts with https://, such as https://example.org/terms, or leave it out to remove the link |
| `name_empty` |  | Name is empty |
| `name_too_long` |  | Name is longer than 200 bytes |
| `no_mail` |  | This Cahut sends no e-mail: an administrator of the Cahut sets its relay first, on the page E-mail of the console |
| `no_mail_to_invite` |  | This Cahut sends no e-mail: give them the address of the Window yourself, with its Key if it has one |
| `no_password_link` |  | This Cahut sends no e-mail: an administrator of the Cahut can give you a new link |
| `password_too_long` | `length`, `limit` | The password has {length} characters: use at most {limit} |
| `password_too_short` | `length`, `limit` | The password has {length} characters: use at least {limit} |
| `room_name_invalid` | `name` | Name {name}: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit |
| `room_name_reserved` | `name` | Name {name} is reserved: it names a page of every Cahut |
| `room_name_taken` | `name` | Name {name} is taken, by a Room or a person: choose another |
| `room_name_unavailable` | `name` | Name {name} is not available: choose another |
| `room_name_yours` | `name` | You already have a Room {name}: choose another name |
| `room_needs_an_owner` | `room` | The Room {room} needs an owner: make another person one first |
| `server_account_disabled` |  | This account is disabled: another administrator of the server can enable it |
| `server_activation_link_gone` |  | This activation link is unknown, used or expired: ask another administrator of the server for a new one, or run cahut admin link on the server |
| `session_ended` |  | Invalid or ended session |
| `setup_done` | `domain` | The Cahut of {domain} has people already: sign in at its console |
| `setup_link_gone` |  | This link to the first start does not work any more: it was used, it expired, or it is for another address; cahut setup, on the server, gives a new one |
| `share_link_end_past` | `ends_at`, `latest` | Ends_at {ends_at} is past, and a share link ends between now and {latest}: give an end in between, or none for 7 days |
| `share_link_ended` | `link` | Share link {link} has ended already: it opens nothing more |
| `share_link_label` | `limit` | Label: write it on one line, in {limit} characters at most |
| `share_link_openings` | `given`, `limit` | Max_openings {given}: give a number from 1 to {limit}, or none for a link that opens to any number of readers |
| `share_link_public` | `room`, `window` | {room}/{window} is public: its permanent address opens it to anyone, so share that address rather than a link |
| `share_link_too_long` | `ends_at`, `latest` | Ends_at {ends_at}: a share link lasts 30 days at most, until {latest}: give an end before it, or none for 7 days |
| `title_empty` |  | Title is empty: give one |
| `title_too_long` |  | Title is longer than 200 bytes: shorten it |
| `too_many_attempts_from_address` |  | Too many attempts from this address: wait a minute |
| `too_many_invitations` | `sent`, `handle`, `limit` | {sent} invitations went out for {handle} in the last 24 hours, and {limit} is the most: send the rest later |
| `too_many_refused_sign_ins` | `minutes` | Too many refused sign-ins for this handle: try again in {minutes} minutes |
| `too_many_share_links` | `room`, `window`, `limit` | {room}/{window} has {limit} share links that stand, the most: take one back before you make another |
| `too_many_share_links_today` | `made`, `handle`, `limit` | {made} share links were made for {handle} in the last 24 hours, and {limit} is the most: make the rest later |
| `too_many_sign_ins_at_once` |  | Too many sign-ins at once: try again in a few seconds |
| `window_name_invalid` | `name` | Name {name}: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit |
| `window_name_taken` | `name`, `room` | Name {name} is taken in the Room {room}, by a Window or one in its trash: choose another |
| `wrong_current_password` |  | The current password is wrong |
| `wrong_handle_or_password` |  | Wrong handle or password |

## Operations of a Cahut

| Route | Who calls it | Session |
| --- | --- | --- |
| `POST /api/activate_account` | Anyone, to sign in | no |
| `POST /api/add_room_member` | Any person | yes |
| `POST /api/approve_authorization` | Any person | yes |
| `POST /api/approve_knock` | Any person | yes |
| `POST /api/change_email` | Any person | yes |
| `POST /api/change_handle` | Any person | yes |
| `POST /api/change_name` | Any person | yes |
| `POST /api/change_password` | Any person | yes |
| `POST /api/check_handle` | Administrators of the Cahut | yes |
| `POST /api/confirm_email` | Anyone, to sign in | no |
| `POST /api/create_activation_link` | Administrators of the Cahut | yes |
| `POST /api/create_room` | Any person | yes |
| `POST /api/create_service_token` | Administrators of the server | yes |
| `POST /api/create_share_link` | Any person | yes |
| `POST /api/create_token` | Any person | yes |
| `POST /api/create_user` | Administrators of the Cahut | yes |
| `POST /api/delete_room` | Any person | yes |
| `POST /api/delete_user` | Administrators of the Cahut | yes |
| `POST /api/deny_authorization` | Any person | yes |
| `POST /api/deny_knock` | Any person | yes |
| `POST /api/describe_authorization` | Any person | yes |
| `POST /api/describe_sign_in` | Anyone, to sign in | no |
| `POST /api/disable_user` | Administrators of the Cahut | yes |
| `POST /api/disconnect_ai` | Any person | yes |
| `POST /api/enable_user` | Administrators of the Cahut | yes |
| `POST /api/get_account` | Any person | yes |
| `POST /api/get_cahut` | Any person | yes |
| `POST /api/get_invitation` | Anyone, to sign in | no |
| `POST /api/get_knock` | Any person | yes |
| `POST /api/get_mail_relay` | Administrators of the Cahut | yes |
| `POST /api/get_room` | Any person | yes |
| `POST /api/get_user` | Administrators of the Cahut | yes |
| `POST /api/get_window` | Any person | yes |
| `POST /api/invite_to_window` | Any person | yes |
| `POST /api/list_connected_ais` | Any person | yes |
| `POST /api/list_invitations` | Any person | yes |
| `POST /api/list_knocks` | Any person | yes |
| `POST /api/list_logbook` | Any person | yes |
| `POST /api/list_rooms` | Any person | yes |
| `POST /api/list_service_tokens` | Administrators of the server | yes |
| `POST /api/list_tokens` | Any person | yes |
| `POST /api/list_users` | Any person | yes |
| `POST /api/list_windows` | Any person | yes |
| `POST /api/open_invitation` | Anyone, to sign in | no |
| `POST /api/open_share_link` | Anyone, to sign in | no |
| `POST /api/preview_windows` | Any person | yes |
| `POST /api/publish_window` | Any person | yes |
| `POST /api/purge_window` | Any person | yes |
| `POST /api/remove_room_member` | Any person | yes |
| `POST /api/rename_window` | Any person | yes |
| `POST /api/request_password_link` | Anyone, to sign in | no |
| `POST /api/restore_window` | Any person | yes |
| `POST /api/resume_cahut` | Administrators of the server | yes |
| `POST /api/retitle_window` | Any person | yes |
| `POST /api/revoke_invitation` | Any person | yes |
| `POST /api/revoke_service_token` | Administrators of the server | yes |
| `POST /api/revoke_share_link` | Any person | yes |
| `POST /api/revoke_token` | Any person | yes |
| `POST /api/rollback_window` | Any person | yes |
| `POST /api/send_test_mail` | Administrators of the Cahut | yes |
| `POST /api/set_ai_reach` | Any person | yes |
| `POST /api/set_cahut_settings` | Administrators of the server | yes |
| `POST /api/set_legal_links` | Administrators of the server | yes |
| `POST /api/set_mail_notices` | Any person | yes |
| `POST /api/set_mail_relay` | Administrators of the Cahut | yes |
| `POST /api/set_room_home` | Any person | yes |
| `POST /api/set_room_member_role` | Any person | yes |
| `POST /api/set_user_email` | Administrators of the Cahut | yes |
| `POST /api/set_user_role` | Administrators of the Cahut | yes |
| `POST /api/set_window_access` | Any person | yes |
| `POST /api/sign_in` | Anyone, to sign in | no |
| `POST /api/sign_in_to_windows` | Any person | yes |
| `POST /api/sign_out` | Any person | yes |
| `POST /api/suspend_cahut` | Administrators of the server | yes |
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
| `POST /api/server/disable_admin` | Administrators of the server | yes |
| `POST /api/server/enable_admin` | Administrators of the server | yes |
| `POST /api/server/get_setup` | Anyone, to sign in | no |
| `POST /api/server/get_updates` | Administrators of the server | yes |
| `POST /api/server/list_admins` | Administrators of the server | yes |
| `POST /api/server/list_domains` | Administrators of the server | yes |
| `POST /api/server/list_logbook` | Administrators of the server | yes |
| `POST /api/server/rename_admin` | Administrators of the server | yes |
| `POST /api/server/rename_domain` | Administrators of the server | yes |
| `POST /api/server/request_update` | Administrators of the server | yes |
| `POST /api/server/set_up` | Anyone, to sign in | no |
| `POST /api/server/sign_in` | Anyone, to sign in | no |
| `POST /api/server/sign_out` | Administrators of the server | yes |

## Admin API

A program on the server, such as a site where people sign up, manages the people of a Cahut through the admin API, with a service token, and without a Knock. The API is off unless `cahut serve --admin-api <address>`, or `CAHUT_ADMIN_API`, gives it an address of the machine itself, such as `127.0.0.1:8090`: Cahut refuses to start with another one, and the API refuses every client that is not on the machine.

| Route | What it runs |
| --- | --- |
| `POST /api/<operation>` | An operation of the Cahut that the host of the request names, among those open to services |

Calls, inputs, outputs and errors are those of the API of the console, with a service token in the header `Authorization: Bearer <token>` in place of a session. A service token starts with `cahut_svc_`. It belongs to one Cahut, and opens nothing else: neither the console, nor MCP, nor the Windows. The administrators of the server create one with `cahut service-token create`, see them with `cahut service-token list` and revoke one with `cahut service-token revoke`, which ends its access at once.

```bash
curl -X POST http://127.0.0.1:8090/api/check_handle \
  -H "Host: cahut.example.com" -H "Authorization: Bearer <service token>" -d '{"handle": "ada"}'
```

| Route | What it does |
| --- | --- |
| `POST /api/check_handle` | Say whether a handle can go to a new person: of the form of a handle, not reserved, and held by no person and by no Room. For the administrators of the Cahut. |
| `POST /api/create_activation_link` | Give a person a new single-use link, valid 7 days, where they choose their password, for instance after they forgot it. It goes to their e-mail address too, when they have one and the Cahut sends e-mail. Their password works until they use the link. |
| `POST /api/create_user` | Create a person, with their personal Room, and the link where they choose their password, which goes to their e-mail address when you give one and the Cahut sends e-mail. |
| `POST /api/disable_user` | Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. Their Rooms and Windows stay. |
| `POST /api/enable_user` | Enable a disabled person again: they can sign in, and their AIs can act with their tokens. |
| `POST /api/get_cahut` | Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, whether it sends e-mail, the links to its legal pages, its kind, collaborative or individual, and its limits with what it holds: a call that would cross a limit is refused. |
| `POST /api/get_user` | Describe one person of the Cahut, as list_users does, with their e-mail address. For the administrators of the Cahut. |
| `POST /api/set_user_email` | Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. Their old address gets a notice. For the administrators of the Cahut. |
