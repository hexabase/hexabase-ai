package project

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/project"
	"log/slog"
)

type service struct {
	repo      project.Repository
	k8sRepo   project.KubernetesRepository
	logger    *slog.Logger
}

// NewService creates a new project service
func NewService(
	repo project.Repository,
	k8sRepo project.KubernetesRepository,
	logger *slog.Logger,
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
	existing, err := s.repo.GetProjectByNameAndWorkspace(ctx, req.Name, req.WorkspaceID)
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
	labels := map[string]string{
		"hexabase.ai/project-id":   proj.ID,
		"hexabase.ai/workspace-id": proj.WorkspaceID,
		"hexabase.ai/managed":      "true",
	}

	if err := s.k8sRepo.CreateNamespace(ctx, req.WorkspaceID, proj.Name, labels); err != nil {
		s.logger.Error("failed to create namespace", "error", err)
		// Don't fail project creation if namespace creation fails
		// It can be retried later
	} else {
		// Update project with namespace name
		proj.NamespaceName = proj.Name
		s.repo.UpdateProject(ctx, proj)
	}

	// Log activity
	s.logActivity(ctx, proj.ID, "project_created", fmt.Sprintf("Created project %s", proj.Name), "")

	return proj, nil
}

func (s *service) GetProject(ctx context.Context, projectID string) (*project.Project, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get namespace status
	if proj.NamespaceName != "" {
		namespace, err := s.k8sRepo.GetNamespace(ctx, proj.WorkspaceID, proj.NamespaceName)
		if err != nil {
			s.logger.Warn("failed to get namespace status", "error", err)
		} else {
			if status, ok := namespace["status"].(string); ok {
				proj.Status = status
			}
		}
	}

	// Get resource usage
	if proj.NamespaceName != "" {
		usage, err := s.k8sRepo.GetNamespaceResourceUsage(ctx, proj.WorkspaceID, proj.NamespaceName)
		if err != nil {
			s.logger.Warn("failed to get namespace resource usage", "error", err)
		} else {
			proj.ResourceUsage = usage
		}
	}

	return proj, nil
}

func (s *service) ListProjects(ctx context.Context, filter project.ProjectFilter) (*project.ProjectList, error) {
	projects, total, err := s.repo.ListProjects(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Get namespace status for each project
	for _, proj := range projects {
		if proj.NamespaceName != "" {
			namespace, err := s.k8sRepo.GetNamespace(ctx, proj.WorkspaceID, proj.NamespaceName)
			if err != nil {
				s.logger.Warn("failed to get namespace status", 
					"project_id", proj.ID,
					"error", err)
				continue
			}
			// Set status based on namespace info
			if status, ok := namespace["status"].(string); ok {
				proj.Status = status
			}
		}
	}

	return &project.ProjectList{
		Projects: projects,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
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
	s.logActivity(ctx, proj.ID, "project_updated", fmt.Sprintf("Updated project %s", proj.Name), "")

	return proj, nil
}

func (s *service) DeleteProject(ctx context.Context, projectID string) error {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Delete namespace from vCluster
	if err := s.k8sRepo.DeleteNamespace(ctx, proj.WorkspaceID, proj.Name); err != nil {
		s.logger.Error("failed to delete namespace", "error", err)
		// Continue with project deletion even if namespace deletion fails
	}

	// Delete project
	if err := s.repo.DeleteProject(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Log activity
	s.logActivity(ctx, proj.ID, "project_deleted", fmt.Sprintf("Deleted project %s", proj.Name), "")

	return nil
}

// Namespace management methods

func (s *service) CreateNamespace(ctx context.Context, projectID string, req *project.CreateNamespaceRequest) (*project.Namespace, error) {
	// Get project
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Generate namespace name if not provided
	namespaceName := req.Name
	if namespaceName == "" {
		namespaceName = fmt.Sprintf("%s-%s", proj.Name, generateID()[:8])
	}

	// Create namespace in Kubernetes
	labels := map[string]string{
		"project-id":   projectID,
		"workspace-id": proj.WorkspaceID,
	}
	if req.Labels != nil {
		for k, v := range req.Labels {
			labels[k] = v
		}
	}

	if err := s.k8sRepo.CreateNamespace(ctx, proj.WorkspaceID, namespaceName, labels); err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	// Apply resource quota if provided
	if req.ResourceQuota != nil {
		if err := s.k8sRepo.CreateResourceQuota(ctx, proj.WorkspaceID, namespaceName, req.ResourceQuota); err != nil {
			s.logger.Error("failed to apply resource quota", "error", err)
		}
	}

	// Create namespace record
	namespace := &project.Namespace{
		ID:            generateID(),
		Name:          namespaceName,
		ProjectID:     projectID,
		Description:   req.Description,
		Status:        "active",
		ResourceQuota: req.ResourceQuota,
		Labels:        labels,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateNamespace(ctx, namespace); err != nil {
		// Rollback: delete the created namespace
		s.k8sRepo.DeleteNamespace(ctx, proj.WorkspaceID, namespaceName)
		return nil, fmt.Errorf("failed to save namespace: %w", err)
	}

	// Log activity
	activity := &project.ProjectActivity{
		ID:          generateID(),
		ProjectID:   projectID,
		Type:        "namespace_created",
		Description: fmt.Sprintf("Created namespace %s", namespaceName),
		CreatedAt:   time.Now(),
	}
	s.repo.CreateActivity(ctx, activity)

	return namespace, nil
}

func (s *service) GetNamespace(ctx context.Context, projectID, namespaceID string) (*project.Namespace, error) {
	namespace, err := s.repo.GetNamespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}

	// Verify namespace belongs to project
	if namespace.ProjectID != projectID {
		return nil, fmt.Errorf("namespace does not belong to project")
	}

	// Get resource usage from Kubernetes
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	usage, err := s.k8sRepo.GetNamespaceUsage(ctx, proj.WorkspaceID, namespace.Name)
	if err != nil {
		s.logger.Error("failed to get namespace usage", "error", err)
	} else {
		namespace.ResourceUsage = usage
	}

	return namespace, nil
}

func (s *service) ListNamespaces(ctx context.Context, projectID string) (*project.NamespaceList, error) {
	namespaces, err := s.repo.ListNamespaces(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &project.NamespaceList{
		Namespaces: namespaces,
		Total:      len(namespaces),
	}, nil
}

func (s *service) UpdateNamespace(ctx context.Context, projectID, namespaceID string, req *project.CreateNamespaceRequest) (*project.Namespace, error) {
	namespace, err := s.repo.GetNamespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}

	// Verify namespace belongs to project
	if namespace.ProjectID != projectID {
		return nil, fmt.Errorf("namespace does not belong to project")
	}

	// Update fields
	if req.Description != "" {
		namespace.Description = req.Description
	}
	if req.Labels != nil {
		namespace.Labels = req.Labels
	}
	namespace.UpdatedAt = time.Now()

	// Update resource quota if provided
	if req.ResourceQuota != nil {
		proj, err := s.repo.GetProject(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}

		if err := s.k8sRepo.UpdateResourceQuota(ctx, proj.WorkspaceID, namespace.Name, req.ResourceQuota); err != nil {
			return nil, fmt.Errorf("failed to update resource quota: %w", err)
		}
		namespace.ResourceQuota = req.ResourceQuota
	}

	if err := s.repo.UpdateNamespace(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to update namespace: %w", err)
	}

	return namespace, nil
}

func (s *service) DeleteNamespace(ctx context.Context, projectID, namespaceID string) error {
	namespace, err := s.repo.GetNamespace(ctx, namespaceID)
	if err != nil {
		return err
	}

	// Verify namespace belongs to project
	if namespace.ProjectID != projectID {
		return fmt.Errorf("namespace does not belong to project")
	}

	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Delete namespace from Kubernetes
	if err := s.k8sRepo.DeleteNamespace(ctx, proj.WorkspaceID, namespace.Name); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	// Delete namespace record
	if err := s.repo.DeleteNamespace(ctx, namespaceID); err != nil {
		return fmt.Errorf("failed to delete namespace record: %w", err)
	}

	// Log activity
	activity := &project.ProjectActivity{
		ID:          generateID(),
		ProjectID:   projectID,
		Type:        "namespace_deleted",
		Description: fmt.Sprintf("Deleted namespace %s", namespace.Name),
		CreatedAt:   time.Now(),
	}
	s.repo.CreateActivity(ctx, activity)

	return nil
}

func (s *service) GetNamespaceUsage(ctx context.Context, projectID, namespaceID string) (*project.NamespaceUsage, error) {
	namespace, err := s.repo.GetNamespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}

	// Verify namespace belongs to project
	if namespace.ProjectID != projectID {
		return nil, fmt.Errorf("namespace does not belong to project")
	}

	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.k8sRepo.GetNamespaceUsage(ctx, proj.WorkspaceID, namespace.Name)
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
		s.logger.Error("failed to configure HNC", "error", err)
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
	s.logActivity(ctx, proj.ID, "quota_applied", 
		fmt.Sprintf("Applied resource quota to project %s", proj.Name), "")

	return nil
}

func (s *service) GetResourceUsage(ctx context.Context, projectID string) (*project.ResourceUsage, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	usage, err := s.k8sRepo.GetNamespaceResourceUsage(ctx, proj.WorkspaceID, proj.NamespaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource usage: %w", err)
	}

	return usage, nil
}

func (s *service) AddMember(ctx context.Context, projectID, adderID string, req *project.AddMemberRequest) (*project.ProjectMember, error) {
	// Check if project exists
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.UserEmail)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.UserID == user.ID {
			return nil, fmt.Errorf("user is already a project member")
		}
	}

	// Add member
	member := &project.ProjectMember{
		ID:        generateID(),
		ProjectID: projectID,
		UserID:    user.ID,
		UserEmail: user.Email,
		UserName:  user.DisplayName,
		Role:      req.Role,
		AddedBy:   adderID,
		AddedAt:   time.Now(),
	}

	if err := s.repo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Apply RBAC in namespace
	if err := s.k8sRepo.ApplyRBAC(ctx, proj.WorkspaceID, proj.NamespaceName, user.ID, req.Role); err != nil {
		s.logger.Error("failed to apply RBAC", "error", err)
	}

	// Log activity
	activity := &project.ProjectActivity{
		ID:          generateID(),
		ProjectID:   projectID,
		Type:        "member_added",
		Description: fmt.Sprintf("Added user %s to project with role %s", user.Email, req.Role),
		UserID:      adderID,
		CreatedAt:   time.Now(),
	}
	s.repo.CreateActivity(ctx, activity)

	return member, nil
}

func (s *service) GetMember(ctx context.Context, projectID, memberID string) (*project.ProjectMember, error) {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member belongs to project
	if member.ProjectID != projectID {
		return nil, fmt.Errorf("member does not belong to project")
	}

	return member, nil
}

func (s *service) ListMembers(ctx context.Context, projectID string) (*project.MemberList, error) {
	members, err := s.repo.ListMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &project.MemberList{
		Members: members,
		Total:   len(members),
	}, nil
}

func (s *service) UpdateMemberRole(ctx context.Context, projectID, memberID string, req *project.UpdateMemberRoleRequest) (*project.ProjectMember, error) {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member belongs to project
	if member.ProjectID != projectID {
		return nil, fmt.Errorf("member does not belong to project")
	}

	// Update role
	member.Role = req.Role

	if err := s.repo.UpdateMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	// Update RBAC in namespace
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	if err := s.k8sRepo.ApplyRBAC(ctx, proj.WorkspaceID, proj.NamespaceName, member.UserID, req.Role); err != nil {
		s.logger.Error("failed to update RBAC", "error", err)
	}

	// Log activity
	activity := &project.ProjectActivity{
		ID:          generateID(),
		ProjectID:   projectID,
		Type:        "member_role_updated",
		Description: fmt.Sprintf("Updated member %s role to %s", member.UserEmail, req.Role),
		CreatedAt:   time.Now(),
	}
	s.repo.CreateActivity(ctx, activity)

	return member, nil
}

func (s *service) RemoveMember(ctx context.Context, projectID, memberID, removerID string) error {
	member, err := s.repo.GetMemberByID(ctx, memberID)
	if err != nil {
		return err
	}

	// Verify member belongs to project
	if member.ProjectID != projectID {
		return fmt.Errorf("member does not belong to project")
	}

	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Remove member
	if err := s.repo.RemoveMember(ctx, memberID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Remove RBAC from namespace
	if err := s.k8sRepo.RemoveRBAC(ctx, proj.WorkspaceID, proj.NamespaceName, member.UserID); err != nil {
		s.logger.Error("failed to remove RBAC", "error", err)
	}

	// Log activity
	activity := &project.ProjectActivity{
		ID:          generateID(),
		ProjectID:   projectID,
		Type:        "member_removed",
		Description: fmt.Sprintf("Removed member %s from project", member.UserEmail),
		UserID:      removerID,
		CreatedAt:   time.Now(),
	}
	s.repo.CreateActivity(ctx, activity)

	return nil
}

func (s *service) AddProjectMember(ctx context.Context, projectID string, req *project.AddMemberRequest) error {
	// Use the existing AddMember method
	adderID := req.AddedBy
	if adderID == "" {
		// If no adder ID provided, use a system ID
		adderID = "system"
	}
	
	_, err := s.AddMember(ctx, projectID, adderID, req)
	return err
}

func (s *service) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Get member first
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Remove member
	if err := s.repo.RemoveMember(ctx, member.ID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Remove RBAC from namespace
	if err := s.k8sRepo.RemoveRBAC(ctx, proj.WorkspaceID, proj.Name, userID); err != nil {
		s.logger.Error("failed to remove RBAC", "error", err)
	}

	// Log activity
	s.logActivity(ctx, proj.ID, "member_removed", 
		fmt.Sprintf("Removed user %s from project %s", userID, proj.Name), "")

	return nil
}

func (s *service) ListProjectMembers(ctx context.Context, projectID string) ([]*project.ProjectMember, error) {
	return s.repo.ListMembers(ctx, projectID)
}

func (s *service) GetActivityLogs(ctx context.Context, projectID string, filter project.ActivityFilter) ([]*project.Activity, error) {
	filter.ProjectID = projectID
	return s.repo.ListActivities(ctx, filter)
}

func (s *service) GetProjectStats(ctx context.Context, projectID string) (*project.ProjectStats, error) {
	proj, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get counts
	namespaceCount := 0
	if proj.Settings != nil && proj.Settings["namespaces"] != nil {
		if namespaces, ok := proj.Settings["namespaces"].([]interface{}); ok {
			namespaceCount = len(namespaces)
		}
	}
	memberCount, err := s.repo.CountMembers(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to count members", "error", err)
	}

	// Get resource usage
	usage, err := s.repo.GetProjectResourceUsage(ctx, projectID)
	if err != nil {
		s.logger.Error("failed to get resource usage", "error", err)
	}

	// Get last activity
	lastActivity, err := s.repo.GetLastActivity(ctx, projectID)
	var lastActivityTime *time.Time
	if err == nil && lastActivity != nil {
		lastActivityTime = &lastActivity.CreatedAt
	}

	return &project.ProjectStats{
		ProjectID:      projectID,
		NamespaceCount: namespaceCount,
		MemberCount:    memberCount,
		ResourceUsage:  usage,
		LastActivity:   lastActivityTime,
	}, nil
}

func (s *service) ListActivities(ctx context.Context, projectID string, limit int) (*project.ActivityList, error) {
	filter := project.ActivityFilter{
		ProjectID: projectID,
		PageSize:  limit,
	}
	activities, err := s.repo.ListActivities(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &project.ActivityList{
		Activities: activities,
		Total:      len(activities),
	}, nil
}

func (s *service) LogActivity(ctx context.Context, activity *project.ProjectActivity) error {
	if activity.ID == "" {
		activity.ID = generateID()
	}
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	return s.repo.CreateActivity(ctx, activity)
}

func (s *service) ValidateProjectAccess(ctx context.Context, userID, projectID string, requiredRole string) error {
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("access denied: user is not a project member")
	}

	// Check role hierarchy
	if !hasRequiredRole(member.Role, requiredRole) {
		return fmt.Errorf("access denied: insufficient permissions")
	}

	return nil
}

func (s *service) GetUserProjectRole(ctx context.Context, userID, projectID string) (string, error) {
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err != nil {
		return "", fmt.Errorf("user is not a project member")
	}

	return member.Role, nil
}

// hasRequiredRole checks if the user's role meets the required role
func hasRequiredRole(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"viewer":    1,
		"developer": 2,
		"admin":     3,
	}

	userLevel, ok1 := roleHierarchy[userRole]
	requiredLevel, ok2 := roleHierarchy[requiredRole]

	if !ok1 || !ok2 {
		return false
	}

	return userLevel >= requiredLevel
}

// Helper functions

func generateID() string {
	return uuid.New().String()
}

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

func (s *service) logActivity(ctx context.Context, projectID, activityType, description, userID string) {
	activity := &project.ProjectActivity{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Type:        activityType,
		Description: description,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateActivity(ctx, activity); err != nil {
		s.logger.Error("failed to log activity", "error", err)
	}
}