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
			AllowedUpstreams: "127.0.0.1", // Allow localhost for tests
		}
	}

	assets := fstest.MapFS{
		"mermaid.min.js":            {Data: []byte("// mermaid mock")},
		"github-markdown-light.css": {Data: []byte("/* light */")},
		"github-markdown-dark.css":  {Data: []byte("/* dark */")},
	}

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

	resp, err := http.Get(srv.URL + "/_cooked/mermaid.min.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/javascript" {
		t.Errorf("Content-Type = %q, want application/javascript", got)
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
