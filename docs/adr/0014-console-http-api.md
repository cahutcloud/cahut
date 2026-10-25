# 14. The console calls each operation at its own HTTP route

- Status: Accepted, superseded in part by [ADR 20](0020-console-under-a-path-of-port-443.md) (the API is under `/console/api/` on port 443), extended by [ADR 37](0037-refusals-with-a-code-for-the-console.md) (the body of a refusal that people meet gives its code and its values)
- Date: 2026-10-22

## Context

The console calls only the public API, which is the registry of operations ([ADR 1](0001-operation-registry.md), [ADR 5](0005-preact-console-embedded.md)). It is served on port 8443 of every Cahut's domain, and it uses no cookie and no CORS ([ADR 11](0011-console-on-its-own-port.md)). MCP already offers the operations to AIs, with JSON Schemas generated from their Go types. The console needs the same operations over plain HTTP, with types that its build can check.

## Decision

- Each operation of a Cahut is at `POST https://<domain>:8443/api/<operation>`, such as `/api/list_windows`. The body of the request is the operation's input, in JSON, and the body of the response is its output, with the same names and schemas as the MCP tools. Unknown fields are refused, as over MCP.
- The operations of the server are at `/api/server/<operation>`, for its administrators only.
- The API takes session tokens only ([ADR 13](0013-people-sessions-and-rights.md)), in `Authorization: Bearer`. `sign_in` and `activate_account` are the only operations that need none. AIs keep MCP and their own tokens.
- The API reads no cookie, answers no CORS preflight, and never sends `Access-Control-Allow-Origin`.
- `/api/` accepts `POST` only, reads included: the API is a set of calls, not of resources, and one method keeps the adapter thin.
- Errors map to status codes:

  | Status | When |
  | --- | --- |
  | 400 | the input is invalid (`ErrInvalid`) |
  | 401 | the session is missing, unknown or expired |
  | 403 | the caller may not do this (`ErrForbidden`) |
  | 404 | the record, or the operation, does not exist (`ErrNotFound`) |
  | 409 | the operation needs the approval of a person (`ErrApprovalRequired`), which only AIs meet |
  | 413 | the body is larger than 1 MiB |
  | 429 | too many sign-ins |
  | 500 | anything else |

  The body of an error is `{"error": "<message>"}`. As over MCP, the message of an internal error is not shown: the Logbook keeps it.
- A request body holds at most 1 MiB. The console does not publish: Windows come from AIs.
- Responses carry `Cache-Control: no-store`, `X-Content-Type-Options: nosniff` and a policy that allows nothing: `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`.
- The TypeScript types of the console are generated from the Go types of the registry, by a command of this repository, as the reference of the docs is. CI checks that they are up to date. A call to an operation that does not exist, or with the wrong input, fails the console's build.
- In development, `cahut serve --dev` serves the console and the API on `localhost:8081`, next to the Windows on `localhost:8080`, so that they keep separate origins, as in production.

## Consequences

- The console and the AIs share one surface. Whatever the console does goes through an operation that an AI acting for the same person can call too, except proving who it is ([ADR 13](0013-people-sessions-and-rights.md)).
- The API can be called by hand, with curl and a session token, for tests and scripts.
- Reads go through `POST`, so HTTP caches and bookmarks do not apply to them. The console needs neither.
- The adapter decodes, calls the registry and encodes, and it is the same for every operation: a new operation needs no new route.
- Renaming an operation breaks the console's build at once, as it breaks AI clients and their skills ([ADR 1](0001-operation-registry.md)).

## Alternatives considered

- **REST resources**, such as `GET /api/rooms/{room}/windows`. Familiar, but each operation would need a route written by hand, and the names would drift from those of the MCP tools.
- **MCP for the console as well**, with an MCP client in the browser. One adapter less, but sessions, streams and tool results weigh on every screen, and MCP is made for models, not for interfaces.
- **GraphQL or gRPC-Web.** A schema and typed clients, for a large dependency on both sides, when the registry already has typed operations.
- **Types written by hand in TypeScript.** No generator, but they drift from the Go types without anyone noticing.
