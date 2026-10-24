-- Invitations to read a Window, sent by e-mail: to a person of the Cahut,
-- named by handle, or to anyone else, by address. For a Window with a Key,
-- an invitation carries a personal link, which stands for the Key for its
-- recipient alone: it gives a pass, until it ends with the Key or the
-- access, or someone revokes it. Only a SHA-256 hash of its code is kept.
-- See docs/adr/0031-e-mail-through-a-relay.md.
CREATE TABLE invitations (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  window_id   INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  invited_by  INTEGER REFERENCES users (id), -- NULL from the command line
  client      TEXT,                          -- the AI that sent it for them, if any
  user_id     INTEGER REFERENCES users (id) ON DELETE CASCADE, -- a person of the Cahut
  email       TEXT,                          -- or someone else
  note        TEXT NOT NULL DEFAULT '',
  hash        BLOB UNIQUE,                   -- the code of its personal link, for a Window with a Key
  created_at  TEXT NOT NULL,
  expires_at  TEXT,                          -- of its personal link
  ended_at    TEXT,                          -- revoked, or ended with its Key or its access
  CHECK ((user_id IS NULL) != (email IS NULL))
) STRICT;

CREATE INDEX invitations_window ON invitations (window_id);
CREATE INDEX invitations_by ON invitations (invited_by, created_at);

-- A pass that a personal link gave ends with its invitation.
ALTER TABLE window_passes ADD COLUMN invitation_id INTEGER REFERENCES invitations (id) ON DELETE CASCADE;

-- Single-use codes, valid a minute, that the console hands to the port of
-- the Windows to give the pass of an invitation, as it does to sign readers
-- in. Only a SHA-256 hash is kept.
CREATE TABLE invitation_openings (
  hash           BLOB PRIMARY KEY,
  invitation_id  INTEGER NOT NULL REFERENCES invitations (id) ON DELETE CASCADE,
  expires_at     TEXT NOT NULL
) STRICT;
