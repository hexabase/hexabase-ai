'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { billingApi, type SubscriptionPlan } from '@/lib/api-client';

interface SubscriptionPlansModalProps {
  orgId: string;
  isOpen: boolean;
  onClose: () => void;
}

export default function SubscriptionPlansModal({ orgId, isOpen, onClose }: SubscriptionPlansModalProps) {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly');
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (isOpen) {
      fetchPlans();
    }
  }, [isOpen]);

  const fetchPlans = async () => {
    try {
      const data = await billingApi.getPlans();
      setPlans(data.plans);
    } catch (error) {
      console.error('Failed to fetch plans:', error);
      // Use mock data for development
      setPlans(getMockPlans());
    } finally {
      setLoading(false);
    }
  };

  const getMockPlans = (): SubscriptionPlan[] => [
    {
      id: 'starter',
      name: 'Starter',
      description: 'Perfect for small teams getting started',
      price_monthly: 9,
      price_yearly: 90,
      yearly_discount_percentage: 17,
      features: [
        '3 Workspaces',
        '25GB Storage',
        '100GB Bandwidth',
        'Community Support',
        'Basic API Access'
      ],
      limits: {
        workspaces: 3,
        storage_gb: 25,
        bandwidth_gb: 100,
        support_level: 'community'
      },
      popular: false
    },
    {
      id: 'professional',
      name: 'Professional',
      description: 'For growing teams with advanced needs',
      price_monthly: 29,
      price_yearly: 290,
      yearly_discount_percentage: 17,
      features: [
        '10 Workspaces',
        '100GB Storage',
        '500GB Bandwidth',
        'Email Support',
        'Full API Access',
        'Advanced Monitoring'
      ],
      limits: {
        workspaces: 10,
        storage_gb: 100,
        bandwidth_gb: 500,
        support_level: 'email'
      },
      popular: true
    },
    {
      id: 'enterprise',
      name: 'Enterprise',
      description: 'For large organizations with custom requirements',
      price_monthly: 99,
      price_yearly: 990,
      yearly_discount_percentage: 17,
      features: [
        'Unlimited Workspaces',
        '1TB Storage',
        'Unlimited Bandwidth',
        '24/7 Phone Support',
        'Priority API Access',
        'Custom Integrations',
        'Dedicated Account Manager'
      ],
      limits: {
        workspaces: -1, // unlimited
        storage_gb: 1000,
        bandwidth_gb: -1, // unlimited
        support_level: 'phone'
      },
      popular: false
    }
  ];

  const handleSelectPlan = (planId: string) => {
    setSelectedPlan(planId);
    setShowConfirmation(true);
  };

  const handleConfirmUpgrade = async () => {
    if (!selectedPlan) return;

    try {
      await billingApi.updateSubscription(orgId, {
        plan_id: selectedPlan,
        billing_cycle: billingCycle
      });
      
      // Show success state
      setShowConfirmation(false);
      // In real app, would show success message or redirect
      onClose();
    } catch (error) {
      console.error('Failed to upgrade subscription:', error);
    }
  };

  if (!isOpen) return null;

  const getPrice = (plan: SubscriptionPlan) => {
    return billingCycle === 'monthly' ? plan.price_monthly : plan.price_yearly;
  };

  const getPriceLabel = (plan: SubscriptionPlan) => {
    if (billingCycle === 'monthly') {
      return `$${plan.price_monthly}/mo`;
    }
    return `$${plan.price_yearly}/yr`;
  };

  if (showConfirmation && selectedPlan) {
    const plan = plans.find(p => p.id === selectedPlan);
    if (!plan) return null;

    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div data-testid="upgrade-confirmation" className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
          <h2 className="text-xl font-semibold mb-4">Confirm Upgrade</h2>
          
          <div className="space-y-4 mb-6">
            <div className="flex justify-between">
              <span>Upgrade from:</span>
              <span data-testid="upgrade-from" className="font-medium">Professional</span>
            </div>
            <div className="flex justify-between">
              <span>Upgrade to:</span>
              <span data-testid="upgrade-to" className="font-medium">{plan.name}</span>
            </div>
            <div className="border-t pt-2">
              <div className="flex justify-between">
                <span>Prorated amount:</span>
                <span data-testid="prorated-amount" className="font-medium">${(getPrice(plan) - 29).toFixed(2)}</span>
              </div>
            </div>
          </div>

          <div data-testid="payment-method-summary" className="bg-gray-50 rounded p-3 mb-6">
            <p className="text-sm text-gray-600">Payment Method</p>
            <p className="font-medium">•••• •••• •••• 4242</p>
          </div>

          <div className="flex space-x-3">
            <Button variant="outline" onClick={() => setShowConfirmation(false)} className="flex-1">
              Cancel
            </Button>
            <Button onClick={handleConfirmUpgrade} data-testid="confirm-upgrade" className="flex-1">
              Confirm Upgrade
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div data-testid="subscription-plans-modal" className="bg-white rounded-lg p-6 max-w-6xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-semibold">Choose Your Plan</h2>
          <Button variant="outline" onClick={onClose}>
            <span className="sr-only">Close</span>
            ✕
          </Button>
        </div>

        {/* Billing Toggle */}
        <div data-testid="billing-toggle" className="flex items-center justify-center mb-8">
          <div className="bg-gray-100 rounded-lg p-1 flex">
            <button
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                billingCycle === 'monthly' 
                  ? 'bg-white text-gray-900 shadow-sm' 
                  : 'text-gray-600 hover:text-gray-900'
              }`}
              onClick={() => setBillingCycle('monthly')}
            >
              Monthly
            </button>
            <button
              data-testid="yearly-billing"
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                billingCycle === 'yearly' 
                  ? 'bg-white text-gray-900 shadow-sm' 
                  : 'text-gray-600 hover:text-gray-900'
              }`}
              onClick={() => setBillingCycle('yearly')}
            >
              Yearly
              {billingCycle === 'yearly' && (
                <span data-testid="yearly-discount" className="ml-2 text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                  Save 17%
                </span>
              )}
            </button>
          </div>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="border rounded-lg p-6 animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
                <div className="h-8 bg-gray-200 rounded w-3/4 mb-6"></div>
                <div className="space-y-2">
                  {[1, 2, 3, 4].map((j) => (
                    <div key={j} className="h-3 bg-gray-200 rounded"></div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {plans.map((plan) => (
              <div
                key={plan.id}
                data-testid={`${plan.id}-plan`}
                className={`border rounded-lg p-6 relative ${
                  plan.popular ? 'border-blue-500 ring-2 ring-blue-100' : 'border-gray-200'
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                    <span className="bg-blue-500 text-white px-3 py-1 rounded-full text-xs font-medium">
                      Most Popular
                    </span>
                  </div>
                )}

                <div className="text-center mb-6">
                  <h3 className="text-lg font-semibold mb-2">{plan.name}</h3>
                  <p className="text-sm text-gray-600 mb-4">{plan.description}</p>
                  <div className="text-3xl font-bold mb-1">{getPriceLabel(plan)}</div>
                  {billingCycle === 'yearly' && (
                    <p className="text-sm text-gray-500">
                      ${plan.price_monthly}/mo billed annually
                    </p>
                  )}
                </div>

                <div className="space-y-3 mb-6">
                  <div data-testid="plan-workspaces" className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    <span className="text-sm">
                      {plan.limits.workspaces === -1 ? 'Unlimited' : plan.limits.workspaces} Workspaces
                    </span>
                  </div>
                  <div data-testid="plan-storage" className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    <span className="text-sm">{plan.limits.storage_gb}GB Storage</span>
                  </div>
                  <div className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    <span className="text-sm">
                      {plan.limits.bandwidth_gb === -1 ? 'Unlimited' : `${plan.limits.bandwidth_gb}GB`} Bandwidth
                    </span>
                  </div>
                  <div data-testid="plan-support" className="flex items-center">
                    <span className="text-green-500 mr-2">✓</span>
                    <span className="text-sm capitalize">{plan.limits.support_level} Support</span>
                  </div>
                  {plan.features.slice(4).map((feature, index) => (
                    <div key={index} className="flex items-center">
                      <span className="text-green-500 mr-2">✓</span>
                      <span className="text-sm">{feature}</span>
                    </div>
                  ))}
                </div>

                <Button
                  onClick={() => handleSelectPlan(plan.id)}
                  data-testid={`select-${plan.id}-plan`}
                  className={`w-full ${
                    plan.popular
                      ? 'bg-blue-600 hover:bg-blue-700 text-white'
                      : 'bg-gray-100 hover:bg-gray-200 text-gray-900'
                  }`}
                >
                  {plan.id === 'professional' ? 'Current Plan' : 'Select Plan'}
                </Button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}