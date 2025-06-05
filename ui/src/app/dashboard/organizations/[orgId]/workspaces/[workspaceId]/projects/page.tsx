'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Plus, Folder, Activity, Database, Cpu, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { CreateProjectDialog } from '@/components/create-project-dialog';
import { projectsApi, workspacesApi, type Project, type Workspace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

export default function ProjectsPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const [projects, setProjects] = useState<Project[]>([]);
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [loading, setLoading] = useState(true);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  const orgId = params.orgId as string;
  const workspaceId = params.workspaceId as string;

  const loadData = async () => {
    try {
      setLoading(true);
      
      // Load workspace info
      const wsResponse = await workspacesApi.get(orgId, workspaceId);
      setWorkspace(wsResponse);
      
      // Load projects
      const response = await projectsApi.list(orgId, { workspace_id: workspaceId });
      setProjects(response.projects);
    } catch (error) {
      console.error('Failed to load projects:', error);
      toast({
        title: 'Error',
        description: 'Failed to load projects',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [orgId, workspaceId]);

  const handleProjectCreated = async () => {
    setCreateDialogOpen(false);
    await loadData();
    toast({
      title: 'Success',
      description: 'Project created successfully',
    });
  };

  const handleProjectClick = (projectId: string) => {
    router.push(`/dashboard/organizations/${orgId}/projects/${projectId}`);
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'active': { label: 'Active', variant: 'default' as const },
      'inactive': { label: 'Inactive', variant: 'secondary' as const },
      'archived': { label: 'Archived', variant: 'outline' as const },
    };

    const config = statusConfig[status] || { label: status, variant: 'outline' as const };

    return (
      <Badge variant={config.variant} data-testid="project-status">
        {config.label}
      </Badge>
    );
  };

  const filteredProjects = projects.filter(project => {
    const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.description?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || project.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  if (loading) {
    return (
      <div className="container mx-auto p-6 space-y-6">
        <div className="flex justify-between items-center">
          <Skeleton className="h-8 w-32" />
          <Skeleton className="h-10 w-40" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
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
          <h1 className="text-3xl font-bold">Projects</h1>
          <p className="text-gray-600">
            Manage projects in {workspace?.name || 'workspace'}
          </p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)} data-testid="create-project-button">
          <Plus className="w-4 h-4 mr-2" />
          Create Project
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search projects..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
              data-testid="project-search-input"
            />
          </div>
        </div>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[180px]" data-testid="project-status-filter">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
            <SelectItem value="archived">Archived</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Project List */}
      {filteredProjects.length === 0 ? (
        <Card data-testid="empty-state">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Folder className="h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-semibold mb-2">
              {searchTerm || statusFilter !== 'all' ? 'No projects found' : 'No projects yet'}
            </h3>
            <p className="text-gray-600 text-center mb-6">
              {searchTerm || statusFilter !== 'all' 
                ? 'Try adjusting your search or filters'
                : 'Create your first project to organize namespaces and resources.'}
            </p>
            {!searchTerm && statusFilter === 'all' && (
              <Button onClick={() => setCreateDialogOpen(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Project
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" data-testid="projects-list">
          {filteredProjects.map((project) => (
            <Card
              key={project.id}
              className="cursor-pointer hover:shadow-lg transition-shadow"
              onClick={() => handleProjectClick(project.id)}
              data-testid={`project-card-${project.id}`}
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <CardTitle className="text-lg" data-testid="project-name">
                      {project.name}
                    </CardTitle>
                    <CardDescription className="mt-1" data-testid="project-description">
                      {project.description || 'No description'}
                    </CardDescription>
                  </div>
                  {getStatusBadge(project.status)}
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex items-center text-sm text-gray-600">
                    <Database className="h-4 w-4 mr-2" />
                    <span data-testid="project-namespace-count">
                      {project.namespace_count || 0} namespaces
                    </span>
                  </div>
                  {project.resource_usage && (
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center text-gray-600">
                        <Cpu className="h-4 w-4 mr-1" />
                        <span>{project.resource_usage.cpu}</span>
                      </div>
                      <div className="flex items-center text-gray-600">
                        <Activity className="h-4 w-4 mr-1" />
                        <span>{project.resource_usage.memory}</span>
                      </div>
                    </div>
                  )}
                  <div className="pt-2 text-xs text-gray-500">
                    Created {new Date(project.created_at).toLocaleDateString()}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Project Dialog */}
      <CreateProjectDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        organizationId={orgId}
        workspaceId={workspaceId}
        onSuccess={handleProjectCreated}
      />
    </div>
  );
}