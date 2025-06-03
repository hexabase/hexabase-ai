'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Loader2, Package, Server } from 'lucide-react';
import { type Workspace } from '@/lib/api-client';

const formSchema = z.object({
  name: z.string()
    .min(3, 'Project name must be at least 3 characters')
    .max(50, 'Project name must be less than 50 characters')
    .regex(/^[a-zA-Z0-9-_\s]+$/, 'Project name can only contain letters, numbers, hyphens, underscores, and spaces'),
  description: z.string()
    .max(200, 'Description must be less than 200 characters')
    .optional(),
  workspace_id: z.string().min(1, 'Please select a workspace'),
});

type FormData = z.infer<typeof formSchema>;

interface CreateProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: FormData) => Promise<void>;
  workspaces: Workspace[];
}

export function CreateProjectDialog({ open, onOpenChange, onSubmit, workspaces }: CreateProjectDialogProps) {
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      workspace_id: '',
    },
  });

  const handleSubmit = async (data: FormData) => {
    try {
      setSubmitting(true);
      await onSubmit(data);
      form.reset();
    } catch {
      // Error handling is done in parent component
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    form.reset();
    onOpenChange(false);
  };

  const selectedWorkspace = workspaces.find(w => w.id === form.watch('workspace_id'));

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]" data-testid="create-project-modal">
        <DialogHeader>
          <DialogTitle>Create New Project</DialogTitle>
          <DialogDescription>
            Create a new project to organize your applications and services within Kubernetes namespaces.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
            {/* Project Name */}
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Project Name</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="Enter project name" 
                      {...field}
                      data-testid="project-name-input"
                    />
                  </FormControl>
                  <FormDescription>
                    Choose a descriptive name for your project
                  </FormDescription>
                  <FormMessage data-testid="name-error" />
                </FormItem>
              )}
            />

            {/* Project Description */}
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description (Optional)</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="Enter project description" 
                      {...field}
                      data-testid="project-description-input"
                    />
                  </FormControl>
                  <FormDescription>
                    Brief description of what this project will contain
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Workspace Selection */}
            <FormField
              control={form.control}
              name="workspace_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Workspace</FormLabel>
                  <FormControl>
                    <Select 
                      onValueChange={field.onChange} 
                      value={field.value}
                      data-testid="workspace-selection"
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select a workspace" />
                      </SelectTrigger>
                      <SelectContent>
                        {workspaces.length === 0 ? (
                          <div className="flex items-center justify-center p-4">
                            <span className="text-sm text-gray-500">No workspaces available</span>
                          </div>
                        ) : (
                          workspaces.map((workspace) => (
                            <SelectItem key={workspace.id} value={workspace.id}>
                              <div className="flex items-center justify-between w-full">
                                <span>{workspace.name}</span>
                                <Badge 
                                  variant={workspace.vcluster_status === 'running' ? 'default' : 'secondary'}
                                  className="ml-2"
                                >
                                  {workspace.vcluster_status}
                                </Badge>
                              </div>
                            </SelectItem>
                          ))
                        )}
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormDescription>
                    Select the workspace where this project will be deployed
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Workspace Details */}
            {selectedWorkspace && (
              <div className="space-y-4">
                <h4 className="text-sm font-medium">Selected Workspace</h4>
                <Card>
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Server className="h-4 w-4 text-blue-500" />
                        <CardTitle className="text-base">{selectedWorkspace.name}</CardTitle>
                      </div>
                      <Badge 
                        variant={selectedWorkspace.vcluster_status === 'running' ? 'default' : 'secondary'}
                      >
                        {selectedWorkspace.vcluster_status}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-2">
                    <div className="space-y-2 text-sm">
                      <div className="flex items-center">
                        <Package className="h-4 w-4 text-green-600 mr-2" />
                        <span>Kubernetes namespace isolation</span>
                      </div>
                      <div className="flex items-center">
                        <Package className="h-4 w-4 text-green-600 mr-2" />
                        <span>Resource quotas and limits</span>
                      </div>
                      <div className="flex items-center">
                        <Package className="h-4 w-4 text-green-600 mr-2" />
                        <span>Network policies and security</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}

            {/* Submit Buttons */}
            <div className="flex justify-end space-x-3">
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={submitting}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={submitting}
                data-testid="submit-project"
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Project
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}