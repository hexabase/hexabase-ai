package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateGroupRequest represents the request payload for creating a group
type CreateGroupRequest struct {
	Name          string  `json:"name" binding:"required"`
	ParentGroupID *string `json:"parent_group_id,omitempty"`
}

// UpdateGroupRequest represents the request payload for updating a group
type UpdateGroupRequest struct {
	Name *string `json:"name,omitempty"`
}

// AddGroupMemberRequest represents the request payload for adding a member to a group
type AddGroupMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GroupHandler handles group and member related endpoints
type GroupHandler struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *GroupHandler {
	return &GroupHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// CreateGroup creates a new group in workspace
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		}
		return
	}

	// Validate group name
	if len(strings.TrimSpace(req.Name)) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
		return
	}

	// Check if group name already exists in workspace
	var existingGroup db.Group
	if err := h.db.First(&existingGroup, "workspace_id = ? AND name = ?", wsID, strings.TrimSpace(req.Name)).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "group with this name already exists"})
		return
	}

	// If parent group is specified, validate it
	if req.ParentGroupID != nil {
		var parentGroup db.Group
		if err := h.db.First(&parentGroup, "id = ? AND workspace_id = ?", *req.ParentGroupID, wsID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "parent group not found"})
			} else {
				h.logger.Error("Failed to get parent group", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get parent group"})
			}
			return
		}

		// Note: No circular reference check needed for new groups
		// since they don't exist in the hierarchy yet
	}

	// Create group
	group := &db.Group{
		WorkspaceID:   wsID,
		Name:          strings.TrimSpace(req.Name),
		ParentGroupID: req.ParentGroupID,
	}

	if err := h.db.Create(group).Error; err != nil {
		h.logger.Error("Failed to create group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create group"})
		return
	}

	h.logger.Info("Group created successfully",
		zap.String("group_id", group.ID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusCreated, group)
}

// ListGroups lists groups in workspace
func (h *GroupHandler) ListGroups(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var groups []db.Group
	if err := h.db.Where("workspace_id = ?", wsID).Find(&groups).Error; err != nil {
		h.logger.Error("Failed to list groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list groups"})
		return
	}

	h.logger.Info("List groups successful",
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Int("count", len(groups)))

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}

// GetGroup gets group details
func (h *GroupHandler) GetGroup(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	h.logger.Info("Get group successful",
		zap.String("group_id", groupID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, group)
}

// UpdateGroup updates group details
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// Get existing group
	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	// Track what needs to be updated
	updates := make(map[string]interface{})

	// Update name if provided
	if req.Name != nil {
		if len(strings.TrimSpace(*req.Name)) < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
			return
		}

		// Check if new name already exists (excluding current group)
		var existingGroup db.Group
		if err := h.db.First(&existingGroup, "workspace_id = ? AND name = ? AND id != ?", wsID, strings.TrimSpace(*req.Name), groupID).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "group with this name already exists"})
			return
		}

		updates["name"] = strings.TrimSpace(*req.Name)
	}

	// Apply updates if any
	if len(updates) > 0 {
		if err := h.db.Model(&group).Updates(updates).Error; err != nil {
			h.logger.Error("Failed to update group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update group"})
			return
		}

		// Reload group to get updated data
		if err := h.db.First(&group, "id = ?", groupID).Error; err != nil {
			h.logger.Error("Failed to reload group", zap.Error(err))
		}
	}

	h.logger.Info("Update group successful",
		zap.String("group_id", groupID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Any("updates", updates))

	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Get existing group
	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	// Check if group has child groups
	var childCount int64
	if err := h.db.Model(&db.Group{}).Where("parent_group_id = ?", groupID).Count(&childCount).Error; err != nil {
		h.logger.Error("Failed to count child groups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check child groups"})
		return
	}

	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete group with child groups"})
		return
	}

	// Check if group has members
	var memberCount int64
	if err := h.db.Model(&db.GroupMembership{}).Where("group_id = ?", groupID).Count(&memberCount).Error; err != nil {
		h.logger.Error("Failed to count group members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check group members"})
		return
	}

	if memberCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete group with existing members"})
		return
	}

	// Delete group
	if err := h.db.Delete(&group).Error; err != nil {
		h.logger.Error("Failed to delete group", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group"})
		return
	}

	h.logger.Info("Delete group successful",
		zap.String("group_id", groupID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "group deleted successfully",
	})
}

// AddGroupMember adds a member to group
func (h *GroupHandler) AddGroupMember(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	var req AddGroupMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if req.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		}
		return
	}

	// Verify group exists and belongs to workspace
	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	// Verify user exists and is a member of the organization
	var orgUser db.OrganizationUser
	if err := h.db.First(&orgUser, "user_id = ? AND organization_id = ?", req.UserID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user not found in organization"})
		} else {
			h.logger.Error("Failed to check organization membership", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check organization membership"})
		}
		return
	}

	// Check if user is already in group
	var existingMembership db.GroupMembership
	if err := h.db.First(&existingMembership, "group_id = ? AND user_id = ?", groupID, req.UserID).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already in group"})
		return
	}

	// Add user to group
	membership := &db.GroupMembership{
		GroupID:  groupID,
		UserID:   req.UserID,
		JoinedAt: time.Now(),
	}

	if err := h.db.Create(membership).Error; err != nil {
		h.logger.Error("Failed to add group member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add group member"})
		return
	}

	h.logger.Info("Add group member successful",
		zap.String("group_id", groupID),
		zap.String("member_user_id", req.UserID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "user added to group successfully",
	})
}

// RemoveGroupMember removes a member from group
func (h *GroupHandler) RemoveGroupMember(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	memberUserID := c.Param("userId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Verify group exists and belongs to workspace
	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	// Get and delete membership
	var membership db.GroupMembership
	if err := h.db.First(&membership, "group_id = ? AND user_id = ?", groupID, memberUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found in group"})
		} else {
			h.logger.Error("Failed to get group membership", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group membership"})
		}
		return
	}

	if err := h.db.Delete(&membership).Error; err != nil {
		h.logger.Error("Failed to remove group member", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove group member"})
		return
	}

	h.logger.Info("Remove group member successful",
		zap.String("group_id", groupID),
		zap.String("member_user_id", memberUserID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)))

	c.JSON(http.StatusOK, gin.H{
		"message": "user removed from group successfully",
	})
}

// ListGroupMembers lists members in a group
func (h *GroupHandler) ListGroupMembers(c *gin.Context) {
	orgID := c.Param("orgId")
	wsID := c.Param("wsId")
	groupID := c.Param("groupId")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Verify user has access to the organization
	if !h.hasOrgAccess(userID.(string), orgID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to access this organization"})
		return
	}

	// Verify workspace exists and belongs to organization
	var workspace db.Workspace
	if err := h.db.First(&workspace, "id = ? AND organization_id = ?", wsID, orgID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		} else {
			h.logger.Error("Failed to get workspace", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace"})
		}
		return
	}

	// Verify group exists and belongs to workspace
	var group db.Group
	if err := h.db.First(&group, "id = ? AND workspace_id = ?", groupID, wsID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		} else {
			h.logger.Error("Failed to get group", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group"})
		}
		return
	}

	// Get group members with user details
	var memberships []struct {
		db.GroupMembership
		User db.User `json:"user"`
	}

	if err := h.db.Table("group_memberships").
		Select("group_memberships.*, users.*").
		Joins("JOIN users ON users.id = group_memberships.user_id").
		Where("group_memberships.group_id = ?", groupID).
		Scan(&memberships).Error; err != nil {
		h.logger.Error("Failed to list group members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list group members"})
		return
	}

	h.logger.Info("List group members successful",
		zap.String("group_id", groupID),
		zap.String("workspace_id", wsID),
		zap.String("user_id", userID.(string)),
		zap.Int("count", len(memberships)))

	c.JSON(http.StatusOK, gin.H{
		"members": memberships,
		"total":   len(memberships),
	})
}

// hasOrgAccess checks if user has access to the organization
func (h *GroupHandler) hasOrgAccess(userID, orgID string) bool {
	var membership db.OrganizationUser
	err := h.db.First(&membership, "user_id = ? AND organization_id = ?", userID, orgID).Error
	return err == nil
}

// wouldCreateCircularReference checks if adding parentGroupID as parent would create a circular reference
func (h *GroupHandler) wouldCreateCircularReference(parentGroupID string, currentGroup *db.Group) bool {
	// Walk up the parent chain from the proposed parent
	checkGroupID := parentGroupID
	for checkGroupID != "" {
		if checkGroupID == currentGroup.ID {
			return true // Found circular reference
		}

		var parentGroup db.Group
		if err := h.db.First(&parentGroup, "id = ?", checkGroupID).Error; err != nil {
			break // Parent not found, stop checking
		}

		if parentGroup.ParentGroupID == nil {
			break // Reached root group
		}

		checkGroupID = *parentGroup.ParentGroupID
	}

	return false
}