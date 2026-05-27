package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/internal/logger"
	"go.uber.org/zap"
)

// Migration represents a single migration file.
type Migration struct {
	Version  int
	Name     string
	FilePath string
}

// LoadMigrations reads and sorts migration files from a directory.
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		// File format: 001_name.sql
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(parts[1], ".sql")
		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			FilePath: filepath.Join(dir, e.Name()),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// RunMigrations applies pending migrations in order.
func RunMigrations(db *sql.DB, dir string) error {
	migrations, err := LoadMigrations(dir)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		applied, err := isMigrationApplied(db, "public", m.Version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if applied {
			logger.L.Debug("migration already applied", zap.Int("version", m.Version), zap.String("name", m.Name))
			continue
		}

		logger.L.Info("applying migration", zap.Int("version", m.Version), zap.String("name", m.Name))

		content, err := os.ReadFile(m.FilePath)
		if err != nil {
			return fmt.Errorf("read migration %d: %w", m.Version, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec migration %d (%s): %w", m.Version, m.Name, err)
		}

		logger.L.Info("migration applied", zap.Int("version", m.Version), zap.String("name", m.Name))
	}

	return nil
}

// MigrateTenantSchemas applies tenant-level migrations to all active tenants.
// This is called after system-level migrations to ensure all tenant schemas are up-to-date.
func MigrateTenantSchemas(db *sql.DB, cfg config.DatabaseConfig) error {
	rows, err := db.Query(`SELECT id, schema_name FROM tenants WHERE status = 'active'`)
	if err != nil {
		return fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tenantID, schemaName string
		if err := rows.Scan(&tenantID, &schemaName); err != nil {
			return fmt.Errorf("scan tenant: %w", err)
		}

		// Check if tenant schema exists; if not, create and install tables
		var exists bool
		err = db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", schemaName,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check schema %s: %w", schemaName, err)
		}

		if !exists {
			logger.L.Info("creating tenant schema", zap.String("schema", schemaName))
			if _, err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)); err != nil {
				return fmt.Errorf("create schema %s: %w", schemaName, err)
			}
		}

		// Run tenant-level migration functions in order
		appliedVersion, err := getLastAppliedVersion(db, schemaName)
		if err != nil {
			return fmt.Errorf("get last version for %s: %w", schemaName, err)
		}

		// Apply tenant install functions based on what's missing
		if appliedVersion < 1 {
			logger.L.Info("installing tenant tables v1", zap.String("schema", schemaName))
			if _, err := db.Exec(`SELECT install_tenant_tables($1)`, schemaName); err != nil {
				return fmt.Errorf("install v1 for %s: %w", schemaName, err)
			}
		}
		if appliedVersion < 2 {
			logger.L.Info("installing tenant tables v2", zap.String("schema", schemaName))
			if _, err := db.Exec(`SELECT install_tenant_v2_tables($1)`, schemaName); err != nil {
				return fmt.Errorf("install v2 for %s: %w", schemaName, err)
			}
		}
		if appliedVersion < 3 {
			logger.L.Info("installing tenant tables v3", zap.String("schema", schemaName))
			if _, err := db.Exec(`SELECT install_tenant_v3_tables($1)`, schemaName); err != nil {
				return fmt.Errorf("install v3 for %s: %w", schemaName, err)
			}
		}
	}

	return rows.Err()
}

func isMigrationApplied(db *sql.DB, schemaName string, version int) (bool, error) {
	var count int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM schema_migrations WHERE schema_name = $1 AND version = $2",
		schemaName, version,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getLastAppliedVersion(db *sql.DB, schemaName string) (int, error) {
	var version int
	err := db.QueryRow(
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE schema_name = $1",
		schemaName,
	).Scan(&version)
	return version, err
}
