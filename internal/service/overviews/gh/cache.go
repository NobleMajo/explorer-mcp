package gh

import (
	"sync"
	"time"
)

type rawCache struct {
	mu      sync.Mutex
	entries map[string]rawCacheEntry
}

type rawCacheEntry struct {
	raw       string
	expiresAt time.Time
}

func newRawCache() *rawCache {
	return &rawCache{entries: make(map[string]rawCacheEntry)}
}

func (c *rawCache) getOrFetch(key string, ttl time.Duration, fetch func() (string, error)) (string, error) {
	now := time.Now()

	c.mu.Lock()
	if entry, ok := c.entries[key]; ok && now.Before(entry.expiresAt) {
		raw := entry.raw
		c.mu.Unlock()
		return raw, nil
	}
	c.mu.Unlock()

	raw, err := fetch()
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.entries[key] = rawCacheEntry{
		raw:       raw,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()

	return raw, nil
}

var (
	orgListCache  = newRawCache()
	repoListCache = newRawCache()
)

const (
	orgListCacheTTL  = 5 * time.Minute
	repoListCacheTTL = 3 * time.Minute
)
