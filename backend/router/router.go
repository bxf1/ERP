package router

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/handler"
	"github.com/bxf1/ERP/backend/internal/middleware"
	"github.com/bxf1/ERP/backend/pkg/database"
	"github.com/bxf1/ERP/backend/pkg/embedding"
	permcache "github.com/bxf1/ERP/backend/pkg/permission/cache"
	permhandlers "github.com/bxf1/ERP/backend/pkg/permission/handlers"
	permservices "github.com/bxf1/ERP/backend/pkg/permission/services"
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
	setupPermissions(r, db, cfg)

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

func setupPermissions(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	permCache, err := permcache.NewPermissionCache(cfg.Redis.URL, 5*time.Minute)
	if err != nil {
		log.Printf("WARNING: permission cache init failed (redis may be unavailable): %v", err)
		return
	}

	rbacSvc := permservices.NewRBACService(db, permCache)
	roleHandler := permhandlers.NewRoleHandler(db, rbacSvc)
	permHandler := permhandlers.NewPermissionHandler(db, rbacSvc)
	userRoleHandler := permhandlers.NewUserRoleHandler(db, rbacSvc)

	api := r.Group("/api/v1")
	{
		roles := api.Group("/roles")
		{
			roles.GET("", roleHandler.ListRoles)
			roles.GET("/:id", roleHandler.GetRole)
			roles.POST("", roleHandler.CreateRole)
			roles.PUT("/:id", roleHandler.UpdateRole)
			roles.DELETE("/:id", roleHandler.DeleteRole)
			roles.POST("/:id/permissions", roleHandler.AssignPermissions)
			roles.POST("/:id/data-scope", roleHandler.SetDataScope)
		}

		permissions := api.Group("/permissions")
		{
			permissions.GET("", permHandler.ListPermissions)
			permissions.GET("/flat", permHandler.ListPermissionsFlat)
			permissions.POST("", permHandler.CreatePermission)
			permissions.PUT("/:id", permHandler.UpdatePermission)
			permissions.DELETE("/:id", permHandler.DeletePermission)
		}

		api.GET("/roles/available", userRoleHandler.ListAllRoles)
		api.GET("/users/:userId/roles", userRoleHandler.GetUserRoles)
		api.GET("/users/:userId/permissions", userRoleHandler.GetUserPermissions)
		api.POST("/users/:userId/roles", userRoleHandler.AssignRole)
		api.DELETE("/users/:userId/roles/:roleId", userRoleHandler.RemoveRole)
	}
}
