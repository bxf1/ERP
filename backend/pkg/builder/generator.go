package builder

import (
	"fmt"
	"regexp"
	"strings"
)

// ConfigGenerator converts parsed build intents into concrete ModelConfig objects.
type ConfigGenerator struct {
	ctx *ContextProvider
}

// NewConfigGenerator creates a ConfigGenerator with context awareness.
func NewConfigGenerator(ctx *ContextProvider) *ConfigGenerator {
	return &ConfigGenerator{ctx: ctx}
}

// GenerateFromIntent converts a parsed BuildIntent into one or more ModelConfig objects.
// It resolves relation fields against existing models and applies naming conventions.
func (g *ConfigGenerator) GenerateFromIntent(intent *BuildIntent) ([]ModelConfig, error) {
	switch intent.Intent {
	case "create_model":
		return g.generateCreate(intent)
	case "update_model":
		return g.generateUpdate(intent)
	default:
		return nil, fmt.Errorf("unsupported intent: %s", intent.Intent)
	}
}

func (g *ConfigGenerator) generateCreate(intent *BuildIntent) ([]ModelConfig, error) {
	cfg := ModelConfig{
		Name:        normalizeModelName(intent.ModelName),
		DisplayName: intent.DisplayName,
		Description: intent.Description,
		Fields:      make([]FieldConfig, 0, len(intent.Fields)),
	}

	for _, f := range intent.Fields {
		normalized := normalizeField(f)
		cfg.Fields = append(cfg.Fields, normalized)
	}

	// Auto-generate indexes for unique fields.
	for _, f := range cfg.Fields {
		if f.Unique {
			cfg.Indexes = append(cfg.Indexes, IndexConfig{
				Name:   fmt.Sprintf("idx_%s_%s", cfg.Name, f.Name),
				Fields: []string{f.Name},
				Unique: true,
			})
		}
	}

	// Auto-add common timestamp fields if not explicitly provided.
	if !hasField(cfg.Fields, "created_at") {
		cfg.Fields = append(cfg.Fields, FieldConfig{
			Name:        "created_at",
			DisplayName: "创建时间",
			Type:        FieldDateTime,
			Required:    true,
		})
	}
	if !hasField(cfg.Fields, "updated_at") {
		cfg.Fields = append(cfg.Fields, FieldConfig{
			Name:        "updated_at",
			DisplayName: "更新时间",
			Type:        FieldDateTime,
			Required:    true,
		})
	}

	return []ModelConfig{cfg}, nil
}

func (g *ConfigGenerator) generateUpdate(intent *BuildIntent) ([]ModelConfig, error) {
	// For updates, validate that the target model exists.
	if g.ctx != nil {
		if _, err := g.ctx.GetModel(intent.ModelName); err != nil {
			return nil, fmt.Errorf("cannot update: %w", err)
		}
	}
	return g.generateCreate(intent)
}

// SuggestRelations analyzes proposed fields and recommends relation replacements
// to avoid data duplication.
func (g *ConfigGenerator) SuggestRelations(config ModelConfig) []Suggestion {
	if g.ctx == nil {
		return nil
	}
	var suggestions []Suggestion
	overlaps := g.ctx.FindOverlappingFields(config)
	for _, o := range overlaps {
		if o.Confidence >= 0.5 {
			suggestions = append(suggestions, Suggestion{
				Message: fmt.Sprintf(
					"字段 '%s' 可能与已有模型 '%s'（%s）中的 '%s' 字段重叠，建议使用 relation 类型关联而非重复定义。",
					o.ProposedField, o.ExistingModel, o.ModelLabel, o.ExistingField,
				),
				ReplaceField:  o.ProposedField,
				RelationModel: o.ExistingModel,
				Confidence:    o.Confidence,
			})
		}
	}
	return suggestions
}

// Suggestion is a recommendation to use a relation instead of a duplicated field.
type Suggestion struct {
	Message       string  `json:"message"`
	ReplaceField  string  `json:"replace_field"`
	RelationModel string  `json:"relation_model"`
	Confidence    float64 `json:"confidence"`
}

var nameRe = regexp.MustCompile(`[^a-z0-9_]+`)

func normalizeModelName(name string) string {
	name = strings.ToLower(name)
	name = nameRe.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "unnamed"
	}
	return name
}

func normalizeField(f FieldConfig) FieldConfig {
	f.Name = nameRe.ReplaceAllString(strings.ToLower(f.Name), "_")
	f.Name = strings.Trim(f.Name, "_")
	if f.Type == "" {
		f.Type = FieldString
	}
	return f
}

func hasField(fields []FieldConfig, name string) bool {
	for _, f := range fields {
		if f.Name == name {
			return true
		}
	}
	return false
}
