package model

import (
	"encoding/json"
	"time"
)

// Tenant is the public.tenants row.
type Tenant struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Slug       string          `json:"slug"`
	SchemaName string          `json:"schema_name"`
	Status     string          `json:"status"`
	Config     json.RawMessage `json:"config"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Model represents a metadata model definition (per-tenant).
type Model struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	TableName   string          `json:"table_name"`
	Label       string          `json:"label"`
	Description *string         `json:"description,omitempty"`
	IsSystem    bool            `json:"is_system"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Field represents a field definition within a model (per-tenant).
type Field struct {
	ID              string          `json:"id"`
	ModelID         string          `json:"model_id"`
	Name            string          `json:"name"`
	ColumnName      string          `json:"column_name"`
	Label           string          `json:"label"`
	FieldType       string          `json:"field_type"`
	IsRequired      bool            `json:"is_required"`
	IsUnique        bool            `json:"is_unique"`
	DefaultValue    *string         `json:"default_value,omitempty"`
	ValidationRules json.RawMessage `json:"validation_rules"`
	UIConfig        json.RawMessage `json:"ui_config"`
	OrderIndex      int             `json:"order_index"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// Menu represents a navigation menu item (per-tenant).
type Menu struct {
	ID            string    `json:"id"`
	ParentID      *string   `json:"parent_id,omitempty"`
	Name          string    `json:"name"`
	Label         string    `json:"label"`
	Icon          *string   `json:"icon,omitempty"`
	Path          string    `json:"path"`
	Component     *string   `json:"component,omitempty"`
	PermissionKey *string   `json:"permission_key,omitempty"`
	OrderIndex    int       `json:"order_index"`
	IsVisible     bool      `json:"is_visible"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Role represents an RBAC role (per-tenant).
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a single permission entry (per-tenant).
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description *string   `json:"description,omitempty"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModelRelation defines a relationship between two models.
type ModelRelation struct {
	ID             string    `json:"id"`
	SourceModelID  string    `json:"source_model_id"`
	TargetModelID  string    `json:"target_model_id"`
	RelationType   string    `json:"relation_type"`
	SourceField    string    `json:"source_field"`
	TargetField    *string   `json:"target_field,omitempty"`
	JunctionTable  *string   `json:"junction_table,omitempty"`
	Label          *string   `json:"label,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AuditLog records an audit trail entry.
type AuditLog struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Action     string          `json:"action"`
	Resource   string          `json:"resource"`
	ResourceID *string         `json:"resource_id,omitempty"`
	OldValues  json.RawMessage `json:"old_values,omitempty"`
	NewValues  json.RawMessage `json:"new_values,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	UserAgent  *string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
