package semantic

import (
	"fmt"
	"sync"
	"time"

	"github.com/bxf1/ERP/backend/pkg/datadict"
)

// Service is the public API of the semantic layer.
// It holds the parsed configuration and provides metric lookup,
// SQL generation, and LLM context assembly.
type Service struct {
	mu         sync.RWMutex
	cfg        *Config
	builder    *SQLBuilder
	ctxBuilder *ContextBuilder

	// Metric index for fast lookup.
	metricIndex map[string]Metric
	modelIndex  map[string]SemanticModel
}

// NewService creates a semantic-layer Service wired to the data dictionary
// and loaded from the given Config.
func NewService(dictSvc *datadict.Service, cfg *Config) *Service {
	svc := &Service{
		cfg:         cfg,
		builder:     NewSQLBuilder("postgres"),
		ctxBuilder:  NewContextBuilder(dictSvc, cfg),
		metricIndex: make(map[string]Metric),
		modelIndex:  make(map[string]SemanticModel),
	}
	svc.buildIndex()
	return svc
}

func (s *Service) buildIndex() {
	for _, m := range s.cfg.Models {
		s.modelIndex[m.Name] = m
		for _, metric := range m.Metrics {
			s.metricIndex[metric.Name] = metric
		}
	}
}

// ReloadConfig replaces the current configuration and rebuilds indexes.
func (s *Service) ReloadConfig(cfg *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
	s.buildIndex()
}

// GetMetric returns a single metric by name.
func (s *Service) GetMetric(name string) (*Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metricIndex[name]
	if !ok {
		return nil, fmt.Errorf("metric %q not found", name)
	}
	return &m, nil
}

// ListMetrics returns all metrics, optionally filtered by model name.
func (s *Service) ListMetrics(modelName string) []Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if modelName != "" {
		m, ok := s.modelIndex[modelName]
		if !ok {
			return nil
		}
		return m.Metrics
	}

	var all []Metric
	for _, m := range s.cfg.Models {
		all = append(all, m.Metrics...)
	}
	return all
}

// ListModels returns all semantic model definitions.
func (s *Service) ListModels() []SemanticModel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Models
}

// BuildSQL generates a SQL query for the given metric names.
func (s *Service) BuildSQL(metricNames []string) (*QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var metrics []Metric
	for _, name := range metricNames {
		m, ok := s.metricIndex[name]
		if !ok {
			return nil, fmt.Errorf("metric %q not found", name)
		}
		metrics = append(metrics, m)
	}

	sql := s.builder.BuildMultiMetric(metrics)
	return &QueryResult{
		SQL:         sql,
		Metrics:     metricNames,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

// BuildMetricSQL generates SQL for a single metric.
func (s *Service) BuildMetricSQL(name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metricIndex[name]
	if !ok {
		return "", fmt.Errorf("metric %q not found", name)
	}
	return s.builder.BuildSelect(m), nil
}

// BuildLLMContext returns the full LLM prompt context.
func (s *Service) BuildLLMContext() (*LLMContext, error) {
	return s.ctxBuilder.Build()
}

// BuildPromptFragment returns a prompt-ready string with schema, metrics, and joins.
func (s *Service) BuildPromptFragment() (string, error) {
	return s.ctxBuilder.PromptFragment()
}

// Config returns a copy of the current configuration.
func (s *Service) Config() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}
