// Package model provides data model definitions and CRUD operations
// for the ERP metadata engine.
package model

import (
	"time"

	"github.com/google/uuid"
)

// DataModel represents a user-defined business data model (metadata-driven).
type DataModel struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	TableName   string    `json:"table_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModelField represents a field within a data model.
type ModelField struct {
	ID          string    `json:"id"`
	ModelID     string    `json:"model_id"`
	Name        string    `json:"name"`
	Label       string    `json:"label"`
	Type        string    `json:"type"`
	Required    bool      `json:"required"`
	Unique      bool      `json:"unique"`
	Default     *string   `json:"default,omitempty"`
	MaxLength   int       `json:"max_length"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewDataModel creates a DataModel with auto-generated ID.
func NewDataModel(tenantID, name, label, description, tableName string) *DataModel {
	now := time.Now().UTC()
	return &DataModel{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Label:       label,
		Description: description,
		TableName:   tableName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewModelField creates a ModelField with auto-generated ID.
func NewModelField(modelID string, config FieldConfig) *ModelField {
	var def *string
	if config.Default != nil {
		s := toString(config.Default)
		def = &s
	}
	return &ModelField{
		ID:          uuid.New().String(),
		ModelID:     modelID,
		Name:        config.Name,
		Label:       config.Label,
		Type:        config.Type,
		Required:    config.Required,
		Unique:      config.Unique,
		Default:     def,
		MaxLength:   config.MaxLength,
		Description: config.Description,
		CreatedAt:   time.Now().UTC(),
	}
}

// FieldConfig is the input config for creating a field.
type FieldConfig struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Unique      bool   `json:"unique,omitempty"`
	Default     any    `json:"default,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	Description string `json:"description,omitempty"`
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return ""
	}
}
