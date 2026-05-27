// Package semantic defines business metrics mapped to SQL fragments,
// providing the context layer that bridges natural-language questions
// to database queries.
package semantic

import "time"

// MetricKind classifies what shape of data a metric produces.
type MetricKind string

const (
	MetricSum   MetricKind = "sum"
	MetricCount MetricKind = "count"
	MetricAvg   MetricKind = "avg"
	MetricMin   MetricKind = "min"
	MetricMax   MetricKind = "max"
)

// Metric defines a single business KPI.
type Metric struct {
	Name        string            `json:"name" yaml:"name"`
	Label       string            `json:"label" yaml:"label"`
	Description string            `json:"description" yaml:"description"`
	Kind        MetricKind        `json:"kind" yaml:"kind"`
	Table       string            `json:"table" yaml:"table"`
	Column      string            `json:"column" yaml:"column"`
	Filters     []FilterClause    `json:"filters,omitempty" yaml:"filters,omitempty"`
	Joins       []JoinClause      `json:"joins,omitempty" yaml:"joins,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty" yaml:"group_by,omitempty"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// FilterClause is a single WHERE condition applied to a metric.
type FilterClause struct {
	Column   string `json:"column" yaml:"column"`
	Operator string `json:"operator" yaml:"operator"` // =, !=, >, <, >=, <=, IN, LIKE, IS NULL, IS NOT NULL
	Value    string `json:"value" yaml:"value"`
}

// JoinClause describes how to join another table when computing a metric.
type JoinClause struct {
	Table      string `json:"table" yaml:"table"`
	Alias      string `json:"alias,omitempty" yaml:"alias,omitempty"`
	On         string `json:"on" yaml:"on"` // e.g. "sales_orders.customer_id = customers.id"
	JoinType   string `json:"join_type" yaml:"join_type"` // INNER, LEFT, RIGHT
}

// Dimension describes a field that can be used to slice a metric.
type Dimension struct {
	Name        string `json:"name" yaml:"name"`
	Label       string `json:"label" yaml:"label"`
	Table       string `json:"table" yaml:"table"`
	Column      string `json:"column" yaml:"column"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// SemanticModel groups related metrics and dimensions that belong to
// a single business domain (e.g. "sales", "inventory").
type SemanticModel struct {
	Name        string      `json:"name" yaml:"name"`
	Label       string      `json:"label" yaml:"label"`
	Description string      `json:"description" yaml:"description"`
	Metrics     []Metric    `json:"metrics" yaml:"metrics"`
	Dimensions  []Dimension `json:"dimensions" yaml:"dimensions"`
}

// Config is the top-level configuration file structure.
type Config struct {
	Version string          `json:"version" yaml:"version"`
	Models  []SemanticModel `json:"models" yaml:"models"`
}

// QueryRequest is a user-submitted natural-language query request.
type QueryRequest struct {
	Question string   `json:"question"`
	Metrics  []string `json:"metrics,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// QueryResult holds the SQL generated from a semantic-layer resolution.
type QueryResult struct {
	SQL        string              `json:"sql"`
	Metrics    []string            `json:"metrics_used"`
	GeneratedAt time.Time          `json:"generated_at"`
	Params     map[string]string   `json:"params,omitempty"`
}

// LLMContext is the prompt payload injected into an NL2SQL LLM call.
type LLMContext struct {
	Tables      []TableRef    `json:"tables"`
	Metrics     []MetricRef   `json:"metrics"`
	Dimensions  []Dimension   `json:"dimensions"`
	Joins       []JoinHint    `json:"joins"`
	GeneratedAt time.Time     `json:"generated_at"`
}

// TableRef is a compact table reference for LLM context.
type TableRef struct {
	Name    string       `json:"name"`
	Columns []ColumnRef  `json:"columns"`
}

// ColumnRef is a compact column reference for LLM context.
type ColumnRef struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment,omitempty"`
}

// MetricRef is a compact metric reference for LLM context.
type MetricRef struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SQL         string `json:"sql"` // Rendered SQL fragment
}

// JoinHint describes available join paths between tables.
type JoinHint struct {
	FromTable string `json:"from_table"`
	ToTable   string `json:"to_table"`
	Condition string `json:"condition"`
}
