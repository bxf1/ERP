package database

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"

	"github.com/bxf1/ERP/backend/config"
)

var (
	poolMu   sync.RWMutex
	tenantDB = make(map[string]*sql.DB)
)

// Connect opens a connection to the primary database.
func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// RegisterTenant associates a tenant schema's sql.DB with a tenant ID.
// The connection is opened with search_path set to the tenant's schema.
func RegisterTenant(tenantID string, cfg config.DatabaseConfig, schema string) error {
	poolMu.Lock()
	defer poolMu.Unlock()

	if _, exists := tenantDB[tenantID]; exists {
		return nil
	}

	connStr := cfg.DSN() + fmt.Sprintf(" search_path=%s,public", schema)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("open tenant db %s: %w", tenantID, err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping tenant db %s: %w", tenantID, err)
	}

	tenantDB[tenantID] = db
	return nil
}

// RemoveTenant removes a tenant's sql.DB from the pool.
func RemoveTenant(tenantID string) {
	poolMu.Lock()
	defer poolMu.Unlock()

	if db, ok := tenantDB[tenantID]; ok {
		db.Close()
		delete(tenantDB, tenantID)
	}
}

// GetTenantDB returns the sql.DB for a given tenant.
func GetTenantDB(tenantID string) (*sql.DB, bool) {
	poolMu.RLock()
	defer poolMu.RUnlock()
	db, ok := tenantDB[tenantID]
	return db, ok
}
