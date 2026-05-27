package builder

import (
	"context"
	"encoding/json"
	"fmt"
)

// Agent is the top-level Builder Agent that orchestrates the four-phase
// conversational workflow to convert natural language into ERP metadata configurations.
type Agent struct {
	cm       *ConversationManager
	ctx      *ContextProvider
	gen      *ConfigGenerator
	val      *ConfigValidator
	executor ToolExecutor
}

// AgentConfig holds the dependencies for creating a Builder Agent.
type AgentConfig struct {
	ConversationStore ConversationStore
	ExistingModels    []ModelConfig
	ToolExecutor      ToolExecutor
}

// NewAgent creates a fully wired Builder Agent.
func NewAgent(cfg AgentConfig) *Agent {
	ctxProv := NewContextProvider(cfg.ExistingModels)
	exec := cfg.ToolExecutor
	if exec == nil {
		exec = NewMockToolExecutor()
	}
	return &Agent{
		cm:       NewConversationManager(cfg.ConversationStore),
		ctx:      ctxProv,
		gen:      NewConfigGenerator(ctxProv),
		val:      NewConfigValidator(ctxProv),
		executor: exec,
	}
}

// ProcessMessage is the main entry point for the Builder Agent.
// It accepts a user message and returns the agent's response, handling state
// transitions internally based on the conversation flow.
func (a *Agent) ProcessMessage(ctx context.Context, req BuildRequest) (*BuildResponse, error) {
	session, err := a.resolveSession(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("resolve session: %w", err)
	}
	sessionID := session.SessionID

	if err := a.cm.AddMessage(sessionID, "user", req.UserMessage); err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	// Route based on current state and user intent.
	switch session.State {
	case StateRequirements:
		return a.handleRequirements(ctx, session)
	case StateSolution:
		return a.handleSolution(ctx, session)
	case StateConfirmation:
		return a.handleConfirmation(ctx, session)
	case StateCreation:
		return a.handleCreation(ctx, session)
	default:
		return a.handleRequirements(ctx, session)
	}
}

// StartSession begins a new builder conversation.
func (a *Agent) StartSession() *Conversation {
	return a.cm.StartSession()
}

// GetSession returns the current state of a conversation.
func (a *Agent) GetSession(sessionID string) (*Conversation, error) {
	return a.cm.GetSession(sessionID)
}

// GetExistingModels returns summaries of all models in the system.
func (a *Agent) GetExistingModels() []ModelSummary {
	return a.ctx.ListSummaries()
}

// GetTools returns the MCP tool definitions available to the agent.
func (a *Agent) GetTools() []MCPTool {
	return a.executor.ListTools()
}

// ExecuteTool runs a tool call specified by the agent (typically from an LLM function call).
func (a *Agent) ExecuteTool(ctx context.Context, call MCPToolCall) (json.RawMessage, error) {
	return a.executor.Execute(ctx, call)
}

// resolveSession gets an existing session or starts a new one.
func (a *Agent) resolveSession(sessionID string) (*Conversation, error) {
	if sessionID != "" {
		sess, err := a.cm.GetSession(sessionID)
		if err == nil {
			return sess, nil
		}
	}
	return a.cm.StartSession(), nil
}

// handleRequirements processes user input in the requirements gathering phase.
func (a *Agent) handleRequirements(ctx context.Context, session *Conversation) (*BuildResponse, error) {
	lastMsg := lastUserMessage(session)

	// Parse the user's intent from their natural language input.
	intent := ParseIntent(lastMsg.Content, a.ctx.ListSummaries())

	// If the intent is incomplete, ask clarifying questions.
	questions := assessCompleteness(intent)
	if len(questions) > 0 {
		msg := buildClarificationMessage(questions, a.ctx.ListSummaries())
		if err := a.cm.AddMessage(session.SessionID, "assistant", msg); err != nil {
			return nil, err
		}
		return &BuildResponse{
			SessionID:       session.SessionID,
			State:           StateRequirements,
			Message:         msg,
			ExistingModels:  a.ctx.ListSummaries(),
		}, nil
	}

	// Requirements are complete — generate config proposals.
	configs, err := a.gen.GenerateFromIntent(intent)
	if err != nil {
		return nil, fmt.Errorf("generate configs: %w", err)
	}

	if err := a.cm.TransitionState(session.SessionID, StateSolution); err != nil {
		return nil, err
	}
	if err := a.cm.SetProposedConfigs(session.SessionID, configs); err != nil {
		return nil, err
	}

	// Build the solution presentation message.
	msg := a.buildSolutionMessage(intent, configs)
	if err := a.cm.AddMessage(session.SessionID, "assistant", msg); err != nil {
		return nil, err
	}

	return &BuildResponse{
		SessionID:       session.SessionID,
		State:           StateSolution,
		Message:         msg,
		ProposedConfigs: configs,
		ExistingModels:  a.ctx.ListSummaries(),
	}, nil
}

// handleSolution processes user feedback on a generated solution.
func (a *Agent) handleSolution(ctx context.Context, session *Conversation) (*BuildResponse, error) {
	lastMsg := lastUserMessage(session)
	content := lastMsg.Content

	// Check if the user wants modifications or accepts the proposal.
	if isAffirmative(content) {
		// User agrees — move to confirmation.
		if err := a.cm.TransitionState(session.SessionID, StateConfirmation); err != nil {
			return nil, err
		}

		// Run validation and present results.
		errs := a.val.ValidateAll(session.ProposedConfigs)
		msg := FormatValidationErrors(errs)
		if len(errs) == 0 || onlyWarnings(errs) {
			msg += "\n\n如确认无误，请回复"确认创建"以执行创建操作。如需修改，请说明调整内容。"
		}
		if err := a.cm.AddMessage(session.SessionID, "assistant", msg); err != nil {
			return nil, err
		}

		return &BuildResponse{
			SessionID:        session.SessionID,
			State:            StateConfirmation,
			Message:          msg,
			ProposedConfigs:  session.ProposedConfigs,
			ValidationErrors: errs,
		}, nil
	}

	// User wants modifications — go back to requirements with the feedback.
	if err := a.cm.TransitionState(session.SessionID, StateRequirements); err != nil {
		return nil, err
	}

	// Re-parse with the modification feedback.
	modifiedMsg := content + "\n\n(上述是对之前方案的修改意见)"
	if err := a.cm.AddMessage(session.SessionID, "user", modifiedMsg); err != nil {
		return nil, err
	}

	return a.handleRequirements(ctx, session)
}

// handleConfirmation processes the final confirmation before creation.
func (a *Agent) handleConfirmation(ctx context.Context, session *Conversation) (*BuildResponse, error) {
	lastMsg := lastUserMessage(session)

	if !isAffirmative(lastMsg.Content) {
		// User changed mind — go back to requirements.
		if err := a.cm.TransitionState(session.SessionID, StateRequirements); err != nil {
			return nil, err
		}
		msg := "好的，请告诉我需要如何调整？"
		if err := a.cm.AddMessage(session.SessionID, "assistant", msg); err != nil {
			return nil, err
		}
		return &BuildResponse{
			SessionID: session.SessionID,
			State:     StateRequirements,
			Message:   msg,
		}, nil
	}

	// User confirmed — execute creation.
	if err := a.cm.TransitionState(session.SessionID, StateCreation); err != nil {
		return nil, err
	}

	var created []ModelSummary
	var failures []string

	// Create models in dependency order (non-relation models first).
	ordered := sortByDependency(session.ProposedConfigs)
	for _, cfg := range ordered {
		result, err := a.executor.Execute(ctx, MCPToolCall{
			ToolName: "create_model",
			Args: map[string]interface{}{
				"config": cfg,
			},
		})
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", cfg.Name, err))
			continue
		}
		_ = result
		a.cm.RecordCreatedModel(session.SessionID, cfg.Name)
		created = append(created, ModelSummary{
			ID:          cfg.Name,
			Name:        cfg.Name,
			DisplayName: cfg.DisplayName,
			FieldCount:  len(cfg.Fields),
		})
	}

	msg := buildCreationResult(created, failures)
	if err := a.cm.AddMessage(session.SessionID, "assistant", msg); err != nil {
		return nil, err
	}

	if len(failures) == 0 {
		a.cm.TransitionState(session.SessionID, StateDone)
	}

	return &BuildResponse{
		SessionID:     session.SessionID,
		State:         StateDone,
		Message:       msg,
		CreatedModels: created,
	}, nil
}

// handleCreation processes input after models have been created.
func (a *Agent) handleCreation(ctx context.Context, session *Conversation) (*BuildResponse, error) {
	// After creation, any new request starts a fresh requirements cycle.
	if err := a.cm.TransitionState(session.SessionID, StateRequirements); err != nil {
		return nil, err
	}
	return a.handleRequirements(ctx, session)
}

// buildSolutionMessage constructs the human-readable solution presentation.
func (a *Agent) buildSolutionMessage(intent *BuildIntent, configs []ModelConfig) string {
	msg := "根据您的需求，我生成了以下元数据配置方案：\n\n"
	for i, cfg := range configs {
		msg += fmt.Sprintf("### 模型 %d: %s (%s)\n", i+1, cfg.DisplayName, cfg.Name)
		if cfg.Description != "" {
			msg += fmt.Sprintf("> %s\n\n", cfg.Description)
		}
		msg += "| 字段名 | 显示名 | 类型 | 必填 | 说明 |\n"
		msg += "|--------|--------|------|------|------|\n"
		for _, f := range cfg.Fields {
			required := ""
			if f.Required {
				required = "是"
			}
			typeDesc := string(f.Type)
			if f.Type == FieldRelation && f.RelationModel != "" {
				typeDesc = fmt.Sprintf("→ %s", f.RelationModel)
			}
			desc := f.Description
			if desc == "" {
				desc = "-"
			}
			msg += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				f.Name, f.DisplayName, typeDesc, required, desc)
		}
		msg += "\n"

		// Add relation reuse suggestions.
		suggestions := a.gen.SuggestRelations(cfg)
		for _, s := range suggestions {
			msg += fmt.Sprintf("💡 %s\n", s.Message)
		}
		msg += "\n"
	}
	msg += "---\n请确认以上方案是否符合您的预期？您可以直接确认，或提出修改意见。"
	return msg
}

func lastUserMessage(session *Conversation) Message {
	for i := len(session.Messages) - 1; i >= 0; i-- {
		if session.Messages[i].Role == "user" {
			return session.Messages[i]
		}
	}
	return Message{}
}

func isAffirmative(content string) bool {
	affirmatives := []string{"确认", "是的", "好的", "可以", "没问题", "同意", "创建", "ok", "yes", "对", "行", "好"}
	content = normalizeForMatch(content)
	for _, a := range affirmatives {
		if len(content) >= len(a) && content[:len(a)] == a {
			return true
		}
	}
	return false
}

func normalizeForMatch(s string) string {
	// Strip common sentence-leading words.
	result := s
	for _, prefix := range []string{"确认创建", "请", "那就", "那么就", "帮我"} {
		if len(result) >= len(prefix) && result[:len(prefix)] == prefix {
			result = result[len(prefix):]
		}
	}
	return result
}

func onlyWarnings(errs []ValidationError) bool {
	for _, e := range errs {
		if e.Severity == "error" {
			return false
		}
	}
	return true
}

// sortByDependency orders configs so that models without relation dependencies
// are created first.
func sortByDependency(configs []ModelConfig) []ModelConfig {
	if len(configs) <= 1 {
		return configs
	}

	// Collect names of models in this batch.
	inBatch := make(map[string]bool)
	for _, c := range configs {
		inBatch[c.Name] = true
	}

	// Topological sort: models with no in-batch dependencies first.
	hasDep := make(map[string]bool)
	for _, c := range configs {
		for _, f := range c.Fields {
			if f.Type == FieldRelation && inBatch[f.RelationModel] {
				hasDep[c.Name] = true
			}
		}
	}

	var ordered, deferred []ModelConfig
	for _, c := range configs {
		if hasDep[c.Name] {
			deferred = append(deferred, c)
		} else {
			ordered = append(ordered, c)
		}
	}
	ordered = append(ordered, deferred...)
	return ordered
}

func buildCreationResult(created []ModelSummary, failures []string) string {
	if len(created) > 0 {
		msg := "✅ 已成功创建以下模型：\n\n"
		for _, m := range created {
			msg += fmt.Sprintf("- **%s** (%s)：%d 个字段\n", m.DisplayName, m.Name, m.FieldCount)
		}
		if len(failures) > 0 {
			msg += "\n❌ 以下模型创建失败：\n"
			for _, f := range failures {
				msg += fmt.Sprintf("- %s\n", f)
			}
		}
		msg += "\n如有其他需求，请继续描述。"
		return msg
	}
	return "未能创建任何模型，请检查错误信息后重试。"
}
