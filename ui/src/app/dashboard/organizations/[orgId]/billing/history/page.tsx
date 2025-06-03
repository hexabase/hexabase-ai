'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { billingApi, type Invoice } from '@/lib/api-client';
import { formatDistanceToNow, format } from 'date-fns';
import { ArrowLeft, Download, Filter } from 'lucide-react';

interface BillingHistoryPageProps {
  params: {
    orgId: string;
  };
}

export default function BillingHistoryPage({ params }: BillingHistoryPageProps) {
  const router = useRouter();
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [filter, setFilter] = useState<'all' | 'paid' | 'open' | 'overdue'>('all');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchInvoices();
  }, [filter]);

  const fetchInvoices = async () => {
    try {
      const filterStatus = filter === 'all' ? undefined : filter;
      const data = await billingApi.getInvoices(params.orgId, { status: filterStatus });
      setInvoices(data.invoices);
    } catch (error) {
      console.error('Failed to fetch invoices:', error);
      // Use mock data for development
      setInvoices(getMockInvoices());
    } finally {
      setLoading(false);
    }
  };

  const getMockInvoices = (): Invoice[] => [
    {
      id: 'inv_001',
      organization_id: params.orgId,
      subscription_id: 'sub_123',
      amount_due: 29.00,
      amount_paid: 29.00,
      currency: 'usd',
      status: 'paid',
      period_start: '2024-11-01T00:00:00Z',
      period_end: '2024-12-01T00:00:00Z',
      due_date: '2024-12-01T00:00:00Z',
      created_at: '2024-11-01T00:00:00Z',
      line_items: [
        {
          id: 'li_001',
          description: 'Professional Plan - Monthly',
          amount: 29.00,
          quantity: 1,
          unit_price: 29.00,
          period_start: '2024-11-01T00:00:00Z',
          period_end: '2024-12-01T00:00:00Z'
        }
      ]
    },
    {
      id: 'inv_002',
      organization_id: params.orgId,
      subscription_id: 'sub_123',
      amount_due: 29.00,
      amount_paid: 29.00,
      currency: 'usd',
      status: 'paid',
      period_start: '2024-10-01T00:00:00Z',
      period_end: '2024-11-01T00:00:00Z',
      due_date: '2024-11-01T00:00:00Z',
      created_at: '2024-10-01T00:00:00Z',
      line_items: [
        {
          id: 'li_002',
          description: 'Professional Plan - Monthly',
          amount: 29.00,
          quantity: 1,
          unit_price: 29.00,
          period_start: '2024-10-01T00:00:00Z',
          period_end: '2024-11-01T00:00:00Z'
        }
      ]
    },
    {
      id: 'inv_003',
      organization_id: params.orgId,
      subscription_id: 'sub_123',
      amount_due: 29.00,
      amount_paid: 29.00,
      currency: 'usd',
      status: 'paid',
      period_start: '2024-09-01T00:00:00Z',
      period_end: '2024-10-01T00:00:00Z',
      due_date: '2024-10-01T00:00:00Z',
      created_at: '2024-09-01T00:00:00Z',
      line_items: [
        {
          id: 'li_003',
          description: 'Professional Plan - Monthly',
          amount: 29.00,
          quantity: 1,
          unit_price: 29.00,
          period_start: '2024-09-01T00:00:00Z',
          period_end: '2024-10-01T00:00:00Z'
        }
      ]
    }
  ];

  const handleDownloadPdf = async (invoiceId: string) => {
    try {
      const blob = await billingApi.downloadInvoicePdf(params.orgId, invoiceId);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `invoice-${invoiceId}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Failed to download invoice:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'paid':
        return 'bg-green-100 text-green-800';
      case 'open':
        return 'bg-blue-100 text-blue-800';
      case 'overdue':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            onClick={() => router.push(`/dashboard/organizations/${params.orgId}/billing`)}
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Billing
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Billing History</h1>
            <p className="text-gray-600 mt-1">View and download your past invoices</p>
          </div>
        </div>
      </div>

      <div data-testid="billing-history" className="bg-white rounded-lg border">
        {/* Filters */}
        <div className="border-b border-gray-200 p-4">
          <div className="flex items-center space-x-4">
            <Filter className="w-4 h-4 text-gray-500" />
            <select
              data-testid="invoice-filter"
              value={filter}
              onChange={(e) => setFilter(e.target.value as any)}
              className="border border-gray-300 rounded-md px-3 py-1 text-sm"
            >
              <option value="all">All Invoices</option>
              <option value="paid">Paid</option>
              <option value="open">Open</option>
              <option value="overdue">Overdue</option>
            </select>
          </div>
        </div>

        {/* Invoice List */}
        {loading ? (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-2 text-gray-500">Loading invoices...</p>
          </div>
        ) : (
          <div data-testid="invoice-list" className="divide-y divide-gray-200">
            {invoices.map((invoice) => (
              <div key={invoice.id} data-testid="invoice-item" className="p-4 hover:bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-2">
                      <p className="font-medium text-gray-900">
                        Invoice #{invoice.id.slice(-8)}
                      </p>
                      <span
                        data-testid="invoice-status"
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(invoice.status)}`}
                      >
                        {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                      </span>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
                      <div>
                        <p className="font-medium">Amount</p>
                        <p data-testid="invoice-amount" className="text-lg font-semibold text-gray-900">
                          ${invoice.amount_due.toFixed(2)}
                        </p>
                      </div>
                      <div>
                        <p className="font-medium">Date</p>
                        <p data-testid="invoice-date">
                          {format(new Date(invoice.created_at), 'MMM dd, yyyy')}
                        </p>
                      </div>
                      <div>
                        <p className="font-medium">Period</p>
                        <p>
                          {format(new Date(invoice.period_start), 'MMM dd')} - {format(new Date(invoice.period_end), 'MMM dd, yyyy')}
                        </p>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDownloadPdf(invoice.id)}
                      data-testid="download-pdf"
                    >
                      <Download className="w-4 h-4 mr-1" />
                      PDF
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {!loading && invoices.length > 0 && (
          <div data-testid="invoice-pagination" className="border-t border-gray-200 p-4">
            <div className="flex items-center justify-between">
              <p className="text-sm text-gray-700">
                Showing <span className="font-medium">1</span> to <span className="font-medium">{invoices.length}</span> of{' '}
                <span className="font-medium">{invoices.length}</span> results
              </p>
              <div className="flex space-x-2">
                <Button variant="outline" size="sm" disabled>
                  Previous
                </Button>
                <Button variant="outline" size="sm" disabled>
                  Next
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}