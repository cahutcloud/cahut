package moderation

import (
	"context"
	"database/sql"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Without a report, a moderator opens what anyone opens: the public
// Windows of the Cahut, which they list, the latest made public first, so
// that they see what the Cahut serves to anyone without waiting for a
// report. See docs/adr/0054-reporting-and-withdrawing-a-window.md.

// ListPublicInput is the input of list_public_windows.
type ListPublicInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"how many Windows, 100 by default, at most 500"`
}

// PublicWindows is the output of list_public_windows.
type PublicWindows struct {
	Windows []PublicWindow `json:"windows" jsonschema:"the public Windows of the Cahut, out of the trash and not withdrawn, the latest made public first"`
}

// PublicWindow is a public Window, as its moderators list it.
type PublicWindow struct {
	Reported
	By       string    `json:"by,omitempty" jsonschema:"the handle of the person it was first published for; empty once they are deleted"`
	PublicAt time.Time `json:"public_at" jsonschema:"when its access was last set, or when it was published"`
}

func listPublic(ctx context.Context, c *operation.Call, in ListPublicInput) (PublicWindows, error) {
	c.Log("public Windows", "")
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 100
	case limit < 0 || limit > 500:
		return PublicWindows{}, operation.Invalidf("limit %d: use 1 to 500", limit)
	}
	rows, err := c.Tx.QueryContext(ctx, `
		SELECT w.uid, r.name, w.name, w.title, w.access, coalesce(u.handle, ''), coalesce(w.access_set_at, w.created_at) AS public_at
		FROM windows w JOIN rooms r ON r.id = w.room_id LEFT JOIN users u ON u.id = w.created_by
		WHERE w.access = 'public' AND w.trashed_at IS NULL AND w.withdrawn_at IS NULL
		ORDER BY public_at DESC, w.id DESC LIMIT ?`, limit)
	if err != nil {
		return PublicWindows{}, err
	}
	defer rows.Close()
	out := PublicWindows{Windows: []PublicWindow{}}
	for rows.Next() {
		var p PublicWindow
		var at string
		if err := rows.Scan(&p.ID, &p.Room, &p.Name, &p.Title, &p.Access, &p.By, &at); err != nil {
			return PublicWindows{}, err
		}
		if p.PublicAt, err = store.ParseTime(at); err != nil {
			return PublicWindows{}, err
		}
		out.Windows = append(out.Windows, p)
	}
	return out, rows.Err()
}

// querier runs a query, in a transaction or not.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Waiting returns how many reports about the Windows of the Cahut wait for
// its moderators, for the list of the Cahuts of the server.
func Waiting(ctx context.Context, q querier) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, "SELECT count(*) FROM reports WHERE decision IS NULL").Scan(&n)
	return n, err
}
