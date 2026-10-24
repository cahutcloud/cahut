-- An administrator of the server may have an e-mail address, where their
-- links and the notices of the server go, unique whatever its case. An
-- address they give for themself replaces theirs once they use the link
-- sent to it. See docs/adr/0031-e-mail-through-a-relay.md.
ALTER TABLE admins ADD COLUMN email TEXT;

CREATE UNIQUE INDEX admins_email ON admins (email COLLATE NOCASE);

CREATE TABLE admin_email_changes (
  admin_id    INTEGER PRIMARY KEY REFERENCES admins (id) ON DELETE CASCADE,
  email       TEXT NOT NULL,
  hash        BLOB NOT NULL UNIQUE,
  expires_at  TEXT NOT NULL
) STRICT;
