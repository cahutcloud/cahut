-- The records of the server as a whole. Each Cahut keeps its own records in
-- a database of its own, in cahuts/<id>/. Times are UTC text in the layout
-- of store.FormatTime.

CREATE TABLE cahuts (
  id          INTEGER PRIMARY KEY,
  domain      TEXT NOT NULL UNIQUE,
  created_at  TEXT NOT NULL
) STRICT;

-- The Logbook of the operations on the server, laid out as the Logbook of
-- a Cahut.
CREATE TABLE logbook (
  id           INTEGER PRIMARY KEY,
  at           TEXT NOT NULL,
  operation    TEXT NOT NULL,
  actor        TEXT NOT NULL CHECK (actor IN ('admin', 'person', 'ai')),
  user_id      INTEGER,
  user_handle  TEXT,
  token_id     INTEGER,
  client       TEXT,
  via          TEXT NOT NULL,
  target       TEXT,
  detail       TEXT,
  outcome      TEXT NOT NULL CHECK (outcome IN ('ok', 'refused', 'failed')),
  error        TEXT
) STRICT;

CREATE INDEX logbook_at ON logbook (at);
