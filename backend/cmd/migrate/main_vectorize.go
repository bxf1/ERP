package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/pkg/database"
	"github.com/bxf1/ERP/backend/pkg/embedding"
	"github.com/bxf1/ERP/backend/repository"
	"github.com/bxf1/ERP/backend/service"
)

func main() {
	dir := flag.String("dir", "", "directory containing .md files to vectorize")
	flag.Parse()

	if *dir == "" {
		log.Fatal("--dir is required")
	}

	cfg := config.Load()

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("database connect: %v", err)
	}

	if err := database.EnablePgvector(db); err != nil {
		log.Fatalf("enable pgvector: %v", err)
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
		log.Fatalf("doc migration: %v", err)
	}
	if err := qaRepo.AutoMigrate(dim); err != nil {
		log.Fatalf("qa migration: %v", err)
	}

	svc := service.NewRAGService(cfg, embedder, docRepo, qaRepo)
	ctx := context.Background()

	count := 0
	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(*dir, path)
		title := strings.TrimSuffix(info.Name(), ".md")

		if err := svc.IngestDocument(ctx, service.IngestDocInput{
			Title:   title,
			Content: string(data),
			Source:  relPath,
		}); err != nil {
			return fmt.Errorf("ingest %s: %w", path, err)
		}

		count++
		fmt.Printf("ingested: %s\n", relPath)
		return nil
	})

	if err != nil {
		log.Fatalf("walk: %v", err)
	}

	fmt.Printf("\nDone. %d documents vectorized.\n", count)
}
