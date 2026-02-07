package config

import (
	"testing"
	"time"
)

func TestParse_Defaults(t *testing.T) {
	cfg, err := Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Listen != "127.0.0.1:8080" {
		t.Errorf("Listen = %q, want 127.0.0.1:8080", cfg.Listen)
	}
	if cfg.CacheTTL != 5*time.Minute {
		t.Errorf("CacheTTL = %v, want 5m", cfg.CacheTTL)
	}
	if cfg.CacheMaxSize != 100*1024*1024 {
		t.Errorf("CacheMaxSize = %d, want %d", cfg.CacheMaxSize, 100*1024*1024)
	}
	if cfg.FetchTimeout != 30*time.Second {
		t.Errorf("FetchTimeout = %v, want 30s", cfg.FetchTimeout)
	}
	if cfg.MaxFileSize != 5*1024*1024 {
		t.Errorf("MaxFileSize = %d, want %d", cfg.MaxFileSize, 5*1024*1024)
	}
	if cfg.AllowedUpstreams != "" {
		t.Errorf("AllowedUpstreams = %q, want empty", cfg.AllowedUpstreams)
	}
	if cfg.BaseURL != "" {
		t.Errorf("BaseURL = %q, want empty", cfg.BaseURL)
	}
	if cfg.DefaultTheme != "auto" {
		t.Errorf("DefaultTheme = %q, want auto", cfg.DefaultTheme)
	}
	if cfg.TLSSkipVerify {
		t.Error("TLSSkipVerify = true, want false")
	}
}

func TestParse_Flags(t *testing.T) {
	args := []string{
		"--listen", ":9090",
		"--cache-ttl", "10m",
		"--cache-max-size", "200MB",
		"--fetch-timeout", "15s",
		"--max-file-size", "10MB",
		"--allowed-upstreams", "cgit.internal,s3.internal",
		"--base-url", "https://cooked.example.com",
		"--default-theme", "dark",
		"--tls-skip-verify",
	}

	cfg, err := Parse(args)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Listen != ":9090" {
		t.Errorf("Listen = %q, want :9090", cfg.Listen)
	}
	if cfg.CacheTTL != 10*time.Minute {
		t.Errorf("CacheTTL = %v, want 10m", cfg.CacheTTL)
	}
	if cfg.CacheMaxSize != 200*1024*1024 {
		t.Errorf("CacheMaxSize = %d, want %d", cfg.CacheMaxSize, 200*1024*1024)
	}
	if cfg.FetchTimeout != 15*time.Second {
		t.Errorf("FetchTimeout = %v, want 15s", cfg.FetchTimeout)
	}
	if cfg.MaxFileSize != 10*1024*1024 {
		t.Errorf("MaxFileSize = %d, want %d", cfg.MaxFileSize, 10*1024*1024)
	}
	if cfg.AllowedUpstreams != "cgit.internal,s3.internal" {
		t.Errorf("AllowedUpstreams = %q, want cgit.internal,s3.internal", cfg.AllowedUpstreams)
	}
	if cfg.BaseURL != "https://cooked.example.com" {
		t.Errorf("BaseURL = %q, want https://cooked.example.com", cfg.BaseURL)
	}
	if cfg.DefaultTheme != "dark" {
		t.Errorf("DefaultTheme = %q, want dark", cfg.DefaultTheme)
	}
	if !cfg.TLSSkipVerify {
		t.Error("TLSSkipVerify = false, want true")
	}
}

func TestParse_EnvFallback(t *testing.T) {
	t.Setenv("COOKED_LISTEN", ":7070")
	t.Setenv("COOKED_CACHE_TTL", "2m")
	t.Setenv("COOKED_TLS_SKIP_VERIFY", "true")

	cfg, err := Parse([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Listen != ":7070" {
		t.Errorf("Listen = %q, want :7070", cfg.Listen)
	}
	if cfg.CacheTTL != 2*time.Minute {
		t.Errorf("CacheTTL = %v, want 2m", cfg.CacheTTL)
	}
	if !cfg.TLSSkipVerify {
		t.Error("TLSSkipVerify = false, want true")
	}
}

func TestParse_FlagOverridesEnv(t *testing.T) {
	t.Setenv("COOKED_LISTEN", ":7070")

	cfg, err := Parse([]string{"--listen", ":9090"})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Listen != ":9090" {
		t.Errorf("Listen = %q, want :9090 (flag should override env)", cfg.Listen)
	}
}

func TestParse_InvalidTheme(t *testing.T) {
	_, err := Parse([]string{"--default-theme", "neon"})
	if err == nil {
		t.Error("expected error for invalid theme, got nil")
	}
}

func TestParse_InvalidCacheMaxSize(t *testing.T) {
	_, err := Parse([]string{"--cache-max-size", "notasize"})
	if err == nil {
		t.Error("expected error for invalid cache-max-size, got nil")
	}
}

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100B", 100},
		{"1KB", 1024},
		{"5MB", 5 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"100MB", 100 * 1024 * 1024},
		{"100", 100},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := parseByteSize(tc.input)
			if err != nil {
				t.Fatalf("parseByteSize(%q) error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseByteSize(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseByteSize_Errors(t *testing.T) {
	tests := []string{
		"",
		"notasize",
		"100TB",
		"MB",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			_, err := parseByteSize(tc)
			if err == nil {
				t.Errorf("parseByteSize(%q) expected error, got nil", tc)
			}
		})
	}
}
