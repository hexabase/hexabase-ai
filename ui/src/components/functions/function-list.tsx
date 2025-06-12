'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { FunctionConfig, apiClient } from '@/lib/api-client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Play,
  Upload,
  FileText,
  Clock,
  Cpu,
  Globe,
  Calendar,
  Zap,
  AlertCircle,
  Filter,
} from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { DeployFunctionDialog } from './deploy-function-dialog';
import { FunctionInvocationDialog } from './function-invocation-dialog';
import { FunctionLogsDialog } from './function-logs-dialog';

export function FunctionList() {
  const router = useRouter();
  const params = useParams();
  const { toast } = useToast();
  const orgId = params?.orgId as string;
  const workspaceId = params?.workspaceId as string;
  const projectId = params?.projectId as string;

  const [functions, setFunctions] = useState<FunctionConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [runtimeFilter, setRuntimeFilter] = useState('all');
  const [selectedFunction, setSelectedFunction] = useState<FunctionConfig | null>(null);
  const [deployDialogOpen, setDeployDialogOpen] = useState(false);
  const [invokeDialogOpen, setInvokeDialogOpen] = useState(false);
  const [logsDialogOpen, setLogsDialogOpen] = useState(false);

  const fetchFunctions = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.functions.list(orgId, workspaceId, projectId, {
        runtime: runtimeFilter === 'all' ? '' : runtimeFilter,
      });
      setFunctions(response.data.functions);
    } catch (error) {
      setError('Failed to load functions');
      console.error('Error fetching functions:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFunctions();
  }, [orgId, workspaceId, projectId, runtimeFilter]);

  const handleFunctionClick = (func: FunctionConfig) => {
    router.push(`/dashboard/organizations/${orgId}/workspaces/${workspaceId}/projects/${projectId}/functions/${func.id}`);
  };

  const handleDeploy = (func: FunctionConfig) => {
    setSelectedFunction(func);
    setDeployDialogOpen(true);
  };

  const handleInvoke = (func: FunctionConfig) => {
    setSelectedFunction(func);
    setInvokeDialogOpen(true);
  };

  const handleViewLogs = (func: FunctionConfig) => {
    setSelectedFunction(func);
    setLogsDialogOpen(true);
  };

  const handleDeploySuccess = () => {
    fetchFunctions();
    toast({
      title: 'Function deployed',
      description: 'New version has been deployed successfully.',
    });
  };

  const getRuntimeIcon = (runtime: string) => {
    if (runtime.includes('node')) return '🟨';
    if (runtime.includes('python')) return '🐍';
    if (runtime.includes('go')) return '🐹';
    if (runtime.includes('java')) return '☕';
    return '📦';
  };

  const getStatusVariant = (status: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case 'active':
        return 'default';
      case 'updating':
        return 'secondary';
      case 'error':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  const getTriggerBadge = (trigger: string) => {
    switch (trigger) {
      case 'http':
        return <Badge variant="outline" className="flex items-center gap-1"><Globe className="h-3 w-3" />http</Badge>;
      case 'event':
        return <Badge variant="outline" className="flex items-center gap-1"><Zap className="h-3 w-3" />event</Badge>;
      case 'schedule':
        return <Badge variant="outline" className="flex items-center gap-1"><Calendar className="h-3 w-3" />schedule</Badge>;
      default:
        return <Badge variant="outline">{trigger}</Badge>;
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-40 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <AlertCircle className="h-8 w-8 text-destructive mb-4" />
        <p className="text-lg font-medium">{error}</p>
        <Button onClick={fetchFunctions} className="mt-4">
          Retry
        </Button>
      </div>
    );
  }

  if (functions.length === 0 && runtimeFilter === 'all') {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <Zap className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium">No functions found</h3>
        <p className="text-muted-foreground mt-2">
          Create a serverless function to run code on demand.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Functions</h2>
        <Select
          value={runtimeFilter}
          onValueChange={setRuntimeFilter}
        >
          <SelectTrigger className="w-40" data-testid="runtime-filter">
            <Filter className="h-4 w-4 mr-2" />
            <SelectValue placeholder="All runtimes" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All runtimes</SelectItem>
            <SelectItem value="nodejs">Node.js</SelectItem>
            <SelectItem value="python">Python</SelectItem>
            <SelectItem value="go">Go</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-4">
        {functions.map((func) => (
          <Card key={func.id} className="hover:shadow-lg transition-shadow">
            <div
              role="button"
              onClick={() => handleFunctionClick(func)}
              className="cursor-pointer"
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{getRuntimeIcon(func.runtime)}</span>
                    <div>
                      <CardTitle className="text-lg">{func.name}</CardTitle>
                      {func.description && (
                        <p className="text-sm text-muted-foreground mt-1">{func.description}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={getStatusVariant(func.status)} className="text-xs">
                      {func.status}
                    </Badge>
                    <Badge variant="secondary">{func.version}</Badge>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div className="flex items-center gap-2">
                    <Cpu className="h-4 w-4 text-muted-foreground" />
                    <span>{func.runtime}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span>{func.memory} MB</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span>{func.timeout}s timeout</span>
                  </div>
                  <div className="flex items-center gap-2">
                    {func.triggers.map((trigger) => getTriggerBadge(trigger))}
                  </div>
                </div>
              </CardContent>
            </div>
            <CardContent className="pt-0 border-t">
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleInvoke(func);
                  }}
                  data-testid={`invoke-${func.id}`}
                >
                  <Play className="h-4 w-4 mr-1" />
                  Invoke
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeploy(func);
                  }}
                  data-testid={`deploy-${func.id}`}
                >
                  <Upload className="h-4 w-4 mr-1" />
                  Deploy
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleViewLogs(func);
                  }}
                  data-testid={`logs-${func.id}`}
                >
                  <FileText className="h-4 w-4 mr-1" />
                  Logs
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {selectedFunction && (
        <>
          <DeployFunctionDialog
            open={deployDialogOpen}
            onClose={() => setDeployDialogOpen(false)}
            onSuccess={handleDeploySuccess}
            functionData={selectedFunction}
            orgId={orgId}
            workspaceId={workspaceId}
          />
          <FunctionInvocationDialog
            open={invokeDialogOpen}
            onClose={() => setInvokeDialogOpen(false)}
            functionData={selectedFunction}
            orgId={orgId}
            workspaceId={workspaceId}
          />
          <FunctionLogsDialog
            open={logsDialogOpen}
            onClose={() => setLogsDialogOpen(false)}
            functionData={selectedFunction}
            orgId={orgId}
            workspaceId={workspaceId}
          />
        </>
      )}
    </div>
  );
}