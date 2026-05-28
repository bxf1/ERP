package repository

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
)

const supplierColumns = `id, code, name, contact, phone, email, address, bank_account, tax_id, status, created_at, updated_at`

type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) FindAll(keyword string, page, pageSize int) ([]model.Supplier, int64, error) {
	var total int64
	args := []any{}

	where := "WHERE 1=1"
	if keyword != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR contact ILIKE $%d)",
			len(args)+1, len(args)+2, len(args)+3)
		k := "%" + keyword + "%"
		args = append(args, k, k, k)
	}

	err := r.db.QueryRow("SELECT COUNT(*) FROM suppliers "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM suppliers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		supplierColumns, where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Supplier
	for rows.Next() {
		var s model.Supplier
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Contact, &s.Phone, &s.Email,
			&s.Address, &s.BankAccount, &s.TaxID, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *SupplierRepository) FindByID(id string) (*model.Supplier, error) {
	var s model.Supplier
	err := r.db.QueryRow(
		"SELECT "+supplierColumns+" FROM suppliers WHERE id = $1", id,
	).Scan(&s.ID, &s.Code, &s.Name, &s.Contact, &s.Phone, &s.Email,
		&s.Address, &s.BankAccount, &s.TaxID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) Create(s *model.Supplier) error {
	return r.db.QueryRow(
		`INSERT INTO suppliers (code, name, contact, phone, email, address, bank_account, tax_id, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at`,
		s.Code, s.Name, s.Contact, s.Phone, s.Email, s.Address, s.BankAccount, s.TaxID, s.Status,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SupplierRepository) Update(s *model.Supplier) error {
	_, err := r.db.Exec(
		`UPDATE suppliers SET code=$1, name=$2, contact=$3, phone=$4, email=$5, address=$6,
		 bank_account=$7, tax_id=$8, status=$9, updated_at=now() WHERE id=$10`,
		s.Code, s.Name, s.Contact, s.Phone, s.Email, s.Address,
		s.BankAccount, s.TaxID, s.Status, s.ID,
	)
	return err
}

func (r *SupplierRepository) GenerateCode() (string, error) {
	var maxCode sql.NullString
	err := r.db.QueryRow(
		"SELECT MAX(code) FROM suppliers WHERE code ~ '^SUP-[0-9]+$'",
	).Scan(&maxCode)
	if err != nil {
		return "", err
	}
	num := 1
	if maxCode.Valid {
		fmt.Sscanf(maxCode.String, "SUP-%d", &num)
		num++
	}
	return fmt.Sprintf("SUP-%04d", num), nil
}

// CustomerRepository

const customerColumns = `id, code, name, contact, phone, email, address, credit_limit, status, created_at, updated_at`

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindAll(keyword string, page, pageSize int) ([]model.Customer, int64, error) {
	var total int64
	args := []any{}
	where := "WHERE 1=1"
	if keyword != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR contact ILIKE $%d)",
			len(args)+1, len(args)+2, len(args)+3)
		k := "%" + keyword + "%"
		args = append(args, k, k, k)
	}

	err := r.db.QueryRow("SELECT COUNT(*) FROM customers "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM customers %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		customerColumns, where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Contact, &c.Phone, &c.Email,
			&c.Address, &c.CreditLimit, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}

func (r *CustomerRepository) FindByID(id string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.QueryRow(
		"SELECT "+customerColumns+" FROM customers WHERE id = $1", id,
	).Scan(&c.ID, &c.Code, &c.Name, &c.Contact, &c.Phone, &c.Email,
		&c.Address, &c.CreditLimit, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Create(c *model.Customer) error {
	return r.db.QueryRow(
		`INSERT INTO customers (code, name, contact, phone, email, address, credit_limit, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`,
		c.Code, c.Name, c.Contact, c.Phone, c.Email, c.Address, c.CreditLimit, c.Status,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CustomerRepository) Update(c *model.Customer) error {
	_, err := r.db.Exec(
		`UPDATE customers SET code=$1, name=$2, contact=$3, phone=$4, email=$5, address=$6,
		 credit_limit=$7, status=$8, updated_at=now() WHERE id=$9`,
		c.Code, c.Name, c.Contact, c.Phone, c.Email, c.Address,
		c.CreditLimit, c.Status, c.ID,
	)
	return err
}

func (r *CustomerRepository) GenerateCode() (string, error) {
	var maxCode sql.NullString
	err := r.db.QueryRow(
		"SELECT MAX(code) FROM customers WHERE code ~ '^CUS-[0-9]+$'",
	).Scan(&maxCode)
	if err != nil {
		return "", err
	}
	num := 1
	if maxCode.Valid {
		fmt.Sscanf(maxCode.String, "CUS-%d", &num)
		num++
	}
	return fmt.Sprintf("CUS-%04d", num), nil
}
