"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { organizationsApi, type Organization } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/ui/loading";
import OrganizationsList from "@/components/organizations-list";
import CreateOrganizationDialog from "@/components/create-organization-dialog";

export default function DashboardPage() {
  const { user, logout } = useAuth();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadOrganizations = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await organizationsApi.list();
      setOrganizations(response.organizations);
    } catch (error) {
      console.error('Failed to load organizations:', error);
      setError('Failed to load organizations. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadOrganizations();
  }, []);

  const handleCreateOrganization = async (name: string) => {
    try {
      const newOrg = await organizationsApi.create({ 
        name, 
        display_name: name // Use the same value for display_name 
      });
      setOrganizations(prev => [...prev, newOrg]);
      setIsCreateDialogOpen(false);
    } catch (error) {
      console.error('Failed to create organization:', error);
      throw error; // Re-throw to let dialog handle error
    }
  };

  const handleDeleteOrganization = async (id: string) => {
    try {
      await organizationsApi.delete(id);
      setOrganizations(prev => prev.filter(org => org.id !== id));
    } catch (error) {
      console.error('Failed to delete organization:', error);
      setError('Failed to delete organization. Please try again.');
    }
  };

  const handleUpdateOrganization = async (id: string, name: string) => {
    try {
      const updatedOrg = await organizationsApi.update(id, { name });
      setOrganizations(prev => 
        prev.map(org => org.id === id ? updatedOrg : org)
      );
    } catch (error) {
      console.error('Failed to update organization:', error);
      throw error; // Re-throw to let component handle error
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <div className="h-10 w-10 bg-gradient-to-br from-primary-500 to-primary-700 rounded-xl flex items-center justify-center mr-3 shadow-lg">
                <svg
                  className="h-6 w-6 text-white"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                  />
                </svg>
              </div>
              <h1 className="text-xl font-bold text-gray-900">Hexabase KaaS</h1>
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
        <div className="max-w-7xl mx-auto">
          {/* Page Header */}
          <div className="mb-8">
            <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 md:p-8">
              <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <div>
                  <h2 className="text-3xl font-bold text-gray-900">Organizations</h2>
                  <p className="mt-2 text-base text-gray-600">
                    Manage your organizations and access your Kubernetes workspaces
                  </p>
                </div>
                <Button
                  onClick={() => setIsCreateDialogOpen(true)}
                  className="rounded-lg shadow-md hover:shadow-lg"
                  size="lg"
                >
                  <svg
                    className="h-5 w-5 mr-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 4v16m8-8H4"
                    />
                  </svg>
                  New Organization
                </Button>
              </div>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-6 bg-danger-50 border border-danger-200 rounded-xl p-4 animate-slide-in">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-danger-500 mt-0.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                </div>
                <div className="ml-3 flex-1">
                  <p className="text-sm font-medium text-danger-800">{error}</p>
                </div>
                <div className="ml-auto pl-3">
                  <button
                    onClick={() => setError(null)}
                    className="inline-flex rounded-lg p-1.5 text-danger-500 hover:bg-danger-100 focus:outline-none focus:ring-2 focus:ring-danger-600 focus:ring-offset-2"
                  >
                    <span className="sr-only">Dismiss</span>
                    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Organizations List */}
          {isLoading ? (
            <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-12">
              <div className="flex flex-col items-center justify-center">
                <LoadingSpinner size="lg" className="mb-4" />
                <span className="text-gray-600 font-medium">Loading organizations...</span>
              </div>
            </div>
          ) : (
            <OrganizationsList
              organizations={organizations}
              onDelete={handleDeleteOrganization}
              onUpdate={handleUpdateOrganization}
            />
          )}

          {/* Create Organization Dialog */}
          <CreateOrganizationDialog
            isOpen={isCreateDialogOpen}
            onClose={() => setIsCreateDialogOpen(false)}
            onCreate={handleCreateOrganization}
          />
        </div>
      </main>
    </div>
  );
}