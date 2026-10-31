package window

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
)

// Declaration is what the AI declares that the pages of a version do. Cahut
// makes the Content Security Policy of their files from it: a Window that
// declares no call can make none. See
// docs/adr/0019-window-versions-access-and-declaration.md.
type Declaration struct {
	Scripts bool     `json:"scripts" jsonschema:"whether the pages run scripts, inline or in files"`
	Origins []string `json:"origins,omitempty" jsonschema:"every origin the pages load anything from or call, in the browser of their readers, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make"`
	Forms   []string `json:"forms,omitempty" jsonschema:"the origins that the forms of the pages send to, if any"`
	Data    *Data    `json:"data,omitempty" jsonschema:"a SQLite database of the Window on the server, for a poll, a form or a tracker, which its pages call at .data/ under its root"`
}

// Data is what the pages of a version keep and read in the database of
// their Window. The database belongs to the Window, not to the version:
// it keeps its tables and rows across versions and rollbacks. See
// docs/adr/0056-a-database-for-each-window.md.
type Data struct {
	Migrations []string            `json:"migrations,omitempty" jsonschema:"the SQL that makes the database, in order: create tables, indexes, views and triggers, add columns, insert rows. Each runs once: keep those that ran as they are, add new ones at the end"`
	Readable   map[string][]string `json:"readable,omitempty" jsonschema:"the tables, and their columns, that readers SELECT themselves at .data/sql; other columns read as NULL. List nothing private"`
	Queries    map[string]Query    `json:"queries,omitempty" jsonschema:"the statements that the pages call at .data/queries/<name>, named with lowercase letters, digits and _"`
	Closes     *time.Time          `json:"closes,omitempty" jsonschema:"when the queries that write stop"`
}

// Query is a statement of a declaration, which pages call by its name.
type Query struct {
	SQL  string `json:"sql" jsonschema:"one statement that reads, adds, changes or deletes rows; parameters such as :name come from the page, :reader from Cahut"`
	Who  string `json:"who" jsonschema:"anyone, people signed in to the Cahut, or members of its Room"`
	Kind string `json:"kind,omitempty" jsonschema:"reads, adds, changes or deletes, which Cahut sets"`
}

// Who may call a query of a declaration.
const (
	Anyone  = "anyone"
	People  = "people"
	Members = "members"
)

// Bounds of the data of a declaration.
const (
	MaxMigrations    = 100
	MaxMigrationSize = 100_000
	MaxQueries       = 50
	MaxQuerySize     = 10_000
	MaxReadable      = 50
)

var (
	validQueryName = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)
	validSQLName   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)
)

// check returns the data, or nil when it declares nothing, or an error
// wrapping operation.ErrInvalid. What its SQL does, the database of the
// Window checks.
func (d *Data) check() (*Data, error) {
	if d == nil || len(d.Migrations) == 0 && len(d.Readable) == 0 && len(d.Queries) == 0 {
		return nil, nil
	}
	if len(d.Migrations) > MaxMigrations {
		return nil, operation.Invalidf("declaration.data has %d migrations, more than the %d allowed", len(d.Migrations), MaxMigrations)
	}
	for i, m := range d.Migrations {
		if strings.TrimSpace(m) == "" {
			return nil, operation.Invalidf("declaration.data: migration %d is empty", i+1)
		}
		if len(m) > MaxMigrationSize {
			return nil, operation.Invalidf("declaration.data: migration %d is longer than %d bytes", i+1, MaxMigrationSize)
		}
	}
	if len(d.Readable) > MaxReadable {
		return nil, operation.Invalidf("declaration.data.readable lists %d tables, more than the %d allowed", len(d.Readable), MaxReadable)
	}
	for t, cols := range d.Readable {
		if !validSQLName.MatchString(t) {
			return nil, operation.Invalidf("declaration.data.readable: %q is not the name of a table: use letters, digits and underscores", t)
		}
		for _, c := range cols {
			if !validSQLName.MatchString(c) {
				return nil, operation.Invalidf("declaration.data.readable.%s: %q is not the name of a column: use letters, digits and underscores", t, c)
			}
		}
	}
	if len(d.Queries) > MaxQueries {
		return nil, operation.Invalidf("declaration.data has %d queries, more than the %d allowed", len(d.Queries), MaxQueries)
	}
	out := &Data{Migrations: d.Migrations, Readable: d.Readable, Closes: d.Closes, Queries: make(map[string]Query, len(d.Queries))}
	for name, q := range d.Queries {
		if !validQueryName.MatchString(name) {
			return nil, operation.Invalidf("declaration.data.queries: %q is not the name of a query: use 1 to 64 lowercase letters, digits and underscores", name)
		}
		if strings.TrimSpace(q.SQL) == "" {
			return nil, operation.Invalidf("declaration.data.queries.%s: sql is empty", name)
		}
		if len(q.SQL) > MaxQuerySize {
			return nil, operation.Invalidf("declaration.data.queries.%s: sql is longer than %d bytes", name, MaxQuerySize)
		}
		switch q.Who {
		case Anyone, People, Members:
		default:
			return nil, operation.Invalidf("declaration.data.queries.%s: who %q: use anyone, people or members", name, q.Who)
		}
		out.Queries[name] = Query{SQL: q.SQL, Who: q.Who} // the database says what it does
	}
	if d.Closes != nil {
		closes := d.Closes.UTC().Truncate(time.Second)
		out.Closes = &closes
	}
	return out, nil
}

// statements returns the SQL of each query, by name.
func (d *Data) statements() map[string]string {
	out := make(map[string]string, len(d.Queries))
	for name, q := range d.Queries {
		out[name] = q.SQL
	}
	return out
}

// writesForAnyone returns the names of the queries that change or delete
// rows for anyone, as they are, sorted.
func (d *Data) writesForAnyone() []string {
	if d == nil {
		return nil
	}
	var out []string
	for _, name := range slices.Sorted(maps.Keys(d.Queries)) {
		q := d.Queries[name]
		if q.Who == Anyone && (q.Kind == "changes" || q.Kind == "deletes") {
			out = append(out, name+" "+q.SQL)
		}
	}
	return out
}

// String sums the data up for the Logbook, without its SQL.
func (d *Data) String() string {
	s := fmt.Sprintf("data: %d migrations, %d queries, %d readable tables", len(d.Migrations), len(d.Queries), len(d.Readable))
	if d.Closes != nil {
		s += ", writes close " + d.Closes.Format(time.RFC3339)
	}
	return s
}

// MaxOrigins bounds each list of a declaration.
const MaxOrigins = 20

// validOrigin matches an origin of a declaration: https or wss, a host
// that may start with *. for its subdomains, and an optional port.
var validOrigin = regexp.MustCompile(`^(https|wss)://(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(:[0-9]{1,5})?$`)

// check returns the declaration with its origins in lowercase and without
// a final slash, or an error wrapping operation.ErrInvalid.
func (d *Declaration) check() (*Declaration, error) {
	if d == nil {
		return nil, operation.Invalidf("declaration is missing: say whether the pages run scripts, " +
			"and list the origins they load from or call, and those their forms send to")
	}
	out := &Declaration{Scripts: d.Scripts}
	var err error
	if out.Origins, err = origins("origins", d.Origins); err != nil {
		return nil, err
	}
	if out.Forms, err = origins("forms", d.Forms); err != nil {
		return nil, err
	}
	if out.Data, err = d.Data.check(); err != nil {
		return nil, err
	}
	if out.Data != nil && !out.Scripts {
		return nil, operation.Invalidf("declaration.data: the pages call their data with scripts: set scripts to true")
	}
	return out, nil
}

func origins(field string, list []string) ([]string, error) {
	if len(list) > MaxOrigins {
		return nil, operation.Invalidf("declaration.%s lists %d origins, more than the %d allowed", field, len(list), MaxOrigins)
	}
	var out []string
	seen := map[string]bool{}
	for _, o := range list {
		o = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(o)), "/")
		if !validOrigin.MatchString(o) {
			return nil, operation.Invalidf("declaration.%s: %q is not an origin: give https:// or wss://, a host and an optional port, "+
				"such as https://cdn.jsdelivr.net, without a path", field, o)
		}
		if !seen[o] {
			seen[o] = true
			out = append(out, o)
		}
	}
	return out, nil
}

// String sums the declaration up for the Logbook.
func (d *Declaration) String() string {
	if d == nil {
		return "not declared"
	}
	s := "no scripts"
	if d.Scripts {
		s = "scripts"
	}
	list := func(name string, origins []string) string {
		if len(origins) == 0 {
			return "; no " + name
		}
		return fmt.Sprintf("; %s %s", name, strings.Join(origins, " "))
	}
	s += list("origins", d.Origins) + list("forms", d.Forms)
	if d.Data != nil {
		s += "; " + d.Data.String()
	}
	return s
}

// encode returns the declaration as the store keeps it.
func (d *Declaration) encode() (string, error) {
	b, err := json.Marshal(d)
	return string(b), err
}

// decodeDeclaration reads a declaration as the store keeps it: nil when the
// version was published before declarations.
func decodeDeclaration(s *string) (*Declaration, error) {
	if s == nil {
		return nil, nil
	}
	d := &Declaration{}
	return d, json.Unmarshal([]byte(*s), d)
}

// Sandbox returns the flags of the sandbox of the pages that d declares:
// without allow-scripts when they run none.
func (d *Declaration) Sandbox() string {
	flags := "allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads"
	if d == nil || d.Scripts {
		flags = "allow-scripts " + flags
	}
	return flags
}

// Policy returns the Content Security Policy of the files of a Window that
// d declares, served at the origin self, such as https://cahut.example.com.
// A Window published before declarations keeps the sandbox alone. Pages may
// fetch the blob: and data: URLs they make: these hold what the page already
// has, and the browser resolves a blob: URL only for the opaque origin that
// made it. See docs/adr/0045-windows-fetch-the-blob-and-data-urls-they-make.md.
func (d *Declaration) Policy(self string) string {
	sandbox := "sandbox " + d.Sandbox()
	if d == nil {
		return sandbox
	}
	from := func(sources ...string) string {
		return strings.Join(append(append([]string{self}, sources...), d.Origins...), " ")
	}
	directives := []string{sandbox, "default-src 'none'"}
	if d.Scripts {
		directives = append(directives, "script-src "+from("'unsafe-inline'", "'unsafe-eval'"))
	}
	forms := "'none'"
	if len(d.Forms) > 0 {
		forms = strings.Join(d.Forms, " ")
	}
	directives = append(directives,
		"style-src "+from("'unsafe-inline'"),
		"img-src "+from("data:", "blob:"),
		"font-src "+from("data:"),
		"media-src "+from("data:", "blob:"),
		"connect-src "+from("data:", "blob:"),
		"frame-src "+from(),
		"worker-src "+self+" blob:",
		"form-action "+forms,
		"base-uri 'none'",
	)
	return strings.Join(directives, "; ")
}
