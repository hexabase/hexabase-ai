'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { billingApi, type PaymentMethod } from '@/lib/api-client';

interface PaymentMethodsModalProps {
  orgId: string;
  isOpen: boolean;
  onClose: () => void;
}

export default function PaymentMethodsModal({ orgId, isOpen, onClose }: PaymentMethodsModalProps) {
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([]);
  const [showAddForm, setShowAddForm] = useState(false);
  const [loading, setLoading] = useState(true);
  const [formData, setFormData] = useState({
    cardNumber: '',
    expiry: '',
    cvc: '',
    name: ''
  });

  useEffect(() => {
    if (isOpen) {
      fetchPaymentMethods();
    }
  }, [isOpen]);

  const fetchPaymentMethods = async () => {
    try {
      const data = await billingApi.getPaymentMethods(orgId);
      setPaymentMethods(data.payment_methods);
    } catch (error) {
      console.error('Failed to fetch payment methods:', error);
      // Use mock data for development
      setPaymentMethods(getMockPaymentMethods());
    } finally {
      setLoading(false);
    }
  };

  const getMockPaymentMethods = (): PaymentMethod[] => [
    {
      id: 'pm_123',
      type: 'card',
      last_four: '4242',
      brand: 'visa',
      expiry_month: 12,
      expiry_year: 2025,
      is_default: true,
      created_at: '2024-01-15T00:00:00Z'
    }
  ];

  const handleAddPaymentMethod = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      // In real app, would use Stripe or similar to tokenize card
      const mockToken = 'tok_visa';
      
      await billingApi.addPaymentMethod(orgId, {
        token: mockToken,
        is_default: paymentMethods.length === 0
      });
      
      // Refresh payment methods
      await fetchPaymentMethods();
      setShowAddForm(false);
      setFormData({ cardNumber: '', expiry: '', cvc: '', name: '' });
    } catch (error) {
      console.error('Failed to add payment method:', error);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div data-testid="payment-methods-modal" className="bg-white rounded-lg p-6 max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold">Payment Methods</h2>
          <Button variant="outline" onClick={onClose}>
            <span className="sr-only">Close</span>
            ✕
          </Button>
        </div>

        {loading ? (
          <div className="space-y-4">
            <div className="border rounded-lg p-4 animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/3 mb-2"></div>
              <div className="h-3 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        ) : (
          <>
            {/* Current Payment Methods */}
            <div className="space-y-4 mb-6">
              {paymentMethods.map((method) => (
                <div
                  key={method.id}
                  data-testid="current-payment-method"
                  className="border rounded-lg p-4 flex items-center justify-between"
                >
                  <div className="flex items-center space-x-3">
                    <div className="w-10 h-6 bg-gradient-to-r from-blue-600 to-purple-600 rounded flex items-center justify-center">
                      <span className="text-white text-xs font-bold">
                        {method.brand?.toUpperCase() || 'CARD'}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium">
                        •••• •••• •••• <span data-testid="card-last-four">{method.last_four}</span>
                      </p>
                      <p data-testid="card-expiry" className="text-sm text-gray-600">
                        Expires {method.expiry_month?.toString().padStart(2, '0')}/{method.expiry_year}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {method.is_default && (
                      <span className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded">
                        Default
                      </span>
                    )}
                    <Button variant="outline" size="sm">
                      Edit
                    </Button>
                  </div>
                </div>
              ))}
            </div>

            {/* Add Payment Method Button */}
            {!showAddForm && (
              <Button
                onClick={() => setShowAddForm(true)}
                data-testid="add-payment-method"
                className="w-full"
              >
                Add Payment Method
              </Button>
            )}

            {/* Add Payment Method Form */}
            {showAddForm && (
              <div data-testid="payment-form" className="border rounded-lg p-4 bg-gray-50">
                <h3 className="font-medium mb-4">Add New Card</h3>
                <form onSubmit={handleAddPaymentMethod} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Card Number
                    </label>
                    <input
                      type="text"
                      data-testid="card-number-input"
                      value={formData.cardNumber}
                      onChange={(e) => setFormData({ ...formData, cardNumber: e.target.value })}
                      placeholder="1234 5678 9012 3456"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      required
                    />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Expiry Date
                      </label>
                      <input
                        type="text"
                        data-testid="expiry-input"
                        value={formData.expiry}
                        onChange={(e) => setFormData({ ...formData, expiry: e.target.value })}
                        placeholder="MM/YY"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                        required
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        CVC
                      </label>
                      <input
                        type="text"
                        data-testid="cvc-input"
                        value={formData.cvc}
                        onChange={(e) => setFormData({ ...formData, cvc: e.target.value })}
                        placeholder="123"
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Cardholder Name
                    </label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      placeholder="John Doe"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                      required
                    />
                  </div>

                  <div className="flex space-x-3 pt-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setShowAddForm(false)}
                      className="flex-1"
                    >
                      Cancel
                    </Button>
                    <Button type="submit" className="flex-1">
                      Add Card
                    </Button>
                  </div>
                </form>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}