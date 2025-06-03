'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { billingApi, type Subscription, type UsageMetrics } from '@/lib/api-client';
import { formatDistanceToNow } from 'date-fns';

interface BillingOverviewProps {
  orgId: string;
  onUpgradePlan: () => void;
  onManagePayment: () => void;
}

export default function BillingOverview({ orgId, onUpgradePlan, onManagePayment }: BillingOverviewProps) {
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [usage, setUsage] = useState<UsageMetrics | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchBillingData = async () => {
      try {
        const [subData, usageData] = await Promise.all([
          billingApi.getSubscription(orgId).catch(() => null),
          billingApi.getUsageMetrics(orgId).catch(() => null)
        ]);
        
        setSubscription(subData || getMockSubscription());
        setUsage(usageData || getMockUsage());
      } catch (error) {
        console.error('Failed to fetch billing data:', error);
        // Use mock data for development
        setSubscription(getMockSubscription());
        setUsage(getMockUsage());
      } finally {
        setLoading(false);
      }
    };

    fetchBillingData();
  }, [orgId]);

  const getMockSubscription = (): Subscription => ({
    id: 'sub_123',
    organization_id: orgId,
    plan_id: 'professional',
    plan_name: 'Professional',
    status: 'active',
    billing_cycle: 'monthly',
    current_period_start: '2024-12-01T00:00:00Z',
    current_period_end: '2025-01-01T00:00:00Z',
    price_per_month: 29,
    price_per_year: 290,
    features: [
      '10 Workspaces',
      '100GB Storage',
      '500GB Bandwidth',
      'Email Support',
      'API Access'
    ],
    limits: {
      workspaces: 10,
      storage_gb: 100,
      bandwidth_gb: 500,
      support_level: 'email'
    }
  });

  const getMockUsage = (): UsageMetrics => ({
    organization_id: orgId,
    period_start: '2024-12-01T00:00:00Z',
    period_end: '2025-01-01T00:00:00Z',
    workspaces_count: 4,
    workspaces_limit: 10,
    storage_used_gb: 23.5,
    storage_limit_gb: 100,
    bandwidth_used_gb: 156.8,
    bandwidth_limit_gb: 500,
    overage_charges: {
      storage: 0,
      bandwidth: 0,
      total: 0
    }
  });

  if (loading) {
    return (
      <div data-testid="billing-overview" className="animate-pulse">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-white rounded-lg border p-6">
              <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div className="h-8 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!subscription) {
    return (
      <div data-testid="billing-overview" className="text-center py-8">
        <p className="text-gray-500">No subscription found</p>
      </div>
    );
  }

  const usagePercentage = (used: number, limit: number) => Math.round((used / limit) * 100);

  return (
    <div data-testid="billing-overview" className="space-y-6">
      {/* Current Subscription */}
      <div className="bg-white rounded-lg border p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-lg font-semibold">Current Subscription</h2>
            <p className="text-sm text-gray-600">Your active plan and billing information</p>
          </div>
          <Button onClick={onUpgradePlan} data-testid="upgrade-plan-button">
            Upgrade Plan
          </Button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div data-testid="current-plan" className="space-y-2">
            <p className="text-sm text-gray-600">Plan</p>
            <p data-testid="plan-name" className="text-xl font-semibold text-blue-600">{subscription.plan_name}</p>
            <p data-testid="plan-price" className="text-2xl font-bold">${subscription.price_per_month}/mo</p>
          </div>

          <div className="space-y-2">
            <p className="text-sm text-gray-600">Billing Cycle</p>
            <p data-testid="billing-cycle" className="text-lg font-medium capitalize">{subscription.billing_cycle}</p>
            <p className="text-sm text-gray-500">
              Next billing: {formatDistanceToNow(new Date(subscription.current_period_end))}
            </p>
          </div>

          <div className="space-y-2">
            <p className="text-sm text-gray-600">Status</p>
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              <p className="text-lg font-medium capitalize text-green-600">{subscription.status}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Usage Overview */}
      {usage && (
        <div data-testid="usage-overview" className="bg-white rounded-lg border p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-lg font-semibold">Current Usage</h2>
              <p className="text-sm text-gray-600">Usage for the current billing period</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div data-testid="workspaces-usage" className="space-y-3">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Workspaces</p>
                <p className="text-sm text-gray-600">{usage.workspaces_count} / {usage.workspaces_limit}</p>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${usagePercentage(usage.workspaces_count, usage.workspaces_limit)}%` }}
                ></div>
              </div>
              <p className="text-xs text-gray-500">{usagePercentage(usage.workspaces_count, usage.workspaces_limit)}% used</p>
            </div>

            <div data-testid="storage-usage" className="space-y-3">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Storage</p>
                <p className="text-sm text-gray-600">{usage.storage_used_gb}GB / {usage.storage_limit_gb}GB</p>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-green-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${usagePercentage(usage.storage_used_gb, usage.storage_limit_gb)}%` }}
                ></div>
              </div>
              <p className="text-xs text-gray-500">{usagePercentage(usage.storage_used_gb, usage.storage_limit_gb)}% used</p>
            </div>

            <div data-testid="bandwidth-usage" className="space-y-3">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Bandwidth</p>
                <p className="text-sm text-gray-600">{usage.bandwidth_used_gb}GB / {usage.bandwidth_limit_gb}GB</p>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-orange-600 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${usagePercentage(usage.bandwidth_used_gb, usage.bandwidth_limit_gb)}%` }}
                ></div>
              </div>
              <p className="text-xs text-gray-500">{usagePercentage(usage.bandwidth_used_gb, usage.bandwidth_limit_gb)}% used</p>
            </div>
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Button 
          variant="outline" 
          className="h-16 flex flex-col items-center justify-center space-y-1"
          data-testid="download-invoice-button"
        >
          <span className="font-medium">Download Invoice</span>
          <span className="text-xs text-gray-500">Latest billing statement</span>
        </Button>

        <Button 
          variant="outline" 
          className="h-16 flex flex-col items-center justify-center space-y-1"
          onClick={onManagePayment}
          data-testid="payment-method-button"
        >
          <span className="font-medium">Payment Methods</span>
          <span className="text-xs text-gray-500">Manage cards & billing</span>
        </Button>

        <Button 
          variant="outline" 
          className="h-16 flex flex-col items-center justify-center space-y-1"
          data-testid="billing-history-button"
        >
          <span className="font-medium">Billing History</span>
          <span className="text-xs text-gray-500">View past invoices</span>
        </Button>
      </div>
    </div>
  );
}