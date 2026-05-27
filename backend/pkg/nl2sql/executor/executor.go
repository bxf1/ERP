package executor

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

// Config holds retry and timeout settings.
type Config struct {
	MaxRetries      int           `json:"max_retries"`
	BaseBackoff     time.Duration `json:"base_backoff"`
	MaxBackoff      time.Duration `json:"max_backoff"`
	QueryTimeout    time.Duration `json:"query_timeout"`
	DegradeLimitRow int           `json:"degrade_limit_row"` // fallback row limit on repeated failure
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxRetries:      3,
		BaseBackoff:     200 * time.Millisecond,
		MaxBackoff:      5 * time.Second,
		QueryTimeout:    30 * time.Second,
		DegradeLimitRow: 100,
	}
}

// Executor runs validated SQL queries against the database with retry logic.
type Executor struct {
	db     *sql.DB
	config Config
}

// New creates a new Executor.
func New(db *sql.DB, config Config) *Executor {
	return &Executor{db: db, config: config}
}

// QueryResult holds the result of a successful query.
type QueryResult struct {
	Columns  []string
	Rows     []map[string]interface{}
	Duration time.Duration
}

// Execute runs a SQL query with retry and degradation.
func (e *Executor) Execute(ctx context.Context, sqlQuery string) (*QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, e.config.QueryTimeout)
	defer cancel()

	var lastErr error
	backoff := e.config.BaseBackoff

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
			backoff = time.Duration(math.Min(float64(backoff*2), float64(e.config.MaxBackoff)))
		}

		start := time.Now()
		result, err := e.runQuery(ctx, sqlQuery)
		if err == nil {
			result.Duration = time.Since(start)
			return result, nil
		}

		lastErr = err

		// On final attempt, try a degraded version.
		if attempt == e.config.MaxRetries {
			result, degradeErr := e.runDegraded(ctx, sqlQuery)
			if degradeErr == nil {
				result.Duration = time.Since(start)
				return result, nil
			}
			lastErr = fmt.Errorf("all retries exhausted (last: %w), degraded also failed: %w", lastErr, degradeErr)
		}
	}

	return nil, lastErr
}

func (e *Executor) runQuery(ctx context.Context, sqlQuery string) (*QueryResult, error) {
	rows, err := e.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	results, err := scanRows(rows, columns)
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		Columns: columns,
		Rows:    results,
	}, nil
}

func (e *Executor) runDegraded(ctx context.Context, sqlQuery string) (*QueryResult, error) {
	// Append a lower row limit to reduce load.
	degradedSQL := sqlQuery
	if len(degradedSQL) > 6 && degradedSQL[len(degradedSQL)-1] == ';' {
		degradedSQL = degradedSQL[:len(degradedSQL)-1]
	}
	// Replace existing LIMIT with a lower one if present, or append.
	if idx := lastIndexOfLimit(degradedSQL); idx >= 0 {
		degradedSQL = degradedSQL[:idx] + fmt.Sprintf(" LIMIT %d", e.config.DegradeLimitRow)
	} else {
		degradedSQL = fmt.Sprintf("%s LIMIT %d", degradedSQL, e.config.DegradeLimitRow)
	}
	degradedSQL += ";"

	ctx2, cancel := context.WithTimeout(ctx, e.config.QueryTimeout/2)
	defer cancel()

	return e.runQuery(ctx2, degradedSQL)
}

func scanRows(rows *sql.Rows, columns []string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("row scan: %w", err)
		}

		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			// Convert byte slices to strings for JSON-friendliness.
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return results, nil
}

func lastIndexOfLimit(sql string) int {
	upper := []byte(sql)
	for i := len(upper) - 1; i >= 0; i-- {
		if upper[i] >= 'a' && upper[i] <= 'z' {
			upper[i] -= 32
		}
	}
	// Simple substring search for "LIMIT" from end.
	for i := len(upper) - 5; i >= 0; i-- {
		if string(upper[i:i+5]) == "LIMIT" {
			return i
		}
	}
	return -1
}
