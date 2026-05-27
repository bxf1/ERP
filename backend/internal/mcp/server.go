package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bxf1/ERP/backend/internal/gateway"
	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/internal/semantic"
	"github.com/gin-gonic/gin"
)

// Server is the MCP HTTP server that exposes tools for LLM Function Calling.
type Server struct {
	router   *gin.Engine
	modelSvc *model.Service
	gateway  *gateway.SecurityGateway
	semantic *semantic.Layer
}

// NewServer creates and configures the MCP server.
func NewServer(modelSvc *model.Service, gw *gateway.SecurityGateway, sem *semantic.Layer) *Server {
	s := &Server{
		modelSvc: modelSvc,
		gateway:  gw,
		semantic: sem,
	}

	router := gin.Default()

	// MCP JSON-RPC endpoint
	router.POST("/mcp", s.handleJSONRPC)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	s.router = router
	return s
}

// Run starts the MCP server on the given address.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// ToolDefinitions returns all registered MCP tools.
func ToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "list_models",
			Description: "获取所有数据模型列表。返回当前租户下所有已定义的数据模型。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"}},"required":["tenant_id"]}`),
		},
		{
			Name:        "get_model",
			Description: "获取模型详情，包括字段定义、关系、校验规则。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"},"name":{"type":"string","description":"模型名称"}},"required":["tenant_id","name"]}`),
		},
		{
			Name:        "create_model",
			Description: "创建新数据模型。写操作，需用户二次确认。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"},"name":{"type":"string","description":"模型名称（英文标识）"},"label":{"type":"string","description":"模型显示名"},"description":{"type":"string","description":"模型描述"},"table_name":{"type":"string","description":"数据库表名"},"fields":{"type":"array","description":"初始字段定义"},"confirmed":{"type":"boolean","description":"用户确认标志，首次调用为false，确认后为true"}},"required":["tenant_id","name","label","table_name","confirmed"]}`),
		},
		{
			Name:        "add_field",
			Description: "为已有数据模型添加字段。写操作，需用户二次确认。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"},"model":{"type":"string","description":"目标模型名称"},"field_config":{"type":"object","description":"字段配置"},"confirmed":{"type":"boolean","description":"用户确认标志"}},"required":["tenant_id","model","field_config","confirmed"]}`),
		},
		{
			Name:        "query_data",
			Description: "执行只读SQL查询，自动注入租户过滤条件并进行权限检查。仅支持SELECT语句。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"},"sql":{"type":"string","description":"SQL SELECT查询语句"}},"required":["tenant_id","sql"]}`),
		},
		{
			Name:        "get_semantic_layer",
			Description: "获取语义层定义，包含业务指标和数据维度的映射关系。",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"tenant_id":{"type":"string","description":"租户ID"}},"required":["tenant_id"]}`),
		},
	}
}

// handleJSONRPC processes MCP JSON-RPC requests.
func (s *Server) handleJSONRPC(c *gin.Context) {
	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error: " + err.Error()},
		})
		return
	}

	log.Printf("[MCP] method=%s id=%v", req.Method, req.ID)

	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "tools/list":
		resp.Result = gin.H{"tools": ToolDefinitions()}

	case "tools/call":
		var callParams struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = &RPCError{Code: -32602, Message: "Invalid params: " + err.Error()}
		} else {
			result := s.dispatchTool(callParams.Name, callParams.Arguments)
			resp.Result = result
		}

	default:
		resp.Error = &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}
	}

	c.JSON(http.StatusOK, resp)
}

// dispatchTool routes a tool call to the correct handler.
func (s *Server) dispatchTool(name string, args json.RawMessage) ToolCallResult {
	userID := extractUserID(args)
	tenantID := extractTenantID(args)

	switch name {
	case "list_models":
		return s.handleListModels(args, tenantID, userID)
	case "get_model":
		return s.handleGetModel(args, tenantID, userID)
	case "create_model":
		return s.handleCreateModel(args, tenantID, userID)
	case "add_field":
		return s.handleAddField(args, tenantID, userID)
	case "query_data":
		return s.handleQueryData(args, tenantID, userID)
	case "get_semantic_layer":
		return s.handleGetSemanticLayer(args, tenantID, userID)
	default:
		return ToolCallResult{
			Success: false,
			Error:   fmt.Sprintf("unknown tool: %s", name),
		}
	}
}

func (s *Server) handleListModels(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermReadModels); err != nil {
		s.gateway.LogAction(tenantID, userID, "list_models", "models", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	var params ListModelsParams
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{Success: false, Error: "invalid params: " + err.Error()}
	}

	models := s.modelSvc.ListModels(params.TenantID)

	var result []ModelInfo
	for _, m := range models {
		fields := convertFieldsFromService(s.modelSvc.GetFieldsByModelID(m.ID))
		result = append(result, toModelInfo(m, fields))
	}

	s.gateway.LogAction(tenantID, userID, "list_models", "models",
		fmt.Sprintf("returned %d models", len(result)), "success")
	return ToolCallResult{Success: true, Data: result}
}

func (s *Server) handleGetModel(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermReadModels); err != nil {
		s.gateway.LogAction(tenantID, userID, "get_model", "models", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	var params GetModelParams
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{Success: false, Error: "invalid params: " + err.Error()}
	}

	m, fields, err := s.modelSvc.GetModel(params.TenantID, params.Name)
	if err != nil {
		s.gateway.LogAction(tenantID, userID, "get_model", params.Name, "", "not_found")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	info := toModelInfo(m, convertFieldsFromService(fields))
	s.gateway.LogAction(tenantID, userID, "get_model", params.Name, "", "success")
	return ToolCallResult{Success: true, Data: info}
}

func (s *Server) handleCreateModel(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermWriteModels); err != nil {
		s.gateway.LogAction(tenantID, userID, "create_model", "models", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	var params CreateModelParams
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{Success: false, Error: "invalid params: " + err.Error()}
	}

	// Require user confirmation for write operations.
	if !params.Confirmed {
		summary := map[string]interface{}{
			"name":        params.Name,
			"label":       params.Label,
			"description": params.Description,
			"table_name":  params.TableName,
			"field_count": len(params.Fields),
		}
		s.gateway.LogAction(tenantID, userID, "create_model", params.Name, "confirmation_requested", "pending")
		return ToolCallResult{
			Success: false,
			Warning: "写操作需要用户二次确认。请将 confirmed 设为 true 后重试。",
			Data: ConfirmationRequest{
				NeedsConfirmation: true,
				Message:           fmt.Sprintf("确认创建数据模型 %q（表名: %s）？此操作将创建新的数据表并记录审计日志。", params.Label, params.TableName),
				Operation:         "create_model",
				Summary:           summary,
			},
		}
	}

	var fieldConfigs []model.FieldConfig
	for _, f := range params.Fields {
		fieldConfigs = append(fieldConfigs, model.FieldConfig{
			Name:        f.Name,
			Label:       f.Label,
			Type:        f.Type,
			Required:    f.Required,
			Unique:      f.Unique,
			Default:     f.Default,
			MaxLength:   f.MaxLength,
			Description: f.Description,
		})
	}

	m, fields, err := s.modelSvc.CreateModel(params.TenantID, params.Name, params.Label,
		params.Description, params.TableName, fieldConfigs)
	if err != nil {
		s.gateway.LogAction(tenantID, userID, "create_model", params.Name, err.Error(), "failed")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	info := toModelInfo(m, convertFieldsFromService(fields))
	s.gateway.LogAction(tenantID, userID, "create_model", params.Name,
		fmt.Sprintf("created model with %d fields", len(fields)), "success")
	return ToolCallResult{Success: true, Data: info}
}

func (s *Server) handleAddField(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermWriteModels); err != nil {
		s.gateway.LogAction(tenantID, userID, "add_field", "models", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	var params AddFieldParams
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{Success: false, Error: "invalid params: " + err.Error()}
	}

	// Require user confirmation for write operations.
	if !params.Confirmed {
		summary := map[string]interface{}{
			"model":     params.ModelName,
			"field":     params.Field.Name,
			"type":      params.Field.Type,
			"required":  params.Field.Required,
		}
		s.gateway.LogAction(tenantID, userID, "add_field", params.ModelName+"."+params.Field.Name,
			"confirmation_requested", "pending")
		return ToolCallResult{
			Success: false,
			Warning: "写操作需要用户二次确认。请将 confirmed 设为 true 后重试。",
			Data: ConfirmationRequest{
				NeedsConfirmation: true,
				Message: fmt.Sprintf("确认向模型 %q 添加字段 %q（类型: %s）？",
					params.ModelName, params.Field.Name, params.Field.Type),
				Operation: "add_field",
				Summary:   summary,
			},
		}
	}

	fc := model.FieldConfig{
		Name:        params.Field.Name,
		Label:       params.Field.Label,
		Type:        params.Field.Type,
		Required:    params.Field.Required,
		Unique:      params.Field.Unique,
		Default:     params.Field.Default,
		MaxLength:   params.Field.MaxLength,
		Description: params.Field.Description,
	}

	field, err := s.modelSvc.AddField(params.TenantID, params.ModelName, fc)
	if err != nil {
		s.gateway.LogAction(tenantID, userID, "add_field", params.ModelName+"."+params.Field.Name,
			err.Error(), "failed")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	result := FieldInfo{
		ID:          field.ID,
		Name:        field.Name,
		Label:       field.Label,
		Type:        field.Type,
		Required:    field.Required,
		Unique:      field.Unique,
		MaxLength:   field.MaxLength,
		Description: field.Description,
	}

	s.gateway.LogAction(tenantID, userID, "add_field", params.ModelName+"."+params.Field.Name,
		"", "success")
	return ToolCallResult{Success: true, Data: result}
}

func (s *Server) handleQueryData(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermQueryData); err != nil {
		s.gateway.LogAction(tenantID, userID, "query_data", "data", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	var params QueryDataParams
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolCallResult{Success: false, Error: "invalid params: " + err.Error()}
	}

	// Inject tenant filter into SQL.
	filteredSQL, err := gateway.InjectTenantFilter(params.SQL, params.TenantID)
	if err != nil {
		s.gateway.LogAction(tenantID, userID, "query_data", "data", err.Error(), "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	s.gateway.LogAction(tenantID, userID, "query_data", "data",
		fmt.Sprintf("original=%s, filtered=%s", truncate(params.SQL, 100), truncate(filteredSQL, 100)),
		"success")

	return ToolCallResult{
		Success: true,
		Data: map[string]interface{}{
			"executed_sql": filteredSQL,
			"note":         "Query validated and tenant filter injected. In production, results would be returned here.",
		},
	}
}

func (s *Server) handleGetSemanticLayer(args json.RawMessage, tenantID, userID string) ToolCallResult {
	if err := s.gateway.CheckPermission(tenantID, userID, gateway.PermReadSemantic); err != nil {
		s.gateway.LogAction(tenantID, userID, "get_semantic_layer", "semantic", "", "denied")
		return ToolCallResult{Success: false, Error: err.Error()}
	}

	s.gateway.LogAction(tenantID, userID, "get_semantic_layer", "semantic", "", "success")
	return ToolCallResult{Success: true, Data: s.semantic}
}

// ---------- helpers ----------

func toModelInfo(m *model.DataModel, fields []FieldInfo) ModelInfo {
	return ModelInfo{
		ID:          m.ID,
		Name:        m.Name,
		Label:       m.Label,
		Description: m.Description,
		TableName:   m.TableName,
		Fields:      fields,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func convertFieldsFromService(fields []*model.ModelField) []FieldInfo {
	var result []FieldInfo
	for _, f := range fields {
		var def any
		if f.Default != nil {
			def = *f.Default
		}
		result = append(result, FieldInfo{
			ID:          f.ID,
			Name:        f.Name,
			Label:       f.Label,
			Type:        f.Type,
			Required:    f.Required,
			Unique:      f.Unique,
			Default:     def,
			MaxLength:   f.MaxLength,
			Description: f.Description,
		})
	}
	return result
}

func extractTenantID(args json.RawMessage) string {
	var m map[string]interface{}
	if json.Unmarshal(args, &m) == nil {
		if tid, ok := m["tenant_id"].(string); ok {
			return tid
		}
	}
	return "unknown"
}

func extractUserID(args json.RawMessage) string {
	var m map[string]interface{}
	if json.Unmarshal(args, &m) == nil {
		if uid, ok := m["user_id"].(string); ok {
			return uid
		}
	}
	return "system"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
