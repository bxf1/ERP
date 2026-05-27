package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/pkg/response"
)

type UserHandler struct{}

func ListUsers(c *gin.Context) {
	response.Success(c, gin.H{"users": []interface{}{}})
}

func GetUser(c *gin.Context) {
	response.Success(c, gin.H{"user": nil})
}

func CreateUser(c *gin.Context) {
	response.Created(c, gin.H{"id": "placeholder"})
}

func UpdateUser(c *gin.Context) {
	response.Success(c, nil)
}

func DeleteUser(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}
