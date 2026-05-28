package service

import (
	"database/sql"
	"sync"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

type DashboardService struct {
	supplierRepo    *repository.SupplierRepository
	customerRepo    *repository.CustomerRepository
	purchaseRepo    *repository.PurchaseOrderRepository
	salesRepo       *repository.SalesOrderRepository
	inventoryRepo   *repository.InventoryRepository
	stocktakingRepo *repository.StocktakingRepository
}

func NewDashboardService(db *sql.DB) *DashboardService {
	return &DashboardService{
		supplierRepo:    repository.NewSupplierRepository(db),
		customerRepo:    repository.NewCustomerRepository(db),
		purchaseRepo:    repository.NewPurchaseOrderRepository(db),
		salesRepo:       repository.NewSalesOrderRepository(db),
		inventoryRepo:   repository.NewInventoryRepository(db),
		stocktakingRepo: repository.NewStocktakingRepository(db),
	}
}

func (s *DashboardService) GetDashboard() (*model.DashboardData, error) {
	var (
		d             model.DashboardData
		mu            sync.Mutex
		wg            sync.WaitGroup
		errs          []error
	)

	run := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	run(func() error {
		_, total, err := s.supplierRepo.FindAll("", 1, 1)
		mu.Lock()
		d.TotalSuppliers = total
		mu.Unlock()
		return err
	})

	run(func() error {
		_, total, err := s.customerRepo.FindAll("", 1, 1)
		mu.Lock()
		d.TotalCustomers = total
		mu.Unlock()
		return err
	})

	run(func() error {
		_, total, err := s.purchaseRepo.FindAll("", 1, 1)
		mu.Lock()
		d.PurchaseThisMonth = total
		mu.Unlock()
		return err
	})

	run(func() error {
		_, total, err := s.salesRepo.FindAll("", 1, 1)
		mu.Lock()
		d.SalesThisMonth = total
		mu.Unlock()
		return err
	})

	run(func() error {
		_, total, err := s.purchaseRepo.FindAll("submitted", 1, 1)
		mu.Lock()
		d.PendingPO = total
		mu.Unlock()
		return err
	})

	run(func() error {
		count, err := s.inventoryRepo.CountLowStock()
		mu.Lock()
		d.LowStockCount = count
		mu.Unlock()
		return err
	})

	run(func() error {
		val, err := s.inventoryRepo.TotalValue()
		mu.Lock()
		d.InventoryValue = val
		mu.Unlock()
		return err
	})

	run(func() error {
		var total float64
		orders, _, err := s.purchaseRepo.FindAll("", 1, 10000)
		if err == nil {
			for _, o := range orders {
				total += o.TotalAmount
			}
		}
		mu.Lock()
		d.PurchaseAmount = total
		mu.Unlock()
		return err
	})

	run(func() error {
		var total float64
		orders, _, err := s.salesRepo.FindAll("", 1, 10000)
		if err == nil {
			for _, o := range orders {
				total += o.TotalAmount
			}
		}
		mu.Lock()
		d.SalesAmount = total
		mu.Unlock()
		return err
	})

	wg.Wait()

	if len(errs) > 0 && d.TotalSuppliers == 0 {
		return nil, errs[0]
	}

	return &d, nil
}
