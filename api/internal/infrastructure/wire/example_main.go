// Example of how to use Wire in main.go
// This file shows how to integrate the refactored code

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/api/routes"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/hexabase/kaas-api/internal/infrastructure/wire"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Initialize database
	database, err := db.NewConnection(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		logger.Fatal("Failed to get Kubernetes config", zap.Error(err))
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", zap.Error(err))
	}

	// Initialize app with Wire
	app, err := wire.InitializeApp(cfg, database, k8sClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize app", zap.Error(err))
	}

	// Setup router
	router := gin.Default()
	api := router.Group("/api/v1")

	// Register routes with dependency-injected handlers
	routes.RegisterMonitoringRoutes(api, app.MonitoringHandler, authMiddleware)
	
	// Add other routes as they are refactored
	// routes.RegisterWorkspaceRoutes(api, app.WorkspaceHandler, authMiddleware)
	// routes.RegisterBillingRoutes(api, app.BillingHandler, authMiddleware)

	// Start server
	if err := router.Run(cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// Placeholder for auth middleware
func authMiddleware(c *gin.Context) {
	// Implementation
}