package account

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

// A reader session lets a person open the Windows of their Cahut, where it
// lives in a cookie. The person signs in to the console, which reads no
// cookie and hands a single-use code to /-/session. See
// docs/adr/0019-window-versions-access-and-declaration.md and
// docs/adr/0020-console-under-a-path-of-port-443.md.

// ReaderPrefix starts every reader session token.
const ReaderPrefix = "cahut_rdr_"

// ReaderCodeTTL is how long the code of sign_in_to_windows lasts.
const ReaderCodeTTL = time.Minute

// SignInToWindowsInput is the input of sign_in_to_windows.
type SignInToWindowsInput struct{}

// ReaderCode is a single-use code that opens a reader session. Code is
// shown once and never stored.
type ReaderCode struct {
	Code      string    `json:"code" jsonschema:"send the browser to <domain>/-/session?code=<code>, on the port of the Windows"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SignInToWindows gives a person, signed in to the console, the code that
// signs them in to read Windows.
var SignInToWindows = &operation.Op[SignInToWindowsInput, ReaderCode]{
	Name:  "sign_in_to_windows",
	Title: "Sign in to read Windows",
	Description: "Get a single-use code, valid one minute, that signs you in to read the Windows of this Cahut " +
		"at its address: the browser takes it to /-/session.",
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, _ SignInToWindowsInput) (ReaderCode, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.Kind != operation.Person {
			return ReaderCode{}, operation.Invalidf("the administrators of the server read no Windows: " +
				"sign in as a person of this Cahut")
		}
		_, err := c.Tx.ExecContext(ctx, "DELETE FROM reader_codes WHERE expires_at <= ?", store.FormatTime(c.Now))
		if err != nil {
			return ReaderCode{}, err
		}
		rc := ReaderCode{Code: credential.Token(""), ExpiresAt: c.Now.Add(ReaderCodeTTL)}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO reader_codes (hash, user_id, expires_at) VALUES (?, ?, ?)",
			credential.Hash(rc.Code), c.Actor.UserID, store.FormatTime(rc.ExpiresAt))
		return rc, err
	},
}

// ReaderSession is a reader session just opened. Token is shown once, in a
// cookie, and never stored.
type ReaderSession struct {
	Token     string
	ExpiresAt time.Time
}

// OpenReaderSession uses the code of sign_in_to_windows, once, and opens a
// reader session for its person. It returns ErrInvalidSession when the code
// is unknown, used or expired, or its person is disabled. Like the exchanges
// of OAuth, it is not an operation: the code was logged when it was made.
func OpenReaderSession(ctx context.Context, db *store.DB, code string, now time.Time) (ReaderSession, error) {
	tx, err := db.Write.BeginTx(ctx, nil)
	if err != nil {
		return ReaderSession{}, err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit

	var userID int64
	err = tx.QueryRowContext(ctx, `
		DELETE FROM reader_codes WHERE hash = ? AND expires_at > ?
		AND user_id IN (SELECT id FROM users WHERE disabled_at IS NULL) RETURNING user_id`,
		credential.Hash(strings.TrimSpace(code)), store.FormatTime(now)).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ReaderSession{}, ErrInvalidSession
	}
	if err != nil {
		return ReaderSession{}, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= ? OR last_used_at <= ?",
		store.FormatTime(now), store.FormatTime(now.Add(-credential.SessionIdle)))
	if err != nil {
		return ReaderSession{}, err
	}
	s := ReaderSession{Token: credential.Token(ReaderPrefix), ExpiresAt: now.Add(credential.SessionMax)}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (user_id, hash, created_at, last_used_at, expires_at, reader) VALUES (?, ?, ?, ?, ?, 1)`,
		userID, credential.Hash(s.Token), store.FormatTime(now), store.FormatTime(now), store.FormatTime(s.ExpiresAt))
	if err != nil {
		return ReaderSession{}, err
	}
	return s, tx.Commit()
}

// VerifyReader checks a reader session token and returns who it belongs
// to. It records when the session was last used, at most once a minute.
func VerifyReader(ctx context.Context, db *store.DB, token string, now time.Time) (SessionOwner, error) {
	if !strings.HasPrefix(token, ReaderPrefix) {
		return SessionOwner{}, ErrInvalidSession
	}
	return verify(ctx, db, "s.hash = ? AND s.reader = 1", credential.Hash(token), now)
}

// Reader returns who the reader session id belongs to, while it lasts.
func Reader(ctx context.Context, db *store.DB, id int64, now time.Time) (SessionOwner, error) {
	return verify(ctx, db, "s.id = ? AND s.reader = 1", id, now)
}

// Console returns who the console session id belongs to, while it lasts,
// for the preview of a Window. It does not record a use: a preview that
// keeps loading, as a dashboard does, must not keep an idle console open.
// See docs/adr/0027-previews-of-windows-in-the-console.md.
func Console(ctx context.Context, db *store.DB, id int64, now time.Time) (SessionOwner, error) {
	return lookupSession(ctx, db, "s.id = ? AND s.reader = 0", id, now)
}

// EndReaderSession ends the reader session of token, if it is one.
func EndReaderSession(ctx context.Context, db *store.DB, token string) error {
	_, err := db.Write.ExecContext(ctx, "DELETE FROM sessions WHERE hash = ? AND reader = 1", credential.Hash(token))
	return err
}

// verify returns the owner of the session that where finds, with arg, if it
// lasts and its person is not disabled, and records its use.
func verify(ctx context.Context, db *store.DB, where string, arg any, now time.Time) (SessionOwner, error) {
	o, err := lookupSession(ctx, db, where, arg, now)
	if err != nil {
		return SessionOwner{}, err
	}
	_, err = db.Write.ExecContext(ctx, "UPDATE sessions SET last_used_at = ? WHERE id = ? AND last_used_at < ?",
		store.FormatTime(now), o.SessionID, store.FormatTime(now.Add(-time.Minute)))
	return o, err
}

// lookupSession returns the owner of the session that where finds, with
// arg, if it lasts and its person is not disabled.
func lookupSession(ctx context.Context, db *store.DB, where string, arg any, now time.Time) (SessionOwner, error) {
	var o SessionOwner
	err := db.Read.QueryRowContext(ctx, `
		SELECT s.id, u.id, u.handle FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE `+where+` AND s.expires_at > ? AND s.last_used_at > ? AND u.disabled_at IS NULL`,
		arg, store.FormatTime(now), store.FormatTime(now.Add(-credential.SessionIdle))).Scan(&o.SessionID, &o.UserID, &o.Handle)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionOwner{}, ErrInvalidSession
	}
	return o, err
}
