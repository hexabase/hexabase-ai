package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/organization"
	"go.uber.org/zap"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	service organization.Service
	logger  *zap.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(service organization.Service, logger *zap.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
		logger:  logger,
	}
}

// CreateOrganization handles organization creation
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	userID := c.GetString("user_id")

	var req organization.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.OwnerID = userID

	org, err := h.service.CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization created",
		zap.String("org_id", org.ID),
		zap.String("user_id", userID))

	c.JSON(http.StatusCreated, org)
}

// GetOrganization handles getting an organization
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	org, err := h.service.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get organization", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations handles listing user's organizations
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	userID := c.GetString("user_id")

	orgs, err := h.service.ListOrganizations(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list organizations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgs,
		"total":         len(orgs),
	})
}

// UpdateOrganization handles updating an organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.GetString("user_id")

	var req organization.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.UpdatedBy = userID

	org, err := h.service.UpdateOrganization(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to update organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization updated",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles deleting an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.GetString("user_id")

	err := h.service.DeleteOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to delete organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization deleted",
		zap.String("org_id", orgID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}

// AddMember handles adding a member to organization
func (h *OrganizationHandler) AddMember(c *gin.Context) {
	orgID := c.Param("orgId")
	addedBy := c.GetString("user_id")

	var req organization.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.AddedBy = addedBy

	err := h.service.AddMember(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to add member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

// RemoveMember handles removing a member from organization
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	err := h.service.RemoveMember(c.Request.Context(), orgID, userID)
	if err != nil {
		h.logger.Error("failed to remove member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed successfully"})
}

// UpdateMemberRole handles updating a member's role
func (h *OrganizationHandler) UpdateMemberRole(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	var req struct {
		Role string `json:"role" binding:"required,oneof=admin member"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	err := h.service.UpdateMemberRole(c.Request.Context(), orgID, userID, req.Role)
	if err != nil {
		h.logger.Error("failed to update member role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member role updated successfully"})
}

// ListMembers handles listing organization members
func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	orgID := c.Param("orgId")

	members, err := h.service.ListMembers(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   len(members),
	})
}

// CreateInvitation handles creating an invitation
func (h *OrganizationHandler) CreateInvitation(c *gin.Context) {
	orgID := c.Param("orgId")
	invitedBy := c.GetString("user_id")

	var req organization.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	req.InvitedBy = invitedBy

	invitation, err := h.service.CreateInvitation(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create invitation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// AcceptInvitation handles accepting an invitation
func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")
	userID := c.GetString("user_id")

	err := h.service.AcceptInvitation(c.Request.Context(), invitationID, userID)
	if err != nil {
		h.logger.Error("failed to accept invitation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation accepted successfully"})
}

// RevokeInvitation handles revoking an invitation
func (h *OrganizationHandler) RevokeInvitation(c *gin.Context) {
	orgID := c.Param("orgId")
	invitationID := c.Param("invitationId")

	err := h.service.RevokeInvitation(c.Request.Context(), invitationID)
	if err != nil {
		h.logger.Error("failed to revoke invitation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation revoked successfully"})
}

// ListInvitations handles listing organization invitations
func (h *OrganizationHandler) ListInvitations(c *gin.Context) {
	orgID := c.Param("orgId")

	invitations, err := h.service.ListInvitations(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list invitations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       len(invitations),
	})
}

// GetActivityLogs handles getting organization activity logs
func (h *OrganizationHandler) GetActivityLogs(c *gin.Context) {
	orgID := c.Param("orgId")

	var filter organization.ActivityFilter
	// Parse query parameters for filtering

	logs, err := h.service.GetActivityLogs(c.Request.Context(), orgID, filter)
	if err != nil {
		h.logger.Error("failed to get activity logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": logs,
		"total":      len(logs),
	})
}