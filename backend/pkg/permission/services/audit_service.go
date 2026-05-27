package services

import (
	"context"
	"time"

	"github.com/bxf1/ERP/backend/pkg/permission/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditService struct {
	db      *gorm.DB
	enabled bool
}

func NewAuditService(db *gorm.DB, enabled bool) *AuditService {
	return &AuditService{db: db, enabled: enabled}
}

func (s *AuditService) Log(ctx context.Context, userID uuid.UUID, action, resource, targetID, requestPath, ip string, allowed bool, reason string) {
	if !s.enabled {
		return
	}
	result := "allow"
	if !allowed {
		result = "deny"
	}
	log := models.PermissionAuditLog{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		TargetID:    targetID,
		Result:      result,
		Reason:      reason,
		RequestPath: requestPath,
		IPAddress:   ip,
		CreatedAt:   time.Now(),
	}
	// Fire and forget — don't block on audit logging
	go func() {
		_ = s.db.Create(&log).Error
	}()
}

// GetAuditLogs returns paginated audit logs.
func (s *AuditService) GetAuditLogs(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.PermissionAuditLog, int64, error) {
	var logs []models.PermissionAuditLog
	var total int64

	query := s.db.WithContext(ctx).Model(&models.PermissionAuditLog{})
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
