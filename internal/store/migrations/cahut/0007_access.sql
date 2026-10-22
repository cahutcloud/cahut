-- Each Window opens to its Room, to everyone signed in to its Cahut, to
-- the holders of its Key, or to the public, and readers sign in on the
-- port of the Windows. See docs/adr/0019-window-versions-access-and-declaration.md.

-- The access of a Window may end, after which the Window opens to its Room
-- only. Windows published before this migration open to their Room only.
ALTER TABLE windows ADD COLUMN access TEXT NOT NULL DEFAULT 'room' CHECK (access IN ('room', 'internal', 'key', 'public'));
ALTER TABLE windows ADD COLUMN access_ends_at TEXT;
ALTER TABLE windows ADD COLUMN key_hash TEXT; -- Argon2id, as a PHC string

-- A pass is what a reader gets for typing the Key of a Window: a cookie of
-- its own, whose SHA-256 hash only is kept. A new Key ends the passes of
-- the old one.
CREATE TABLE window_passes (
  id          INTEGER PRIMARY KEY,
  window_id   INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  hash        BLOB NOT NULL UNIQUE,
  created_at  TEXT NOT NULL,
  expires_at  TEXT NOT NULL
) STRICT;

CREATE INDEX window_passes_window ON window_passes (window_id);

-- A reader session lets a person open Windows on the port of the Windows,
-- in a cookie. It lives and ends as a session of the console does.
ALTER TABLE sessions ADD COLUMN reader INTEGER NOT NULL DEFAULT 0 CHECK (reader IN (0, 1));

-- Single-use codes, valid a minute, that the console hands to the port of
-- the Windows to open a reader session. Only a SHA-256 hash is kept.
CREATE TABLE reader_codes (
  hash        BLOB PRIMARY KEY,
  user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  expires_at  TEXT NOT NULL
) STRICT;

-- Random secrets of the Cahut, drawn on first use, such as the key that
-- signs the capabilities of framed Windows.
CREATE TABLE secrets (
  name   TEXT PRIMARY KEY,
  value  BLOB NOT NULL
) STRICT;
