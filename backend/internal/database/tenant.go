package database

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/internal/logger"
	"go.uber.org/zap"
)

// Tenant represents a tenant record from the public.tenants table.
type Tenant struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	SchemaName string `json:"schema_name"`
	Status     string `json:"status"`
}

// TenantResolver resolves a tenant from a request identifier (hostname, header, or path param).
type TenantResolver struct {
	db  *sql.DB
	cfg config.DatabaseConfig
}

// NewTenantResolver creates a new resolver backed by the system database.
func NewTenantResolver(db *sql.DB, cfg config.DatabaseConfig) *TenantResolver {
	return &TenantResolver{db: db, cfg: cfg}
}

// ResolveBySlug looks up a tenant by its URL slug.
// If the tenant is not yet in the connection pool, registers it.
func (r *TenantResolver) ResolveBySlug(slug string) (*Tenant, error) {
	t, err := r.lookup("slug = $1", slug)
	if err != nil {
		return nil, err
	}

	if err := RegisterTenant(t.ID, r.cfg, t.SchemaName); err != nil {
		logger.L.Warn("failed to register tenant connection",
			zap.String("tenant_id", t.ID),
			zap.Error(err),
		)
	}

	return t, nil
}

// ResolveByID looks up a tenant by its UUID.
func (r *TenantResolver) ResolveByID(id string) (*Tenant, error) {
	t, err := r.lookup("id = $1", id)
	if err != nil {
		return nil, err
	}

	if err := RegisterTenant(t.ID, r.cfg, t.SchemaName); err != nil {
		logger.L.Warn("failed to register tenant connection",
			zap.String("tenant_id", t.ID),
			zap.Error(err),
		)
	}

	return t, nil
}

func (r *TenantResolver) lookup(where string, arg any) (*Tenant, error) {
	query := fmt.Sprintf(
		"SELECT id, name, slug, schema_name, status FROM tenants WHERE %s AND status = 'active'",
		where,
	)

	var t Tenant
	err := r.db.QueryRow(query, arg).Scan(&t.ID, &t.Name, &t.Slug, &t.SchemaName, &t.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}

	return &t, nil
}
