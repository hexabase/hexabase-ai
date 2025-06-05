'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { TaskMonitor } from '@/components/task-monitor';
import { vclusterApi } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { Loader2, Shield, Download, Upload, RefreshCw } from 'lucide-react';

interface WorkspaceOperationsProps {
  organizationId: string;
  workspaceId: string;
}

interface OperationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  operation: 'upgrade' | 'backup' | 'restore' | null;
  organizationId: string;
  workspaceId: string;
  onSuccess?: () => void;
}

function OperationDialog({ 
  open, 
  onOpenChange, 
  operation, 
  organizationId, 
  workspaceId,
  onSuccess 
}: OperationDialogProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    target_version: '',
    backup_name: '',
    retention: '7d',
    strategy: 'rolling',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      setLoading(true);
      let response;
      
      switch (operation) {
        case 'upgrade':
          if (!formData.target_version) {
            toast({
              title: 'Error',
              description: 'Please specify target version',
              variant: 'destructive',
            });
            return;
          }
          response = await vclusterApi.upgrade(organizationId, workspaceId, {
            target_version: formData.target_version,
            strategy: formData.strategy,
          });
          break;
          
        case 'backup':
          if (!formData.backup_name) {
            toast({
              title: 'Error',
              description: 'Please specify backup name',
              variant: 'destructive',
            });
            return;
          }
          response = await vclusterApi.backup(organizationId, workspaceId, {
            backup_name: formData.backup_name,
            retention: formData.retention,
          });
          break;
          
        case 'restore':
          if (!formData.backup_name) {
            toast({
              title: 'Error',
              description: 'Please select a backup to restore',
              variant: 'destructive',
            });
            return;
          }
          response = await vclusterApi.restore(organizationId, workspaceId, {
            backup_name: formData.backup_name,
            strategy: formData.strategy,
          });
          break;
      }
      
      if (response?.task_id) {
        setTaskId(response.task_id);
        toast({
          title: 'Operation Started',
          description: response.message || `${operation} operation initiated`,
        });
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: `Failed to start ${operation} operation`,
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const getDialogTitle = () => {
    switch (operation) {
      case 'upgrade':
        return 'Upgrade vCluster';
      case 'backup':
        return 'Backup Workspace';
      case 'restore':
        return 'Restore Workspace';
      default:
        return '';
    }
  };

  const getDialogDescription = () => {
    switch (operation) {
      case 'upgrade':
        return 'Upgrade your vCluster to a newer version with zero downtime';
      case 'backup':
        return 'Create a backup of your workspace configuration and data';
      case 'restore':
        return 'Restore your workspace from a previous backup';
      default:
        return '';
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        {!taskId ? (
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{getDialogTitle()}</DialogTitle>
              <DialogDescription>{getDialogDescription()}</DialogDescription>
            </DialogHeader>
            
            <div className="space-y-4 py-4">
              {operation === 'upgrade' && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="target_version">Target Version</Label>
                    <Input
                      id="target_version"
                      placeholder="e.g., v0.16.0"
                      value={formData.target_version}
                      onChange={(e) => setFormData({ ...formData, target_version: e.target.value })}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="strategy">Upgrade Strategy</Label>
                    <Select
                      value={formData.strategy}
                      onValueChange={(value) => setFormData({ ...formData, strategy: value })}
                    >
                      <SelectTrigger id="strategy">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="rolling">Rolling Update</SelectItem>
                        <SelectItem value="recreate">Recreate</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </>
              )}
              
              {operation === 'backup' && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="backup_name">Backup Name</Label>
                    <Input
                      id="backup_name"
                      placeholder="e.g., pre-upgrade-backup"
                      value={formData.backup_name}
                      onChange={(e) => setFormData({ ...formData, backup_name: e.target.value })}
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="retention">Retention Period</Label>
                    <Select
                      value={formData.retention}
                      onValueChange={(value) => setFormData({ ...formData, retention: value })}
                    >
                      <SelectTrigger id="retention">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1d">1 Day</SelectItem>
                        <SelectItem value="7d">7 Days</SelectItem>
                        <SelectItem value="30d">30 Days</SelectItem>
                        <SelectItem value="90d">90 Days</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </>
              )}
              
              {operation === 'restore' && (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="backup_select">Select Backup</Label>
                    <Select
                      value={formData.backup_name}
                      onValueChange={(value) => setFormData({ ...formData, backup_name: value })}
                    >
                      <SelectTrigger id="backup_select">
                        <SelectValue placeholder="Choose a backup to restore" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="backup-2024-01-15">backup-2024-01-15 (3 days ago)</SelectItem>
                        <SelectItem value="pre-upgrade-backup">pre-upgrade-backup (1 week ago)</SelectItem>
                        <SelectItem value="weekly-backup-2024-01">weekly-backup-2024-01 (2 weeks ago)</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="restore_strategy">Restore Strategy</Label>
                    <Select
                      value={formData.strategy}
                      onValueChange={(value) => setFormData({ ...formData, strategy: value })}
                    >
                      <SelectTrigger id="restore_strategy">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="replace">Replace Current</SelectItem>
                        <SelectItem value="merge">Merge with Current</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </>
              )}
            </div>
            
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={loading}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Starting...
                  </>
                ) : (
                  <>Start {operation}</>
                )}
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="py-4">
            <DialogHeader>
              <DialogTitle>Operation in Progress</DialogTitle>
              <DialogDescription>
                Your {operation} operation is being processed
              </DialogDescription>
            </DialogHeader>
            
            <div className="mt-4">
              <TaskMonitor
                taskId={taskId}
                organizationId={organizationId}
                onComplete={() => {
                  toast({
                    title: 'Success',
                    description: `${operation} completed successfully`,
                  });
                  onSuccess?.();
                  onOpenChange(false);
                  setTaskId(null);
                  setFormData({
                    target_version: '',
                    backup_name: '',
                    retention: '7d',
                    strategy: 'rolling',
                  });
                }}
                onError={(error) => {
                  toast({
                    title: 'Error',
                    description: error || `${operation} operation failed`,
                    variant: 'destructive',
                  });
                  setTaskId(null);
                }}
                showActions={true}
              />
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

export function WorkspaceOperations({ organizationId, workspaceId }: WorkspaceOperationsProps) {
  const [operationDialog, setOperationDialog] = useState<{
    open: boolean;
    operation: 'upgrade' | 'backup' | 'restore' | null;
  }>({ open: false, operation: null });

  const openDialog = (operation: 'upgrade' | 'backup' | 'restore') => {
    setOperationDialog({ open: true, operation });
  };

  return (
    <>
      <div className="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => openDialog('upgrade')}
        >
          <Shield className="w-4 h-4 mr-2" />
          Upgrade
        </Button>
        
        <Button
          variant="outline"
          size="sm"
          onClick={() => openDialog('backup')}
        >
          <Download className="w-4 h-4 mr-2" />
          Backup
        </Button>
        
        <Button
          variant="outline"
          size="sm"
          onClick={() => openDialog('restore')}
        >
          <Upload className="w-4 h-4 mr-2" />
          Restore
        </Button>
      </div>

      <OperationDialog
        open={operationDialog.open}
        onOpenChange={(open) => setOperationDialog({ ...operationDialog, open })}
        operation={operationDialog.operation}
        organizationId={organizationId}
        workspaceId={workspaceId}
        onSuccess={() => {
          // Optionally refresh workspace data
        }}
      />
    </>
  );
}