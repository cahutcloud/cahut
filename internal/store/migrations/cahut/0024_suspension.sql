-- The suspension of the Cahut, which the administrators of its server
-- decide: since when, and why. Both are NULL while the Cahut is active. See
-- docs/adr/0048-suspending-a-cahut.md.
ALTER TABLE settings ADD COLUMN suspended_at TEXT;
ALTER TABLE settings ADD COLUMN suspension_reason TEXT CHECK (suspension_reason <> '');
