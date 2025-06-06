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

func (s *service) CreateWorkspace(ctx context.Context, req *workspace.CreateWorkspaceRequest) (*workspace.Workspace, *workspace.Task, error) {
	// Validate request
	if req.Name == "" {
		return nil, nil, fmt.Errorf("workspace name is required")
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
		return nil, nil, fmt.Errorf("failed to create workspace: %w", err)
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

	return ws, task, nil
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
			ws.ClusterInfo = map[string]interface{}{
				"status": status,
			}
		}
	}

	return ws, nil
}

func (s *service) ListWorkspaces(ctx context.Context, filter workspace.WorkspaceFilter) (*workspace.WorkspaceList, error) {
	workspaces, total, err := s.repo.ListWorkspaces(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &workspace.WorkspaceList{
		Workspaces: workspaces,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
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

func (s *service) GetWorkspaceStatus(ctx context.Context, workspaceID string) (*workspace.WorkspaceStatus, error) {
	// Get workspace
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Get current status from vCluster
	vclusterStatus, err := s.k8sRepo.GetVClusterStatus(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to get vcluster status", zap.Error(err))
		vclusterStatus = "unknown"
	}

	// Get resource usage
	usage, err := s.k8sRepo.GetResourceMetrics(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to get resource metrics", zap.Error(err))
	}

	// Get cluster info
	clusterInfo, err := s.k8sRepo.GetVClusterInfo(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to get cluster info", zap.Error(err))
	}

	status := &workspace.WorkspaceStatus{
		WorkspaceID:   workspaceID,
		Status:        ws.Status,
		Healthy:       vclusterStatus == "running",
		Message:       fmt.Sprintf("vCluster is %s", vclusterStatus),
		ResourceUsage: usage,
		LastChecked:   time.Now(),
	}

	if clusterInfo != nil {
		status.ClusterInfo = map[string]interface{}{
			"endpoint":   clusterInfo.Endpoint,
			"api_server": clusterInfo.APIServer,
			"status":     clusterInfo.Status,
		}
	}

	// Save status
	if err := s.repo.SaveWorkspaceStatus(ctx, status); err != nil {
		s.logger.Error("failed to save workspace status", zap.Error(err))
	}

	return status, nil
}

func (s *service) ExecuteOperation(ctx context.Context, workspaceID string, req *workspace.WorkspaceOperationRequest) (*workspace.Task, error) {
	// Get workspace
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Create task based on operation type
	task := &workspace.Task{
		ID:          generateID(),
		WorkspaceID: workspaceID,
		Type:        req.Operation,
		Status:      "pending",
		Progress:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	switch req.Operation {
	case "backup":
		task.Message = fmt.Sprintf("Creating backup of workspace %s", ws.Name)
		if req.Metadata != nil {
			task.Metadata = req.Metadata
		}
	case "restore":
		task.Message = fmt.Sprintf("Restoring workspace %s", ws.Name)
		if req.Metadata != nil && req.Metadata["backup_id"] != nil {
			task.Payload = map[string]interface{}{
				"backup_id": req.Metadata["backup_id"],
			}
		}
	case "upgrade":
		task.Message = fmt.Sprintf("Upgrading workspace %s", ws.Name)
		if req.Metadata != nil && req.Metadata["target_version"] != nil {
			task.Payload = map[string]interface{}{
				"target_version": req.Metadata["target_version"],
			}
		}
	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}

	// Save task
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Publish task to message queue for async processing
	if err := s.publishTask(ctx, task); err != nil {
		s.logger.Error("failed to publish task", zap.Error(err))
	}

	return task, nil
}

func (s *service) DeleteWorkspace(ctx context.Context, workspaceID string) (*workspace.Task, error) {
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.Status == "deleting" {
		return nil, fmt.Errorf("workspace is already being deleted")
	}

	// Update status to deleting
	ws.Status = "deleting"
	ws.UpdatedAt = time.Now()

	if err := s.repo.UpdateWorkspace(ctx, ws); err != nil {
		return nil, fmt.Errorf("failed to update workspace status: %w", err)
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

	return task, nil
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
	usage, err := s.k8sRepo.GetResourceMetrics(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource metrics: %w", err)
	}

	// Set workspace ID and timestamp
	usage.WorkspaceID = workspaceID
	usage.Timestamp = time.Now()

	// Store usage record
	if err := s.repo.CreateResourceUsage(ctx, usage); err != nil {
		s.logger.Error("failed to store resource usage", zap.Error(err))
	}

	return usage, nil
}

func (s *service) GetKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	// Get workspace
	ws, err := s.repo.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("workspace not found: %w", err)
	}

	if ws.Status != "active" {
		return "", fmt.Errorf("workspace is not active")
	}

	// Get kubeconfig from repository
	kubeconfig, err := s.repo.GetKubeconfig(ctx, workspaceID)
	if err != nil {
		// Try to get from vCluster
		info, err := s.k8sRepo.GetVClusterInfo(ctx, workspaceID)
		if err != nil {
			return "", fmt.Errorf("failed to get vcluster info: %w", err)
		}
		
		kubeconfig = info.KubeConfig
		
		// Save kubeconfig for future use
		if err := s.repo.SaveKubeconfig(ctx, workspaceID, kubeconfig); err != nil {
			s.logger.Error("failed to save kubeconfig", zap.Error(err))
		}
	}

	return kubeconfig, nil
}

func (s *service) updateKubeconfigWithToken(kubeconfig, token string) string {
	// Implementation to update kubeconfig with OIDC token
	// This would parse the kubeconfig and update the user auth section
	return kubeconfig
}

func (s *service) AddWorkspaceMember(ctx context.Context, workspaceID string, req *workspace.AddMemberRequest) (*workspace.WorkspaceMember, error) {
	// Check if workspace exists
	if _, err := s.repo.GetWorkspace(ctx, workspaceID); err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	// Check if user is already a member
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	for _, member := range members {
		if member.UserID == req.UserID {
			return nil, fmt.Errorf("user is already a member")
		}
	}

	// Add member
	member := &workspace.WorkspaceMember{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		UserID:      req.UserID,
		Role:        req.Role,
		AddedBy:     req.AddedBy,
		AddedAt:     time.Now(),
	}

	if err := s.repo.AddWorkspaceMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Update OIDC configuration
	oidcConfig := map[string]interface{}{
		"users": []string{member.UserID},
	}
	if err := s.k8sRepo.UpdateOIDCConfig(ctx, workspaceID, oidcConfig); err != nil {
		s.logger.Error("failed to update OIDC config", zap.Error(err))
	}

	return member, nil
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

	// Update OIDC configuration with remaining users
	remainingUsers := []string{}
	for _, m := range members {
		if m.UserID != userID {
			remainingUsers = append(remainingUsers, m.UserID)
		}
	}
	
	oidcConfig := map[string]interface{}{
		"users": remainingUsers,
	}
	if err := s.k8sRepo.UpdateOIDCConfig(ctx, workspaceID, oidcConfig); err != nil {
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

func (s *service) GetTask(ctx context.Context, taskID string) (*workspace.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *service) GetWorkspaceTask(ctx context.Context, taskID string) (*workspace.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *service) ListTasks(ctx context.Context, workspaceID string) ([]*workspace.Task, error) {
	return s.repo.ListTasks(ctx, workspaceID)
}

func (s *service) ListWorkspaceTasks(ctx context.Context, workspaceID string) ([]*workspace.Task, error) {
	return s.repo.ListTasks(ctx, workspaceID)
}

func (s *service) ProcessProvisioningTask(ctx context.Context, taskID string) error {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.Type != "create" {
		return fmt.Errorf("invalid task type for provisioning: %s", task.Type)
	}

	// Update task to running
	task.Status = "running"
	task.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Provision vCluster
	if err := s.provisionVCluster(ctx, task); err != nil {
		// Update task to failed
		task.Status = "failed"
		task.Error = err.Error()
		task.UpdatedAt = time.Now()
		task.CompletedAt = &task.UpdatedAt
		s.repo.UpdateTask(ctx, task)
		return err
	}

	// Update task to completed
	task.Status = "completed"
	task.Progress = 100
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *service) ProcessDeletionTask(ctx context.Context, taskID string) error {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if task.Type != "delete" {
		return fmt.Errorf("invalid task type for deletion: %s", task.Type)
	}

	// Update task to running
	task.Status = "running"
	task.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Delete vCluster
	if err := s.deleteVCluster(ctx, task); err != nil {
		// Update task to failed
		task.Status = "failed"
		task.Error = err.Error()
		task.UpdatedAt = time.Now()
		task.CompletedAt = &task.UpdatedAt
		s.repo.UpdateTask(ctx, task)
		return err
	}

	// Update task to completed
	task.Status = "completed"
	task.Progress = 100
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.UpdatedAt = time.Now()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Helper functions

func generateID() string {
	return uuid.New().String()
}

func (s *service) publishTask(ctx context.Context, task *workspace.Task) error {
	// TODO: Implement message queue publishing
	// For now, just log the task
	s.logger.Info("task created", 
		zap.String("task_id", task.ID),
		zap.String("type", task.Type),
		zap.String("status", task.Status))
	return nil
}

func (s *service) ValidateWorkspaceAccess(ctx context.Context, userID, workspaceID string) error {
	// Check if user is a member of the workspace
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to list workspace members: %w", err)
	}

	for _, member := range members {
		if member.UserID == userID {
			return nil // User has access
		}
	}

	return fmt.Errorf("user does not have access to workspace")
}