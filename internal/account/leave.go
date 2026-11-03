package account

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
)

// A person deletes their own account with their current password, and
// their handle typed in full. Cahut checks the password before the
// transaction, as sign_in does. delete_account lives with the Windows,
// whose files it deletes. See
// docs/adr/0059-people-delete-their-own-account.md.

// errHandleMismatch refuses a handle that is not that of the person who
// deletes their account.
var errHandleMismatch = operation.NewRefusal("account_handle_mismatch", operation.ErrInvalid,
	"this is not your handle: type it as you sign in with it")

// DeleteAccountInput is the input of delete_account.
type DeleteAccountInput struct {
	Password string `json:"password" jsonschema:"your current password"`
	Handle   string `json:"handle" jsonschema:"your handle, typed in full"`
}

// PrepareLeaving checks the password of the person who deletes their
// account, before the transaction of the call, and keeps the hash it
// checked for Leaving. It is the Prepare of delete_account.
func PrepareLeaving(ctx context.Context, c *operation.Call, in DeleteAccountInput) error {
	id, err := self(c, "delete_account")
	if err != nil {
		return err
	}
	// The Logbook names the person, never what they typed: a password
	// typed in the wrong field would be kept.
	c.Log(c.Actor.Handle, "")
	var hash sql.NullString
	err = c.Read(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id = ?", id).Scan(&hash)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return operation.Forbiddenf("this person is not in the Cahut")
	}
	if err != nil {
		return err
	}
	if !hash.Valid {
		return operation.Invalidf("%s has no password yet: use an activation link", c.Actor.Handle)
	}
	ok, err := credential.CheckPassword(ctx, hash.String, in.Password)
	if err != nil {
		return err
	}
	if !ok {
		return credential.ErrWrongCurrentPassword
	}
	c.Keep(hash.String)
	return nil
}

// Leaving returns the person who deletes their account, in the transaction
// of the call, once PrepareLeaving checked their password: it refuses a
// password changed since, a handle that is not theirs, and the last
// administrator of the Cahut.
func Leaving(ctx context.Context, c *operation.Call, in DeleteAccountInput) (User, error) {
	id, err := self(c, "delete_account")
	if err != nil {
		return User{}, err
	}
	var handle string
	err = c.Tx.QueryRowContext(ctx, "SELECT handle FROM users WHERE id = ?", id).Scan(&handle)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, operation.Forbiddenf("this person is not in the Cahut")
	}
	if err != nil {
		return User{}, err
	}
	p, err := findUser(ctx, c.Tx, handle)
	if err != nil {
		return User{}, err
	}
	c.Log(p.Handle, "")
	if checked, _ := c.Kept().(string); checked == "" || checked != p.passwordHash {
		return User{}, credential.ErrWrongCurrentPassword // its password changed since it was checked
	}
	if strings.ToLower(strings.TrimSpace(in.Handle)) != p.Handle {
		return User{}, errHandleMismatch
	}
	last, err := IsLastAdmin(ctx, c.Tx, p.ID)
	if err != nil {
		return User{}, err
	}
	if last {
		return User{}, errLastAdmin
	}
	return p.User, nil
}

// DeletedLetter tells a person that their account is deleted. It is
// written while the account is there, and goes once it is gone.
func DeletedLetter(p mail.Person, name, where string) mail.Letter {
	return mail.Letter{
		Subject: "Your account is deleted",
		Lines: []string{"Hello " + p.Name + ",", "Your account " + p.Handle + " on " + where +
			" is deleted, with your personal Room and the Rooms that were yours alone."},
		After: []string{"If you did not delete it, tell an administrator of " + name + " now."},
	}
}
