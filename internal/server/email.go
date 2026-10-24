package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// The e-mail address of an administrator of the server, where their links
// and the notices of the server go, through the relay of the server. See
// docs/adr/0031-e-mail-through-a-relay.md.

// adminAddressFree returns an error wrapping operation.ErrInvalid when email
// is the address of an administrator other than id.
func adminAddressFree(ctx context.Context, tx *sql.Tx, email string, id int64) error {
	var taken bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM admins WHERE email = ? COLLATE NOCASE AND id != ?)", email, id).Scan(&taken)
	if err != nil {
		return err
	}
	if taken {
		return operation.Invalidf("%s is already the e-mail address of another administrator of the server", email)
	}
	return nil
}

// Pages of the console where the links of administrators open.
const (
	adminActivatePage = "server/activate#"
	adminResetPage    = "server/activate?reset#"
	adminConfirmPage  = "server/confirm-email#"
)

// console returns the address of the console where the links of the call c
// open: that of the Cahut the call came to, or of the first Cahut of the
// server on the command line; empty when the server has none.
func console(ctx context.Context, c *operation.Call, o *mail.Outbox) (domain, address string, err error) {
	domain = c.Actor.Domain
	if domain == "" {
		err = c.Tx.QueryRowContext(ctx, "SELECT domain FROM cahuts ORDER BY id LIMIT 1").Scan(&domain)
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		if err != nil {
			return "", "", err
		}
	}
	return domain, o.Consoles(domain), nil
}

// adminLetter makes a message to the administrator a, who signs in at the
// console of domain.
type adminLetter func(domain string, a adminRecord, link string, expires time.Time) mail.Letter

func adminActivationLetter(domain string, a adminRecord, link string, expires time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Choose your password for the Cahut server",
		Lines: []string{"Hello " + a.Name + ",", "You are an administrator of the Cahut server of " + domain +
			". Your handle is " + a.Handle + "."},
		Action: &mail.Action{Label: "Choose your password", URL: link},
		After: []string{"This link works once, until " + mail.Day(expires) + ".",
			"If you did not expect this message, you can ignore it."},
	}
}

func adminResetLetter(domain string, a adminRecord, link string, _ time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Choose a new password for the Cahut server",
		Lines: []string{"Hello " + a.Name + ",", "Someone asked for a new password for your account " + a.Handle +
			" of administrator of the Cahut server of " + domain + "."},
		Action: &mail.Action{Label: "Choose a new password", URL: link},
		After: []string{"This link works once, for one hour. Your password works until you use it.",
			"If you did not ask for a new password, you can ignore this message."},
	}
}

func adminConfirmLetter(domain string, a adminRecord, link string, _ time.Time) mail.Letter {
	return mail.Letter{
		Subject: "Confirm your e-mail address",
		Lines: []string{"Hello " + a.Name + ",", "Confirm that this is your e-mail address for your account " + a.Handle +
			" of administrator of the Cahut server of " + domain + "."},
		Action: &mail.Action{Label: "Confirm this address", URL: link},
		After:  []string{"This link works for 24 hours. If you did not ask for this, you can ignore this message: nothing changes."},
	}
}

// sendAdminLink sends the link of the activation l, which opens at page of
// the console, to the address of the administrator a, in the message that
// write makes, once the call is committed. It reports whether it went.
func sendAdminLink(ctx context.Context, c *operation.Call, a adminRecord, l account.Activation, page string, write adminLetter) (bool, error) {
	o := mail.OutboxOf(c)
	if a.Email == "" || o == nil {
		return false, nil
	}
	if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
		return false, err
	}
	domain, at, err := console(ctx, c, o)
	if err != nil || at == "" {
		return false, err
	}
	m := write(domain, a, at+page+l.Code, l.ExpiresAt)
	m.To = a.Email
	o.Post(c, m)
	return true, nil
}

// noticeAdminChange tells the old address of the administrator a that
// their address is now email.
func noticeAdminChange(ctx context.Context, c *operation.Call, a adminRecord, email string) error {
	o := mail.OutboxOf(c)
	if a.Email == "" || o == nil || strings.EqualFold(a.Email, email) {
		return nil
	}
	if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok {
		return err
	}
	now := "The e-mail address of your account " + a.Handle + " of administrator of the Cahut server is now " + email + "."
	if email == "" {
		now = "The e-mail address of your account " + a.Handle + " of administrator of the Cahut server was removed: " +
			"links no longer come to this address."
	}
	o.Post(c, mail.Letter{
		To:      a.Email,
		Subject: "Your e-mail address changed",
		Lines:   []string{"Hello " + a.Name + ",", now},
		After:   []string{"If you did not ask for this, tell another administrator of the server."},
	})
	return nil
}

// ChangeEmail starts the change of the address of the administrator who
// calls: a link goes to the new address, which replaces theirs once they
// use it.
var ChangeEmail = &operation.Op[account.ChangeEmailInput, account.EmailChanged]{
	Name:  "change_email",
	Title: "Change your e-mail address",
	Description: "Change your e-mail address of administrator of the server: a link goes to the new one, which replaces yours " +
		"once you use the link, within 24 hours, and your old address gets a notice. An empty address removes yours at once.",
	Level:      operation.ServerAdmin,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in account.ChangeEmailInput) (account.EmailChanged, error) {
		c.Log(c.Actor.Handle, "")
		if c.Actor.Handle == "" {
			return account.EmailChanged{}, operation.Invalidf("change_email acts for an administrator who signed in: " +
				"on the command line, there is none")
		}
		a, err := findAdmin(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return account.EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_email_changes WHERE admin_id = ?", a.id); err != nil {
			return account.EmailChanged{}, err
		}
		if strings.TrimSpace(in.Email) == "" {
			c.Log(a.Handle, "address removed")
			if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET email = NULL WHERE id = ?", a.id); err != nil {
				return account.EmailChanged{}, err
			}
			return account.EmailChanged{}, noticeAdminChange(ctx, c, a, "")
		}
		email, err := mail.Address(in.Email)
		if err != nil {
			return account.EmailChanged{}, operation.Invalidf("email: %v", err)
		}
		if strings.EqualFold(email, a.Email) {
			return account.EmailChanged{}, operation.Invalidf("%s is already your e-mail address", email)
		}
		o := mail.OutboxOf(c)
		domain, at, err := "", "", error(nil)
		if o != nil {
			domain, at, err = console(ctx, c, o)
			if err != nil {
				return account.EmailChanged{}, err
			}
		}
		if ok, err := o.Sends(ctx, c.Tx); err != nil || !ok || at == "" {
			return account.EmailChanged{}, errors.Join(err, operation.Invalidf("this server sends no e-mail, so the address cannot be confirmed: "+
				"set the relay of the server first"))
		}
		code := credential.Token("")
		w := account.EmailChange{Email: email, ExpiresAt: c.Now.Add(credential.ConfirmTTL)}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO admin_email_changes (admin_id, email, hash, expires_at) VALUES (?, ?, ?, ?)",
			a.id, email, credential.Hash(code), store.FormatTime(w.ExpiresAt))
		if err != nil {
			return account.EmailChanged{}, err
		}
		c.Log(a.Handle, "confirmation sent")
		m := adminConfirmLetter(domain, a, at+adminConfirmPage+code, w.ExpiresAt)
		m.To = email
		o.Post(c, m)
		return account.EmailChanged{Email: a.Email, Waiting: &w}, nil
	},
}

// ConfirmEmail makes the address that a link went to the address of its
// administrator.
var ConfirmEmail = &operation.Op[account.ConfirmEmailInput, account.EmailChanged]{
	Name:        "confirm_email",
	Title:       "Confirm an e-mail address",
	Description: "Confirm a new e-mail address of an administrator of the server, from the link sent to it.",
	PeopleOnly:  true,
	Visitors:    true,
	Run: func(ctx context.Context, c *operation.Call, in account.ConfirmEmailInput) (account.EmailChanged, error) {
		var handle, email string
		err := c.Tx.QueryRowContext(ctx, `
			SELECT a.handle, e.email FROM admin_email_changes e JOIN admins a ON a.id = e.admin_id
			WHERE e.hash = ? AND e.expires_at > ?`, credential.Hash(strings.TrimSpace(in.Code)), store.FormatTime(c.Now)).Scan(&handle, &email)
		if errors.Is(err, sql.ErrNoRows) {
			return account.EmailChanged{}, operation.Forbiddenf("this link is unknown, used or expired: change your address again from your account")
		}
		if err != nil {
			return account.EmailChanged{}, err
		}
		a, err := findAdmin(ctx, c.Tx, handle)
		if err != nil {
			return account.EmailChanged{}, err
		}
		c.Identify(operation.Admin, 0, a.Handle)
		c.Log(a.Handle, "address confirmed")
		if a.Status == account.Disabled {
			return account.EmailChanged{}, operation.Forbiddenf("this account is disabled: another administrator of the server can enable it")
		}
		if err := adminAddressFree(ctx, c.Tx, email, a.id); err != nil {
			return account.EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE admins SET email = ? WHERE id = ?", email, a.id); err != nil {
			return account.EmailChanged{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_email_changes WHERE admin_id = ?", a.id); err != nil {
			return account.EmailChanged{}, err
		}
		return account.EmailChanged{Email: email}, noticeAdminChange(ctx, c, a, email)
	},
}

// RequestPasswordLink sends an administrator of the server who forgot their
// password a link where they choose a new one, at the address of their
// account.
var RequestPasswordLink = &operation.Op[account.RequestPasswordLinkInput, account.PasswordLinkRequested]{
	Name:  "request_password_link",
	Title: "Ask for a new password",
	Description: "Send the link where an administrator of the server chooses a new password to the e-mail address of their account, " +
		"given its handle or its address. The link works for one hour. The answer is the same whether the account exists or not.",
	PeopleOnly: true,
	Visitors:   true,
	Run: func(ctx context.Context, c *operation.Call, in account.RequestPasswordLinkInput) (account.PasswordLinkRequested, error) {
		answer := account.PasswordLinkRequested{Message: "If an administrator of this server has this handle or this address, " +
			"a link to choose a new password went to their address. It works for one hour."}
		who := strings.TrimSpace(in.Who)
		handle := strings.ToLower(who)
		if strings.Contains(who, "@") {
			err := c.Tx.QueryRowContext(ctx, "SELECT handle FROM admins WHERE email = ? COLLATE NOCASE", who).Scan(&handle)
			if errors.Is(err, sql.ErrNoRows) {
				handle = ""
			} else if err != nil {
				return account.PasswordLinkRequested{}, err
			}
		}
		a, err := findAdmin(ctx, c.Tx, handle)
		if errors.Is(err, operation.ErrNotFound) {
			c.Log("", "unknown account")
			return answer, nil
		}
		if err != nil {
			return account.PasswordLinkRequested{}, err
		}
		c.Identify(operation.Admin, 0, a.Handle)
		var recent bool
		err = c.Tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM admin_activations WHERE admin_id = ? AND created_at > ?)",
			a.id, store.FormatTime(c.Now.Add(-credential.ResetEvery))).Scan(&recent)
		if err != nil {
			return account.PasswordLinkRequested{}, err
		}
		sends, err := mail.OutboxOf(c).Sends(ctx, c.Tx)
		if err != nil {
			return account.PasswordLinkRequested{}, err
		}
		switch {
		case a.Status == account.Disabled:
			c.Log(a.Handle, "not sent: disabled")
		case a.Email == "":
			c.Log(a.Handle, "not sent: no address")
		case !sends:
			c.Log(a.Handle, "not sent: the server sends no e-mail")
		case recent:
			c.Log(a.Handle, "not sent: a link went out less than 5 minutes ago")
		default:
			link, err := newActivation(ctx, c, a.id, a.Handle, credential.ResetTTL)
			if err != nil {
				return account.PasswordLinkRequested{}, err
			}
			if _, err := sendAdminLink(ctx, c, a, link, adminResetPage, adminResetLetter); err != nil {
				return account.PasswordLinkRequested{}, err
			}
			c.Log(a.Handle, "link sent")
		}
		return answer, nil
	},
}
