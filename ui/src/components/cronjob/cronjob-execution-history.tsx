'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { 
  Loader2, 
  CheckCircle2,
  XCircle,
  Clock,
  Terminal,
  FileText,
  X
} from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { formatDistanceToNow, format } from 'date-fns'

interface CronJobExecution {
  id: string
  application_id: string
  job_name: string
  started_at: string
  completed_at?: string
  status: 'running' | 'succeeded' | 'failed' | 'cancelled'
  exit_code?: number
  logs?: string
  created_at: string
  updated_at: string
}

interface CronJobExecutionHistoryProps {
  orgId: string
  workspaceId: string
  cronJobId: string
  cronJobName: string
  onClose?: () => void
}

export function CronJobExecutionHistory({
  orgId,
  workspaceId,
  cronJobId,
  cronJobName,
  onClose,
}: CronJobExecutionHistoryProps) {
  const { toast } = useToast()
  const [executions, setExecutions] = useState<CronJobExecution[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selectedExecution, setSelectedExecution] = useState<CronJobExecution | null>(null)
  const [logsDialogOpen, setLogsDialogOpen] = useState(false)
  const pageSize = 10

  useEffect(() => {
    fetchExecutions()
    
    // Poll for updates if there are running executions
    const interval = setInterval(() => {
      if (executions.some(e => e.status === 'running')) {
        fetchExecutions()
      }
    }, 5000)

    return () => clearInterval(interval)
  }, [cronJobId, page])

  const fetchExecutions = async () => {
    try {
      setLoading(true)
      const response = await apiClient.applicationsApi.getCronJobExecutions(
        orgId,
        workspaceId,
        cronJobId,
        {
          page,
          page_size: pageSize,
        }
      )
      setExecutions(response.data.executions)
      setTotal(response.data.total)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch execution history',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleViewLogs = (execution: CronJobExecution) => {
    setSelectedExecution(execution)
    setLogsDialogOpen(true)
  }

  const getStatusIcon = (status: CronJobExecution['status']) => {
    switch (status) {
      case 'succeeded':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />
      case 'running':
        return <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
      case 'cancelled':
        return <XCircle className="h-4 w-4 text-gray-600" />
    }
  }

  const getStatusBadgeVariant = (status: CronJobExecution['status']) => {
    switch (status) {
      case 'succeeded':
        return 'default'
      case 'failed':
        return 'destructive'
      case 'running':
        return 'secondary'
      case 'cancelled':
        return 'outline'
    }
  }

  const calculateDuration = (execution: CronJobExecution) => {
    if (!execution.completed_at) return 'In progress'
    const start = new Date(execution.started_at)
    const end = new Date(execution.completed_at)
    const duration = end.getTime() - start.getTime()
    const minutes = Math.floor(duration / 60000)
    const seconds = Math.floor((duration % 60000) / 1000)
    
    if (minutes === 0) return `${seconds} seconds`
    return `${minutes} minutes`
  }

  if (loading && executions.length === 0) {
    return (
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Execution History - {cronJobName}</CardTitle>
          {onClose && (
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          )}
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
          <CardTitle>Execution History - {cronJobName}</CardTitle>
          {onClose && (
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {executions.length === 0 ? (
            <div className="text-center py-8">
              <Clock className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">
                No executions yet. The CronJob will run according to its schedule.
              </p>
            </div>
          ) : (
            <ScrollArea className="h-[400px]">
              <div className="space-y-4">
                {executions.map((execution) => (
                  <div
                    key={execution.id}
                    className="border rounded-lg p-4 space-y-3"
                  >
                    <div className="flex items-start justify-between">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          {getStatusIcon(execution.status)}
                          <span className="font-mono text-sm">{execution.job_name}</span>
                          <Badge variant={getStatusBadgeVariant(execution.status)}>
                            {execution.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-foreground">
                          Started: {format(new Date(execution.started_at), 'PPpp')}
                        </p>
                      </div>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleViewLogs(execution)}
                        data-execution-id={execution.id}
                      >
                        <FileText className="mr-2 h-4 w-4" />
                        View Logs
                      </Button>
                    </div>

                    <div className="grid grid-cols-3 gap-4 text-sm">
                      <div>
                        <p className="text-muted-foreground">Duration</p>
                        <p className="font-medium">{calculateDuration(execution)}</p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Exit Code</p>
                        <p className="font-medium">
                          {execution.exit_code !== undefined ? (
                            <span className={execution.exit_code === 0 ? 'text-green-600' : 'text-red-600'}>
                              {execution.exit_code}
                            </span>
                          ) : (
                            'N/A'
                          )}
                        </p>
                      </div>
                      <div>
                        <p className="text-muted-foreground">Completed</p>
                        <p className="font-medium">
                          {execution.completed_at
                            ? formatDistanceToNow(new Date(execution.completed_at), { addSuffix: true })
                            : 'Running'}
                        </p>
                      </div>
                    </div>

                    {execution.status === 'failed' && execution.logs && (
                      <div className="text-sm text-red-600 bg-red-50 rounded p-2 font-mono">
                        {execution.logs.split('\n').slice(-2).join('\n')}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}

          {total > pageSize && (
            <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page - 1)}
                disabled={page === 1}
              >
                Previous
              </Button>
              <span className="text-sm text-muted-foreground">
                Page {page} of {Math.ceil(total / pageSize)}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page + 1)}
                disabled={page >= Math.ceil(total / pageSize)}
              >
                Next
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={logsDialogOpen} onOpenChange={setLogsDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Terminal className="h-5 w-5" />
              Execution Logs
            </DialogTitle>
          </DialogHeader>
          {selectedExecution && (
            <div className="space-y-4">
              <div className="flex items-center gap-4 text-sm">
                <div className="flex items-center gap-2">
                  {getStatusIcon(selectedExecution.status)}
                  <Badge variant={getStatusBadgeVariant(selectedExecution.status)}>
                    {selectedExecution.status}
                  </Badge>
                </div>
                <span className="text-muted-foreground">
                  {selectedExecution.job_name}
                </span>
                {selectedExecution.exit_code !== undefined && (
                  <span className="text-muted-foreground">
                    Exit code: {selectedExecution.exit_code}
                  </span>
                )}
              </div>
              
              <ScrollArea className="h-[400px] w-full rounded border bg-black p-4">
                <pre className="text-xs text-green-400 font-mono whitespace-pre-wrap">
                  {selectedExecution.logs || 'No logs available'}
                </pre>
              </ScrollArea>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}