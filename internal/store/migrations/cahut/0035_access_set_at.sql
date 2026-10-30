-- When the access of a Window was last set, so that the moderators of a
-- Cahut list its public Windows, the latest made public first. A Window
-- whose access was set before counts from when it was published. See
-- docs/adr/0054-reporting-and-withdrawing-a-window.md.
ALTER TABLE windows ADD COLUMN access_set_at TEXT;
