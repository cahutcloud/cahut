# Architecture decision records

Each record states one architecture decision for Cahut: its context, what was decided, what follows from it, and the options set aside. Once accepted, a record changes only for typos and links. A new record replaces it, and the old one links to the new one.

| Record | Decision | Status |
| --- | --- | --- |
| [1](0001-operation-registry.md) | One registry of typed operations is the only way in | Accepted, superseded in part by 13 |
| [2](0002-built-in-https-with-certmagic.md) | The binary serves HTTPS itself, with CertMagic | Accepted, superseded in part by 8, 11 and 20 |
| [3](0003-one-origin-per-window.md) | Each Window gets its own web origin | Superseded by 8, 9 and 10 |
| [4](0004-sqlite-in-pure-go.md) | SQLite in pure Go | Accepted, extended by 8 |
| [5](0005-preact-console-embedded.md) | The console is a Preact app embedded in the binary | Accepted, superseded in part by 11 |
| [6](0006-in-house-oauth-server.md) | Cahut runs its own minimal OAuth authorization server | Accepted, extended by 16 |
| [7](0007-module-path-repository-and-releases.md) | Module path, repository and releases | Accepted |
| [8](0008-several-cahuts-per-server.md) | A server holds several Cahuts, each at its own domain | Accepted |
| [9](0009-rooms-and-windows-as-paths.md) | Rooms and Windows are paths of their Cahut's domain | Accepted |
| [10](0010-windows-in-a-csp-sandbox.md) | Windows run in a CSP sandbox | Accepted |
| [11](0011-console-on-its-own-port.md) | The console has its own port | Accepted, superseded in part by 13, 17 and 20 |
| [12](0012-website-and-docs-as-astro-sites.md) | The website and the docs are Astro sites in this repository | Accepted, extended by 18 |
| [13](0013-people-sessions-and-rights.md) | People sign in to their Cahut, and the registry checks their rights | Accepted, superseded in part by 21 |
| [14](0014-console-http-api.md) | The console calls each operation at its own HTTP route | Accepted, superseded in part by 20 |
| [15](0015-license.md) | AGPL-3.0 for the repository, the name and the logo apart | Accepted |
| [16](0016-oauth-for-ai-clients.md) | AI clients connect over OAuth, to the Rooms and the scope a person grants | Accepted, superseded in part by 17, 20 and 21 |
| [17](0017-oauth-endpoints-on-port-443.md) | The OAuth endpoints that clients' servers call are on port 443 | Accepted, superseded in part by 20 |
| [18](0018-project-pages-from-community-files.md) | The project pages are built from the community files of the repository | Accepted |
| [19](0019-window-versions-access-and-declaration.md) | Windows keep two versions, open to an access, and declare what they call | Accepted, extended by 21 |
| [20](0020-console-under-a-path-of-port-443.md) | The console is served under `/console/` on port 443 | Accepted |
| [21](0021-ais-knock-first.md) | An AI knocks first, and the person it acts for answers | Accepted |

Records 1 to 7 call an instance of Cahut a home, and speak of Doors. Since record 8, an instance is a Cahut, and each Window has its own access level, without Doors.

## Writing a record

- Copy the layout of an existing record: a title that states the decision, then Status, Date, Context, Decision, Consequences and Alternatives considered.
- Number records in order and never reuse a number. The file name is the number on four digits, then the title in a few words.
- Write in English: the records are public at launch.
- Add the record to the table above in the same pull request.
