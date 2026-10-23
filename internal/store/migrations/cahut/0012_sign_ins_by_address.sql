-- Refused sign-ins are counted at each address, as well as for each
-- handle from everywhere: a stranger who knows a handle can no longer lock
-- its person out by guessing wrong. The Logbook keeps the address of the
-- calls that sign in, as limits count it: an IPv4 address, or the /64 of
-- an IPv6 one. See docs/revue-securite-v0.md.
ALTER TABLE logbook ADD COLUMN address TEXT;

CREATE INDEX logbook_sign_in_address ON logbook (address, at) WHERE operation = 'sign_in' AND outcome = 'refused';
