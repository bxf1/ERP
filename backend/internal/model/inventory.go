package model

import "time"

// Supplier represents a vendor/supplier.
type Supplier struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Contact     string    `json:"contact"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	BankAccount string    `json:"bankAccount"`
	TaxID       string    `json:"taxId"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Customer represents a customer.
type Customer struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Contact     string    `json:"contact"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	CreditLimit float64   `json:"creditLimit"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PurchaseOrder represents a purchase order header.
type PurchaseOrder struct {
	ID           string              `json:"id"`
	OrderNo      string              `json:"orderNo"`
	SupplierID   string              `json:"supplierId"`
	SupplierName string              `json:"supplierName"`
	OrderDate    string              `json:"orderDate"`
	DeliveryDate string              `json:"deliveryDate"`
	Items        []PurchaseOrderItem `json:"items"`
	TotalAmount  float64             `json:"totalAmount"`
	Status       string              `json:"status"`
	Remark       string              `json:"remark"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
}

// PurchaseOrderItem represents a line item in a purchase order.
type PurchaseOrderItem struct {
	ID              string  `json:"id"`
	PurchaseOrderID string  `json:"purchaseOrderId"`
	ProductID       string  `json:"productId"`
	ProductCode     string  `json:"productCode"`
	ProductName     string  `json:"productName"`
	Specification   string  `json:"specification"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unitPrice"`
	Amount          float64 `json:"amount"`
}

// SalesOrder represents a sales order header.
type SalesOrder struct {
	ID           string            `json:"id"`
	OrderNo      string            `json:"orderNo"`
	CustomerID   string            `json:"customerId"`
	CustomerName string            `json:"customerName"`
	OrderDate    string            `json:"orderDate"`
	DeliveryDate string            `json:"deliveryDate"`
	Items        []SalesOrderItem  `json:"items"`
	TotalAmount  float64           `json:"totalAmount"`
	Status       string            `json:"status"`
	Remark       string            `json:"remark"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

// SalesOrderItem represents a line item in a sales order.
type SalesOrderItem struct {
	ID             string  `json:"id"`
	SalesOrderID   string  `json:"salesOrderId"`
	ProductID      string  `json:"productId"`
	ProductCode    string  `json:"productCode"`
	ProductName    string  `json:"productName"`
	Specification  string  `json:"specification"`
	Unit           string  `json:"unit"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Amount         float64 `json:"amount"`
}

// InventoryRecord represents a stock record for a product in a warehouse.
type InventoryRecord struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"productId"`
	ProductCode   string    `json:"productCode"`
	ProductName   string    `json:"productName"`
	Specification string    `json:"specification"`
	Unit          string    `json:"unit"`
	WarehouseID   string    `json:"warehouseId"`
	WarehouseName string    `json:"warehouseName"`
	Quantity      float64   `json:"quantity"`
	SafetyStock   float64   `json:"safetyStock"`
	CostPrice     float64   `json:"costPrice"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Stocktaking represents a stocktaking task.
type Stocktaking struct {
	ID            string             `json:"id"`
	TaskNo        string             `json:"taskNo"`
	WarehouseID   string             `json:"warehouseId"`
	WarehouseName string             `json:"warehouseName"`
	StartDate     string             `json:"startDate"`
	EndDate       string             `json:"endDate"`
	Status        string             `json:"status"`
	Items         []StocktakingItem  `json:"items"`
	Remark        string             `json:"remark"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

// StocktakingItem represents a line item in a stocktaking task.
type StocktakingItem struct {
	ID             string  `json:"id"`
	StocktakingID  string  `json:"stocktakingId"`
	ProductID      string  `json:"productId"`
	ProductCode    string  `json:"productCode"`
	ProductName    string  `json:"productName"`
	Specification  string  `json:"specification"`
	Unit           string  `json:"unit"`
	BookQuantity   float64 `json:"bookQuantity"`
	ActualQuantity float64 `json:"actualQuantity"`
	DiffQuantity   float64 `json:"diffQuantity"`
	Remark         string  `json:"remark"`
}

// DashboardData aggregates KPIs for the dashboard page.
type DashboardData struct {
	TotalSuppliers   int64   `json:"totalSuppliers"`
	TotalCustomers   int64   `json:"totalCustomers"`
	PurchaseThisMonth int64  `json:"purchaseThisMonth"`
	SalesThisMonth   int64   `json:"salesThisMonth"`
	PurchaseAmount   float64 `json:"purchaseAmount"`
	SalesAmount      float64 `json:"salesAmount"`
	InventoryValue   float64 `json:"inventoryValue"`
	LowStockCount    int64   `json:"lowStockCount"`
	PendingPO        int64   `json:"pendingPO"`
	PendingSO        int64   `json:"pendingSO"`
}
