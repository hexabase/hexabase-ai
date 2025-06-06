package workspace

import (
	"context"
)

// Service defines the workspace business logic interface
type Service interface {
	// CreateWorkspace creates a new workspace
	CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*Workspace, *Task, error)
	
	// GetWorkspace retrieves a workspace by ID
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	
	// ListWorkspaces lists workspaces with filtering
	ListWorkspaces(ctx context.Context, filter WorkspaceFilter) (*WorkspaceList, error)
	
	// UpdateWorkspace updates a workspace
	UpdateWorkspace(ctx context.Context, workspaceID string, req *UpdateWorkspaceRequest) (*Workspace, error)
	
	// DeleteWorkspace deletes a workspace
	DeleteWorkspace(ctx context.Context, workspaceID string) (*Task, error)
	
	// GetWorkspaceStatus gets the current status of a workspace
	GetWorkspaceStatus(ctx context.Context, workspaceID string) (*WorkspaceStatus, error)
	
	// GetKubeconfig retrieves the kubeconfig for a workspace
	GetKubeconfig(ctx context.Context, workspaceID string) (string, error)
	
	// ExecuteOperation performs an operation on a workspace (backup, restore, upgrade)
	ExecuteOperation(ctx context.Context, workspaceID string, req *WorkspaceOperationRequest) (*Task, error)
	
	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)
	
	// ListTasks lists tasks for a workspace
	ListTasks(ctx context.Context, workspaceID string) ([]*Task, error)
	
	// ProcessProvisioningTask processes a workspace provisioning task
	ProcessProvisioningTask(ctx context.Context, taskID string) error
	
	// ProcessDeletionTask processes a workspace deletion task
	ProcessDeletionTask(ctx context.Context, taskID string) error
	
	// ValidateWorkspaceAccess validates if a user has access to a workspace
	ValidateWorkspaceAccess(ctx context.Context, userID, workspaceID string) error
	
	// GetResourceUsage gets the resource usage for a workspace
	GetResourceUsage(ctx context.Context, workspaceID string) (*ResourceUsage, error)
	
	// AddWorkspaceMember adds a member to a workspace
	AddWorkspaceMember(ctx context.Context, workspaceID string, req *AddMemberRequest) (*WorkspaceMember, error)
	
	// ListWorkspaceMembers lists members of a workspace
	ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]*WorkspaceMember, error)
	
	// RemoveWorkspaceMember removes a member from a workspace
	RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error
}