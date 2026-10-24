-- A Cahut is collaborative, as every Cahut was until now, or individual:
-- each person works alone there and does not see the others, and limits
-- can bound each person. See docs/adr/0030-kind-and-limits-of-each-cahut.md.
ALTER TABLE settings ADD COLUMN kind TEXT NOT NULL DEFAULT 'collaborative' CHECK (kind IN ('collaborative', 'individual'));
ALTER TABLE settings ADD COLUMN rooms_per_person INTEGER CHECK (rooms_per_person >= 0);     -- Rooms beside the personal one
ALTER TABLE settings ADD COLUMN windows_per_person INTEGER CHECK (windows_per_person >= 0); -- Windows out of the trash, over the Rooms of the person
ALTER TABLE settings ADD COLUMN disk_per_person INTEGER CHECK (disk_per_person >= 0);       -- bytes of the versions of the Windows of the person
