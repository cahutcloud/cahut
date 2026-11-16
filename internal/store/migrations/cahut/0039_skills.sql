-- The skills of a Cahut and of its people, in the format of Agent Skills,
-- with two versions each, their files included. See
-- docs/adr/0064-the-skills-of-a-cahut.md.

CREATE TABLE skills (
  id               INTEGER PRIMARY KEY,
  scope            TEXT NOT NULL CHECK (scope IN ('cahut', 'personal')),
  owner_id         INTEGER REFERENCES users (id) ON DELETE CASCADE, -- the person of a personal skill; NULL for a skill of the Cahut
  name             TEXT NOT NULL,
  every_window     INTEGER NOT NULL DEFAULT 0 CHECK (every_window IN (0, 1)),
  current_version  INTEGER,
  previous_version INTEGER,
  created_at       TEXT NOT NULL,
  CHECK ((scope = 'cahut') = (owner_id IS NULL)),
  CHECK (every_window = 0 OR scope = 'cahut')
) STRICT;
CREATE UNIQUE INDEX skills_of_the_cahut ON skills (name) WHERE owner_id IS NULL;
CREATE UNIQUE INDEX skills_of_a_person ON skills (owner_id, name) WHERE owner_id IS NOT NULL;

CREATE TABLE skill_versions (
  id          INTEGER PRIMARY KEY,
  skill_id    INTEGER NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
  number      INTEGER NOT NULL,
  digest      TEXT NOT NULL,
  text        TEXT NOT NULL,    -- the SKILL.md as it came
  description TEXT NOT NULL,
  size        INTEGER NOT NULL, -- the bytes of the SKILL.md and of the files
  written_at  TEXT NOT NULL,
  written_by  INTEGER REFERENCES users (id) ON DELETE SET NULL, -- NULL once that person is deleted
  via         TEXT NOT NULL CHECK (via IN ('console', 'import')),
  UNIQUE (skill_id, number)
) STRICT;

CREATE TABLE skill_files (
  version_id INTEGER NOT NULL REFERENCES skill_versions (id) ON DELETE CASCADE,
  path       TEXT NOT NULL,
  type       TEXT NOT NULL,
  content    BLOB NOT NULL,
  PRIMARY KEY (version_id, path)
) STRICT;

-- The index: a row for each skill, whose rowid is its id, with its name,
-- and the description and the text of its current version. Words match
-- whatever their case and accents, as for Windows (0034).
CREATE VIRTUAL TABLE skill_search USING fts5(name, description, text, tokenize = 'unicode61 remove_diacritics 2');

-- A skill that goes takes its row, those of a person who is deleted
-- included.
CREATE TRIGGER skills_leave_search AFTER DELETE ON skills BEGIN
  DELETE FROM skill_search WHERE rowid = old.id;
END;

-- The revision of the catalog that AIs see, which every change to a skill
-- of the Cahut raises in its transaction.
CREATE TABLE skill_catalog (
  id       INTEGER PRIMARY KEY CHECK (id = 1),
  revision INTEGER NOT NULL
) STRICT;
INSERT INTO skill_catalog (id, revision) VALUES (1, 0);
