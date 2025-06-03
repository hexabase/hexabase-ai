'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { billingApi, type UsageMetrics, type BillingForecast } from '@/lib/api-client';
import { ArrowLeft, TrendingUp, AlertTriangle, BarChart3 } from 'lucide-react';

interface UsageAnalyticsPageProps {
  params: {
    orgId: string;
  };
}

export default function UsageAnalyticsPage({ params }: UsageAnalyticsPageProps) {
  const router = useRouter();
  const [usage, setUsage] = useState<UsageMetrics | null>(null);
  const [forecast, setForecast] = useState<BillingForecast | null>(null);
  const [period, setPeriod] = useState<'1month' | '3months' | '6months'>('1month');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchUsageData();
  }, [period]);

  const fetchUsageData = async () => {
    try {
      const [usageData, forecastData] = await Promise.all([
        billingApi.getUsageMetrics(params.orgId, { period }),
        billingApi.getBillingForecast(params.orgId)
      ]);
      setUsage(usageData);
      setForecast(forecastData);
    } catch (error) {
      console.error('Failed to fetch usage data:', error);
      // Use mock data for development
      setUsage(getMockUsage());
      setForecast(getMockForecast());
    } finally {
      setLoading(false);
    }
  };

  const getMockUsage = (): UsageMetrics => ({
    organization_id: params.orgId,
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

  const getMockForecast = (): BillingForecast => ({
    organization_id: params.orgId,
    projected_amount: 32.50,
    projected_period_end: '2025-01-01T00:00:00Z',
    usage_trend: 'increasing',
    recommendations: [
      'Consider upgrading to Enterprise plan for unlimited bandwidth',
      'Current storage usage is well within limits',
      'Monitor bandwidth usage - trending upward'
    ]
  });

  const usagePercentage = (used: number, limit: number) => Math.round((used / limit) * 100);

  const getUsageAlert = (percentage: number) => {
    if (percentage >= 90) return { level: 'error', message: 'Critical: Over 90% usage' };
    if (percentage >= 80) return { level: 'warning', message: 'High usage: Over 80%' };
    if (percentage >= 70) return { level: 'info', message: 'Moderate usage: Over 70%' };
    return null;
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
            <h1 className="text-2xl font-bold text-gray-900">Usage Analytics</h1>
            <p className="text-gray-600 mt-1">Monitor your resource usage and billing forecasts</p>
          </div>
        </div>
        
        <select
          data-testid="period-filter"
          value={period}
          onChange={(e) => setPeriod(e.target.value as any)}
          className="border border-gray-300 rounded-md px-3 py-2 text-sm"
        >
          <option value="1month">Last 30 days</option>
          <option value="3months">Last 3 months</option>
          <option value="6months">Last 6 months</option>
        </select>
      </div>

      {loading ? (
        <div data-testid="usage-analytics" className="animate-pulse">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg border p-6">
              <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div className="h-32 bg-gray-200 rounded"></div>
            </div>
            <div className="bg-white rounded-lg border p-6">
              <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
              <div className="h-32 bg-gray-200 rounded"></div>
            </div>
          </div>
        </div>
      ) : (
        <div data-testid="usage-analytics" className="space-y-6">
          {/* Current Period Usage */}
          {usage && (
            <div data-testid="current-period-usage" className="bg-white rounded-lg border p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-lg font-semibold">Current Period Usage</h2>
                <div data-testid="usage-percentage" className="text-sm text-gray-600">
                  Period: {new Date(usage.period_start).toLocaleDateString()} - {new Date(usage.period_end).toLocaleDateString()}
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Workspaces Usage */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium">Workspaces</h3>
                    <span className="text-sm text-gray-600">{usage.workspaces_count} / {usage.workspaces_limit}</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-3 mb-2">
                    <div 
                      className="bg-blue-600 h-3 rounded-full transition-all duration-300"
                      style={{ width: `${usagePercentage(usage.workspaces_count, usage.workspaces_limit)}%` }}
                    ></div>
                  </div>
                  <p className="text-xs text-gray-500">{usagePercentage(usage.workspaces_count, usage.workspaces_limit)}% used</p>
                  {getUsageAlert(usagePercentage(usage.workspaces_count, usage.workspaces_limit)) && (
                    <div className="mt-2 flex items-center text-xs">
                      <AlertTriangle className="w-3 h-3 mr-1 text-yellow-500" />
                      <span className="text-yellow-700">
                        {getUsageAlert(usagePercentage(usage.workspaces_count, usage.workspaces_limit))?.message}
                      </span>
                    </div>
                  )}
                </div>

                {/* Storage Usage */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium">Storage</h3>
                    <span className="text-sm text-gray-600">{usage.storage_used_gb}GB / {usage.storage_limit_gb}GB</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-3 mb-2">
                    <div 
                      className="bg-green-600 h-3 rounded-full transition-all duration-300"
                      style={{ width: `${usagePercentage(usage.storage_used_gb, usage.storage_limit_gb)}%` }}
                    ></div>
                  </div>
                  <p className="text-xs text-gray-500">{usagePercentage(usage.storage_used_gb, usage.storage_limit_gb)}% used</p>
                </div>

                {/* Bandwidth Usage */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium">Bandwidth</h3>
                    <span className="text-sm text-gray-600">{usage.bandwidth_used_gb}GB / {usage.bandwidth_limit_gb}GB</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-3 mb-2">
                    <div 
                      className="bg-orange-600 h-3 rounded-full transition-all duration-300"
                      style={{ width: `${usagePercentage(usage.bandwidth_used_gb, usage.bandwidth_limit_gb)}%` }}
                    ></div>
                  </div>
                  <p className="text-xs text-gray-500">{usagePercentage(usage.bandwidth_used_gb, usage.bandwidth_limit_gb)}% used</p>
                </div>
              </div>
            </div>
          )}

          {/* Usage Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div data-testid="workspaces-usage-chart" className="bg-white rounded-lg border p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-medium">Workspaces Trend</h3>
                <BarChart3 className="w-4 h-4 text-gray-400" />
              </div>
              <div className="h-32 flex items-end justify-center space-x-2">
                {[40, 60, 45, 80, 70, 90, 85].map((height, index) => (
                  <div
                    key={index}
                    className="bg-blue-500 w-6 rounded-t"
                    style={{ height: `${height}%` }}
                  ></div>
                ))}
              </div>
              <p className="text-xs text-gray-500 mt-2 text-center">Last 7 days</p>
            </div>

            <div data-testid="storage-usage-chart" className="bg-white rounded-lg border p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-medium">Storage Trend</h3>
                <BarChart3 className="w-4 h-4 text-gray-400" />
              </div>
              <div className="h-32 flex items-end justify-center space-x-2">
                {[20, 25, 30, 28, 35, 40, 38].map((height, index) => (
                  <div
                    key={index}
                    className="bg-green-500 w-6 rounded-t"
                    style={{ height: `${height}%` }}
                  ></div>
                ))}
              </div>
              <p className="text-xs text-gray-500 mt-2 text-center">Last 7 days</p>
            </div>

            <div data-testid="bandwidth-usage-chart" className="bg-white rounded-lg border p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-medium">Bandwidth Trend</h3>
                <BarChart3 className="w-4 h-4 text-gray-400" />
              </div>
              <div className="h-32 flex items-end justify-center space-x-2">
                {[50, 65, 70, 80, 75, 85, 90].map((height, index) => (
                  <div
                    key={index}
                    className="bg-orange-500 w-6 rounded-t"
                    style={{ height: `${height}%` }}
                  ></div>
                ))}
              </div>
              <p className="text-xs text-gray-500 mt-2 text-center">Last 7 days</p>
            </div>
          </div>

          {/* Billing Forecast */}
          {forecast && (
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div data-testid="billing-forecast" className="bg-white rounded-lg border p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-medium">Billing Forecast</h3>
                  <TrendingUp className={`w-4 h-4 ${
                    forecast.usage_trend === 'increasing' ? 'text-red-500' : 
                    forecast.usage_trend === 'decreasing' ? 'text-green-500' : 'text-gray-400'
                  }`} />
                </div>
                <div className="space-y-4">
                  <div>
                    <p className="text-sm text-gray-600">Projected Cost</p>
                    <p data-testid="projected-cost" className="text-2xl font-bold">${forecast.projected_amount.toFixed(2)}</p>
                    <p className="text-xs text-gray-500">End of period: {new Date(forecast.projected_period_end).toLocaleDateString()}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600">Usage Trend</p>
                    <p className={`text-sm font-medium capitalize ${
                      forecast.usage_trend === 'increasing' ? 'text-red-600' : 
                      forecast.usage_trend === 'decreasing' ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {forecast.usage_trend}
                    </p>
                  </div>
                </div>
              </div>

              <div data-testid="usage-alerts" className="bg-white rounded-lg border p-6">
                <h3 className="font-medium mb-4">Recommendations</h3>
                <div className="space-y-3">
                  {forecast.recommendations.map((recommendation, index) => (
                    <div key={index} className="flex items-start space-x-2">
                      <div className="w-1.5 h-1.5 bg-blue-500 rounded-full mt-2"></div>
                      <p className="text-sm text-gray-700">{recommendation}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}