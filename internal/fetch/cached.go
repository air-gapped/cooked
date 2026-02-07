package fetch

import (
	"github.com/air-gapped/cooked/internal/cache"
)

// CachedClient wraps a fetch Client with an in-memory cache.
type CachedClient struct {
	client *Client
	cache  *cache.Cache
}

// CachedResult extends Result with cache status.
type CachedResult struct {
	*Result
	CacheStatus cache.Status
	RenderMs    int64 // set by caller after rendering
}

// NewCachedClient creates a fetch client with caching.
func NewCachedClient(client *Client, c *cache.Cache) *CachedClient {
	return &CachedClient{client: client, cache: c}
}

// Fetch retrieves content from upstream, using the cache when possible.
// On cache hit with ETag/LastModified, performs conditional GET.
// On 304, serves from cache and resets TTL.
// On miss/expired, fetches fresh content.
// The caller is responsible for rendering and storing the result in the cache.
func (cc *CachedClient) Fetch(rawURL string) (*CachedResult, *cache.Entry, error) {
	// Check cache
	entry, status := cc.cache.Get(rawURL)

	switch status {
	case cache.StatusHit:
		return &CachedResult{
			Result:      &Result{StatusCode: 200, FetchMs: 0},
			CacheStatus: cache.StatusHit,
		}, entry, nil

	case cache.StatusExpired:
		// Attempt revalidation with conditional GET
		result, err := cc.client.Fetch(rawURL, entry.ETag, entry.LastModified)
		if err != nil {
			// On error, serve stale cache
			return &CachedResult{
				Result:      &Result{StatusCode: 200, FetchMs: 0},
				CacheStatus: cache.StatusHit,
			}, entry, nil
		}

		if result.StatusCode == 304 {
			cc.cache.RefreshTTL(rawURL)
			return &CachedResult{
				Result:      result,
				CacheStatus: cache.StatusRevalidated,
			}, entry, nil
		}

		// Got fresh content
		return &CachedResult{
			Result:      result,
			CacheStatus: cache.StatusExpired,
		}, nil, nil

	default: // miss
		result, err := cc.client.Fetch(rawURL, "", "")
		if err != nil {
			return nil, nil, err
		}

		return &CachedResult{
			Result:      result,
			CacheStatus: cache.StatusMiss,
		}, nil, nil
	}
}

// Store caches a rendered page entry.
func (cc *CachedClient) Store(key string, entry cache.Entry) {
	cc.cache.Put(key, entry)
}

// Cache returns the underlying cache for direct access.
func (cc *CachedClient) Cache() *cache.Cache {
	return cc.cache
}
