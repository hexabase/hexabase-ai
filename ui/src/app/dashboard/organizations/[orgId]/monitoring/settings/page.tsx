'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { monitoringApi, type AlertRule } from '@/lib/api-client';
import { ArrowLeft, Plus, Bell, Mail, MessageSquare, Webhook, Save, Trash2 } from 'lucide-react';
import AlertRuleModal from '@/components/monitoring/alert-rule-modal';

interface MonitoringSettingsPageProps {
  params: {
    orgId: string;
  };
}

export default function MonitoringSettingsPage({ params }: MonitoringSettingsPageProps) {
  const router = useRouter();
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [showRuleModal, setShowRuleModal] = useState(false);
  const [selectedRule, setSelectedRule] = useState<AlertRule | null>(null);
  const [thresholds, setThresholds] = useState({
    cpu: 80,
    memory: 85,
    storage: 90,
    pod_restart: 5
  });
  const [notifications, setNotifications] = useState({
    email: true,
    slack: false,
    webhook: false
  });
  const [retention, setRetention] = useState({
    metrics: 30,
    logs: 7,
    alerts: 90
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, [params.orgId]);

  const fetchSettings = async () => {
    try {
      const data = await monitoringApi.getAlertRules(params.orgId);
      setAlertRules(data.rules);
    } catch (error) {
      console.error('Failed to fetch monitoring settings:', error);
      // Use mock data for development
      setAlertRules(getMockAlertRules());
    } finally {
      setLoading(false);
    }
  };

  const getMockAlertRules = (): AlertRule[] => [
    {
      id: 'rule-1',
      name: 'High CPU Usage',
      description: 'Alert when CPU usage exceeds threshold',
      metric: 'cpu_usage_percentage',
      condition: 'above',
      threshold: 80,
      duration_minutes: 5,
      severity: 'warning',
      enabled: true,
      notification_channels: ['email', 'slack'],
      created_at: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'rule-2',
      name: 'Memory Pressure',
      description: 'Alert when memory usage is critical',
      metric: 'memory_usage_percentage',
      condition: 'above',
      threshold: 90,
      duration_minutes: 3,
      severity: 'critical',
      enabled: true,
      notification_channels: ['email', 'slack', 'webhook'],
      created_at: new Date(Date.now() - 15 * 24 * 60 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'rule-3',
      name: 'Pod Restart Loop',
      description: 'Alert when pod restarts exceed threshold',
      metric: 'pod_restart_count',
      condition: 'above',
      threshold: 3,
      duration_minutes: 10,
      severity: 'error',
      enabled: true,
      notification_channels: ['email'],
      created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString()
    }
  ];

  const handleSaveThresholds = async () => {
    setSaving(true);
    try {
      // In real app, would save thresholds via API
      await new Promise(resolve => setTimeout(resolve, 1000));
      console.log('Thresholds saved:', thresholds);
    } catch (error) {
      console.error('Failed to save thresholds:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteRule = async (ruleId: string) => {
    if (!confirm('Are you sure you want to delete this alert rule?')) return;
    
    try {
      await monitoringApi.deleteAlertRule(params.orgId, ruleId);
      await fetchSettings();
    } catch (error) {
      console.error('Failed to delete alert rule:', error);
    }
  };

  const handleEditRule = (rule: AlertRule) => {
    setSelectedRule(rule);
    setShowRuleModal(true);
  };

  const handleRuleSaved = async () => {
    setShowRuleModal(false);
    setSelectedRule(null);
    await fetchSettings();
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'text-red-600 bg-red-50';
      case 'error':
        return 'text-red-600 bg-red-50';
      case 'warning':
        return 'text-yellow-600 bg-yellow-50';
      case 'info':
        return 'text-blue-600 bg-blue-50';
      default:
        return 'text-gray-600 bg-gray-50';
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
            <h1 className="text-2xl font-bold text-gray-900">Monitoring Settings</h1>
            <p className="text-gray-600 mt-1">Configure alerts, thresholds, and notifications</p>
          </div>
        </div>
      </div>

      <div data-testid="monitoring-settings" className="space-y-6">
        {/* Alert Rules */}
        <div data-testid="alert-rules" className="bg-white rounded-lg border">
          <div className="p-6 border-b flex items-center justify-between">
            <h2 className="text-lg font-semibold">Alert Rules</h2>
            <Button
              onClick={() => {
                setSelectedRule(null);
                setShowRuleModal(true);
              }}
              data-testid="add-alert-rule"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Rule
            </Button>
          </div>
          
          {loading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : alertRules.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              No alert rules configured
            </div>
          ) : (
            <div className="divide-y">
              {alertRules.map((rule) => (
                <div key={rule.id} className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-1">
                        <h3 className="font-medium">{rule.name}</h3>
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(rule.severity)}`}>
                          {rule.severity}
                        </span>
                        {!rule.enabled && (
                          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
                            Disabled
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 mb-2">{rule.description}</p>
                      <div className="flex items-center space-x-4 text-sm text-gray-500">
                        <span>Metric: {rule.metric}</span>
                        <span>Condition: {rule.condition} {rule.threshold}</span>
                        <span>Duration: {rule.duration_minutes} min</span>
                        <span>Channels: {rule.notification_channels.join(', ')}</span>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleEditRule(rule)}
                      >
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleDeleteRule(rule.id)}
                        className="text-red-600 hover:bg-red-50"
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Metric Thresholds */}
        <div data-testid="metric-thresholds" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Default Metric Thresholds</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                CPU Usage Threshold (%)
              </label>
              <input
                type="number"
                min="1"
                max="100"
                data-testid="cpu-threshold"
                value={thresholds.cpu}
                onChange={(e) => setThresholds({ ...thresholds, cpu: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Memory Usage Threshold (%)
              </label>
              <input
                type="number"
                min="1"
                max="100"
                data-testid="memory-threshold"
                value={thresholds.memory}
                onChange={(e) => setThresholds({ ...thresholds, memory: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Storage Usage Threshold (%)
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={thresholds.storage}
                onChange={(e) => setThresholds({ ...thresholds, storage: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Pod Restart Threshold
              </label>
              <input
                type="number"
                min="1"
                value={thresholds.pod_restart}
                onChange={(e) => setThresholds({ ...thresholds, pod_restart: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
          
          <div className="mt-4 flex justify-end">
            <Button onClick={handleSaveThresholds} disabled={saving}>
              <Save className="w-4 h-4 mr-2" />
              {saving ? 'Saving...' : 'Save Thresholds'}
            </Button>
          </div>
        </div>

        {/* Notification Channels */}
        <div data-testid="notification-channels" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Notification Channels</h2>
          
          <div className="space-y-4">
            <div data-testid="email-channel" className="flex items-center justify-between p-4 border rounded-lg">
              <div className="flex items-center space-x-3">
                <Mail className="w-5 h-5 text-gray-500" />
                <div>
                  <h3 className="font-medium">Email Notifications</h3>
                  <p className="text-sm text-gray-600">Send alerts to organization admins</p>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={notifications.email}
                  onChange={(e) => setNotifications({ ...notifications, email: e.target.checked })}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
            
            <div data-testid="slack-channel" className="flex items-center justify-between p-4 border rounded-lg">
              <div className="flex items-center space-x-3">
                <MessageSquare className="w-5 h-5 text-gray-500" />
                <div>
                  <h3 className="font-medium">Slack Integration</h3>
                  <p className="text-sm text-gray-600">Post alerts to Slack channels</p>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={notifications.slack}
                  onChange={(e) => setNotifications({ ...notifications, slack: e.target.checked })}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
            
            <div data-testid="webhook-channel" className="flex items-center justify-between p-4 border rounded-lg">
              <div className="flex items-center space-x-3">
                <Webhook className="w-5 h-5 text-gray-500" />
                <div>
                  <h3 className="font-medium">Webhook</h3>
                  <p className="text-sm text-gray-600">Send alerts to custom webhook endpoints</p>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={notifications.webhook}
                  onChange={(e) => setNotifications({ ...notifications, webhook: e.target.checked })}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
            </div>
          </div>
        </div>

        {/* Data Retention */}
        <div data-testid="data-retention" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Data Retention Settings</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Metrics Retention (days)
              </label>
              <input
                type="number"
                min="1"
                max="365"
                data-testid="metrics-retention"
                value={retention.metrics}
                onChange={(e) => setRetention({ ...retention, metrics: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Logs Retention (days)
              </label>
              <input
                type="number"
                min="1"
                max="90"
                data-testid="logs-retention"
                value={retention.logs}
                onChange={(e) => setRetention({ ...retention, logs: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Alerts History (days)
              </label>
              <input
                type="number"
                min="1"
                max="365"
                value={retention.alerts}
                onChange={(e) => setRetention({ ...retention, alerts: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
          </div>
          
          <p className="text-sm text-gray-600 mt-4">
            Data older than the retention period will be automatically deleted to optimize storage.
          </p>
        </div>
      </div>

      {showRuleModal && (
        <AlertRuleModal
          rule={selectedRule}
          orgId={params.orgId}
          isOpen={showRuleModal}
          onClose={() => {
            setShowRuleModal(false);
            setSelectedRule(null);
          }}
          onSave={handleRuleSaved}
        />
      )}
    </div>
  );
}