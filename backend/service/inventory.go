package service

import (
	"database/sql"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(db *sql.DB) *InventoryService {
	return &InventoryService{repo: repository.NewInventoryRepository(db)}
}

func (s *InventoryService) List(keyword string, page, pageSize int) ([]model.InventoryRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(keyword, page, pageSize)
}

func (s *InventoryService) LowStockAlerts() ([]model.InventoryRecord, error) {
	return s.repo.FindLowStock()
}
