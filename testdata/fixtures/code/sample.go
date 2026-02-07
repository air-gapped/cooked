// Package sample demonstrates Go syntax highlighting.
package sample

import (
	"context"
	"fmt"
	"sync"
)

// Cache is a concurrent-safe in-memory cache.
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// New creates a new Cache.
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

// Get retrieves a value from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

// Set stores a value in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Range iterates over all items, stopping if fn returns false.
func (c *Cache[K, V]) Range(ctx context.Context, fn func(K, V) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for k, v := range c.items {
		if ctx.Err() != nil {
			return
		}
		if !fn(k, v) {
			return
		}
	}
}

func Example() {
	c := New[string, int]()
	c.Set("answer", 42)

	if v, ok := c.Get("answer"); ok {
		fmt.Printf("answer = %d\n", v)
	}
}
