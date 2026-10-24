package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strconv"
	"time"

	"cahut.cloud/internal/address"
)

// timeout bounds a whole conversation with a relay.
const timeout = time.Minute

// greeting bounds the wait for the greeting of a relay. One that speaks TLS
// from the start says nothing to a client that waits for its greeting.
var greeting = 10 * time.Second

// session connects to the relay r, at a public address, with TLS, and signs
// in when r has a user name.
func (s *Service) session(ctx context.Context, r *Relay) (*smtp.Client, error) {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	if !s.dialAnywhere {
		dialer.Control = address.DialPublic
	}
	addr := net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	hello := time.Now().Add(greeting)
	if deadline.Before(hello) {
		hello = deadline
	}
	_ = conn.SetDeadline(hello)
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()

	conf := &tls.Config{ServerName: r.Host, RootCAs: s.roots, MinVersion: tls.VersionTLS12}
	if r.Security == TLS {
		tc := tls.Client(conn, conf)
		if err := tc.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, fmt.Errorf("TLS with %s: %w", addr, err)
		}
		conn = tc
	}
	c, err := smtp.NewClient(conn, r.Host)
	if err != nil {
		conn.Close()
		if r.Security == StartTLS {
			return nil, fmt.Errorf("%s does not answer as an SMTP relay (%w): if it speaks TLS from the start, often on port 465, use security tls", addr, err)
		}
		return nil, fmt.Errorf("%s does not answer as an SMTP relay: %w", addr, err)
	}
	_ = conn.SetDeadline(deadline)
	fail := func(err error) (*smtp.Client, error) {
		c.Close()
		return nil, err
	}
	if err := c.Hello(s.helo); err != nil {
		return fail(fmt.Errorf("greet %s: %w", addr, err))
	}
	if r.Security == StartTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fail(fmt.Errorf("%s does not offer STARTTLS: use security tls if it speaks TLS from the start, often on port 465", addr))
		}
		if err := c.StartTLS(conf); err != nil {
			return fail(fmt.Errorf("STARTTLS with %s: %w", addr, err))
		}
	}
	if r.Username != "" {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fail(fmt.Errorf("%s takes no authentication: leave out the user name and the password", addr))
		}
		if err := c.Auth(smtp.PlainAuth("", r.Username, r.Password, r.Host)); err != nil {
			return fail(fmt.Errorf("sign in to %s as %s: %w", addr, r.Username, err))
		}
	}
	return c, nil
}

// try connects to the relay r and signs in, as set_mail_relay does before
// it keeps a relay.
func (s *Service) try(ctx context.Context, r *Relay) error {
	c, err := s.session(ctx, r)
	if err != nil {
		return err
	}
	_ = c.Quit()
	return nil
}

// deliver hands msg for the address to to the relay r.
func (s *Service) deliver(ctx context.Context, r *Relay, to string, msg []byte) error {
	c, err := s.session(ctx, r)
	if err != nil {
		return err
	}
	defer c.Close()
	if err := c.Mail(r.From); err != nil {
		return fmt.Errorf("the relay refuses mail from %s: %w", r.From, err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("the relay refuses mail to %s: %w", to, err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("the relay refuses the message: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("send the message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("the relay refuses the message: %w", err)
	}
	_ = c.Quit()
	return nil
}

// temporary reports whether err may pass if the message is tried again
// later: the relay answered with a code of 4xx, or could not be reached.
// A code of 5xx is final.
func temporary(err error) bool {
	var te *textproto.Error
	if errors.As(err, &te) {
		return te.Code >= 400 && te.Code < 500
	}
	return true
}
