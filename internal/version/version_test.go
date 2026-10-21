package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Error(`Version is empty, want a version or "dev"`)
	}
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("GoVersion = %q, want a Go toolchain version", info.GoVersion)
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, want GOOS/GOARCH", info.Platform)
	}
}

func TestFill(t *testing.T) {
	tests := []struct {
		name   string
		linked string // version set by the linker
		main   string // main module version recorded by the toolchain
		want   string
	}{
		{name: "from build info", main: "v1.2.3", want: "1.2.3"},
		{name: "linker wins", linked: "2.0.0", main: "v1.2.3", want: "2.0.0"},
		{name: "devel build", main: "(devel)", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bi := &debug.BuildInfo{
				Main: debug.Module{Version: tt.main},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "0123abcdef"},
					{Key: "vcs.time", Value: "2026-10-21T08:00:00Z"},
					{Key: "vcs.modified", Value: "true"},
				},
			}
			info := Info{Version: tt.linked}
			info.fill(bi)
			if info.Version != tt.want {
				t.Errorf("Version = %q, want %q", info.Version, tt.want)
			}
			if info.Commit != "0123abcdef" || info.Date != "2026-10-21T08:00:00Z" || !info.Modified {
				t.Errorf("vcs settings not read: %+v", info)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "release",
			info: Info{Version: "1.2.3", Commit: "0123abcdef", Date: "2026-10-21T08:00:00Z", GoVersion: "go1.26.2", Platform: "linux/amd64"},
			want: "cahut 1.2.3 (commit 0123abc, 2026-10-21T08:00:00Z, go1.26.2 linux/amd64)",
		},
		{
			name: "modified tree",
			info: Info{Version: "dev", Commit: "0123abcdef", Modified: true, GoVersion: "go1.26.2", Platform: "darwin/arm64"},
			want: "cahut dev (commit 0123abc, modified, go1.26.2 darwin/arm64)",
		},
		{
			name: "no vcs",
			info: Info{Version: "dev", GoVersion: "go1.26.2", Platform: "linux/arm64"},
			want: "cahut dev (go1.26.2 linux/arm64)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
