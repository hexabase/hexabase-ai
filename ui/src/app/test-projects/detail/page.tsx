'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Plus, Settings, Activity, Users, Cpu, Database, Package, RefreshCw } from 'lucide-react';
import { type Project, type Namespace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { NamespaceCard } from '@/components/projects/namespace-card';
import { CreateNamespaceDialog } from '@/components/projects/create-namespace-dialog';
import { ProjectSettingsDialog } from '@/components/projects/project-settings-dialog';

export default function TestProjectDetailPage() {
  const { toast } = useToast();
  const [isCreateNamespaceOpen, setIsCreateNamespaceOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  const project: Project = {
    id: 'project-123',
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

  const namespaces: Namespace[] = [
    {
      id: 'ns-1',
      name: 'development',
      description: 'Development environment',
      project_id: 'project-123',
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
      project_id: 'project-123',
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
      project_id: 'project-123',
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

  const stats = {
    namespaces: 3,
    pods: 12,
    cpu_usage: '65%',
    memory_usage: '42%',
  };

  const handleCreateNamespace = async (data: { name: string; description?: string; resource_quota?: { cpu: string; memory: string; pods: number } }) => {
    setIsCreateNamespaceOpen(false);
    toast({
      title: 'Success',
      description: 'Namespace created successfully',
      variant: 'default',
    });
  };

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto space-y-6" data-testid="project-detail-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button variant="ghost">
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
                <Badge variant="default">
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
              <div className="text-2xl font-bold">{stats.namespaces}</div>
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
              <div className="text-2xl font-bold">{stats.pods}</div>
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
              <div className="text-2xl font-bold">{stats.cpu_usage}</div>
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
              <div className="text-2xl font-bold">{stats.memory_usage}</div>
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
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {namespaces.map((namespace) => (
                <NamespaceCard
                  key={namespace.id}
                  namespace={namespace}
                  onUpdate={() => {}}
                />
              ))}
            </div>
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
          onUpdate={() => {}}
        />
      </div>
    </div>
  );
}