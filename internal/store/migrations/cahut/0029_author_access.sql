-- foreign keys off
-- A Window may open to its author alone, the person it was first published
-- for, while they are in its Room: windows is made again, so that its
-- access admits author. The entries of the Logbook about such a Window go
-- to its author, not to the members of its Room: an entry names the person
-- who reads it in their place. See
-- docs/adr/0053-windows-for-their-author-alone.md.

CREATE TABLE windows_again (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id         INTEGER NOT NULL REFERENCES rooms (id),
  name            TEXT NOT NULL,
  title           TEXT NOT NULL,
  kind            TEXT NOT NULL CHECK (kind IN ('document', 'deck', 'dashboard', 'site')),
  version         INTEGER NOT NULL,
  size            INTEGER NOT NULL, -- bytes of the current version's files
  file_count      INTEGER NOT NULL,
  created_by      INTEGER REFERENCES users (id) ON DELETE SET NULL, -- its author; NULL once that person is deleted
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  previous        INTEGER, -- the number of the previous version, if any
  trashed_at      TEXT,
  access          TEXT NOT NULL DEFAULT 'room' CHECK (access IN ('author', 'room', 'internal', 'key', 'public')),
  access_ends_at  TEXT,
  key_hash        TEXT, -- Argon2id, as a PHC string
  access_set_by   INTEGER REFERENCES users (id) ON DELETE SET NULL,
  uid             TEXT NOT NULL DEFAULT '',
  description     TEXT NOT NULL DEFAULT '',
  UNIQUE (room_id, name),
  CHECK (access != 'author' OR created_by IS NOT NULL)
) STRICT;

INSERT INTO windows_again (id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash, access_set_by, uid, description)
  SELECT id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash, access_set_by, uid, description
  FROM windows;

-- AUTOINCREMENT goes on from the largest id ever given, not from the largest
-- left: a Window purged before keeps its id.
DELETE FROM sqlite_sequence WHERE name = 'windows_again';
INSERT INTO sqlite_sequence (name, seq)
  SELECT 'windows_again', max(coalesce((SELECT seq FROM sqlite_sequence WHERE name = 'windows'), 0),
    coalesce((SELECT max(id) FROM windows_again), 0));

DROP TABLE windows;
ALTER TABLE windows_again RENAME TO windows;

CREATE UNIQUE INDEX windows_uid ON windows (uid);

ALTER TABLE logbook ADD COLUMN reader_id INTEGER; -- the person who reads the entry in place of the members of its Room

CREATE INDEX logbook_reader ON logbook (reader_id) WHERE reader_id IS NOT NULL;
