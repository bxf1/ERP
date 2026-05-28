package repository

import (
	"database/sql"
	"fmt"

	"github.com/bxf1/ERP/backend/internal/model"
)

const invColumns = `id, product_id, product_code, product_name, specification, unit, warehouse_id, warehouse_name, quantity, safety_stock, cost_price, created_at, updated_at`

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) FindAll(keyword string, page, pageSize int) ([]model.InventoryRecord, int64, error) {
	var total int64
	args := []any{}
	where := "WHERE 1=1"
	if keyword != "" {
		where += fmt.Sprintf(" AND (product_name ILIKE $%d OR product_code ILIKE $%d)",
			len(args)+1, len(args)+2)
		k := "%" + keyword + "%"
		args = append(args, k, k)
	}

	err := r.db.QueryRow("SELECT COUNT(*) FROM inventory "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM inventory %s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d",
		invColumns, where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.InventoryRecord
	for rows.Next() {
		var inv model.InventoryRecord
		if err := rows.Scan(&inv.ID, &inv.ProductID, &inv.ProductCode, &inv.ProductName,
			&inv.Specification, &inv.Unit, &inv.WarehouseID, &inv.WarehouseName,
			&inv.Quantity, &inv.SafetyStock, &inv.CostPrice, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, inv)
	}
	return items, total, rows.Err()
}

func (r *InventoryRepository) FindLowStock() ([]model.InventoryRecord, error) {
	rows, err := r.db.Query(
		"SELECT "+invColumns+" FROM inventory WHERE quantity <= safety_stock AND safety_stock > 0 ORDER BY quantity ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.InventoryRecord
	for rows.Next() {
		var inv model.InventoryRecord
		if err := rows.Scan(&inv.ID, &inv.ProductID, &inv.ProductCode, &inv.ProductName,
			&inv.Specification, &inv.Unit, &inv.WarehouseID, &inv.WarehouseName,
			&inv.Quantity, &inv.SafetyStock, &inv.CostPrice, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, inv)
	}
	if items == nil {
		items = []model.InventoryRecord{}
	}
	return items, rows.Err()
}

func (r *InventoryRepository) Upsert(rec *model.InventoryRecord) error {
	return r.db.QueryRow(
		`INSERT INTO inventory (product_id, product_code, product_name, specification, unit, warehouse_id, warehouse_name, quantity, safety_stock, cost_price)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (product_id, warehouse_id)
		 DO UPDATE SET quantity=$8, safety_stock=$9, cost_price=$10, updated_at=now()
		 RETURNING id, created_at, updated_at`,
		rec.ProductID, rec.ProductCode, rec.ProductName, rec.Specification, rec.Unit,
		rec.WarehouseID, rec.WarehouseName, rec.Quantity, rec.SafetyStock, rec.CostPrice,
	).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt)
}

func (r *InventoryRepository) AdjustQuantity(productID, warehouseID string, delta float64) error {
	_, err := r.db.Exec(
		`UPDATE inventory SET quantity = quantity + $1, updated_at = now()
		 WHERE product_id = $2 AND warehouse_id = $3`,
		delta, productID, warehouseID,
	)
	return err
}

func (r *InventoryRepository) CountLowStock() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM inventory WHERE quantity <= safety_stock AND safety_stock > 0",
	).Scan(&count)
	return count, err
}

func (r *InventoryRepository) TotalValue() (float64, error) {
	var total sql.NullFloat64
	err := r.db.QueryRow(
		"SELECT SUM(quantity * cost_price) FROM inventory",
	).Scan(&total)
	if !total.Valid {
		return 0, err
	}
	return total.Float64, err
}

// StocktakingRepository

const stColumns = `id, task_no, warehouse_id, warehouse_name, start_date, end_date, status, remark, created_at, updated_at`
const stiColumns = `id, stocktaking_id, product_id, product_code, product_name, specification, unit, book_quantity, actual_quantity, diff_quantity, remark`

type StocktakingRepository struct {
	db *sql.DB
}

func NewStocktakingRepository(db *sql.DB) *StocktakingRepository {
	return &StocktakingRepository{db: db}
}

func (r *StocktakingRepository) FindAll(page, pageSize int) ([]model.Stocktaking, int64, error) {
	var total int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM stocktaking").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(
		"SELECT "+stColumns+" FROM stocktaking ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Stocktaking
	for rows.Next() {
		var s model.Stocktaking
		if err := rows.Scan(&s.ID, &s.TaskNo, &s.WarehouseID, &s.WarehouseName,
			&s.StartDate, &s.EndDate, &s.Status, &s.Remark, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		s.Items = []model.StocktakingItem{}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *StocktakingRepository) FindByID(id string) (*model.Stocktaking, error) {
	var s model.Stocktaking
	err := r.db.QueryRow(
		"SELECT "+stColumns+" FROM stocktaking WHERE id = $1", id,
	).Scan(&s.ID, &s.TaskNo, &s.WarehouseID, &s.WarehouseName,
		&s.StartDate, &s.EndDate, &s.Status, &s.Remark, &s.CreatedAt, &s.UpdatedAt)
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
	s.Items = items
	return &s, nil
}

func (r *StocktakingRepository) findItems(taskID string) ([]model.StocktakingItem, error) {
	rows, err := r.db.Query(
		"SELECT "+stiColumns+" FROM stocktaking_items WHERE stocktaking_id = $1 ORDER BY created_at",
		taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.StocktakingItem
	for rows.Next() {
		var i model.StocktakingItem
		if err := rows.Scan(&i.ID, &i.StocktakingID, &i.ProductID, &i.ProductCode,
			&i.ProductName, &i.Specification, &i.Unit, &i.BookQuantity,
			&i.ActualQuantity, &i.DiffQuantity, &i.Remark); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if items == nil {
		items = []model.StocktakingItem{}
	}
	return items, rows.Err()
}

func (r *StocktakingRepository) Create(s *model.Stocktaking) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO stocktaking (task_no, warehouse_id, warehouse_name, start_date, end_date, status, remark)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at, updated_at`,
		s.TaskNo, s.WarehouseID, s.WarehouseName, s.StartDate, s.EndDate, s.Status, s.Remark,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return err
	}

	for i := range s.Items {
		s.Items[i].StocktakingID = s.ID
		s.Items[i].DiffQuantity = s.Items[i].ActualQuantity - s.Items[i].BookQuantity
		err = tx.QueryRow(
			`INSERT INTO stocktaking_items (stocktaking_id, product_id, product_code, product_name, specification, unit, book_quantity, actual_quantity, diff_quantity, remark)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
			s.Items[i].StocktakingID, s.Items[i].ProductID, s.Items[i].ProductCode,
			s.Items[i].ProductName, s.Items[i].Specification, s.Items[i].Unit,
			s.Items[i].BookQuantity, s.Items[i].ActualQuantity, s.Items[i].DiffQuantity, s.Items[i].Remark,
		).Scan(&s.Items[i].ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *StocktakingRepository) GenerateTaskNo() (string, error) {
	var maxNo sql.NullString
	err := r.db.QueryRow(
		"SELECT MAX(task_no) FROM stocktaking WHERE task_no ~ '^ST-[0-9]+$'",
	).Scan(&maxNo)
	if err != nil {
		return "", err
	}
	num := 1
	if maxNo.Valid {
		fmt.Sscanf(maxNo.String, "ST-%d", &num)
		num++
	}
	return fmt.Sprintf("ST-%06d", num), nil
}
