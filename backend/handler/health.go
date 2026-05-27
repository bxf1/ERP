package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/internal/response"
)

func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
