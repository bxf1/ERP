package datadict

import (
	"context"
	"log"
	"sync"
	"time"
)

// SchemaCache holds an in-memory copy of the latest DataDict and supports
// TTL-based refresh and manual invalidation so callers always get a fast,
// reasonably fresh snapshot without hitting information_schema on every request.
type SchemaCache struct {
	mu       sync.RWMutex
	dict     *DataDict
	ext      *Extractor
	ttl      time.Duration
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewSchemaCache creates a cache that auto-refreshes on every call to Get
// if the cached copy is older than ttl.
func NewSchemaCache(ext *Extractor, ttl time.Duration) *SchemaCache {
	return &SchemaCache{
		ext:    ext,
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
}

// Get returns the cached DataDict, refreshing it if the entry is stale.
func (c *SchemaCache) Get(ctx context.Context) (*DataDict, error) {
	c.mu.RLock()
	valid := c.dict != nil && time.Since(c.dict.ExtractedAt) < c.ttl
	c.mu.RUnlock()

	if valid {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return c.dict, nil
	}

	return c.Refresh(ctx)
}

// Refresh forces an immediate extraction and replaces the cached copy.
func (c *SchemaCache) Refresh(ctx context.Context) (*DataDict, error) {
	dict, err := c.ext.Extract(ctx)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.dict = dict
	c.mu.Unlock()

	return dict, nil
}

// StartPeriodicRefresh launches a background goroutine that refreshes the
// schema at the given interval. Call Stop to shut it down.
func (c *SchemaCache) StartPeriodicRefresh(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, err := c.Refresh(ctx); err != nil {
					log.Printf("datadict: periodic refresh failed: %v", err)
				}
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop shuts down the periodic refresh goroutine.
func (c *SchemaCache) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopCh)
	})
}

// Diff compares the cached schema against a freshly extracted one and reports
// additions, removals, and modifications.
func (c *SchemaCache) Diff(ctx context.Context) (*SchemaDiff, error) {
	old, err := c.Get(ctx)
	if err != nil {
		return nil, err
	}

	newDict, err := c.ext.Extract(ctx)
	if err != nil {
		return nil, err
	}

	return computeDiff(old, newDict), nil
}

func computeDiff(old, new *DataDict) *SchemaDiff {
	oldTables := make(map[string]TableInfo, len(old.Tables))
	newTables := make(map[string]TableInfo, len(new.Tables))

	for _, t := range old.Tables {
		oldTables[tableKey(t.Schema, t.Name)] = t
	}
	for _, t := range new.Tables {
		newTables[tableKey(t.Schema, t.Name)] = t
	}

	diff := &SchemaDiff{}

	for key := range newTables {
		if _, ok := oldTables[key]; !ok {
			diff.AddedTables = append(diff.AddedTables, key)
		}
	}
	for key := range oldTables {
		if _, ok := newTables[key]; !ok {
			diff.RemovedTables = append(diff.RemovedTables, key)
		}
	}

	for key, oldT := range oldTables {
		newT, ok := newTables[key]
		if !ok {
			continue
		}
		// Compare columns.
		oldCols := make(map[string]ColumnInfo, len(oldT.Columns))
		newCols := make(map[string]ColumnInfo, len(newT.Columns))
		for _, c := range oldT.Columns {
			oldCols[c.Name] = c
		}
		for _, c := range newT.Columns {
			newCols[c.Name] = c
		}

		modified := false
		for name, newC := range newCols {
			oldC, exists := oldCols[name]
			if !exists {
				diff.ColumnChanges = append(diff.ColumnChanges, ColumnDiff{
					TableName: key, ColumnName: name, ChangeType: "added",
				})
				modified = true
				continue
			}
			if oldC.DataType != newC.DataType || oldC.Nullable != newC.Nullable {
				diff.ColumnChanges = append(diff.ColumnChanges, ColumnDiff{
					TableName: key, ColumnName: name, ChangeType: "modified",
					OldValue: oldC.DataType, NewValue: newC.DataType,
				})
				modified = true
			}
		}
		for name := range oldCols {
			if _, exists := newCols[name]; !exists {
				diff.ColumnChanges = append(diff.ColumnChanges, ColumnDiff{
					TableName: key, ColumnName: name, ChangeType: "removed",
				})
				modified = true
			}
		}
		if modified {
			diff.ModifiedTables = append(diff.ModifiedTables, key)
		}
	}

	return diff
}
