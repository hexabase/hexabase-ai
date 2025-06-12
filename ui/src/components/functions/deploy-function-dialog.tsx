'use client';

import { useState, useEffect } from 'react';
import { FunctionConfig, FunctionVersion, functionsApi } from '@/lib/api-client';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Upload,
  Plus,
  Trash2,
  Clock,
  Package,
  AlertCircle,
  RotateCcw,
} from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { format } from 'date-fns';

interface DeployFunctionDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  functionData: FunctionConfig;
  orgId: string;
  workspaceId: string;
}

export function DeployFunctionDialog({
  open,
  onClose,
  onSuccess,
  functionData,
  orgId,
  workspaceId,
}: DeployFunctionDialogProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [versions, setVersions] = useState<FunctionVersion[]>([]);
  const [sourceFile, setSourceFile] = useState<File | null>(null);
  const [versionTag, setVersionTag] = useState('');
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [rollbackVersion, setRollbackVersion] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      fetchVersions();
      // Initialize with existing env vars
      if (functionData.environment_vars) {
        setEnvVars(
          Object.entries(functionData.environment_vars).map(([key, value]) => ({
            key,
            value,
          }))
        );
      }
    }
  }, [open]);

  const fetchVersions = async () => {
    try {
      setVersionsLoading(true);
      const response = await functionsApi.getVersions(orgId, workspaceId, functionData.id);
      setVersions(response.data.versions);
    } catch (error) {
      console.error('Failed to fetch versions:', error);
    } finally {
      setVersionsLoading(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSourceFile(e.target.files[0]);
    }
  };

  const addEnvVar = () => {
    setEnvVars([...envVars, { key: '', value: '' }]);
  };

  const removeEnvVar = (index: number) => {
    setEnvVars(envVars.filter((_, i) => i !== index));
  };

  const updateEnvVar = (index: number, field: 'key' | 'value', value: string) => {
    const updated = [...envVars];
    updated[index][field] = value;
    setEnvVars(updated);
  };

  const handleDeploy = async () => {
    setError(null);

    if (!sourceFile && !rollbackVersion) {
      setError('Please upload source code');
      return;
    }

    try {
      setLoading(true);
      const envVarsObject = envVars.reduce((acc, { key, value }) => {
        if (key) acc[key] = value;
        return acc;
      }, {} as Record<string, string>);

      await functionsApi.deploy(orgId, workspaceId, functionData.id, {
        version: versionTag || undefined,
        source: sourceFile || undefined,
        environment_vars: Object.keys(envVarsObject).length > 0 ? envVarsObject : undefined,
        rollback_to: rollbackVersion || undefined,
      });

      onSuccess();
      onClose();
    } catch (error) {
      setError('Deployment failed');
      console.error('Deploy error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRollback = (version: string) => {
    setRollbackVersion(version);
  };

  const confirmRollback = async () => {
    if (!rollbackVersion) return;

    try {
      setLoading(true);
      await functionsApi.deploy(orgId, workspaceId, functionData.id, {
        rollback_to: rollbackVersion,
      });
      onSuccess();
      onClose();
    } catch (error) {
      setError('Rollback failed');
      console.error('Rollback error:', error);
    } finally {
      setLoading(false);
      setRollbackVersion(null);
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={onClose}>
        <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Deploy Function</DialogTitle>
            <DialogDescription>
              Deploy a new version of {functionData.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            <div className="p-4 bg-muted rounded-md">
              <p className="text-sm font-medium">{functionData.name}</p>
              <p className="text-sm text-muted-foreground">Current version: {functionData.version}</p>
            </div>

            {/* Version History */}
            <div>
              <h3 className="text-sm font-medium mb-3">Version History</h3>
              {versionsLoading ? (
                <div className="space-y-2">
                  {[1, 2].map((i) => (
                    <Skeleton key={i} className="h-16 w-full" />
                  ))}
                </div>
              ) : versions.length > 0 ? (
                <div className="space-y-2">
                  {versions.map((version) => (
                    <div
                      key={version.version}
                      className="flex items-center justify-between p-3 border rounded-md"
                    >
                      <div className="flex items-center gap-3">
                        <Package className="h-4 w-4 text-muted-foreground" />
                        <div>
                          <p className="text-sm font-medium">{version.version}</p>
                          <p className="text-xs text-muted-foreground">
                            {format(new Date(version.deployed_at), 'MMM d, yyyy HH:mm')}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant={version.status === 'active' ? 'default' : 'secondary'}>
                          {version.status}
                        </Badge>
                        {version.status === 'inactive' && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleRollback(version.version)}
                            data-testid={`rollback-${version.version}`}
                          >
                            <RotateCcw className="h-3 w-3 mr-1" />
                            Rollback
                          </Button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No previous versions</p>
              )}
            </div>

            {/* Source Code Upload */}
            <div>
              <Label htmlFor="source">Source Code</Label>
              <div className="mt-1">
                <Input
                  id="source"
                  type="file"
                  accept=".zip,.tar,.tar.gz"
                  onChange={handleFileChange}
                />
                {sourceFile && (
                  <p className="text-sm text-muted-foreground mt-1">
                    Selected: {sourceFile.name}
                  </p>
                )}
              </div>
            </div>

            {/* Version Tag */}
            <div>
              <Label htmlFor="version">Version Tag (optional)</Label>
              <Input
                id="version"
                value={versionTag}
                onChange={(e) => setVersionTag(e.target.value)}
                placeholder="e.g., v1.3.0"
              />
            </div>

            {/* Environment Variables */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <Label>Environment Variables</Label>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={addEnvVar}
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add Variable
                </Button>
              </div>
              <div className="space-y-2">
                {envVars.map((envVar, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      placeholder="Key"
                      value={envVar.key}
                      onChange={(e) => updateEnvVar(index, 'key', e.target.value)}
                    />
                    <Input
                      placeholder="Value"
                      value={envVar.value}
                      onChange={(e) => updateEnvVar(index, 'value', e.target.value)}
                    />
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => removeEnvVar(index)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            </div>

            {error && (
              <div className="flex items-center gap-2 p-3 bg-destructive/10 text-destructive rounded-md">
                <AlertCircle className="h-4 w-4" />
                <p className="text-sm">{error}</p>
              </div>
            )}

            <div className="flex justify-end gap-2 pt-4">
              <Button variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button onClick={handleDeploy} disabled={loading}>
                {loading ? (
                  <>
                    <Upload className="h-4 w-4 mr-2 animate-spin" />
                    Deploying...
                  </>
                ) : (
                  <>
                    <Upload className="h-4 w-4 mr-2" />
                    Deploy
                  </>
                )}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!rollbackVersion} onOpenChange={() => setRollbackVersion(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rollback to {rollbackVersion}?</AlertDialogTitle>
            <AlertDialogDescription>
              This will activate version {rollbackVersion} and deactivate the current version.
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={confirmRollback}>
              Confirm
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}