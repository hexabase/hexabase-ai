package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, handlers *Handlers) {
	// API version 1
	v1 := router.Group("/api/v1")
	
	// Authentication routes
	auth := router.Group("/auth")
	{
		auth.POST("/login/:provider", handlers.Auth.LoginProvider)
		auth.GET("/callback/:provider", handlers.Auth.CallbackProvider)
		auth.POST("/logout", handlers.Auth.Logout)
		auth.GET("/me", handlers.Auth.AuthMiddleware(), handlers.Auth.GetCurrentUser)
	}

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(handlers.Auth.AuthMiddleware())
	
	// Organization routes
	orgs := protected.Group("/organizations")
	{
		orgs.POST("/", handlers.Organizations.CreateOrganization)
		orgs.GET("/", handlers.Organizations.ListOrganizations)
		orgs.GET("/:orgId", handlers.Organizations.GetOrganization)
		orgs.PUT("/:orgId", handlers.Organizations.UpdateOrganization)
		orgs.DELETE("/:orgId", handlers.Organizations.DeleteOrganization)
		
		// Organization users
		orgs.POST("/:orgId/users", handlers.Organizations.InviteUser)
		orgs.GET("/:orgId/users", handlers.Organizations.ListUsers)
		orgs.DELETE("/:orgId/users/:userId", handlers.Organizations.RemoveUser)
		
		// Organization billing
		orgs.POST("/:orgId/billing/portal-session", handlers.Organizations.CreatePortalSession)
		orgs.GET("/:orgId/billing/subscriptions", handlers.Organizations.GetSubscriptions)
		orgs.GET("/:orgId/billing/payment-methods", handlers.Organizations.GetPaymentMethods)
		orgs.POST("/:orgId/billing/payment-methods/setup-intent", handlers.Organizations.CreateSetupIntent)
		
		// Workspaces under organizations
		workspaces := orgs.Group("/:orgId/workspaces")
		{
			workspaces.POST("/", handlers.Workspaces.CreateWorkspace)
			workspaces.GET("/", handlers.Workspaces.ListWorkspaces)
			workspaces.GET("/:wsId", handlers.Workspaces.GetWorkspace)
			workspaces.PUT("/:wsId", handlers.Workspaces.UpdateWorkspace)
			workspaces.DELETE("/:wsId", handlers.Workspaces.DeleteWorkspace)
			workspaces.GET("/:wsId/kubeconfig", handlers.Workspaces.GetKubeconfig)
		}
	}

	// Workspace-scoped routes
	workspaceScoped := protected.Group("/workspaces/:wsId")
	{
		// Groups
		groups := workspaceScoped.Group("/groups")
		{
			groups.POST("/", handlers.Workspaces.CreateGroup)
			groups.GET("/", handlers.Workspaces.ListGroups)
			groups.GET("/:groupId", handlers.Workspaces.GetGroup)
			groups.PUT("/:groupId", handlers.Workspaces.UpdateGroup)
			groups.DELETE("/:groupId", handlers.Workspaces.DeleteGroup)
			groups.POST("/:groupId/members", handlers.Workspaces.AddGroupMember)
			groups.DELETE("/:groupId/members/:userId", handlers.Workspaces.RemoveGroupMember)
		}
		
		// Projects
		projects := workspaceScoped.Group("/projects")
		{
			projects.POST("/", handlers.Projects.CreateProject)
			projects.GET("/", handlers.Projects.ListProjects)
			projects.GET("/:projectId", handlers.Projects.GetProject)
			projects.PUT("/:projectId", handlers.Projects.UpdateProject)
			projects.DELETE("/:projectId", handlers.Projects.DeleteProject)
		}
		
		// Cluster role assignments
		workspaceScoped.POST("/clusterroleassignments", handlers.Workspaces.CreateClusterRoleAssignment)
		workspaceScoped.GET("/clusterroleassignments", handlers.Workspaces.ListClusterRoleAssignments)
		workspaceScoped.DELETE("/clusterroleassignments/:assignmentId", handlers.Workspaces.DeleteClusterRoleAssignment)
	}

	// Project-scoped routes
	projectScoped := protected.Group("/projects/:projectId")
	{
		// Roles
		roles := projectScoped.Group("/roles")
		{
			roles.POST("/", handlers.Projects.CreateRole)
			roles.GET("/", handlers.Projects.ListRoles)
			roles.GET("/:roleId", handlers.Projects.GetRole)
			roles.PUT("/:roleId", handlers.Projects.UpdateRole)
			roles.DELETE("/:roleId", handlers.Projects.DeleteRole)
		}
		
		// Role assignments
		projectScoped.POST("/roleassignments", handlers.Projects.CreateRoleAssignment)
		projectScoped.GET("/roleassignments", handlers.Projects.ListRoleAssignments)
		projectScoped.DELETE("/roleassignments/:assignmentId", handlers.Projects.DeleteRoleAssignment)
	}

	// Webhook routes (no authentication required)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/stripe", handlers.Webhooks.HandleStripeWebhook)
	}

	// OIDC provider routes
	oidc := router.Group("/.well-known")
	{
		oidc.GET("/openid-configuration", handlers.Auth.OIDCDiscovery)
		oidc.GET("/jwks.json", handlers.Auth.JWKS)
	}

	// Catch-all for undefined routes
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})
}