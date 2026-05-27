package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Code        string         `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"size:20;not null;default:active" json:"status"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	DataScopes  []DataScope  `gorm:"foreignKey:RoleID" json:"data_scopes,omitempty"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type Permission struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Code        string     `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Resource    string     `gorm:"size:50;not null;index" json:"resource"`
	Action      string     `gorm:"size:50;not null" json:"action"`
	Description string     `gorm:"type:text" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Children []*Permission `gorm:"-" json:"children,omitempty"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserRole struct {
	UserID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"role_id"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`
	CreatedAt     time.Time  `json:"created_at"`

	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

type DataScope struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	RoleID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_role_model" json:"role_id"`
	TargetModel string    `gorm:"size:100;not null;uniqueIndex:idx_role_model" json:"target_model"`
	ScopeType   string    `gorm:"size:20;not null" json:"scope_type"`
	ScopeRule   string    `gorm:"type:jsonb" json:"scope_rule"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (d *DataScope) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type PermissionAuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Action      string    `gorm:"size:100;not null" json:"action"`
	Resource    string    `gorm:"size:100;not null" json:"resource"`
	TargetID    string    `gorm:"size:100" json:"target_id"`
	Result      string    `gorm:"size:20;not null" json:"result"`
	Reason      string    `gorm:"type:text" json:"reason"`
	RequestPath string    `gorm:"size:500" json:"request_path"`
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}

func (a *PermissionAuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
