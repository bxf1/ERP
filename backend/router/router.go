package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/handler"
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
	}

	setupRAG(r, db, cfg)
	setupWorkflow(r, db)

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

func setupWorkflow(r *gin.Engine, db *gorm.DB) {
	wfRepo := repository.NewWorkflowRepository(db)
	wfSvc := service.NewWorkflowService(db, wfRepo)
	wfHandler := handler.NewWorkflowHandler(wfSvc)

	wf := r.Group("/api/v1/workflow")
	{
		// Definitions
		defs := wf.Group("/definitions")
		{
			defs.GET("", wfHandler.ListDefinitions)
			defs.POST("", wfHandler.CreateDefinition)
			defs.GET("/:id", wfHandler.GetDefinition)
			defs.PUT("/:id", wfHandler.UpdateDefinition)
			defs.DELETE("/:id", wfHandler.DeleteDefinition)

			// Nodes under a definition
			defs.POST("/:id/nodes", wfHandler.CreateNode)
		}

		// Nodes (independent routes)
		nodes := wf.Group("/nodes")
		{
			nodes.PUT("/:id", wfHandler.UpdateNode)
			nodes.DELETE("/:id", wfHandler.DeleteNode)
		}

		// Edges
		edges := wf.Group("/edges")
		{
			edges.PUT("/:id", wfHandler.UpdateEdge)
			edges.DELETE("/:id", wfHandler.DeleteEdge)
		}

		// Edges under a definition
		defs.POST("/:id/edges", wfHandler.CreateEdge)

		// Instances
		instances := wf.Group("/instances")
		{
			instances.GET("", wfHandler.ListInstances)
			instances.POST("", wfHandler.StartInstance)
			instances.GET("/:id", wfHandler.GetInstance)
			instances.POST("/:id/approve", wfHandler.ApproveInstance)
			instances.POST("/:id/reject", wfHandler.RejectInstance)
			instances.POST("/:id/transfer", wfHandler.TransferInstance)
			instances.POST("/:id/add-signer", wfHandler.AddSigner)
			instances.POST("/:id/cancel", wfHandler.CancelInstance)
		}
	}
}
