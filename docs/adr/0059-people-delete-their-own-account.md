# 59. People delete their own account, unless the administrators of their Cahut turn it off

- Status: Accepted
- Date: 2026-11-02

## Context

The administrators of a Cahut delete a person for good, with the Rooms that are theirs alone ([ADR 47](0047-deleting-a-person.md)). The person cannot: ADR 47 lets no one delete themself. A person who wants to leave asks an administrator, and waits. A person who signed up to a Cahut on their own, as in an individual Cahut ([ADR 30](0030-kind-and-limits-of-each-cahut.md)), expects to close their account as they opened it, and the law of many countries gives them the right to have their data erased.

Leaving is not the same as being removed. An administrator who deletes a person keeps the work of the teams: the team Rooms that have another owner stay, with the Windows the person published in them, and the Windows they kept for themselves open to their Room ([ADR 53](0053-windows-for-their-author-alone.md)). A person who leaves should take what is theirs alone, without taking away a Room that a team works in, and without leaving their drafts open to that team.

Not every Cahut wants it. An organization may keep the Windows of the people who leave, and handle each departure itself. A host may delete accounts through an app of its own, with a delay that its terms promise.

## Decision

### Who deletes their account

- `delete_account` deletes the account of the person who calls, for good and at once. It is for people only: no AI is offered it, with or without a Knock ([ADR 21](0021-ais-knock-first.md)), and the administrator on the command line is not a person.
- The person gives their current password and types their handle. Cahut checks the password before the transaction, and refuses a handle that is not theirs.
- `check_account_deletion` says, and changes nothing, what would go, what would stay, and what stops it. The console shows it before it asks for the password.
- The administrators of the Cahut allow it or not with `set_account_deletion`, from the console or the command line. Every Cahut allows it until they turn it off. It is for people only, as the e-mail relay is.

### What stops it

Each refusal has a code, which the console says in its language ([ADR 37](0037-refusals-with-a-code-for-the-console.md)):

| Code | When |
| --- | --- |
| `account_deletion_off` | The administrators of the Cahut turned it off: the person asks one of them |
| `cahut_needs_an_admin` | The person is the last administrator of the Cahut who is not disabled, as when someone else would take that role from them |
| `account_rooms_shared` | Team Rooms that the person is the only owner of have other members. The refusal names them: the person gives each another owner, or deletes it, first |
| `account_handle_mismatch` | The handle typed is not that of the person |

A wrong password is refused as `change_password` refuses it.

### What goes and what stays

- With the person go, as ADR 47 has it, their account, their address, their sessions, this one included, their tokens, the AIs they connected and their Knocks. Their personal Room goes, and each team Room whose only member they are, with the Windows of these Rooms, those in their trash included, and their versions, files and data. In an individual Cahut, every Room of the person is such a Room. A Window that its moderators withdrew ([ADR 54](0054-reporting-and-withdrawing-a-window.md)) goes as any other, with its reports, as when its members purge it.
- The Windows they kept for themselves (`author`, ADR 53) in the Rooms that stay go with them too, rather than open to their Room: no one else has seen them.
- The team Rooms that have another owner stay, without the person. What they published in them stays, without an author, and so do their invitations and share links, without their name, as ADR 47 has it.
- The Logbook keeps every entry of the person, with their id and their handle as they were. Their handle becomes free.
- When the Cahut sends e-mail, a message tells the person, at their address, that their account is deleted.

### In the console

- The page of the account ends with a section that deletes it, shown only where the Cahut allows it. It opens on what `check_account_deletion` says: the Rooms that go, with how many Windows, what stays, and what stops it, with a link to each Room to change. The form then asks for the password and for the handle, typed in full. Its button stays off until the handle is right.
- Once the account is deleted, the console forgets its session and shows the sign-in page, which says that the account is deleted.
- The page of the people tells the administrators of the Cahut whether people delete their own account, and lets them change it.

## Consequences

- A person leaves a Cahut without asking anyone, where its administrators allow it, and takes with them what was theirs alone.
- A team does not lose a Room because one person left: its only owner names another first, or deletes it knowingly.
- A deleted account does not come back. The password, the typed handle and what the console shows before are the only guards, as for a Room or a Window deleted for good.
- Every Cahut allows it after the update that brings it. A host or an organization that deletes accounts its own way turns it off right after the update.
- An app that keeps the accounts of a Cahut beside it, such as a sign-up app, learns of a deletion from the Logbook only, unless its Cahut turns it off.

## Alternatives considered

- **A delay of grace**: the account disabled at once, and deleted after 30 days unless the person signs in again. It forgives a mistake, but it needs a new state of the account, Windows closed while they wait, and a task that runs every hour. A host that promises one keeps it in its own app, with `disable_user` then `delete_user`.
- **Deleting the team Rooms that the person owns alone, members and all**, as ADR 47 does. A team would lose its work by the decision of one of its people.
- **Opening the Windows kept for themselves to their Room**, as ADR 47 and ADR 53 do. The team would see drafts that the person never shared.
- **`delete_user` on oneself, with a password.** It is an operation of the administrators, offered to AIs with a Knock, and its refusals are not those of a person who leaves.
- **Always allowed, without a setting.** An organization could not keep the Windows of the people who leave.
- **A setting of the administrators of the server**, beside the kind and the limits ([ADR 30](0030-kind-and-limits-of-each-cahut.md)). Who may leave, and with what, is the business of the Cahut, not of its host.
