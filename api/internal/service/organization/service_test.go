package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

// Mock implementations for dependencies

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *organization.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Organization), args.Error(1)
}

func (m *MockRepository) GetOrganizationByName(ctx context.Context, name string) (*organization.Organization, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Organization), args.Error(1)
}

func (m *MockRepository) ListOrganizations(ctx context.Context, filter organization.OrganizationFilter) ([]*organization.Organization, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*organization.Organization), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateOrganization(ctx context.Context, org *organization.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) DeleteOrganization(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockRepository) AddMember(ctx context.Context, member *organization.OrganizationUser) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRepository) GetMember(ctx context.Context, orgID, userID string) (*organization.OrganizationUser, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.OrganizationUser), args.Error(1)
}

func (m *MockRepository) ListMembers(ctx context.Context, filter organization.MemberFilter) ([]*organization.OrganizationUser, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*organization.OrganizationUser), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateMember(ctx context.Context, member *organization.OrganizationUser) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockRepository) CountMembers(ctx context.Context, orgID string) (int, error) {
	args := m.Called(ctx, orgID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetUser(ctx context.Context, userID string) (*organization.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*organization.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.User), args.Error(1)
}

func (m *MockRepository) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*organization.User, error) {
	args := m.Called(ctx, userIDs)
	return args.Get(0).([]*organization.User), args.Error(1)
}

func (m *MockRepository) CreateInvitation(ctx context.Context, invitation *organization.Invitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockRepository) GetInvitation(ctx context.Context, invitationID string) (*organization.Invitation, error) {
	args := m.Called(ctx, invitationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Invitation), args.Error(1)
}

func (m *MockRepository) GetInvitationByToken(ctx context.Context, token string) (*organization.Invitation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Invitation), args.Error(1)
}

func (m *MockRepository) ListInvitations(ctx context.Context, orgID string, status string) ([]*organization.Invitation, error) {
	args := m.Called(ctx, orgID, status)
	return args.Get(0).([]*organization.Invitation), args.Error(1)
}

func (m *MockRepository) UpdateInvitation(ctx context.Context, invitation *organization.Invitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockRepository) DeleteInvitation(ctx context.Context, invitationID string) error {
	args := m.Called(ctx, invitationID)
	return args.Error(0)
}

func (m *MockRepository) DeleteExpiredInvitations(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockRepository) GetOrganizationStats(ctx context.Context, orgID string) (*organization.OrganizationStats, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.OrganizationStats), args.Error(1)
}

func (m *MockRepository) GetWorkspaceCount(ctx context.Context, orgID string) (total int, active int, err error) {
	args := m.Called(ctx, orgID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetProjectCount(ctx context.Context, orgID string) (int, error) {
	args := m.Called(ctx, orgID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetResourceUsage(ctx context.Context, orgID string) (*organization.Usage, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Usage), args.Error(1)
}

func (m *MockRepository) ListWorkspaces(ctx context.Context, orgID string) ([]*organization.WorkspaceInfo, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*organization.WorkspaceInfo), args.Error(1)
}

func (m *MockRepository) CreateActivity(ctx context.Context, activity *organization.Activity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockRepository) ListActivities(ctx context.Context, filter organization.ActivityFilter) ([]*organization.Activity, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*organization.Activity), args.Error(1)
}

func (m *MockRepository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	args := m.Called(ctx, orgID, userID, role)
	return args.Error(0)
}

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) GetUser(ctx context.Context, userID string) (*organization.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*organization.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

type MockBillingRepository struct {
	mock.Mock
}

func (m *MockBillingRepository) CreateCustomer(ctx context.Context, org *organization.Organization) (string, error) {
	args := m.Called(ctx, org)
	return args.String(0), args.Error(1)
}

func (m *MockBillingRepository) DeleteCustomer(ctx context.Context, customerID string) error {
	args := m.Called(ctx, customerID)
	return args.Error(0)
}

func (m *MockBillingRepository) GetOrganizationSubscription(ctx context.Context, orgID string) (*organization.Subscription, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organization.Subscription), args.Error(1)
}

func (m *MockBillingRepository) CancelSubscription(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

// Test cases

func TestCreateOrganization(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful organization creation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		req := &organization.CreateOrganizationRequest{
			Name:        "test-org",
			DisplayName: "Test Organization",
			Description: "A test organization",
			Email:       "test@example.com",
			Website:     "https://test.com",
		}

		userID := "user-123"

		// Mock expectations
		mockRepo.On("CreateOrganization", ctx, mock.MatchedBy(func(org *organization.Organization) bool {
			return org.Name == "test-org" &&
				org.DisplayName == "Test Organization" &&
				org.OwnerID == userID &&
				org.Status == "active"
		})).Return(nil)

		mockRepo.On("AddMember", ctx, mock.MatchedBy(func(member *organization.OrganizationUser) bool {
			return member.UserID == userID &&
				member.Role == "admin" &&
				member.Status == "active"
		})).Return(nil)

		mockBillingRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*organization.Organization")).Return("cus-123", nil)

		// Execute
		org, err := service.CreateOrganization(ctx, userID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, "test-org", org.Name)
		assert.Equal(t, "Test Organization", org.DisplayName)
		assert.Equal(t, userID, org.OwnerID)
		assert.Equal(t, "active", org.Status)

		mockRepo.AssertExpectations(t)
		mockBillingRepo.AssertExpectations(t)
	})

	t.Run("organization creation with empty display name", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		req := &organization.CreateOrganizationRequest{
			Name: "test-org",
		}

		userID := "user-123"

		// Mock expectations
		mockRepo.On("CreateOrganization", ctx, mock.MatchedBy(func(org *organization.Organization) bool {
			return org.DisplayName == "test-org" // Should default to Name
		})).Return(nil)

		mockRepo.On("AddMember", ctx, mock.AnythingOfType("*organization.OrganizationUser")).Return(nil)
		mockBillingRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*organization.Organization")).Return("cus-123", nil)

		// Execute
		org, err := service.CreateOrganization(ctx, userID, req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test-org", org.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("organization creation with empty name", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		req := &organization.CreateOrganizationRequest{
			Name: "",
		}

		userID := "user-123"

		// Execute
		org, err := service.CreateOrganization(ctx, userID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "organization name is required")
	})

	t.Run("organization creation fails on repository error", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		req := &organization.CreateOrganizationRequest{
			Name: "test-org",
		}

		userID := "user-123"

		// Mock expectations
		mockRepo.On("CreateOrganization", ctx, mock.AnythingOfType("*organization.Organization")).
			Return(errors.New("database error"))

		// Execute
		org, err := service.CreateOrganization(ctx, userID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "failed to create organization")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetOrganization(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful organization retrieval", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		expectedOrg := &organization.Organization{
			ID:          orgID,
			Name:        "test-org",
			DisplayName: "Test Organization",
			Status:      "active",
		}

		subscription := &organization.Subscription{
			PlanID:           "plan-123",
			PlanName:         "Professional",
			Status:           "active",
			CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(expectedOrg, nil)
		mockRepo.On("ListMembers", ctx, mock.MatchedBy(func(filter organization.MemberFilter) bool {
			return filter.OrganizationID == orgID && filter.PageSize == 1000
		})).Return([]*organization.OrganizationUser{}, 5, nil)
		mockBillingRepo.On("GetOrganizationSubscription", ctx, orgID).Return(subscription, nil)

		// Execute
		org, err := service.GetOrganization(ctx, orgID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, org)
		assert.Equal(t, orgID, org.ID)
		assert.Equal(t, 5, org.MemberCount)
		assert.NotNil(t, org.SubscriptionInfo)
		assert.Equal(t, "Professional", org.SubscriptionInfo.PlanName)

		mockRepo.AssertExpectations(t)
		mockBillingRepo.AssertExpectations(t)
	})

	t.Run("organization not found", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-notfound"

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(nil, errors.New("not found"))

		// Execute
		org, err := service.GetOrganization(ctx, orgID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, org)
		assert.Contains(t, err.Error(), "failed to get organization")

		mockRepo.AssertExpectations(t)
	})
}

func TestListOrganizations(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("list organizations by user", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		userID := "user-123"
		filter := organization.OrganizationFilter{
			UserID:   userID,
			Page:     1,
			PageSize: 10,
		}

		orgIDs := []string{"org-1", "org-2", "org-3"}
		orgs := []*organization.Organization{
			{ID: "org-1", Name: "org1"},
			{ID: "org-2", Name: "org2"},
			{ID: "org-3", Name: "org3"},
		}

		// Mock expectations
		mockAuthRepo.On("GetUserOrganizations", ctx, userID).Return(orgIDs, nil)
		for i, orgID := range orgIDs {
			mockRepo.On("GetOrganization", ctx, orgID).Return(orgs[i], nil)
		}

		// Execute
		result, err := service.ListOrganizations(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Organizations, 3)
		assert.Equal(t, 3, result.Total)

		mockAuthRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("list all organizations", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		filter := organization.OrganizationFilter{
			Page:     1,
			PageSize: 10,
		}

		orgs := []*organization.Organization{
			{ID: "org-1", Name: "org1"},
			{ID: "org-2", Name: "org2"},
		}

		// Mock expectations
		mockRepo.On("ListOrganizations", ctx, filter).Return(orgs, 2, nil)

		// Execute
		result, err := service.ListOrganizations(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Organizations, 2)
		assert.Equal(t, 2, result.Total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user has no organizations", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		userID := "user-123"
		filter := organization.OrganizationFilter{
			UserID:   userID,
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockAuthRepo.On("GetUserOrganizations", ctx, userID).Return([]string{}, nil)

		// Execute
		result, err := service.ListOrganizations(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Organizations, 0)
		assert.Equal(t, 0, result.Total)

		mockAuthRepo.AssertExpectations(t)
	})
}

func TestUpdateOrganization(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful organization update", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		existingOrg := &organization.Organization{
			ID:          orgID,
			Name:        "test-org",
			DisplayName: "Old Display Name",
			Description: "Old description",
		}

		req := &organization.UpdateOrganizationRequest{
			DisplayName: "New Display Name",
			Description: "New description",
			Settings: map[string]interface{}{
				"theme": "dark",
			},
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(existingOrg, nil)
		mockRepo.On("UpdateOrganization", ctx, mock.MatchedBy(func(org *organization.Organization) bool {
			return org.DisplayName == "New Display Name" &&
				org.Description == "New description" &&
				org.Settings["theme"] == "dark"
		})).Return(nil)

		// Execute
		updatedOrg, err := service.UpdateOrganization(ctx, orgID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, updatedOrg)
		assert.Equal(t, "New Display Name", updatedOrg.DisplayName)
		assert.Equal(t, "New description", updatedOrg.Description)

		mockRepo.AssertExpectations(t)
	})

	t.Run("organization not found for update", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-notfound"
		req := &organization.UpdateOrganizationRequest{
			DisplayName: "New Name",
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(nil, errors.New("not found"))

		// Execute
		updatedOrg, err := service.UpdateOrganization(ctx, orgID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, updatedOrg)
		assert.Contains(t, err.Error(), "failed to get organization")

		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteOrganization(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful organization deletion", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"

		// Mock expectations
		mockRepo.On("ListWorkspaces", ctx, orgID).Return([]*organization.WorkspaceInfo{}, nil)
		mockBillingRepo.On("CancelSubscription", ctx, orgID).Return(nil)
		mockBillingRepo.On("DeleteCustomer", ctx, orgID).Return(nil)
		mockRepo.On("DeleteOrganization", ctx, orgID).Return(nil)

		// Execute
		err := service.DeleteOrganization(ctx, orgID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockBillingRepo.AssertExpectations(t)
	})

	t.Run("cannot delete organization with active workspaces", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		workspaces := []*organization.WorkspaceInfo{
			{ID: "ws-1", Name: "workspace1"},
		}

		// Mock expectations
		mockRepo.On("ListWorkspaces", ctx, orgID).Return(workspaces, nil)

		// Execute
		err := service.DeleteOrganization(ctx, orgID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete organization with active workspaces")

		mockRepo.AssertExpectations(t)
	})
}

func TestInviteUser(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful user invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		inviterID := "user-123"
		req := &organization.InviteUserRequest{
			Email: "newuser@example.com",
			Role:  "member",
		}

		org := &organization.Organization{
			ID:   orgID,
			Name: "test-org",
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(nil, errors.New("not found"))
		mockRepo.On("ListInvitations", ctx, orgID, "pending").Return([]*organization.Invitation{}, nil)
		mockRepo.On("CreateInvitation", ctx, mock.MatchedBy(func(inv *organization.Invitation) bool {
			return inv.OrganizationID == orgID &&
				inv.Email == req.Email &&
				inv.Role == req.Role &&
				inv.InvitedBy == inviterID &&
				inv.Status == "pending"
		})).Return(nil)

		// Execute
		invitation, err := service.InviteUser(ctx, orgID, inviterID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, req.Email, invitation.Email)
		assert.Equal(t, req.Role, invitation.Role)
		assert.Equal(t, "pending", invitation.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot invite existing member", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		inviterID := "user-123"
		existingUserID := "user-456"
		req := &organization.InviteUserRequest{
			Email: "existinguser@example.com",
			Role:  "member",
		}

		org := &organization.Organization{
			ID:   orgID,
			Name: "test-org",
		}

		existingUser := &organization.User{
			ID:    existingUserID,
			Email: req.Email,
		}

		members := []*organization.OrganizationUser{
			{UserID: existingUserID, Email: req.Email},
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(existingUser, nil)
		mockRepo.On("ListMembers", ctx, mock.MatchedBy(func(filter organization.MemberFilter) bool {
			return filter.OrganizationID == orgID
		})).Return(members, 1, nil)

		// Execute
		invitation, err := service.InviteUser(ctx, orgID, inviterID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "user is already a member")

		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot invite with existing pending invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		inviterID := "user-123"
		req := &organization.InviteUserRequest{
			Email: "newuser@example.com",
			Role:  "member",
		}

		org := &organization.Organization{
			ID:   orgID,
			Name: "test-org",
		}

		existingInvitations := []*organization.Invitation{
			{Email: req.Email, Status: "pending"},
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("GetUserByEmail", ctx, req.Email).Return(nil, errors.New("not found"))
		mockRepo.On("ListInvitations", ctx, orgID, "pending").Return(existingInvitations, nil)

		// Execute
		invitation, err := service.InviteUser(ctx, orgID, inviterID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "invitation already exists")

		mockRepo.AssertExpectations(t)
	})
}

func TestAcceptInvitation(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful invitation acceptance", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		token := "invite-token-123"
		userID := "user-123"
		orgID := "org-123"

		invitation := &organization.Invitation{
			ID:             "inv-123",
			OrganizationID: orgID,
			Email:          "user@example.com",
			Role:           "member",
			Status:         "pending",
			Token:          token,
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}

		// Mock expectations
		mockRepo.On("GetInvitationByToken", ctx, token).Return(invitation, nil)
		mockRepo.On("UpdateInvitation", ctx, mock.MatchedBy(func(inv *organization.Invitation) bool {
			return inv.Status == "accepted" && inv.AcceptedAt != nil
		})).Return(nil)
		mockRepo.On("AddMember", ctx, mock.MatchedBy(func(member *organization.OrganizationUser) bool {
			return member.OrganizationID == orgID &&
				member.UserID == userID &&
				member.Email == invitation.Email &&
				member.Role == invitation.Role &&
				member.Status == "active"
		})).Return(nil)

		// Execute
		member, err := service.AcceptInvitation(ctx, token, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, orgID, member.OrganizationID)
		assert.Equal(t, "member", member.Role)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot accept expired invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		token := "invite-token-123"
		userID := "user-123"

		invitation := &organization.Invitation{
			ID:        "inv-123",
			Status:    "pending",
			Token:     token,
			ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		}

		// Mock expectations
		mockRepo.On("GetInvitationByToken", ctx, token).Return(invitation, nil)

		// Execute
		member, err := service.AcceptInvitation(ctx, token, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "invitation has expired")

		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot accept non-pending invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		token := "invite-token-123"
		userID := "user-123"

		invitation := &organization.Invitation{
			ID:        "inv-123",
			Status:    "accepted", // Already accepted
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		// Mock expectations
		mockRepo.On("GetInvitationByToken", ctx, token).Return(invitation, nil)

		// Execute
		member, err := service.AcceptInvitation(ctx, token, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "invitation is not pending")

		mockRepo.AssertExpectations(t)
	})
}

func TestListMembers(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful member listing", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		filter := organization.MemberFilter{
			OrganizationID: "org-123",
			Page:           1,
			PageSize:       10,
		}

		orgUsers := []*organization.OrganizationUser{
			{
				ID:     "ou-1",
				UserID: "user-1",
				Role:   "admin",
				Status: "active",
			},
			{
				ID:     "ou-2",
				UserID: "user-2",
				Role:   "member",
				Status: "active",
			},
		}

		users := map[string]*organization.User{
			"user-1": {
				ID:          "user-1",
				Email:       "admin@example.com",
				DisplayName: "Admin User",
			},
			"user-2": {
				ID:          "user-2",
				Email:       "member@example.com",
				DisplayName: "Member User",
			},
		}

		// Mock expectations
		mockRepo.On("ListMembers", ctx, filter).Return(orgUsers, 2, nil)
		for userID, user := range users {
			mockAuthRepo.On("GetUser", ctx, userID).Return(user, nil)
		}

		// Execute
		result, err := service.ListMembers(ctx, filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Members, 2)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, "admin@example.com", result.Members[0].Email)
		assert.Equal(t, "member@example.com", result.Members[1].Email)

		mockRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}

func TestRemoveMember(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful member removal", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-456"
		removerID := "user-123"

		org := &organization.Organization{
			ID:      orgID,
			OwnerID: "user-123", // Different from userID
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("RemoveMember", ctx, orgID, userID).Return(nil)

		// Execute
		err := service.RemoveMember(ctx, orgID, userID, removerID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cannot remove organization owner", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		ownerID := "user-123"
		removerID := "user-456"

		org := &organization.Organization{
			ID:      orgID,
			OwnerID: ownerID,
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)

		// Execute
		err := service.RemoveMember(ctx, orgID, ownerID, removerID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot remove organization owner")

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateMemberRole(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful role update", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-456"
		req := &organization.UpdateMemberRoleRequest{
			Role: "admin",
		}

		org := &organization.Organization{
			ID:      orgID,
			OwnerID: "user-123", // Different from userID
		}

		updatedMember := &organization.OrganizationUser{
			ID:     "ou-1",
			UserID: userID,
			Role:   "admin",
			Status: "active",
		}

		user := &organization.User{
			ID:          userID,
			Email:       "user@example.com",
			DisplayName: "Test User",
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)
		mockRepo.On("UpdateMemberRole", ctx, orgID, userID, "admin").Return(nil)
		mockRepo.On("GetMember", ctx, orgID, userID).Return(updatedMember, nil)
		mockAuthRepo.On("GetUser", ctx, userID).Return(user, nil)

		// Execute
		member, err := service.UpdateMemberRole(ctx, orgID, userID, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, "admin", member.Role)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid role", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-456"
		req := &organization.UpdateMemberRoleRequest{
			Role: "invalid-role",
		}

		// Execute
		member, err := service.UpdateMemberRole(ctx, orgID, userID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "invalid role")
	})

	t.Run("cannot change owner role from admin", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		ownerID := "user-123"
		req := &organization.UpdateMemberRoleRequest{
			Role: "member", // Trying to downgrade owner
		}

		org := &organization.Organization{
			ID:      orgID,
			OwnerID: ownerID,
		}

		// Mock expectations
		mockRepo.On("GetOrganization", ctx, orgID).Return(org, nil)

		// Execute
		member, err := service.UpdateMemberRole(ctx, orgID, ownerID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "cannot change owner role from admin")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetMember(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful member retrieval", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"

		orgUser := &organization.OrganizationUser{
			ID:       "ou-1",
			UserID:   userID,
			Role:     "admin",
			Status:   "active",
			JoinedAt: time.Now(),
		}

		user := &organization.User{
			ID:          userID,
			Email:       "user@example.com",
			DisplayName: "Test User",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(orgUser, nil)
		mockAuthRepo.On("GetUser", ctx, userID).Return(user, nil)

		// Execute
		member, err := service.GetMember(ctx, orgID, userID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, "user@example.com", member.Email)
		assert.Equal(t, "Test User", member.DisplayName)
		assert.Equal(t, "admin", member.Role)

		mockRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}

func TestGetOrganizationStats(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful stats retrieval", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		expectedStats := &organization.OrganizationStats{
			OrganizationID:   orgID,
			TotalMembers:     10,
			ActiveMembers:    8,
			TotalWorkspaces:  5,
			ActiveWorkspaces: 4,
			TotalProjects:    15,
			ResourceUsage: &organization.Usage{
				CPU:     4.5,
				Memory:  16.0,
				Storage: 100.0,
				Cost:    250.0,
			},
			LastUpdated: time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetOrganizationStats", ctx, orgID).Return(expectedStats, nil)

		// Execute
		stats, err := service.GetOrganizationStats(ctx, orgID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, expectedStats, stats)

		mockRepo.AssertExpectations(t)
	})
}

func TestValidateOrganizationAccess(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("valid access with required role", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"
		requiredRole := "admin"

		member := &organization.OrganizationUser{
			UserID: userID,
			Role:   "admin",
			Status: "active",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(member, nil)

		// Execute
		err := service.ValidateOrganizationAccess(ctx, userID, orgID, requiredRole)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("access denied - user not a member", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"
		requiredRole := "member"

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(nil, errors.New("not found"))

		// Execute
		err := service.ValidateOrganizationAccess(ctx, userID, orgID, requiredRole)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied: user is not a member")

		mockRepo.AssertExpectations(t)
	})

	t.Run("access denied - inactive member", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"
		requiredRole := "member"

		member := &organization.OrganizationUser{
			UserID: userID,
			Role:   "member",
			Status: "suspended",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(member, nil)

		// Execute
		err := service.ValidateOrganizationAccess(ctx, userID, orgID, requiredRole)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied: member status is suspended")

		mockRepo.AssertExpectations(t)
	})

	t.Run("access denied - insufficient role", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"
		requiredRole := "admin"

		member := &organization.OrganizationUser{
			UserID: userID,
			Role:   "member", // Has member role but admin required
			Status: "active",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(member, nil)

		// Execute
		err := service.ValidateOrganizationAccess(ctx, userID, orgID, requiredRole)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied: admin role required")

		mockRepo.AssertExpectations(t)
	})
}

func TestGetUserRole(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful role retrieval", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"

		member := &organization.OrganizationUser{
			UserID: userID,
			Role:   "admin",
			Status: "active",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(member, nil)

		// Execute
		role, err := service.GetUserRole(ctx, userID, orgID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "admin", role)

		mockRepo.AssertExpectations(t)
	})
}

func TestInvitationManagement(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("get invitation by ID", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-123"
		expectedInvitation := &organization.Invitation{
			ID:     invitationID,
			Email:  "user@example.com",
			Status: "pending",
		}

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(expectedInvitation, nil)

		// Execute
		invitation, err := service.GetInvitation(ctx, invitationID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, expectedInvitation, invitation)

		mockRepo.AssertExpectations(t)
	})

	t.Run("list pending invitations", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		expectedInvitations := []*organization.Invitation{
			{ID: "inv-1", Email: "user1@example.com", Status: "pending"},
			{ID: "inv-2", Email: "user2@example.com", Status: "pending"},
		}

		// Mock expectations
		mockRepo.On("ListInvitations", ctx, orgID, "pending").Return(expectedInvitations, nil)

		// Execute
		invitations, err := service.ListPendingInvitations(ctx, orgID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, invitations)
		assert.Len(t, invitations, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("resend invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-123"
		invitation := &organization.Invitation{
			ID:        invitationID,
			Status:    "pending",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(invitation, nil)
		mockRepo.On("UpdateInvitation", ctx, mock.MatchedBy(func(inv *organization.Invitation) bool {
			return inv.ExpiresAt.After(time.Now())
		})).Return(nil)

		// Execute
		err := service.ResendInvitation(ctx, invitationID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cancel invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-123"
		invitation := &organization.Invitation{
			ID:     invitationID,
			Status: "pending",
		}

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(invitation, nil)
		mockRepo.On("UpdateInvitation", ctx, mock.MatchedBy(func(inv *organization.Invitation) bool {
			return inv.Status == "canceled"
		})).Return(nil)

		// Execute
		err := service.CancelInvitation(ctx, invitationID)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cleanup expired invitations", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		// Mock expectations
		mockRepo.On("DeleteExpiredInvitations", ctx, mock.AnythingOfType("time.Time")).Return(nil)

		// Execute
		err := service.CleanupExpiredInvitations(ctx)

		// Assert
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestErrorScenarios(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("resend invitation not found", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-notfound"

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(nil, errors.New("not found"))

		// Execute
		err := service.ResendInvitation(ctx, invitationID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("cancel invitation not found", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-notfound"

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(nil, errors.New("not found"))

		// Execute
		err := service.CancelInvitation(ctx, invitationID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("get member user details not found", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-123"

		orgUser := &organization.OrganizationUser{
			ID:     "ou-1",
			UserID: userID,
			Role:   "admin",
			Status: "active",
		}

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(orgUser, nil)
		mockAuthRepo.On("GetUser", ctx, userID).Return(nil, errors.New("user not found"))

		// Execute
		member, err := service.GetMember(ctx, orgID, userID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, member)
		assert.Contains(t, err.Error(), "failed to get user details")

		mockRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("get user role member not found", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"
		userID := "user-notfound"

		// Mock expectations
		mockRepo.On("GetMember", ctx, orgID, userID).Return(nil, errors.New("not found"))

		// Execute
		role, err := service.GetUserRole(ctx, userID, orgID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, role)
		assert.Contains(t, err.Error(), "failed to get member")

		mockRepo.AssertExpectations(t)
	})

	t.Run("list workspaces error in delete organization", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		orgID := "org-123"

		// Mock expectations
		mockRepo.On("ListWorkspaces", ctx, orgID).Return([]*organization.WorkspaceInfo(nil), errors.New("database error"))

		// Execute
		err := service.DeleteOrganization(ctx, orgID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check workspaces")

		mockRepo.AssertExpectations(t)
	})

	t.Run("resend non-pending invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-123"
		invitation := &organization.Invitation{
			ID:     invitationID,
			Status: "accepted", // Not pending
		}

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(invitation, nil)

		// Execute
		err := service.ResendInvitation(ctx, invitationID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only resend pending invitations")

		mockRepo.AssertExpectations(t)
	})

	t.Run("cancel non-pending invitation", func(t *testing.T) {
		// Setup mocks
		mockRepo := new(MockRepository)
		mockAuthRepo := new(MockAuthRepository)
		mockBillingRepo := new(MockBillingRepository)

		service := NewService(mockRepo, mockAuthRepo, mockBillingRepo, logger)

		invitationID := "inv-123"
		invitation := &organization.Invitation{
			ID:     invitationID,
			Status: "expired", // Not pending
		}

		// Mock expectations
		mockRepo.On("GetInvitation", ctx, invitationID).Return(invitation, nil)

		// Execute
		err := service.CancelInvitation(ctx, invitationID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only cancel pending invitations")

		mockRepo.AssertExpectations(t)
	})
}