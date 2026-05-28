package handler

import (
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bxf1/ERP/backend/internal/database"
	"github.com/bxf1/ERP/backend/internal/middleware"
	apperrors "github.com/bxf1/ERP/backend/internal/errors"
	"github.com/bxf1/ERP/backend/internal/response"
	"github.com/bxf1/ERP/backend/service"
)

func getTenantDB(c *gin.Context) (*sql.DB, error) {
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "tenant not resolved")
	}
	db, ok := database.GetTenantDB(tenant.ID)
	if !ok {
		return nil, apperrors.New(apperrors.CodeInternalError, "tenant database not available")
	}
	return db, nil
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}

// ---- Suppliers ----

func ListSuppliers(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSupplierService(db)
	items, total, err := svc.List(c.Query("keyword"), queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list suppliers", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetSupplier(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSupplierService(db)
	item, err := svc.Get(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("get supplier", err))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("supplier not found"))
		return
	}
	response.OK(c, item)
}

func CreateSupplier(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.SupplierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewSupplierService(db)
	item, err := svc.Create(input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.Created(c, item)
}

func UpdateSupplier(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.SupplierInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewSupplierService(db)
	item, err := svc.Update(c.Param("id"), input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("supplier not found"))
		return
	}
	response.OK(c, item)
}

// ---- Customers ----

func ListCustomers(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewCustomerService(db)
	items, total, err := svc.List(c.Query("keyword"), queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list customers", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetCustomer(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewCustomerService(db)
	item, err := svc.Get(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("get customer", err))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("customer not found"))
		return
	}
	response.OK(c, item)
}

func CreateCustomer(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewCustomerService(db)
	item, err := svc.Create(input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.Created(c, item)
}

func UpdateCustomer(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewCustomerService(db)
	item, err := svc.Update(c.Param("id"), input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("customer not found"))
		return
	}
	response.OK(c, item)
}

// ---- Purchase Orders ----

func ListPurchaseOrders(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewPurchaseOrderService(db)
	items, total, err := svc.List(c.Query("status"), queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list purchase orders", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetPurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Get(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("get purchase order", err))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("purchase order not found"))
		return
	}
	response.OK(c, item)
}

func CreatePurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.PurchaseOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Create(input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.Created(c, item)
}

func UpdatePurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.PurchaseOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Update(c.Param("id"), input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("purchase order not found"))
		return
	}
	response.OK(c, item)
}

func SubmitPurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Submit(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

func ApprovePurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Approve(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

func ReceivePurchaseOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewPurchaseOrderService(db)
	item, err := svc.Receive(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

// ---- Sales Orders ----

func ListSalesOrders(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSalesOrderService(db)
	items, total, err := svc.List(c.Query("status"), queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list sales orders", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Get(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("get sales order", err))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("sales order not found"))
		return
	}
	response.OK(c, item)
}

func CreateSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.SalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Create(input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.Created(c, item)
}

func UpdateSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.SalesOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Update(c.Param("id"), input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("sales order not found"))
		return
	}
	response.OK(c, item)
}

func ConfirmSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Confirm(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

func ShipSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Ship(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

func InvoiceSalesOrder(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewSalesOrderService(db)
	item, err := svc.Invoice(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.BadRequest(err.Error()))
		return
	}
	response.OK(c, item)
}

// ---- Inventory ----

func ListInventory(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	repo := service.NewInventoryService(db)
	items, total, err := repo.List(c.Query("keyword"), queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list inventory", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetLowStockAlerts(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	repo := service.NewInventoryService(db)
	items, err := repo.LowStockAlerts()
	if err != nil {
		response.Error(c, apperrors.DatabaseError("low stock alerts", err))
		return
	}
	response.OK(c, items)
}

// ---- Stocktaking ----

func ListStocktaking(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewStocktakingService(db)
	items, total, err := svc.List(queryInt(c, "page", 1), queryInt(c, "pageSize", 10))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("list stocktaking", err))
		return
	}
	response.Paged(c, items, &response.Page{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10), Total: total})
}

func GetStocktaking(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewStocktakingService(db)
	item, err := svc.Get(c.Param("id"))
	if err != nil {
		response.Error(c, apperrors.DatabaseError("get stocktaking", err))
		return
	}
	if item == nil {
		response.Error(c, apperrors.NotFound("stocktaking not found"))
		return
	}
	response.OK(c, item)
}

func CreateStocktaking(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var input service.StocktakingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperrors.ValidationFailed(err.Error()))
		return
	}

	svc := service.NewStocktakingService(db)
	item, err := svc.Create(input)
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.Created(c, item)
}

// ---- Dashboard ----

func GetDashboard(c *gin.Context) {
	db, err := getTenantDB(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	svc := service.NewDashboardService(db)
	data, err := svc.GetDashboard()
	if err != nil {
		response.Error(c, apperrors.Internal(err.Error()))
		return
	}
	response.OK(c, data)
}
