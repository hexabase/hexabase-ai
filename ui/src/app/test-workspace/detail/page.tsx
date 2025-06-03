'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Download, Play, Square, RefreshCw, Activity, Users, Cpu, Database } from 'lucide-react';
import { type Workspace, type VClusterHealth } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

export default function TestWorkspaceDetailPage() {
  const { toast } = useToast();
  const [workspace] = useState<Workspace>({
    id: 'test-workspace-123',
    name: 'Test Workspace',
    plan_id: 'plan-basic',
    vcluster_status: 'running',
    vcluster_instance_name: 'vcluster-test-123',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  });
  
  const [health] = useState<VClusterHealth>({
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

  const handleVClusterAction = async (action: 'start' | 'stop') => {
    toast({
      title: `vCluster ${action}`,
      description: `vCluster ${action} action triggered`,
    });
  };

  const handleDownloadKubeconfig = () => {
    toast({
      title: 'Download Started',
      description: 'Kubeconfig file download started',
    });
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

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto space-y-6" data-testid="workspace-detail-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button variant="ghost">
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
                data-testid="stop-vcluster"
              >
                <Square className="w-4 h-4 mr-2" />
                Stop vCluster
              </Button>
            ) : (
              <Button
                onClick={() => handleVClusterAction('start')}
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
              className="mt-4"
              data-testid="refresh-health"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}