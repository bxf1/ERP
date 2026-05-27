package builder

import (
	"regexp"
	"strings"
)

// ParseIntent extracts structured build intent from a natural language user message.
// This is a lightweight rule-based parser; in production it would be backed by an LLM
// that has been primed with the Builder Agent system prompt.
func ParseIntent(userMessage string, existingModels []ModelSummary) *BuildIntent {
	intent := &BuildIntent{
		Intent: "create_model",
	}

	msg := strings.TrimSpace(userMessage)

	// Detect intent type.
	if match, _ := regexp.MatchString(`(修改|更新|添加字段|增加字段|加字段)`, msg); match {
		intent.Intent = "update_model"
		// Extract model name from patterns like "修改订单模型"
		intent.ModelName = extractModelName(msg)
	}

	// Extract model display name.
	intent.DisplayName = extractDisplayName(msg, existingModels)

	// Extract model name (from explicit mentions or derive from display name).
	if intent.ModelName == "" && intent.Intent == "create_model" {
		intent.ModelName = normalizeModelName(extractModelNameFromDescription(msg))
		if intent.ModelName == "" && intent.DisplayName != "" {
			intent.ModelName = normalizeModelName(intent.DisplayName)
		}
	}

	// Extract description.
	intent.Description = extractDescription(msg)

	// Extract field definitions from the message.
	intent.Fields = extractFields(msg, existingModels)

	return intent
}

// assessCompleteness checks if the parsed intent has enough information to proceed.
// Returns clarifying questions for missing information.
func assessCompleteness(intent *BuildIntent) []string {
	var questions []string

	if intent.DisplayName == "" {
		questions = append(questions, "请问您想创建什么类型的业务对象？例如：客户、订单、产品等。")
	}

	if len(intent.Fields) == 0 {
		questions = append(questions, "请描述该对象需要包含哪些字段？例如：客户名称、联系电话、地址等。")
	}

	if intent.Intent == "update_model" && intent.ModelName == "" {
		questions = append(questions, "请问您想修改哪个已有模型？")
	}

	return questions
}

func buildClarificationMessage(questions []string, existingModels []ModelSummary) string {
	msg := "我还需要了解一些信息：\n\n"
	for i, q := range questions {
		msg += strings.Repeat(" ", 0) + string(rune('1'+i)) + ". " + q + "\n"
	}
	return msg
}

func extractModelName(msg string) string {
	// Pattern: "模型名" or "XXX 模型"
	re := regexp.MustCompile(`(?:修改|更新|的)([\p{Han}a-zA-Z0-9]+)(?:模型|模块)?`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) >= 2 {
		return normalizeModelName(matches[1])
	}

	// Pattern: explicit "模型名：xxx" or "模型: xxx"
	re = regexp.MustCompile(`模型[名]?[：:]\s*([a-z_]+)`)
	matches = re.FindStringSubmatch(msg)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

func extractDisplayName(msg string, existingModels []ModelSummary) string {
	// Pattern: "创建一个 XXX" or "新建一个 XXX"
	re := regexp.MustCompile(`(?:创建|新建|建一个?|做一个?|定义一个?|加一个?)(?:新的?)?[\p{Han}a-zA-Z0-9_]+`)
	if loc := re.FindStringIndex(msg); loc != nil {
		rest := msg[loc[1]:]
		// Take first 2-4 Chinese characters as the display name.
		displayRe := regexp.MustCompile(`^[\p{Han}]{2,4}`)
		if m := displayRe.FindString(rest); m != "" {
			return m
		}
	}

	// Pattern: named "XXX 模块" or "XXX 表"
	re = regexp.MustCompile(`[\p{Han}]+(?:模型|模块|表|管理|信息)`)
	if m := re.FindString(msg); m != "" {
		return m
	}

	return ""
}

func extractModelNameFromDescription(msg string) string {
	// Look for explicit English name: "英文名/模型名 xxx" or "name: xxx"
	re := regexp.MustCompile(`(?:模型名|英文名|name)[：:]\s*([a-z_]+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func extractDescription(msg string) string {
	// Look for a description pattern.
	re := regexp.MustCompile(`(?:描述|说明|用途)[：:]\s*(.+?)(?:[。\n]|$)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractFields attempts to extract field definitions from natural language.
// This is a best-effort parser; the LLM-based prompt is the primary mechanism.
func extractFields(msg string, existingModels []ModelSummary) []FieldConfig {
	var fields []FieldConfig

	// Pattern: "字段名：类型" or "字段名（类型）"
	// Example: "客户名称：string", "联系电话：string", "地址：text"
	fieldRe := regexp.MustCompile(`[\p{Han}]{2,6}(?:名称|编号|编码|地址|电话|手机|邮箱|日期|时间|金额|数量|价格|状态|类型|描述|备注)[：:]\s*([a-zA-Z]+)`)
	matches := fieldRe.FindAllStringSubmatch(msg, -1)
	if len(matches) > 0 {
		seen := make(map[string]bool)
		for _, m := range matches {
			full := m[0]
			typeStr := m[1]
			// Extract field name part (before colon).
			parts := strings.SplitN(full, "：", 2)
			if len(parts) != 2 {
				parts = strings.SplitN(full, ":", 2)
			}
			if len(parts) != 2 {
				continue
			}
			displayName := strings.TrimSpace(parts[0])
			name := normalizeModelName(displayName)
			if seen[name] {
				continue
			}
			seen[name] = true

			ft := parseFieldType(typeStr)
			f := FieldConfig{
				Name:        name,
				DisplayName: displayName,
				Type:        ft,
				Required:    strings.Contains(msg, "必填") || strings.Contains(msg, "必输"),
			}
			fields = append(fields, f)
		}
	}

	// Pattern: list-style " - 字段名 (类型, 约束)"
	listRe := regexp.MustCompile(`[-*]\s*([\p{Han}a-zA-Z]{2,10})[：:（(]\s*([\p{Han}a-zA-Z]+)`)
	listMatches := listRe.FindAllStringSubmatch(msg, -1)
	if len(listMatches) > len(fields) {
		seen := make(map[string]bool)
		for _, f := range fields {
			seen[f.Name] = true
		}
		for _, m := range listMatches {
			displayName := strings.TrimSpace(m[1])
			name := normalizeModelName(displayName)
			if seen[name] {
				continue
			}
			seen[name] = true
			ft := parseFieldType(strings.TrimSpace(m[2]))
			fields = append(fields, FieldConfig{
				Name:        name,
				DisplayName: displayName,
				Type:        ft,
			})
		}
	}

	// Check for relation hints in existing models.
	for i, f := range fields {
		referencedModel := findReferencedModel(f.DisplayName, existingModels)
		if referencedModel != "" {
			fields[i].Type = FieldRelation
			fields[i].RelationModel = referencedModel
		}
	}

	return fields
}

func parseFieldType(s string) FieldType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "string", "字符串", "文本", "str":
		return FieldString
	case "text", "长文本", "大文本":
		return FieldText
	case "number", "int", "integer", "整数", "数字":
		return FieldNumber
	case "decimal", "float", "double", "小数", "金额", "价格":
		return FieldDecimal
	case "boolean", "bool", "布尔", "是否", "开关":
		return FieldBoolean
	case "date", "日期":
		return FieldDate
	case "datetime", "时间", "日期时间":
		return FieldDateTime
	case "enum", "枚举", "选项", "下拉":
		return FieldEnum
	case "relation", "关联", "外键", "引用":
		return FieldRelation
	case "file", "文件", "附件", "图片":
		return FieldFile
	case "json", "jsonb":
		return FieldJSON
	default:
		return FieldString
	}
}

func findReferencedModel(displayName string, existingModels []ModelSummary) string {
	name := normalizeModelName(displayName)
	for _, m := range existingModels {
		if m.Name == name || m.DisplayName == displayName {
			return m.Name
		}
	}
	return ""
}
