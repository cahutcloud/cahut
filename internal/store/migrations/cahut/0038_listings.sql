-- The address of the listing of a Cahut in the directory of connectors of
-- Claude, which the administrators of its server set; NULL when it has
-- none. See docs/adr/0061-a-guide-to-the-first-ai.md.
ALTER TABLE settings ADD COLUMN claude_listing TEXT;
