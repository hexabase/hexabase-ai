package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// WorkspaceTestSuite is the test suite for workspace handlers
type WorkspaceTestSuite struct {
	suite.Suite
	db       *gorm.DB
	handlers *Handlers
	router   *gin.Engine
	authUser *db.User
	authOrg  *db.Organization
	authToken string
	testPlan *db.Plan
}

// SetupSuite runs once before all tests
func (suite *WorkspaceTestSuite) SetupSuite() {
	// Setup test database with SQLite in-memory
	var err error
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	cfg := config.Config{
		Auth: config.AuthConfig{
			JWTSecret:     "test-secret-key-for-testing-only",
			JWTExpiration: 3600,
			OIDCIssuer:    "https://test.example.com",
		},
	}

	// Auto-migrate all models
	err = suite.db.AutoMigrate(
		&db.User{},
		&db.Organization{},
		&db.OrganizationUser{},
		&db.Plan{},
		&db.Workspace{},
		&db.Project{},
		&db.Group{},
		&db.GroupMembership{},
		&db.Role{},
		&db.RoleAssignment{},
		&db.VClusterProvisioningTask{},
		&db.StripeEvent{},
	)
	suite.Require().NoError(err)

	// Setup handlers and router
	logger, _ := zap.NewDevelopment()
	suite.handlers = NewHandlers(suite.db, &cfg, logger)
	
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup routes manually for testing to avoid auth service issues
	v1 := suite.router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(suite.testAuthMiddleware())
	
	// Organization routes with workspaces
	orgs := protected.Group("/organizations")
	workspaces := orgs.Group("/:orgId/workspaces")
	{
		workspaces.POST("/", suite.handlers.Workspaces.CreateWorkspace)
		workspaces.GET("/", suite.handlers.Workspaces.ListWorkspaces)
		workspaces.GET("/:wsId", suite.handlers.Workspaces.GetWorkspace)
		workspaces.PUT("/:wsId", suite.handlers.Workspaces.UpdateWorkspace)
		workspaces.DELETE("/:wsId", suite.handlers.Workspaces.DeleteWorkspace)
		workspaces.GET("/:wsId/kubeconfig", suite.handlers.Workspaces.GetKubeconfig)
	}
}

// SetupTest runs before each test
func (suite *WorkspaceTestSuite) SetupTest() {
	// Clean up database (SQLite doesn't support TRUNCATE CASCADE)
	suite.db.Exec("DELETE FROM v_cluster_provisioning_tasks")
	suite.db.Exec("DELETE FROM role_assignments")
	suite.db.Exec("DELETE FROM roles")
	suite.db.Exec("DELETE FROM group_memberships")
	suite.db.Exec("DELETE FROM groups")
	suite.db.Exec("DELETE FROM projects")
	suite.db.Exec("DELETE FROM workspaces")
	suite.db.Exec("DELETE FROM organization_users")
	suite.db.Exec("DELETE FROM organizations")
	suite.db.Exec("DELETE FROM users")
	suite.db.Exec("DELETE FROM plans")

	// Create test user
	suite.authUser = &db.User{
		ID:          "test-user-1",
		ExternalID:  "google-123456",
		Provider:    "google",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	suite.Require().NoError(suite.db.Create(suite.authUser).Error)

	// Create test organization
	suite.authOrg = &db.Organization{
		ID:   "test-org-1",
		Name: "Test Organization",
	}
	suite.Require().NoError(suite.db.Create(suite.authOrg).Error)

	// Link user to organization as admin
	orgUser := &db.OrganizationUser{
		OrganizationID: suite.authOrg.ID,
		UserID:         suite.authUser.ID,
		Role:           "admin",
		JoinedAt:       time.Now(),
	}
	suite.Require().NoError(suite.db.Create(orgUser).Error)

	// Create test plan
	suite.testPlan = &db.Plan{
		ID:            "plan-basic",
		Name:          "Basic Plan",
		Description:   "Basic workspace plan",
		Price:         9.99,
		Currency:      "USD",
		StripePriceID: "price_test123",
		ResourceLimits: `{
			"cpu": "4",
			"memory": "8Gi",
			"storage": "100Gi"
		}`,
		MaxProjectsPerWorkspace: intPtr(10),
		MaxMembersPerWorkspace:  intPtr(5),
		IsActive:                true,
	}
	suite.Require().NoError(suite.db.Create(suite.testPlan).Error)

	// Generate auth token (simplified for testing)
	suite.authToken = "Bearer test-token-" + suite.authUser.ID
}

// Helper function for int pointers
func intPtr(i int) *int {
	return &i
}

// Test workspace CRUD operations
func (suite *WorkspaceTestSuite) TestCreateWorkspace() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name: "successful workspace creation",
			payload: map[string]interface{}{
				"name":    "Development Workspace",
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing workspace name",
			payload: map[string]interface{}{
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace name is required",
		},
		{
			name: "missing plan_id",
			payload: map[string]interface{}{
				"name": "Development Workspace",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "plan_id is required",
		},
		{
			name: "invalid plan_id",
			payload: map[string]interface{}{
				"name":    "Development Workspace",
				"plan_id": "invalid-plan",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid plan selected",
		},
		{
			name: "workspace name too short",
			payload: map[string]interface{}{
				"name":    "ws",
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace name must be at least 3 characters",
		},
		{
			name: "duplicate workspace name in same org",
			payload: map[string]interface{}{
				"name":    "Existing Workspace",
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "workspace with this name already exists",
			setup: func() {
				// Create existing workspace
				ws := &db.Workspace{
					OrganizationID: suite.authOrg.ID,
					Name:           "Existing Workspace",
					PlanID:         suite.testPlan.ID,
					VClusterStatus: "RUNNING",
				}
				suite.Require().NoError(suite.db.Create(ws).Error)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/", suite.authOrg.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// Debug: print response for debugging
			if w.Code == 307 {
				suite.T().Logf("Unexpected redirect - Status: %d, Body: %s, Headers: %v", w.Code, w.Body.String(), w.Header())
			}

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			} else if tt.expectedStatus == http.StatusCreated {
				var workspace db.Workspace
				err := json.Unmarshal(w.Body.Bytes(), &workspace)
				suite.NoError(err)
				suite.NotEmpty(workspace.ID)
				suite.Equal("Development Workspace", workspace.Name)
				suite.Equal(suite.testPlan.ID, workspace.PlanID)
				suite.Equal("PENDING_CREATION", workspace.VClusterStatus)

				// Check if provisioning task was created
				var task db.VClusterProvisioningTask
				err = suite.db.Where("workspace_id = ?", workspace.ID).First(&task).Error
				suite.NoError(err)
				suite.Equal("CREATE", task.TaskType)
				suite.Equal("PENDING", task.Status)
			}
		})
	}
}

func (suite *WorkspaceTestSuite) TestListWorkspaces() {
	// Create test workspaces
	workspaces := []db.Workspace{
		{
			ID:                   "ws-1",
			OrganizationID:       suite.authOrg.ID,
			Name:                 "Production",
			PlanID:               suite.testPlan.ID,
			VClusterStatus:       "RUNNING",
			VClusterInstanceName: stringPtr("prod-cluster"),
		},
		{
			ID:             "ws-2",
			OrganizationID: suite.authOrg.ID,
			Name:           "Staging",
			PlanID:         suite.testPlan.ID,
			VClusterStatus: "PENDING_CREATION",
		},
	}

	for _, ws := range workspaces {
		suite.Require().NoError(suite.db.Create(&ws).Error)
	}

	// Create workspace in different org (should not be listed)
	otherOrg := &db.Organization{ID: "other-org", Name: "Other Org"}
	suite.Require().NoError(suite.db.Create(otherOrg).Error)
	
	otherWs := &db.Workspace{
		ID:             "ws-other",
		OrganizationID: otherOrg.ID,
		Name:           "Other Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(&otherWs).Error)

	// Test listing workspaces
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/", suite.authOrg.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Workspaces []db.Workspace `json:"workspaces"`
		Total      int           `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Total)
	suite.Len(response.Workspaces, 2)

	// Verify workspace details
	wsMap := make(map[string]db.Workspace)
	for _, ws := range response.Workspaces {
		wsMap[ws.ID] = ws
	}

	suite.Contains(wsMap, "ws-1")
	suite.Contains(wsMap, "ws-2")
	suite.NotContains(wsMap, "ws-other")
}

func (suite *WorkspaceTestSuite) TestGetWorkspace() {
	// Create test workspace
	ws := &db.Workspace{
		ID:                   "ws-test",
		OrganizationID:       suite.authOrg.ID,
		Name:                 "Test Workspace",
		PlanID:               suite.testPlan.ID,
		VClusterStatus:       "RUNNING",
		VClusterInstanceName: stringPtr("test-cluster"),
		VClusterConfig: `{
			"version": "v0.15.0",
			"values": {
				"syncer": {
					"extraArgs": ["--fake-kubelets=5"]
				}
			}
		}`,
	}
	suite.Require().NoError(suite.db.Create(ws).Error)

	tests := []struct {
		name           string
		workspaceID    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful get workspace",
			workspaceID:    "ws-test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "workspace not found",
			workspaceID:    "ws-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "workspace not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			} else {
				var workspace db.Workspace
				err := json.Unmarshal(w.Body.Bytes(), &workspace)
				suite.NoError(err)
				suite.Equal("ws-test", workspace.ID)
				suite.Equal("Test Workspace", workspace.Name)
				suite.NotEmpty(workspace.VClusterConfig)
			}
		})
	}
}

func (suite *WorkspaceTestSuite) TestUpdateWorkspace() {
	// Create test workspace
	ws := &db.Workspace{
		ID:             "ws-update",
		OrganizationID: suite.authOrg.ID,
		Name:           "Old Name",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(ws).Error)

	// Create upgraded plan
	upgradedPlan := &db.Plan{
		ID:            "plan-pro",
		Name:          "Pro Plan",
		Price:         29.99,
		Currency:      "USD",
		StripePriceID: "price_test456",
		ResourceLimits: `{
			"cpu": "8",
			"memory": "16Gi",
			"storage": "500Gi"
		}`,
		IsActive: true,
	}
	suite.Require().NoError(suite.db.Create(upgradedPlan).Error)

	tests := []struct {
		name           string
		workspaceID    string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful name update",
			workspaceID: "ws-update",
			payload: map[string]interface{}{
				"name": "New Name",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful plan upgrade",
			workspaceID: "ws-update",
			payload: map[string]interface{}{
				"plan_id": "plan-pro",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid plan downgrade",
			workspaceID: "ws-update",
			payload: map[string]interface{}{
				"plan_id": "plan-basic",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot downgrade plan",
		},
		{
			name:        "workspace not found",
			workspaceID: "ws-nonexistent",
			payload: map[string]interface{}{
				"name": "New Name",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "workspace not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s", suite.authOrg.ID, tt.workspaceID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			}
		})
	}
}

func (suite *WorkspaceTestSuite) TestDeleteWorkspace() {
	tests := []struct {
		name           string
		workspaceID    string
		expectedStatus int
		expectedError  string
		setup          func() *db.Workspace
	}{
		{
			name:           "successful delete empty workspace",
			workspaceID:    "ws-delete-empty",
			expectedStatus: http.StatusOK,
			setup: func() *db.Workspace {
				ws := &db.Workspace{
					ID:             "ws-delete-empty",
					OrganizationID: suite.authOrg.ID,
					Name:           "Empty Workspace",
					PlanID:         suite.testPlan.ID,
					VClusterStatus: "RUNNING",
				}
				suite.Require().NoError(suite.db.Create(ws).Error)
				return ws
			},
		},
		{
			name:           "cannot delete workspace with projects",
			workspaceID:    "ws-delete-projects",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot delete workspace with existing projects",
			setup: func() *db.Workspace {
				ws := &db.Workspace{
					ID:             "ws-delete-projects",
					OrganizationID: suite.authOrg.ID,
					Name:           "Workspace with Projects",
					PlanID:         suite.testPlan.ID,
					VClusterStatus: "RUNNING",
				}
				suite.Require().NoError(suite.db.Create(ws).Error)

				// Create project in workspace
				project := &db.Project{
					WorkspaceID: ws.ID,
					Name:        "Test Project",
				}
				suite.Require().NoError(suite.db.Create(project).Error)
				return ws
			},
		},
		{
			name:           "workspace not found",
			workspaceID:    "ws-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "workspace not found",
			setup:          func() *db.Workspace { return nil },
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ws := tt.setup()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			} else {
				// Verify workspace is marked for deletion
				if ws != nil {
					var updatedWs db.Workspace
					err := suite.db.First(&updatedWs, "id = ?", ws.ID).Error
					suite.NoError(err)
					suite.Equal("DELETING", updatedWs.VClusterStatus)

					// Check if deletion task was created
					var task db.VClusterProvisioningTask
					err = suite.db.Where("workspace_id = ? AND task_type = ?", ws.ID, "DELETE").First(&task).Error
					suite.NoError(err)
					suite.Equal("PENDING", task.Status)
				}
			}
		})
	}
}

func (suite *WorkspaceTestSuite) TestGetKubeconfig() {
	// Create test workspace
	ws := &db.Workspace{
		ID:                   "ws-kubeconfig",
		OrganizationID:       suite.authOrg.ID,
		Name:                 "Kubeconfig Test",
		PlanID:               suite.testPlan.ID,
		VClusterStatus:       "RUNNING",
		VClusterInstanceName: stringPtr("kubeconfig-test"),
	}
	suite.Require().NoError(suite.db.Create(ws).Error)

	tests := []struct {
		name           string
		workspaceID    string
		expectedStatus int
		expectedError  string
		setupStatus    string
	}{
		{
			name:           "successful get kubeconfig",
			workspaceID:    "ws-kubeconfig",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "workspace not ready",
			workspaceID:    "ws-pending",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace is not ready",
			setupStatus:    "PENDING_CREATION",
		},
		{
			name:           "workspace not found",
			workspaceID:    "ws-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "workspace not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setupStatus != "" {
				pendingWs := &db.Workspace{
					ID:             "ws-pending",
					OrganizationID: suite.authOrg.ID,
					Name:           "Pending Workspace",
					PlanID:         suite.testPlan.ID,
					VClusterStatus: tt.setupStatus,
				}
				suite.Require().NoError(suite.db.Create(pendingWs).Error)
			}

			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/kubeconfig", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response, "kubeconfig")
				suite.NotEmpty(response["kubeconfig"])
			}
		})
	}
}

// Test workspace authorization
func (suite *WorkspaceTestSuite) TestWorkspaceAuthorization() {
	// Create another user and organization
	otherUser := &db.User{
		ID:          "other-user",
		ExternalID:  "google-999999",
		Provider:    "google",
		Email:       "other@example.com",
		DisplayName: "Other User",
	}
	suite.Require().NoError(suite.db.Create(otherUser).Error)

	otherOrg := &db.Organization{
		ID:   "other-org",
		Name: "Other Organization",
	}
	suite.Require().NoError(suite.db.Create(otherOrg).Error)

	// Link other user to other org
	orgUser := &db.OrganizationUser{
		OrganizationID: otherOrg.ID,
		UserID:         otherUser.ID,
		Role:           "admin",
		JoinedAt:       time.Now(),
	}
	suite.Require().NoError(suite.db.Create(orgUser).Error)

	// Create workspace in our org
	ws := &db.Workspace{
		ID:             "ws-auth-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "Auth Test Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(ws).Error)

	// Try to access workspace with other user's token
	otherToken := "Bearer test-token-" + otherUser.ID

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s", suite.authOrg.ID, ws.ID), nil)
	req.Header.Set("Authorization", otherToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"], "not authorized")
}

// testAuthMiddleware creates a mock auth middleware for testing
func (suite *WorkspaceTestSuite) testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Simple mock authentication - extract user ID from token
		if strings.HasPrefix(authHeader, "Bearer test-token-") {
			userID := strings.TrimPrefix(authHeader, "Bearer test-token-")
			
			// Verify user exists and get their org memberships
			var user db.User
			if err := suite.db.First(&user, "id = ?", userID).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				c.Abort()
				return
			}

			var orgMemberships []db.OrganizationUser
			suite.db.Where("user_id = ?", userID).Find(&orgMemberships)

			var orgIDs []string
			for _, membership := range orgMemberships {
				orgIDs = append(orgIDs, membership.OrganizationID)
			}

			c.Set("user_id", userID)
			c.Set("user_email", user.Email)
			c.Set("user_name", user.DisplayName)
			c.Set("org_ids", orgIDs)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// TestWorkspaceSuite runs the test suite
func TestWorkspaceSuite(t *testing.T) {
	suite.Run(t, new(WorkspaceTestSuite))
}