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
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/store"
)

// Share links open a Window to whoever holds them, without an account or a
// Key, until they end. A link adds to the access of its Window, which keeps
// it; a Window has several, each with its end, an optional number of
// openings and a label, and each is taken back alone. Its page opens
// nothing: a click on it opens the Window, and gives a pass, which opens it
// whatever its access until the link ends. See docs/adr/0052-share-links.md.

// LinkPrefix starts the path of every share link, /-/s/<token>.
const LinkPrefix = "/-/s/"

// Bounds of share links.
const (
	LinkTTL         = 7 * 24 * time.Hour  // of a link whose end is not given
	MaxLinkTTL      = 30 * 24 * time.Hour // of any link
	MaxLinks        = 20                  // links that stand on one Window
	MaxLinksPerDay  = 100                 // links that one person makes in 24 hours, their AIs included
	MaxLinkOpenings = 1000
	MaxLabel        = 100 // characters of a label
	// OnceTTL is how long the pass of a link that opens once lasts: a
	// session of reading, as long as a frame. See
	// docs/adr/0053-windows-for-their-author-alone.md.
	OnceTTL = FrameTTL
	// The clock of a browser, or the idea that an AI has of the time, may
	// be a little ahead of the server's.
	linkSkew = 10 * time.Minute
)

// The refusals of a share link that people meet in the console, which says
// them in their language.
var (
	linkToPublic = operation.NewRefusalOf[At]("share_link_public", operation.ErrInvalid,
		"{room}/{window} is public: its permanent address opens it to anyone, so share that address rather than a link")
	linkEndPast = operation.NewRefusalOf[LinkEnd]("share_link_end_past", operation.ErrInvalid,
		"ends_at {ends_at} is past, and a share link ends between now and {latest}: give an end in between, or none for 7 days")
	linkTooLong = operation.NewRefusalOf[LinkEnd]("share_link_too_long", operation.ErrInvalid,
		"ends_at {ends_at}: a share link lasts 30 days at most, until {latest}: give an end before it, or none for 7 days")
	linkOpenings = operation.NewRefusalOf[LinkOpenings]("share_link_openings", operation.ErrInvalid,
		"max_openings {given}: give a number from 1 to {limit}, or none for a link that opens to any number of readers")
	linkLabel = operation.NewRefusalOf[LinkLabel]("share_link_label", operation.ErrInvalid,
		"label: write it on one line, in {limit} characters at most")
	tooManyLinks = operation.NewRefusalOf[LinksOn]("too_many_share_links", operation.ErrForbidden,
		"{room}/{window} has {limit} share links that stand, the most: take one back before you make another")
	tooManyLinksToday = operation.NewRefusalOf[LinksMade]("too_many_share_links_today", operation.ErrTooMany,
		"{made} share links were made for {handle} in the last 24 hours, and {limit} is the most: make the rest later")
	linkEnded = operation.NewRefusalOf[LinkRef]("share_link_ended", operation.ErrInvalid,
		"share link {link} has ended already: it opens nothing more")
)

// errLinkGone tells a reader that a share link opens nothing, without
// telling why.
var errLinkGone = fmt.Errorf("%w: this link opens nothing: it ended, was taken back, or was opened as many times as it could. "+
	"Ask whoever gave it to you for a new one", operation.ErrNotFound)

// LinkEnd fills the refusal of an end that a share link cannot have.
type LinkEnd struct {
	EndsAt string `json:"ends_at" jsonschema:"the end given, in RFC 3339"`
	Latest string `json:"latest" jsonschema:"the latest end that a link made now may have, in RFC 3339"`
}

// LinkOpenings fills the refusal of a number of openings out of bounds.
type LinkOpenings struct {
	Given int `json:"given"`
	Limit int `json:"limit"`
}

// LinkLabel fills the refusal of a label that is too long or holds a line
// break.
type LinkLabel struct {
	Limit int `json:"limit" jsonschema:"the most characters of a label"`
}

// LinksOn fills the refusal of a link past the bound of a Window.
type LinksOn struct {
	Room   string `json:"room" jsonschema:"the Room of the Window"`
	Window string `json:"window" jsonschema:"the name of the Window"`
	Limit  int    `json:"limit" jsonschema:"the most links that stand on a Window"`
}

// LinksMade fills the refusal of links past the bound of a day.
type LinksMade struct {
	Made   int    `json:"made" jsonschema:"how many share links were made in the last 24 hours"`
	Handle string `json:"handle" jsonschema:"the person who makes them, or whom the AI acts for"`
	Limit  int    `json:"limit" jsonschema:"the most links of one person in 24 hours"`
}

// LinkRef fills the refusal of a share link that has ended.
type LinkRef struct {
	Link int64 `json:"link" jsonschema:"the id of the link"`
}

// ShareLinkInput is the input of create_share_link.
type ShareLinkInput struct {
	Window      string     `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	EndsAt      *time.Time `json:"ends_at,omitempty" jsonschema:"when the link stops opening the Window, within 30 days, such as 2026-11-05T18:00:00Z; 7 days from now when absent"`
	MaxOpenings int        `json:"max_openings,omitempty" jsonschema:"how many readers may open it, from 1 to 1000; any number when absent"`
	Label       string     `json:"label,omitempty" jsonschema:"a few words that tell the links of the Window apart, such as For Acme, 100 characters at most, which only the members of its Room see"`
}

// ShareLink is a share link of a Window, without its token.
type ShareLink struct {
	ID          int64      `json:"id"`
	Label       string     `json:"label,omitempty"`
	By          string     `json:"by,omitempty" jsonschema:"the handle of the person who made it"`
	AI          string     `json:"ai,omitempty" jsonschema:"the AI that made it for them"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at" jsonschema:"when it stops opening the Window, to those who opened it too"`
	MaxOpenings int        `json:"max_openings,omitempty" jsonschema:"how many readers may open it; absent when any number may"`
	Openings    int        `json:"openings" jsonschema:"how many readers opened it"`
	OpenedAt    *time.Time `json:"opened_at,omitempty" jsonschema:"when it was last opened"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty" jsonschema:"when it was taken back"`
	Stands      bool       `json:"stands" jsonschema:"whether it opens the Window to someone new: neither ended nor taken back, with openings left"`
}

// ShareLinkMade is the output of create_share_link.
type ShareLinkMade struct {
	ShareLink
	Window string `json:"window" jsonschema:"the Window, as room/name"`
	Link   string `json:"link" jsonschema:"the share link, returned this once: give it to the person, who passes it on. Whoever holds it opens the Window"`
}

// stands selects the share links l that open their Window to someone new.
const stands = "l.revoked_at IS NULL AND l.expires_at > ? AND (l.max_openings IS NULL OR l.openings < l.max_openings)"

func (s *Service) createLink(ctx context.Context, c *operation.Call, in ShareLinkInput) (ShareLinkMade, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return ShareLinkMade{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return ShareLinkMade{}, err
	}
	at := w.Room + "/" + w.Name
	if w.Access == AccessPublic {
		return ShareLinkMade{}, linkToPublic(At{Room: w.Room, Window: w.Name})
	}
	label, err := cleanLabel(in.Label)
	if err != nil {
		return ShareLinkMade{}, err
	}
	expires := c.Now.Add(LinkTTL)
	if in.EndsAt != nil {
		expires = in.EndsAt.UTC().Truncate(time.Millisecond)
		end := LinkEnd{EndsAt: expires.Format(time.RFC3339), Latest: c.Now.Add(MaxLinkTTL).Format(time.RFC3339)}
		switch {
		case !expires.After(c.Now):
			return ShareLinkMade{}, linkEndPast(end)
		case expires.After(c.Now.Add(MaxLinkTTL + linkSkew)):
			return ShareLinkMade{}, linkTooLong(end)
		}
	}
	if in.MaxOpenings < 0 || in.MaxOpenings > MaxLinkOpenings {
		return ShareLinkMade{}, linkOpenings(LinkOpenings{Given: in.MaxOpenings, Limit: MaxLinkOpenings})
	}
	if c.Actor.UserID != 0 {
		var made int
		err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM share_links WHERE created_by = ? AND created_at > ?",
			c.Actor.UserID, store.FormatTime(c.Now.Add(-24*time.Hour))).Scan(&made)
		if err != nil {
			return ShareLinkMade{}, err
		}
		if made >= MaxLinksPerDay {
			return ShareLinkMade{}, tooManyLinksToday(LinksMade{Made: made, Handle: c.Actor.Handle, Limit: MaxLinksPerDay})
		}
	}
	var standing int
	err = c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM share_links l WHERE l.window_id = ? AND "+stands,
		w.id, store.FormatTime(c.Now)).Scan(&standing)
	if err != nil {
		return ShareLinkMade{}, err
	}
	if standing >= MaxLinks {
		return ShareLinkMade{}, tooManyLinks(LinksOn{Room: w.Room, Window: w.Name, Limit: MaxLinks})
	}

	l := ShareLink{Label: label, By: c.Actor.Handle, CreatedAt: c.Now, ExpiresAt: expires, MaxOpenings: in.MaxOpenings, Stands: true}
	if c.Actor.Kind == operation.AI {
		l.AI = c.Actor.Client
	}
	token := credential.Token("")
	err = c.Tx.QueryRowContext(ctx, `
		INSERT INTO share_links (window_id, hash, label, created_by, client, created_at, expires_at, max_openings)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		w.id, credential.Hash(token), label, nullID(c.Actor.UserID), nullString(l.AI), store.FormatTime(c.Now), store.FormatTime(expires),
		sql.NullInt64{Int64: int64(in.MaxOpenings), Valid: in.MaxOpenings != 0}).Scan(&l.ID)
	if err != nil {
		return ShareLinkMade{}, err
	}
	// Never the token, nor the label, which may name people.
	detail := fmt.Sprintf("link %d, until %s", l.ID, expires.Format(time.RFC3339))
	if in.MaxOpenings != 0 {
		detail += fmt.Sprintf(", %d openings at most", in.MaxOpenings)
	}
	c.Log(at, detail)
	return ShareLinkMade{ShareLink: l, Window: at, Link: strings.TrimSuffix(s.base, "/") + LinkPrefix + token}, nil
}

// cleanLabel returns the label of a share link, trimmed, or an error when it
// is too long or holds control characters.
func cleanLabel(label string) (string, error) {
	label = strings.TrimSpace(label)
	bad := utf8.RuneCountInString(label) > MaxLabel
	for _, r := range label {
		bad = bad || unicode.IsControl(r) || r == utf8.RuneError
	}
	if bad {
		return "", linkLabel(LinkLabel{Limit: MaxLabel})
	}
	return label, nil
}

// linksOf returns the share links of the Window id, the latest first, and
// how many stand.
func linksOf(ctx context.Context, tx *sql.Tx, id int64, now time.Time) ([]ShareLink, int, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT l.id, l.label, coalesce(b.handle, ''), coalesce(l.client, ''), l.created_at, l.expires_at,
			coalesce(l.max_openings, 0), l.openings, l.opened_at, l.revoked_at, `+stands+`
		FROM share_links l LEFT JOIN users b ON b.id = l.created_by
		WHERE l.window_id = ? ORDER BY l.id DESC LIMIT 100`, store.FormatTime(now), id)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list, standing := []ShareLink{}, 0
	for rows.Next() {
		var l ShareLink
		var created, expires string
		var opened, revoked sql.NullString
		if err := rows.Scan(&l.ID, &l.Label, &l.By, &l.AI, &created, &expires, &l.MaxOpenings, &l.Openings, &opened, &revoked, &l.Stands); err != nil {
			return nil, 0, err
		}
		if l.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, 0, err
		}
		if l.ExpiresAt, err = store.ParseTime(expires); err != nil {
			return nil, 0, err
		}
		if l.OpenedAt, err = parseNull(opened); err != nil {
			return nil, 0, err
		}
		if l.RevokedAt, err = parseNull(revoked); err != nil {
			return nil, 0, err
		}
		if l.Stands {
			standing++
		}
		list = append(list, l)
	}
	return list, standing, rows.Err()
}

// RevokeShareLinkInput is the input of revoke_share_link.
type RevokeShareLinkInput struct {
	Link int64 `json:"link" jsonschema:"the id of the share link, which get_window gives"`
}

func (s *Service) revokeLink(ctx context.Context, c *operation.Call, in RevokeShareLinkInput) (ShareLink, error) {
	var roomName, name, expires, access string
	var windowID, author int64
	var revoked sql.NullString
	err := c.Tx.QueryRowContext(ctx, `
		SELECT r.name, w.name, w.id, w.access, coalesce(w.created_by, 0), l.expires_at, l.revoked_at
		FROM share_links l JOIN windows w ON w.id = l.window_id JOIN rooms r ON r.id = w.room_id
		WHERE l.id = ?`, in.Link).Scan(&roomName, &name, &windowID, &access, &author, &expires, &revoked)
	gone := operation.NotFoundf("no share link %d: get_window lists those of a Window", in.Link)
	if errors.Is(err, sql.ErrNoRows) {
		return ShareLink{}, gone
	}
	if err != nil {
		return ShareLink{}, err
	}
	c.Log(roomName+"/"+name, fmt.Sprintf("link %d", in.Link))
	r, err := room.Require(ctx, c, roomName, room.Member)
	if err != nil {
		return ShareLink{}, err
	}
	c.InRoom(r.ID)
	if access == AccessAuthor {
		c.ReadBy(author)
	}
	// A Window that its author keeps for themself is theirs to manage.
	if !c.Actor.Reaches(r.ID) || access == AccessAuthor && author != c.Actor.UserID && !c.Actor.CommandLine() {
		return ShareLink{}, gone
	}
	if revoked.Valid || expires <= store.FormatTime(c.Now) {
		return ShareLink{}, linkEnded(LinkRef(in))
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE share_links SET revoked_at = ? WHERE id = ?", store.FormatTime(c.Now), in.Link); err != nil {
		return ShareLink{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_passes WHERE link_id = ?", in.Link); err != nil {
		return ShareLink{}, err
	}
	list, _, err := linksOf(ctx, c.Tx, windowID, c.Now)
	if err != nil {
		return ShareLink{}, err
	}
	for _, l := range list {
		if l.ID == in.Link {
			return l, nil
		}
	}
	return ShareLink{}, gone
}

// OpenShareLinkInput is the input of open_share_link.
type OpenShareLinkInput struct {
	Token string `json:"token" jsonschema:"the token of the share link, the last part of its address"`
}

// ShareLinkOpening is the output of open_share_link: the Window that the
// link opens, and the pass that opens it, which only the page of the link
// hands over, in a cookie.
type ShareLinkOpening struct {
	Window    string    `json:"window" jsonschema:"the Window, as room/name"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the pass ends: with the link, or 4 hours after the opening of a link that opens once"`
	WindowID  int64     `json:"-"`
	Pass      string    `json:"-"`
	Once      bool      `json:"-"` // the link opens once: its pass is for a session of reading
}

// openLink spends an opening of a share link, and gives its reader a pass
// that lasts as long as the link, or a session of reading for a link that
// opens once. The Logbook records the opening in the Room of the Window, or
// for its author alone, without the reader, signed in or not, their address
// or their browser.
func openLink(ctx context.Context, c *operation.Call, in OpenShareLinkInput) (ShareLinkOpening, error) {
	c.Anonymize()
	var o ShareLinkOpening
	var id, roomID, author int64
	var expires, access string
	var limit, openings int
	err := c.Tx.QueryRowContext(ctx, `
		SELECT l.id, w.id, w.room_id, r.name || '/' || w.name, w.access, coalesce(w.created_by, 0), l.expires_at, coalesce(l.max_openings, 0), l.openings
		FROM share_links l JOIN windows w ON w.id = l.window_id JOIN rooms r ON r.id = w.room_id
		WHERE l.hash = ? AND w.trashed_at IS NULL AND `+stands,
		credential.Hash(strings.TrimSpace(in.Token)), store.FormatTime(c.Now)).
		Scan(&id, &o.WindowID, &roomID, &o.Window, &access, &author, &expires, &limit, &openings)
	if errors.Is(err, sql.ErrNoRows) {
		return ShareLinkOpening{}, errLinkGone
	}
	if err != nil {
		return ShareLinkOpening{}, err
	}
	c.InRoom(roomID)
	if access == AccessAuthor {
		c.ReadBy(author)
	}
	if o.ExpiresAt, err = store.ParseTime(expires); err != nil {
		return ShareLinkOpening{}, err
	}
	o.Once = limit == 1
	if o.Once && c.Now.Add(OnceTTL).Before(o.ExpiresAt) {
		o.ExpiresAt = c.Now.Add(OnceTTL)
	}
	openings++
	detail := fmt.Sprintf("link %d, opening %d", id, openings)
	if limit != 0 {
		detail += fmt.Sprintf(" of %d", limit)
	}
	c.Log(o.Window, detail)
	now := store.FormatTime(c.Now)
	if _, err := c.Tx.ExecContext(ctx, "UPDATE share_links SET openings = ?, opened_at = ? WHERE id = ?", openings, now, id); err != nil {
		return ShareLinkOpening{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_passes WHERE expires_at <= ?", now); err != nil {
		return ShareLinkOpening{}, err
	}
	o.Pass = credential.Token(PassPrefix)
	_, err = c.Tx.ExecContext(ctx, "INSERT INTO window_passes (window_id, hash, created_at, expires_at, link_id) VALUES (?, ?, ?, ?, ?)",
		o.WindowID, credential.Hash(o.Pass), now, store.FormatTime(o.ExpiresAt), id)
	return o, err
}

// Link is a share link as its page shows it: the Window it opens, until
// when, and how many openings it has left.
type Link struct {
	Window    Served
	ExpiresAt time.Time
	Left      int  // openings left, when it has a number of them
	Limited   bool // it has a number of openings
	Once      bool // it opens once, and its pass lasts a session of reading
	Stands    bool // it opens the Window to someone new
}

// ErrNoLink means that a token names no share link, or the Window of the
// link is in the trash.
var ErrNoLink = errors.New("no share link")

// FindLink returns the share link whose token is token, with its Window,
// whether it stands or not, for its page. Visits are not logged.
func FindLink(ctx context.Context, db *store.DB, token string, now time.Time) (Link, error) {
	var l Link
	var windowID int64
	var expires string
	var limit, openings int
	err := db.Read.QueryRowContext(ctx, `
		SELECT l.window_id, l.expires_at, coalesce(l.max_openings, 0), l.openings, `+stands+`
		FROM share_links l WHERE l.hash = ?`,
		store.FormatTime(now), credential.Hash(strings.TrimSpace(token))).Scan(&windowID, &expires, &limit, &openings, &l.Stands)
	if errors.Is(err, sql.ErrNoRows) {
		return Link{}, ErrNoLink
	}
	if err != nil {
		return Link{}, err
	}
	if l.ExpiresAt, err = store.ParseTime(expires); err != nil {
		return Link{}, err
	}
	l.Limited, l.Left, l.Once = limit != 0, limit-openings, limit == 1
	l.Window, err = LookupID(ctx, db, windowID, now)
	if errors.Is(err, operation.ErrNotFound) {
		return Link{}, ErrNoLink
	}
	return l, err
}
