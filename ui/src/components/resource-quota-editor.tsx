'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent } from '@/components/ui/card';
import { projectsApi } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { Loader2, AlertCircle, Cpu, HardDrive, Database, Package } from 'lucide-react';

interface ResourceQuotaEditorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
  projectId: string;
  currentQuotas?: {
    cpu_limit: string;
    memory_limit: string;
    storage_limit: string;
    pod_limit?: string;
  };
  onSuccess?: () => void;
}

export function ResourceQuotaEditor({
  open,
  onOpenChange,
  organizationId,
  projectId,
  currentQuotas,
  onSuccess,
}: ResourceQuotaEditorProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    cpu_limit: currentQuotas?.cpu_limit?.replace(/[^0-9.]/g, '') || '4',
    memory_limit: currentQuotas?.memory_limit?.replace(/[^0-9.]/g, '') || '8',
    storage_limit: currentQuotas?.storage_limit?.replace(/[^0-9.]/g, '') || '100',
    pod_limit: currentQuotas?.pod_limit?.replace(/[^0-9]/g, '') || '50',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = () => {
    const newErrors: Record<string, string> = {};
    
    const cpuLimit = parseFloat(formData.cpu_limit);
    if (isNaN(cpuLimit) || cpuLimit < 1 || cpuLimit > 64) {
      newErrors.cpu_limit = 'CPU limit must be between 1 and 64 cores';
    }
    
    const memoryLimit = parseFloat(formData.memory_limit);
    if (isNaN(memoryLimit) || memoryLimit < 1 || memoryLimit > 256) {
      newErrors.memory_limit = 'Memory limit must be between 1 and 256 GB';
    }
    
    const storageLimit = parseFloat(formData.storage_limit);
    if (isNaN(storageLimit) || storageLimit < 10 || storageLimit > 10000) {
      newErrors.storage_limit = 'Storage limit must be between 10 and 10000 GB';
    }
    
    const podLimit = parseInt(formData.pod_limit);
    if (isNaN(podLimit) || podLimit < 1 || podLimit > 1000) {
      newErrors.pod_limit = 'Pod limit must be between 1 and 1000';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      
      const updateData = {
        resource_quotas: {
          cpu_limit: formData.cpu_limit,
          memory_limit: formData.memory_limit,
          storage_limit: formData.storage_limit,
          pod_limit: formData.pod_limit,
        },
      };
      
      await projectsApi.update(organizationId, projectId, updateData);
      
      toast({
        title: 'Success',
        description: 'Resource quotas updated successfully',
      });
      
      onSuccess?.();
      onOpenChange(false);
    } catch (error) {
      console.error('Failed to update resource quotas:', error);
      toast({
        title: 'Error',
        description: 'Failed to update resource quotas',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const calculateChange = (current: string, new_value: string): string => {
    const currentNum = parseFloat(current);
    const newNum = parseFloat(new_value);
    
    if (isNaN(currentNum) || isNaN(newNum)) return '';
    
    const diff = newNum - currentNum;
    const percentChange = ((diff / currentNum) * 100).toFixed(1);
    
    if (diff > 0) {
      return `+${diff} (+${percentChange}%)`;
    } else if (diff < 0) {
      return `${diff} (${percentChange}%)`;
    }
    return 'No change';
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Resource Quotas</DialogTitle>
            <DialogDescription>
              Update project-wide resource limits. Changes will apply to all namespaces.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 py-4">
            <Card className="bg-amber-50 border-amber-200">
              <CardContent className="pt-6">
                <div className="flex">
                  <AlertCircle className="h-5 w-5 text-amber-600 mr-3 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm text-amber-800 font-medium">Important</p>
                    <p className="text-sm text-amber-700 mt-1">
                      Reducing resource quotas may affect running workloads. Ensure your applications 
                      can operate within the new limits before applying changes.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            <div className="grid grid-cols-2 gap-6">
              {/* CPU Limit */}
              <div className="space-y-2">
                <Label htmlFor="cpu_limit" className="flex items-center gap-2">
                  <Cpu className="h-4 w-4" />
                  CPU Limit (cores)
                </Label>
                <Input
                  id="cpu_limit"
                  type="number"
                  step="1"
                  min="1"
                  max="64"
                  value={formData.cpu_limit}
                  onChange={(e) => setFormData({ ...formData, cpu_limit: e.target.value })}
                  className={errors.cpu_limit ? 'border-red-500' : ''}
                />
                {errors.cpu_limit && (
                  <p className="text-sm text-red-500">{errors.cpu_limit}</p>
                )}
                {currentQuotas && (
                  <p className="text-xs text-gray-500">
                    Current: {currentQuotas.cpu_limit} • Change: {calculateChange(
                      currentQuotas.cpu_limit.replace(/[^0-9.]/g, ''),
                      formData.cpu_limit
                    )}
                  </p>
                )}
              </div>

              {/* Memory Limit */}
              <div className="space-y-2">
                <Label htmlFor="memory_limit" className="flex items-center gap-2">
                  <Database className="h-4 w-4" />
                  Memory Limit (GB)
                </Label>
                <Input
                  id="memory_limit"
                  type="number"
                  step="1"
                  min="1"
                  max="256"
                  value={formData.memory_limit}
                  onChange={(e) => setFormData({ ...formData, memory_limit: e.target.value })}
                  className={errors.memory_limit ? 'border-red-500' : ''}
                />
                {errors.memory_limit && (
                  <p className="text-sm text-red-500">{errors.memory_limit}</p>
                )}
                {currentQuotas && (
                  <p className="text-xs text-gray-500">
                    Current: {currentQuotas.memory_limit} • Change: {calculateChange(
                      currentQuotas.memory_limit.replace(/[^0-9.]/g, ''),
                      formData.memory_limit
                    )}
                  </p>
                )}
              </div>

              {/* Storage Limit */}
              <div className="space-y-2">
                <Label htmlFor="storage_limit" className="flex items-center gap-2">
                  <HardDrive className="h-4 w-4" />
                  Storage Limit (GB)
                </Label>
                <Input
                  id="storage_limit"
                  type="number"
                  step="10"
                  min="10"
                  max="10000"
                  value={formData.storage_limit}
                  onChange={(e) => setFormData({ ...formData, storage_limit: e.target.value })}
                  className={errors.storage_limit ? 'border-red-500' : ''}
                />
                {errors.storage_limit && (
                  <p className="text-sm text-red-500">{errors.storage_limit}</p>
                )}
                {currentQuotas && (
                  <p className="text-xs text-gray-500">
                    Current: {currentQuotas.storage_limit} • Change: {calculateChange(
                      currentQuotas.storage_limit.replace(/[^0-9.]/g, ''),
                      formData.storage_limit
                    )}
                  </p>
                )}
              </div>

              {/* Pod Limit */}
              <div className="space-y-2">
                <Label htmlFor="pod_limit" className="flex items-center gap-2">
                  <Package className="h-4 w-4" />
                  Pod Limit
                </Label>
                <Input
                  id="pod_limit"
                  type="number"
                  step="10"
                  min="1"
                  max="1000"
                  value={formData.pod_limit}
                  onChange={(e) => setFormData({ ...formData, pod_limit: e.target.value })}
                  className={errors.pod_limit ? 'border-red-500' : ''}
                />
                {errors.pod_limit && (
                  <p className="text-sm text-red-500">{errors.pod_limit}</p>
                )}
                {currentQuotas && (
                  <p className="text-xs text-gray-500">
                    Current: {currentQuotas.pod_limit || 'Unlimited'} • Change: {
                      currentQuotas.pod_limit 
                        ? calculateChange(currentQuotas.pod_limit.replace(/[^0-9]/g, ''), formData.pod_limit)
                        : `Set to ${formData.pod_limit}`
                    }
                  </p>
                )}
              </div>
            </div>

            {/* Summary */}
            <Card className="bg-blue-50 border-blue-200">
              <CardContent className="pt-6">
                <h4 className="text-sm font-medium text-blue-900 mb-3">New Resource Limits Summary</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-blue-700">CPU:</span>
                    <span className="ml-2 font-medium text-blue-900">{formData.cpu_limit} cores</span>
                  </div>
                  <div>
                    <span className="text-blue-700">Memory:</span>
                    <span className="ml-2 font-medium text-blue-900">{formData.memory_limit} GB</span>
                  </div>
                  <div>
                    <span className="text-blue-700">Storage:</span>
                    <span className="ml-2 font-medium text-blue-900">{formData.storage_limit} GB</span>
                  </div>
                  <div>
                    <span className="text-blue-700">Pods:</span>
                    <span className="ml-2 font-medium text-blue-900">{formData.pod_limit} max</span>
                  </div>
                </div>
              </CardContent>
            </Card>
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
                  Updating...
                </>
              ) : (
                'Save Changes'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}