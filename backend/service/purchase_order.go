package service

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

type PurchaseOrderService struct {
	repo      *repository.PurchaseOrderRepository
	inventory *repository.InventoryRepository
}

func NewPurchaseOrderService(db *sql.DB) *PurchaseOrderService {
	return &PurchaseOrderService{
		repo:      repository.NewPurchaseOrderRepository(db),
		inventory: repository.NewInventoryRepository(db),
	}
}

func (s *PurchaseOrderService) List(status string, page, pageSize int) ([]model.PurchaseOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(status, page, pageSize)
}

func (s *PurchaseOrderService) Get(id string) (*model.PurchaseOrder, error) {
	return s.repo.FindByID(id)
}

func (s *PurchaseOrderService) Create(input PurchaseOrderInput) (*model.PurchaseOrder, error) {
	orderNo, err := s.repo.GenerateOrderNo()
	if err != nil {
		return nil, fmt.Errorf("generate order no: %w", err)
	}

	po := &model.PurchaseOrder{
		OrderNo:      orderNo,
		SupplierID:   input.SupplierID,
		SupplierName: input.SupplierName,
		OrderDate:    input.OrderDate,
		DeliveryDate: input.DeliveryDate,
		TotalAmount:  input.TotalAmount,
		Status:       "draft",
		Remark:       input.Remark,
		Items:        input.Items,
	}
	if po.Items == nil {
		po.Items = []model.PurchaseOrderItem{}
	}

	if err := s.repo.Create(po); err != nil {
		return nil, err
	}
	return po, nil
}

func (s *PurchaseOrderService) Update(id string, input PurchaseOrderInput) (*model.PurchaseOrder, error) {
	po, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if po == nil {
		return nil, nil
	}
	if po.Status != "draft" {
		return nil, fmt.Errorf("only draft orders can be edited")
	}

	po.SupplierID = input.SupplierID
	po.SupplierName = input.SupplierName
	po.OrderDate = input.OrderDate
	po.DeliveryDate = input.DeliveryDate
	po.TotalAmount = input.TotalAmount
	po.Remark = input.Remark
	if input.Items != nil {
		po.Items = input.Items
	}

	if err := s.repo.Update(po); err != nil {
		return nil, err
	}
	return po, nil
}

func (s *PurchaseOrderService) Submit(id string) (*model.PurchaseOrder, error) {
	po, err := s.repo.FindByID(id)
	if err != nil || po == nil {
		return po, err
	}
	if po.Status != "draft" {
		return nil, fmt.Errorf("only draft orders can be submitted")
	}
	if err := s.repo.UpdateStatus(id, "submitted"); err != nil {
		return nil, err
	}
	po.Status = "submitted"
	return po, nil
}

func (s *PurchaseOrderService) Approve(id string) (*model.PurchaseOrder, error) {
	po, err := s.repo.FindByID(id)
	if err != nil || po == nil {
		return po, err
	}
	if po.Status != "submitted" {
		return nil, fmt.Errorf("only submitted orders can be approved")
	}
	if err := s.repo.UpdateStatus(id, "approved"); err != nil {
		return nil, err
	}
	po.Status = "approved"
	return po, nil
}

func (s *PurchaseOrderService) Receive(id string) (*model.PurchaseOrder, error) {
	po, err := s.repo.FindByID(id)
	if err != nil || po == nil {
		return po, err
	}
	if po.Status != "approved" {
		return nil, fmt.Errorf("only approved orders can be received")
	}

	// Increase inventory for each item
	for _, item := range po.Items {
		rec := model.InventoryRecord{
			ProductID:     item.ProductID,
			ProductCode:   item.ProductCode,
			ProductName:   item.ProductName,
			Specification: item.Specification,
			Unit:          item.Unit,
			WarehouseID:   "default",
			WarehouseName: "默认仓库",
			Quantity:      item.Quantity,
			SafetyStock:   0,
			CostPrice:     item.UnitPrice,
		}
		if err := s.inventory.Upsert(&rec); err != nil {
			return nil, fmt.Errorf("update inventory for %s: %w", item.ProductName, err)
		}
	}

	if err := s.repo.UpdateStatus(id, "received"); err != nil {
		return nil, err
	}
	po.Status = "received"
	return po, nil
}

type PurchaseOrderInput struct {
	SupplierID   string                     `json:"supplierId"`
	SupplierName string                     `json:"supplierName"`
	OrderDate    string                     `json:"orderDate"`
	DeliveryDate string                     `json:"deliveryDate"`
	TotalAmount  float64                    `json:"totalAmount"`
	Remark       string                     `json:"remark"`
	Items        []model.PurchaseOrderItem  `json:"items"`
}
