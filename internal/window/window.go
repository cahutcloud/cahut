// Package window publishes Windows, the living pages that AIs make, and
// keeps their files on disk. Each Window has its own web origin, named after
// its DNS label. See docs/adr/0003-one-origin-per-window.md.
package window

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Kinds of Windows.
const (
	Document  = "document"
	Deck      = "deck"
	Dashboard = "dashboard"
	Site      = "site"
)

// Service runs the Window operations and knows where their files live.
type Service struct {
	dir string
	url func(name string) string

	// Published, when set, is called with the name of each Window once it
	// is published. The server uses it to request the Window's certificate.
	Published func(name string)
}

// NewService returns a service that keeps Window files under dir and gives
// Window addresses with url.
func NewService(dir string, url func(name string) string) *Service {
	return &Service{dir: dir, url: url}
}

// Register adds the Window operations to r.
func (s *Service) Register(r *operation.Registry) {
	operation.Register(r, &operation.Op[PublishInput, Window]{
		Name: "publish_window",
		Description: "Publish a new Window: a document, deck, dashboard or site made of HTML and its files, " +
			"in the personal Room of the person you act for. Returns the address where it opens.",
		Run: s.publish,
	})
	operation.Register(r, &operation.Op[GetInput, Detail]{
		Name:        "get_window",
		Description: "Describe a Window and list its files, optionally with their contents.",
		ReadOnly:    true,
		Run:         s.get,
	})
	operation.Register(r, &operation.Op[ListInput, List]{
		Name:        "list_windows",
		Description: "List the Windows in the Rooms of the person you act for, most recently updated first.",
		ReadOnly:    true,
		Run:         s.list,
	})
}

// Window describes a published Window.
type Window struct {
	Name      string    `json:"name" jsonschema:"the Window's label, first part of its address"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
	URL       string    `json:"url" jsonschema:"where the Window opens"`
	Room      string    `json:"room"`
	Version   int       `json:"version"`
	Size      int64     `json:"size" jsonschema:"total size of the files, in bytes"`
	FileCount int       `json:"file_count"`
	CreatedBy string    `json:"created_by" jsonschema:"handle of the person it was published for"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublishInput is the input of publish_window.
type PublishInput struct {
	Title string `json:"title" jsonschema:"title of the Window, also used in its address"`
	Kind  string `json:"kind" jsonschema:"one of document, deck, dashboard or site"`
	Files []File `json:"files" jsonschema:"the files of the Window; index.html at the root is the page that opens"`
}

// File is a file of a publication.
type File struct {
	Path     string `json:"path" jsonschema:"relative path with / as separator, such as index.html or img/logo.svg"`
	Content  string `json:"content"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 (the default) or base64, for binary files"`
}

func (s *Service) publish(ctx context.Context, c *operation.Call, in PublishInput) (Window, error) {
	title := strings.TrimSpace(in.Title)
	c.Log(title, "")
	if title == "" {
		return Window{}, operation.Invalidf("title is empty")
	}
	if len(title) > 200 {
		return Window{}, operation.Invalidf("title is longer than 200 bytes")
	}
	switch in.Kind {
	case Document, Deck, Dashboard, Site:
	default:
		return Window{}, operation.Invalidf("kind %q: use document, deck, dashboard or site", in.Kind)
	}
	files, size, err := decode(in.Files)
	if err != nil {
		return Window{}, err
	}
	if c.Actor.UserID == 0 {
		return Window{}, operation.Invalidf("publish_window acts for a person, and the administrator is not one")
	}

	var roomID int64
	var room string
	err = c.Tx.QueryRowContext(ctx, `
		SELECT r.id, r.name FROM rooms r JOIN room_members m ON m.room_id = r.id
		WHERE r.kind = 'personal' AND m.user_id = ?`, c.Actor.UserID).Scan(&roomID, &room)
	if err != nil {
		return Window{}, fmt.Errorf("find personal Room: %w", err)
	}

	w := Window{Title: title, Kind: in.Kind, Room: room, Version: 1, Size: size, FileCount: len(files),
		CreatedBy: c.Actor.Handle, CreatedAt: c.Now, UpdatedAt: c.Now}
	if w.Name, err = freeName(ctx, c.Tx, title); err != nil {
		return Window{}, err
	}
	w.URL = s.url(w.Name)
	now := store.FormatTime(c.Now)
	_, err = c.Tx.ExecContext(ctx, `
		INSERT INTO windows (name, room_id, title, kind, version, size, file_count, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		w.Name, roomID, w.Title, w.Kind, w.Version, w.Size, w.FileCount, c.Actor.UserID, now, now)
	if err != nil {
		return Window{}, err
	}
	c.Log(w.Name, fmt.Sprintf("%s %q, %d files, %d bytes", w.Kind, w.Title, w.FileCount, w.Size))

	c.OnFailure(func() { os.RemoveAll(filepath.Join(s.dir, w.Name)) })
	if err := writeFiles(s.Dir(w.Name, w.Version), files); err != nil {
		return Window{}, err
	}
	if s.Published != nil {
		c.AfterCommit(func() { s.Published(w.Name) })
	}
	return w, nil
}

// Dir returns the directory that holds the files of a version of a Window.
func (s *Service) Dir(name string, version int) string {
	return filepath.Join(s.dir, name, strconv.Itoa(version))
}

// CurrentVersion returns the version of the Window called name that is
// served to readers. Serving a page is not an operation: readers are not
// actors, and their visits are not logged.
func CurrentVersion(ctx context.Context, db *store.DB, name string) (int, error) {
	var version int
	err := db.Read.QueryRowContext(ctx, "SELECT version FROM windows WHERE name = ?", name).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, operation.NotFoundf("no Window %q", name)
	}
	return version, err
}

// GetInput is the input of get_window.
type GetInput struct {
	Window         string `json:"window" jsonschema:"the Window's name or address"`
	IncludeContent bool   `json:"include_content,omitempty" jsonschema:"also return the content of each file"`
}

// Detail is the output of get_window.
type Detail struct {
	Window
	Files []FileInfo `json:"files"`
}

// FileInfo describes a file of a Window.
type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Type     string `json:"type" jsonschema:"media type"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 or base64"`
}

func (s *Service) get(ctx context.Context, c *operation.Call, in GetInput) (Detail, error) {
	name := nameFrom(in.Window)
	c.Log(name, "")
	w, err := s.query(ctx, c, "w.name = ?", name)
	if err != nil {
		return Detail{}, err
	}
	if len(w) == 0 {
		return Detail{}, operation.NotFoundf("no Window %q in your Rooms", in.Window)
	}
	files, err := listFiles(s.Dir(name, w[0].Version), in.IncludeContent)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Window: w[0], Files: files}, nil
}

// ListInput is the input of list_windows.
type ListInput struct{}

// List is the output of list_windows.
type List struct {
	Windows []Window `json:"windows"`
}

func (s *Service) list(ctx context.Context, c *operation.Call, _ ListInput) (List, error) {
	w, err := s.query(ctx, c, "1 = 1")
	if err != nil {
		return List{}, err
	}
	c.Log("", fmt.Sprintf("%d Windows", len(w)))
	return List{Windows: w}, nil
}

// query returns the Windows matching where that the actor may see: all of
// them for the administrator, those of their Rooms for anyone else.
func (s *Service) query(ctx context.Context, c *operation.Call, where string, args ...any) ([]Window, error) {
	q := `
		SELECT w.name, w.title, w.kind, r.name, w.version, w.size, w.file_count, u.handle, w.created_at, w.updated_at
		FROM windows w JOIN rooms r ON r.id = w.room_id JOIN users u ON u.id = w.created_by
		WHERE ` + where
	if c.Actor.Kind != operation.Admin {
		q += " AND w.room_id IN (SELECT room_id FROM room_members WHERE user_id = ?)"
		args = append(args, c.Actor.UserID)
	}
	rows, err := c.Tx.QueryContext(ctx, q+" ORDER BY w.updated_at DESC, w.id DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	windows := []Window{}
	for rows.Next() {
		var w Window
		var created, updated string
		if err := rows.Scan(&w.Name, &w.Title, &w.Kind, &w.Room, &w.Version, &w.Size, &w.FileCount,
			&w.CreatedBy, &created, &updated); err != nil {
			return nil, err
		}
		if w.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if w.UpdatedAt, err = store.ParseTime(updated); err != nil {
			return nil, err
		}
		w.URL = s.url(w.Name)
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

// freeName returns a name for a new Window: its title as a DNS label, then
// eight random characters. The random part keeps names unique and hard to
// guess while Windows have no Doors yet.
func freeName(ctx context.Context, tx *sql.Tx, title string) (string, error) {
	for range 3 {
		name := slug(title) + "-" + strings.ToLower(rand.Text()[:8])
		var taken bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM windows WHERE name = ?)", name).Scan(&taken); err != nil {
			return "", err
		}
		if !taken {
			return name, nil
		}
	}
	return "", errors.New("no free Window name after 3 tries")
}

// folds turns the accented letters of Latin scripts into ASCII for slugs.
var folds = map[rune]string{
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae",
	'ç': "c", 'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i", 'ñ': "n",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o", 'œ': "oe",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u", 'ý': "y", 'ÿ': "y", 'ß': "ss",
}

// slug turns a title into the start of a DNS label: at most 40 lowercase
// ASCII letters, digits and hyphens.
func slug(title string) string {
	var b strings.Builder
	hyphen := false
	for _, r := range strings.ToLower(title) {
		s, ok := folds[r]
		switch {
		case ok:
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			s = string(r)
		default:
			if b.Len() > 0 && !hyphen {
				b.WriteByte('-')
				hyphen = true
			}
			continue
		}
		if b.Len()+len(s) > 40 {
			break
		}
		b.WriteString(s)
		hyphen = false
	}
	if s := strings.Trim(b.String(), "-"); s != "" {
		return s
	}
	return "window"
}

var validName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// nameFrom accepts a Window's name or any address of it, and returns its
// name, or the input unchanged when it is neither.
func nameFrom(s string) string {
	s = strings.TrimSpace(s)
	host := s
	if strings.Contains(s, "://") {
		if u, err := url.Parse(s); err == nil {
			host = u.Hostname()
		}
	}
	label, _, _ := strings.Cut(host, ".")
	if validName.MatchString(label) {
		return label
	}
	return s
}
