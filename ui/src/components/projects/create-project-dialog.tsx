'use client';

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { CreateProjectRequest } from '@/lib/api-client';

interface CreateProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateProjectRequest) => void | Promise<void>;
}

export function CreateProjectDialog({
  open,
  onOpenChange,
  onSubmit,
}: CreateProjectDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    description: '',
    namespace: '',
    resource_quota: {
      cpu: '2',
      memory: '4',
      storage: '10',
    },
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [namespaceManuallySet, setNamespaceManuallySet] = useState(false);

  const generateNamespace = (name: string) => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '');
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const name = e.target.value;
    setFormData(prev => ({
      ...prev,
      name,
      namespace: namespaceManuallySet ? prev.namespace : generateNamespace(name),
    }));
  };

  const handleNamespaceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const namespace = e.target.value;
    setFormData(prev => ({ ...prev, namespace }));
    setNamespaceManuallySet(true);
  };

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Project name is required';
    }

    if (formData.namespace && !/^[a-z0-9-]+$/.test(formData.namespace)) {
      newErrors.namespace = 'Namespace can only contain lowercase letters, numbers, and hyphens';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    try {
      await onSubmit({
        ...formData,
        resource_quota: {
          cpu: formData.resource_quota?.cpu || '2',
          memory: `${formData.resource_quota?.memory || '4'}Gi`,
          storage: `${formData.resource_quota?.storage || '10'}Gi`,
        },
      });
      handleClose();
    } catch (error) {
      // Error handling would be done by parent component
    } finally {
      setIsSubmitting(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      namespace: '',
      resource_quota: {
        cpu: '2',
        memory: '4',
        storage: '10',
      },
    });
    setErrors({});
    setNamespaceManuallySet(false);
  };

  const handleClose = () => {
    onOpenChange(false);
  };

  useEffect(() => {
    if (!open) {
      resetForm();
    }
  }, [open]);

  const applyResourceTemplate = (template: 'small' | 'medium' | 'large') => {
    const templates = {
      small: { cpu: '1', memory: '2', storage: '10' },
      medium: { cpu: '4', memory: '8', storage: '50' },
      large: { cpu: '8', memory: '16', storage: '100' },
    };

    setFormData(prev => ({
      ...prev,
      resource_quota: templates[template],
    }));
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Project</DialogTitle>
          <DialogDescription>
            Create a new project within this workspace.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <Label htmlFor="name">Project Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={handleNameChange}
                placeholder="My Project"
              />
              {errors.name && (
                <p className="text-sm text-destructive mt-1">{errors.name}</p>
              )}
            </div>

            <div>
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="Project description"
              />
            </div>

            <div>
              <Label htmlFor="namespace">Namespace</Label>
              <Input
                id="namespace"
                value={formData.namespace || ''}
                onChange={handleNamespaceChange}
                placeholder="my-project"
              />
              {errors.namespace && (
                <p className="text-sm text-destructive mt-1">{errors.namespace}</p>
              )}
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-medium">Resource Quotas</h4>
              
              <div className="flex gap-2 mb-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => applyResourceTemplate('small')}
                >
                  Small
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => applyResourceTemplate('medium')}
                >
                  Medium
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => applyResourceTemplate('large')}
                >
                  Large
                </Button>
              </div>

              <div className="grid grid-cols-3 gap-2">
                <div>
                  <Label htmlFor="cpu">CPU Limit</Label>
                  <Input
                    id="cpu"
                    type="number"
                    value={formData.resource_quota?.cpu || ''}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      resource_quota: {
                        ...prev.resource_quota!,
                        cpu: e.target.value,
                      },
                    }))}
                    placeholder="2"
                  />
                </div>
                <div>
                  <Label htmlFor="memory">Memory Limit (GB)</Label>
                  <Input
                    id="memory"
                    type="number"
                    value={formData.resource_quota?.memory || ''}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      resource_quota: {
                        ...prev.resource_quota!,
                        memory: e.target.value,
                      },
                    }))}
                    placeholder="4"
                  />
                </div>
                <div>
                  <Label htmlFor="storage">Storage Limit (GB)</Label>
                  <Input
                    id="storage"
                    type="number"
                    value={formData.resource_quota?.storage || ''}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      resource_quota: {
                        ...prev.resource_quota!,
                        storage: e.target.value,
                      },
                    }))}
                    placeholder="10"
                  />
                </div>
              </div>
            </div>
          </div>

          <DialogFooter className="mt-6">
            <Button 
              type="button" 
              variant="outline" 
              onClick={handleClose}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Creating...' : 'Create Project'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}