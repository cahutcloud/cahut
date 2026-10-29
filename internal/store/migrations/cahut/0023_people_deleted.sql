-- foreign keys off
-- An administrator deletes a person for good. The Logbook keeps the id of
-- each person, whose entries they read: users is made again with
-- AUTOINCREMENT, so that a deleted person never hands their id to the next
-- one, as rooms and windows were in 0010. The Windows and the versions that
-- the person published in the team Rooms that stay keep no author: their
-- created_by and published_by become NULL, and so windows and
-- window_versions are made again. See docs/adr/0047-deleting-a-person.md.

CREATE TABLE users_again (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  handle         TEXT NOT NULL UNIQUE,
  name           TEXT NOT NULL,
  role           TEXT NOT NULL CHECK (role IN ('admin', 'member')),
  created_at     TEXT NOT NULL,
  password_hash  TEXT, -- Argon2id, as a PHC string
  disabled_at    TEXT,
  email          TEXT,
  notices_off    INTEGER NOT NULL DEFAULT 0 CHECK (notices_off IN (0, 1))
) STRICT;

INSERT INTO users_again (id, handle, name, role, created_at, password_hash, disabled_at, email, notices_off)
  SELECT id, handle, name, role, created_at, password_hash, disabled_at, email, notices_off FROM users;

DROP TABLE users;
ALTER TABLE users_again RENAME TO users;

CREATE UNIQUE INDEX users_email ON users (email COLLATE NOCASE);

CREATE TABLE windows_again (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id         INTEGER NOT NULL REFERENCES rooms (id),
  name            TEXT NOT NULL,
  title           TEXT NOT NULL,
  kind            TEXT NOT NULL CHECK (kind IN ('document', 'deck', 'dashboard', 'site')),
  version         INTEGER NOT NULL,
  size            INTEGER NOT NULL, -- bytes of the current version's files
  file_count      INTEGER NOT NULL,
  created_by      INTEGER REFERENCES users (id) ON DELETE SET NULL, -- NULL once that person is deleted
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL,
  previous        INTEGER, -- the number of the previous version, if any
  trashed_at      TEXT,
  access          TEXT NOT NULL DEFAULT 'room' CHECK (access IN ('room', 'internal', 'key', 'public')),
  access_ends_at  TEXT,
  key_hash        TEXT, -- Argon2id, as a PHC string
  access_set_by   INTEGER REFERENCES users (id) ON DELETE SET NULL,
  UNIQUE (room_id, name)
) STRICT;

INSERT INTO windows_again (id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash, access_set_by)
  SELECT id, room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at,
    previous, trashed_at, access, access_ends_at, key_hash, access_set_by
  FROM windows;

-- AUTOINCREMENT goes on from the largest id ever given, not from the largest
-- left: a Window purged before keeps its id.
DELETE FROM sqlite_sequence WHERE name = 'windows_again';
INSERT INTO sqlite_sequence (name, seq)
  SELECT 'windows_again', max(coalesce((SELECT seq FROM sqlite_sequence WHERE name = 'windows'), 0),
    coalesce((SELECT max(id) FROM windows_again), 0));

DROP TABLE windows;
ALTER TABLE windows_again RENAME TO windows;

CREATE TABLE window_versions_again (
  window_id     INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  number        INTEGER NOT NULL,
  size          INTEGER NOT NULL, -- bytes of its files
  file_count    INTEGER NOT NULL,
  published_by  INTEGER REFERENCES users (id) ON DELETE SET NULL, -- NULL once that person is deleted
  published_at  TEXT NOT NULL,
  declaration   TEXT, -- JSON
  PRIMARY KEY (window_id, number)
) STRICT;

INSERT INTO window_versions_again (window_id, number, size, file_count, published_by, published_at, declaration)
  SELECT window_id, number, size, file_count, published_by, published_at, declaration FROM window_versions;

DROP TABLE window_versions;
ALTER TABLE window_versions_again RENAME TO window_versions;
