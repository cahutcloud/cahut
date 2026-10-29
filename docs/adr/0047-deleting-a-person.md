# 47. The administrators of a Cahut delete a person for good, with the Rooms that are theirs alone

- Status: Accepted
- Date: 2026-10-29

## Context

The administrators of a Cahut add people, change their role and disable them ([ADR 13](0013-people-sessions-and-rights.md)). A disabled person can no longer sign in, and their AIs can no longer act, but everything of theirs stays: their account, their address, their personal Room and its Windows, the team Rooms they own. Nothing removes a person who left, nor frees their handle, nor takes away what they published when the Cahut should no longer hold it.

A person leaves traces outside their own Rooms. They publish Windows and versions in team Rooms that others share, invite people to read them, set the house style, and the Logbook names them in its entries. The Logbook keeps the id of each person, without a foreign key, and a person reads the entries of their id. The ids of people are not AUTOINCREMENT: a deleted person would hand theirs to the next one, who would read the entries of the first.

## Decision

- `delete_user` deletes a person for good. The administrators of the Cahut, and those of the server, call it; an AI that calls it knocks first ([ADR 21](0021-ais-knock-first.md)). No one deletes themself, nor the last administrator of the Cahut who is not disabled.
- With the person go their account, their address, their sessions, their tokens, the AIs they connected and their Knocks, their personal Room, and each team Room they are the only owner of, with its members. The Windows of these Rooms go with them, those in their trash included, with their versions and their files, once the call commits, as `delete_room` deletes them.
- A team Room that has another owner stays, without the person. The Windows and the versions they published in it stay too, and name no one: `created_by` and `published_by` are empty, and the console leaves out whom they were published for. The invitations they sent and the house styles they set stay, without their name.
- The Logbook keeps every entry of the person, with their id and their handle as they were. The people of the Cahut are made again with AUTOINCREMENT, so that an id is never given twice, as rooms and windows were in migration 0010. Their handle becomes free, and a new person may take it, with a new personal Room.
- The console asks for the handle of the person, typed in full, before it deletes them, as it asks for the name of a Room.

## Consequences

- A Cahut can let go of a person and of what was theirs alone, without asking each team to clean up after them.
- What a team made together stays where it was, whoever published it, at the cost of an author that the console no longer names.
- The migration makes `users`, `windows` and `window_versions` again, with their rows, before the server serves: on a large Cahut, the first start after the update takes longer.
- A person deleted by mistake does not come back: the Knock of an AI and the typed handle in the console are the only guards. `disable_user` remains the way to stop a person and keep what is theirs.

## Alternatives considered

- Keeping the row of a deleted person, emptied of their name and address, so that their Windows keep an author. Their handle would stay taken, and a deleted person would linger in every list.
- Deleting every Window the person published, in every Room. It would take away the work of a team because one of its people left.
- Deleting every team Room the person owns. A Room that another owner keeps belongs to them as much.
