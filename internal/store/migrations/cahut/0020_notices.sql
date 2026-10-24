-- Notices by e-mail. A person may turn off those that are neither about
-- their account nor about a Knock. The person who set the access of a
-- Window last hears that it ends. Each notice that time brings goes out
-- once: notices_sent remembers it. See docs/adr/0031-e-mail-through-a-relay.md.
ALTER TABLE users ADD COLUMN notices_off INTEGER NOT NULL DEFAULT 0 CHECK (notices_off IN (0, 1));
ALTER TABLE windows ADD COLUMN access_set_by INTEGER REFERENCES users (id);

CREATE TABLE notices_sent (
  kind     TEXT NOT NULL, -- such as access_ends or limit
  key      TEXT NOT NULL, -- what it is about, such as a Window and the end of its access
  sent_at  TEXT NOT NULL,
  PRIMARY KEY (kind, key)
) STRICT;
