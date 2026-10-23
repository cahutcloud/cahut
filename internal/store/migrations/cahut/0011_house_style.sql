-- Each Cahut has a house style: the name of its organization, its colors,
-- its fonts, its logo and guidance for AIs. It keeps the current one and
-- the one before, which rollback_house_style restores. A Cahut without
-- one has the neutral house style. See
-- docs/adr/0022-house-style-of-each-cahut.md.
CREATE TABLE house_styles (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  state     TEXT NOT NULL CHECK (state IN ('current', 'previous')),
  settings  TEXT NOT NULL, -- JSON: the name, the colors, the fonts and the guidance
  set_by    INTEGER REFERENCES users (id), -- NULL when set from the command line
  set_at    TEXT NOT NULL
) STRICT;

-- The files of a house style: its logos and the files of its fonts. They
-- are small, and follow the transaction of the call that sets them.
CREATE TABLE house_style_files (
  house_style_id  INTEGER NOT NULL REFERENCES house_styles (id) ON DELETE CASCADE,
  name            TEXT NOT NULL, -- logo, logo-inverse, or fonts/<file>
  type            TEXT NOT NULL, -- its media type
  size            INTEGER NOT NULL,
  sha256          TEXT NOT NULL, -- in hex, for its ETag
  content         BLOB NOT NULL,
  PRIMARY KEY (house_style_id, name)
) STRICT;
