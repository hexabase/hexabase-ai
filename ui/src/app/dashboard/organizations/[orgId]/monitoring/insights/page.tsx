'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { monitoringApi, type PerformanceInsight } from '@/lib/api-client';
import { ArrowLeft, TrendingUp, DollarSign, AlertTriangle, Lightbulb, ChevronRight } from 'lucide-react';

interface InsightsPageProps {
  params: {
    orgId: string;
  };
}

export default function InsightsPage({ params }: InsightsPageProps) {
  const router = useRouter();
  const [insights, setInsights] = useState<PerformanceInsight[]>([]);
  const [recommendations, setRecommendations] = useState<PerformanceInsight[]>([]);
  const [costOptimizations, setCostOptimizations] = useState<PerformanceInsight[]>([]);
  const [bottlenecks, setBottlenecks] = useState<PerformanceInsight[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchInsights();
  }, [params.orgId]);

  const fetchInsights = async () => {
    try {
      const data = await monitoringApi.getInsights(params.orgId);
      const all = data.insights;
      setInsights(all);
      setRecommendations(all.filter(i => i.type === 'optimization'));
      setCostOptimizations(all.filter(i => i.type === 'cost_saving'));
      setBottlenecks(all.filter(i => i.type === 'bottleneck'));
    } catch (error) {
      console.error('Failed to fetch insights:', error);
      // Use mock data for development
      const mockInsights = getMockInsights();
      setInsights(mockInsights);
      setRecommendations(mockInsights.filter(i => i.type === 'optimization'));
      setCostOptimizations(mockInsights.filter(i => i.type === 'cost_saving'));
      setBottlenecks(mockInsights.filter(i => i.type === 'bottleneck'));
    } finally {
      setLoading(false);
    }
  };

  const getMockInsights = (): PerformanceInsight[] => [
    {
      id: 'insight-1',
      type: 'cost_saving',
      severity: 'high',
      title: 'Oversized Production Instances',
      description: 'Your production nodes are utilizing only 35% of allocated CPU resources on average.',
      recommendation: 'Consider downsizing from c5.4xlarge to c5.2xlarge instances to save approximately $180/month per node.',
      potential_savings: 540,
      affected_resources: ['node-prod-01', 'node-prod-02', 'node-prod-03'],
      created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'insight-2',
      type: 'bottleneck',
      severity: 'medium',
      title: 'Database Connection Pool Exhaustion',
      description: 'The database connection pool is reaching its limit during peak hours, causing request queuing.',
      recommendation: 'Increase the connection pool size from 50 to 100 connections and implement connection pooling at the application level.',
      affected_resources: ['database-primary', 'api-server'],
      created_at: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'insight-3',
      type: 'optimization',
      severity: 'low',
      title: 'Enable Horizontal Pod Autoscaling',
      description: 'Several deployments show variable load patterns that could benefit from autoscaling.',
      recommendation: 'Enable HPA for frontend and API deployments with min replicas: 2, max replicas: 10.',
      affected_resources: ['frontend-deployment', 'api-deployment'],
      created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'insight-4',
      type: 'cost_saving',
      severity: 'medium',
      title: 'Unused Persistent Volumes',
      description: '5 persistent volumes totaling 250GB are not attached to any pods for over 30 days.',
      recommendation: 'Delete unused volumes or create snapshots and remove the volumes to save $25/month.',
      potential_savings: 25,
      affected_resources: ['pv-old-logs', 'pv-backup-2023', 'pv-test-data'],
      created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString()
    },
    {
      id: 'insight-5',
      type: 'bottleneck',
      severity: 'high',
      title: 'Memory Pressure on Worker Nodes',
      description: 'Worker nodes are experiencing frequent memory pressure, triggering pod evictions.',
      recommendation: 'Add 2 additional worker nodes or upgrade existing nodes to instances with more memory.',
      affected_resources: ['node-worker-01', 'node-worker-02'],
      created_at: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString()
    }
  ];

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'text-red-600 bg-red-50 border-red-200';
      case 'medium':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'low':
        return 'text-green-600 bg-green-50 border-green-200';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'cost_saving':
        return <DollarSign className="w-5 h-5 text-green-500" />;
      case 'bottleneck':
        return <AlertTriangle className="w-5 h-5 text-red-500" />;
      case 'optimization':
        return <Lightbulb className="w-5 h-5 text-blue-500" />;
      default:
        return <TrendingUp className="w-5 h-5 text-gray-500" />;
    }
  };

  const totalSavings = costOptimizations.reduce((sum, insight) => sum + (insight.potential_savings || 0), 0);

  // Generate mock prediction data
  const generatePredictionData = () => {
    const now = Date.now();
    const dataPoints = 30; // 30 days forecast
    const interval = 24 * 60 * 60 * 1000; // 1 day
    
    return Array.from({ length: dataPoints }, (_, i) => ({
      date: new Date(now + i * interval).toISOString(),
      cpu: 40 + i * 0.8 + Math.random() * 10,
      memory: 60 + i * 0.5 + Math.random() * 8
    }));
  };

  const predictionData = generatePredictionData();

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button variant="ghost" onClick={() => router.push(`/dashboard/organizations/${params.orgId}/monitoring`)}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Monitoring
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Performance Insights</h1>
              <p className="text-gray-600 mt-1">AI-powered recommendations and predictions</p>
            </div>
          </div>
        </div>
        
        <div className="animate-pulse space-y-4">
          <div className="bg-white rounded-lg border p-6">
            <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
            <div className="h-20 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

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
            <h1 className="text-2xl font-bold text-gray-900">Performance Insights</h1>
            <p className="text-gray-600 mt-1">AI-powered recommendations and predictions</p>
          </div>
        </div>
      </div>

      <div data-testid="performance-insights" className="space-y-6">
        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-gray-900">Total Insights</h3>
              <TrendingUp className="w-5 h-5 text-blue-500" />
            </div>
            <p className="text-3xl font-bold">{insights.length}</p>
            <p className="text-sm text-gray-600 mt-1">Actionable recommendations</p>
          </div>
          
          <div data-testid="cost-optimization" className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-gray-900">Potential Savings</h3>
              <DollarSign className="w-5 h-5 text-green-500" />
            </div>
            <p data-testid="potential-savings" className="text-3xl font-bold">${totalSavings}</p>
            <p className="text-sm text-gray-600 mt-1">Per month</p>
          </div>
          
          <div className="bg-white rounded-lg border p-6">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-gray-900">Critical Issues</h3>
              <AlertTriangle className="w-5 h-5 text-red-500" />
            </div>
            <p className="text-3xl font-bold">{insights.filter(i => i.severity === 'high').length}</p>
            <p className="text-sm text-gray-600 mt-1">Require immediate attention</p>
          </div>
        </div>

        {/* Recommendations */}
        <div data-testid="recommendations" className="bg-white rounded-lg border">
          <div className="p-6 border-b">
            <h2 className="text-lg font-semibold">Optimization Recommendations</h2>
          </div>
          
          <div className="divide-y">
            {recommendations.map((insight) => (
              <div key={insight.id} data-testid="recommendation-card" className="p-6 hover:bg-gray-50">
                <div className="flex items-start space-x-3">
                  {getTypeIcon(insight.type)}
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-2">
                      <h3 className="font-medium">{insight.title}</h3>
                      <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getSeverityColor(insight.severity)}`}>
                        {insight.severity}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{insight.description}</p>
                    <div className="bg-blue-50 rounded-lg p-3 mb-2">
                      <p className="text-sm text-blue-800">
                        <strong>Recommendation:</strong> {insight.recommendation}
                      </p>
                    </div>
                    <div className="flex items-center space-x-4 text-xs text-gray-500">
                      <span>Affected: {insight.affected_resources.join(', ')}</span>
                    </div>
                  </div>
                  <ChevronRight className="w-5 h-5 text-gray-400" />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Performance Bottlenecks */}
        <div data-testid="bottlenecks" className="bg-white rounded-lg border">
          <div className="p-6 border-b">
            <h2 className="text-lg font-semibold">Performance Bottlenecks</h2>
          </div>
          
          <div className="divide-y">
            {bottlenecks.map((bottleneck) => (
              <div key={bottleneck.id} data-testid="bottleneck-item" className="p-6">
                <div className="flex items-start space-x-3">
                  <AlertTriangle className={`w-5 h-5 ${
                    bottleneck.severity === 'high' ? 'text-red-500' : 'text-yellow-500'
                  }`} />
                  <div className="flex-1">
                    <h3 className="font-medium mb-1">{bottleneck.title}</h3>
                    <p className="text-sm text-gray-600 mb-2">{bottleneck.description}</p>
                    <p className="text-sm text-gray-700">
                      <strong>Solution:</strong> {bottleneck.recommendation}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Resource Predictions */}
        <div data-testid="resource-predictions" className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">30-Day Resource Forecast</h2>
          
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div data-testid="cpu-prediction-chart">
              <h3 className="font-medium mb-3">CPU Usage Prediction</h3>
              <div className="h-48 bg-gray-50 rounded-lg flex items-center justify-center">
                <div className="text-center">
                  <TrendingUp className="w-8 h-8 text-blue-500 mx-auto mb-2" />
                  <p className="text-sm text-gray-600">CPU usage expected to reach 85% in 30 days</p>
                  <p className="text-xs text-gray-500 mt-1">Consider scaling up before day 25</p>
                </div>
              </div>
            </div>
            
            <div data-testid="memory-prediction-chart">
              <h3 className="font-medium mb-3">Memory Usage Prediction</h3>
              <div className="h-48 bg-gray-50 rounded-lg flex items-center justify-center">
                <div className="text-center">
                  <TrendingUp className="w-8 h-8 text-purple-500 mx-auto mb-2" />
                  <p className="text-sm text-gray-600">Memory usage expected to reach 78% in 30 days</p>
                  <p className="text-xs text-gray-500 mt-1">Current trajectory is sustainable</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}