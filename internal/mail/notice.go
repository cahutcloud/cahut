package mail

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Notices tell a person what happened that concerns them. Those about their
// account or about a Knock always go; the others go unless the person
// turned them off. See docs/adr/0031-e-mail-through-a-relay.md.

// Person is a person of a Cahut, as a notice reaches them.
type Person struct {
	Handle, Name, Email string
}

// Notice says whether a notice always goes, or only to those who did not
// turn notices off.
type Notice bool

// Kinds of notices.
const (
	Always   Notice = true  // about the account of the person, or a Knock
	Optional Notice = false // such as being added to a Room
)

// Prepare returns the letter that write makes for the person userID of the
// Cahut, as q reads them, addressed to them, or nil when nothing goes: the
// Cahut sends no e-mail, the person has no address or is disabled, or, for
// an optional notice, they turned notices off. A nil o prepares nothing.
func (o *Outbox) Prepare(ctx context.Context, q querier, userID int64, kind Notice, write func(p Person, name, where string) Letter) (*Letter, error) {
	if o == nil || userID == 0 {
		return nil, nil
	}
	var p Person
	var email sql.NullString
	var off bool
	err := q.QueryRowContext(ctx, "SELECT handle, name, email, notices_off FROM users WHERE id = ? AND disabled_at IS NULL", userID).
		Scan(&p.Handle, &p.Name, &email, &off)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !email.Valid || off && kind == Optional {
		return nil, nil
	}
	if ok, err := o.Service.Sends(ctx, q); err != nil || !ok {
		return nil, err
	}
	p.Email = email.String
	name, where, err := o.Where(ctx)
	if err != nil {
		return nil, err
	}
	l := write(p, name, where)
	l.To = p.Email
	return &l, nil
}

// Notify sends to the person userID, once the call c is committed, the
// letter that write makes, as Prepare decides.
func (o *Outbox) Notify(ctx context.Context, c *operation.Call, userID int64, kind Notice, write func(p Person, name, where string) Letter) error {
	l, err := o.Prepare(ctx, c.Tx, userID, kind, write)
	if err != nil || l == nil {
		return err
	}
	o.Post(c, *l)
	return nil
}

// Once reports whether the notice of kind about key may go now, at most
// once every every, or once for good when every is zero, and records that
// it goes. tx is in the database of the Cahut, which remembers the notices
// sent.
func Once(ctx context.Context, tx *sql.Tx, kind, key string, every time.Duration, now time.Time) (bool, error) {
	var sent string
	err := tx.QueryRowContext(ctx, "SELECT sent_at FROM notices_sent WHERE kind = ? AND key = ?", kind, key).Scan(&sent)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return false, err
	default:
		at, err := store.ParseTime(sent)
		if err != nil {
			return false, err
		}
		if every == 0 || now.Sub(at) < every {
			return false, nil
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO notices_sent (kind, key, sent_at) VALUES (?, ?, ?)
		ON CONFLICT (kind, key) DO UPDATE SET sent_at = excluded.sent_at`, kind, key, store.FormatTime(now))
	return err == nil, err
}
