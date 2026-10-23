-- An AI that calls an operation that cannot be undone, or that changes what
-- a person may do, makes a Knock: a request that waits for the person it
-- acts for, who approves or denies it in the console. See
-- docs/adr/0021-ais-knock-first.md.

-- The input of a Knock holds no secret. A Knock that waits past expires_at
-- has expired, and one whose AI lost access is cancelled: neither is
-- written, both are read from the time and from the AI's token or
-- connection.
CREATE TABLE knocks (
  id             INTEGER PRIMARY KEY,
  user_id        INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  token_id       INTEGER REFERENCES tokens (id) ON DELETE CASCADE,
  connection_id  INTEGER REFERENCES connections (id) ON DELETE CASCADE,
  client         TEXT NOT NULL,     -- the AI, as the Logbook names it
  operation      TEXT NOT NULL,
  input          TEXT NOT NULL,     -- the input of the call, in JSON
  target         TEXT,              -- what the call acts on, as the Logbook says it
  detail         TEXT,
  room_id        INTEGER,           -- the Room the call concerns; no key, as it may be deleted
  created_at     TEXT NOT NULL,
  expires_at     TEXT NOT NULL,
  state          TEXT NOT NULL DEFAULT 'waiting' CHECK (state IN ('waiting', 'done', 'failed', 'denied')),
  answered_at    TEXT,
  error          TEXT,              -- why the operation failed, once approved
  CHECK ((token_id IS NULL) != (connection_id IS NULL))
) STRICT;

CREATE INDEX knocks_user ON knocks (user_id, state);

-- The Logbook records a call that knocks with the outcome knocked, and
-- the Knock that a call made, answered or ran for. SQLite cannot change the
-- check of a column: the table is made again, with its entries.
CREATE TABLE logbook_knocks (
  id             INTEGER PRIMARY KEY,
  at             TEXT NOT NULL,
  operation      TEXT NOT NULL,
  actor          TEXT NOT NULL CHECK (actor IN ('admin', 'person', 'ai')),
  user_id        INTEGER, -- the person, or the person an AI acts for
  user_handle    TEXT,
  token_id       INTEGER,
  client         TEXT,    -- the AI, as its token and its software name it
  via            TEXT NOT NULL,
  target         TEXT,
  detail         TEXT,
  outcome        TEXT NOT NULL CHECK (outcome IN ('ok', 'knocked', 'refused', 'failed')),
  error          TEXT,
  room_id        INTEGER,
  connection_id  INTEGER,
  knock_id       INTEGER
) STRICT;

INSERT INTO logbook_knocks (id, at, operation, actor, user_id, user_handle, token_id, client, via, target, detail,
    outcome, error, room_id, connection_id)
  SELECT id, at, operation, actor, user_id, user_handle, token_id, client, via, target, detail,
    outcome, error, room_id, connection_id
  FROM logbook;

DROP TABLE logbook;
ALTER TABLE logbook_knocks RENAME TO logbook;

CREATE INDEX logbook_at ON logbook (at);
CREATE INDEX logbook_sign_in ON logbook (user_id, at) WHERE operation = 'sign_in' AND outcome = 'refused';
CREATE INDEX logbook_room ON logbook (room_id);
