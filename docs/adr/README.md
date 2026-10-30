# Architecture decision records

Each record states one architecture decision for Cahut: its context, what was decided, what follows from it, and the options set aside. Once accepted, a record changes only for typos and links. A new record replaces it, and the old one links to the new one.

| Record | Decision | Status |
| --- | --- | --- |
| [1](0001-operation-registry.md) | One registry of typed operations is the only way in | Accepted, superseded in part by 13 |
| [2](0002-built-in-https-with-certmagic.md) | The binary serves HTTPS itself, with CertMagic | Accepted, superseded in part by 8, 11 and 20, extended by 41 |
| [3](0003-one-origin-per-window.md) | Each Window gets its own web origin | Superseded by 8, 9 and 10 |
| [4](0004-sqlite-in-pure-go.md) | SQLite in pure Go | Accepted, extended by 8 |
| [5](0005-preact-console-embedded.md) | The console is a Preact app embedded in the binary | Accepted, superseded in part by 11 |
| [6](0006-in-house-oauth-server.md) | Cahut runs its own minimal OAuth authorization server | Accepted, extended by 16 |
| [7](0007-module-path-repository-and-releases.md) | Module path, repository and releases | Accepted, extended by 24 |
| [8](0008-several-cahuts-per-server.md) | A server holds several Cahuts, each at its own domain | Accepted, extended by 26, 30 and 48 |
| [9](0009-rooms-and-windows-as-paths.md) | Rooms and Windows are paths of their Cahut's domain | Accepted, extended by 22, 30, 32 and 50, superseded in part by 38 |
| [10](0010-windows-in-a-csp-sandbox.md) | Windows run in a CSP sandbox | Accepted, extended by 27 |
| [11](0011-console-on-its-own-port.md) | The console has its own port | Accepted, superseded in part by 13, 17 and 20 |
| [12](0012-website-and-docs-as-astro-sites.md) | The website and the docs are Astro sites in this repository | Accepted, extended by 18 |
| [13](0013-people-sessions-and-rights.md) | People sign in to their Cahut, and the registry checks their rights | Accepted, superseded in part by 21 and 54, extended by 30, 31, 43 and 48 |
| [14](0014-console-http-api.md) | The console calls each operation at its own HTTP route | Accepted, superseded in part by 20, extended by 37 |
| [15](0015-license.md) | AGPL-3.0 for the repository, the name and the logo apart | Accepted |
| [16](0016-oauth-for-ai-clients.md) | AI clients connect over OAuth, to the Rooms and the scope a person grants | Accepted, superseded in part by 17, 20 and 21 |
| [17](0017-oauth-endpoints-on-port-443.md) | The OAuth endpoints that clients' servers call are on port 443 | Accepted, superseded in part by 20 |
| [18](0018-project-pages-from-community-files.md) | The project pages are built from the community files of the repository | Accepted |
| [19](0019-window-versions-access-and-declaration.md) | Windows keep two versions, open to an access, and declare what they call | Accepted, extended by 21, 22, 27, 30, 31, 32, 52, 53 and 54, superseded in part by 45, 49 and 50 |
| [20](0020-console-under-a-path-of-port-443.md) | The console is served under `/console/` on port 443 | Accepted, extended by 43 |
| [21](0021-ais-knock-first.md) | An AI knocks first, and the person it acts for answers | Accepted, extended by 31 |
| [22](0022-house-style-of-each-cahut.md) | Each Cahut has a house style, served to its Windows and given to AIs | Superseded by 49, but for the skill for Claude |
| [23](0023-house-style-inactive-for-ais.md) | The house style can be inactive for AIs | Superseded by 49 |
| [24](0024-signed-releases-and-updates.md) | Releases are signed, and servers update themselves with a rollback | Accepted |
| [25](0025-first-start-in-the-browser.md) | The first start is guided in the browser, from a single-use link | Accepted, extended by 31 and 43, superseded in part by 49 |
| [26](0026-cahut-at-an-ip-address.md) | A Cahut can live at the IP address of its server until it has a domain | Accepted, superseded in part by 28 |
| [27](0027-previews-of-windows-in-the-console.md) | The console shows each Window in a live preview | Accepted, extended by 32, superseded in part by 50 |
| [28](0028-ip-address-of-a-cahut-is-ipv4.md) | The IP address of a Cahut is an IPv4 address, for now | Accepted |
| [29](0029-chatgpt-second-ai-client.md) | ChatGPT is the second AI client of the V0, through the same tools as Claude | Accepted, superseded in part by 49 |
| [30](0030-kind-and-limits-of-each-cahut.md) | Each Cahut has a kind and limits, set by the administrators of its server | Accepted, extended by 42 and 48, superseded in part by 49 |
| [31](0031-e-mail-through-a-relay.md) | Each Cahut sends e-mail through its own relay or that of its server | Accepted, superseded in part by 34 and 46 |
| [32](0032-the-page-of-a-room.md) | The page of a Room shows its Windows in previews, or a Window that its owners choose, in the brand of Cahut until the house style | Accepted, extended by 33, superseded in part by 38 and 49 |
| [33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md) | The pages of Cahut carry its brand until the house style has a look of its own | Superseded by 49 |
| [34](0034-e-mail-through-the-relay-of-each-cahut-only.md) | E-mail goes through the relay of each Cahut only | Accepted |
| [35](0035-a-program-on-the-server-manages-people-with-a-service-token.md) | A program on the server manages the people of a Cahut with a service token | Accepted |
| [36](0036-console-in-english-and-french.md) | The console speaks English and French, in the language of each browser | Accepted, superseded in part by 37 |
| [37](0037-refusals-with-a-code-for-the-console.md) | The refusals that people meet carry a code, which the console says in its language | Accepted |
| [38](0038-the-root-of-a-cahut-opens-its-console.md) | The root of a Cahut opens its console | Accepted |
| [39](0039-people-change-their-name-and-handle.md) | People change their name and their handle, and their personal Room follows | Accepted, extended by 50 |
| [40](0040-the-icon-of-a-cahut.md) | Each Cahut gives its icon to browsers, favicon services and AI clients | Accepted, superseded in part by 49 |
| [41](0041-the-proxy-protocol-from-trusted-proxies.md) | Cahut reads the PROXY protocol from the proxies it trusts | Accepted |
| [42](0042-the-legal-pages-of-a-cahut.md) | The pages of a Cahut link its legal pages, which the administrators of its server give | Accepted, extended by 54 |
| [43](0043-the-console-holds-two-sessions.md) | The console holds the session of a person and that of an administrator of the server together | Accepted, superseded in part by 44 |
| [44](0044-two-bars-for-the-server-and-the-cahut.md) | The console shows the server and the Cahut in two bars, and each signs out of its own session | Accepted |
| [45](0045-windows-fetch-the-blob-and-data-urls-they-make.md) | The pages of a Window fetch the blob: and data: URLs they make | Accepted |
| [46](0046-styled-notes-in-invitations.md) | The note of an invitation carries bold, italics and lists | Accepted |
| [47](0047-deleting-a-person.md) | The administrators of a Cahut delete a person for good, with the Rooms that are theirs alone | Accepted |
| [48](0048-suspending-a-cahut.md) | The administrators of the server suspend a Cahut, which keeps everything and serves nothing until they resume it | Accepted |
| [49](0049-no-house-style.md) | Cahut has no house style, until a module of skills brings the brand of an organization | Accepted |
| [50](0050-window-ids-permanent-addresses-and-names.md) | Each Window has an id that never changes and a permanent address, and its people change its name, its title and its description | Accepted, extended by 52 |
| [51](0051-sharing-a-window-and-its-preview.md) | People share a Window by its permanent address, which shows the robots of previews a card of Cahut | Accepted, extended by 52 |
| [52](0052-share-links.md) | A share link opens a Window to whoever holds it, without an account or a Key, until it ends | Accepted, extended by 53 and 54 |
| [53](0053-windows-for-their-author-alone.md) | A Window can open to its author alone, and a share link that opens once gives it a single reader | Accepted |
| [54](0054-reporting-and-withdrawing-a-window.md) | Anyone reports a Window, and the moderators of its Cahut open it and withdraw it | Accepted |

Records 1 to 7 call an instance of Cahut a home, and speak of Doors. Since record 8, an instance is a Cahut, and each Window has its own access level, without Doors.

## Writing a record

- Copy the layout of an existing record: a title that states the decision, then Status, Date, Context, Decision, Consequences and Alternatives considered.
- Number records in order and never reuse a number. The file name is the number on four digits, then the title in a few words.
- Write in English: the records are public at launch.
- Add the record to the table above in the same pull request.
