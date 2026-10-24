package address

import (
	"net/netip"
	"testing"
)

func TestPublicAddresses(t *testing.T) {
	for ip, want := range map[string]bool{
		"160.79.104.10":   true,
		"2606:4700::1111": true,
		"127.0.0.1":       false,
		"10.0.0.1":        false,
		"172.16.0.1":      false,
		"192.168.1.1":     false,
		"169.254.169.254": false, // cloud metadata
		"100.64.0.1":      false,
		"0.0.0.0":         false,
		"::1":             false,
		"fe80::1":         false,
		"fc00::1":         false,
		"::ffff:10.0.0.1": false,
		"64:ff9b::a00:1":  false,
		"224.0.0.1":       false,
		"::127.0.0.1":     false,
		"::ffff:0:a00:1":  false,
	} {
		if got := Public(netip.MustParseAddr(ip)); got != want {
			t.Errorf("public(%s) = %v, want %v", ip, got, want)
		}
	}
}
