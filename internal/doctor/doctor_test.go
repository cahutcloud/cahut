package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/update"
)

var now = time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)

// fake is a healthy AlmaLinux server with one Cahut, which tests break.
type fake struct {
	root     bool
	files    map[string]string
	commands map[string]string // by command line; absent fails
	owner    string
	label    string
	free     uint64
	local    []net.IP
	dns      map[string][]net.IP
	closed   map[string]bool
	gets     map[string]string // body by URL; absent fails
	health   update.Health
	cert     *Certificate
}

func healthy() *fake {
	f := &fake{
		root:  true,
		files: map[string]string{"/etc/os-release": "ID=\"almalinux\"\nVERSION_ID=\"10.2\"\nPRETTY_NAME=\"AlmaLinux 10.2 (Lavender Lion)\"\n"},
		commands: map[string]string{
			"timedatectl show --property=NTPSynchronized --value": "yes",
			"systemctl is-active cahut":                           "active",
			"systemctl is-enabled cahut":                          "enabled",
			"systemctl is-active firewalld":                       "active",
			"firewall-cmd --query-service=http":                   "yes",
			"firewall-cmd --query-service=https":                  "yes",
			"getenforce":                                          "Enforcing",
			"systemctl is-active cahut-update.timer":              "active",
		},
		owner:  "cahut",
		label:  "system_u:object_r:bin_t:s0",
		free:   40 << 30,
		local:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("203.0.113.7")},
		dns:    map[string][]net.IP{"cahut.example.com": {net.ParseIP("203.0.113.7")}},
		closed: map[string]bool{},
		gets: map[string]string{
			"https://cahut.example.com/-/healthz":            "ok\n",
			"https://acme-v02.api.letsencrypt.org/directory": "{}",
			"https://api.github.com/":                        "{}",
		},
		health: update.Health{Version: "0.4.0", OK: true, Cahuts: []update.CahutHealth{{Domain: "cahut.example.com", OK: true}}},
		cert:   &Certificate{NotBefore: now.AddDate(0, 0, -10), NotAfter: now.AddDate(0, 0, 80), Issuer: "Let's Encrypt"},
	}
	return f
}

func (f *fake) Command(_ context.Context, name string, args ...string) (string, error) {
	out, ok := f.commands[strings.Join(append([]string{name}, args...), " ")]
	if !ok {
		return "", errors.New("exit status 1")
	}
	return out, nil
}

func (f *fake) ReadFile(name string) ([]byte, error) {
	s, ok := f.files[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return []byte(s), nil
}

func (f *fake) Stat(name string) (fs.FileInfo, error) {
	if name == "/var/lib/cahut" && f.owner != "" {
		return os.Stat(".")
	}
	return nil, fs.ErrNotExist
}

func (f *fake) Owner(string) (string, error)     { return f.owner, nil }
func (f *fake) Label(string) (string, error)     { return f.label, nil }
func (f *fake) FreeBytes(string) (uint64, error) { return f.free, nil }
func (f *fake) LocalAddrs() ([]net.IP, error)    { return f.local, nil }
func (f *fake) Root() bool                       { return f.root }
func (f *fake) Now() time.Time                   { return now }
func (f *fake) LookupIP(_ context.Context, host string) ([]net.IP, error) {
	if ips, ok := f.dns[host]; ok {
		return ips, nil
	}
	return nil, errors.New("no such host")
}

func (f *fake) Dial(_ context.Context, addr string) error {
	if f.closed[addr] {
		return errors.New("connection refused")
	}
	return nil
}

func (f *fake) Get(_ context.Context, url string) (int, []byte, error) {
	if url == update.HealthURL {
		if f.closed["127.0.0.1:80"] {
			return 0, nil, errors.New("connection refused")
		}
		data, _ := json.Marshal(f.health)
		return 200, data, nil
	}
	body, ok := f.gets[url]
	if !ok {
		return 0, nil, errors.New("connection refused")
	}
	return 200, []byte(body), nil
}

func (f *fake) Certificate(context.Context, string, string) (Certificate, error) {
	if f.cert == nil {
		return Certificate{}, errors.New("remote error: tls: internal error")
	}
	return *f.cert, nil
}

func run(t *testing.T, f *fake, state string) []Result {
	t.Helper()
	return Run(context.Background(), f, Config{
		BinPath: "/usr/local/bin/cahut", DataDir: "/var/lib/cahut", StateDir: state,
		HealthURL: update.HealthURL, Version: "0.4.0",
	})
}

// managed returns a state directory of updates, with st in it.
func managed(t *testing.T, st update.Status) string {
	t.Helper()
	dir := t.TempDir()
	data, _ := json.Marshal(st)
	if err := os.WriteFile(filepath.Join(dir, update.StatusFile), data, 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// find returns the first result of topic whose text holds text.
func find(results []Result, topic, text string) (Result, bool) {
	for _, r := range results {
		if r.Topic == topic && strings.Contains(r.Text, text) {
			return r, true
		}
	}
	return Result{}, false
}

func TestHealthyServer(t *testing.T) {
	results := run(t, healthy(), managed(t, update.Status{Policy: update.Patch}))
	for _, r := range results {
		if r.Status == Fail || r.Status == Warn {
			t.Errorf("%s %s: %s", r.Status, r.Topic, r.Text)
		}
	}
	for _, want := range []struct{ topic, text string }{
		{"System", "AlmaLinux 10.2"},
		{"Service", "Cahut 0.4.0 answers, with 1 Cahut"},
		{"Firewall", "firewalld lets in"},
		{"SELinux", "Enforcing"},
		{"cahut.example.com", "Points to this machine (203.0.113.7)"},
		{"cahut.example.com", "valid until 11 January 2027"},
		{"cahut.example.com", "https://cahut.example.com/ answers"},
		{"Updates", "CAHUT_UPDATES=patch"},
		{"Hosting provider", "Ports 80 and 443"},
	} {
		if _, ok := find(results, want.topic, want.text); !ok {
			t.Errorf("no %s result saying %q", want.topic, want.text)
		}
	}
}

func TestProblems(t *testing.T) {
	for _, tt := range []struct {
		name   string
		breaks func(*fake)
		status Status
		topic  string
		text   string
	}{
		{"untested system", func(f *fake) { f.files["/etc/os-release"] = "ID=debian\nVERSION_ID=\"12\"\n" }, Warn, "System", "tested on AlmaLinux"},
		{"clock", func(f *fake) { f.commands["timedatectl show --property=NTPSynchronized --value"] = "no" }, Fail, "Clock", "timedatectl set-ntp true"},
		{"service stopped", func(f *fake) { f.commands["systemctl is-active cahut"] = "failed" }, Fail, "Service", "sudo systemctl start cahut"},
		{"not at boot", func(f *fake) { f.commands["systemctl is-enabled cahut"] = "disabled" }, Warn, "Service", "systemctl enable cahut"},
		{"no answer", func(f *fake) { f.closed["127.0.0.1:80"] = true }, Fail, "Service", "does not answer on port 80"},
		{"old version runs", func(f *fake) { f.health.Version = "0.3.0" }, Warn, "Service", "systemctl restart cahut"},
		{"owner", func(f *fake) { f.owner = "root" }, Fail, "Data", "chown -R cahut:cahut"},
		{"disk full", func(f *fake) { f.free = 300 << 20 }, Fail, "Data", "0.3 GiB free"},
		{"disk low", func(f *fake) { f.free = 3 << 30 }, Warn, "Data", "3.0 GiB free"},
		{"port taken", func(f *fake) { f.closed["127.0.0.1:443"] = true }, Fail, "Ports", "port 443"},
		{"firewalld closed", func(f *fake) { delete(f.commands, "firewall-cmd --query-service=https") }, Fail, "Firewall", "--add-service=http --add-service=https"},
		{"ufw closed", func(f *fake) {
			f.commands["systemctl is-active firewalld"] = "inactive"
			f.commands["ufw status"] = "Status: active\n\nTo     Action  From\n22/tcp ALLOW   Anywhere\n80/tcp ALLOW   Anywhere\n"
		}, Fail, "Firewall", "ufw allow 80/tcp && sudo ufw allow 443/tcp"},
		{"ufw open", func(f *fake) {
			f.commands["systemctl is-active firewalld"] = "inactive"
			f.commands["ufw status"] = "Status: active\n\n80/tcp ALLOW Anywhere\n443/tcp ALLOW Anywhere\n443/tcp (v6) ALLOW Anywhere (v6)\n"
		}, OK, "Firewall", "ufw lets in"},
		{"no firewall", func(f *fake) { f.commands["systemctl is-active firewalld"] = "inactive" }, OK, "Firewall", "No firewall"},
		{"not root", func(f *fake) { f.root = false }, Info, "Firewall", "as root"},
		{"SELinux label", func(f *fake) { f.label = "unconfined_u:object_r:user_tmp_t:s0" }, Fail, "SELinux", "restorecon /usr/local/bin/cahut"},
		{"Cahut does not open", func(f *fake) {
			f.health.OK, f.health.Cahuts[0].OK, f.health.Cahuts[0].Error = false, false, "migration 0012: no such table"
		}, Fail, "cahut.example.com", "migration 0012"},
		{"no Cahut", func(f *fake) { f.health.Cahuts = nil }, Fail, "Cahuts", "domain add"},
		{"no DNS", func(f *fake) { delete(f.dns, "cahut.example.com") }, Fail, "cahut.example.com", "add an A record"},
		{"DNS on loopback", func(f *fake) { f.dns["cahut.example.com"] = []net.IP{net.ParseIP("127.0.0.1")} }, Fail, "cahut.example.com", "public address"},
		{"DNS elsewhere", func(f *fake) { f.dns["cahut.example.com"] = []net.IP{net.ParseIP("198.51.100.1")} }, Warn, "cahut.example.com", "not an address of this machine"},
		{"no certificate", func(f *fake) { f.cert = nil }, Fail, "cahut.example.com", "Let's Encrypt must reach port 80"},
		{"staging certificate", func(f *fake) { f.cert.Trusted = errors.New("x509: certificate signed by unknown authority") }, Warn, "cahut.example.com", "staging"},
		{"expired", func(f *fake) { f.cert.NotAfter = now.Add(-time.Hour) }, Fail, "cahut.example.com", "expired on 23 October 2026"},
		{"late renewal", func(f *fake) { f.cert.NotBefore, f.cert.NotAfter = now.AddDate(0, 0, -85), now.AddDate(0, 0, 5) }, Warn, "cahut.example.com", "should have been renewed"},
		{"unreachable", func(f *fake) { delete(f.gets, "https://cahut.example.com/-/healthz") }, Warn, "cahut.example.com", "hosting provider"},
		{"Let's Encrypt", func(f *fake) { delete(f.gets, "https://acme-v02.api.letsencrypt.org/directory") }, Fail, "Let's Encrypt", "Cannot be reached"},
		{"GitHub", func(f *fake) { delete(f.gets, "https://api.github.com/") }, Warn, "GitHub", "cannot read new releases"},
		{"timer off", func(f *fake) { f.commands["systemctl is-active cahut-update.timer"] = "inactive" }, Warn, "Updates", "enable --now cahut-update.timer"},
		{"development", func(f *fake) { f.health.Cahuts[0].Domain = "beta.localhost" }, Info, "beta.localhost", "development"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := healthy()
			tt.breaks(f)
			results := run(t, f, managed(t, update.Status{}))
			r, ok := find(results, tt.topic, tt.text)
			if !ok || r.Status != tt.status {
				t.Errorf("want %s %s saying %q; got:", tt.status, tt.topic, tt.text)
				for _, r := range results {
					t.Logf("  %s %s: %s", r.Status, r.Topic, r.Text)
				}
			}
		})
	}
}

func TestUpdates(t *testing.T) {
	at := time.Date(2026, 10, 22, 4, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name   string
		st     *update.Status
		status Status
		text   string
	}{
		{"not managed", nil, Warn, "does not update itself"},
		{"updated", &update.Status{Last: &update.Run{From: "0.3.0", To: "0.4.0", Result: update.Updated, Ended: at}}, OK, "Updated from 0.3.0 to 0.4.0 on 22 October 2026"},
		{"rolled back", &update.Status{Last: &update.Run{To: "0.4.1", Result: update.RolledBack, Reason: "a Cahut did not open", Ended: at}}, Warn, "rolled back: a Cahut did not open"},
		{"failed", &update.Status{Last: &update.Run{To: "0.4.1", Result: update.Failed, Reason: "down", Ended: at}}, Fail, "so did going back"},
		{"check failed", &update.Status{Check: &update.Check{At: at, Error: "GitHub answers 401"}}, Warn, "could not read the releases"},
		{"available", &update.Status{Check: &update.Check{At: at, Latest: "0.5.0"}}, Info, "0.5.0 is available"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := filepath.Join(t.TempDir(), "none")
			if tt.st != nil {
				state = managed(t, *tt.st)
			}
			results := run(t, healthy(), state)
			if r, ok := find(results, "Updates", tt.text); !ok || r.Status != tt.status {
				t.Errorf("want %s saying %q; got %+v", tt.status, tt.text, results)
			}
		})
	}
}
