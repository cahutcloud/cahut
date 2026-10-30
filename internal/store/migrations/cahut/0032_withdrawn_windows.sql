-- The moderators of a Cahut, its administrators and those of its server,
-- withdraw a Window with a reason, until they reinstate it: it keeps
-- everything, and serves nothing to its readers meanwhile. A Window that
-- the administrators of the server withdrew is reinstated by them only. See
-- docs/adr/0054-reporting-and-withdrawing-a-window.md.
ALTER TABLE windows ADD COLUMN withdrawn_at TEXT;
ALTER TABLE windows ADD COLUMN withdrawn_reason TEXT;
ALTER TABLE windows ADD COLUMN withdrawn_by_server INTEGER NOT NULL DEFAULT 0 CHECK (withdrawn_by_server IN (0, 1));
