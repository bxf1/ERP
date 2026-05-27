package main

import (
	"log"
	"os"

	"github.com/bxf1/ERP/backend/internal/gateway"
	"github.com/bxf1/ERP/backend/internal/mcp"
	"github.com/bxf1/ERP/backend/internal/model"
	"github.com/bxf1/ERP/backend/internal/semantic"
)

func main() {
	// Initialize core components.
	repo := model.NewRepository()
	modelSvc := model.NewService(repo)
	auditLogger := gateway.NewAuditLogger()
	secGateway := gateway.NewSecurityGateway(auditLogger)
	semLayer := semantic.DefaultLayer()

	// Seed default permissions for the system agent.
	secGateway.GrantPermission("system", gateway.PermReadModels)
	secGateway.GrantPermission("system", gateway.PermWriteModels)
	secGateway.GrantPermission("system", gateway.PermQueryData)
	secGateway.GrantPermission("system", gateway.PermReadSemantic)

	// Create the MCP server.
	server := mcp.NewServer(modelSvc, secGateway, semLayer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[ERP-MCP] Starting MCP interface server on :%s", port)
	log.Printf("[ERP-MCP] Registered %d tools for LLM Function Calling", len(mcp.ToolDefinitions()))

	if err := server.Run(":" + port); err != nil {
		log.Fatalf("[ERP-MCP] Failed to start server: %v", err)
	}
}
