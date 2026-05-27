package builder

import (
	"fmt"
	"strings"
)

// ContextProvider supplies the Builder Agent with information about the existing
// ERP model landscape — models, fields, and relationships — to enable reuse.
type ContextProvider struct {
	models map[string]*ModelConfig
	byName map[string]string // display_name → name lookup
}

// NewContextProvider creates a ContextProvider preloaded with existing models.
func NewContextProvider(models []ModelConfig) *ContextProvider {
	cp := &ContextProvider{
		models: make(map[string]*ModelConfig),
		byName: make(map[string]string),
	}
	for i := range models {
		cp.models[models[i].Name] = &models[i]
		cp.byName[models[i].DisplayName] = models[i].Name
	}
	return cp
}

// ListSummaries returns lightweight summaries of all models in the system.
func (cp *ContextProvider) ListSummaries() []ModelSummary {
	summaries := make([]ModelSummary, 0, len(cp.models))
	for _, m := range cp.models {
		summaries = append(summaries, ModelSummary{
			ID:          m.Name,
			Name:        m.Name,
			DisplayName: m.DisplayName,
			FieldCount:  len(m.Fields),
		})
	}
	return summaries
}

// GetModel returns the full config of a model by name.
func (cp *ContextProvider) GetModel(name string) (*ModelConfig, error) {
	m, ok := cp.models[name]
	if !ok {
		return nil, fmt.Errorf("model %s not found", name)
	}
	return m, nil
}

// FindOverlappingFields checks if the proposed fields overlap with fields already
// defined in existing models. Returns matching field→model pairs that the agent
// can suggest as relation targets instead of creating duplicates.
func (cp *ContextProvider) FindOverlappingFields(proposed ModelConfig) []OverlapResult {
	var overlaps []OverlapResult
	for _, existingModel := range cp.models {
		if existingModel.Name == proposed.Name {
			continue
		}
		for _, existingField := range existingModel.Fields {
			for _, proposedField := range proposed.Fields {
				if fieldsOverlap(existingField, proposedField, existingModel.Name) {
					overlaps = append(overlaps, OverlapResult{
						ProposedField:  proposedField.Name,
						ExistingModel:  existingModel.Name,
						ExistingField:  existingField.Name,
						ModelLabel:     existingModel.DisplayName,
						Confidence:     overlapConfidence(existingField, proposedField),
					})
				}
			}
		}
	}
	return overlaps
}

// SearchModels finds models whose name or display name match a query string.
func (cp *ContextProvider) SearchModels(query string) []ModelSummary {
	query = strings.ToLower(query)
	var results []ModelSummary
	for name, m := range cp.models {
		if strings.Contains(strings.ToLower(name), query) ||
			strings.Contains(strings.ToLower(m.DisplayName), query) {
			results = append(results, ModelSummary{
				ID:          name,
				Name:        name,
				DisplayName: m.DisplayName,
				FieldCount:  len(m.Fields),
			})
		}
	}
	return results
}

// OverlapResult describes a field in the proposal that overlaps with an existing field.
type OverlapResult struct {
	ProposedField string  `json:"proposed_field"`
	ExistingModel string  `json:"existing_model"`
	ExistingField string  `json:"existing_field"`
	ModelLabel    string  `json:"model_label"`
	Confidence    float64 `json:"confidence"`
}

func fieldsOverlap(a, existing FieldConfig, modelName string) bool {
	// Direct name match
	if a.Name == existing.Name {
		return true
	}
	// Same type + similar name hints at conceptual overlap
	if a.Type == existing.Type && semanticOverlap(a.Name, existing.Name) {
		return true
	}
	// A relation field that already references this model
	if a.Type == FieldRelation && a.RelationModel == modelName {
		return true
	}
	return false
}

func semanticOverlap(a, b string) bool {
	return a == b || strings.Contains(a, b) || strings.Contains(b, a)
}

func overlapConfidence(existing FieldConfig, proposed FieldConfig) float64 {
	if existing.Name == proposed.Name && existing.Type == proposed.Type {
		return 0.95
	}
	if existing.Name == proposed.Name {
		return 0.7
	}
	if existing.Type == proposed.Type && (strings.Contains(proposed.Name, existing.Name) ||
		strings.Contains(existing.Name, proposed.Name)) {
		return 0.5
	}
	return 0.3
}
