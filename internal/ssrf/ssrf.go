package ssrf

import "net"

// cgnatRange is the Carrier-Grade NAT range (100.64.0.0/10), which is not
// covered by Go's net.IP.IsPrivate() but must be blocked for SSRF protection.
var cgnatRange = mustParseCIDR("100.64.0.0/10")

func mustParseCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}

// IsBlockedIP returns true if the given IP should be blocked for SSRF protection.
// It covers loopback, private (RFC 1918), link-local unicast/multicast, unspecified,
// multicast, and CGNAT (100.64.0.0/10) addresses.
func IsBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() ||
		cgnatRange.Contains(ip)
}
