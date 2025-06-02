package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrganizationHandler handles organization-related endpoints
type OrganizationHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name string `json:"name" binding:"omitempty,min=3,max=100"`
}

// OrganizationResponse represents an organization in API responses
type OrganizationResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      string    `json:"role,omitempty"` // User's role in this org
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create organization
	org := &db.Organization{
		Name: req.Name,
	}

	if err := tx.Create(org).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to create organization", 
			zap.Error(err),
			zap.String("name", req.Name))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	// Add creator as admin
	orgUser := &db.OrganizationUser{
		OrganizationID: org.ID,
		UserID:         userID.(string),
		Role:           "admin",
	}

	if err := tx.Create(orgUser).Error; err != nil {
		tx.Rollback()
		h.logger.Error("Failed to add user to organization", 
			zap.Error(err),
			zap.String("org_id", org.ID),
			zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.logger.Error("Failed to commit transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create organization"})
		return
	}

	h.logger.Info("Organization created",
		zap.String("org_id", org.ID),
		zap.String("name", org.Name),
		zap.String("creator_id", userID.(string)))

	c.JSON(http.StatusCreated, OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
		Role:      "admin",
	})
}

// ListOrganizations lists organizations for the current user
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Query organizations with user's role
	var orgUsers []db.OrganizationUser
	if err := h.db.Preload("Organization").Where("user_id = ?", userID).Find(&orgUsers).Error; err != nil {
		h.logger.Error("Failed to list organizations", 
			zap.Error(err),
			zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations"})
		return
	}

	// Convert to response format
	organizations := make([]OrganizationResponse, len(orgUsers))
	for i, ou := range orgUsers {
		organizations[i] = OrganizationResponse{
			ID:        ou.Organization.ID,
			Name:      ou.Organization.Name,
			CreatedAt: ou.Organization.CreatedAt,
			UpdatedAt: ou.Organization.UpdatedAt,
			Role:      ou.Role,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": organizations,
		"total":         len(organizations),
	})
}

// GetOrganization gets organization details
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orgID := c.Param("orgId")

	// Check if user has access to this organization
	var orgUser db.OrganizationUser
	err := h.db.Preload("Organization").
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&orgUser).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	} else if err != nil {
		h.logger.Error("Failed to get organization", 
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get organization"})
		return
	}

	c.JSON(http.StatusOK, OrganizationResponse{
		ID:        orgUser.Organization.ID,
		Name:      orgUser.Organization.Name,
		CreatedAt: orgUser.Organization.CreatedAt,
		UpdatedAt: orgUser.Organization.UpdatedAt,
		Role:      orgUser.Role,
	})
}

// UpdateOrganization updates organization details
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orgID := c.Param("orgId")

	// Check if user is admin of this organization
	var orgUser db.OrganizationUser
	err := h.db.Where("organization_id = ? AND user_id = ? AND role = ?", orgID, userID, "admin").
		First(&orgUser).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	} else if err != nil {
		h.logger.Error("Failed to check permissions", 
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update organization
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	// Get organization for response
	var org db.Organization
	if err := h.db.Model(&org).Where("id = ?", orgID).Updates(updates).Error; err != nil {
		h.logger.Error("Failed to update organization", 
			zap.Error(err),
			zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	// Fetch updated organization
	if err := h.db.First(&org, "id = ?", orgID).Error; err != nil {
		h.logger.Error("Failed to fetch updated organization", 
			zap.Error(err),
			zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update organization"})
		return
	}

	h.logger.Info("Organization updated",
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)),
		zap.Any("updates", updates))

	c.JSON(http.StatusOK, OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
		Role:      "admin",
	})
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	orgID := c.Param("orgId")

	// Check if user is admin of this organization
	var orgUser db.OrganizationUser
	err := h.db.Where("organization_id = ? AND user_id = ? AND role = ?", orgID, userID, "admin").
		First(&orgUser).Error

	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	} else if err != nil {
		h.logger.Error("Failed to check permissions", 
			zap.Error(err),
			zap.String("org_id", orgID),
			zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	// Check if organization has workspaces
	var workspaceCount int64
	if err := h.db.Model(&db.Workspace{}).Where("organization_id = ?", orgID).Count(&workspaceCount).Error; err != nil {
		h.logger.Error("Failed to check workspaces", 
			zap.Error(err),
			zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	if workspaceCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete organization with active workspaces"})
		return
	}

	// Delete organization (cascades to organization_users)
	if err := h.db.Delete(&db.Organization{}, "id = ?", orgID).Error; err != nil {
		h.logger.Error("Failed to delete organization", 
			zap.Error(err),
			zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete organization"})
		return
	}

	h.logger.Info("Organization deleted",
		zap.String("org_id", orgID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

// InviteUser invites a user to the organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("Invite user endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// ListUsers lists users in the organization
func (h *OrganizationHandler) ListUsers(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("List users endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// RemoveUser removes a user from the organization
func (h *OrganizationHandler) RemoveUser(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")
	h.logger.Info("Remove user endpoint called", 
		zap.String("org_id", orgID),
		zap.String("user_id", userID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// CreatePortalSession creates Stripe billing portal session
func (h *OrganizationHandler) CreatePortalSession(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("Create portal session endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// GetSubscriptions gets organization subscriptions
func (h *OrganizationHandler) GetSubscriptions(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("Get subscriptions endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// GetPaymentMethods gets organization payment methods
func (h *OrganizationHandler) GetPaymentMethods(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("Get payment methods endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}

// CreateSetupIntent creates Stripe setup intent for payment methods
func (h *OrganizationHandler) CreateSetupIntent(c *gin.Context) {
	orgID := c.Param("orgId")
	h.logger.Info("Create setup intent endpoint called", zap.String("org_id", orgID))
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "endpoint not implemented yet",
	})
}