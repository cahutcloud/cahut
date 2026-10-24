-- The Window that a Room shows at its address, in place of the page that
-- lists its Windows; none by default. Purging that Window takes the Room
-- back to its list. See docs/adr/0032-the-page-of-a-room.md.
ALTER TABLE rooms ADD COLUMN home_window_id INTEGER REFERENCES windows (id) ON DELETE SET NULL;
