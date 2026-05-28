package service

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/repository"
)

// SupplierService handles supplier business logic.
type SupplierService struct {
	repo *repository.SupplierRepository
}

func NewSupplierService(db *sql.DB) *SupplierService {
	return &SupplierService{repo: repository.NewSupplierRepository(db)}
}

func (s *SupplierService) List(keyword string, page, pageSize int) ([]model.Supplier, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(keyword, page, pageSize)
}

func (s *SupplierService) Get(id string) (*model.Supplier, error) {
	return s.repo.FindByID(id)
}

func (s *SupplierService) Create(input SupplierInput) (*model.Supplier, error) {
	code, err := s.repo.GenerateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}
	sup := &model.Supplier{
		Code:        code,
		Name:        input.Name,
		Contact:     input.Contact,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		BankAccount: input.BankAccount,
		TaxID:       input.TaxID,
		Status:      "active",
	}
	if err := s.repo.Create(sup); err != nil {
		return nil, err
	}
	return sup, nil
}

func (s *SupplierService) Update(id string, input SupplierInput) (*model.Supplier, error) {
	sup, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if sup == nil {
		return nil, nil
	}
	sup.Name = input.Name
	sup.Contact = input.Contact
	sup.Phone = input.Phone
	sup.Email = input.Email
	sup.Address = input.Address
	sup.BankAccount = input.BankAccount
	sup.TaxID = input.TaxID
	if input.Status != "" {
		sup.Status = input.Status
	}
	if err := s.repo.Update(sup); err != nil {
		return nil, err
	}
	return sup, nil
}

type SupplierInput struct {
	Name        string `json:"name"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	BankAccount string `json:"bankAccount"`
	TaxID       string `json:"taxId"`
	Status      string `json:"status"`
}

// CustomerService handles customer business logic.
type CustomerService struct {
	repo *repository.CustomerRepository
}

func NewCustomerService(db *sql.DB) *CustomerService {
	return &CustomerService{repo: repository.NewCustomerRepository(db)}
}

func (s *CustomerService) List(keyword string, page, pageSize int) ([]model.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.FindAll(keyword, page, pageSize)
}

func (s *CustomerService) Get(id string) (*model.Customer, error) {
	return s.repo.FindByID(id)
}

func (s *CustomerService) Create(input CustomerInput) (*model.Customer, error) {
	code, err := s.repo.GenerateCode()
	if err != nil {
		return nil, fmt.Errorf("generate code: %w", err)
	}
	cust := &model.Customer{
		Code:        code,
		Name:        input.Name,
		Contact:     input.Contact,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		CreditLimit: input.CreditLimit,
		Status:      "active",
	}
	if err := s.repo.Create(cust); err != nil {
		return nil, err
	}
	return cust, nil
}

func (s *CustomerService) Update(id string, input CustomerInput) (*model.Customer, error) {
	cust, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if cust == nil {
		return nil, nil
	}
	cust.Name = input.Name
	cust.Contact = input.Contact
	cust.Phone = input.Phone
	cust.Email = input.Email
	cust.Address = input.Address
	cust.CreditLimit = input.CreditLimit
	if input.Status != "" {
		cust.Status = input.Status
	}
	if err := s.repo.Update(cust); err != nil {
		return nil, err
	}
	return cust, nil
}

type CustomerInput struct {
	Name        string  `json:"name"`
	Contact     string  `json:"contact"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Address     string  `json:"address"`
	CreditLimit float64 `json:"creditLimit"`
	Status      string  `json:"status"`
}
