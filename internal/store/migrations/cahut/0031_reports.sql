-- Anyone reports a Window to the moderators of its Cahut, its
-- administrators and those of its server, who decide: they leave it as it
-- is, or withdraw it. The name and the address of whoever reports are kept
-- until the decision, and their address on the network never. Reports go
-- with their Window when it is purged. Their numbers are never reused, since
-- the Logbook names reports by them. See
-- docs/adr/0054-reporting-and-withdrawing-a-window.md.
CREATE TABLE reports (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  window_id    INTEGER NOT NULL REFERENCES windows (id) ON DELETE CASCADE,
  category     TEXT NOT NULL CHECK (category IN ('phishing', 'malware', 'illegal', 'rights', 'other')),
  explanation  TEXT NOT NULL,
  name         TEXT,                -- of whoever reports, until the decision
  email        TEXT,                -- where the receipt and the decision go, until the decision
  created_at   TEXT NOT NULL,
  decision     TEXT CHECK (decision IN ('dismissed', 'withdrawn')), -- NULL while it waits
  decided_at   TEXT,
  decided_by   TEXT,                -- the handle of the moderator who decided
  by_server    INTEGER NOT NULL DEFAULT 0 CHECK (by_server IN (0, 1)), -- an administrator of the server decided
  note         TEXT NOT NULL DEFAULT '' -- for the moderators, when they leave the Window as it is
) STRICT;

CREATE INDEX reports_window ON reports (window_id);
CREATE INDEX reports_waiting ON reports (id) WHERE decision IS NULL;
