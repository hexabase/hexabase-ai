'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import BillingOverview from '@/components/billing/billing-overview';
import SubscriptionPlansModal from '@/components/billing/subscription-plans-modal';
import PaymentMethodsModal from '@/components/billing/payment-methods-modal';

interface BillingPageProps {
  params: {
    orgId: string;
  };
}

export default function BillingPage({ params }: BillingPageProps) {
  const [showPlansModal, setShowPlansModal] = useState(false);
  const [showPaymentModal, setShowPaymentModal] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Billing & Subscription</h1>
          <p className="text-gray-600 mt-1">Manage your subscription, payment methods, and billing history</p>
        </div>
      </div>

      <BillingOverview 
        orgId={params.orgId}
        onUpgradePlan={() => setShowPlansModal(true)}
        onManagePayment={() => setShowPaymentModal(true)}
      />

      {showPlansModal && (
        <SubscriptionPlansModal
          orgId={params.orgId}
          isOpen={showPlansModal}
          onClose={() => setShowPlansModal(false)}
        />
      )}

      {showPaymentModal && (
        <PaymentMethodsModal
          orgId={params.orgId}
          isOpen={showPaymentModal}
          onClose={() => setShowPaymentModal(false)}
        />
      )}
    </div>
  );
}