-- The Room that a Logbook entry concerns, so that the members of a Room
-- read what happens in it. Like the other ids of the Logbook, it has no
-- foreign key: an entry outlives its Room.
-- See docs/adr/0013-people-sessions-and-rights.md.
ALTER TABLE logbook ADD COLUMN room_id INTEGER;

CREATE INDEX logbook_room ON logbook (room_id);
