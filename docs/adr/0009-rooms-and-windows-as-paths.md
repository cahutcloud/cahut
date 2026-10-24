# 9. Rooms and Windows are paths of their Cahut's domain

- Status: Accepted, extended by [ADR 22](0022-house-style-of-each-cahut.md) (Windows link the house style of their Cahut under `/-/`, by root-relative paths) and by [ADR 30](0030-kind-and-limits-of-each-cahut.md) (in an individual Cahut, the name of a Room starts with the handle of its owner)
- Date: 2026-10-21

## Context

A Cahut has one domain ([ADR 8](0008-several-cahuts-per-server.md)). Its Rooms and Windows need addresses that read well, stay stable and can be typed, like `github.com/<organization>/<repository>`.

## Decision

- A Room is at `https://<domain>/<room>/` and a Window at `https://<domain>/<room>/<window>/`. The files of a Window sit under that prefix, such as `https://<domain>/<room>/<window>/img/logo.svg`. An address without its final slash redirects to the one with it, so that relative links resolve.
- The names of Rooms and Windows use 1 to 63 lowercase ASCII letters, digits and hyphens, and start and end with a letter or a digit.
- The name of a Room is unique in its Cahut, and the name of a Window is unique in its Room. Personal and team Rooms share one set of names: a person's personal Room is named after their handle, which no team Room can take.
- A Window's name comes from its title, with accents folded, and `-2`, `-3` added on a clash. It holds no random part: the access level of each Window protects it, not a hard-to-guess address.
- Cahut's own pages and endpoints on the domain live under `/-/`, which no name can take: the Key page, the reader's sign-in, the house style sheet, the health check. A few names are reserved at the root, because clients expect them or Cahut may need them: `mcp`, for the MCP endpoint at `https://<domain>/mcp`, then `api`, `admin`, `console` and `help`. Paths such as `/.well-known/`, `/favicon.ico` and `/robots.txt` hold a dot, which no name can.
- `https://<domain>/` is a page of the Cahut, and the address of a Room lists the Windows its reader may open. A Room or a Window that the reader may not open answers `404 Not Found`, like one that does not exist, as GitHub does for private repositories, with a link to sign in.
- A Window keeps its address across its versions. Renaming a Room or a Window, and moving a Window to another Room, come after the V0, with a redirect from the old address.
- Pages link with relative paths. A page at `/<room>/<window>/` that loads `/style.css` reaches the root of the Cahut, not its own folder. The MCP instructions and the skills say so.

## Consequences

- Addresses read like the organization: `cahut.example.com/marketing/q3-review/`.
- Addresses can be guessed. A closed Window and a missing one answer alike, and the access level is checked on every request.
- AI-made pages that use root-relative links break until the AI fixes them. `publish_window` can warn when it finds such links.
- The files of a Window are stored under its id, not its name, so that a rename moves no file.

## Alternatives considered

- **A prefix for Rooms**, such as `/r/<room>/`, as an early example in the brand guidelines had it. It frees the root for Cahut's own pages, but lengthens every address; `/-/` does the same job.
- **A random part in every Window name**, as batch 1 first had it. It hides addresses, but access levels protect Windows, and readable addresses are worth more.
- **A subdomain per Room.** It needs a wildcard record again.
