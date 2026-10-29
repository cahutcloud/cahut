-- Share links open a Window to whoever holds them, without an account or a
-- Key, until they end: beside the access of the Window, several to a
-- Window, each with its end, an optional number of openings and a label.
-- Only a SHA-256 hash of the token of a link is kept. Their numbers are
-- never reused, since the Logbook names links by them. See
-- docs/adr/0052-share-links.md.
CREATE TABLE share_links (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  window_id     INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  hash          BLOB NOT NULL UNIQUE,
  label         TEXT NOT NULL DEFAULT '',
  created_by    INTEGER REFERENCES users (id), -- NULL from the command line, and once the person is deleted
  client        TEXT,                          -- the AI that made it for them, if any
  created_at    TEXT NOT NULL,
  expires_at    TEXT NOT NULL,
  max_openings  INTEGER CHECK (max_openings BETWEEN 1 AND 1000), -- NULL when it has no limit
  openings      INTEGER NOT NULL DEFAULT 0,
  opened_at     TEXT,                          -- the last opening
  revoked_at    TEXT
) STRICT;

CREATE INDEX share_links_window ON share_links (window_id);
CREATE INDEX share_links_by ON share_links (created_by, created_at);

-- A pass that a link gave opens its Window whatever its access, and ends
-- with the link.
ALTER TABLE window_passes ADD COLUMN link_id INTEGER REFERENCES share_links (id) ON DELETE CASCADE;
