'use client';

import { useState, useEffect } from 'react';
import { Plus, Loader2, AlertCircle, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { WorkspaceCard } from './workspace-card';
import { CreateWorkspaceDialog } from './create-workspace-dialog';
import { workspacesApi, vclusterApi, type Workspace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

interface WorkspaceListProps {
  organizationId: string;
}

export function WorkspaceList({ organizationId }: WorkspaceListProps) {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const { toast } = useToast();

  const loadWorkspaces = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await workspacesApi.list(organizationId);
      setWorkspaces(response.workspaces);
    } catch (err: unknown) {
      const errorMessage = (err as any)?.response?.data?.error || (err as any)?.message || 'Failed to load workspaces';
      setError(errorMessage);
      
      if ((err as any)?.code === 'NETWORK_ERROR' || !(err as any)?.response) {
        setError('Unable to connect to server');
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkspaces();
  }, [organizationId, loadWorkspaces]);

  const handleCreateWorkspace = async (data: { name: string; plan_id: string }) => {
    try {
      const newWorkspace = await workspacesApi.create(organizationId, data);
      setWorkspaces(prev => [...prev, newWorkspace]);
      setIsCreateDialogOpen(false);
      toast({
        title: 'Success',
        description: 'Workspace created successfully',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to create workspace',
        variant: 'destructive',
      });
      throw err;
    }
  };

  const handleStartWorkspace = async (workspaceId: string) => {
    try {
      setActionLoading(workspaceId);
      await vclusterApi.start(organizationId, workspaceId);
      
      // Update workspace status optimistically
      setWorkspaces(prev => 
        prev.map(ws => 
          ws.id === workspaceId 
            ? { ...ws, vcluster_status: 'STARTING' }
            : ws
        )
      );
      
      toast({
        title: 'Success',
        description: 'vCluster start initiated',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to start vCluster',
        variant: 'destructive',
      });
    } finally {
      setActionLoading(null);
    }
  };

  const handleStopWorkspace = async (workspaceId: string) => {
    try {
      setActionLoading(workspaceId);
      await vclusterApi.stop(organizationId, workspaceId);
      
      // Update workspace status optimistically
      setWorkspaces(prev => 
        prev.map(ws => 
          ws.id === workspaceId 
            ? { ...ws, vcluster_status: 'STOPPING' }
            : ws
        )
      );
      
      toast({
        title: 'Success',
        description: 'vCluster stop initiated',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to stop vCluster',
        variant: 'destructive',
      });
    } finally {
      setActionLoading(null);
    }
  };

  const handleViewWorkspace = (workspaceId: string) => {
    // Navigate to workspace detail page
    window.location.href = `/organizations/${organizationId}/workspaces/${workspaceId}`;
  };

  const handleDeleteWorkspace = async (workspaceId: string) => {
    if (!confirm('Are you sure you want to delete this workspace? This action cannot be undone.')) {
      return;
    }

    try {
      await workspacesApi.delete(organizationId, workspaceId);
      setWorkspaces(prev => prev.filter(ws => ws.id !== workspaceId));
      toast({
        title: 'Success',
        description: 'Workspace deleted successfully',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to delete workspace',
        variant: 'destructive',
      });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin" />
        <span className="ml-2">Loading workspaces...</span>
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center p-8">
          <div className="text-center space-y-4">
            <AlertCircle className="h-12 w-12 text-red-500 mx-auto" />
            <div>
              <h3 className="text-lg font-medium text-gray-900">Error Loading Workspaces</h3>
              <p className="text-gray-600 mt-1" data-testid="error-message">
                {error.includes('Unable to connect') ? (
                  <span data-testid="network-error">Unable to connect to server</span>
                ) : (
                  'Failed to load workspaces'
                )}
              </p>
            </div>
            <Button onClick={loadWorkspaces} variant="outline">
              <RefreshCw className="h-4 w-4 mr-2" />
              Try Again
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (workspaces.length === 0) {
    return (
      <div className="text-center py-12" data-testid="workspaces-empty-state">
        <div className="mx-auto max-w-md">
          <div className="mx-auto h-12 w-12 bg-gray-100 rounded-lg flex items-center justify-center">
            <Plus className="h-6 w-6 text-gray-600" />
          </div>
          <h3 className="mt-4 text-lg font-medium text-gray-900">No workspaces found</h3>
          <p className="mt-2 text-gray-600">
            Create your first workspace to get started with Kubernetes clusters.
          </p>
          <Button 
            className="mt-4"
            onClick={() => setIsCreateDialogOpen(true)}
            data-testid="create-workspace-button"
          >
            <Plus className="h-4 w-4 mr-2" />
            Create Workspace
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Workspaces</h2>
          <p className="text-gray-600">Manage your Kubernetes clusters</p>
        </div>
        <Button 
          onClick={() => setIsCreateDialogOpen(true)}
          data-testid="create-workspace-button"
        >
          <Plus className="h-4 w-4 mr-2" />
          Create Workspace
        </Button>
      </div>

      {/* Workspace Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {workspaces.map((workspace) => (
          <div key={workspace.id} className={actionLoading === workspace.id ? 'opacity-50' : ''}>
            <WorkspaceCard
              workspace={workspace}
              onStart={handleStartWorkspace}
              onStop={handleStopWorkspace}
              onView={handleViewWorkspace}
              onDelete={handleDeleteWorkspace}
            />
          </div>
        ))}
      </div>

      {/* Mobile Layout */}
      <div className="block md:hidden" data-testid="mobile-workspace-list">
        <div className="space-y-4">
          {workspaces.map((workspace) => (
            <WorkspaceCard
              key={workspace.id}
              workspace={workspace}
              onStart={handleStartWorkspace}
              onStop={handleStopWorkspace}
              onView={handleViewWorkspace}
              onDelete={handleDeleteWorkspace}
            />
          ))}
        </div>
      </div>

      {/* Create Workspace Dialog */}
      <CreateWorkspaceDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSubmit={handleCreateWorkspace}
      />
    </div>
  );
}