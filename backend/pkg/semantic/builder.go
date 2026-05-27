package semantic

import (
	"fmt"
	"strings"
)

// SQLBuilder renders a Metric into a runnable SQL SELECT fragment.
type SQLBuilder struct {
	dialect string // postgres (default)
}

// NewSQLBuilder creates a SQLBuilder for the given SQL dialect.
func NewSQLBuilder(dialect string) *SQLBuilder {
	return &SQLBuilder{dialect: dialect}
}

// BuildMetricSQL renders a single metric into its SQL expression.
// Example output: SUM(sales_orders.amount)
func (b *SQLBuilder) BuildMetricSQL(m Metric) string {
	col := b.qualify(m.Table, m.Column)
	switch m.Kind {
	case MetricCount:
		return fmt.Sprintf("COUNT(%s)", col)
	case MetricAvg:
		return fmt.Sprintf("AVG(%s)", col)
	case MetricMin:
		return fmt.Sprintf("MIN(%s)", col)
	case MetricMax:
		return fmt.Sprintf("MAX(%s)", col)
	default:
		return fmt.Sprintf("SUM(%s)", col)
	}
}

// BuildFromClause constructs the FROM + JOIN clauses for a metric.
func (b *SQLBuilder) BuildFromClause(m Metric) string {
	from := b.quoteIdent(m.Table)
	for _, j := range m.Joins {
		alias := j.Alias
		if alias == "" {
			alias = j.Table
		}
		from += fmt.Sprintf(" %s JOIN %s AS %s ON %s",
			j.JoinType, b.quoteIdent(j.Table), b.quoteIdent(alias), j.On)
	}
	return "FROM " + from
}

// BuildWhereClause builds the WHERE clause from metric filters.
func (b *SQLBuilder) BuildWhereClause(m Metric) string {
	if len(m.Filters) == 0 {
		return ""
	}
	var conds []string
	for _, f := range m.Filters {
		col := b.qualify(m.Table, f.Column)
		switch strings.ToUpper(f.Operator) {
		case "IN":
			conds = append(conds, fmt.Sprintf("%s IN (%s)", col, f.Value))
		case "IS NULL":
			conds = append(conds, fmt.Sprintf("%s IS NULL", col))
		case "IS NOT NULL":
			conds = append(conds, fmt.Sprintf("%s IS NOT NULL", col))
		case "LIKE":
			conds = append(conds, fmt.Sprintf("%s LIKE '%s'", col, escapeSQL(f.Value)))
		default:
			conds = append(conds, fmt.Sprintf("%s %s '%s'", col, f.Operator, escapeSQL(f.Value)))
		}
	}
	return "WHERE " + strings.Join(conds, " AND ")
}

// BuildGroupBy builds the GROUP BY clause.
func (b *SQLBuilder) BuildGroupBy(m Metric) string {
	if len(m.GroupBy) == 0 {
		return ""
	}
	var cols []string
	for _, g := range m.GroupBy {
		cols = append(cols, b.qualify(m.Table, g))
	}
	return "GROUP BY " + strings.Join(cols, ", ")
}

// BuildSelect assembles the full SELECT statement for a single metric.
func (b *SQLBuilder) BuildSelect(m Metric) string {
	parts := []string{
		"SELECT " + b.BuildMetricSQL(m),
		b.BuildFromClause(m),
	}
	if where := b.BuildWhereClause(m); where != "" {
		parts = append(parts, where)
	}
	if groupBy := b.BuildGroupBy(m); groupBy != "" {
		parts = append(parts, groupBy)
	}
	return strings.Join(parts, "\n")
}

// BuildMultiMetric assembles a SELECT that computes several metrics together,
// joining their tables as needed.
func (b *SQLBuilder) BuildMultiMetric(metrics []Metric) string {
	if len(metrics) == 0 {
		return ""
	}
	if len(metrics) == 1 {
		return b.BuildSelect(metrics[0])
	}

	var exprs []string
	var fromParts []string
	var whereParts []string
	seenTables := make(map[string]bool)

	for _, m := range metrics {
		qual := b.qualify(m.Table, m.Column)
		alias := m.Name
		switch m.Kind {
		case MetricCount:
			exprs = append(exprs, fmt.Sprintf("COUNT(%s) AS %s", qual, b.quoteIdent(alias)))
		case MetricAvg:
			exprs = append(exprs, fmt.Sprintf("AVG(%s) AS %s", qual, b.quoteIdent(alias)))
		case MetricMin:
			exprs = append(exprs, fmt.Sprintf("MIN(%s) AS %s", qual, b.quoteIdent(alias)))
		case MetricMax:
			exprs = append(exprs, fmt.Sprintf("MAX(%s) AS %s", qual, b.quoteIdent(alias)))
		default:
			exprs = append(exprs, fmt.Sprintf("SUM(%s) AS %s", qual, b.quoteIdent(alias)))
		}

		if !seenTables[m.Table] {
			seenTables[m.Table] = true
			fromParts = append(fromParts, b.quoteIdent(m.Table))
		}
		for _, j := range m.Joins {
			alias := j.Alias
			if alias == "" {
				alias = j.Table
			}
			if !seenTables[alias] {
				seenTables[alias] = true
				fromParts = append(fromParts,
					fmt.Sprintf("%s JOIN %s AS %s ON %s", j.JoinType, b.quoteIdent(j.Table), b.quoteIdent(alias), j.On))
			}
		}
		for _, f := range m.Filters {
			col := b.qualify(m.Table, f.Column)
			whereParts = append(whereParts, fmt.Sprintf("%s %s '%s'", col, f.Operator, escapeSQL(f.Value)))
		}
	}

	result := "SELECT " + strings.Join(exprs, ",\n       ")
	result += "\nFROM " + strings.Join(fromParts, ", ")
	if len(whereParts) > 0 {
		result += "\nWHERE " + strings.Join(whereParts, " AND ")
	}
	return result
}

func (b *SQLBuilder) qualify(table, column string) string {
	return b.quoteIdent(table) + "." + b.quoteIdent(column)
}

func (b *SQLBuilder) quoteIdent(s string) string {
	if b.dialect == "postgres" && !strings.HasPrefix(s, `"`) && !strings.HasSuffix(s, `"`) {
		return `"` + s + `"`
	}
	return s
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
