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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, Package, Cpu, Database, Users } from 'lucide-react';

const formSchema = z.object({
  name: z.string()
    .min(3, 'Namespace name must be at least 3 characters')
    .max(50, 'Namespace name must be less than 50 characters')
    .regex(/^[a-z0-9-]+$/, 'Namespace name can only contain lowercase letters, numbers, and hyphens'),
  description: z.string()
    .max(200, 'Description must be less than 200 characters')
    .optional(),
  cpu_limit: z.string().optional(),
  memory_limit: z.string().optional(),
  pod_limit: z.string().optional(),
});

type FormData = z.infer<typeof formSchema>;

interface CreateNamespaceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { 
    name: string; 
    description?: string; 
    resource_quota?: { cpu: string; memory: string; pods: number } 
  }) => Promise<void>;
}

export function CreateNamespaceDialog({ open, onOpenChange, onSubmit }: CreateNamespaceDialogProps) {
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      description: '',
      cpu_limit: '1',
      memory_limit: '2',
      pod_limit: '10',
    },
  });

  const handleSubmit = async (data: FormData) => {
    try {
      setSubmitting(true);
      const submitData = {
        name: data.name,
        description: data.description,
        resource_quota: {
          cpu: `${data.cpu_limit} cores`,
          memory: `${data.memory_limit} GB`,
          pods: parseInt(data.pod_limit || '10'),
        },
      };
      await onSubmit(submitData);
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

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]" data-testid="create-namespace-modal">
        <DialogHeader>
          <DialogTitle>Create Namespace</DialogTitle>
          <DialogDescription>
            Create a new Kubernetes namespace to isolate your applications and resources.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
            {/* Namespace Name */}
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Namespace Name</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="e.g., development, staging, production" 
                      {...field}
                      data-testid="namespace-name-input"
                    />
                  </FormControl>
                  <FormDescription>
                    Must be lowercase letters, numbers, and hyphens only
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Description */}
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description (Optional)</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="Brief description of this namespace" 
                      {...field}
                      data-testid="namespace-description-input"
                    />
                  </FormControl>
                  <FormDescription>
                    Help others understand the purpose of this namespace
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Resource Limits */}
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Resource Limits</CardTitle>
                <CardDescription>
                  Set resource quotas to prevent any single namespace from consuming too many cluster resources
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {/* CPU Limit */}
                  <FormField
                    control={form.control}
                    name="cpu_limit"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="flex items-center">
                          <Cpu className="h-4 w-4 mr-2" />
                          CPU Cores
                        </FormLabel>
                        <FormControl>
                          <Input 
                            type="number"
                            min="0.1"
                            step="0.1"
                            placeholder="1"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription className="text-xs">
                          Maximum CPU cores
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Memory Limit */}
                  <FormField
                    control={form.control}
                    name="memory_limit"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="flex items-center">
                          <Database className="h-4 w-4 mr-2" />
                          Memory (GB)
                        </FormLabel>
                        <FormControl>
                          <Input 
                            type="number"
                            min="0.1"
                            step="0.1"
                            placeholder="2"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription className="text-xs">
                          Maximum memory in GB
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {/* Pod Limit */}
                  <FormField
                    control={form.control}
                    name="pod_limit"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel className="flex items-center">
                          <Users className="h-4 w-4 mr-2" />
                          Max Pods
                        </FormLabel>
                        <FormControl>
                          <Input 
                            type="number"
                            min="1"
                            placeholder="10"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription className="text-xs">
                          Maximum pod count
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                {/* Resource Summary */}
                <div className="bg-blue-50 rounded-lg p-4 border border-blue-200">
                  <div className="flex items-center mb-2">
                    <Package className="h-4 w-4 text-blue-600 mr-2" />
                    <span className="text-sm font-medium text-blue-900">Resource Allocation Summary</span>
                  </div>
                  <div className="grid grid-cols-3 gap-4 text-sm">
                    <div>
                      <span className="text-blue-700">CPU:</span>
                      <span className="ml-1 font-medium">{form.watch('cpu_limit') || '1'} cores</span>
                    </div>
                    <div>
                      <span className="text-blue-700">Memory:</span>
                      <span className="ml-1 font-medium">{form.watch('memory_limit') || '2'} GB</span>
                    </div>
                    <div>
                      <span className="text-blue-700">Pods:</span>
                      <span className="ml-1 font-medium">{form.watch('pod_limit') || '10'} max</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Submit Buttons */}
            <div className="flex justify-end space-x-3">
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={submitting}
                data-testid="cancel-namespace"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={submitting}
                data-testid="submit-namespace"
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Namespace
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}