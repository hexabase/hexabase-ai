'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Application, applicationsApi } from '@/lib/api-client';
import { ApplicationCard } from './application-card';
import { DeployApplicationDialog } from './deploy-application-dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, AlertCircle, Filter } from 'lucide-react';
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

export function ApplicationList() {
  const router = useRouter();
  const params = useParams();
  const { toast } = useToast();
  const orgId = params?.orgId as string;
  const workspaceId = params?.workspaceId as string;
  const projectId = params?.projectId as string;

  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deployDialogOpen, setDeployDialogOpen] = useState(false);
  const [deleteApp, setDeleteApp] = useState<Application | null>(null);
  const [filters, setFilters] = useState({
    type: '',
    status: '',
  });

  const fetchApplications = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await applicationsApi.list(orgId, workspaceId, projectId, filters);
      setApplications(response.data.applications);
    } catch (error) {
      setError('Failed to load applications');
      console.error('Error fetching applications:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchApplications();
  }, [orgId, workspaceId, projectId, filters]);

  const handleApplicationClick = (applicationId: string) => {
    router.push(
      `/dashboard/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/applications/${applicationId}`
    );
  };

  const handleDeploy = async (data: any) => {
    try {
      const newApp = await applicationsApi.create(orgId, workspaceId, projectId, data);
      setDeployDialogOpen(false);
      toast({
        title: 'Application deployed',
        description: `${newApp.data.name} has been deployed successfully.`,
      });
      fetchApplications(); // Refresh list
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to deploy application. Please try again.',
        variant: 'destructive',
      });
      throw error; // Re-throw for dialog to handle
    }
  };

  const handleEditApplication = (application: Application) => {
    // Edit functionality would be implemented here
    console.log('Edit application:', application);
  };

  const handleDeleteApplication = async (applicationId: string) => {
    try {
      await applicationsApi.delete(orgId, workspaceId, projectId, applicationId);
      setApplications(applications.filter(app => app.id !== applicationId));
      setDeleteApp(null);
      toast({
        title: 'Application deleted',
        description: 'Application has been deleted successfully.',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete application. Please try again.',
        variant: 'destructive',
      });
    }
  };

  const handleStatusChange = async (applicationId: string, newStatus: string) => {
    try {
      const response = await applicationsApi.updateStatus(orgId, workspaceId, applicationId, { 
        status: newStatus 
      });
      setApplications(applications.map(app => 
        app.id === applicationId ? response.data : app
      ));
      toast({
        title: 'Status updated',
        description: `Application status changed to ${newStatus}.`,
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update application status.',
        variant: 'destructive',
      });
      throw error;
    }
  };

  if (loading) {
    return (
      <div className="space-y-4" data-testid="applications-skeleton">
        {[1, 2, 3].map((i) => (
          <div key={i} className="rounded-lg border p-6">
            <div className="space-y-3">
              <Skeleton className="h-5 w-1/3" />
              <Skeleton className="h-4 w-2/3" />
              <div className="flex gap-4">
                <Skeleton className="h-8 w-20" />
                <Skeleton className="h-8 w-20" />
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <AlertCircle className="h-8 w-8 text-destructive mb-4" />
        <p className="text-lg font-medium">{error}</p>
        <Button onClick={fetchApplications} className="mt-4">
          Retry
        </Button>
      </div>
    );
  }

  if (applications.length === 0 && !filters.type && !filters.status) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <h3 className="text-lg font-medium">No applications found</h3>
        <p className="text-muted-foreground mt-2">Deploy your first application to get started.</p>
        <Button onClick={() => setDeployDialogOpen(true)} className="mt-4">
          <Plus className="h-4 w-4 mr-2" />
          Deploy Application
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Applications</h2>
        <div className="flex gap-2">
          <Select
            value={filters.type}
            onValueChange={(value) => setFilters({ ...filters, type: value })}
          >
            <SelectTrigger className="w-40" data-testid="type-filter">
              <Filter className="h-4 w-4 mr-2" />
              <SelectValue placeholder="All types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">All types</SelectItem>
              <SelectItem value="stateless">Stateless</SelectItem>
              <SelectItem value="stateful">Stateful</SelectItem>
              <SelectItem value="cronjob">CronJob</SelectItem>
              <SelectItem value="function">Function</SelectItem>
            </SelectContent>
          </Select>

          <Select
            value={filters.status}
            onValueChange={(value) => setFilters({ ...filters, status: value })}
          >
            <SelectTrigger className="w-40" data-testid="status-filter">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="">All statuses</SelectItem>
              <SelectItem value="running">Running</SelectItem>
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="error">Error</SelectItem>
              <SelectItem value="suspended">Suspended</SelectItem>
            </SelectContent>
          </Select>

          <Button onClick={() => setDeployDialogOpen(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Deploy Application
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {applications.map((application) => (
          <ApplicationCard
            key={application.id}
            application={application}
            onClick={handleApplicationClick}
            onEdit={handleEditApplication}
            onDelete={(applicationId: string) => setDeleteApp(applications.find(app => app.id === applicationId) || null)}
            onStatusChange={handleStatusChange}
          />
        ))}
      </div>

      <DeployApplicationDialog
        open={deployDialogOpen}
        onOpenChange={setDeployDialogOpen}
        onSubmit={handleDeploy}
      />

      <AlertDialog open={!!deleteApp} onOpenChange={() => setDeleteApp(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Application</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{deleteApp?.name}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteApp && handleDeleteApplication(deleteApp.id)}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
