package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/api/handlers"
)

// RegisterMonitoringRoutes registers all monitoring-related routes
func RegisterMonitoringRoutes(router *gin.RouterGroup, handler *handlers.MonitoringHandler, authMiddleware gin.HandlerFunc) {
	// Workspace monitoring endpoints
	workspaces := router.Group("/workspaces/:workspace_id")
	workspaces.Use(authMiddleware)
	{
		workspaces.GET("/metrics", handler.GetMetrics)
		workspaces.GET("/health", handler.GetClusterHealth)
		workspaces.GET("/resources", handler.GetResourceUsage)
		workspaces.GET("/alerts", handler.GetAlerts)
		workspaces.POST("/alerts", handler.CreateAlert)
	}

	// Alert management endpoints
	alerts := router.Group("/alerts")
	alerts.Use(authMiddleware)
	{
		alerts.PUT("/:alert_id/acknowledge", handler.AcknowledgeAlert)
		alerts.PUT("/:alert_id/resolve", handler.ResolveAlert)
	}

	// Organization monitoring overview
	orgs := router.Group("/organizations/:org_id/monitoring")
	orgs.Use(authMiddleware)
	{
		orgs.GET("/overview", handler.GetWorkspaceOverview)
	}
}