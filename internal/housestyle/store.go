package housestyle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cahut.cloud/internal/store"
)

// settings is what a house style keeps besides its files, as the store
// keeps it. An empty field takes the value of the neutral house style.
type settings struct {
	Name        string   `json:"name,omitempty"`
	Colors      colors   `json:"colors,omitzero"`
	ChartColors []string `json:"chart_colors,omitempty"`
	Heading     *font    `json:"heading_font,omitempty"` // the body font when nil
	Body        *font    `json:"body_font,omitempty"`    // the font of the system when nil
	Guidance    string   `json:"guidance,omitempty"`
	// Inactive house styles are not given to AIs. See
	// docs/adr/0023-house-style-inactive-for-ais.md.
	Inactive bool `json:"inactive,omitempty"`
}

type colors struct {
	Text       string `json:"text,omitempty"`
	Background string `json:"background,omitempty"`
	Primary    string `json:"primary,omitempty"`
	Secondary  string `json:"secondary,omitempty"`
}

type font struct {
	Family   string     `json:"family,omitempty"`
	Fallback string     `json:"fallback,omitempty"`
	Files    []fontFile `json:"files,omitempty"`
}

type fontFile struct {
	Name      string `json:"name"`   // under fonts/
	Format    string `json:"format"` // woff2, woff, truetype or opentype, as @font-face names them
	Weight    int    `json:"weight"`
	MaxWeight int    `json:"max_weight,omitempty"`
	Style     string `json:"style"`
}

// The names of the files of a house style.
const (
	logoFile        = "logo"
	logoInverseFile = "logo-inverse"
	fontsDir        = "fonts/"
)

// file is a file of a house style. Its content is loaded only when needed.
type file struct {
	typ     string
	size    int64
	sha256  string
	content []byte
}

// filesSize returns the bytes of the files of v, or 0 when v is nil.
func filesSize(v *version) int64 {
	var n int64
	if v != nil {
		for _, f := range v.files {
			n += f.size
		}
	}
	return n
}

// version is a house style as the store keeps it: the current one, or the
// previous one.
type version struct {
	id       int64
	settings settings
	files    map[string]file // by name
	setAt    time.Time
	setBy    string // the handle of the person; empty from the command line
}

// neutral reports whether v is the neutral house style.
func (v *version) neutral() bool {
	b, _ := json.Marshal(v.settings)
	return string(b) == "{}" && len(v.files) == 0
}

// querier runs queries, in a transaction or not.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// load returns the current house style and the previous one, nil when there
// is none, with the contents of their files when contents is set.
func load(ctx context.Context, q querier, contents bool) (cur, prev *version, err error) {
	if cur, err = loadState(ctx, q, "current", contents); err != nil {
		return nil, nil, err
	}
	if prev, err = loadState(ctx, q, "previous", contents); err != nil {
		return nil, nil, err
	}
	return cur, prev, nil
}

func loadState(ctx context.Context, q querier, state string, contents bool) (*version, error) {
	v := &version{files: map[string]file{}}
	var raw, at string
	var by sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT h.id, h.settings, h.set_at, u.handle FROM house_styles h LEFT JOIN users u ON u.id = h.set_by
		WHERE h.state = ? ORDER BY h.id DESC LIMIT 1`, state).Scan(&v.id, &raw, &at, &by)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(raw), &v.settings); err != nil {
		return nil, fmt.Errorf("read the settings of house style %d: %w", v.id, err)
	}
	if v.setAt, err = store.ParseTime(at); err != nil {
		return nil, err
	}
	v.setBy = by.String
	content := "NULL"
	if contents {
		content = "content"
	}
	rows, err := q.QueryContext(ctx, `SELECT name, type, size, sha256, `+content+` FROM house_style_files WHERE house_style_id = ?`, v.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var f file
		if err := rows.Scan(&name, &f.typ, &f.size, &f.sha256, &f.content); err != nil {
			return nil, err
		}
		v.files[name] = f
	}
	return v, rows.Err()
}

// loadFile returns the file called name of the current house style, with
// its content, or false when it has none.
func loadFile(ctx context.Context, q querier, name string) (file, bool, error) {
	var f file
	err := q.QueryRowContext(ctx, `
		SELECT f.type, f.size, f.sha256, f.content FROM house_style_files f
		JOIN house_styles h ON h.id = f.house_style_id AND h.state = 'current'
		WHERE f.name = ? ORDER BY h.id DESC LIMIT 1`, name).Scan(&f.typ, &f.size, &f.sha256, &f.content)
	if errors.Is(err, sql.ErrNoRows) {
		return file{}, false, nil
	}
	return f, err == nil, err
}

// save makes v the current house style in tx: the current one becomes the
// previous one, or the neutral house style does when there is none, and the
// previous one is deleted with its files. It sets the id of v.
func save(ctx context.Context, tx *sql.Tx, v *version, userID int64, now time.Time) error {
	at := store.FormatTime(now)
	if _, err := tx.ExecContext(ctx, "DELETE FROM house_styles WHERE state = 'previous'"); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, "UPDATE house_styles SET state = 'previous' WHERE state = 'current'")
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		// The Cahut had the neutral house style: it is the one to go back to.
		if _, err := tx.ExecContext(ctx, "INSERT INTO house_styles (state, settings, set_at) VALUES ('previous', '{}', ?)", at); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(v.settings)
	if err != nil {
		return err
	}
	var by any
	if userID != 0 {
		by = userID
	}
	err = tx.QueryRowContext(ctx, "INSERT INTO house_styles (state, settings, set_by, set_at) VALUES ('current', ?, ?, ?) RETURNING id",
		string(raw), by, at).Scan(&v.id)
	if err != nil {
		return err
	}
	for name, f := range v.files {
		_, err := tx.ExecContext(ctx, "INSERT INTO house_style_files (house_style_id, name, type, size, sha256, content) VALUES (?, ?, ?, ?, ?, ?)",
			v.id, name, f.typ, f.size, f.sha256, f.content)
		if err != nil {
			return err
		}
	}
	return nil
}

// swap makes the previous house style the current one, and the current
// one the previous one.
func swap(ctx context.Context, tx *sql.Tx, cur, prev *version) error {
	if _, err := tx.ExecContext(ctx, "UPDATE house_styles SET state = 'previous' WHERE id = ?", cur.id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, "UPDATE house_styles SET state = 'current' WHERE id = ?", prev.id)
	return err
}
