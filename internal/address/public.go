package address

import (
	"fmt"
	"net/netip"
	"syscall"
)

// nonPublic holds the ranges that netip does not classify, and that a
// server must not be made to reach.
var nonPublic = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // this network
	netip.MustParsePrefix("100.64.0.0/10"),   // shared address space, carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved, and broadcast
	netip.MustParsePrefix("64:ff9b::/96"),    // NAT64, which reaches IPv4 addresses
	netip.MustParsePrefix("::/96"),           // IPv4-compatible, deprecated, such as ::127.0.0.1
	netip.MustParsePrefix("::ffff:0:0:0/96"), // IPv4-translated, such as ::ffff:0:127.0.0.1
	netip.MustParsePrefix("64:ff9b:1::/48"),  // local NAT64
	netip.MustParsePrefix("2001:db8::/32"),   // documentation
	netip.MustParsePrefix("2002::/16"),       // 6to4, which reaches IPv4 addresses
	netip.MustParsePrefix("2001::/32"),       // Teredo
	netip.MustParsePrefix("fec0::/10"),       // site-local, deprecated
	netip.MustParsePrefix("100::/64"),        // discard
	netip.MustParsePrefix("192.88.99.0/24"),  // 6to4 relays
	netip.MustParsePrefix("198.51.100.0/24"), // documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // documentation
	netip.MustParsePrefix("192.0.2.0/24"),    // documentation
}

// Public reports whether ip is an address on the internet, as opposed to
// this machine, its networks or a reserved range.
func Public(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, p := range nonPublic {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}

// DialPublic is the Control of a net.Dialer that connects to public
// addresses only. It checks the address it connects to, after the name is
// resolved, so that a name that resolves first to a public address and then
// to a private one gains nothing.
func DialPublic(_, address string, _ syscall.RawConn) error {
	ap, err := netip.ParseAddrPort(address)
	if err != nil {
		return err
	}
	if !Public(ap.Addr()) {
		return fmt.Errorf("%s is not a public address", ap.Addr())
	}
	return nil
}
