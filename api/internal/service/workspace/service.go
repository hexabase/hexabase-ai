package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/workspace"
	"go.uber.org/zap"
)

type service struct {
	repo      workspace.Repository
	k8sRepo   workspace.KubernetesRepository
	authRepo  workspace.AuthRepository
	logger    *zap.Logger
}

// NewService creates a new workspace service
func NewService(
	repo workspace.Repository,
	k8sRepo workspace.KubernetesRepository,
	authRepo workspace.AuthRepository,
	logger *zap.Logger,
) workspace.Service {
	return &service{
		repo:     repo,
		k8sRepo:  k8sRepo,
		authRepo: authRepo,
		logger:   logger,
	}
}

func (s *service) CreateWorkspace(ctx context.Context, req *workspace.CreateWorkspaceRequest) (*workspace.Workspace, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("workspace name is required")
	}

	// Create workspace record
	ws := &workspace.Workspace{
		ID:             uuid.New().String(),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Plan:           req.Plan,
		Status:         "creating",
		Settings:       req.Settings,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateWorkspace(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Create provisioning task
	task := &workspace.Task{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Type:        "provision_vcluster",
		Status:      "pending",
		Payload:     map[string]interface{}{"workspace_id": ws.ID},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		s.logger.Error("failed to create provisioning task", zap.Error(err))
	}

	return ws, nil
}

func (s *service) GetWorkspace(ctx context.Context, workspaceID string) (*workspace.Workspace, error) {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Fetch cluster status if active
	if ws.Status == "active" {
		status, err := s.k8sRepo.GetVClusterStatus(ctx, workspaceID)
		if err != nil {
			s.logger.Warn("failed to get vcluster status", zap.Error(err))
		} else {
			ws.ClusterInfo = status
		}
	}

	return ws, nil
}

func (s *service) ListWorkspaces(ctx context.Context, orgID string, filter workspace.WorkspaceFilter) ([]*workspace.Workspace, int, error) {
	filter.OrganizationID = orgID
	return s.repo.ListWorkspaces(ctx, filter)
}

func (s *service) UpdateWorkspace(ctx context.Context, workspaceID string, req *workspace.UpdateWorkspaceRequest) (*workspace.Workspace, error) {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Update fields
	if req.Name != "" {
		ws.Name = req.Name
	}
	if req.Description != "" {
		ws.Description = req.Description
	}
	if req.Settings != nil {
		ws.Settings = req.Settings
	}

	ws.UpdatedAt = time.Now()

	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return ws, nil
}

func (s *service) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status == "deleting" {
		return fmt.Errorf("workspace is already being deleted")
	}

	// Update status to deleting
	ws.Status = "deleting"
	ws.UpdatedAt = time.Now()

	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	// Create deletion task
	task := &workspace.Task{
		ID:          uuid.New().String(),
		WorkspaceID: ws.ID,
		Type:        "delete_vcluster",
		Status:      "pending",
		Payload:     map[string]interface{}{"workspace_id": ws.ID},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		s.logger.Error("failed to create deletion task", zap.Error(err))
	}

	return nil
}

func (s *service) SuspendWorkspace(ctx context.Context, workspaceID string, reason string) error {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status != "active" {
		return fmt.Errorf("can only suspend active workspaces")
	}

	// Update status
	ws.Status = "suspended"
	ws.UpdatedAt = time.Now()

	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	// Scale down vCluster
	if err := s.k8sRepo.ScaleVCluster(ctx, workspaceID, 0); err != nil {
		s.logger.Error("failed to scale down vcluster", zap.Error(err))
	}

	return nil
}

func (s *service) ReactivateWorkspace(ctx context.Context, workspaceID string) error {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status != "suspended" {
		return fmt.Errorf("can only reactivate suspended workspaces")
	}

	// Update status
	ws.Status = "active"
	ws.UpdatedAt = time.Now()

	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	// Scale up vCluster
	if err := s.k8sRepo.ScaleVCluster(ctx, workspaceID, 1); err != nil {
		s.logger.Error("failed to scale up vcluster", zap.Error(err))
	}

	return nil
}

func (s *service) GetResourceUsage(ctx context.Context, workspaceID string) (*workspace.ResourceUsage, error) {
	metrics, err := s.k8sRepo.GetResourceMetrics(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource metrics: %w", err)
	}

	usage := &workspace.ResourceUsage{
		WorkspaceID: workspaceID,
		CPU:         metrics["cpu"],
		Memory:      metrics["memory"],
		Storage:     metrics["storage"],
		Pods:        int(metrics["pods"]),
		Timestamp:   time.Now(),
	}

	// Store usage record
	if err := s.repo.CreateResourceUsage(ctx, usage); err != nil {
		s.logger.Error("failed to store resource usage", zap.Error(err))
	}

	return usage, nil
}

func (s *service) GetKubeconfig(ctx context.Context, workspaceID, userID string) (string, error) {
	// Check if user has access
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to list workspace members: %w", err)
	}

	hasAccess := false
	for _, member := range members {
		if member.UserID == userID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return "", fmt.Errorf("user does not have access to workspace")
	}

	// Get OIDC token for user
	token, err := s.authRepo.GenerateWorkspaceToken(ctx, userID, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to generate workspace token: %w", err)
	}

	// Get kubeconfig from vCluster
	kubeconfig, err := s.k8sRepo.GetVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to get vcluster kubeconfig: %w", err)
	}

	// Update kubeconfig with OIDC token
	kubeconfig = s.updateKubeconfigWithToken(kubeconfig, token)

	return kubeconfig, nil
}

func (s *service) updateKubeconfigWithToken(kubeconfig, token string) string {
	// Implementation to update kubeconfig with OIDC token
	// This would parse the kubeconfig and update the user auth section
	return kubeconfig
}

func (s *service) AddWorkspaceMember(ctx context.Context, workspaceID string, req *workspace.AddMemberRequest) error {
	// Check if workspace exists
	if _, err := s.repo.GetWorkspace(ctx, workspaceID); err != nil {
		return fmt.Errorf("workspace not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.UserID == req.UserID {
			return fmt.Errorf("user is already a member")
		}
	}

	// Add member
	member := &workspace.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
		AddedAt:     time.Now(),
	}

	if err := s.repo.AddWorkspaceMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// Update OIDC configuration
	if err := s.k8sRepo.UpdateOIDCConfig(ctx, workspaceID); err != nil {
		s.logger.Error("failed to update OIDC config", zap.Error(err))
	}

	return nil
}

func (s *service) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	// Check if user is a member
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list members: %w", err)
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return fmt.Errorf("user is not a member")
	}

	// Remove member
	if err := s.repo.RemoveWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Update OIDC configuration
	if err := s.k8sRepo.UpdateOIDCConfig(ctx, workspaceID); err != nil {
		s.logger.Error("failed to update OIDC config", zap.Error(err))
	}

	return nil
}

func (s *service) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	return s.repo.ListWorkspaceMembers(ctx, workspaceID)
}

func (s *service) ProcessTask(ctx context.Context, taskID string) error {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status != "pending" {
		return fmt.Errorf("task is not pending")
	}

	// Update task status
	task.Status = "processing"
	task.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Process based on task type
	var processErr error
	switch task.Type {
	case "provision_vcluster":
		processErr = s.provisionVCluster(ctx, task)
	case "delete_vcluster":
		processErr = s.deleteVCluster(ctx, task)
	default:
		processErr = fmt.Errorf("unknown task type: %s", task.Type)
	}

	// Update task status based on result
	if processErr != nil {
		task.Status = "failed"
		task.Error = processErr.Error()
	} else {
		task.Status = "completed"
	}
	task.CompletedAt = &[]time.Time{time.Now()}[0]
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		s.logger.Error("failed to update task status", zap.Error(err))
	}

	return processErr
}

func (s *service) provisionVCluster(ctx context.Context, task *workspace.Task) error {
	workspaceID := task.WorkspaceID

	// Get workspace details
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Create vCluster
	if err := s.k8sRepo.CreateVCluster(ctx, workspaceID, ws.Plan); err != nil {
		return fmt.Errorf("failed to create vcluster: %w", err)
	}

	// Wait for vCluster to be ready
	if err := s.k8sRepo.WaitForVClusterReady(ctx, workspaceID); err != nil {
		return fmt.Errorf("vcluster did not become ready: %w", err)
	}

	// Configure OIDC
	if err := s.k8sRepo.ConfigureOIDC(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to configure OIDC: %w", err)
	}

	// Apply resource quotas
	if err := s.k8sRepo.ApplyResourceQuotas(ctx, workspaceID, ws.Plan); err != nil {
		return fmt.Errorf("failed to apply resource quotas: %w", err)
	}

	// Update workspace status
	ws.Status = "active"
	ws.UpdatedAt = time.Now()
	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return fmt.Errorf("failed to update workspace status: %w", err)
	}

	return nil
}

func (s *service) deleteVCluster(ctx context.Context, task *workspace.Task) error {
	workspaceID := task.WorkspaceID

	// Delete vCluster
	if err := s.k8sRepo.DeleteVCluster(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to delete vcluster: %w", err)
	}

	// Wait for deletion
	if err := s.k8sRepo.WaitForVClusterDeleted(ctx, workspaceID); err != nil {
		return fmt.Errorf("vcluster deletion did not complete: %w", err)
	}

	// Mark workspace as deleted
	if err := s.repo.DeleteWorkspace(ctx, workspaceID); err != nil {
		return fmt.Errorf("failed to delete workspace record: %w", err)
	}

	return nil
}

func (s *service) GetWorkspaceTask(ctx context.Context, taskID string) (*workspace.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *service) ListWorkspaceTasks(ctx context.Context, workspaceID string) ([]*workspace.Task, error) {
	return s.repo.ListTasks(ctx, workspaceID)
}