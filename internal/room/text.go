package room

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
)

// The title of a Room or a Window, in bytes, and its description, in
// characters: a sentence or two. See
// docs/adr/0050-window-ids-permanent-addresses-and-names.md.
const (
	MaxTitle       = 200
	MaxDescription = 300
)

// The refusals of a title or a description, which people meet in the
// console.
var (
	titleEmpty         = operation.NewRefusal("title_empty", operation.ErrInvalid, "title is empty: give one")
	titleTooLong       = operation.NewRefusal("title_too_long", operation.ErrInvalid, "title is longer than 200 bytes: shorten it")
	descriptionTooLong = operation.NewRefusalOf[DescriptionLength]("description_too_long", operation.ErrInvalid,
		"the description has {length} characters: use at most {limit}, a sentence or two")
	descriptionControl = operation.NewRefusal("description_control", operation.ErrInvalid,
		"the description holds a control character: write it in plain text")
)

// DescriptionLength fills the refusal of a description that is too long.
type DescriptionLength struct {
	Length int `json:"length" jsonschema:"how many characters it has"`
	Limit  int `json:"limit" jsonschema:"the most it may have"`
}

// CheckTitle returns the title s without the spaces around it, or an error
// wrapping operation.ErrInvalid when it is empty or too long.
func CheckTitle(s string) (string, error) {
	title := strings.TrimSpace(s)
	switch {
	case title == "":
		return "", titleEmpty
	case len(title) > MaxTitle:
		return "", titleTooLong
	}
	return title, nil
}

// CheckDescription returns the description s on one line, its spaces and
// line breaks folded into single spaces, or an error wrapping
// operation.ErrInvalid when it is too long or holds a control character.
// An empty description is none.
func CheckDescription(s string) (string, error) {
	d := strings.Join(strings.Fields(s), " ")
	if strings.ContainsFunc(d, unicode.IsControl) {
		return "", descriptionControl
	}
	if n := utf8.RuneCountInString(d); n > MaxDescription {
		return "", descriptionTooLong(DescriptionLength{Length: n, Limit: MaxDescription})
	}
	return d, nil
}
