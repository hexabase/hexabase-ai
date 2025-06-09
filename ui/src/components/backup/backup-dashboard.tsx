'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { 
  Loader2, 
  HardDrive, 
  AlertTriangle,
  TrendingUp,
  Database,
  Archive,
  Calendar
} from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { format, formatDistanceToNow } from 'date-fns'

interface BackupStorageUsage {
  storage_id: string
  total_gb: number
  used_gb: number
  available_gb: number
  usage_percent: number
  backup_count: number
  oldest_backup?: string
  latest_backup?: string
}

interface BackupDashboardProps {
  orgId: string
  workspaceId: string
  workspacePlan: 'shared' | 'dedicated'
}

export function BackupDashboard({ orgId, workspaceId, workspacePlan }: BackupDashboardProps) {
  const { toast } = useToast()
  const [usage, setUsage] = useState<BackupStorageUsage[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (workspacePlan === 'dedicated') {
      fetchUsage()
    } else {
      setLoading(false)
    }
  }, [workspaceId, workspacePlan])

  const fetchUsage = async () => {
    try {
      setLoading(true)
      const response = await apiClient.backupApi.getStorageUsage(orgId, workspaceId)
      setUsage(response.data)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch storage usage',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  if (workspacePlan !== 'dedicated') {
    return null
  }

  if (loading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Loading...</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-center py-4">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  const totalStorage = usage.reduce((sum, u) => sum + u.total_gb, 0)
  const totalUsed = usage.reduce((sum, u) => sum + u.used_gb, 0)
  const totalBackups = usage.reduce((sum, u) => sum + u.backup_count, 0)
  const averageUsage = usage.length > 0 ? usage.reduce((sum, u) => sum + u.usage_percent, 0) / usage.length : 0

  const criticalStorages = usage.filter(u => u.usage_percent > 95)
  const warningStorages = usage.filter(u => u.usage_percent > 90 && u.usage_percent <= 95)

  return (
    <div className="space-y-4">
      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Storage</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalStorage} GB</div>
            <p className="text-xs text-muted-foreground">
              Across {usage.length} storage{usage.length !== 1 ? 's' : ''}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Used Storage</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalUsed} GB</div>
            <Progress value={totalStorage > 0 ? (totalUsed / totalStorage) * 100 : 0} className="mt-2" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Backups</CardTitle>
            <Archive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalBackups}</div>
            <p className="text-xs text-muted-foreground">
              Avg {averageUsage.toFixed(1)}% usage
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Storage Health</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {criticalStorages.length > 0 ? (
                <span className="text-red-600">Critical</span>
              ) : warningStorages.length > 0 ? (
                <span className="text-yellow-600">Warning</span>
              ) : (
                <span className="text-green-600">Healthy</span>
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {criticalStorages.length + warningStorages.length} storage{criticalStorages.length + warningStorages.length !== 1 ? 's' : ''} need attention
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Alerts */}
      {criticalStorages.length > 0 && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Storage usage critical</AlertTitle>
          <AlertDescription>
            {criticalStorages.length} storage{criticalStorages.length !== 1 ? 's are' : ' is'} above 95% capacity.
            Immediate action required to prevent backup failures.
          </AlertDescription>
        </Alert>
      )}

      {warningStorages.length > 0 && criticalStorages.length === 0 && (
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>High storage usage</AlertTitle>
          <AlertDescription>
            {warningStorages.length} storage{warningStorages.length !== 1 ? 's are' : ' is'} above 90% capacity.
            Consider expanding storage or cleaning up old backups.
          </AlertDescription>
        </Alert>
      )}

      {/* Individual Storage Usage */}
      {usage.length > 0 && (
        <div className="grid gap-4 md:grid-cols-2">
          {usage.map((storageUsage) => (
            <Card key={storageUsage.storage_id}>
              <CardHeader>
                <CardTitle className="text-base">Storage {storageUsage.storage_id}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Usage</span>
                    <span className={storageUsage.usage_percent > 95 ? 'text-red-600 font-semibold' : ''}>
                      {storageUsage.usage_percent.toFixed(1)}% Used
                    </span>
                  </div>
                  <Progress 
                    value={storageUsage.usage_percent} 
                    className={`h-2 ${
                      storageUsage.usage_percent > 95 
                        ? '[&>div]:bg-red-600' 
                        : storageUsage.usage_percent > 90 
                        ? '[&>div]:bg-yellow-600' 
                        : ''
                    }`}
                  />
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>{storageUsage.used_gb} GB / {storageUsage.total_gb} GB</span>
                    <span>{storageUsage.available_gb} GB Available</span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <p className="text-muted-foreground">Backups</p>
                    <p className="font-medium">{storageUsage.backup_count}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Latest Backup</p>
                    <p className="font-medium">
                      {storageUsage.latest_backup ? (
                        <span title={format(new Date(storageUsage.latest_backup), 'PPpp')}>
                          {formatDistanceToNow(new Date(storageUsage.latest_backup), { addSuffix: true })}
                        </span>
                      ) : (
                        'Never'
                      )}
                    </p>
                  </div>
                </div>

                {storageUsage.oldest_backup && storageUsage.latest_backup && (
                  <div className="text-sm">
                    <p className="text-muted-foreground">Retention Period</p>
                    <p className="font-medium">
                      <Calendar className="inline h-3 w-3 mr-1" />
                      {format(new Date(storageUsage.oldest_backup), 'MMM d, yyyy')} - {format(new Date(storageUsage.latest_backup), 'MMM d, yyyy')}
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}