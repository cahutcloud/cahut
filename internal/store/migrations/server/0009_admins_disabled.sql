-- An administrator of the server can be disabled, then enabled again: a
-- disabled administrator cannot sign in, and their sessions and their link
-- end. The server keeps at least one administrator who is not disabled.
ALTER TABLE admins ADD COLUMN disabled_at TEXT;
