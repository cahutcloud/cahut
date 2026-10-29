package address

import (
	"crypto/rand"
	"encoding/base32"
)

// ids writes the id of a Window in base32, in lowercase and in the alphabet
// of Crockford, which has no i, l, o or u, without padding. See
// docs/adr/0050-window-ids-permanent-addresses-and-names.md.
var ids = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

// idBytes is the size of an id: 128 bits, enough for the ids that Cahuts
// draw on their own to be unique in the world.
const idBytes = 16

// IDLength is the length of an id, in characters.
const IDLength = 26

// NewID returns a new id for a Window, drawn at random.
func NewID() string {
	b := make([]byte, idBytes)
	rand.Read(b) // never fails: it crashes the program rather than return an error
	return ids.EncodeToString(b)
}

// ValidID reports whether s is an id, as NewID writes it: 26 characters
// whose last one carries 3 bits only.
func ValidID(s string) bool {
	if len(s) != IDLength {
		return false
	}
	b, err := ids.DecodeString(s)
	return err == nil && len(b) == idBytes && ids.EncodeToString(b) == s
}
