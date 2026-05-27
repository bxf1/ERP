package gateway

import (
	"strings"
	"testing"
)

func TestInjectTenantFilterWithWhere(t *testing.T) {
	sql := "SELECT * FROM orders WHERE status = 'active'"
	result, err := InjectTenantFilter(sql, "t-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "tenant_id = 't-001'") {
		t.Errorf("tenant filter not injected, got: %s", result)
	}

	if !strings.HasPrefix(result, "SELECT") {
		t.Errorf("result should start with SELECT, got: %s", result)
	}
}

func TestInjectTenantFilterNoWhere(t *testing.T) {
	sql := "SELECT * FROM orders ORDER BY created_at DESC"
	result, err := InjectTenantFilter(sql, "t-002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "tenant_id = 't-002'") {
		t.Errorf("tenant filter not injected, got: %s", result)
	}
}

func TestInjectTenantFilterSimpleSelect(t *testing.T) {
	sql := "SELECT COUNT(*) FROM orders"
	result, err := InjectTenantFilter(sql, "t-003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "tenant_id = 't-003'") {
		t.Errorf("tenant filter not injected, got: %s", result)
	}
}

func TestInjectTenantFilterEmptySQL(t *testing.T) {
	_, err := InjectTenantFilter("", "t1")
	if err == nil {
		t.Error("expected error for empty SQL")
	}
}

func TestInjectTenantFilterEmptyTenant(t *testing.T) {
	_, err := InjectTenantFilter("SELECT 1", "")
	if err == nil {
		t.Error("expected error for empty tenant_id")
	}
}

func TestInjectTenantFilterRejectsDelete(t *testing.T) {
	_, err := InjectTenantFilter("DELETE FROM users WHERE id = 1", "t1")
	if err == nil {
		t.Error("DELETE should be rejected")
	}
}

func TestInjectTenantFilterRejectsInsert(t *testing.T) {
	_, err := InjectTenantFilter("INSERT INTO users (name) VALUES ('test')", "t1")
	if err == nil {
		t.Error("INSERT should be rejected")
	}
}

func TestInjectTenantFilterRejectsDrop(t *testing.T) {
	_, err := InjectTenantFilter("DROP TABLE users", "t1")
	if err == nil {
		t.Error("DROP should be rejected")
	}
}

func TestInjectTenantFilterRejectsSqlComment(t *testing.T) {
	_, err := InjectTenantFilter("SELECT * FROM users -- dangerous", "t1")
	if err == nil {
		t.Error("SQL with comments should be rejected")
	}
}

func TestPermissionCheck(t *testing.T) {
	audit := NewAuditLogger()
	gw := NewSecurityGateway(audit)

	gw.GrantPermission("user-1", PermReadModels)

	if err := gw.CheckPermission("t1", "user-1", PermReadModels); err != nil {
		t.Errorf("user-1 should have read permission: %v", err)
	}

	if err := gw.CheckPermission("t1", "user-1", PermWriteModels); err == nil {
		t.Error("user-1 should NOT have write permission")
	}

	if err := gw.CheckPermission("t1", "unknown-user", PermReadModels); err == nil {
		t.Error("unknown user should be denied")
	}
}

func TestAuditLogging(t *testing.T) {
	audit := NewAuditLogger()
	gw := NewSecurityGateway(audit)

	gw.LogAction("t1", "user-1", "test_action", "test_resource", "details", "success")
	gw.LogAction("t1", "user-1", "another_action", "another_resource", "", "failed")

	entries := audit.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(entries))
	}

	if entries[0].Action != "test_action" {
		t.Errorf("expected 'test_action', got %s", entries[0].Action)
	}

	if entries[1].Result != "failed" {
		t.Errorf("expected 'failed', got %s", entries[1].Result)
	}
}

func TestEscapeSQL(t *testing.T) {
	if s := escapeSQL("test'value"); s != "test''value" {
		t.Errorf("expected 'test''value', got %s", s)
	}
}
