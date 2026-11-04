package windata

import "testing"

func TestScopes(t *testing.T) {
	for _, tt := range []struct {
		abi    int
		errata uint64
		want   uint64
	}{
		{5, erratumSignals, 0},
		// Without the fix of erratum 2, the signal scope would stop the
		// worker as it restricts its threads.
		{6, 0, scopeAbstractUnixSocket},
		{6, 1, scopeAbstractUnixSocket},
		{6, erratumSignals, scopeAbstractUnixSocket | scopeSignal},
		{8, 0x7, scopeAbstractUnixSocket | scopeSignal},
	} {
		if got := scopes(tt.abi, tt.errata); got != tt.want {
			t.Errorf("scopes(%d, %#x) = %#x, want %#x", tt.abi, tt.errata, got, tt.want)
		}
	}
}
