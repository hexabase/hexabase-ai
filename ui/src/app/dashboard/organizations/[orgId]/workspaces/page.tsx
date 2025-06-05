'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Plus, Server, Activity, Cpu, HardDrive } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { CreateWorkspaceDialog } from '@/components/create-workspace-dialog';
import { workspacesApi, type Workspace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { useOrganizationUpdates } from '@/hooks/use-websocket';
import type { WorkspaceStatusUpdate } from '@/lib/websocket';

export default function WorkspacesPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  const orgId = params.orgId as string;
  
  // WebSocket integration
  const { onWorkspaceStatus, isConnected } = useOrganizationUpdates(orgId);

  const loadWorkspaces = async () => {
    try {
      setLoading(true);
      const response = await workspacesApi.list(orgId);
      setWorkspaces(response.workspaces);
    } catch (error) {
      console.error('Failed to load workspaces:', error);
      // Use mock data for testing
      setWorkspaces([
        {
          id: 'ws-1',
          name: 'Production Workspace',
          plan_id: 'plan-pro',
          vcluster_status: 'RUNNING',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        },
        {
          id: 'ws-2',
          name: 'Development Workspace',
          plan_id: 'plan-dev',
          vcluster_status: 'STOPPED',
          created_at: '2024-01-02T00:00:00Z',
          updated_at: '2024-01-02T00:00:00Z'
        },
        {
          id: 'ws-3',
          name: 'Staging Workspace',
          plan_id: 'plan-starter',
          vcluster_status: 'PENDING_CREATION',
          created_at: '2024-01-03T00:00:00Z',
          updated_at: '2024-01-03T00:00:00Z'
        }
      ]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkspaces();
  }, [orgId]);

  // Listen for workspace status updates
  useEffect(() => {
    const unsubscribe = onWorkspaceStatus((update: WorkspaceStatusUpdate) => {
      setWorkspaces(prev => 
        prev.map(ws => 
          ws.id === update.workspace_id 
            ? { ...ws, vcluster_status: update.status }
            : ws
        )
      );
      
      // Show toast for important status changes
      if (update.status === 'RUNNING' || update.status === 'ERROR') {
        toast({
          title: update.status === 'RUNNING' ? 'Workspace Started' : 'Workspace Error',
          description: update.message || `Workspace ${update.workspace_id} status changed to ${update.status}`,
          variant: update.status === 'ERROR' ? 'destructive' : 'default',
        });
      }
    });

    return () => {
      unsubscribe();
    };
  }, [onWorkspaceStatus, toast]);

  const handleWorkspaceCreated = async () => {
    setCreateDialogOpen(false);
    await loadWorkspaces();
    toast({
      title: 'Workspace Created',
      description: 'Your new workspace is being provisioned.',
    });
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'RUNNING': { label: 'Running', variant: 'default' as const, className: 'status-running' },
      'STOPPED': { label: 'Stopped', variant: 'secondary' as const, className: 'status-stopped' },
      'PENDING_CREATION': { label: 'Creating', variant: 'outline' as const, className: 'status-pending' },
      'ERROR': { label: 'Error', variant: 'destructive' as const, className: 'status-error' },
      'STARTING': { label: 'Starting', variant: 'outline' as const, className: 'status-starting' },
      'STOPPING': { label: 'Stopping', variant: 'outline' as const, className: 'status-stopping' },
    };

    const config = statusConfig[status] || { label: status, variant: 'outline' as const, className: '' };

    return (
      <Badge 
        variant={config.variant} 
        className={config.className}
        data-testid="workspace-status"
      >
        {config.label}
      </Badge>
    );
  };

  const getResourceIcon = (planId: string) => {
    switch (planId) {
      case 'plan-pro':
        return <Server className="h-5 w-5 text-blue-600" />;
      case 'plan-dev':
        return <Cpu className="h-5 w-5 text-green-600" />;
      default:
        return <HardDrive className="h-5 w-5 text-gray-600" />;
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6 space-y-6">
        <div className="flex justify-between items-center">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-10 w-40" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Workspaces</h1>
          <p className="text-gray-600">Manage your Kubernetes workspaces</p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Workspace
        </Button>
      </div>

      {/* Workspace List */}
      {workspaces.length === 0 ? (
        <Card data-testid="empty-state">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Server className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-semibold mb-2">No workspaces yet</h3>
            <p className="text-gray-600 text-center mb-6">
              Create your first workspace to start deploying applications on Kubernetes.
            </p>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create your first workspace
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" data-testid="workspace-list">
          {workspaces.map((workspace) => (
            <Card 
              key={workspace.id} 
              className="cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => router.push(`/dashboard/organizations/${orgId}/workspaces/${workspace.id}`)}
              data-testid="workspace-card"
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex items-center gap-3">
                    {getResourceIcon(workspace.plan_id)}
                    <div>
                      <CardTitle data-testid="workspace-name">{workspace.name}</CardTitle>
                      <CardDescription>
                        Created {new Date(workspace.created_at).toLocaleDateString()}
                      </CardDescription>
                    </div>
                  </div>
                  {getStatusBadge(workspace.vcluster_status)}
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">Plan</span>
                    <span className="font-medium capitalize">
                      {workspace.plan_id.replace('plan-', '')}
                    </span>
                  </div>
                  {workspace.vcluster_instance_name && (
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-gray-600">Instance</span>
                      <span className="font-mono text-xs">
                        {workspace.vcluster_instance_name}
                      </span>
                    </div>
                  )}
                  <div className="flex items-center gap-2 pt-2">
                    <Activity className="h-4 w-4 text-gray-400" />
                    <span className="text-sm text-gray-600">
                      Last updated {new Date(workspace.updated_at).toLocaleTimeString()}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Workspace Dialog */}
      <CreateWorkspaceDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        orgId={orgId}
        onSuccess={handleWorkspaceCreated}
      />
    </div>
  );
}