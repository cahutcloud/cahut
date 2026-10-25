-- The links to the legal pages of the Cahut, which the administrators of
-- its server set: its legal notice, terms, privacy policy and support, each
-- the address of a page that the host publishes elsewhere, and another
-- address in French where it differs. See
-- docs/adr/0042-the-legal-pages-of-a-cahut.md.
CREATE TABLE legal_links (
  page  TEXT NOT NULL CHECK (page IN ('legal_notice', 'terms', 'privacy', 'support')),
  lang  TEXT NOT NULL CHECK (lang IN ('', 'fr')), -- '' for every language without an address of its own
  url   TEXT NOT NULL CHECK (url LIKE 'https://%'),
  PRIMARY KEY (page, lang)
) STRICT;
