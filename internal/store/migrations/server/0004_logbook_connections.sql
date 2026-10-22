-- The connection of an AI that acted, as in the Logbook of a Cahut, so that
-- both Logbooks keep the same structure. No AI acts on the server.
ALTER TABLE logbook ADD COLUMN connection_id INTEGER;
