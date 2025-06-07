package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/infrastructure/wire"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, app *wire.App) {
	// API version 1
	v1 := router.Group("/api/v1")

	// Authentication routes
	auth := router.Group("/auth")
	{
		auth.POST("/login/:provider", app.AuthHandler.Login)
		auth.GET("/callback/:provider", app.AuthHandler.Callback)
		auth.POST("/callback/:provider", app.AuthHandler.Callback) // For PKCE flow
		auth.POST("/refresh", app.AuthHandler.RefreshToken)
		auth.POST("/logout", app.AuthHandler.AuthMiddleware(), app.AuthHandler.Logout)
		auth.GET("/me", app.AuthHandler.AuthMiddleware(), app.AuthHandler.GetCurrentUser)

		// Session management
		auth.GET("/sessions", app.AuthHandler.AuthMiddleware(), app.AuthHandler.GetSessions)
		auth.DELETE("/sessions/:sessionId", app.AuthHandler.AuthMiddleware(), app.AuthHandler.RevokeSession)
		auth.POST("/sessions/revoke-all", app.AuthHandler.AuthMiddleware(), app.AuthHandler.RevokeAllSessions)

		// Security logs
		auth.GET("/security-logs", app.AuthHandler.AuthMiddleware(), app.AuthHandler.GetSecurityLogs)
	}

	// OIDC Discovery endpoints (public)
	router.GET("/.well-known/openid-configuration", app.AuthHandler.OIDCDiscovery)
	router.GET("/.well-known/jwks.json", app.AuthHandler.JWKS)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(app.AuthHandler.AuthMiddleware())

	// Organization routes
	orgs := protected.Group("/organizations")
	{
		orgs.POST("/", app.OrganizationHandler.CreateOrganization)
		orgs.GET("/", app.OrganizationHandler.ListOrganizations)
		orgs.GET("/:orgId", app.OrganizationHandler.GetOrganization)
		orgs.PUT("/:orgId", app.OrganizationHandler.UpdateOrganization)
		orgs.DELETE("/:orgId", app.OrganizationHandler.DeleteOrganization)

		// Organization members
		// orgs.POST("/:orgId/members", app.OrganizationHandler.AddMember) // Removed - use invitations instead
		orgs.GET("/:orgId/members", app.OrganizationHandler.ListMembers)
		orgs.DELETE("/:orgId/members/:userId", app.OrganizationHandler.RemoveMember)
		orgs.PUT("/:orgId/members/:userId/role", app.OrganizationHandler.UpdateMemberRole)

		// Organization invitations
		orgs.POST("/:orgId/invitations", app.OrganizationHandler.InviteUser)
		orgs.GET("/:orgId/invitations", app.OrganizationHandler.ListPendingInvitations)
		orgs.POST("/invitations/:token/accept", app.OrganizationHandler.AcceptInvitation)
		orgs.DELETE("/invitations/:invitationId", app.OrganizationHandler.CancelInvitation)

		// Organization activity
		orgs.GET("/:orgId/stats", app.OrganizationHandler.GetOrganizationStats)

		// Billing under organizations
		billing := orgs.Group("/:orgId/billing")
		{
			billing.GET("/subscription", app.BillingHandler.GetSubscription)
			billing.POST("/subscription", app.BillingHandler.CreateSubscription)
			billing.PUT("/subscription", app.BillingHandler.UpdateSubscription)
			billing.DELETE("/subscription", app.BillingHandler.CancelSubscription)

			billing.GET("/payment-methods", app.BillingHandler.ListPaymentMethods)
			billing.POST("/payment-methods", app.BillingHandler.AddPaymentMethod)
			billing.PUT("/payment-methods/:methodId/default", app.BillingHandler.SetDefaultPaymentMethod)
			billing.DELETE("/payment-methods/:methodId", app.BillingHandler.RemovePaymentMethod)

			billing.GET("/invoices", app.BillingHandler.ListInvoices)
			billing.GET("/invoices/upcoming", app.BillingHandler.GetUpcomingInvoice)
			billing.GET("/invoices/:invoiceId", app.BillingHandler.GetInvoice)
			billing.GET("/invoices/:invoiceId/download", app.BillingHandler.DownloadInvoice)

			billing.GET("/usage/current", app.BillingHandler.GetCurrentUsage)
			billing.GET("/overview", app.BillingHandler.GetBillingOverview)
			billing.GET("/settings", app.BillingHandler.GetBillingSettings)
			billing.PUT("/settings", app.BillingHandler.UpdateBillingSettings)
		}

		// Workspaces under organizations
		workspaces := orgs.Group("/:orgId/workspaces")
		{
			workspaces.POST("/", app.WorkspaceHandler.CreateWorkspace)
			workspaces.GET("/", app.WorkspaceHandler.ListWorkspaces)
			workspaces.GET("/:wsId", app.WorkspaceHandler.GetWorkspace)
			workspaces.PUT("/:wsId", app.WorkspaceHandler.UpdateWorkspace)
			workspaces.DELETE("/:wsId", app.WorkspaceHandler.DeleteWorkspace)
			workspaces.GET("/:wsId/kubeconfig", app.WorkspaceHandler.GetKubeconfig)
			workspaces.GET("/:wsId/resource-usage", app.WorkspaceHandler.GetResourceUsage)

			// Workspace members
			workspaces.POST("/:wsId/members", app.WorkspaceHandler.AddWorkspaceMember)
			workspaces.GET("/:wsId/members", app.WorkspaceHandler.ListWorkspaceMembers)
			workspaces.DELETE("/:wsId/members/:userId", app.WorkspaceHandler.RemoveWorkspaceMember)
		}
	}

	// Workspace-scoped routes
	workspaceScoped := protected.Group("/workspaces/:wsId")
	{
		// Projects
		projects := workspaceScoped.Group("/projects")
		{
			projects.POST("/", app.ProjectHandler.CreateProject)
			projects.GET("/", app.ProjectHandler.ListProjects)
			projects.GET("/:projectId", app.ProjectHandler.GetProject)
			projects.PUT("/:projectId", app.ProjectHandler.UpdateProject)
			projects.DELETE("/:projectId", app.ProjectHandler.DeleteProject)

			// Sub-projects
			projects.POST("/:projectId/subprojects", app.ProjectHandler.CreateSubProject)
			projects.GET("/:projectId/hierarchy", app.ProjectHandler.GetProjectHierarchy)

			// Resource management
			projects.POST("/:projectId/resource-quota", app.ProjectHandler.ApplyResourceQuota)
			projects.GET("/:projectId/resource-usage", app.ProjectHandler.GetResourceUsage)

			// Project members
			projects.POST("/:projectId/members", app.ProjectHandler.AddProjectMember)
			projects.GET("/:projectId/members", app.ProjectHandler.ListProjectMembers)
			projects.DELETE("/:projectId/members/:userId", app.ProjectHandler.RemoveProjectMember)

			// Activity logs
			projects.GET("/:projectId/activity", app.ProjectHandler.GetActivityLogs)
		}

		// Monitoring (workspace level)
		monitoring := workspaceScoped.Group("/monitoring")
		{
			monitoring.GET("/metrics", app.MonitoringHandler.GetMetrics)
			monitoring.GET("/health", app.MonitoringHandler.GetClusterHealth)
			monitoring.GET("/alerts", app.MonitoringHandler.GetAlerts)
			monitoring.POST("/alerts", app.MonitoringHandler.CreateAlert)
			monitoring.PUT("/alerts/:alertId/acknowledge", app.MonitoringHandler.AcknowledgeAlert)
			monitoring.PUT("/alerts/:alertId/resolve", app.MonitoringHandler.ResolveAlert)
		}
	}

	// Billing plans (public)
	plans := v1.Group("/plans")
	{
		plans.GET("/", app.BillingHandler.ListPlans)
		plans.GET("/compare", app.BillingHandler.ComparePlans)
	}

	// Webhook routes (no authentication required)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/stripe", app.BillingHandler.HandleStripeWebhook)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Catch-all for undefined routes
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})
}