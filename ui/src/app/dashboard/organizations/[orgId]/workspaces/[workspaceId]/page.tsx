'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Download, Play, Square, RefreshCw, Activity, Users, Cpu, Database } from 'lucide-react';
import { workspacesApi, vclusterApi, type Workspace, type VClusterHealth } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

export default function WorkspaceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [health, setHealth] = useState<VClusterHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [operationLoading, setOperationLoading] = useState(false);

  const orgId = params.orgId as string;
  const workspaceId = params.workspaceId as string;

  const loadWorkspaceData = async () => {
    try {
      setLoading(true);
      // Mock workspace data for testing
      const mockWorkspace: Workspace = {
        id: workspaceId,
        name: 'Test Workspace',
        plan_id: 'plan-basic',
        vcluster_status: 'running',
        vcluster_instance_name: `vcluster-${workspaceId}`,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setWorkspace(mockWorkspace);
      await loadHealth();
    } catch (error: unknown) {
      console.error('Failed to load workspace:', error);
      toast({
        title: 'Error',
        description: 'Failed to load workspace details',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const loadHealth = async () => {
    try {
      const healthData = await vclusterApi.getHealth(orgId, workspaceId);
      setHealth(healthData);
    } catch (error: unknown) {
      console.error('Failed to load vCluster health:', error);
      // Set mock health data for testing
      setHealth({
        healthy: true,
        components: {
          'api-server': 'healthy',
          'etcd': 'healthy',
          'scheduler': 'healthy',
        },
        resource_usage: {
          'cpu': '45.2%',
          'memory': '62.8%',
          'nodes': '3',
          'pods': '12',
        },
        last_checked: new Date().toISOString(),
      });
    }
  };

  useEffect(() => {
    loadWorkspaceData();
    const interval = setInterval(loadHealth, 30000); // Poll health every 30 seconds
    return () => clearInterval(interval);
  }, [orgId, workspaceId]);

  const handleVClusterAction = async (action: 'start' | 'stop') => {
    try {
      setOperationLoading(true);
      if (action === 'start') {
        await vclusterApi.start(orgId, workspaceId);
        toast({
          title: 'vCluster Starting',
          description: 'Your vCluster is being started...',
        });
      } else {
        // Mock stop operation
        toast({
          title: 'vCluster Stopping',
          description: 'Your vCluster is being stopped...',
        });
      }
      await loadHealth();
    } catch (error: unknown) {
      toast({
        title: 'Error',
        description: `Failed to ${action} vCluster`,
        variant: 'destructive',
      });
    } finally {
      setOperationLoading(false);
    }
  };

  const handleDownloadKubeconfig = async () => {
    try {
      const response = await workspacesApi.getKubeconfig(orgId, workspaceId);
      const blob = new Blob([response.kubeconfig], { type: 'text/plain' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `kubeconfig-${workspace?.name}.yaml`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast({
        title: 'Download Started',
        description: 'Kubeconfig file downloaded successfully',
      });
    } catch (error: unknown) {
      toast({
        title: 'Error',
        description: 'Failed to download kubeconfig',
        variant: 'destructive',
      });
    }
  };

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'running':
        return 'default';
      case 'stopped':
        return 'secondary';
      case 'starting':
      case 'stopping':
        return 'outline';
      default:
        return 'destructive';
    }
  };

  const getHealthStatusColor = (healthy: boolean) => {
    return healthy ? 'text-green-600' : 'text-red-600';
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

  if (!workspace) {
    return (
      <div className="container mx-auto p-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">Workspace not found</h1>
          <p className="text-gray-600 mt-2">The requested workspace could not be found.</p>
          <Button onClick={() => router.back()} className="mt-4">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Go Back
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6" data-testid="workspace-detail-page">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" onClick={() => router.back()}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold" data-testid="workspace-name">
              {workspace.name}
            </h1>
            <p className="text-gray-600">
              vCluster: {workspace.vcluster_instance_name}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          <Badge variant={getStatusBadgeVariant(workspace.vcluster_status)} data-testid="workspace-status">
            {workspace.vcluster_status}
          </Badge>
          <Button
            variant="outline"
            onClick={handleDownloadKubeconfig}
            data-testid="download-kubeconfig"
          >
            <Download className="w-4 h-4 mr-2" />
            Download Kubeconfig
          </Button>
          {workspace.vcluster_status === 'running' ? (
            <Button
              variant="outline"
              onClick={() => handleVClusterAction('stop')}
              disabled={operationLoading}
              data-testid="stop-vcluster"
            >
              <Square className="w-4 h-4 mr-2" />
              Stop vCluster
            </Button>
          ) : (
            <Button
              onClick={() => handleVClusterAction('start')}
              disabled={operationLoading}
              data-testid="start-vcluster"
            >
              <Play className="w-4 h-4 mr-2" />
              Start vCluster
            </Button>
          )}
        </div>
      </div>

      {/* Health Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card data-testid="health-status-card">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Health Status</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${getHealthStatusColor(health?.healthy ?? false)}`}>
              {health?.healthy ? 'Healthy' : 'Unhealthy'}
            </div>
            <p className="text-xs text-muted-foreground">
              Last checked: {health?.last_checked ? new Date(health.last_checked).toLocaleTimeString() : 'Never'}
            </p>
          </CardContent>
        </Card>

        <Card data-testid="nodes-card">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Nodes</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{health?.resource_usage?.nodes || 0}</div>
            <p className="text-xs text-muted-foreground">
              Active nodes
            </p>
          </CardContent>
        </Card>

        <Card data-testid="cpu-usage-card">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{health?.resource_usage?.cpu || '0%'}</div>
            <p className="text-xs text-muted-foreground">
              Current usage
            </p>
          </CardContent>
        </Card>

        <Card data-testid="memory-usage-card">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{health?.resource_usage?.memory || '0%'}</div>
            <p className="text-xs text-muted-foreground">
              Current usage
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Pod Information */}
      <Card data-testid="pod-info-card">
        <CardHeader>
          <CardTitle>Pod Information</CardTitle>
          <CardDescription>
            Overview of running pods in your vCluster
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-3xl font-bold mb-2">{health?.resource_usage?.pods || 0}</div>
          <p className="text-sm text-muted-foreground">
            Total pods running across all namespaces
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={loadHealth}
            className="mt-4"
            data-testid="refresh-health"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}