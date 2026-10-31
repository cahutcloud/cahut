package window

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Access levels of a Window: who opens it. See
// docs/adr/0019-window-versions-access-and-declaration.md and
// docs/adr/0053-windows-for-their-author-alone.md.
const (
	AccessAuthor   = "author"   // its author, signed in, while they are in its Room
	AccessRoom     = "room"     // the members of its Room, signed in
	AccessInternal = "internal" // everyone signed in to its Cahut
	AccessKey      = "key"      // anyone who types its Key, and the members of its Room
	AccessPublic   = "public"   // anyone who has its address
)

// The refusals of a Window that its author keeps for themself, which the
// console says in the language of its person. See
// docs/adr/0053-windows-for-their-author-alone.md.
var (
	authorOnly = operation.NewRefusalOf[ByAuthor]("window_author_only", operation.ErrForbidden,
		"only the author of {room}/{window}, {author}, keeps it for themself")
	withoutAuthor = operation.NewRefusalOf[At]("window_without_author", operation.ErrInvalid,
		"{room}/{window} has no author since the person it was published for was deleted: "+
			"it opens to its Room, its Cahut, a Key or the public, and to no author alone")
)

// ByAuthor fills the refusal of a Window that someone else than its author
// would keep for its author alone.
type ByAuthor struct {
	Room   string `json:"room" jsonschema:"the Room of the Window"`
	Window string `json:"window" jsonschema:"the name of the Window"`
	Author string `json:"author" jsonschema:"the handle of its author"`
}

// keptBy returns an error unless the person the call acts for may keep the
// Window w for themself: its author.
func (w found) keptBy(c *operation.Call) error {
	switch {
	case w.authorID == 0:
		return withoutAuthor(At{Room: w.Room, Window: w.Name})
	case w.authorID != c.Actor.UserID:
		return authorOnly(ByAuthor{Room: w.Room, Window: w.Name, Author: w.CreatedBy})
	}
	return nil
}

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
	Window string     `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	Access string     `json:"access" jsonschema:"author, room, internal, key or public"`
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
	if err := w.changeable(); err != nil {
		return AccessSet{}, err
	}
	switch in.Access {
	case AccessAuthor, AccessRoom, AccessInternal, AccessKey, AccessPublic:
	default:
		return AccessSet{}, operation.Invalidf("access %q: use author, room, internal, key or public", in.Access)
	}
	if in.Access == AccessAuthor {
		if err := w.keptBy(c); err != nil {
			return AccessSet{}, err
		}
	}
	// The Room reads the entry that takes the Window from it, and the one
	// that gives it back.
	if (w.Access == AccessAuthor) != (in.Access == AccessAuthor) {
		c.ReadBy(0)
	}
	if in.Access == AccessInternal {
		individual, err := settings.IsIndividual(ctx, c.Tx)
		if err != nil {
			return AccessSet{}, err
		}
		if individual {
			return AccessSet{}, operation.Invalidf("this Cahut is individual, and its people do not see one another: " +
				"a Window opens to its Room, to anyone with its Key, or to the public, not to everyone signed in to the Cahut")
		}
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
		case in.Access == AccessAuthor:
			return AccessSet{}, operation.Invalidf("a Window for its author alone has no end: give no ends_at")
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
	// Who sets the access hears when it ends. See
	// docs/adr/0031-e-mail-through-a-relay.md. When it is set tells the
	// moderators which Windows were made public last. See
	// docs/adr/0054-reporting-and-withdrawing-a-window.md.
	set, args := "access = ?, access_ends_at = ?, access_set_by = ?, access_set_at = ?",
		[]any{in.Access, ends, sql.NullInt64{Int64: c.Actor.UserID, Valid: c.Actor.UserID != 0}, store.FormatTime(c.Now)}
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
	// A new Key, or no Key at all, ends the passes of the old one, and the
	// invitations that stood for it. The passes of share links stand.
	if key != "" || in.Access != AccessKey {
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_passes WHERE window_id = ? AND link_id IS NULL", w.id); err != nil {
			return AccessSet{}, err
		}
		if err := endKeyInvitations(ctx, c, w.id); err != nil {
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
	// Its Knock says what anyone will write. See
	// docs/adr/0056-a-database-for-each-window.md.
	if writes := w.Declaration.writesForAnyone(); in.Access == AccessPublic && len(writes) > 0 {
		detail += "; anyone writes with " + strings.Join(writes, ", ")
	}
	c.Log(w.Room+"/"+w.Name, detail)
	return AccessSet{Window: w.Window, Key: key}, nil
}

// Served is a Window, out of the trash, as its readers reach it. Serving a
// Window is not an operation: readers are not actors, and their visits are
// not logged.
type Served struct {
	ID, RoomID        int64
	UID               string // the id that people and AIs see, which never changes
	Room, Name, Title string
	Description       string
	Kind              string
	Version           int
	Access            string // as it is now: an access that has ended is room
	AuthorID          int64  // the person it was first published for; zero once they are deleted
	EndsAt            *time.Time
	UpdatedAt         time.Time
	Declaration       *Declaration // of the current version; nil when it was published before declarations
	keyHash           string
	withdrawn         bool
}

// servedAll selects the Windows out of the trash, as readers reach them, and
// served those among them that the moderators have not withdrawn.
const (
	servedAll = `
	SELECT w.id, w.room_id, w.uid, r.name, w.name, w.title, w.description, w.kind, w.version, w.access, coalesce(w.created_by, 0), w.access_ends_at, w.updated_at, w.key_hash, v.declaration,
		w.withdrawn_at IS NOT NULL
	FROM windows w JOIN rooms r ON r.id = w.room_id JOIN window_versions v ON v.window_id = w.id AND v.number = w.version
	WHERE w.trashed_at IS NULL AND `
	served = servedAll + "w.withdrawn_at IS NULL AND "
)

// Withdrawn is the error of a lookup that finds a Window that the
// moderators of its Cahut withdrew: it serves nothing to its readers until
// they reinstate it. It is an ErrNotFound, so that a reader who may not
// learn of it meets what is not found; the pages that may say it was
// withdrawn find the Window in it, and the console still previews it for
// the members of its Room. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
type Withdrawn struct {
	Window Served
}

func (e *Withdrawn) Error() string {
	return "not found: " + e.Window.Room + "/" + e.Window.Name + " was withdrawn by the moderators of its Cahut"
}

func (e *Withdrawn) Unwrap() error { return operation.ErrNotFound }

// lookup returns the Window that where finds with args, or a Withdrawn
// error when the moderators withdrew it, or what is not found, which
// missing says.
func lookup(ctx context.Context, db *store.DB, now time.Time, missing error, where string, args ...any) (Served, error) {
	return lookupIn(ctx, db.Read, now, missing, where, args...)
}

func lookupIn(ctx context.Context, q account.RowQuerier, now time.Time, missing error, where string, args ...any) (Served, error) {
	w, err := scanServed(q.QueryRowContext(ctx, servedAll+where, args...), now)
	if errors.Is(err, sql.ErrNoRows) {
		return Served{}, missing
	}
	if err == nil && w.withdrawn {
		return Served{}, &Withdrawn{Window: w}
	}
	return w, err
}

// Lookup returns the Window called name in the Room called room.
func Lookup(ctx context.Context, db *store.DB, room, name string, now time.Time) (Served, error) {
	return lookup(ctx, db, now, operation.NotFoundf("no Window %s/%s", room, name), "r.name = ? AND w.name = ?", room, name)
}

// LookupID returns the Window with the given id.
func LookupID(ctx context.Context, db *store.DB, id int64, now time.Time) (Served, error) {
	return lookup(ctx, db, now, operation.NotFoundf("no Window %d", id), "w.id = ?", id)
}

// LookupUID returns the Window whose id, as people and AIs see it, is uid.
func LookupUID(ctx context.Context, db *store.DB, uid string, now time.Time) (Served, error) {
	return lookup(ctx, db, now, operation.NotFoundf("no Window %s", uid), "w.uid = ?", uid)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanServed(row scanner, now time.Time) (Served, error) {
	var w Served
	var ends, keyHash sql.NullString
	var updated string
	var declared *string
	err := row.Scan(&w.ID, &w.RoomID, &w.UID, &w.Room, &w.Name, &w.Title, &w.Description, &w.Kind, &w.Version, &w.Access, &w.AuthorID, &ends, &updated, &keyHash, &declared,
		&w.withdrawn)
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
// zero. The pass of a Key opens it while its access is key, and the pass of
// a share link whatever its access. A Window for its author alone opens to
// no other session.
func (w Served) Opens(ctx context.Context, db *store.DB, userID, pass int64) (Opening, error) {
	return w.opensIn(ctx, db.Read, userID, pass)
}

func (w Served) opensIn(ctx context.Context, q account.RowQuerier, userID, pass int64) (Opening, error) {
	if w.Access == AccessPublic {
		return ByAddress, nil
	}
	if userID != 0 && (w.Access != AccessAuthor || userID == w.AuthorID) {
		if w.Access == AccessInternal {
			return BySession, nil
		}
		var member bool
		err := q.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = ? AND user_id = ?)",
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
	if pass != 0 {
		var link bool
		err := q.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM window_passes WHERE id = ? AND window_id = ? AND link_id IS NOT NULL)",
			pass, w.ID).Scan(&link)
		if err != nil {
			return Closed, err
		}
		if link {
			return ByPass, nil
		}
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
	return pass(ctx, db, windowID, token, now, "")
}

// pass returns the id of the pass token for the Window id, while it lasts
// and its row matches the clause and, or zero.
func pass(ctx context.Context, db *store.DB, windowID int64, token string, now time.Time, and string) (int64, error) {
	if !strings.HasPrefix(token, PassPrefix) {
		return 0, nil
	}
	var id int64
	err := db.Read.QueryRowContext(ctx, "SELECT id FROM window_passes WHERE hash = ? AND window_id = ? AND expires_at > ?"+and, /* #nosec G202 -- and is a constant clause */
		credential.Hash(token), windowID, store.FormatTime(now)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// PassLasts reports whether the pass id still opens the Window windowID,
// as far as the pass goes: the access of the Window still applies.
func PassLasts(ctx context.Context, db *store.DB, windowID, id int64, now time.Time) (bool, error) {
	return passLastsIn(ctx, db.Read, windowID, id, now)
}

func passLastsIn(ctx context.Context, q account.RowQuerier, windowID, id int64, now time.Time) (bool, error) {
	var ok bool
	err := q.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM window_passes WHERE id = ? AND window_id = ? AND expires_at > ?)",
		id, windowID, store.FormatTime(now)).Scan(&ok)
	return ok, err
}

// RoomView is a Room as a reader sees it, with the Windows they open.
type RoomView struct {
	Name, Title, Description string
	Member                   bool
	Windows                  []Opened
	// Home is the Window that the Room shows at its address, when it has
	// one and the reader opens it; nil otherwise, and the address lists
	// Windows. See docs/adr/0032-the-page-of-a-room.md.
	Home *Opened
}

// Opened is a Window that a reader opens, and how.
type Opened struct {
	Served
	How  Opening
	Pass int64 // the id of the reader's pass, when they open it by its Key
}

// ViewRoom returns the Room called name, with the Windows that a reader
// opens, most recently updated first: userID is the person of their reader
// session, or zero, and passes their passes by the id of their Window.
func ViewRoom(ctx context.Context, db *store.DB, name string, userID int64, passes map[int64]string, now time.Time) (RoomView, error) {
	v := RoomView{Name: name}
	var roomID int64
	var home sql.NullInt64
	err := db.Read.QueryRowContext(ctx, `
		SELECT r.id, r.title, r.description, r.home_window_id, EXISTS (SELECT 1 FROM room_members WHERE room_id = r.id AND user_id = ?)
		FROM rooms r WHERE r.name = ?`, userID, name).Scan(&roomID, &v.Title, &v.Description, &home, &v.Member)
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
	// The Windows open as Opens has it, from what the query read.
	for _, w := range all {
		o := Opened{Served: w}
		switch {
		case w.Access == AccessPublic:
			o.How = ByAddress
		case w.Access == AccessAuthor:
			if userID != 0 && v.Member && userID == w.AuthorID {
				o.How = BySession
			}
		case userID != 0 && (v.Member || w.Access == AccessInternal):
			o.How = BySession
		case w.Access == AccessKey && passes[w.ID] != "":
			// A share link opens its Window, not the page of its Room.
			if o.Pass, err = pass(ctx, db, w.ID, passes[w.ID], now, " AND link_id IS NULL"); err != nil {
				return RoomView{}, err
			}
			if o.Pass != 0 {
				o.How = ByPass
			}
		}
		if o.How == Closed {
			continue
		}
		v.Windows = append(v.Windows, o)
		if home.Valid && w.ID == home.Int64 {
			v.Home = &o
		}
	}
	return v, nil
}
