package model

import (
	"fmt"
	"sync"
	"time"
)

// Repository provides in-memory storage for data models and fields.
// In production this would be backed by PostgreSQL JSONB columns.
type Repository struct {
	mu     sync.RWMutex
	models map[string]*DataModel // keyed by model ID
	fields map[string][]*ModelField // keyed by model ID
}

// NewRepository creates a new in-memory repository.
func NewRepository() *Repository {
	return &Repository{
		models: make(map[string]*DataModel),
		fields: make(map[string][]*ModelField),
	}
}

// ListByTenant returns all models for a tenant.
func (r *Repository) ListByTenant(tenantID string) []*DataModel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*DataModel
	for _, m := range r.models {
		if m.TenantID == tenantID {
			result = append(result, m)
		}
	}
	return result
}

// GetByName returns a model by tenant and name.
func (r *Repository) GetByName(tenantID, name string) (*DataModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.models {
		if m.TenantID == tenantID && m.Name == name {
			return m, nil
		}
	}
	return nil, fmt.Errorf("model %q not found for tenant %q", name, tenantID)
}

// GetByID returns a model by ID.
func (r *Repository) GetByID(id string) (*DataModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.models[id]
	if !ok {
		return nil, fmt.Errorf("model %q not found", id)
	}
	return m, nil
}

// Create inserts a new data model.
func (r *Repository) Create(m *DataModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check uniqueness of name per tenant.
	for _, existing := range r.models {
		if existing.TenantID == m.TenantID && existing.Name == m.Name {
			return fmt.Errorf("model with name %q already exists in tenant %q", m.Name, m.TenantID)
		}
	}

	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
	r.models[m.ID] = m
	r.fields[m.ID] = []*ModelField{}
	return nil
}

// Update modifies an existing model.
func (r *Repository) Update(m *DataModel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.models[m.ID]; !ok {
		return fmt.Errorf("model %q not found", m.ID)
	}
	m.UpdatedAt = time.Now().UTC()
	r.models[m.ID] = m
	return nil
}

// GetFields returns all fields for a model.
func (r *Repository) GetFields(modelID string) []*ModelField {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fields := r.fields[modelID]
	result := make([]*ModelField, len(fields))
	copy(result, fields)
	return result
}

// AddField adds a field to a model.
func (r *Repository) AddField(f *ModelField) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fields := r.fields[f.ModelID]
	for _, existing := range fields {
		if existing.Name == f.Name {
			return fmt.Errorf("field %q already exists in model", f.Name)
		}
	}
	r.fields[f.ModelID] = append(r.fields[f.ModelID], f)
	return nil
}
