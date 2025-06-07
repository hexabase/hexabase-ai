package db_test

import (
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := &db.User{
		ExternalID:  "google-123456",
		Provider:    "google",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	assert.Equal(t, "google-123456", user.ExternalID)
	assert.Equal(t, "google", user.Provider)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.DisplayName)
}

func TestOrganizationModel(t *testing.T) {
	org := &db.Organization{
		Name: "Test Organization",
	}

	assert.Equal(t, "Test Organization", org.Name)
	assert.Nil(t, org.StripeCustomerID)
	assert.Nil(t, org.StripeSubscriptionID)
}

func TestWorkspaceModel(t *testing.T) {
	workspace := &db.Workspace{
		OrganizationID: "org-123",
		Name:           "Test Workspace",
		PlanID:         "free-tier",
		VClusterStatus: "PENDING_CREATION",
	}

	assert.Equal(t, "org-123", workspace.OrganizationID)
	assert.Equal(t, "Test Workspace", workspace.Name)
	assert.Equal(t, "free-tier", workspace.PlanID)
	assert.Equal(t, "PENDING_CREATION", workspace.VClusterStatus)
}

func TestProjectModel(t *testing.T) {
	project := &db.Project{
		WorkspaceID: "ws-123",
		Name:        "Backend Services",
	}

	assert.Equal(t, "ws-123", project.WorkspaceID)
	assert.Equal(t, "Backend Services", project.Name)
	assert.Nil(t, project.ParentProjectID)
}

func TestGroupModel(t *testing.T) {
	group := &db.Group{
		WorkspaceID: "ws-123",
		Name:        "WSAdmins",
	}

	assert.Equal(t, "ws-123", group.WorkspaceID)
	assert.Equal(t, "WSAdmins", group.Name)
	assert.Nil(t, group.ParentGroupID)
}

func TestRoleModel(t *testing.T) {
	role := &db.Role{
		Name:     "custom-deployer",
		Rules:    `[{"apiGroups":["apps"],"resources":["deployments"],"verbs":["create","get","list"]}]`,
		IsCustom: true,
	}

	assert.Equal(t, "custom-deployer", role.Name)
	assert.True(t, role.IsCustom)
	assert.Contains(t, role.Rules, "deployments")
}

func TestVClusterTaskModel(t *testing.T) {
	task := &db.VClusterProvisioningTask{
		WorkspaceID: "ws-123",
		TaskType:    "CREATE",
		Status:      "PENDING",
		Payload:     `{"plan_id": "free-tier"}`,
	}

	assert.Equal(t, "ws-123", task.WorkspaceID)
	assert.Equal(t, "CREATE", task.TaskType)
	assert.Equal(t, "PENDING", task.Status)
	assert.Contains(t, task.Payload, "free-tier")
}

func TestPlanModel(t *testing.T) {
	plan := &db.Plan{
		ID:            "free-tier",
		Name:          "Free Tier",
		Price:         0.0,
		Currency:      "usd",
		StripePriceID: "price_test_123",
		IsActive:      true,
	}

	assert.Equal(t, "free-tier", plan.ID)
	assert.Equal(t, "Free Tier", plan.Name)
	assert.Equal(t, 0.0, plan.Price)
	assert.Equal(t, "usd", plan.Currency)
	assert.True(t, plan.IsActive)
}