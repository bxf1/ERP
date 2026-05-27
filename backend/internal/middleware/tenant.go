package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bxf1/ERP/backend/internal/database"
	apperrors "github.com/bxf1/ERP/backend/internal/errors"
	"github.com/bxf1/ERP/backend/internal/response"
)

const (
	// TenantKey is the gin context key for the resolved tenant.
	TenantKey = "tenant"
)

// TenantConfig holds tenant middleware configuration.
type TenantConfig struct {
	// Resolver is the tenant resolver.
	Resolver *database.TenantResolver
	// HeaderName is the HTTP header to read the tenant slug from (default: X-Tenant).
	HeaderName string
	// PathPrefix extracts tenant from the first URL segment when true (e.g., /{tenant}/api/...).
	UsePathPrefix bool
}

// Tenant is a Gin middleware that resolves the current tenant.
// Resolution order: path prefix > X-Tenant header > subdomain.
func Tenant(cfg TenantConfig) gin.HandlerFunc {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Tenant"
	}

	return func(c *gin.Context) {
		var slug string

		// 1. Try path prefix: /{tenant-slug}/...
		if cfg.UsePathPrefix {
			segments := strings.SplitN(strings.TrimPrefix(c.Request.URL.Path, "/"), "/", 2)
			if len(segments) > 0 {
				slug = segments[0]
			}
		}

		// 2. Try header
		if slug == "" {
			slug = c.GetHeader(cfg.HeaderName)
		}

		// 3. Try subdomain (rough extraction)
		if slug == "" {
			host := c.Request.Host
			host = strings.SplitN(host, ":", 1)[0]
			parts := strings.SplitN(host, ".", 2)
			if len(parts) > 1 && len(parts[0]) > 0 {
				slug = parts[0]
			}
		}

		if slug == "" {
			response.Error(c, apperrors.New(apperrors.CodeBadRequest, "tenant identifier is required"))
			return
		}

		tenant, err := cfg.Resolver.ResolveBySlug(slug)
		if err != nil {
			response.Error(c, apperrors.New(apperrors.CodeInternalError, "failed to resolve tenant"))
			return
		}

		if tenant == nil {
			response.Error(c, apperrors.New(apperrors.CodeTenantNotFound, "tenant not found: "+slug))
			return
		}

		c.Set(TenantKey, tenant)
		c.Next()
	}
}

// GetTenant extracts the resolved tenant from the gin context.
func GetTenant(c *gin.Context) *database.Tenant {
	if t, ok := c.Get(TenantKey); ok {
		if tenant, ok := t.(*database.Tenant); ok {
			return tenant
		}
	}
	return nil
}
