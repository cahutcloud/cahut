-- The frame of a reported Window that an administrator of the server opens
-- names their session by its id, and holds while that session lasts, at
-- any domain of the server. A session that ends must not hand its id to the
-- next one, which would open the old frames again: admin_sessions is made
-- again with AUTOINCREMENT, as sessions were in the 0013 of each Cahut. See
-- docs/revue-securite-v0.md and docs/adr/0058-a-console-for-the-server.md.

CREATE TABLE admin_sessions_again (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  admin_id      INTEGER NOT NULL REFERENCES admins (id) ON DELETE CASCADE,
  domain        TEXT NOT NULL,
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT NOT NULL,
  expires_at    TEXT NOT NULL
) STRICT;

INSERT INTO admin_sessions_again (id, admin_id, domain, hash, created_at, last_used_at, expires_at)
  SELECT id, admin_id, domain, hash, created_at, last_used_at, expires_at FROM admin_sessions;

DROP TABLE admin_sessions;
ALTER TABLE admin_sessions_again RENAME TO admin_sessions;

CREATE INDEX admin_sessions_admin ON admin_sessions (admin_id);
