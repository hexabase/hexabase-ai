'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Loader2, Info } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { backupApi } from '@/lib/api-client'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

interface BackupStorage {
  id: string
  name: string
  type: string
  capacity_gb: number
  used_gb: number
  status: string
}

interface CreateBackupPolicyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgId: string
  workspaceId: string
  applicationId: string
  existingPolicy?: any
  onSuccess?: () => void
}

export function CreateBackupPolicyDialog({
  open,
  onOpenChange,
  orgId,
  workspaceId,
  applicationId,
  existingPolicy,
  onSuccess,
}: CreateBackupPolicyDialogProps) {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [storages, setStorages] = useState<BackupStorage[]>([])
  const [formData, setFormData] = useState({
    storage_id: existingPolicy?.storage_id || '',
    schedule: existingPolicy?.schedule || '0 2 * * *',
    retention_days: existingPolicy?.retention_days || 30,
    backup_type: existingPolicy?.backup_type || 'full',
    include_volumes: existingPolicy?.include_volumes ?? true,
    include_database: existingPolicy?.include_database ?? true,
    include_config: existingPolicy?.include_config ?? true,
    compression_enabled: existingPolicy?.compression_enabled ?? true,
    encryption_enabled: existingPolicy?.encryption_enabled ?? true,
    enabled: existingPolicy?.enabled ?? true,
  })

  useEffect(() => {
    if (open) {
      fetchStorages()
    }
  }, [open, workspaceId])

  const fetchStorages = async () => {
    try {
      const response = await backupApi.listBackupStorages(orgId, workspaceId)
      setStorages(response.data.filter((s: BackupStorage) => s.status === 'active'))
      
      // Auto-select first storage if none selected
      if (!formData.storage_id && response.data.length > 0) {
        setFormData(prev => ({ ...prev, storage_id: response.data[0].id }))
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch backup storages',
        variant: 'destructive',
      })
    }
  }

  const handleSubmit = async () => {
    if (!formData.storage_id) {
      toast({
        title: 'Error',
        description: 'Please select a backup storage',
        variant: 'destructive',
      })
      return
    }

    try {
      setLoading(true)
      
      if (existingPolicy) {
        await backupApi.updateBackupPolicy(orgId, workspaceId, existingPolicy.id, formData)
        toast({
          title: 'Success',
          description: 'Backup policy updated successfully',
        })
      } else {
        await backupApi.createBackupPolicy(orgId, workspaceId, applicationId, formData)
        toast({
          title: 'Success',
          description: 'Backup policy created successfully',
        })
      }
      
      onSuccess?.()
      onOpenChange(false)
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to save backup policy',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const cronExamples = [
    { label: 'Daily at 2 AM', value: '0 2 * * *' },
    { label: 'Every 6 hours', value: '0 */6 * * *' },
    { label: 'Weekly on Sunday', value: '0 0 * * 0' },
    { label: 'Monthly on 1st', value: '0 0 1 * *' },
  ]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>
            {existingPolicy ? 'Edit Backup Policy' : 'Create Backup Policy'}
          </DialogTitle>
          <DialogDescription>
            Configure automated backups for your application. Backups will run according
            to the schedule you define.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="storage_id">Backup Storage</Label>
            <Select
              name="storage_id"
              value={formData.storage_id}
              onValueChange={(value) => setFormData({ ...formData, storage_id: value })}
              disabled={loading}
            >
              <SelectTrigger id="storage_id">
                <SelectValue placeholder="Select backup storage" />
              </SelectTrigger>
              <SelectContent>
                {storages.map((storage) => (
                  <SelectItem key={storage.id} value={storage.id}>
                    <div className="flex items-center justify-between w-full">
                      <span>{storage.name}</span>
                      <span className="text-xs text-muted-foreground ml-2">
                        {storage.used_gb}/{storage.capacity_gb} GB
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {storages.length === 0 && (
              <p className="text-xs text-muted-foreground">
                No active backup storage found. Please create one first.
              </p>
            )}
          </div>

          <div className="grid gap-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="schedule">Backup Schedule (Cron Expression)</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="h-4 w-4 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-xs">
                    <p>Use standard cron syntax: minute hour day month weekday</p>
                    <p className="mt-1 text-xs">Examples:</p>
                    {cronExamples.map((ex) => (
                      <p key={ex.value} className="text-xs">
                        {ex.value} = {ex.label}
                      </p>
                    ))}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <div className="flex gap-2">
              <Input
                id="schedule"
                name="schedule"
                value={formData.schedule}
                onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
                disabled={loading}
                placeholder="0 2 * * *"
              />
              <Select
                value={formData.schedule}
                onValueChange={(value) => setFormData({ ...formData, schedule: value })}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {cronExamples.map((ex) => (
                    <SelectItem key={ex.value} value={ex.value}>
                      {ex.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="retention_days">Retention Period (Days)</Label>
            <Input
              id="retention_days"
              name="retention_days"
              type="number"
              min="1"
              max="365"
              value={formData.retention_days}
              onChange={(e) => setFormData({ ...formData, retention_days: parseInt(e.target.value) || 30 })}
              disabled={loading}
            />
            <p className="text-xs text-muted-foreground">
              Backups older than this will be automatically deleted
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="backup_type">Backup Type</Label>
            <Select
              name="backup_type"
              value={formData.backup_type}
              onValueChange={(value) => setFormData({ ...formData, backup_type: value as any })}
              disabled={loading}
            >
              <SelectTrigger id="backup_type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="full">Full Backup</SelectItem>
                <SelectItem value="incremental">Incremental Backup</SelectItem>
                <SelectItem value="application">Application Backup</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Backup Components</Label>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="include_volumes"
                  name="include_volumes"
                  checked={formData.include_volumes}
                  onCheckedChange={(checked) => 
                    setFormData({ ...formData, include_volumes: checked as boolean })
                  }
                  disabled={loading}
                />
                <label
                  htmlFor="include_volumes"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Include persistent volumes
                </label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="include_database"
                  name="include_database"
                  checked={formData.include_database}
                  onCheckedChange={(checked) => 
                    setFormData({ ...formData, include_database: checked as boolean })
                  }
                  disabled={loading}
                />
                <label
                  htmlFor="include_database"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Include database dumps
                </label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="include_config"
                  checked={formData.include_config}
                  onCheckedChange={(checked) => 
                    setFormData({ ...formData, include_config: checked as boolean })
                  }
                  disabled={loading}
                />
                <label
                  htmlFor="include_config"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Include configuration files
                </label>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Backup Options</Label>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="compression_enabled"
                  name="compression_enabled"
                  checked={formData.compression_enabled}
                  onCheckedChange={(checked) => 
                    setFormData({ ...formData, compression_enabled: checked as boolean })
                  }
                  disabled={loading}
                />
                <label
                  htmlFor="compression_enabled"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Enable compression
                </label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="encryption_enabled"
                  name="encryption_enabled"
                  checked={formData.encryption_enabled}
                  onCheckedChange={(checked) => 
                    setFormData({ ...formData, encryption_enabled: checked as boolean })
                  }
                  disabled={loading}
                />
                <label
                  htmlFor="encryption_enabled"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Enable encryption
                </label>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={loading || storages.length === 0}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {existingPolicy ? 'Update Policy' : 'Create Policy'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
