package release

import "testing"

func TestParseVersion(t *testing.T) {
	for in, want := range map[string]string{
		"0.4.0":                               "0.4.0",
		"v0.4.0":                              "0.4.0",
		"1.12.3-rc.1":                         "1.12.3-rc.1",
		"0.3.1-dev+e5c9df9":                   "0.3.1-dev",
		"0.3.1-0.20261023105550-97b30eb61391": "0.3.1-0.20261023105550-97b30eb61391",
	} {
		v, err := ParseVersion(in)
		if err != nil {
			t.Errorf("ParseVersion(%q): %v", in, err)
			continue
		}
		if v.String() != want {
			t.Errorf("ParseVersion(%q) = %s, want %s", in, v, want)
		}
	}
	for _, bad := range []string{"", "dev", "0.4", "0.4.0.1", "0.4.x", "0.04.0", "-1.0.0", "0.4.0-", "v"} {
		if _, err := ParseVersion(bad); err == nil {
			t.Errorf("ParseVersion(%q) succeeded", bad)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	// In order, oldest first, as Semantic Versioning orders them.
	ordered := []string{
		"0.3.0",
		"0.3.1-0.20261023105550-97b30eb61391",
		"0.3.1-dev",
		"0.3.1",
		"0.4.0-alpha",
		"0.4.0-alpha.1",
		"0.4.0-alpha.beta",
		"0.4.0-beta.2",
		"0.4.0-beta.11",
		"0.4.0-rc.1",
		"0.4.0",
		"0.4.1",
		"0.10.0",
		"1.0.0",
	}
	for i, a := range ordered {
		va, _ := ParseVersion(a)
		for j, b := range ordered {
			vb, _ := ParseVersion(b)
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := va.Compare(vb); got != want {
				t.Errorf("Compare(%s, %s) = %d, want %d", a, b, got, want)
			}
		}
	}
}

func TestSameMinor(t *testing.T) {
	v, _ := ParseVersion("0.4.0")
	for s, want := range map[string]bool{"0.4.3": true, "0.4.0-rc.1": true, "0.5.0": false, "1.4.0": false} {
		w, _ := ParseVersion(s)
		if got := v.SameMinor(w); got != want {
			t.Errorf("0.4.0 SameMinor %s = %v, want %v", s, got, want)
		}
	}
}
