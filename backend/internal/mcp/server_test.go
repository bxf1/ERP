package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bxf1/ERP/backend/internal/gateway"
	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/internal/semantic"
)

func setupTestServer() *Server {
	repo := model.NewRepository()
	modelSvc := model.NewService(repo)
	auditLogger := gateway.NewAuditLogger()
	secGateway := gateway.NewSecurityGateway(auditLogger)

	secGateway.GrantPermission("test-user", gateway.PermReadModels)
	secGateway.GrantPermission("test-user", gateway.PermWriteModels)
	secGateway.GrantPermission("test-user", gateway.PermQueryData)
	secGateway.GrantPermission("test-user", gateway.PermReadSemantic)

	semLayer := semantic.DefaultLayer()
	return NewServer(modelSvc, secGateway, semLayer)
}

func makeRPCRequest(t *testing.T, server *Server, method string, params interface{}) *httptest.ResponseRecorder {
	t.Helper()

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  method,
	}
	if params != nil {
		body["params"] = params
	}

	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	return w
}

func TestToolsList(t *testing.T) {
	srv := setupTestServer()
	w := makeRPCRequest(t, srv, "tools/list", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("tools is not an array")
	}

	if len(tools) != 6 {
		t.Errorf("expected 6 tools, got %d", len(tools))
	}
}

func TestListModels(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "list_models",
		"arguments": map[string]interface{}{"tenant_id": "t1", "user_id": "test-user"},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map: %v", resp.Result)
	}

	if result["success"] != true {
		t.Fatalf("expected success, got %v", result)
	}
}

func TestCreateModelConfirmationRequired(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "create_model",
		"arguments": map[string]interface{}{
			"tenant_id":  "t1",
			"user_id":    "test-user",
			"name":       "products",
			"label":      "产品",
			"table_name": "products",
			"confirmed":  false,
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != false {
		t.Errorf("expected success=false (confirmation required), got %v", result["success"])
	}

	if result["warning"] == nil || result["warning"].(string) == "" {
		t.Error("expected warning about confirmation")
	}
}

func TestCreateModelConfirmed(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "create_model",
		"arguments": map[string]interface{}{
			"tenant_id":   "t1",
			"user_id":     "test-user",
			"name":        "orders",
			"label":       "订单",
			"table_name":  "orders",
			"description": "销售订单表",
			"confirmed":   true,
			"fields": []map[string]interface{}{
				{"name": "order_no", "label": "订单号", "type": "string", "required": true},
				{"name": "amount", "label": "金额", "type": "float", "required": true},
			},
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != true {
		t.Fatalf("expected success, got error: %v", result["error"])
	}
}

func TestGetModel(t *testing.T) {
	srv := setupTestServer()

	// First create a model.
	createParams := map[string]interface{}{
		"name":      "create_model",
		"arguments": map[string]interface{}{
			"tenant_id":  "t1",
			"user_id":    "test-user",
			"name":       "inventory",
			"label":      "库存",
			"table_name": "inventory",
			"confirmed":  true,
		},
	}
	makeRPCRequest(t, srv, "tools/call", createParams)

	// Now get it.
	getParams := map[string]interface{}{
		"name":      "get_model",
		"arguments": map[string]interface{}{
			"tenant_id": "t1",
			"user_id":   "test-user",
			"name":      "inventory",
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", getParams)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != true {
		t.Fatalf("expected success, got %v", result)
	}
}

func TestAddField(t *testing.T) {
	srv := setupTestServer()

	// Create a model first.
	createParams := map[string]interface{}{
		"name":      "create_model",
		"arguments": map[string]interface{}{
			"tenant_id":  "t1",
			"user_id":    "test-user",
			"name":       "customers",
			"label":      "客户",
			"table_name": "customers",
			"confirmed":  true,
		},
	}
	makeRPCRequest(t, srv, "tools/call", createParams)

	// Add a field with confirmation.
	addParams := map[string]interface{}{
		"name":      "add_field",
		"arguments": map[string]interface{}{
			"tenant_id": "t1",
			"user_id":   "test-user",
			"model":     "customers",
			"confirmed": true,
			"field_config": map[string]interface{}{
				"name":     "email",
				"label":    "邮箱",
				"type":     "string",
				"required": true,
			},
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", addParams)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != true {
		t.Fatalf("expected success, got error: %v / %v", result["error"], result["warning"])
	}
}

func TestQueryDataTenantFilterInjection(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "query_data",
		"arguments": map[string]interface{}{
			"tenant_id": "t-123",
			"user_id":   "test-user",
			"sql":       "SELECT * FROM sales_orders WHERE status = 'completed'",
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != true {
		t.Fatalf("expected success, got %v", result["error"])
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map")
	}

	executedSQL, ok := data["executed_sql"].(string)
	if !ok {
		t.Fatal("executed_sql not found")
	}

	if !strings.Contains(executedSQL, "tenant_id = 't-123'") {
		t.Errorf("tenant filter not injected, got: %s", executedSQL)
	}
}

func TestQueryDataRejectsNonSelect(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "query_data",
		"arguments": map[string]interface{}{
			"tenant_id": "t1",
			"user_id":   "test-user",
			"sql":       "DELETE FROM users WHERE id = 1",
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != false {
		t.Error("DELETE query should be rejected")
	}
}

func TestGetSemanticLayer(t *testing.T) {
	srv := setupTestServer()

	params := map[string]interface{}{
		"name":      "get_semantic_layer",
		"arguments": map[string]interface{}{
			"tenant_id": "t1",
			"user_id":   "test-user",
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != true {
		t.Fatalf("expected success, got %v", result["error"])
	}
}

func TestPermissionDenied(t *testing.T) {
	repo := model.NewRepository()
	modelSvc := model.NewService(repo)
	auditLogger := gateway.NewAuditLogger()
	secGateway := gateway.NewSecurityGateway(auditLogger)
	// No permissions granted.
	semLayer := semantic.DefaultLayer()
	srv := NewServer(modelSvc, secGateway, semLayer)

	params := map[string]interface{}{
		"name":      "list_models",
		"arguments": map[string]interface{}{
			"tenant_id": "t1",
			"user_id":   "no-perms-user",
		},
	}
	w := makeRPCRequest(t, srv, "tools/call", params)

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map")
	}

	if result["success"] != false {
		t.Error("user without permissions should be denied")
	}
}

func TestDuplicateModelNameRejected(t *testing.T) {
	srv := setupTestServer()

	createArgs := map[string]interface{}{
		"name":      "create_model",
		"arguments": map[string]interface{}{
			"tenant_id":  "t1",
			"user_id":    "test-user",
			"name":       "dup_model",
			"label":      "重复模型",
			"table_name": "dup_model",
			"confirmed":  true,
		},
	}

	// First create should succeed.
	w1 := makeRPCRequest(t, srv, "tools/call", createArgs)
	var resp1 JSONRPCResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	r1, _ := resp1.Result.(map[string]interface{})
	if r1["success"] != true {
		t.Fatalf("first create should succeed: %v", r1["error"])
	}

	// Second create with same name should fail.
	w2 := makeRPCRequest(t, srv, "tools/call", createArgs)
	var resp2 JSONRPCResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	r2, _ := resp2.Result.(map[string]interface{})
	if r2["success"] != false {
		t.Error("duplicate model name should be rejected")
	}
}
