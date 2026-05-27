package security

import (
	"strings"
	"testing"
)

func TestValidateSQL_AcceptsSelect(t *testing.T) {
	_, err := ValidateSQL("SELECT * FROM orders", "t-001")
	if err != nil {
		t.Fatalf("expected valid SELECT, got: %v", err)
	}
}

func TestValidateSQL_RejectsNonSelect(t *testing.T) {
	tests := []string{
		"INSERT INTO orders VALUES (1)",
		"UPDATE orders SET x=1",
		"DELETE FROM orders",
		"DROP TABLE orders",
		"ALTER TABLE orders ADD COLUMN x INT",
		"TRUNCATE orders",
	}
	for _, sql := range tests {
		_, err := ValidateSQL(sql, "t-001")
		if err == nil {
			t.Errorf("expected rejection for: %s", sql)
		}
	}
}

func TestValidateSQL_RejectsMultipleStatements(t *testing.T) {
	_, err := ValidateSQL("SELECT 1; SELECT 2", "t-001")
	if err == nil {
		t.Fatal("expected rejection for multiple statements")
	}
}

func TestValidateSQL_InjectsTenantFilter(t *testing.T) {
	sql, err := ValidateSQL("SELECT * FROM orders", "t-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "tenant_id = 't-42'") {
		t.Errorf("expected tenant filter, got: %s", sql)
	}
}

func TestValidateSQL_KeepsExistingTenantFilter(t *testing.T) {
	sql, err := ValidateSQL("SELECT * FROM orders WHERE tenant_id = 't-99'", "t-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should keep the existing tenant_id, not inject a second one.
	count := strings.Count(sql, "tenant_id")
	if count != 1 {
		t.Errorf("expected 1 tenant_id reference, got %d: %s", count, sql)
	}
}

func TestValidateSQL_InjectsLimit(t *testing.T) {
	sql, err := ValidateSQL("SELECT * FROM orders", "t-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sql, "LIMIT 1000") {
		t.Errorf("expected LIMIT injection, got: %s", sql)
	}
}

func TestValidateSQL_AcceptsCTE(t *testing.T) {
	_, err := ValidateSQL("WITH recent AS (SELECT * FROM orders WHERE created_at > NOW()) SELECT * FROM recent", "t-001")
	if err != nil {
		t.Fatalf("expected CTE SELECT to be accepted, got: %v", err)
	}
}

func TestValidateSQL_RejectsDangerousFunctions(t *testing.T) {
	tests := []string{
		"SELECT pg_sleep(10)",
		"SELECT pg_read_file('/etc/passwd')",
		"SELECT current_setting('password')",
	}
	for _, sql := range tests {
		_, err := ValidateSQL(sql, "t-001")
		if err == nil {
			t.Errorf("expected rejection for dangerous function: %s", sql)
		}
	}
}

func TestValidateSQL_HandlesQuotedSemicolons(t *testing.T) {
	sql, err := ValidateSQL("SELECT * FROM orders WHERE name = 'test;value'", "t-001")
	if err != nil {
		t.Fatalf("semicolons in string literals should be allowed, got: %v", err)
	}
	if !strings.Contains(sql, "tenant_id = 't-001'") {
		t.Errorf("expected tenant filter: %s", sql)
	}
}

func TestHasMultipleStatements_FalsePositiveInString(t *testing.T) {
	if hasMultipleStatements("SELECT 'hello;world' AS greeting") {
		t.Error("semicolon in string literal should not count as statement separator")
	}
}

func TestHasMultipleStatements_TruePositive(t *testing.T) {
	if !hasMultipleStatements("SELECT 1; DROP TABLE users") {
		t.Error("should detect multiple statements")
	}
}
