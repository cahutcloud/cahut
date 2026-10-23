package housestyle

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"time"

	"cahut.cloud/internal/store"
)

// The kit of the house style: the classes of its sheet, the deck engine,
// the templates of each kind of Window and the guide that AIs read. They
// are under Apache-2.0, stated in their directory. See
// docs/adr/0022-house-style-of-each-cahut.md.
//
//go:embed kit/kit.css kit/deck.js kit/guide.md kit/templates
var kit embed.FS

var (
	kitCSS    = mustRead("kit/kit.css")
	deckJS    = asset("text/javascript; charset=utf-8", mustRead("kit/deck.js"))
	guide     = string(mustRead("kit/guide.md"))
	templates = template.Must(template.ParseFS(kit, "kit/templates/*.html"))
)

func mustRead(name string) []byte {
	b, err := kit.ReadFile(name)
	if err != nil {
		panic(err) // an embedded file: a programming error
	}
	return b
}

// Guide returns the guide to the variables and the classes of the style
// sheet, in Markdown, as get_house_style gives it to AIs.
func Guide() string {
	return guide
}

// DeckPath is the address of the deck engine, from the root of the domain.
const DeckPath = "/-/deck.js"

// DeckEngine returns the deck engine.
func DeckEngine() Asset {
	return deckJS
}

// Template is the template of a kind of Window, in the house style.
type Template struct {
	Kind    string `json:"kind" jsonschema:"the kind of Window it makes"`
	HTML    string `json:"html" jsonschema:"index.html of a Window of this kind: keep its structure and its classes, replace its sample content, and give it its own title"`
	Scripts bool   `json:"scripts" jsonschema:"whether it runs scripts, which the Window then declares"`
}

// Kinds that have a template: a site starts from the document's.
var kinds = map[string]string{"document": "document", "deck": "deck", "dashboard": "dashboard", "site": "document"}

// render returns the template of kind, filled with the name of the
// organization and the date of now.
func render(kind, name string, now time.Time) (*Template, error) {
	file, ok := kinds[kind]
	if !ok {
		return nil, nil
	}
	var b bytes.Buffer
	err := templates.ExecuteTemplate(&b, file+".html", struct{ Name, Date string }{name, now.Format("2 January 2006")})
	if err != nil {
		return nil, err
	}
	return &Template{Kind: kind, HTML: b.String(), Scripts: file == "deck"}, nil
}

// TemplatePage returns the template of kind in the house style of the
// Cahut at domain, for people to look at, or false when kind has none.
func TemplatePage(ctx context.Context, db *store.DB, domain, kind string, now time.Time) ([]byte, bool, error) {
	if _, ok := kinds[kind]; !ok || kind == "site" {
		return nil, false, nil
	}
	b, err := LoadBrand(ctx, db, domain)
	if err != nil {
		return nil, false, err
	}
	t, err := render(kind, b.Name, now)
	if err != nil {
		return nil, false, err
	}
	return []byte(t.HTML), true, nil
}
