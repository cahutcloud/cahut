-- Refused sign-ins of administrators are counted at each address, as well
-- as for each handle from everywhere, as in each Cahut. See
-- docs/revue-securite-v0.md.
ALTER TABLE logbook ADD COLUMN address TEXT;

CREATE INDEX logbook_sign_in_address ON logbook (address, at) WHERE operation = 'sign_in' AND outcome = 'refused';
