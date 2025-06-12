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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Loader2 } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { backupApi } from '@/lib/api-client'
import { TaskMonitor } from '@/components/task-monitor'

interface CreateBackupStorageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgId: string
  workspaceId: string
  onSuccess?: () => void
}

export function CreateBackupStorageDialog({
  open,
  onOpenChange,
  orgId,
  workspaceId,
  onSuccess,
}: CreateBackupStorageDialogProps) {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [taskId, setTaskId] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    name: '',
    type: 'proxmox',
    capacity_gb: 100,
  })

  const handleCreate = async () => {
    if (!formData.name.trim()) {
      toast({
        title: 'Error',
        description: 'Storage name is required',
        variant: 'destructive',
      })
      return
    }

    if (formData.capacity_gb < 10 || formData.capacity_gb > 10000) {
      toast({
        title: 'Error',
        description: 'Capacity must be between 10 GB and 10,000 GB',
        variant: 'destructive',
      })
      return
    }

    try {
      setLoading(true)
      const response = await backupApi.createBackupStorage(orgId, workspaceId, {
        name: formData.name,
        type: formData.type as 'nfs' | 'ceph' | 'local' | 'proxmox',
        capacity_gb: formData.capacity_gb,
      })

      if (response.data.task_id) {
        setTaskId(response.data.task_id)
      } else {
        toast({
          title: 'Success',
          description: 'Backup storage created successfully',
        })
        onSuccess?.()
        handleClose()
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create backup storage',
        variant: 'destructive',
      })
      setLoading(false)
    }
  }

  const handleClose = () => {
    setFormData({
      name: '',
      type: 'proxmox',
      capacity_gb: 100,
    })
    setTaskId(null)
    setLoading(false)
    onOpenChange(false)
  }

  const handleTaskComplete = () => {
    toast({
      title: 'Success',
      description: 'Backup storage created successfully',
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Backup Storage</DialogTitle>
          <DialogDescription>
            Configure a new backup storage for your workspace. Storage will be provisioned
            on your dedicated nodes.
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
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Storage Name</Label>
                <Input
                  id="name"
                  name="name"
                  placeholder="production-backups"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  disabled={loading}
                />
              </div>

              <div className="grid gap-2">
                <Label htmlFor="type">Storage Type</Label>
                <Select
                  name="type"
                  value={formData.type}
                  onValueChange={(value) => setFormData({ ...formData, type: value as any })}
                  disabled={loading}
                >
                  <SelectTrigger id="type">
                    <SelectValue placeholder="Select storage type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="proxmox">Proxmox</SelectItem>
                    <SelectItem value="nfs">NFS</SelectItem>
                    <SelectItem value="ceph">Ceph</SelectItem>
                    <SelectItem value="local">Local</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="capacity_gb">Capacity (GB)</Label>
                <Input
                  id="capacity_gb"
                  name="capacity_gb"
                  type="number"
                  min="10"
                  max="10000"
                  step="10"
                  value={formData.capacity_gb}
                  onChange={(e) => setFormData({ ...formData, capacity_gb: parseInt(e.target.value) || 100 })}
                  disabled={loading}
                />
                <p className="text-xs text-muted-foreground">
                  Minimum: 10 GB, Maximum: 10,000 GB
                </p>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={handleClose} disabled={loading}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={loading}>
                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
