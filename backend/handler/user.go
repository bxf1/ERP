package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/internal/response"
)

type UserHandler struct{}

func ListUsers(c *gin.Context) {
	response.OK(c, gin.H{"users": []interface{}{}})
}

func GetUser(c *gin.Context) {
	response.OK(c, gin.H{"user": nil})
}

func CreateUser(c *gin.Context) {
	response.Created(c, gin.H{"id": "placeholder"})
}

func UpdateUser(c *gin.Context) {
	response.OK(c, nil)
}

func DeleteUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}
