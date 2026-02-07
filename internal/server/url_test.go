package server

import (
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

		// Exact prefix match
		{"cgit.internal", "cgit.internal", true},
		{"s3.internal", "cgit.internal,s3.internal", true},

		// Prefix matching
		{"cgit.internal.example.com", "cgit.internal", true},

		// Not in list
		{"evil.com", "cgit.internal,s3.internal", false},

		// With port
		{"cgit.internal:8080", "cgit.internal", true},

		// Whitespace in allowed list
		{"cgit.internal", " cgit.internal , s3.internal ", true},
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
