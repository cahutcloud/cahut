# 48. The administrators of the server suspend a Cahut, which keeps everything and serves nothing until they resume it

- Status: Proposed
- Date: 2026-10-29

## Context

A server holds several Cahuts ([ADR 8](0008-several-cahuts-per-server.md)), and a host gives each organization its share of a server ([ADR 30](0030-kind-and-limits-of-each-cahut.md)). ADR 30 left for later records the operations that a host calls to create, suspend, delete or move Cahuts.

A host must be able to stop a Cahut without deleting it: after a decision of moderation, when its organization stops paying, or when the organization asks for it. The Cahut must come back whole, as if nothing had happened.

Today, the administrators of the server have no way to do it. Moving the Cahut to another domain (`rename_domain`) breaks every link and every connection of an AI. Disabling each person (`disable_user`) stops people and their AIs, but public Windows and Windows with a Key stay open, and each person must be enabled again. Removing its directory loses everything.

A Cahut serves in many places: its Windows and the pages of its Rooms, to readers; its console, to people; its MCP endpoint and its OAuth endpoints, to AIs; the admin API, to the programs of the host ([ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)); and its e-mail, through its relay ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)).

## Decision

### One state

- A Cahut is active or suspended. A suspended Cahut keeps everything: its people and their sessions, their tokens and the AIs they connected, its Rooms, its Windows and their versions, its house style, its settings and its Logbook. It serves none of it to anyone but the administrators of its server, until they resume it. Then everything works as before, and nobody signs in or connects again.
- `suspend_cahut` suspends the Cahut, with a reason that the call must give, of 500 characters at most. `resume_cahut` resumes it. Both are operations of the Cahut at the level of the administrators of the server, as `set_cahut_settings` is ([ADR 30](0030-kind-and-limits-of-each-cahut.md)). No AI and no service is offered them.
- The administrators of the server call them from the command line, `cahut domain suspend` and `cahut domain resume`, and from their session in the console of the Cahut, on its page of settings, beside its kind and its limits ([ADR 43](0043-the-console-holds-two-sessions.md)).
- Suspending a Cahut that is suspended already changes its reason only. Resuming a Cahut that is active is refused.

### Where the state is kept

- The state, its date and its reason live in the database of the Cahut, beside its kind and its limits, so that they travel with it when it is backed up, restored or moved, as its settings do. A host that moves a suspended Cahut to another of its servers finds it suspended there.
- The registry reads the state in the transaction of every call, as it reads the kind of the Cahut.
- Each suspension and each resumption goes to the Logbook of the Cahut, with its reason, in the transaction that changes the state.
- The list of the Cahuts of the server, in its console and in `cahut domain list`, says which are suspended, since when and why.

### What a suspended Cahut serves

| Who | What they meet |
| --- | --- |
| Readers | Every Window, page of a Room, frame, page of a Key and link of an invitation answers with a page of Cahut, "This Cahut is suspended", with the status 503. The page carries the brand of Cahut, whatever the house style of the Cahut, and the links to its legal pages, which its host gives ([ADR 42](0042-the-legal-pages-of-a-cahut.md)). It names nothing that the Cahut holds. |
| People | The page of sign-in of the console says that the Cahut is suspended, in place of its form: `describe_sign_in` says so, and is the only operation of the Cahut that answers them. Every other call of a person, signing in included, is refused with the code `cahut_suspended`, and the console shows the same notice on every page. Sessions are kept, and end by themselves as usual. |
| AIs | Their tokens and their connections are kept, and the OAuth endpoints keep answering, so that they stay connected. Each call of a tool is refused with a message that says the Cahut is suspended by the administrators of its server, that nothing it holds is lost, and to tell the person. A new connection is refused with the sign-in of its person. |
| Services | Each call of the admin API is refused with the same code. |
| E-mail | The Cahut sends none while it is suspended: no notice of an access that ends or of a limit, and the operations that send e-mail are refused with all the others. Only the e-mail of the suspension and of the resumption go out, and the messages already in the queue. |
| Administrators of the server | They keep the console of the Cahut and every operation they may call there, to look into it and to resume it. The command line works as before. They still open no Window ([ADR 13](0013-people-sessions-and-rights.md)). |

- The files of the brand of Cahut, the style sheet and the scripts of its pages, its icon and `/-/healthz` keep answering. What the Cahut serves under `/-/` of its own, its house style, its templates, the deck engine and the skill, answers with the page.
- The health of the server, which the updater reads, opens and migrates a suspended Cahut as every other, and lists it as suspended: a suspended Cahut does not fail an update. `cahut doctor` names the suspended Cahuts.
- Time goes on: a Knock that waits expires, an access ends at its date, an invitation expires, as they would in an active Cahut.

### The reason

- The reason is for the administrators of the server and of the Cahut. Its administrators read it in the e-mail of the suspension, and in the Logbook once the Cahut is resumed.
- When the Cahut has a relay, its administrators who are active and have an address receive an e-mail when it is suspended, with the reason and the links to its legal pages, and another when it is resumed. Without a relay, the Cahut sends nothing, and the host tells them as it tells them anything else.
- Readers, people and AIs never read the reason. People are told to ask the administrators of the Cahut.

## Consequences

- A host stops a Cahut in one command and gives it back in another, without loss: people find their sessions, AIs their connections and readers their links.
- A Cahut where people read and no longer publish needs no second state: lowering its limits of disk and of Windows to what it holds refuses every new Window and every new version, and keeps everything readable ([ADR 30](0030-kind-and-limits-of-each-cahut.md)).
- A suspended Cahut costs its disk, and nothing else.
- Every call reads one more value in its transaction, and every request to the Cahut one more row of its database.
- The console says the refusal in English and French, with its code ([ADR 37](0037-refusals-with-a-code-for-the-console.md)). The documentation for administrators says how to suspend and resume a Cahut.
- A person who stops paying in a Cahut that many people share, such as an individual Cahut, is not a case for this record: `disable_user` stops one person.
- A migration adds the state to the database of every Cahut: active.
- This record extends ADR 8 (a Cahut can be suspended), ADR 13 (only the administrators of the server act in a suspended Cahut) and ADR 30 (its state is kept with its settings).

## Alternatives considered

- **Two levels, read-only and closed.** Reading only would need a decision for each operation that changes something without publishing: signing in, a Key, an invitation, an approved Knock, a new password. The limits already give a Cahut where nothing new is published. A second level can come later, beside this one, if hosts ask for it.
- **The state in `server.db`, beside the domain.** The server would know which Cahuts are suspended before it opens one, and its list would read nothing else. But the registry would read a second database, outside the transaction of each call; the Logbook of the Cahut and its e-mail would need a second step after the change; and the state would not travel with the directory of the Cahut.
- **The reason on the page and in the console.** Clearer for people, but a page that anyone opens would tell why an organization lost its Cahut: a debt, or a decision of moderation.
- **Ending the sessions, the tokens and the connections of AIs.** Safer if one was stolen, but resuming would ask every person to sign in again and every AI to connect again. A refused call protects as much while the Cahut is suspended.
- **Serving nothing at the domain**, with 421 as for a domain that no Cahut declares. Readers and AI clients would read an error of the network rather than a suspended Cahut, and the administrators of the server would lose its console.
- **The page in the house style of the Cahut.** It is the host who speaks, not the organization.
