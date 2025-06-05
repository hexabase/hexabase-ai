package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/project"
	"go.uber.org/zap"
)

type service struct {
	repo      project.Repository
	k8sRepo   project.KubernetesRepository
	logger    *zap.Logger
}

// NewService creates a new project service
func NewService(
	repo project.Repository,
	k8sRepo project.KubernetesRepository,
	logger *zap.Logger,
) project.Service {
	return &service{
		repo:    repo,
		k8sRepo: k8sRepo,
		logger:  logger,
	}
}

func (s *service) CreateProject(ctx context.Context, req *project.CreateProjectRequest) (*project.Project, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Validate name format (RFC 1123)
	if !isValidProjectName(req.Name) {
		return nil, fmt.Errorf("invalid project name: must be lowercase alphanumeric or '-', and must start and end with alphanumeric")
	}

	// Check if project name is unique within workspace
	existing, err := s.repo.GetProjectByName(ctx, req.WorkspaceID, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("project with name %s already exists in workspace", req.Name)
	}

	// Create project
	proj := &project.Project{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Settings:    req.Settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if proj.DisplayName == "" {
		proj.DisplayName = proj.Name
	}

	if err := s.repo.CreateProject(ctx, proj); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create namespace in vCluster
	namespace := &project.Namespace{
		ProjectID: proj.ID,
		Name:      proj.Name,
		Labels: map[string]string{
			"hexabase.io/project-id":   proj.ID,
			"hexabase.io/workspace-id": proj.WorkspaceID,
			"hexabase.io/managed":      "true",
		},
	}

	if err := s.k8sRepo.CreateNamespace(ctx, req.WorkspaceID, namespace); err != nil {
		s.logger.Error("failed to create namespace", zap.Error(err))
		// Don't fail project creation if namespace creation fails
		// It can be retried later
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "project_created", fmt.Sprintf("Created project %s", proj.Name), req.CreatedBy)

	return proj, nil
}

func (s *service) GetProject(ctx context.Context, projectID string) (*project.Project, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get namespace status
	namespace, err := s.k8sRepo.GetNamespace(ctx, proj.WorkspaceID, proj.Name)
	if err != nil {
		s.logger.Warn("failed to get namespace status", zap.Error(err))
	} else {
		proj.NamespaceStatus = namespace.Status
	}

	// Get resource usage
	usage, err := s.k8sRepo.GetNamespaceResourceUsage(ctx, proj.WorkspaceID, proj.Name)
	if err != nil {
		s.logger.Warn("failed to get namespace resource usage", zap.Error(err))
	} else {
		proj.ResourceUsage = usage
	}

	return proj, nil
}

func (s *service) ListProjects(ctx context.Context, workspaceID string) ([]*project.Project, error) {
	projects, err := s.repo.ListProjects(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Get namespace status for each project
	for _, proj := range projects {
		namespace, err := s.k8sRepo.GetNamespace(ctx, workspaceID, proj.Name)
		if err != nil {
			s.logger.Warn("failed to get namespace status", 
				zap.String("project_id", proj.ID),
				zap.Error(err))
			continue
		}
		proj.NamespaceStatus = namespace.Status
	}

	return projects, nil
}

func (s *service) UpdateProject(ctx context.Context, projectID string, req *project.UpdateProjectRequest) (*project.Project, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Update fields
	if req.DisplayName != "" {
		proj.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		proj.Description = req.Description
	}
	if req.Settings != nil {
		proj.Settings = req.Settings
	}

	proj.UpdatedAt = time.Now()

	if err := s.repo.UpdateProject(ctx, proj); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "project_updated", fmt.Sprintf("Updated project %s", proj.Name), req.UpdatedBy)

	return proj, nil
}

func (s *service) DeleteProject(ctx context.Context, projectID string) error {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Delete namespace from vCluster
	if err := s.k8sRepo.DeleteNamespace(ctx, proj.WorkspaceID, proj.Name); err != nil {
		s.logger.Error("failed to delete namespace", zap.Error(err))
		// Continue with project deletion even if namespace deletion fails
	}

	// Delete project
	if err := s.repo.DeleteProject(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "project_deleted", fmt.Sprintf("Deleted project %s", proj.Name), "")

	return nil
}

func (s *service) CreateSubProject(ctx context.Context, parentID string, req *project.CreateProjectRequest) (*project.Project, error) {
	// Get parent project
	parent, err := s.repo.GetProject(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("parent project not found: %w", err)
	}

	// Ensure workspace ID matches
	req.WorkspaceID = parent.WorkspaceID
	
	// Create sub-project
	subProj, err := s.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	// Set parent relationship
	subProj.ParentID = &parentID
	subProj.UpdatedAt = time.Now()

	if err := s.repo.UpdateProject(ctx, subProj); err != nil {
		// Rollback: delete the created project
		s.repo.DeleteProject(ctx, subProj.ID)
		return nil, fmt.Errorf("failed to set parent relationship: %w", err)
	}

	// Apply HNC (Hierarchical Namespace Controller) configuration
	if err := s.k8sRepo.ConfigureHNC(ctx, parent.WorkspaceID, parent.Name, subProj.Name); err != nil {
		s.logger.Error("failed to configure HNC", zap.Error(err))
	}

	return subProj, nil
}

func (s *service) GetProjectHierarchy(ctx context.Context, projectID string) (*project.ProjectHierarchy, error) {
	root, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	hierarchy := &project.ProjectHierarchy{
		Project:  root,
		Children: []*project.ProjectHierarchy{},
	}

	// Recursively get children
	if err := s.buildHierarchy(ctx, hierarchy); err != nil {
		return nil, fmt.Errorf("failed to build hierarchy: %w", err)
	}

	return hierarchy, nil
}

func (s *service) buildHierarchy(ctx context.Context, node *project.ProjectHierarchy) error {
	children, err := s.repo.GetChildProjects(ctx, node.Project.ID)
	if err != nil {
		return err
	}

	for _, child := range children {
		childNode := &project.ProjectHierarchy{
			Project:  child,
			Children: []*project.ProjectHierarchy{},
		}

		// Recursively build children
		if err := s.buildHierarchy(ctx, childNode); err != nil {
			return err
		}

		node.Children = append(node.Children, childNode)
	}

	return nil
}

func (s *service) ApplyResourceQuota(ctx context.Context, projectID string, quota *project.ResourceQuota) error {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Apply quota to namespace
	if err := s.k8sRepo.ApplyResourceQuota(ctx, proj.WorkspaceID, proj.Name, quota); err != nil {
		return fmt.Errorf("failed to apply resource quota: %w", err)
	}

	// Store quota in database
	if proj.Settings == nil {
		proj.Settings = make(map[string]interface{})
	}
	proj.Settings["resource_quota"] = quota
	proj.UpdatedAt = time.Now()

	if err := s.repo.UpdateProject(ctx, proj); err != nil {
		return fmt.Errorf("failed to update project settings: %w", err)
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "quota_applied", 
		fmt.Sprintf("Applied resource quota to project %s", proj.Name), "")

	return nil
}

func (s *service) GetResourceUsage(ctx context.Context, projectID string) (map[string]interface{}, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	usage, err := s.k8sRepo.GetNamespaceResourceUsage(ctx, proj.WorkspaceID, proj.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage: %w", err)
	}

	return usage, nil
}

func (s *service) AddProjectMember(ctx context.Context, projectID string, req *project.AddMemberRequest) error {
	// Check if project exists
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListProjectMembers(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.UserID == req.UserID {
			return fmt.Errorf("user is already a project member")
		}
	}

	// Add member
	member := &project.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
		AddedAt:   time.Now(),
	}

	if err := s.repo.AddProjectMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Apply RBAC in namespace
	if err := s.k8sRepo.ApplyRBAC(ctx, proj.WorkspaceID, proj.Name, req.UserID, req.Role); err != nil {
		s.logger.Error("failed to apply RBAC", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "member_added", 
		fmt.Sprintf("Added user %s to project %s with role %s", req.UserID, proj.Name, req.Role), req.AddedBy)

	return nil
}

func (s *service) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Remove member
	if err := s.repo.RemoveProjectMember(ctx, projectID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Remove RBAC from namespace
	if err := s.k8sRepo.RemoveRBAC(ctx, proj.WorkspaceID, proj.Name, userID); err != nil {
		s.logger.Error("failed to remove RBAC", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, proj.WorkspaceID, "member_removed", 
		fmt.Sprintf("Removed user %s from project %s", userID, proj.Name), "")

	return nil
}

func (s *service) ListProjectMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	return s.repo.ListProjectMembers(ctx, projectID)
}

func (s *service) GetActivityLogs(ctx context.Context, projectID string, filter project.ActivityFilter) ([]*project.Activity, error) {
	filter.ProjectID = projectID
	return s.repo.ListActivities(ctx, filter)
}

// Helper functions

func isValidProjectName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}

	// Must start and end with alphanumeric
	if !isAlphaNumeric(name[0]) || !isAlphaNumeric(name[len(name)-1]) {
		return false
	}

	// Must contain only lowercase alphanumeric or '-'
	for _, c := range name {
		if !isAlphaNumeric(byte(c)) && c != '-' {
			return false
		}
	}

	return true
}

func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

func (s *service) logActivity(ctx context.Context, workspaceID, activityType, description, userID string) {
	activity := &project.Activity{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		Type:        activityType,
		Description: description,
		UserID:      userID,
		Timestamp:   time.Now(),
	}

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		s.logger.Error("failed to log activity", zap.Error(err))
	}
}