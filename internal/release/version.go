package release

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// Version is the number of a release, as Semantic Versioning 2.0.0 has it:
// 0.4.0, 0.4.1-rc.1. Builds that are not releases carry a pre-release too,
// such as 0.3.1-dev+e5c9df9 for a snapshot of GoReleaser, or
// 0.3.1-0.20261023105550-97b30eb61391 for a build of a commit after v0.3.0.
type Version struct {
	Major, Minor, Patch int
	Pre                 string // the pre-release, without its hyphen
}

// ParseVersion reads a version, with or without a leading v. Build metadata,
// after a plus sign, is dropped: it does not order versions.
func ParseVersion(s string) (Version, error) {
	text := strings.TrimPrefix(s, "v")
	text, _, _ = strings.Cut(text, "+")
	core, pre, hasPre := strings.Cut(text, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 || hasPre && pre == "" {
		return Version{}, fmt.Errorf("%q is not a version such as 0.4.0", s)
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 || p != strconv.Itoa(v) {
			return Version{}, fmt.Errorf("%q is not a version such as 0.4.0", s)
		}
		n[i] = v
	}
	return Version{Major: n[0], Minor: n[1], Patch: n[2], Pre: pre}, nil
}

func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// Compare returns -1, 0 or +1 as v is older than, the same as, or newer
// than w. A pre-release is older than its release.
func (v Version) Compare(w Version) int {
	if c := cmp.Compare(v.Major, w.Major); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Minor, w.Minor); c != 0 {
		return c
	}
	if c := cmp.Compare(v.Patch, w.Patch); c != 0 {
		return c
	}
	switch {
	case v.Pre == w.Pre:
		return 0
	case v.Pre == "":
		return 1
	case w.Pre == "":
		return -1
	}
	a, b := strings.Split(v.Pre, "."), strings.Split(w.Pre, ".")
	for i := range min(len(a), len(b)) {
		if c := comparePre(a[i], b[i]); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(a), len(b))
}

// comparePre orders two identifiers of a pre-release: numbers by value,
// below words, which compare as text.
func comparePre(a, b string) int {
	na, errA := strconv.Atoi(a)
	nb, errB := strconv.Atoi(b)
	switch {
	case errA == nil && errB == nil:
		return cmp.Compare(na, nb)
	case errA == nil:
		return -1
	case errB == nil:
		return 1
	}
	return strings.Compare(a, b)
}

// SameMinor reports whether v and w share their major and minor versions,
// as a patch release does with the release it fixes.
func (v Version) SameMinor(w Version) bool {
	return v.Major == w.Major && v.Minor == w.Minor
}
