-- The settings of the Cahut, which the administrators of its server set:
-- the limits on what it holds. A single row; a NULL limit counts nothing.
-- See docs/adr/0030-kind-and-limits-of-each-cahut.md.
CREATE TABLE settings (
  id                INTEGER PRIMARY KEY CHECK (id = 1),
  people            INTEGER CHECK (people >= 1),           -- people who are not disabled, administrators included
  admins            INTEGER CHECK (admins >= 1),           -- administrators who are not disabled
  rooms             INTEGER CHECK (rooms >= 0),            -- team Rooms
  windows           INTEGER CHECK (windows >= 0),          -- Windows out of the trash
  windows_per_room  INTEGER CHECK (windows_per_room >= 0), -- Windows out of the trash in one Room
  disk              INTEGER CHECK (disk >= 0),             -- bytes of the versions of Windows and of the files of the house style
  set_at            TEXT                                   -- NULL until the administrators of the server set them
) STRICT;

INSERT INTO settings (id) VALUES (1);
