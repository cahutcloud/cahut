package mail

import (
	"context"
	"errors"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Outbox is what the operations of a Cahut, or of the server, need to send
// e-mail: the service of the server, whose e-mail it is, and where the links
// of its messages lead.
type Outbox struct {
	Service *Service
	DB      *store.DB // of the Cahut; nil for the server
	Domain  string    // of the Cahut; empty for the server
	// Name returns the name of the Cahut, which its messages come from.
	Name func(ctx context.Context) (string, error)
	// Base is the address of the Cahut, such as https://example.com/, under
	// which its Rooms and Windows open.
	Base string
	// Console is the address of the console of the Cahut, such as
	// https://example.com/console/, under which links open.
	Console string
	// Consoles returns, for the server, the address of the console of the
	// Cahut at a domain, where its administrators sign in.
	Consoles func(domain string) string
}

type outboxKey struct{}

// Provide makes o available to the operations of the registry r of its
// Cahut, which OutboxOf returns.
func Provide(r *operation.Registry, o *Outbox) {
	r.Provide(outboxKey{}, o)
}

// OutboxOf returns the outbox of the Cahut of the call c, or nil where it
// sends no e-mail, as on a Cahut opened without its server.
func OutboxOf(c *operation.Call) *Outbox {
	o, _ := c.Value(outboxKey{}).(*Outbox)
	return o
}

// OutboxIn returns the outbox that the registry r provides, or nil, for
// what runs outside a call.
func OutboxIn(r *operation.Registry) *Outbox {
	o, _ := r.Value(outboxKey{}).(*Outbox)
	return o
}

// Sends reports, in the transaction of a call, whether the Cahut sends
// e-mail. A nil o sends none.
func (o *Outbox) Sends(ctx context.Context, tx querier) (bool, error) {
	if o == nil {
		return false, nil
	}
	return o.Service.Sends(ctx, tx)
}

// Post sends l once the call c is committed, so that a call that fails or
// knocks sends nothing. A message that cannot go out is logged by the
// service. A nil o sends nothing.
func (o *Outbox) Post(c *operation.Call, l Letter) {
	if o == nil {
		return
	}
	c.AfterCommit(func() { o.Send(l) })
}

// Send sends l now, for what runs outside a call or once it is committed.
// A message that cannot go out is logged by the service. A nil o sends
// nothing.
func (o *Outbox) Send(l Letter) {
	if o == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	name, err := o.Name(ctx)
	if err == nil {
		err = o.Service.Post(ctx, From{DB: o.DB, Domain: o.Domain, Name: name}, l)
	}
	if err != nil && !errors.Is(err, ErrNoRelay) {
		o.Service.log.Warn("post e-mail", "for", o.Domain, "error", err)
	}
}

// Where returns how a message names the Cahut of the outbox: its name, and
// where it is, such as "Acme, at cahut.acme.com", or "the Cahut at
// cahut.acme.com" when it has no name of its own.
func (o *Outbox) Where(ctx context.Context) (name, where string, err error) {
	name, err = o.Name(ctx)
	if err != nil {
		return "", "", err
	}
	if name == o.Domain {
		return name, "the Cahut at " + o.Domain, nil
	}
	return name, name + ", at " + o.Domain, nil
}

// Link returns the address of the page path of the console, such as
// "activate#<code>".
func (o *Outbox) Link(path string) string {
	return o.Console + path
}

// Day writes the time t for a message, such as "31 October 2026 at 14:05
// UTC".
func Day(t time.Time) string {
	return t.UTC().Format("2 January 2006 at 15:04 UTC")
}
