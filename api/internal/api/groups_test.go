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

// GroupTestSuite is the test suite for group handlers
type GroupTestSuite struct {
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
func (suite *GroupTestSuite) SetupSuite() {
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
	
	// Organization routes with workspaces and groups
	orgs := protected.Group("/organizations")
	workspaces := orgs.Group("/:orgId/workspaces")
	
	// Group management endpoints
	groups := workspaces.Group("/:wsId/groups")
	{
		groups.POST("/", suite.handlers.Groups.CreateGroup)
		groups.GET("/", suite.handlers.Groups.ListGroups)
		groups.GET("/:groupId", suite.handlers.Groups.GetGroup)
		groups.PUT("/:groupId", suite.handlers.Groups.UpdateGroup)
		groups.DELETE("/:groupId", suite.handlers.Groups.DeleteGroup)
		groups.POST("/:groupId/members", suite.handlers.Groups.AddGroupMember)
		groups.DELETE("/:groupId/members/:userId", suite.handlers.Groups.RemoveGroupMember)
		groups.GET("/:groupId/members", suite.handlers.Groups.ListGroupMembers)
	}
}

// SetupTest runs before each test
func (suite *GroupTestSuite) SetupTest() {
	// Clean up database (SQLite doesn't support TRUNCATE CASCADE)
	suite.db.Exec("DELETE FROM group_memberships")
	suite.db.Exec("DELETE FROM groups")
	suite.db.Exec("DELETE FROM role_assignments")
	suite.db.Exec("DELETE FROM roles")
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

// Test group creation and management
func (suite *GroupTestSuite) TestCreateGroup() {
	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name: "successful root group creation",
			payload: map[string]interface{}{
				"name": "engineering",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing group name",
			payload: map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "group name is required",
		},
		{
			name: "duplicate group name in same workspace",
			payload: map[string]interface{}{
				"name": "existing-group",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "group with this name already exists",
			setup: func() {
				group := &db.Group{
					WorkspaceID: suite.testWs.ID,
					Name:        "existing-group",
				}
				suite.Require().NoError(suite.db.Create(group).Error)
			},
		},
		{
			name: "create hierarchical group (child group)",
			payload: map[string]interface{}{
				"name":            "frontend-team",
				"parent_group_id": "hierarchical-parent-group-id",
			},
			expectedStatus: http.StatusCreated,
			setup: func() {
				parentGroup := &db.Group{
					ID:          "hierarchical-parent-group-id",
					WorkspaceID: suite.testWs.ID,
					Name:        "engineering",
				}
				suite.Require().NoError(suite.db.Create(parentGroup).Error)
			},
		},
		{
			name: "create group with invalid parent",
			payload: map[string]interface{}{
				"name":            "orphan-team",
				"parent_group_id": "nonexistent-parent",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "parent group not found",
		},
		{
			name: "create child group successfully",
			payload: map[string]interface{}{
				"name":            "circular-group",
				"parent_group_id": "circular-child-group-id",
			},
			expectedStatus: http.StatusCreated,
			setup: func() {
				// Create parent group
				parent := &db.Group{
					ID:          "circular-parent-group-id",
					WorkspaceID: suite.testWs.ID,
					Name:        "parent",
				}
				suite.Require().NoError(suite.db.Create(parent).Error)
				
				// Create child group
				child := &db.Group{
					ID:              "circular-child-group-id",
					WorkspaceID:     suite.testWs.ID,
					Name:            "child",
					ParentGroupID:   &parent.ID,
				}
				suite.Require().NoError(suite.db.Create(child).Error)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/", suite.authOrg.ID, suite.testWs.ID), bytes.NewBuffer(body))
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
				var group db.Group
				err := json.Unmarshal(w.Body.Bytes(), &group)
				suite.NoError(err)
				suite.NotEmpty(group.ID)
				suite.Equal(tt.payload.(map[string]interface{})["name"], group.Name)
			}
		})
	}
}

func (suite *GroupTestSuite) TestListGroups() {
	// Create test groups with hierarchy
	rootGroup := &db.Group{
		ID:          "grp-root",
		WorkspaceID: suite.testWs.ID,
		Name:        "engineering",
	}
	suite.Require().NoError(suite.db.Create(rootGroup).Error)

	child1 := &db.Group{
		ID:            "grp-child1",
		WorkspaceID:   suite.testWs.ID,
		Name:          "backend",
		ParentGroupID: &rootGroup.ID,
	}
	suite.Require().NoError(suite.db.Create(child1).Error)

	child2 := &db.Group{
		ID:            "grp-child2",
		WorkspaceID:   suite.testWs.ID,
		Name:          "frontend",
		ParentGroupID: &rootGroup.ID,
	}
	suite.Require().NoError(suite.db.Create(child2).Error)

	grandchild := &db.Group{
		ID:            "grp-grandchild",
		WorkspaceID:   suite.testWs.ID,
		Name:          "react-team",
		ParentGroupID: &child2.ID,
	}
	suite.Require().NoError(suite.db.Create(grandchild).Error)

	// Test listing groups
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/", suite.authOrg.ID, suite.testWs.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response struct {
		Groups []db.Group `json:"groups"`
		Total  int        `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(4, response.Total)
	suite.Len(response.Groups, 4)
}

func (suite *GroupTestSuite) TestGetGroup() {
	// Create test group
	group := &db.Group{
		ID:          "grp-test",
		WorkspaceID: suite.testWs.ID,
		Name:        "test-group",
	}
	suite.Require().NoError(suite.db.Create(group).Error)

	tests := []struct {
		name           string
		groupID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful get group",
			groupID:        "grp-test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "group not found",
			groupID:        "grp-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "group not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/%s", suite.authOrg.ID, suite.testWs.ID, tt.groupID), nil)
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
				var group db.Group
				err := json.Unmarshal(w.Body.Bytes(), &group)
				suite.NoError(err)
				suite.Equal("grp-test", group.ID)
				suite.Equal("test-group", group.Name)
			}
		})
	}
}

func (suite *GroupTestSuite) TestUpdateGroup() {
	// Create test group
	group := &db.Group{
		ID:          "grp-update",
		WorkspaceID: suite.testWs.ID,
		Name:        "old-name",
	}
	suite.Require().NoError(suite.db.Create(group).Error)

	tests := []struct {
		name           string
		groupID        string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name:    "successful name update",
			groupID: "grp-update",
			payload: map[string]interface{}{
				"name": "new-name",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "group not found",
			groupID: "grp-nonexistent",
			payload: map[string]interface{}{
				"name": "new-name",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "group not found",
		},
		{
			name:    "duplicate name",
			groupID: "grp-update",
			payload: map[string]interface{}{
				"name": "existing-name",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "group with this name already exists",
			setup: func() {
				existing := &db.Group{
					WorkspaceID: suite.testWs.ID,
					Name:        "existing-name",
				}
				suite.Require().NoError(suite.db.Create(existing).Error)
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/%s", suite.authOrg.ID, suite.testWs.ID, tt.groupID), bytes.NewBuffer(body))
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

func (suite *GroupTestSuite) TestDeleteGroup() {
	tests := []struct {
		name           string
		groupID        string
		expectedStatus int
		expectedError  string
		setup          func() *db.Group
	}{
		{
			name:           "successful delete empty group",
			groupID:        "grp-delete-empty",
			expectedStatus: http.StatusOK,
			setup: func() *db.Group {
				group := &db.Group{
					ID:          "grp-delete-empty",
					WorkspaceID: suite.testWs.ID,
					Name:        "empty-group",
				}
				suite.Require().NoError(suite.db.Create(group).Error)
				return group
			},
		},
		{
			name:           "cannot delete group with members",
			groupID:        "grp-delete-members",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot delete group with existing members",
			setup: func() *db.Group {
				group := &db.Group{
					ID:          "grp-delete-members",
					WorkspaceID: suite.testWs.ID,
					Name:        "group-with-members",
				}
				suite.Require().NoError(suite.db.Create(group).Error)

				// Add member to group
				membership := &db.GroupMembership{
					GroupID:  group.ID,
					UserID:   suite.authUser.ID,
					JoinedAt: time.Now(),
				}
				suite.Require().NoError(suite.db.Create(membership).Error)
				return group
			},
		},
		{
			name:           "cannot delete group with child groups",
			groupID:        "grp-delete-parent",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cannot delete group with child groups",
			setup: func() *db.Group {
				parent := &db.Group{
					ID:          "grp-delete-parent",
					WorkspaceID: suite.testWs.ID,
					Name:        "parent-group",
				}
				suite.Require().NoError(suite.db.Create(parent).Error)

				// Create child group
				child := &db.Group{
					WorkspaceID:   suite.testWs.ID,
					Name:          "child-group",
					ParentGroupID: &parent.ID,
				}
				suite.Require().NoError(suite.db.Create(child).Error)
				return parent
			},
		},
		{
			name:           "group not found",
			groupID:        "grp-nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "group not found",
			setup:          func() *db.Group { return nil },
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/%s", suite.authOrg.ID, suite.testWs.ID, tt.groupID), nil)
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

func (suite *GroupTestSuite) TestAddGroupMember() {
	// Create test group
	group := &db.Group{
		ID:          "grp-add-member",
		WorkspaceID: suite.testWs.ID,
		Name:        "test-group",
	}
	suite.Require().NoError(suite.db.Create(group).Error)

	// Create another user in the organization
	member := &db.User{
		ID:          "member-to-add",
		ExternalID:  "google-999",
		Provider:    "google",
		Email:       "member@example.com",
		DisplayName: "Member User",
	}
	suite.Require().NoError(suite.db.Create(member).Error)

	orgUser := &db.OrganizationUser{
		OrganizationID: suite.authOrg.ID,
		UserID:         member.ID,
		Role:           "member",
		JoinedAt:       time.Now(),
	}
	suite.Require().NoError(suite.db.Create(orgUser).Error)

	tests := []struct {
		name           string
		groupID        string
		payload        interface{}
		expectedStatus int
		expectedError  string
		setup          func()
	}{
		{
			name:    "successful add member",
			groupID: "grp-add-member",
			payload: map[string]interface{}{
				"user_id": "member-to-add",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "user already in group",
			groupID: "grp-add-member",
			payload: map[string]interface{}{
				"user_id": suite.authUser.ID,
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "user already in group",
			setup: func() {
				membership := &db.GroupMembership{
					GroupID:  group.ID,
					UserID:   suite.authUser.ID,
					JoinedAt: time.Now(),
				}
				suite.Require().NoError(suite.db.Create(membership).Error)
			},
		},
		{
			name:    "user not in organization",
			groupID: "grp-add-member",
			payload: map[string]interface{}{
				"user_id": "external-user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "user not found in organization",
		},
		{
			name:    "group not found",
			groupID: "grp-nonexistent",
			payload: map[string]interface{}{
				"user_id": "member-to-add",
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "group not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/%s/members", suite.authOrg.ID, suite.testWs.ID, tt.groupID), bytes.NewBuffer(body))
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

func (suite *GroupTestSuite) TestListGroupMembers() {
	// Create test group
	group := &db.Group{
		ID:          "grp-list-members",
		WorkspaceID: suite.testWs.ID,
		Name:        "test-group",
	}
	suite.Require().NoError(suite.db.Create(group).Error)

	// Create and add members
	for i := 0; i < 3; i++ {
		user := &db.User{
			ID:          fmt.Sprintf("member-%d", i),
			ExternalID:  fmt.Sprintf("google-%d", i),
			Provider:    "google",
			Email:       fmt.Sprintf("member%d@example.com", i),
			DisplayName: fmt.Sprintf("Member %d", i),
		}
		suite.Require().NoError(suite.db.Create(user).Error)

		orgUser := &db.OrganizationUser{
			OrganizationID: suite.authOrg.ID,
			UserID:         user.ID,
			Role:           "member",
			JoinedAt:       time.Now(),
		}
		suite.Require().NoError(suite.db.Create(orgUser).Error)

		membership := &db.GroupMembership{
			GroupID:  group.ID,
			UserID:   user.ID,
			JoinedAt: time.Now(),
		}
		suite.Require().NoError(suite.db.Create(membership).Error)
	}

	// Test listing members
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/%s/members", suite.authOrg.ID, suite.testWs.ID, group.ID), nil)
	req.Header.Set("Authorization", suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	
	members, ok := response["members"].([]interface{})
	suite.True(ok)
	suite.Equal(3, len(members))
}

// Test workspace authorization
func (suite *GroupTestSuite) TestGroupAuthorization() {
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

	// Try to create group with other user's token
	otherToken := "Bearer test-token-" + otherUser.ID

	payload := map[string]interface{}{
		"name": "unauthorized-group",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/groups/", suite.authOrg.ID, suite.testWs.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
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
func (suite *GroupTestSuite) testAuthMiddleware() gin.HandlerFunc {
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

// TestGroupSuite runs the test suite
func TestGroupSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}