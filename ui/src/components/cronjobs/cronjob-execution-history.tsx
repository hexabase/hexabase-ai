'use client';

import { useState } from 'react';
import { CronJobExecution } from '@/lib/api-client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  CheckCircle,
  XCircle,
  Loader2,
  Clock,
  FileText,
  ChevronRight,
  ChevronDown,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { format, formatDistanceToNow } from 'date-fns';

interface CronJobExecutionHistoryProps {
  executions: CronJobExecution[];
  loading: boolean;
  onViewLogs?: (execution: CronJobExecution) => void;
  totalExecutions?: number;
  currentPage?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
}

export function CronJobExecutionHistory({
  executions,
  loading,
  onViewLogs,
  totalExecutions,
  currentPage = 1,
  pageSize = 10,
  onPageChange,
}: CronJobExecutionHistoryProps) {
  const [expandedExecutions, setExpandedExecutions] = useState<Set<string>>(new Set());

  const getStatusVariant = (status: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case 'succeeded':
        return 'default';
      case 'failed':
        return 'destructive';
      case 'running':
        return 'secondary';
      case 'cancelled':
        return 'outline';
      default:
        return 'secondary';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'succeeded':
        return <CheckCircle className="h-3 w-3" />;
      case 'failed':
        return <XCircle className="h-3 w-3" />;
      case 'running':
        return <Loader2 className="h-3 w-3 animate-spin" />;
      case 'cancelled':
        return <XCircle className="h-3 w-3" />;
      default:
        return <Clock className="h-3 w-3" />;
    }
  };

  const calculateDuration = (startedAt: string, completedAt?: string) => {
    if (!completedAt) return null;
    const start = new Date(startedAt).getTime();
    const end = new Date(completedAt).getTime();
    const durationMs = end - start;
    const minutes = Math.floor(durationMs / 60000);
    const seconds = Math.floor((durationMs % 60000) / 1000);
    
    if (minutes > 0) {
      return `${minutes} minutes${seconds > 0 ? ` ${seconds}s` : ''}`;
    }
    return `${seconds} seconds`;
  };

  const toggleExpanded = (executionId: string) => {
    const newExpanded = new Set(expandedExecutions);
    if (newExpanded.has(executionId)) {
      newExpanded.delete(executionId);
    } else {
      newExpanded.add(executionId);
    }
    setExpandedExecutions(newExpanded);
  };

  if (loading) {
    return (
      <div className="space-y-2" data-testid="executions-skeleton">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-20 w-full" />
        ))}
      </div>
    );
  }

  if (executions.length === 0) {
    return (
      <Card>
        <CardContent className="text-center py-6">
          <Clock className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
          <p className="text-muted-foreground">No executions yet</p>
        </CardContent>
      </Card>
    );
  }

  const totalPages = totalExecutions ? Math.ceil(totalExecutions / pageSize) : 1;

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        {executions.map((execution) => {
          const isExpanded = expandedExecutions.has(execution.id);
          const duration = calculateDuration(execution.started_at, execution.completed_at);

          return (
            <Card key={execution.id} className="overflow-hidden">
              <div
                className="cursor-pointer"
                onClick={() => toggleExpanded(execution.id)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {isExpanded ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                      <CardTitle className="text-base">{execution.job_name}</CardTitle>
                      <Badge
                        variant={getStatusVariant(execution.status)}
                        className="flex items-center gap-1"
                      >
                        {getStatusIcon(execution.status)}
                        {execution.status}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      {execution.exit_code !== undefined && (
                        <span>Exit code: {execution.exit_code}</span>
                      )}
                      {duration && <span>{duration}</span>}
                      <span>{format(new Date(execution.started_at), 'MMM d, yyyy HH:mm')}</span>
                    </div>
                  </div>
                </CardHeader>
              </div>

              {isExpanded && (
                <CardContent className="pt-0">
                  <div className="space-y-3">
                    <div className="text-sm">
                      <div className="flex items-center gap-4 text-muted-foreground">
                        <span>
                          Started: {format(new Date(execution.started_at), 'MMM d, yyyy HH:mm:ss')}
                        </span>
                        {execution.completed_at && (
                          <span>
                            Completed: {format(new Date(execution.completed_at), 'MMM d, yyyy HH:mm:ss')}
                          </span>
                        )}
                      </div>
                    </div>

                    {execution.logs && (
                      <div className="space-y-2">
                        <div className="bg-muted rounded-md p-3">
                          <pre className="text-xs whitespace-pre-wrap">{execution.logs}</pre>
                        </div>
                        {onViewLogs && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={(e) => {
                              e.stopPropagation();
                              onViewLogs(execution);
                            }}
                          >
                            <FileText className="h-4 w-4 mr-1" />
                            View Logs
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                </CardContent>
              )}
            </Card>
          );
        })}
      </div>

      {onPageChange && totalPages > 1 && (
        <div className="flex justify-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage === 1}
          >
            Previous
          </Button>
          <span className="flex items-center px-3 text-sm">
            Page {currentPage} of {totalPages}
          </span>
          <Button
            size="sm"
            variant="outline"
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}