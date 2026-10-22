-- The Logbook of the server keeps the layout of the Logbook of a Cahut,
-- which now records the Room an entry concerns. The server has no Rooms:
-- the column stays empty.
ALTER TABLE logbook ADD COLUMN room_id INTEGER;
