'use client';

import { useState } from 'react';
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
import { Card } from '@/components/ui/card';
import { AlertCircle, Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Plan {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  resource_limits?: {
    cpu: string;
    memory: string;
    storage: string;
  };
}

interface CreateWorkspaceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  plans: Plan[];
  onSubmit: (name: string, planId: string) => Promise<void>;
}

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
  plans,
  onSubmit,
}: CreateWorkspaceDialogProps) {
  const [name, setName] = useState('');
  const [selectedPlanId, setSelectedPlanId] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError('Workspace name is required');
      return;
    }
    
    if (!selectedPlanId) {
      setError('Please select a plan');
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      await onSubmit(name.trim(), selectedPlanId);
      setName('');
      setSelectedPlanId('');
    } catch (err) {
      setError('Failed to create workspace. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setName('');
      setSelectedPlanId('');
      setError('');
    }
    onOpenChange(open);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create New Workspace</DialogTitle>
            <DialogDescription>
              Choose a plan and name for your new workspace.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6 py-4">
            <div className="space-y-2">
              <Label htmlFor="workspace-name">Workspace Name</Label>
              <Input
                id="workspace-name"
                placeholder="e.g., Production, Development"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isLoading}
              />
            </div>

            <div className="space-y-3">
              <Label>Select a Plan</Label>
              <RadioGroup value={selectedPlanId} onValueChange={setSelectedPlanId}>
                {plans.map((plan) => (
                  <Card
                    key={plan.id}
                    className={cn(
                      'relative p-4 cursor-pointer hover:shadow-md transition-shadow',
                      selectedPlanId === plan.id && 'ring-2 ring-primary'
                    )}
                    onClick={() => setSelectedPlanId(plan.id)}
                  >
                    <div className="flex items-start space-x-3">
                      <RadioGroupItem
                        value={plan.id}
                        id={plan.id}
                        aria-label={plan.name}
                        className="mt-1"
                      />
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <Label htmlFor={plan.id} className="text-base font-medium cursor-pointer">
                            {plan.name}
                          </Label>
                          <span className="text-lg font-semibold">
                            {plan.price === 0 ? (
                              <span className="text-green-600">$0/month</span>
                            ) : (
                              `$${plan.price}/month`
                            )}
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                          {plan.description}
                        </p>
                        {plan.resource_limits && (
                          <div className="mt-2 text-sm text-muted-foreground">
                            <span className="font-medium">Resources:</span>{' '}
                            {plan.resource_limits.cpu} CPU,{' '}
                            {plan.resource_limits.memory} Memory,{' '}
                            {plan.resource_limits.storage} Storage
                          </div>
                        )}
                      </div>
                      {selectedPlanId === plan.id && (
                        <Check className="h-5 w-5 text-primary" />
                      )}
                    </div>
                  </Card>
                ))}
              </RadioGroup>
            </div>

            {error && (
              <div className="flex items-center gap-2 text-sm text-destructive">
                <AlertCircle className="h-4 w-4" />
                {error}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || !name || !selectedPlanId}>
              {isLoading ? 'Creating...' : 'Create Workspace'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}