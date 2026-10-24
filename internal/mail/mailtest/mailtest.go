// Package mailtest runs an SMTP relay in memory, for tests. It speaks enough
// SMTP for package mail: EHLO, STARTTLS or TLS from the start, AUTH PLAIN,
// MAIL, RCPT and DATA, and it keeps the messages it receives.
package mailtest

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Options say how a relay behaves.
type Options struct {
	// TLS makes the relay speak TLS from the start, as on port 465, instead
	// of offering STARTTLS.
	TLS bool
	// NoStartTLS makes the relay offer no STARTTLS, nor TLS.
	NoStartTLS bool
	// User and Password, when User is set, are the credentials the relay
	// asks for: it offers AUTH PLAIN, and refuses mail without it.
	User, Password string
}

// Message is a message the relay received.
type Message struct {
	From string
	To   []string
	Data []byte // as it came, with CRLF line endings and without the final dot
}

// Relay is a running relay.
type Relay struct {
	Host  string // 127.0.0.1
	Port  int
	Roots *x509.CertPool // trusts the certificate of the relay

	opts Options
	cert tls.Certificate
	ln   net.Listener

	mu     sync.Mutex
	msgs   []Message
	refuse string // the reply to MAIL, when set
	got    chan struct{}
}

// Start starts a relay on a loopback address, which stops with the test.
func Start(t testing.TB, opts Options) *Relay {
	t.Helper()
	cert, roots := certificate(t)
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	r := &Relay{Host: "127.0.0.1", Port: ln.Addr().(*net.TCPAddr).Port, Roots: roots, opts: opts, cert: cert, ln: ln, got: make(chan struct{}, 1000)}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.serve(conn)
			}()
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		wg.Wait()
	})
	return r
}

// Messages returns the messages received so far.
func (r *Relay) Messages() []Message {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Message(nil), r.msgs...)
}

// Wait waits until the relay has received n messages, and returns them. It
// fails the test after ten seconds.
func (r *Relay) Wait(t testing.TB, n int) []Message {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		if msgs := r.Messages(); len(msgs) >= n {
			return msgs
		}
		select {
		case <-r.got:
		case <-deadline:
			t.Fatalf("the relay received %d messages, not %d", len(r.Messages()), n)
		}
	}
}

// Refuse makes the relay answer MAIL with reply, such as
// "451 4.3.0 Try again later"; an empty reply accepts mail again.
func (r *Relay) Refuse(reply string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refuse = reply
}

func (r *Relay) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	secure := false
	if r.opts.TLS {
		conn = tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{r.cert}, MinVersion: tls.VersionTLS12})
		secure = true
	}
	tp := textproto.NewConn(conn)
	reply := func(lines ...string) bool {
		for _, l := range lines {
			if err := tp.PrintfLine("%s", l); err != nil {
				return false
			}
		}
		return true
	}
	if !reply("220 mailtest ESMTP") {
		return
	}
	authed := false
	var from string
	var to []string
	for {
		line, err := tp.ReadLine()
		if err != nil {
			return
		}
		verb, arg, _ := strings.Cut(line, " ")
		switch strings.ToUpper(verb) {
		case "EHLO", "HELO":
			ext := []string{"8BITMIME"}
			if !secure && !r.opts.NoStartTLS {
				ext = append(ext, "STARTTLS")
			}
			if r.opts.User != "" {
				ext = append(ext, "AUTH PLAIN")
			}
			lines := []string{"250-mailtest"}
			for i, e := range ext {
				sep := "-"
				if i == len(ext)-1 {
					sep = " "
				}
				lines = append(lines, "250"+sep+e)
			}
			if !reply(lines...) {
				return
			}
		case "STARTTLS":
			if secure || r.opts.NoStartTLS {
				reply("503 5.5.1 TLS already active, or not offered")
				continue
			}
			if !reply("220 2.0.0 Ready to start TLS") {
				return
			}
			tc := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{r.cert}, MinVersion: tls.VersionTLS12})
			if tc.HandshakeContext(context.Background()) != nil {
				return
			}
			conn, secure, authed = tc, true, false
			tp = textproto.NewConn(conn)
		case "AUTH":
			mech, initial, _ := strings.Cut(arg, " ")
			if r.opts.User == "" || !strings.EqualFold(mech, "PLAIN") {
				reply("504 5.5.4 Unrecognized authentication type")
				continue
			}
			raw, err := base64.StdEncoding.DecodeString(initial)
			parts := strings.Split(string(raw), "\x00")
			if err != nil || len(parts) != 3 || parts[1] != r.opts.User || parts[2] != r.opts.Password {
				reply("535 5.7.8 Authentication credentials invalid")
				continue
			}
			authed = true
			reply("235 2.7.0 Authentication successful")
		case "MAIL":
			r.mu.Lock()
			refuse := r.refuse
			r.mu.Unlock()
			switch {
			case r.opts.User != "" && !authed:
				reply("530 5.7.0 Authentication required")
			case refuse != "":
				reply(refuse)
			default:
				from, to = address(arg), nil
				reply("250 2.1.0 OK")
			}
		case "RCPT":
			to = append(to, address(arg))
			reply("250 2.1.5 OK")
		case "DATA":
			if !reply("354 Go ahead") {
				return
			}
			data, err := tp.ReadDotBytes()
			if err != nil {
				return
			}
			r.mu.Lock()
			// ReadDotBytes turns CRLF into LF: put them back, as they came.
			r.msgs = append(r.msgs, Message{From: from, To: to, Data: []byte(strings.ReplaceAll(string(data), "\n", "\r\n"))})
			r.mu.Unlock()
			r.got <- struct{}{}
			reply("250 2.0.0 OK queued")
		case "RSET", "NOOP":
			reply("250 2.0.0 OK")
		case "QUIT":
			reply("221 2.0.0 Bye")
			return
		default:
			reply("502 5.5.2 Command not implemented")
		}
	}
}

// address returns the address in an argument such as FROM:<a@b> BODY=8BITMIME.
func address(arg string) string {
	_, rest, _ := strings.Cut(arg, "<")
	addr, _, _ := strings.Cut(rest, ">")
	return addr
}

// certificate makes a certificate for 127.0.0.1 and localhost, and a pool
// that trusts it.
func certificate(t testing.TB) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mailtest"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         true,
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},

		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: cert}, roots
}

// Addr returns the address of the relay, host:port.
func (r *Relay) Addr() string {
	return net.JoinHostPort(r.Host, strconv.Itoa(r.Port))
}

// Reader returns a reader of the data of m, for net/mail.
func (m Message) Reader() *bufio.Reader {
	return bufio.NewReader(strings.NewReader(string(m.Data)))
}
