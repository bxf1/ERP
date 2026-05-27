package model

// Service provides business logic for model CRUD operations.
type Service struct {
	repo *Repository
}

// NewService creates a new model service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListModels returns all models for a tenant.
func (s *Service) ListModels(tenantID string) []*DataModel {
	return s.repo.ListByTenant(tenantID)
}

// GetModel returns a model with its fields.
func (s *Service) GetModel(tenantID, name string) (*DataModel, []*ModelField, error) {
	m, err := s.repo.GetByName(tenantID, name)
	if err != nil {
		return nil, nil, err
	}
	fields := s.repo.GetFields(m.ID)
	return m, fields, nil
}

// CreateModel creates a new data model with optional initial fields.
func (s *Service) CreateModel(tenantID, name, label, description, tableName string, fieldConfigs []FieldConfig) (*DataModel, []*ModelField, error) {
	m := NewDataModel(tenantID, name, label, description, tableName)
	if err := s.repo.Create(m); err != nil {
		return nil, nil, err
	}

	var fields []*ModelField
	for _, fc := range fieldConfigs {
		f := NewModelField(m.ID, fc)
		if err := s.repo.AddField(f); err != nil {
			return nil, nil, err
		}
		fields = append(fields, f)
	}
	return m, fields, nil
}

// GetFieldsByModelID returns all fields for a model by its ID.
func (s *Service) GetFieldsByModelID(modelID string) []*ModelField {
	return s.repo.GetFields(modelID)
}

// AddField adds a field to an existing model.
func (s *Service) AddField(tenantID, modelName string, config FieldConfig) (*ModelField, error) {
	m, err := s.repo.GetByName(tenantID, modelName)
	if err != nil {
		return nil, err
	}
	f := NewModelField(m.ID, config)
	if err := s.repo.AddField(f); err != nil {
		return nil, err
	}
	return f, nil
}
