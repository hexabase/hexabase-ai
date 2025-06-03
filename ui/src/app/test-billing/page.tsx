'use client';

import { useState } from 'react';
import BillingOverview from '@/components/billing/billing-overview';
import SubscriptionPlansModal from '@/components/billing/subscription-plans-modal';
import PaymentMethodsModal from '@/components/billing/payment-methods-modal';
import { Button } from '@/components/ui/button';

export default function TestBillingPage() {
  const [showPlansModal, setShowPlansModal] = useState(false);
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [upgradeSuccess, setUpgradeSuccess] = useState(false);

  const handleUpgradeSuccess = () => {
    setUpgradeSuccess(true);
    setShowPlansModal(false);
  };

  return (
    <div data-testid="billing-test-page" className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-6xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Billing Management Test</h1>
            <p className="text-gray-600 mt-1">Test billing components and functionality</p>
          </div>
          <div className="flex space-x-3">
            <Button onClick={() => setShowPlansModal(true)}>
              Test Plans Modal
            </Button>
            <Button onClick={() => setShowPaymentModal(true)} variant="outline">
              Test Payment Modal
            </Button>
          </div>
        </div>

        {upgradeSuccess && (
          <div data-testid="upgrade-success" className="bg-green-50 border border-green-200 rounded-lg p-4">
            <p className="text-green-800 font-medium">Subscription upgraded successfully</p>
            <p className="text-green-600 text-sm mt-1">Your new plan is now active</p>
          </div>
        )}

        <BillingOverview 
          orgId="test-org"
          onUpgradePlan={() => setShowPlansModal(true)}
          onManagePayment={() => setShowPaymentModal(true)}
        />

        {/* Mock Usage Analytics Section */}
        <div data-testid="usage-analytics" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Usage Analytics</h2>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div data-testid="workspaces-usage-chart" className="bg-gray-50 rounded p-4">
              <h3 className="font-medium mb-2">Workspaces Usage</h3>
              <div className="h-32 bg-blue-100 rounded flex items-center justify-center">
                <span className="text-blue-600">Chart: 4/10 Workspaces</span>
              </div>
            </div>
            <div data-testid="storage-usage-chart" className="bg-gray-50 rounded p-4">
              <h3 className="font-medium mb-2">Storage Usage</h3>
              <div className="h-32 bg-green-100 rounded flex items-center justify-center">
                <span className="text-green-600">Chart: 23.5/100 GB</span>
              </div>
            </div>
            <div data-testid="bandwidth-usage-chart" className="bg-gray-50 rounded p-4">
              <h3 className="font-medium mb-2">Bandwidth Usage</h3>
              <div className="h-32 bg-orange-100 rounded flex items-center justify-center">
                <span className="text-orange-600">Chart: 156/500 GB</span>
              </div>
            </div>
          </div>

          <div className="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div data-testid="current-period-usage" className="bg-gray-50 rounded p-4">
              <h3 className="font-medium mb-2">Current Period</h3>
              <div data-testid="usage-percentage" className="text-sm text-gray-600">
                December 2024 usage summary
              </div>
            </div>
            <div data-testid="billing-forecast" className="bg-gray-50 rounded p-4">
              <h3 className="font-medium mb-2">Billing Forecast</h3>
              <div data-testid="projected-cost" className="text-lg font-bold">$32.50</div>
              <div className="text-sm text-gray-600">Projected next month</div>
            </div>
          </div>
        </div>

        {/* Mock Billing History */}
        <div data-testid="billing-history" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Recent Invoices</h2>
          <div data-testid="invoice-list" className="space-y-3">
            <div data-testid="invoice-item" className="flex items-center justify-between p-3 bg-gray-50 rounded">
              <div>
                <div className="flex items-center space-x-2">
                  <span className="font-medium">Invoice #12345</span>
                  <span data-testid="invoice-status" className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">Paid</span>
                </div>
                <div className="text-sm text-gray-600">
                  <span data-testid="invoice-date">Dec 01, 2024</span> • 
                  <span data-testid="invoice-amount" className="ml-1">$29.00</span>
                </div>
              </div>
              <Button size="sm" variant="outline" data-testid="download-pdf">
                Download PDF
              </Button>
            </div>
          </div>
          <div data-testid="invoice-pagination" className="mt-4 text-center">
            <Button variant="outline" size="sm">View All Invoices</Button>
          </div>
        </div>

        {/* Mock Billing Settings */}
        <div data-testid="billing-settings" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">Billing Settings</h2>
          <div data-testid="billing-alerts" className="space-y-4">
            <div data-testid="usage-threshold-alert">
              <h3 className="font-medium mb-2">Usage Alerts</h3>
              <div className="flex items-center space-x-3">
                <input 
                  data-testid="usage-threshold-input"
                  type="number" 
                  defaultValue="80" 
                  className="w-20 px-2 py-1 border rounded"
                />
                <span>% threshold</span>
                <Button size="sm" data-testid="save-threshold">Save</Button>
              </div>
            </div>
            <div data-testid="billing-email-alert">
              <label className="flex items-center space-x-2">
                <input type="checkbox" defaultChecked />
                <span>Email notifications</span>
              </label>
            </div>
          </div>
          <div data-testid="notification-preferences" className="mt-4 space-y-3">
            <div data-testid="email-notifications">
              <h4 className="font-medium">Email Preferences</h4>
              <div className="ml-4 space-y-1">
                <label className="flex items-center space-x-2">
                  <input type="checkbox" defaultChecked />
                  <span className="text-sm">Invoice notifications</span>
                </label>
              </div>
            </div>
            <div data-testid="slack-notifications">
              <h4 className="font-medium">Slack Integration</h4>
              <input 
                type="url" 
                placeholder="Webhook URL" 
                className="mt-1 w-full px-3 py-2 border rounded"
              />
            </div>
          </div>
          <div data-testid="usage-alerts" className="mt-4 p-3 bg-blue-50 rounded">
            <h4 className="font-medium text-blue-800">Recommendations</h4>
            <ul className="mt-2 text-sm text-blue-700 space-y-1">
              <li>• Consider upgrading for unlimited bandwidth</li>
              <li>• Storage usage is within normal limits</li>
            </ul>
          </div>
        </div>

        {showPlansModal && (
          <SubscriptionPlansModal
            orgId="test-org"
            isOpen={showPlansModal}
            onClose={() => setShowPlansModal(false)}
          />
        )}

        {showPaymentModal && (
          <PaymentMethodsModal
            orgId="test-org"
            isOpen={showPaymentModal}
            onClose={() => setShowPaymentModal(false)}
          />
        )}
      </div>
    </div>
  );
}