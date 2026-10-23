package address

import "net/netip"

// shared is the shared address space of carrier-grade NAT, RFC 6598, which
// netip does not treat as private.
var shared = netip.MustParsePrefix("100.64.0.0/10")

// PublicIP returns s as the IPv4 address of a Cahut that has no domain, in
// its canonical form, and reports whether it can be one: a global unicast
// address, not private, loopback, link-local or shared. IPv6 waits: a
// Content-Security-Policy cannot name an IPv6 address, and the policy of a
// Window names the address of its Cahut. See
// docs/adr/0026-cahut-at-an-ip-address.md and
// docs/adr/0028-ip-address-of-a-cahut-is-ipv4.md.
func PublicIP(s string) (string, bool) {
	ip, err := netip.ParseAddr(s)
	if err != nil || !ip.Is4() || ip.String() != s {
		return "", false
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || shared.Contains(ip) {
		return "", false
	}
	return ip.String(), true
}

// ValidAddress reports whether s can be the address of a Cahut: a domain,
// or the public IPv4 address of its server.
func ValidAddress(s string) bool {
	_, ip := PublicIP(s)
	return ip || ValidDomain(s)
}

// IsIP reports whether the address of a Cahut is an IP address.
func IsIP(s string) bool {
	_, err := netip.ParseAddr(s)
	return err == nil
}
