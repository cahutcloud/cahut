package window

import (
	"errors"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

func TestDeclaration(t *testing.T) {
	f := setup(t)
	publish := func(declaration string) (Window, error) {
		t.Helper()
		in := `{"title": "Chart", "kind": "dashboard", "files": [{"path": "index.html", "content": "x"}]`
		if declaration != "" {
			in += `, "declaration": ` + declaration
		}
		out, err := f.invoke(t, f.olivier, "publish_window", in+`}`)
		if err != nil {
			return Window{}, err
		}
		return out.(Window), nil
	}
	w, err := publish(`{"scripts": true, "origins": ["https://cdn.jsdelivr.net/", "HTTPS://*.Example.com", "wss://live.example.com:8443", "https://cdn.jsdelivr.net"], "forms": ["https://forms.example.com"]}`)
	if err != nil {
		t.Fatal(err)
	}
	d := w.Declaration
	if d == nil || !d.Scripts || strings.Join(d.Origins, " ") != "https://cdn.jsdelivr.net https://*.example.com wss://live.example.com:8443" ||
		strings.Join(d.Forms, " ") != "https://forms.example.com" {
		t.Errorf("declaration = %+v", d)
	}
	out, err := f.invoke(t, f.olivier, "get_window", `{"window": "`+w.URL+`"}`)
	if err != nil || out.(Detail).Declaration == nil || !out.(Detail).Declaration.Scripts {
		t.Errorf("get_window = %+v, %v", out, err)
	}
	entries, err := logbook.List(t.Context(), f.db.Read, 2)
	if err != nil {
		t.Fatal(err)
	}
	if e := entries[1]; e.Operation != "publish_window" || !strings.HasSuffix(e.Detail, "; scripts; origins https://cdn.jsdelivr.net https://*.example.com wss://live.example.com:8443; forms https://forms.example.com") {
		t.Errorf("logbook = %+v", e)
	}

	// An update declares again, and a rollback brings back the previous declaration.
	out, err = f.invoke(t, f.olivier, "update_window", `{"window": "`+w.URL+`", "declaration": {"scripts": false}, "files": [{"path": "index.html", "content": "y"}]}`)
	if err != nil || out.(Window).Declaration.Scripts {
		t.Errorf("update_window = %+v, %v", out, err)
	}
	out, err = f.invoke(t, f.olivier, "rollback_window", `{"window": "`+w.URL+`"}`)
	if err != nil || !out.(Window).Declaration.Scripts {
		t.Errorf("rollback_window = %+v, %v", out, err)
	}

	for _, bad := range []string{
		"",
		`{"scripts": false, "origins": ["http://cdn.example.com"]}`,
		`{"scripts": false, "origins": ["https://cdn.example.com/lib.js"]}`,
		`{"scripts": false, "origins": ["*"]}`,
		`{"scripts": false, "origins": ["https://*"]}`,
		`{"scripts": false, "forms": ["javascript:alert(1)"]}`,
		`{"scripts": false, "origins": ["https://a.example.com 'unsafe-inline'"]}`,
		`{"scripts": false, "origins": ["https://a.example.com;script-src *"]}`,
		`{"scripts": false, "origins": ["` + strings.Repeat(`https://a.example.com", "`, MaxOrigins) + `https://b.example.com"]}`,
	} {
		if _, err := publish(bad); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("declaration %s: err = %v, want ErrInvalid", bad, err)
		}
	}
}

func TestPolicy(t *testing.T) {
	const self = "https://cahut.example.com"
	if got, want := (*Declaration)(nil).Policy(self), "sandbox allow-scripts allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads"; got != want {
		t.Errorf("undeclared policy = %q, want %q", got, want)
	}
	none := (&Declaration{}).Policy(self)
	for _, want := range []string{"sandbox allow-forms ", "default-src 'none'", "connect-src https://cahut.example.com;", "form-action 'none'", "base-uri 'none'"} {
		if !strings.Contains(none, want) {
			t.Errorf("policy without calls = %q, want %q in it", none, want)
		}
	}
	if strings.Contains(none, "script-src") || strings.Contains(none, "allow-scripts") {
		t.Errorf("policy without scripts = %q", none)
	}
	d := &Declaration{Scripts: true, Origins: []string{"https://cdn.jsdelivr.net"}, Forms: []string{"https://forms.example.com"}}
	p := d.Policy(self)
	for _, want := range []string{
		"sandbox allow-scripts ",
		"script-src https://cahut.example.com 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net",
		"connect-src https://cahut.example.com https://cdn.jsdelivr.net",
		"img-src https://cahut.example.com data: blob: https://cdn.jsdelivr.net",
		"form-action https://forms.example.com",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("policy = %q, want %q in it", p, want)
		}
	}
}
