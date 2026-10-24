-- The SMTP relay of the server, which its administrators set: the e-mail
-- of the server, and that of every Cahut without a relay of its own, goes
-- through it. A single row, absent while the server has none. It may be on
-- this machine or its network, and without TLS on its loopback address
-- only. The password is sealed under the key in mail.key. See
-- docs/adr/0031-e-mail-through-a-relay.md.
CREATE TABLE mail_relay (
  id           INTEGER PRIMARY KEY CHECK (id = 1),
  host         TEXT NOT NULL,
  port         INTEGER NOT NULL CHECK (port BETWEEN 1 AND 65535),
  security     TEXT NOT NULL CHECK (security IN ('tls', 'starttls', 'none')),
  username     TEXT NOT NULL DEFAULT '', -- empty: the relay takes no authentication
  password     BLOB,                     -- sealed with AES-256-GCM; NULL without a user name
  sender       TEXT NOT NULL,            -- the address that messages come from
  sender_name  TEXT NOT NULL DEFAULT '', -- the name they come from; empty for Cahut
  set_at       TEXT NOT NULL
) STRICT;
