package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/pkg/response"
	"github.com/bxf1/ERP/backend/service"
)

type RAGHandler struct {
	svc *service.RAGService
}

func NewRAGHandler(svc *service.RAGService) *RAGHandler {
	return &RAGHandler{svc: svc}
}

type ingestDocRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Source  string `json:"source"`
}

func (h *RAGHandler) IngestDocument(c *gin.Context) {
	var req ingestDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.IngestDocument(c.Request.Context(), service.IngestDocInput{
		Title:   req.Title,
		Content: req.Content,
		Source:  req.Source,
	}); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, nil)
}

type ingestQARequest struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	Category string `json:"category"`
}

func (h *RAGHandler) IngestQA(c *gin.Context) {
	var req ingestQARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.IngestQA(c.Request.Context(), service.IngestQAInput{
		Question: req.Question,
		Answer:   req.Answer,
		Category: req.Category,
	}); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, nil)
}

type searchRequest struct {
	Query string `json:"query" binding:"required"`
}

func (h *RAGHandler) Search(c *gin.Context) {
	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Search(c.Request.Context(), req.Query)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"documents": result.Docs,
		"qas":       result.QAs,
		"context":   h.svc.BuildContext(result),
	})
}

func (h *RAGHandler) Stats(c *gin.Context) {
	docCount, _ := h.svc.GetDocCount()
	qaCount, _ := h.svc.GetQACount()

	response.Success(c, gin.H{
		"document_count": docCount,
		"qa_count":       qaCount,
	})
}

func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	var docID uint
	if _, err := fmt.Sscanf(id, "%d", &docID); err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.DeleteDocument(docID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *RAGHandler) DeleteQA(c *gin.Context) {
	id := c.Param("id")
	var qaID uint
	if _, err := fmt.Sscanf(id, "%d", &qaID); err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.DeleteQA(qaID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
