-- The administrators of the server, apart from the people of the Cahuts.
-- They sign in at the console of any Cahut, and each of their sessions holds
-- at the domain where it opened. See docs/adr/0013-people-sessions-and-rights.md.

-- An administrator without a password has not activated their account yet.
CREATE TABLE admins (
  id             INTEGER PRIMARY KEY,
  handle         TEXT NOT NULL UNIQUE,
  name           TEXT NOT NULL,
  password_hash  TEXT, -- Argon2id, as a PHC string
  created_at     TEXT NOT NULL
) STRICT;

-- Single-use links where an administrator chooses their password. Only a
-- SHA-256 hash of each code is kept, and a new link replaces the previous one.
CREATE TABLE admin_activations (
  id          INTEGER PRIMARY KEY,
  admin_id    INTEGER NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
  hash        BLOB NOT NULL UNIQUE,
  created_at  TEXT NOT NULL,
  expires_at  TEXT NOT NULL
) STRICT;

CREATE INDEX admin_activations_admin ON admin_activations (admin_id);

-- Sessions of administrators, each good at one domain. Only a SHA-256 hash
-- of each token is kept. A session ends 7 days after its last use, and at
-- expires_at in any case; an ended session is deleted.
CREATE TABLE admin_sessions (
  id            INTEGER PRIMARY KEY,
  admin_id      INTEGER NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
  domain        TEXT NOT NULL,
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT NOT NULL,
  expires_at    TEXT NOT NULL
) STRICT;

CREATE INDEX admin_sessions_admin ON admin_sessions (admin_id);

-- Refused sign-ins of an administrator, counted to slow down password
-- guessing.
CREATE INDEX logbook_sign_in ON logbook (user_handle, at) WHERE operation = 'sign_in' AND outcome = 'refused';
