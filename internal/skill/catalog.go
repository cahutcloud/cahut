package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

// CatalogBytes is the budget of the lines of the catalog, in bytes of
// JSON, within the bound of the list of tools (ADR 57).
const CatalogBytes = 1200

// CatalogLine is how many characters of its description a skill keeps in
// the catalog.
const CatalogLine = 160

// catalogHeading opens the catalog, after the description of find_skills:
// the words of the trials that AIs followed.
const catalogHeading = "The skills of this Cahut today:"

// catalog is the catalog of a Cahut, as it was built.
type catalog struct {
	text     string   // what the description of find_skills carries
	revision int64    // the revision it follows, or an earlier one
	bytes    int      // the bytes of JSON that the text takes
	left     []string // the names of the skills that do not fit
}

// Catalog returns the lines that the description of find_skills carries:
// a line for each skill of the Cahut, its name then its description cut at
// 160 characters on a word, those for every Window first, then by name,
// within budget, and a last line for those that do not fit. It returns ""
// for an individual Cahut. Revision is the revision of the catalog.
// See docs/adr/0064-the-skills-of-a-cahut.md.
func Catalog(ctx context.Context, q Querier, budget int) (lines string, revision int64, err error) {
	c, err := buildCatalog(ctx, q, budget)
	return c.text, c.revision, err
}

// Revision returns the revision of the catalog of a Cahut.
func Revision(ctx context.Context, q Querier) (int64, error) {
	var n int64
	err := q.QueryRowContext(ctx, "SELECT revision FROM skill_catalog WHERE id = 1").Scan(&n)
	return n, err
}

func buildCatalog(ctx context.Context, q Querier, budget int) (catalog, error) {
	// The revision comes first: out of a transaction, the skills read after
	// it are as new as it is, or newer, and a catalog never stands for a
	// revision later than its own.
	revision, err := Revision(ctx, q)
	if err != nil {
		return catalog{}, err
	}
	c := catalog{revision: revision}
	individual, err := settings.IsIndividual(ctx, q)
	if err != nil || individual {
		return c, err
	}
	rows, err := q.QueryContext(ctx, `
		SELECT s.name, v.description FROM skills s JOIN skill_versions v ON v.id = s.current_version
		WHERE s.owner_id IS NULL ORDER BY s.every_window DESC, s.name`)
	if err != nil {
		return catalog{}, err
	}
	defer rows.Close()
	var names, lines []string
	for rows.Next() {
		var name, description string
		if err := rows.Scan(&name, &description); err != nil {
			return catalog{}, err
		}
		names = append(names, name)
		lines = append(lines, "- "+name+": "+cutDescription(description))
	}
	if err := rows.Err(); err != nil {
		return catalog{}, err
	}
	if len(lines) == 0 {
		return c, nil
	}
	// The most lines that fit, with the line that counts the others.
	for n := len(lines); n >= 0; n-- {
		text := strings.Join(append([]string{catalogHeading}, lines[:n]...), "\n")
		if more := len(lines) - n; more > 0 {
			text += "\n" + moreSkills(more)
		}
		if b := jsonBytes(text); b <= budget || n == 0 {
			c.text, c.bytes, c.left = text, b, names[n:]
			return c, nil
		}
	}
	return c, nil
}

// cutDescription returns a description on one line, cut at CatalogLine
// characters at the end of a word, or within one when the words are long.
func cutDescription(d string) string {
	d = strings.Join(strings.Fields(d), " ")
	if utf8.RuneCountInString(d) <= CatalogLine {
		return d
	}
	cut := string([]rune(d)[:CatalogLine+1])
	// A description without spaces, in Japanese or with a long address,
	// keeps its first characters, not what comes before its first space.
	if i := strings.LastIndexByte(cut, ' '); i > 0 && utf8.RuneCountInString(cut[:i]) >= CatalogLine/2 {
		cut = cut[:i]
	} else {
		cut = string([]rune(cut)[:CatalogLine])
	}
	return strings.TrimRight(cut, " ,;:") + "…"
}

// moreSkills is the last line of a catalog that leaves n skills out.
func moreSkills(n int) string {
	if n == 1 {
		return "and 1 more skill: find_skills lists it"
	}
	return fmt.Sprintf("and %d more skills: find_skills lists them", n)
}

// jsonBytes returns how many bytes s takes in a string of JSON, its quotes
// left out.
func jsonBytes(s string) int {
	b, _ := json.Marshal(s) // a string always encodes
	return len(b) - 2
}

// PreviewCatalogInput is the input of preview_skill_catalog.
type PreviewCatalogInput struct{}

// CatalogPreview is the catalog as AIs receive it.
type CatalogPreview struct {
	Catalog string   `json:"catalog" jsonschema:"what the description of find_skills carries after its own words; empty when the Cahut has no skills, or is individual"`
	Bytes   int      `json:"bytes" jsonschema:"the bytes of JSON that it takes in the list of tools"`
	Budget  int      `json:"budget" jsonschema:"the bytes that it may take"`
	Left    []string `json:"left" jsonschema:"the names of the skills that do not fit, which AIs find with find_skills"`
}

// PreviewSkillCatalog shows the administrators of a Cahut the catalog of
// its skills as AIs receive it. See docs/adr/0064-the-skills-of-a-cahut.md.
var PreviewSkillCatalog = &operation.Op[PreviewCatalogInput, CatalogPreview]{
	Name:  "preview_skill_catalog",
	Title: "Preview what AIs see of skills",
	Description: "Show the catalog of the skills of this Cahut as AIs receive it in the description of find_skills, " +
		"the bytes it takes, its budget, and the skills that do not fit. For the administrators of the Cahut.",
	Level:      operation.CahutAdmin,
	ReadOnly:   true,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, _ PreviewCatalogInput) (CatalogPreview, error) {
		cat, err := buildCatalog(ctx, c.Tx, CatalogBytes)
		if err != nil {
			return CatalogPreview{}, err
		}
		left := cat.left
		if left == nil {
			left = []string{}
		}
		return CatalogPreview{Catalog: cat.text, Bytes: cat.bytes, Budget: CatalogBytes, Left: left}, nil
	},
}
