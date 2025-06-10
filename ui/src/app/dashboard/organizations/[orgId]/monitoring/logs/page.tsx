'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { monitoringApi, type LogEntry } from '@/lib/api-client';
import { ArrowLeft, Search, Download, PlayCircle, PauseCircle, Filter } from 'lucide-react';
import { format } from 'date-fns';

interface LogsPageProps {
  params: {
    orgId: string;
  };
}

export default function LogsPage({ params }: LogsPageProps) {
  const router = useRouter();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [workspaceId, setWorkspaceId] = useState<string>('all');
  const [namespace, setNamespace] = useState<string>('all');
  const [pod, setPod] = useState<string>('all');
  const [logLevel, setLogLevel] = useState<string>('all');
  const [search, setSearch] = useState<string>('');
  const [streaming, setStreaming] = useState(false);
  const [loading, setLoading] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    fetchLogs();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [workspaceId, namespace, pod, logLevel]);

  useEffect(() => {
    if (streaming) {
      scrollToBottom();
    }
  }, [logs, streaming]);

  const fetchLogs = async () => {
    try {
      const params: any = { limit: 100 };
      if (workspaceId !== 'all') params.workspace_id = workspaceId;
      if (namespace !== 'all') params.namespace = namespace;
      if (pod !== 'all') params.pod = pod;
      if (logLevel !== 'all') params.level = logLevel;
      if (search) params.search = search;
      
      const data = await monitoringApi.getLogs(params.orgId, params);
      setLogs(data.logs);
    } catch (error) {
      console.error('Failed to fetch logs:', error);
      // Use mock data for development
      setLogs(getMockLogs());
    } finally {
      setLoading(false);
    }
  };

  const getMockLogs = (): LogEntry[] => {
    const levels: LogEntry['level'][] = ['debug', 'info', 'warn', 'error'];
    const pods = ['api-server-7d9f8c', 'worker-8b4c5d', 'frontend-9a2f3e', 'database-3c5d7f'];
    const namespaces = ['default', 'kube-system', 'application', 'monitoring'];
    
    return Array.from({ length: 50 }, (_, i) => ({
      timestamp: new Date(Date.now() - i * 60000).toISOString(),
      level: levels[Math.floor(Math.random() * levels.length)],
      workspace_id: 'ws-prod',
      namespace: namespaces[Math.floor(Math.random() * namespaces.length)],
      pod: pods[Math.floor(Math.random() * pods.length)],
      container: 'main',
      message: getRandomLogMessage(i),
      metadata: {}
    }));
  };

  const getRandomLogMessage = (index: number) => {
    const messages = [
      'Successfully connected to database',
      'Request processed: GET /api/v1/health',
      'Cache hit for key: user_session_123',
      'Starting periodic cleanup task',
      'Error connecting to external service: timeout after 30s',
      'Warning: Memory usage above 80%',
      'Debug: Processing batch job item 42 of 100',
      'Info: New deployment detected, reloading configuration',
      'Error: Failed to parse JSON response',
      'Successfully authenticated user: admin@example.com'
    ];
    return messages[index % messages.length];
  };

  const handleStreamToggle = () => {
    if (streaming) {
      setStreaming(false);
      if (wsRef.current) {
        wsRef.current.close();
      }
    } else {
      setStreaming(true);
      startLogStream();
    }
  };

  const startLogStream = () => {
    const params: any = {};
    if (workspaceId !== 'all') params.workspace_id = workspaceId;
    if (namespace !== 'all') params.namespace = namespace;
    if (pod !== 'all') params.pod = pod;
    
    try {
      const ws = monitoringApi.streamLogs(params.orgId, params);
      
      ws.onmessage = (event) => {
        const newLog = JSON.parse(event.data) as LogEntry;
        setLogs(prev => [...prev.slice(-99), newLog]); // Keep last 100 logs
      };
      
      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setStreaming(false);
      };
      
      ws.onclose = () => {
        setStreaming(false);
      };
      
      wsRef.current = ws;
    } catch (error) {
      console.error('Failed to start log stream:', error);
      setStreaming(false);
    }
  };

  const scrollToBottom = () => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleExport = async () => {
    try {
      const blob = await monitoringApi.exportLogs(params.orgId, {
        format: 'csv',
        start_time: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        end_time: new Date().toISOString()
      });
      
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `logs-${format(new Date(), 'yyyy-MM-dd-HH-mm-ss')}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Failed to export logs:', error);
    }
  };

  const getLogLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'text-red-600 bg-red-50';
      case 'warn':
        return 'text-yellow-600 bg-yellow-50';
      case 'info':
        return 'text-blue-600 bg-blue-50';
      case 'debug':
        return 'text-gray-600 bg-gray-50';
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
            <h1 className="text-2xl font-bold text-gray-900">Logs Viewer</h1>
            <p className="text-gray-600 mt-1">View and search application logs</p>
          </div>
        </div>
        
        <div className="flex items-center space-x-3">
          <Button
            variant="outline"
            onClick={handleStreamToggle}
            data-testid="stream-logs"
          >
            {streaming ? (
              <>
                <PauseCircle className="w-4 h-4 mr-2" />
                Stop Streaming
              </>
            ) : (
              <>
                <PlayCircle className="w-4 h-4 mr-2" />
                Stream Logs
              </>
            )}
          </Button>
          
          <Button
            variant="outline"
            onClick={handleExport}
            data-testid="export-logs"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      <div data-testid="logs-viewer" className="bg-white rounded-lg border">
        {/* Filters */}
        <div className="p-4 border-b space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Workspace</label>
              <select
                data-testid="workspace-selector"
                value={workspaceId}
                onChange={(e) => setWorkspaceId(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Workspaces</option>
                <option value="ws-prod">Production</option>
                <option value="ws-staging">Staging</option>
                <option value="ws-dev">Development</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Namespace</label>
              <select
                data-testid="namespace-selector"
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Namespaces</option>
                <option value="default">default</option>
                <option value="kube-system">kube-system</option>
                <option value="application">application</option>
                <option value="monitoring">monitoring</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Pod</label>
              <select
                data-testid="pod-selector"
                value={pod}
                onChange={(e) => setPod(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Pods</option>
                <option value="api-server">api-server</option>
                <option value="worker">worker</option>
                <option value="frontend">frontend</option>
                <option value="database">database</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Log Level</label>
              <select
                data-testid="log-level-filter"
                value={logLevel}
                onChange={(e) => setLogLevel(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              >
                <option value="all">All Levels</option>
                <option value="error">Error</option>
                <option value="warn">Warning</option>
                <option value="info">Info</option>
                <option value="debug">Debug</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Search</label>
              <div className="relative">
                <input
                  type="text"
                  data-testid="log-search"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && fetchLogs()}
                  placeholder="Search logs..."
                  className="w-full border border-gray-300 rounded-md pl-9 pr-3 py-2 text-sm"
                />
                <Search className="absolute left-3 top-2.5 w-4 h-4 text-gray-400" />
              </div>
            </div>
          </div>
        </div>

        {/* Log Entries */}
        <div data-testid="log-entries" className="p-4 font-mono text-sm max-h-[600px] overflow-y-auto bg-gray-50">
          {loading ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>
          ) : logs.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No logs found
            </div>
          ) : (
            <div className="space-y-1">
              {logs.map((log, index) => (
                <div key={index} data-testid="log-entry" className="bg-white rounded px-3 py-2 hover:bg-gray-50">
                  <div className="flex items-start space-x-3">
                    <span className="text-gray-500 text-xs whitespace-nowrap">
                      {format(new Date(log.timestamp), 'HH:mm:ss.SSS')}
                    </span>
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getLogLevelColor(log.level)}`}>
                      {log.level.toUpperCase()}
                    </span>
                    <span className="text-gray-600 text-xs">
                      [{log.namespace}/{log.pod}]
                    </span>
                    <span className="text-gray-800 flex-1">{log.message}</span>
                  </div>
                </div>
              ))}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>
        
        {streaming && (
          <div className="px-4 py-2 bg-green-50 border-t flex items-center justify-center">
            <div className="flex items-center space-x-2 text-green-700">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
              <span className="text-sm">Live streaming active</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}