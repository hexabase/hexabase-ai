'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { monitoringApi, type Alert } from '@/lib/api-client';
import { ArrowLeft, AlertCircle, AlertTriangle, Info, XCircle, CheckCircle, Clock } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface AlertsPageProps {
  params: {
    orgId: string;
  };
}

export default function AlertsPage({ params }: AlertsPageProps) {
  const router = useRouter();
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [activeAlerts, setActiveAlerts] = useState<Alert[]>([]);
  const [incidentHistory, setIncidentHistory] = useState<Alert[]>([]);
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [workspaceFilter, setWorkspaceFilter] = useState<string>('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAlerts();
    const interval = setInterval(fetchAlerts, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [severityFilter, workspaceFilter]);

  const fetchAlerts = async () => {
    try {
      const params: any = {};
      if (severityFilter !== 'all') params.severity = severityFilter;
      if (workspaceFilter !== 'all') params.workspace_id = workspaceFilter;
      
      const data = await monitoringApi.getAlerts(params.orgId, params);
      const all = data.alerts;
      setAlerts(all);
      setActiveAlerts(all.filter(a => a.status === 'active'));
      setIncidentHistory(all.filter(a => a.status === 'resolved'));
    } catch (error) {
      console.error('Failed to fetch alerts:', error);
      // Use mock data for development
      const mockAlerts = getMockAlerts();
      setAlerts(mockAlerts);
      setActiveAlerts(mockAlerts.filter(a => a.status === 'active'));
      setIncidentHistory(mockAlerts.filter(a => a.status === 'resolved'));
    } finally {
      setLoading(false);
    }
  };

  const getMockAlerts = (): Alert[] => [
    {
      id: 'alert-1',
      severity: 'critical',
      title: 'High Memory Usage',
      description: 'Memory usage exceeded 90% threshold on production workspace',
      workspace_id: 'ws-prod',
      workspace_name: 'Production',
      resource_type: 'node',
      resource_name: 'node-prod-01',
      metric_name: 'memory_usage_percentage',
      current_value: 92.5,
      threshold_value: 90,
      triggered_at: new Date(Date.now() - 15 * 60000).toISOString(),
      status: 'active'
    },
    {
      id: 'alert-2',
      severity: 'warning',
      title: 'CPU Usage Warning',
      description: 'CPU usage is approaching threshold on staging workspace',
      workspace_id: 'ws-staging',
      workspace_name: 'Staging',
      resource_type: 'pod',
      resource_name: 'api-server-7d9f8c',
      metric_name: 'cpu_usage_percentage',
      current_value: 78.3,
      threshold_value: 80,
      triggered_at: new Date(Date.now() - 30 * 60000).toISOString(),
      status: 'active'
    },
    {
      id: 'alert-3',
      severity: 'info',
      title: 'Deployment Completed',
      description: 'New version deployed successfully to development workspace',
      workspace_id: 'ws-dev',
      workspace_name: 'Development',
      resource_type: 'deployment',
      resource_name: 'frontend-app',
      metric_name: 'deployment_status',
      current_value: 1,
      threshold_value: 1,
      triggered_at: new Date(Date.now() - 2 * 60 * 60000).toISOString(),
      resolved_at: new Date(Date.now() - 2 * 60 * 60000).toISOString(),
      status: 'resolved'
    },
    {
      id: 'alert-4',
      severity: 'error',
      title: 'Pod Crash Loop',
      description: 'Pod is in crash loop back off state',
      workspace_id: 'ws-prod',
      workspace_name: 'Production',
      resource_type: 'pod',
      resource_name: 'worker-8b4c5d',
      metric_name: 'restart_count',
      current_value: 5,
      threshold_value: 3,
      triggered_at: new Date(Date.now() - 45 * 60000).toISOString(),
      status: 'acknowledged'
    }
  ];

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <XCircle className="w-5 h-5 text-red-500" />;
      case 'error':
        return <AlertCircle className="w-5 h-5 text-red-500" />;
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
      case 'info':
        return <Info className="w-5 h-5 text-blue-500" />;
      default:
        return <AlertCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'error':
        return 'bg-red-50 text-red-700 border-red-100';
      case 'warning':
        return 'bg-yellow-50 text-yellow-800 border-yellow-200';
      case 'info':
        return 'bg-blue-50 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-50 text-gray-800 border-gray-200';
    }
  };

  const handleAcknowledge = async (alertId: string) => {
    try {
      await monitoringApi.acknowledgeAlert(params.orgId, alertId);
      await fetchAlerts();
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/monitoring`)}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Monitoring
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Alerts & Incidents</h1>
            <p className="text-gray-600 mt-1">Monitor and manage system alerts</p>
          </div>
        </div>
      </div>

      <div data-testid="alerts-dashboard" className="space-y-6">
        {/* Filters */}
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center space-x-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Severity</label>
              <select
                data-testid="severity-filter"
                value={severityFilter}
                onChange={(e) => setSeverityFilter(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Severities</option>
                <option value="critical">Critical</option>
                <option value="error">Error</option>
                <option value="warning">Warning</option>
                <option value="info">Info</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Workspace</label>
              <select
                data-testid="workspace-filter"
                value={workspaceFilter}
                onChange={(e) => setWorkspaceFilter(e.target.value)}
                className="border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Workspaces</option>
                <option value="ws-prod">Production</option>
                <option value="ws-staging">Staging</option>
                <option value="ws-dev">Development</option>
              </select>
            </div>
          </div>
        </div>

        {/* Active Alerts */}
        <div data-testid="active-alerts" className="bg-white rounded-lg border">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">Active Alerts ({activeAlerts.length})</h2>
          </div>
          
          {loading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : activeAlerts.length === 0 ? (
            <div className="p-8 text-center">
              <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-2" />
              <p className="text-gray-500">No active alerts</p>
            </div>
          ) : (
            <div className="divide-y">
              {activeAlerts.map((alert) => (
                <div key={alert.id} data-testid="alert-item" className="p-4 hover:bg-gray-50">
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3">
                      {getSeverityIcon(alert.severity)}
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <h3 data-testid="alert-title" className="font-medium">{alert.title}</h3>
                          <span data-testid="alert-severity" className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(alert.severity)}`}>
                            {alert.severity.charAt(0).toUpperCase() + alert.severity.slice(1)}
                          </span>
                          {alert.status === 'acknowledged' && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                              Acknowledged
                            </span>
                          )}
                        </div>
                        <p className="text-sm text-gray-600 mb-2">{alert.description}</p>
                        <div className="flex items-center space-x-4 text-xs text-gray-500">
                          <span>Workspace: {alert.workspace_name}</span>
                          <span>Resource: {alert.resource_name}</span>
                          <span>Value: {alert.current_value.toFixed(1)} (threshold: {alert.threshold_value})</span>
                          <span data-testid="alert-time" className="flex items-center">
                            <Clock className="w-3 h-3 mr-1" />
                            {formatDistanceToNow(new Date(alert.triggered_at))} ago
                          </span>
                        </div>
                      </div>
                    </div>
                    
                    {alert.status !== 'acknowledged' && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleAcknowledge(alert.id)}
                      >
                        Acknowledge
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Incident History */}
        <div data-testid="incident-history" className="bg-white rounded-lg border">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">Incident History</h2>
          </div>
          
          {incidentHistory.length === 0 ? (
            <div className="p-8 text-center">
              <p className="text-gray-500">No resolved incidents</p>
            </div>
          ) : (
            <div className="divide-y">
              {incidentHistory.map((incident) => (
                <div key={incident.id} className="p-4">
                  <div className="flex items-start space-x-3">
                    <CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-1">
                        <h3 className="font-medium text-gray-700">{incident.title}</h3>
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(incident.severity)}`}>
                          {incident.severity.charAt(0).toUpperCase() + incident.severity.slice(1)}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mb-2">{incident.description}</p>
                      <div className="flex items-center space-x-4 text-xs text-gray-500">
                        <span>Workspace: {incident.workspace_name}</span>
                        <span>Duration: {formatDistanceToNow(new Date(incident.triggered_at))}</span>
                        <span>Resolved: {formatDistanceToNow(new Date(incident.resolved_at!))} ago</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}