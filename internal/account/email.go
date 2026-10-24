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

// The e-mail address of a person, where their links go. The person and the
// administrators of the Cahut read it; no operation gives it to an AI. See
// docs/adr/0031-e-mail-through-a-relay.md.

// addressFree returns an error wrapping operation.ErrInvalid when email is
// the address of a person other than userID.
func addressFree(ctx context.Context, tx *sql.Tx, email string, userID int64) error {
	var taken bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = ? COLLATE NOCASE AND id != ?)", email, userID).Scan(&taken)
	if err != nil {
		return err
	}
	if taken {
		return operation.Invalidf("%s is already the e-mail address of another person of this Cahut", email)
	}
	return nil
}

// EmailChange is an address that waits for its person to confirm it.
type EmailChange struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the link sent to it stops working"`
}

// pendingChange returns the address that waits for the person userID to
// confirm it, or nil.
func pendingChange(ctx context.Context, c *operation.Call, userID int64) (*EmailChange, error) {
	var e EmailChange
	var expires string
	err := c.Tx.QueryRowContext(ctx, "SELECT email, expires_at FROM email_changes WHERE user_id = ? AND expires_at > ?",
		userID, store.FormatTime(c.Now)).Scan(&e.Email, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.ExpiresAt, err = store.ParseTime(expires)
	return &e, err
}

// cahutName returns how a message names the Cahut of the outbox o: its
// name, and where it is, such as "Acme, at cahut.acme.com", or "the Cahut
// at cahut.acme.com" when it has no name of its own.
func cahutName(ctx context.Context, o *mail.Outbox) (name, where string, err error) {
	name, err = o.Name(ctx)
	if err != nil {
		return "", "", err
	}
	if name == o.Domain {
		return name, "the Cahut at " + o.Domain, nil
	}
	return name, name + ", at " + o.Domain, nil
}

// letter makes a message to the person p from the Cahut of the outbox o.
type letter func(name, where string, p person, link string, expires time.Time) mail.Letter

func activationLetter(name, where string, p person, link string, expires time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Choose your password for " + name,
		Lines:   []string{"Hello " + p.Name + ",", "An account waits for you on " + where + ". Your handle is " + p.Handle + "."},
		Action:  &mail.Action{Label: "Choose your password", URL: link},
		After: []string{"This link works once, until " + mail.Day(expires) + ".",
			"If you did not expect this message, you can ignore it."},
	}
}

func resetLetter(name, where string, p person, link string, _ time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Choose a new password for " + name,
		Lines:   []string{"Hello " + p.Name + ",", "Someone asked for a new password for your account " + p.Handle + " on " + where + "."},
		Action:  &mail.Action{Label: "Choose a new password", URL: link},
		After: []string{"This link works once, for one hour. Your password works until you use it.",
			"If you did not ask for a new password, you can ignore this message."},
	}
}

func confirmLetter(_, where string, p person, link string, _ time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Confirm your e-mail address",
		Lines:   []string{"Hello " + p.Name + ",", "Confirm that this is your e-mail address for your account " + p.Handle + " on " + where + "."},
		Action:  &mail.Action{Label: "Confirm this address", URL: link},
		After:   []string{"This link works for 24 hours. If you did not ask for this, you can ignore this message: nothing changes."},
	}
}

// Pages of the console where the links of activations open: that of a
// forgotten password says so.
const (
	activatePage = "activate#"
	resetPage    = "activate?reset#"
)

// sendLink sends the link a, which opens at page of the console, to the
// address of the person p, in the message that write makes, once the call
// is committed. It reports whether it went: the person has an address, and
// the Cahut sends e-mail.
func sendLink(ctx context.Context, c *operation.Call, p person, a Activation, page string, write letter) (bool, error) {
	o := mail.OutboxOf(c)
	if p.email == "" {
		return false, nil
	}
	if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
		return false, err
	}
	name, where, err := cahutName(ctx, o)
	if err != nil {
		return false, err
	}
	l := write(name, where, p, o.Link(page+a.Code), a.ExpiresAt)
	l.To = p.email
	o.Post(c, l)
	return true, nil
}

// noticeChange tells the old address of the person p that their address is
// now email, or was removed when email is empty.
func noticeChange(ctx context.Context, c *operation.Call, p person, old, email string) error {
	o := mail.OutboxOf(c)
	if old == "" || strings.EqualFold(old, email) {
		return nil
	}
	if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
		return err
	}
	name, where, err := cahutName(ctx, o)
	if err != nil {
		return err
	}
	now := "The e-mail address of your account " + p.Handle + " on " + where + " is now " + email + "."
	if email == "" {
		now = "The e-mail address of your account " + p.Handle + " on " + where + " was removed: links no longer come to this address."
	}
	o.Post(c, mail.Letter{
		To:      old,
		Subject: "Your e-mail address changed",
		Lines:   []string{"Hello " + p.Name + ",", now},
		After:   []string{"If you did not ask for this, tell an administrator of " + name + "."},
	})
	return nil
}

// SetUserEmailInput is the input of set_user_email.
type SetUserEmailInput struct {
	User  string `json:"user" jsonschema:"handle of the person"`
	Email string `json:"email" jsonschema:"their e-mail address; empty removes it"`
}

// SetUserEmail gives a person an e-mail address, or removes it, for the
// administrators of the Cahut. It decides where the links of the person go:
// an AI knocks first.
var SetUserEmail = &operation.Op[SetUserEmailInput, User]{
	Name:  "set_user_email",
	Title: "Set the e-mail address of a person",
	Description: "Give a person an e-mail address, where their activation links and notices go, or remove it with an empty address. " +
		"Their old address gets a notice. For the administrators of the Cahut.",
	Level:      operation.CahutAdmin,
	KnockFirst: true,
	Run: func(ctx context.Context, c *operation.Call, in SetUserEmailInput) (User, error) {
		c.Log(in.User, "")
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return User{}, err
		}
		var email sql.NullString
		if strings.TrimSpace(in.Email) != "" {
			e, err := mail.Address(in.Email)
			if err != nil {
				return User{}, operation.Invalidf("email: %v", err)
			}
			if err := addressFree(ctx, c.Tx, e, p.ID); err != nil {
				return User{}, err
			}
			email = sql.NullString{String: e, Valid: true}
			c.Log(p.Handle, "address set")
		} else {
			c.Log(p.Handle, "address removed")
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET email = ? WHERE id = ?", email, p.ID); err != nil {
			return User{}, err
		}
		// The address an administrator gives replaces one that waited.
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM email_changes WHERE user_id = ?", p.ID); err != nil {
			return User{}, err
		}
		if err := noticeChange(ctx, c, p, p.email, email.String); err != nil {
			return User{}, err
		}
		u := p.User
		if c.Actor.Kind != operation.AI || c.Approved() {
			u.Email = email.String
		}
		return u, nil
	},
}

// ChangeEmailInput is the input of change_email.
type ChangeEmailInput struct {
	Email string `json:"email" jsonschema:"your new e-mail address, which a link sent to it confirms; empty removes yours"`
}

// EmailChanged is the output of change_email and confirm_email.
type EmailChanged struct {
	Email   string       `json:"email,omitempty" jsonschema:"your e-mail address now"`
	Waiting *EmailChange `json:"waiting,omitempty" jsonschema:"the address that waits for you to confirm it, from the link sent to it"`
}

// ChangeEmail starts the change of the address of the person who calls: a
// link goes to the new address, which replaces theirs once they use it. No
// AI changes it, as no AI changes a password.
var ChangeEmail = &operation.Op[ChangeEmailInput, EmailChanged]{
	Name:  "change_email",
	Title: "Change your e-mail address",
	Description: "Change your e-mail address: a link goes to the new one, which replaces yours once you use the link, " +
		"within 24 hours, and your old address gets a notice. An empty address removes yours at once.",
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in ChangeEmailInput) (EmailChanged, error) {
		if _, err := self(c, "change_email"); err != nil {
			return EmailChanged{}, err
		}
		c.Log(c.Actor.Handle, "")
		p, err := findUser(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM email_changes WHERE user_id = ?", p.ID); err != nil {
			return EmailChanged{}, err
		}
		if strings.TrimSpace(in.Email) == "" {
			c.Log(p.Handle, "address removed")
			if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET email = NULL WHERE id = ?", p.ID); err != nil {
				return EmailChanged{}, err
			}
			return EmailChanged{}, noticeChange(ctx, c, p, p.email, "")
		}
		email, err := mail.Address(in.Email)
		if err != nil {
			return EmailChanged{}, operation.Invalidf("email: %v", err)
		}
		if strings.EqualFold(email, p.email) {
			return EmailChanged{}, operation.Invalidf("%s is already your e-mail address", email)
		}
		o := mail.OutboxOf(c)
		if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
			return EmailChanged{}, errors.Join(err, operation.Invalidf("this Cahut sends no e-mail, so the address cannot be confirmed: "+
				"ask an administrator of the Cahut to set it"))
		}
		// Whether another person has this address is not said here, but
		// once the link is used: only the owner of the address learns it.
		code := credential.Token("")
		w := EmailChange{Email: email, ExpiresAt: c.Now.Add(credential.ConfirmTTL)}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO email_changes (user_id, email, hash, expires_at) VALUES (?, ?, ?, ?)",
			p.ID, email, credential.Hash(code), store.FormatTime(w.ExpiresAt))
		if err != nil {
			return EmailChanged{}, err
		}
		c.Log(p.Handle, "confirmation sent")
		name, where, err := cahutName(ctx, o)
		if err != nil {
			return EmailChanged{}, err
		}
		l := confirmLetter(name, where, p, o.Link("confirm-email#"+code), w.ExpiresAt)
		l.To = email
		o.Post(c, l)
		return EmailChanged{Email: p.email, Waiting: &w}, nil
	},
}

// ConfirmEmailInput is the input of confirm_email.
type ConfirmEmailInput struct {
	Code string `json:"code" jsonschema:"the code of the link sent to the new address"`
}

// ConfirmEmail makes the address that a link went to the address of its
// person, who need not be signed in: the link is the proof.
var ConfirmEmail = &operation.Op[ConfirmEmailInput, EmailChanged]{
	Name:        "confirm_email",
	Title:       "Confirm an e-mail address",
	Description: "Confirm a new e-mail address from the link sent to it, which makes it the address of its person.",
	PeopleOnly:  true,
	Visitors:    true,
	Run: func(ctx context.Context, c *operation.Call, in ConfirmEmailInput) (EmailChanged, error) {
		var handle, email string
		err := c.Tx.QueryRowContext(ctx, `
			SELECT u.handle, e.email FROM email_changes e JOIN users u ON u.id = e.user_id
			WHERE e.hash = ? AND e.expires_at > ?`, credential.Hash(strings.TrimSpace(in.Code)), store.FormatTime(c.Now)).Scan(&handle, &email)
		if errors.Is(err, sql.ErrNoRows) {
			return EmailChanged{}, operation.Forbiddenf("this link is unknown, used or expired: change your address again from your account")
		}
		if err != nil {
			return EmailChanged{}, err
		}
		p, err := findUser(ctx, c.Tx, handle)
		if err != nil {
			return EmailChanged{}, err
		}
		c.Identify(operation.Person, p.ID, p.Handle)
		c.Log(p.Handle, "address confirmed")
		if p.Status == Disabled {
			return EmailChanged{}, operation.Forbiddenf("this account is disabled: an administrator of the Cahut can enable it")
		}
		if err := addressFree(ctx, c.Tx, email, p.ID); err != nil {
			return EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET email = ? WHERE id = ?", email, p.ID); err != nil {
			return EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM email_changes WHERE user_id = ?", p.ID); err != nil {
			return EmailChanged{}, err
		}
		return EmailChanged{Email: email}, noticeChange(ctx, c, p, p.email, email)
	},
}

// RequestPasswordLinkInput is the input of request_password_link.
type RequestPasswordLinkInput struct {
	Who string `json:"who" jsonschema:"the handle of the account, or its e-mail address"`
}

// PasswordLinkRequested is the output of request_password_link.
type PasswordLinkRequested struct {
	Message string `json:"message"`
}

// passwordLinkAnswer is the answer of request_password_link, the same
// whether the account exists, has an address or is disabled: it tells a
// visitor nothing about the people of the Cahut.
const passwordLinkAnswer = "If an account of this Cahut has this handle or this address, " +
	"a link to choose a new password went to its address. It works for one hour."

// RequestPasswordLink sends a person who forgot their password a link where
// they choose a new one, at the address of their account.
var RequestPasswordLink = &operation.Op[RequestPasswordLinkInput, PasswordLinkRequested]{
	Name:  "request_password_link",
	Title: "Ask for a new password",
	Description: "Send the link where a person chooses a new password to the e-mail address of their account, " +
		"given its handle or its address. The link works for one hour; the password works until it is used. " +
		"The answer is the same whether the account exists or not.",
	PeopleOnly: true,
	Visitors:   true,
	Run: func(ctx context.Context, c *operation.Call, in RequestPasswordLinkInput) (PasswordLinkRequested, error) {
		answer := PasswordLinkRequested{Message: passwordLinkAnswer}
		who := strings.TrimSpace(in.Who)
		handle := strings.ToLower(who)
		if strings.Contains(who, "@") {
			err := c.Tx.QueryRowContext(ctx, "SELECT handle FROM users WHERE email = ? COLLATE NOCASE", who).Scan(&handle)
			if errors.Is(err, sql.ErrNoRows) {
				handle = ""
			} else if err != nil {
				return PasswordLinkRequested{}, err
			}
		}
		p, err := findUser(ctx, c.Tx, handle)
		if errors.Is(err, operation.ErrNotFound) {
			// What a visitor typed is not logged: it may be anyone's address.
			c.Log("", "unknown account")
			return answer, nil
		}
		if err != nil {
			return PasswordLinkRequested{}, err
		}
		c.Identify(operation.Person, p.ID, p.Handle)
		var recent bool
		err = c.Tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM activations WHERE user_id = ? AND created_at > ?)",
			p.ID, store.FormatTime(c.Now.Add(-credential.ResetEvery))).Scan(&recent)
		if err != nil {
			return PasswordLinkRequested{}, err
		}
		sends, err := mail.OutboxOf(c).Sends(ctx, c.Tx)
		if err != nil {
			return PasswordLinkRequested{}, err
		}
		switch {
		case p.Status == Disabled:
			c.Log(p.Handle, "not sent: disabled")
		case p.email == "":
			c.Log(p.Handle, "not sent: no address")
		case !sends:
			c.Log(p.Handle, "not sent: the Cahut sends no e-mail")
		case recent:
			c.Log(p.Handle, "not sent: a link went out less than 5 minutes ago")
		default:
			a, err := newActivation(ctx, c, p.ID, p.Handle, credential.ResetTTL)
			if err != nil {
				return PasswordLinkRequested{}, err
			}
			if _, err := sendLink(ctx, c, p, a, resetPage, resetLetter); err != nil {
				return PasswordLinkRequested{}, err
			}
			c.Log(p.Handle, "link sent")
		}
		return answer, nil
	},
}
