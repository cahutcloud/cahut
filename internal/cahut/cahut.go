// Package cahut assembles a Cahut: its database, its registry of operations
// and the services behind them. A server holds one or more Cahuts, each at
// its own domain. See docs/adr/0008-several-cahuts-per-server.md.
package cahut

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/housestyle"
	"cahut.cloud/internal/knock"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/oauth"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// Config says where a Cahut keeps its data and how it is reached.
type Config struct {
	// Dir holds the database and the files of the Windows.
	Dir string
	// Domain is the domain of the Cahut, such as cahut.example.com.
	Domain string
	// BaseURL is the address of the Cahut, such as https://cahut.example.com/,
	// from which the addresses of its Windows are made. It may be empty where
	// addresses are not needed, as on the command line.
	BaseURL string
	// ConsoleURL is the address of the console of the Cahut, such as
	// https://cahut.example.com/console/, where people consent to AI clients.
	// It may be empty, as BaseURL may.
	ConsoleURL string
}

// Cahut is an open Cahut.
type Cahut struct {
	Domain   string
	DB       *store.DB
	Registry *operation.Registry
	Windows  *window.Service
	OAuth    *oauth.Server
	// About describes the Cahut: get_cahut.
	About *operation.Op[GetCahutInput, About]
	// FrameKey signs the capabilities of the frames of Windows: those of
	// protected Windows and those of previews in the console. See
	// docs/adr/0019-window-versions-access-and-declaration.md and
	// docs/adr/0027-previews-of-windows-in-the-console.md.
	FrameKey []byte
}

// Open opens the Cahut in cfg.Dir, creating it if needed.
func Open(ctx context.Context, cfg Config) (*Cahut, error) {
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, err
	}
	db, err := store.Open(ctx, filepath.Join(cfg.Dir, "cahut.db"), store.CahutSchema)
	if err != nil {
		return nil, fmt.Errorf("open the database of %s: %w", cfg.Domain, err)
	}
	frameKey, err := secret(ctx, db, "frames")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read the secrets of %s: %w", cfg.Domain, err)
	}
	url := func(room, name string) string {
		if cfg.BaseURL == "" {
			return ""
		}
		return cfg.BaseURL + room + "/" + name + "/"
	}
	r := operation.NewRegistry(db)
	account.Register(r)
	room.Register(r)
	w := window.NewService(filepath.Join(cfg.Dir, "windows"), url, frameKey)
	w.Register(r)
	auth := oauth.New(db, r, oauth.Config{
		Issuer:    strings.TrimSuffix(cfg.BaseURL, "/"),
		Authorize: cfg.ConsoleURL + "oauth/authorize",
		Resource:  cfg.BaseURL + "mcp",
	})
	auth.Register(r)
	knock.Register(r, func(id int64) string {
		if cfg.ConsoleURL == "" {
			return ""
		}
		return cfg.ConsoleURL + "knocks/" + strconv.FormatInt(id, 10)
	})
	housestyle.Register(r, cfg.Domain)
	operation.Register(r, settings.Set)
	operation.Register(r, ListLogbook)
	about := &operation.Op[GetCahutInput, About]{
		Name:  "get_cahut",
		Title: "Describe the Cahut",
		Description: "Describe this Cahut: its domain, the address where its Rooms and Windows open, that of its MCP endpoint, " +
			"and its limits with what it holds: a call that would cross a limit is refused.",
		ReadOnly: true,
		Run: func(ctx context.Context, c *operation.Call, _ GetCahutInput) (About, error) {
			c.Log(cfg.Domain, "")
			a := About{Domain: cfg.Domain, URL: cfg.BaseURL}
			if cfg.BaseURL != "" {
				a.MCP = cfg.BaseURL + "mcp"
			}
			var err error
			a.Settings, err = settings.Read(ctx, c.Tx)
			return a, err
		},
	}
	operation.Register(r, about)
	return &Cahut{Domain: cfg.Domain, DB: db, Registry: r, Windows: w, OAuth: auth, About: about, FrameKey: frameKey}, nil
}

// secret returns the random secret of the Cahut called name, drawn on first
// use.
func secret(ctx context.Context, db *store.DB, name string) ([]byte, error) {
	_, err := db.Write.ExecContext(ctx, "INSERT INTO secrets (name, value) VALUES (?, ?) ON CONFLICT DO NOTHING", name, credential.Random(32))
	if err != nil {
		return nil, err
	}
	var value []byte
	err = db.Write.QueryRowContext(ctx, "SELECT value FROM secrets WHERE name = ?", name).Scan(&value)
	return value, err
}

// Close closes the database.
func (c *Cahut) Close() error {
	return c.DB.Close()
}

// GetCahutInput is the input of get_cahut.
type GetCahutInput struct{}

// About is the output of get_cahut.
type About struct {
	Domain string `json:"domain"`
	URL    string `json:"url,omitempty" jsonschema:"the address of the Cahut, under which its Rooms and Windows open"`
	MCP    string `json:"mcp,omitempty" jsonschema:"the address of its MCP endpoint, where AI clients connect"`
	settings.Settings
}

// ListLogbookInput is the input of list_logbook.
type ListLogbookInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many entries, 50 by default, at most 1000"`
}

// LogbookPage is the output of list_logbook.
type LogbookPage struct {
	Entries []logbook.Entry `json:"entries"`
}

// ListLogbook returns the latest entries of the Logbook, newest first: all
// of them for an administrator, and for anyone else what they did and what
// their AIs did. An AI granted some Rooms only reads what happened in them.
var ListLogbook = &operation.Op[ListLogbookInput, LogbookPage]{
	Name:  "list_logbook",
	Title: "Read the Logbook",
	Description: "Read the latest entries of the Logbook, newest first: all of them for an administrator of the Cahut, " +
		"and for anyone else what they did and what AIs did on their behalf.",
	ReadOnly: true,
	Run:      readLogbook,
}

// ListServerLogbook returns the latest entries of the server's Logbook,
// newest first.
var ListServerLogbook = &operation.Op[ListLogbookInput, LogbookPage]{
	Name:        "list_logbook",
	Title:       "Read the server's Logbook",
	Description: "Read the latest entries of the server's Logbook, newest first.",
	Level:       operation.ServerAdmin,
	ReadOnly:    true,
	Run:         readLogbook,
}

func readLogbook(ctx context.Context, c *operation.Call, in ListLogbookInput) (LogbookPage, error) {
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 50
	case limit < 0 || limit > 1000:
		return LogbookPage{}, operation.Invalidf("limit %d: use 1 to 1000", limit)
	}
	var entries []logbook.Entry
	var err error
	switch {
	case c.Actor.Limited() && c.Admin:
		entries, err = logbook.ListIn(ctx, c.Tx, c.Actor.Rooms, 0, limit)
	case c.Actor.Limited():
		entries, err = logbook.ListIn(ctx, c.Tx, c.Actor.Rooms, c.Actor.UserID, limit)
	case c.Admin:
		entries, err = logbook.List(ctx, c.Tx, limit)
	default:
		entries, err = logbook.ListOf(ctx, c.Tx, c.Actor.UserID, limit)
	}
	if err != nil {
		return LogbookPage{}, err
	}
	if entries == nil {
		entries = []logbook.Entry{}
	}
	return LogbookPage{Entries: entries}, nil
}
