'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { projectsApi, type CreateProjectRequest } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { Loader2, Folder, Settings, Database } from 'lucide-react';

interface CreateProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
  workspaceId: string;
  onSuccess?: () => void;
}

export function CreateProjectDialog({
  open,
  onOpenChange,
  organizationId,
  workspaceId,
  onSuccess,
}: CreateProjectDialogProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(1);
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    description: '',
    workspace_id: workspaceId,
    namespace_name: '',
    resource_quotas: {
      cpu_limit: '4',
      memory_limit: '8',
      storage_limit: '100',
      pod_limit: '50',
    },
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateStep1 = () => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.name.trim()) {
      newErrors.name = 'Project name is required';
    } else if (formData.name.length < 3) {
      newErrors.name = 'Project name must be at least 3 characters';
    } else if (!/^[a-zA-Z0-9-_\s]+$/.test(formData.name)) {
      newErrors.name = 'Project name can only contain letters, numbers, spaces, hyphens, and underscores';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validateStep2 = () => {
    const newErrors: Record<string, string> = {};
    
    if (formData.namespace_name && !/^[a-z0-9-]+$/.test(formData.namespace_name)) {
      newErrors.namespace_name = 'Namespace name can only contain lowercase letters, numbers, and hyphens';
    }
    
    const cpuLimit = parseInt(formData.resource_quotas?.cpu_limit || '0');
    if (cpuLimit < 1 || cpuLimit > 64) {
      newErrors.cpu_limit = 'CPU limit must be between 1 and 64 cores';
    }
    
    const memoryLimit = parseInt(formData.resource_quotas?.memory_limit || '0');
    if (memoryLimit < 1 || memoryLimit > 256) {
      newErrors.memory_limit = 'Memory limit must be between 1 and 256 GB';
    }
    
    const storageLimit = parseInt(formData.resource_quotas?.storage_limit || '0');
    if (storageLimit < 1 || storageLimit > 1000) {
      newErrors.storage_limit = 'Storage limit must be between 1 and 1000 GB';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = () => {
    if (step === 1 && validateStep1()) {
      setStep(2);
    }
  };

  const handleBack = () => {
    setStep(1);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateStep2()) {
      return;
    }

    try {
      setLoading(true);
      
      // Generate namespace name if not provided
      const projectData = {
        ...formData,
        namespace_name: formData.namespace_name || formData.name.toLowerCase().replace(/\s+/g, '-'),
      };
      
      await projectsApi.create(organizationId, projectData);
      
      toast({
        title: 'Success',
        description: 'Project created successfully',
      });
      
      onSuccess?.();
      onOpenChange(false);
      
      // Reset form
      setFormData({
        name: '',
        description: '',
        workspace_id: workspaceId,
        namespace_name: '',
        resource_quotas: {
          cpu_limit: '4',
          memory_limit: '8',
          storage_limit: '100',
          pod_limit: '50',
        },
      });
      setStep(1);
      setErrors({});
    } catch (error) {
      console.error('Failed to create project:', error);
      toast({
        title: 'Error',
        description: 'Failed to create project. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]" data-testid="create-project-modal">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Project</DialogTitle>
            <DialogDescription>
              {step === 1 ? 'Set up a new project to organize your namespaces and resources' : 'Configure namespace and resource quotas'}
            </DialogDescription>
          </DialogHeader>

          <div className="py-6">
            {step === 1 ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Project Name</Label>
                  <Input
                    id="name"
                    name="name"
                    placeholder="e.g., Frontend App, Backend API"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    data-testid="project-name-input"
                    className={errors.name ? 'border-red-500' : ''}
                  />
                  {errors.name && (
                    <p className="text-sm text-red-500" data-testid="name-error">
                      {errors.name}
                    </p>
                  )}
                </div>

                <div className="space-y-2">
                  <Label htmlFor="description">Description (Optional)</Label>
                  <Textarea
                    id="description"
                    name="description"
                    placeholder="Describe the purpose of this project"
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    rows={3}
                    data-testid="project-description-input"
                  />
                </div>

                <Card className="bg-blue-50 border-blue-200">
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm flex items-center">
                      <Folder className="h-4 w-4 mr-2" />
                      Project Organization
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-gray-600">
                      Projects help you organize related namespaces and resources within a workspace. 
                      Each project can have multiple namespaces with their own resource quotas.
                    </p>
                  </CardContent>
                </Card>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="namespace_name">Default Namespace Name (Optional)</Label>
                  <Input
                    id="namespace_name"
                    name="namespace_name"
                    placeholder={formData.name.toLowerCase().replace(/\s+/g, '-')}
                    value={formData.namespace_name}
                    onChange={(e) => setFormData({ ...formData, namespace_name: e.target.value })}
                    className={errors.namespace_name ? 'border-red-500' : ''}
                  />
                  {errors.namespace_name && (
                    <p className="text-sm text-red-500">
                      {errors.namespace_name}
                    </p>
                  )}
                  <p className="text-xs text-gray-500">
                    A namespace will be created automatically. Leave empty to use project name.
                  </p>
                </div>

                <div className="space-y-4">
                  <h3 className="text-sm font-medium flex items-center">
                    <Settings className="h-4 w-4 mr-2" />
                    Resource Quotas
                  </h3>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="cpu_limit">CPU Limit (cores)</Label>
                      <Input
                        id="cpu_limit"
                        name="cpu_limit"
                        type="number"
                        min="1"
                        max="64"
                        value={formData.resource_quotas?.cpu_limit}
                        onChange={(e) => setFormData({
                          ...formData,
                          resource_quotas: {
                            ...formData.resource_quotas!,
                            cpu_limit: e.target.value,
                          },
                        })}
                        className={errors.cpu_limit ? 'border-red-500' : ''}
                      />
                      {errors.cpu_limit && (
                        <p className="text-xs text-red-500">{errors.cpu_limit}</p>
                      )}
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="memory_limit">Memory Limit (GB)</Label>
                      <Input
                        id="memory_limit"
                        name="memory_limit"
                        type="number"
                        min="1"
                        max="256"
                        value={formData.resource_quotas?.memory_limit}
                        onChange={(e) => setFormData({
                          ...formData,
                          resource_quotas: {
                            ...formData.resource_quotas!,
                            memory_limit: e.target.value,
                          },
                        })}
                        className={errors.memory_limit ? 'border-red-500' : ''}
                      />
                      {errors.memory_limit && (
                        <p className="text-xs text-red-500">{errors.memory_limit}</p>
                      )}
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="storage_limit">Storage Limit (GB)</Label>
                      <Input
                        id="storage_limit"
                        name="storage_limit"
                        type="number"
                        min="1"
                        max="1000"
                        value={formData.resource_quotas?.storage_limit}
                        onChange={(e) => setFormData({
                          ...formData,
                          resource_quotas: {
                            ...formData.resource_quotas!,
                            storage_limit: e.target.value,
                          },
                        })}
                        className={errors.storage_limit ? 'border-red-500' : ''}
                      />
                      {errors.storage_limit && (
                        <p className="text-xs text-red-500">{errors.storage_limit}</p>
                      )}
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="pod_limit">Pod Limit</Label>
                      <Input
                        id="pod_limit"
                        name="pod_limit"
                        type="number"
                        min="1"
                        max="1000"
                        value={formData.resource_quotas?.pod_limit}
                        onChange={(e) => setFormData({
                          ...formData,
                          resource_quotas: {
                            ...formData.resource_quotas!,
                            pod_limit: e.target.value,
                          },
                        })}
                      />
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            {step === 2 && (
              <Button
                type="button"
                variant="outline"
                onClick={handleBack}
                disabled={loading}
              >
                Back
              </Button>
            )}
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            {step === 1 ? (
              <Button type="button" onClick={handleNext}>
                Next
              </Button>
            ) : (
              <Button type="submit" disabled={loading} data-testid="submit-project">
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  'Create Project'
                )}
              </Button>
            )}
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}