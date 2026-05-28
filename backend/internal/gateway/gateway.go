// Package gateway implements the security gateway for MCP tool calls:
// permission verification, audit logging, tenant-isolated SQL filtering,
// and write-operation confirmation.
package gateway

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Permission represents a granted permission.
type Permission string

const (
	PermReadModels   Permission = "models:read"
	PermWriteModels  Permission = "models:write"
	PermQueryData    Permission = "data:query"
	PermReadSemantic Permission = "semantic:read"
)

// AuditEntry is a single audit log record.
type AuditEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	Result    string    `json:"result"`
}

// AuditLogger records audit entries to an in-memory store.
type AuditLogger struct {
	entries []AuditEntry
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{entries: make([]AuditEntry, 0)}
}

// Log records an audit entry.
func (a *AuditLogger) Log(tenantID, userID, action, resource, details, result string) {
	entry := AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		TenantID:  tenantID,
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   details,
		Result:    result,
	}
	a.entries = append(a.entries, entry)
	log.Printf("[AUDIT] tenant=%s user=%s action=%s resource=%s result=%s", tenantID, userID, action, resource, result)
}

// Entries returns all audit entries.
func (a *AuditLogger) Entries() []AuditEntry {
	return a.entries
}

// SecurityGateway orchestrates permission checks, audit logging,
// write confirmation, and SQL tenant-filter injection.
type SecurityGateway struct {
	audit *AuditLogger
	// In production, this would be backed by the RBAC permission engine.
	permissions map[string][]Permission // userID -> permissions
}

// NewSecurityGateway creates a new security gateway.
func NewSecurityGateway(audit *AuditLogger) *SecurityGateway {
	return &SecurityGateway{
		audit:       audit,
		permissions: make(map[string][]Permission),
	}
}

// GrantPermission assigns a permission to a user.
func (s *SecurityGateway) GrantPermission(userID string, perm Permission) {
	s.permissions[userID] = append(s.permissions[userID], perm)
}

// CheckPermission verifies that the user has the required permission.
// Returns nil if allowed, or an error if denied.
func (s *SecurityGateway) CheckPermission(tenantID, userID string, required Permission) error {
	perms, ok := s.permissions[userID]
	if !ok {
		s.audit.Log(tenantID, userID, "check_permission", string(required),
			"user has no permissions", "denied")
		return fmt.Errorf("permission denied: user %q has no permissions in tenant %q", userID, tenantID)
	}

	for _, p := range perms {
		if p == required {
			s.audit.Log(tenantID, userID, "check_permission", string(required),
				"permission granted", "allowed")
			return nil
		}
	}

	s.audit.Log(tenantID, userID, "check_permission", string(required),
		fmt.Sprintf("user has %v, needed %s", perms, required), "denied")
	return fmt.Errorf("permission denied: missing %s", required)
}

// LogAction records an audit entry for an operation.
func (s *SecurityGateway) LogAction(tenantID, userID, action, resource, details, result string) {
	s.audit.Log(tenantID, userID, action, resource, details, result)
}

// InjectTenantFilter modifies a SQL query to include a tenant filter.
// It wraps the original query to ensure data isolation by tenant_id.
func InjectTenantFilter(sql, tenantID string) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required for query filtering")
	}

	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", fmt.Errorf("SQL query must not be empty")
	}

	// Reject dangerous SQL patterns (allowlist approach for read-only queries).
	upperSQL := strings.ToUpper(sql)
	if !strings.HasPrefix(upperSQL, "SELECT") {
		return "", fmt.Errorf("only SELECT queries are allowed via query_data, got %q", sql)
	}

	dangerousWords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE",
		"TRUNCATE", "EXEC", "EXECUTE",
	}
	wordBoundary := regexp.MustCompile(`\b(` + strings.Join(dangerousWords, "|") + `)\b`)
	if wordBoundary.MatchString(upperSQL) {
		return "", fmt.Errorf("SQL contains forbidden keyword: %s", wordBoundary.FindString(upperSQL))
	}
	if strings.Contains(upperSQL, "--") || strings.Contains(upperSQL, "/*") {
		return "", fmt.Errorf("SQL contains forbidden comment syntax")
	}

	// Inject tenant_id = '<tenantID>' into WHERE clause.
	injected, err := injectWhereClause(sql, tenantID)
	if err != nil {
		return "", fmt.Errorf("tenant filter injection failed: %w", err)
	}

	return injected, nil
}

var whereRegex = regexp.MustCompile(`(?i)\bWHERE\b`)

func injectWhereClause(sql, tenantID string) (string, error) {
	tenantCondition := fmt.Sprintf("tenant_id = '%s'", escapeSQL(tenantID))

	if whereRegex.MatchString(sql) {
		return whereRegex.ReplaceAllString(sql, "WHERE "+tenantCondition+" AND "), nil
	}

	// No existing WHERE clause: look for GROUP BY, ORDER BY, LIMIT, etc.
	insertPoints := []string{"GROUP BY", "ORDER BY", "LIMIT", "HAVING"}
	upperSQL := strings.ToUpper(sql)
	earliest := len(sql)

	for _, kw := range insertPoints {
		idx := strings.Index(upperSQL, kw)
		if idx >= 0 && idx < earliest {
			earliest = idx
		}
	}

	if earliest < len(sql) {
		return sql[:earliest] + " WHERE " + tenantCondition + " " + sql[earliest:], nil
	}

	return sql + " WHERE " + tenantCondition, nil
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
