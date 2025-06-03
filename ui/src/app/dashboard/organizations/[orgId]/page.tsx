'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { organizationsApi, type Organization } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowLeft, Building2, Users, Package, CreditCard } from 'lucide-react';
import { WorkspaceList } from '@/components/workspaces/workspace-list';
import { ProjectList } from '@/components/projects/project-list';
import BillingOverview from '@/components/billing/billing-overview';

export default function OrganizationDashboardPage() {
  const params = useParams();
  const router = useRouter();
  const { user, logout } = useAuth();
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'workspaces' | 'projects' | 'members' | 'billing'>('workspaces');

  const orgId = params.orgId as string;

  const loadOrganization = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const org = await organizationsApi.get(orgId);
      setOrganization(org);
    } catch (error) {
      console.error('Failed to load organization:', error);
      setError('Failed to load organization. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadOrganization();
  }, [orgId, loadOrganization]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <LoadingSpinner size="lg" className="mb-4" />
          <span className="text-gray-600 font-medium">Loading organization...</span>
        </div>
      </div>
    );
  }

  if (error || !organization) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center max-w-md">
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
            <div className="text-red-500 mb-4">
              <svg className="h-12 w-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">Organization Not Found</h2>
            <p className="text-gray-600 mb-6">{error || 'The organization could not be found.'}</p>
            <Button onClick={() => router.push('/dashboard')} variant="outline">
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Dashboard
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Button 
                variant="ghost" 
                onClick={() => router.push('/dashboard')}
                className="mr-4"
                data-testid="back-to-dashboard"
              >
                <ArrowLeft className="w-4 h-4 mr-2" />
                Dashboard
              </Button>
              <div className="h-10 w-10 bg-gradient-to-br from-primary-500 to-primary-700 rounded-xl flex items-center justify-center mr-3 shadow-lg">
                <Building2 className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900" data-testid="organization-name">
                  {organization.name}
                </h1>
                <p className="text-sm text-gray-600">Organization Dashboard</p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              <div className="hidden sm:flex items-center space-x-1 bg-gray-100 rounded-lg px-3 py-1.5">
                <svg className="h-4 w-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                <span className="text-sm font-medium text-gray-700">
                  {user?.name || user?.email}
                </span>
              </div>
              <Button
                onClick={logout}
                variant="secondary"
                size="sm"
                className="rounded-lg"
              >
                <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
                Sign Out
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="px-4 sm:px-6 lg:px-8 py-8">
        <div className="max-w-7xl mx-auto space-y-8">
          {/* Quick Stats */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <Card data-testid="workspaces-stat">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Workspaces</CardTitle>
                <Package className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">3</div>
                <p className="text-xs text-muted-foreground">
                  Active workspaces
                </p>
              </CardContent>
            </Card>

            <Card data-testid="members-stat">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Members</CardTitle>
                <Users className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">12</div>
                <p className="text-xs text-muted-foreground">
                  Organization members
                </p>
              </CardContent>
            </Card>

            <Card data-testid="projects-stat">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Projects</CardTitle>
                <Building2 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">28</div>
                <p className="text-xs text-muted-foreground">
                  Total projects
                </p>
              </CardContent>
            </Card>

            <Card data-testid="billing-stat">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Monthly Cost</CardTitle>
                <CreditCard className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">$186</div>
                <p className="text-xs text-muted-foreground">
                  Current month
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Navigation Tabs */}
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100">
            <div className="border-b border-gray-200">
              <nav className="flex space-x-8 px-6" aria-label="Tabs">
                <button 
                  className={`whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === 'workspaces' 
                      ? 'border-primary-500 text-primary-600' 
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setActiveTab('workspaces')}
                  data-testid="workspaces-tab"
                >
                  Workspaces
                </button>
                <button 
                  className={`whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === 'projects' 
                      ? 'border-primary-500 text-primary-600' 
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setActiveTab('projects')}
                  data-testid="projects-tab"
                >
                  Projects
                </button>
                <button 
                  className={`whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === 'members' 
                      ? 'border-primary-500 text-primary-600' 
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setActiveTab('members')}
                  data-testid="members-tab"
                >
                  Members
                </button>
                <button 
                  className={`whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === 'billing' 
                      ? 'border-primary-500 text-primary-600' 
                      : 'border-transparent text-gray-500 hover:text-gray-700'
                  }`}
                  onClick={() => setActiveTab('billing')}
                  data-testid="billing-tab"
                >
                  Billing
                </button>
              </nav>
            </div>

            {/* Tab Content */}
            <div className="p-6" data-testid={`${activeTab}-section`}>
              {activeTab === 'workspaces' && (
                <>
                  <div className="mb-6">
                    <h3 className="text-lg font-medium text-gray-900 mb-2">Kubernetes Workspaces</h3>
                    <p className="text-sm text-gray-600">
                      Manage your isolated Kubernetes environments with vCluster technology
                    </p>
                  </div>
                  <WorkspaceList organizationId={orgId} />
                </>
              )}

              {activeTab === 'projects' && (
                <>
                  <div className="mb-6">
                    <h3 className="text-lg font-medium text-gray-900 mb-2">Projects</h3>
                    <p className="text-sm text-gray-600">
                      Organize your applications and services within Kubernetes namespaces
                    </p>
                  </div>
                  <ProjectList organizationId={orgId} />
                </>
              )}

              {activeTab === 'members' && (
                <div className="text-center py-12">
                  <Users className="mx-auto h-12 w-12 text-gray-400" />
                  <h3 className="mt-2 text-sm font-medium text-gray-900">Members Management</h3>
                  <p className="mt-1 text-sm text-gray-500">Coming soon - manage organization members and permissions</p>
                </div>
              )}

              {activeTab === 'billing' && (
                <>
                  <div className="mb-6">
                    <h3 className="text-lg font-medium text-gray-900 mb-2">Billing & Subscriptions</h3>
                    <p className="text-sm text-gray-600">
                      Manage your subscription plans, payment methods, and billing history
                    </p>
                  </div>
                  <BillingOverview 
                    orgId={orgId}
                    onUpgradePlan={() => router.push(`/dashboard/organizations/${orgId}/billing`)}
                    onManagePayment={() => router.push(`/dashboard/organizations/${orgId}/billing`)}
                  />
                </>
              )}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}