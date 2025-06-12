'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Application, CronJobExecution, apiClient } from '@/lib/api-client';
import { CronJobExecutionHistory } from './cronjob-execution-history';
import { CronJobScheduleEditor } from './cronjob-schedule-editor';
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
  Pause,
  Clock,
  History,
  Edit,
  AlertCircle,
  CheckCircle,
  Filter,
} from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { format, formatDistanceToNow } from 'date-fns';
import { cn } from '@/lib/utils';

export function CronJobList() {
  const router = useRouter();
  const params = useParams();
  const { toast } = useToast();
  const orgId = params?.orgId as string;
  const workspaceId = params?.workspaceId as string;
  const projectId = params?.projectId as string;

  const [cronJobs, setCronJobs] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState('all');
  const [editingSchedule, setEditingSchedule] = useState<Application | null>(null);
  const [viewingHistory, setViewingHistory] = useState<Application | null>(null);
  const [executions, setExecutions] = useState<CronJobExecution[]>([]);
  const [executionsLoading, setExecutionsLoading] = useState(false);

  const fetchCronJobs = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.applications.list(orgId, workspaceId, projectId, {
        type: 'cronjob',
        status: statusFilter === 'all' ? '' : statusFilter,
      });
      setCronJobs(response.data.applications);
    } catch (error) {
      setError('Failed to load cronjobs');
      console.error('Error fetching cronjobs:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchExecutions = async (appId: string) => {
    try {
      setExecutionsLoading(true);
      const response = await apiClient.applications.getCronJobExecutions(orgId, workspaceId, appId);
      setExecutions(response.data.executions);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load execution history',
        variant: 'destructive',
      });
    } finally {
      setExecutionsLoading(false);
    }
  };

  useEffect(() => {
    fetchCronJobs();
  }, [orgId, workspaceId, projectId, statusFilter]);

  const handleTrigger = async (cronJob: Application) => {
    try {
      const response = await apiClient.applications.triggerCronJob(orgId, workspaceId, cronJob.id);
      toast({
        title: 'CronJob triggered',
        description: `${cronJob.name} has been triggered successfully.`,
      });
      fetchCronJobs(); // Refresh to get updated last_execution_at
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to trigger cronjob',
        variant: 'destructive',
      });
    }
  };

  const handleToggleStatus = async (cronJob: Application) => {
    const newStatus = cronJob.status === 'active' ? 'suspended' : 'active';
    try {
      await apiClient.applications.updateStatus(orgId, workspaceId, cronJob.id, {
        status: newStatus,
      });
      setCronJobs(cronJobs.map(cj => 
        cj.id === cronJob.id ? { ...cj, status: newStatus } : cj
      ));
      toast({
        title: 'Status updated',
        description: `${cronJob.name} is now ${newStatus}.`,
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update status',
        variant: 'destructive',
      });
    }
  };

  const handleScheduleUpdate = async (schedule: string) => {
    if (!editingSchedule) return;

    try {
      const response = await apiClient.applications.updateCronSchedule(
        orgId,
        workspaceId,
        editingSchedule.id,
        { schedule }
      );
      setCronJobs(cronJobs.map(cj => 
        cj.id === editingSchedule.id 
          ? { ...cj, cron_schedule: schedule, next_execution_at: response.data.next_execution_at }
          : cj
      ));
      setEditingSchedule(null);
      toast({
        title: 'Schedule updated',
        description: 'CronJob schedule has been updated successfully.',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update schedule',
        variant: 'destructive',
      });
    }
  };

  const handleViewHistory = async (cronJob: Application) => {
    setViewingHistory(cronJob);
    await fetchExecutions(cronJob.id);
  };

  if (loading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-32 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <AlertCircle className="h-8 w-8 text-destructive mb-4" />
        <p className="text-lg font-medium">{error}</p>
        <Button onClick={fetchCronJobs} className="mt-4">
          Retry
        </Button>
      </div>
    );
  }

  if (cronJobs.length === 0 && statusFilter === 'all') {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center">
        <Clock className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium">No CronJobs found</h3>
        <p className="text-muted-foreground mt-2">
          Create a CronJob to run scheduled tasks.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">CronJobs</h2>
        <Select
          value={statusFilter}
          onValueChange={setStatusFilter}
        >
          <SelectTrigger className="w-40" data-testid="status-filter">
            <Filter className="h-4 w-4 mr-2" />
            <SelectValue placeholder="All statuses" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="suspended">Suspended</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-4">
        {cronJobs.map((cronJob) => (
          <Card key={cronJob.id}>
            <CardHeader>
              <div className="flex justify-between items-start">
                <div>
                  <CardTitle className="text-lg">{cronJob.name}</CardTitle>
                  <div className="flex items-center gap-2 mt-1">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <code className="text-sm bg-muted px-2 py-0.5 rounded">
                      {cronJob.cron_schedule}
                    </code>
                  </div>
                </div>
                <Badge
                  variant={cronJob.status === 'active' ? 'default' : 'secondary'}
                  className="flex items-center gap-1"
                >
                  {cronJob.status === 'active' ? (
                    <CheckCircle className="h-3 w-3" />
                  ) : (
                    <Pause className="h-3 w-3" />
                  )}
                  {cronJob.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="text-sm text-muted-foreground">
                  {cronJob.last_execution_at && (
                    <p>
                      Last run: {formatDistanceToNow(new Date(cronJob.last_execution_at), {
                        addSuffix: true,
                      })}
                    </p>
                  )}
                  {cronJob.next_execution_at && cronJob.status === 'active' && (
                    <p>
                      Next run: {format(new Date(cronJob.next_execution_at), 'MMM d, yyyy HH:mm')}
                    </p>
                  )}
                </div>

                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleTrigger(cronJob)}
                    disabled={cronJob.status === 'suspended'}
                    data-testid={`trigger-${cronJob.id}`}
                  >
                    <Play className="h-4 w-4 mr-1" />
                    Run Now
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleToggleStatus(cronJob)}
                    data-testid={`toggle-${cronJob.id}`}
                  >
                    {cronJob.status === 'active' ? (
                      <>
                        <Pause className="h-4 w-4 mr-1" />
                        Suspend
                      </>
                    ) : (
                      <>
                        <Play className="h-4 w-4 mr-1" />
                        Activate
                      </>
                    )}
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setEditingSchedule(cronJob)}
                    data-testid={`edit-schedule-${cronJob.id}`}
                  >
                    <Edit className="h-4 w-4 mr-1" />
                    Edit Schedule
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => handleViewHistory(cronJob)}
                    data-testid={`view-history-${cronJob.id}`}
                  >
                    <History className="h-4 w-4 mr-1" />
                    History
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Dialog
        open={!!editingSchedule}
        onOpenChange={() => setEditingSchedule(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Schedule</DialogTitle>
            <DialogDescription>
              Update the cron schedule for {editingSchedule?.name}
            </DialogDescription>
          </DialogHeader>
          {editingSchedule && (
            <CronJobScheduleEditor
              currentSchedule={editingSchedule.cron_schedule || ''}
              onSave={handleScheduleUpdate}
              onCancel={() => setEditingSchedule(null)}
            />
          )}
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!viewingHistory}
        onOpenChange={() => setViewingHistory(null)}
      >
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Execution History</DialogTitle>
            <DialogDescription>
              Recent executions for {viewingHistory?.name}
            </DialogDescription>
          </DialogHeader>
          <CronJobExecutionHistory
            executions={executions}
            loading={executionsLoading}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}