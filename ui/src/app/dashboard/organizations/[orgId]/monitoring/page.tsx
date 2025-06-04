'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import ClusterHealthCard from '@/components/monitoring/cluster-health-card';
import ResourceUtilization from '@/components/monitoring/resource-utilization';
import MetricsCharts from '@/components/monitoring/metrics-charts';
import WorkspaceMetrics from '@/components/monitoring/workspace-metrics';
import { monitoringApi } from '@/lib/api-client';
import { ArrowLeft, Activity, AlertCircle, BarChart3, FileText } from 'lucide-react';

interface MonitoringPageProps {
  params: {
    orgId: string;
  };
}

export default function MonitoringPage({ params }: MonitoringPageProps) {
  const router = useRouter();
  const [timeRange, setTimeRange] = useState('1h');
  const [loading, setLoading] = useState(true);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}`)}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Monitoring & Observability</h1>
            <p className="text-gray-600 mt-1">Real-time cluster monitoring and performance insights</p>
          </div>
        </div>
        
        <div className="flex items-center space-x-3">
          <select
            data-testid="time-range-selector"
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="5m">Last 5 minutes</option>
            <option value="15m">Last 15 minutes</option>
            <option value="1h">Last 1 hour</option>
            <option value="6h">Last 6 hours</option>
            <option value="24h">Last 24 hours</option>
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
          </select>
          
          <Button
            variant="outline"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/monitoring/alerts`)}
          >
            <AlertCircle className="w-4 h-4 mr-2" />
            Alerts
          </Button>
          
          <Button
            variant="outline"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/monitoring/logs`)}
          >
            <FileText className="w-4 h-4 mr-2" />
            Logs
          </Button>
          
          <Button
            variant="outline"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/monitoring/insights`)}
          >
            <BarChart3 className="w-4 h-4 mr-2" />
            Insights
          </Button>
        </div>
      </div>

      <div data-testid="monitoring-dashboard" className="space-y-6">
        {/* Cluster Health Overview */}
        <ClusterHealthCard orgId={params.orgId} />

        {/* Resource Utilization */}
        <ResourceUtilization orgId={params.orgId} />

        {/* Metrics Charts */}
        <MetricsCharts orgId={params.orgId} timeRange={timeRange} />

        {/* Workspace Metrics */}
        <WorkspaceMetrics orgId={params.orgId} />
      </div>
    </div>
  );
}