package mail

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// SetMailRelayInput is the input of set_mail_relay.
type SetMailRelayInput struct {
	Host     string `json:"host" jsonschema:"the name or the address of the SMTP relay, such as smtp.example.com, without a scheme or a port; empty removes the relay"`
	Port     int    `json:"port,omitempty" jsonschema:"its port: 465 for TLS from the start, 587 for STARTTLS; 465 with tls and 587 otherwise when left out"`
	Security string `json:"security,omitempty" jsonschema:"tls, from the start, or starttls; tls on port 465 and starttls otherwise when left out"`
	Username string `json:"username,omitempty" jsonschema:"the user name at the relay; leave it out when the relay takes no authentication"`
	Password string `json:"password,omitempty" jsonschema:"the password at the relay; when left out, the password kept stays, as long as the host and the user name do not change"`
	From     string `json:"from,omitempty" jsonschema:"the address that messages come from, one that the relay accepts, such as cahut@example.com"`
	Name     string `json:"name,omitempty" jsonschema:"the name that messages come from; the domain of the Cahut when left out"`
}

// GetMailRelayInput is the input of get_mail_relay.
type GetMailRelayInput struct{}

// SendTestMailInput is the input of send_test_mail.
type SendTestMailInput struct {
	To string `json:"to,omitempty" jsonschema:"on the command line, or for an administrator of the server, the address to send the test to; a person's test goes to their own address"`
}

// Delivery is the output of set_mail_relay and get_mail_relay: the relay of
// the Cahut, and how its e-mail goes out.
type Delivery struct {
	Relay *RelayInfo `json:"relay,omitempty" jsonschema:"the relay of the Cahut; absent when there is none, and the Cahut then sends no e-mail"`
	Sends bool       `json:"sends" jsonschema:"whether the Cahut sends e-mail: it has a relay, or the server runs for development and writes its messages to disk"`
	Status
}

// RelayInfo describes a relay, without its password.
type RelayInfo struct {
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	Security string    `json:"security" jsonschema:"tls or starttls"`
	Username string    `json:"username,omitempty"`
	Password bool      `json:"password" jsonschema:"whether a password is kept; it is never given back"`
	From     string    `json:"from" jsonschema:"the address that messages come from"`
	Name     string    `json:"name,omitempty" jsonschema:"the name that messages come from; the domain of the Cahut when absent"`
	SetAt    time.Time `json:"set_at"`
}

// TestSent is the output of send_test_mail.
type TestSent struct {
	To string `json:"to"`
}

// scope is the Cahut whose relay the operations set.
type scope struct {
	s      *Service
	db     *store.DB
	domain string
}

// errNoService means that the Cahut was opened without the e-mail of its
// server, which only happens where nothing sends e-mail.
var errNoService = errors.New("mail: this Cahut was opened without the e-mail of its server")

// RegisterCahut registers, in the registry r of the Cahut at domain whose
// database is db, the operations that set its relay: set_mail_relay,
// get_mail_relay and send_test_mail. s may be nil, where the Cahut sends
// no e-mail.
func (s *Service) RegisterCahut(r *operation.Registry, db *store.DB, domain string) {
	sc := &scope{s: s, db: db, domain: domain}
	operation.Register(r, &operation.Op[SetMailRelayInput, Delivery]{
		Name:  "set_mail_relay",
		Title: "Set the e-mail relay",
		Description: "Set the SMTP relay that sends the e-mail of this Cahut, such as activation links, from an address of its own. " +
			"The relay is at a public address and speaks TLS. Cahut connects to it and signs in before it keeps it. " +
			"An empty host removes it: the Cahut then sends no e-mail, and its console offers nothing that would. " +
			"For the administrators of the Cahut.",
		Level:      operation.CahutAdmin,
		PeopleOnly: true,
		Prepare:    sc.prepareSet,
		Run:        sc.set,
	})
	operation.Register(r, &operation.Op[GetMailRelayInput, Delivery]{
		Name:  "get_mail_relay",
		Title: "Describe the e-mail relay",
		Description: "Describe the SMTP relay of this Cahut, without its password, or say that it has none and sends no e-mail; " +
			"and when a message last went out, or last failed and why. For the administrators of the Cahut.",
		Level:      operation.CahutAdmin,
		PeopleOnly: true,
		ReadOnly:   true,
		Run:        sc.get,
	})
	operation.Register(r, &operation.Op[SendTestMailInput, TestSent]{
		Name:  "send_test_mail",
		Title: "Send a test e-mail",
		Description: "Send a test message through the relay of this Cahut, and say whether the relay took it. " +
			"For the administrators of the Cahut.",
		Level:      operation.CahutAdmin,
		PeopleOnly: true,
		Prepare:    sc.prepareTest,
		Run:        sc.test,
	})
}

// prepared is what set_mail_relay prepares, before its transaction: the
// relay to keep, or nil to remove it, once it answered.
type prepared struct {
	relay *stored
	kept  bool // the password kept stays
}

func (sc *scope) prepareSet(ctx context.Context, c *operation.Call, in SetMailRelayInput) error {
	if sc.s == nil {
		return errNoService
	}
	if strings.TrimSpace(in.Host) == "" {
		c.Keep(prepared{})
		return nil
	}
	r := Relay{Host: in.Host, Port: in.Port, Security: in.Security, Username: in.Username, Password: in.Password, From: in.From, Name: in.Name}
	c.Log("mail relay", strings.TrimSpace(in.Host))
	if err := check(&r); err != nil {
		return operation.Invalidf("%v", err)
	}
	if !sc.s.dialAnywhere {
		if ip, err := netip.ParseAddr(r.Host); r.Host == "localhost" || err == nil && !address.Public(ip) {
			return operation.Invalidf("%s is not a public address: the relay of a Cahut is on the internet, "+
				"such as the SMTP relay of your provider of e-mail", r.Host)
		}
	}
	var cur *stored
	err := c.Read(ctx, func(tx *sql.Tx) error {
		var err error
		cur, err = load(ctx, tx)
		return err
	})
	if err != nil {
		return err
	}
	kept := false
	if r.Username != "" && r.Password == "" {
		if cur == nil || cur.sealed == nil || cur.Host != r.Host || cur.Username != r.Username {
			return operation.Invalidf("give the password of %s at the relay: the password kept stays only when the host and the user name do not change", r.Username)
		}
		if r.Password, err = sc.s.unseal(cur.sealed, cur.Host, cur.Username); err != nil {
			return operation.Invalidf("%v", err)
		}
		kept = true
	}
	if err := sc.s.try(ctx, &r); err != nil {
		return operation.Invalidf("the relay did not take this setting: %v", err)
	}
	st := &stored{Relay: r}
	if r.Username != "" {
		if st.sealed, err = sc.s.seal(r.Password, r.Host, r.Username); err != nil {
			return err
		}
	}
	st.Password = ""
	c.Keep(prepared{relay: st, kept: kept})
	return nil
}

func (sc *scope) set(ctx context.Context, c *operation.Call, _ SetMailRelayInput) (Delivery, error) {
	p, ok := c.Kept().(prepared)
	if !ok {
		return Delivery{}, errors.New("mail: set_mail_relay ran without its Prepare")
	}
	if p.relay == nil {
		c.Log("mail relay", "removed")
		if err := remove(ctx, c.Tx); err != nil {
			return Delivery{}, err
		}
		return sc.view(ctx, c.Tx)
	}
	r := p.relay
	r.SetAt = c.Now
	detail := fmt.Sprintf("%s:%d %s, from %s", r.Host, r.Port, r.Security, r.From)
	switch {
	case r.Username == "":
		detail += ", without authentication"
	case p.kept:
		detail += ", as " + r.Username + ", password kept"
	default:
		detail += ", as " + r.Username + ", password set"
	}
	c.Log("mail relay", detail)
	if err := save(ctx, c.Tx, r); err != nil {
		return Delivery{}, err
	}
	return sc.view(ctx, c.Tx)
}

func (sc *scope) get(ctx context.Context, c *operation.Call, _ GetMailRelayInput) (Delivery, error) {
	if sc.s == nil {
		return Delivery{}, errNoService
	}
	return sc.view(ctx, c.Tx)
}

// view describes the relay of the Cahut, as tx reads it.
func (sc *scope) view(ctx context.Context, tx *sql.Tx) (Delivery, error) {
	own, err := load(ctx, tx)
	if err != nil {
		return Delivery{}, err
	}
	out := Delivery{Status: sc.s.statusOf(sc.domain), Sends: sc.s.cfg.Dev}
	if own != nil {
		out.Relay = &RelayInfo{
			Host: own.Host, Port: own.Port, Security: own.Security, Username: own.Username, Password: own.sealed != nil,
			From: own.From, Name: own.Name, SetAt: own.SetAt,
		}
		out.Sends = true
	}
	return out, nil
}

func (sc *scope) prepareTest(ctx context.Context, c *operation.Call, in SendTestMailInput) error {
	if sc.s == nil {
		return errNoService
	}
	to, err := sc.recipient(ctx, c, in.To)
	if err != nil {
		return err
	}
	c.Log("mail relay", "test to "+to)
	from := From{DB: sc.db, Domain: sc.domain, Name: sc.domain}
	r, err := sc.s.relayOf(ctx, from)
	if err != nil {
		return operation.Invalidf("%v", err)
	}
	if r == nil {
		return operation.Invalidf("%v", ErrNoRelay)
	}
	l := Letter{
		To:      to,
		Subject: "A test from " + from.Name,
		Lines:   []string{"This message tests the e-mail of the Cahut at " + sc.domain + "."},
		After:   []string{"It reached you: there is nothing else to do."},
	}
	msg, err := build(r, from, l, c.Now)
	if err != nil {
		return err
	}
	err = sc.s.deliver(ctx, r, to, msg)
	sc.s.record(sc.domain, err)
	if err != nil {
		return operation.Invalidf("the relay did not take the test: %v", err)
	}
	c.Keep(TestSent{To: to})
	return nil
}

// recipient returns where the test of the call c goes: to the address of
// the person who calls, or to the address given, on the command line and
// for an administrator of the server, who has none.
func (sc *scope) recipient(ctx context.Context, c *operation.Call, given string) (string, error) {
	if c.Actor.CommandLine() || c.Actor.UserID == 0 {
		to, err := Address(given)
		if err != nil {
			return "", operation.Invalidf("to: %v", err)
		}
		return to, nil
	}
	if strings.TrimSpace(given) != "" {
		return "", operation.Invalidf("a test goes to your own e-mail address: leave out to")
	}
	var email sql.NullString
	err := c.Read(ctx, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, "SELECT email FROM users WHERE id = ?", c.Actor.UserID).Scan(&email)
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if email.String == "" {
		return "", operation.Invalidf("your account has no e-mail address for the test to go to: add one on the page of your account")
	}
	return email.String, nil
}

func (sc *scope) test(_ context.Context, c *operation.Call, _ SendTestMailInput) (TestSent, error) {
	sent, ok := c.Kept().(TestSent)
	if !ok {
		return TestSent{}, errors.New("mail: send_test_mail ran without its Prepare")
	}
	c.Log("mail relay", "test to "+sent.To)
	return sent, nil
}

// Sends reports, in the transaction tx of a Cahut, whether the Cahut sends
// e-mail: when it has a relay, and in development always, since messages
// without a relay go to disk. A nil s sends none.
func (s *Service) Sends(ctx context.Context, tx querier) (bool, error) {
	if s == nil {
		return false, nil
	}
	if s.cfg.Dev {
		return true, nil
	}
	r, err := load(ctx, tx)
	return r != nil, err
}
