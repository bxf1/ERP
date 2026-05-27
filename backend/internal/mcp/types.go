// Package mcp defines the MCP (Model Context Protocol) interface types and tool registry
// for AI agents to interact with the ERP system via Function Calling.
package mcp

import "encoding/json"

// JSONRPCRequest represents a standard JSON-RPC 2.0 request from an MCP client.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a standard JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error object.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolDefinition describes an MCP tool exposed for LLM Function Calling.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListModelsParams is the input for list_models.
type ListModelsParams struct {
	TenantID string `json:"tenant_id"`
}

// GetModelParams is the input for get_model.
type GetModelParams struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
}

// CreateModelParams is the input for create_model.
type CreateModelParams struct {
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Label       string          `json:"label"`
	Description string          `json:"description,omitempty"`
	TableName   string          `json:"table_name"`
	Fields      []FieldConfig   `json:"fields,omitempty"`
	Confirmed   bool            `json:"confirmed"`
}

// AddFieldParams is the input for add_field.
type AddFieldParams struct {
	TenantID  string      `json:"tenant_id"`
	ModelName string      `json:"model"`
	Field     FieldConfig `json:"field_config"`
	Confirmed bool        `json:"confirmed"`
}

// FieldConfig defines a field on a data model.
type FieldConfig struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"` // string, integer, float, boolean, date, datetime, text, jsonb
	Required    bool   `json:"required"`
	Unique      bool   `json:"unique,omitempty"`
	Default     any    `json:"default,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	Description string `json:"description,omitempty"`
}

// QueryDataParams is the input for query_data.
type QueryDataParams struct {
	TenantID string `json:"tenant_id"`
	SQL      string `json:"sql"`
}

// SemanticLayerParams is the input for get_semantic_layer.
type SemanticLayerParams struct {
	TenantID string `json:"tenant_id"`
}

// ModelInfo is the response type for list_models and get_model.
type ModelInfo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Label       string        `json:"label"`
	Description string        `json:"description"`
	TableName   string        `json:"table_name"`
	Fields      []FieldInfo   `json:"fields"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

// FieldInfo describes a field in a model response.
type FieldInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Unique      bool   `json:"unique"`
	Default     any    `json:"default,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	Description string `json:"description,omitempty"`
}

// SemanticLayer represents business-to-data mapping.
type SemanticLayer struct {
	Metrics   []MetricDef  `json:"metrics"`
	Dimensions []DimensionDef `json:"dimensions"`
	Models    []string     `json:"models"`
}

// MetricDef defines a business metric.
type MetricDef struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
	Unit        string `json:"unit,omitempty"`
}

// DimensionDef defines a business dimension.
type DimensionDef struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	FieldPath   string `json:"field_path"`
}

// ToolCallResult holds the result of a tool execution.
type ToolCallResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Warning string      `json:"warning,omitempty"`
}

// ConfirmationRequest is returned when a write operation needs user confirmation.
type ConfirmationRequest struct {
	NeedsConfirmation bool   `json:"needs_confirmation"`
	Message           string `json:"message"`
	Operation         string `json:"operation"`
	Summary           any    `json:"summary"`
}
