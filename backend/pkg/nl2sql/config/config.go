package config

import "time"

// Config holds the NL2SQL module configuration.
type Config struct {
	DBSchema     string        `json:"db_schema"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	CacheMaxSize int           `json:"cache_max_size"`
	MaxRetries   int           `json:"max_retries"`
	BaseBackoff  time.Duration `json:"base_backoff"`
	MaxBackoff   time.Duration `json:"max_backoff"`
	QueryTimeout time.Duration `json:"query_timeout"`
}

// DefaultConfig returns sensible module defaults.
func DefaultConfig() Config {
	return Config{
		DBSchema:     "public",
		CacheTTL:     5 * time.Minute,
		CacheMaxSize: 1000,
		MaxRetries:   3,
		BaseBackoff:  200 * time.Millisecond,
		MaxBackoff:   5 * time.Second,
		QueryTimeout: 30 * time.Second,
	}
}
