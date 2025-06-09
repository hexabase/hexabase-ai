'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Loader2, 
  Plus, 
  Clock, 
  PlayCircle, 
  PauseCircle,
  MoreHorizontal,
  CalendarClock,
  Activity,
  AlertCircle
} from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { CreateCronJobDialog } from './create-cronjob-dialog'
import { CronJobExecutionHistory } from './cronjob-execution-history'
import { formatDistanceToNow, format } from 'date-fns'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface CronJob {
  id: string
  name: string
  type: 'cronjob'
  status: 'active' | 'suspended' | 'error'
  cron_schedule: string
  cron_command?: string[]
  cron_args?: string[]
  template_app_id?: string
  last_execution_at?: string
  next_execution_at?: string
  source_type: string
  source_image?: string
  created_at: string
  updated_at: string
}

interface CronJobListProps {
  orgId: string
  workspaceId: string
  projectId: string
}

export function CronJobList({ orgId, workspaceId, projectId }: CronJobListProps) {
  const { toast } = useToast()
  const [cronJobs, setCronJobs] = useState<CronJob[]>([])
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [selectedCronJob, setSelectedCronJob] = useState<CronJob | null>(null)
  const [showHistory, setShowHistory] = useState(false)

  useEffect(() => {
    fetchCronJobs()
  }, [orgId, workspaceId, projectId])

  const fetchCronJobs = async () => {
    try {
      setLoading(true)
      const response = await apiClient.applicationsApi.list(orgId, workspaceId, projectId, {
        type: 'cronjob'
      })
      setCronJobs(response.data.applications)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch CronJobs',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleToggleStatus = async (cronJob: CronJob) => {
    const newStatus = cronJob.status === 'active' ? 'suspended' : 'active'
    
    try {
      await apiClient.applicationsApi.updateStatus(orgId, workspaceId, cronJob.id, {
        status: newStatus
      })
      
      toast({
        title: 'Success',
        description: `CronJob ${newStatus === 'active' ? 'activated' : 'suspended'}`,
      })
      
      fetchCronJobs()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update CronJob status',
        variant: 'destructive',
      })
    }
  }

  const handleTriggerNow = async (cronJob: CronJob) => {
    try {
      const response = await apiClient.applicationsApi.triggerCronJob(orgId, workspaceId, cronJob.id)
      
      toast({
        title: 'Success',
        description: 'CronJob triggered successfully',
      })
      
      // Show execution history to see the new execution
      setSelectedCronJob(cronJob)
      setShowHistory(true)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to trigger CronJob',
        variant: 'destructive',
      })
    }
  }

  const handleDelete = async (cronJob: CronJob) => {
    try {
      await apiClient.applicationsApi.delete(orgId, workspaceId, projectId, cronJob.id)
      
      toast({
        title: 'Success',
        description: 'CronJob deleted successfully',
      })
      
      fetchCronJobs()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete CronJob',
        variant: 'destructive',
      })
    }
  }

  const parseCronExpression = (expression: string): string => {
    // Simple cron expression parser for common patterns
    const patterns: { [key: string]: string } = {
      '* * * * *': 'Every minute',
      '*/5 * * * *': 'Every 5 minutes',
      '*/10 * * * *': 'Every 10 minutes',
      '*/15 * * * *': 'Every 15 minutes',
      '*/30 * * * *': 'Every 30 minutes',
      '0 * * * *': 'Every hour',
      '0 */2 * * *': 'Every 2 hours',
      '0 */3 * * *': 'Every 3 hours',
      '0 */6 * * *': 'Every 6 hours',
      '0 */12 * * *': 'Every 12 hours',
      '0 0 * * *': 'Daily at midnight',
      '0 2 * * *': 'Daily at 2:00 AM',
      '0 0 * * 0': 'Weekly on Sunday',
      '0 0 * * 1': 'Weekly on Monday',
      '0 0 1 * *': 'Monthly on the 1st',
      '0 0 15 * *': 'Monthly on the 15th',
    }
    
    return patterns[expression] || expression
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>CronJobs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>CronJobs</CardTitle>
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create CronJob
          </Button>
        </CardHeader>
        <CardContent>
          {cronJobs.length === 0 ? (
            <div className="text-center py-8">
              <Clock className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">
                No CronJobs configured yet.
              </p>
              <Button
                className="mt-4"
                variant="outline"
                onClick={() => setCreateDialogOpen(true)}
              >
                Create your first CronJob
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {cronJobs.map((cronJob) => (
                <CronJobCard
                  key={cronJob.id}
                  cronJob={cronJob}
                  onToggleStatus={() => handleToggleStatus(cronJob)}
                  onTriggerNow={() => handleTriggerNow(cronJob)}
                  onViewHistory={() => {
                    setSelectedCronJob(cronJob)
                    setShowHistory(true)
                  }}
                  onDelete={() => handleDelete(cronJob)}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <CreateCronJobDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        orgId={orgId}
        workspaceId={workspaceId}
        projectId={projectId}
        onSuccess={fetchCronJobs}
      />

      {selectedCronJob && showHistory && (
        <CronJobExecutionHistory
          orgId={orgId}
          workspaceId={workspaceId}
          cronJobId={selectedCronJob.id}
          cronJobName={selectedCronJob.name}
          onClose={() => {
            setShowHistory(false)
            setSelectedCronJob(null)
          }}
        />
      )}
    </>
  )
}

function CronJobCard({ 
  cronJob, 
  onToggleStatus,
  onTriggerNow,
  onViewHistory,
  onDelete
}: { 
  cronJob: CronJob
  onToggleStatus: () => void
  onTriggerNow: () => void
  onViewHistory: () => void
  onDelete: () => void
}) {
  const parseCronExpression = (expression: string): string => {
    const patterns: { [key: string]: string } = {
      '* * * * *': 'Every minute',
      '*/5 * * * *': 'Every 5 minutes',
      '*/10 * * * *': 'Every 10 minutes',
      '*/15 * * * *': 'Every 15 minutes',
      '*/30 * * * *': 'Every 30 minutes',
      '0 * * * *': 'Every hour',
      '0 */2 * * *': 'Every 2 hours',
      '0 */3 * * *': 'Every 3 hours',
      '0 */6 * * *': 'Every 6 hours',
      '0 */12 * * *': 'Every 12 hours',
      '0 0 * * *': 'Daily at midnight',
      '0 2 * * *': 'Daily at 2:00 AM',
      '0 0 * * 0': 'Weekly on Sunday',
      '0 0 * * 1': 'Weekly on Monday',
      '0 0 1 * *': 'Monthly on the 1st',
      '0 0 15 * *': 'Monthly on the 15th',
    }
    
    return patterns[expression] || expression
  }

  const getStatusIcon = (status: CronJob['status']) => {
    switch (status) {
      case 'active':
        return <PlayCircle className="h-4 w-4 text-green-600" />
      case 'suspended':
        return <PauseCircle className="h-4 w-4 text-gray-600" />
      case 'error':
        return <AlertCircle className="h-4 w-4 text-red-600" />
    }
  }

  const getStatusBadgeVariant = (status: CronJob['status']) => {
    switch (status) {
      case 'active':
        return 'default'
      case 'suspended':
        return 'secondary'
      case 'error':
        return 'destructive'
    }
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              {getStatusIcon(cronJob.status)}
              <h3 className="font-semibold">{cronJob.name}</h3>
              <Badge variant={getStatusBadgeVariant(cronJob.status)}>
                {cronJob.status}
              </Badge>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger className="flex items-center gap-1">
                    <CalendarClock className="h-3 w-3" />
                    <code>{cronJob.cron_schedule}</code>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{parseCronExpression(cronJob.cron_schedule)}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              
              {cronJob.source_image && (
                <span className="flex items-center gap-1">
                  <Activity className="h-3 w-3" />
                  {cronJob.source_image}
                </span>
              )}
            </div>
          </div>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={onViewHistory}>
                <Clock className="mr-2 h-4 w-4" />
                View History
              </DropdownMenuItem>
              <DropdownMenuItem onClick={onTriggerNow} disabled={cronJob.status !== 'active'}>
                <PlayCircle className="mr-2 h-4 w-4" />
                Trigger Now
              </DropdownMenuItem>
              <DropdownMenuItem onClick={onToggleStatus}>
                {cronJob.status === 'active' ? (
                  <>
                    <PauseCircle className="mr-2 h-4 w-4" />
                    Suspend
                  </>
                ) : (
                  <>
                    <PlayCircle className="mr-2 h-4 w-4" />
                    Activate
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Edit</DropdownMenuItem>
              <DropdownMenuItem>Save as Template</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem 
                className="text-destructive"
                onClick={onDelete}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-muted-foreground">Last Execution</p>
            <p className="font-medium">
              {cronJob.last_execution_at ? (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      {formatDistanceToNow(new Date(cronJob.last_execution_at), { addSuffix: true })}
                    </TooltipTrigger>
                    <TooltipContent>
                      {format(new Date(cronJob.last_execution_at), 'PPpp')}
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ) : (
                'Never'
              )}
            </p>
          </div>
          <div>
            <p className="text-muted-foreground">Next Execution</p>
            <p className="font-medium">
              {cronJob.next_execution_at && cronJob.status === 'active' ? (
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      {formatDistanceToNow(new Date(cronJob.next_execution_at), { addSuffix: true })}
                    </TooltipTrigger>
                    <TooltipContent>
                      {format(new Date(cronJob.next_execution_at), 'PPpp')}
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ) : (
                'N/A'
              )}
            </p>
          </div>
        </div>

        {cronJob.cron_command && cronJob.cron_command.length > 0 && (
          <div className="mt-3 text-sm">
            <p className="text-muted-foreground">Command</p>
            <code className="text-xs bg-muted px-2 py-1 rounded">
              {cronJob.cron_command.join(' ')} {cronJob.cron_args?.join(' ')}
            </code>
          </div>
        )}
      </CardContent>
    </Card>
  )
}