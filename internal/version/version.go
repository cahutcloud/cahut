// Package version reports which build of Cahut is running.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// version is set by the release build:
//
//	-ldflags "-X cahut.cloud/internal/version.version=1.2.3"
//
// When it is empty, the version comes from the build info that the Go
// toolchain records, which covers go install and local builds.
var version string

// Info describes the running binary.
type Info struct {
	Version   string // release version without the leading v, or "dev"
	Commit    string // full commit hash, empty when unknown
	Date      string // commit time in RFC 3339, empty when unknown
	Modified  bool   // built from a working tree with uncommitted changes
	GoVersion string
	Platform  string // GOOS/GOARCH
}

// Get returns the build information of the running binary.
func Get() Info {
	info := Info{
		Version:   version,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		info.fill(bi)
	}
	if info.Version == "" {
		info.Version = "dev"
	}
	return info
}

func (i *Info) fill(bi *debug.BuildInfo) {
	if i.Version == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		i.Version = strings.TrimPrefix(bi.Main.Version, "v")
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			i.Commit = s.Value
		case "vcs.time":
			i.Date = s.Value
		case "vcs.modified":
			i.Modified = s.Value == "true"
		}
	}
}

// String formats the information on one line, for example:
//
//	cahut 1.2.3 (commit 0123abc, 2026-10-21T08:00:00Z, go1.26.2 linux/amd64)
func (i Info) String() string {
	var details []string
	if i.Commit != "" {
		commit := "commit " + shorten(i.Commit)
		if i.Modified {
			commit += ", modified"
		}
		details = append(details, commit)
	}
	if i.Date != "" {
		details = append(details, i.Date)
	}
	details = append(details, i.GoVersion+" "+i.Platform)
	return fmt.Sprintf("cahut %s (%s)", i.Version, strings.Join(details, ", "))
}

func shorten(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}
