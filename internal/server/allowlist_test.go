package server

import (
	"testing"
)

func TestAllowlist_NilAllowsAll(t *testing.T) {
	var a *Allowlist
	if !a.Allows("anything.com") {
		t.Error("nil allowlist should allow all hosts")
	}
	if !a.Allows("10.0.0.1") {
		t.Error("nil allowlist should allow IPs")
	}
}

func TestAllowlist_EmptyStringReturnsNil(t *testing.T) {
	a := ParseAllowlist("")
	if a != nil {
		t.Error("empty string should return nil allowlist")
	}
}

func TestAllowlist_Exact(t *testing.T) {
	tests := []struct {
		host   string
		wantOK bool
	}{
		// Exact match
		{"cgit.internal", true},
		{"s3.internal", true},

		// Subdomain match
		{"sub.cgit.internal", true},
		{"deep.sub.cgit.internal", true},

		// Must NOT match — attacker-controlled suffix (F-04)
		{"cgit.internal.attacker.com", false},

		// Not in list
		{"evil.com", false},

		// With port
		{"cgit.internal:8080", true},

		// Case insensitive
		{"CGIT.INTERNAL", true},
		{"S3.Internal", true},
	}

	a := ParseAllowlist("cgit.internal, s3.internal")
	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			if got := a.Allows(tc.host); got != tc.wantOK {
				t.Errorf("Allows(%q) = %v, want %v", tc.host, got, tc.wantOK)
			}
		})
	}
}

func TestAllowlist_CIDR(t *testing.T) {
	tests := []struct {
		host   string
		wantOK bool
	}{
		{"10.0.0.1", true},
		{"10.0.1.50", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"11.0.0.1", false},
		{"192.168.1.1", false},
		{"8.8.8.8", false},

		// With port
		{"10.0.0.1:8080", true},

		// Hostname (not IP) should not match CIDR
		{"example.com", false},
	}

	a := ParseAllowlist("10.0.0.0/8, 172.16.0.0/12")
	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			if got := a.Allows(tc.host); got != tc.wantOK {
				t.Errorf("Allows(%q) = %v, want %v", tc.host, got, tc.wantOK)
			}
		})
	}
}

func TestAllowlist_CIDRIPv6(t *testing.T) {
	a := ParseAllowlist("fd00::/8")

	tests := []struct {
		host   string
		wantOK bool
	}{
		{"fd00::1", true},
		{"fdff::1", true},
		{"fe80::1", false},
		{"::1", false},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			if got := a.Allows(tc.host); got != tc.wantOK {
				t.Errorf("Allows(%q) = %v, want %v", tc.host, got, tc.wantOK)
			}
		})
	}
}

func TestAllowlist_Wildcard(t *testing.T) {
	tests := []struct {
		host   string
		wantOK bool
	}{
		// Matches
		{"foo.internal", true},
		{"bar.internal", true},
		{"a.b.internal", true},
		{"deep.sub.corp.example.com", true},

		// Must NOT match the bare domain
		{"internal", false},

		// Must NOT match unrelated suffix
		{"notinternal", false},
		{"evil.com", false},
	}

	a := ParseAllowlist("*.internal, *.corp.example.com")
	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			if got := a.Allows(tc.host); got != tc.wantOK {
				t.Errorf("Allows(%q) = %v, want %v", tc.host, got, tc.wantOK)
			}
		})
	}
}

func TestAllowlist_Mixed(t *testing.T) {
	a := ParseAllowlist("*.internal, 10.0.0.0/8, gitea.specific.host")

	tests := []struct {
		host   string
		wantOK bool
	}{
		// Exact
		{"gitea.specific.host", true},
		{"sub.gitea.specific.host", true},

		// Wildcard
		{"foo.internal", true},

		// CIDR
		{"10.0.1.50", true},

		// None match
		{"evil.com", false},
		{"11.0.0.1", false},
	}

	for _, tc := range tests {
		t.Run(tc.host, func(t *testing.T) {
			if got := a.Allows(tc.host); got != tc.wantOK {
				t.Errorf("Allows(%q) = %v, want %v", tc.host, got, tc.wantOK)
			}
		})
	}
}

func TestAllowlist_WhitespaceHandling(t *testing.T) {
	a := ParseAllowlist(" cgit.internal , , s3.internal ")
	if !a.Allows("cgit.internal") {
		t.Error("should allow cgit.internal despite whitespace")
	}
	if !a.Allows("s3.internal") {
		t.Error("should allow s3.internal despite whitespace")
	}
}

func TestAllowlist_InvalidCIDRIgnored(t *testing.T) {
	// ParseAllowlist silently skips unparseable CIDRs.
	// Validation is done at config parse time.
	a := ParseAllowlist("10.0.0.0/33, valid.host")
	if !a.Allows("valid.host") {
		t.Error("valid.host should still be allowed")
	}
	if a.Allows("10.0.0.1") {
		t.Error("10.0.0.1 should not be allowed (invalid CIDR was skipped)")
	}
}

func FuzzAllowlist(f *testing.F) {
	seeds := []struct {
		host    string
		allowed string
	}{
		{"example.com", ""},
		{"cgit.internal", "cgit.internal"},
		{"s3.internal", "cgit.internal,s3.internal"},
		{"evil.com", "cgit.internal,s3.internal"},
		{"cgit.internal:8080", "cgit.internal"},
		{"", ""},
		{"", "example.com"},
		{"host", " host , other "},
		{"10.0.0.1", "10.0.0.0/8"},
		{"foo.internal", "*.internal"},
		{"a.b.c.d.e", "a.b,c.d"},
		{"10.0.0.1", "10.0.0.0/8,*.internal,exact.host"},
	}
	for _, s := range seeds {
		f.Add(s.host, s.allowed)
	}

	f.Fuzz(func(t *testing.T, host, allowed string) {
		a := ParseAllowlist(allowed)

		// Must never panic
		a.Allows(host)

		// Nil allowlist allows everything
		if a == nil && !a.Allows(host) {
			t.Error("nil allowlist should allow all hosts")
		}
	})
}
