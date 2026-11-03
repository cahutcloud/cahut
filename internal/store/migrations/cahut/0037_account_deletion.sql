-- The people of a Cahut delete their own account unless its administrators
-- turn it off: every Cahut allows it until they do. See
-- docs/adr/0059-people-delete-their-own-account.md.
ALTER TABLE settings ADD COLUMN account_deletion INTEGER NOT NULL DEFAULT 1 CHECK (account_deletion IN (0, 1));
