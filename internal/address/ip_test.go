package address

import "testing"

func TestPublicIP(t *testing.T) {
	for _, s := range []string{"203.0.113.7", "51.38.1.2", "8.8.8.8"} {
		if got, ok := PublicIP(s); !ok || got != s {
			t.Errorf("PublicIP(%q) = %q, %v", s, got, ok)
		}
	}
	for _, s := range []string{
		"10.0.0.1", "172.16.5.4", "192.168.1.1", // private
		"127.0.0.1", "169.254.1.1", "100.64.0.1", // loopback, link-local, shared
		"0.0.0.0", "255.255.255.255", "224.0.0.1", // unspecified, broadcast, multicast
		"2001:db8::7", "::ffff:203.0.113.7", "[2001:db8::7]", // IPv6 waits
		"203.0.113.07", " 203.0.113.7", "cahut.example.com", "",
	} {
		if _, ok := PublicIP(s); ok {
			t.Errorf("PublicIP(%q) accepted it", s)
		}
	}
}

func TestValidAddress(t *testing.T) {
	for s, want := range map[string]bool{
		"cahut.example.com": true, "203.0.113.7": true, "localhost": true,
		"10.0.0.1": false, "203.0.113": false, "*.example.com": false,
	} {
		if got := ValidAddress(s); got != want {
			t.Errorf("ValidAddress(%q) = %v, want %v", s, got, want)
		}
	}
	if !IsIP("203.0.113.7") || IsIP("cahut.example.com") {
		t.Error("IsIP")
	}
}
