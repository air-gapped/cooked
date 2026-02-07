package ssrf

import (
	"net"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		// Loopback
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"127.255.255.255", true},
		{"::1", true},

		// Private (RFC 1918)
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},

		// IPv6 unique local address
		{"fd00::1", true},
		{"fdff::1", true},

		// Link-local unicast
		{"169.254.1.1", true},
		{"169.254.169.254", true},
		{"fe80::1", true},

		// Link-local multicast
		{"ff02::1", true},

		// Unspecified
		{"0.0.0.0", true},
		{"::", true},

		// Multicast
		{"224.0.0.1", true},
		{"239.255.255.255", true},
		{"ff01::1", true},

		// CGNAT (100.64.0.0/10)
		{"100.64.0.1", true},
		{"100.100.100.100", true},
		{"100.127.255.255", true},

		// Public IPs — must NOT be blocked
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false},
		{"104.16.0.1", false},
		{"2606:4700::1", false},
		{"2001:db8::1", false}, // documentation range, but not blocked by stdlib

		// Edge of CGNAT range
		{"100.63.255.255", false},
		{"100.128.0.0", false},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP %q", tc.ip)
			}
			got := IsBlockedIP(ip)
			if got != tc.blocked {
				t.Errorf("IsBlockedIP(%s) = %v, want %v", tc.ip, got, tc.blocked)
			}
		})
	}
}
