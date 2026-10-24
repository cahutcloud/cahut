-- A person may have an e-mail address, where their links go: one, unique
-- in the Cahut whatever its case. The person and the administrators of the
-- Cahut read it, and no AI. See docs/adr/0031-e-mail-through-a-relay.md.
ALTER TABLE users ADD COLUMN email TEXT;

CREATE UNIQUE INDEX users_email ON users (email COLLATE NOCASE);

-- An address that a person gives for themself replaces theirs once they use
-- the link sent to it. A person has one change waiting at most; only a
-- SHA-256 hash of its code is kept.
CREATE TABLE email_changes (
  user_id     INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  email       TEXT NOT NULL,
  hash        BLOB NOT NULL UNIQUE,
  expires_at  TEXT NOT NULL
) STRICT;
