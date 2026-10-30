package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// The moderators withdraw a Window that a report names, or a public
// Window, with a reason for the members of its Room. It keeps everything,
// and serves nothing to its readers until they reinstate it; a Window that
// the administrators of the server withdrew is reinstated by them only.
// Withdrawing decides the reports that wait about it. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.

// MaxReason is how many characters the reason of a withdrawal may hold.
const MaxReason = 500

// The refusals of a withdrawal that people meet in the console.
var (
	withdrawnByHost = operation.NewRefusalOf[WindowRef]("window_withdrawn_by_host", operation.ErrForbidden,
		"the administrators of the server withdrew {room}/{window}: only they change its reason or reinstate it")
	reasonMissing = operation.NewRefusalOf[ReasonLimit]("withdrawal_reason_invalid", operation.ErrInvalid,
		"give the reason of the withdrawal, in {limit} characters at most: the members of the Room of the Window read it")
)

// WindowRef fills the refusals that name a Window.
type WindowRef struct {
	Room   string `json:"room" jsonschema:"the Room of the Window"`
	Window string `json:"window" jsonschema:"the name of the Window"`
}

// ReasonLimit fills the refusal of a reason that is missing or too long.
type ReasonLimit struct {
	Limit int `json:"limit" jsonschema:"the most characters of a reason"`
}

// WithdrawInput is the input of withdraw_window.
type WithdrawInput struct {
	Window string `json:"window" jsonschema:"the Window, by its id or as room/name"`
	Reason string `json:"reason" jsonschema:"why it is withdrawn, 500 characters at most, which the members of its Room read"`
}

// ReinstateInput is the input of reinstate_window.
type ReinstateInput struct {
	Window string `json:"window" jsonschema:"the Window, by its id or as room/name"`
}

// Moderated is the output of withdraw_window and reinstate_window.
type Moderated struct {
	Window    Reported           `json:"window"`
	Withdrawn *window.Withdrawal `json:"withdrawn,omitempty" jsonschema:"when and why it is withdrawn; absent once it is reinstated"`
	Decided   []int64            `json:"decided,omitempty" jsonschema:"the reports that waited about it, which the withdrawal decided"`
}

// target is a Window as the moderators see it.
type target struct {
	id, roomID, author int64
	Reported
	withdrawal *window.Withdrawal
}

// locate returns the Window that ref names, by its id or as room/name, in
// the trash or not, and logs the call about it in its Room.
func locate(ctx context.Context, c *operation.Call, ref string) (target, error) {
	ref = strings.TrimSpace(ref)
	where, args := "w.uid = ?", []any{strings.ToLower(ref)}
	if !address.ValidID(strings.ToLower(ref)) {
		roomName, name, ok := strings.Cut(strings.Trim(ref, "/"), "/")
		if !ok || !address.ValidName(roomName) || !address.ValidName(name) {
			c.Log(ref, "")
			return target{}, operation.Invalidf("window %q: give the Window's id, or its Room and its name as room/name", ref)
		}
		where, args = "r.name = ? AND w.name = ?", []any{roomName, name}
	}
	var t target
	var withdrawn, reason sql.NullString
	var byServer bool
	err := c.Tx.QueryRowContext(ctx, `
		SELECT w.id, w.room_id, coalesce(w.created_by, 0), w.uid, r.name, w.name, w.title, w.access, w.trashed_at IS NOT NULL,
			w.withdrawn_at, w.withdrawn_reason, w.withdrawn_by_server
		FROM windows w JOIN rooms r ON r.id = w.room_id WHERE `+where, args...).
		Scan(&t.id, &t.roomID, &t.author, &t.ID, &t.Room, &t.Name, &t.Title, &t.Access, &t.Trashed, &withdrawn, &reason, &byServer)
	if errors.Is(err, sql.ErrNoRows) {
		c.Log(ref, "")
		return target{}, operation.NotFoundf("no Window %s in this Cahut", ref)
	}
	if err != nil {
		return target{}, err
	}
	if withdrawn.Valid {
		at, err := store.ParseTime(withdrawn.String)
		if err != nil {
			return target{}, err
		}
		t.withdrawal = &window.Withdrawal{At: at, Reason: reason.String, ByServer: byServer}
		t.Withdrawn = true
	}
	// The members of its Room read what the moderators did, or its author
	// alone, for a Window that is theirs alone.
	c.Log(t.Room+"/"+t.Name, "")
	c.InRoom(t.roomID)
	if t.Access == window.AccessAuthor {
		c.ReadBy(t.author)
	}
	return t, nil
}

// byHost reports whether the actor of c administers the server.
func byHost(c *operation.Call) bool {
	return c.Actor.Kind == operation.Admin
}

func withdraw(ctx context.Context, c *operation.Call, in WithdrawInput) (Moderated, error) {
	t, err := locate(ctx, c, in.Window)
	if err != nil {
		return Moderated{}, err
	}
	// One line, as the Logbook and the console show it.
	reason := strings.Join(strings.Fields(in.Reason), " ")
	c.Log(t.Room+"/"+t.Name, reason)
	if n := utf8.RuneCountInString(reason); n == 0 || n > MaxReason {
		return Moderated{}, reasonMissing(ReasonLimit{Limit: MaxReason})
	}
	ref := WindowRef{Room: t.Room, Window: t.Name}
	if t.withdrawal != nil && t.withdrawal.ByServer && !byHost(c) {
		return Moderated{}, withdrawnByHost(ref)
	}
	var waiting []int64
	rows, err := c.Tx.QueryContext(ctx, "SELECT id FROM reports WHERE window_id = ? AND decision IS NULL ORDER BY id", t.id)
	if err != nil {
		return Moderated{}, err
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return Moderated{}, err
		}
		waiting = append(waiting, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Moderated{}, err
	}
	switch {
	case t.Trashed:
		return Moderated{}, operation.Invalidf("%s/%s is in the trash: it serves nothing already", t.Room, t.Name)
	case t.withdrawal == nil && len(waiting) == 0 && t.Access != window.AccessPublic:
		return Moderated{}, operation.Forbiddenf("%s/%s is neither public nor named by a report that waits: "+
			"the moderators withdraw a Window that anyone opens, or that someone reported", t.Room, t.Name)
	}
	w := window.Withdrawal{At: c.Now, Reason: reason, ByServer: byHost(c)}
	if t.withdrawal != nil {
		// Withdrawn already: only the reason changes, and since when it
		// is withdrawn stays.
		w.At, w.ByServer = t.withdrawal.At, w.ByServer || t.withdrawal.ByServer
	}
	_, err = c.Tx.ExecContext(ctx, "UPDATE windows SET withdrawn_at = ?, withdrawn_reason = ?, withdrawn_by_server = ? WHERE id = ?",
		store.FormatTime(w.At), w.Reason, w.ByServer, t.id)
	if err != nil {
		return Moderated{}, err
	}
	for _, id := range waiting {
		if err := decide(ctx, c, id, Withdrawn, ""); err != nil {
			return Moderated{}, err
		}
	}
	detail := reason
	if len(waiting) > 0 {
		detail = fmt.Sprintf("reports %s: %s", joinIDs(waiting), reason)
	}
	c.Log(t.Room+"/"+t.Name, detail)
	t.Withdrawn = true
	return Moderated{Window: t.Reported, Withdrawn: &w, Decided: waiting}, nil
}

func reinstate(ctx context.Context, c *operation.Call, in ReinstateInput) (Moderated, error) {
	t, err := locate(ctx, c, in.Window)
	if err != nil {
		return Moderated{}, err
	}
	if t.withdrawal == nil {
		return Moderated{}, operation.Invalidf("%s/%s is not withdrawn", t.Room, t.Name)
	}
	c.Log(t.Room+"/"+t.Name, "withdrawn since "+t.withdrawal.At.Format("2006-01-02")+": "+t.withdrawal.Reason)
	if t.withdrawal.ByServer && !byHost(c) {
		return Moderated{}, withdrawnByHost(WindowRef{Room: t.Room, Window: t.Name})
	}
	_, err = c.Tx.ExecContext(ctx, "UPDATE windows SET withdrawn_at = NULL, withdrawn_reason = NULL, withdrawn_by_server = 0 WHERE id = ?", t.id)
	if err != nil {
		return Moderated{}, err
	}
	t.Withdrawn = false
	return Moderated{Window: t.Reported}, nil
}

// joinIDs writes the numbers of reports as 3, 4 and 7.
func joinIDs(ids []int64) string {
	words := make([]string, len(ids))
	for i, id := range ids {
		words[i] = fmt.Sprint(id)
	}
	if len(words) == 1 {
		return words[0]
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}
