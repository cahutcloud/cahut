-- AI clients connect over OAuth, to the Rooms and the scope a person
-- grants them. See docs/adr/0016-oauth-for-ai-clients.md.

-- Clients that registered dynamically (RFC 7591). Clients known by a Client
-- ID Metadata Document are fetched, not stored. A client that receives no
-- connection within 7 days is deleted.
CREATE TABLE oauth_clients (
  id             TEXT PRIMARY KEY,  -- the client ID, random
  name           TEXT NOT NULL,
  uri            TEXT,              -- its web page, as it gives it
  redirect_uris  TEXT NOT NULL,     -- a JSON array
  created_at     TEXT NOT NULL,
  connected_at   TEXT               -- when a person last connected it
) STRICT;

-- A connection is what a person granted an AI client: a scope, and every
-- Room of theirs or some of them. The console calls it a connected AI.
CREATE TABLE connections (
  id             INTEGER PRIMARY KEY,
  user_id        INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  client_id      TEXT NOT NULL,
  client_name    TEXT NOT NULL,
  redirect_host  TEXT NOT NULL,
  scope          TEXT NOT NULL CHECK (scope IN ('read', 'write')),
  all_rooms      INTEGER NOT NULL CHECK (all_rooms IN (0, 1)),
  created_at     TEXT NOT NULL,
  last_used_at   TEXT,
  ended_at       TEXT,
  end_reason     TEXT
) STRICT;

CREATE INDEX connections_user ON connections (user_id);

-- The Rooms of a connection that does not reach every Room of its person.
CREATE TABLE connection_rooms (
  connection_id  INTEGER NOT NULL REFERENCES connections (id) ON DELETE CASCADE,
  room_id        INTEGER NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
  PRIMARY KEY (connection_id, room_id)
) STRICT;

-- Authorization codes, used once. Only a SHA-256 hash of each code is kept.
CREATE TABLE oauth_codes (
  hash           BLOB PRIMARY KEY,
  connection_id  INTEGER NOT NULL REFERENCES connections (id) ON DELETE CASCADE,
  client_id      TEXT NOT NULL,
  redirect_uri   TEXT NOT NULL,
  challenge      TEXT NOT NULL,  -- PKCE, S256
  resource       TEXT NOT NULL,
  expires_at     TEXT NOT NULL,
  used_at        TEXT
) STRICT;

-- Access and refresh tokens. Only a SHA-256 hash of each token is kept. A
-- refresh token is replaced at each use, and kept a while after, so that a
-- replay is recognized.
CREATE TABLE oauth_tokens (
  id             INTEGER PRIMARY KEY,
  hash           BLOB NOT NULL UNIQUE,
  connection_id  INTEGER NOT NULL REFERENCES connections (id) ON DELETE CASCADE,
  kind           TEXT NOT NULL CHECK (kind IN ('access', 'refresh')),
  created_at     TEXT NOT NULL,
  expires_at     TEXT NOT NULL,
  replaced_at    TEXT
) STRICT;

CREATE INDEX oauth_tokens_connection ON oauth_tokens (connection_id);

-- The connection of an AI that acted, next to the personal token of one.
ALTER TABLE logbook ADD COLUMN connection_id INTEGER;
