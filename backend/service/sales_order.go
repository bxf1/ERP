package service

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

type SalesOrderService struct {
	repo      *repository.SalesOrderRepository
	inventory *repository.InventoryRepository
}

func NewSalesOrderService(db *sql.DB) *SalesOrderService {
	return &SalesOrderService{
		repo:      repository.NewSalesOrderRepository(db),
		inventory: repository.NewInventoryRepository(db),
	}
}

func (s *SalesOrderService) List(status string, page, pageSize int) ([]model.SalesOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(status, page, pageSize)
}

func (s *SalesOrderService) Get(id string) (*model.SalesOrder, error) {
	return s.repo.FindByID(id)
}

func (s *SalesOrderService) Create(input SalesOrderInput) (*model.SalesOrder, error) {
	orderNo, err := s.repo.GenerateOrderNo()
	if err != nil {
		return nil, fmt.Errorf("generate order no: %w", err)
	}

	so := &model.SalesOrder{
		OrderNo:      orderNo,
		CustomerID:   input.CustomerID,
		CustomerName: input.CustomerName,
		OrderDate:    input.OrderDate,
		DeliveryDate: input.DeliveryDate,
		TotalAmount:  input.TotalAmount,
		Status:       "draft",
		Remark:       input.Remark,
		Items:        input.Items,
	}
	if so.Items == nil {
		so.Items = []model.SalesOrderItem{}
	}

	if err := s.repo.Create(so); err != nil {
		return nil, err
	}
	return so, nil
}

func (s *SalesOrderService) Update(id string, input SalesOrderInput) (*model.SalesOrder, error) {
	so, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if so == nil {
		return nil, nil
	}
	if so.Status != "draft" {
		return nil, fmt.Errorf("only draft orders can be edited")
	}

	so.CustomerID = input.CustomerID
	so.CustomerName = input.CustomerName
	so.OrderDate = input.OrderDate
	so.DeliveryDate = input.DeliveryDate
	so.TotalAmount = input.TotalAmount
	so.Remark = input.Remark
	if input.Items != nil {
		so.Items = input.Items
	}

	if err := s.repo.Update(so); err != nil {
		return nil, err
	}
	return so, nil
}

func (s *SalesOrderService) Confirm(id string) (*model.SalesOrder, error) {
	so, err := s.repo.FindByID(id)
	if err != nil || so == nil {
		return so, err
	}
	if so.Status != "draft" {
		return nil, fmt.Errorf("only draft orders can be confirmed")
	}
	if err := s.repo.UpdateStatus(id, "confirmed"); err != nil {
		return nil, err
	}
	so.Status = "confirmed"
	return so, nil
}

func (s *SalesOrderService) Ship(id string) (*model.SalesOrder, error) {
	so, err := s.repo.FindByID(id)
	if err != nil || so == nil {
		return so, err
	}
	if so.Status != "confirmed" {
		return nil, fmt.Errorf("only confirmed orders can be shipped")
	}

	// Deduct inventory for each item
	for _, item := range so.Items {
		if err := s.inventory.AdjustQuantity(item.ProductID, "default", -item.Quantity); err != nil {
			return nil, fmt.Errorf("deduct inventory for %s: %w", item.ProductName, err)
		}
	}

	if err := s.repo.UpdateStatus(id, "shipped"); err != nil {
		return nil, err
	}
	so.Status = "shipped"
	return so, nil
}

func (s *SalesOrderService) Invoice(id string) (*model.SalesOrder, error) {
	so, err := s.repo.FindByID(id)
	if err != nil || so == nil {
		return so, err
	}
	if so.Status != "shipped" {
		return nil, fmt.Errorf("only shipped orders can be invoiced")
	}
	if err := s.repo.UpdateStatus(id, "invoiced"); err != nil {
		return nil, err
	}
	so.Status = "invoiced"
	return so, nil
}

type SalesOrderInput struct {
	CustomerID   string                   `json:"customerId"`
	CustomerName string                   `json:"customerName"`
	OrderDate    string                   `json:"orderDate"`
	DeliveryDate string                   `json:"deliveryDate"`
	TotalAmount  float64                  `json:"totalAmount"`
	Remark       string                   `json:"remark"`
	Items        []model.SalesOrderItem   `json:"items"`
}
