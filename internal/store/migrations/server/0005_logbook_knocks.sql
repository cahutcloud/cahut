-- The Knock of a call, as in the Logbook of a Cahut, so that both Logbooks
-- keep the same columns. No AI acts on the server, so no call knocks there.
ALTER TABLE logbook ADD COLUMN knock_id INTEGER;
