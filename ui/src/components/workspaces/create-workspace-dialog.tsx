'use client';

import { useState, useEffect } from 'react';
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
import { Loader2, Check } from 'lucide-react';
import { plansApi, type Plan } from '@/lib/api-client';

const formSchema = z.object({
  name: z.string()
    .min(3, 'Workspace name must be at least 3 characters')
    .max(50, 'Workspace name must be less than 50 characters')
    .regex(/^[a-zA-Z0-9-_\s]+$/, 'Workspace name can only contain letters, numbers, hyphens, underscores, and spaces'),
  plan_id: z.string().min(1, 'Please select a plan'),
});

type FormData = z.infer<typeof formSchema>;

interface CreateWorkspaceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: FormData) => Promise<void>;
}

export function CreateWorkspaceDialog({ open, onOpenChange, onSubmit }: CreateWorkspaceDialogProps) {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [plansLoading, setPlansLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: '',
      plan_id: '',
    },
  });

  // Load plans when dialog opens
  useEffect(() => {
    if (open) {
      loadPlans();
    }
  }, [open]);

  const loadPlans = async () => {
    try {
      setPlansLoading(true);
      const response = await plansApi.list();
      setPlans(response.plans);
    } catch (error) {
      console.error('Failed to load plans:', error);
      // Set mock plans for testing
      setPlans([
        {
          id: 'plan-basic',
          name: 'Basic Plan',
          description: 'Basic resources for development',
          price: 10.00,
          currency: 'usd',
        },
        {
          id: 'plan-pro',
          name: 'Pro Plan',
          description: 'Enhanced resources for production',
          price: 50.00,
          currency: 'usd',
        },
      ]);
    } finally {
      setPlansLoading(false);
    }
  };

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

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]" data-testid="create-workspace-modal">
        <DialogHeader>
          <DialogTitle>Create New Workspace</DialogTitle>
          <DialogDescription>
            Set up a new Kubernetes workspace with vCluster isolation.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
            {/* Workspace Name */}
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Workspace Name</FormLabel>
                  <FormControl>
                    <Input 
                      placeholder="Enter workspace name" 
                      {...field}
                      data-testid="workspace-name-input"
                    />
                  </FormControl>
                  <FormDescription>
                    Choose a descriptive name for your workspace
                  </FormDescription>
                  <FormMessage data-testid="name-error" />
                </FormItem>
              )}
            />

            {/* Plan Selection */}
            <FormField
              control={form.control}
              name="plan_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Plan</FormLabel>
                  <FormControl>
                    <Select 
                      onValueChange={field.onChange} 
                      value={field.value}
                      data-testid="plan-selection"
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select a plan" />
                      </SelectTrigger>
                      <SelectContent>
                        {plansLoading ? (
                          <div className="flex items-center justify-center p-4">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            <span className="ml-2">Loading plans...</span>
                          </div>
                        ) : (
                          plans.map((plan) => (
                            <SelectItem key={plan.id} value={plan.id}>
                              <div className="flex items-center justify-between w-full">
                                <span>{plan.name}</span>
                                <Badge variant="outline">
                                  ${plan.price}/{plan.currency}
                                </Badge>
                              </div>
                            </SelectItem>
                          ))
                        )}
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormDescription>
                    Choose the resource plan for your workspace
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Plan Details */}
            {form.watch('plan_id') && (
              <div className="space-y-4">
                <h4 className="text-sm font-medium">Plan Details</h4>
                {plans
                  .filter(plan => plan.id === form.watch('plan_id'))
                  .map((plan) => (
                    <Card key={plan.id}>
                      <CardHeader className="pb-2">
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-base">{plan.name}</CardTitle>
                          <Badge>
                            ${plan.price}/{plan.currency}
                          </Badge>
                        </div>
                        <CardDescription>{plan.description}</CardDescription>
                      </CardHeader>
                      <CardContent className="pt-2">
                        <div className="space-y-2 text-sm">
                          <div className="flex items-center">
                            <Check className="h-4 w-4 text-green-600 mr-2" />
                            <span>Dedicated vCluster instance</span>
                          </div>
                          <div className="flex items-center">
                            <Check className="h-4 w-4 text-green-600 mr-2" />
                            <span>Kubernetes namespace isolation</span>
                          </div>
                          <div className="flex items-center">
                            <Check className="h-4 w-4 text-green-600 mr-2" />
                            <span>Resource quotas and limits</span>
                          </div>
                          {plan.id === 'plan-pro' && (
                            <div className="flex items-center">
                              <Check className="h-4 w-4 text-green-600 mr-2" />
                              <span>Enhanced monitoring and alerts</span>
                            </div>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
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
                data-testid="submit-workspace"
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Create Workspace
              </Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}