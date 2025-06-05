'use client';

import { useEffect, useState } from 'react';
import { Progress } from '@/components/ui/progress';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { RefreshCw, X, CheckCircle, XCircle, Clock, Loader2 } from 'lucide-react';
import { taskApi, type Task } from '@/lib/api-client';
import { useTaskMonitoring } from '@/hooks/use-websocket';
import type { TaskProgressUpdate } from '@/lib/websocket';
import { useToast } from '@/hooks/use-toast';

interface TaskMonitorProps {
  taskId: string;
  organizationId?: string;
  onComplete?: (task: Task) => void;
  onError?: (error: string) => void;
  onCancel?: () => void;
  showActions?: boolean;
}

export function TaskMonitor({ 
  taskId, 
  organizationId,
  onComplete, 
  onError, 
  onCancel,
  showActions = true 
}: TaskMonitorProps) {
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();
  const { onTaskProgress, isConnected } = useTaskMonitoring(taskId, organizationId);

  // Load initial task data
  useEffect(() => {
    const loadTask = async () => {
      try {
        const data = await taskApi.get(taskId);
        setTask(data);
        setLoading(false);
        
        // Check if task is already complete
        if (data.status === 'completed' && onComplete) {
          onComplete(data);
        } else if (data.status === 'failed' && onError) {
          onError(data.error || 'Task failed');
        }
      } catch (error) {
        console.error('Failed to load task:', error);
        setLoading(false);
      }
    };
    
    loadTask();
  }, [taskId, onComplete, onError]);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribe = onTaskProgress((update: TaskProgressUpdate) => {
      if (update.task_id === taskId) {
        setTask(prev => prev ? {
          ...prev,
          status: update.status,
          progress: update.progress,
          message: update.message,
          error: update.error,
          updated_at: update.timestamp,
          completed_at: update.status === 'completed' || update.status === 'failed' ? update.timestamp : prev.completed_at,
        } : null);
        
        // Handle completion
        if (update.status === 'completed' && onComplete && task) {
          onComplete({ ...task, status: 'completed', completed_at: update.timestamp });
        } else if (update.status === 'failed' && onError) {
          onError(update.error || 'Task failed');
        }
      }
    });

    return () => {
      unsubscribe();
    };
  }, [taskId, task, onTaskProgress, onComplete, onError]);

  const handleCancel = async () => {
    try {
      await taskApi.cancel(taskId);
      toast({
        title: 'Task Cancelled',
        description: 'The task has been cancelled.',
      });
      if (onCancel) {
        onCancel();
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to cancel task',
        variant: 'destructive',
      });
    }
  };

  const handleRetry = async () => {
    try {
      const newTask = await taskApi.retry(taskId);
      setTask(newTask);
      toast({
        title: 'Task Retried',
        description: 'The task has been restarted.',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to retry task',
        variant: 'destructive',
      });
    }
  };

  const getStatusIcon = () => {
    switch (task?.status) {
      case 'pending':
        return <Clock className="h-4 w-4" />;
      case 'in_progress':
        return <Loader2 className="h-4 w-4 animate-spin" />;
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />;
      default:
        return null;
    }
  };

  const getStatusVariant = () => {
    switch (task?.status) {
      case 'pending':
        return 'secondary';
      case 'in_progress':
        return 'default';
      case 'completed':
        return 'default';
      case 'failed':
        return 'destructive';
      default:
        return 'outline';
    }
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center space-x-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Loading task...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!task) {
    return (
      <Card>
        <CardContent className="p-6">
          <p className="text-sm text-gray-600">Task not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <CardTitle className="text-lg">{task.type || 'Task'}</CardTitle>
            <Badge variant={getStatusVariant()} className="flex items-center gap-1">
              {getStatusIcon()}
              {task.status}
            </Badge>
          </div>
          {showActions && task.status === 'in_progress' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleCancel}
              disabled={!isConnected}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
          {showActions && task.status === 'failed' && (
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRetry}
              disabled={!isConnected}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
          )}
        </div>
        {task.message && (
          <CardDescription className="mt-2">{task.message}</CardDescription>
        )}
      </CardHeader>
      <CardContent>
        {task.progress !== undefined && task.status === 'in_progress' && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress</span>
              <span>{task.progress}%</span>
            </div>
            <Progress value={task.progress} className="h-2" />
          </div>
        )}
        
        {task.error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-800">{task.error}</p>
          </div>
        )}
        
        {!isConnected && task.status === 'in_progress' && (
          <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
            <p className="text-sm text-yellow-800">
              Real-time updates unavailable. Task is still running in the background.
            </p>
          </div>
        )}
        
        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Started:</span>
            <p className="font-medium">{new Date(task.created_at).toLocaleString()}</p>
          </div>
          {task.completed_at && (
            <div>
              <span className="text-gray-500">Completed:</span>
              <p className="font-medium">{new Date(task.completed_at).toLocaleString()}</p>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}