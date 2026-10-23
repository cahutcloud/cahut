-- The updates that the Logbook of the server records, by the identifier
-- that the updater gives each of them, so that each is recorded once. See
-- docs/adr/0024-signed-releases-and-updates.md.
CREATE TABLE updates_logged (
  id  TEXT PRIMARY KEY
) STRICT, WITHOUT ROWID;
