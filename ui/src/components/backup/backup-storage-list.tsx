'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Loader2, Plus, HardDrive, AlertCircle, MoreHorizontal } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import { CreateBackupStorageDialog } from './create-backup-storage-dialog'
import { BackupStorageActionsMenu } from './backup-storage-actions-menu'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

interface BackupStorage {
  id: string
  name: string
  type: 'nfs' | 'ceph' | 'local' | 'proxmox'
  proxmox_storage_id?: string
  proxmox_node_id?: string
  capacity_gb: number
  used_gb: number
  status: 'provisioning' | 'active' | 'error' | 'deleting'
  error_message?: string
  created_at: string
  updated_at: string
}

interface BackupStorageListProps {
  orgId: string
  workspaceId: string
  workspacePlan: 'shared' | 'dedicated'
}

export function BackupStorageList({ orgId, workspaceId, workspacePlan }: BackupStorageListProps) {
  const { toast } = useToast()
  const [storages, setStorages] = useState<BackupStorage[]>([])
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  useEffect(() => {
    if (workspacePlan === 'dedicated') {
      fetchStorages()
    } else {
      setLoading(false)
    }
  }, [workspaceId, workspacePlan])

  const fetchStorages = async () => {
    try {
      setLoading(true)
      const response = await apiClient.backupApi.listBackupStorages(orgId, workspaceId)
      setStorages(response.data)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch backup storages',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleStorageCreated = () => {
    fetchStorages()
    setCreateDialogOpen(false)
  }

  const handleDeleteStorage = async (storageId: string) => {
    try {
      await apiClient.backupApi.deleteBackupStorage(orgId, workspaceId, storageId)
      toast({
        title: 'Success',
        description: 'Backup storage deleted successfully',
      })
      fetchStorages()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete backup storage',
        variant: 'destructive',
      })
    }
  }

  if (workspacePlan !== 'dedicated') {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Backup Storage</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <AlertCircle className="mx-auto h-12 w-12 text-muted-foreground" />
            <p className="mt-2 text-sm text-muted-foreground">
              Backup storage is only available for dedicated plan workspaces.
            </p>
            <Button className="mt-4" variant="outline">
              Upgrade to Dedicated Plan
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Backup Storage</CardTitle>
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
          <CardTitle>Backup Storage</CardTitle>
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Backup Storage
          </Button>
        </CardHeader>
        <CardContent>
          {storages.length === 0 ? (
            <div className="text-center py-8">
              <HardDrive className="mx-auto h-12 w-12 text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">
                No backup storage configured yet.
              </p>
              <Button
                className="mt-4"
                variant="outline"
                onClick={() => setCreateDialogOpen(true)}
              >
                Create your first backup storage
              </Button>
            </div>
          ) : (
            <div className="space-y-4">
              {storages.map((storage) => (
                <BackupStorageCard
                  key={storage.id}
                  storage={storage}
                  onDelete={() => handleDeleteStorage(storage.id)}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <CreateBackupStorageDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        orgId={orgId}
        workspaceId={workspaceId}
        onSuccess={handleStorageCreated}
      />
    </>
  )
}

function BackupStorageCard({ 
  storage, 
  onDelete 
}: { 
  storage: BackupStorage
  onDelete: () => void 
}) {
  const usagePercent = (storage.used_gb / storage.capacity_gb) * 100
  const isHighUsage = usagePercent > 90
  const isCriticalUsage = usagePercent > 95

  const getStatusBadgeVariant = (status: BackupStorage['status']) => {
    switch (status) {
      case 'active':
        return 'default'
      case 'provisioning':
        return 'secondary'
      case 'error':
        return 'destructive'
      case 'deleting':
        return 'outline'
      default:
        return 'default'
    }
  }

  const getStorageTypeLabel = (type: BackupStorage['type']) => {
    switch (type) {
      case 'nfs':
        return 'NFS'
      case 'ceph':
        return 'Ceph'
      case 'local':
        return 'Local'
      case 'proxmox':
        return 'Proxmox'
      default:
        return type.toUpperCase()
    }
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold">{storage.name}</h3>
              <Badge variant={getStatusBadgeVariant(storage.status)}>
                {storage.status}
              </Badge>
              <Badge variant="outline">
                {getStorageTypeLabel(storage.type)}
              </Badge>
            </div>
            {storage.error_message && (
              <p className="text-sm text-destructive">{storage.error_message}</p>
            )}
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem>View Details</DropdownMenuItem>
              <DropdownMenuItem>Edit</DropdownMenuItem>
              <DropdownMenuItem 
                className="text-destructive"
                onClick={onDelete}
              >
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="mt-4 space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Storage Usage</span>
            <span className={isCriticalUsage ? 'text-red-600 font-semibold' : ''}>
              {storage.used_gb} GB / {storage.capacity_gb} GB
            </span>
          </div>
          <Progress 
            value={usagePercent} 
            className={`h-2 ${isCriticalUsage ? '[&>div]:bg-red-600' : isHighUsage ? '[&>div]:bg-yellow-600' : ''}`}
          />
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>{usagePercent.toFixed(1)}% Used</span>
            <span>{storage.capacity_gb - storage.used_gb} GB Available</span>
          </div>
        </div>

        {storage.proxmox_node_id && (
          <div className="mt-3 text-sm text-muted-foreground">
            <span>Node: {storage.proxmox_node_id}</span>
            {storage.proxmox_storage_id && (
              <span className="ml-3">Storage: {storage.proxmox_storage_id}</span>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}