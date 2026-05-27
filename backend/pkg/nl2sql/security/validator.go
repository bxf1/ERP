package security

import (
	"fmt"
	"regexp"
	"strings"
)

// AllowedMaxRows is the maximum rows a SELECT can return.
const AllowedMaxRows = 1000

// DangerousPatterns that indicate SQL injection or unauthorized operations.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i);\s*(INSERT|UPDATE|DELETE|DROP|TRUNCATE|ALTER|CREATE|GRANT|REVOKE)\s`),
	regexp.MustCompile(`(?i)COPY\s+`),
	regexp.MustCompile(`(?i)EXECUTE\s+`),
	regexp.MustCompile(`(?i)pg_sleep\s*\(`),
	regexp.MustCompile(`(?i)pg_read_file\s*\(`),
	regexp.MustCompile(`(?i)pg_write_file\s*\(`),
	regexp.MustCompile(`(?i)LO_IMPORT\s*\(`),
	regexp.MustCompile(`(?i)LO_EXPORT\s*\(`),
	regexp.MustCompile(`(?i)current_setting\s*\(`),
	regexp.MustCompile(`(?i)set_config\s*\(`),
}

// reSelect checks the statement begins with SELECT (optionally after WITH/CTE).
var reSelect = regexp.MustCompile(`(?i)^\s*(WITH\s+.+\s+)?SELECT\s`)

// ValidateSQL checks that a SQL string is a safe SELECT query.
// Returns the validated (and possibly rewritten) SQL, or an error.
func ValidateSQL(sql string, tenantID string) (string, error) {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimRight(sql, ";")
	sql = strings.TrimSpace(sql)

	if sql == "" {
		return "", fmt.Errorf("empty SQL statement")
	}

	// Must start with SELECT (or WITH ... SELECT for CTEs).
	if !reSelect.MatchString(sql) {
		return "", fmt.Errorf("only SELECT statements are allowed")
	}

	// Block multiple statements (semicolons not inside string literals).
	if hasMultipleStatements(sql) {
		return "", fmt.Errorf("multiple SQL statements are not allowed")
	}

	// Check for injected dangerous patterns (defense in depth).
	for _, pat := range dangerousPatterns {
		if pat.MatchString(sql) {
			return "", fmt.Errorf("SQL contains dangerous pattern: %s", pat.String())
		}
	}

	// Check for no LIMIT clause and inject one if needed.
	if !hasLimitClause(sql) {
		sql = injectLimit(sql, AllowedMaxRows)
	}

	// Enforce tenant filtering.
	var err error
	sql, err = enforceTenantFilter(sql, tenantID)
	if err != nil {
		return "", fmt.Errorf("tenant filter enforcement: %w", err)
	}

	return sql + ";", nil
}

func hasLimitClause(sql string) bool {
	upper := strings.ToUpper(sql)
	return regexp.MustCompile(`\bLIMIT\s+\d+`).MatchString(upper)
}

func injectLimit(sql string, limit int) string {
	return fmt.Sprintf("%s LIMIT %d", sql, limit)
}

// hasMultipleStatements checks for semicolons that would indicate statement chaining.
// It strips single-quoted strings before checking to avoid false positives.
func hasMultipleStatements(sql string) bool {
	clean := stripQuotedStrings(sql)
	return strings.Contains(clean, ";")
}

func stripQuotedStrings(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	inSingle := false
	inDouble := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			result.WriteByte('\'')
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			result.WriteByte('"')
			continue
		}
		if inSingle || inDouble {
			result.WriteByte('X') // Mask content inside quotes.
		} else {
			result.WriteByte(ch)
		}
	}
	return result.String()
}
