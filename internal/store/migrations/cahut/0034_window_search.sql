-- People and AIs search their Windows by the words of their titles,
-- descriptions, tags and pages. See
-- docs/adr/0055-tags-and-search-of-windows.md.

-- The text of the HTML pages of each version, without their markup, up to
-- 100 KB, read when the version is published. NULL for the versions
-- published before, until Cahut reads it from their files.
ALTER TABLE window_versions ADD COLUMN text TEXT;

-- The index: a row for each Window, whose rowid is its id, with its title,
-- its description, its tags and the text of its current version. Words
-- match whatever their case and accents.
CREATE VIRTUAL TABLE window_search USING fts5(title, description, tags, text, tokenize = 'unicode61 remove_diacritics 2');

INSERT INTO window_search (rowid, title, description, tags, text)
  SELECT id, title, description, coalesce((SELECT group_concat(tag, ' ' ORDER BY tag) FROM window_tags WHERE window_id = windows.id), ''), ''
  FROM windows;

-- A Window that goes takes its row. A migration that makes windows again
-- makes this trigger again too.
CREATE TRIGGER windows_leave_search AFTER DELETE ON windows BEGIN
  DELETE FROM window_search WHERE rowid = old.id;
END;
