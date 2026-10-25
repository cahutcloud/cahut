# 39. People change their name and their handle, and their personal Room follows

- Status: Accepted
- Date: 2026-10-25

## Context

An administrator gives each person of a Cahut a handle and a name when they add them ([ADR 13](0013-people-sessions-and-rights.md)). The person signs in with the handle, and the people of the Cahut see the name. Since batch 9, a person changes their e-mail address themselves, through a link sent to the new one ([ADR 31](0031-e-mail-through-a-relay.md)), but nothing changed a name or a handle: a typo, a new name or a handle chosen in haste stayed.

A handle is also the name of the person's personal Room, and so the first part of the address of its Windows, `<domain>/<handle>/<window>/` ([ADR 9](0009-rooms-and-windows-as-paths.md)). In an individual Cahut, the name of each Room of a person starts with their handle and a hyphen ([ADR 30](0030-kind-and-limits-of-each-cahut.md)). The records of a Cahut hold people and Rooms by their ids: the files of Windows, the members of Rooms, sessions, tokens, connected AIs and the Logbook. Only the handle of a person and the name of a Room hold the words.

## Decision

### The name

- A person changes their name on the Account page of the console, with `change_name`. A name is not empty and holds 200 bytes at most, as the title of a Room does; `create_user` checks the same.
- The personal Room takes the new name as its title, unless the person gave it another title.
- The name stays one field, as the person writes it: a split into a first name and a last name fits some names and not others.

### The handle

- A person changes their handle on the Account page, with `change_handle`, if the new one is free: of the form of a handle, not reserved, and held by no person and by no Room.
- Their personal Room takes the new handle as its name. In an individual Cahut, their other Rooms, whose names start with the old handle, take the new one in its place; if one of the new names is taken or too long, nothing changes and the person chooses another handle.
- The addresses of the Windows of these Rooms change with them, and the old ones stop working: there is no redirect. The old handle is free for another person or Room.
- The console says it before the change, with the old and the new addresses, and asks again on the spot.
- A Knock that waits for the person names the old addresses in its call: the person answers their Knocks first.
- Sessions, tokens and connected AIs go on: they hold the person and their Rooms by their ids. The console keeps the new handle and name in the session of its browser, and another browser takes them when it opens. The Logbook keeps the handle of each entry as it was.
- The person gets a notice by e-mail when the Cahut sends e-mail, as for a new password.

### Who changes them

- `change_name` and `change_handle` serve the console only, as `change_email` and `change_password` do: no AI changes who a person is.
- The e-mail address does not change: with a relay, the person changes it through a link sent to the new address; without one, the Cahut shows no feature of e-mail ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)), and an administrator sets it.

## Consequences

- A person fixes their name and handle without an administrator.
- A new handle breaks the links to the Windows of the person's Rooms that they gave out, public ones included. The person decides it, warned.
- The old addresses may later lead to the Windows of another person or Room that takes the old name, if they publish Windows of the same names there. Whoever follows an old link then opens what that Room shows to them.
- A Knock made before the change by another person's AI, which names the old handle, finds the person who holds it when it is approved: whoever approves it reads the handle on the Knock.
- The refusal of a taken handle tells that a person or a Room holds it, as a form of sign-up does. In an individual Cahut, where people do not see one another, it is the one place where one learns that a handle is held.
- The command line and the administrators do not change a handle or a name for a person: that can come later, with a Knock.

## Alternatives considered

- **Redirects from the old addresses**, with the old handle kept for the person. Shared links would keep working, and no one else could take the old addresses, but the handle would stay taken for good, and a table of old names and a redirect at each address would remain. The handle is the person's own, and they choose a clean break.
- **A personal Room that keeps its name**: only the handle changes. Links keep working, but the address no longer says whose the Room is, and the old handle stays taken by the Room.
- **A first name and a last name**: two fields for the console and for e-mail, and a migration that would split the names already there, which no rule splits well.
