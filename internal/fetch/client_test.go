package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newTestClient creates a fetch client with SSRF protection disabled,
// so tests can connect to httptest servers on 127.0.0.1.
func newTestClient(timeout time.Duration, maxFileSize int64) *Client {
	return NewClient(timeout, maxFileSize, false, WithSSRFProtection(false))
}

func TestFetch_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Thu, 06 Feb 2026 12:00:00 GMT")
		w.Write([]byte("# Hello World"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	result, err := c.Fetch(upstream.URL+"/README.md", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if string(result.Body) != "# Hello World" {
		t.Errorf("Body = %q, want # Hello World", string(result.Body))
	}
	if result.ETag != `"abc123"` {
		t.Errorf("ETag = %q, want \"abc123\"", result.ETag)
	}
	if result.LastModified != "Thu, 06 Feb 2026 12:00:00 GMT" {
		t.Errorf("LastModified = %q", result.LastModified)
	}
	if result.FetchMs < 0 {
		t.Errorf("FetchMs = %d, want >= 0", result.FetchMs)
	}
}

func TestFetch_ConditionalGet304(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"abc123"` {
			w.WriteHeader(304)
			return
		}
		w.Write([]byte("content"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	result, err := c.Fetch(upstream.URL+"/file.md", `"abc123"`, "")
	if err != nil {
		t.Fatal(err)
	}

	if result.StatusCode != 304 {
		t.Errorf("StatusCode = %d, want 304", result.StatusCode)
	}
	if len(result.Body) != 0 {
		t.Errorf("Body should be empty for 304, got %d bytes", len(result.Body))
	}
}

func TestFetch_FileTooLarge_ContentLength(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10000000")
		w.Write([]byte("big"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 1024) // 1KB limit
	_, err := c.Fetch(upstream.URL+"/big.md", "", "")
	if err == nil {
		t.Error("expected error for large file, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %v, want 'too large'", err)
	}
}

func TestFetch_FileTooLarge_StreamingBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't set Content-Length, stream data
		w.Header().Set("Transfer-Encoding", "chunked")
		for i := 0; i < 100; i++ {
			fmt.Fprint(w, strings.Repeat("x", 100))
		}
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 1024) // 1KB limit
	_, err := c.Fetch(upstream.URL+"/big.md", "", "")
	if err == nil {
		t.Error("expected error for large streamed file, got nil")
	}
}

func TestFetch_NoCredentialForwarding(t *testing.T) {
	var gotAuth string
	var gotCookie string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	_, err := c.Fetch(upstream.URL+"/file.md", "", "")
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "" {
		t.Errorf("Authorization header was forwarded: %q", gotAuth)
	}
	if gotCookie != "" {
		t.Errorf("Cookie header was forwarded: %q", gotCookie)
	}
}

func TestFetch_Timeout(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("slow"))
	}))
	defer upstream.Close()

	c := newTestClient(100*time.Millisecond, 5*1024*1024)
	_, err := c.Fetch(upstream.URL+"/slow.md", "", "")
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestFetch_UpstreamNon200(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	result, err := c.Fetch(upstream.URL+"/missing.md", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

// F-03: DNS TOCTOU — the custom DialContext blocks connections to private IPs.
func TestFetch_SSRFDialBlock(t *testing.T) {
	// httptest.NewServer binds to 127.0.0.1, which is a loopback address.
	// With SSRF protection enabled (default), the dial guard blocks it.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not reach here"))
	}))
	defer upstream.Close()

	c := NewClient(5*time.Second, 5*1024*1024, false) // SSRF enabled by default
	_, err := c.Fetch(upstream.URL+"/secret", "", "")
	if err == nil {
		t.Fatal("expected SSRF dial block error, got nil")
	}
	if !strings.Contains(err.Error(), "blocked IP") {
		t.Errorf("unexpected error: %v, want 'blocked IP'", err)
	}
}

// F-01: Redirect to max limit and redirect validator.
func TestFetch_RedirectChainExceedsMax(t *testing.T) {
	redirectCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}))
	defer srv.Close()

	// Use SSRF disabled so we can reach the loopback test server.
	c := NewClient(5*time.Second, 5*1024*1024, false, WithSSRFProtection(false))
	_, err := c.Fetch(srv.URL+"/loop", "", "")
	if err == nil {
		t.Fatal("expected too many redirects error, got nil")
	}
	if !strings.Contains(err.Error(), "too many redirects") {
		t.Errorf("unexpected error: %v, want 'too many redirects'", err)
	}
}

// F-01: Redirect validator blocks redirect to disallowed host.
func TestFetch_RedirectValidatorBlocks(t *testing.T) {
	// Server A redirects to Server B.
	srvB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not reach B"))
	}))
	defer srvB.Close()

	srvA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srvB.URL+"/evil", http.StatusFound)
	}))
	defer srvA.Close()

	// Validator blocks any host that isn't srvA's host.
	srvAHost := srvA.Listener.Addr().String()
	validator := func(target *url.URL) error {
		if target.Host != srvAHost {
			return fmt.Errorf("host %q not allowed", target.Host)
		}
		return nil
	}

	c := NewClient(5*time.Second, 5*1024*1024, false,
		WithSSRFProtection(false),
		WithRedirectValidator(validator),
	)
	_, err := c.Fetch(srvA.URL+"/start", "", "")
	if err == nil {
		t.Fatal("expected redirect validator error, got nil")
	}
	if !strings.Contains(err.Error(), "redirect blocked") {
		t.Errorf("unexpected error: %v, want 'redirect blocked'", err)
	}
}
