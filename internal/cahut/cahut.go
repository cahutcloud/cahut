// Package cahut assembles a Cahut: its database, its registry of operations
// and the services behind them. A server holds one or more Cahuts, each at
// its own domain. See docs/adr/0008-several-cahuts-per-server.md.
package cahut

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
}

// Cahut is an open Cahut.
type Cahut struct {
	Domain   string
	DB       *store.DB
	Registry *operation.Registry
	Windows  *window.Service
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
	url := func(room, name string) string {
		if cfg.BaseURL == "" {
			return ""
		}
		return cfg.BaseURL + room + "/" + name + "/"
	}
	r := operation.NewRegistry(db)
	account.Register(r)
	w := window.NewService(filepath.Join(cfg.Dir, "windows"), url)
	w.Register(r)
	operation.Register(r, ListLogbook)
	return &Cahut{Domain: cfg.Domain, DB: db, Registry: r, Windows: w}, nil
}

// Close closes the database.
func (c *Cahut) Close() error {
	return c.DB.Close()
}

// ListLogbookInput is the input of list_logbook.
type ListLogbookInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many entries, 50 by default, at most 1000"`
}

// LogbookPage is the output of list_logbook.
type LogbookPage struct {
	Entries []logbook.Entry `json:"entries"`
}

// ListLogbook returns the latest entries of the Logbook, newest first. The
// server registers it too, for its own Logbook.
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
