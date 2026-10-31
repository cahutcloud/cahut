# 58. The administrators of the server have a console of their own, under `/server/`

- Status: Accepted
- Date: 2026-10-31

## Context

The console of a domain holds the session of a person of its Cahut and that of an administrator of the server together ([ADR 43](0043-the-console-holds-two-sessions.md)), and shows them in two bars ([ADR 44](0044-two-bars-for-the-server-and-the-cahut.md)). Its sidebar mixes two sections, This server and This Cahut. Some pages of This Cahut call the API with the session of the server: the kind, the limits, the legal pages, the address of moderation and the suspension. The others call it with that of the person. An administrator of the server alone in the browser acts in the Cahut as its administrators do ([ADR 13](0013-people-sessions-and-rights.md)), and sees its people, its Rooms and its Logbook.

In use, nobody knows who does what. Someone alone on their server holds both sessions and reads one console that changes shape with them. The administrators of a Cahut cannot tell their pages from those of the server. A host who looks after several Cahuts signs in again at the console of each, since a session of the server is good at one domain (ADR 13).

The two roles do not do the same work. The administrators of a Cahut look after its people, its Rooms, its e-mail and its moderation. The administrators of the server declare Cahuts, set the kind, the limits and the legal pages of each, suspend and resume them, moderate them as the host ([ADR 54](0054-reporting-and-withdrawing-a-window.md)), and look after the administrators of the server, its Logbook and its updates.

## Decision

### Two consoles

- **The console of the Cahut**, `https://<domain>/console/`, is for the people of the Cahut: its members and its administrators. For a person, it stays what it was, without the bar of the server and without the pages of the server. It no longer opens for an administrator of the server.
- **The console of the server**, `https://<domain>/server/`, is for the administrators of the server. It is served at every domain of the server, and at its IP address ([ADR 26](0026-cahut-at-an-ip-address.md)).
- The console of the Cahut does not name the console of the server. The console of the server gives the address of each Cahut, as a plain link that carries no session.

### The console of the Cahut

- It signs in the people of the Cahut only. Its page of the account no longer adds the account of the server.
- Its administrators see what they saw: its people, its e-mail, its moderation, and the kind, the limits, the legal pages and the address of moderation, which they read. What needed the session of the server leaves it: setting the kind, the limits, the legal pages and the address of moderation, suspending and resuming the Cahut, and reinstating a Window that the administrators of the server withdrew.
- A suspended Cahut says that the administrators of the server resume it ([ADR 48](0048-suspending-a-cahut.md)).

### The console of the server

- Its top bar is Ink, Paper in the dark theme, and says Server beside the logo, with the name of the administrator and the button that signs them out. Its sidebar holds the server alone: Cahuts, Administrators, Logbook and Updates, with Account at its foot.
- Cahuts lists the Cahuts of the server, with their domain, their kind, the reports that wait in each and their suspension. It declares a Cahut, gives a link of first start, and moves a Cahut to another domain.
- The page of a Cahut, `/server/cahuts/<domain>`, sets its kind, its limits, its legal pages and its address of moderation, suspends and resumes it, and creates and revokes its service tokens ([ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)). It moderates it too: its reports, its public Windows, the preview of a reported Window, its withdrawal and its reinstatement (ADR 54).
- Account changes the password of the server, and signs out.
- It has its own sign-in, `/server/sign-in`, and its own activation, `/server/activate#<code>`. The links that the command line prints for an administrator of the server name them.
- It speaks English and French, as the console of the Cahut does ([ADR 36](0036-console-in-english-and-french.md)). Both are built from one app, each with its own entry: they share its components, its words, its client of the API and the brand.

### The API

- `POST /server/api/<operation>` runs an operation of the server. It replaces `/console/api/server/<operation>` ([ADR 20](0020-console-under-a-path-of-port-443.md)).
- `POST /server/api/cahuts/<domain>/<operation>` runs an operation in the Cahut of `<domain>`, which the server opens from the path, as the command line does from `--domain`. A domain that the server does not hold is not found.
- `POST /console/api/<operation>` takes the sessions of people only.
- Both take a session token in the `Authorization` header, read no cookie and grant no CORS access ([ADR 14](0014-console-http-api.md)).

### What the administrators of the server do in a Cahut

- They call there the operations that serve the server only: those at the level of the administrators of the server, those of moderation, and `get_cahut`, which reads the settings that they set. An operation says so in a field of its own, which the registry reads, and a test keeps the list.
- They no longer act in a Cahut as its administrators do: neither its people, nor its Rooms, nor its e-mail, nor its Logbook. To act there, they are one of its people.
- What they do in a Cahut goes to its Logbook, under `admin` and their handle, as before. Its administrators read it there.
- The command line still acts on the server and in every Cahut.

### Sessions

- A session of the server lives in `server.db`, and is good at the domain where it opened (ADR 13). From that domain, it reaches every Cahut of the server, by naming it.
- The console of the server keeps it in the storage of the origin, under the key where the console kept it before: a browser that held it stays signed in. The console of the Cahut no longer reads it.
- The preview of a reported Window frames it at the domain of its Cahut ([ADR 27](0027-previews-of-windows-in-the-console.md)). The policy of the console of the server allows frames from the domains of the Cahuts of the server, which the server knows, and the capability of a preview takes a session of the server wherever it opened.

### The first start

- The link of the first start does not change, nor do its steps You and Your Cahut ([ADR 25](0025-first-start-in-the-browser.md)).
- When the first start makes the administrator of the server, it opens both sessions and ends in the console of the server, on the page of the Cahut, which links to the console of the Cahut. When the server has administrators already, it ends in the console of the Cahut.
- The step Your first AI leaves the first start. The overview of the console of the Cahut offers it, as long as the person has neither a connected AI nor a token, as ADR 25 has it.
- `install.sh` prints the address of the console of the server beside that of the console of the Cahut.

### A reserved name

- `server` becomes a reserved name, as `console` and `admin` are ([ADR 20](0020-console-under-a-path-of-port-443.md)): no Room and no handle takes it.

This record supersedes ADR 43 and ADR 44. It replaces, in ADR 13, the lines under which an administrator of the server acts in a Cahut as its administrators do and signs in at the console of each Cahut; in ADR 20, the address of the API of the server; in ADR 25, the end of the first start and its third step; and in ADR 54, the console where the administrators of the server moderate.

## Consequences

- Each console says who acts in it: a person in the console of the Cahut, an administrator of the server in that of the server. No screen holds two accounts.
- A host looks after every Cahut of the server from one place, without signing in at each.
- Someone alone on their server keeps two accounts, which the first start makes with one password, and two addresses. The first start ends in the console of the server, and `install.sh` prints its address; after that, they type `/server/` to reach it.
- An administrator of the server who looked after the people or the Rooms of a Cahut needs an account of a person there, as one of its administrators.
- A request of the console of the server reads `server.db` for its session, then the database of the Cahut that it names. The server opens that Cahut, as the command line does: no Cahut reads the records of another ([ADR 8](0008-several-cahuts-per-server.md)).
- The console of the server hides, at `/server/`, a Room or a handle named `server` made before this record. A person changes such a handle ([ADR 39](0039-people-change-their-name-and-handle.md)); a team Room keeps its Windows at their permanent addresses ([ADR 50](0050-window-ids-permanent-addresses-and-names.md)).
- The links of activation that the command line printed for an administrator of the server before this record name `/console/activate`, and stop working. `cahut admin link` gives a new one.
- The bar of the server leaves the console of the Cahut, and its 48 pixels with it.
- The pages of the server load in their own entry: the browser of a member does not load them.

## Alternatives considered

- **One console with clearer sections.** ADR 44 tried it with a bar of the server: the pages of the two roles still mix, and a page of the Cahut acts as one account or the other.
- **A domain of its own for the server**, such as `server.example.com`. An origin of its own and an address that says where one is, but a DNS record and a certificate more, and no address at the first start, before the server has a domain.
- **A port of its own.** An origin of its own without DNS, but out of reach on the networks that let port 443 out only, as ADR 20 found.
- **A session of the server good at the Cahut of its domain only**, as before. Fewer changes, but a host still signs in at each Cahut.
- **The pages of the people, the Rooms and the Logbook of each Cahut in the console of the server.** The administrators of the server would keep all they reach today, but every page would exist twice, and the two consoles would overlap again.
- **A link from the console of the Cahut to that of the server.** One address less to remember, but a console that names the other brings back the mix that this record removes.
- **One app with two sets of routes**, which chooses at start which console it is. Fewer changes to the build, but one router for both, and the code of the server in the browser of every member.
- **Operations of the server that wrap those of a Cahut**, such as `set_cahut_settings` with the domain in its input. One prefix for the API, but each operation twice, with two ways in for one action, against [ADR 1](0001-operation-registry.md).
