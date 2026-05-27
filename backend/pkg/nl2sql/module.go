package nl2sql

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bxf1/ERP/backend/pkg/nl2sql/api"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/config"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/models"
)

// Module wraps the full NL2SQL service.
type Module struct {
	handler *api.Handler
}

// NewModule creates the NL2SQL module with all components wired up.
func NewModule(db *sql.DB, llm api.LLMClient, customSemantics []models.SemanticMapping, cfg config.Config) *Module {
	h := api.NewHandler(db, llm, customSemantics, cfg)
	return &Module{handler: h}
}

// RegisterRoutes mounts the NL2SQL endpoints.
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
	m.handler.RegisterRoutes(r)
}

// HealthCheck verifies the database connection.
func (m *Module) HealthCheck() error {
	return m.handler.CheckDB()
}

// StartServer is a convenience function that starts the NL2SQL service standalone.
func StartServer(db *sql.DB, llm api.LLMClient, addr string, cfg config.Config) error {
	mod := NewModule(db, llm, nil, cfg)

	r := gin.Default()
	api := r.Group("/api/nl2sql")
	mod.RegisterRoutes(api)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("NL2SQL service starting on %s", addr)
	return r.Run(addr)
}
