package builder

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the Builder Agent API endpoints on a Gin router group.
func (a *Agent) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/sessions", a.handleStartSession)
	rg.GET("/sessions/:id", a.handleGetSession)
	rg.POST("/sessions/:id/messages", a.handleProcessMessage)
	rg.GET("/models", a.handleListModels)
	rg.GET("/tools", a.handleListTools)
	rg.POST("/tools/execute", a.handleExecuteTool)
}

func (a *Agent) handleStartSession(c *gin.Context) {
	session := a.StartSession()
	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.SessionID,
		"state":      session.State,
		"created_at": session.CreatedAt,
	})
}

func (a *Agent) handleGetSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, err := a.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (a *Agent) handleProcessMessage(c *gin.Context) {
	sessionID := c.Param("id")

	var req BuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.SessionID = sessionID

	resp, err := a.ProcessMessage(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *Agent) handleListModels(c *gin.Context) {
	models := a.GetExistingModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (a *Agent) handleListTools(c *gin.Context) {
	tools := a.GetTools()
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

func (a *Agent) handleExecuteTool(c *gin.Context) {
	var call MCPToolCall
	if err := c.ShouldBindJSON(&call); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := a.ExecuteTool(c.Request.Context(), call)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var parsed interface{}
	json.Unmarshal(result, &parsed)
	c.JSON(http.StatusOK, gin.H{"result": parsed})
}
