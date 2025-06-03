package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// RBACTestSuite is the test suite for RBAC handlers
type RBACTestSuite struct {
	suite.Suite
	db            *gorm.DB
	handlers      *Handlers
	router        *gin.Engine
	authUser      *db.User
	authOrg       *db.Organization
	authToken     string
	testWorkspace *db.Workspace
	testProject   *db.Project
	testGroup     *db.Group
}

// SetupSuite runs once before all tests
func (suite *RBACTestSuite) SetupSuite() {
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
		&db.RoleBinding{},
		&db.RoleAssignment{},
	)
	suite.Require().NoError(err)

	// Setup handlers and router
	logger, _ := zap.NewDevelopment()
	suite.handlers = NewHandlers(suite.db, &cfg, logger)
	
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	
	// Setup routes
	v1 := suite.router.Group("/api/v1")
	protected := v1.Group("")
	protected.Use(suite.testAuthMiddleware())
	
	// RBAC routes
	orgs := protected.Group("/organizations")
	workspaces := orgs.Group("/:orgId/workspaces/:wsId")
	rbac := workspaces.Group("/rbac")
	{
		// Role management
		rbac.GET("/roles", suite.handlers.RBAC.ListRoles)
		rbac.POST("/roles", suite.handlers.RBAC.CreateRole)
		rbac.GET("/roles/:roleId", suite.handlers.RBAC.GetRole)
		rbac.PUT("/roles/:roleId", suite.handlers.RBAC.UpdateRole)
		rbac.DELETE("/roles/:roleId", suite.handlers.RBAC.DeleteRole)
		
		// RoleBinding management
		rbac.GET("/rolebindings", suite.handlers.RBAC.ListRoleBindings)
		rbac.POST("/rolebindings", suite.handlers.RBAC.CreateRoleBinding)
		rbac.GET("/rolebindings/:bindingId", suite.handlers.RBAC.GetRoleBinding)
		rbac.DELETE("/rolebindings/:bindingId", suite.handlers.RBAC.DeleteRoleBinding)
		
		// Permission checking
		rbac.POST("/permissions/check", suite.handlers.RBAC.CheckPermissions)
		
		// Project-specific RBAC
		projects := workspaces.Group("/projects/:projectId/rbac")
		projects.GET("/roles", suite.handlers.RBAC.ListProjectRoles)
		projects.POST("/roles", suite.handlers.RBAC.CreateProjectRole)
		projects.GET("/rolebindings", suite.handlers.RBAC.ListProjectRoleBindings)
		projects.POST("/rolebindings", suite.handlers.RBAC.CreateProjectRoleBinding)
	}
}

// SetupTest runs before each test
func (suite *RBACTestSuite) SetupTest() {
	// Clean up database
	suite.db.Exec("DELETE FROM role_bindings")
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
	plan := &db.Plan{
		ID:            "plan-test",
		Name:          "Test Plan",
		Description:   "Test plan for RBAC",
		Price:         9.99,
		Currency:      "USD",
		StripePriceID: "price_test",
		IsActive:      true,
	}
	suite.Require().NoError(suite.db.Create(plan).Error)

	// Create test workspace
	suite.testWorkspace = &db.Workspace{
		ID:             "ws-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "Test Workspace",
		PlanID:         "plan-test",
		VClusterStatus: "RUNNING",
	}
	suite.Require().NoError(suite.db.Create(suite.testWorkspace).Error)

	// Create test project
	suite.testProject = &db.Project{
		ID:              "proj-test",
		WorkspaceID:     suite.testWorkspace.ID,
		Name:            "Test Project",
		Description:     "Test project for RBAC",
		NamespaceStatus: "ACTIVE",
	}
	suite.Require().NoError(suite.db.Create(suite.testProject).Error)

	// Create test group
	suite.testGroup = &db.Group{
		ID:          "group-test",
		WorkspaceID: suite.testWorkspace.ID,
		Name:        "Test Group",
	}
	suite.Require().NoError(suite.db.Create(suite.testGroup).Error)

	// Generate auth token
	suite.authToken = "Bearer test-token-" + suite.authUser.ID
}

// Test role management
func (suite *RBACTestSuite) TestCreateRole() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful namespace-scoped role creation",
			payload: map[string]interface{}{
				"name":        "pod-reader",
				"description": "Read access to pods",
				"scope":       "namespace",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{""},
						"resources": []string{"pods"},
						"verbs":     []string{"get", "list", "watch"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful cluster-scoped role creation",
			payload: map[string]interface{}{
				"name":        "node-reader",
				"description": "Read access to nodes",
				"scope":       "cluster",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{""},
						"resources": []string{"nodes"},
						"verbs":     []string{"get", "list"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "custom service account role",
			payload: map[string]interface{}{
				"name":        "sa-manager",
				"description": "Manage service accounts",
				"scope":       "namespace",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{""},
						"resources": []string{"serviceaccounts"},
						"verbs":     []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing name",
			payload:        map[string]interface{}{"scope": "namespace"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "invalid scope",
			payload: map[string]interface{}{
				"name":  "invalid-scope",
				"scope": "invalid",
				"rules": []map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid scope",
		},
		{
			name: "missing rules",
			payload: map[string]interface{}{
				"name":  "no-rules",
				"scope": "namespace",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "rules are required",
		},
		{
			name: "duplicate role name",
			payload: map[string]interface{}{
				"name":  "pod-reader",
				"scope": "namespace",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{""},
						"resources": []string{"pods"},
						"verbs":     []string{"get"},
					},
				},
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "role already exists",
		},
	}

	for i, tt := range tests {
		suite.Run(tt.name, func() {
			// For duplicate test, create the role first
			if i == 6 {
				role := &db.Role{
					Name:        "pod-reader",
					WorkspaceID: &suite.testWorkspace.ID,
					Scope:       "namespace",
					Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get"]}]`,
				}
				suite.db.Create(role)
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/rbac/roles", suite.authOrg.ID, suite.testWorkspace.ID), bytes.NewBuffer(body))
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

func (suite *RBACTestSuite) TestListRoles() {
	// Create test roles
	roles := []db.Role{
		{
			Name:        "pod-reader",
			WorkspaceID: &suite.testWorkspace.ID,
			Scope:       "namespace",
			Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list"]}]`,
			IsActive:    true,
		},
		{
			Name:        "deployment-manager",
			WorkspaceID: &suite.testWorkspace.ID,
			Scope:       "namespace", 
			Rules:       `[{"apiGroups":["apps"],"resources":["deployments"],"verbs":["*"]}]`,
			IsActive:    true,
		},
		{
			Name:        "cluster-admin",
			WorkspaceID: &suite.testWorkspace.ID,
			Scope:       "cluster",
			Rules:       `[{"apiGroups":["*"],"resources":["*"],"verbs":["*"]}]`,
			IsActive:    true,
		},
	}
	for _, role := range roles {
		suite.Require().NoError(suite.db.Create(&role).Error)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/rbac/roles", suite.authOrg.ID, suite.testWorkspace.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Roles []db.Role `json:"roles"`
		Total int        `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(3, response.Total)
	suite.Len(response.Roles, 3)
}

// Test role binding management
func (suite *RBACTestSuite) TestCreateRoleBinding() {
	// Create test role
	role := &db.Role{
		ID:          "role-test",
		Name:        "pod-reader",
		WorkspaceID: &suite.testWorkspace.ID,
		Scope:       "namespace",
		Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list"]}]`,
		IsActive:    true,
	}
	suite.Require().NoError(suite.db.Create(role).Error)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user role binding",
			payload: map[string]interface{}{
				"role_id":      "role-test",
				"subject_type": "User",
				"subject_id":   suite.authUser.ID,
				"subject_name": suite.authUser.Email,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "successful group role binding",
			payload: map[string]interface{}{
				"role_id":      "role-test",
				"subject_type": "Group",
				"subject_id":   suite.testGroup.ID,
				"subject_name": suite.testGroup.Name,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "project-scoped role binding",
			payload: map[string]interface{}{
				"role_id":      "role-test",
				"project_id":   suite.testProject.ID,
				"subject_type": "User",
				"subject_id":   suite.authUser.ID,
				"subject_name": suite.authUser.Email,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing role_id",
			payload:        map[string]interface{}{"subject_type": "User"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "role_id is required",
		},
		{
			name: "invalid subject_type",
			payload: map[string]interface{}{
				"role_id":      "role-test",
				"subject_type": "Invalid",
				"subject_id":   "test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid subject_type",
		},
		{
			name: "role not found",
			payload: map[string]interface{}{
				"role_id":      "role-nonexistent",
				"subject_type": "User",
				"subject_id":   suite.authUser.ID,
				"subject_name": suite.authUser.Email,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "role not found",
		},
		{
			name: "duplicate role binding",
			payload: map[string]interface{}{
				"role_id":      "role-test",
				"subject_type": "User",
				"subject_id":   suite.authUser.ID,
				"subject_name": suite.authUser.Email,
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "role binding already exists",
		},
	}

	for i, tt := range tests {
		suite.Run(tt.name, func() {
			// For duplicate test, create the binding first
			if i == 6 {
				binding := &db.RoleBinding{
					WorkspaceID: suite.testWorkspace.ID,
					RoleID:      "role-test",
					SubjectType: "User",
					SubjectID:   suite.authUser.ID,
					SubjectName: suite.authUser.Email,
				}
				suite.db.Create(binding)
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/rbac/rolebindings", suite.authOrg.ID, suite.testWorkspace.ID), bytes.NewBuffer(body))
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

func (suite *RBACTestSuite) TestListRoleBindings() {
	// Create test role
	role := &db.Role{
		ID:          "role-test",
		Name:        "pod-reader",
		WorkspaceID: &suite.testWorkspace.ID,
		Scope:       "namespace",
		Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list"]}]`,
	}
	suite.Require().NoError(suite.db.Create(role).Error)

	// Create test role bindings
	bindings := []db.RoleBinding{
		{
			WorkspaceID: suite.testWorkspace.ID,
			RoleID:      "role-test",
			SubjectType: "User",
			SubjectID:   suite.authUser.ID,
			SubjectName: suite.authUser.Email,
			IsActive:    true,
		},
		{
			WorkspaceID: suite.testWorkspace.ID,
			RoleID:      "role-test",
			SubjectType: "Group",
			SubjectID:   suite.testGroup.ID,
			SubjectName: suite.testGroup.Name,
			IsActive:    true,
		},
	}
	for _, binding := range bindings {
		suite.Require().NoError(suite.db.Create(&binding).Error)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/rbac/rolebindings", suite.authOrg.ID, suite.testWorkspace.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		RoleBindings []db.RoleBinding `json:"role_bindings"`
		Total        int              `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Total)
	suite.Len(response.RoleBindings, 2)
}

// Test permission checking
func (suite *RBACTestSuite) TestCheckPermissions() {
	// Create test role with specific permissions
	role := &db.Role{
		ID:          "role-test",
		Name:        "pod-reader",
		WorkspaceID: &suite.testWorkspace.ID,
		Scope:       "namespace",
		Rules:       `[{"apiGroups":[""],"resources":["pods"],"verbs":["get","list","watch"]}]`,
	}
	suite.Require().NoError(suite.db.Create(role).Error)

	// Create role binding for user
	binding := &db.RoleBinding{
		WorkspaceID: suite.testWorkspace.ID,
		RoleID:      "role-test",
		SubjectType: "User",
		SubjectID:   suite.authUser.ID,
		SubjectName: suite.authUser.Email,
		IsActive:    true,
	}
	suite.Require().NoError(suite.db.Create(binding).Error)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedResult bool
	}{
		{
			name: "allowed permission - get pods",
			payload: map[string]interface{}{
				"user_id":    suite.authUser.ID,
				"api_group":  "",
				"resource":   "pods",
				"verb":       "get",
				"namespace":  "default",
			},
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name: "allowed permission - list pods",
			payload: map[string]interface{}{
				"user_id":   suite.authUser.ID,
				"api_group": "",
				"resource":  "pods",
				"verb":      "list",
			},
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name: "denied permission - create pods",
			payload: map[string]interface{}{
				"user_id":   suite.authUser.ID,
				"api_group": "",
				"resource":  "pods",
				"verb":      "create",
			},
			expectedStatus: http.StatusOK,
			expectedResult: false,
		},
		{
			name: "denied permission - different resource",
			payload: map[string]interface{}{
				"user_id":   suite.authUser.ID,
				"api_group": "",
				"resource":  "secrets",
				"verb":      "get",
			},
			expectedStatus: http.StatusOK,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/rbac/permissions/check", suite.authOrg.ID, suite.testWorkspace.ID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			var response struct {
				Allowed bool   `json:"allowed"`
				Reason  string `json:"reason"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)
			suite.Equal(tt.expectedResult, response.Allowed)
		})
	}
}

// Test project-specific RBAC
func (suite *RBACTestSuite) TestCreateProjectRole() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful project-scoped role creation",
			payload: map[string]interface{}{
				"name":        "project-admin",
				"description": "Full access to project resources",
				"scope":       "namespace",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{"*"},
						"resources": []string{"*"},
						"verbs":     []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "project deployment manager",
			payload: map[string]interface{}{
				"name":        "deployment-manager",
				"description": "Manage deployments in project",
				"scope":       "namespace",
				"rules": []map[string]interface{}{
					{
						"apiGroups": []string{"apps"},
						"resources": []string{"deployments", "replicasets"},
						"verbs":     []string{"get", "list", "create", "update", "patch", "delete"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/projects/%s/rbac/roles", suite.authOrg.ID, suite.testWorkspace.ID, suite.testProject.ID), bytes.NewBuffer(body))
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

// testAuthMiddleware creates a mock auth middleware for testing
func (suite *RBACTestSuite) testAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Simple mock authentication - extract user ID from token
		if authHeader == "Bearer test-token-"+suite.authUser.ID {
			// Get user's organizations
			var orgMemberships []db.OrganizationUser
			suite.db.Where("user_id = ?", suite.authUser.ID).Find(&orgMemberships)
			
			var orgIDs []string
			for _, membership := range orgMemberships {
				orgIDs = append(orgIDs, membership.OrganizationID)
			}
			
			c.Set("user_id", suite.authUser.ID)
			c.Set("user_email", suite.authUser.Email)
			c.Set("user_name", suite.authUser.DisplayName)
			c.Set("org_ids", orgIDs)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TestRBACTestSuite runs the test suite
func TestRBACTestSuite(t *testing.T) {
	suite.Run(t, new(RBACTestSuite))
}