// Package window publishes Windows, the living pages that AIs make, and
// keeps their files on disk. A Window is at <domain>/<room>/<window>/, under
// the domain of its Cahut. See docs/adr/0009-rooms-and-windows-as-paths.md.
package window

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
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
	url func(room, name string) string
}

// NewService returns a service that keeps Window files under dir and gives
// Window addresses with url.
func NewService(dir string, url func(room, name string) string) *Service {
	return &Service{dir: dir, url: url}
}

// Register adds the Window operations to r.
func (s *Service) Register(r *operation.Registry) {
	operation.Register(r, &operation.Op[PublishInput, Window]{
		Name: "publish_window",
		Description: "Publish a new Window: a document, deck, dashboard or site made of HTML and its files, " +
			"in a Room of the person you act for, by default their personal Room. Returns the address where it opens.",
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
		Description: "List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first.",
		ReadOnly:    true,
		Run:         s.list,
	})
}

// Window describes a published Window.
type Window struct {
	Name      string    `json:"name" jsonschema:"the Window's name, last part of its address"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
	URL       string    `json:"url" jsonschema:"where the Window opens"`
	Room      string    `json:"room" jsonschema:"the name of its Room, the part of its address before its name"`
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
	Room  string `json:"room,omitempty" jsonschema:"name of the Room to publish in; the personal Room of the person you act for when empty"`
	Files []File `json:"files" jsonschema:"the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css"`
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

	// A person publishes in their personal Room, named after their handle,
	// or in a Room they are in.
	name := in.Room
	if name == "" {
		name = c.Actor.Handle
	}
	r, err := room.Require(ctx, c, name, room.Member)
	if err != nil {
		return Window{}, err
	}
	c.InRoom(r.ID)
	roomID := r.ID
	w := Window{Title: title, Kind: in.Kind, Room: r.Name, Version: 1, Size: size, FileCount: len(files),
		CreatedBy: c.Actor.Handle, CreatedAt: c.Now, UpdatedAt: c.Now}
	if w.Name, err = freeName(ctx, c.Tx, roomID, title); err != nil {
		return Window{}, err
	}
	w.URL = s.url(w.Room, w.Name)
	now := store.FormatTime(c.Now)
	var id int64
	err = c.Tx.QueryRowContext(ctx, `
		INSERT INTO windows (room_id, name, title, kind, version, size, file_count, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		roomID, w.Name, w.Title, w.Kind, w.Version, w.Size, w.FileCount, c.Actor.UserID, now, now).Scan(&id)
	if err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("%s %q, %d files, %d bytes", w.Kind, w.Title, w.FileCount, w.Size))

	c.OnFailure(func() { os.RemoveAll(filepath.Join(s.dir, strconv.FormatInt(id, 10))) })
	if err := writeFiles(s.Dir(id, w.Version), files); err != nil {
		return Window{}, err
	}
	return w, nil
}

// Dir returns the directory that holds the files of a version of the Window
// with the given id.
func (s *Service) Dir(id int64, version int) string {
	return filepath.Join(s.dir, strconv.FormatInt(id, 10), strconv.Itoa(version))
}

// Current returns the id and the version served to readers of the Window
// called name in the Room called room. Serving a page is not an operation:
// readers are not actors, and their visits are not logged.
func Current(ctx context.Context, db *store.DB, room, name string) (id int64, version int, err error) {
	err = db.Read.QueryRowContext(ctx, `
		SELECT w.id, w.version FROM windows w JOIN rooms r ON r.id = w.room_id
		WHERE r.name = ? AND w.name = ?`, room, name).Scan(&id, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, operation.NotFoundf("no Window %s/%s", room, name)
	}
	return id, version, err
}

// GetInput is the input of get_window.
type GetInput struct {
	Window         string `json:"window" jsonschema:"the Window's address, or its Room and its name as room/name"`
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
	room, name, ok := locate(in.Window)
	if !ok {
		c.Log(in.Window, "")
		return Detail{}, operation.Invalidf("window %q: give the Window's address, or its Room and its name as room/name", in.Window)
	}
	c.Log(room+"/"+name, "")
	w, err := s.query(ctx, c, "r.name = ? AND w.name = ?", room, name)
	if err != nil {
		return Detail{}, err
	}
	if len(w) == 0 {
		return Detail{}, operation.NotFoundf("no Window %s/%s in your Rooms", room, name)
	}
	c.InRoom(w[0].roomID)
	files, err := listFiles(s.Dir(w[0].id, w[0].Version), in.IncludeContent)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Window: w[0].Window, Files: files}, nil
}

// ListInput is the input of list_windows.
type ListInput struct {
	Room string `json:"room,omitempty" jsonschema:"only the Windows of this Room, by its name"`
}

// List is the output of list_windows.
type List struct {
	Windows []Window `json:"windows"`
}

func (s *Service) list(ctx context.Context, c *operation.Call, in ListInput) (List, error) {
	where, args := "1 = 1", []any{}
	if in.Room != "" {
		r, err := room.Require(ctx, c, in.Room, room.Member)
		if err != nil {
			return List{}, err
		}
		c.InRoom(r.ID)
		where, args = "w.room_id = ?", []any{r.ID}
	}
	found, err := s.query(ctx, c, where, args...)
	if err != nil {
		return List{}, err
	}
	c.Log(in.Room, fmt.Sprintf("%d Windows", len(found)))
	list := List{Windows: make([]Window, len(found))}
	for i, w := range found {
		list.Windows[i] = w.Window
	}
	return list, nil
}

// found is a Window that query found, with its id and that of its Room.
type found struct {
	Window
	id, roomID int64
}

// query returns the Windows matching where that the actor may see: all of
// them for the command line, those of their Rooms for anyone else, and of
// the Rooms it was granted for an AI.
// Administrators manage people and Rooms, not content: an administrator of
// the server, who belongs to no Room, sees none.
func (s *Service) query(ctx context.Context, c *operation.Call, where string, args ...any) ([]found, error) {
	q := `
		SELECT w.id, w.room_id, w.name, w.title, w.kind, r.name, w.version, w.size, w.file_count, u.handle, w.created_at, w.updated_at
		FROM windows w JOIN rooms r ON r.id = w.room_id JOIN users u ON u.id = w.created_by
		WHERE ` + where
	if !c.Actor.CommandLine() {
		q += " AND w.room_id IN (SELECT room_id FROM room_members WHERE user_id = ?)"
		args = append(args, c.Actor.UserID)
	}
	rows, err := c.Tx.QueryContext(ctx, q+" ORDER BY w.updated_at DESC, w.id DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	windows := []found{}
	for rows.Next() {
		var w found
		var created, updated string
		if err := rows.Scan(&w.id, &w.roomID, &w.Name, &w.Title, &w.Kind, &w.Room, &w.Version, &w.Size, &w.FileCount,
			&w.CreatedBy, &created, &updated); err != nil {
			return nil, err
		}
		if w.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if w.UpdatedAt, err = store.ParseTime(updated); err != nil {
			return nil, err
		}
		if !c.Actor.Reaches(w.roomID) {
			continue
		}
		w.URL = s.url(w.Room, w.Name)
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

// freeName returns a name for a new Window in the Room roomID: its title as
// a name, then -2, -3 and so on when the Room already has a Window of that
// name.
func freeName(ctx context.Context, tx *sql.Tx, roomID int64, title string) (string, error) {
	base := address.Slug(title, "window")
	for n := 1; n <= 1000; n++ {
		name := base
		if n > 1 {
			name += "-" + strconv.Itoa(n)
		}
		var taken bool
		err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM windows WHERE room_id = ? AND name = ?)", roomID, name).Scan(&taken)
		if err != nil {
			return "", err
		}
		if !taken {
			return name, nil
		}
	}
	return "", operation.Invalidf("the Room has too many Windows called %q: choose another title", title)
}

// locate reads the Room and the name of a Window from its address, such as
// https://cahut.example.com/marketing/q3-review/index.html, or from
// room/name. It reports false when s is neither.
func locate(s string) (room, name string, ok bool) {
	p := strings.TrimSpace(s)
	if strings.Contains(p, "://") {
		u, err := url.Parse(p)
		if err != nil {
			return "", "", false
		}
		p = u.Path
	}
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) < 2 || !address.ValidName(parts[0]) || !address.ValidName(parts[1]) {
		return "", "", false
	}
	return parts[0], parts[1], true
}
