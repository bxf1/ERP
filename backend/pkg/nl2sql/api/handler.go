package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bxf1/ERP/backend/pkg/nl2sql/cache"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/config"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/executor"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/models"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/prompt"
	"github.com/bxf1/ERP/backend/pkg/nl2sql/security"
)

// Handler holds all dependencies for the NL2SQL HTTP API.
type Handler struct {
	db        *sql.DB
	llm       LLMClient
	dict      *prompt.Dictionary
	semantics []models.SemanticMapping
	cache     *cache.QueryCache
	executor  *executor.Executor
	cfg       config.Config
}

// LLMClient is the interface for calling an LLM to generate SQL.
type LLMClient interface {
	GenerateSQL(systemPrompt, userPrompt string) (string, error)
}

// NewHandler creates the API handler with all dependencies.
func NewHandler(db *sql.DB, llm LLMClient, customSemantics []models.SemanticMapping, cfg config.Config) *Handler {
	dict := prompt.NewDictionary(db)
	execCfg := executor.Config{
		MaxRetries:      cfg.MaxRetries,
		BaseBackoff:     cfg.BaseBackoff,
		MaxBackoff:      cfg.MaxBackoff,
		QueryTimeout:    cfg.QueryTimeout,
		DegradeLimitRow: 100,
	}

	return &Handler{
		db:        db,
		llm:       llm,
		dict:      dict,
		semantics: prompt.MergeSemantics(prompt.DefaultSemantics(), customSemantics),
		cache:     cache.New(cfg.CacheTTL, cfg.CacheMaxSize),
		executor:  executor.New(db, execCfg),
		cfg:       cfg,
	}
}

// RegisterRoutes wires up the NL2SQL endpoints on a Gin router group.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/query", h.handleQuery)
	r.GET("/schema", h.handleSchema)
	r.GET("/cache/stats", h.handleCacheStats)
}

// handleQuery is the main NL2SQL endpoint.
func (h *Handler) handleQuery(c *gin.Context) {
	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MaxRows <= 0 || req.MaxRows > 1000 {
		req.MaxRows = 1000
	}

	// Load fresh schema metadata.
	tables, err := h.dict.LoadTables(h.cfg.DBSchema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schema: " + err.Error()})
		return
	}

	rels, err := h.dict.LoadRelationships(h.cfg.DBSchema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load relationships: " + err.Error()})
		return
	}

	// Build the LLM prompt.
	tmpl := prompt.NewTemplate(tables, rels, h.semantics, req.TenantID)
	systemPrompt := tmpl.BuildSystemPrompt()
	userPrompt := prompt.BuildUserPrompt(req.Question)

	// Call LLM to generate SQL.
	rawSQL, err := h.llm.GenerateSQL(systemPrompt, userPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM generation failed: " + err.Error()})
		return
	}

	// Security validation.
	safeSQL, err := security.ValidateSQL(rawSQL, req.TenantID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":         "SQL validation failed",
			"detail":        err.Error(),
			"generated_sql": rawSQL,
		})
		return
	}

	resp := models.QueryResponse{
		SQL:   safeSQL,
		Explanation: "Query generated from: " + truncate(req.Question, 200),
	}

	// Check cache.
	cacheKey := cache.CacheKey(safeSQL, req.TenantID)
	if req.UseCache {
		if entry := h.cache.Get(cacheKey); entry != nil {
			resp.Columns = entry.Columns
			resp.Rows = entry.Rows
			resp.RowCount = len(entry.Rows)
			resp.FromCache = true
			resp.Duration = 0
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// Execute the validated SQL.
	start := time.Now()
	result, err := h.executor.Execute(c.Request.Context(), safeSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "query execution failed",
			"detail": err.Error(),
			"sql":    safeSQL,
		})
		return
	}

	resp.Columns = result.Columns
	resp.Rows = result.Rows
	resp.RowCount = len(result.Rows)
	resp.Duration = time.Since(start)

	// Store in cache.
	if req.UseCache {
		h.cache.Set(cacheKey, result.Columns, result.Rows)
	}

	c.JSON(http.StatusOK, resp)
}

// handleSchema returns the current data dictionary and semantics.
func (h *Handler) handleSchema(c *gin.Context) {
	tables, err := h.dict.LoadTables(h.cfg.DBSchema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rels, err := h.dict.LoadRelationships(h.cfg.DBSchema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SchemaResponse{
		Tables:        tables,
		Relationships: rels,
		Semantics:     h.semantics,
	})
}

// handleCacheStats returns cache metrics.
func (h *Handler) handleCacheStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"size":     h.cache.Size(),
		"max_size": h.cfg.CacheMaxSize,
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// CheckDB verifies the database connection is alive.
func (h *Handler) CheckDB() error {
	return h.db.Ping()
}
