# 35. A program on the server manages the people of a Cahut with a service token

- Status: Accepted
- Date: 2026-10-24

## Context

A host that runs a Cahut for many people, such as an individual Cahut where people sign up on a site of the host ([ADR 30](0030-kind-and-limits-of-each-cahut.md)), needs a program to create those people, give them their activation link, and later disable them. That program works at any hour, for people it has never met.

Today, the people of a Cahut change through three doors, and none fits:

- **The console**, with the session of a person, which expires, and which a program could only open with a password it would have to keep.
- **The command line**, run as the service user, which opens the databases outside the service, from a shell.
- **An AI**, over MCP, with the personal token of an administrator. Adding a person knocks first ([ADR 21](0021-ais-knock-first.md)): every sign-up would wait for a person to approve it.

ADR 30 said that the operations a program calls for a host would come with later records. This record covers the people of one Cahut.

## Decision

### Service tokens

- A service token belongs to one Cahut. The administrators of the server create, list and revoke them with `create_service_token`, `list_service_tokens` and `revoke_service_token`, operations of the Cahut at their level, from the command line (`cahut service-token`) or from their session in the console. No AI is offered them.
- The token is shown once. The Cahut keeps its SHA-256, as for personal tokens, and records its last use, at most once a minute. It starts with `cahut_svc_`, so that secret scanners recognize it. Revoking it ends its access at once.
- A call with a service token is the call of a new kind of actor, a service. It acts with the rights of the administrators of the Cahut, for the operations open to services only, and it never knocks: the administrator of the server who created the token approved, once, what the program does.
- Each operation says whether services may call it (`Op.Services`). Open to services: `check_handle`, `get_user`, `create_user`, `create_activation_link`, `set_user_email`, `disable_user`, `enable_user` and `get_cahut`. None of them makes an administrator, or touches the house style, the settings of the Cahut, its Rooms or its Windows.
- The Logbook records a call of a service as one of an administrator, through `admin-api`, under the name of its token, such as `service token account`. Its table keeps its three kinds of actors.

### The admin API

- `cahut serve --admin-api <address>`, or `CAHUT_ADMIN_API`, serves the operations of every Cahut of the server at `POST /api/<operation>`, as [ADR 14](0014-console-http-api.md) has it for the console: the input and the output in JSON, with the names and schemas of the registry, the same status codes, and the token in `Authorization: Bearer`. The host of the request names the Cahut, as for every other request.
- The address is one of the machine itself: Cahut refuses to start with another one, and the API refuses every request whose client is not on the machine.
- It takes service tokens only, and they open nothing else: neither a session nor a personal token reaches the admin API, and a service token reaches neither the console, nor MCP, nor the Windows.
- It is off unless the address is given.

### What the operations gain

- `check_handle` says whether a handle can go to a new person: valid, not reserved, held by no person and by no Room. It serves the administrators of the Cahut and services, and is not offered to AIs: `list_users` tells them who is there.
- `get_user` describes one person, as `list_users` does, with their address for those who may read it. It serves the same callers.
- `create_user` and `create_activation_link` take `send_activation_email`, true when absent. False keeps the link from the address of the person: the caller gives it to them itself, so that they receive one e-mail, not two.

## Consequences

- A host signs people up from a site of its own, on the same machine, without a person in the loop, and the Logbook keeps every change that the site makes.
- Whoever reads a service token acts as an administrator of its Cahut for the operations open to services: it lives in the configuration of the program only, and the administrators of the server revoke it in one command.
- The machine is the boundary. A program on another machine needs a tunnel, which is the choice of the host.
- The console's API and the tools of MCP do not change. The reference of the docs lists the operations open to services, and the routes of the admin API.
- Opening an operation to services later is a change of one field, and a pull request that says why.

## Alternatives considered

- **The command line, from the program.** Every call would open the databases outside the service, from a shell of the service user, against the single writer of each Cahut.
- **The personal token of an administrator.** Its calls are those of an AI: adding a person knocks first. And the token stops working when its person is disabled.
- **Knocks approved by a program.** It undoes the rule of ADR 21 for every AI, not for one program.
- **The admin API on port 443, behind its token only.** The token would be the only barrier between the internet and the people of the Cahut.
- **A fourth kind of actor in the Logbook.** Its table would need to be made again, with every entry, for a word that `via` and the name of the token already give.
