package semantic

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts semantic-layer endpoints on the given Gin router group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	h := &handler{svc: svc}

	sem := rg.Group("/semantic")
	{
		sem.GET("/models", h.ListModels)
		sem.GET("/metrics", h.ListMetrics)
		sem.GET("/metrics/:name", h.GetMetric)
		sem.GET("/metrics/:name/sql", h.GetMetricSQL)
		sem.POST("/query", h.BuildQuery)
		sem.GET("/context", h.GetLLMContext)
		sem.GET("/prompt", h.GetPromptFragment)
	}
}

type handler struct {
	svc *Service
}

func (h *handler) ListModels(c *gin.Context) {
	models := h.svc.ListModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (h *handler) ListMetrics(c *gin.Context) {
	modelName := c.Query("model")
	metrics := h.svc.ListMetrics(modelName)
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

func (h *handler) GetMetric(c *gin.Context) {
	name := c.Param("name")
	m, err := h.svc.GetMetric(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *handler) GetMetricSQL(c *gin.Context) {
	name := c.Param("name")
	sql, err := h.svc.BuildMetricSQL(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"metric": name, "sql": sql})
}

func (h *handler) BuildQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Metrics) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one metric name is required"})
		return
	}
	result, err := h.svc.BuildSQL(req.Metrics)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *handler) GetLLMContext(c *gin.Context) {
	ctx, err := h.svc.BuildLLMContext()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ctx)
}

func (h *handler) GetPromptFragment(c *gin.Context) {
	fragment, err := h.svc.BuildPromptFragment()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, fragment)
}
