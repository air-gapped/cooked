package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/air-gapped/cooked/internal/config"
)

func newTestServer(t *testing.T, cfg *config.Config) *Server {
	t.Helper()
	if cfg == nil {
		cfg = &config.Config{
			Listen:           ":8080",
			CacheTTL:         5 * time.Minute,
			CacheMaxSize:     100 * 1024 * 1024,
			FetchTimeout:     10 * time.Second,
			MaxFileSize:      5 * 1024 * 1024,
			DefaultTheme:     "auto",
			AllowedUpstreams: "127.0.0.0/8", // Allow loopback CIDR for tests
		}
	}

	assets := fstest.MapFS{
		"mermaid.min.js":            {Data: []byte("// mermaid mock")},
		"github-markdown-light.css": {Data: []byte("/* light */")},
		"github-markdown-dark.css":  {Data: []byte("/* dark */")},
	}

	// When AllowedUpstreams is set, SSRF dial protection is automatically
	// disabled by server.New(). No extra fetch options needed.
	return New(cfg, "v0.1.0-test", assets)
}

func TestHealthz(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("body = %q, want OK", string(body))
	}
}

func TestLandingPage(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "cooked") {
		t.Error("landing page missing 'cooked' text")
	}
	if !strings.Contains(string(body), `href="/_cooked/docs"`) {
		t.Error("landing page missing docs link")
	}
}

func TestRenderMarkdown(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Last-Modified", "Thu, 06 Feb 2026 12:00:00 GMT")
		w.Write([]byte("# Hello World\n\nThis is a test.\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/README.md")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}

	// Check X-Cooked headers
	headers := map[string]string{
		"X-Cooked-Version":      "v0.1.0-test",
		"X-Cooked-Cache":        "miss",
		"X-Cooked-Content-Type": "markdown",
	}
	for k, want := range headers {
		if got := resp.Header.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// Check content rendered
	if !strings.Contains(html, "Hello World") {
		t.Error("rendered HTML missing 'Hello World'")
	}

	// Check required structure
	if !strings.Contains(html, `id="cooked-header"`) {
		t.Error("missing #cooked-header")
	}
	if !strings.Contains(html, `id="cooked-content"`) {
		t.Error("missing #cooked-content")
	}
}

func TestRenderCode(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("def hello():\n    print('world')\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/script.py")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Cooked-Content-Type"); got != "code" {
		t.Errorf("X-Cooked-Content-Type = %q, want code", got)
	}
}

func TestCacheHitMiss(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Hello\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	url := srv.URL + "/" + upstream.URL + "/README.md"

	// First request — miss
	resp1, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	if got := resp1.Header.Get("X-Cooked-Cache"); got != "miss" {
		t.Errorf("first request cache = %q, want miss", got)
	}

	// Second request — hit
	resp2, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if got := resp2.Header.Get("X-Cooked-Cache"); got != "hit" {
		t.Errorf("second request cache = %q, want hit", got)
	}
}

func TestBlockedUpstream(t *testing.T) {
	s := newTestServer(t, &config.Config{
		Listen:           ":8080",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "allowed.internal",
	})
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/https://evil.com/README.md")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestSSRFProtection(t *testing.T) {
	s := newTestServer(t, &config.Config{
		Listen:       ":8080",
		CacheTTL:     5 * time.Minute,
		CacheMaxSize: 100 * 1024 * 1024,
		FetchTimeout: 10 * time.Second,
		MaxFileSize:  5 * 1024 * 1024,
		DefaultTheme: "auto",
		// No AllowedUpstreams — SSRF protection is active
	})
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/http://127.0.0.1:8080/secret")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("status = %d, want 403 for SSRF", resp.StatusCode)
	}
}

// Regression test: private IPs must work when allowlisted via CIDR.
// This is the key use case for air-gapped intranets.
func TestAllowlistedPrivateIP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Internal Doc\n"))
	}))
	defer upstream.Close()

	// Use CIDR that covers 127.0.0.0/8 — the httptest server binds to 127.0.0.1.
	// No explicit WithSSRFProtection(false) — the allowlist should handle it.
	s := newTestServer(t, &config.Config{
		Listen:           ":8080",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "127.0.0.0/8",
	})
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/README.md")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200 (private IP should be allowed via CIDR); body: %s", resp.StatusCode, body)
	}
}

// Test wildcard allowlist blocks non-matching hosts.
func TestWildcardAllowlistBlocks(t *testing.T) {
	s := newTestServer(t, &config.Config{
		Listen:           ":8080",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "*.internal",
	})
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/https://evil.com/README.md")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("status = %d, want 403 for host not matching *.internal", resp.StatusCode)
	}
}

// Test CIDR allowlist blocks IPs outside the range.
func TestCIDRAllowlistBlocks(t *testing.T) {
	s := newTestServer(t, &config.Config{
		Listen:           ":8080",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "10.0.0.0/8",
	})
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/http://11.0.0.1:8080/README.md")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 403 {
		t.Errorf("status = %d, want 403 for IP outside CIDR", resp.StatusCode)
	}
}

func TestInvalidURL(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ftp://example.com/file")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestEmbeddedAsset(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantType   string
	}{
		{"javascript", "/_cooked/mermaid.min.js", 200, "application/javascript"},
		{"css light", "/_cooked/github-markdown-light.css", 200, "text/css"},
		{"css dark", "/_cooked/github-markdown-dark.css", 200, "text/css"},
		{"not found", "/_cooked/nonexistent.txt", 404, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + tc.path)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.wantStatus)
			}
			if tc.wantType != "" {
				if got := resp.Header.Get("Content-Type"); got != tc.wantType {
					t.Errorf("Content-Type = %q, want %q", got, tc.wantType)
				}
			}
		})
	}
}

func TestErrorPageStructure(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ftp://bad/url")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, `id="cooked-error"`) {
		t.Error("error page missing #cooked-error")
	}
	if !strings.Contains(html, `data-error-type=`) {
		t.Error("error page missing data-error-type")
	}
	if !strings.Contains(html, `data-status-code=`) {
		t.Error("error page missing data-status-code")
	}
}

func TestCacheControlHeader(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Hello\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/README.md")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=300" {
		t.Errorf("Cache-Control = %q, want 'public, max-age=300'", cc)
	}
}

// F-10: Security response headers
func TestSecurityHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Hello\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/README.md")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	tests := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "no-referrer",
		"X-Frame-Options":        "DENY",
	}
	for header, want := range tests {
		if got := resp.Header.Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

// F-10: Security headers on error responses too
func TestSecurityHeaders_ErrorResponse(t *testing.T) {
	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ftp://bad/url")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options on error = %q, want nosniff", got)
	}
}

// F-09: Redact upstream URLs — query params stripped from header
func TestDocsPage(t *testing.T) {
	cfg := &config.Config{
		Listen:           ":8080",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "127.0.0.0/8",
	}

	assets := fstest.MapFS{
		"mermaid.min.js":            {Data: []byte("// mermaid mock")},
		"github-markdown-light.css": {Data: []byte("/* light */")},
		"github-markdown-dark.css":  {Data: []byte("/* dark */")},
		"project-readme.md":         {Data: []byte("# Project Docs\n\nWelcome to cooked.\n")},
	}

	s := New(cfg, "v0.1.0-test", assets)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/_cooked/docs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}

	cc := resp.Header.Get("Cache-Control")
	if cc != "public, max-age=86400" {
		t.Errorf("Cache-Control = %q, want 'public, max-age=86400'", cc)
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "Project Docs") {
		t.Error("docs page missing rendered content 'Project Docs'")
	}
	if !strings.Contains(html, "Welcome to cooked") {
		t.Error("docs page missing rendered content 'Welcome to cooked'")
	}
}

func TestDocsPageMissingReadme(t *testing.T) {
	s := newTestServer(t, nil) // no project-readme.md in test assets
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/_cooked/docs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404 when README not embedded", resp.StatusCode)
	}
}

func TestUpstreamURLRedaction(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Test\n"))
	}))
	defer upstream.Close()

	s := newTestServer(t, nil)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/" + upstream.URL + "/file.md?token=secret&key=abc")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	got := resp.Header.Get("X-Cooked-Upstream")
	if strings.Contains(got, "token") || strings.Contains(got, "secret") || strings.Contains(got, "key=abc") {
		t.Errorf("X-Cooked-Upstream contains sensitive query: %q", got)
	}
	// Should still contain the base URL without query
	if !strings.Contains(got, "/file.md") {
		t.Errorf("X-Cooked-Upstream missing path: %q", got)
	}
}
