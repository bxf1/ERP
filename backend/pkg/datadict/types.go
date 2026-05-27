// Package datadict provides automated extraction and serving of database schema metadata.
// It scans PostgreSQL information_schema to build a structured data dictionary
// that the AI layer uses for NL2SQL, form generation, and schema exploration.
package datadict

import "time"

// DataDict is the top-level snapshot of the entire database schema.
type DataDict struct {
	ExtractedAt time.Time       `json:"extracted_at"`
	Tables      []TableInfo     `json:"tables"`
	Relations   []RelationInfo  `json:"relations"`
}

// TableInfo describes a single database table or view.
type TableInfo struct {
	Schema      string       `json:"schema"`
	Name        string       `json:"name"`
	Comment     string       `json:"comment"`
	Type        string       `json:"type"` // TABLE, VIEW, MATERIALIZED VIEW
	Columns     []ColumnInfo `json:"columns"`
	PrimaryKeys []string     `json:"primary_keys"`
	Indexes     []IndexInfo  `json:"indexes,omitempty"`
}

// ColumnInfo describes a single column in a table.
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	Comment      string `json:"comment,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	MaxLength    *int   `json:"max_length,omitempty"`
	NumericScale *int   `json:"numeric_scale,omitempty"`
}

// IndexInfo describes a database index.
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// RelationInfo describes a foreign key relationship between two tables.
type RelationInfo struct {
	Name           string `json:"name"`
	SourceSchema   string `json:"source_schema"`
	SourceTable    string `json:"source_table"`
	SourceColumn   string `json:"source_column"`
	TargetSchema   string `json:"target_schema"`
	TargetTable    string `json:"target_table"`
	TargetColumn   string `json:"target_column"`
}

// ColumnDiff represents a change detected in a column between two schema snapshots.
type ColumnDiff struct {
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	ChangeType string `json:"change_type"` // added, removed, modified
	OldValue   string `json:"old_value,omitempty"`
	NewValue   string `json:"new_value,omitempty"`
}

// SchemaDiff is the result of comparing two DataDict snapshots.
type SchemaDiff struct {
	AddedTables    []string     `json:"added_tables"`
	RemovedTables  []string     `json:"removed_tables"`
	ModifiedTables []string     `json:"modified_tables"`
	ColumnChanges  []ColumnDiff `json:"column_changes"`
}
