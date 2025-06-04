'use client';

import { useState, useEffect } from 'react';
import { monitoringApi, type ClusterHealth } from '@/lib/api-client';
import { Activity, CheckCircle, AlertTriangle, XCircle, Server } from 'lucide-react';

interface ClusterHealthCardProps {
  orgId: string;
}

export default function ClusterHealthCard({ orgId }: ClusterHealthCardProps) {
  const [health, setHealth] = useState<ClusterHealth | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchClusterHealth();
    const interval = setInterval(fetchClusterHealth, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [orgId]);

  const fetchClusterHealth = async () => {
    try {
      const data = await monitoringApi.getClusterHealth(orgId);
      setHealth(data);
    } catch (error) {
      console.error('Failed to fetch cluster health:', error);
      // Use mock data for development
      setHealth(getMockHealth());
    } finally {
      setLoading(false);
    }
  };

  const getMockHealth = (): ClusterHealth => ({
    status: 'healthy',
    uptime_seconds: 2592000, // 30 days
    last_check: new Date().toISOString(),
    nodes_total: 5,
    nodes_healthy: 5,
    nodes_unhealthy: 0,
    services_total: 23,
    services_healthy: 23
  });

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="w-8 h-8 text-green-500" />;
      case 'degraded':
        return <AlertTriangle className="w-8 h-8 text-yellow-500" />;
      case 'critical':
        return <XCircle className="w-8 h-8 text-red-500" />;
      default:
        return <Activity className="w-8 h-8 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'degraded':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`;
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  };

  if (loading) {
    return (
      <div data-testid="cluster-health" className="bg-white rounded-lg border p-6 animate-pulse">
        <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
        <div className="h-8 bg-gray-200 rounded w-1/2"></div>
      </div>
    );
  }

  if (!health) {
    return null;
  }

  return (
    <div data-testid="cluster-health" className="bg-white rounded-lg border">
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold">Cluster Health</h2>
          <span className="text-sm text-gray-500">
            Last updated: {new Date(health.last_check).toLocaleTimeString()}
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          {/* Overall Status */}
          <div className="flex items-center space-x-4">
            <div>{getStatusIcon(health.status)}</div>
            <div>
              <p className="text-sm text-gray-600">Status</p>
              <p data-testid="health-status" className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(health.status)}`}>
                {health.status.charAt(0).toUpperCase() + health.status.slice(1)}
              </p>
            </div>
          </div>

          {/* Uptime */}
          <div className="flex items-center space-x-4">
            <Activity className="w-8 h-8 text-blue-500" />
            <div>
              <p className="text-sm text-gray-600">Uptime</p>
              <p data-testid="cluster-uptime" className="text-lg font-medium">
                {formatUptime(health.uptime_seconds)}
              </p>
            </div>
          </div>

          {/* Nodes */}
          <div data-testid="nodes-overview" className="flex items-center space-x-4">
            <Server className="w-8 h-8 text-indigo-500" />
            <div>
              <p className="text-sm text-gray-600">Nodes</p>
              <div className="flex items-center space-x-2">
                <span data-testid="healthy-nodes" className="text-lg font-medium text-green-600">
                  {health.nodes_healthy}
                </span>
                <span className="text-gray-400">/</span>
                <span data-testid="total-nodes" className="text-lg font-medium">
                  {health.nodes_total}
                </span>
              </div>
            </div>
          </div>

          {/* Services */}
          <div className="flex items-center space-x-4">
            <div className="w-8 h-8 bg-purple-100 rounded-lg flex items-center justify-center">
              <span className="text-purple-600 font-bold">S</span>
            </div>
            <div>
              <p className="text-sm text-gray-600">Services</p>
              <div className="flex items-center space-x-2">
                <span className="text-lg font-medium text-green-600">
                  {health.services_healthy}
                </span>
                <span className="text-gray-400">/</span>
                <span className="text-lg font-medium">
                  {health.services_total}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Alert if cluster is not healthy */}
        {health.status !== 'healthy' && (
          <div className={`mt-4 p-3 rounded-lg ${
            health.status === 'degraded' ? 'bg-yellow-50 text-yellow-800' : 'bg-red-50 text-red-800'
          }`}>
            <p className="text-sm font-medium">
              {health.status === 'degraded' 
                ? 'Cluster is experiencing degraded performance'
                : 'Critical issues detected in the cluster'}
            </p>
            <p className="text-sm mt-1">
              {health.nodes_unhealthy} unhealthy nodes detected
            </p>
          </div>
        )}
      </div>
    </div>
  );
}