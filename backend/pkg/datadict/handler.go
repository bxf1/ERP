package datadict

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts data dictionary endpoints on the given Gin router group.
func RegisterRoutes(rg *gin.RouterGroup, svc *Service) {
	h := &handler{svc: svc}

	dict := rg.Group("/datadict")
	{
		dict.GET("/schema", h.GetSchema)
		dict.POST("/schema/refresh", h.RefreshSchema)
		dict.GET("/schema/diff", h.DiffSchema)
		dict.GET("/tables", h.ListTables)
		dict.GET("/tables/:schema/:table", h.GetTable)
		dict.GET("/relations", h.GetRelations)
		dict.GET("/summary", h.SchemaSummary)
	}
}

type handler struct {
	svc *Service
}

func (h *handler) GetSchema(c *gin.Context) {
	dict, err := h.svc.GetSchema(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dict)
}

func (h *handler) RefreshSchema(c *gin.Context) {
	dict, err := h.svc.RefreshSchema(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dict)
}

func (h *handler) DiffSchema(c *gin.Context) {
	diff, err := h.svc.DiffSchema(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, diff)
}

func (h *handler) ListTables(c *gin.Context) {
	schema := c.Query("schema")
	tables, err := h.svc.ListTables(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

func (h *handler) GetTable(c *gin.Context) {
	schema := c.Param("schema")
	table := c.Param("table")
	t, err := h.svc.GetTable(c.Request.Context(), schema, table)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *handler) GetRelations(c *gin.Context) {
	rels, err := h.svc.GetRelations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"relations": rels})
}

func (h *handler) SchemaSummary(c *gin.Context) {
	summary, err := h.svc.SchemaSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.String(http.StatusOK, summary)
}
