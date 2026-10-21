// Package home assembles a Cahut home: its database, its registry of
// operations and the services behind them.
package home

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// Config says where a home keeps its data and how it is reached.
type Config struct {
	// DataDir holds the database and the files of the Windows.
	DataDir string
	// WindowURL returns the address of a Window from its name. It may be
	// nil where addresses are not needed, as on the command line.
	WindowURL func(name string) string
}

// Home is an open home.
type Home struct {
	DB       *store.DB
	Registry *operation.Registry
	Windows  *window.Service
}

// Open opens the home in cfg.DataDir, creating it if needed.
func Open(ctx context.Context, cfg Config) (*Home, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return nil, err
	}
	db, err := store.Open(ctx, filepath.Join(cfg.DataDir, "cahut.db"))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	url := cfg.WindowURL
	if url == nil {
		url = func(string) string { return "" }
	}
	r := operation.NewRegistry(db)
	account.Register(r)
	w := window.NewService(filepath.Join(cfg.DataDir, "windows"), url)
	w.Register(r)
	operation.Register(r, ListLogbook)
	return &Home{DB: db, Registry: r, Windows: w}, nil
}

// Close closes the database.
func (h *Home) Close() error {
	return h.DB.Close()
}

// ListLogbookInput is the input of list_logbook.
type ListLogbookInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many entries, 50 by default, at most 1000"`
}

// LogbookPage is the output of list_logbook.
type LogbookPage struct {
	Entries []logbook.Entry `json:"entries"`
}

// ListLogbook returns the latest entries of the Logbook, newest first.
var ListLogbook = &operation.Op[ListLogbookInput, LogbookPage]{
	Name:        "list_logbook",
	Description: "Read the latest entries of the Logbook, newest first.",
	AdminOnly:   true,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in ListLogbookInput) (LogbookPage, error) {
		limit := in.Limit
		switch {
		case limit == 0:
			limit = 50
		case limit < 0 || limit > 1000:
			return LogbookPage{}, operation.Invalidf("limit %d: use 1 to 1000", limit)
		}
		entries, err := logbook.List(ctx, c.Tx, limit)
		if err != nil {
			return LogbookPage{}, err
		}
		if entries == nil {
			entries = []logbook.Entry{}
		}
		return LogbookPage{Entries: entries}, nil
	},
}
