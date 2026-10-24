// Package settings keeps the settings of a Cahut, which the administrators
// of its server set: the limits on what it holds. Every operation that could
// cross a limit checks it here, in its own transaction. See
// docs/adr/0030-kind-and-limits-of-each-cahut.md.
package settings

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Limits bound what a Cahut holds. A limit that is nil counts nothing.
type Limits struct {
	People         *int64 `json:"people,omitempty" jsonschema:"how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count"`
	Admins         *int64 `json:"admins,omitempty" jsonschema:"how many of them may be administrators of the Cahut"`
	Rooms          *int64 `json:"rooms,omitempty" jsonschema:"how many team Rooms the Cahut may have; personal Rooms come with their person and do not count"`
	Windows        *int64 `json:"windows,omitempty" jsonschema:"how many Windows out of the trash the Cahut may hold"`
	WindowsPerRoom *int64 `json:"windows_per_room,omitempty" jsonschema:"how many Windows out of the trash each Room may hold, personal Rooms included"`
	Disk           *int64 `json:"disk,omitempty" jsonschema:"how many bytes the Cahut may keep: every version of every Window, those in the trash included, and the files of the house style"`
}

// Usage is what a Cahut holds, as its limits count it.
type Usage struct {
	People  int64 `json:"people" jsonschema:"people who are not disabled, administrators included"`
	Admins  int64 `json:"admins" jsonschema:"administrators who are not disabled"`
	Rooms   int64 `json:"rooms" jsonschema:"team Rooms"`
	Windows int64 `json:"windows" jsonschema:"Windows out of the trash"`
	Disk    int64 `json:"disk" jsonschema:"bytes of every version of every Window, those in the trash included, and of the files of the house style"`
}

// Settings are the settings of a Cahut, with what it holds.
type Settings struct {
	Limits Limits `json:"limits" jsonschema:"the limits that the administrators of the server set; a limit that is absent counts nothing"`
	Usage  Usage  `json:"usage" jsonschema:"what the Cahut holds, as its limits count it"`
}

// Load returns the limits of the Cahut, in the transaction tx.
func Load(ctx context.Context, tx *sql.Tx) (Limits, error) {
	var l Limits
	var people, admins, rooms, windows, perRoom, disk sql.NullInt64
	err := tx.QueryRowContext(ctx, "SELECT people, admins, rooms, windows, windows_per_room, disk FROM settings WHERE id = 1").
		Scan(&people, &admins, &rooms, &windows, &perRoom, &disk)
	if err != nil {
		return Limits{}, err
	}
	l.People, l.Admins, l.Rooms = ptr(people), ptr(admins), ptr(rooms)
	l.Windows, l.WindowsPerRoom, l.Disk = ptr(windows), ptr(perRoom), ptr(disk)
	return l, nil
}

func ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

// Use returns what the Cahut holds, in the transaction tx.
func Use(ctx context.Context, tx *sql.Tx) (Usage, error) {
	var u Usage
	err := tx.QueryRowContext(ctx, `
		SELECT
		  (SELECT count(*) FROM users WHERE disabled_at IS NULL),
		  (SELECT count(*) FROM users WHERE disabled_at IS NULL AND role = 'admin'),
		  (SELECT count(*) FROM rooms WHERE kind = 'team'),
		  (SELECT count(*) FROM windows WHERE trashed_at IS NULL)`).
		Scan(&u.People, &u.Admins, &u.Rooms, &u.Windows)
	if err != nil {
		return Usage{}, err
	}
	u.Disk, err = disk(ctx, tx)
	return u, err
}

// Read returns the settings of the Cahut, with what it holds, in the
// transaction tx.
func Read(ctx context.Context, tx *sql.Tx) (Settings, error) {
	l, err := Load(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	u, err := Use(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	return Settings{Limits: l, Usage: u}, nil
}

// disk returns the bytes that the Cahut keeps, as the limit of its disk
// counts them: the sizes that the database records for each version of a
// Window and each file of the house style.
func disk(ctx context.Context, tx *sql.Tx) (int64, error) {
	var n int64
	err := tx.QueryRowContext(ctx, `
		SELECT (SELECT coalesce(sum(size), 0) FROM window_versions) + (SELECT coalesce(sum(size), 0) FROM house_style_files)`).Scan(&n)
	return n, err
}

// SetInput is the input of set_cahut_settings.
type SetInput struct {
	Limits Limits `json:"limits" jsonschema:"every limit of the Cahut: a limit left out is lifted"`
}

// Set sets the limits of the Cahut, for the administrators of the server.
var Set = &operation.Op[SetInput, Settings]{
	Name:  "set_cahut_settings",
	Title: "Set the limits of the Cahut",
	Description: "Set the limits of this Cahut: how many people, administrators, team Rooms and Windows it may have, " +
		"how many Windows each Room may hold, and how many bytes it may keep. Give every limit: a limit left out is lifted. " +
		"Lowering a limit deletes nothing: the Cahut keeps what it holds and refuses what would add to it. " +
		"For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SetInput) (Settings, error) {
		c.Log("limits", "")
		l := in.Limits
		for _, f := range fields(&l) {
			if *f.limit == nil {
				continue
			}
			if n := **f.limit; n < f.least {
				return Settings{}, operation.Invalidf("%s: give a limit of %d or more, or leave it out to lift it", f.name, f.least)
			}
		}
		before, err := Load(ctx, c.Tx)
		if err != nil {
			return Settings{}, err
		}
		_, err = c.Tx.ExecContext(ctx, `
			UPDATE settings SET people = ?, admins = ?, rooms = ?, windows = ?, windows_per_room = ?, disk = ?, set_at = ? WHERE id = 1`,
			l.People, l.Admins, l.Rooms, l.Windows, l.WindowsPerRoom, l.Disk, store.FormatTime(c.Now))
		if err != nil {
			return Settings{}, err
		}
		c.Log("limits", changes(before, l))
		return Read(ctx, c.Tx)
	},
}

// field is a limit, as set_cahut_settings checks it and the Logbook names it.
type field struct {
	name  string
	least int64 // the lowest limit it takes
	bytes bool  // it counts bytes
	limit **int64
}

func fields(l *Limits) []field {
	return []field{
		{name: "people", least: 1, limit: &l.People},
		{name: "admins", least: 1, limit: &l.Admins},
		{name: "rooms", limit: &l.Rooms},
		{name: "windows", limit: &l.Windows},
		{name: "windows_per_room", limit: &l.WindowsPerRoom},
		{name: "disk", bytes: true, limit: &l.Disk},
	}
}

// changes describes, for the Logbook, how the limits went from before to
// after, such as "people 50 → 20, disk none → 10 GB".
func changes(before, after Limits) string {
	b, a := fields(&before), fields(&after)
	var out []string
	for i := range b {
		was, is := b[i].show(), a[i].show()
		if was != is {
			out = append(out, fmt.Sprintf("%s %s → %s", b[i].name, was, is))
		}
	}
	if len(out) == 0 {
		return "unchanged"
	}
	return strings.Join(out, ", ")
}

func (f field) show() string {
	switch {
	case *f.limit == nil:
		return "none"
	case f.bytes:
		return Bytes(**f.limit)
	default:
		return fmt.Sprint(**f.limit)
	}
}

// Bytes writes n bytes in decimal units, as hosts sell disk: 1 GB is 10⁹
// bytes.
func Bytes(n int64) string {
	const unit = 1000
	if n < unit {
		return fmt.Sprintf("%d bytes", n)
	}
	v, units := float64(n), []string{"kB", "MB", "GB", "TB", "PB", "EB"}
	i := -1
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	s := fmt.Sprintf("%.1f", v)
	return strings.TrimSuffix(s, ".0") + " " + units[i]
}
