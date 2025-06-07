package db_test

import (
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *DatabaseTestSuite) SetupSuite() {
	// For now, we'll use a test database connection string
	// In a real implementation, we'd use testcontainers
	dsn := "host=localhost user=postgres password=postgres dbname=hexabase_test port=5433 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	
	if err != nil {
		// Skip tests if no test database is available
		suite.T().Skip("Test database not available, skipping database tests")
		return
	}
	
	suite.db = database
	
	// Migrate all models
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
	
	if err != nil {
		suite.T().Fatalf("Failed to migrate database: %v", err)
	}
}

func (suite *DatabaseTestSuite) TearDownSuite() {
	if suite.db != nil {
		// Clean up test data
		suite.cleanupTables()
	}
}

func (suite *DatabaseTestSuite) SetupTest() {
	if suite.db != nil {
		suite.cleanupTables()
	}
}

func (suite *DatabaseTestSuite) cleanupTables() {
	tables := []string{
		"stripe_events",
		"v_cluster_provisioning_tasks", 
		"role_assignments",
		"roles",
		"group_memberships",
		"groups",
		"projects",
		"workspaces",
		"organization_users",
		"plans",
		"organizations",
		"users",
	}
	
	for _, table := range tables {
		suite.db.Exec("DELETE FROM " + table)
	}
}

func (suite *DatabaseTestSuite) TestCreateUser() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	user := &db.User{
		ExternalID:  "google-123456",
		Provider:    "google",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	
	err := suite.db.Create(user).Error
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.ID)
	assert.True(suite.T(), len(user.ID) > 10) // Should have UUID format
	assert.Contains(suite.T(), user.ID, "hxb-usr-")
	
	// Verify user creation
	var found db.User
	err = suite.db.First(&found, "id = ?", user.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Email, found.Email)
	assert.Equal(suite.T(), user.ExternalID, found.ExternalID)
	assert.Equal(suite.T(), user.Provider, found.Provider)
}

func (suite *DatabaseTestSuite) TestCreateOrganization() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	org := &db.Organization{
		Name: "Test Organization",
	}
	
	err := suite.db.Create(org).Error
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), org.ID)
	assert.Contains(suite.T(), org.ID, "org-")
	
	// Verify organization creation
	var found db.Organization
	err = suite.db.First(&found, "id = ?", org.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), org.Name, found.Name)
}

func (suite *DatabaseTestSuite) TestOrganizationUserRelationship() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	// Create user and organization
	user := &db.User{
		ExternalID:  "test-123",
		Provider:    "google",
		Email:       "user@test.com",
		DisplayName: "Test User",
	}
	suite.db.Create(user)
	
	org := &db.Organization{
		Name: "Test Org",
	}
	suite.db.Create(org)
	
	// Create relationship
	orgUser := &db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           "admin",
	}
	
	err := suite.db.Create(orgUser).Error
	assert.NoError(suite.T(), err)
	
	// Verify relationship
	var found db.OrganizationUser
	err = suite.db.Preload("User").Preload("Organization").
		First(&found, "organization_id = ? AND user_id = ?", org.ID, user.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "admin", found.Role)
	assert.Equal(suite.T(), user.Email, found.User.Email)
	assert.Equal(suite.T(), org.Name, found.Organization.Name)
}

func (suite *DatabaseTestSuite) TestCreateWorkspace() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	// Create organization first
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	
	// Create plan
	plan := &db.Plan{
		ID:            "free-tier",
		Name:          "Free Tier",
		Price:         0.0,
		Currency:      "usd",
		StripePriceID: "price_test_123",
	}
	suite.db.Create(plan)
	
	workspace := &db.Workspace{
		OrganizationID: org.ID,
		Name:           "Test Workspace",
		PlanID:         plan.ID,
		VClusterStatus: "PENDING_CREATION",
	}
	
	err := suite.db.Create(workspace).Error
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), workspace.ID)
	assert.Contains(suite.T(), workspace.ID, "ws-")
	
	// Verify workspace creation with associations
	var found db.Workspace
	err = suite.db.Preload("Organization").Preload("Plan").
		First(&found, "id = ?", workspace.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), workspace.Name, found.Name)
	assert.Equal(suite.T(), org.Name, found.Organization.Name)
	assert.Equal(suite.T(), plan.Name, found.Plan.Name)
}

func (suite *DatabaseTestSuite) TestProjectHierarchy() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	// Create prerequisites
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	
	plan := &db.Plan{
		ID:            "test-plan",
		Name:          "Test Plan", 
		Price:         10.0,
		Currency:      "usd",
		StripePriceID: "price_test_456",
	}
	suite.db.Create(plan)
	
	workspace := &db.Workspace{
		OrganizationID: org.ID,
		Name:           "Test Workspace",
		PlanID:         plan.ID,
	}
	suite.db.Create(workspace)
	
	// Create parent project
	parentProject := &db.Project{
		WorkspaceID: workspace.ID,
		Name:        "Backend Services",
	}
	suite.db.Create(parentProject)
	
	// Create child project
	childProject := &db.Project{
		WorkspaceID:     workspace.ID,
		Name:            "Auth Service",
		ParentProjectID: &parentProject.ID,
	}
	suite.db.Create(childProject)
	
	// Verify hierarchy
	var foundParent db.Project
	err := suite.db.Preload("ChildProjects").First(&foundParent, "id = ?", parentProject.ID).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundParent.ChildProjects, 1)
	assert.Equal(suite.T(), childProject.Name, foundParent.ChildProjects[0].Name)
	
	var foundChild db.Project
	err = suite.db.Preload("ParentProject").First(&foundChild, "id = ?", childProject.ID).Error
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundChild.ParentProject)
	assert.Equal(suite.T(), parentProject.Name, foundChild.ParentProject.Name)
}

func (suite *DatabaseTestSuite) TestGroupHierarchy() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	// Create prerequisites
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	
	plan := &db.Plan{
		ID:            "test-plan-2",
		Name:          "Test Plan 2",
		Price:         10.0,
		Currency:      "usd", 
		StripePriceID: "price_test_789",
	}
	suite.db.Create(plan)
	
	workspace := &db.Workspace{
		OrganizationID: org.ID,
		Name:           "Test Workspace",
		PlanID:         plan.ID,
	}
	suite.db.Create(workspace)
	
	// Create group hierarchy
	parentGroup := &db.Group{
		WorkspaceID: workspace.ID,
		Name:        "WorkspaceMembers",
	}
	suite.db.Create(parentGroup)
	
	childGroup := &db.Group{
		WorkspaceID:   workspace.ID,
		Name:          "WSAdmins",
		ParentGroupID: &parentGroup.ID,
	}
	suite.db.Create(childGroup)
	
	// Verify hierarchy
	var foundParent db.Group
	err := suite.db.Preload("ChildGroups").First(&foundParent, "id = ?", parentGroup.ID).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundParent.ChildGroups, 1)
	assert.Equal(suite.T(), childGroup.Name, foundParent.ChildGroups[0].Name)
}

func (suite *DatabaseTestSuite) TestVClusterTask() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}
	
	// Create prerequisites
	org := &db.Organization{Name: "Test Org"}
	suite.db.Create(org)
	
	plan := &db.Plan{
		ID:            "test-plan-3",
		Name:          "Test Plan 3",
		Price:         10.0,
		Currency:      "usd",
		StripePriceID: "price_test_999",
	}
	suite.db.Create(plan)
	
	workspace := &db.Workspace{
		OrganizationID: org.ID,
		Name:           "Test Workspace",
		PlanID:         plan.ID,
	}
	suite.db.Create(workspace)
	
	task := &db.VClusterProvisioningTask{
		WorkspaceID: workspace.ID,
		TaskType:    "CREATE",
		Status:      "PENDING",
		Payload:     `{"plan_id": "test-plan-3"}`,
	}
	
	err := suite.db.Create(task).Error
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), task.ID, "task-")
	
	// Verify task creation
	var found db.VClusterProvisioningTask
	err = suite.db.Preload("Workspace").First(&found, "id = ?", task.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "CREATE", found.TaskType)
	assert.Equal(suite.T(), workspace.Name, found.Workspace.Name)
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}