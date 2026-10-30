package window

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"cahut.cloud/internal/operation"
)

// The tags of a Window: 10 at most, each 1 to 32 letters, digits and
// hyphens, in lower case. They belong to the Window, not to a version: a
// new version and a rollback keep them. See
// docs/adr/0055-tags-and-search-of-windows.md.
const (
	MaxTags   = 10
	MaxTagLen = 32
)

// The refusals of tags, which people meet in the console.
var (
	badTag = operation.NewRefusalOf[GivenTag]("tag_invalid", operation.ErrInvalid,
		"tag {tag}: use 1 to 32 letters, digits and hyphens, such as q3-review")
	tooManyTags = operation.NewRefusalOf[TagCount]("tags_too_many", operation.ErrInvalid,
		"{count} tags: give a Window {limit} at most")
)

// GivenTag fills the refusal of a tag that has not the form of one.
type GivenTag struct {
	Tag operation.Quoted `json:"tag" jsonschema:"the tag, as it was given"`
}

// TagCount fills the refusal of too many tags.
type TagCount struct {
	Count int `json:"count" jsonschema:"how many tags were given"`
	Limit int `json:"limit" jsonschema:"the most a Window carries"`
}

// CheckTags returns the tags given, each as CheckTag writes it, without the
// empty ones and the repeated ones, in order. It returns an error wrapping
// operation.ErrInvalid when one has not the form of a tag, or when there
// are more than MaxTags.
func CheckTags(given []string) ([]string, error) {
	tags := []string{}
	for _, g := range given {
		tag, err := CheckTag(g)
		if err != nil {
			return nil, err
		}
		if tag != "" && !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}
	if len(tags) > MaxTags {
		return nil, tooManyTags(TagCount{Count: len(tags), Limit: MaxTags})
	}
	slices.Sort(tags)
	return tags, nil
}

// CheckTag returns the tag s as Cahut keeps it, as a person may type it,
// such as "#Q3 Review": in lower case, without a leading #, its words
// joined by single hyphens, q3-review. It returns "" when s holds nothing,
// and an error wrapping operation.ErrInvalid when s holds anything but
// letters, digits, spaces, hyphens and underscores, or when the tag is
// longer than MaxTagLen.
func CheckTag(s string) (string, error) {
	var b strings.Builder
	hyphen := false
	for _, r := range strings.TrimPrefix(strings.TrimSpace(norm.NFC.String(s)), "#") {
		switch {
		case unicode.IsSpace(r) || r == '-' || r == '_':
			hyphen = b.Len() > 0
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r):
			if hyphen {
				b.WriteByte('-')
				hyphen = false
			}
			b.WriteString(strings.ToLower(string(r)))
		default:
			return "", badTag(GivenTag{Tag: operation.Quoted(s)})
		}
	}
	tag := b.String()
	if utf8.RuneCountInString(tag) > MaxTagLen {
		return "", badTag(GivenTag{Tag: operation.Quoted(s)})
	}
	return tag, nil
}

// setTags gives the Window with the given id the tags tags, in place of
// those it had.
func setTags(ctx context.Context, tx *sql.Tx, id int64, tags []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM window_tags WHERE window_id = ?", id); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, "INSERT INTO window_tags (window_id, tag) VALUES (?, ?)", id, tag); err != nil {
			return err
		}
	}
	return nil
}

// tagsDetail says the tags tags in an entry of the Logbook.
func tagsDetail(tags []string) string {
	if len(tags) == 0 {
		return "no tags"
	}
	return "tags " + strings.Join(tags, ", ")
}
