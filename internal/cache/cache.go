package cache

import (
	"container/list"
	"sync"
	"time"
)

// Status represents the cache lookup result.
type Status string

const (
	StatusHit         Status = "hit"
	StatusMiss        Status = "miss"
	StatusRevalidated Status = "revalidated"
	StatusExpired     Status = "expired"
)

// Entry holds a cached rendered page.
type Entry struct {
	HTML         []byte
	ETag         string
	LastModified string
	Size         int64
	ContentType  string
	ExpiresAt    time.Time
}

// Cache is a thread-safe, in-memory LRU cache with TTL and byte-counting eviction.
type Cache struct {
	mu      sync.Mutex
	items   map[string]*list.Element
	order   *list.List
	ttl     time.Duration
	maxSize int64
	curSize int64
	now     func() time.Time // injectable for testing
}

type cacheItem struct {
	key   string
	entry Entry
}

// New creates a cache with the given TTL and max size in bytes.
func New(ttl time.Duration, maxSize int64) *Cache {
	return &Cache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		ttl:     ttl,
		maxSize: maxSize,
		now:     time.Now,
	}
}

// Get retrieves a cached entry. Returns the entry, true if found (may be expired).
// The status indicates hit, miss, or expired.
func (c *Cache) Get(key string) (*Entry, Status) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, StatusMiss
	}

	item := elem.Value.(*cacheItem)

	// Check expiry
	if c.now().After(item.entry.ExpiresAt) {
		return &item.entry, StatusExpired
	}

	// Move to front (most recently used)
	c.order.MoveToFront(elem)
	return &item.entry, StatusHit
}

// Put stores an entry in the cache. Evicts LRU entries if necessary.
func (c *Cache) Put(key string, entry Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry.ExpiresAt = c.now().Add(c.ttl)

	// Update existing
	if elem, ok := c.items[key]; ok {
		old := elem.Value.(*cacheItem)
		c.curSize -= old.entry.Size
		old.entry = entry
		c.curSize += entry.Size
		c.order.MoveToFront(elem)
		c.evict()
		return
	}

	// Insert new
	item := &cacheItem{key: key, entry: entry}
	elem := c.order.PushFront(item)
	c.items[key] = elem
	c.curSize += entry.Size

	c.evict()
}

// RefreshTTL resets the TTL for an existing entry (used after 304 revalidation).
func (c *Cache) RefreshTTL(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		item := elem.Value.(*cacheItem)
		item.entry.ExpiresAt = c.now().Add(c.ttl)
		c.order.MoveToFront(elem)
	}
}

// evict removes LRU entries until curSize <= maxSize. Must be called with mu held.
func (c *Cache) evict() {
	for c.curSize > c.maxSize && c.order.Len() > 0 {
		oldest := c.order.Back()
		if oldest == nil {
			break
		}
		item := oldest.Value.(*cacheItem)
		c.curSize -= item.entry.Size
		delete(c.items, item.key)
		c.order.Remove(oldest)
	}
}

// Len returns the number of cached entries.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Size returns the current byte size of the cache.
func (c *Cache) Size() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.curSize
}
