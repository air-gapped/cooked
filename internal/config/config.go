package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Config holds all runtime configuration for cooked.
type Config struct {
	Listen           string
	CacheTTL         time.Duration
	CacheMaxSize     int64
	FetchTimeout     time.Duration
	MaxFileSize      int64
	AllowedUpstreams string
	BaseURL          string
	DefaultTheme     string
	TLSSkipVerify    bool
}

// Parse reads configuration from CLI flags with environment variable fallback.
func Parse(args []string) (*Config, error) {
	fs := flag.NewFlagSet("cooked", flag.ContinueOnError)

	cfg := &Config{}

	fs.StringVar(&cfg.Listen, "listen", envOr("COOKED_LISTEN", "127.0.0.1:8080"), "Listen address")
	fs.DurationVar(&cfg.CacheTTL, "cache-ttl", envDurationOr("COOKED_CACHE_TTL", 5*time.Minute), "Cache TTL duration")
	cacheMaxSize := fs.String("cache-max-size", envOr("COOKED_CACHE_MAX_SIZE", "100MB"), "Max cache size (e.g. 100MB)")
	fs.DurationVar(&cfg.FetchTimeout, "fetch-timeout", envDurationOr("COOKED_FETCH_TIMEOUT", 30*time.Second), "Upstream fetch timeout")
	maxFileSize := fs.String("max-file-size", envOr("COOKED_MAX_FILE_SIZE", "5MB"), "Max file size to render (e.g. 5MB)")
	fs.StringVar(&cfg.AllowedUpstreams, "allowed-upstreams", envOr("COOKED_ALLOWED_UPSTREAMS", ""), "Comma-separated allowed upstreams: hostnames, *.wildcard, or CIDR ranges (e.g. \"cgit.internal,*.corp,10.0.0.0/8\")")
	fs.StringVar(&cfg.BaseURL, "base-url", envOr("COOKED_BASE_URL", ""), "Public base URL of cooked (auto-detect from Host header if empty)")
	fs.StringVar(&cfg.DefaultTheme, "default-theme", envOr("COOKED_DEFAULT_THEME", "auto"), "Default theme: auto, light, or dark")
	fs.BoolVar(&cfg.TLSSkipVerify, "tls-skip-verify", envBoolOr("COOKED_TLS_SKIP_VERIFY", false), "Disable TLS certificate verification for upstream fetches")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	var err error
	cfg.CacheMaxSize, err = parseByteSize(*cacheMaxSize)
	if err != nil {
		return nil, fmt.Errorf("parse cache-max-size: %w", err)
	}

	cfg.MaxFileSize, err = parseByteSize(*maxFileSize)
	if err != nil {
		return nil, fmt.Errorf("parse max-file-size: %w", err)
	}

	if err := validateAllowedUpstreams(cfg.AllowedUpstreams); err != nil {
		return nil, fmt.Errorf("invalid allowed-upstreams: %w", err)
	}

	switch cfg.DefaultTheme {
	case "auto", "light", "dark":
	default:
		return nil, fmt.Errorf("invalid default-theme %q: must be auto, light, or dark", cfg.DefaultTheme)
	}

	return cfg, nil
}

// validateAllowedUpstreams checks that all entries in the comma-separated
// allowlist are well-formed at startup time.
func validateAllowedUpstreams(raw string) error {
	if raw == "" {
		return nil
	}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		switch {
		case strings.Contains(entry, "/"):
			if _, _, err := net.ParseCIDR(entry); err != nil {
				return fmt.Errorf("invalid CIDR %q: %w", entry, err)
			}
		case strings.HasPrefix(entry, "*."):
			if len(entry) <= 2 {
				return fmt.Errorf("empty wildcard suffix in %q", entry)
			}
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envDurationOr(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func envBoolOr(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		return v == "1" || v == "true" || v == "yes"
	}
	return fallback
}

// parseByteSize parses a human-readable byte size like "100MB", "5KB", "1GB".
func parseByteSize(s string) (int64, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("empty size string")
	}

	// Find where the numeric part ends
	i := 0
	for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.') {
		i++
	}

	numStr := s[:i]
	unit := s[i:]

	var num float64
	if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}

	var multiplier int64
	switch unit {
	case "", "B":
		multiplier = 1
	case "KB", "kb":
		multiplier = 1024
	case "MB", "mb":
		multiplier = 1024 * 1024
	case "GB", "gb":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown size unit %q in %q", unit, s)
	}

	return int64(num * float64(multiplier)), nil
}
