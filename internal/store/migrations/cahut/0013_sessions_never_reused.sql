-- foreign keys off
-- The frame of a protected Window, and the preview of a Window in the
-- console, name the session or the pass they open for by its id. A session
-- or a pass that ends must not hand its id to the next one, which would
-- open the old frames again: sessions and window_passes are made again with
-- AUTOINCREMENT, as rooms and windows were in 0010.
--
-- A reader session also records the console session that opened it, so
-- that signing out of the console signs out of the Windows too, and the
-- codes that open reader sessions record it until then. See
-- docs/revue-securite-v0.md.

CREATE TABLE sessions_again (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT NOT NULL,
  expires_at    TEXT NOT NULL,
  reader        INTEGER NOT NULL DEFAULT 0 CHECK (reader IN (0, 1)),
  opened_by     INTEGER -- the console session that opened a reader session
) STRICT;

INSERT INTO sessions_again (id, user_id, hash, created_at, last_used_at, expires_at, reader)
  SELECT id, user_id, hash, created_at, last_used_at, expires_at, reader FROM sessions;

DROP TABLE sessions;
ALTER TABLE sessions_again RENAME TO sessions;

CREATE INDEX sessions_user ON sessions (user_id);
CREATE INDEX sessions_opened_by ON sessions (opened_by) WHERE opened_by IS NOT NULL;

CREATE TABLE window_passes_again (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  window_id   INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  hash        BLOB NOT NULL UNIQUE,
  created_at  TEXT NOT NULL,
  expires_at  TEXT NOT NULL
) STRICT;

INSERT INTO window_passes_again (id, window_id, hash, created_at, expires_at)
  SELECT id, window_id, hash, created_at, expires_at FROM window_passes;

DROP TABLE window_passes;
ALTER TABLE window_passes_again RENAME TO window_passes;

CREATE INDEX window_passes_window ON window_passes (window_id);

ALTER TABLE reader_codes ADD COLUMN session_id INTEGER;
