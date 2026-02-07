package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCache_PutGet(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	c.Put("key1", Entry{HTML: []byte("<h1>Hello</h1>"), Size: 14, ETag: `"abc"`})

	entry, status := c.Get("key1")
	if status != StatusHit {
		t.Errorf("status = %q, want hit", status)
	}
	if string(entry.HTML) != "<h1>Hello</h1>" {
		t.Errorf("HTML = %q", string(entry.HTML))
	}
	if entry.ETag != `"abc"` {
		t.Errorf("ETag = %q", entry.ETag)
	}
}

func TestCache_Miss(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	entry, status := c.Get("nonexistent")
	if status != StatusMiss {
		t.Errorf("status = %q, want miss", status)
	}
	if entry != nil {
		t.Error("entry should be nil for miss")
	}
}

func TestCache_TTLExpiry(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	now := time.Now()
	c.now = func() time.Time { return now }

	c.Put("key1", Entry{HTML: []byte("data"), Size: 4})

	// Advance time past TTL
	c.now = func() time.Time { return now.Add(6 * time.Minute) }

	entry, status := c.Get("key1")
	if status != StatusExpired {
		t.Errorf("status = %q, want expired", status)
	}
	if entry == nil {
		t.Error("expired entries should still be returned for revalidation")
	}
}

func TestCache_RefreshTTL(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	now := time.Now()
	c.now = func() time.Time { return now }

	c.Put("key1", Entry{HTML: []byte("data"), Size: 4})

	// Advance to near expiry
	c.now = func() time.Time { return now.Add(4 * time.Minute) }

	// Refresh TTL (as if after 304 revalidation)
	c.RefreshTTL("key1")

	// Advance to what would have been past original expiry
	c.now = func() time.Time { return now.Add(6 * time.Minute) }

	_, status := c.Get("key1")
	if status != StatusHit {
		t.Errorf("after refresh, status = %q, want hit", status)
	}
}

func TestCache_LRUEviction(t *testing.T) {
	// Max 100 bytes
	c := New(5*time.Minute, 100)

	c.Put("a", Entry{HTML: []byte("aaa"), Size: 40})
	c.Put("b", Entry{HTML: []byte("bbb"), Size: 40})
	c.Put("c", Entry{HTML: []byte("ccc"), Size: 40})

	// "a" should be evicted (LRU), cache has "b" and "c"
	_, status := c.Get("a")
	if status != StatusMiss {
		t.Errorf("'a' should be evicted, got status %q", status)
	}

	_, status = c.Get("b")
	if status != StatusHit {
		t.Errorf("'b' should still be cached, got status %q", status)
	}
}

func TestCache_LRUEviction_AccessOrder(t *testing.T) {
	c := New(5*time.Minute, 100)

	c.Put("a", Entry{HTML: []byte("aaa"), Size: 40})
	c.Put("b", Entry{HTML: []byte("bbb"), Size: 40})

	// Access "a" to make it recently used
	c.Get("a")

	// Adding "c" should evict "b" (least recently used), not "a"
	c.Put("c", Entry{HTML: []byte("ccc"), Size: 40})

	_, status := c.Get("a")
	if status != StatusHit {
		t.Error("'a' was accessed recently and should not be evicted")
	}

	_, status = c.Get("b")
	if status != StatusMiss {
		t.Error("'b' should be evicted as LRU")
	}
}

func TestCache_UpdateExisting(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	c.Put("key1", Entry{HTML: []byte("old"), Size: 3})
	c.Put("key1", Entry{HTML: []byte("new data"), Size: 8})

	entry, status := c.Get("key1")
	if status != StatusHit {
		t.Errorf("status = %q, want hit", status)
	}
	if string(entry.HTML) != "new data" {
		t.Errorf("HTML = %q, want new data", string(entry.HTML))
	}

	if c.Size() != 8 {
		t.Errorf("Size = %d, want 8 (updated entry size)", c.Size())
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i%10)
			c.Put(key, Entry{HTML: []byte("data"), Size: 4})
			c.Get(key)
		}(i)
	}
	wg.Wait()

	if c.Len() > 10 {
		t.Errorf("Len = %d, expected <= 10", c.Len())
	}
}

func TestCache_KeyIsFullURL(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	// Same path, different query strings should be different cache keys
	c.Put("https://example.com/file.md?token=abc", Entry{HTML: []byte("abc"), Size: 3})
	c.Put("https://example.com/file.md?token=xyz", Entry{HTML: []byte("xyz"), Size: 3})

	entry1, status1 := c.Get("https://example.com/file.md?token=abc")
	if status1 != StatusHit {
		t.Errorf("status = %q, want hit", status1)
	}
	if string(entry1.HTML) != "abc" {
		t.Errorf("HTML = %q, want abc", string(entry1.HTML))
	}

	entry2, status2 := c.Get("https://example.com/file.md?token=xyz")
	if status2 != StatusHit {
		t.Errorf("status = %q, want hit", status2)
	}
	if string(entry2.HTML) != "xyz" {
		t.Errorf("HTML = %q, want xyz", string(entry2.HTML))
	}

	// URL without query string is a separate key
	_, status3 := c.Get("https://example.com/file.md")
	if status3 != StatusMiss {
		t.Errorf("bare URL status = %q, want miss", status3)
	}

	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}

func TestCache_SizeAccounting(t *testing.T) {
	c := New(5*time.Minute, 1024*1024)

	c.Put("a", Entry{Size: 100})
	c.Put("b", Entry{Size: 200})

	if c.Size() != 300 {
		t.Errorf("Size = %d, want 300", c.Size())
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}
