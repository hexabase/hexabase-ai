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

// ProjectTestSuite is the test suite for project handlers
type ProjectTestSuite struct {
	suite.Suite
	db          *gorm.DB
	handlers    *Handlers
	router      *gin.Engine
	authUser    *db.User
	authOrg     *db.Organization
	authToken   string
	testPlan    *db.Plan
	testWs      *db.Workspace
}

// SetupSuite runs once before all tests
func (suite *ProjectTestSuite) SetupSuite() {
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
	
	// Organization routes with workspaces and projects
	orgs := protected.Group("/organizations")
	workspaces := orgs.Group("/:orgId/workspaces")
	projects := workspaces.Group("/:wsId/projects")
	{
		projects.POST("/", suite.handlers.Projects.CreateProject)
		projects.GET("/", suite.handlers.Projects.ListProjects)
		projects.GET("/:projectId", suite.handlers.Projects.GetProject)
		projects.PUT("/:projectId", suite.handlers.Projects.UpdateProject)
		projects.DELETE("/:projectId", suite.handlers.Projects.DeleteProject)
		
		// Role management within projects
		projectRoles := projects.Group("/:projectId/roles")
		{
			projectRoles.POST("/", suite.handlers.Projects.CreateRole)
			projectRoles.GET("/", suite.handlers.Projects.ListRoles)
			projectRoles.GET("/:roleId", suite.handlers.Projects.GetRole)
			projectRoles.PUT("/:roleId", suite.handlers.Projects.UpdateRole)
			projectRoles.DELETE("/:roleId", suite.handlers.Projects.DeleteRole)
		}
		
		// Role assignments within projects
		projectAssignments := projects.Group("/:projectId/role-assignments")
		{
			projectAssignments.POST("/", suite.handlers.Projects.CreateRoleAssignment)
			projectAssignments.GET("/", suite.handlers.Projects.ListRoleAssignments)
			projectAssignments.DELETE("/:assignmentId", suite.handlers.Projects.DeleteRoleAssignment)
		}
	}
}

// SetupTest runs before each test
func (suite *ProjectTestSuite) SetupTest() {
	// Clean up database (SQLite doesn't support TRUNCATE CASCADE)
	suite.db.Exec("DELETE FROM role_assignments")
	suite.db.Exec("DELETE FROM roles")
	suite.db.Exec("DELETE FROM group_memberships")
	suite.db.Exec("DELETE FROM groups")
	suite.db.Exec("DELETE FROM projects")
	suite.db.Exec("DELETE FROM v_cluster_provisioning_tasks")
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

	// Create test workspace
	suite.testWs = &db.Workspace{
		ID:             "test-ws-1",
		OrganizationID: suite.authOrg.ID,
		Name:           "Test Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
		VClusterConfig: "{}",
		DedicatedNodeConfig: "{}",
	}
	suite.Require().NoError(suite.db.Create(suite.testWs).Error)

	// Generate auth token (simplified for testing)
	suite.authToken = "Bearer test-token-" + suite.authUser.ID
}

// Helper function for int pointers (moved to avoid duplication)

// Test project CRUD operations
func (suite *ProjectTestSuite) TestCreateProject() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name: "successful project creation",
			payload: map[string]interface{}{
				"name":        "development-project",
				"description": "Development environment project",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing project name",
			payload: map[string]interface{}{
				"description": "Project without name",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "project name is required",
		},
		{
			name: "project name too short",
			payload: map[string]interface{}{
				"name": "pr",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "project name must be at least 3 characters",
		},
		{
			name: "invalid project name (special characters)",
			payload: map[string]interface{}{
				"name": "project-with-UPPERCASE",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "project name must be lowercase alphanumeric with hyphens",
		},
		{
			name: "duplicate project name in same workspace",
			payload: map[string]interface{}{
				"name": "existing-project",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "project with this name already exists",
			setup: func() {
				// Create existing project
				proj := &db.Project{
					WorkspaceID: suite.testWs.ID,
					Name:        "existing-project",
					Description: "Existing project",
					NamespaceStatus: "ACTIVE",
				}
				suite.Require().NoError(suite.db.Create(proj).Error)
			},
		},
		{
			name: "exceed project limit per workspace",
			payload: map[string]interface{}{
				"name": "project-over-limit",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "workspace has reached maximum number of projects",
			setup: func() {
				// Create projects up to the limit (10 projects)
				for i := 0; i < 10; i++ {
					proj := &db.Project{
						WorkspaceID: suite.testWs.ID,
						Name:        fmt.Sprintf("project-%d", i),
						Description: fmt.Sprintf("Project %d", i),
						NamespaceStatus: "ACTIVE",
					}
					suite.Require().NoError(suite.db.Create(proj).Error)
				}
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/", suite.authOrg.ID, suite.testWs.ID), bytes.NewBuffer(body))
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
			} else if tt.expectedStatus == http.StatusCreated {
				var project db.Project
				err := json.Unmarshal(w.Body.Bytes(), &project)
				suite.NoError(err)
				suite.NotEmpty(project.ID)
				suite.Equal("development-project", project.Name)
				suite.Equal("Development environment project", project.Description)
				suite.Equal("PENDING_CREATION", project.NamespaceStatus)
			}
		})
	}
}

func (suite *ProjectTestSuite) TestListProjects() {
	// Create test projects
	projects := []db.Project{
		{
			ID:          "proj-1",
			WorkspaceID: suite.testWs.ID,
			Name:        "frontend",
			Description: "Frontend application project",
			NamespaceStatus: "ACTIVE",
		},
		{
			ID:          "proj-2",
			WorkspaceID: suite.testWs.ID,
			Name:        "backend",
			Description: "Backend API project",
			NamespaceStatus: "ACTIVE",
		},
	}

	for _, proj := range projects {
		suite.Require().NoError(suite.db.Create(&proj).Error)
	}

	// Create project in different workspace (should not be listed)
	otherWs := &db.Workspace{
		ID:             "other-ws",
		OrganizationID: suite.authOrg.ID,
		Name:           "Other Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(otherWs).Error)
	
	otherProj := &db.Project{
		ID:          "proj-other",
		WorkspaceID: otherWs.ID,
		Name:        "other-project",
		Description: "Project in other workspace",
		NamespaceStatus: "ACTIVE",
	}
	suite.Require().NoError(suite.db.Create(&otherProj).Error)

	// Test listing projects
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/", suite.authOrg.ID, suite.testWs.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Projects []db.Project `json:"projects"`
		Total    int          `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Total)
	suite.Len(response.Projects, 2)

	// Verify project details
	projMap := make(map[string]db.Project)
	for _, proj := range response.Projects {
		projMap[proj.ID] = proj
	}

	suite.Contains(projMap, "proj-1")
	suite.Contains(projMap, "proj-2")
	suite.NotContains(projMap, "proj-other")
}

func (suite *ProjectTestSuite) TestGetProject() {
	// Create test project
	proj := &db.Project{
		ID:          "proj-test",
		WorkspaceID: suite.testWs.ID,
		Name:        "test-project",
		Description: "Test project description",
		NamespaceStatus: "ACTIVE",
		KubernetesNamespace: stringPtr("test-project-ns"),
	}
	suite.Require().NoError(suite.db.Create(proj).Error)

	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful get project",
			projectID:      "proj-test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "project not found",
			projectID:      "proj-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "project not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/%s", suite.authOrg.ID, suite.testWs.ID, tt.projectID), nil)
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
				var project db.Project
				err := json.Unmarshal(w.Body.Bytes(), &project)
				suite.NoError(err)
				suite.Equal("proj-test", project.ID)
				suite.Equal("test-project", project.Name)
				suite.Equal("Test project description", project.Description)
			}
		})
	}
}

func (suite *ProjectTestSuite) TestUpdateProject() {
	// Create test project
	proj := &db.Project{
		ID:          "proj-update",
		WorkspaceID: suite.testWs.ID,
		Name:        "old-name",
		Description: "Old description",
		NamespaceStatus: "ACTIVE",
	}
	suite.Require().NoError(suite.db.Create(proj).Error)

	tests := []struct {
		name           string
		projectID      string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful description update",
			projectID: "proj-update",
			payload: map[string]interface{}{
				"description": "New description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "project not found",
			projectID: "proj-nonexistent",
			payload: map[string]interface{}{
				"description": "New description",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "project not found",
		},
		{
			name:      "invalid name update (names are immutable)",
			projectID: "proj-update",
			payload: map[string]interface{}{
				"name": "new-name",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "project name cannot be changed",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/%s", suite.authOrg.ID, suite.testWs.ID, tt.projectID), bytes.NewBuffer(body))
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

func (suite *ProjectTestSuite) TestDeleteProject() {
	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
		expectedError  string
		setup          func() *db.Project
	}{
		{
			name:           "successful delete empty project",
			projectID:      "proj-delete-empty",
			expectedStatus: http.StatusOK,
			setup: func() *db.Project {
				proj := &db.Project{
					ID:          "proj-delete-empty",
					WorkspaceID: suite.testWs.ID,
					Name:        "empty-project",
					Description: "Empty project to delete",
					NamespaceStatus: "ACTIVE",
				}
				suite.Require().NoError(suite.db.Create(proj).Error)
				return proj
			},
		},
		{
			name:           "cannot delete project with roles",
			projectID:      "proj-delete-roles",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot delete project with existing roles",
			setup: func() *db.Project {
				proj := &db.Project{
					ID:          "proj-delete-roles",
					WorkspaceID: suite.testWs.ID,
					Name:        "project-with-roles",
					Description: "Project with roles",
					NamespaceStatus: "ACTIVE",
				}
				suite.Require().NoError(suite.db.Create(proj).Error)

				// Create role in project
				role := &db.Role{
					ProjectID:   &proj.ID,
					Name:        "test-role",
					Description: "Test role",
					Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list"]}]`,
				}
				suite.Require().NoError(suite.db.Create(role).Error)
				return proj
			},
		},
		{
			name:           "project not found",
			projectID:      "proj-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "project not found",
			setup:          func() *db.Project { return nil },
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			proj := tt.setup()

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/%s", suite.authOrg.ID, suite.testWs.ID, tt.projectID), nil)
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
				// Verify project is marked for deletion
				if proj != nil {
					var updatedProj db.Project
					err := suite.db.First(&updatedProj, "id = ?", proj.ID).Error
					suite.NoError(err)
					suite.Equal("DELETING", updatedProj.NamespaceStatus)
				}
			}
		})
	}
}

// Test project authorization
func (suite *ProjectTestSuite) TestProjectAuthorization() {
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

	// Create project in our workspace
	proj := &db.Project{
		ID:          "proj-auth-test",
		WorkspaceID: suite.testWs.ID,
		Name:        "auth-test-project",
		Description: "Auth test project",
		NamespaceStatus: "ACTIVE",
	}
	suite.Require().NoError(suite.db.Create(proj).Error)

	// Try to access project with other user's token
	otherToken := "Bearer test-token-" + otherUser.ID

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/%s", suite.authOrg.ID, suite.testWs.ID, proj.ID), nil)
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
func (suite *ProjectTestSuite) testAuthMiddleware() gin.HandlerFunc {
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

// Helper function for string pointers (moved to avoid duplication)

// TestProjectSuite runs the test suite
func TestProjectSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}