package window

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Invitations send the address of a Window to people by e-mail. A person
// of the Cahut, named by handle, signs in to read it. Anyone else, named by
// address, reads a Window with a Key through a personal link, which stands
// for the Key for them alone and is revoked alone, or a public Window at its
// address. The Key itself never goes by e-mail. See
// docs/adr/0031-e-mail-through-a-relay.md.

// Bounds of invitations.
const (
	MaxInvited       = 20                  // recipients of one call
	MaxInvitedPerDay = 100                 // invitations of one person in 24 hours, their AIs included
	InvitationTTL    = 30 * 24 * time.Hour // of a personal link, and never beyond the end of the access
	MaxNote          = 500                 // characters of a note
	openingTTL       = time.Minute         // of the code that the console hands to /-/invitation
)

// The refusals of an invitation that people meet on the page of a Window,
// which the console says in their language.
var (
	errNoMail = operation.NewRefusal("no_mail_to_invite", operation.ErrInvalid,
		"this Cahut sends no e-mail: give them the address of the Window yourself, with its Key if it has one")
	tooManyInvited = operation.NewRefusalOf[Sent]("too_many_invitations", operation.ErrTooMany,
		"{sent} invitations went out for {handle} in the last 24 hours, and {limit} is the most: send the rest later")
	badAddress = operation.NewRefusalOf[Addressee]("invite_bad_address", operation.ErrInvalid,
		"to: {to} is not an e-mail address: give one address alone, such as jane@example.com")
	addressToRoom = operation.NewRefusalOf[At]("invite_address_room_only", operation.ErrInvalid,
		"{room}/{window} opens to its Room only: invite people of the Cahut by handle, "+
			"or set its access to key first, so that each address gets a personal link")
	addressToInternal = operation.NewRefusalOf[At]("invite_address_internal", operation.ErrInvalid,
		"{room}/{window} opens to the people of the Cahut only: invite people of the Cahut by handle, "+
			"or set its access to key first, so that each address gets a personal link")
	handleInIndividual = operation.NewRefusalOf[Addressee]("invite_by_address", operation.ErrInvalid,
		"{to}: in this Cahut, which is individual, invite by e-mail address, to a Window with a Key or public")
	noInvitee = operation.NewRefusalOf[Addressee]("invite_no_person", operation.ErrNotFound,
		"no person {to} in this Cahut: invite someone else by e-mail address")
	inviteeDisabled = operation.NewRefusalOf[Invitee]("invite_disabled", operation.ErrInvalid,
		"{handle} is disabled, and cannot read the Window")
	outsider = operation.NewRefusalOf[Outsider]("invite_not_in_room", operation.ErrInvalid,
		"{handle} is not in the Room {room}, and {room}/{window} opens to its Room only: "+
			"add them to the Room, or set its access to internal or key first")
)

// Sent fills the refusal of invitations past the limit of a day.
type Sent struct {
	Sent   int    `json:"sent" jsonschema:"how many invitations went out in the last 24 hours"`
	Handle string `json:"handle" jsonschema:"the person who invites, or whom the AI acts for"`
	Limit  int    `json:"limit" jsonschema:"the most invitations of one person in 24 hours"`
}

// Addressee fills the refusal of a recipient of an invitation.
type Addressee struct {
	To operation.Quoted `json:"to" jsonschema:"the recipient, as the invitation names them"`
}

// At fills the refusal of an invitation by address to a Window that opens
// to people of the Cahut only.
type At struct {
	Room   string `json:"room" jsonschema:"the Room of the Window"`
	Window string `json:"window" jsonschema:"the name of the Window"`
}

// Invitee fills the refusal of a person of the Cahut whom an invitation
// cannot reach.
type Invitee struct {
	Handle string `json:"handle" jsonschema:"the handle of the person"`
}

// Outsider fills the refusal of a person outside the Room of a Window that
// opens to its Room only.
type Outsider struct {
	Handle string `json:"handle" jsonschema:"the handle of the person"`
	Room   string `json:"room" jsonschema:"the Room of the Window"`
	Window string `json:"window" jsonschema:"the name of the Window"`
}

// InviteInput is the input of invite_to_window.
type InviteInput struct {
	Window string   `json:"window" jsonschema:"the Window's address, or its Room and its name as room/name"`
	To     []string `json:"to" jsonschema:"who to invite, 20 at most: handles of people of the Cahut, who sign in to read it, or e-mail addresses of anyone else, for a Window with a Key or public"`
	Note   string   `json:"note,omitempty" jsonschema:"a few words for them, 500 characters at most, in plain text with **bold**, *italics*, and lists whose lines start with - or 1."`
}

// Invitation is an invitation to read a Window.
type Invitation struct {
	ID        int64      `json:"id"`
	To        string     `json:"to" jsonschema:"the handle of the person invited, or the address"`
	By        string     `json:"by,omitempty" jsonschema:"the handle of the person who invited"`
	AI        string     `json:"ai,omitempty" jsonschema:"the AI that sent the invitation for them"`
	Link      bool       `json:"link" jsonschema:"whether it carries a personal link, which opens the Window for this recipient alone"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" jsonschema:"when its personal link stops working"`
	EndedAt   *time.Time `json:"ended_at,omitempty" jsonschema:"when it was revoked, or ended with the Key or the access it stood for"`
}

// Invited is the output of invite_to_window.
type Invited struct {
	Window      string       `json:"window"`
	Invitations []Invitation `json:"invitations"`
	Unreachable []string     `json:"unreachable,omitempty" jsonschema:"people of the Cahut who have no e-mail address, to whom nothing went"`
}

// recipient is someone to invite, once checked.
type recipient struct {
	to     string // as the invitation names them: a handle or an address
	userID int64  // the person of the Cahut, or zero
	name   string // the name of the person, or empty
	email  string // where the invitation goes
	link   bool   // it carries a personal link
}

func (s *Service) invite(ctx context.Context, c *operation.Call, in InviteInput) (Invited, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Invited{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return Invited{}, err
	}
	at := w.Room + "/" + w.Name
	to := unique(in.To)
	switch {
	case len(to) == 0:
		return Invited{}, operation.Invalidf("to: name someone to invite, by handle or by e-mail address")
	case len(to) > MaxInvited:
		return Invited{}, operation.Invalidf("to: %d people, and an invitation goes to %d at most: send several", len(to), MaxInvited)
	}
	note, err := cleanNote(in.Note)
	if err != nil {
		return Invited{}, err
	}
	o := mail.OutboxOf(c)
	if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
		return Invited{}, errors.Join(err, errNoMail)
	}
	if c.Actor.UserID != 0 {
		var sent int
		err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM invitations WHERE invited_by = ? AND created_at > ?",
			c.Actor.UserID, store.FormatTime(c.Now.Add(-24*time.Hour))).Scan(&sent)
		if err != nil {
			return Invited{}, err
		}
		if sent+len(to) > MaxInvitedPerDay {
			return Invited{}, tooManyInvited(Sent{Sent: sent, Handle: c.Actor.Handle, Limit: MaxInvitedPerDay})
		}
	}
	individual, err := settings.IsIndividual(ctx, c.Tx)
	if err != nil {
		return Invited{}, err
	}
	var list []recipient
	var unreachable []string
	for _, t := range to {
		r, err := s.recipient(ctx, c, w, t, individual)
		if err != nil {
			return Invited{}, err
		}
		if r.email == "" {
			unreachable = append(unreachable, r.to)
			continue
		}
		list = append(list, r)
	}
	c.Log(at, "to "+strings.Join(to, ", "))

	inviter, err := inviterName(ctx, c)
	if err != nil {
		return Invited{}, err
	}
	_, where, err := o.Where(ctx)
	if err != nil {
		return Invited{}, err
	}
	expires := c.Now.Add(InvitationTTL)
	if w.AccessEndsAt != nil && w.AccessEndsAt.Before(expires) {
		expires = *w.AccessEndsAt
	}
	out := Invited{Window: at, Invitations: []Invitation{}, Unreachable: unreachable}
	for _, r := range list {
		inv := Invitation{To: r.to, By: c.Actor.Handle, Link: r.link, CreatedAt: c.Now}
		var hash []byte
		var ends sql.NullString
		link := w.URL
		if r.link {
			code := credential.Token("")
			hash, ends = credential.Hash(code), sql.NullString{String: store.FormatTime(expires), Valid: true}
			inv.ExpiresAt = &expires
			link = o.Link("invitation#" + code)
		}
		if c.Actor.Kind == operation.AI {
			inv.AI = c.Actor.Client
		}
		err := c.Tx.QueryRowContext(ctx, `
			INSERT INTO invitations (window_id, invited_by, client, user_id, email, note, hash, created_at, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			w.id, nullID(c.Actor.UserID), nullString(inv.AI), nullID(r.userID), nullString(emailOf(r)), note, hash,
			store.FormatTime(c.Now), ends).Scan(&inv.ID)
		if err != nil {
			return Invited{}, err
		}
		o.Post(c, invitationLetter(r, w.Window, inviter, inv.AI, where, note, link, expires))
		out.Invitations = append(out.Invitations, inv)
	}
	return out, nil
}

// recipient checks the recipient t, a handle or an address, of an
// invitation to the Window w.
func (s *Service) recipient(ctx context.Context, c *operation.Call, w found, t string, individual bool) (recipient, error) {
	if strings.Contains(t, "@") {
		email, err := mail.Address(t)
		if err != nil {
			return recipient{}, badAddress(Addressee{To: operation.Quoted(t)}) // t is trimmed, as mail.Address trims it
		}
		switch w.Access {
		case AccessKey:
			return recipient{to: email, email: email, link: true}, nil
		case AccessPublic:
			return recipient{to: email, email: email}, nil
		case AccessInternal:
			return recipient{}, addressToInternal(At{Room: w.Room, Window: w.Name})
		}
		return recipient{}, addressToRoom(At{Room: w.Room, Window: w.Name})
	}
	if individual {
		return recipient{}, handleInIndividual(Addressee{To: operation.Quoted(t)})
	}
	handle := strings.ToLower(t)
	var r recipient
	var email, disabled sql.NullString
	var member bool
	err := c.Tx.QueryRowContext(ctx, `
		SELECT id, name, email, disabled_at, EXISTS (SELECT 1 FROM room_members m WHERE m.room_id = ? AND m.user_id = u.id)
		FROM users u WHERE handle = ?`, w.roomID, handle).Scan(&r.userID, &r.name, &email, &disabled, &member)
	if errors.Is(err, sql.ErrNoRows) {
		return recipient{}, noInvitee(Addressee{To: operation.Quoted(handle)})
	}
	if err != nil {
		return recipient{}, err
	}
	if disabled.Valid {
		return recipient{}, inviteeDisabled(Invitee{Handle: handle})
	}
	r.to, r.email = handle, email.String
	switch {
	case member || w.Access == AccessInternal || w.Access == AccessPublic:
	case w.Access == AccessKey:
		r.link = true
	default:
		return recipient{}, outsider(Outsider{Handle: handle, Room: w.Room, Window: w.Name})
	}
	return r, nil
}

// unique trims the recipients of an invitation and drops the repeated
// ones, whatever their case.
func unique(to []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, t := range to {
		t = strings.TrimSpace(t)
		if t == "" || seen[strings.ToLower(t)] {
			continue
		}
		seen[strings.ToLower(t)] = true
		out = append(out, t)
	}
	return out
}

// cleanNote returns the note of an invitation, trimmed, or an error when it
// is too long or holds control characters other than line breaks.
func cleanNote(note string) (string, error) {
	note = strings.TrimSpace(strings.ReplaceAll(note, "\r\n", "\n"))
	if n := utf8.RuneCountInString(note); n > MaxNote {
		return "", operation.Invalidf("note: %d characters, and a note holds %d at most", n, MaxNote)
	}
	for _, r := range note {
		if unicode.IsControl(r) && r != '\n' || r == utf8.RuneError {
			return "", operation.Invalidf("note: write it without control characters")
		}
	}
	return note, nil
}

// inviterName returns the name of the person who invites, or who the AI
// that invites acts for.
func inviterName(ctx context.Context, c *operation.Call) (string, error) {
	if c.Actor.UserID == 0 {
		return "An administrator of the server", nil
	}
	var name string
	err := c.Tx.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", c.Actor.UserID).Scan(&name)
	return name, err
}

// invitationLetter makes the message of an invitation to r to read the
// Window w, from inviter, sent by the AI ai when one did.
func invitationLetter(r recipient, w Window, inviter, ai, where, note, link string, expires time.Time) mail.Letter {
	l := mail.Letter{
		To:      r.email,
		Subject: inviter + " invites you to read " + w.Title,
		Lines:   []string{inviter + " invites you to read " + w.Title + ", on " + where + "."},
		Action:  &mail.Action{Label: "Open " + w.Title, URL: link},
	}
	if r.name != "" {
		l.Lines = append([]string{"Hello " + r.name + ","}, l.Lines...)
	}
	if ai != "" {
		l.Lines = append(l.Lines, ai+" sent this invitation for "+inviter+".")
	}
	if note != "" {
		by := inviter
		if ai != "" {
			by = ai + ", for " + inviter
		}
		l.Quote = &mail.Quote{By: by, Text: note}
	}
	switch {
	case r.link:
		l.After = []string{"This link is yours: it opens " + w.Title + " on any device until " + mail.Day(expires) +
			", unless the invitation is taken back. Keep it to yourself."}
	case r.userID != 0:
		l.After = []string{"Sign in to " + where + " as " + r.to + " to read it."}
	default:
		l.After = []string{"Anyone with this address can read it."}
	}
	return l
}

func emailOf(r recipient) string {
	if r.userID != 0 {
		return ""
	}
	return r.email
}

func nullID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id != 0}
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ListInvitationsInput is the input of list_invitations.
type ListInvitationsInput struct {
	Window string `json:"window" jsonschema:"the Window's address, or its Room and its name as room/name"`
}

// InvitationList is the output of list_invitations.
type InvitationList struct {
	Invitations []Invitation `json:"invitations" jsonschema:"the latest first"`
}

func (s *Service) listInvitations(ctx context.Context, c *operation.Call, in ListInvitationsInput) (InvitationList, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return InvitationList{}, err
	}
	list, err := invitations(ctx, c.Tx, "i.window_id = ?", w.id)
	return InvitationList{Invitations: list}, err
}

// invitations returns the invitations that where selects, the latest first.
func invitations(ctx context.Context, tx *sql.Tx, where string, args ...any) ([]Invitation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT i.id, coalesce(t.handle, i.email), coalesce(b.handle, ''), coalesce(i.client, ''), i.hash IS NOT NULL,
			i.created_at, i.expires_at, i.ended_at
		FROM invitations i LEFT JOIN users t ON t.id = i.user_id LEFT JOIN users b ON b.id = i.invited_by
		WHERE `+where+` ORDER BY i.id DESC LIMIT 500`, /* #nosec G202 -- where is a constant clause */ args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Invitation{}
	for rows.Next() {
		var inv Invitation
		var created string
		var expires, ended sql.NullString
		if err := rows.Scan(&inv.ID, &inv.To, &inv.By, &inv.AI, &inv.Link, &created, &expires, &ended); err != nil {
			return nil, err
		}
		if inv.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if inv.ExpiresAt, err = parseNull(expires); err != nil {
			return nil, err
		}
		if inv.EndedAt, err = parseNull(ended); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func parseNull(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := store.ParseTime(s.String)
	return &t, err
}

// RevokeInvitationInput is the input of revoke_invitation.
type RevokeInvitationInput struct {
	Invitation int64 `json:"invitation" jsonschema:"the id of the invitation, which list_invitations gives"`
}

func (s *Service) revokeInvitation(ctx context.Context, c *operation.Call, in RevokeInvitationInput) (Invitation, error) {
	var roomName, name string
	var ended sql.NullString
	err := c.Tx.QueryRowContext(ctx, `
		SELECT r.name, w.name, i.ended_at FROM invitations i JOIN windows w ON w.id = i.window_id JOIN rooms r ON r.id = w.room_id
		WHERE i.id = ?`, in.Invitation).Scan(&roomName, &name, &ended)
	if errors.Is(err, sql.ErrNoRows) {
		return Invitation{}, operation.NotFoundf("no invitation %d: list_invitations lists those of a Window", in.Invitation)
	}
	if err != nil {
		return Invitation{}, err
	}
	c.Log(roomName+"/"+name, fmt.Sprintf("invitation %d", in.Invitation))
	r, err := room.Require(ctx, c, roomName, room.Member)
	if err != nil {
		return Invitation{}, err
	}
	c.InRoom(r.ID)
	if !c.Actor.Reaches(r.ID) {
		return Invitation{}, operation.NotFoundf("no invitation %d: list_invitations lists those of a Window", in.Invitation)
	}
	if ended.Valid {
		return Invitation{}, operation.Invalidf("invitation %d has ended already", in.Invitation)
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE invitations SET ended_at = ? WHERE id = ?", store.FormatTime(c.Now), in.Invitation); err != nil {
		return Invitation{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_passes WHERE invitation_id = ?", in.Invitation); err != nil {
		return Invitation{}, err
	}
	list, err := invitations(ctx, c.Tx, "i.id = ?", in.Invitation)
	if err != nil {
		return Invitation{}, err
	}
	return list[0], nil
}

// endKeyInvitations ends the invitations of the Window id that stand for its
// Key, when the Key changes or the access is no longer key.
func endKeyInvitations(ctx context.Context, c *operation.Call, id int64) error {
	_, err := c.Tx.ExecContext(ctx, "UPDATE invitations SET ended_at = ? WHERE window_id = ? AND hash IS NOT NULL AND ended_at IS NULL",
		store.FormatTime(c.Now), id)
	return err
}

// InvitationCodeInput is the input of get_invitation and open_invitation.
type InvitationCodeInput struct {
	Code string `json:"code" jsonschema:"the code of the personal link of the invitation"`
}

// InvitationPage is the output of get_invitation: what the page of an
// invitation shows.
type InvitationPage struct {
	Window     string           `json:"window" jsonschema:"the Window, as room/name"`
	Title      string           `json:"title"`
	By         string           `json:"by" jsonschema:"the name of the person who invites"`
	AI         string           `json:"ai,omitempty" jsonschema:"the AI that sent the invitation for them"`
	Note       string           `json:"note,omitempty"`
	NoteBlocks []mail.NoteBlock `json:"note_blocks,omitempty" jsonschema:"the note in paragraphs and lists, as the page shows it"`
	ExpiresAt  time.Time        `json:"expires_at"`
}

// open is an invitation that a personal link still opens.
type open struct {
	id, windowID int64
	page         InvitationPage
}

// errInvitationGone tells a visitor that a personal link no longer works.
var errInvitationGone = operation.NewRefusal("invitation_gone", operation.ErrNotFound,
	"this invitation does not open the Window any more: it was taken back, it expired, "+
		"or the Window changed its Key or its access. Ask the person who invited you for a new one")

// findOpen returns the invitation of code, while its personal link opens
// its Window.
func findOpen(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, code string, now time.Time) (open, error) {
	var o open
	var expires, access string
	var ends sql.NullString
	err := q.QueryRowContext(ctx, `
		SELECT i.id, i.window_id, r.name || '/' || w.name, w.title, coalesce(b.name, 'An administrator of the server'),
			coalesce(i.client, ''), i.note, i.expires_at, w.access, w.access_ends_at
		FROM invitations i JOIN windows w ON w.id = i.window_id JOIN rooms r ON r.id = w.room_id
		LEFT JOIN users b ON b.id = i.invited_by
		WHERE i.hash = ? AND i.ended_at IS NULL AND i.expires_at > ? AND w.trashed_at IS NULL`,
		credential.Hash(strings.TrimSpace(code)), store.FormatTime(now)).
		Scan(&o.id, &o.windowID, &o.page.Window, &o.page.Title, &o.page.By, &o.page.AI, &o.page.Note, &expires, &access, &ends)
	if errors.Is(err, sql.ErrNoRows) {
		return open{}, errInvitationGone
	}
	if err != nil {
		return open{}, err
	}
	if a, _, err := effective(access, ends, now); err != nil || a != AccessKey {
		return open{}, errors.Join(err, errInvitationGone)
	}
	o.page.ExpiresAt, err = store.ParseTime(expires)
	return o, err
}

func getInvitation(ctx context.Context, c *operation.Call, in InvitationCodeInput) (InvitationPage, error) {
	o, err := findOpen(ctx, c.Tx, in.Code, c.Now)
	o.page.NoteBlocks = mail.ParseNote(o.page.Note)
	return o.page, err
}

// InvitationOpening is the output of open_invitation: a code, valid a
// minute, that the console hands to /-/invitation?code=, which gives the
// pass.
type InvitationOpening struct {
	Code   string `json:"code"`
	Window string `json:"window" jsonschema:"the Window, as room/name"`
}

func openInvitation(ctx context.Context, c *operation.Call, in InvitationCodeInput) (InvitationOpening, error) {
	o, err := findOpen(ctx, c.Tx, in.Code, c.Now)
	if err != nil {
		return InvitationOpening{}, err
	}
	c.Log(o.page.Window, fmt.Sprintf("invitation %d opened", o.id))
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM invitation_openings WHERE expires_at <= ?", store.FormatTime(c.Now)); err != nil {
		return InvitationOpening{}, err
	}
	code := credential.Token("")
	_, err = c.Tx.ExecContext(ctx, "INSERT INTO invitation_openings (hash, invitation_id, expires_at) VALUES (?, ?, ?)",
		credential.Hash(code), o.id, store.FormatTime(c.Now.Add(openingTTL)))
	return InvitationOpening{Code: code, Window: o.page.Window}, err
}

// ErrOpeningGone means that the code that opens an invitation is unknown,
// used or expired.
var ErrOpeningGone = errors.New("the code of this invitation is unknown, used or expired")

// OpenInvitation takes the single-use code of open_invitation and gives a
// pass to the Window of its invitation, for /-/invitation. The pass lasts
// PassTTL, and never beyond the invitation or the access.
func OpenInvitation(ctx context.Context, db *store.DB, code string, now time.Time) (w Served, token string, expires time.Time, err error) {
	tx, err := db.Write.BeginTx(ctx, nil)
	if err != nil {
		return Served{}, "", time.Time{}, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	var id int64
	var invExpires string
	err = tx.QueryRowContext(ctx, `
		DELETE FROM invitation_openings WHERE hash = ? AND expires_at > ? RETURNING invitation_id`,
		credential.Hash(strings.TrimSpace(code)), store.FormatTime(now)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return Served{}, "", time.Time{}, ErrOpeningGone
	}
	if err != nil {
		return Served{}, "", time.Time{}, err
	}
	var windowID int64
	err = tx.QueryRowContext(ctx, "SELECT window_id, expires_at FROM invitations WHERE id = ? AND ended_at IS NULL AND expires_at > ?",
		id, store.FormatTime(now)).Scan(&windowID, &invExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return Served{}, "", time.Time{}, ErrOpeningGone
	}
	if err != nil {
		return Served{}, "", time.Time{}, err
	}
	w, err = LookupID(ctx, db, windowID, now)
	if err != nil {
		return Served{}, "", time.Time{}, err
	}
	if w.Access != AccessKey {
		return Served{}, "", time.Time{}, ErrOpeningGone
	}
	expires = now.Add(PassTTL)
	if t, err := store.ParseTime(invExpires); err == nil && t.Before(expires) {
		expires = t
	}
	if w.EndsAt != nil && w.EndsAt.Before(expires) {
		expires = *w.EndsAt
	}
	token = credential.Token(PassPrefix)
	_, err = tx.ExecContext(ctx, "INSERT INTO window_passes (window_id, hash, created_at, expires_at, invitation_id) VALUES (?, ?, ?, ?, ?)",
		windowID, credential.Hash(token), store.FormatTime(now), store.FormatTime(expires), id)
	if err != nil {
		return Served{}, "", time.Time{}, err
	}
	return w, token, expires, tx.Commit()
}
