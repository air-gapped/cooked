package fetch

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/air-gapped/cooked/internal/cache"
)

func TestCachedClient_Miss(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("# Hello"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	cc := NewCachedClient(c, cache.New(5*time.Minute, 100*1024*1024))

	result, entry, err := cc.Fetch(upstream.URL + "/README.md")
	if err != nil {
		t.Fatal(err)
	}

	if result.CacheStatus != cache.StatusMiss {
		t.Errorf("CacheStatus = %q, want miss", result.CacheStatus)
	}
	if entry != nil {
		t.Error("entry should be nil on miss")
	}
	if string(result.Body) != "# Hello" {
		t.Errorf("Body = %q", string(result.Body))
	}
}

func TestCachedClient_Hit(t *testing.T) {
	fetchCount := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Write([]byte("# Hello"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	memCache := cache.New(5*time.Minute, 100*1024*1024)
	cc := NewCachedClient(c, memCache)

	url := upstream.URL + "/README.md"

	// First fetch — miss
	result1, _, err := cc.Fetch(url)
	if err != nil {
		t.Fatal(err)
	}
	if result1.CacheStatus != cache.StatusMiss {
		t.Fatalf("first fetch CacheStatus = %q, want miss", result1.CacheStatus)
	}

	// Store in cache
	cc.Store(url, cache.Entry{
		HTML: []byte("<h1>Hello</h1>"),
		Size: 14,
		ETag: `"abc"`,
	})

	// Second fetch — hit
	result2, entry, err := cc.Fetch(url)
	if err != nil {
		t.Fatal(err)
	}
	if result2.CacheStatus != cache.StatusHit {
		t.Errorf("second fetch CacheStatus = %q, want hit", result2.CacheStatus)
	}
	if entry == nil {
		t.Fatal("entry should not be nil on hit")
	}
	if string(entry.HTML) != "<h1>Hello</h1>" {
		t.Errorf("cached HTML = %q", string(entry.HTML))
	}
	if result2.FetchMs != 0 {
		t.Errorf("FetchMs = %d, want 0 for cache hit", result2.FetchMs)
	}

	// Should not have fetched again
	if fetchCount != 1 {
		t.Errorf("upstream was fetched %d times, want 1", fetchCount)
	}
}

func TestCachedClient_Revalidation304(t *testing.T) {
	fetchCount := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		if r.Header.Get("If-None-Match") == `"etag1"` {
			w.WriteHeader(304)
			return
		}
		w.Write([]byte("content"))
	}))
	defer upstream.Close()

	c := newTestClient(10*time.Second, 5*1024*1024)
	memCache := cache.New(1*time.Millisecond, 100*1024*1024) // very short TTL
	cc := NewCachedClient(c, memCache)

	url := upstream.URL + "/file.md"

	// Store with ETag
	cc.Store(url, cache.Entry{
		HTML: []byte("<p>cached</p>"),
		Size: 13,
		ETag: `"etag1"`,
	})

	// Wait for TTL to expire
	time.Sleep(5 * time.Millisecond)

	// Fetch — should revalidate
	result, entry, err := cc.Fetch(url)
	if err != nil {
		t.Fatal(err)
	}
	if result.CacheStatus != cache.StatusRevalidated {
		t.Errorf("CacheStatus = %q, want revalidated", result.CacheStatus)
	}
	if entry == nil {
		t.Fatal("entry should not be nil after revalidation")
	}
	if string(entry.HTML) != "<p>cached</p>" {
		t.Errorf("cached HTML = %q", string(entry.HTML))
	}
}

func TestCachedClient_RevalidationError_ServesStale(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("content"))
	}))

	c := newTestClient(10*time.Second, 5*1024*1024)
	memCache := cache.New(1*time.Millisecond, 100*1024*1024) // very short TTL
	cc := NewCachedClient(c, memCache)

	url := upstream.URL + "/file.md"

	// Store with ETag
	cc.Store(url, cache.Entry{
		HTML: []byte("<p>stale</p>"),
		Size: 12,
		ETag: `"etag1"`,
	})

	// Wait for TTL to expire
	time.Sleep(5 * time.Millisecond)

	// Shut down upstream so revalidation fails
	upstream.Close()

	// Fetch — should serve stale content
	result, entry, err := cc.Fetch(url)
	if err != nil {
		t.Fatal(err)
	}
	if result.CacheStatus != cache.StatusStale {
		t.Errorf("CacheStatus = %q, want stale", result.CacheStatus)
	}
	if entry == nil {
		t.Fatal("entry should not be nil when serving stale")
	}
	if string(entry.HTML) != "<p>stale</p>" {
		t.Errorf("cached HTML = %q, want <p>stale</p>", string(entry.HTML))
	}
	if result.FetchMs != 0 {
		t.Errorf("FetchMs = %d, want 0 for stale serve", result.FetchMs)
	}
}
