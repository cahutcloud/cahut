-- foreign keys off
-- Deleting a Room or purging a Window must not hand its id to the next one:
-- the Logbook and the Knocks keep the id of a Room, whose members read its
-- entries, and the frames and passes of a Window name its id. SQLite gives
-- a new row the largest id plus one, unless the table has AUTOINCREMENT,
-- which only a new table can have: rooms and windows are made again, with
-- their rows. See docs/adr/0021-ais-knock-first.md.

CREATE TABLE rooms_again (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL UNIQUE,
  kind        TEXT NOT NULL CHECK (kind IN ('personal', 'team')),
  title       TEXT NOT NULL,
  created_at  TEXT NOT NULL
) STRICT;

INSERT INTO rooms_again (id, name, kind, title, created_at)
  SELECT id, name, kind, title, created_at FROM rooms;

DROP TABLE rooms;
ALTER TABLE rooms_again RENAME TO rooms;

CREATE TABLE windows_again (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id         INTEGER NOT NULL REFERENCES rooms (id),
  name            TEXT NOT NULL,
  title           TEXT NOT NULL,
  kind            TEXT NOT NULL CHECK (kind IN ('document', 'deck', 'dashboard', 'site')),
  version         INTEGER NOT NULL,
  size            INTEGER NOT NULL, -- bytes of the current version's files
  file_count      INTEGER NOT NULL,
  created_by      INTEGER NOT NULL REFERENCES users (id),
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  previous        INTEGER, -- the number of the previous version, if any
  trashed_at      TEXT,
  access          TEXT NOT NULL DEFAULT 'room' CHECK (access IN ('room', 'internal', 'key', 'public')),
  access_ends_at  TEXT,
  key_hash        TEXT, -- Argon2id, as a PHC string
  UNIQUE (room_id, name)
) STRICT;

INSERT INTO windows_again (id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash)
  SELECT id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash
  FROM windows;

DROP TABLE windows;
ALTER TABLE windows_again RENAME TO windows;
