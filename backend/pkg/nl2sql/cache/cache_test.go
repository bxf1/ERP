package cache

import (
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	c := New(10*time.Minute, 100)

	cols := []string{"id", "name"}
	rows := []map[string]interface{}{
		{"id": int64(1), "name": "test"},
	}

	key := "test-key"
	c.Set(key, cols, rows)

	entry := c.Get(key)
	if entry == nil {
		t.Fatal("expected cache hit")
	}
	if len(entry.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(entry.Columns))
	}
	if len(entry.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(entry.Rows))
	}
}

func TestCacheExpiry(t *testing.T) {
	c := New(50*time.Millisecond, 100)

	c.Set("key", []string{"x"}, []map[string]interface{}{{"x": 1}})

	time.Sleep(100 * time.Millisecond)

	entry := c.Get("key")
	if entry != nil {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCacheEviction(t *testing.T) {
	c := New(10*time.Minute, 2)

	c.Set("key1", nil, nil)
	c.Set("key2", nil, nil)
	c.Set("key3", nil, nil) // should evict key1 (oldest)

	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := New(10*time.Minute, 100)

	c.Set("key", nil, nil)
	c.Invalidate("key")
	if c.Get("key") != nil {
		t.Error("expected nil after invalidation")
	}
}

func TestCacheKey_Deterministic(t *testing.T) {
	k1 := CacheKey("SELECT * FROM orders", "t-001")
	k2 := CacheKey("SELECT * FROM orders", "t-001")
	if k1 != k2 {
		t.Errorf("same inputs should produce same key: %s vs %s", k1, k2)
	}
}

func TestCacheKey_Different(t *testing.T) {
	k1 := CacheKey("SELECT * FROM orders", "t-001")
	k2 := CacheKey("SELECT * FROM orders", "t-002")
	if k1 == k2 {
		t.Error("different tenants should produce different keys")
	}
}
