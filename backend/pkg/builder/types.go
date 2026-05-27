// Package builder implements the AI-powered Builder Agent that converts
// natural language requirements into ERP metadata configurations.
package builder

import "time"

// ConversationState represents the current phase of the builder conversation.
type ConversationState string

const (
	StateRequirements ConversationState = "requirements"
	StateSolution     ConversationState = "solution"
	StateConfirmation ConversationState = "confirmation"
	StateCreation     ConversationState = "creation"
	StateDone         ConversationState = "done"
)

// FieldType enumerates supported ERP field types.
type FieldType string

const (
	FieldString   FieldType = "string"
	FieldText     FieldType = "text"
	FieldNumber   FieldType = "number"
	FieldDecimal  FieldType = "decimal"
	FieldBoolean  FieldType = "boolean"
	FieldDate     FieldType = "date"
	FieldDateTime FieldType = "datetime"
	FieldEnum     FieldType = "enum"
	FieldRelation FieldType = "relation"
	FieldFile     FieldType = "file"
	FieldJSON     FieldType = "json"
)

// ModelConfig is the metadata configuration for a single ERP model.
type ModelConfig struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description,omitempty"`
	TableName   string        `json:"table_name,omitempty"`
	Fields      []FieldConfig `json:"fields"`
	Indexes     []IndexConfig `json:"indexes,omitempty"`
}

// FieldConfig defines a single field within a model.
type FieldConfig struct {
	Name         string      `json:"name"`
	DisplayName  string      `json:"display_name"`
	Type         FieldType   `json:"type"`
	Required     bool        `json:"required"`
	Unique       bool        `json:"unique,omitempty"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	MaxLength    int         `json:"max_length,omitempty"`
	MinValue     *float64    `json:"min_value,omitempty"`
	MaxValue     *float64    `json:"max_value,omitempty"`
	EnumValues   []string    `json:"enum_values,omitempty"`
	RelationModel string     `json:"relation_model,omitempty"`
	RelationField string     `json:"relation_field,omitempty"`
	Description  string      `json:"description,omitempty"`
}

// IndexConfig defines a database index on a model.
type IndexConfig struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
	Unique bool     `json:"unique,omitempty"`
}

// BuildRequest is the input to a Builder Agent session.
type BuildRequest struct {
	SessionID     string `json:"session_id,omitempty"`
	UserMessage   string `json:"user_message"`
	PreferredLang string `json:"preferred_lang,omitempty"`
}

// BuildResponse is the output from a single Builder Agent turn.
type BuildResponse struct {
	SessionID        string            `json:"session_id"`
	State            ConversationState `json:"state"`
	Message          string            `json:"message"`
	ProposedConfigs  []ModelConfig     `json:"proposed_configs,omitempty"`
	ExistingModels   []ModelSummary    `json:"existing_models,omitempty"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
	CreatedModels    []ModelSummary    `json:"created_models,omitempty"`
}

// ModelSummary is a lightweight reference to an existing model.
type ModelSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	FieldCount  int    `json:"field_count"`
}

// ValidationError describes a problem with a generated configuration.
type ValidationError struct {
	ModelName string `json:"model_name,omitempty"`
	FieldName string `json:"field_name,omitempty"`
	Message   string `json:"message"`
	Severity  string `json:"severity"` // error, warning
}

// Conversation stores the full session state for a builder conversation.
type Conversation struct {
	SessionID        string            `json:"session_id"`
	State            ConversationState `json:"state"`
	Messages         []Message         `json:"messages"`
	ProposedConfigs  []ModelConfig     `json:"proposed_configs,omitempty"`
	CreatedModels    []string          `json:"created_models,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// Message is a single turn in the builder conversation.
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// MCPTool defines a tool callable by the Builder Agent.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// MCPToolCall represents a single tool invocation.
type MCPToolCall struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

// BuildIntent is the structured output parsed from a user's natural language input.
// It represents what the user wants to build.
type BuildIntent struct {
	Intent      string        `json:"intent"`      // create_model, update_model, query_models
	ModelName   string        `json:"model_name"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description"`
	Fields      []FieldConfig `json:"fields"`
	ReuseModels []string      `json:"reuse_models"` // models to reference via relations
}
