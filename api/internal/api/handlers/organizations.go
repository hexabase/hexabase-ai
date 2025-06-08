package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/domain/organization"
)

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	service organization.Service
	logger  *slog.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(service organization.Service, logger *slog.Logger) *OrganizationHandler {
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

	org, err := h.service.CreateOrganization(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("failed to create organization", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization created",
		"org_id", org.ID,
		"user_id", userID)

	c.JSON(http.StatusCreated, org)
}

// GetOrganization handles getting an organization
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID := c.Param("orgId")

	org, err := h.service.GetOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get organization", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations handles listing user's organizations
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	userID := c.GetString("user_id")

	filter := organization.OrganizationFilter{
		UserID:   userID,
		Page:     1,
		PageSize: 100,
	}

	orgList, err := h.service.ListOrganizations(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list organizations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"organizations": orgList.Organizations,
		"total":         orgList.Total,
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


	org, err := h.service.UpdateOrganization(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to update organization", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization updated",
		"org_id", orgID,
		"user_id", userID)

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles deleting an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.GetString("user_id")

	err := h.service.DeleteOrganization(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to delete organization", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("organization deleted",
		"org_id", orgID,
		"user_id", userID)

	c.JSON(http.StatusOK, gin.H{"message": "organization deleted successfully"})
}


// RemoveMember handles removing a member from organization
func (h *OrganizationHandler) RemoveMember(c *gin.Context) {
	orgID := c.Param("orgId")
	userID := c.Param("userId")

	removerID := c.GetString("user_id")
	err := h.service.RemoveMember(c.Request.Context(), orgID, userID, removerID)
	if err != nil {
		h.logger.Error("failed to remove member", "error", err)
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

	updateReq := &organization.UpdateMemberRoleRequest{
		Role: req.Role,
	}

	member, err := h.service.UpdateMemberRole(c.Request.Context(), orgID, userID, updateReq)
	if err != nil {
		h.logger.Error("failed to update member role", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// ListMembers handles listing organization members
func (h *OrganizationHandler) ListMembers(c *gin.Context) {
	orgID := c.Param("orgId")

	filter := organization.MemberFilter{
		OrganizationID: orgID,
		Page:           1,
		PageSize:       100,
	}

	memberList, err := h.service.ListMembers(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list members", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": memberList.Members,
		"total":   memberList.Total,
	})
}

// InviteUser handles inviting a user to organization
func (h *OrganizationHandler) InviteUser(c *gin.Context) {
	orgID := c.Param("orgId")
	inviterID := c.GetString("user_id")

	var req organization.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	invitation, err := h.service.InviteUser(c.Request.Context(), orgID, inviterID, &req)
	if err != nil {
		h.logger.Error("failed to invite user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// AcceptInvitation handles accepting an invitation
func (h *OrganizationHandler) AcceptInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")
	userID := c.GetString("user_id")

	member, err := h.service.AcceptInvitation(c.Request.Context(), invitationID, userID)
	if err != nil {
		h.logger.Error("failed to accept invitation", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// CancelInvitation handles canceling an invitation
func (h *OrganizationHandler) CancelInvitation(c *gin.Context) {
	invitationID := c.Param("invitationId")

	err := h.service.CancelInvitation(c.Request.Context(), invitationID)
	if err != nil {
		h.logger.Error("failed to cancel invitation", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation canceled successfully"})
}

// ListPendingInvitations handles listing pending invitations
func (h *OrganizationHandler) ListPendingInvitations(c *gin.Context) {
	orgID := c.Param("orgId")

	invitations, err := h.service.ListPendingInvitations(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list invitations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       len(invitations),
	})
}

// GetOrganizationStats handles getting organization statistics
func (h *OrganizationHandler) GetOrganizationStats(c *gin.Context) {
	orgID := c.Param("orgId")

	stats, err := h.service.GetOrganizationStats(c.Request.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get organization stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}