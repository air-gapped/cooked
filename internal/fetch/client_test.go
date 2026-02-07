package fetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetch_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Thu, 06 Feb 2026 12:00:00 GMT")
		w.Write([]byte("# Hello World"))
	}))
	defer upstream.Close()

	c := NewClient(10*time.Second, 5*1024*1024, false)
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

	c := NewClient(10*time.Second, 5*1024*1024, false)
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

	c := NewClient(10*time.Second, 1024, false) // 1KB limit
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

	c := NewClient(10*time.Second, 1024, false) // 1KB limit
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

	c := NewClient(10*time.Second, 5*1024*1024, false)
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

	c := NewClient(100*time.Millisecond, 5*1024*1024, false)
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

	c := NewClient(10*time.Second, 5*1024*1024, false)
	result, err := c.Fetch(upstream.URL+"/missing.md", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}
