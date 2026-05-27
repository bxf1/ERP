package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// ReWhere matches an existing WHERE clause.
	reWhere = regexp.MustCompile(`(?i)\bWHERE\b`)
	// ReGroupBy matches GROUP BY to know where to insert WHERE before GROUP BY.
	reGroupBy = regexp.MustCompile(`(?i)\bGROUP\s+BY\b`)
	// ReOrderBy matches ORDER BY.
	reOrderBy = regexp.MustCompile(`(?i)\bORDER\s+BY\b`)
	// ReLimit matches LIMIT clause.
	reLimit = regexp.MustCompile(`(?i)\bLIMIT\b`)
	// ReTenantID checks if tenant_id is already in the query.
	reTenantFilter = regexp.MustCompile(`(?i)tenant_id\s*=\s*'[^']*'`)
	// ReTableAliases finds table aliases to help inject qualified tenant_id.
	reTableAlias = regexp.MustCompile(`(?i)(?:FROM|JOIN)\s+\w+\s+(?:AS\s+)?(\w+)`)
)

// enforceTenantFilter ensures the query includes a tenant_id filter.
// If the query already filters by tenant_id, it's returned unchanged.
// Otherwise, the tenant condition is injected into the WHERE clause.
func enforceTenantFilter(sql string, tenantID string) (string, error) {
	if reTenantFilter.MatchString(sql) {
		return sql, nil
	}

	tenantCondition := fmt.Sprintf("tenant_id = '%s'", tenantID)

	if reWhere.MatchString(sql) {
		// Has WHERE clause: inject tenant_id condition into it.
		sql = reWhere.ReplaceAllStringFunc(sql, func(match string) string {
			return match + " " + tenantCondition + " AND "
		})
	} else {
		// No WHERE clause: add one before GROUP BY / ORDER BY / LIMIT.
		sql = injectWhereClause(sql, tenantCondition)
	}

	return sql, nil
}

// injectWhereClause adds a WHERE clause at the right position in a query without one.
func injectWhereClause(sql, condition string) string {
	// Find the last position before GROUP BY, ORDER BY, or LIMIT.
	insertPos := len(sql)

	for _, re := range []*regexp.Regexp{reGroupBy, reOrderBy, reLimit} {
		loc := re.FindStringIndex(sql)
		if loc != nil && loc[0] < insertPos {
			insertPos = loc[0]
		}
	}

	prefix := strings.TrimSpace(sql[:insertPos])
	suffix := ""
	if insertPos < len(sql) {
		suffix = " " + strings.TrimSpace(sql[insertPos:])
	}

	return fmt.Sprintf("%s WHERE %s%s", prefix, condition, suffix)
}
