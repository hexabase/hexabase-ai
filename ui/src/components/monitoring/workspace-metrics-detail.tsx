'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { AlertCircle, Cpu, HardDrive, MemoryStick, Activity, RefreshCw } from 'lucide-react';
import { apiClient } from '@/lib/api-client';

interface MetricData {
  cpu_usage: number;
  memory_usage: number;
  storage_usage: number;
  network_ingress: number;
  network_egress: number;
  pod_count: number;
  container_count: number;
  timestamp: string;
}

interface HistoryData {
  cpu: Array<{ timestamp: string; value: number }>;
  memory: Array<{ timestamp: string; value: number }>;
}

interface WorkspaceMetricsProps {
  workspaceId: string;
  showHistory?: boolean;
}

export function WorkspaceMetrics({ workspaceId, showHistory = false }: WorkspaceMetricsProps) {
  const [metrics, setMetrics] = useState<MetricData | null>(null);
  const [history, setHistory] = useState<HistoryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timePeriod, setTimePeriod] = useState('1h');
  const [refreshInterval, setRefreshInterval] = useState<NodeJS.Timeout | null>(null);

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const fetchMetrics = async () => {
    try {
      setError(null);
      const response = await apiClient.monitoring.getWorkspaceMetrics(workspaceId);
      setMetrics(response.metrics);
      
      if (showHistory) {
        const historyResponse = await apiClient.monitoring.getResourceUsageHistory(
          workspaceId,
          { period: timePeriod }
        );
        setHistory(historyResponse.history);
      }
    } catch (err) {
      setError('Failed to load metrics');
      console.error('Error fetching metrics:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();

    // Set up periodic refresh every 30 seconds
    const interval = setInterval(fetchMetrics, 30000);
    setRefreshInterval(interval);

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [workspaceId, timePeriod, showHistory]);

  const handleRetry = () => {
    setLoading(true);
    fetchMetrics();
  };

  const handleTimePeriodChange = (value: string) => {
    setTimePeriod(value);
  };

  const isCritical = (value: number, threshold: number = 90) => value >= threshold;

  if (loading) {
    return (
      <div data-testid="metrics-skeleton" className="space-y-4">
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardContent className="flex flex-col items-center justify-center py-8">
          <AlertCircle className="h-12 w-12 text-destructive mb-4" />
          <p className="text-sm text-muted-foreground mb-4">{error}</p>
          <Button onClick={handleRetry} size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!metrics) return null;

  const showCriticalAlert = isCritical(metrics.cpu_usage) || isCritical(metrics.memory_usage);

  return (
    <div className="space-y-4">
      {showCriticalAlert && (
        <Card className="border-destructive bg-destructive/10">
          <CardHeader>
            <CardTitle className="text-destructive flex items-center gap-2">
              <AlertCircle className="h-5 w-5" />
              Critical Resource Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {isCritical(metrics.cpu_usage) && (
                <p className="text-sm">CPU usage is at {metrics.cpu_usage}%</p>
              )}
              {isCritical(metrics.memory_usage) && (
                <p className="text-sm">Memory usage is at {metrics.memory_usage}%</p>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">CPU Usage</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.cpu_usage}%</div>
            <div className="mt-2">
              <div className="w-full bg-secondary rounded-full h-2">
                <div
                  className={`h-2 rounded-full ${
                    isCritical(metrics.cpu_usage) ? 'bg-destructive' : 'bg-primary'
                  }`}
                  style={{ width: `${metrics.cpu_usage}%` }}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Memory Usage</CardTitle>
            <MemoryStick className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.memory_usage}%</div>
            <div className="mt-2">
              <div className="w-full bg-secondary rounded-full h-2">
                <div
                  className={`h-2 rounded-full ${
                    isCritical(metrics.memory_usage) ? 'bg-destructive' : 'bg-primary'
                  }`}
                  style={{ width: `${metrics.memory_usage}%` }}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Storage Usage</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.storage_usage}%</div>
            <div className="mt-2">
              <div className="w-full bg-secondary rounded-full h-2">
                <div
                  className="h-2 rounded-full bg-primary"
                  style={{ width: `${metrics.storage_usage}%` }}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Network</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Ingress:</span>
                <span className="font-medium">{formatBytes(metrics.network_ingress)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Egress:</span>
                <span className="font-medium">{formatBytes(metrics.network_egress)}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Resources</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="space-y-1">
              <div className="text-sm">{metrics.pod_count} pods</div>
              <div className="text-sm">{metrics.container_count} containers</div>
            </div>
          </CardContent>
        </Card>
      </div>

      {showHistory && history && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Resource Usage History</h3>
            <div className="flex items-center gap-2">
              <Label htmlFor="time-period" className="text-sm">Time Period</Label>
              <Select value={timePeriod} onValueChange={handleTimePeriodChange}>
                <SelectTrigger id="time-period" className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1h">Last 1 hour</SelectItem>
                  <SelectItem value="6h">Last 6 hours</SelectItem>
                  <SelectItem value="24h">Last 24 hours</SelectItem>
                  <SelectItem value="7d">Last 7 days</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">CPU History</CardTitle>
              </CardHeader>
              <CardContent>
                <div data-testid="cpu-history-chart" className="h-48 bg-muted rounded flex items-center justify-center">
                  <span className="text-muted-foreground text-sm">CPU usage chart</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Memory History</CardTitle>
              </CardHeader>
              <CardContent>
                <div data-testid="memory-history-chart" className="h-48 bg-muted rounded flex items-center justify-center">
                  <span className="text-muted-foreground text-sm">Memory usage chart</span>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      <div className="flex justify-end">
        <Button onClick={fetchMetrics} size="sm" variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>
    </div>
  );
}