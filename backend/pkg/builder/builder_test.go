package builder

import (
	"context"
	"strings"
	"testing"
)

func TestParseIntent_CreateModel(t *testing.T) {
	msg := "创建一个客户管理模块，包含客户名称(string)、联系电话(string)、邮箱地址(string)、客户等级(enum)，客户名称必填且唯一"

	intent := ParseIntent(msg, nil)

	if intent.Intent != "create_model" {
		t.Errorf("expected create_model intent, got %s", intent.Intent)
	}
	if intent.DisplayName == "" {
		t.Error("expected a display name to be extracted")
	}
	if len(intent.Fields) == 0 {
		t.Error("expected fields to be extracted")
	}
}

func TestParseIntent_UpdateModel(t *testing.T) {
	msg := "修改客户模型，添加一个信用额度的decimal字段"

	intent := ParseIntent(msg, []ModelSummary{
		{Name: "customer", DisplayName: "客户", FieldCount: 5},
	})

	if intent.Intent != "update_model" {
		t.Errorf("expected update_model intent, got %s", intent.Intent)
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	v := NewConfigValidator(nil)
	cfg := ModelConfig{
		Name:        "supplier",
		DisplayName: "供应商",
		Fields: []FieldConfig{
			{Name: "name", DisplayName: "供应商名称", Type: FieldString, Required: true},
			{Name: "contact_phone", DisplayName: "联系电话", Type: FieldString},
			{Name: "status", DisplayName: "状态", Type: FieldEnum, EnumValues: []string{"active", "inactive"}},
		},
	}

	errs := v.Validate(cfg)
	for _, e := range errs {
		if e.Severity == "error" {
			t.Errorf("unexpected validation error: %s", e.Message)
		}
	}
}

func TestValidate_EmptyModelName(t *testing.T) {
	v := NewConfigValidator(nil)
	cfg := ModelConfig{
		Name:        "",
		DisplayName: "测试",
		Fields:      []FieldConfig{{Name: "name", DisplayName: "名称", Type: FieldString}},
	}

	errs := v.Validate(cfg)
	if len(errs) == 0 {
		t.Error("expected validation errors for empty model name")
	}
}

func TestValidate_EmptyFields(t *testing.T) {
	v := NewConfigValidator(nil)
	cfg := ModelConfig{
		Name:        "test",
		DisplayName: "测试",
		Fields:      []FieldConfig{},
	}

	errs := v.Validate(cfg)
	hasFieldError := false
	for _, e := range errs {
		if e.Severity == "error" && strings.Contains(e.Message, "至少需要") {
			hasFieldError = true
		}
	}
	if !hasFieldError {
		t.Error("expected error about empty fields")
	}
}

func TestValidate_DuplicateFields(t *testing.T) {
	v := NewConfigValidator(nil)
	cfg := ModelConfig{
		Name:        "test",
		DisplayName: "测试",
		Fields: []FieldConfig{
			{Name: "name", DisplayName: "名称", Type: FieldString},
			{Name: "name", DisplayName: "名称2", Type: FieldString},
		},
	}

	errs := v.Validate(cfg)
	hasDupError := false
	for _, e := range errs {
		if e.Severity == "error" && strings.Contains(e.Message, "重复") {
			hasDupError = true
		}
	}
	if !hasDupError {
		t.Error("expected duplicate field error")
	}
}

func TestValidate_RelationMissingTarget(t *testing.T) {
	v := NewConfigValidator(nil)
	cfg := ModelConfig{
		Name:        "order",
		DisplayName: "订单",
		Fields: []FieldConfig{
			{Name: "customer_id", DisplayName: "客户", Type: FieldRelation},
		},
	}

	errs := v.Validate(cfg)
	hasRelationError := false
	for _, e := range errs {
		if e.Severity == "error" && strings.Contains(e.Message, "relation_model") {
			hasRelationError = true
		}
	}
	if !hasRelationError {
		t.Error("expected error about missing relation_model")
	}
}

func TestValidateAll_CrossModelRelation(t *testing.T) {
	cp := NewContextProvider([]ModelConfig{
		{Name: "customer", DisplayName: "客户", Fields: []FieldConfig{{Name: "name", DisplayName: "名称", Type: FieldString}}},
	})
	v := NewConfigValidator(cp)

	configs := []ModelConfig{
		{
			Name:        "order",
			DisplayName: "订单",
			Fields: []FieldConfig{
				{Name: "customer_id", DisplayName: "客户", Type: FieldRelation, RelationModel: "customer"},
			},
		},
	}

	errs := v.ValidateAll(configs)
	for _, e := range errs {
		if e.Severity == "error" {
			t.Errorf("unexpected validation error: %s", e.Message)
		}
	}
}

func TestValidate_ReservedFields(t *testing.T) {
	v := NewConfigValidator(nil)
	reserved := []string{"id", "tenant_id", "deleted_at"}

	for _, name := range reserved {
		cfg := ModelConfig{
			Name:        "test",
			DisplayName: "测试",
			Fields:      []FieldConfig{{Name: name, DisplayName: "系统字段", Type: FieldString}},
		}
		errs := v.Validate(cfg)
		hasReservedError := false
		for _, e := range errs {
			if e.Severity == "error" && strings.Contains(e.Message, "保留") {
				hasReservedError = true
			}
		}
		if !hasReservedError {
			t.Errorf("expected reserved field error for %s", name)
		}
	}
}

func TestContextProvider_FindOverlappingFields(t *testing.T) {
	cp := NewContextProvider([]ModelConfig{
		{
			Name:        "customer",
			DisplayName: "客户",
			Fields: []FieldConfig{
				{Name: "name", DisplayName: "客户名称", Type: FieldString},
				{Name: "phone", DisplayName: "联系电话", Type: FieldString},
			},
		},
	})

	proposed := ModelConfig{
		Name: "order",
		Fields: []FieldConfig{
			{Name: "name", DisplayName: "订单名称", Type: FieldString},
			{Name: "amount", DisplayName: "金额", Type: FieldDecimal},
		},
	}

	overlaps := cp.FindOverlappingFields(proposed)
	if len(overlaps) == 0 {
		t.Error("expected overlap between order.name and customer.name")
	}
}

func TestConfigGenerator_SuggestRelations(t *testing.T) {
	cp := NewContextProvider([]ModelConfig{
		{
			Name:        "customer",
			DisplayName: "客户",
			Fields: []FieldConfig{
				{Name: "name", DisplayName: "客户名称", Type: FieldString},
			},
		},
	})

	gen := NewConfigGenerator(cp)
	config := ModelConfig{
		Name: "order",
		Fields: []FieldConfig{
			{Name: "name", DisplayName: "订单名称", Type: FieldString},
		},
	}

	suggestions := gen.SuggestRelations(config)
	if len(suggestions) == 0 {
		t.Error("expected suggestions about overlapping field 'name'")
	}
}

func TestConversationManager_StateTransitions(t *testing.T) {
	cm := NewConversationManager(nil)
	session := cm.StartSession()

	// Valid: requirements → solution
	if err := cm.TransitionState(session.SessionID, StateSolution); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Valid: solution → confirmation
	if err := cm.TransitionState(session.SessionID, StateConfirmation); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid: confirmation → done (must go through creation first)
	if err := cm.TransitionState(session.SessionID, StateDone); err == nil {
		t.Error("expected error for invalid transition confirmation → done")
	}

	// Valid: confirmation → creation
	if err := cm.TransitionState(session.SessionID, StateCreation); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_FullWorkflow(t *testing.T) {
	// Pre-populate with an existing model to test reuse.
	existingModels := []ModelConfig{
		{
			Name:        "customer",
			DisplayName: "客户",
			Fields: []FieldConfig{
				{Name: "name", DisplayName: "客户名称", Type: FieldString, Required: true},
				{Name: "phone", DisplayName: "联系电话", Type: FieldString},
			},
		},
	}

	agent := NewAgent(AgentConfig{
		ExistingModels: existingModels,
	})

	// Phase 1: Start session and send requirements.
	resp1, err := agent.ProcessMessage(context.Background(), BuildRequest{
		UserMessage: "创建一个订单管理模型，包含订单编号(string, 必填, 唯一)、订单金额(decimal)、关联客户(relation→customer)、订单状态(enum: 待审核/已确认/已完成)",
	})
	if err != nil {
		t.Fatalf("phase 1 failed: %v", err)
	}
	if resp1.State != StateSolution {
		t.Errorf("expected state solution, got %s", resp1.State)
	}
	if len(resp1.ProposedConfigs) == 0 {
		t.Error("expected proposed configs")
	}

	// Phase 2: User confirms the solution.
	resp2, err := agent.ProcessMessage(context.Background(), BuildRequest{
		SessionID:   resp1.SessionID,
		UserMessage: "确认，没问题",
	})
	if err != nil {
		t.Fatalf("phase 2 failed: %v", err)
	}
	if resp2.State != StateConfirmation {
		t.Errorf("expected state confirmation, got %s", resp2.State)
	}

	// Phase 3: User confirms creation.
	resp3, err := agent.ProcessMessage(context.Background(), BuildRequest{
		SessionID:   resp1.SessionID,
		UserMessage: "确认创建",
	})
	if err != nil {
		t.Fatalf("phase 3 failed: %v", err)
	}
	if resp3.State != StateDone {
		t.Errorf("expected state done, got %s", resp3.State)
	}
	if len(resp3.CreatedModels) == 0 {
		t.Error("expected created models to be populated")
	}
}

func TestMockToolExecutor_CreateAndList(t *testing.T) {
	exec := NewMockToolExecutor()

	// Create a model.
	cfg := ModelConfig{
		Name:        "product",
		DisplayName: "产品",
		Fields: []FieldConfig{
			{Name: "name", DisplayName: "产品名称", Type: FieldString},
		},
	}

	result, err := exec.Execute(context.Background(), MCPToolCall{
		ToolName: "create_model",
		Args:     map[string]interface{}{"config": cfg},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !strings.Contains(string(result), "created") {
		t.Errorf("expected created status, got %s", string(result))
	}

	// List models.
	list, err := exec.Execute(context.Background(), MCPToolCall{
		ToolName: "list_models",
		Args:     map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(string(list), "product") {
		t.Errorf("expected product in list, got %s", string(list))
	}
}

func TestSortByDependency(t *testing.T) {
	configs := []ModelConfig{
		{
			Name: "order",
			Fields: []FieldConfig{
				{Name: "customer_id", Type: FieldRelation, RelationModel: "customer"},
			},
		},
		{
			Name:   "customer",
			Fields: []FieldConfig{{Name: "name", Type: FieldString}},
		},
	}

	ordered := sortByDependency(configs)
	if ordered[0].Name != "customer" {
		t.Errorf("expected customer first (no deps), got %s first", ordered[0].Name)
	}
}

func TestSystemPrompt_Content(t *testing.T) {
	prompt := SystemPrompt([]ModelSummary{
		{Name: "customer", DisplayName: "客户", FieldCount: 5},
	})

	if !strings.Contains(prompt, "Builder Agent") {
		t.Error("prompt should mention Builder Agent")
	}
	if !strings.Contains(prompt, "customer") {
		t.Error("prompt should include existing models")
	}
	if !strings.Contains(prompt, "list_models") {
		t.Error("prompt should mention available tools")
	}
	if !strings.Contains(prompt, "requirements") {
		t.Error("prompt should describe the four phases")
	}
}
