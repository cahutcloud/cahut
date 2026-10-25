# 30. Each Cahut has a kind and limits, set by the administrators of its server

- Status: Accepted, extended by [ADR 42](0042-the-legal-pages-of-a-cahut.md) (the links to the legal pages of each Cahut)
- Date: 2026-10-24

## Context

[ADR 8](0008-several-cahuts-per-server.md) lets a server hold several Cahuts: one per brand, per client or per team, or, for a host, one per organization. It left quotas for later, and one Cahut can fill the disk for all the others.

A host who runs Cahuts for organizations that will not run a server themselves needs more. It gives each organization a share of a server: so many people, so much disk. It holds hundreds of Cahuts on a few servers, and none of them may take what the others were given.

Some organizations also want a Cahut where people work alone, side by side: the clients of one consultant, the students of one course. Each person keeps their own Windows, and none should learn who the others are.

Today, any person of a Cahut lists the others with their names (`list_users`). Any person creates a team Room and adds anyone to it, and a Window can open to everyone signed in to its Cahut ([ADR 19](0019-window-versions-access-and-declaration.md)). Rooms and people share one set of names ([ADR 9](0009-rooms-and-windows-as-paths.md)): creating a Room under a name that someone holds answers that the name is taken, by a Room or a person.

## Decision

### The settings of a Cahut

- Each Cahut has a kind and limits, kept in its own database, so that they travel with it when it is backed up or moved.
- Only the administrators of the server set them, through `set_cahut_settings`, an operation of the Cahut at their level ([ADR 1](0001-operation-registry.md)), from the console of the Cahut or from the command line. No AI is offered it.
- Every change goes to the Logbook of the Cahut, with the values before and after, so that its administrators see when their limits change. Setting the old values again undoes it.
- `get_cahut` gives the kind, the limits and what the Cahut uses, to people and to AIs, so that an AI that is refused knows why. In an individual Cahut, a member gets their own use and their own limits only.
- A Cahut is collaborative and has no limit until the administrators of its server say otherwise. Every Cahut that exists stays as it is.

### Two kinds

- **Collaborative**, as every Cahut is today: people create team Rooms, add one another to them, and open Windows to the whole Cahut.
- **Individual**: people share the Cahut, not their work, and do not see one another.
  - Every Room has one member, its owner. A person creates Rooms of their own beside their personal Room. Nobody can be added to a Room: the operations that manage the members of a Room are refused, and not offered to AIs.
  - Only the administrators of the Cahut list its people. No operation open to a member names another person.
  - A Window opens to its Room, to anyone with its Key, or to the public, and never to everyone signed in to the Cahut.
  - The name of a Room starts with the handle of its owner and a hyphen, such as `bob-projects`. Cahut adds the prefix to a name given without it. A person never meets the names of the others, and the address says whose Room it is. A clash remains possible only between two handles where one extends the other, such as `bob` and `bob-x`; the refusal then says that the name is not available, without saying who holds it.
  - The administrators of the Cahut see everyone, every Room and the whole Logbook, and open only the Windows of their own Rooms, as [ADR 13](0013-people-sessions-and-rights.md) has it. They run the Cahut and are not peers: where they act in the Room of a person, for instance to create it, that person's Logbook names them.
- A collaborative Cahut becomes individual only when no Room has more than one member and no Window opens to the whole Cahut. The refusal names those that do. An individual Cahut can always become collaborative. Rooms keep their names either way.

### Limits

Each limit is optional. Without it, nothing is counted.

| Limit | What counts |
| --- | --- |
| `people` | The people who are active or wait for their activation, administrators included, and not those who are disabled |
| `admins` | The administrators of the Cahut, counted the same way |
| `rooms` | Team Rooms, or, in an individual Cahut, the Rooms beside the personal ones. A personal Room comes with its person, and `people` already bounds them |
| `windows` | The Windows out of the trash, in the whole Cahut |
| `windows_per_room` | The Windows out of the trash, in one Room, personal Rooms included |
| `disk` | The bytes of every version of every Window, those in the trash included, and of the files of the house style, its previous version included |

In an individual Cahut, `rooms_per_person`, `windows_per_person` and `disk_per_person` also bound each person, over the Rooms they own, the same for every person. No person can take what the Cahut gives to all, and a refusal speaks of the person's own use. A collaborative Cahut has no limit per person, since its Rooms belong to several people.

- Each operation checks the limits it could cross, in its transaction: creating a person, enabling one or making one an administrator; creating a Room; publishing, updating or restoring a Window; setting the house style. A Cahut has a single write connection, so two calls at once cannot cross a limit together.
- A call that would cross a limit is refused with `operation.ErrForbidden`, and a message that says what to do: put Windows in the trash, purge some, or ask the administrators of the server for more. In an individual Cahut, a refusal for a limit of the whole Cahut gives no count, which would tell a member how much the others use.
- Lowering a limit below what the Cahut uses deletes nothing. The Cahut keeps what it holds, refuses what would add to it, and the console shows the overrun.
- The disk is counted from the sizes that the database already keeps for each version of a Window and each file of the house style: an exact count, read in the transaction, without walking the disk. The console takes it in gigabytes of 10⁹ bytes, as hosts sell it.

## Consequences

- The Cahuts that people install for themselves change in nothing: collaborative, without limits.
- The quotas that ADR 8 left for later exist: a host gives each organization its share of a server.
- A refusal for a limit tells an AI what to do, and `get_cahut` lets it look before it publishes. The instructions of the MCP server do not grow.
- An individual Cahut offers its AIs fewer tools, without those that manage the members of a Room. The tools offered change when the kind changes.
- The tests of every operation run over the matrix of callers of ADR 13 in an individual Cahut too. A test walks every operation that a member may call in an individual Cahut of two people, and fails when a result names the other.
- The disk counted is the content of files, not the blocks they take, and leaves out the rest of the database and the Logbook, which people do not control and which weigh little. The check of free space on the server stays beside it.
- A migration adds the settings to the database of every Cahut: collaborative, without limits.
- A view of every Cahut of a server with its use, and the operations that a program calls to create, suspend, delete or move Cahuts for a host, come with later records.
- This record extends ADR 8 (limits per Cahut), ADR 9 (the names of Rooms in an individual Cahut), ADR 13 (who lists people, and who can be added to a Room) and ADR 19 (no access to the whole Cahut in an individual Cahut).

## Alternatives considered

- **A Village: several Cahuts under one domain**, at `<domain>/<cahut>/`, with an administrator above those of each Cahut. The Cahuts of a Village would share one origin, and with it the storage where the console keeps its session tokens ([ADR 20](0020-console-under-a-path-of-port-443.md)), for every organization of the Village. The root-relative links of Windows to the house style under `/-/` ([ADR 22](0022-house-style-of-each-cahut.md)) would need a prefix per Cahut, and a Window could not move between a Village and a domain of its own. The OAuth issuer and the MCP endpoint would take a path. All the Cahuts of a Village would stay on one server, whose address its domain names. A subdomain per Cahut, such as `acme.<the host's domain>`, keeps ADR 8 as it is, and the administrators of the server play the part of the Village's.
- **The settings in `server.db`**, beside the domain of each Cahut. The server would hold what it grants, but every check would read a second database, outside the transaction of the call, and the settings would not travel with the directory of the Cahut.
- **A switch for each rule**: team Rooms, members, access to the whole Cahut, the list of people. More combinations, most of which make no sense, such as people hidden from one another who share Rooms. Two kinds make a promise that a person can read and a test can check.
- **An individual Cahut as a limit of zero team Rooms.** No new setting, but the people of an individual Cahut create Rooms of their own, and hiding people takes more than a count.
- **Names of Rooms left free, with a refusal that says only "not available".** Shorter addresses, but a person could still try the handles of others, one by one, to learn who is there.
- **The trash in the count of Windows.** A Cahut at its limit would have to purge to publish again, which cannot be undone and knocks first for an AI.
- **The disk measured on the disk.** It would count blocks and every file, but walking the files of a Cahut is slow, and happens outside the transaction where the limit is checked.
- **Limits that differ from one person to another.** One person could get more than the others. It can come later, if hosts ask for it.
