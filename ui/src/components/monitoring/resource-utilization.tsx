'use client';

import { useState, useEffect } from 'react';
import { monitoringApi, type ResourceMetrics } from '@/lib/api-client';
import { Cpu, HardDrive, Wifi, Database } from 'lucide-react';

interface ResourceUtilizationProps {
  orgId: string;
}

export default function ResourceUtilization({ orgId }: ResourceUtilizationProps) {
  const [metrics, setMetrics] = useState<ResourceMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, [orgId]);

  const fetchMetrics = async () => {
    try {
      const data = await monitoringApi.getResourceMetrics(orgId, { time_range: '5m', interval: '1m' });
      setMetrics(data[data.length - 1]); // Get latest metrics
    } catch (error) {
      console.error('Failed to fetch resource metrics:', error);
      // Use mock data for development
      setMetrics(getMockMetrics());
    } finally {
      setLoading(false);
    }
  };

  const getMockMetrics = (): ResourceMetrics => ({
    timestamp: new Date().toISOString(),
    cpu: {
      usage_percentage: 42.5,
      cores_used: 17,
      cores_total: 40
    },
    memory: {
      usage_percentage: 68.3,
      used_gb: 109.3,
      total_gb: 160
    },
    storage: {
      usage_percentage: 35.7,
      used_gb: 1428,
      total_gb: 4000
    },
    network: {
      ingress_mbps: 245.8,
      egress_mbps: 189.3
    }
  });

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'text-red-600 bg-red-100';
    if (percentage >= 80) return 'text-orange-600 bg-orange-100';
    if (percentage >= 70) return 'text-yellow-600 bg-yellow-100';
    return 'text-green-600 bg-green-100';
  };

  const getProgressColor = (percentage: number) => {
    if (percentage >= 90) return 'bg-red-500';
    if (percentage >= 80) return 'bg-orange-500';
    if (percentage >= 70) return 'bg-yellow-500';
    return 'bg-green-500';
  };

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-lg border p-4 animate-pulse">
            <div className="h-4 bg-gray-200 rounded w-1/2 mb-3"></div>
            <div className="h-6 bg-gray-200 rounded w-3/4"></div>
          </div>
        ))}
      </div>
    );
  }

  if (!metrics) {
    return null;
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
      {/* CPU Utilization */}
      <div data-testid="cpu-utilization" className="bg-white rounded-lg border p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-2">
            <Cpu className="w-5 h-5 text-blue-500" />
            <h3 className="font-medium">CPU Usage</h3>
          </div>
          <span className={`text-sm font-medium px-2 py-1 rounded ${getUsageColor(metrics.cpu.usage_percentage)}`}>
            {metrics.cpu.usage_percentage.toFixed(1)}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={`h-2 rounded-full transition-all duration-300 ${getProgressColor(metrics.cpu.usage_percentage)}`}
              style={{ width: `${metrics.cpu.usage_percentage}%` }}
            ></div>
          </div>
          <p className="text-sm text-gray-600">
            {metrics.cpu.cores_used} / {metrics.cpu.cores_total} cores
          </p>
        </div>
      </div>

      {/* Memory Utilization */}
      <div data-testid="memory-utilization" className="bg-white rounded-lg border p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-2">
            <Database className="w-5 h-5 text-purple-500" />
            <h3 className="font-medium">Memory Usage</h3>
          </div>
          <span className={`text-sm font-medium px-2 py-1 rounded ${getUsageColor(metrics.memory.usage_percentage)}`}>
            {metrics.memory.usage_percentage.toFixed(1)}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={`h-2 rounded-full transition-all duration-300 ${getProgressColor(metrics.memory.usage_percentage)}`}
              style={{ width: `${metrics.memory.usage_percentage}%` }}
            ></div>
          </div>
          <p className="text-sm text-gray-600">
            {metrics.memory.used_gb.toFixed(1)} / {metrics.memory.total_gb} GB
          </p>
        </div>
      </div>

      {/* Storage Utilization */}
      <div data-testid="storage-utilization" className="bg-white rounded-lg border p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-2">
            <HardDrive className="w-5 h-5 text-green-500" />
            <h3 className="font-medium">Storage Usage</h3>
          </div>
          <span className={`text-sm font-medium px-2 py-1 rounded ${getUsageColor(metrics.storage.usage_percentage)}`}>
            {metrics.storage.usage_percentage.toFixed(1)}%
          </span>
        </div>
        <div className="space-y-2">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={`h-2 rounded-full transition-all duration-300 ${getProgressColor(metrics.storage.usage_percentage)}`}
              style={{ width: `${metrics.storage.usage_percentage}%` }}
            ></div>
          </div>
          <p className="text-sm text-gray-600">
            {(metrics.storage.used_gb / 1000).toFixed(1)} / {(metrics.storage.total_gb / 1000).toFixed(1)} TB
          </p>
        </div>
      </div>

      {/* Network Utilization */}
      <div data-testid="network-utilization" className="bg-white rounded-lg border p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-2">
            <Wifi className="w-5 h-5 text-orange-500" />
            <h3 className="font-medium">Network I/O</h3>
          </div>
        </div>
        <div className="space-y-3">
          <div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Ingress</span>
              <span className="font-medium">{metrics.network.ingress_mbps.toFixed(1)} Mbps</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-1.5 mt-1">
              <div 
                className="bg-blue-500 h-1.5 rounded-full transition-all duration-300"
                style={{ width: `${Math.min((metrics.network.ingress_mbps / 1000) * 100, 100)}%` }}
              ></div>
            </div>
          </div>
          <div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Egress</span>
              <span className="font-medium">{metrics.network.egress_mbps.toFixed(1)} Mbps</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-1.5 mt-1">
              <div 
                className="bg-purple-500 h-1.5 rounded-full transition-all duration-300"
                style={{ width: `${Math.min((metrics.network.egress_mbps / 1000) * 100, 100)}%` }}
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}