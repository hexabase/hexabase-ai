'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Project, apiClient } from '@/lib/api-client';
import { ProjectCard } from './project-card';
import { CreateProjectDialog } from './create-project-dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Plus, AlertCircle } from 'lucide-react';
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

export function ProjectList() {
  const router = useRouter();
  const params = useParams();
  const { toast } = useToast();
  const orgId = params?.orgId as string;
  const workspaceId = params?.workspaceId as string;

  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [deleteProject, setDeleteProject] = useState<Project | null>(null);

  const fetchProjects = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.projects.list(orgId, { workspace_id: workspaceId });
      setProjects(response.projects);
    } catch (error) {
      setError('Failed to load projects');
      console.error('Error fetching projects:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProjects();
  }, [orgId, workspaceId]);

  const handleProjectClick = (projectId: string) => {
    router.push(`/dashboard/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}`);
  };

  const handleCreateProject = async (data: any) => {
    try {
      const newProject = await apiClient.projects.create(orgId, workspaceId, data);
      setProjects([...projects, newProject]);
      setCreateDialogOpen(false);
      toast({
        title: 'Project created',
        description: `${newProject.name} has been created successfully.`,
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to create project. Please try again.',
        variant: 'destructive',
      });
    }
  };

  const handleEditProject = (project: Project) => {
    // Edit functionality would be implemented here
    console.log('Edit project:', project);
  };

  const handleDeleteProject = async (projectId: string) => {
    try {
      await apiClient.projects.delete(orgId, workspaceId, projectId);
      setProjects(projects.filter(p => p.id !== projectId));
      setDeleteProject(null);
      toast({
        title: 'Project deleted',
        description: 'Project has been deleted successfully.',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete project. Please try again.',
        variant: 'destructive',
      });
    }
  };

  if (loading) {
    return (
      <div className="space-y-4" data-testid="projects-skeleton">
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
        <Button onClick={fetchProjects} className="mt-4">
          Retry
        </Button>
      </div>
    );
  }

  if (projects.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <h3 className="text-lg font-medium">No projects found</h3>
        <p className="text-muted-foreground mt-2">Create your first project to get started.</p>
        <Button onClick={() => setCreateDialogOpen(true)} className="mt-4">
          <Plus className="h-4 w-4 mr-2" />
          Create Project
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Projects</h2>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Project
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {projects.map((project) => (
          <ProjectCard
            key={project.id}
            project={project}
            onClick={handleProjectClick}
            onEdit={handleEditProject}
            onDelete={(projectId) => setDeleteProject(projects.find(p => p.id === projectId) || null)}
          />
        ))}
      </div>

      <CreateProjectDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSubmit={handleCreateProject}
      />

      <AlertDialog open={!!deleteProject} onOpenChange={() => setDeleteProject(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Project</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{deleteProject?.name}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteProject && handleDeleteProject(deleteProject.id)}
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