package address

import (
	"regexp"
	"testing"
)

// anID is what the migration of the ids writes too: 25 characters of 5
// bits, then one of 3 bits.
var anID = regexp.MustCompile(`^[0-9a-hjkmnp-tv-z]{25}[048cgmrw]$`)

func TestNewID(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		id := NewID()
		if !anID.MatchString(id) || !ValidID(id) {
			t.Fatalf("NewID() = %q, which is not an id", id)
		}
		if seen[id] {
			t.Fatalf("NewID() gave %q twice", id)
		}
		seen[id] = true
	}
}

func TestValidID(t *testing.T) {
	if !ValidID("3f9k2m7q8x4r1t6v0b5n8c2d4g") {
		t.Error("ValidID refuses the id of the ADR")
	}
	for _, s := range []string{
		"",
		"3f9k2m7q8x4r1t6v0b5n8c2d4",   // too short
		"3f9k2m7q8x4r1t6v0b5n8c2d4gg", // too long
		"3f9k2m7q8x4r1t6v0b5n8c2d4h",  // a last character of more than 3 bits
		"3F9K2M7Q8X4R1T6V0B5N8C2D4G",  // uppercase
		"3f9k2m7q8x4r1t6v0b5n8c2dig",  // i is not in the alphabet
		"3f9k2m7q8x4r1t6v0b5n8c2d-g",
	} {
		if ValidID(s) {
			t.Errorf("ValidID(%q) = true", s)
		}
	}
}
