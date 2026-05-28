package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bxf1/ERP/backend/internal/response"
	apperrors "github.com/bxf1/ERP/backend/internal/errors"
	"github.com/bxf1/ERP/backend/service"
)

type WorkflowHandler struct {
	svc *service.WorkflowService
}

func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// =========================================================================
// Definitions
// =========================================================================

func (h *WorkflowHandler) ListDefinitions(c *gin.Context) {
	status := c.Query("status")
	defs, err := h.svc.ListDefinitions(status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"definitions": defs})
}

func (h *WorkflowHandler) GetDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid definition id"))
		return
	}

	def, err := h.svc.GetDefinition(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, def)
}

func (h *WorkflowHandler) CreateDefinition(c *gin.Context) {
	var input service.CreateDefinitionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	def, err := h.svc.CreateDefinition(input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, def)
}

func (h *WorkflowHandler) UpdateDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid definition id"))
		return
	}

	var input service.UpdateDefinitionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	def, err := h.svc.UpdateDefinition(id, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, def)
}

func (h *WorkflowHandler) DeleteDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid definition id"))
		return
	}

	if err := h.svc.DeleteDefinition(id); err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// =========================================================================
// Nodes
// =========================================================================

func (h *WorkflowHandler) CreateNode(c *gin.Context) {
	defID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid definition id"))
		return
	}

	var input service.CreateNodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	node, err := h.svc.CreateNode(defID, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, node)
}

func (h *WorkflowHandler) UpdateNode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid node id"))
		return
	}

	var input service.CreateNodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	node, err := h.svc.UpdateNode(id, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, node)
}

func (h *WorkflowHandler) DeleteNode(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid node id"))
		return
	}

	if err := h.svc.DeleteNode(id); err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// =========================================================================
// Edges
// =========================================================================

func (h *WorkflowHandler) CreateEdge(c *gin.Context) {
	defID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid definition id"))
		return
	}

	var input service.CreateEdgeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	edge, err := h.svc.CreateEdge(defID, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, edge)
}

func (h *WorkflowHandler) UpdateEdge(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid edge id"))
		return
	}

	var input service.CreateEdgeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	edge, err := h.svc.UpdateEdge(id, input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, edge)
}

func (h *WorkflowHandler) DeleteEdge(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid edge id"))
		return
	}

	if err := h.svc.DeleteEdge(id); err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// =========================================================================
// Instances
// =========================================================================

func (h *WorkflowHandler) StartInstance(c *gin.Context) {
	var input service.StartInstanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	inst, err := h.svc.StartInstance(input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, inst)
}

func (h *WorkflowHandler) GetInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	inst, err := h.svc.GetInstance(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func (h *WorkflowHandler) ListInstances(c *gin.Context) {
	status := c.Query("status")
	var submittedBy *uuid.UUID
	if sb := c.Query("submitted_by"); sb != "" {
		if uid, err := uuid.Parse(sb); err == nil {
			submittedBy = &uid
		}
	}

	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	instances, total, err := h.svc.ListInstances(status, submittedBy, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paged(c, gin.H{"instances": instances}, &response.Page{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

// =========================================================================
// Approval operations
// =========================================================================

type approvalRequest struct {
	OperatorID string `json:"operator_id" binding:"required"`
	Comment    string `json:"comment"`
	TargetNodeID string `json:"target_node_id"`
}

func (h *WorkflowHandler) ApproveInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	var req approvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	opID, err := uuid.Parse(req.OperatorID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid operator_id"))
		return
	}

	inst, err := h.svc.ApproveInstance(id, service.ApprovalInput{
		OperatorID: opID,
		Comment:    req.Comment,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func (h *WorkflowHandler) RejectInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	var req approvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	opID, err := uuid.Parse(req.OperatorID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid operator_id"))
		return
	}

	inst, err := h.svc.RejectInstance(id, service.ApprovalInput{
		OperatorID: opID,
		Comment:    req.Comment,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func (h *WorkflowHandler) TransferInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	var req approvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	opID, err := uuid.Parse(req.OperatorID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid operator_id"))
		return
	}

	targetNodeID, err := uuid.Parse(req.TargetNodeID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid target_node_id"))
		return
	}

	inst, err := h.svc.TransferInstance(id, service.ApprovalInput{
		OperatorID: opID,
		Comment:    req.Comment,
	}, targetNodeID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func (h *WorkflowHandler) AddSigner(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	var req approvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	opID, err := uuid.Parse(req.OperatorID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid operator_id"))
		return
	}

	inst, err := h.svc.AddSigner(id, service.ApprovalInput{
		OperatorID: opID,
		Comment:    req.Comment,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func (h *WorkflowHandler) CancelInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid instance id"))
		return
	}

	var req struct {
		OperatorID string `json:"operator_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}

	opID, err := uuid.Parse(req.OperatorID)
	if err != nil {
		response.Error(c, apperrors.BadRequest("invalid operator_id"))
		return
	}

	inst, err := h.svc.CancelInstance(id, opID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, inst)
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
