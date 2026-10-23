// Package update installs new releases of Cahut on a server, as root: it
// checks their signature, copies the databases, swaps the binary, checks
// that the service answers with the new version, and rolls back when it
// does not. The service, which runs unprivileged, only reads the state that
// the updater leaves, and writes a request for it. See
// docs/adr/0024-signed-releases-and-updates.md.
package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cahut.cloud/internal/release"
)

// Where install.sh puts things.
const (
	// DefaultStateDir holds the state of updates, owned by root and readable
	// by the group of the service.
	DefaultStateDir = "/var/lib/cahut-update"
	// DefaultBinPath is the binary that the service runs.
	DefaultBinPath = "/usr/local/bin/cahut"
	// RequestFile, in the data directory, is where the service asks for a
	// check or an update.
	RequestFile = "update-request"
	// StatusFile, in the state directory, holds the last check and the last
	// update.
	StatusFile = "status.json"
	// HealthURL is where the service says how it is, to the machine itself.
	HealthURL = "http://127.0.0.1:80/-/healthz"
)

// Policy says what the daily run of the updater does.
type Policy string

// Policies, as CAHUT_UPDATES sets them.
const (
	Patch     Policy = "patch" // install the patch releases of the minor version running; report the others
	All       Policy = "all"   // install any newer release
	CheckOnly Policy = "check" // report newer releases only
	Off       Policy = "off"   // ask GitHub nothing
)

// ParsePolicy reads the value of CAHUT_UPDATES, Patch when it is empty.
func ParsePolicy(s string) (Policy, error) {
	switch p := Policy(strings.TrimSpace(s)); p {
	case "":
		return Patch, nil
	case Patch, All, CheckOnly, Off:
		return p, nil
	}
	return "", fmt.Errorf("CAHUT_UPDATES=%s: use patch, all, check or off", s)
}

// Results of an update.
const (
	Updated    = "updated"     // the new version runs
	RolledBack = "rolled_back" // the new version failed its check: the previous one runs again
	Refused    = "refused"     // nothing was installed: download, signature, checksum or binary
	Failed     = "failed"      // the previous version failed its check too: the service is down
)

// Who started an update.
const (
	ByCommandLine = "command line"
	ByTimer       = "timer"
	ByConsole     = "console"
)

// Status is the state that the updater leaves for the service.
type Status struct {
	Policy Policy `json:"policy,omitempty"`
	Check  *Check `json:"check,omitempty" jsonschema:"the last time the updater read the releases"`
	Last   *Run   `json:"last,omitempty" jsonschema:"the last update, done or undone"`
	// History holds the last updates, newest first, so that the service
	// records each of them even when two come within a minute.
	History []Run `json:"history,omitempty"`
}

// keep is how many updates History holds.
const keep = 10

// record makes run the last update.
func (s *Status) record(run *Run) {
	s.Last = run
	s.History = append([]Run{*run}, s.History...)
	if len(s.History) > keep {
		s.History = s.History[:keep]
	}
}

// Runs returns the updates of s that have ended, oldest first.
func (s Status) Runs() []Run {
	var runs []Run
	for i := len(s.History) - 1; i >= 0; i-- {
		runs = append(runs, s.History[i])
	}
	if len(runs) == 0 && s.Last != nil {
		runs = append(runs, *s.Last)
	}
	ended := runs[:0]
	for _, r := range runs {
		if !r.Ended.IsZero() {
			ended = append(ended, r)
		}
	}
	return ended
}

// Check is a reading of the published releases.
type Check struct {
	At     time.Time `json:"at"`
	Latest string    `json:"latest,omitempty" jsonschema:"the newest published release"`
	URL    string    `json:"url,omitempty" jsonschema:"the page of that release, with its notes"`
	Error  string    `json:"error,omitempty" jsonschema:"why the releases could not be read"`
}

// Run is one update.
type Run struct {
	ID      string    `json:"id"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	By      string    `json:"by" jsonschema:"command line, timer or console"`
	Started time.Time `json:"started"`
	Ended   time.Time `json:"ended"`
	Result  string    `json:"result" jsonschema:"updated, rolled_back, refused or failed"`
	Reason  string    `json:"reason,omitempty" jsonschema:"why it did not update"`
}

// ReadStatus reads the state in dir. It returns false when the server does
// not update itself, as when dir does not exist.
func ReadStatus(dir string) (Status, bool, error) {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return Status{}, false, nil
	}
	data, err := os.ReadFile(filepath.Join(dir, StatusFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Status{}, true, nil
	}
	if err != nil {
		return Status{}, true, err
	}
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return Status{}, true, fmt.Errorf("read %s: %w", StatusFile, err)
	}
	return s, true, nil
}

// writeStatus replaces the state in dir, at once. The file takes the group
// of dir, which install.sh makes the service's, and the group may read it.
func writeStatus(dir string, s Status) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, StatusFile), append(data, '\n'), 0o640)
}

// writeFile replaces file with data, through a new file that it renames.
func writeFile(file string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(file), "."+filepath.Base(file)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // a no-op once renamed
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), file)
}

// checkRequest is what a request holds when it names no version.
const checkRequest = "check"

// maxRequest bounds what the updater reads of a request.
const maxRequest = 64

// WriteRequest asks the updater, from the service, to check the releases
// (version empty) or to install a version. A systemd path unit starts the
// updater when the file appears.
func WriteRequest(dataDir, version string) error {
	what := checkRequest
	if version != "" {
		v, err := release.ParseVersion(version)
		if err != nil {
			return err
		}
		what = v.String()
	}
	return writeFile(filepath.Join(dataDir, RequestFile), []byte(what+"\n"), 0o640)
}

// ReadRequest returns what the service asked for, without taking it: the
// console shows it until the updater does. It returns "" for a check, a
// version to install, and false when there is no request.
func ReadRequest(dataDir string) (string, bool, error) {
	return readRequest(dataDir, false)
}

// takeRequest reads and removes the request in dataDir. It returns "" for a
// check, a version to install, and false when there is no request. The
// service owns the data directory, so the request is read without
// following links, and only a version or a check is accepted.
func takeRequest(dataDir string) (string, bool, error) {
	return readRequest(dataDir, true)
}

func readRequest(dataDir string, take bool) (string, bool, error) {
	file := filepath.Join(dataDir, RequestFile)
	f, err := os.OpenFile(file, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if take {
		// Whatever it holds, a request is taken once.
		defer os.Remove(file)
	}
	if err != nil {
		return "", true, fmt.Errorf("the request of the console cannot be read: %w", err)
	}
	defer f.Close()
	if fi, err := f.Stat(); err != nil || !fi.Mode().IsRegular() {
		return "", true, errors.New("the request of the console is not a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(f, maxRequest+1))
	if err != nil {
		return "", true, err
	}
	if len(data) > maxRequest {
		return "", true, errors.New("the request of the console is too long")
	}
	what := strings.TrimSpace(string(data))
	if what == checkRequest {
		return "", true, nil
	}
	v, err := release.ParseVersion(what)
	if err != nil {
		return "", true, errors.New("the request of the console names no version")
	}
	return v.String(), true, nil
}

// Health is what the service answers at HealthURL.
type Health struct {
	Version string        `json:"version"`
	OK      bool          `json:"ok" jsonschema:"whether every Cahut opened"`
	Cahuts  []CahutHealth `json:"cahuts"`
}

// CahutHealth says whether a Cahut opened, with its migrations.
type CahutHealth struct {
	Domain string `json:"domain"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
}
