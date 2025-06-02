package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/api"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type OrganizationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *api.OrganizationHandler
	router  *gin.Engine
	user    *db.User
}

func (suite *OrganizationTestSuite) SetupSuite() {
	// Setup test database
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Migrate models
	err = database.AutoMigrate(
		&db.User{},
		&db.Organization{},
		&db.OrganizationUser{},
		&db.Workspace{},
	)
	suite.Require().NoError(err)

	// Create test user
	user := &db.User{
		ExternalID:  "test-external-id",
		Provider:    "test",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	err = database.Create(user).Error
	suite.Require().NoError(err)

	// Setup handler and router
	logger := zap.NewNop()
	cfg := &config.Config{}
	handler := api.NewOrganizationHandler(database, cfg, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup routes with mock auth middleware
	orgGroup := router.Group("/api/v1/organizations")
	orgGroup.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Set("user_email", user.Email)
		c.Next()
	})
	{
		orgGroup.POST("", handler.CreateOrganization)
		orgGroup.GET("", handler.ListOrganizations)
		orgGroup.GET("/:orgId", handler.GetOrganization)
		orgGroup.PUT("/:orgId", handler.UpdateOrganization)
		orgGroup.DELETE("/:orgId", handler.DeleteOrganization)
	}

	suite.db = database
	suite.handler = handler
	suite.router = router
	suite.user = user
}

func (suite *OrganizationTestSuite) TearDownTest() {
	// Clean up data after each test
	suite.db.Exec("DELETE FROM organization_users")
	suite.db.Exec("DELETE FROM organizations")
}

func (suite *OrganizationTestSuite) TestCreateOrganization() {
	// Test successful creation
	reqBody := api.CreateOrganizationRequest{
		Name: "Test Organization",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var resp api.OrganizationResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Organization", resp.Name)
	assert.Equal(suite.T(), "admin", resp.Role)
	assert.NotEmpty(suite.T(), resp.ID)

	// Verify organization was created in database
	var org db.Organization
	err = suite.db.First(&org, "id = ?", resp.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Test Organization", org.Name)

	// Verify user was added as admin
	var orgUser db.OrganizationUser
	err = suite.db.Where("organization_id = ? AND user_id = ?", resp.ID, suite.user.ID).First(&orgUser).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "admin", orgUser.Role)
}

func (suite *OrganizationTestSuite) TestCreateOrganization_ValidationError() {
	// Test with empty name
	reqBody := api.CreateOrganizationRequest{
		Name: "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/organizations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *OrganizationTestSuite) TestListOrganizations() {
	// Create test organizations
	org1 := &db.Organization{Name: "Org 1"}
	org2 := &db.Organization{Name: "Org 2"}
	suite.db.Create(org1)
	suite.db.Create(org2)

	// Add user to both organizations with different roles
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org1.ID,
		UserID:         suite.user.ID,
		Role:           "admin",
	})
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org2.ID,
		UserID:         suite.user.ID,
		Role:           "member",
	})

	// Test listing
	req := httptest.NewRequest("GET", "/api/v1/organizations", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(2), resp["total"])

	orgs := resp["organizations"].([]interface{})
	assert.Len(suite.T(), orgs, 2)

	// Check roles are included
	org1Data := orgs[0].(map[string]interface{})
	org2Data := orgs[1].(map[string]interface{})
	
	if org1Data["name"] == "Org 1" {
		assert.Equal(suite.T(), "admin", org1Data["role"])
		assert.Equal(suite.T(), "member", org2Data["role"])
	} else {
		assert.Equal(suite.T(), "member", org1Data["role"])
		assert.Equal(suite.T(), "admin", org2Data["role"])
	}
}

func (suite *OrganizationTestSuite) TestGetOrganization() {
	// Create test organization
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         suite.user.ID,
		Role:           "admin",
	})

	// Test getting organization
	req := httptest.NewRequest("GET", "/api/v1/organizations/"+org.ID, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp api.OrganizationResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.ID, resp.ID)
	assert.Equal(suite.T(), "Test Org", resp.Name)
	assert.Equal(suite.T(), "admin", resp.Role)
}

func (suite *OrganizationTestSuite) TestGetOrganization_NotFound() {
	req := httptest.NewRequest("GET", "/api/v1/organizations/non-existent-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *OrganizationTestSuite) TestUpdateOrganization() {
	// Create test organization
	org := &db.Organization{Name: "Old Name"}
	suite.db.Create(org)
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         suite.user.ID,
		Role:           "admin",
	})

	// Update organization
	reqBody := api.UpdateOrganizationRequest{
		Name: "New Name",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/v1/organizations/"+org.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp api.OrganizationResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Name", resp.Name)

	// Verify in database
	var updatedOrg db.Organization
	suite.db.First(&updatedOrg, "id = ?", org.ID)
	assert.Equal(suite.T(), "New Name", updatedOrg.Name)
}

func (suite *OrganizationTestSuite) TestUpdateOrganization_Forbidden() {
	// Create organization where user is not admin
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         suite.user.ID,
		Role:           "member", // Not admin
	})

	reqBody := api.UpdateOrganizationRequest{Name: "New Name"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PUT", "/api/v1/organizations/"+org.ID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *OrganizationTestSuite) TestDeleteOrganization() {
	// Create test organization
	org := &db.Organization{Name: "To Delete"}
	suite.db.Create(org)
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         suite.user.ID,
		Role:           "admin",
	})

	req := httptest.NewRequest("DELETE", "/api/v1/organizations/"+org.ID, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify organization was deleted
	var count int64
	suite.db.Model(&db.Organization{}).Where("id = ?", org.ID).Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *OrganizationTestSuite) TestDeleteOrganization_WithWorkspaces() {
	// Create organization with workspace
	org := &db.Organization{Name: "Has Workspaces"}
	suite.db.Create(org)
	suite.db.Create(&db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         suite.user.ID,
		Role:           "admin",
	})
	
	// Create a workspace
	workspace := &db.Workspace{
		OrganizationID: org.ID,
		Name:           "Test Workspace",
	}
	suite.db.Create(workspace)

	req := httptest.NewRequest("DELETE", "/api/v1/organizations/"+org.ID, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(suite.T(), resp["error"], "cannot delete organization with active workspaces")
}

func TestOrganizationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTestSuite))
}