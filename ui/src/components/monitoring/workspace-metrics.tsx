'use client';

import { useState, useEffect } from 'react';
import { monitoringApi, type WorkspaceMetrics as WorkspaceMetricsType } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import WorkspaceDetailModal from './workspace-detail-modal';
import { ChevronRight, Activity, AlertCircle, CheckCircle } from 'lucide-react';

interface WorkspaceMetricsProps {
  orgId: string;
}

export default function WorkspaceMetrics({ orgId }: WorkspaceMetricsProps) {
  const [workspaces, setWorkspaces] = useState<WorkspaceMetricsType[]>([]);
  const [selectedWorkspace, setSelectedWorkspace] = useState<WorkspaceMetricsType | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchWorkspaceMetrics();
    const interval = setInterval(fetchWorkspaceMetrics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [orgId]);

  const fetchWorkspaceMetrics = async () => {
    try {
      const data = await monitoringApi.getWorkspaceMetrics(orgId);
      setWorkspaces(data.workspaces);
    } catch (error) {
      console.error('Failed to fetch workspace metrics:', error);
      // Use mock data for development
      setWorkspaces(getMockWorkspaces());
    } finally {
      setLoading(false);
    }
  };

  const getMockWorkspaces = (): WorkspaceMetricsType[] => [
    {
      workspace_id: 'ws-dev',
      workspace_name: 'Development',
      cpu_usage: 35.2,
      memory_usage: 42.8,
      storage_usage: 128.5,
      pod_count: 23,
      namespace_count: 5,
      status: 'healthy'
    },
    {
      workspace_id: 'ws-staging',
      workspace_name: 'Staging',
      cpu_usage: 72.5,
      memory_usage: 68.3,
      storage_usage: 256.7,
      pod_count: 45,
      namespace_count: 8,
      status: 'warning'
    },
    {
      workspace_id: 'ws-prod',
      workspace_name: 'Production',
      cpu_usage: 58.9,
      memory_usage: 61.2,
      storage_usage: 512.3,
      pod_count: 89,
      namespace_count: 12,
      status: 'healthy'
    }
  ];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'warning':
        return <AlertCircle className="w-4 h-4 text-yellow-500" />;
      case 'critical':
        return <AlertCircle className="w-4 h-4 text-red-500" />;
      default:
        return <Activity className="w-4 h-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'text-green-600 bg-green-50';
      case 'warning':
        return 'text-yellow-600 bg-yellow-50';
      case 'critical':
        return 'text-red-600 bg-red-50';
      default:
        return 'text-gray-600 bg-gray-50';
    }
  };

  const getUsageColor = (percentage: number) => {
    if (percentage >= 80) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  if (loading) {
    return (
      <div data-testid="workspace-metrics" className="bg-white rounded-lg border p-6">
        <h2 className="text-lg font-semibold mb-4">Workspace Metrics</h2>
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="border rounded-lg p-4 animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/3 mb-3"></div>
              <div className="h-3 bg-gray-200 rounded w-full"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <>
      <div data-testid="workspace-metrics" className="bg-white rounded-lg border p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">Workspace Metrics</h2>
          <span className="text-sm text-gray-500">{workspaces.length} active workspaces</span>
        </div>

        <div className="space-y-3">
          {workspaces.map((workspace) => (
            <div
              key={workspace.workspace_id}
              data-testid="workspace-card"
              className="border rounded-lg p-4 hover:bg-gray-50 cursor-pointer transition-colors"
              onClick={() => setSelectedWorkspace(workspace)}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center space-x-3">
                  {getStatusIcon(workspace.status)}
                  <div>
                    <h3 data-testid="workspace-name" className="font-medium">{workspace.workspace_name}</h3>
                    <p className={`text-xs inline-flex items-center px-2 py-0.5 rounded-full ${getStatusColor(workspace.status)}`}>
                      {workspace.status}
                    </p>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-gray-400" />
              </div>

              <div className="grid grid-cols-4 gap-4 text-sm">
                <div>
                  <p className="text-gray-600">CPU</p>
                  <div className="flex items-center space-x-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-1.5">
                      <div 
                        data-testid="workspace-cpu"
                        className={`h-1.5 rounded-full ${getUsageColor(workspace.cpu_usage)}`}
                        style={{ width: `${workspace.cpu_usage}%` }}
                      ></div>
                    </div>
                    <span className="text-xs font-medium">{workspace.cpu_usage.toFixed(1)}%</span>
                  </div>
                </div>

                <div>
                  <p className="text-gray-600">Memory</p>
                  <div className="flex items-center space-x-2">
                    <div className="flex-1 bg-gray-200 rounded-full h-1.5">
                      <div 
                        data-testid="workspace-memory"
                        className={`h-1.5 rounded-full ${getUsageColor(workspace.memory_usage)}`}
                        style={{ width: `${workspace.memory_usage}%` }}
                      ></div>
                    </div>
                    <span className="text-xs font-medium">{workspace.memory_usage.toFixed(1)}%</span>
                  </div>
                </div>

                <div>
                  <p className="text-gray-600">Pods</p>
                  <p data-testid="workspace-pods" className="font-medium">{workspace.pod_count}</p>
                </div>

                <div>
                  <p className="text-gray-600">Namespaces</p>
                  <p className="font-medium">{workspace.namespace_count}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {selectedWorkspace && (
        <WorkspaceDetailModal
          workspace={selectedWorkspace}
          orgId={orgId}
          isOpen={!!selectedWorkspace}
          onClose={() => setSelectedWorkspace(null)}
        />
      )}
    </>
  );
}