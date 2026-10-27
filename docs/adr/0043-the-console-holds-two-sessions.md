# 43. The console holds the session of a person and that of an administrator of the server together

- Status: Accepted, superseded in part by [ADR 44](0044-two-bars-for-the-server-and-the-cahut.md) (the console shows the server and the Cahut in two bars, and each signs out of its own session)
- Date: 2026-10-26

## Context

The administrators of the server have accounts of their own, in `server.db`, apart from the people of the Cahuts ([ADR 13](0013-people-sessions-and-rights.md)). Someone who installs Cahut alone has two accounts, which the first start makes with the same handle and the same password ([ADR 25](0025-first-start-in-the-browser.md)): the administrator of the server, and the first person of their Cahut, its administrator.

Both sign in at the console of the same domain, under `/console/` ([ADR 20](0020-console-under-a-path-of-port-443.md)), and the console keeps one session in the storage of its origin. Signing in with one account forgets the other. Someone alone on their server signs out of the account they use every day, with its Rooms, its Knocks and its AIs, to reach what only the account of the server reaches: the Cahuts of the server, its administrators, its Logbook and its updates; then signs out again to come back. The consent screen of an AI client, the page of a Knock and the sign-in of readers sign in as a person on the spot, and forget the session of the server without saying so.

The two consoles also overlap. An administrator of the server acts in the Cahut of the domain where they signed in as its administrators do ([ADR 13](0013-people-sessions-and-rights.md)), so both show its people, its house style, its kind and limits, its e-mail, its Rooms and its Logbook. Two consoles that look alike, at one address, for one person, read as one account that changes shape.

## Decision

### Two places in the browser

- The console of a domain holds up to two sessions in the storage of its origin: that of a person of its Cahut, and that of an administrator of the server. Each sign-in takes its own place and leaves the other as it is: signing in as a person keeps the session of the server, and signing in to the server keeps that of the person.
- A sign-in that takes a place already held ends, on the server, the session it replaces.
- Signing out ends both sessions. Leaving the server, from the page of the account, ends its session only.
- A session that ends by itself, 7 days after its last use or 30 days after it opened, frees its own place only, and the console asks to sign in again with that account.

### One console

- With both sessions, the console is that of the person, with a section of its sidebar for the server: its Cahuts, its administrators, its Logbook and its updates. The pages of the server call the API with the session of the server, and so do the kind, the limits and the legal pages of the Cahut, which only the administrators of the server set ([ADR 30](0030-kind-and-limits-of-each-cahut.md), [ADR 42](0042-the-legal-pages-of-a-cahut.md)). Every other page, those of the Cahut included, calls it with the session of the person, within their rights. The top bar names the person.
- With the session of the server alone, the console is what it was: the pages of the server, and those of the Cahut that its administrators reach.
- The page of the account names the accounts that the browser holds and changes the password of each. It adds the other account: an administrator of the Cahut signs in to the server from it, and an administrator of the server signs in as a person of the Cahut.
- The first start, which makes both accounts with one password, opens both sessions.

### Nothing changes on the server

- Two accounts, two sessions, and each call carries one of them. No request touches two databases, and a session stays good at the domain where it opened.

## Consequences

- Someone alone on their server signs in once with each account, in the same browser, and then uses one console. The first start does it for them.
- The consent screen, the page of a Knock and the sign-in of readers keep the session of the server.
- A browser reaches what it holds, as before: whoever uses a browser that holds the session of the server reaches the server, whether or not a person signs in beside it. The page of the account names both accounts, and signing out ends both.
- A browser that held the session of an administrator of the server before keeps it, in its new place.
- A person who is a member of the Cahut, and administers the server, sees the pages of the Cahut as a member, but for its kind and limits. To act in the Cahut as its administrators do, an administrator of the Cahut makes them one, or they sign in to the server alone.

## Alternatives considered

- **One account for both, a person with a flag of the server.** [ADR 13](0013-people-sessions-and-rights.md) set it aside: signing in at another domain would read the database of the Cahut that holds the account.
- **A person linked to an account of the server, which the server follows.** One sign-in, but a session of a Cahut would reach `server.db`, and one request would touch two databases.
- **Two consoles, each with its session, and a switch between them.** No sign-in to repeat, but two consoles that overlap stay two, and the person goes back and forth between them.
- **Signing in to both accounts at once when one handle and one password open both.** One sign-in, but every sign-in of a person would check a second password against `server.db`, and the two accounts would part the day one password changes.
