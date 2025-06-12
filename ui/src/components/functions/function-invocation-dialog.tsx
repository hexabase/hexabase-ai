'use client';

import { useState } from 'react';
import { FunctionConfig, FunctionInvocation, functionsApi } from '@/lib/api-client';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Play,
  CheckCircle,
  XCircle,
  Clock,
  Code,
  FileText,
  AlertCircle,
  RefreshCw,
  Globe,
  Zap,
  Calendar,
} from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { cn } from '@/lib/utils';

interface FunctionInvocationDialogProps {
  open: boolean;
  onClose: () => void;
  functionData: FunctionConfig;
  orgId: string;
  workspaceId: string;
}

export function FunctionInvocationDialog({
  open,
  onClose,
  functionData,
  orgId,
  workspaceId,
}: FunctionInvocationDialogProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [triggerType, setTriggerType] = useState(functionData.triggers[0] || 'http');
  const [payload, setPayload] = useState('{}');
  const [httpMethod, setHttpMethod] = useState('POST');
  const [headers, setHeaders] = useState('{}');
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<FunctionInvocation | null>(null);

  const handleInvoke = async () => {
    setError(null);
    setResult(null);

    // Validate JSON payload
    try {
      if (payload) {
        JSON.parse(payload);
      }
      if (triggerType === 'http' && headers) {
        JSON.parse(headers);
      }
    } catch (e) {
      setError('Invalid JSON format');
      return;
    }

    try {
      setLoading(true);
      const invocationData: any = {
        trigger_type: triggerType,
        payload: payload ? JSON.parse(payload) : undefined,
      };

      if (triggerType === 'http') {
        invocationData.http_method = httpMethod;
        invocationData.headers = headers ? JSON.parse(headers) : undefined;
      }

      const response = await functionsApi.invoke(
        orgId,
        workspaceId,
        functionData.id,
        invocationData
      );

      setResult(response.data);
    } catch (error) {
      setError('Failed to invoke function');
      console.error('Invocation error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleInvokeAgain = () => {
    setResult(null);
    setError(null);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-4 w-4" />;
      case 'error':
        return <XCircle className="h-4 w-4" />;
      case 'timeout':
        return <Clock className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const getStatusVariant = (status: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case 'success':
        return 'default';
      case 'error':
        return 'destructive';
      case 'timeout':
        return 'secondary';
      default:
        return 'secondary';
    }
  };

  const getTriggerIcon = (trigger: string) => {
    switch (trigger) {
      case 'http':
        return <Globe className="h-4 w-4" />;
      case 'event':
        return <Zap className="h-4 w-4" />;
      case 'schedule':
        return <Calendar className="h-4 w-4" />;
      default:
        return null;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Invoke Function</DialogTitle>
          <DialogDescription>
            Test {functionData.name} with custom payload
          </DialogDescription>
        </DialogHeader>

        {!result ? (
          <div className="space-y-4">
            <div>
              <Label htmlFor="trigger-type">Trigger Type</Label>
              <Select
                value={triggerType}
                onValueChange={setTriggerType}
              >
                <SelectTrigger id="trigger-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {functionData.triggers.map((trigger) => (
                    <SelectItem key={trigger} value={trigger}>
                      <div className="flex items-center gap-2">
                        {getTriggerIcon(trigger)}
                        <span className="capitalize">{trigger}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {triggerType === 'http' && (
              <>
                <div>
                  <Label htmlFor="http-method">HTTP Method</Label>
                  <Select
                    value={httpMethod}
                    onValueChange={setHttpMethod}
                  >
                    <SelectTrigger id="http-method">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="GET">GET</SelectItem>
                      <SelectItem value="POST">POST</SelectItem>
                      <SelectItem value="PUT">PUT</SelectItem>
                      <SelectItem value="DELETE">DELETE</SelectItem>
                      <SelectItem value="PATCH">PATCH</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label htmlFor="headers">Headers (JSON)</Label>
                  <Textarea
                    id="headers"
                    value={headers}
                    onChange={(e) => setHeaders(e.target.value)}
                    placeholder='{"Content-Type": "application/json"}'
                    className="font-mono text-sm"
                    rows={3}
                  />
                </div>
              </>
            )}

            <div>
              <Label htmlFor="payload">Payload (JSON)</Label>
              <Textarea
                id="payload"
                value={payload}
                onChange={(e) => setPayload(e.target.value)}
                placeholder='{"key": "value"}'
                className="font-mono text-sm"
                rows={8}
              />
            </div>

            {error && (
              <div className="flex items-center gap-2 p-3 bg-destructive/10 text-destructive rounded-md">
                <AlertCircle className="h-4 w-4" />
                <p className="text-sm">{error}</p>
              </div>
            )}

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={onClose}>
                Close
              </Button>
              <Button onClick={handleInvoke} disabled={loading}>
                {loading ? (
                  <>
                    <Play className="h-4 w-4 mr-2 animate-spin" />
                    Invoking...
                  </>
                ) : (
                  <>
                    <Play className="h-4 w-4 mr-2" />
                    Invoke
                  </>
                )}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="p-4 bg-muted rounded-md">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-medium">Invocation Result</h3>
                <Badge variant={getStatusVariant(result.status)} className="flex items-center gap-1">
                  {getStatusIcon(result.status)}
                  {result.status}
                </Badge>
              </div>
              <div className="space-y-1 text-sm">
                <p>Invocation ID: <code className="text-xs">{result.invocation_id}</code></p>
                {result.duration_ms && (
                  <p>Duration: {result.duration_ms}ms</p>
                )}
              </div>
            </div>

            <Tabs defaultValue="output" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="output">
                  <Code className="h-4 w-4 mr-2" />
                  Output
                </TabsTrigger>
                <TabsTrigger value="logs">
                  <FileText className="h-4 w-4 mr-2" />
                  Logs
                </TabsTrigger>
              </TabsList>
              <TabsContent value="output" className="mt-4">
                <div className="p-4 bg-muted rounded-md">
                  {result.error ? (
                    <div className="text-destructive">
                      <p className="font-medium mb-2">Error:</p>
                      <pre className="text-sm whitespace-pre-wrap">{result.error}</pre>
                    </div>
                  ) : (
                    <pre className="text-sm whitespace-pre-wrap">
                      {JSON.stringify(result.output, null, 2)}
                    </pre>
                  )}
                </div>
              </TabsContent>
              <TabsContent value="logs" className="mt-4">
                <div className="p-4 bg-muted rounded-md">
                  <h4 className="font-medium mb-2">Function Logs</h4>
                  <pre className="text-sm whitespace-pre-wrap text-muted-foreground">
                    {result.logs || 'No logs available'}
                  </pre>
                </div>
              </TabsContent>
            </Tabs>

            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={onClose}>
                Close
              </Button>
              <Button onClick={handleInvokeAgain}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Invoke Again
              </Button>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}