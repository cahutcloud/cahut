-- A Window keeps its current version and the one before it, so that a
-- rollback goes back to it, and it can go to the trash and come back.
-- See docs/adr/0019-window-versions-access-and-declaration.md.

-- The versions that a Window keeps: its current one, whose number is
-- windows.version, and its previous one, windows.previous. Their files live
-- on disk under the id of the Window, then the number of the version.
CREATE TABLE window_versions (
  window_id     INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  number        INTEGER NOT NULL,
  size          INTEGER NOT NULL, -- bytes of its files
  file_count    INTEGER NOT NULL,
  published_by  INTEGER NOT NULL REFERENCES users (id),
  published_at  TEXT NOT NULL,
  PRIMARY KEY (window_id, number)
) STRICT;

INSERT INTO window_versions (window_id, number, size, file_count, published_by, published_at)
  SELECT id, version, size, file_count, created_by, updated_at FROM windows;

ALTER TABLE windows ADD COLUMN previous INTEGER; -- the number of the previous version, if any
ALTER TABLE windows ADD COLUMN trashed_at TEXT;
