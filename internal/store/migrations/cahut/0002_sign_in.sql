-- People sign in with a password, which they choose from an activation
-- link, and keep a session. See docs/adr/0013-people-sessions-and-rights.md.

-- A person without a password has not activated their account yet. A
-- disabled person can neither sign in nor act through an AI.
ALTER TABLE users ADD COLUMN password_hash TEXT; -- Argon2id, as a PHC string
ALTER TABLE users ADD COLUMN disabled_at TEXT;

-- Single-use links where a person chooses their password. Only a SHA-256
-- hash of each code is kept, and a new link replaces the previous one.
CREATE TABLE activations (
  id          INTEGER PRIMARY KEY,
  user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  hash        BLOB NOT NULL UNIQUE,
  created_at  TEXT NOT NULL,
  expires_at  TEXT NOT NULL
) STRICT;

CREATE INDEX activations_user ON activations (user_id);

-- Sessions of people in the console. Only a SHA-256 hash of each token is
-- kept. A session ends 7 days after its last use, and at expires_at in any
-- case; an ended session is deleted.
CREATE TABLE sessions (
  id            INTEGER PRIMARY KEY,
  user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT NOT NULL,
  expires_at    TEXT NOT NULL
) STRICT;

CREATE INDEX sessions_user ON sessions (user_id);

-- Refused sign-ins of a person, counted to slow down password guessing.
CREATE INDEX logbook_sign_in ON logbook (user_id, at) WHERE operation = 'sign_in' AND outcome = 'refused';
