package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DataScopeService struct {
	db           *gorm.DB
	rbacService  *RBACService
	superAdminBypass bool
}

func NewDataScopeService(db *gorm.DB, rbac *RBACService, superAdminBypass bool) *DataScopeService {
	return &DataScopeService{
		db:               db,
		rbacService:      rbac,
		superAdminBypass: superAdminBypass,
	}
}

// ScopeRule represents the filtering rule applied to a query.
type ScopeRule struct {
	ScopeType string          `json:"scope_type"`
	Field     string          `json:"field,omitempty"`
	Condition json.RawMessage `json:"condition,omitempty"`
}

// GetDataScopeFilter returns GORM scopes that filter data by the user's data scope.
// targetModel is the model name being accessed (e.g., "order", "customer").
// userContext provides the current user's identity (user_id, department_id, etc.).
func (s *DataScopeService) GetDataScopeFilter(ctx context.Context, userID uuid.UUID, targetModel string, userContext map[string]interface{}) (func(*gorm.DB) *gorm.DB, error) {
	if s.superAdminBypass {
		isSuper, _ := s.rbacService.CheckPermission(ctx, userID, "*")
		if isSuper {
			return func(db *gorm.DB) *gorm.DB { return db }, nil
		}
	}

	roles, err := s.rbacService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db.Where("1 = 0") // no roles = no data access
		}, nil
	}

	scopes, err := s.getDataScopes(ctx, roles, targetModel)
	if err != nil {
		return nil, err
	}

	return s.buildScopeFunc(scopes, userContext), nil
}

// getDataScopes fetches data scope rules for the user's roles on the target model.
func (s *DataScopeService) getDataScopes(ctx context.Context, roleCodes []string, targetModel string) ([]models.DataScope, error) {
	var roleIDs []uuid.UUID
	if err := s.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("code IN ? AND status = 'active'", roleCodes).
		Pluck("id", &roleIDs).Error; err != nil {
		return nil, fmt.Errorf("find role IDs: %w", err)
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}

	var scopes []models.DataScope
	err := s.db.WithContext(ctx).
		Where("role_id IN ?", roleIDs).
		Where("target_model = ? OR target_model = '*'", targetModel).
		Find(&scopes).Error
	if err != nil {
		return nil, fmt.Errorf("load data scopes: %w", err)
	}
	return scopes, nil
}

// buildScopeFunc converts data scope rules into a GORM scope function.
// The most permissive scope wins (all > department > self > custom).
func (s *DataScopeService) buildScopeFunc(scopes []models.DataScope, userCtx map[string]interface{}) func(*gorm.DB) *gorm.DB {
	if len(scopes) == 0 {
		return func(db *gorm.DB) *gorm.DB { return db.Where("1 = 0") }
	}

	// Check for "all" scope first — most permissive wins
	for _, scope := range scopes {
		if scope.ScopeType == "all" {
			return func(db *gorm.DB) *gorm.DB { return db }
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		conditions := db.Session(&gorm.Session{NewDB: true})
		for i, scope := range scopes {
			cond := s.scopeToCondition(scope, userCtx)
			if cond == nil {
				continue
			}
			if i == 0 {
				conditions = conditions.Where(cond.query, cond.args...)
			} else {
				conditions = conditions.Or(cond.query, cond.args...)
			}
		}
		return db.Where(conditions)
	}
}

type scopeCondition struct {
	query string
	args  []interface{}
}

// scopeToCondition converts a single scope rule to a SQL WHERE condition.
func (s *DataScopeService) scopeToCondition(scope models.DataScope, userCtx map[string]interface{}) *scopeCondition {
	switch scope.ScopeType {
	case "self":
		field := "created_by"
		userID, ok := userCtx["user_id"]
		if !ok {
			return nil
		}
		var rule map[string]interface{}
		if scope.ScopeRule != "" {
			_ = json.Unmarshal([]byte(scope.ScopeRule), &rule)
			if f, ok := rule["field"].(string); ok {
				field = f
			}
		}
		return &scopeCondition{
			query: fmt.Sprintf("%s = ?", field),
			args:  []interface{}{userID},
		}

	case "department":
		deptID, ok := userCtx["department_id"]
		if !ok {
			return nil
		}
		field := "department_id"
		var rule map[string]interface{}
		if scope.ScopeRule != "" {
			_ = json.Unmarshal([]byte(scope.ScopeRule), &rule)
			if f, ok := rule["field"].(string); ok {
				field = f
			}
		}
		return &scopeCondition{
			query: fmt.Sprintf("%s = ?", field),
			args:  []interface{}{deptID},
		}

	case "custom":
		if scope.ScopeRule == "" {
			return nil
		}
		var rule map[string]interface{}
		if err := json.Unmarshal([]byte(scope.ScopeRule), &rule); err != nil {
			return nil
		}
		return s.parseCustomRule(rule, userCtx)

	default:
		return nil
	}
}

// parseCustomRule converts a custom JSON rule to SQL.
// Supported operators: eq, neq, gt, gte, lt, lte, in, like, between.
func (s *DataScopeService) parseCustomRule(rule map[string]interface{}, userCtx map[string]interface{}) *scopeCondition {
	op, _ := rule["op"].(string)
	field, _ := rule["field"].(string)
	value := rule["value"]

	if field == "" || op == "" {
		return nil
	}

	// Resolve template variables like {{user_id}}, {{department_id}}
	resolvedVal := s.resolveValue(value, userCtx)

	switch op {
	case "eq":
		return &scopeCondition{query: fmt.Sprintf("%s = ?", field), args: []interface{}{resolvedVal}}
	case "neq":
		return &scopeCondition{query: fmt.Sprintf("%s != ?", field), args: []interface{}{resolvedVal}}
	case "gt":
		return &scopeCondition{query: fmt.Sprintf("%s > ?", field), args: []interface{}{resolvedVal}}
	case "gte":
		return &scopeCondition{query: fmt.Sprintf("%s >= ?", field), args: []interface{}{resolvedVal}}
	case "lt":
		return &scopeCondition{query: fmt.Sprintf("%s < ?", field), args: []interface{}{resolvedVal}}
	case "lte":
		return &scopeCondition{query: fmt.Sprintf("%s <= ?", field), args: []interface{}{resolvedVal}}
	case "in":
		if arr, ok := resolvedVal.([]interface{}); ok {
			if len(arr) == 0 {
				return nil
			}
			return &scopeCondition{
				query: fmt.Sprintf("%s IN ?", field),
				args:  []interface{}{arr},
			}
		}
	case "like":
		return &scopeCondition{query: fmt.Sprintf("%s LIKE ?", field), args: []interface{}{fmt.Sprintf("%%%v%%", resolvedVal)}}
	case "between":
		if arr, ok := resolvedVal.([]interface{}); ok && len(arr) == 2 {
			return &scopeCondition{
				query: fmt.Sprintf("%s BETWEEN ? AND ?", field),
				args:  []interface{}{arr[0], arr[1]},
			}
		}
	}
	return nil
}

// resolveValue resolves template placeholders in value strings.
func (s *DataScopeService) resolveValue(value interface{}, userCtx map[string]interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Resolve {{key}} placeholders
		if len(v) > 4 && v[:2] == "{{" && v[len(v)-2:] == "}}" {
			key := v[2 : len(v)-2]
			if ctxVal, ok := userCtx[key]; ok {
				return ctxVal
			}
		}
		return v
	case []interface{}:
		resolved := make([]interface{}, len(v))
		for i, item := range v {
			resolved[i] = s.resolveValue(item, userCtx)
		}
		return resolved
	default:
		return v
	}
}
