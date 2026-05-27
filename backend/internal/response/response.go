package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/bxf1/ERP/backend/internal/errors"
)

// APIResponse is the standard API response envelope.
type APIResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Page represents pagination info.
type Page struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// PagedResponse is the standard paginated response.
type PagedResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Page    *Page  `json:"page,omitempty"`
}

// OK sends a success response (HTTP 200).
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Code:    "OK",
		Message: "success",
		Data:    data,
	})
}

// Created sends a success response (HTTP 201).
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Code:    "OK",
		Message: "created",
		Data:    data,
	})
}

// NoContent sends a success response with no body.
func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// Paged sends a paginated success response.
func Paged(c *gin.Context, data any, page *Page) {
	c.JSON(http.StatusOK, PagedResponse{
		Success: true,
		Code:    "OK",
		Message: "success",
		Data:    data,
		Page:    page,
	})
}

// Error sends an error response.
// If the error is an AppError, uses its code and status; otherwise treats as internal error.
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if e, ok := err.(*apperrors.AppError); ok {
		appErr = e
	} else {
		appErr = apperrors.Internal(err.Error())
	}

	status := appErr.Code.StatusCode()
	if status < 400 {
		status = http.StatusInternalServerError
	}

	c.AbortWithStatusJSON(status, APIResponse{
		Success: false,
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Data:    appErr.Details,
	})
}
