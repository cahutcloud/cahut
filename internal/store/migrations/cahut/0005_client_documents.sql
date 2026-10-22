-- The last valid copy of each Client ID Metadata Document that a person
-- connected, used when the document no longer answers, as when its host is
-- down. See docs/adr/0016-oauth-for-ai-clients.md.
CREATE TABLE client_documents (
  client_id   TEXT PRIMARY KEY,  -- the address of the document
  document    TEXT NOT NULL,     -- as it was fetched, in JSON
  fetched_at  TEXT NOT NULL
) STRICT;
