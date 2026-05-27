package semantic

import (
	"context"
	"database/sql"
	"time"

	"github.com/bxf1/ERP/backend/pkg/datadict"
	"github.com/gin-gonic/gin"
)

// Module bundles the data dictionary and semantic layer together as a single
// installable unit. Call NewModule during app startup, then mount routes
// with Module.RegisterRoutes.
type Module struct {
	DataDict *datadict.Service
	Semantic *Service
}

// ModuleConfig holds initialization parameters.
type ModuleConfig struct {
	DB               *sql.DB
	SemanticCfgPath  string
	CacheTTL         time.Duration
	AutoRefresh      time.Duration // 0 disables periodic background refresh
}

// NewModule creates both services and wires them together.
// Returns an error if the semantic config file can't be loaded.
func NewModule(cfg ModuleConfig) (*Module, error) {
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 5 * time.Minute
	}

	dictSvc := datadict.NewService(cfg.DB, cfg.CacheTTL)

	semCfg, err := LoadConfig(cfg.SemanticCfgPath)
	if err != nil {
		return nil, err
	}
	semSvc := NewService(dictSvc, semCfg)

	if cfg.AutoRefresh > 0 {
		dictSvc.StartAutoRefresh(context.Background(), cfg.AutoRefresh)
	}

	return &Module{DataDict: dictSvc, Semantic: semSvc}, nil
}

// RegisterRoutes mounts all data dictionary and semantic layer API routes
// under the given Gin router group (typically /api/v1).
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	datadict.RegisterRoutes(rg, m.DataDict)
	RegisterRoutes(rg, m.Semantic)
}
