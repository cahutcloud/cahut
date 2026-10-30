-- The address of moderation of a Cahut receives each report about its
-- Windows, beside its administrators: an abuse mailbox of its host, since
-- the administrators of the server have no e-mail. They set it. See
-- docs/adr/0054-reporting-and-withdrawing-a-window.md.
ALTER TABLE settings ADD COLUMN moderation_email TEXT;
