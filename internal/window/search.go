package window

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"cahut.cloud/internal/store"
)

// People and AIs search their Windows with list_windows, by the words of
// their titles, descriptions, tags and pages, in an FTS5 table of the
// database of each Cahut, window_search. See
// docs/adr/0055-tags-and-search-of-windows.md.

// MaxText is the most text that Cahut keeps of a version of a Window, in
// bytes.
const MaxText = 100 << 10

// MaxWords is the most words that a search takes.
const MaxWords = 10

// isPage reports whether the file at p is an HTML page, whose text the
// search reads.
func isPage(p string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".html", ".htm":
		return true
	}
	return false
}

// pagesOf returns the HTML pages among files.
func pagesOf(files []decoded) []decoded {
	var pages []decoded
	for _, f := range files {
		if isPage(f.path) {
			pages = append(pages, f)
		}
	}
	return pages
}

// textOf returns the text of the HTML pages pages: that of index.html
// first, then of the others in the order of their paths, without markup,
// scripts and styles, with the text that stands for images, its spaces
// folded, and cut at MaxText.
func textOf(pages []decoded) string {
	slices.SortFunc(pages, func(a, b decoded) int {
		switch {
		case a.path == b.path:
			return 0
		case a.path == "index.html":
			return -1
		case b.path == "index.html":
			return 1
		}
		return strings.Compare(a.path, b.path)
	})
	var b strings.Builder
	for _, p := range pages {
		if !pageText(&b, p.data) {
			break
		}
	}
	return strings.ToValidUTF8(strings.TrimSpace(b.String()), "�")
}

// pageText writes the words of the HTML page data to b, after those that b
// holds, and reports whether b has room for more.
func pageText(b *strings.Builder, data []byte) bool {
	z := html.NewTokenizer(bytes.NewReader(data))
	skip := 0 // depth in elements whose content is not text
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// The end of the page, or markup too broken to read further.
			return true
		case html.StartTagToken, html.SelfClosingTagToken:
			name, attrs := z.TagName()
			switch atom.Lookup(name) {
			case atom.Script, atom.Style, atom.Noscript:
				// The tokenizer reads their content as one text, up to
				// their end tag, even when they close themselves.
				skip++
			case atom.Template:
				if tt == html.StartTagToken {
					skip++
				}
			case atom.Img, atom.Area:
				for attrs {
					var key, value []byte
					key, value, attrs = z.TagAttr()
					if string(key) == "alt" && skip == 0 && !words(b, value) {
						return false
					}
				}
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			switch atom.Lookup(name) {
			case atom.Script, atom.Style, atom.Template, atom.Noscript:
				skip = max(skip-1, 0)
			}
		case html.TextToken:
			if skip == 0 && !words(b, z.Text()) {
				return false
			}
		}
	}
}

// words writes the words of s to b, after a space, up to MaxText bytes, and
// reports whether b has room for more.
func words(b *strings.Builder, s []byte) bool {
	for _, w := range strings.Fields(string(s)) {
		if b.Len()+1+len(w) > MaxText {
			return false
		}
		b.WriteByte(' ')
		b.WriteString(w)
	}
	return true
}

// reindex writes the row of the Window with the given id in the index of
// the search, from what its transaction holds: its title, its description,
// its tags and the text of its current version.
func reindex(ctx context.Context, tx *sql.Tx, id int64) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM window_search WHERE rowid = ?", id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO window_search (rowid, title, description, tags, text)
		SELECT w.id, w.title, w.description, coalesce((SELECT group_concat(tag, ' ' ORDER BY tag) FROM window_tags WHERE window_id = w.id), ''),
			coalesce(v.text, '')
		FROM windows w JOIN window_versions v ON v.window_id = w.id AND v.number = w.version
		WHERE w.id = ?`, id)
	return err
}

// matchOf returns the query of FTS5 that finds the Windows holding each
// word of s, at the start of a word: "q3"* "bret"*. A word is a run of
// letters and digits, as the index cuts them. It returns "" when s holds no
// word.
func matchOf(s string) string {
	ws := strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r) })
	if len(ws) > MaxWords {
		ws = ws[:MaxWords]
	}
	for i, w := range ws {
		ws[i] = `"` + w + `"*`
	}
	return strings.Join(ws, " ")
}

// The marks around the words found, in the excerpt that FTS5 cuts: an
// excerpt without them holds none of the words.
const (
	markStart = "\x02"
	markEnd   = "\x03"
)

// excerptOf returns the excerpt that FTS5 cut, without the marks around the
// words found, or "" when it holds none of them.
func excerptOf(s string) string {
	if !strings.Contains(s, markStart) {
		return ""
	}
	return strings.NewReplacer(markStart, "", markEnd, "").Replace(s)
}

// ReadTexts reads the text of the versions of Windows published before the
// search, from their files, and indexes it: a few versions in each
// transaction, until none is left or ctx is done. A version whose files
// cannot be read gets no text. It returns how many versions it read.
func (s *Service) ReadTexts(ctx context.Context, db *store.DB) (int, error) {
	read := 0
	for ctx.Err() == nil {
		type version struct {
			window  int64
			number  int
			current bool
		}
		rows, err := db.Read.QueryContext(ctx, `
			SELECT v.window_id, v.number, v.number = w.version
			FROM window_versions v JOIN windows w ON w.id = v.window_id
			WHERE v.text IS NULL LIMIT 20`)
		if err != nil {
			return read, err
		}
		var versions []version
		for rows.Next() {
			var v version
			if err := rows.Scan(&v.window, &v.number, &v.current); err != nil {
				rows.Close()
				return read, err
			}
			versions = append(versions, v)
		}
		rows.Close()
		if err := rows.Err(); err != nil || len(versions) == 0 {
			return read, err
		}
		// The files are read before the transaction, which holds the only
		// connection that writes.
		texts := make([]string, len(versions))
		for i, v := range versions {
			pages, err := readPages(s.Dir(v.window, v.number))
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return read, err
			}
			texts[i] = textOf(pages)
		}
		tx, err := db.Write.BeginTx(ctx, nil)
		if err != nil {
			return read, err
		}
		for i, v := range versions {
			// An update may have dropped the version since.
			res, err := tx.ExecContext(ctx, "UPDATE window_versions SET text = ? WHERE window_id = ? AND number = ? AND text IS NULL", texts[i], v.window, v.number)
			if err == nil && v.current {
				if n, _ := res.RowsAffected(); n > 0 {
					err = reindex(ctx, tx, v.window)
				}
			}
			if err != nil {
				_ = tx.Rollback() // the error that matters is err
				return read, err
			}
		}
		if err := tx.Commit(); err != nil {
			return read, err
		}
		read += len(versions)
	}
	return read, ctx.Err()
}

// readPages returns the HTML pages of the version of a Window whose files
// are in dir.
func readPages(dir string) ([]decoded, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	fsys := root.FS()
	var pages []decoded
	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() || !isPage(p) {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		pages = append(pages, decoded{p, data})
		return nil
	})
	return pages, err
}
