package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/handler"
	intdb "github.com/bxf1/ERP/backend/internal/database"
	"github.com/bxf1/ERP/backend/internal/middleware"
	"github.com/bxf1/ERP/backend/pkg/database"
	"github.com/bxf1/ERP/backend/pkg/embedding"
	"github.com/bxf1/ERP/backend/repository"
	"github.com/bxf1/ERP/backend/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	gin.SetMode(gin.DebugMode)

	r := gin.New()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}
	tenantResolver := intdb.NewTenantResolver(sqlDB, cfg.Database)
	tenantMW := middleware.Tenant(middleware.TenantConfig{
		Resolver:      tenantResolver,
		HeaderName:    "X-Tenant",
		UsePathPrefix: false,
	})

	api := r.Group("/api/v1")
	{
		api.GET("/health", handler.Health)

		user := api.Group("/users")
		{
			user.GET("", handler.ListUsers)
			user.GET("/:id", handler.GetUser)
			user.POST("", handler.CreateUser)
			user.PUT("/:id", handler.UpdateUser)
			user.DELETE("/:id", handler.DeleteUser)
		}

		nl2sql := api.Group("/nl2sql")
		{
			nl2sql.POST("/query", handler.NL2SQLQuery)
			nl2sql.GET("/history", handler.NL2SQLHistory)
		}

		// Inventory / ERP business routes — all tenant-isolated
		biz := api.Group("")
		biz.Use(tenantMW)
		{
			suppliers := biz.Group("/suppliers")
			{
				suppliers.GET("", handler.ListSuppliers)
				suppliers.GET("/:id", handler.GetSupplier)
				suppliers.POST("", handler.CreateSupplier)
				suppliers.PUT("/:id", handler.UpdateSupplier)
			}

			customers := biz.Group("/customers")
			{
				customers.GET("", handler.ListCustomers)
				customers.GET("/:id", handler.GetCustomer)
				customers.POST("", handler.CreateCustomer)
				customers.PUT("/:id", handler.UpdateCustomer)
			}

			purchaseOrders := biz.Group("/purchase-orders")
			{
				purchaseOrders.GET("", handler.ListPurchaseOrders)
				purchaseOrders.GET("/:id", handler.GetPurchaseOrder)
				purchaseOrders.POST("", handler.CreatePurchaseOrder)
				purchaseOrders.PUT("/:id", handler.UpdatePurchaseOrder)
				purchaseOrders.POST("/:id/submit", handler.SubmitPurchaseOrder)
				purchaseOrders.POST("/:id/approve", handler.ApprovePurchaseOrder)
				purchaseOrders.POST("/:id/receive", handler.ReceivePurchaseOrder)
			}

			salesOrders := biz.Group("/sales-orders")
			{
				salesOrders.GET("", handler.ListSalesOrders)
				salesOrders.GET("/:id", handler.GetSalesOrder)
				salesOrders.POST("", handler.CreateSalesOrder)
				salesOrders.PUT("/:id", handler.UpdateSalesOrder)
				salesOrders.POST("/:id/confirm", handler.ConfirmSalesOrder)
				salesOrders.POST("/:id/ship", handler.ShipSalesOrder)
				salesOrders.POST("/:id/invoice", handler.InvoiceSalesOrder)
			}

			inventory := biz.Group("/inventory")
			{
				inventory.GET("", handler.ListInventory)
				inventory.GET("/alerts", handler.GetLowStockAlerts)
			}

			stocktaking := biz.Group("/stocktaking")
			{
				stocktaking.GET("", handler.ListStocktaking)
				stocktaking.GET("/:id", handler.GetStocktaking)
				stocktaking.POST("", handler.CreateStocktaking)
			}

			biz.GET("/dashboard", handler.GetDashboard)
		}
	}

	setupRAG(r, db, cfg)

	return r
}

func setupRAG(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	if err := database.EnablePgvector(db); err != nil {
		log.Printf("WARNING: failed to enable pgvector extension: %v", err)
		return
	}

	var embedder embedding.Provider
	switch cfg.Embedding.Provider {
	case "openai":
		embedder = embedding.NewOpenaiProvider(
			cfg.Embedding.APIKey,
			cfg.Embedding.BaseURL,
			cfg.Embedding.Model,
		)
	default:
		embedder = embedding.NewMockProvider(1536)
	}

	docRepo := repository.NewKnowledgeDocRepository(db)
	qaRepo := repository.NewKnowledgeQARepository(db)

	dim := embedder.Dimensions()
	if err := docRepo.AutoMigrate(dim); err != nil {
		log.Printf("WARNING: knowledge doc migration failed: %v", err)
		return
	}
	if err := qaRepo.AutoMigrate(dim); err != nil {
		log.Printf("WARNING: knowledge qa migration failed: %v", err)
		return
	}

	ragSvc := service.NewRAGService(cfg, embedder, docRepo, qaRepo)
	ragHandler := handler.NewRAGHandler(ragSvc)

	knowledge := r.Group("/api/v1/knowledge")
	{
		knowledge.POST("/documents", ragHandler.IngestDocument)
		knowledge.POST("/qa", ragHandler.IngestQA)
		knowledge.POST("/search", ragHandler.Search)
		knowledge.GET("/stats", ragHandler.Stats)
		knowledge.DELETE("/documents/:id", ragHandler.DeleteDocument)
		knowledge.DELETE("/qa/:id", ragHandler.DeleteQA)
	}
}
