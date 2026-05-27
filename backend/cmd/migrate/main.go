package main

import (
	"fmt"
	"os"

	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/internal/database"
	"github.com/bxf1/ERP/backend/internal/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	if err := logger.Init(cfg.Log); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := database.Connect(cfg.Database)
	if err != nil {
		logger.L.Fatal("connect database", zap.Error(err))
	}
	defer db.Close()

	migrationsDir := "migrations"
	if len(os.Args) > 1 {
		migrationsDir = os.Args[1]
	}

	logger.L.Info("running system migrations...")
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		logger.L.Fatal("system migrations failed", zap.Error(err))
	}

	logger.L.Info("running tenant migrations...")
	if err := database.MigrateTenantSchemas(db, cfg.Database); err != nil {
		logger.L.Fatal("tenant migrations failed", zap.Error(err))
	}

	logger.L.Info("all migrations complete")
}
