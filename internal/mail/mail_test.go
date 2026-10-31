package mail_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	cmail "cahut.cloud/internal/mail"
	"cahut.cloud/internal/mail/mailtest"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

var cli = operation.Actor{Kind: operation.Admin, Via: "cli"}

// fixture is a server with one Cahut, example.com.
type fixture struct {
	dir   string
	srv   *server.Server
	cahut *cahut.Cahut
}

func setup(t *testing.T, cfg server.Config) *fixture {
	t.Helper()
	ctx := context.Background()
	cfg.DataDir = t.TempDir()
	srv, err := server.Open(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { srv.Close() })
	if _, err := operation.Do(ctx, srv.Registry, srv.AddDomain, cli, server.AddDomainInput{Domain: "example.com"}); err != nil {
		t.Fatal(err)
	}
	c, err := srv.OpenCahut(ctx, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &fixture{dir: cfg.DataDir, srv: srv, cahut: c}
}

// call runs the operation called name in r as the command line, with in.
func call[Out any](t *testing.T, r *operation.Registry, name string, in any) (Out, error) {
	t.Helper()
	var zero Out
	args, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := r.Invoke(context.Background(), name, cli, args)
	if err != nil {
		return zero, err
	}
	return out.(Out), nil
}

func relayInput(r *mailtest.Relay, security string) cmail.SetMailRelayInput {
	return cmail.SetMailRelayInput{Host: r.Host, Port: r.Port, Security: security, Username: "cahut", Password: "s3cret pass", From: "cahut@example.net"}
}

// cahutRelay starts a relay in memory, as the relay of the Cahut of f,
// which reaches it on this machine as tests may. The relay offers STARTTLS.
func cahutRelay(t *testing.T, f *fixture, opts mailtest.Options) *mailtest.Relay {
	t.Helper()
	relay := mailtest.Start(t, opts)
	f.srv.Mail.UseRoots(relay.Roots)
	f.srv.Mail.DialAnywhere()
	in := relayInput(relay, cmail.StartTLS)
	if opts.User == "" {
		in.Username, in.Password = "", ""
	}
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", in); err != nil {
		t.Fatal(err)
	}
	return relay
}

func TestRelay(t *testing.T) {
	f := setup(t, server.Config{})
	relay := mailtest.Start(t, mailtest.Options{User: "cahut", Password: "s3cret pass"})
	f.srv.Mail.UseRoots(relay.Roots)
	ctx := context.Background()

	// Without a relay, the Cahut sends nothing, and nothing falls back to
	// the server, which has no relay.
	about, err := operation.Do(ctx, f.cahut.Registry, f.cahut.About, cli, cahut.GetCahutInput{})
	if err != nil || about.SendsMail {
		t.Fatalf("get_cahut without a relay: %+v %v", about, err)
	}
	if got, err := call[cmail.Delivery](t, f.cahut.Registry, "get_mail_relay", cmail.GetMailRelayInput{}); err != nil || got.Relay != nil || got.Sends {
		t.Fatalf("get_mail_relay without a relay: %+v %v", got, err)
	}
	if _, err := call[cmail.TestSent](t, f.cahut.Registry, "send_test_mail", cmail.SendTestMailInput{To: "jane@example.org"}); !errors.Is(err, operation.ErrInvalid) ||
		!strings.Contains(err.Error(), "sends no e-mail") {
		t.Fatalf("send_test_mail without a relay: %v", err)
	}
	if err := f.srv.Mail.Post(ctx, cmail.From{DB: f.cahut.DB, Domain: "example.com"}, cmail.Letter{To: "jane@example.org", Subject: "Hello"}); !errors.Is(err, cmail.ErrNoRelay) {
		t.Errorf("Post without a relay: %v", err)
	}
	if _, ok := f.srv.Registry.Lookup("set_mail_relay"); ok {
		t.Error("the server has an operation set_mail_relay")
	}
	var tables int
	if err := f.srv.DB.Read.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name IN ('mail_relay', 'notices_sent', 'admin_email_changes')").Scan(&tables); err != nil || tables != 0 {
		t.Errorf("server.db keeps %d tables of e-mail: %v", tables, err)
	}

	// The relay of a Cahut is on the internet, and speaks TLS.
	for _, host := range []string{"127.0.0.1", "localhost", "10.0.0.1", "::1"} {
		in := relayInput(relay, cmail.StartTLS)
		in.Host = host
		if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", in); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "public") {
			t.Errorf("set_mail_relay at %s: %v", host, err)
		}
	}
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", relayInput(relay, "none")); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "TLS") {
		t.Errorf("set_mail_relay without TLS: %v", err)
	}
	f.srv.Mail.DialAnywhere()

	// A wrong password is refused, with the answer of the relay, and nothing is kept.
	in := relayInput(relay, cmail.StartTLS)
	in.Password = "wrong"
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", in); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "535") {
		t.Fatalf("set_mail_relay with a wrong password: %v", err)
	}
	if got, _ := call[cmail.Delivery](t, f.cahut.Registry, "get_mail_relay", cmail.GetMailRelayInput{}); got.Relay != nil {
		t.Fatalf("a refused relay was kept: %+v", got.Relay)
	}

	set, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", relayInput(relay, cmail.StartTLS))
	if err != nil {
		t.Fatalf("set_mail_relay: %v", err)
	}
	if r := set.Relay; r == nil || r.Host != "127.0.0.1" || r.Security != "starttls" || !r.Password || r.From != "cahut@example.net" || !set.Sends {
		t.Fatalf("set_mail_relay = %+v", set)
	}
	out, _ := json.Marshal(set)
	if strings.Contains(string(out), "s3cret") {
		t.Errorf("the output gives the password back: %s", out)
	}

	// The password is sealed in the database, under a key of the server
	// that only its user reads.
	var sealed []byte
	if err := f.cahut.DB.Read.QueryRow("SELECT password FROM mail_relay").Scan(&sealed); err != nil || bytes.Contains(sealed, []byte("s3cret")) {
		t.Errorf("the password in the database: %q %v", sealed, err)
	}
	if fi, err := os.Stat(filepath.Join(f.dir, "mail.key")); err != nil || fi.Mode().Perm() != 0o600 || fi.Size() != 32 {
		t.Errorf("mail.key: %v %v", fi, err)
	}

	about, err = operation.Do(ctx, f.cahut.Registry, f.cahut.About, cli, cahut.GetCahutInput{})
	if err != nil || !about.SendsMail {
		t.Errorf("get_cahut with a relay: %+v %v", about, err)
	}
	sent, err := call[cmail.TestSent](t, f.cahut.Registry, "send_test_mail", cmail.SendTestMailInput{To: "Jane@Example.ORG"})
	if err != nil || sent.To != "Jane@example.org" {
		t.Fatalf("send_test_mail: %+v %v", sent, err)
	}
	msgs := relay.Wait(t, 1)
	m, err := mail.ReadMessage(msgs[0].Reader())
	if err != nil {
		t.Fatal(err)
	}
	if msgs[0].From != "cahut@example.net" || len(msgs[0].To) != 1 || msgs[0].To[0] != "Jane@example.org" ||
		m.Header.Get("Subject") != "A test from example.com" || m.Header.Get("Auto-Submitted") != "auto-generated" {
		t.Errorf("the test: %+v %v", msgs[0], m.Header)
	}
	if from, err := mail.ParseAddress(m.Header.Get("From")); err != nil || from.Name != "example.com" || from.Address != "cahut@example.net" {
		t.Errorf("From = %q", m.Header.Get("From"))
	}
	if got, _ := call[cmail.Delivery](t, f.cahut.Registry, "get_mail_relay", cmail.GetMailRelayInput{}); got.LastSent == nil || got.LastError != "" {
		t.Errorf("status after the test: %+v", got.Status)
	}

	// Without the password, the one kept stays, while the host and the
	// user name do not change.
	in = relayInput(relay, cmail.StartTLS)
	in.Password, in.Name = "", "Example"
	if set, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", in); err != nil || set.Relay.Name != "Example" || !set.Relay.Password {
		t.Errorf("set_mail_relay keeping the password: %+v %v", set, err)
	}
	in.Username = "other"
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", in); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "give the password") {
		t.Errorf("set_mail_relay with another user and no password: %v", err)
	}

	// The Logbook says what changed, never the password.
	entries, err := logbook.List(ctx, f.cahut.DB.Read, 20)
	if err != nil {
		t.Fatal(err)
	}
	var details []string
	for _, e := range entries {
		if e.Operation == "set_mail_relay" && e.Error == "" {
			details = append(details, e.Detail)
		}
		if strings.Contains(e.Detail+e.Error, "s3cret") || strings.Contains(e.Detail+e.Error, "wrong") {
			t.Errorf("a password in the Logbook: %+v", e)
		}
	}
	if len(details) != 2 || !strings.Contains(details[0], "password kept") || !strings.Contains(details[1], "password set") {
		t.Errorf("the Logbook of set_mail_relay: %q", details)
	}

	// Removing the relay stops the e-mail of the Cahut.
	if set, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", cmail.SetMailRelayInput{}); err != nil || set.Relay != nil || set.Sends {
		t.Errorf("set_mail_relay with an empty host: %+v %v", set, err)
	}
	if about, _ := operation.Do(ctx, f.cahut.Registry, f.cahut.About, cli, cahut.GetCahutInput{}); about.SendsMail {
		t.Error("get_cahut says the Cahut sends e-mail after its relay is removed")
	}
}

func TestRelayOfTLS(t *testing.T) {
	f := setup(t, server.Config{})
	relay := mailtest.Start(t, mailtest.Options{TLS: true, User: "cahut", Password: "s3cret pass"})
	f.srv.Mail.UseRoots(relay.Roots)
	f.srv.Mail.DialAnywhere()
	cmail.SetGreeting(300 * time.Millisecond)
	t.Cleanup(func() { cmail.SetGreeting(10 * time.Second) })
	// starttls by default, which a relay that speaks TLS at once does not
	// answer.
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", relayInput(relay, "")); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "use security tls") {
		t.Errorf("set_mail_relay, STARTTLS with a relay of TLS: %v", err)
	}
	set, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", relayInput(relay, cmail.TLS))
	if err != nil || set.Relay == nil || set.Relay.Security != "tls" {
		t.Fatalf("set_mail_relay: %+v %v", set, err)
	}

	// A message of the Cahut goes through its relay, from its address.
	ctx := context.Background()
	err = f.srv.Mail.Post(ctx, cmail.From{DB: f.cahut.DB, Domain: "example.com", Name: "Example"}, cmail.Letter{
		To: "jane@example.org", Subject: "Hello", Lines: []string{"Hello Jane."}, Action: &cmail.Action{Label: "Open", URL: "https://example.com/a/b/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	msgs := relay.Wait(t, 1)
	if msgs[0].From != "cahut@example.net" || !bytes.Contains(msgs[0].Data, []byte("From: \"Example\" <cahut@example.net>")) {
		t.Errorf("the message of the Cahut: %s", msgs[0].Data)
	}
}

func TestRelayWithoutStartTLS(t *testing.T) {
	f := setup(t, server.Config{})
	relay := mailtest.Start(t, mailtest.Options{NoStartTLS: true, User: "cahut", Password: "s3cret pass"})
	f.srv.Mail.DialAnywhere()
	if _, err := call[cmail.Delivery](t, f.cahut.Registry, "set_mail_relay", relayInput(relay, cmail.StartTLS)); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "STARTTLS") {
		t.Errorf("set_mail_relay with STARTTLS, which the relay does not offer: %v", err)
	}
}

func TestQueueTriesAgain(t *testing.T) {
	f := setup(t, server.Config{})
	relay := cahutRelay(t, f, mailtest.Options{})
	cmail.SetRetries(f.srv.Mail, 50*time.Millisecond, 50*time.Millisecond)
	ctx := context.Background()
	from := cmail.From{DB: f.cahut.DB, Domain: "example.com"}

	// A relay that answers 4xx gets the message again later.
	relay.Refuse("451 4.3.0 Try again later")
	if err := f.srv.Mail.Post(ctx, from, cmail.Letter{To: "jane@example.org", Subject: "Once"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		got, _ := call[cmail.Delivery](t, f.cahut.Registry, "get_mail_relay", cmail.GetMailRelayInput{})
		return strings.Contains(got.LastError, "451")
	})
	relay.Refuse("")
	relay.Wait(t, 1)

	// A relay that answers 5xx refuses it for good.
	relay.Refuse("550 5.7.1 Relaying denied")
	if err := f.srv.Mail.Post(ctx, from, cmail.Letter{To: "jane@example.org", Subject: "Never"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		got, _ := call[cmail.Delivery](t, f.cahut.Registry, "get_mail_relay", cmail.GetMailRelayInput{})
		return strings.Contains(got.LastError, "550")
	})
	relay.Refuse("")
	time.Sleep(200 * time.Millisecond)
	if n := len(relay.Messages()); n != 1 {
		t.Errorf("the relay got %d messages, want 1: the refused one is not tried again", n)
	}

	// Close sends what waits in the queue.
	if err := f.srv.Mail.Post(ctx, from, cmail.Letter{To: "jane@example.org", Subject: "Last"}); err != nil {
		t.Fatal(err)
	}
	if err := f.srv.Mail.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if n := len(relay.Messages()); n != 2 {
		t.Errorf("after Close, the relay got %d messages, want 2", n)
	}
	if err := f.srv.Mail.Post(ctx, from, cmail.Letter{To: "jane@example.org", Subject: "Late"}); err == nil {
		t.Error("Post after Close: no error")
	}
}

func waitFor(t *testing.T, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestDevelopmentWritesMessages(t *testing.T) {
	f := setup(t, server.Config{MailDev: true})
	err := f.srv.Mail.Post(context.Background(), cmail.From{DB: f.cahut.DB, Domain: "example.com"}, cmail.Letter{
		To: "jane@example.org", Subject: "Choose your password", Action: &cmail.Action{Label: "Choose", URL: "http://example.com:8080/console/activate#code"},
	})
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(f.dir, "mail", "*.eml"))
	if err != nil || len(files) != 1 {
		t.Fatalf("messages written: %v %v", files, err)
	}
	data, _ := os.ReadFile(files[0])
	if m, err := mail.ReadMessage(bytes.NewReader(data)); err != nil || m.Header.Get("Subject") != "Choose your password" {
		t.Errorf("the message written: %v", err)
	}
}

func TestLetterLayout(t *testing.T) {
	r := &cmail.Relay{From: "cahut@example.net"}
	from := cmail.From{Domain: "example.com", Name: "Équipe <Example>"}
	l := cmail.Letter{
		To:      "jane@example.org",
		Subject: "Olivier invites you\r\nBcc: evil@example.net",
		Lines:   []string{"Olivier Durand invites you to read Q3 review."},
		Quote:   &cmail.Quote{By: "Olivier Durand", Text: "Have a look <script>alert(1)</script>\nbefore **Friday**, at \\*noon\\*.\n\n- the *deck*\n- the notes"},
		Action:  &cmail.Action{Label: "Open Q3 review", URL: "https://example.com/console/invitation#abc"},
		After:   []string{"This link works for 30 days."},
	}
	data, err := cmail.Build(r, from, l, time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	m, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	h := m.Header
	if h.Get("Bcc") != "" || h.Get("Subject") != "Olivier invites you Bcc: evil@example.net" {
		t.Errorf("a line break in the subject made a header: %v", h)
	}
	if from, err := (&mime.WordDecoder{}).DecodeHeader(h.Get("From")); err != nil || !strings.Contains(from, "Équipe <Example>") {
		t.Errorf("From = %q %v", h.Get("From"), err)
	}
	if !strings.HasSuffix(h.Get("Message-ID"), "@example.net>") || h.Get("Date") != "Sat, 24 Oct 2026 12:00:00 +0000" {
		t.Errorf("headers: %v", h)
	}
	_, params, err := mime.ParseMediaType(h.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	related := multipart.NewReader(m.Body, params["boundary"])
	alt, err := related.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	_, altParams, _ := mime.ParseMediaType(alt.Header.Get("Content-Type"))
	parts := multipart.NewReader(alt, altParams["boundary"])
	var text, html string
	for {
		p, err := parts.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(p) // NextPart decodes quoted-printable
		switch typ, _, _ := mime.ParseMediaType(p.Header.Get("Content-Type")); typ {
		case "text/plain":
			text = string(b)
		case "text/html":
			html = string(b)
		}
	}
	for _, want := range []string{"Olivier Durand invites you", "Olivier Durand wrote:", "> before **Friday**, at *noon*.", "> - the *deck*",
		"Open Q3 review: https://example.com/console/invitation#abc", "This link works for 30 days.", "Sent by the Cahut at example.com."} {
		if !strings.Contains(text, want) {
			t.Errorf("the text lacks %q:\n%s", want, text)
		}
	}
	for _, want := range []string{`href="https://example.com/console/invitation#abc"`, "&lt;script&gt;", "<br>before <strong>Friday</strong>, at *noon*.",
		"<li>the <em>deck</em></li><li>the notes</li></ul>", "cid:logo@cahut", "#1F2A33"} {
		if !strings.Contains(html, want) {
			t.Errorf("the HTML lacks %q", want)
		}
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "{ink}") {
		t.Errorf("the HTML:\n%s", html)
	}
	logo, err := related.NextPart()
	if err != nil || logo.Header.Get("Content-ID") != "<logo@cahut>" || logo.Header.Get("Content-Type") != "image/png" {
		t.Errorf("the logo: %v %v", logo, err)
	}

	// An address that is not one, or a link that is not one, goes nowhere.
	if _, err := cmail.Build(r, from, cmail.Letter{To: "jane@example.org, evil@example.net"}, time.Now()); err == nil {
		t.Error("two recipients: no error")
	}
	data, _ = cmail.Build(r, from, cmail.Letter{To: "jane@example.org", Action: &cmail.Action{Label: "Go", URL: "javascript:alert(1)"}}, time.Now())
	if bytes.Contains(data, []byte("javascript:")) && bytes.Contains(data, []byte(`href=3D"javascript`)) {
		t.Error("a javascript: link in the HTML")
	}
}

// TestBrand checks that e-mail takes the colors and the logo of the brand
// guidelines as they are.
func TestBrand(t *testing.T) {
	colors, logo := cmail.Brand()
	raw, err := os.ReadFile("../../docs/charte-graphique/tokens/cahut-tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens map[string]any
	if err := json.Unmarshal(raw, &tokens); err != nil {
		t.Fatal(err)
	}
	token := func(path ...string) string {
		var v any = tokens
		for _, p := range path {
			v = v.(map[string]any)[p]
		}
		return v.(map[string]any)["$value"].(string)
	}
	for name, path := range map[string][]string{
		"paper": {"color", "brand", "paper"},
		"ink":   {"color", "brand", "ink"},
		"muted": {"color", "ink", "600"},
		"link":  {"color", "sky", "700"},
		"rule":  {"color", "brand", "light-grey"},
		"white": {"color", "neutral", "white"},
	} {
		if want := token(path...); !strings.EqualFold(colors[name], want) {
			t.Errorf("%s = %s, the token %s is %s", name, colors[name], strings.Join(path, "."), want)
		}
	}
	if len(colors) != 6 {
		t.Errorf("e-mail takes %d colors: check each against a token", len(colors))
	}
	want, err := os.ReadFile("../../docs/charte-graphique/logos/png/cahut-logo-horizontal-800px.png")
	if err != nil || !bytes.Equal(logo, want) {
		t.Errorf("internal/mail/logo.png differs from cahut-logo-horizontal-800px.png of the brand guidelines: copy it again (%v)", err)
	}
}

func TestSealedPasswords(t *testing.T) {
	a := cmail.New(cmail.Config{DataDir: t.TempDir()})
	b := cmail.New(cmail.Config{DataDir: t.TempDir()})
	t.Cleanup(func() {
		a.Close(context.Background())
		b.Close(context.Background())
	})
	sealed, err := cmail.Seal(a, "s3cret", "smtp.example.com", "cahut")
	if err != nil {
		t.Fatal(err)
	}
	if pw, err := cmail.Unseal(a, sealed, "smtp.example.com", "cahut"); err != nil || pw != "s3cret" {
		t.Errorf("Unseal = %q %v", pw, err)
	}
	// Bound to its relay: moved to another host in the database, it does not open.
	if _, err := cmail.Unseal(a, sealed, "smtp.evil.example", "cahut"); err == nil {
		t.Error("a password opened for another host")
	}
	// Another server, with another key, cannot read it.
	if _, err := cmail.Unseal(b, sealed, "smtp.example.com", "cahut"); err == nil || !strings.Contains(err.Error(), "set the relay again") {
		t.Errorf("Unseal on another server: %v", err)
	}
}

func TestAddress(t *testing.T) {
	for in, want := range map[string]string{
		"jane@example.org":      "jane@example.org",
		" Jane@Example.ORG ":    "Jane@example.org",
		"jane.doe+x@sub.ex.org": "jane.doe+x@sub.ex.org",
	} {
		if got, err := cmail.Address(in); err != nil || got != want {
			t.Errorf("Address(%q) = %q %v, want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "jane", "Jane <jane@example.org>", "jane@example.org, joe@example.org", "jane@example.org\r\nBcc: x@y.z", "<jane@example.org>"} {
		if got, err := cmail.Address(in); err == nil {
			t.Errorf("Address(%q) = %q, want an error", in, got)
		}
	}
}

func TestTestGoesToThePersonWhoAsks(t *testing.T) {
	f := setup(t, server.Config{})
	ctx := context.Background()
	relay := cahutRelay(t, f, mailtest.Options{})
	u, err := operation.Do(ctx, f.cahut.Registry, account.CreateUser, cli, account.CreateUserInput{Handle: "olivier", Name: "Olivier", Admin: true, Email: "olivier@example.org"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := operation.Do(ctx, f.cahut.Registry, account.CreateUser, cli, account.CreateUserInput{Handle: "ana", Name: "Ana", Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	as := func(id int64, handle string, to string) error {
		args, _ := json.Marshal(cmail.SendTestMailInput{To: to})
		_, err := f.cahut.Registry.Invoke(ctx, "send_test_mail", operation.Actor{Kind: operation.Person, Via: "api", UserID: id, Handle: handle}, args)
		return err
	}
	if err := as(u.ID, "olivier", "someone@example.net"); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("send_test_mail to another address: %v", err)
	}
	if err := as(other.ID, "ana", ""); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "no e-mail address") {
		t.Errorf("send_test_mail without an address: %v", err)
	}
	if err := as(u.ID, "olivier", ""); err != nil {
		t.Fatalf("send_test_mail: %v", err)
	}
	if m := relay.Wait(t, 1); m[0].To[0] != "olivier@example.org" {
		t.Errorf("the test went to %v", m[0].To)
	}
}

// TestDevelopmentSends checks that in development, without a relay, a
// Cahut sends e-mail, to disk, so that its flows can be tried.
func TestDevelopmentSends(t *testing.T) {
	f := setup(t, server.Config{MailDev: true})
	about, err := operation.Do(context.Background(), f.cahut.Registry, f.cahut.About, cli, cahut.GetCahutInput{})
	if err != nil || !about.SendsMail {
		t.Errorf("get_cahut in development: %+v %v", about, err)
	}
}

// TestTestFromTheCommandLine checks that the command line, which has no
// address, names where the test goes, and that an administrator of the
// server no longer sends it: the e-mail of a Cahut is for its
// administrators. See docs/adr/0058-a-console-for-the-server.md.
func TestTestFromTheCommandLine(t *testing.T) {
	f := setup(t, server.Config{})
	ctx := context.Background()
	relay := cahutRelay(t, f, mailtest.Options{})
	if _, err := operation.Do(ctx, f.srv.Registry, server.CreateAdmin, cli, server.CreateAdminInput{Handle: "root", Name: "Root"}); err != nil {
		t.Fatal(err)
	}
	root := operation.Actor{Kind: operation.Admin, Via: "api", Domain: "example.com", Handle: "root", Session: 1}
	if _, err := f.cahut.Registry.Invoke(ctx, "send_test_mail", root, []byte(`{"to":"root@example.org"}`)); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("send_test_mail by an administrator of the server: %v", err)
	}
	if _, err := f.cahut.Registry.Invoke(ctx, "send_test_mail", cli, []byte(`{}`)); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "to:") {
		t.Errorf("send_test_mail from the command line, without an address: %v", err)
	}
	if _, err := f.cahut.Registry.Invoke(ctx, "send_test_mail", cli, []byte(`{"to":"root@example.org"}`)); err != nil {
		t.Fatalf("send_test_mail from the command line: %v", err)
	}
	if m := relay.Wait(t, 1); m[0].To[0] != "root@example.org" {
		t.Errorf("the test went to %v", m[0].To)
	}
}
