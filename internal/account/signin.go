package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// How long links and sessions last, and how many refused sign-ins a person
// takes before they must wait.
const (
	ActivationTTL = 7 * 24 * time.Hour
	SessionIdle   = 7 * 24 * time.Hour  // a session ends this long after its last use
	SessionMax    = 30 * 24 * time.Hour // and this long after it opened in any case

	maxRefusals   = 10
	refusalWindow = 15 * time.Minute
)

// SessionPrefix starts every session token, so that secret scanners and
// people can recognize one.
const SessionPrefix = "cahut_ses_"

// Activation is a single-use link just created, where a person chooses
// their password. Code is shown once and never stored.
type Activation struct {
	User      string    `json:"user" jsonschema:"handle of the person"`
	Code      string    `json:"code" jsonschema:"the person opens the console of the Cahut at /activate#<code> to choose their password"`
	ExpiresAt time.Time `json:"expires_at"`
}

// newActivation creates an activation link for the person userID, which
// replaces any link they had.
func newActivation(ctx context.Context, c *operation.Call, userID int64, handle string) (Activation, error) {
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM activations WHERE user_id = ?", userID); err != nil {
		return Activation{}, err
	}
	a := Activation{User: handle, Code: credential.Token(""), ExpiresAt: c.Now.Add(ActivationTTL)}
	_, err := c.Tx.ExecContext(ctx, "INSERT INTO activations (user_id, hash, created_at, expires_at) VALUES (?, ?, ?, ?)",
		userID, credential.Hash(a.Code), store.FormatTime(c.Now), store.FormatTime(a.ExpiresAt))
	return a, err
}

// CreateActivationLink gives a person a new activation link, for instance
// after they forgot their password.
var CreateActivationLink = &operation.Op[UserInput, Activation]{
	Name: "create_activation_link",
	Description: "Give a person a new single-use link, valid 7 days, where they choose their password, " +
		"for instance after they forgot it. Their password works until they use the link.",
	Level:      operation.CahutAdmin,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in UserInput) (Activation, error) {
		c.Log(in.User, "")
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return Activation{}, err
		}
		if p.Status == Disabled {
			return Activation{}, operation.Invalidf("%s is disabled: enable them first", p.Handle)
		}
		return newActivation(ctx, c, p.ID, p.Handle)
	},
}

// Session is a session just opened. Token is shown once and never stored.
type Session struct {
	Token     string    `json:"token" jsonschema:"send it in the Authorization header: Bearer <token>"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the session ends at the latest; it ends sooner after 7 days without use"`
	User      User      `json:"user"`
}

// openSession opens a session for the person p, and deletes the sessions
// of the Cahut that have ended.
func openSession(ctx context.Context, c *operation.Call, p person) (Session, error) {
	_, err := c.Tx.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= ? OR last_used_at <= ?",
		store.FormatTime(c.Now), store.FormatTime(c.Now.Add(-SessionIdle)))
	if err != nil {
		return Session{}, err
	}
	s := Session{Token: credential.Token(SessionPrefix), ExpiresAt: c.Now.Add(SessionMax), User: p.User}
	now := store.FormatTime(c.Now)
	_, err = c.Tx.ExecContext(ctx, "INSERT INTO sessions (user_id, hash, created_at, last_used_at, expires_at) VALUES (?, ?, ?, ?, ?)",
		p.ID, credential.Hash(s.Token), now, now, store.FormatTime(s.ExpiresAt))
	return s, err
}

// endSessions ends the sessions of the person userID, but the session keep.
func endSessions(ctx context.Context, tx *sql.Tx, userID, keep int64) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ? AND id != ?", userID, keep)
	return err
}

// ActivateInput is the input of activate_account.
type ActivateInput struct {
	Code     string `json:"code" jsonschema:"the code of the activation link"`
	Password string `json:"password" jsonschema:"the new password, of 15 characters at least"`
}

// ActivateAccount sets a person's password from an activation link, and
// signs them in. It ends their other sessions.
var ActivateAccount = &operation.Op[ActivateInput, Session]{
	Name:        "activate_account",
	Description: "Choose a password from a single-use activation link, and sign in.",
	PeopleOnly:  true,
	Visitors:    true,
	Run: func(ctx context.Context, c *operation.Call, in ActivateInput) (Session, error) {
		if err := credential.CheckPolicy(in.Password); err != nil {
			return Session{}, err
		}
		var handle string
		err := c.Tx.QueryRowContext(ctx, `
			SELECT u.handle FROM activations a JOIN users u ON u.id = a.user_id
			WHERE a.hash = ? AND a.expires_at > ?`, credential.Hash(strings.TrimSpace(in.Code)), store.FormatTime(c.Now)).Scan(&handle)
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, operation.Forbiddenf("this activation link is unknown, used or expired: " +
				"ask an administrator of the Cahut for a new one")
		}
		if err != nil {
			return Session{}, err
		}
		p, err := findUser(ctx, c.Tx, handle)
		if err != nil {
			return Session{}, err
		}
		c.Identify(p.ID, p.Handle)
		c.Log(p.Handle, "")
		if p.Status == Disabled {
			return Session{}, operation.Forbiddenf("this account is disabled: an administrator of the Cahut can enable it")
		}
		hash, err := credential.HashPassword(ctx, in.Password)
		if err != nil {
			return Session{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET password_hash = ? WHERE id = ?", hash, p.ID); err != nil {
			return Session{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM activations WHERE user_id = ?", p.ID); err != nil {
			return Session{}, err
		}
		if err := endSessions(ctx, c.Tx, p.ID, 0); err != nil {
			return Session{}, err
		}
		p.Status = Active
		return openSession(ctx, c, p)
	},
}

// SignInInput is the input of sign_in.
type SignInInput struct {
	Handle   string `json:"handle"`
	Password string `json:"password"`
}

// errWrongPassword answers every failed sign-in alike, so that it does not
// tell whether a handle exists.
var errWrongPassword = operation.Forbiddenf("wrong handle or password")

// SignIn opens a session for a person who gives their handle and password.
var SignIn = &operation.Op[SignInInput, Session]{
	Name:        "sign_in",
	Description: "Sign in with a handle and a password, and open a session.",
	PeopleOnly:  true,
	Visitors:    true,
	Run: func(ctx context.Context, c *operation.Call, in SignInInput) (Session, error) {
		// refuse spends the time of a password check before it answers, so
		// that response times tell nothing either.
		refuse := func(refusal error) (Session, error) {
			if err := credential.CheckNoPassword(ctx, in.Password); err != nil {
				return Session{}, err
			}
			return Session{}, refusal
		}
		// An unknown handle is not logged: it may be a password typed in the
		// wrong field.
		handle := strings.ToLower(strings.TrimSpace(in.Handle))
		p, err := findUser(ctx, c.Tx, handle)
		if errors.Is(err, operation.ErrNotFound) {
			c.Log("", "unknown handle")
			return refuse(errWrongPassword)
		}
		if err != nil {
			return Session{}, err
		}
		c.Identify(p.ID, p.Handle)
		c.Log(p.Handle, "")

		var refusals int
		err = c.Tx.QueryRowContext(ctx, `
			SELECT count(*) FROM logbook
			WHERE operation = 'sign_in' AND outcome = 'refused' AND user_id = ? AND at > ?`,
			p.ID, store.FormatTime(c.Now.Add(-refusalWindow))).Scan(&refusals)
		if err != nil {
			return Session{}, err
		}
		if refusals >= maxRefusals {
			return refuse(fmt.Errorf("%w: too many refused sign-ins for this handle: try again in 15 minutes", operation.ErrTooMany))
		}
		if p.passwordHash == "" {
			c.Log(p.Handle, "account not activated")
			return refuse(errWrongPassword)
		}
		ok, err := credential.CheckPassword(ctx, p.passwordHash, in.Password)
		if err != nil {
			return Session{}, err
		}
		if !ok {
			return Session{}, errWrongPassword
		}
		if p.Status == Disabled {
			return Session{}, operation.Forbiddenf("this account is disabled: an administrator of the Cahut can enable it")
		}
		return openSession(ctx, c, p)
	},
}

// SignOutInput is the input of sign_out.
type SignOutInput struct{}

// SignedOut is the output of sign_out.
type SignedOut struct{}

// SignOut ends the session of the call.
var SignOut = &operation.Op[SignOutInput, SignedOut]{
	Name:        "sign_out",
	Description: "End the session you signed in with.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, _ SignOutInput) (SignedOut, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.Session == 0 {
			return SignedOut{}, operation.Invalidf("sign_out ends a session, and this call comes with none")
		}
		_, err := c.Tx.ExecContext(ctx, "DELETE FROM sessions WHERE id = ? AND user_id = ?", c.Actor.Session, c.Actor.UserID)
		return SignedOut{}, err
	},
}

// ChangePasswordInput is the input of change_password.
type ChangePasswordInput struct {
	Password    string `json:"password" jsonschema:"the current password"`
	NewPassword string `json:"new_password" jsonschema:"the new password, of 15 characters at least"`
}

// PasswordChanged is the output of change_password.
type PasswordChanged struct{}

// ChangePassword changes the password of the person who calls, and ends
// their other sessions.
var ChangePassword = &operation.Op[ChangePasswordInput, PasswordChanged]{
	Name:        "change_password",
	Description: "Change your password. Your other sessions end.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in ChangePasswordInput) (PasswordChanged, error) {
		if _, err := self(c, "change_password"); err != nil {
			return PasswordChanged{}, err
		}
		c.Log(c.Actor.Handle, "")
		if err := credential.CheckPolicy(in.NewPassword); err != nil {
			return PasswordChanged{}, err
		}
		p, err := findUser(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return PasswordChanged{}, err
		}
		if p.passwordHash == "" {
			return PasswordChanged{}, operation.Invalidf("%s has no password yet: use an activation link", p.Handle)
		}
		ok, err := credential.CheckPassword(ctx, p.passwordHash, in.Password)
		if err != nil {
			return PasswordChanged{}, err
		}
		if !ok {
			return PasswordChanged{}, operation.Forbiddenf("the current password is wrong")
		}
		hash, err := credential.HashPassword(ctx, in.NewPassword)
		if err != nil {
			return PasswordChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET password_hash = ? WHERE id = ?", hash, p.ID); err != nil {
			return PasswordChanged{}, err
		}
		return PasswordChanged{}, endSessions(ctx, c.Tx, p.ID, c.Actor.Session)
	},
}

// SessionOwner is who a valid session belongs to.
type SessionOwner struct {
	SessionID int64
	UserID    int64
	Handle    string
}

// ErrInvalidSession means the session is unknown or has ended.
var ErrInvalidSession = errors.New("invalid or ended session")

// VerifySession checks a session token and returns who it belongs to. It
// records when the session was last used, at most once a minute.
func VerifySession(ctx context.Context, db *store.DB, token string, now time.Time) (SessionOwner, error) {
	if !strings.HasPrefix(token, SessionPrefix) {
		return SessionOwner{}, ErrInvalidSession
	}
	var o SessionOwner
	err := db.Read.QueryRowContext(ctx, `
		SELECT s.id, u.id, u.handle FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.hash = ? AND s.expires_at > ? AND s.last_used_at > ? AND u.disabled_at IS NULL`,
		credential.Hash(token), store.FormatTime(now), store.FormatTime(now.Add(-SessionIdle))).Scan(&o.SessionID, &o.UserID, &o.Handle)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionOwner{}, ErrInvalidSession
	}
	if err != nil {
		return SessionOwner{}, err
	}
	_, err = db.Write.ExecContext(ctx, "UPDATE sessions SET last_used_at = ? WHERE id = ? AND last_used_at < ?",
		store.FormatTime(now), o.SessionID, store.FormatTime(now.Add(-time.Minute)))
	return o, err
}
