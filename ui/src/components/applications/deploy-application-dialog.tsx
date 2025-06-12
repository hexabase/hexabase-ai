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
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CreateApplicationRequest } from '@/lib/api-client';
import { Card, CardContent } from '@/components/ui/card';
import { Server, Package, Clock, GitBranch, Image, FileCode } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DeployApplicationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateApplicationRequest) => void | Promise<void>;
}

export function DeployApplicationDialog({
  open,
  onOpenChange,
  onSubmit,
}: DeployApplicationDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formData, setFormData] = useState<CreateApplicationRequest>({
    name: '',
    type: 'stateless',
    source_type: 'image',
    source_image: '',
    source_git_url: '',
    source_git_ref: '',
    config: {
      replicas: 1,
      cpu: '100m',
      memory: '128Mi',
    },
    cron_schedule: '',
    function_runtime: '',
    function_handler: '',
    function_timeout: 30,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'stateless',
      source_type: 'image',
      source_image: '',
      source_git_url: '',
      source_git_ref: '',
      config: {
        replicas: 1,
        cpu: '100m',
        memory: '128Mi',
      },
      cron_schedule: '',
      function_runtime: '',
      function_handler: '',
      function_timeout: 30,
    });
    setErrors({});
  };

  useEffect(() => {
    if (!open) {
      resetForm();
    }
  }, [open]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Application name is required';
    }

    if (formData.source_type === 'image' && !formData.source_image) {
      newErrors.source_image = 'Image name is required';
    }

    if (formData.source_type === 'git' && !formData.source_git_url) {
      newErrors.source_git_url = 'Repository URL is required';
    }

    if (formData.type === 'cronjob' && !formData.cron_schedule) {
      newErrors.cron_schedule = 'Cron schedule is required';
    }

    if (formData.type === 'function') {
      if (!formData.function_runtime) {
        newErrors.function_runtime = 'Runtime is required';
      }
      if (!formData.function_handler) {
        newErrors.function_handler = 'Handler is required';
      }
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
      await onSubmit(formData);
      onOpenChange(false);
    } catch (error) {
      // Error handling would be done by parent component
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Deploy New Application</DialogTitle>
          <DialogDescription>
            Choose your application type and configure deployment settings.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="space-y-6">
            {/* Application Name */}
            <div>
              <Label htmlFor="name">Application Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="my-app"
              />
              {errors.name && (
                <p className="text-sm text-destructive mt-1">{errors.name}</p>
              )}
            </div>

            {/* Application Type */}
            <div className="space-y-3">
              <Label>Application Type</Label>
              <div className="grid grid-cols-2 gap-3">
                <Card
                  className={cn(
                    "cursor-pointer transition-colors",
                    formData.type === 'stateless' && "border-primary"
                  )}
                  onClick={() => setFormData({ ...formData, type: 'stateless' })}
                >
                  <CardContent className="flex items-center gap-3 p-4">
                    <Server className="h-5 w-5" />
                    <div>
                      <p className="font-medium">Stateless Application</p>
                      <p className="text-xs text-muted-foreground">Web apps, APIs</p>
                    </div>
                  </CardContent>
                </Card>

                <Card
                  className={cn(
                    "cursor-pointer transition-colors",
                    formData.type === 'stateful' && "border-primary"
                  )}
                  onClick={() => setFormData({ ...formData, type: 'stateful' })}
                >
                  <CardContent className="flex items-center gap-3 p-4">
                    <Package className="h-5 w-5" />
                    <div>
                      <p className="font-medium">Stateful Application</p>
                      <p className="text-xs text-muted-foreground">Databases, caches</p>
                    </div>
                  </CardContent>
                </Card>

                <Card
                  className={cn(
                    "cursor-pointer transition-colors",
                    formData.type === 'cronjob' && "border-primary"
                  )}
                  onClick={() => setFormData({ ...formData, type: 'cronjob' })}
                >
                  <CardContent className="flex items-center gap-3 p-4">
                    <Clock className="h-5 w-5" />
                    <div>
                      <p className="font-medium">CronJob</p>
                      <p className="text-xs text-muted-foreground">Scheduled tasks</p>
                    </div>
                  </CardContent>
                </Card>

                <Card
                  className={cn(
                    "cursor-pointer transition-colors",
                    formData.type === 'function' && "border-primary"
                  )}
                  onClick={() => setFormData({ ...formData, type: 'function' })}
                >
                  <CardContent className="flex items-center gap-3 p-4">
                    <GitBranch className="h-5 w-5" />
                    <div>
                      <p className="font-medium">Serverless Function</p>
                      <p className="text-xs text-muted-foreground">Event-driven code</p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>

            {/* Source Type */}
            <div className="space-y-3">
              <Label>Source</Label>
              <RadioGroup
                value={formData.source_type}
                onValueChange={(value: any) => setFormData({ ...formData, source_type: value })}
              >
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="image" id="image" />
                  <Label htmlFor="image" className="flex items-center gap-2 cursor-pointer">
                    <Image className="h-4 w-4" />
                    Container Image
                  </Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="git" id="git" />
                  <Label htmlFor="git" className="flex items-center gap-2 cursor-pointer">
                    <GitBranch className="h-4 w-4" />
                    Git Repository
                  </Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="buildpack" id="buildpack" />
                  <Label htmlFor="buildpack" className="flex items-center gap-2 cursor-pointer">
                    <FileCode className="h-4 w-4" />
                    Buildpack
                  </Label>
                </div>
              </RadioGroup>
            </div>

            {/* Source Configuration */}
            {formData.source_type === 'image' && (
              <div>
                <Label htmlFor="image-name">Image Name</Label>
                <Input
                  id="image-name"
                  value={formData.source_image || ''}
                  onChange={(e) => setFormData({ ...formData, source_image: e.target.value })}
                  placeholder="nginx:latest"
                />
                {errors.source_image && (
                  <p className="text-sm text-destructive mt-1">{errors.source_image}</p>
                )}
              </div>
            )}

            {formData.source_type === 'git' && (
              <div className="space-y-3">
                <div>
                  <Label htmlFor="git-url">Repository URL</Label>
                  <Input
                    id="git-url"
                    value={formData.source_git_url || ''}
                    onChange={(e) => setFormData({ ...formData, source_git_url: e.target.value })}
                    placeholder="https://github.com/user/repo"
                  />
                  {errors.source_git_url && (
                    <p className="text-sm text-destructive mt-1">{errors.source_git_url}</p>
                  )}
                </div>
                <div>
                  <Label htmlFor="git-ref">Branch/Tag</Label>
                  <Input
                    id="git-ref"
                    value={formData.source_git_ref || ''}
                    onChange={(e) => setFormData({ ...formData, source_git_ref: e.target.value })}
                    placeholder="main"
                  />
                </div>
              </div>
            )}

            {/* Type-specific Configuration */}
            {(formData.type === 'stateless' || formData.type === 'stateful') && (
              <div className="space-y-3">
                <h4 className="text-sm font-medium">Resources</h4>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <Label htmlFor="replicas">Replicas</Label>
                    <Input
                      id="replicas"
                      type="number"
                      min="1"
                      value={formData.config?.replicas || 1}
                      onChange={(e) => setFormData({
                        ...formData,
                        config: { ...formData.config, replicas: parseInt(e.target.value) }
                      })}
                    />
                  </div>
                  <div>
                    <Label htmlFor="cpu">CPU Request</Label>
                    <Input
                      id="cpu"
                      value={formData.config?.cpu || '100m'}
                      onChange={(e) => setFormData({
                        ...formData,
                        config: { ...formData.config, cpu: e.target.value }
                      })}
                      placeholder="100m"
                    />
                  </div>
                  <div>
                    <Label htmlFor="memory">Memory Request</Label>
                    <Input
                      id="memory"
                      value={formData.config?.memory || '128Mi'}
                      onChange={(e) => setFormData({
                        ...formData,
                        config: { ...formData.config, memory: e.target.value }
                      })}
                      placeholder="128Mi"
                    />
                  </div>
                </div>
              </div>
            )}

            {formData.type === 'stateful' && (
              <div className="space-y-3">
                <h4 className="text-sm font-medium">Storage</h4>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <Label htmlFor="storage-size">Storage Size</Label>
                    <Input
                      id="storage-size"
                      placeholder="10Gi"
                    />
                  </div>
                  <div>
                    <Label htmlFor="storage-class">Storage Class</Label>
                    <Select>
                      <SelectTrigger id="storage-class">
                        <SelectValue placeholder="Select storage class" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="standard">Standard</SelectItem>
                        <SelectItem value="fast-ssd">Fast SSD</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>
            )}

            {formData.type === 'cronjob' && (
              <div>
                <Label htmlFor="schedule">Cron Schedule</Label>
                <Input
                  id="schedule"
                  value={formData.cron_schedule || ''}
                  onChange={(e) => setFormData({ ...formData, cron_schedule: e.target.value })}
                  placeholder="0 * * * *"
                />
                {errors.cron_schedule && (
                  <p className="text-sm text-destructive mt-1">{errors.cron_schedule}</p>
                )}
              </div>
            )}

            {formData.type === 'function' && (
              <div className="space-y-3">
                <div>
                  <Label htmlFor="runtime">Runtime</Label>
                  <Select
                    value={formData.function_runtime}
                    onValueChange={(value) => setFormData({ ...formData, function_runtime: value })}
                  >
                    <SelectTrigger id="runtime">
                      <SelectValue placeholder="Select runtime" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="nodejs18">Node.js 18</SelectItem>
                      <SelectItem value="python39">Python 3.9</SelectItem>
                      <SelectItem value="go119">Go 1.19</SelectItem>
                    </SelectContent>
                  </Select>
                  {errors.function_runtime && (
                    <p className="text-sm text-destructive mt-1">{errors.function_runtime}</p>
                  )}
                </div>
                <div>
                  <Label htmlFor="handler">Handler</Label>
                  <Input
                    id="handler"
                    value={formData.function_handler || ''}
                    onChange={(e) => setFormData({ ...formData, function_handler: e.target.value })}
                    placeholder="index.handler"
                  />
                  {errors.function_handler && (
                    <p className="text-sm text-destructive mt-1">{errors.function_handler}</p>
                  )}
                </div>
                <div>
                  <Label htmlFor="timeout">Timeout (seconds)</Label>
                  <Input
                    id="timeout"
                    type="number"
                    min="1"
                    max="900"
                    value={formData.function_timeout || 30}
                    onChange={(e) => setFormData({ ...formData, function_timeout: parseInt(e.target.value) })}
                  />
                </div>
              </div>
            )}
          </div>

          <DialogFooter className="mt-6">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Deploying...' : 'Deploy'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

