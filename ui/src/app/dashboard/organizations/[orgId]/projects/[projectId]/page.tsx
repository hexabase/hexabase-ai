'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Plus, Settings, Activity, Users, Cpu, Database, Package, RefreshCw } from 'lucide-react';
import { projectsApi, namespacesApi, type Project, type Namespace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { NamespaceCard } from '@/components/projects/namespace-card';
import { CreateNamespaceDialog } from '@/components/projects/create-namespace-dialog';
import { ProjectSettingsDialog } from '@/components/projects/project-settings-dialog';

export default function ProjectDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const [project, setProject] = useState<Project | null>(null);
  const [namespaces, setNamespaces] = useState<Namespace[]>([]);
  const [stats, setStats] = useState<{ namespaces: number; pods: number; cpu_usage: string; memory_usage: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [isCreateNamespaceOpen, setIsCreateNamespaceOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  const orgId = params.orgId as string;
  const projectId = params.projectId as string;

  const loadProjectData = useCallback(async () => {
    try {
      setLoading(true);
      
      // Mock project data for testing
      const mockProject: Project = {
        id: projectId,
        name: 'Frontend Application',
        description: 'React-based frontend application with microservices architecture',
        workspace_id: 'workspace-1',
        workspace_name: 'Production Workspace',
        status: 'active',
        namespace_count: 3,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        resource_usage: {
          cpu: '2.5 cores',
          memory: '4.2 GB',
          pods: 12,
        },
      };
      setProject(mockProject);

      // Mock stats
      setStats({
        namespaces: 3,
        pods: 12,
        cpu_usage: '65%',
        memory_usage: '42%',
      });

      await loadNamespaces();
    } catch (error: unknown) {
      console.error('Failed to load project:', error);
      toast({
        title: 'Error',
        description: 'Failed to load project details',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  }, [orgId, projectId, toast]);

  const loadNamespaces = useCallback(async () => {
    try {
      // Mock namespaces data for testing
      const mockNamespaces: Namespace[] = [
        {
          id: 'ns-1',
          name: 'development',
          description: 'Development environment',
          project_id: projectId,
          status: 'active',
          resource_quota: {
            cpu: '2 cores',
            memory: '4 GB',
            pods: 10,
          },
          resource_usage: {
            cpu: '1.2 cores',
            memory: '2.1 GB',
            pods: 5,
          },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'ns-2',
          name: 'staging',
          description: 'Staging environment for testing',
          project_id: projectId,
          status: 'active',
          resource_quota: {
            cpu: '1 core',
            memory: '2 GB',
            pods: 5,
          },
          resource_usage: {
            cpu: '0.8 cores',
            memory: '1.5 GB',
            pods: 3,
          },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'ns-3',
          name: 'production',
          description: 'Production environment',
          project_id: projectId,
          status: 'active',
          resource_quota: {
            cpu: '4 cores',
            memory: '8 GB',
            pods: 20,
          },
          resource_usage: {
            cpu: '2.1 cores',
            memory: '4.2 GB',
            pods: 8,
          },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ];
      setNamespaces(mockNamespaces);
    } catch (error: unknown) {
      console.error('Failed to load namespaces:', error);
    }
  }, [orgId, projectId]);

  useEffect(() => {
    loadProjectData();
  }, [loadProjectData]);

  const handleCreateNamespace = async (data: { name: string; description?: string; resource_quota?: { cpu: string; memory: string; pods: number } }) => {
    try {
      const newNamespace = await namespacesApi.create(orgId, projectId, data);
      setNamespaces(prev => [...prev, newNamespace]);
      setIsCreateNamespaceOpen(false);
      toast({
        title: 'Success',
        description: 'Namespace created successfully',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to create namespace',
        variant: 'destructive',
      });
      throw err;
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'active':
        return 'default';
      case 'inactive':
        return 'secondary';
      case 'archived':
        return 'outline';
      default:
        return 'secondary';
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-32 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!project) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Project not found</h1>
          <p className="text-gray-600 mt-2">The requested project could not be found.</p>
          <Button onClick={() => router.back()} className="mt-4">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6" data-testid="project-detail-page">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" onClick={() => router.back()}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold" data-testid="project-name">
              {project.name}
            </h1>
            <p className="text-gray-600" data-testid="project-description">
              {project.description || 'No description provided'}
            </p>
            <div className="flex items-center space-x-2 mt-2">
              <Badge variant={getStatusBadgeVariant(project.status)}>
                {project.status}
              </Badge>
              <span className="text-sm text-gray-500">•</span>
              <span className="text-sm text-gray-500">
                Workspace: {project.workspace_name}
              </span>
            </div>
          </div>
        </div>
        <Button
          variant="outline"
          onClick={() => setIsSettingsOpen(true)}
          data-testid="project-settings-button"
        >
          <Settings className="w-4 h-4 mr-2" />
          Settings
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card data-testid="total-namespaces-stat">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Namespaces</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.namespaces || 0}</div>
            <p className="text-xs text-muted-foreground">
              Total namespaces
            </p>
          </CardContent>
        </Card>

        <Card data-testid="total-pods-stat">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pods</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.pods || 0}</div>
            <p className="text-xs text-muted-foreground">
              Running pods
            </p>
          </CardContent>
        </Card>

        <Card data-testid="cpu-usage-stat">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.cpu_usage || '0%'}</div>
            <p className="text-xs text-muted-foreground">
              Of allocated resources
            </p>
          </CardContent>
        </Card>

        <Card data-testid="memory-usage-stat">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.memory_usage || '0%'}</div>
            <p className="text-xs text-muted-foreground">
              Of allocated resources
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Resource Usage Chart */}
      <Card data-testid="resource-usage-chart">
        <CardHeader>
          <CardTitle>Resource Usage Overview</CardTitle>
          <CardDescription>
            Current resource utilization across all namespaces
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-gray-500">
            <Activity className="mx-auto h-12 w-12 mb-4" />
            <p>Resource usage charts coming soon</p>
          </div>
        </CardContent>
      </Card>

      {/* Namespaces Section */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle data-testid="namespaces-section">Namespaces</CardTitle>
              <CardDescription>
                Manage Kubernetes namespaces for environment isolation
              </CardDescription>
            </div>
            <div className="flex space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={loadNamespaces}
              >
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
              <Button
                onClick={() => setIsCreateNamespaceOpen(true)}
                data-testid="create-namespace-button"
              >
                <Plus className="w-4 h-4 mr-2" />
                Create Namespace
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {namespaces.length === 0 ? (
            <div className="text-center py-12">
              <Package className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">No namespaces</h3>
              <p className="mt-1 text-sm text-gray-500">
                Get started by creating your first namespace
              </p>
              <Button
                onClick={() => setIsCreateNamespaceOpen(true)}
                className="mt-4"
              >
                <Plus className="w-4 h-4 mr-2" />
                Create Namespace
              </Button>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {namespaces.map((namespace) => (
                <NamespaceCard
                  key={namespace.id}
                  namespace={namespace}
                  onUpdate={loadNamespaces}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Dialogs */}
      <CreateNamespaceDialog
        open={isCreateNamespaceOpen}
        onOpenChange={setIsCreateNamespaceOpen}
        onSubmit={handleCreateNamespace}
      />

      <ProjectSettingsDialog
        open={isSettingsOpen}
        onOpenChange={setIsSettingsOpen}
        project={project}
        onUpdate={(updatedProject) => setProject(updatedProject)}
      />
    </div>
  );
}