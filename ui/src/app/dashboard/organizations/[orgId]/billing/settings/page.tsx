'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { billingApi } from '@/lib/api-client';
import { ArrowLeft, Bell, Mail, MessageSquare, Save } from 'lucide-react';

interface BillingSettingsPageProps {
  params: {
    orgId: string;
  };
}

export default function BillingSettingsPage({ params }: BillingSettingsPageProps) {
  const router = useRouter();
  const [settings, setSettings] = useState({
    usage_threshold: 80,
    email_notifications: true,
    slack_webhook: ''
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const data = await billingApi.getBillingSettings(params.orgId);
      setSettings(data);
    } catch (error) {
      console.error('Failed to fetch billing settings:', error);
      // Use default settings if API fails
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSettings = async () => {
    setSaving(true);
    setMessage(null);

    try {
      await billingApi.updateBillingSettings(params.orgId, settings);
      setMessage({ type: 'success', text: 'Billing settings updated successfully' });
    } catch (error) {
      console.error('Failed to update billing settings:', error);
      setMessage({ type: 'error', text: 'Failed to update settings. Please try again.' });
    } finally {
      setSaving(false);
    }
  };

  const handleThresholdSave = async () => {
    setSaving(true);
    setMessage(null);

    try {
      await billingApi.updateBillingSettings(params.orgId, { 
        usage_threshold: settings.usage_threshold 
      });
      setMessage({ type: 'success', text: 'Alert threshold updated' });
    } catch (error) {
      console.error('Failed to update threshold:', error);
      setMessage({ type: 'error', text: 'Failed to update threshold' });
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/billing`)}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Billing
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Billing Settings</h1>
            <p className="text-gray-600 mt-1">Configure billing alerts and notifications</p>
          </div>
        </div>
      </div>

      {message && (
        <div className={`rounded-md p-4 ${
          message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'
        }`}>
          {message.text}
        </div>
      )}

      {loading ? (
        <div data-testid="billing-settings" className="animate-pulse">
          <div className="bg-white rounded-lg border p-6">
            <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
            <div className="h-32 bg-gray-200 rounded"></div>
          </div>
        </div>
      ) : (
        <div data-testid="billing-settings" className="space-y-6">
          {/* Billing Alerts */}
          <div data-testid="billing-alerts" className="bg-white rounded-lg border p-6">
            <div className="flex items-center space-x-2 mb-6">
              <Bell className="w-5 h-5 text-gray-500" />
              <h2 className="text-lg font-semibold">Billing Alerts</h2>
            </div>

            <div className="space-y-6">
              {/* Usage Threshold Alert */}
              <div data-testid="usage-threshold-alert" className="space-y-3">
                <h3 className="font-medium">Usage Threshold Alert</h3>
                <p className="text-sm text-gray-600">
                  Get notified when your usage exceeds a certain percentage of your plan limits
                </p>
                <div className="flex items-center space-x-4">
                  <div className="flex-1 max-w-xs">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Threshold Percentage
                    </label>
                    <div className="flex items-center space-x-2">
                      <input
                        type="number"
                        min="1"
                        max="100"
                        data-testid="usage-threshold-input"
                        value={settings.usage_threshold}
                        onChange={(e) => setSettings({ ...settings, usage_threshold: parseInt(e.target.value) })}
                        className="w-20 px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      />
                      <span className="text-sm text-gray-500">%</span>
                      <Button
                        onClick={handleThresholdSave}
                        size="sm"
                        disabled={saving}
                        data-testid="save-threshold"
                      >
                        <Save className="w-4 h-4 mr-1" />
                        Save
                      </Button>
                    </div>
                  </div>
                </div>
              </div>

              {/* Billing Email Alerts */}
              <div data-testid="billing-email-alert" className="space-y-3">
                <h3 className="font-medium">Invoice and Payment Alerts</h3>
                <p className="text-sm text-gray-600">
                  Receive email notifications for new invoices and payment confirmations
                </p>
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={settings.email_notifications}
                    onChange={(e) => setSettings({ ...settings, email_notifications: e.target.checked })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm">Enable email notifications</span>
                </label>
              </div>
            </div>
          </div>

          {/* Notification Preferences */}
          <div data-testid="notification-preferences" className="bg-white rounded-lg border p-6">
            <div className="flex items-center space-x-2 mb-6">
              <MessageSquare className="w-5 h-5 text-gray-500" />
              <h2 className="text-lg font-semibold">Notification Preferences</h2>
            </div>

            <div className="space-y-6">
              {/* Email Notifications */}
              <div data-testid="email-notifications">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center space-x-2">
                    <Mail className="w-4 h-4 text-gray-500" />
                    <h3 className="font-medium">Email Notifications</h3>
                  </div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={settings.email_notifications}
                      onChange={(e) => setSettings({ ...settings, email_notifications: e.target.checked })}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm">Enabled</span>
                  </label>
                </div>
                <div className="pl-6 space-y-2">
                  <div className="flex items-center space-x-2">
                    <input type="checkbox" defaultChecked className="rounded border-gray-300 text-blue-600" />
                    <span className="text-sm text-gray-600">New invoices</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <input type="checkbox" defaultChecked className="rounded border-gray-300 text-blue-600" />
                    <span className="text-sm text-gray-600">Payment confirmations</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <input type="checkbox" defaultChecked className="rounded border-gray-300 text-blue-600" />
                    <span className="text-sm text-gray-600">Usage alerts</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <input type="checkbox" className="rounded border-gray-300 text-blue-600" />
                    <span className="text-sm text-gray-600">Plan changes</span>
                  </div>
                </div>
              </div>

              {/* Slack Notifications */}
              <div data-testid="slack-notifications">
                <div className="flex items-center space-x-2 mb-3">
                  <MessageSquare className="w-4 h-4 text-gray-500" />
                  <h3 className="font-medium">Slack Notifications</h3>
                </div>
                <p className="text-sm text-gray-600 mb-3">
                  Send billing notifications to a Slack channel via webhook
                </p>
                <div className="space-y-3">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Slack Webhook URL
                    </label>
                    <input
                      type="url"
                      value={settings.slack_webhook}
                      onChange={(e) => setSettings({ ...settings, slack_webhook: e.target.value })}
                      placeholder="https://hooks.slack.com/services/..."
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <p className="text-xs text-gray-500">
                    <a href="https://api.slack.com/messaging/webhooks" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                      Learn how to create a Slack webhook
                    </a>
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Save Button */}
          <div className="flex justify-end">
            <Button
              onClick={handleSaveSettings}
              disabled={saving}
              className="px-6"
            >
              {saving ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="w-4 h-4 mr-2" />
                  Save All Settings
                </>
              )}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}