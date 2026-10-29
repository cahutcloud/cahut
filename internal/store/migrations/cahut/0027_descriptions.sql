-- A Window and a Room say what they are in a description: a sentence or
-- two of plain text, on one line, empty until someone writes one. See
-- docs/adr/0050-window-ids-permanent-addresses-and-names.md.
ALTER TABLE windows ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE rooms ADD COLUMN description TEXT NOT NULL DEFAULT '';
