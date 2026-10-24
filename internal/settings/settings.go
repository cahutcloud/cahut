// Package settings keeps the settings of a Cahut, which the administrators
// of its server set: its kind, collaborative or individual, and the limits
// on what it holds. Every operation that could cross a limit checks it
// here, in its own transaction. See
// docs/adr/0030-kind-and-limits-of-each-cahut.md.
package settings

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Kinds of Cahuts.
const (
	// Collaborative Cahuts are those where people create team Rooms, add one
	// another to them, and open Windows to the whole Cahut.
	Collaborative = "collaborative"
	// Individual Cahuts are those where people share the Cahut, not their
	// work: every Room has one member, and people do not see one another.
	Individual = "individual"
)

// Limits bound what a Cahut holds. A limit that is nil counts nothing.
type Limits struct {
	People           *int64 `json:"people,omitempty" jsonschema:"how many people the Cahut may have, active or waiting for their activation, administrators included; disabled people do not count"`
	Admins           *int64 `json:"admins,omitempty" jsonschema:"how many of them may be administrators of the Cahut"`
	Rooms            *int64 `json:"rooms,omitempty" jsonschema:"how many team Rooms the Cahut may have, or in an individual Cahut how many Rooms beside the personal ones; personal Rooms come with their person and do not count"`
	Windows          *int64 `json:"windows,omitempty" jsonschema:"how many Windows out of the trash the Cahut may hold"`
	WindowsPerRoom   *int64 `json:"windows_per_room,omitempty" jsonschema:"how many Windows out of the trash each Room may hold, personal Rooms included"`
	Disk             *int64 `json:"disk,omitempty" jsonschema:"how many bytes the Cahut may keep: every version of every Window, those in the trash included, and the files of the house style"`
	RoomsPerPerson   *int64 `json:"rooms_per_person,omitempty" jsonschema:"in an individual Cahut, how many Rooms each person may have beside their personal Room"`
	WindowsPerPerson *int64 `json:"windows_per_person,omitempty" jsonschema:"in an individual Cahut, how many Windows out of the trash each person may hold, over their Rooms"`
	DiskPerPerson    *int64 `json:"disk_per_person,omitempty" jsonschema:"in an individual Cahut, how many bytes the Windows of each person may keep, every version included"`
}

// perPerson keeps of l the limits that bound each person.
func (l Limits) perPerson() Limits {
	return Limits{RoomsPerPerson: l.RoomsPerPerson, WindowsPerPerson: l.WindowsPerPerson, DiskPerPerson: l.DiskPerPerson}
}

// Usage is what a Cahut holds, as its limits count it.
type Usage struct {
	People  int64 `json:"people" jsonschema:"people who are not disabled, administrators included"`
	Admins  int64 `json:"admins" jsonschema:"administrators who are not disabled"`
	Rooms   int64 `json:"rooms" jsonschema:"team Rooms, or in an individual Cahut the Rooms beside the personal ones"`
	Windows int64 `json:"windows" jsonschema:"Windows out of the trash"`
	Disk    int64 `json:"disk" jsonschema:"bytes of every version of every Window, those in the trash included, and of the files of the house style"`
}

// PersonUsage is what a person of an individual Cahut holds, as the limits
// per person count it.
type PersonUsage struct {
	Rooms   int64 `json:"rooms" jsonschema:"Rooms beside the personal one"`
	Windows int64 `json:"windows" jsonschema:"Windows out of the trash, over their Rooms"`
	Disk    int64 `json:"disk" jsonschema:"bytes of every version of the Windows of their Rooms, those in the trash included"`
}

// Settings are the settings of a Cahut, with what it holds.
type Settings struct {
	Kind   string       `json:"kind" jsonschema:"collaborative, where people share team Rooms, or individual, where each person works alone and does not see the others"`
	Limits Limits       `json:"limits" jsonschema:"the limits that the administrators of the server set; a limit that is absent counts nothing. A member of an individual Cahut gets the limits per person only"`
	Usage  *Usage       `json:"usage,omitempty" jsonschema:"what the Cahut holds, as its limits count it; absent for a member of an individual Cahut"`
	Yours  *PersonUsage `json:"yours,omitempty" jsonschema:"in an individual Cahut, what the person you act for holds, as the limits per person count it"`
}

// querier runs a query, in a transaction or not.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// get returns the kind and the limits of the Cahut.
func get(ctx context.Context, q querier) (string, Limits, error) {
	var kind string
	var people, admins, rooms, windows, perRoom, disk, roomsPP, windowsPP, diskPP sql.NullInt64
	err := q.QueryRowContext(ctx, `
		SELECT kind, people, admins, rooms, windows, windows_per_room, disk, rooms_per_person, windows_per_person, disk_per_person
		FROM settings WHERE id = 1`).
		Scan(&kind, &people, &admins, &rooms, &windows, &perRoom, &disk, &roomsPP, &windowsPP, &diskPP)
	if err != nil {
		return "", Limits{}, err
	}
	return kind, Limits{
		People: ptr(people), Admins: ptr(admins), Rooms: ptr(rooms),
		Windows: ptr(windows), WindowsPerRoom: ptr(perRoom), Disk: ptr(disk),
		RoomsPerPerson: ptr(roomsPP), WindowsPerPerson: ptr(windowsPP), DiskPerPerson: ptr(diskPP),
	}, nil
}

func ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

// IsIndividual reports whether the Cahut is individual.
func IsIndividual(ctx context.Context, q querier) (bool, error) {
	var individual bool
	err := q.QueryRowContext(ctx, "SELECT kind = 'individual' FROM settings WHERE id = 1").Scan(&individual)
	return individual, err
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

// UseOf returns what the person userID holds over the Rooms they are in,
// which in an individual Cahut are theirs alone.
func UseOf(ctx context.Context, tx *sql.Tx, userID int64) (PersonUsage, error) {
	var u PersonUsage
	err := tx.QueryRowContext(ctx, `
		SELECT
		  (SELECT count(*) FROM rooms r JOIN room_members m ON m.room_id = r.id WHERE r.kind = 'team' AND m.user_id = ?1),
		  (SELECT count(*) FROM windows w JOIN room_members m ON m.room_id = w.room_id WHERE w.trashed_at IS NULL AND m.user_id = ?1),
		  (SELECT coalesce(sum(v.size), 0) FROM window_versions v JOIN windows w ON w.id = v.window_id
		     JOIN room_members m ON m.room_id = w.room_id WHERE m.user_id = ?1)`, userID).
		Scan(&u.Rooms, &u.Windows, &u.Disk)
	return u, err
}

// Read returns the settings of the Cahut, with what it holds, as the
// person userID may read them: a member of an individual Cahut gets the
// limits per person and their own use only. An administrator of the Cahut
// gets everything; userID is zero for the command line.
func Read(ctx context.Context, tx *sql.Tx, userID int64, admin bool) (Settings, error) {
	kind, l, err := get(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	s := Settings{Kind: kind, Limits: l}
	if kind == Individual && userID != 0 {
		u, err := UseOf(ctx, tx, userID)
		if err != nil {
			return Settings{}, err
		}
		s.Yours = &u
	}
	if kind == Individual && !admin {
		s.Limits = l.perPerson()
		return s, nil
	}
	u, err := Use(ctx, tx)
	if err != nil {
		return Settings{}, err
	}
	s.Usage = &u
	return s, nil
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

// SetCahutSettingsInput is the input of set_cahut_settings.
type SetCahutSettingsInput struct {
	Kind   string `json:"kind,omitempty" jsonschema:"collaborative or individual; the kind stays when empty. A Cahut becomes individual only when no Room has more than one member and no Window opens to the whole Cahut"`
	Limits Limits `json:"limits" jsonschema:"every limit of the Cahut: a limit left out is lifted"`
}

// Set sets the kind and the limits of the Cahut, for the administrators of
// the server.
var Set = &operation.Op[SetCahutSettingsInput, Settings]{
	Name:  "set_cahut_settings",
	Title: "Set the kind and limits of the Cahut",
	Description: "Set the kind of this Cahut, collaborative or individual, and its limits: how many people, administrators, " +
		"Rooms and Windows it may have, how many Windows each Room may hold, how many bytes it may keep, and, in an individual Cahut, " +
		"how many Rooms, Windows and bytes each person may have. Give every limit: a limit left out is lifted. " +
		"Lowering a limit deletes nothing: the Cahut keeps what it holds and refuses what would add to it. " +
		"For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SetCahutSettingsInput) (Settings, error) {
		c.Log("settings", "")
		if in.Kind != "" && in.Kind != Collaborative && in.Kind != Individual {
			return Settings{}, operation.Invalidf("kind %q: use collaborative or individual, or leave it out to keep the kind", in.Kind)
		}
		l := in.Limits
		for _, f := range fields(&l) {
			if *f.limit != nil && **f.limit < f.least {
				return Settings{}, operation.Invalidf("%s: give a limit of %d or more, or leave it out to lift it", f.name, f.least)
			}
		}
		was, before, err := get(ctx, c.Tx)
		if err != nil {
			return Settings{}, err
		}
		kind := cmp.Or(in.Kind, was)
		if kind == Collaborative && l.perPerson() != (Limits{}) {
			return Settings{}, operation.Invalidf("rooms_per_person, windows_per_person and disk_per_person bound the people of an individual Cahut: " +
				"leave them out, or make the Cahut individual")
		}
		if was == Collaborative && kind == Individual {
			if err := canBeIndividual(ctx, c); err != nil {
				return Settings{}, err
			}
		}
		_, err = c.Tx.ExecContext(ctx, `
			UPDATE settings SET kind = ?, people = ?, admins = ?, rooms = ?, windows = ?, windows_per_room = ?, disk = ?,
			  rooms_per_person = ?, windows_per_person = ?, disk_per_person = ?, set_at = ? WHERE id = 1`,
			kind, l.People, l.Admins, l.Rooms, l.Windows, l.WindowsPerRoom, l.Disk,
			l.RoomsPerPerson, l.WindowsPerPerson, l.DiskPerPerson, store.FormatTime(c.Now))
		if err != nil {
			return Settings{}, err
		}
		detail := changes(before, l)
		if kind != was {
			detail = strings.TrimSuffix(fmt.Sprintf("kind %s → %s, %s", was, kind, detail), ", unchanged")
		}
		c.Log("settings", detail)
		return Read(ctx, c.Tx, c.Actor.UserID, c.Admin)
	},
}

// canBeIndividual returns an error wrapping operation.ErrInvalid when a
// Room of the Cahut has more than one member, or a Window opens to the
// whole Cahut: an individual Cahut has neither. It names them.
func canBeIndividual(ctx context.Context, c *operation.Call) error {
	shared, err := names(ctx, c.Tx, `
		SELECT r.name FROM rooms r JOIN room_members m ON m.room_id = r.id
		GROUP BY r.id HAVING count(*) > 1 ORDER BY r.name`)
	if err != nil {
		return err
	}
	internal, err := names(ctx, c.Tx, `
		SELECT r.name || '/' || w.name FROM windows w JOIN rooms r ON r.id = w.room_id
		WHERE w.access = 'internal' AND (w.access_ends_at IS NULL OR w.access_ends_at > ?) ORDER BY r.name, w.name`,
		store.FormatTime(c.Now))
	if err != nil {
		return err
	}
	var why []string
	if shared != "" {
		why = append(why, "these Rooms have more than one member: "+shared+"; remove their members first")
	}
	if internal != "" {
		why = append(why, "these Windows open to everyone signed in to the Cahut: "+internal+"; set their access first")
	}
	if len(why) > 0 {
		return operation.Invalidf("this Cahut cannot become individual, where each Room keeps one member and people do not see one another: %s",
			strings.Join(why, ", and "))
	}
	return nil
}

// names returns the names that query lists, the first few of them, joined.
func names(ctx context.Context, tx *sql.Tx, query string, args ...any) (string, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	const shown = 5
	var list []string
	more := 0
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", err
		}
		if len(list) < shown {
			list = append(list, name)
		} else {
			more++
		}
	}
	if more > 0 {
		list = append(list, fmt.Sprintf("and %d more", more))
	}
	return strings.Join(list, ", "), rows.Err()
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
		{name: "rooms_per_person", limit: &l.RoomsPerPerson},
		{name: "windows_per_person", limit: &l.WindowsPerPerson},
		{name: "disk_per_person", bytes: true, limit: &l.DiskPerPerson},
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
