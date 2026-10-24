-- The server sends no e-mail: only a Cahut has a relay, and the
-- administrators of the server have no address and receive no notice. See
-- docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
DROP TABLE mail_relay;

DROP TABLE admin_email_changes;

DROP INDEX admins_email;

ALTER TABLE admins DROP COLUMN email;

DROP TABLE notices_sent;
