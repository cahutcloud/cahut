package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cahut.cloud/internal/release"
)

// Updater updates the server it runs on. Its zero value is not usable: the
// command line fills it from the paths of install.sh.
type Updater struct {
	Repo     string // GitHub repository of the releases, owner/name
	API      string // GitHub API, https://api.github.com
	Token    string // to read the releases of a private repository
	Client   *http.Client
	Keys     []ed25519.PublicKey // that sign releases
	Arch     string              // amd64 or arm64
	Running  string              // version of the binary installed now
	Policy   Policy
	BinPath  string // the binary that the service runs
	DataDir  string // the data of the service
	StateDir string // the state of updates
	Service  Service
	// HealthURL is where the service answers with its version and the
	// state of each Cahut; Wait bounds how long the updater waits for it.
	HealthURL string
	Wait      time.Duration
	Out       io.Writer // what the person reads
	Now       func() time.Time
}

// Service starts and stops the service of Cahut.
type Service interface {
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}

// Systemd is the service of Cahut under systemd.
type Systemd struct{ Unit string }

// Stop stops the service, and waits until it has stopped.
func (s Systemd) Stop(ctx context.Context) error { return systemctl(ctx, "stop", s.Unit) }

// Start starts the service.
func (s Systemd) Start(ctx context.Context) error { return systemctl(ctx, "start", s.Unit) }

func systemctl(ctx context.Context, verb, unit string) error {
	out, err := exec.CommandContext(ctx, "systemctl", verb, unit).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s %s: %w: %s", verb, unit, err, bytes.TrimSpace(out))
	}
	return nil
}

// Options say what a run of the updater does.
type Options struct {
	Version    string // a release to install, newer than the one running; empty for the newest
	Binary     string // a binary to install instead of a release, without signature
	Check      bool   // read the releases, install nothing
	Unattended bool   // run as systemd starts it: the request of the console, or else the policy
}

// Run checks or updates, and records what it did in the state directory.
// The error says why nothing was installed; an update undone is not an
// error, and Status.Last says how it ended.
func (u *Updater) Run(ctx context.Context, opts Options) (Status, error) {
	unlock, err := u.lock()
	if err != nil {
		return Status{}, err
	}
	defer unlock()
	st, _, err := ReadStatus(u.StateDir)
	if err != nil {
		u.say("the state of the last update cannot be read, and starts again: %v", err)
	}
	st.Policy = u.Policy

	by := ByCommandLine
	if opts.Unattended {
		by = ByTimer
		what, asked, err := takeRequest(u.DataDir)
		switch {
		case err != nil:
			return st, err
		case asked:
			by = ByConsole
			opts.Version, opts.Check = what, what == ""
		case u.Policy == Off:
			return st, nil
		case u.Policy == CheckOnly:
			opts.Check = true
		}
	}

	if opts.Binary != "" {
		run := u.newRun(by)
		u.installBinary(ctx, opts.Binary, run)
		st.record(run)
		return st, u.save(st, run)
	}

	releases, err := u.releases(ctx)
	st.Check = &Check{At: u.Now().UTC()}
	if err != nil {
		st.Check.Error = err.Error()
		return st, errors.Join(err, writeStatus(u.StateDir, st))
	}
	if newest := newestOf(releases, false); newest != nil {
		st.Check.Latest, st.Check.URL = newest.Version.String(), newest.URL
	}
	target, err := u.choose(releases, opts, by)
	if err != nil || target == nil || opts.Check {
		if err == nil && opts.Check {
			u.reportCheck(releases)
		}
		return st, errors.Join(err, writeStatus(u.StateDir, st))
	}

	run := u.newRun(by)
	run.To = target.Version.String()
	u.installRelease(ctx, *target, run)
	st.record(run)
	return st, u.save(st, run)
}

// save records st, and returns an error when run installed nothing.
func (u *Updater) save(st Status, run *Run) error {
	err := writeStatus(u.StateDir, st)
	if run.Result == Refused {
		err = errors.Join(errors.New(run.Reason), err)
	}
	return err
}

func (u *Updater) newRun(by string) *Run {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	now := u.Now().UTC()
	return &Run{
		ID:      now.Format("20060102T150405Z") + "-" + hex.EncodeToString(b),
		From:    u.Running,
		By:      by,
		Started: now,
	}
}

// choose returns the release to install, or nil when there is none.
func (u *Updater) choose(releases []Release, opts Options, by string) (*Release, error) {
	running, runErr := release.ParseVersion(u.Running)
	if opts.Version != "" {
		want, err := release.ParseVersion(opts.Version)
		if err != nil {
			return nil, err
		}
		for i := range releases {
			if releases[i].Version.Compare(want) == 0 {
				if runErr == nil && want.Compare(running) <= 0 {
					return nil, fmt.Errorf("%s is not newer than %s, the version running: an update never goes back", want, running)
				}
				return &releases[i], nil
			}
		}
		return nil, fmt.Errorf("no published release %s in %s", want, u.Repo)
	}
	var candidates []Release
	for _, r := range releases {
		if runErr == nil && r.Version.Compare(running) <= 0 {
			continue
		}
		// The timer keeps to the patch releases of the minor version
		// running, unless the policy says otherwise, and leaves a build
		// that is not a release alone.
		if by == ByTimer && (runErr != nil || u.Policy == Patch && !r.Version.SameMinor(running)) {
			continue
		}
		candidates = append(candidates, r)
	}
	target := newestOf(candidates, false)
	if target == nil && !opts.Check { // a check reports on its own
		u.say("cahut %s is up to date", u.Running)
		if newest := newestOf(releases, false); newest != nil && runErr == nil && newest.Version.Compare(running) > 0 {
			u.say("%s is available: install it with sudo %s update", newest.Version, u.BinPath)
		}
	}
	return target, nil
}

// newestOf returns the newest of releases, which are newest first,
// pre-releases included only when pre is true.
func newestOf(releases []Release, pre bool) *Release {
	for i := range releases {
		if pre || !releases[i].Pre {
			return &releases[i]
		}
	}
	return nil
}

func (u *Updater) reportCheck(releases []Release) {
	newest := newestOf(releases, false)
	if newest == nil {
		u.say("%s has no published release", u.Repo)
		return
	}
	running, err := release.ParseVersion(u.Running)
	if err == nil && newest.Version.Compare(running) <= 0 {
		u.say("cahut %s is up to date", u.Running)
		return
	}
	u.say("cahut %s runs, and %s is available: %s", u.Running, newest.Version, newest.URL)
}

// installRelease downloads target, checks it and installs it.
func (u *Updater) installRelease(ctx context.Context, target Release, run *Run) {
	u.say("downloading Cahut %s for linux/%s", target.Version, u.Arch)
	next, err := u.fetch(ctx, target)
	if err != nil {
		run.Result, run.Reason, run.Ended = Refused, err.Error(), u.Now().UTC()
		u.say("nothing was installed: %v", err)
		return
	}
	u.say("signature and checksum checked")
	u.install(ctx, next, run)
}

// installBinary installs a binary given on the command line.
func (u *Updater) installBinary(ctx context.Context, file string, run *Run) {
	data, err := readLimited(file, maxBinary)
	if err == nil {
		var next string
		if next, err = u.stage(data); err == nil {
			u.install(ctx, next, run)
			return
		}
	}
	run.Result, run.Reason, run.Ended = Refused, err.Error(), u.Now().UTC()
	u.say("nothing was installed: %v", err)
}

// fetch downloads the archive of target for this machine, checks the
// signature of the checksum file and the checksum of the archive, and
// stages its binary.
func (u *Updater) fetch(ctx context.Context, target Release) (string, error) {
	sums, err := u.download(ctx, target, release.ChecksumsFile, maxChecksums)
	if err != nil {
		return "", err
	}
	sig, err := u.download(ctx, target, release.SignatureFile, ed25519.SignatureSize)
	if err != nil {
		return "", fmt.Errorf("%w: releases are signed from 0.4.0 on", err)
	}
	if err := release.Verify(u.Keys, sums, sig); err != nil {
		return "", err
	}
	name := fmt.Sprintf("cahut_%s_linux_%s.tar.gz", target.Version, u.Arch)
	archive, err := u.download(ctx, target, name, maxArchive)
	if err != nil {
		return "", err
	}
	if err := release.Check(sums, name, archive); err != nil {
		return "", err
	}
	bin, err := extract(archive)
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	next, err := u.stage(bin)
	if err != nil {
		return "", err
	}
	if v, _ := versionOf(ctx, next); v != target.Version.String() {
		return "", fmt.Errorf("the binary of release %s says it is %q", target.Version, v)
	}
	return next, nil
}

// extract returns the cahut binary of a release archive.
func extract(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil, errors.New("no cahut binary in the archive")
		}
		if err != nil {
			return nil, err
		}
		if h.Name == "cahut" && h.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(io.LimitReader(tr, maxBinary+1))
			if err == nil && len(data) > maxBinary {
				err = errors.New("the binary is too large")
			}
			return data, err
		}
	}
}

// stage writes the new binary in the state directory, and checks that it
// runs on this machine.
func (u *Updater) stage(bin []byte) (string, error) {
	dir := filepath.Join(u.StateDir, "next")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	next := filepath.Join(dir, "cahut")
	if err := writeFile(next, bin, 0o755); err != nil {
		return "", err
	}
	if _, err := versionOf(context.Background(), next); err != nil {
		return "", fmt.Errorf("the new binary does not run on this machine: %w", err)
	}
	return next, nil
}

// versionOf runs bin version and returns the version it prints.
func versionOf(ctx context.Context, bin string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin, "version").Output()
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(out))
	if len(fields) < 2 || fields[0] != "cahut" {
		return "", fmt.Errorf("%s version prints %q", bin, bytes.TrimSpace(out))
	}
	return fields[1], nil
}

// install puts the staged binary next in place of the one the service
// runs, after a snapshot of the databases, then checks the service, and
// rolls back when it does not answer with the new version.
func (u *Updater) install(ctx context.Context, next string, run *Run) {
	defer func() { run.Ended = u.Now().UTC() }()
	to, err := versionOf(ctx, next)
	if err != nil {
		run.Result, run.Reason = Refused, err.Error()
		return
	}
	run.To = to
	refuse := func(err error) {
		run.Result, run.Reason = Refused, err.Error()
		u.say("nothing was installed: %v", err)
	}

	u.say("stopping the service")
	if err := u.Service.Stop(ctx); err != nil {
		refuse(err)
		return
	}
	previous := filepath.Join(u.StateDir, "previous", "cahut")
	err = errors.Join(u.snapshot(), copyFile(u.BinPath, previous))
	if err != nil {
		refuse(fmt.Errorf("keep the previous version: %w", err))
		if err := u.Service.Start(ctx); err != nil {
			u.say("the service does not start again: %v", err)
		}
		return
	}
	u.say("installing Cahut %s", to)
	err = u.put(ctx, next)
	if err == nil {
		err = u.Service.Start(ctx)
	}
	if err == nil {
		err = u.waitHealthy(ctx, to)
	}
	if err == nil {
		run.Result = Updated
		u.say("Cahut %s is running", to)
		return
	}

	run.Reason = err.Error()
	u.say("Cahut %s failed: %v", to, err)
	u.say("going back to Cahut %s", run.From)
	err = u.Service.Stop(ctx)
	if err == nil {
		err = errors.Join(u.put(ctx, previous), u.restore())
	}
	if err == nil {
		err = u.Service.Start(ctx)
	}
	if err == nil {
		err = u.waitHealthy(ctx, run.From)
	}
	if err != nil {
		run.Result = Failed
		run.Reason += "; then going back failed: " + err.Error()
		u.say("going back failed: %v. See journalctl -u cahut", err)
		return
	}
	run.Result = RolledBack
	u.say("Cahut %s is running again", run.From)
}

// put replaces the binary of the service with src, at once, and gives it
// the SELinux label of its place where SELinux runs.
func (u *Updater) put(ctx context.Context, src string) error {
	if err := copyFile(src, u.BinPath); err != nil {
		return err
	}
	if restorecon, err := exec.LookPath("restorecon"); err == nil {
		_ = exec.CommandContext(ctx, restorecon, u.BinPath).Run()
	}
	return nil
}

// waitHealthy waits until the service answers with version want, having
// opened every Cahut.
func (u *Updater) waitHealthy(ctx context.Context, want string) error {
	ctx, cancel := context.WithTimeout(ctx, u.Wait)
	defer cancel()
	var last error
	for {
		h, err := u.health(ctx)
		switch {
		case err != nil:
			last = fmt.Errorf("the service did not answer: %w", err)
		case h.Version != want:
			last = fmt.Errorf("the service answers with version %s, not %s", h.Version, want)
		case !h.OK:
			last = errors.New("a Cahut did not open")
			for _, c := range h.Cahuts {
				if !c.OK {
					last = fmt.Errorf("the Cahut of %s did not open: %s", c.Domain, c.Error)
					break
				}
			}
		default:
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("after %s, %w", u.Wait, last)
		case <-time.After(time.Second):
		}
	}
}

func (u *Updater) health(ctx context.Context) (Health, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.HealthURL, nil)
	if err != nil {
		return Health{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Health{}, err
	}
	defer resp.Body.Close()
	var h Health
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&h); err != nil {
		return Health{}, fmt.Errorf("%s: %w", resp.Status, err)
	}
	return h, nil
}

// databases returns the files of the databases in dir, relative to it: the
// server's and those of each Cahut, with their journals.
func databases(dir string) ([]string, error) {
	var files []string
	for _, pattern := range []string{"server.db*", filepath.Join("cahuts", "*", "cahut.db*")} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			rel, err := filepath.Rel(dir, m)
			if err != nil {
				return nil, err
			}
			files = append(files, rel)
		}
	}
	return files, nil
}

// snapshot copies the databases, while the service is stopped, to the
// state directory, in place of the previous snapshot.
func (u *Updater) snapshot() error {
	dir := filepath.Join(u.StateDir, "snapshot")
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	files, err := databases(u.DataDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := copyFile(filepath.Join(u.DataDir, f), filepath.Join(dir, f)); err != nil {
			return err
		}
	}
	return nil
}

// restore puts back the databases of the snapshot, and removes those made
// since, while the service is stopped.
func (u *Updater) restore() error {
	dir := filepath.Join(u.StateDir, "snapshot")
	now, err := databases(u.DataDir)
	if err != nil {
		return err
	}
	for _, f := range now {
		if err := os.Remove(filepath.Join(u.DataDir, f)); err != nil {
			return err
		}
	}
	files, err := databases(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := copyFile(filepath.Join(dir, f), filepath.Join(u.DataDir, f)); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies src to dst, at once, with its mode and its owner, making
// the directories of dst (readable by root only) as needed.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	fi, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // a no-op once renamed
	_, err = io.Copy(tmp, in)
	if err == nil {
		err = tmp.Chmod(fi.Mode().Perm())
	}
	if st, ok := fi.Sys().(*syscall.Stat_t); ok && err == nil && (int(st.Uid) != os.Getuid() || int(st.Gid) != os.Getgid()) {
		err = tmp.Chown(int(st.Uid), int(st.Gid))
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

func readLimited(file string, limit int64) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err == nil && int64(len(data)) > limit {
		err = fmt.Errorf("%s holds more than %d bytes", file, limit)
	}
	return data, err
}

// lock keeps two runs of the updater apart, such as the timer and a person.
func (u *Updater) lock() (func(), error) {
	if err := os.MkdirAll(u.StateDir, 0o750); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(u.StateDir, "lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, errors.New("another update is running")
	}
	return func() { f.Close() }, nil
}

func (u *Updater) say(format string, args ...any) {
	if u.Out != nil {
		fmt.Fprintf(u.Out, "cahut: "+format+"\n", args...)
	}
}
