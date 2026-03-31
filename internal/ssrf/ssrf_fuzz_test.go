package ssrf

import (
	"net"
	"testing"
)

func FuzzIsBlockedIP(f *testing.F) {
	// Seed with known IPs
	f.Add([]byte{127, 0, 0, 1})        // loopback
	f.Add([]byte{10, 0, 0, 1})         // private
	f.Add([]byte{8, 8, 8, 8})          // public
	f.Add([]byte{100, 64, 0, 1})       // CGNAT
	f.Add([]byte{169, 254, 1, 1})      // link-local
	f.Add([]byte{0, 0, 0, 0})          // unspecified
	f.Add([]byte{224, 0, 0, 1})        // multicast
	f.Add([]byte(net.IPv6loopback))    // ::1
	f.Add([]byte(net.IPv6unspecified)) // ::

	f.Fuzz(func(t *testing.T, data []byte) {
		ip := net.IP(data)
		if ip.To4() == nil && ip.To16() == nil {
			return // not a valid IP
		}
		// Must not panic
		IsBlockedIP(ip)
	})
}
