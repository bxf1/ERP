package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	wf "github.com/bxf1/ERP/backend/pkg/workflow/models"
)

type WorkflowRepository struct {
	*BaseRepository
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{BaseRepository: NewBaseRepository(db)}
}

// =========================================================================
// Definitions
// =========================================================================

func (r *WorkflowRepository) ListDefinitions(status string) ([]wf.WorkflowDefinition, error) {
	var defs []wf.WorkflowDefinition
	q := r.DB().Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&defs).Error
	return defs, err
}

func (r *WorkflowRepository) GetDefinition(id uuid.UUID) (*wf.WorkflowDefinition, error) {
	var def wf.WorkflowDefinition
	err := r.DB().Preload("Nodes").Preload("Edges").First(&def, "id = ?", id).Error
	return &def, err
}

func (r *WorkflowRepository) CreateDefinition(def *wf.WorkflowDefinition) error {
	return r.DB().Create(def).Error
}

func (r *WorkflowRepository) UpdateDefinition(def *wf.WorkflowDefinition) error {
	return r.DB().Save(def).Error
}

func (r *WorkflowRepository) DeleteDefinition(id uuid.UUID) error {
	return r.DB().Delete(&wf.WorkflowDefinition{}, "id = ?", id).Error
}

// =========================================================================
// Nodes
// =========================================================================

func (r *WorkflowRepository) GetNode(id uuid.UUID) (*wf.WorkflowNode, error) {
	var node wf.WorkflowNode
	err := r.DB().First(&node, "id = ?", id).Error
	return &node, err
}

func (r *WorkflowRepository) CreateNode(node *wf.WorkflowNode) error {
	return r.DB().Create(node).Error
}

func (r *WorkflowRepository) UpdateNode(node *wf.WorkflowNode) error {
	return r.DB().Save(node).Error
}

func (r *WorkflowRepository) DeleteNode(id uuid.UUID) error {
	return r.DB().Delete(&wf.WorkflowNode{}, "id = ?", id).Error
}

func (r *WorkflowRepository) ListNodesByDefinition(defID uuid.UUID) ([]wf.WorkflowNode, error) {
	var nodes []wf.WorkflowNode
	err := r.DB().Where("definition_id = ?", defID).Order("created_at ASC").Find(&nodes).Error
	return nodes, err
}

// =========================================================================
// Edges
// =========================================================================

func (r *WorkflowRepository) GetEdge(id uuid.UUID) (*wf.WorkflowEdge, error) {
	var edge wf.WorkflowEdge
	err := r.DB().First(&edge, "id = ?", id).Error
	return &edge, err
}

func (r *WorkflowRepository) CreateEdge(edge *wf.WorkflowEdge) error {
	return r.DB().Create(edge).Error
}

func (r *WorkflowRepository) UpdateEdge(edge *wf.WorkflowEdge) error {
	return r.DB().Save(edge).Error
}

func (r *WorkflowRepository) DeleteEdge(id uuid.UUID) error {
	return r.DB().Delete(&wf.WorkflowEdge{}, "id = ?", id).Error
}

func (r *WorkflowRepository) ListEdgesByDefinition(defID uuid.UUID) ([]wf.WorkflowEdge, error) {
	var edges []wf.WorkflowEdge
	err := r.DB().Where("definition_id = ?", defID).Order("created_at ASC").Find(&edges).Error
	return edges, err
}

// =========================================================================
// Instances
// =========================================================================

func (r *WorkflowRepository) ListInstances(status string, submittedBy *uuid.UUID, limit, offset int) ([]wf.WorkflowInstance, int64, error) {
	var instances []wf.WorkflowInstance
	var total int64

	q := r.DB().Model(&wf.WorkflowInstance{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if submittedBy != nil {
		q = q.Where("submitted_by = ?", *submittedBy)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Preload("Definition").Preload("CurrentNode").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&instances).Error
	return instances, total, err
}

func (r *WorkflowRepository) GetInstance(id uuid.UUID) (*wf.WorkflowInstance, error) {
	var inst wf.WorkflowInstance
	err := r.DB().Preload("Definition").Preload("CurrentNode").
		Preload("Approvals").Preload("History").
		First(&inst, "id = ?", id).Error
	return &inst, err
}

func (r *WorkflowRepository) CreateInstance(inst *wf.WorkflowInstance) error {
	return r.DB().Create(inst).Error
}

func (r *WorkflowRepository) UpdateInstance(inst *wf.WorkflowInstance) error {
	return r.DB().Save(inst).Error
}

// =========================================================================
// Approvals
// =========================================================================

func (r *WorkflowRepository) CreateApproval(approval *wf.WorkflowApproval) error {
	return r.DB().Create(approval).Error
}

func (r *WorkflowRepository) ListApprovalsByInstance(instID uuid.UUID) ([]wf.WorkflowApproval, error) {
	var approvals []wf.WorkflowApproval
	err := r.DB().Where("instance_id = ?", instID).Order("created_at ASC").Find(&approvals).Error
	return approvals, err
}

// =========================================================================
// History
// =========================================================================

func (r *WorkflowRepository) CreateHistory(h *wf.InstanceHistory) error {
	return r.DB().Create(h).Error
}

func (r *WorkflowRepository) ListHistoryByInstance(instID uuid.UUID) ([]wf.InstanceHistory, error) {
	var history []wf.InstanceHistory
	err := r.DB().Where("instance_id = ?", instID).Order("created_at DESC").Find(&history).Error
	return history, err
}

// =========================================================================
// Graph traversal helpers
// =========================================================================

func (r *WorkflowRepository) GetEdgesFromNode(nodeID uuid.UUID) ([]wf.WorkflowEdge, error) {
	var edges []wf.WorkflowEdge
	err := r.DB().Where("source_node_id = ?", nodeID).Find(&edges).Error
	return edges, err
}

func (r *WorkflowRepository) GetStartNode(defID uuid.UUID) (*wf.WorkflowNode, error) {
	var node wf.WorkflowNode
	err := r.DB().Where("definition_id = ? AND node_type = ?", defID, wf.NodeTypeStart).
		First(&node).Error
	return &node, err
}
