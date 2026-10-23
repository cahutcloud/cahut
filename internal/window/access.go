package window

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Access levels of a Window: who opens it. See
// docs/adr/0019-window-versions-access-and-declaration.md.
const (
	AccessRoom     = "room"     // the members of its Room, signed in
	AccessInternal = "internal" // everyone signed in to its Cahut
	AccessKey      = "key"      // anyone who types its Key, and the members of its Room
	AccessPublic   = "public"   // anyone who has its address
)

// A pass is what a reader gets for typing the Key of a Window. It lasts
// PassTTL, and never beyond the end of the access.
const (
	PassPrefix = "cahut_key_" //nolint:gosec // a prefix, not a credential
	PassTTL    = 30 * 24 * time.Hour
)

// effective returns the access of a Window at now, and its end: an access
// that has ended is room.
func effective(access string, ends sql.NullString, now time.Time) (string, *time.Time, error) {
	if !ends.Valid {
		return access, nil, nil
	}
	t, err := store.ParseTime(ends.String)
	if err != nil {
		return "", nil, err
	}
	if !now.Before(t) {
		return AccessRoom, nil, nil
	}
	return access, &t, nil
}

// AccessInput is the input of set_window_access.
type AccessInput struct {
	Window string     `json:"window" jsonschema:"the Window's address, or its Room and its name as room/name"`
	Access string     `json:"access" jsonschema:"room, internal, key or public"`
	EndsAt *time.Time `json:"ends_at,omitempty" jsonschema:"when the access ends, such as 2026-12-31T18:00:00Z; after it, the Window opens to its Room only. It does not end when absent"`
	Key    string     `json:"key,omitempty" jsonschema:"for key: the Key to set, of 10 to 256 characters. When empty, the Window keeps its Key, or Cahut draws one"`
	NewKey bool       `json:"new_key,omitempty" jsonschema:"for key: draw a new Key; the old one no longer opens the Window"`
}

// AccessSet is the output of set_window_access.
type AccessSet struct {
	Window
	Key string `json:"key,omitempty" jsonschema:"the new Key, returned this once: give it to the person, who passes it on with the address"`
}

func (s *Service) setAccess(ctx context.Context, c *operation.Call, in AccessInput) (AccessSet, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return AccessSet{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return AccessSet{}, err
	}
	switch in.Access {
	case AccessRoom, AccessInternal, AccessKey, AccessPublic:
	default:
		return AccessSet{}, operation.Invalidf("access %q: use room, internal, key or public", in.Access)
	}
	if in.Access == AccessPublic {
		c.Knock() // once a public Window is read, that cannot be undone
	}
	var ends sql.NullString
	if in.EndsAt != nil {
		at := in.EndsAt.UTC().Truncate(time.Millisecond)
		switch {
		case in.Access == AccessRoom:
			return AccessSet{}, operation.Invalidf("a Window open to its Room only has no end: give no ends_at")
		case !at.After(c.Now):
			return AccessSet{}, operation.Invalidf("ends_at %s is past", at.Format(time.RFC3339))
		}
		ends = sql.NullString{String: store.FormatTime(at), Valid: true}
		in.EndsAt = &at
	}
	switch {
	case in.Access != AccessKey && (in.Key != "" || in.NewKey):
		return AccessSet{}, operation.Invalidf("key and new_key go with the access key")
	case in.Key != "" && in.NewKey:
		return AccessSet{}, operation.Invalidf("give a key, or ask for a new one, not both")
	}

	var key string
	switch {
	case in.Access != AccessKey:
	case in.Key != "":
		if err := credential.CheckKey(in.Key); err != nil {
			return AccessSet{}, err
		}
		key = in.Key
	case in.NewKey || !w.hasKey:
		key = credential.Key()
	}
	set, args := "access = ?, access_ends_at = ?", []any{in.Access, ends}
	switch {
	case key != "":
		hash, err := credential.HashPassword(ctx, key)
		if err != nil {
			return AccessSet{}, err
		}
		set, args = set+", key_hash = ?", append(args, hash)
	case in.Access != AccessKey:
		set += ", key_hash = NULL"
	}
	// A new Key, or no Key at all, ends the passes of the old one.
	if key != "" || in.Access != AccessKey {
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_passes WHERE window_id = ?", w.id); err != nil {
			return AccessSet{}, err
		}
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE windows SET "+set+" WHERE id = ?" /* #nosec G202 -- set holds constant clauses */, append(args, w.id)...); err != nil {
		return AccessSet{}, err
	}

	w.Access, w.AccessEndsAt = in.Access, in.EndsAt
	detail := in.Access
	if key != "" {
		detail += ", new Key" // never the Key itself
	}
	if in.EndsAt != nil {
		detail += " until " + in.EndsAt.Format(time.RFC3339)
	}
	c.Log(w.Room+"/"+w.Name, detail)
	return AccessSet{Window: w.Window, Key: key}, nil
}

// Served is a Window, out of the trash, as its readers reach it. Serving a
// Window is not an operation: readers are not actors, and their visits are
// not logged.
type Served struct {
	ID, RoomID        int64
	Room, Name, Title string
	Kind              string
	Version           int
	Access            string // as it is now: an access that has ended is room
	EndsAt            *time.Time
	UpdatedAt         time.Time
	Declaration       *Declaration // of the current version; nil when it was published before declarations
	keyHash           string
}

const served = `
	SELECT w.id, w.room_id, r.name, w.name, w.title, w.kind, w.version, w.access, w.access_ends_at, w.updated_at, w.key_hash, v.declaration
	FROM windows w JOIN rooms r ON r.id = w.room_id JOIN window_versions v ON v.window_id = w.id AND v.number = w.version
	WHERE w.trashed_at IS NULL AND `

// Lookup returns the Window called name in the Room called room.
func Lookup(ctx context.Context, db *store.DB, room, name string, now time.Time) (Served, error) {
	w, err := scanServed(db.Read.QueryRowContext(ctx, served+"r.name = ? AND w.name = ?", room, name), now)
	if errors.Is(err, sql.ErrNoRows) {
		return Served{}, operation.NotFoundf("no Window %s/%s", room, name)
	}
	return w, err
}

// LookupID returns the Window with the given id.
func LookupID(ctx context.Context, db *store.DB, id int64, now time.Time) (Served, error) {
	w, err := scanServed(db.Read.QueryRowContext(ctx, served+"w.id = ?", id), now)
	if errors.Is(err, sql.ErrNoRows) {
		return Served{}, operation.NotFoundf("no Window %d", id)
	}
	return w, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanServed(row scanner, now time.Time) (Served, error) {
	var w Served
	var ends, keyHash sql.NullString
	var updated string
	var declared *string
	err := row.Scan(&w.ID, &w.RoomID, &w.Room, &w.Name, &w.Title, &w.Kind, &w.Version, &w.Access, &ends, &updated, &keyHash, &declared)
	if err != nil {
		return Served{}, err
	}
	if w.Declaration, err = decodeDeclaration(declared); err != nil {
		return Served{}, err
	}
	if w.UpdatedAt, err = store.ParseTime(updated); err != nil {
		return Served{}, err
	}
	w.keyHash = keyHash.String
	w.Access, w.EndsAt, err = effective(w.Access, ends, now)
	return w, err
}

// Opening says how a reader opens a Window.
type Opening int

// Ways to open a Window.
const (
	Closed    Opening = iota // the reader may not open it
	ByAddress                // it is public
	BySession                // the person of the reader's session may open it
	ByPass                   // the reader typed its Key
)

// Opens says how a reader opens the Window w: userID is the person of their
// reader session, or zero, and pass the id of a pass of theirs for w, or
// zero.
func (w Served) Opens(ctx context.Context, db *store.DB, userID, pass int64) (Opening, error) {
	if w.Access == AccessPublic {
		return ByAddress, nil
	}
	if userID != 0 {
		if w.Access == AccessInternal {
			return BySession, nil
		}
		var member bool
		err := db.Read.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = ? AND user_id = ?)",
			w.RoomID, userID).Scan(&member)
		if err != nil {
			return Closed, err
		}
		if member {
			return BySession, nil
		}
	}
	if pass != 0 && w.Access == AccessKey {
		return ByPass, nil
	}
	return Closed, nil
}

// ErrWrongKey means that a Key does not open a Window.
var ErrWrongKey = errors.New("this Key does not open this Window")

// OpenWithKey checks the Key that a reader typed for the Window w, and
// gives them a pass when it opens w. Like Keys, passes are not logged.
func OpenWithKey(ctx context.Context, db *store.DB, w Served, key string, now time.Time) (token string, expires time.Time, err error) {
	if w.Access != AccessKey || w.keyHash == "" {
		return "", time.Time{}, ErrWrongKey
	}
	ok, err := credential.CheckPassword(ctx, w.keyHash, strings.TrimSpace(key))
	if err != nil {
		return "", time.Time{}, err
	}
	if !ok {
		return "", time.Time{}, ErrWrongKey
	}
	expires = now.Add(PassTTL)
	if w.EndsAt != nil && w.EndsAt.Before(expires) {
		expires = *w.EndsAt
	}
	if _, err := db.Write.ExecContext(ctx, "DELETE FROM window_passes WHERE expires_at <= ?", store.FormatTime(now)); err != nil {
		return "", time.Time{}, err
	}
	// The Key may have changed since it was read: the pass is only for the
	// Key that the reader typed.
	token = credential.Token(PassPrefix)
	res, err := db.Write.ExecContext(ctx, `
		INSERT INTO window_passes (window_id, hash, created_at, expires_at)
		SELECT id, ?, ?, ? FROM windows WHERE id = ? AND key_hash = ?`,
		credential.Hash(token), store.FormatTime(now), store.FormatTime(expires), w.ID, w.keyHash)
	if err != nil {
		return "", time.Time{}, err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return "", time.Time{}, errors.Join(ErrWrongKey, err)
	}
	return token, expires, nil
}

// Pass returns the id of the pass token for the Window id, while it lasts,
// or zero.
func Pass(ctx context.Context, db *store.DB, windowID int64, token string, now time.Time) (int64, error) {
	if !strings.HasPrefix(token, PassPrefix) {
		return 0, nil
	}
	var id int64
	err := db.Read.QueryRowContext(ctx, "SELECT id FROM window_passes WHERE hash = ? AND window_id = ? AND expires_at > ?",
		credential.Hash(token), windowID, store.FormatTime(now)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// PassLasts reports whether the pass id still opens the Window windowID,
// as far as the pass goes: the access of the Window still applies.
func PassLasts(ctx context.Context, db *store.DB, windowID, id int64, now time.Time) (bool, error) {
	var ok bool
	err := db.Read.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM window_passes WHERE id = ? AND window_id = ? AND expires_at > ?)",
		id, windowID, store.FormatTime(now)).Scan(&ok)
	return ok, err
}

// RoomView is a Room as a reader sees it, with the Windows they open.
type RoomView struct {
	Name, Title string
	Member      bool
	Windows     []Served
}

// ViewRoom returns the Room called name, with the Windows that a reader
// opens, most recently updated first: userID is the person of their reader
// session, or zero, and passes their passes by the id of their Window.
func ViewRoom(ctx context.Context, db *store.DB, name string, userID int64, passes map[int64]string, now time.Time) (RoomView, error) {
	v := RoomView{Name: name}
	var roomID int64
	err := db.Read.QueryRowContext(ctx, `
		SELECT r.id, r.title, EXISTS (SELECT 1 FROM room_members WHERE room_id = r.id AND user_id = ?)
		FROM rooms r WHERE r.name = ?`, userID, name).Scan(&roomID, &v.Title, &v.Member)
	if errors.Is(err, sql.ErrNoRows) {
		return RoomView{}, operation.NotFoundf("no Room %s", name)
	}
	if err != nil {
		return RoomView{}, err
	}
	rows, err := db.Read.QueryContext(ctx, served+"w.room_id = ? ORDER BY w.updated_at DESC, w.id DESC", roomID)
	if err != nil {
		return RoomView{}, err
	}
	var all []Served
	for rows.Next() {
		w, err := scanServed(rows, now)
		if err != nil {
			rows.Close()
			return RoomView{}, err
		}
		all = append(all, w)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return RoomView{}, err
	}
	for _, w := range all {
		open := v.Member || w.Access == AccessPublic || w.Access == AccessInternal && userID != 0
		if !open && w.Access == AccessKey && passes[w.ID] != "" {
			pass, err := Pass(ctx, db, w.ID, passes[w.ID], now)
			if err != nil {
				return RoomView{}, err
			}
			open = pass != 0
		}
		if open {
			v.Windows = append(v.Windows, w)
		}
	}
	return v, nil
}
