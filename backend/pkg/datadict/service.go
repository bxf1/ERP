package datadict

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Service is the public API of the data dictionary module.
// It wraps extraction, caching, and diffing behind a clean interface
// consumed by Gin handlers and the semantic layer.
type Service struct {
	cache *SchemaCache
	db    *sql.DB
}

// NewService creates a Service. ttl controls how long a cached schema
// is considered fresh before the next Get triggers a re-extraction.
func NewService(db *sql.DB, ttl time.Duration) *Service {
	ext := NewExtractor(db)
	cache := NewSchemaCache(ext, ttl)
	return &Service{cache: cache, db: db}
}

// GetSchema returns the current data dictionary, refreshing if stale.
func (s *Service) GetSchema(ctx context.Context) (*DataDict, error) {
	return s.cache.Get(ctx)
}

// RefreshSchema forces a re-extraction from the database.
func (s *Service) RefreshSchema(ctx context.Context) (*DataDict, error) {
	return s.cache.Refresh(ctx)
}

// GetTable returns a single table by fully qualified name (schema.table).
func (s *Service) GetTable(ctx context.Context, schema, table string) (*TableInfo, error) {
	dict, err := s.cache.Get(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range dict.Tables {
		if t.Schema == schema && t.Name == table {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("table %s.%s not found", schema, table)
}

// ListTables returns all table names, optionally filtered by schema.
func (s *Service) ListTables(ctx context.Context, schema string) ([]TableInfo, error) {
	dict, err := s.cache.Get(ctx)
	if err != nil {
		return nil, err
	}
	if schema == "" {
		return dict.Tables, nil
	}
	var filtered []TableInfo
	for _, t := range dict.Tables {
		if t.Schema == schema {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// GetRelations returns all foreign key relationships.
func (s *Service) GetRelations(ctx context.Context) ([]RelationInfo, error) {
	dict, err := s.cache.Get(ctx)
	if err != nil {
		return nil, err
	}
	return dict.Relations, nil
}

// DiffSchema compares the cached schema against the current database state.
func (s *Service) DiffSchema(ctx context.Context) (*SchemaDiff, error) {
	return s.cache.Diff(ctx)
}

// StartAutoRefresh begins periodic background schema extraction.
func (s *Service) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	s.cache.StartPeriodicRefresh(ctx, interval)
}

// StopAutoRefresh stops the background refresh goroutine.
func (s *Service) StopAutoRefresh() {
	s.cache.Stop()
}

// SchemaSummary returns a compact representation used as LLM context.
func (s *Service) SchemaSummary(ctx context.Context) (string, error) {
	dict, err := s.cache.Get(ctx)
	if err != nil {
		return "", err
	}

	var out string
	for _, t := range dict.Tables {
		out += fmt.Sprintf("Table: %s.%s", t.Schema, t.Name)
		if t.Comment != "" {
			out += fmt.Sprintf(" (%s)", t.Comment)
		}
		out += "\n"
		for _, c := range t.Columns {
			nullable := ""
			if c.Nullable {
				nullable = "?"
			}
			out += fmt.Sprintf("  - %s: %s%s", c.Name, c.DataType, nullable)
			if c.Comment != "" {
				out += fmt.Sprintf(" -- %s", c.Comment)
			}
			out += "\n"
		}
		out += "\n"
	}
	return out, nil
}
