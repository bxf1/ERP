package repository

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
)

const soColumns = `id, order_no, customer_id, customer_name, order_date, delivery_date, total_amount, status, remark, created_at, updated_at`
const soiColumns = `id, sales_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount`

type SalesOrderRepository struct {
	db *sql.DB
}

func NewSalesOrderRepository(db *sql.DB) *SalesOrderRepository {
	return &SalesOrderRepository{db: db}
}

func (r *SalesOrderRepository) FindAll(status string, page, pageSize int) ([]model.SalesOrder, int64, error) {
	var total int64
	args := []any{}
	where := "WHERE 1=1"
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, status)
	}

	err := r.db.QueryRow("SELECT COUNT(*) FROM sales_orders "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM sales_orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		soColumns, where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.SalesOrder
	for rows.Next() {
		var so model.SalesOrder
		if err := rows.Scan(&so.ID, &so.OrderNo, &so.CustomerID, &so.CustomerName,
			&so.OrderDate, &so.DeliveryDate, &so.TotalAmount, &so.Status, &so.Remark,
			&so.CreatedAt, &so.UpdatedAt); err != nil {
			return nil, 0, err
		}
		so.Items = []model.SalesOrderItem{}
		items = append(items, so)
	}
	return items, total, rows.Err()
}

func (r *SalesOrderRepository) FindByID(id string) (*model.SalesOrder, error) {
	var so model.SalesOrder
	err := r.db.QueryRow(
		"SELECT "+soColumns+" FROM sales_orders WHERE id = $1", id,
	).Scan(&so.ID, &so.OrderNo, &so.CustomerID, &so.CustomerName,
		&so.OrderDate, &so.DeliveryDate, &so.TotalAmount, &so.Status, &so.Remark,
		&so.CreatedAt, &so.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.findItems(id)
	if err != nil {
		return nil, err
	}
	so.Items = items
	return &so, nil
}

func (r *SalesOrderRepository) findItems(orderID string) ([]model.SalesOrderItem, error) {
	rows, err := r.db.Query(
		"SELECT "+soiColumns+" FROM sales_order_items WHERE sales_order_id = $1 ORDER BY created_at",
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.SalesOrderItem
	for rows.Next() {
		var i model.SalesOrderItem
		if err := rows.Scan(&i.ID, &i.SalesOrderID, &i.ProductID, &i.ProductCode,
			&i.ProductName, &i.Specification, &i.Unit, &i.Quantity, &i.UnitPrice, &i.Amount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if items == nil {
		items = []model.SalesOrderItem{}
	}
	return items, rows.Err()
}

func (r *SalesOrderRepository) Create(so *model.SalesOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO sales_orders (order_no, customer_id, customer_name, order_date, delivery_date, total_amount, status, remark)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`,
		so.OrderNo, so.CustomerID, so.CustomerName, so.OrderDate, so.DeliveryDate,
		so.TotalAmount, so.Status, so.Remark,
	).Scan(&so.ID, &so.CreatedAt, &so.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range so.Items {
		so.Items[i].SalesOrderID = so.ID
		err = tx.QueryRow(
			`INSERT INTO sales_order_items (sales_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
			so.Items[i].SalesOrderID, so.Items[i].ProductID, so.Items[i].ProductCode,
			so.Items[i].ProductName, so.Items[i].Specification, so.Items[i].Unit,
			so.Items[i].Quantity, so.Items[i].UnitPrice, so.Items[i].Amount,
		).Scan(&so.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SalesOrderRepository) Update(so *model.SalesOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE sales_orders SET order_no=$1, customer_id=$2, customer_name=$3, order_date=$4,
		 delivery_date=$5, total_amount=$6, status=$7, remark=$8, updated_at=now() WHERE id=$9`,
		so.OrderNo, so.CustomerID, so.CustomerName, so.OrderDate, so.DeliveryDate,
		so.TotalAmount, so.Status, so.Remark, so.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM sales_order_items WHERE sales_order_id = $1", so.ID)
	if err != nil {
		return err
	}

	for i := range so.Items {
		so.Items[i].SalesOrderID = so.ID
		err = tx.QueryRow(
			`INSERT INTO sales_order_items (sales_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
			so.Items[i].SalesOrderID, so.Items[i].ProductID, so.Items[i].ProductCode,
			so.Items[i].ProductName, so.Items[i].Specification, so.Items[i].Unit,
			so.Items[i].Quantity, so.Items[i].UnitPrice, so.Items[i].Amount,
		).Scan(&so.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SalesOrderRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(
		"UPDATE sales_orders SET status=$1, updated_at=now() WHERE id=$2",
		status, id,
	)
	return err
}

func (r *SalesOrderRepository) GenerateOrderNo() (string, error) {
	var maxNo sql.NullString
	err := r.db.QueryRow(
		"SELECT MAX(order_no) FROM sales_orders WHERE order_no ~ '^SO-[0-9]+$'",
	).Scan(&maxNo)
	if err != nil {
		return "", err
	}
	num := 1
	if maxNo.Valid {
		fmt.Sscanf(maxNo.String, "SO-%d", &num)
		num++
	}
	return fmt.Sprintf("SO-%06d", num), nil
}
