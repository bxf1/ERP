package repository

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
)

const poColumns = `id, order_no, supplier_id, supplier_name, order_date, delivery_date, total_amount, status, remark, created_at, updated_at`
const poiColumns = `id, purchase_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount`

type PurchaseOrderRepository struct {
	db *sql.DB
}

func NewPurchaseOrderRepository(db *sql.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{db: db}
}

func (r *PurchaseOrderRepository) FindAll(status string, page, pageSize int) ([]model.PurchaseOrder, int64, error) {
	var total int64
	args := []any{}
	where := "WHERE 1=1"
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, status)
	}

	err := r.db.QueryRow("SELECT COUNT(*) FROM purchase_orders "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM purchase_orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		poColumns, where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.PurchaseOrder
	for rows.Next() {
		var po model.PurchaseOrder
		if err := rows.Scan(&po.ID, &po.OrderNo, &po.SupplierID, &po.SupplierName,
			&po.OrderDate, &po.DeliveryDate, &po.TotalAmount, &po.Status, &po.Remark,
			&po.CreatedAt, &po.UpdatedAt); err != nil {
			return nil, 0, err
		}
		po.Items = []model.PurchaseOrderItem{}
		items = append(items, po)
	}
	return items, total, rows.Err()
}

func (r *PurchaseOrderRepository) FindByID(id string) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	err := r.db.QueryRow(
		"SELECT "+poColumns+" FROM purchase_orders WHERE id = $1", id,
	).Scan(&po.ID, &po.OrderNo, &po.SupplierID, &po.SupplierName,
		&po.OrderDate, &po.DeliveryDate, &po.TotalAmount, &po.Status, &po.Remark,
		&po.CreatedAt, &po.UpdatedAt)
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
	po.Items = items
	return &po, nil
}

func (r *PurchaseOrderRepository) findItems(orderID string) ([]model.PurchaseOrderItem, error) {
	rows, err := r.db.Query(
		"SELECT "+poiColumns+" FROM purchase_order_items WHERE purchase_order_id = $1 ORDER BY created_at",
		orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.PurchaseOrderItem
	for rows.Next() {
		var i model.PurchaseOrderItem
		if err := rows.Scan(&i.ID, &i.PurchaseOrderID, &i.ProductID, &i.ProductCode,
			&i.ProductName, &i.Specification, &i.Unit, &i.Quantity, &i.UnitPrice, &i.Amount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if items == nil {
		items = []model.PurchaseOrderItem{}
	}
	return items, rows.Err()
}

func (r *PurchaseOrderRepository) Create(po *model.PurchaseOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO purchase_orders (order_no, supplier_id, supplier_name, order_date, delivery_date, total_amount, status, remark)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at`,
		po.OrderNo, po.SupplierID, po.SupplierName, po.OrderDate, po.DeliveryDate,
		po.TotalAmount, po.Status, po.Remark,
	).Scan(&po.ID, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range po.Items {
		po.Items[i].PurchaseOrderID = po.ID
		err = tx.QueryRow(
			`INSERT INTO purchase_order_items (purchase_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
			po.Items[i].PurchaseOrderID, po.Items[i].ProductID, po.Items[i].ProductCode,
			po.Items[i].ProductName, po.Items[i].Specification, po.Items[i].Unit,
			po.Items[i].Quantity, po.Items[i].UnitPrice, po.Items[i].Amount,
		).Scan(&po.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PurchaseOrderRepository) Update(po *model.PurchaseOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE purchase_orders SET order_no=$1, supplier_id=$2, supplier_name=$3, order_date=$4,
		 delivery_date=$5, total_amount=$6, status=$7, remark=$8, updated_at=now() WHERE id=$9`,
		po.OrderNo, po.SupplierID, po.SupplierName, po.OrderDate, po.DeliveryDate,
		po.TotalAmount, po.Status, po.Remark, po.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM purchase_order_items WHERE purchase_order_id = $1", po.ID)
	if err != nil {
		return err
	}

	for i := range po.Items {
		po.Items[i].PurchaseOrderID = po.ID
		err = tx.QueryRow(
			`INSERT INTO purchase_order_items (purchase_order_id, product_id, product_code, product_name, specification, unit, quantity, unit_price, amount)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
			po.Items[i].PurchaseOrderID, po.Items[i].ProductID, po.Items[i].ProductCode,
			po.Items[i].ProductName, po.Items[i].Specification, po.Items[i].Unit,
			po.Items[i].Quantity, po.Items[i].UnitPrice, po.Items[i].Amount,
		).Scan(&po.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PurchaseOrderRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(
		"UPDATE purchase_orders SET status=$1, updated_at=now() WHERE id=$2",
		status, id,
	)
	return err
}

func (r *PurchaseOrderRepository) GenerateOrderNo() (string, error) {
	var maxNo sql.NullString
	err := r.db.QueryRow(
		"SELECT MAX(order_no) FROM purchase_orders WHERE order_no ~ '^PO-[0-9]+$'",
	).Scan(&maxNo)
	if err != nil {
		return "", err
	}
	num := 1
	if maxNo.Valid {
		fmt.Sscanf(maxNo.String, "PO-%d", &num)
		num++
	}
	return fmt.Sprintf("PO-%06d", num), nil
}
