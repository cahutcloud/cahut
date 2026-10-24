-- The SMTP relay of the Cahut, which its administrators set: its e-mail
-- goes through it, or else through the relay of the server. A single row,
-- absent while the Cahut has no relay of its own. It speaks TLS, and is at
-- a public address. The password is sealed under a key of the server, kept
-- outside the directory of every Cahut. See
-- docs/adr/0031-e-mail-through-a-relay.md.
CREATE TABLE mail_relay (
  id           INTEGER PRIMARY KEY CHECK (id = 1),
  host         TEXT NOT NULL,
  port         INTEGER NOT NULL CHECK (port BETWEEN 1 AND 65535),
  security     TEXT NOT NULL CHECK (security IN ('tls', 'starttls')),
  username     TEXT NOT NULL DEFAULT '', -- empty: the relay takes no authentication
  password     BLOB,                     -- sealed with AES-256-GCM; NULL without a user name
  sender       TEXT NOT NULL,            -- the address that messages come from
  sender_name  TEXT NOT NULL DEFAULT '', -- the name they come from; empty for the name of the Cahut
  set_at       TEXT NOT NULL
) STRICT;
