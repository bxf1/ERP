package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	wf "github.com/bxf1/ERP/backend/pkg/workflow/models"
	"github.com/bxf1/ERP/backend/repository"
	apperrors "github.com/bxf1/ERP/backend/internal/errors"
)

type WorkflowService struct {
	db   *gorm.DB
	repo *repository.WorkflowRepository
}

func NewWorkflowService(db *gorm.DB, repo *repository.WorkflowRepository) *WorkflowService {
	return &WorkflowService{db: db, repo: repo}
}

// =========================================================================
// Definition DTOs
// =========================================================================

type CreateDefinitionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FormConfig  string `json:"form_config"`
}

type UpdateDefinitionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FormConfig  string `json:"form_config"`
	Status      string `json:"status"`
}

type CreateNodeInput struct {
	Name         string  `json:"name"`
	NodeType     string  `json:"node_type"`
	ApproverRule string  `json:"approver_rule"`
	FormViewID   *uuid.UUID `json:"form_view_id"`
	PositionX    float64 `json:"position_x"`
	PositionY    float64 `json:"position_y"`
	Config       string  `json:"config"`
}

type CreateEdgeInput struct {
	SourceNodeID  uuid.UUID `json:"source_node_id"`
	TargetNodeID  uuid.UUID `json:"target_node_id"`
	ConditionExpr string    `json:"condition_expr"`
	Label         string    `json:"label"`
}

// =========================================================================
// Definition management
// =========================================================================

func (s *WorkflowService) ListDefinitions(status string) ([]wf.WorkflowDefinition, error) {
	return s.repo.ListDefinitions(status)
}

func (s *WorkflowService) GetDefinition(id uuid.UUID) (*wf.WorkflowDefinition, error) {
	def, err := s.repo.GetDefinition(id)
	if err != nil {
		return nil, apperrors.NotFound("workflow definition not found")
	}
	return def, nil
}

func (s *WorkflowService) CreateDefinition(input CreateDefinitionInput) (*wf.WorkflowDefinition, error) {
	if input.Name == "" {
		return nil, apperrors.BadRequest("name is required")
	}

	def := &wf.WorkflowDefinition{
		Name:        input.Name,
		Description: input.Description,
		Status:      "draft",
		Version:     1,
		FormConfig:  input.FormConfig,
	}

	if err := s.repo.CreateDefinition(def); err != nil {
		return nil, apperrors.DatabaseError("create definition failed", err)
	}

	return def, nil
}

func (s *WorkflowService) UpdateDefinition(id uuid.UUID, input UpdateDefinitionInput) (*wf.WorkflowDefinition, error) {
	def, err := s.repo.GetDefinition(id)
	if err != nil {
		return nil, apperrors.NotFound("workflow definition not found")
	}

	if input.Name != "" {
		def.Name = input.Name
	}
	if input.Description != "" {
		def.Description = input.Description
	}
	if input.FormConfig != "" {
		def.FormConfig = input.FormConfig
	}
	if input.Status != "" {
		def.Status = input.Status
	}
	def.UpdatedAt = time.Now()

	if err := s.repo.UpdateDefinition(def); err != nil {
		return nil, apperrors.DatabaseError("update definition failed", err)
	}

	return def, nil
}

func (s *WorkflowService) DeleteDefinition(id uuid.UUID) error {
	_, err := s.repo.GetDefinition(id)
	if err != nil {
		return apperrors.NotFound("workflow definition not found")
	}
	return s.repo.DeleteDefinition(id)
}

// =========================================================================
// Node management
// =========================================================================

func (s *WorkflowService) CreateNode(defID uuid.UUID, input CreateNodeInput) (*wf.WorkflowNode, error) {
	_, err := s.repo.GetDefinition(defID)
	if err != nil {
		return nil, apperrors.NotFound("workflow definition not found")
	}

	if !isValidNodeType(input.NodeType) {
		return nil, apperrors.BadRequest("invalid node type: " + input.NodeType)
	}

	node := &wf.WorkflowNode{
		DefinitionID: defID,
		Name:         input.Name,
		NodeType:     input.NodeType,
		ApproverRule: input.ApproverRule,
		FormViewID:   input.FormViewID,
		PositionX:    input.PositionX,
		PositionY:    input.PositionY,
		Config:       input.Config,
	}

	if err := s.repo.CreateNode(node); err != nil {
		return nil, apperrors.DatabaseError("create node failed", err)
	}

	return node, nil
}

func (s *WorkflowService) UpdateNode(id uuid.UUID, input CreateNodeInput) (*wf.WorkflowNode, error) {
	node, err := s.repo.GetNode(id)
	if err != nil {
		return nil, apperrors.NotFound("workflow node not found")
	}

	node.Name = input.Name
	node.NodeType = input.NodeType
	node.ApproverRule = input.ApproverRule
	node.FormViewID = input.FormViewID
	node.PositionX = input.PositionX
	node.PositionY = input.PositionY
	node.Config = input.Config
	node.UpdatedAt = time.Now()

	if err := s.repo.UpdateNode(node); err != nil {
		return nil, apperrors.DatabaseError("update node failed", err)
	}

	return node, nil
}

func (s *WorkflowService) DeleteNode(id uuid.UUID) error {
	_, err := s.repo.GetNode(id)
	if err != nil {
		return apperrors.NotFound("workflow node not found")
	}
	return s.repo.DeleteNode(id)
}

// =========================================================================
// Edge management
// =========================================================================

func (s *WorkflowService) CreateEdge(defID uuid.UUID, input CreateEdgeInput) (*wf.WorkflowEdge, error) {
	_, err := s.repo.GetDefinition(defID)
	if err != nil {
		return nil, apperrors.NotFound("workflow definition not found")
	}

	edge := &wf.WorkflowEdge{
		DefinitionID:  defID,
		SourceNodeID:  input.SourceNodeID,
		TargetNodeID:  input.TargetNodeID,
		ConditionExpr: input.ConditionExpr,
		Label:         input.Label,
	}

	if err := s.repo.CreateEdge(edge); err != nil {
		return nil, apperrors.DatabaseError("create edge failed", err)
	}

	return edge, nil
}

func (s *WorkflowService) UpdateEdge(id uuid.UUID, input CreateEdgeInput) (*wf.WorkflowEdge, error) {
	edge, err := s.repo.GetEdge(id)
	if err != nil {
		return nil, apperrors.NotFound("workflow edge not found")
	}

	edge.SourceNodeID = input.SourceNodeID
	edge.TargetNodeID = input.TargetNodeID
	edge.ConditionExpr = input.ConditionExpr
	edge.Label = input.Label

	if err := s.repo.UpdateEdge(edge); err != nil {
		return nil, apperrors.DatabaseError("update edge failed", err)
	}

	return edge, nil
}

func (s *WorkflowService) DeleteEdge(id uuid.UUID) error {
	_, err := s.repo.GetEdge(id)
	if err != nil {
		return apperrors.NotFound("workflow edge not found")
	}
	return s.repo.DeleteEdge(id)
}

// =========================================================================
// Instance management
// =========================================================================

type StartInstanceInput struct {
	DefinitionID uuid.UUID `json:"definition_id"`
	FormData     string    `json:"form_data"`
	SubmittedBy  uuid.UUID `json:"submitted_by"`
}

func (s *WorkflowService) StartInstance(input StartInstanceInput) (*wf.WorkflowInstance, error) {
	def, err := s.repo.GetDefinition(input.DefinitionID)
	if err != nil {
		return nil, apperrors.NotFound("workflow definition not found")
	}

	if def.Status != "published" {
		return nil, apperrors.BadRequest("workflow definition is not published")
	}

	startNode, err := s.repo.GetStartNode(input.DefinitionID)
	if err != nil {
		return nil, apperrors.BadRequest("workflow has no start node")
	}

	// Find first approval node after start
	firstApproval, err := s.findNextApprovalNode(startNode.ID)
	if err != nil {
		return nil, apperrors.BadRequest("no approval node found after start")
	}

	now := time.Now()
	inst := &wf.WorkflowInstance{
		DefinitionID:  input.DefinitionID,
		CurrentNodeID: &firstApproval.ID,
		Status:        wf.InstanceStatusRunning,
		FormData:      input.FormData,
		SubmittedBy:   &input.SubmittedBy,
		SubmittedAt:   now,
	}

	if err := s.repo.CreateInstance(inst); err != nil {
		return nil, apperrors.DatabaseError("create instance failed", err)
	}

	// Log history
	s.repo.CreateHistory(&wf.InstanceHistory{
		InstanceID: inst.ID,
		NodeID:     &startNode.ID,
		Action:     wf.ApprovalActionSubmit,
		OperatorID: &input.SubmittedBy,
		ToNodeID:   &firstApproval.ID,
	})

	return inst, nil
}

func (s *WorkflowService) GetInstance(id uuid.UUID) (*wf.WorkflowInstance, error) {
	inst, err := s.repo.GetInstance(id)
	if err != nil {
		return nil, apperrors.NotFound("workflow instance not found")
	}
	return inst, nil
}

func (s *WorkflowService) ListInstances(status string, submittedBy *uuid.UUID, page, pageSize int) ([]wf.WorkflowInstance, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListInstances(status, submittedBy, pageSize, offset)
}

// =========================================================================
// Approval operations
// =========================================================================

type ApprovalInput struct {
	OperatorID uuid.UUID `json:"operator_id"`
	Comment    string    `json:"comment"`
}

func (s *WorkflowService) ApproveInstance(instanceID uuid.UUID, input ApprovalInput) (*wf.WorkflowInstance, error) {
	return s.executeApprovalAction(instanceID, input, wf.ApprovalActionApprove, func(inst *wf.WorkflowInstance) error {
		if inst.CurrentNodeID == nil {
			return apperrors.BadRequest("instance has no current node")
		}

		nextNode, err := s.findNextApprovalNode(*inst.CurrentNodeID)
		if err != nil {
			// No more approval nodes; check if the next edge leads to an end node
			endNode, endErr := s.findEndNode(*inst.CurrentNodeID)
			if endErr != nil {
				return apperrors.BadRequest("no next node found in workflow")
			}
			now := time.Now()
			inst.Status = wf.InstanceStatusApproved
			inst.CurrentNodeID = &endNode.ID
			inst.CompletedAt = &now
			return nil
		}

		inst.CurrentNodeID = &nextNode.ID
		return nil
	})
}

func (s *WorkflowService) RejectInstance(instanceID uuid.UUID, input ApprovalInput) (*wf.WorkflowInstance, error) {
	return s.executeApprovalAction(instanceID, input, wf.ApprovalActionReject, func(inst *wf.WorkflowInstance) error {
		now := time.Now()
		inst.Status = wf.InstanceStatusRejected
		inst.CompletedAt = &now
		return nil
	})
}

func (s *WorkflowService) TransferInstance(instanceID uuid.UUID, input ApprovalInput, targetNodeID uuid.UUID) (*wf.WorkflowInstance, error) {
	return s.executeApprovalAction(instanceID, input, wf.ApprovalActionTransfer, func(inst *wf.WorkflowInstance) error {
		// Verify target node exists in same definition
		targetNode, err := s.repo.GetNode(targetNodeID)
		if err != nil || targetNode.DefinitionID != inst.DefinitionID {
			return apperrors.BadRequest("target node not found in this workflow")
		}

		inst.CurrentNodeID = &targetNode.ID
		return nil
	})
}

func (s *WorkflowService) AddSigner(instanceID uuid.UUID, input ApprovalInput) (*wf.WorkflowInstance, error) {
	return s.executeApprovalAction(instanceID, input, wf.ApprovalActionAddSigner, func(inst *wf.WorkflowInstance) error {
		// Add-signer just records the action; instance stays at current node
		return nil
	})
}

// =========================================================================
// Cancellation
// =========================================================================

func (s *WorkflowService) CancelInstance(instanceID uuid.UUID, operatorID uuid.UUID) (*wf.WorkflowInstance, error) {
	inst, err := s.repo.GetInstance(instanceID)
	if err != nil {
		return nil, apperrors.NotFound("workflow instance not found")
	}

	if inst.Status != wf.InstanceStatusRunning {
		return nil, apperrors.BadRequest("only running instances can be cancelled")
	}

	now := time.Now()
	inst.Status = wf.InstanceStatusCancelled
	inst.CompletedAt = &now

	if err := s.repo.UpdateInstance(inst); err != nil {
		return nil, apperrors.DatabaseError("cancel instance failed", err)
	}

	s.repo.CreateHistory(&wf.InstanceHistory{
		InstanceID: inst.ID,
		NodeID:     inst.CurrentNodeID,
		Action:     "cancel",
		OperatorID: &operatorID,
	})

	return inst, nil
}

// =========================================================================
// Internal helpers
// =========================================================================

func (s *WorkflowService) executeApprovalAction(
	instanceID uuid.UUID,
	input ApprovalInput,
	action string,
	mutate func(*wf.WorkflowInstance) error,
) (*wf.WorkflowInstance, error) {
	inst, err := s.repo.GetInstance(instanceID)
	if err != nil {
		return nil, apperrors.NotFound("workflow instance not found")
	}

	if inst.Status != wf.InstanceStatusRunning {
		return nil, apperrors.BadRequest("instance is not in running status")
	}

	prevNodeID := inst.CurrentNodeID

	// Record approval
	s.repo.CreateApproval(&wf.WorkflowApproval{
		InstanceID: inst.ID,
		NodeID:     *inst.CurrentNodeID,
		ApproverID: &input.OperatorID,
		Action:     action,
		Comment:    input.Comment,
	})

	// Apply state mutation
	if err := mutate(inst); err != nil {
		return nil, err
	}

	inst.UpdatedAt = time.Now()
	if err := s.repo.UpdateInstance(inst); err != nil {
		return nil, apperrors.DatabaseError("update instance failed", err)
	}

	// Log history
	s.repo.CreateHistory(&wf.InstanceHistory{
		InstanceID: inst.ID,
		NodeID:     prevNodeID,
		Action:     action,
		OperatorID: &input.OperatorID,
		Comment:    input.Comment,
		ToNodeID:   inst.CurrentNodeID,
	})

	return inst, nil
}

// findNextApprovalNode follows edges from nodeID to find the next approval or end node.
func (s *WorkflowService) findNextApprovalNode(nodeID uuid.UUID) (*wf.WorkflowNode, error) {
	edges, err := s.repo.GetEdgesFromNode(nodeID)
	if err != nil || len(edges) == 0 {
		return nil, fmt.Errorf("no outgoing edges from node %s", nodeID)
	}

	// Follow the first edge (simple linear flow; condition-based routing can be added later)
	targetID := edges[0].TargetNodeID
	target, err := s.repo.GetNode(targetID)
	if err != nil {
		return nil, err
	}

	switch target.NodeType {
	case wf.NodeTypeApproval:
		return target, nil
	case wf.NodeTypeCondition:
		// Recurse through condition nodes
		return s.findNextApprovalNode(target.ID)
	case wf.NodeTypeEnd:
		return nil, fmt.Errorf("reached end node")
	default:
		return nil, fmt.Errorf("unexpected node type: %s", target.NodeType)
	}
}

// findEndNode checks if the given node directly connects to an end node.
func (s *WorkflowService) findEndNode(nodeID uuid.UUID) (*wf.WorkflowNode, error) {
	edges, err := s.repo.GetEdgesFromNode(nodeID)
	if err != nil || len(edges) == 0 {
		return nil, fmt.Errorf("no outgoing edges")
	}

	for _, edge := range edges {
		target, err := s.repo.GetNode(edge.TargetNodeID)
		if err != nil {
			continue
		}
		if target.NodeType == wf.NodeTypeEnd {
			return target, nil
		}
		if target.NodeType == wf.NodeTypeCondition {
			if endNode, err := s.findEndNode(target.ID); err == nil {
				return endNode, nil
			}
		}
	}

	return nil, fmt.Errorf("no end node reachable")
}

func isValidNodeType(t string) bool {
	switch t {
	case wf.NodeTypeStart, wf.NodeTypeApproval, wf.NodeTypeEnd, wf.NodeTypeCondition:
		return true
	}
	return false
}

// PaginatedResult wraps list data with pagination info.
type PaginatedResult struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func (r *PaginatedResult) UnmarshalItems(v interface{}) error {
	data, err := json.Marshal(r.Items)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
