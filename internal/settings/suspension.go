package settings

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// The administrators of the server suspend a Cahut, after a decision of
// moderation or when its organization stops paying, and resume it later. A
// suspended Cahut keeps everything, and serves nothing to anyone but them:
// the registry refuses every other call, and the web server answers the
// readers of its Windows with a page of its own. The state is kept with
// the settings of the Cahut, so that it travels with it. See
// docs/adr/0048-suspending-a-cahut.md.

// Suspension says since when, and why, the administrators of the server
// suspended the Cahut.
type Suspension struct {
	Since  time.Time `json:"since" jsonschema:"when the administrators of the server suspended the Cahut"`
	Reason string    `json:"reason" jsonschema:"why the administrators of the server suspended the Cahut; its administrators read it, and readers, people and AIs never do"`
}

// MaxReason is how many characters the reason of a suspension may hold.
const MaxReason = 500

// ErrSuspended refuses every call in a suspended Cahut, but those of the
// administrators of its server.
var ErrSuspended = operation.NewRefusal("cahut_suspended", operation.ErrForbidden,
	"this Cahut is suspended by the administrators of its server: it keeps everything it holds, "+
		"and nothing in it can be read or changed until they resume it. Ask the administrators of the Cahut, "+
		"or, for an AI, tell the person you act for")

// SuspensionOf returns the suspension of the Cahut, or nil while it is
// active.
func SuspensionOf(ctx context.Context, q querier) (*Suspension, error) {
	var since, reason sql.NullString
	err := q.QueryRowContext(ctx, "SELECT suspended_at, suspension_reason FROM settings WHERE id = 1").Scan(&since, &reason)
	if err != nil || !since.Valid {
		return nil, err
	}
	t, err := store.ParseTime(since.String)
	if err != nil {
		return nil, err
	}
	return &Suspension{Since: t, Reason: reason.String}, nil
}

// CheckActive returns ErrSuspended when the Cahut is suspended. The
// registry calls it in the transaction of every call that the
// administrators of the server do not make.
func CheckActive(ctx context.Context, tx *sql.Tx) error {
	s, err := SuspensionOf(ctx, tx)
	if err != nil {
		return err
	}
	if s != nil {
		return ErrSuspended
	}
	return nil
}

// SuspendCahutInput is the input of suspend_cahut.
type SuspendCahutInput struct {
	Reason string `json:"reason" jsonschema:"why the Cahut is suspended, in 500 characters at most, such as an unpaid invoice or a decision of moderation; its administrators read it, and readers, people and AIs never do"`
}

// Suspend suspends the Cahut, for the administrators of the server.
var Suspend = &operation.Op[SuspendCahutInput, Settings]{
	Name:  "suspend_cahut",
	Title: "Suspend the Cahut",
	Description: "Suspend this Cahut, with a reason, after a decision of moderation or when its organization stops paying. " +
		"It keeps everything: its people and their sessions, their AIs, its Rooms, its Windows and its Logbook. " +
		"Until resume_cahut, it serves nothing to anyone but the administrators of the server: its Windows answer with a page " +
		"that says it is suspended, and every call of a person, an AI or a service is refused. " +
		"Its administrators receive the reason by e-mail when it sends e-mail. Suspending a suspended Cahut changes its reason only. " +
		"For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SuspendCahutInput) (Settings, error) {
		// One line, as the Logbook and the list of the Cahuts show it.
		reason := strings.Join(strings.Fields(in.Reason), " ")
		c.Log("cahut", reason)
		switch n := utf8.RuneCountInString(reason); {
		case n == 0:
			return Settings{}, operation.Invalidf("give the reason of the suspension: its administrators read it")
		case n > MaxReason:
			return Settings{}, operation.Invalidf("the reason holds %d characters: keep it to %d", n, MaxReason)
		}
		was, err := SuspensionOf(ctx, c.Tx)
		if err != nil {
			return Settings{}, err
		}
		if was != nil {
			// The Cahut stays suspended since then: only the reason changes,
			// without a new e-mail.
			if _, err := c.Tx.ExecContext(ctx, "UPDATE settings SET suspension_reason = ? WHERE id = 1", reason); err != nil {
				return Settings{}, err
			}
			c.Log("cahut", "new reason: "+reason)
			return Read(ctx, c.Tx, c.Actor.UserID, c.Admin)
		}
		_, err = c.Tx.ExecContext(ctx, "UPDATE settings SET suspended_at = ?, suspension_reason = ? WHERE id = 1",
			store.FormatTime(c.Now), reason)
		if err != nil {
			return Settings{}, err
		}
		support, err := supportPage(ctx, c.Tx)
		if err != nil {
			return Settings{}, err
		}
		err = tellAdmins(ctx, c, func(p mail.Person, name, where string) mail.Letter {
			l := mail.Letter{
				Subject: name + " is suspended",
				Lines: []string{
					"Hello " + p.Name + ",",
					"The administrators of the server suspended " + where + " on " + mail.Day(c.Now) + ". " +
						"It keeps everything it holds: its people, their AIs, its Rooms, its Windows and its Logbook.",
					"Until they resume it, its Windows say that it is suspended, people cannot sign in, and AIs cannot act in it. " +
						"Its people and their AIs find everything as it was once it is resumed.",
				},
				Quote: &mail.Quote{By: "The administrators of the server", Text: reason},
				After: []string{"To have it resumed, ask the administrators of the server."},
			}
			if support != "" {
				l.Action = &mail.Action{Label: "Get support", URL: support}
			}
			return l
		})
		if err != nil {
			return Settings{}, err
		}
		return Read(ctx, c.Tx, c.Actor.UserID, c.Admin)
	},
}

// ResumeCahutInput is the input of resume_cahut.
type ResumeCahutInput struct{}

// Resume resumes a suspended Cahut, for the administrators of the server.
var Resume = &operation.Op[ResumeCahutInput, Settings]{
	Name:  "resume_cahut",
	Title: "Resume the Cahut",
	Description: "Resume this Cahut, which suspend_cahut suspended: its Windows open again, its people sign in, " +
		"and their AIs act in it as before, without connecting again. Its administrators hear of it by e-mail when it sends e-mail. " +
		"For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, _ ResumeCahutInput) (Settings, error) {
		c.Log("cahut", "")
		was, err := SuspensionOf(ctx, c.Tx)
		if err != nil {
			return Settings{}, err
		}
		if was == nil {
			return Settings{}, operation.Invalidf("this Cahut is not suspended")
		}
		_, err = c.Tx.ExecContext(ctx, "UPDATE settings SET suspended_at = NULL, suspension_reason = NULL WHERE id = 1")
		if err != nil {
			return Settings{}, err
		}
		c.Log("cahut", "suspended since "+mail.Day(was.Since)+": "+was.Reason)
		err = tellAdmins(ctx, c, func(p mail.Person, name, where string) mail.Letter {
			o := mail.OutboxOf(c)
			return mail.Letter{
				Subject: name + " is open again",
				Lines: []string{
					"Hello " + p.Name + ",",
					"The administrators of the server resumed " + where + ", which they had suspended on " + mail.Day(was.Since) + ". " +
						"Its Windows open again, its people sign in, and their AIs act in it as before.",
				},
				Action: &mail.Action{Label: "Open the console", URL: o.Link("")},
			}
		})
		if err != nil {
			return Settings{}, err
		}
		return Read(ctx, c.Tx, c.Actor.UserID, c.Admin)
	},
}

// supportPage returns the address of the support page of the Cahut, which
// its host gives, or "" when it has none.
func supportPage(ctx context.Context, tx *sql.Tx) (string, error) {
	legal, err := Legal(ctx, tx)
	return legal.Support, err
}

// tellAdmins sends the letter that write makes to each administrator of
// the Cahut who is not disabled and has an address, once the call c is
// committed, when the Cahut sends e-mail.
func tellAdmins(ctx context.Context, c *operation.Call, write func(p mail.Person, name, where string) mail.Letter) error {
	o := mail.OutboxOf(c)
	if o == nil {
		return nil
	}
	rows, err := c.Tx.QueryContext(ctx, "SELECT id FROM users WHERE role = 'admin' AND disabled_at IS NULL ORDER BY id")
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	for _, id := range ids {
		l, err := o.Prepare(ctx, c.Tx, id, mail.Always, write)
		if err != nil {
			return err
		}
		if l != nil {
			o.Post(c, *l)
		}
	}
	return nil
}
