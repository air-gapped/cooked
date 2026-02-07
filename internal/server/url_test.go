package server

import (
	"strings"
	"testing"
)

func TestParseUpstreamURL_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantPath string
	}{
		{"https://example.com/README.md", "example.com", "/README.md"},
		{"http://cgit.internal/repo/plain/file.md", "cgit.internal", "/repo/plain/file.md"},
		{"https://s3.internal/bucket/path/file.md?X-Amz-Signature=abc", "s3.internal", "/bucket/path/file.md"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			u, err := ParseUpstreamURL(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Host != tc.wantHost {
				t.Errorf("Host = %q, want %q", u.Host, tc.wantHost)
			}
			if u.Path != tc.wantPath {
				t.Errorf("Path = %q, want %q", u.Path, tc.wantPath)
			}
		})
	}
}

func TestParseUpstreamURL_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"ftp scheme", "ftp://example.com/file"},
		{"no scheme", "example.com/file.md"},
		{"file scheme", "file:///etc/passwd"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseUpstreamURL(tc.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCheckAllowedUpstream(t *testing.T) {
	tests := []struct {
		host    string
		allowed string
		wantOK  bool
	}{
		// Empty allowed = allow all
		{"example.com", "", true},
		{"anything.com", "", true},

		// Exact match
		{"cgit.internal", "cgit.internal", true},
		{"s3.internal", "cgit.internal,s3.internal", true},

		// Subdomain match
		{"sub.cgit.internal", "cgit.internal", true},
		{"deep.sub.cgit.internal", "cgit.internal", true},

		// F-04: Must NOT match — attacker-controlled suffix
		{"cgit.internal.attacker.com", "cgit.internal", false},

		// Not in list
		{"evil.com", "cgit.internal,s3.internal", false},

		// With port
		{"cgit.internal:8080", "cgit.internal", true},

		// Whitespace in allowed list
		{"cgit.internal", " cgit.internal , s3.internal ", true},

		// Case insensitive
		{"CGIT.INTERNAL", "cgit.internal", true},
		{"cgit.internal", "CGIT.INTERNAL", true},
	}

	for _, tc := range tests {
		t.Run(tc.host+"_"+tc.allowed, func(t *testing.T) {
			got := CheckAllowedUpstream(tc.host, tc.allowed)
			if got != tc.wantOK {
				t.Errorf("CheckAllowedUpstream(%q, %q) = %v, want %v", tc.host, tc.allowed, got, tc.wantOK)
			}
		})
	}
}

func TestIsPrivateAddress_IPs(t *testing.T) {
	tests := []struct {
		addr    string
		private bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false},
	}

	for _, tc := range tests {
		t.Run(tc.addr, func(t *testing.T) {
			got, err := IsPrivateAddress(tc.addr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.private {
				t.Errorf("IsPrivateAddress(%q) = %v, want %v", tc.addr, got, tc.private)
			}
		})
	}
}

func TestIsPrivateAddress_WithPort(t *testing.T) {
	got, err := IsPrivateAddress("127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected 127.0.0.1:8080 to be private")
	}
}

func FuzzParseUpstreamURL(f *testing.F) {
	// Seed corpus: valid URLs, edge cases
	seeds := []string{
		"https://example.com/README.md",
		"http://cgit.internal/repo/plain/file.md",
		"https://s3.internal/bucket/path/file.md?X-Amz-Signature=abc",
		"",
		"ftp://example.com/file",
		"file:///etc/passwd",
		"example.com/file.md",
		"https://",
		"http://",
		"https://host",
		"https://host:8080/path",
		"https://user:pass@host/path",
		"https://例え.jp/日本語.md",
		"https://host/" + string(make([]byte, 8192)),
		"://missing-scheme",
		"https://host/path?q=1&r=2#frag",
		"HTTPS://EXAMPLE.COM/FILE.MD",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		u, err := ParseUpstreamURL(input)
		if err != nil {
			return
		}
		// If parsing succeeded, the URL must have a valid scheme and host
		if u.Scheme != "http" && u.Scheme != "https" {
			t.Errorf("accepted non-http(s) scheme: %q", u.Scheme)
		}
		if u.Host == "" {
			t.Error("accepted URL with empty host")
		}
	})
}

func FuzzCheckAllowedUpstream(f *testing.F) {
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
		{"例え.jp", "例え.jp"},
		{"a.b.c.d.e", "a.b,c.d"},
	}
	for _, s := range seeds {
		f.Add(s.host, s.allowed)
	}

	f.Fuzz(func(t *testing.T, host, allowed string) {
		// Must never panic
		got := CheckAllowedUpstream(host, allowed)

		// If allowed list is empty, everything should be allowed
		if allowed == "" && !got {
			t.Error("empty allowed list should allow all hosts")
		}
	})
}

func FuzzExtractUpstreamFromPath(f *testing.F) {
	seeds := []struct {
		path  string
		query string
	}{
		{"/https://example.com/file.md", ""},
		{"/https://s3.internal/file.md", "X-Amz-Signature=abc"},
		{"/http://server/doc.md", ""},
		{"/http:/server/doc.md", ""},
		{"/https:/server/doc.md", ""},
		{"", ""},
		{"/", ""},
		{"/plainpath", "key=val"},
		{"/https://例え.jp/日本語.md", ""},
		{"/http://host/" + string(make([]byte, 4096)), ""},
	}
	for _, s := range seeds {
		f.Add(s.path, s.query)
	}

	f.Fuzz(func(t *testing.T, path, query string) {
		result := ExtractUpstreamFromPath(path, query)

		// Determinism: same inputs must produce same output
		result2 := ExtractUpstreamFromPath(path, query)
		if result != result2 {
			t.Errorf("non-deterministic: %q vs %q", result, result2)
		}

		// If query is non-empty, result must end with "?<query>"
		if query != "" {
			want := "?" + query
			if !strings.HasSuffix(result, want) {
				t.Errorf("result %q does not end with query %q", result, want)
			}
		}

		// Result should not have the leading slash from the request path
		// (only the first "/" is stripped, so "//" → "/" is correct)
		trimmed := strings.TrimPrefix(path, "/")
		if query == "" && result != trimmed {
			// Without scheme-fix, result should equal path minus leading "/"
			// But scheme-fix changes "http:/" to "http://", so skip that case
			if !strings.HasPrefix(trimmed, "http:/") && !strings.HasPrefix(trimmed, "https:/") {
				if result != trimmed {
					t.Errorf("unexpected result %q from path %q (expected %q)", result, path, trimmed)
				}
			}
		}
	})
}

func TestExtractUpstreamFromPath(t *testing.T) {
	tests := []struct {
		path  string
		query string
		want  string
	}{
		{"/https://example.com/file.md", "", "https://example.com/file.md"},
		{"/https://s3.internal/file.md", "X-Amz-Signature=abc", "https://s3.internal/file.md?X-Amz-Signature=abc"},
		{"/http://server/doc.md", "", "http://server/doc.md"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			got := ExtractUpstreamFromPath(tc.path, tc.query)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
