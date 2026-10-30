-- A Window carries tags, which its people and their AIs give it so that
-- they find it among many: 10 at most, each a word or a few joined by
-- hyphens, in lower case. They belong to the Window, not to a version, and
-- go with it. See docs/adr/0055-tags-and-search-of-windows.md.
CREATE TABLE window_tags (
  window_id  INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  tag        TEXT NOT NULL,
  PRIMARY KEY (window_id, tag)
) STRICT, WITHOUT ROWID;

CREATE INDEX window_tags_tag ON window_tags (tag);
