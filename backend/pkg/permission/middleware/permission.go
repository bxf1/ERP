package middleware

import (
	"net/http"
	"strings"

	"github.com/bxf1/ERP/backend/pkg/permission/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionMiddleware checks API-level permissions.
// Usage: r.POST("/users", PermissionMiddleware(rbac, audit, "user:create"), handler)
func PermissionMiddleware(rbac *services.RBACService, audit *services.AuditService, requiredPerms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := extractUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "user not authenticated"})
			c.Abort()
			return
		}

		allowed, err := rbac.CheckPermissions(c.Request.Context(), userID, requiredPerms)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "permission check failed"})
			c.Abort()
			return
		}

		reason := ""
		if !allowed {
			reason = "missing required permissions: " + strings.Join(requiredPerms, ", ")
		}

		if audit != nil {
			audit.Log(
				c.Request.Context(),
				userID,
				c.Request.Method,
				c.Request.URL.Path,
				c.Param("id"),
				c.Request.URL.Path,
				c.ClientIP(),
				allowed,
				reason,
			)
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks if the user has ANY of the specified permissions.
func RequireAnyPermission(rbac *services.RBACService, audit *services.AuditService, permCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := extractUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		allowed, err := rbac.CheckAnyPermission(c.Request.Context(), userID, permCodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			c.Abort()
			return
		}

		if audit != nil {
			audit.Log(c.Request.Context(), userID, c.Request.Method,
				c.Request.URL.Path, c.Param("id"), c.Request.URL.Path,
				c.ClientIP(), allowed, "")
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// DataScopeMiddleware injects a data scope filter into the request context.
// Downstream handlers use the filter to scope DB queries.
func DataScopeMiddleware(dataScope *services.DataScopeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := extractUserID(c)
		if err != nil {
			c.Next()
			return
		}
		userCtx := extractUserContext(c)
		// Store for downstream use — the actual filter is applied per-model in handlers
		c.Set("data_scope_user_id", userID)
		c.Set("data_scope_user_context", userCtx)
		c.Next()
	}
}

// extractUserID extracts the authenticated user ID from the Gin context.
// Assumes auth middleware has set "user_id" in the context.
func extractUserID(c *gin.Context) (uuid.UUID, error) {
	// Try from context set by auth middleware
	if uid, exists := c.Get("user_id"); exists {
		switch v := uid.(type) {
		case uuid.UUID:
			return v, nil
		case string:
			return uuid.Parse(v)
		}
	}
	// Fallback: try X-User-ID header (for development / internal calls)
	header := c.GetHeader("X-User-ID")
	if header != "" {
		return uuid.Parse(header)
	}
	return uuid.Nil, nil
}

// extractUserContext extracts user context for data scope filtering.
func extractUserContext(c *gin.Context) map[string]interface{} {
	ctx := make(map[string]interface{})

	if deptID, exists := c.Get("department_id"); exists {
		ctx["department_id"] = deptID
	}
	if uid, exists := c.Get("user_id"); exists {
		ctx["user_id"] = uid
	}
	// Allow header overrides for development
	if deptHeader := c.GetHeader("X-Department-ID"); deptHeader != "" {
		ctx["department_id"] = deptHeader
	}

	return ctx
}
