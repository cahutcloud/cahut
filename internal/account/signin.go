package account

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
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
	Emailed   bool      `json:"emailed,omitempty" jsonschema:"the link went to the e-mail address of the person too"`
}

// newActivation creates an activation link for the person userID, valid
// for ttl, which replaces any link they had.
func newActivation(ctx context.Context, c *operation.Call, userID int64, handle string, ttl time.Duration) (Activation, error) {
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM activations WHERE user_id = ?", userID); err != nil {
		return Activation{}, err
	}
	a := Activation{User: handle, Code: credential.Token(""), ExpiresAt: c.Now.Add(ttl)}
	_, err := c.Tx.ExecContext(ctx, "INSERT INTO activations (user_id, hash, created_at, expires_at) VALUES (?, ?, ?, ?)",
		userID, credential.Hash(a.Code), store.FormatTime(c.Now), store.FormatTime(a.ExpiresAt))
	return a, err
}

// ActivationLinkInput is the input of create_activation_link.
type ActivationLinkInput struct {
	User string `json:"user" jsonschema:"handle of the person"`
	// SendActivationEmail false keeps the link from the address: the caller
	// gives it to the person itself.
	SendActivationEmail *bool `json:"send_activation_email,omitempty" jsonschema:"false keeps the link from the e-mail address of the person, when you give it to them yourself; true when absent"`
}

// CreateActivationLink gives a person a new activation link, for instance
// after they forgot their password.
var CreateActivationLink = &operation.Op[ActivationLinkInput, Activation]{
	Name:  "create_activation_link",
	Title: "Give a person a new link",
	Description: "Give a person a new single-use link, valid 7 days, where they choose their password, " +
		"for instance after they forgot it. It goes to their e-mail address too, when they have one and the Cahut sends e-mail. " +
		"Their password works until they use the link.",
	Level:      operation.CahutAdmin,
	PeopleOnly: true,
	Services:   true,
	Run: func(ctx context.Context, c *operation.Call, in ActivationLinkInput) (Activation, error) {
		c.Log(in.User, "")
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return Activation{}, err
		}
		if p.Status == Disabled {
			return Activation{}, operation.Invalidf("%s is disabled: enable them first", p.Handle)
		}
		a, err := newActivation(ctx, c, p.ID, p.Handle, credential.ActivationTTL)
		if err != nil {
			return Activation{}, err
		}
		if sendsLink(in.SendActivationEmail) {
			a.Emailed, err = sendLink(ctx, c, p, a, activatePage, activationLetter)
		}
		return a, err
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
		store.FormatTime(c.Now), store.FormatTime(c.Now.Add(-credential.SessionIdle)))
	if err != nil {
		return Session{}, err
	}
	s := Session{Token: credential.Token(SessionPrefix), ExpiresAt: c.Now.Add(credential.SessionMax), User: p.User}
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
	Title:       "Activate an account",
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
		c.Identify(operation.Person, p.ID, p.Handle)
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
		if p.passwordHash != "" {
			// A link replaced the password the person had.
			if err := mail.OutboxOf(c).Notify(ctx, c, p.ID, mail.Always, passwordChangedLetter); err != nil {
				return Session{}, err
			}
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

// SignIn opens a session for a person who gives their handle and password.
// The password is checked before the call's transaction opens, so that a
// burst of sign-ins does not hold the Cahut's writes behind its hashes.
var SignIn = &operation.Op[SignInInput, Session]{
	Name:        "sign_in",
	Title:       "Sign in",
	Description: "Sign in with a handle and a password, and open a session.",
	PeopleOnly:  true,
	Visitors:    true,
	Prepare: func(ctx context.Context, c *operation.Call, in SignInInput) error {
		handle := strings.ToLower(strings.TrimSpace(in.Handle))
		var p person
		attempt := credential.SignIn{Password: in.Password}
		// A call without an address, from a test, counts as one address.
		address := sql.NullString{String: c.Actor.Address, Valid: c.Actor.Address != ""}
		err := c.Read(ctx, func(tx *sql.Tx) error {
			var err error
			p, err = findUser(ctx, tx, handle)
			if errors.Is(err, operation.ErrNotFound) {
				// An unknown handle is not logged: it may be a password
				// typed in the wrong field. Its refusals count at the
				// address, as those of a handle that exists do.
				return tx.QueryRowContext(ctx, `
					SELECT count(*) FROM logbook
					WHERE operation = 'sign_in' AND outcome = 'refused' AND user_id IS NULL AND address IS ? AND at > ?`,
					address, store.FormatTime(c.Now.Add(-credential.RefusalWindow))).Scan(&attempt.RefusedHere)
			}
			if err != nil {
				return err
			}
			attempt.Hash = p.passwordHash
			return tx.QueryRowContext(ctx, `
				SELECT count(*), count(*) FILTER (WHERE address IS ?) FROM logbook
				WHERE operation = 'sign_in' AND outcome = 'refused' AND user_id = ? AND at > ?`,
				address, p.ID, store.FormatTime(c.Now.Add(-credential.RefusalWindow))).Scan(&attempt.RefusedAll, &attempt.RefusedHere)
		})
		if err != nil {
			return err
		}
		if p.ID == 0 {
			c.Log("", "unknown handle")
			return credential.CheckSignIn(ctx, attempt)
		}
		c.Identify(operation.Person, p.ID, p.Handle)
		c.Log(p.Handle, "")
		if p.passwordHash == "" {
			c.Log(p.Handle, "account not activated")
		}
		if err := credential.CheckSignIn(ctx, attempt); err != nil {
			return err
		}
		c.Keep(p.passwordHash)
		return nil
	},
	Run: func(ctx context.Context, c *operation.Call, in SignInInput) (Session, error) {
		p, err := findUser(ctx, c.Tx, strings.ToLower(strings.TrimSpace(in.Handle)))
		if errors.Is(err, operation.ErrNotFound) {
			return Session{}, credential.ErrWrongPassword // deleted since its password was checked
		}
		if err != nil {
			return Session{}, err
		}
		if checked, _ := c.Kept().(string); checked == "" || checked != p.passwordHash {
			return Session{}, credential.ErrWrongPassword // its password changed since it was checked
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
	Title:       "Sign out",
	Description: "End the session you signed in with, and the sessions it opened to read Windows.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, _ SignOutInput) (SignedOut, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.Session == 0 {
			return SignedOut{}, operation.Invalidf("sign_out ends a session, and this call comes with none")
		}
		// The reader sessions that this session opened end with it: whoever
		// uses the browser next opens no Window as this person.
		_, err := c.Tx.ExecContext(ctx, "DELETE FROM sessions WHERE (id = ? OR opened_by = ? AND reader = 1) AND user_id = ?",
			c.Actor.Session, c.Actor.Session, c.Actor.UserID)
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
	Title:       "Change your password",
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
		if err := endSessions(ctx, c.Tx, p.ID, c.Actor.Session); err != nil {
			return PasswordChanged{}, err
		}
		return PasswordChanged{}, mail.OutboxOf(c).Notify(ctx, c, p.ID, mail.Always, passwordChangedLetter)
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
	return verify(ctx, db, "s.hash = ? AND s.reader = 0", credential.Hash(token), now)
}
