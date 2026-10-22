-- The AI declares what the pages of each version do: whether they run
-- scripts, the origins they load from or call, and those their forms send
-- to. Cahut makes the Content Security Policy of their files from it. The
-- versions published before have none. See
-- docs/adr/0019-window-versions-access-and-declaration.md.
ALTER TABLE window_versions ADD COLUMN declaration TEXT; -- JSON
