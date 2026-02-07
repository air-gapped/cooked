package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/air-gapped/cooked/internal/config"
)

// newIntegrationServer creates a test server with configurable options.
// By default it allows localhost upstreams and has generous limits.
func newIntegrationServer(t *testing.T, opts ...func(*config.Config)) (*httptest.Server, func()) {
	t.Helper()

	cfg := &config.Config{
		Listen:           ":0",
		CacheTTL:         5 * time.Minute,
		CacheMaxSize:     100 * 1024 * 1024,
		FetchTimeout:     10 * time.Second,
		MaxFileSize:      5 * 1024 * 1024,
		DefaultTheme:     "auto",
		AllowedUpstreams: "127.0.0.1",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	assets := fstest.MapFS{
		"mermaid.min.js":            {Data: []byte("// mermaid stub")},
		"github-markdown-light.css": {Data: []byte("/* light */")},
		"github-markdown-dark.css":  {Data: []byte("/* dark */")},
	}

	s := New(cfg, "v0.1.0-test", assets)
	srv := httptest.NewServer(s.Handler())
	return srv, srv.Close
}

// serveFixture creates an upstream httptest.Server that serves the given fixture file.
func serveFixture(t *testing.T, fixturePath string) *httptest.Server {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixturePath, err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(data)
	}))
}

func fixtureDir() string {
	// Tests run from the package directory; testdata is at repo root.
	return filepath.Join("..", "..", "testdata", "fixtures")
}

func getBody(t *testing.T, url string) (int, http.Header, string) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, string(body)
}

// --- Full pipeline: markdown with fixture files ---

func TestIntegration_MarkdownBasic(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "markdown", "basic.md"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/basic.md")

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}

	// Headers
	if got := headers.Get("X-Cooked-Content-Type"); got != "markdown" {
		t.Errorf("X-Cooked-Content-Type = %q, want markdown", got)
	}
	if got := headers.Get("X-Cooked-Cache"); got != "miss" {
		t.Errorf("X-Cooked-Cache = %q, want miss", got)
	}
	if got := headers.Get("X-Cooked-Version"); got != "v0.1.0-test" {
		t.Errorf("X-Cooked-Version = %q, want v0.1.0-test", got)
	}
	if got := headers.Get("X-Cooked-Upstream-Status"); got != "200" {
		t.Errorf("X-Cooked-Upstream-Status = %q, want 200", got)
	}

	// DOM structure
	for _, id := range []string{`id="cooked-header"`, `id="cooked-content"`} {
		if !strings.Contains(body, id) {
			t.Errorf("missing %s in rendered page", id)
		}
	}

	// Rendered content
	if !strings.Contains(body, "<h1") {
		t.Error("missing <h1> for '# Hello World'")
	}
	if !strings.Contains(body, "<strong>bold</strong>") {
		t.Error("bold text not rendered")
	}
	if !strings.Contains(body, "<em>italic</em>") {
		t.Error("italic text not rendered")
	}
	if !strings.Contains(body, "<code>inline code</code>") {
		t.Error("inline code not rendered")
	}
}

func TestIntegration_MarkdownWithTOC(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "markdown", "headings_toc.md"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	_, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/headings_toc.md")

	// headings_toc.md has many headings — should have TOC
	if !strings.Contains(body, `data-has-toc="true"`) {
		t.Error("expected data-has-toc=true for document with many headings")
	}
}

func TestIntegration_MarkdownWithMermaid(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "markdown", "mermaid.md"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	_, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/mermaid.md")

	// Should include mermaid.js script reference
	if !strings.Contains(body, "mermaid") {
		t.Error("mermaid content not found in rendered page")
	}
}

// --- Full pipeline: code rendering ---

func TestIntegration_CodePython(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "code", "sample.py"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/sample.py")

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "code" {
		t.Errorf("X-Cooked-Content-Type = %q, want code", got)
	}

	// Chroma produces span elements with classes for syntax highlighting
	if !strings.Contains(body, "<span") {
		t.Error("expected syntax-highlighted spans in code output")
	}
	if !strings.Contains(body, `id="cooked-content"`) {
		t.Error("missing #cooked-content")
	}
}

func TestIntegration_CodeGo(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "code", "sample.go"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, _ := getBody(t, srv.URL+"/"+upstream.URL+"/sample.go")

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "code" {
		t.Errorf("X-Cooked-Content-Type = %q, want code", got)
	}
}

// --- Full pipeline: MDX rendering ---

func TestIntegration_MDX(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "mdx", "docusaurus.mdx"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/getting-started.mdx")

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "mdx" {
		t.Errorf("X-Cooked-Content-Type = %q, want mdx", got)
	}

	// MDX should still render markdown headings
	if !strings.Contains(body, "<h1") {
		t.Error("MDX: missing <h1>")
	}
	// JSX import/export statements should be stripped by MDX preprocessor
	if strings.Contains(body, "import Tabs from") {
		t.Error("MDX: import statement was not stripped")
	}
}

// --- Full pipeline: plaintext rendering ---

func TestIntegration_Plaintext(t *testing.T) {
	upstream := serveFixture(t, filepath.Join(fixtureDir(), "plaintext", "sample.txt"))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/sample.txt")

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "plaintext" {
		t.Errorf("X-Cooked-Content-Type = %q, want plaintext", got)
	}
	if !strings.Contains(body, "<pre") {
		t.Error("plaintext should be wrapped in <pre>")
	}
}

// --- Cache behavior ---

func TestIntegration_CacheMissToHit(t *testing.T) {
	fetchCount := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Write([]byte("# Cached Content\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	url := srv.URL + "/" + upstream.URL + "/README.md"

	// First request — miss
	_, h1, body1 := getBody(t, url)
	if got := h1.Get("X-Cooked-Cache"); got != "miss" {
		t.Errorf("first request: X-Cooked-Cache = %q, want miss", got)
	}

	// Second request — hit (no upstream fetch)
	_, h2, body2 := getBody(t, url)
	if got := h2.Get("X-Cooked-Cache"); got != "hit" {
		t.Errorf("second request: X-Cooked-Cache = %q, want hit", got)
	}

	// Body should be identical (served from cache)
	if body1 != body2 {
		t.Error("cached response body differs from original")
	}

	// Upstream should only be hit once
	if fetchCount != 1 {
		t.Errorf("upstream fetched %d times, want 1", fetchCount)
	}
}

// --- SSRF protection ---

func TestIntegration_SSRFPrivateRanges(t *testing.T) {
	srv, cleanup := newIntegrationServer(t, func(cfg *config.Config) {
		cfg.AllowedUpstreams = "" // SSRF protection active
	})
	defer cleanup()

	privateURLs := []string{
		"/http://127.0.0.1/README.md",
		"/http://10.0.0.1/README.md",
		"/http://192.168.1.1/README.md",
		"/http://172.16.0.1/README.md",
	}
	for _, path := range privateURLs {
		t.Run(path, func(t *testing.T) {
			status, _, body := getBody(t, srv.URL+path)
			if status != 403 {
				t.Errorf("status = %d, want 403", status)
			}
			if !strings.Contains(body, `id="cooked-error"`) {
				t.Error("SSRF block should render error page")
			}
		})
	}
}

// --- Allowed upstreams ---

func TestIntegration_AllowedUpstreams(t *testing.T) {
	srv, cleanup := newIntegrationServer(t, func(cfg *config.Config) {
		cfg.AllowedUpstreams = "trusted.example.com,internal.corp"
	})
	defer cleanup()

	// Blocked host
	status, _, _ := getBody(t, srv.URL+"/https://evil.attacker.com/README.md")
	if status != 403 {
		t.Errorf("blocked host: status = %d, want 403", status)
	}

	// Blocked host (different)
	status2, _, _ := getBody(t, srv.URL+"/https://other.com/file.md")
	if status2 != 403 {
		t.Errorf("blocked host: status = %d, want 403", status2)
	}
}

func TestIntegration_AllowedUpstreamPrefixMatch(t *testing.T) {
	// Start a real upstream so we get 200 (not 502)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# OK\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t, func(cfg *config.Config) {
		cfg.AllowedUpstreams = "127.0.0.1"
	})
	defer cleanup()

	status, _, _ := getBody(t, srv.URL+"/"+upstream.URL+"/README.md")
	if status != 200 {
		t.Errorf("allowed host: status = %d, want 200", status)
	}
}

// --- File size limit ---

func TestIntegration_FileSizeLimit(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10485760") // 10 MB
		// Write more than the limit
		w.Write(make([]byte, 10*1024*1024))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t, func(cfg *config.Config) {
		cfg.MaxFileSize = 1024 // 1 KB limit
	})
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/big.md")
	if status != 413 {
		t.Errorf("status = %d, want 413", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "error" {
		t.Errorf("X-Cooked-Content-Type = %q, want error", got)
	}
	if !strings.Contains(body, "too-large") {
		t.Error("error page should indicate file too large")
	}
}

// --- Unsupported file type ---

func TestIntegration_UnsupportedFileType(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{0xFF, 0xD8, 0xFF}) // fake binary
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/image.png")
	if status != 415 {
		t.Errorf("status = %d, want 415", status)
	}
	if !strings.Contains(body, "unsupported") {
		t.Error("error page should mention unsupported file type")
	}
}

// --- Upstream errors ---

func TestIntegration_UpstreamNotFound(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+"/missing.md")
	if status != 404 {
		t.Errorf("status = %d, want 404", status)
	}
	if got := headers.Get("X-Cooked-Content-Type"); got != "error" {
		t.Errorf("X-Cooked-Content-Type = %q, want error", got)
	}
	if !strings.Contains(body, `id="cooked-error"`) {
		t.Error("missing #cooked-error on error page")
	}
	if !strings.Contains(body, `data-error-type="upstream-error"`) {
		t.Error("missing data-error-type=upstream-error")
	}
}

func TestIntegration_UpstreamServerError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/broken.md")
	if status != 500 {
		t.Errorf("status = %d, want 500", status)
	}
	if !strings.Contains(body, `id="cooked-error"`) {
		t.Error("missing error page structure")
	}
}

// --- Invalid URL schemes ---

func TestIntegration_InvalidSchemes(t *testing.T) {
	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	schemes := []string{
		"/ftp://example.com/file.md",
		"/file:///etc/passwd",
		"/javascript:alert(1)",
	}
	for _, path := range schemes {
		t.Run(path, func(t *testing.T) {
			status, _, _ := getBody(t, srv.URL+path)
			if status != 400 {
				t.Errorf("status = %d, want 400", status)
			}
		})
	}
}

// --- HTML sanitization through the full pipeline ---

func TestIntegration_SanitizesScriptTags(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Safe Content\n\n<script>alert('xss')</script>\n\nMore text.\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	status, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/evil.md")
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if strings.Contains(body, "alert('xss')") {
		t.Error("script tag content was not sanitized")
	}
	if !strings.Contains(body, "Safe Content") {
		t.Error("safe content should be preserved")
	}
	if !strings.Contains(body, "More text") {
		t.Error("text after script should be preserved")
	}
}

func TestIntegration_SanitizesEventHandlers(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Page\n\n<div onmouseover=\"alert('xss')\">hover me</div>\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	_, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/handler.md")
	if strings.Contains(body, "onmouseover") {
		t.Error("event handler attribute was not stripped")
	}
	if !strings.Contains(body, "hover me") {
		t.Error("element content should be preserved")
	}
}

// --- X-Cooked-* headers comprehensive check ---

func TestIntegration_AllResponseHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Thu, 06 Feb 2026 12:00:00 GMT")
		w.Write([]byte("# Heading\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	_, headers, _ := getBody(t, srv.URL+"/"+upstream.URL+"/test.md")

	required := []string{
		"X-Cooked-Version",
		"X-Cooked-Upstream",
		"X-Cooked-Upstream-Status",
		"X-Cooked-Cache",
		"X-Cooked-Content-Type",
		"X-Cooked-Render-Ms",
		"X-Cooked-Upstream-Ms",
		"Content-Type",
		"Cache-Control",
	}
	for _, h := range required {
		if headers.Get(h) == "" {
			t.Errorf("missing required header %s", h)
		}
	}

	if got := headers.Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html prefix", got)
	}
	if got := headers.Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q, want 'public, max-age=300'", got)
	}
}

// --- Theme configuration ---

func TestIntegration_DefaultThemeInOutput(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Themed\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t, func(cfg *config.Config) {
		cfg.DefaultTheme = "dark"
	})
	defer cleanup()

	_, _, body := getBody(t, srv.URL+"/"+upstream.URL+"/themed.md")
	if !strings.Contains(body, "dark") {
		t.Error("expected dark theme reference in output")
	}
}

// --- Multiple file types served by same server ---

func TestIntegration_MultipleFileTypes(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readme.md":
			w.Write([]byte("# README\n"))
		case "/main.py":
			w.Write([]byte("print('hello')\n"))
		case "/notes.txt":
			w.Write([]byte("plain text notes\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	tests := []struct {
		path        string
		wantType    string
		wantContain string
	}{
		{"/readme.md", "markdown", "<h1"},
		{"/main.py", "code", "<span"},
		{"/notes.txt", "plaintext", "<pre"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			status, headers, body := getBody(t, srv.URL+"/"+upstream.URL+tc.path)
			if status != 200 {
				t.Fatalf("status = %d, want 200", status)
			}
			if got := headers.Get("X-Cooked-Content-Type"); got != tc.wantType {
				t.Errorf("X-Cooked-Content-Type = %q, want %q", got, tc.wantType)
			}
			if !strings.Contains(body, tc.wantContain) {
				t.Errorf("body missing expected content %q", tc.wantContain)
			}
		})
	}
}

// --- Upstream URL passed correctly in headers ---

func TestIntegration_UpstreamURLInHeader(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Test\n"))
	}))
	defer upstream.Close()

	srv, cleanup := newIntegrationServer(t)
	defer cleanup()

	expectedUpstream := upstream.URL + "/path/to/doc.md"
	_, headers, _ := getBody(t, srv.URL+"/"+expectedUpstream)

	got := headers.Get("X-Cooked-Upstream")
	if got != expectedUpstream {
		t.Errorf("X-Cooked-Upstream = %q, want %q", got, expectedUpstream)
	}
}
