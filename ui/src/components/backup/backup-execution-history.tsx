'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { 
  Loader2, 
  Download, 
  RotateCcw, 
  CheckCircle2,
  XCircle,
  Clock,
  HardDrive,
  FileArchive,
  AlertCircle
} from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { RestoreBackupDialog } from './restore-backup-dialog'
import { formatDistanceToNow, format } from 'date-fns'

interface BackupExecution {
  id: string
  policy_id: string
  status: 'running' | 'succeeded' | 'failed' | 'cancelled'
  size_bytes: number
  compressed_size_bytes: number
  backup_path: string
  started_at: string
  completed_at?: string
  error_message?: string
  backup_manifest?: {
    volumes?: string[]
    databases?: string[]
    config_maps?: string[]
  }
}

interface BackupExecutionHistoryProps {
  orgId: string
  workspaceId: string
  policyId: string
  applicationId: string
}

export function BackupExecutionHistory({
  orgId,
  workspaceId,
  policyId,
  applicationId,
}: BackupExecutionHistoryProps) {
  const { toast } = useToast()
  const [executions, setExecutions] = useState<BackupExecution[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [selectedExecution, setSelectedExecution] = useState<BackupExecution | null>(null)
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false)
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
  }, [policyId, page])

  const fetchExecutions = async () => {
    try {
      setLoading(true)
      const response = await apiClient.backupApi.listBackupExecutions(
        orgId,
        workspaceId,
        policyId,
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
        description: 'Failed to fetch backup history',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleRestore = (execution: BackupExecution) => {
    setSelectedExecution(execution)
    setRestoreDialogOpen(true)
  }

  const handleValidate = async (execution: BackupExecution) => {
    try {
      const response = await apiClient.backupApi.validateBackup(
        orgId,
        workspaceId,
        execution.id
      )
      
      if (response.data.valid) {
        toast({
          title: 'Backup Valid',
          description: 'Backup integrity check passed',
        })
      } else {
        toast({
          title: 'Backup Invalid',
          description: 'Backup integrity check failed',
          variant: 'destructive',
        })
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to validate backup',
        variant: 'destructive',
      })
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const getStatusIcon = (status: BackupExecution['status']) => {
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

  const getStatusBadgeVariant = (status: BackupExecution['status']) => {
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

  const calculateDuration = (execution: BackupExecution) => {
    if (!execution.completed_at) return 'In progress'
    const start = new Date(execution.started_at)
    const end = new Date(execution.completed_at)
    const duration = end.getTime() - start.getTime()
    const minutes = Math.floor(duration / 60000)
    const seconds = Math.floor((duration % 60000) / 1000)
    return `${minutes}m ${seconds}s`
  }

  if (loading && executions.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Backup History</CardTitle>
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
        <CardHeader>
          <CardTitle>Backup History</CardTitle>
        </CardHeader>
        <CardContent>
          {executions.length === 0 ? (
            <div className="text-center py-8">
              <Clock className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">
                No backup executions yet. Backups will appear here once they run.
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
                          <Badge variant={getStatusBadgeVariant(execution.status)}>
                            {execution.status}
                          </Badge>
                          <span className="text-sm text-muted-foreground">
                            {formatDistanceToNow(new Date(execution.started_at), { addSuffix: true })}
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground">
                          Started: {format(new Date(execution.started_at), 'PPpp')}
                        </p>
                        {execution.error_message && (
                          <p className="text-sm text-destructive">{execution.error_message}</p>
                        )}
                      </div>
                      <div className="flex items-center gap-2">
                        {execution.status === 'succeeded' && (
                          <>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleValidate(execution)}
                            >
                              Validate
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleRestore(execution)}
                              data-execution-id={execution.id}
                            >
                              <RotateCcw className="mr-2 h-4 w-4" />
                              Restore
                            </Button>
                          </>
                        )}
                      </div>
                    </div>

                    {execution.status === 'succeeded' && (
                      <div className="grid grid-cols-3 gap-4 text-sm">
                        <div className="space-y-1">
                          <p className="text-muted-foreground">Duration</p>
                          <p className="font-medium">{calculateDuration(execution)}</p>
                        </div>
                        <div className="space-y-1">
                          <p className="text-muted-foreground">Size</p>
                          <div className="flex items-center gap-1">
                            <HardDrive className="h-3 w-3" />
                            <span className="font-medium">{formatBytes(execution.size_bytes)}</span>
                            {execution.compressed_size_bytes > 0 && (
                              <>
                                <span className="text-muted-foreground">→</span>
                                <FileArchive className="h-3 w-3" />
                                <span className="font-medium">{formatBytes(execution.compressed_size_bytes)}</span>
                              </>
                            )}
                          </div>
                        </div>
                        <div className="space-y-1">
                          <p className="text-muted-foreground">Compression</p>
                          <p className="font-medium">
                            {execution.compressed_size_bytes > 0
                              ? `${Math.round((1 - execution.compressed_size_bytes / execution.size_bytes) * 100)}%`
                              : 'N/A'}
                          </p>
                        </div>
                      </div>
                    )}

                    {execution.backup_manifest && (
                      <div className="text-sm space-y-1 pt-2 border-t">
                        <p className="text-muted-foreground">Backup Contents:</p>
                        <div className="flex flex-wrap gap-2">
                          {execution.backup_manifest.volumes && execution.backup_manifest.volumes.length > 0 && (
                            <Badge variant="outline">
                              {execution.backup_manifest.volumes.length} volumes
                            </Badge>
                          )}
                          {execution.backup_manifest.databases && execution.backup_manifest.databases.length > 0 && (
                            <Badge variant="outline">
                              {execution.backup_manifest.databases.length} databases
                            </Badge>
                          )}
                          {execution.backup_manifest.config_maps && execution.backup_manifest.config_maps.length > 0 && (
                            <Badge variant="outline">
                              {execution.backup_manifest.config_maps.length} configs
                            </Badge>
                          )}
                        </div>
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

      {selectedExecution && (
        <RestoreBackupDialog
          open={restoreDialogOpen}
          onOpenChange={setRestoreDialogOpen}
          orgId={orgId}
          workspaceId={workspaceId}
          applicationId={applicationId}
          backupExecution={selectedExecution}
          onSuccess={() => {
            setRestoreDialogOpen(false)
            toast({
              title: 'Success',
              description: 'Restore started successfully',
            })
          }}
        />
      )}
    </>
  )
}