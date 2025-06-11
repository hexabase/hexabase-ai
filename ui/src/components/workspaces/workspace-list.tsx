'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { apiClient, workspacesApi, plansApi, Workspace } from '@/lib/api-client';
import { WorkspaceCard } from './workspace-card';
import { CreateWorkspaceDialog } from './create-workspace-dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, RefreshCw, Server } from 'lucide-react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useToast } from '@/hooks/use-toast';

interface Plan {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  resource_limits?: {
    cpu: string;
    memory: string;
    storage: string;
  };
}

export function WorkspaceList() {
  const router = useRouter();
  const params = useParams();
  const orgId = params?.orgId as string;
  const { toast } = useToast();

  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [deletingWorkspaceId, setDeletingWorkspaceId] = useState<string | null>(null);

  const fetchWorkspaces = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [workspacesResponse, plansResponse] = await Promise.all([
        workspacesApi.list(orgId),
        plansApi.list(),
      ]);
      setWorkspaces(workspacesResponse.workspaces);
      setPlans(plansResponse.plans);
    } catch (err) {
      setError('Failed to load workspaces');
      console.error('Error fetching workspaces:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (orgId) {
      fetchWorkspaces();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orgId]);

  const handleWorkspaceClick = (workspaceId: string) => {
    router.push(`/dashboard/organizations/${orgId}/workspaces/${workspaceId}`);
  };

  const handleCreateWorkspace = async (name: string, planId: string) => {
    try {
      const response = await workspacesApi.create(orgId, {
        name,
        plan_id: planId,
      });
      setWorkspaces([...workspaces, response]);
      setIsCreateOpen(false);
      toast({
        title: 'Workspace created',
        description: `${name} has been created successfully.`,
      });
    } catch (err) {
      toast({
        title: 'Error creating workspace',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteWorkspace = async () => {
    if (!deletingWorkspaceId) return;

    try {
      await workspacesApi.delete(orgId, deletingWorkspaceId);
      setWorkspaces(workspaces.filter(ws => ws.id !== deletingWorkspaceId));
      setDeletingWorkspaceId(null);
      toast({
        title: 'Workspace deleted',
        description: 'The workspace has been deleted successfully.',
      });
    } catch (err) {
      toast({
        title: 'Error deleting workspace',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  const handleDownloadKubeconfig = async (workspaceId: string) => {
    try {
      const response = await workspacesApi.getKubeconfig(orgId, workspaceId);
      const { kubeconfig } = response;
      
      // Create a blob and download the kubeconfig
      const blob = new Blob([kubeconfig], { type: 'text/yaml' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `kubeconfig-${workspaceId}.yaml`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast({
        title: 'Kubeconfig downloaded',
        description: 'The kubeconfig file has been downloaded successfully.',
      });
    } catch (err) {
      toast({
        title: 'Error downloading kubeconfig',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  const getPlanById = (planId: string) => {
    return plans.find(plan => plan.id === planId);
  };

  if (isLoading) {
    return (
      <div className="space-y-4" data-testid="workspaces-skeleton">
        <div className="flex justify-between items-center mb-6">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-10 w-40" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map(i => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription className="flex items-center justify-between">
          <span>{error}</span>
          <Button onClick={fetchWorkspaces} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    );
  }

  if (workspaces.length === 0) {
    return (
      <div className="text-center py-12">
        <Server className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">No workspaces found</h3>
        <p className="text-muted-foreground mb-4">
          Create your first workspace to get started
        </p>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Workspace
        </Button>
        <CreateWorkspaceDialog
          open={isCreateOpen}
          onOpenChange={setIsCreateOpen}
          plans={plans}
          onSubmit={handleCreateWorkspace}
        />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Workspaces</h2>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Workspace
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {workspaces.map(workspace => (
          <WorkspaceCard
            key={workspace.id}
            workspace={workspace}
            plan={getPlanById(workspace.plan_id)}
            onClick={handleWorkspaceClick}
            onDelete={setDeletingWorkspaceId}
            onDownloadKubeconfig={handleDownloadKubeconfig}
          />
        ))}
      </div>

      <CreateWorkspaceDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        plans={plans}
        onSubmit={handleCreateWorkspace}
      />

      <AlertDialog open={!!deletingWorkspaceId} onOpenChange={(open) => !open && setDeletingWorkspaceId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this workspace? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDeleteWorkspace}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}