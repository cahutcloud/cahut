-- Each notice of the server that time brings, such as a disk almost full
-- or a certificate that cannot be renewed, goes out once a day at most:
-- notices_sent remembers when it last went. See
-- docs/adr/0031-e-mail-through-a-relay.md.
CREATE TABLE notices_sent (
  kind     TEXT NOT NULL, -- such as disk, certificate or update
  key      TEXT NOT NULL, -- what it is about, such as a domain or a version
  sent_at  TEXT NOT NULL,
  PRIMARY KEY (kind, key)
) STRICT;
