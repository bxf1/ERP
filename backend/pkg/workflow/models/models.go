package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =========================================================================
// Workflow Definition (工作流定义)
// =========================================================================

type WorkflowDefinition struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Name        string     `gorm:"size:200;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Version     int        `gorm:"not null;default:1" json:"version"`
	Status      string     `gorm:"size:20;not null;default:draft" json:"status"`
	FormConfig  string     `gorm:"type:jsonb" json:"form_config"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Nodes []WorkflowNode `gorm:"foreignKey:DefinitionID" json:"nodes,omitempty"`
	Edges []WorkflowEdge `gorm:"foreignKey:DefinitionID" json:"edges,omitempty"`
}

func (w *WorkflowDefinition) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Workflow Node (工作流节点)
// =========================================================================

type WorkflowNode struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	DefinitionID uuid.UUID  `gorm:"type:uuid;not null;index" json:"definition_id"`
	Name         string     `gorm:"size:200;not null" json:"name"`
	NodeType     string     `gorm:"size:50;not null" json:"node_type"`
	ApproverRule string     `gorm:"type:jsonb" json:"approver_rule"`
	FormViewID   *uuid.UUID `gorm:"type:uuid" json:"form_view_id"`
	PositionX    float64    `gorm:"default:0" json:"position_x"`
	PositionY    float64    `gorm:"default:0" json:"position_y"`
	Config       string     `gorm:"type:jsonb" json:"config"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (n *WorkflowNode) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Workflow Edge (工作流边)
// =========================================================================

type WorkflowEdge struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	DefinitionID   uuid.UUID `gorm:"type:uuid;not null;index" json:"definition_id"`
	SourceNodeID   uuid.UUID `gorm:"type:uuid;not null" json:"source_node_id"`
	TargetNodeID   uuid.UUID `gorm:"type:uuid;not null" json:"target_node_id"`
	ConditionExpr  string    `gorm:"type:text" json:"condition_expr"`
	Label          string    `gorm:"size:200" json:"label"`
	CreatedAt      time.Time `json:"created_at"`
}

func (e *WorkflowEdge) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Workflow Instance (流程实例)
// =========================================================================

type WorkflowInstance struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	DefinitionID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"definition_id"`
	CurrentNodeID *uuid.UUID `gorm:"type:uuid;index" json:"current_node_id"`
	Status        string     `gorm:"size:20;not null;default:running" json:"status"`
	FormData      string     `gorm:"type:jsonb" json:"form_data"`
	SubmittedBy   *uuid.UUID `gorm:"type:uuid;index" json:"submitted_by"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Definition  *WorkflowDefinition `gorm:"foreignKey:DefinitionID" json:"definition,omitempty"`
	CurrentNode *WorkflowNode       `gorm:"foreignKey:CurrentNodeID" json:"current_node,omitempty"`
	Approvals   []WorkflowApproval  `gorm:"foreignKey:InstanceID" json:"approvals,omitempty"`
	History     []InstanceHistory   `gorm:"foreignKey:InstanceID" json:"history,omitempty"`
}

func (i *WorkflowInstance) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Workflow Approval (审批记录)
// =========================================================================

type WorkflowApproval struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	InstanceID uuid.UUID  `gorm:"type:uuid;not null;index" json:"instance_id"`
	NodeID     uuid.UUID  `gorm:"type:uuid;not null" json:"node_id"`
	ApproverID *uuid.UUID `gorm:"type:uuid" json:"approver_id"`
	Action     string     `gorm:"size:20;not null" json:"action"`
	Comment    string     `gorm:"type:text" json:"comment"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (a *WorkflowApproval) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Instance History (流程历史)
// =========================================================================

type InstanceHistory struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	InstanceID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"instance_id"`
	NodeID      *uuid.UUID `gorm:"type:uuid" json:"node_id"`
	Action      string     `gorm:"size:50;not null" json:"action"`
	OperatorID  *uuid.UUID `gorm:"type:uuid" json:"operator_id"`
	Comment     string     `gorm:"type:text" json:"comment"`
	FromNodeID  *uuid.UUID `gorm:"type:uuid" json:"from_node_id"`
	ToNodeID    *uuid.UUID `gorm:"type:uuid" json:"to_node_id"`
	Details     string     `gorm:"type:jsonb" json:"details"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (h *InstanceHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

// =========================================================================
// Node type constants
// =========================================================================

const (
	NodeTypeStart     = "start"
	NodeTypeApproval  = "approval"
	NodeTypeEnd       = "end"
	NodeTypeCondition = "condition"
)

// Instance status constants
const (
	InstanceStatusRunning   = "running"
	InstanceStatusApproved  = "approved"
	InstanceStatusRejected  = "rejected"
	InstanceStatusCancelled = "cancelled"
)

// Approval action constants
const (
	ApprovalActionSubmit    = "submit"
	ApprovalActionApprove   = "approve"
	ApprovalActionReject    = "reject"
	ApprovalActionTransfer  = "transfer"
	ApprovalActionAddSigner = "add_signer"
)
