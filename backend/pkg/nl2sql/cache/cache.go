package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Entry holds a cached query result.
type Entry struct {
	Columns  []string
	Rows     []map[string]interface{}
	StoredAt time.Time
}

// QueryCache provides an in-memory cache for SQL query results.
type QueryCache struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	ttl     time.Duration
	maxSize int
}

// New creates a query cache with the given TTL and max entries.
func New(ttl time.Duration, maxSize int) *QueryCache {
	c := &QueryCache{
		entries: make(map[string]*Entry),
		ttl:     ttl,
		maxSize: maxSize,
	}
	go c.reapLoop()
	return c
}

// Get retrieves a cached result. Returns nil if not found or expired.
func (c *QueryCache) Get(key string) *Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil
	}
	if time.Since(entry.StoredAt) > c.ttl {
		return nil
	}
	return entry
}

// Set stores a query result in the cache.
func (c *QueryCache) Set(key string, columns []string, rows []map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictOne()
	}

	rowsCopy := make([]map[string]interface{}, len(rows))
	copy(rowsCopy, rows)

	c.entries[key] = &Entry{
		Columns:  columns,
		Rows:     rowsCopy,
		StoredAt: time.Now(),
	}
}

// Invalidate removes a specific cache entry.
func (c *QueryCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all cached entries.
func (c *QueryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*Entry)
}

// Size returns the current number of cached entries.
func (c *QueryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *QueryCache) evictOne() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range c.entries {
		if first || v.StoredAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.StoredAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *QueryCache) reapLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()
	for range ticker.C {
		c.reap()
	}
}

func (c *QueryCache) reap() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.entries {
		if now.Sub(v.StoredAt) > c.ttl {
			delete(c.entries, k)
		}
	}
}

// CacheKey computes a deterministic cache key from a SQL query and tenant ID.
func CacheKey(sql, tenantID string) string {
	h := sha256.New()
	h.Write([]byte(sql))
	h.Write([]byte(tenantID))
	return hex.EncodeToString(h.Sum(nil))[:32]
}
