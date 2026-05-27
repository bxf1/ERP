package builder

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolExecutor defines the interface for executing MCP tool calls against the ERP backend.
type ToolExecutor interface {
	Execute(ctx context.Context, tool MCPToolCall) (json.RawMessage, error)
	ListTools() []MCPTool
}

// DefaultTools returns the standard MCP tool definitions for the Builder Agent.
func DefaultTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "list_models",
			Description: "列出系统中所有已定义的元数据模型。返回模型名称、显示名和字段数量。",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_model",
			Description: "获取指定模型的完整元数据定义，包括所有字段和索引。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model_name": map[string]string{
						"type":        "string",
						"description": "要查询的模型名称",
					},
				},
				"required": []string{"model_name"},
			},
		},
		{
			Name:        "create_model",
			Description: "创建新的元数据模型。传入完整的模型配置 JSON。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"config": map[string]string{
						"type":        "object",
						"description": "模型配置 JSON，包含 name, display_name, fields 等字段",
					},
				},
				"required": []string{"config"},
			},
		},
		{
			Name:        "update_model",
			Description: "更新已有模型的元数据定义。可以添加、修改或删除字段。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model_name": map[string]string{
						"type":        "string",
						"description": "要更新的模型名称",
					},
					"config": map[string]string{
						"type":        "object",
						"description": "更新的模型配置 JSON",
					},
				},
				"required": []string{"model_name", "config"},
			},
		},
	}
}

// MockToolExecutor is a test-friendly executor that stores models in memory.
// In production, this is replaced by an executor that calls the actual ERP API.
type MockToolExecutor struct {
	models map[string]ModelConfig
}

// NewMockToolExecutor creates a MockToolExecutor.
func NewMockToolExecutor() *MockToolExecutor {
	return &MockToolExecutor{
		models: make(map[string]ModelConfig),
	}
}

// Execute runs a tool call against the in-memory model store.
func (e *MockToolExecutor) Execute(ctx context.Context, call MCPToolCall) (json.RawMessage, error) {
	switch call.ToolName {
	case "list_models":
		return e.listModels()
	case "get_model":
		name, _ := call.Args["model_name"].(string)
		return e.getModel(name)
	case "create_model":
		return e.createModel(call.Args)
	case "update_model":
		return e.updateModel(call.Args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", call.ToolName)
	}
}

// ListTools returns the tool definitions.
func (e *MockToolExecutor) ListTools() []MCPTool {
	return DefaultTools()
}

func (e *MockToolExecutor) listModels() (json.RawMessage, error) {
	summaries := make([]ModelSummary, 0, len(e.models))
	for _, m := range e.models {
		summaries = append(summaries, ModelSummary{
			ID:          m.Name,
			Name:        m.Name,
			DisplayName: m.DisplayName,
			FieldCount:  len(m.Fields),
		})
	}
	return json.Marshal(summaries)
}

func (e *MockToolExecutor) getModel(name string) (json.RawMessage, error) {
	m, ok := e.models[name]
	if !ok {
		return nil, fmt.Errorf("model %s not found", name)
	}
	return json.Marshal(m)
}

func (e *MockToolExecutor) createModel(args map[string]interface{}) (json.RawMessage, error) {
	cfgRaw, ok := args["config"]
	if !ok {
		return nil, fmt.Errorf("missing config argument")
	}

	var cfg ModelConfig
	// Accept both already-parsed objects and raw JSON.
	switch v := cfgRaw.(type) {
	case ModelConfig:
		cfg = v
	case map[string]interface{}:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	default:
		return nil, fmt.Errorf("config must be a ModelConfig object, got %T", cfgRaw)
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("model name is required")
	}
	if _, exists := e.models[cfg.Name]; exists {
		return nil, fmt.Errorf("model %s already exists", cfg.Name)
	}

	e.models[cfg.Name] = cfg
	return json.Marshal(map[string]string{"status": "created", "model": cfg.Name})
}

func (e *MockToolExecutor) updateModel(args map[string]interface{}) (json.RawMessage, error) {
	name, _ := args["model_name"].(string)
	if name == "" {
		return nil, fmt.Errorf("model_name is required")
	}

	existing, ok := e.models[name]
	if !ok {
		return nil, fmt.Errorf("model %s not found", name)
	}

	cfgRaw, ok := args["config"]
	if !ok {
		return nil, fmt.Errorf("missing config argument")
	}

	var cfg ModelConfig
	switch v := cfgRaw.(type) {
	case ModelConfig:
		cfg = v
	case map[string]interface{}:
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &cfg)
	}

	// Merge: keep existing fields, add/update new ones.
	existingFields := make(map[string]FieldConfig)
	for _, f := range existing.Fields {
		existingFields[f.Name] = f
	}
	for _, f := range cfg.Fields {
		existingFields[f.Name] = f
	}
	existing.Fields = make([]FieldConfig, 0, len(existingFields))
	for _, f := range existingFields {
		existing.Fields = append(existing.Fields, f)
	}

	existing.UpdatedAt = "" // would be set by the engine
	e.models[name] = existing
	return json.Marshal(map[string]string{"status": "updated", "model": name})
}
