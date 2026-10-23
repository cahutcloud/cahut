package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// script returns a fake cahut binary: a shell script that prints version.
func script(version string) []byte {
	return []byte("#!/bin/sh\necho \"cahut " + version + " (test)\"\n")
}

func archiveOf(t *testing.T, bin []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, data := range map[string][]byte{"LICENSE": []byte("AGPL"), "cahut": bin} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// fakeGitHub serves releases, signed with key, at /repos/getcahut/cahut.
type fakeGitHub struct {
	*httptest.Server
	mu       sync.Mutex
	files    map[string][]byte // by path
	listing  []map[string]any
	requests int
}

func newGitHub(t *testing.T) *fakeGitHub {
	g := &fakeGitHub{files: map[string][]byte{}}
	g.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.requests++
		if r.URL.Path == "/repos/getcahut/cahut/releases" {
			_ = json.NewEncoder(w).Encode(g.listing)
			return
		}
		data, ok := g.files[r.URL.Path]
		if !ok || r.Header.Get("Accept") != "application/octet-stream" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(g.Close)
	return g
}

// add publishes release version for amd64, signed with key, with files
// changed by tweak when it is not nil.
func (g *fakeGitHub) add(t *testing.T, version string, key ed25519.PrivateKey, tweak func(files map[string][]byte)) {
	t.Helper()
	name := fmt.Sprintf("cahut_%s_linux_amd64.tar.gz", version)
	archive := archiveOf(t, script(version))
	sum := sha256.Sum256(archive)
	sums := []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
	files := map[string][]byte{name: archive, "checksums.txt": sums, "checksums.txt.sig": ed25519.Sign(key, sums)}
	if tweak != nil {
		tweak(files)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	var assets []map[string]any
	for n, data := range files {
		path := "/assets/" + version + "/" + n
		g.files[path] = data
		assets = append(assets, map[string]any{"url": g.URL + path, "name": n, "size": len(data)})
	}
	g.listing = append(g.listing, map[string]any{
		"tag_name": "v" + version, "html_url": "https://github.com/getcahut/cahut/releases/tag/v" + version,
		"draft": false, "prerelease": false, "assets": assets,
	})
}

// fakeService stands for systemd and the service: starting it reads the
// version of the binary in place, and a version in broken does not open
// its Cahuts. A version in migrates changes the databases when it starts.
type fakeService struct {
	t        *testing.T
	bin      string
	data     string
	mu       sync.Mutex
	running  string
	broken   map[string]bool
	migrates map[string]bool
	calls    []string
	failStop bool
}

func (s *fakeService) Stop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "stop")
	if s.failStop {
		return fmt.Errorf("stop refused")
	}
	s.running = ""
	return nil
}

func (s *fakeService) Start(ctx context.Context) error {
	v, err := versionOf(ctx, s.bin)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "start "+v)
	s.running = v
	if s.migrates[v] {
		write(s.t, filepath.Join(s.data, "server.db"), "migrated by "+v)
		write(s.t, filepath.Join(s.data, "cahuts", "2", "cahut.db"), "made by "+v)
	}
	return nil
}

func (s *fakeService) health(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running == "" {
		http.Error(w, "down", http.StatusBadGateway)
		return
	}
	h := Health{Version: s.running, OK: !s.broken[s.running], Cahuts: []CahutHealth{{Domain: "cahut.example.com", OK: !s.broken[s.running]}}}
	if !h.OK {
		h.Cahuts[0].Error = "migration 0012_new.sql: no such table"
	}
	_ = json.NewEncoder(w).Encode(h)
}

func write(t *testing.T, file, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, file string) string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

type fixture struct {
	u       *Updater
	github  *fakeGitHub
	service *fakeService
	key     ed25519.PrivateKey
	out     *bytes.Buffer
}

// newFixture sets up a server running running, with a database for the
// server and one for a Cahut.
func newFixture(t *testing.T, running string) *fixture {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin", "cahut")
	write(t, bin, string(script(running)))
	if err := os.Chmod(bin, 0o755); err != nil { //nolint:gosec // a fake binary must run
		t.Fatal(err)
	}
	data := filepath.Join(dir, "data")
	write(t, filepath.Join(data, "server.db"), "server of "+running)
	write(t, filepath.Join(data, "cahuts", "1", "cahut.db"), "cahut of "+running)
	write(t, filepath.Join(data, "cahuts", "1", "windows", "w", "index.html"), "a Window")

	pub, key, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	svc := &fakeService{t: t, bin: bin, data: data, running: running, broken: map[string]bool{}, migrates: map[string]bool{}}
	health := httptest.NewServer(http.HandlerFunc(svc.health))
	t.Cleanup(health.Close)
	gh := newGitHub(t)
	out := &bytes.Buffer{}
	u := &Updater{
		Repo:      "getcahut/cahut",
		API:       gh.URL,
		Client:    gh.Client(),
		Keys:      []ed25519.PublicKey{pub},
		Arch:      "amd64",
		Running:   running,
		Policy:    Patch,
		BinPath:   bin,
		DataDir:   data,
		StateDir:  filepath.Join(dir, "state"),
		Service:   svc,
		HealthURL: health.URL,
		Wait:      3 * time.Second,
		Out:       out,
		Now:       time.Now,
	}
	return &fixture{u: u, github: gh, service: svc, key: key, out: out}
}

func (f *fixture) installed(t *testing.T) string {
	t.Helper()
	v, err := versionOf(context.Background(), f.u.BinPath)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestUpdateToTheNewestRelease(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	f.github.add(t, "0.5.0", f.key, nil)
	f.service.migrates["0.5.0"] = true

	st, err := f.u.Run(context.Background(), Options{})
	if err != nil {
		t.Fatalf("Run: %v\n%s", err, f.out)
	}
	if v := f.installed(t); v != "0.5.0" {
		t.Errorf("installed %s, want 0.5.0", v)
	}
	if st.Last == nil || st.Last.Result != Updated || st.Last.From != "0.4.0" || st.Last.To != "0.5.0" || st.Last.By != ByCommandLine {
		t.Errorf("last run = %+v", st.Last)
	}
	if st.Check == nil || st.Check.Latest != "0.5.0" {
		t.Errorf("check = %+v, want the latest release 0.5.0", st.Check)
	}
	saved, managed, err := ReadStatus(f.u.StateDir)
	if err != nil || !managed || saved.Last == nil || saved.Last.ID != st.Last.ID || len(saved.Runs()) != 1 {
		t.Errorf("ReadStatus = %+v, %v, %v; want the run saved", saved, managed, err)
	}
	if got := read(t, filepath.Join(f.u.StateDir, "snapshot", "server.db")); got != "server of 0.4.0" {
		t.Errorf("the snapshot holds %q", got)
	}
	if v, _ := versionOf(context.Background(), filepath.Join(f.u.StateDir, "previous", "cahut")); v != "0.4.0" {
		t.Errorf("the previous binary is %q", v)
	}
	if got := read(t, filepath.Join(f.u.DataDir, "server.db")); got != "migrated by 0.5.0" {
		t.Errorf("the database holds %q, want what the new version made", got)
	}
}

func TestTimerKeepsToPatchReleases(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	f.github.add(t, "0.4.2", f.key, nil)
	f.github.add(t, "0.5.0", f.key, nil)
	st, err := f.u.Run(context.Background(), Options{Unattended: true})
	if err != nil {
		t.Fatalf("Run: %v\n%s", err, f.out)
	}
	if v := f.installed(t); v != "0.4.2" {
		t.Errorf("installed %s, want the newest patch release, 0.4.2", v)
	}
	if st.Last.By != ByTimer || st.Check.Latest != "0.5.0" {
		t.Errorf("by %q, latest %q", st.Last.By, st.Check.Latest)
	}

	// With every patch release in, the timer only reports 0.5.0.
	f.u.Running = "0.4.2"
	st, err = f.u.Run(context.Background(), Options{Unattended: true})
	if err != nil {
		t.Fatal(err)
	}
	if v := f.installed(t); v != "0.4.2" {
		t.Errorf("installed %s, want 0.4.2 kept", v)
	}
	if !strings.Contains(f.out.String(), "0.5.0 is available") {
		t.Errorf("the output does not report 0.5.0:\n%s", f.out)
	}

	f.u.Policy = All
	if _, err := f.u.Run(context.Background(), Options{Unattended: true}); err != nil {
		t.Fatal(err)
	}
	if v := f.installed(t); v != "0.5.0" {
		t.Errorf("with policy all, installed %s, want 0.5.0", v)
	}
}

func TestTimerCheckAndOff(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)

	f.u.Policy = Off
	if _, err := f.u.Run(context.Background(), Options{Unattended: true}); err != nil {
		t.Fatal(err)
	}
	if f.github.requests != 0 {
		t.Errorf("with policy off, %d requests to GitHub", f.github.requests)
	}

	f.u.Policy = CheckOnly
	st, err := f.u.Run(context.Background(), Options{Unattended: true})
	if err != nil {
		t.Fatal(err)
	}
	if v := f.installed(t); v != "0.4.0" || st.Last != nil {
		t.Errorf("with policy check, installed %s, last run %+v", v, st.Last)
	}
	if st.Check == nil || st.Check.Latest != "0.4.1" {
		t.Errorf("check = %+v", st.Check)
	}
	if n := strings.Count(f.out.String(), "0.4.1 is available"); n != 1 {
		t.Errorf("the check says %d times that 0.4.1 is available:\n%s", n, f.out)
	}
}

func TestBrokenReleaseRollsBack(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	f.service.broken["0.4.1"] = true
	f.service.migrates["0.4.1"] = true

	st, err := f.u.Run(context.Background(), Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if st.Last.Result != RolledBack || !strings.Contains(st.Last.Reason, "cahut.example.com did not open") {
		t.Errorf("last run = %+v, want rolled back because a Cahut did not open", st.Last)
	}
	if v := f.installed(t); v != "0.4.0" {
		t.Errorf("installed %s, want 0.4.0 back", v)
	}
	if got := read(t, filepath.Join(f.u.DataDir, "server.db")); got != "server of 0.4.0" {
		t.Errorf("the database holds %q, want the snapshot back", got)
	}
	if _, err := os.Stat(filepath.Join(f.u.DataDir, "cahuts", "2", "cahut.db")); !os.IsNotExist(err) {
		t.Error("the database made by the failed version is still there")
	}
	if got := read(t, filepath.Join(f.u.DataDir, "cahuts", "1", "windows", "w", "index.html")); got != "a Window" {
		t.Errorf("a Window file holds %q", got)
	}
	want := []string{"stop", "start 0.4.1", "stop", "start 0.4.0"}
	if fmt.Sprint(f.service.calls) != fmt.Sprint(want) {
		t.Errorf("service calls = %v, want %v", f.service.calls, want)
	}
}

func TestOldVersionAnsweringRollsBack(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.u.Wait = 2 * time.Second
	f.github.add(t, "0.4.1", f.key, nil)
	// The new version starts, but something else answers with the old one.
	f.u.Service = serviceFunc{start: func() {
		f.service.mu.Lock()
		defer f.service.mu.Unlock()
		f.service.running = "0.4.0"
	}}
	st, _ := f.u.Run(context.Background(), Options{})
	if st.Last.Result != RolledBack || !strings.Contains(st.Last.Reason, "answers with version 0.4.0, not 0.4.1") {
		t.Errorf("last run = %+v", st.Last)
	}
}

type serviceFunc struct{ start func() }

func (serviceFunc) Stop(context.Context) error { return nil }
func (s serviceFunc) Start(context.Context) error {
	s.start()
	return nil
}

func TestFailedRollbackIsReported(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.u.Wait = 2 * time.Second
	f.github.add(t, "0.4.1", f.key, nil)
	f.service.broken["0.4.1"] = true
	f.service.broken["0.4.0"] = true
	st, _ := f.u.Run(context.Background(), Options{})
	if st.Last.Result != Failed || !strings.Contains(st.Last.Reason, "then going back failed") {
		t.Errorf("last run = %+v", st.Last)
	}
}

func TestRefusedReleases(t *testing.T) {
	_, other, _ := ed25519.GenerateKey(nil)
	for name, tweak := range map[string]func(t *testing.T, f *fixture){
		"signed by another key": func(t *testing.T, f *fixture) { f.github.add(t, "0.4.1", other, nil) },
		"changed archive": func(t *testing.T, f *fixture) {
			f.github.add(t, "0.4.1", f.key, func(files map[string][]byte) {
				files["cahut_0.4.1_linux_amd64.tar.gz"] = archiveOf(t, script("0.4.1 evil"))
			})
		},
		"no signature": func(t *testing.T, f *fixture) {
			f.github.add(t, "0.4.1", f.key, func(files map[string][]byte) { delete(files, "checksums.txt.sig") })
		},
		"binary of another version": func(t *testing.T, f *fixture) {
			f.github.add(t, "0.4.1", f.key, func(files map[string][]byte) {
				archive := archiveOf(t, script("0.9.0"))
				sum := sha256.Sum256(archive)
				sums := []byte(hex.EncodeToString(sum[:]) + "  cahut_0.4.1_linux_amd64.tar.gz\n")
				files["cahut_0.4.1_linux_amd64.tar.gz"] = archive
				files["checksums.txt"] = sums
				files["checksums.txt.sig"] = ed25519.Sign(f.key, sums)
			})
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t, "0.4.0")
			tweak(t, f)
			st, err := f.u.Run(context.Background(), Options{})
			if err == nil || st.Last == nil || st.Last.Result != Refused {
				t.Errorf("err = %v, last run = %+v; want refused", err, st.Last)
			}
			if v := f.installed(t); v != "0.4.0" {
				t.Errorf("installed %s", v)
			}
			if len(f.service.calls) != 0 {
				t.Errorf("the service was touched: %v", f.service.calls)
			}
		})
	}
}

func TestNeverGoesBack(t *testing.T) {
	f := newFixture(t, "0.4.1")
	f.github.add(t, "0.4.0", f.key, nil)
	f.github.add(t, "0.4.1", f.key, nil)
	for _, v := range []string{"0.4.0", "0.4.1"} {
		if _, err := f.u.Run(context.Background(), Options{Version: v}); err == nil || !strings.Contains(err.Error(), "never goes back") {
			t.Errorf("--version %s: err = %v", v, err)
		}
	}
	if _, err := f.u.Run(context.Background(), Options{Version: "0.9.0"}); err == nil || !strings.Contains(err.Error(), "no published release") {
		t.Errorf("--version 0.9.0: err = %v", err)
	}
	if v := f.installed(t); v != "0.4.1" {
		t.Errorf("installed %s", v)
	}
}

func TestRequestOfTheConsole(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	f.github.add(t, "0.5.0", f.key, nil)
	f.u.Policy = Off // a request runs whatever the policy

	if err := WriteRequest(f.u.DataDir, "0.5.0"); err != nil {
		t.Fatal(err)
	}
	if what, asked, err := ReadRequest(f.u.DataDir); what != "0.5.0" || !asked || err != nil {
		t.Errorf("ReadRequest = %q, %v, %v", what, asked, err)
	}
	st, err := f.u.Run(context.Background(), Options{Unattended: true})
	if err != nil {
		t.Fatal(err)
	}
	if v := f.installed(t); v != "0.5.0" || st.Last.By != ByConsole {
		t.Errorf("installed %s by %q, want 0.5.0 by the console", v, st.Last.By)
	}
	if _, asked, _ := ReadRequest(f.u.DataDir); asked {
		t.Error("the request was not taken")
	}
}

func TestRequestsThatAreRefused(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	file := filepath.Join(f.u.DataDir, RequestFile)
	secret := filepath.Join(t.TempDir(), "secret")
	write(t, secret, "0.4.1")
	for name, make := range map[string]func(){
		"a link":     func() { _ = os.Symlink(secret, file) },
		"too long":   func() { write(t, file, strings.Repeat("0.4.1 ", 20)) },
		"no version": func() { write(t, file, "rm -rf /") },
		"older":      func() { write(t, file, "0.3.0") },
	} {
		make()
		_, err := f.u.Run(context.Background(), Options{Unattended: true})
		if err == nil {
			t.Errorf("%s: the request was taken", name)
		}
		if _, err := os.Lstat(file); !os.IsNotExist(err) {
			t.Errorf("%s: the request is still there", name)
		}
		if v := f.installed(t); v != "0.4.0" {
			t.Errorf("%s: installed %s", name, v)
		}
	}
}

func TestInstallABinary(t *testing.T) {
	f := newFixture(t, "0.4.0")
	file := filepath.Join(t.TempDir(), "cahut")
	write(t, file, string(script("0.4.1-dev")))
	st, err := f.u.Run(context.Background(), Options{Binary: file})
	if err != nil {
		t.Fatalf("Run: %v\n%s", err, f.out)
	}
	if v := f.installed(t); v != "0.4.1-dev" || st.Last.Result != Updated || st.Check != nil {
		t.Errorf("installed %s, last run %+v, check %+v", v, st.Last, st.Check)
	}
	if f.github.requests != 0 {
		t.Error("installing a binary asked GitHub")
	}

	write(t, file, "not a program")
	if _, err := f.u.Run(context.Background(), Options{Binary: file}); err == nil {
		t.Error("a binary that does not run was installed")
	}
}

func TestStopRefusedChangesNothing(t *testing.T) {
	f := newFixture(t, "0.4.0")
	f.github.add(t, "0.4.1", f.key, nil)
	f.service.failStop = true
	st, err := f.u.Run(context.Background(), Options{})
	if err == nil || st.Last.Result != Refused {
		t.Errorf("err = %v, last run = %+v", err, st.Last)
	}
	if v := f.installed(t); v != "0.4.0" {
		t.Errorf("installed %s", v)
	}
}

func TestOneRunAtATime(t *testing.T) {
	f := newFixture(t, "0.4.0")
	unlock, err := f.u.lock()
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := f.u.Run(context.Background(), Options{Check: true}); err == nil || !strings.Contains(err.Error(), "another update") {
		t.Errorf("err = %v, want another update running", err)
	}
}

func TestPolicies(t *testing.T) {
	for in, want := range map[string]Policy{"": Patch, "patch": Patch, " all ": All, "check": CheckOnly, "off": Off} {
		if got, err := ParsePolicy(in); got != want || err != nil {
			t.Errorf("ParsePolicy(%q) = %q, %v", in, got, err)
		}
	}
	if _, err := ParsePolicy("always"); err == nil {
		t.Error("ParsePolicy(always) succeeded")
	}
}

func TestStatusOfAServerThatDoesNotUpdate(t *testing.T) {
	if _, managed, err := ReadStatus(filepath.Join(t.TempDir(), "none")); managed || err != nil {
		t.Errorf("ReadStatus of a missing directory: managed %v, err %v", managed, err)
	}
	dir := t.TempDir()
	if st, managed, err := ReadStatus(dir); !managed || err != nil || st.Last != nil {
		t.Errorf("ReadStatus of an empty directory: %+v, %v, %v", st, managed, err)
	}
}

func TestHistoryKeepsTheLastUpdates(t *testing.T) {
	var st Status
	for i := range keep + 3 {
		st.record(&Run{ID: fmt.Sprint(i), Ended: time.Now()})
	}
	runs := st.Runs()
	if len(runs) != keep || runs[0].ID != "3" || runs[keep-1].ID != fmt.Sprint(keep+2) || st.Last.ID != fmt.Sprint(keep+2) {
		t.Errorf("runs = %+v, last = %+v", runs, st.Last)
	}
}
