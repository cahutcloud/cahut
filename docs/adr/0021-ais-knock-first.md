# 21. An AI knocks first, and the person it acts for answers

- Status: Accepted
- Date: 2026-10-23

## Context

Knock first is a rule of the product: what can be undone, an AI does alone; what cannot waits for a person to approve it in the console, with a notification and an expiry. [ADR 1](0001-operation-registry.md) put the rule in the registry, and [ADR 13](0013-people-sessions-and-rights.md) marked the operations that knock first: those that cannot be undone, and those that change what a person may do. Until batch 5 of the [V0 plan](../plan-v0.md), an AI that calls one is refused, with a message that sends its person to the console. [ADR 19](0019-window-versions-access-and-declaration.md) did the same for a Window made public, and [ADR 16](0016-oauth-for-ai-clients.md) left a connected AI no way to ask for more than its person granted.

Batch 5 turns these refusals into Knocks: a request that waits for a person, who approves or denies it. It must settle what knocks, how a Knock is checked and kept, who answers it, how it runs once approved, what the AI learns, and what the Logbook records.

## Decision

### What knocks

- The operations marked `KnockFirst` knock whenever an AI calls them: `create_user`, `set_user_role`, `disable_user`, `enable_user`, `add_room_member`, `set_room_member_role`, `remove_room_member`, and two new ones, `purge_window` and `delete_room`.
- Some operations knock for some calls only, and decide it while they run: `set_window_access` when it makes a Window public, and `set_ai_reach` when it widens what an AI reaches.
- People never knock. In the console, a person runs these operations at once, after a confirmation ([ADR 1](0001-operation-registry.md)). The command line never knocks.
- This answers the open question of the V0 plan on the list of irreversible actions. Erasing backups or changing their destination will knock when backups come, after the V0.

### Knocking

- When an AI calls an operation that knocks, the registry checks the call as it checks any other: the actor, the level, the reach of the AI, the role and the state of its person. It then runs the operation in its transaction and rolls the transaction back. The call is checked against the records as if it ran, and changes nothing. An AI learns at once that a handle is taken or that a Room has a single owner, rather than its person at approval.
- An operation that knocks keeps its effects in its transaction, and acts on the disk after the commit only. A rolled back call leaves nothing behind.
- An operation that decides to knock while it runs may stop there. What it checks before, the AI learns at once; what it checks after, the approval checks. `set_ai_reach` stops before it checks the Rooms an AI names, so that a limited AI does not learn, by naming them, which Rooms its person belongs to.
- The registry then records a Knock, in a new transaction: the operation, its input, what it acts on as the Logbook says it, the Room it concerns, the AI (its personal token or its connection, and its name), the person it acts for, when it was made and when it expires. The input of a Knock holds no secret: no operation knocks for an input that carries a Key or a password.
- A Knock expires after 24 hours. This answers the open question of the V0 plan on the default expiry. The V0 does not let a Cahut change it.
- The same call by the same AI, while its Knock waits, returns that Knock rather than a new one. An AI has at most 20 Knocks waiting; the next call is refused until its person answers some.

### What the AI learns

- The call returns an error to the AI, whose message names the person who must approve, gives the address of the Knock in the console, `https://<domain>/console/knocks/<id>`, says when it expires, and asks the AI to pass the address on. Over MCP it is a tool result marked as an error: the call did not run, and a result in the shape of the operation's output would read as done.
- `get_knock` and `list_knocks` tell an AI the state of the Knocks it made, so that it learns the outcome once its person answers. An AI reads only its own Knocks.
- The address carries no secret. Only the person can answer the Knock, signed in to the console, wherever the address went.

### Answering

- Only the person an AI acts for answers its Knocks. The AI acts on their behalf, within their rights: they would run the operation themselves in the console. An administrator does not answer for another person, and the AI of a member is not offered the operations of administrators anyway.
- `approve_knock` and `deny_knock` are for people only (`PeopleOnly`): no AI answers a Knock, its own or another's.
- Approving runs the operation as the AI, with the input of the Knock, in the transaction that marks the Knock answered. The registry checks the call again, as it stands then: the role and the state of the person, the reach of the AI, the input against the records. If the operation fails, the Knock is marked failed with its message, and nothing else changes.
- The output of the operation goes to the person who approves, in the console, and never to the AI. Approved, `create_user` gives the person the activation link of the new account, which an AI never gets.
- A Knock ends without running when its person denies it, when it expires, or when its AI loses access before an answer: its token revoked, its connection cut, its person disabled.
- A Knock is in one of six states: `waiting`, `done`, `failed`, `denied`, `expired`, `canceled`.

### Notification

- The console shows how many Knocks wait in its navigation and in the title of its tab. It asks again every 30 seconds while it is open, and when its tab comes back to the front.
- Notifications by e-mail come with e-mail, after the V0.

### Logbook

- The call of an AI that knocks is logged under its operation with the outcome `knocked`, and the number of its Knock. Until batch 5, these calls were `refused`.
- An approved Knock is logged under its operation, as the call of the AI on behalf of its person, through the console, with the number of the Knock: the Logbook shows the AI, the person who approved, and the result.
- A denial is logged as `deny_knock`, by the person. An expiry and a cancellation are not actions: the Knock shows them, and the Logbook does not.

### New operations

- `purge_window` deletes a Window in the trash for good: its versions, its files, the passes of its Key. Its name becomes free. The members of its Room purge it, as they put it in the trash. A Window must be in the trash before it is purged.
- `delete_room` deletes a team Room for good, with its members and its Windows, those in its trash included, and their files. Its name becomes free. Its owners delete it, and the administrators of the Cahut, who manage every Room. A personal Room is not deleted: it keeps its person. The Logbook keeps the entries about the Room; its former members no longer read those that are not about them.
- `set_ai_reach` sets the scope and the Rooms of a connected AI. A person sets those of their own AIs, at once, in the console. An AI sets its own: narrowing acts at once, widening knocks. It is the one operation that changes things which an AI with the `read` scope is offered: it is how that AI asks for `write`. An AI with a personal token already reaches everything its person can, and has nothing to widen. This replaces, in [ADR 16](0016-oauth-for-ai-clients.md), disconnecting an AI and connecting it again to widen its reach.

## Consequences

- The refusals of [ADR 13](0013-people-sessions-and-rights.md) and [ADR 19](0019-window-versions-access-and-declaration.md) become Knocks. AIs no longer send their person to do the work in the console: they send them to approve it.
- A call that knocks runs twice, once to check it and once when approved. The operations that knock are rare and small.
- Every new operation that knocks must keep its effects on the disk until after the commit. The tests of the registry check that a rolled back call leaves the disk as it was.
- The person approves what their AI asked, in the words of the console, with the details of the input. The console writes each operation that knocks as a sentence, and shows any other with its name and its input.
- The address of a Knock goes through the AI and its vendor. It is a number, not a capability.
- An AI that its person has stopped trusting cannot get a Knock approved after the fact: revoking it cancels its Knocks.

## Alternatives considered

- **Any administrator answers.** A team could approve the requests of an absent person, but an administrator would answer for an AI that acts for someone else, on content that administrators do not manage.
- **Knocks checked only at approval.** Simpler, with no rollback, but the person would approve calls that then fail, and the AI would learn its mistakes a day later.
- **A result in the shape of the operation's output, with a status.** No error for the AI, but it would break the output schemas of the tools, and a client could take the call as done.
- **An approval link that carries a secret,** approved in one click without signing in. Faster, but whoever reads the conversation, or the AI itself, could approve.
- **An expiry of an hour, or of seven days.** An hour is short without e-mail: a person away from the console misses it. Seven days leaves requests waiting long after their conversation ended.
- **A Knock that the AI must call again once approved,** so that the AI runs the operation itself. The AI would get the output, and the call would depend on the conversation still being open.
