-- The Logbook of the server keeps the columns of the Logbook of a Cahut,
-- which names the person who reads an entry about a Window for its author
-- alone, in place of the members of its Room. The server has no Rooms, and
-- leaves it empty. See docs/adr/0053-windows-for-their-author-alone.md.
ALTER TABLE logbook ADD COLUMN reader_id INTEGER;
