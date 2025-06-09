'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Loader2, AlertTriangle, Info } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { TaskMonitor } from '@/components/task-monitor'

interface RestoreBackupDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgId: string
  workspaceId: string
  applicationId: string
  backupExecution: {
    id: string
    backup_manifest?: {
      volumes?: string[]
      databases?: string[]
      config_maps?: string[]
    }
  }
  onSuccess?: () => void
}

export function RestoreBackupDialog({
  open,
  onOpenChange,
  orgId,
  workspaceId,
  applicationId,
  backupExecution,
  onSuccess,
}: RestoreBackupDialogProps) {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  const [validating, setValidating] = useState(false)
  const [validationResult, setValidationResult] = useState<any>(null)
  const [formData, setFormData] = useState({
    restore_type: 'in_place',
    restore_volumes: true,
    restore_database: true,
    restore_config: true,
    stop_application: true,
  })

  const handleValidate = async () => {
    try {
      setValidating(true)
      const response = await apiClient.backupApi.validateBackup(
        orgId,
        workspaceId,
        backupExecution.id
      )
      setValidationResult(response.data)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to validate backup',
        variant: 'destructive',
      })
    } finally {
      setValidating(false)
    }
  }

  const handleRestore = async () => {
    try {
      setLoading(true)
      const response = await apiClient.backupApi.restoreBackup(
        orgId,
        workspaceId,
        applicationId,
        {
          backup_execution_id: backupExecution.id,
          restore_type: formData.restore_type,
          restore_options: {
            restore_volumes: formData.restore_volumes,
            restore_database: formData.restore_database,
            restore_config: formData.restore_config,
            stop_application: formData.stop_application,
          },
        }
      )

      if (response.data.task_id) {
        setTaskId(response.data.task_id)
      } else {
        toast({
          title: 'Success',
          description: 'Restore started successfully',
        })
        onSuccess?.()
        handleClose()
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to start restore',
        variant: 'destructive',
      })
      setLoading(false)
    }
  }

  const handleClose = () => {
    setFormData({
      restore_type: 'in_place',
      restore_volumes: true,
      restore_database: true,
      restore_config: true,
      stop_application: true,
    })
    setTaskId(null)
    setLoading(false)
    setValidationResult(null)
    onOpenChange(false)
  }

  const handleTaskComplete = () => {
    toast({
      title: 'Success',
      description: 'Restore completed successfully',
    })
    onSuccess?.()
    handleClose()
  }

  const handleTaskError = (error: string) => {
    toast({
      title: 'Error',
      description: error,
      variant: 'destructive',
    })
    setLoading(false)
    setTaskId(null)
  }

  const manifest = backupExecution.backup_manifest

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Restore from Backup</DialogTitle>
          <DialogDescription>
            Restore your application from a previous backup. This operation will replace
            the current state with the backup data.
          </DialogDescription>
        </DialogHeader>

        {taskId ? (
          <TaskMonitor
            taskId={taskId}
            onComplete={handleTaskComplete}
            onError={handleTaskError}
          />
        ) : (
          <>
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Warning</AlertTitle>
              <AlertDescription>
                Restoring from backup will overwrite the current application state.
                This action cannot be undone. Make sure to backup current state if needed.
              </AlertDescription>
            </Alert>

            {validationResult && (
              <Alert variant={validationResult.valid ? 'default' : 'destructive'}>
                <Info className="h-4 w-4" />
                <AlertTitle>Validation Result</AlertTitle>
                <AlertDescription>
                  {validationResult.valid ? (
                    <>
                      Backup is valid. Integrity check: {validationResult.integrity_check}
                    </>
                  ) : (
                    <>
                      Backup validation failed. This backup may be corrupted.
                    </>
                  )}
                </AlertDescription>
              </Alert>
            )}

            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="restore_type">Restore Type</Label>
                <Select
                  name="restore_type"
                  value={formData.restore_type}
                  onValueChange={(value) => setFormData({ ...formData, restore_type: value as any })}
                  disabled={loading}
                >
                  <SelectTrigger id="restore_type">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="in_place">In-Place Restore</SelectItem>
                    <SelectItem value="new_application" disabled>
                      Restore to New Application (Coming Soon)
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label>Restore Components</Label>
                <div className="space-y-2">
                  {manifest?.volumes && manifest.volumes.length > 0 && (
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id="restore_volumes"
                        name="restore_volumes"
                        checked={formData.restore_volumes}
                        onCheckedChange={(checked) => 
                          setFormData({ ...formData, restore_volumes: checked as boolean })
                        }
                        disabled={loading}
                      />
                      <label
                        htmlFor="restore_volumes"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                      >
                        Restore volumes ({manifest.volumes.length} found)
                      </label>
                    </div>
                  )}
                  {manifest?.databases && manifest.databases.length > 0 && (
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id="restore_database"
                        name="restore_database"
                        checked={formData.restore_database}
                        onCheckedChange={(checked) => 
                          setFormData({ ...formData, restore_database: checked as boolean })
                        }
                        disabled={loading}
                      />
                      <label
                        htmlFor="restore_database"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                      >
                        Restore databases ({manifest.databases.join(', ')})
                      </label>
                    </div>
                  )}
                  {manifest?.config_maps && manifest.config_maps.length > 0 && (
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id="restore_config"
                        checked={formData.restore_config}
                        onCheckedChange={(checked) => 
                          setFormData({ ...formData, restore_config: checked as boolean })
                        }
                        disabled={loading}
                      />
                      <label
                        htmlFor="restore_config"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                      >
                        Restore configuration ({manifest.config_maps.length} configs)
                      </label>
                    </div>
                  )}
                </div>
              </div>

              <div className="space-y-2">
                <Label>Restore Options</Label>
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="stop_application"
                    checked={formData.stop_application}
                    onCheckedChange={(checked) => 
                      setFormData({ ...formData, stop_application: checked as boolean })
                    }
                    disabled={loading}
                  />
                  <label
                    htmlFor="stop_application"
                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                  >
                    Stop application during restore (recommended)
                  </label>
                </div>
              </div>
            </div>

            <DialogFooter>
              {!validationResult && (
                <Button
                  variant="outline"
                  onClick={handleValidate}
                  disabled={loading || validating}
                >
                  {validating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Validate Backup
                </Button>
              )}
              <Button variant="outline" onClick={handleClose} disabled={loading}>
                Cancel
              </Button>
              <Button 
                onClick={handleRestore} 
                disabled={loading || (validationResult && !validationResult.valid)}
                variant="destructive"
              >
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Start Restore
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}