-- Service tokens: a program on the server, such as the site where people
-- sign up, manages the people of the Cahut through the admin API. The Cahut
-- keeps the hash of each token, never the token. See
-- docs/adr/0035-a-program-on-the-server-manages-people-with-a-service-token.md.
CREATE TABLE service_tokens (
  id            INTEGER PRIMARY KEY,
  name          TEXT NOT NULL, -- what uses it, such as the site where people sign up
  hash          BLOB NOT NULL UNIQUE,
  created_at    TEXT NOT NULL,
  last_used_at  TEXT,
  revoked_at    TEXT
) STRICT;
