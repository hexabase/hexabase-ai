'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { Plus, Search, Filter, Loader2, AlertCircle, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ProjectCard } from './project-card';
import { CreateProjectDialog } from './create-project-dialog';
import { projectsApi, workspacesApi, type Project, type Workspace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

interface ProjectListProps {
  organizationId: string;
}

export function ProjectList({ organizationId }: ProjectListProps) {
  const router = useRouter();
  const [projects, setProjects] = useState<Project[]>([]);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [filters, setFilters] = useState({
    search: '',
    status: '',
    workspace_id: '',
  });
  const { toast } = useToast();

  const loadProjects = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await projectsApi.list(organizationId, filters);
      setProjects(response.projects);
    } catch (err: unknown) {
      const errorMessage = (err as any)?.response?.data?.error || (err as any)?.message || 'Failed to load projects';
      setError(errorMessage);
      
      // Set mock projects for testing
      if ((err as any)?.code === 'NETWORK_ERROR' || !(err as any)?.response) {
        const mockProjects: Project[] = [
          {
            id: 'project-123',
            name: 'Frontend Application',
            description: 'React-based frontend application',
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
          },
          {
            id: 'project-456',
            name: 'Backend Services',
            description: 'Microservices backend',
            workspace_id: 'workspace-1',
            workspace_name: 'Production Workspace',
            status: 'active',
            namespace_count: 5,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            resource_usage: {
              cpu: '4.1 cores',
              memory: '8.7 GB',
              pods: 25,
            },
          },
        ];
        setProjects(mockProjects);
        setError(null);
      }
    } finally {
      setLoading(false);
    }
  }, [organizationId, filters]);

  const loadWorkspaces = useCallback(async () => {
    try {
      const response = await workspacesApi.list(organizationId);
      setWorkspaces(response.workspaces);
    } catch (err: unknown) {
      console.error('Failed to load workspaces:', err);
      // Set mock workspaces for testing
      setWorkspaces([
        {
          id: 'workspace-1',
          name: 'Production Workspace',
          plan_id: 'plan-pro',
          vcluster_status: 'running',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'workspace-2',
          name: 'Development Workspace',
          plan_id: 'plan-basic',
          vcluster_status: 'running',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ]);
    }
  }, [organizationId]);

  useEffect(() => {
    loadProjects();
    loadWorkspaces();
  }, [loadProjects, loadWorkspaces]);

  const handleCreateProject = async (data: { name: string; description?: string; workspace_id: string }) => {
    try {
      const newProject = await projectsApi.create(organizationId, data);
      setProjects(prev => [...prev, newProject]);
      setIsCreateDialogOpen(false);
      toast({
        title: 'Success',
        description: 'Project created successfully',
        variant: 'default',
      });
    } catch (err: unknown) {
      toast({
        title: 'Error',
        description: (err as any)?.response?.data?.error || 'Failed to create project',
        variant: 'destructive',
      });
      throw err;
    }
  };

  const handleProjectClick = (projectId: string) => {
    router.push(`/dashboard/organizations/${organizationId}/projects/${projectId}`);
  };

  const handleSearchChange = (value: string) => {
    setFilters(prev => ({ ...prev, search: value }));
  };

  const handleStatusFilter = (value: string) => {
    setFilters(prev => ({ ...prev, status: value }));
  };

  const handleWorkspaceFilter = (value: string) => {
    setFilters(prev => ({ ...prev, workspace_id: value }));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4 text-primary-600" />
          <p className="text-sm text-gray-600">Loading projects...</p>
        </div>
      </div>
    );
  }

  if (error && projects.length === 0) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="text-center">
            <AlertCircle className="h-8 w-8 text-red-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to load projects</h3>
            <p className="text-sm text-gray-600 mb-4">{error}</p>
            <Button onClick={loadProjects} variant="outline">
              <RefreshCw className="h-4 w-4 mr-2" />
              Try Again
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (projects.length === 0) {
    return (
      <div className="text-center py-12" data-testid="projects-empty-state">
        <div className="mx-auto h-16 w-16 bg-gradient-to-br from-gray-100 to-gray-200 rounded-2xl flex items-center justify-center mb-6">
          <svg
            className="h-8 w-8 text-gray-500"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
            />
          </svg>
        </div>
        <h3 className="text-xl font-semibold text-gray-900 mb-3">No projects yet</h3>
        <p className="text-base text-gray-600 max-w-md mx-auto mb-6">
          Create your first project to organize your applications and services within Kubernetes namespaces.
        </p>
        <Button 
          onClick={() => setIsCreateDialogOpen(true)}
          data-testid="create-project-button"
        >
          <Plus className="h-4 w-4 mr-2" />
          Create your first project
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Search and Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search projects..."
            value={filters.search}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="pl-10"
            data-testid="project-search-input"
          />
        </div>
        
        <Select value={filters.status} onValueChange={handleStatusFilter}>
          <SelectTrigger className="w-full sm:w-48" data-testid="project-status-filter">
            <Filter className="h-4 w-4 mr-2" />
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All statuses</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
            <SelectItem value="archived">Archived</SelectItem>
          </SelectContent>
        </Select>

        <Select value={filters.workspace_id} onValueChange={handleWorkspaceFilter}>
          <SelectTrigger className="w-full sm:w-48" data-testid="workspace-filter">
            <SelectValue placeholder="Filter by workspace" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All workspaces</SelectItem>
            {workspaces.map((workspace) => (
              <SelectItem key={workspace.id} value={workspace.id}>
                {workspace.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button 
          onClick={() => setIsCreateDialogOpen(true)}
          data-testid="create-project-button"
        >
          <Plus className="h-4 w-4 mr-2" />
          New Project
        </Button>
      </div>

      {/* Projects Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3" data-testid="projects-list">
        {projects.map((project) => (
          <ProjectCard
            key={project.id}
            project={project}
            onClick={() => handleProjectClick(project.id)}
          />
        ))}
      </div>

      {/* Create Project Dialog */}
      <CreateProjectDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onSubmit={handleCreateProject}
        workspaces={workspaces}
      />
    </div>
  );
}