'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { 
  Loader2, 
  Plus, 
  Calendar, 
  Shield, 
  Database,
  HardDrive,
  PlayCircle,
  MoreHorizontal,
  Clock
} from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { CreateBackupPolicyDialog } from './create-backup-policy-dialog'
import { BackupExecutionHistory } from './backup-execution-history'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface BackupPolicy {
  id: string
  application_id: string
  storage_id: string
  enabled: boolean
  schedule: string
  retention_days: number
  backup_type: 'full' | 'incremental' | 'application'
  include_volumes: boolean
  include_database: boolean
  include_config: boolean
  compression_enabled: boolean
  encryption_enabled: boolean
  created_at: string
  updated_at: string
}

interface BackupPolicyListProps {
  orgId: string
  workspaceId: string
  applicationId: string
  applicationName: string
}

export function BackupPolicyList({ 
  orgId, 
  workspaceId, 
  applicationId,
  applicationName 
}: BackupPolicyListProps) {
  const { toast } = useToast()
  const [policy, setPolicy] = useState<BackupPolicy | null>(null)
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [executing, setExecuting] = useState(false)

  useEffect(() => {
    fetchPolicy()
  }, [applicationId])

  const fetchPolicy = async () => {
    try {
      setLoading(true)
      const response = await apiClient.backupApi.getBackupPolicy(orgId, workspaceId, applicationId)
      setPolicy(response.data)
    } catch (error: any) {
      if (error.response?.status !== 404) {
        toast({
          title: 'Error',
          description: 'Failed to fetch backup policy',
          variant: 'destructive',
        })
      }
    } finally {
      setLoading(false)
    }
  }

  const handlePolicyCreated = () => {
    fetchPolicy()
    setCreateDialogOpen(false)
  }

  const handleTogglePolicy = async (enabled: boolean) => {
    if (!policy) return

    try {
      await apiClient.backupApi.updateBackupPolicy(orgId, workspaceId, policy.id, {
        enabled,
      })
      setPolicy({ ...policy, enabled })
      toast({
        title: 'Success',
        description: `Backup policy ${enabled ? 'enabled' : 'disabled'}`,
      })
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update backup policy',
        variant: 'destructive',
      })
    }
  }

  const handleManualBackup = async () => {
    if (!policy) return

    try {
      setExecuting(true)
      const response = await apiClient.backupApi.executeBackupPolicy(orgId, workspaceId, policy.id)
      toast({
        title: 'Success',
        description: 'Backup started successfully',
      })
      setShowHistory(true)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to start backup',
        variant: 'destructive',
      })
    } finally {
      setExecuting(false)
    }
  }

  const handleDeletePolicy = async () => {
    if (!policy) return

    try {
      await apiClient.backupApi.deleteBackupPolicy(orgId, workspaceId, policy.id)
      toast({
        title: 'Success',
        description: 'Backup policy deleted',
      })
      setPolicy(null)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete backup policy',
        variant: 'destructive',
      })
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Backup Policy</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (!policy) {
    return (
      <>
        <Card>
          <CardHeader>
            <CardTitle>Backup Policy</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8">
              <Shield className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">
                No backup policy configured for {applicationName}.
              </p>
              <Button
                className="mt-4"
                onClick={() => setCreateDialogOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                Create Backup Policy
              </Button>
            </div>
          </CardContent>
        </Card>

        <CreateBackupPolicyDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          orgId={orgId}
          workspaceId={workspaceId}
          applicationId={applicationId}
          onSuccess={handlePolicyCreated}
        />
      </>
    )
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex items-center gap-4">
            <CardTitle>Backup Policy</CardTitle>
            <Switch
              checked={policy.enabled}
              onCheckedChange={handleTogglePolicy}
            />
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowHistory(!showHistory)}
            >
              <Clock className="mr-2 h-4 w-4" />
              History
            </Button>
            <Button
              size="sm"
              onClick={handleManualBackup}
              disabled={!policy.enabled || executing}
            >
              {executing ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <PlayCircle className="mr-2 h-4 w-4" />
              )}
              Backup Now
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm">
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => setCreateDialogOpen(true)}>
                  Edit Policy
                </DropdownMenuItem>
                <DropdownMenuItem 
                  className="text-destructive"
                  onClick={handleDeletePolicy}
                >
                  Delete Policy
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <p className="text-sm font-medium">Schedule</p>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <code className="text-sm">{policy.schedule}</code>
                </div>
              </div>
              <div className="space-y-1">
                <p className="text-sm font-medium">Retention</p>
                <p className="text-sm text-muted-foreground">
                  {policy.retention_days} days
                </p>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-medium">Backup Configuration</p>
              <div className="flex flex-wrap gap-2">
                <Badge variant={policy.backup_type === 'full' ? 'default' : 'secondary'}>
                  {policy.backup_type}
                </Badge>
                {policy.include_volumes && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge variant="outline">
                          <HardDrive className="mr-1 h-3 w-3" />
                          Volumes
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>Persistent volumes will be included</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
                {policy.include_database && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge variant="outline">
                          <Database className="mr-1 h-3 w-3" />
                          Database
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>Database dumps will be included</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
                {policy.compression_enabled && (
                  <Badge variant="outline">Compressed</Badge>
                )}
                {policy.encryption_enabled && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge variant="outline">
                          <Shield className="mr-1 h-3 w-3" />
                          Encrypted
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>Backups are encrypted at rest</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {showHistory && policy && (
        <BackupExecutionHistory
          orgId={orgId}
          workspaceId={workspaceId}
          policyId={policy.id}
          applicationId={applicationId}
        />
      )}

      <CreateBackupPolicyDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        orgId={orgId}
        workspaceId={workspaceId}
        applicationId={applicationId}
        existingPolicy={policy}
        onSuccess={handlePolicyCreated}
      />
    </>
  )
}