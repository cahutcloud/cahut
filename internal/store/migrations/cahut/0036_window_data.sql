-- The data of each Window lives in a database of its own, beside its
-- files. The database of the Cahut keeps its size, with its copies, as it
-- stood after the last migration or change of an AI or a person, which the
-- limit of the disk counts; and the limit of the size of each database. See
-- docs/adr/0056-a-database-for-each-window.md.
ALTER TABLE windows ADD COLUMN data_size INTEGER NOT NULL DEFAULT 0;
ALTER TABLE settings ADD COLUMN data_per_window INTEGER;
