package models

import "time"

// TableMeta describes a database table for the data dictionary.
type TableMeta struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Columns     []ColumnMeta `json:"columns"`
	PrimaryKey  string       `json:"primary_key"`
}

// ColumnMeta describes a single column.
type ColumnMeta struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nullable    bool   `json:"nullable"`
	Description string `json:"description"`
	IsFK        bool   `json:"is_fk"`
	FKRef       string `json:"fk_ref,omitempty"` // e.g. "orders.user_id -> users.id"
}

// Relationship describes a foreign key relationship between tables.
type Relationship struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
}

// SemanticMapping maps a business concept to its SQL representation.
type SemanticMapping struct {
	BusinessTerm string `json:"business_term"`
	SQLFragment  string `json:"sql_fragment"`
	Description  string `json:"description"`
	Table        string `json:"table,omitempty"`
}

// QueryRequest is the input for the NL2SQL API.
type QueryRequest struct {
	Question    string `json:"question" binding:"required"`
	TenantID    string `json:"tenant_id" binding:"required"`
	UserID      string `json:"user_id"`
	MaxRows     int    `json:"max_rows"`
	UseCache    bool   `json:"use_cache"`
}

// QueryResponse is the output of the NL2SQL API.
type QueryResponse struct {
	SQL         string                   `json:"sql"`
	Explanation string                   `json:"explanation"`
	Columns     []string                 `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"row_count"`
	FromCache   bool                     `json:"from_cache"`
	Duration    time.Duration            `json:"duration_ms"`
}

// SchemaResponse returns the data dictionary to the client.
type SchemaResponse struct {
	Tables        []TableMeta     `json:"tables"`
	Relationships []Relationship  `json:"relationships"`
	Semantics     []SemanticMapping `json:"semantics"`
}
