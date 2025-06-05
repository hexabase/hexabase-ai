'use client';

import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Badge } from '@/components/ui/badge';
import { workspacesApi, plansApi, type Plan, type CreateWorkspaceRequest } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { Loader2, Check, Zap, Server, Cloud } from 'lucide-react';

interface CreateWorkspaceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSuccess?: () => void;
}

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
  orgId,
  onSuccess,
}: CreateWorkspaceDialogProps) {
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [formData, setFormData] = useState<CreateWorkspaceRequest>({
    name: '',
    plan_id: '',
  });

  useEffect(() => {
    if (open) {
      loadPlans();
    }
  }, [open]);

  const loadPlans = async () => {
    try {
      const response = await plansApi.list();
      setPlans(response.plans);
      // Select first plan by default
      if (response.plans.length > 0 && !formData.plan_id) {
        setFormData(prev => ({ ...prev, plan_id: response.plans[0].id }));
      }
    } catch (error) {
      console.error('Failed to load plans:', error);
      // Use mock plans for testing
      const mockPlans: Plan[] = [
        {
          id: 'plan-starter',
          name: 'Starter',
          description: 'Perfect for small projects',
          price: 0,
          currency: 'USD',
          resource_limits: JSON.stringify({
            cpu: '2 cores',
            memory: '4 GB',
            storage: '10 GB'
          })
        },
        {
          id: 'plan-pro',
          name: 'Professional',
          description: 'For production workloads',
          price: 99,
          currency: 'USD',
          resource_limits: JSON.stringify({
            cpu: '8 cores',
            memory: '16 GB',
            storage: '100 GB'
          })
        },
        {
          id: 'plan-enterprise',
          name: 'Enterprise',
          description: 'Unlimited resources with SLA',
          price: 499,
          currency: 'USD',
          resource_limits: JSON.stringify({
            cpu: 'Unlimited',
            memory: 'Unlimited',
            storage: 'Unlimited'
          })
        }
      ];
      setPlans(mockPlans);
      if (!formData.plan_id) {
        setFormData(prev => ({ ...prev, plan_id: mockPlans[0].id }));
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name || !formData.plan_id) {
      toast({
        title: 'Validation Error',
        description: 'Please fill in all required fields',
        variant: 'destructive',
      });
      return;
    }

    try {
      setLoading(true);
      await workspacesApi.create(orgId, formData);
      toast({
        title: 'Success',
        description: 'Workspace created successfully',
      });
      onSuccess?.();
      onOpenChange(false);
      // Reset form
      setFormData({ name: '', plan_id: plans[0]?.id || '' });
    } catch (error) {
      console.error('Failed to create workspace:', error);
      toast({
        title: 'Error',
        description: 'Failed to create workspace. Please try again.',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const getPlanIcon = (planName: string) => {
    switch (planName.toLowerCase()) {
      case 'starter':
        return <Zap className="h-5 w-5" />;
      case 'professional':
        return <Server className="h-5 w-5" />;
      case 'enterprise':
        return <Cloud className="h-5 w-5" />;
      default:
        return <Server className="h-5 w-5" />;
    }
  };

  const parseResourceLimits = (limits?: string) => {
    try {
      return limits ? JSON.parse(limits) : {};
    } catch {
      return {};
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Workspace</DialogTitle>
            <DialogDescription>
              Set up a new Kubernetes workspace with vCluster
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 py-4">
            {/* Workspace Name */}
            <div className="space-y-2">
              <Label htmlFor="name">Workspace Name</Label>
              <Input
                id="name"
                name="name"
                placeholder="e.g., Production, Development"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            {/* Plan Selection */}
            <div className="space-y-2">
              <Label>Select a Plan</Label>
              <RadioGroup 
                value={formData.plan_id} 
                onValueChange={(value) => setFormData({ ...formData, plan_id: value })}
                data-testid="plan-selector"
              >
                <div className="grid gap-4">
                  {plans.map((plan) => {
                    const limits = parseResourceLimits(plan.resource_limits);
                    return (
                      <Card 
                        key={plan.id} 
                        className={`cursor-pointer transition-all ${
                          formData.plan_id === plan.id ? 'ring-2 ring-primary' : ''
                        }`}
                        onClick={() => setFormData({ ...formData, plan_id: plan.id })}
                        data-testid={`plan-card-${plan.id}`}
                      >
                        <CardHeader className="pb-3">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              {getPlanIcon(plan.name)}
                              <div>
                                <CardTitle className="text-lg">{plan.name}</CardTitle>
                                <CardDescription className="text-sm">
                                  {plan.description}
                                </CardDescription>
                              </div>
                            </div>
                            <RadioGroupItem value={plan.id} />
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="flex items-center justify-between">
                            <div className="space-y-1">
                              <div className="flex items-center gap-2 text-sm text-gray-600">
                                <span>CPU: {limits.cpu || 'N/A'}</span>
                                <span>•</span>
                                <span>Memory: {limits.memory || 'N/A'}</span>
                                <span>•</span>
                                <span>Storage: {limits.storage || 'N/A'}</span>
                              </div>
                            </div>
                            <div className="text-right">
                              <div className="text-2xl font-bold">
                                ${plan.price}
                                <span className="text-sm font-normal text-gray-600">/mo</span>
                              </div>
                              {plan.price === 0 && (
                                <Badge variant="secondary" className="mt-1">Free</Badge>
                              )}
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    );
                  })}
                </div>
              </RadioGroup>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  Create Workspace
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}