-- Each Window has an id that never changes: 128 bits drawn at random, in 26
-- characters of base32, in lowercase and in the alphabet of Crockford. It
-- gives the Window a permanent address, /-/w/<id>; the integer id stays
-- inside the Cahut, for the files, the passes and the frames. The Windows
-- that are already there draw theirs here, as Go writes 16 bytes in base32:
-- 25 characters of 5 bits, then one of 3 bits. Each term draws again for
-- each row. See docs/adr/0050-window-ids-permanent-addresses-and-names.md.
ALTER TABLE windows ADD COLUMN uid TEXT NOT NULL DEFAULT '';

UPDATE windows SET uid =
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('0123456789abcdefghjkmnpqrstvwxyz', (random() & 31) + 1, 1) ||
  substr('048cgmrw', (random() & 7) + 1, 1);

CREATE UNIQUE INDEX windows_uid ON windows (uid);
