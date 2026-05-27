package builder

import (
	"fmt"
	"regexp"
	"strings"
)

// ConfigValidator checks metadata configurations for correctness before creation.
type ConfigValidator struct {
	ctx *ContextProvider
}

// NewConfigValidator creates a validator with awareness of existing models.
func NewConfigValidator(ctx *ContextProvider) *ConfigValidator {
	return &ConfigValidator{ctx: ctx}
}

// Validate checks a model config and returns all validation errors.
// An empty slice means the config is valid.
func (v *ConfigValidator) Validate(config ModelConfig) []ValidationError {
	var errs []ValidationError

	errs = append(errs, v.validateName(config)...)
	errs = append(errs, v.validateFields(config)...)
	errs = append(errs, v.validateIndexes(config)...)
	errs = append(errs, v.validateUniqueness(config)...)

	return errs
}

// ValidateAll runs validation across multiple configs. It also checks for
// cross-model consistency (e.g., relation targets exist).
func (v *ConfigValidator) ValidateAll(configs []ModelConfig) []ValidationError {
	var errs []ValidationError
	seen := make(map[string]bool)

	for _, cfg := range configs {
		errs = append(errs, v.Validate(cfg)...)
		seen[cfg.Name] = true
	}

	// Cross-model: verify relation targets exist either in existing models
	// or in the batch being created.
	for _, cfg := range configs {
		for _, f := range cfg.Fields {
			if f.Type == FieldRelation && f.RelationModel != "" {
				if !modelExists(f.RelationModel, v.ctx) && !seen[f.RelationModel] {
					errs = append(errs, ValidationError{
						ModelName: cfg.Name,
						FieldName: f.Name,
						Message:   fmt.Sprintf("关联模型 '%s' 不存在，请确认该模型已创建或调整创建顺序", f.RelationModel),
						Severity:  "error",
					})
				}
			}
		}
	}

	return errs
}

var fieldNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (v *ConfigValidator) validateName(config ModelConfig) []ValidationError {
	var errs []ValidationError

	if config.Name == "" {
		errs = append(errs, ValidationError{
			ModelName: config.Name,
			Message:   "模型名不能为空",
			Severity:  "error",
		})
	} else if !fieldNameRe.MatchString(config.Name) {
		errs = append(errs, ValidationError{
			ModelName: config.Name,
			Message:   "模型名只能包含小写字母、数字和下划线，且必须以字母开头",
			Severity:  "error",
		})
	}

	if config.DisplayName == "" {
		errs = append(errs, ValidationError{
			ModelName: config.Name,
			Message:   "模型显示名不能为空",
			Severity:  "error",
		})
	}

	// Check name collision with existing models.
	if v.ctx != nil {
		if _, err := v.ctx.GetModel(config.Name); err == nil {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   fmt.Sprintf("模型 '%s' 已存在，请使用其他名称或使用 update_model", config.Name),
				Severity:  "error",
			})
		}
	}

	return errs
}

func (v *ConfigValidator) validateFields(config ModelConfig) []ValidationError {
	var errs []ValidationError

	if len(config.Fields) == 0 {
		errs = append(errs, ValidationError{
			ModelName: config.Name,
			Message:   "模型至少需要包含一个字段",
			Severity:  "error",
		})
		return errs
	}

	fieldNames := make(map[string]bool)
	for _, f := range config.Fields {
		if f.Name == "" {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   "存在字段名为空",
				Severity:  "error",
			})
			continue
		}

		if !fieldNameRe.MatchString(f.Name) {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				FieldName: f.Name,
				Message:   "字段名只能包含小写字母、数字和下划线",
				Severity:  "error",
			})
		}

		if fieldNames[f.Name] {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				FieldName: f.Name,
				Message:   "字段名重复",
				Severity:  "error",
			})
		}
		fieldNames[f.Name] = true

		errs = append(errs, v.validateFieldType(config.Name, f)...)
	}
	return errs
}

func (v *ConfigValidator) validateFieldType(modelName string, f FieldConfig) []ValidationError {
	var errs []ValidationError

	switch f.Type {
	case FieldRelation:
		if f.RelationModel == "" {
			errs = append(errs, ValidationError{
				ModelName: modelName,
				FieldName: f.Name,
				Message:   "relation 类型字段必须指定 relation_model",
				Severity:  "error",
			})
		}
	case FieldEnum:
		if len(f.EnumValues) == 0 {
			errs = append(errs, ValidationError{
				ModelName: modelName,
				FieldName: f.Name,
				Message:   "enum 类型字段必须提供 enum_values",
				Severity:  "error",
			})
		}
	case FieldString:
		if f.MaxLength == 0 {
			// Warn but don't error — defaults to 255 at the engine level.
			errs = append(errs, ValidationError{
				ModelName: modelName,
				FieldName: f.Name,
				Message:   "string 字段未指定 max_length，将使用默认值 255",
				Severity:  "warning",
			})
		}
		if f.MaxLength > 4000 {
			errs = append(errs, ValidationError{
				ModelName: modelName,
				FieldName: f.Name,
				Message:   "string 字段 max_length 超过 4000，建议使用 text 类型",
				Severity:  "warning",
			})
		}
	case FieldDecimal:
		if f.MinValue != nil && f.MaxValue != nil && *f.MinValue >= *f.MaxValue {
			errs = append(errs, ValidationError{
				ModelName: modelName,
				FieldName: f.Name,
				Message:   "min_value 必须小于 max_value",
				Severity:  "error",
			})
		}
	case "":
		errs = append(errs, ValidationError{
			ModelName: modelName,
			FieldName: f.Name,
			Message:   "字段类型不能为空",
			Severity:  "error",
		})
	}

	// Check for reserved field names that conflict with system columns.
	reservedFields := map[string]bool{
		"id": true, "tenant_id": true, "deleted_at": true,
	}
	if reservedFields[f.Name] {
		errs = append(errs, ValidationError{
			ModelName: modelName,
			FieldName: f.Name,
			Message:   fmt.Sprintf("'%s' 是系统保留字段，请使用其他名称", f.Name),
			Severity:  "error",
		})
	}

	// Check display_name
	if f.DisplayName == "" {
		errs = append(errs, ValidationError{
			ModelName: modelName,
			FieldName: f.Name,
			Message:   "字段显示名不能为空",
			Severity:  "error",
		})
	}

	return errs
}

func (v *ConfigValidator) validateIndexes(config ModelConfig) []ValidationError {
	var errs []ValidationError
	indexNames := make(map[string]bool)
	fieldNames := make(map[string]bool)
	for _, f := range config.Fields {
		fieldNames[f.Name] = true
	}

	for _, idx := range config.Indexes {
		if idx.Name == "" {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   "索引名不能为空",
				Severity:  "error",
			})
			continue
		}
		if indexNames[idx.Name] {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   fmt.Sprintf("索引名 '%s' 重复", idx.Name),
				Severity:  "error",
			})
		}
		indexNames[idx.Name] = true

		if len(idx.Fields) == 0 {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   fmt.Sprintf("索引 '%s' 未指定字段", idx.Name),
				Severity:  "error",
			})
		}
		for _, idxField := range idx.Fields {
			if !fieldNames[idxField] {
				errs = append(errs, ValidationError{
					ModelName: config.Name,
					Message:   fmt.Sprintf("索引 '%s' 引用了不存在的字段 '%s'", idx.Name, idxField),
					Severity:  "error",
				})
			}
		}
	}
	return errs
}

func (v *ConfigValidator) validateUniqueness(config ModelConfig) []ValidationError {
	var errs []ValidationError

	// Recommend: every model should have at least one unique field or a clear
	// natural key. This is a warning, not an error.
	hasUnique := false
	for _, f := range config.Fields {
		if f.Unique {
			hasUnique = true
			break
		}
	}
	hasUniqueIdx := false
	for _, idx := range config.Indexes {
		if idx.Unique {
			hasUniqueIdx = true
			break
		}
	}
	if !hasUnique && !hasUniqueIdx {
		if !hasField(config.Fields, "code") && !hasField(config.Fields, "name") {
			errs = append(errs, ValidationError{
				ModelName: config.Name,
				Message:   "模型没有唯一字段或唯一索引，建议添加业务唯一标识（如 code 或 name）",
				Severity:  "warning",
			})
		}
	}

	return errs
}

func modelExists(name string, ctx *ContextProvider) bool {
	if ctx == nil {
		return false
	}
	_, err := ctx.GetModel(name)
	return err == nil
}

// FormatValidationErrors produces a human-readable summary of validation results.
func FormatValidationErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "校验通过，配置合法。"
	}

	errCount, warnCount := 0, 0
	for _, e := range errs {
		if e.Severity == "error" {
			errCount++
		} else {
			warnCount++
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("校验结果：%d 个错误，%d 个警告\n\n", errCount, warnCount))

	for _, e := range errs {
		prefix := "❌"
		if e.Severity == "warning" {
			prefix = "⚠️"
		}
		location := e.ModelName
		if e.FieldName != "" {
			location += "." + e.FieldName
		}
		sb.WriteString(fmt.Sprintf("%s [%s] %s\n", prefix, location, e.Message))
	}

	return sb.String()
}
