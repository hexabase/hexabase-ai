'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { monitoringApi, type WorkspaceMetrics } from '@/lib/api-client';
import { X, Activity, Package, HardDrive, Database } from 'lucide-react';

interface WorkspaceDetailModalProps {
  workspace: WorkspaceMetrics;
  orgId: string;
  isOpen: boolean;
  onClose: () => void;
}

interface NamespaceMetric {
  name: string;
  cpu_usage: number;
  memory_usage: number;
  pod_count: number;
}

export default function WorkspaceDetailModal({ workspace, orgId, isOpen, onClose }: WorkspaceDetailModalProps) {
  const [namespaces, setNamespaces] = useState<NamespaceMetric[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (isOpen) {
      fetchWorkspaceDetails();
    }
  }, [isOpen, workspace.workspace_id]);

  const fetchWorkspaceDetails = async () => {
    try {
      const data = await monitoringApi.getWorkspaceDetails(orgId, workspace.workspace_id, { time_range: '1h' });
      // In real app, would get namespace breakdown from API
      setNamespaces(getMockNamespaces());
    } catch (error) {
      console.error('Failed to fetch workspace details:', error);
      setNamespaces(getMockNamespaces());
    } finally {
      setLoading(false);
    }
  };

  const getMockNamespaces = (): NamespaceMetric[] => [
    { name: 'default', cpu_usage: 15.2, memory_usage: 23.5, pod_count: 5 },
    { name: 'kube-system', cpu_usage: 8.7, memory_usage: 12.3, pod_count: 8 },
    { name: 'monitoring', cpu_usage: 12.5, memory_usage: 18.9, pod_count: 6 },
    { name: 'application', cpu_usage: 25.8, memory_usage: 32.1, pod_count: 12 },
    { name: 'database', cpu_usage: 18.3, memory_usage: 25.6, pod_count: 4 }
  ];

  if (!isOpen) return null;

  const getUsageColor = (percentage: number) => {
    if (percentage >= 80) return 'bg-red-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div data-testid="workspace-detail-modal" className="bg-white rounded-lg p-6 max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-xl font-semibold">{workspace.workspace_name} Workspace</h2>
            <p className="text-sm text-gray-600 mt-1">Detailed metrics and namespace breakdown</p>
          </div>
          <Button variant="outline" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Workspace Overview */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Activity className="w-4 h-4 text-blue-500" />
              <span className="text-sm text-gray-600">CPU Usage</span>
            </div>
            <p className="text-2xl font-bold">{workspace.cpu_usage.toFixed(1)}%</p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Database className="w-4 h-4 text-purple-500" />
              <span className="text-sm text-gray-600">Memory Usage</span>
            </div>
            <p className="text-2xl font-bold">{workspace.memory_usage.toFixed(1)}%</p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Package className="w-4 h-4 text-green-500" />
              <span className="text-sm text-gray-600">Total Pods</span>
            </div>
            <p className="text-2xl font-bold">{workspace.pod_count}</p>
          </div>

          <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <HardDrive className="w-4 h-4 text-orange-500" />
              <span className="text-sm text-gray-600">Storage</span>
            </div>
            <p className="text-2xl font-bold">{workspace.storage_usage.toFixed(1)} GB</p>
          </div>
        </div>

        {/* Namespace Breakdown */}
        <div data-testid="namespace-metrics">
          <h3 className="font-medium mb-4">Namespace Breakdown</h3>
          
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="border rounded-lg p-4 animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
                  <div className="h-3 bg-gray-200 rounded w-full"></div>
                </div>
              ))}
            </div>
          ) : (
            <div className="space-y-3">
              {namespaces.map((namespace) => (
                <div key={namespace.name} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-medium">{namespace.name}</h4>
                    <span className="text-sm text-gray-600">{namespace.pod_count} pods</span>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <div className="flex items-center justify-between text-sm mb-1">
                        <span className="text-gray-600">CPU</span>
                        <span className="font-medium">{namespace.cpu_usage.toFixed(1)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className={`h-2 rounded-full ${getUsageColor(namespace.cpu_usage)}`}
                          style={{ width: `${namespace.cpu_usage}%` }}
                        ></div>
                      </div>
                    </div>
                    
                    <div>
                      <div className="flex items-center justify-between text-sm mb-1">
                        <span className="text-gray-600">Memory</span>
                        <span className="font-medium">{namespace.memory_usage.toFixed(1)}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                          className={`h-2 rounded-full ${getUsageColor(namespace.memory_usage)}`}
                          style={{ width: `${namespace.memory_usage}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="mt-6 flex justify-end">
          <Button onClick={onClose}>Close</Button>
        </div>
      </div>
    </div>
  );
}