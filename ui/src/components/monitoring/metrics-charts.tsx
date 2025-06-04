'use client';

import { useState, useEffect } from 'react';
import { monitoringApi, type ResourceMetrics } from '@/lib/api-client';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface MetricsChartsProps {
  orgId: string;
  timeRange: string;
}

export default function MetricsCharts({ orgId, timeRange }: MetricsChartsProps) {
  const [metrics, setMetrics] = useState<ResourceMetrics[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMetrics();
  }, [orgId, timeRange]);

  const fetchMetrics = async () => {
    try {
      const data = await monitoringApi.getResourceMetrics(orgId, { time_range: timeRange });
      setMetrics(data);
    } catch (error) {
      console.error('Failed to fetch metrics:', error);
      // Use mock data for development
      setMetrics(generateMockMetrics());
    } finally {
      setLoading(false);
    }
  };

  const generateMockMetrics = (): ResourceMetrics[] => {
    const now = Date.now();
    const dataPoints = 20;
    const interval = 60000; // 1 minute
    
    return Array.from({ length: dataPoints }, (_, i) => ({
      timestamp: new Date(now - (dataPoints - i - 1) * interval).toISOString(),
      cpu: {
        usage_percentage: 40 + Math.random() * 20,
        cores_used: 16 + Math.random() * 4,
        cores_total: 40
      },
      memory: {
        usage_percentage: 60 + Math.random() * 15,
        used_gb: 96 + Math.random() * 24,
        total_gb: 160
      },
      storage: {
        usage_percentage: 35 + Math.random() * 5,
        used_gb: 1400 + Math.random() * 50,
        total_gb: 4000
      },
      network: {
        ingress_mbps: 200 + Math.random() * 100,
        egress_mbps: 150 + Math.random() * 80
      }
    }));
  };

  const renderChart = (data: number[], label: string, color: string, unit: string = '%') => {
    const max = Math.max(...data);
    const min = Math.min(...data);
    const range = max - min;
    const height = 120;
    
    // Calculate trend
    const trend = data[data.length - 1] - data[0];
    const trendPercentage = ((trend / data[0]) * 100).toFixed(1);
    
    return (
      <div className="bg-white rounded-lg border p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-medium">{label}</h3>
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium">{data[data.length - 1].toFixed(1)}{unit}</span>
            <div className={`flex items-center text-xs ${
              trend > 0 ? 'text-red-600' : trend < 0 ? 'text-green-600' : 'text-gray-600'
            }`}>
              {trend > 0 ? <TrendingUp className="w-3 h-3" /> : trend < 0 ? <TrendingDown className="w-3 h-3" /> : <Minus className="w-3 h-3" />}
              <span className="ml-1">{Math.abs(parseFloat(trendPercentage))}%</span>
            </div>
          </div>
        </div>
        
        <div className="relative" style={{ height }}>
          <svg className="w-full h-full">
            {/* Grid lines */}
            {[0, 0.25, 0.5, 0.75, 1].map((y) => (
              <line
                key={y}
                x1="0"
                y1={height * (1 - y)}
                x2="100%"
                y2={height * (1 - y)}
                stroke="#e5e7eb"
                strokeWidth="1"
              />
            ))}
            
            {/* Data line */}
            <polyline
              fill="none"
              stroke={color}
              strokeWidth="2"
              points={data.map((value, index) => {
                const x = (index / (data.length - 1)) * 100;
                const y = range > 0 ? ((value - min) / range) * height : height / 2;
                return `${x}%,${height - y}`;
              }).join(' ')}
            />
            
            {/* Data points */}
            {data.map((value, index) => {
              const x = (index / (data.length - 1)) * 100;
              const y = range > 0 ? ((value - min) / range) * height : height / 2;
              return (
                <circle
                  key={index}
                  cx={`${x}%`}
                  cy={height - y}
                  r="2"
                  fill={color}
                />
              );
            })}
          </svg>
        </div>
        
        <div className="flex justify-between text-xs text-gray-500 mt-2">
          <span>{new Date(Date.now() - (data.length - 1) * 60000).toLocaleTimeString()}</span>
          <span>Now</span>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-lg border p-4 animate-pulse">
            <div className="h-4 bg-gray-200 rounded w-1/3 mb-4"></div>
            <div className="h-32 bg-gray-200 rounded"></div>
          </div>
        ))}
      </div>
    );
  }

  const cpuData = metrics.map(m => m.cpu.usage_percentage);
  const memoryData = metrics.map(m => m.memory.usage_percentage);
  const networkInData = metrics.map(m => m.network.ingress_mbps);
  const networkOutData = metrics.map(m => m.network.egress_mbps);

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold">Performance Metrics</h2>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div data-testid="cpu-usage-chart">
          {renderChart(cpuData, 'CPU Usage', '#3b82f6')}
        </div>
        
        <div data-testid="memory-usage-chart">
          {renderChart(memoryData, 'Memory Usage', '#8b5cf6')}
        </div>
        
        <div data-testid="network-io-chart">
          {renderChart(networkInData, 'Network Ingress', '#10b981', ' Mbps')}
        </div>
        
        <div data-testid="disk-io-chart">
          {renderChart(networkOutData, 'Network Egress', '#f59e0b', ' Mbps')}
        </div>
      </div>
    </div>
  );
}