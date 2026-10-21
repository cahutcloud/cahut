-- Times are UTC text in the layout of store.FormatTime.

CREATE TABLE users (
  id          INTEGER PRIMARY KEY,
  handle      TEXT NOT NULL UNIQUE,
  name        TEXT NOT NULL,
  role        TEXT NOT NULL CHECK (role IN ('admin', 'member')),
  created_at  TEXT NOT NULL
) STRICT;

CREATE TABLE rooms (
  id          INTEGER PRIMARY KEY,
  kind        TEXT NOT NULL CHECK (kind IN ('personal', 'team')),
  name        TEXT NOT NULL,
  created_at  TEXT NOT NULL
) STRICT;

CREATE TABLE room_members (
  room_id  INTEGER NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
  user_id  INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  role     TEXT NOT NULL CHECK (role IN ('owner', 'member')),
  PRIMARY KEY (room_id, user_id)
) STRICT;

CREATE INDEX room_members_user ON room_members (user_id);

-- Personal tokens for AIs that run in a terminal. Only a SHA-256 hash of
-- each token is kept.
CREATE TABLE tokens (
  id            INTEGER PRIMARY KEY,
  user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT,
  revoked_at    TEXT
) STRICT;

CREATE TABLE windows (
  id          INTEGER PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE, -- the DNS label of the Window's origin
  room_id     INTEGER NOT NULL REFERENCES rooms (id),
  title       TEXT NOT NULL,
  kind        TEXT NOT NULL CHECK (kind IN ('document', 'deck', 'dashboard', 'site')),
  version     INTEGER NOT NULL,
  size        INTEGER NOT NULL, -- bytes of the current version's files
  file_count  INTEGER NOT NULL,
  created_by  INTEGER NOT NULL REFERENCES users (id),
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
) STRICT;

CREATE INDEX windows_room ON windows (room_id);

-- The Logbook keeps ids and names as they were, without foreign keys, so
-- that an entry outlives what it refers to.
CREATE TABLE logbook (
  id           INTEGER PRIMARY KEY,
  at           TEXT NOT NULL,
  operation    TEXT NOT NULL,
  actor        TEXT NOT NULL CHECK (actor IN ('admin', 'person', 'ai')),
  user_id      INTEGER, -- the person, or the person an AI acts for
  user_handle  TEXT,
  token_id     INTEGER,
  client       TEXT,    -- the AI, as its token and its software name it
  via          TEXT NOT NULL,
  target       TEXT,
  detail       TEXT,
  outcome      TEXT NOT NULL CHECK (outcome IN ('ok', 'refused', 'failed')),
  error        TEXT
) STRICT;

CREATE INDEX logbook_at ON logbook (at);
