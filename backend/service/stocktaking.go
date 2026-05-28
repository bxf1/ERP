package service

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

type StocktakingService struct {
	repo *repository.StocktakingRepository
}

func NewStocktakingService(db *sql.DB) *StocktakingService {
	return &StocktakingService{repo: repository.NewStocktakingRepository(db)}
}

func (s *StocktakingService) List(page, pageSize int) ([]model.Stocktaking, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(page, pageSize)
}

func (s *StocktakingService) Get(id string) (*model.Stocktaking, error) {
	return s.repo.FindByID(id)
}

func (s *StocktakingService) Create(input StocktakingInput) (*model.Stocktaking, error) {
	taskNo, err := s.repo.GenerateTaskNo()
	if err != nil {
		return nil, fmt.Errorf("generate task no: %w", err)
	}

	st := &model.Stocktaking{
		TaskNo:        taskNo,
		WarehouseID:   input.WarehouseID,
		WarehouseName: input.WarehouseName,
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		Status:        "pending",
		Items:         input.Items,
		Remark:        input.Remark,
	}
	if st.Items == nil {
		st.Items = []model.StocktakingItem{}
	}

	if err := s.repo.Create(st); err != nil {
		return nil, err
	}
	return st, nil
}

type StocktakingInput struct {
	WarehouseID   string                    `json:"warehouseId"`
	WarehouseName string                    `json:"warehouseName"`
	StartDate     string                    `json:"startDate"`
	EndDate       string                    `json:"endDate"`
	Remark        string                    `json:"remark"`
	Items         []model.StocktakingItem   `json:"items"`
}
