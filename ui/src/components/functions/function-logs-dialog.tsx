'use client';

import { useState, useEffect } from 'react';
import { FunctionConfig, functionsApi } from '@/lib/api-client';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { FileText, RefreshCw, Download } from 'lucide-react';
import { format } from 'date-fns';

interface FunctionLogsDialogProps {
  open: boolean;
  onClose: () => void;
  functionData: FunctionConfig;
  orgId: string;
  workspaceId: string;
}

export function FunctionLogsDialog({
  open,
  onClose,
  functionData,
  orgId,
  workspaceId,
}: FunctionLogsDialogProps) {
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      fetchLogs();
    }
  }, [open]);

  const fetchLogs = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await functionsApi.getLogs(orgId, workspaceId, functionData.id, {
        limit: 100,
      });
      setLogs(response.data.logs);
    } catch (error) {
      setError('Failed to fetch logs');
      console.error('Error fetching logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = () => {
    const content = logs.join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${functionData.name}-logs-${format(new Date(), 'yyyy-MM-dd-HHmmss')}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Function Logs</DialogTitle>
          <DialogDescription>
            Recent logs for {functionData.name}
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-end gap-2 mb-4">
          <Button size="sm" variant="outline" onClick={fetchLogs}>
            <RefreshCw className="h-4 w-4 mr-1" />
            Refresh
          </Button>
          <Button size="sm" variant="outline" onClick={handleDownload} disabled={logs.length === 0}>
            <Download className="h-4 w-4 mr-1" />
            Download
          </Button>
        </div>

        <div className="flex-1 overflow-hidden">
          {loading ? (
            <div className="space-y-2">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-4 w-full" />
              ))}
            </div>
          ) : error ? (
            <div className="flex flex-col items-center justify-center p-8 text-center">
              <FileText className="h-8 w-8 text-muted-foreground mb-4" />
              <p className="text-sm text-muted-foreground">{error}</p>
              <Button onClick={fetchLogs} className="mt-4" size="sm">
                Retry
              </Button>
            </div>
          ) : logs.length > 0 ? (
            <div className="bg-muted rounded-md p-4 overflow-auto max-h-[50vh]">
              <pre className="text-sm whitespace-pre-wrap font-mono">
                {logs.join('\n')}
              </pre>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center p-8 text-center">
              <FileText className="h-8 w-8 text-muted-foreground mb-4" />
              <p className="text-sm text-muted-foreground">No logs available</p>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-4">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}