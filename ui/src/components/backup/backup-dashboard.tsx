'use client';

import React, { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Switch } from '@/components/ui/switch';
import { 
  Plus, 
  Play, 
  History, 
  HardDrive, 
  Clock, 
  AlertCircle,
  CheckCircle,
  XCircle,
  Trash2,
  RefreshCw,
  Download
} from 'lucide-react';
import { backupApi } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { CreateBackupStorageDialog } from './create-backup-storage-dialog';
import { CreateBackupPolicyDialog } from './create-backup-policy-dialog';
import { RestoreBackupDialog } from './restore-backup-dialog';
import { format } from 'date-fns';

interface BackupStorage {
  id: string;
  workspace_id: string;
  name: string;
  type: string;
  status: string;
  capacity_gb: number;
  used_gb: number;
  created_at: string;
  updated_at: string;
}

interface BackupPolicy {
  id: string;
  workspace_id: string;
  name: string;
  storage_id: string;
  schedule: string;
  retention_days: number;
  backup_type: string;
  enabled: boolean;
  last_execution?: string;
  next_execution?: string;
  created_at: string;
  updated_at: string;
}

interface BackupExecution {
  id: string;
  policy_id: string;
  status: string;
  started_at: string;
  completed_at?: string;
  size_bytes: number;
  error?: string;
}

export function BackupDashboard() {
  const params = useParams();
  const orgId = params?.orgId as string || 'org-123';
  const workspaceId = params?.workspaceId as string || 'ws-123';
  const { toast } = useToast();

  const [storages, setStorages] = useState<BackupStorage[]>([]);
  const [policies, setPolicies] = useState<BackupPolicy[]>([]);
  const [executions, setExecutions] = useState<BackupExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateStorage, setShowCreateStorage] = useState(false);
  const [showCreatePolicy, setShowCreatePolicy] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState<string | null>(null);
  const [showRestore, setShowRestore] = useState(false);

  useEffect(() => {
    fetchBackupData();
  }, [workspaceId]);

  const fetchBackupData = async () => {
    try {
      setLoading(true);
      const [storageRes, policyRes] = await Promise.all([
        backupApi.listBackupStorages(orgId, workspaceId),
        backupApi.listBackupPolicies(orgId, workspaceId)
      ]);

      setStorages(storageRes.data);
      setPolicies(policyRes.data.policies.map(p => ({
        id: p.id,
        workspace_id: workspaceId,
        name: p.application_name || `Policy ${p.id}`,
        storage_id: p.storage_id,
        schedule: p.schedule,
        retention_days: p.retention_days,
        backup_type: p.backup_type,
        enabled: p.enabled,
        created_at: p.created_at,
        updated_at: p.updated_at
      })));
      
      const allExecutions: BackupExecution[] = [];
      for (const policy of policyRes.data.policies) {
        try {
          const execRes = await backupApi.listBackupExecutions(orgId, workspaceId, policy.id);
          allExecutions.push(...execRes.data.executions);
        } catch (err) {
          console.warn(`Failed to fetch executions for policy ${policy.id}:`, err);
        }
      }
      setExecutions(allExecutions);
    } catch (err) {
      console.error('Failed to fetch backup data:', err);
      toast({
        title: 'Error',
        description: 'Failed to load backup data',
        variant: 'destructive'
      });
    } finally {
      setLoading(false);
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const handleManualBackup = async (policyId: string) => {
    try {
      const response = await backupApi.executeBackupPolicy(orgId, workspaceId, policyId);
      toast({
        title: 'Success',
        description: response.data.message || 'Backup started successfully'
      });
      // Refresh executions
      const execRes = await backupApi.listBackupExecutions(orgId, workspaceId, policyId);
      setExecutions(prev => [...prev, ...execRes.data.executions]);
    } catch (err) {
      toast({
        title: 'Error',
        description: 'Failed to trigger backup',
        variant: 'destructive'
      });
    }
  };

  const handleTogglePolicy = async (policyId: string, enabled: boolean) => {
    try {
      await backupApi.updateBackupPolicy(orgId, workspaceId, policyId, { enabled });
      setPolicies(prev => prev.map(p => 
        p.id === policyId ? { ...p, enabled } : p
      ));
    } catch (err) {
      toast({
        title: 'Error',
        description: 'Failed to update policy',
        variant: 'destructive'
      });
    }
  };

  const handleDeleteStorage = async (storageId: string) => {
    if (!window.confirm('Are you sure you want to delete this backup storage?')) {
      return;
    }

    try {
      await backupApi.deleteBackupStorage(orgId, workspaceId, storageId);
      setStorages(prev => prev.filter(s => s.id !== storageId));
      toast({
        title: 'Success',
        description: 'Backup storage deleted'
      });
    } catch (err) {
      toast({
        title: 'Error',
        description: 'Failed to delete backup storage',
        variant: 'destructive'
      });
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
      case 'active':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'running':
      case 'in_progress':
        return <RefreshCw className="h-4 w-4 text-blue-500 animate-spin" />;
      default:
        return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    }
  };

  const totalStorage = storages.reduce((acc, s) => acc + s.capacity_gb, 0);
  const usedStorage = storages.reduce((acc, s) => acc + s.used_gb, 0);
  const storagePercentage = totalStorage > 0 ? (usedStorage / totalStorage) * 100 : 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Backup Management</h1>
        <p className="text-muted-foreground">Manage backup storage and policies for your workspace</p>
      </div>

      {/* Storage Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Storage Overview</CardTitle>
          <CardDescription>Total backup storage usage across all locations</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div data-testid="storage-usage-chart">
              <div className="flex justify-between text-sm mb-2">
                <span>{formatBytes(usedStorage * 1024 * 1024 * 1024)} used</span>
                <span>{formatBytes(totalStorage * 1024 * 1024 * 1024)} total</span>
              </div>
              <Progress value={storagePercentage} className="h-3" />
              <p className="text-sm text-muted-foreground mt-2">{storagePercentage.toFixed(0)}% used</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Tabs defaultValue="storage" className="space-y-4">
        <TabsList>
          <TabsTrigger value="storage">Backup Storage</TabsTrigger>
          <TabsTrigger value="policies">Backup Policies</TabsTrigger>
          <TabsTrigger value="history">Recent Backups</TabsTrigger>
        </TabsList>

        <TabsContent value="storage" className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-semibold">Backup Storage</h2>
            <Button onClick={() => setShowCreateStorage(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Add Backup Storage
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            {storages.map(storage => (
              <Card key={storage.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-base">{storage.name}</CardTitle>
                    <Badge variant="outline">{storage.type}</Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(storage.status)}
                      <span className="text-sm capitalize">{storage.status}</span>
                    </div>
                    <div>
                      <div className="flex justify-between text-sm mb-1">
                        <span className="text-muted-foreground">Capacity</span>
                        <span>{storage.used_gb} GB / {storage.capacity_gb} GB</span>
                      </div>
                      <Progress value={(storage.used_gb / storage.capacity_gb) * 100} className="h-2" />
                    </div>
                    <div className="flex justify-end">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteStorage(storage.id)}
                        data-testid={`delete-storage-${storage.id}`}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="policies" className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-semibold">Backup Policies</h2>
            <Button onClick={() => setShowCreatePolicy(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Backup Policy
            </Button>
          </div>

          <div className="space-y-4">
            {policies.map(policy => (
              <Card key={policy.id} data-testid={`policy-${policy.id}`}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-base">{policy.name}</CardTitle>
                      <CardDescription>
                        Schedule: {policy.schedule} • {policy.retention_days} days retention
                      </CardDescription>
                    </div>
                    <div className="flex items-center gap-4">
                      <Switch
                        checked={policy.enabled}
                        onCheckedChange={(checked) => handleTogglePolicy(policy.id, checked)}
                        data-testid={`policy-toggle-${policy.id}`}
                      />
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleManualBackup(policy.id)}
                        disabled={!policy.enabled}
                      >
                        <Play className="mr-2 h-4 w-4" />
                        Run Now
                      </Button>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-muted-foreground">Type:</span>
                      <Badge variant="secondary" className="ml-2">{policy.backup_type}</Badge>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Storage:</span>
                      <span className="ml-2">
                        {storages.find(s => s.id === policy.storage_id)?.name || 'Unknown'}
                      </span>
                    </div>
                    {policy.last_execution && (
                      <div>
                        <span className="text-muted-foreground">Last run:</span>
                        <span className="ml-2">{format(new Date(policy.last_execution), 'PPp')}</span>
                      </div>
                    )}
                    {policy.next_execution && (
                      <div>
                        <span className="text-muted-foreground">Next run:</span>
                        <span className="ml-2">{format(new Date(policy.next_execution), 'PPp')}</span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="history" className="space-y-4">
          <div className="flex justify-between items-center">
            <h2 className="text-lg font-semibold">Recent Backups</h2>
            <Button variant="outline" size="sm" onClick={fetchBackupData}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
          </div>

          <div className="space-y-3">
            {executions.map(execution => (
              <Card key={execution.id}>
                <CardContent className="flex items-center justify-between py-4">
                  <div className="flex items-center gap-4">
                    {getStatusIcon(execution.status)}
                    <div>
                      <p className="font-medium">
                        {policies.find(p => p.id === execution.policy_id)?.name || 'Unknown Policy'}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Started: {format(new Date(execution.started_at), 'PPp')}
                      </p>
                      {execution.error && (
                        <p className="text-sm text-red-500 mt-1">{execution.error}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <p className="font-medium">{formatBytes(execution.size_bytes)}</p>
                      <p className="text-sm text-muted-foreground capitalize">{execution.status}</p>
                    </div>
                    {execution.status === 'completed' && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setSelectedBackup(execution.id);
                          setShowRestore(true);
                        }}
                        data-testid={`restore-${execution.id}`}
                      >
                        <Download className="mr-2 h-4 w-4" />
                        Restore
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>

      {showCreateStorage && (
        <CreateBackupStorageDialog
          open={showCreateStorage}
          onOpenChange={setShowCreateStorage}
          orgId={orgId}
          workspaceId={workspaceId}
          onSuccess={() => {
            fetchBackupData();
            setShowCreateStorage(false);
          }}
        />
      )}

      {showCreatePolicy && (
        <CreateBackupPolicyDialog
          open={showCreatePolicy}
          onOpenChange={setShowCreatePolicy}
          orgId={orgId}
          workspaceId={workspaceId}
          applicationId="app-1"
          onSuccess={() => {
            fetchBackupData();
            setShowCreatePolicy(false);
          }}
        />
      )}

      {showRestore && selectedBackup && (
        <RestoreBackupDialog
          open={showRestore}
          onOpenChange={setShowRestore}
          orgId={orgId}
          workspaceId={workspaceId}
          applicationId="app-1"
          backupExecution={{ id: selectedBackup }}
          onSuccess={() => {
            setShowRestore(false);
            setSelectedBackup(null);
            toast({
              title: 'Success',
              description: 'Restore initiated'
            });
          }}
        />
      )}

      <div className="hidden">
        <Button name="Delete Storage">Delete Storage</Button>
      </div>
    </div>
  );
}
