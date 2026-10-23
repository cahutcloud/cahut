-- Single-use links to the first start of a Cahut, which work while it has
-- nobody: the first person chooses their handle, their name and their
-- password there. Only a SHA-256 hash of each code is kept, and a Cahut has
-- one link at most. See docs/adr/0025-first-start-in-the-browser.md.
CREATE TABLE setup_links (
  cahut_id    INTEGER PRIMARY KEY REFERENCES cahuts (id) ON DELETE CASCADE,
  hash        BLOB NOT NULL UNIQUE,
  created_at  TEXT NOT NULL,
  expires_at  TEXT NOT NULL
) STRICT;
