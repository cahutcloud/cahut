package cahut

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

// openWithMail opens a Cahut whose e-mail, in development, goes to files
// under the directory it returns.
func openWithMail(t *testing.T) (*limits, string) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	svc := mail.New(mail.Config{DataDir: dir, Dev: true})
	t.Cleanup(func() { svc.Close(ctx) })
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com", BaseURL: "https://cahut.example.com/",
		ConsoleURL: "https://cahut.example.com/console/", Mail: svc})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &limits{t: t, ctx: ctx, c: c, admin: operation.Actor{Kind: operation.Admin, Via: "cli"}}, filepath.Join(dir, "mail")
}

// letters returns the e-mail written under dir, in the order it went.
func letters(t *testing.T, dir string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.eml"))
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(files)
	var all []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		all = append(all, strings.ReplaceAll(string(b), "=\r\n", ""))
	}
	return all
}

// The administrators of the server suspend a Cahut with a reason, and
// resume it. Meanwhile, the registry refuses every call but theirs, and
// describe_sign_in. See docs/adr/0048-suspending-a-cahut.md.
func TestSuspendCahut(t *testing.T) {
	l, mailDir := openWithMail(t)
	no := false
	for _, p := range []struct {
		handle, email string
		admin         bool
	}{{"olivier", "olivier@example.com", true}, {"ana", "", true}, {"camille", "camille@example.com", false}} {
		_, err := operation.Do(l.ctx, l.c.Registry, account.CreateUser, l.admin,
			account.CreateUserInput{Handle: p.handle, Name: p.handle, Admin: p.admin, Email: p.email, SendActivationEmail: &no})
		if err != nil {
			t.Fatal(err)
		}
	}
	olivier := operation.Actor{Kind: operation.Person, Via: "api", UserID: 1, Handle: "olivier", Session: 1}
	camille := operation.Actor{Kind: operation.Person, Via: "api", UserID: 3, Handle: "camille", Session: 1}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: 1, Handle: "olivier", TokenID: 1, Client: "Claude"}
	service := operation.Actor{Kind: operation.Service, Via: "admin-api", Client: "service token account"}
	suspend := func(who operation.Actor, reason string) (settings.Settings, error) {
		return operation.Do(l.ctx, l.c.Registry, settings.Suspend, who, settings.SuspendCahutInput{Reason: reason})
	}

	// Only the administrators of the server suspend a Cahut, and no AI is
	// offered it.
	if _, err := suspend(olivier, "Unpaid invoice"); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("suspend_cahut by an administrator of the Cahut: %v", err)
	}
	for _, name := range []string{"suspend_cahut", "resume_cahut"} {
		info, _ := l.c.Registry.Lookup(name)
		if info.ForAIs(true, false) || info.Services {
			t.Errorf("%s is offered to AIs or services", name)
		}
	}
	// It takes a reason, which its administrators read.
	for _, reason := range []string{"", "   ", strings.Repeat("é", settings.MaxReason+1)} {
		if _, err := suspend(l.admin, reason); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("suspend_cahut with a reason of %d characters: %v", len(reason), err)
		}
	}
	if _, err := operation.Do(l.ctx, l.c.Registry, settings.Resume, l.admin, settings.ResumeCahutInput{}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("resume_cahut of an active Cahut: %v", err)
	}

	s, err := suspend(l.admin, "  Unpaid\ninvoice of October  ")
	if err != nil {
		t.Fatalf("suspend_cahut: %v", err)
	}
	if s.Suspension == nil || s.Suspension.Reason != "Unpaid invoice of October" || s.Suspension.Since.IsZero() {
		t.Errorf("suspend_cahut = %+v", s.Suspension)
	}
	if e := l.lastEntry(); e.Operation != "suspend_cahut" || e.Actor != "admin" || e.Detail != "Unpaid invoice of October" {
		t.Errorf("logbook = %+v", e)
	}
	// Its administrators who have an address hear of it, with the reason;
	// its members do not.
	sent := letters(t, mailDir)
	if len(sent) != 1 || !strings.Contains(sent[0], "To: olivier@example.com") || !strings.Contains(sent[0], "is suspended") ||
		!strings.Contains(sent[0], "Unpaid invoice of October") {
		t.Errorf("e-mail of the suspension = %q", sent)
	}

	// Every call but those of the administrators of the server is refused,
	// whoever makes it, reads included.
	for who, actor := range map[string]operation.Actor{"an administrator of the Cahut": olivier, "a member": camille, "an AI": ai, "a service": service} {
		_, err := operation.Do(l.ctx, l.c.Registry, l.c.About, actor, GetCahutInput{})
		var r *operation.Refusal
		if !errors.As(err, &r) || r.Code != "cahut_suspended" || !errors.Is(err, operation.ErrForbidden) {
			t.Errorf("get_cahut by %s: %v", who, err)
		}
	}
	// Whoever has not signed in learns that it is suspended, and nothing
	// else: signing in, with the right password, is refused too.
	if p := l.signInPage(); !p.Suspended {
		t.Errorf("describe_sign_in = %+v", p)
	}
	if a := l.about(l.admin); a.Suspension == nil || a.Suspension.Reason != "Unpaid invoice of October" {
		t.Errorf("get_cahut by the administrators of the server = %+v", a.Suspension)
	}
	if _, err := l.createUser(l.admin, "dana", false); err != nil {
		t.Errorf("create_user by the administrators of the server: %v", err)
	}

	// Suspending it again changes the reason only, and sends nothing.
	since := s.Suspension.Since
	if s, err := suspend(l.admin, "Decision of moderation"); err != nil || s.Suspension.Since != since || s.Suspension.Reason != "Decision of moderation" {
		t.Errorf("suspend_cahut again = %+v, %v", s.Suspension, err)
	}
	if e := l.lastEntry(); e.Detail != "new reason: Decision of moderation" {
		t.Errorf("logbook = %+v", e)
	}

	// Resumed, it answers everyone again, and its administrators hear of it.
	s, err = operation.Do(l.ctx, l.c.Registry, settings.Resume, l.admin, settings.ResumeCahutInput{})
	if err != nil || s.Suspension != nil {
		t.Fatalf("resume_cahut = %+v, %v", s.Suspension, err)
	}
	if e := l.lastEntry(); e.Operation != "resume_cahut" || !strings.HasSuffix(e.Detail, ": Decision of moderation") {
		t.Errorf("logbook = %+v", e)
	}
	if a := l.about(camille); a.Suspension != nil {
		t.Errorf("get_cahut once resumed = %+v", a.Suspension)
	}
	if p := l.signInPage(); p.Suspended {
		t.Errorf("describe_sign_in once resumed = %+v", p)
	}
	sent = letters(t, mailDir)
	if len(sent) != 2 || !strings.Contains(sent[1], "To: olivier@example.com") || !strings.Contains(sent[1], "is open again") {
		t.Errorf("e-mail of the resumption = %q", sent[1:])
	}
}
