package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// The administrators of the server have accounts of their own, apart from
// the people of the Cahuts. They sign in at the console of any Cahut, and
// each session holds at the domain where it opened: in a Cahut, an
// administrator acts at the domain of that Cahut only.
// See docs/adr/0013-people-sessions-and-rights.md.

// SessionPrefix starts every session token of an administrator of the
// server, so that the console's API knows where to check it.
const SessionPrefix = "cahut_adm_"

// registerAdmins adds the operations on the administrators of the server.
func registerAdmins(r *operation.Registry) {
	operation.Register(r, CreateAdmin)
	operation.Register(r, ListAdmins)
	operation.Register(r, DisableAdmin)
	operation.Register(r, EnableAdmin)
	operation.Register(r, CreateActivationLink)
	operation.Register(r, ActivateAccount)
	operation.Register(r, SignIn)
	operation.Register(r, SignOut)
	operation.Register(r, ChangePassword)
}

// Admin is an administrator of the server.
type Admin struct {
	Handle    string    `json:"handle"`
	Name      string    `json:"name"`
	Status    string    `json:"status" jsonschema:"pending until the administrator chooses a password, then active; disabled while they may not sign in"`
	CreatedAt time.Time `json:"created_at"`
}

// adminRecord is an administrator as the operations read them.
type adminRecord struct {
	Admin
	id           int64
	passwordHash string
}

func findAdmin(ctx context.Context, tx *sql.Tx, handle string) (adminRecord, error) {
	var a adminRecord
	var hash, disabled sql.NullString
	var created string
	err := tx.QueryRowContext(ctx, "SELECT id, handle, name, password_hash, disabled_at, created_at FROM admins WHERE handle = ?",
		handle).Scan(&a.id, &a.Handle, &a.Name, &hash, &disabled, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return adminRecord{}, operation.NotFoundf("no administrator %q on this server", handle)
	}
	if err != nil {
		return adminRecord{}, err
	}
	a.passwordHash = hash.String
	a.Status = adminStatus(hash, disabled)
	a.CreatedAt, err = store.ParseTime(created)
	return a, err
}

func adminStatus(passwordHash, disabledAt sql.NullString) string {
	switch {
	case disabledAt.Valid:
		return account.Disabled
	case passwordHash.Valid:
		return account.Active
	}
	return account.Pending
}

// CreateAdminInput is the input of create_admin.
type CreateAdminInput struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
}

// NewAdmin is the output of create_admin.
type NewAdmin struct {
	Admin
	Activation account.Activation `json:"activation" jsonschema:"the link where the administrator chooses their password"`
}

// CreateAdmin creates an administrator of the server, with an activation
// link.
var CreateAdmin = &operation.Op[CreateAdminInput, NewAdmin]{
	Name:        "create_admin",
	Title:       "Create a server administrator",
	Description: "Create an administrator of the server, and the link where they choose their password.",
	Level:       operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in CreateAdminInput) (NewAdmin, error) {
		c.Log(in.Handle, "")
		if !account.ValidHandle(in.Handle) {
			return NewAdmin{}, operation.Invalidf("handle %q: use 2 to 32 lowercase letters, digits and hyphens, "+
				"starting with a letter and ending with a letter or a digit", in.Handle)
		}
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return NewAdmin{}, operation.Invalidf("name is empty")
		}
		a := NewAdmin{Admin: Admin{Handle: in.Handle, Name: name, Status: account.Pending, CreatedAt: c.Now}}
		var id int64
		err := c.Tx.QueryRowContext(ctx, "INSERT INTO admins (handle, name, created_at) VALUES (?, ?, ?) ON CONFLICT DO NOTHING RETURNING id",
			a.Handle, a.Name, store.FormatTime(c.Now)).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return NewAdmin{}, operation.Invalidf("handle %q is taken by another administrator", a.Handle)
		}
		if err != nil {
			return NewAdmin{}, err
		}
		a.Activation, err = newActivation(ctx, c, id, a.Handle)
		return a, err
	},
}

// ListAdminsInput is the input of list_admins.
type ListAdminsInput struct{}

// AdminList is the output of list_admins.
type AdminList struct {
	Admins []Admin `json:"admins"`
}

// ListAdmins lists the administrators of the server, by handle.
var ListAdmins = &operation.Op[ListAdminsInput, AdminList]{
	Name:        "list_admins",
	Title:       "List server administrators",
	Description: "List the administrators of the server.",
	Level:       operation.ServerAdmin,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, _ ListAdminsInput) (AdminList, error) {
		rows, err := c.Tx.QueryContext(ctx, "SELECT handle, name, password_hash, disabled_at, created_at FROM admins ORDER BY handle")
		if err != nil {
			return AdminList{}, err
		}
		defer rows.Close()
		list := AdminList{Admins: []Admin{}}
		for rows.Next() {
			var a Admin
			var hash, disabled sql.NullString
			var created string
			if err := rows.Scan(&a.Handle, &a.Name, &hash, &disabled, &created); err != nil {
				return AdminList{}, err
			}
			a.Status = adminStatus(hash, disabled)
			if a.CreatedAt, err = store.ParseTime(created); err != nil {
				return AdminList{}, err
			}
			list.Admins = append(list.Admins, a)
		}
		return list, rows.Err()
	},
}

// newActivation creates an activation link for the administrator id, which
// replaces any link they had.
func newActivation(ctx context.Context, c *operation.Call, id int64, handle string) (account.Activation, error) {
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_activations WHERE admin_id = ?", id); err != nil {
		return account.Activation{}, err
	}
	a := account.Activation{User: handle, Code: credential.Token(""), ExpiresAt: c.Now.Add(credential.ActivationTTL)}
	_, err := c.Tx.ExecContext(ctx, "INSERT INTO admin_activations (admin_id, hash, created_at, expires_at) VALUES (?, ?, ?, ?)",
		id, credential.Hash(a.Code), store.FormatTime(c.Now), store.FormatTime(a.ExpiresAt))
	return a, err
}

// AdminInput names an administrator of the server, by handle.
type AdminInput struct {
	Admin string `json:"admin" jsonschema:"handle of the administrator"`
}

// DisableAdmin disables an administrator of the server: they can no longer
// sign in, and their sessions and their link end.
var DisableAdmin = &operation.Op[AdminInput, Admin]{
	Name:  "disable_admin",
	Title: "Disable a server administrator",
	Description: "Disable an administrator of the server: their sessions and their link end, and they cannot sign in " +
		"until another administrator enables them again. The server keeps at least one administrator who is not disabled.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in AdminInput) (Admin, error) {
		c.Log(in.Admin, "")
		a, err := findAdmin(ctx, c.Tx, in.Admin)
		if err != nil {
			return Admin{}, err
		}
		if a.Status == account.Disabled {
			return Admin{}, operation.Invalidf("%s is already disabled", a.Handle)
		}
		if a.Handle == c.Actor.Handle {
			return Admin{}, operation.Invalidf("you cannot disable yourself: another administrator of the server can")
		}
		var others int
		err = c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM admins WHERE disabled_at IS NULL AND id != ?", a.id).Scan(&others)
		if err != nil {
			return Admin{}, err
		}
		if others == 0 {
			return Admin{}, operation.Invalidf("the server needs an administrator: create or enable another one first")
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET disabled_at = ? WHERE id = ?", store.FormatTime(c.Now), a.id); err != nil {
			return Admin{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_sessions WHERE admin_id = ?", a.id); err != nil {
			return Admin{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_activations WHERE admin_id = ?", a.id); err != nil {
			return Admin{}, err
		}
		a.Status = account.Disabled
		return a.Admin, nil
	},
}

// EnableAdmin enables a disabled administrator of the server again.
var EnableAdmin = &operation.Op[AdminInput, Admin]{
	Name:        "enable_admin",
	Title:       "Enable a server administrator",
	Description: "Enable a disabled administrator of the server again: they can sign in with their password, or with a new link.",
	Level:       operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in AdminInput) (Admin, error) {
		c.Log(in.Admin, "")
		a, err := findAdmin(ctx, c.Tx, in.Admin)
		if err != nil {
			return Admin{}, err
		}
		if a.Status != account.Disabled {
			return Admin{}, operation.Invalidf("%s is not disabled", a.Handle)
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET disabled_at = NULL WHERE id = ?", a.id); err != nil {
			return Admin{}, err
		}
		a.Status = adminStatus(sql.NullString{String: a.passwordHash, Valid: a.passwordHash != ""}, sql.NullString{})
		return a.Admin, nil
	},
}

// CreateActivationLink gives an administrator of the server a new link
// where they choose their password.
var CreateActivationLink = &operation.Op[AdminInput, account.Activation]{
	Name:  "create_activation_link",
	Title: "Give an administrator a new link",
	Description: "Give an administrator of the server a new single-use link, valid 7 days, where they choose their password, " +
		"for instance after they forgot it. Their password works until they use the link.",
	Level:      operation.ServerAdmin,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in AdminInput) (account.Activation, error) {
		c.Log(in.Admin, "")
		a, err := findAdmin(ctx, c.Tx, in.Admin)
		if err != nil {
			return account.Activation{}, err
		}
		if a.Status == account.Disabled {
			return account.Activation{}, operation.Invalidf("%s is disabled: enable them first", a.Handle)
		}
		return newActivation(ctx, c, a.id, a.Handle)
	},
}

// Session is a session of an administrator of the server, just opened.
// Token is shown once and never stored.
type Session struct {
	Token     string    `json:"token" jsonschema:"send it in the Authorization header: Bearer <token>"`
	Domain    string    `json:"domain" jsonschema:"the domain where the session holds"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the session ends at the latest; it ends sooner after 7 days without use"`
	Admin     Admin     `json:"admin"`
}

// openSession opens a session for the administrator a at the domain of the
// call, and deletes the sessions of administrators that have ended.
func openSession(ctx context.Context, c *operation.Call, a adminRecord) (Session, error) {
	if c.Actor.Domain == "" {
		return Session{}, operation.Invalidf("a session holds at the domain of a Cahut, and this call came to none")
	}
	_, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_sessions WHERE expires_at <= ? OR last_used_at <= ?",
		store.FormatTime(c.Now), store.FormatTime(c.Now.Add(-credential.SessionIdle)))
	if err != nil {
		return Session{}, err
	}
	s := Session{Token: credential.Token(SessionPrefix), Domain: c.Actor.Domain, ExpiresAt: c.Now.Add(credential.SessionMax), Admin: a.Admin}
	now := store.FormatTime(c.Now)
	_, err = c.Tx.ExecContext(ctx, `
		INSERT INTO admin_sessions (admin_id, domain, hash, created_at, last_used_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		a.id, s.Domain, credential.Hash(s.Token), now, now, store.FormatTime(s.ExpiresAt))
	c.Log(a.Handle, "at "+s.Domain)
	return s, err
}

// ActivateAccount sets an administrator's password from an activation link,
// and signs them in at the domain of the call. It ends their other sessions.
var ActivateAccount = &operation.Op[account.ActivateInput, Session]{
	Name:        "activate_account",
	Title:       "Activate an administrator account",
	Description: "Choose a password from a single-use activation link, as an administrator of the server, and sign in.",
	PeopleOnly:  true,
	Visitors:    true,
	Run: func(ctx context.Context, c *operation.Call, in account.ActivateInput) (Session, error) {
		c.Identify(operation.Admin, 0, "")
		if err := credential.CheckPolicy(in.Password); err != nil {
			return Session{}, err
		}
		var handle string
		err := c.Tx.QueryRowContext(ctx, `
			SELECT a.handle FROM admin_activations v JOIN admins a ON a.id = v.admin_id
			WHERE v.hash = ? AND v.expires_at > ?`, credential.Hash(strings.TrimSpace(in.Code)), store.FormatTime(c.Now)).Scan(&handle)
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, operation.Forbiddenf("this activation link is unknown, used or expired: " +
				"ask another administrator of the server for a new one, or run cahut admin link on the server")
		}
		if err != nil {
			return Session{}, err
		}
		a, err := findAdmin(ctx, c.Tx, handle)
		if err != nil {
			return Session{}, err
		}
		c.Identify(operation.Admin, 0, a.Handle)
		c.Log(a.Handle, "")
		if a.Status == account.Disabled {
			return Session{}, operation.Forbiddenf("this account is disabled: another administrator of the server can enable it")
		}
		hash, err := credential.HashPassword(ctx, in.Password)
		if err != nil {
			return Session{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET password_hash = ? WHERE id = ?", hash, a.id); err != nil {
			return Session{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_activations WHERE admin_id = ?", a.id); err != nil {
			return Session{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_sessions WHERE admin_id = ?", a.id); err != nil {
			return Session{}, err
		}
		a.Status = account.Active
		return openSession(ctx, c, a)
	},
}

// SignIn opens a session for an administrator of the server, at the domain
// of the call.
var SignIn = &operation.Op[account.SignInInput, Session]{
	Name:        "sign_in",
	Title:       "Sign in to the server",
	Description: "Sign in as an administrator of the server, with a handle and a password, and open a session at this domain.",
	PeopleOnly:  true,
	Visitors:    true,
	Prepare: func(ctx context.Context, c *operation.Call, in account.SignInInput) error {
		// The password is checked before the call's transaction opens, as
		// for the people of a Cahut.
		c.Identify(operation.Admin, 0, "")
		var a adminRecord
		attempt := credential.SignIn{Password: in.Password}
		address := sql.NullString{String: c.Actor.Address, Valid: c.Actor.Address != ""}
		err := c.Read(ctx, func(tx *sql.Tx) error {
			var err error
			a, err = findAdmin(ctx, tx, strings.ToLower(strings.TrimSpace(in.Handle)))
			if errors.Is(err, operation.ErrNotFound) {
				// An unknown handle is not logged: it may be a password
				// typed in the wrong field.
				return tx.QueryRowContext(ctx, `
					SELECT count(*) FROM logbook
					WHERE operation = 'sign_in' AND outcome = 'refused' AND user_handle IS NULL AND address IS ? AND at > ?`,
					address, store.FormatTime(c.Now.Add(-credential.RefusalWindow))).Scan(&attempt.RefusedHere)
			}
			if err != nil {
				return err
			}
			attempt.Hash = a.passwordHash
			return tx.QueryRowContext(ctx, `
				SELECT count(*), count(*) FILTER (WHERE address IS ?) FROM logbook
				WHERE operation = 'sign_in' AND outcome = 'refused' AND user_handle = ? AND at > ?`,
				address, a.Handle, store.FormatTime(c.Now.Add(-credential.RefusalWindow))).Scan(&attempt.RefusedAll, &attempt.RefusedHere)
		})
		if err != nil {
			return err
		}
		if a.id == 0 {
			c.Log("", "unknown handle")
			return credential.CheckSignIn(ctx, attempt)
		}
		c.Identify(operation.Admin, 0, a.Handle)
		c.Log(a.Handle, "at "+c.Actor.Domain)
		if err := credential.CheckSignIn(ctx, attempt); err != nil {
			return err
		}
		c.Keep(a.passwordHash)
		return nil
	},
	Run: func(ctx context.Context, c *operation.Call, in account.SignInInput) (Session, error) {
		a, err := findAdmin(ctx, c.Tx, strings.ToLower(strings.TrimSpace(in.Handle)))
		if errors.Is(err, operation.ErrNotFound) {
			return Session{}, credential.ErrWrongPassword // deleted since its password was checked
		}
		if err != nil {
			return Session{}, err
		}
		if checked, _ := c.Kept().(string); checked == "" || checked != a.passwordHash {
			return Session{}, credential.ErrWrongPassword // its password changed since it was checked
		}
		if a.Status == account.Disabled {
			return Session{}, operation.Forbiddenf("this account is disabled: another administrator of the server can enable it")
		}
		return openSession(ctx, c, a)
	},
}

// SignOut ends the session of the call.
var SignOut = &operation.Op[account.SignOutInput, account.SignedOut]{
	Name:        "sign_out",
	Title:       "Sign out of the server",
	Description: "End the session you signed in with, as an administrator of the server.",
	Level:       operation.ServerAdmin,
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, _ account.SignOutInput) (account.SignedOut, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.Session == 0 {
			return account.SignedOut{}, operation.Invalidf("sign_out ends a session, and this call comes with none")
		}
		_, err := c.Tx.ExecContext(ctx, `
			DELETE FROM admin_sessions WHERE id = ? AND admin_id = (SELECT id FROM admins WHERE handle = ?)`,
			c.Actor.Session, c.Actor.Handle)
		return account.SignedOut{}, err
	},
}

// ChangePassword changes the password of the administrator who calls, and
// ends their other sessions.
var ChangePassword = &operation.Op[account.ChangePasswordInput, account.PasswordChanged]{
	Name:        "change_password",
	Title:       "Change your server password",
	Description: "Change your password, as an administrator of the server. Your other sessions end.",
	Level:       operation.ServerAdmin,
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in account.ChangePasswordInput) (account.PasswordChanged, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.CommandLine() {
			return account.PasswordChanged{}, operation.Invalidf("change_password acts for an administrator who signed in, " +
				"and the command line is none: use cahut admin link")
		}
		if err := credential.CheckPolicy(in.NewPassword); err != nil {
			return account.PasswordChanged{}, err
		}
		a, err := findAdmin(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return account.PasswordChanged{}, err
		}
		ok, err := credential.CheckPassword(ctx, a.passwordHash, in.Password)
		if err != nil {
			return account.PasswordChanged{}, err
		}
		if !ok {
			return account.PasswordChanged{}, operation.Forbiddenf("the current password is wrong")
		}
		hash, err := credential.HashPassword(ctx, in.NewPassword)
		if err != nil {
			return account.PasswordChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET password_hash = ? WHERE id = ?", hash, a.id); err != nil {
			return account.PasswordChanged{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "DELETE FROM admin_sessions WHERE admin_id = ? AND id != ?", a.id, c.Actor.Session)
		return account.PasswordChanged{}, err
	},
}

// AdminSession is who a valid session of an administrator belongs to.
type AdminSession struct {
	SessionID int64
	Handle    string
}

// VerifySession checks the session token of an administrator of the server,
// at domain, and returns who it belongs to. A session of another domain is
// invalid. It records when the session was last used, at most once a minute.
func (s *Server) VerifySession(ctx context.Context, token, domain string, now time.Time) (AdminSession, error) {
	if !strings.HasPrefix(token, SessionPrefix) {
		return AdminSession{}, account.ErrInvalidSession
	}
	var a AdminSession
	err := s.DB.Read.QueryRowContext(ctx, `
		SELECT s.id, a.handle FROM admin_sessions s JOIN admins a ON a.id = s.admin_id
		WHERE s.hash = ? AND s.domain = ? AND s.expires_at > ? AND s.last_used_at > ? AND a.disabled_at IS NULL`,
		credential.Hash(token), domain, store.FormatTime(now), store.FormatTime(now.Add(-credential.SessionIdle))).Scan(&a.SessionID, &a.Handle)
	if errors.Is(err, sql.ErrNoRows) {
		return AdminSession{}, account.ErrInvalidSession
	}
	if err != nil {
		return AdminSession{}, err
	}
	_, err = s.DB.Write.ExecContext(ctx, "UPDATE admin_sessions SET last_used_at = ? WHERE id = ? AND last_used_at < ?",
		store.FormatTime(now), a.SessionID, store.FormatTime(now.Add(-time.Minute)))
	return a, err
}
