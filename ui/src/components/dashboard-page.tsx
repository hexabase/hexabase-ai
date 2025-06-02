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
      const newOrg = await organizationsApi.create({ name });
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
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center mr-3">
                <svg
                  className="h-5 w-5 text-white"
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
              <h1 className="text-xl font-semibold text-gray-900">Hexabase KaaS</h1>
            </div>

            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                Welcome, {user?.name || user?.email}
              </span>
              <Button
                onClick={logout}
                variant="outline"
                size="sm"
              >
                Sign Out
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <div className="flex justify-between items-center">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">Organizations</h2>
              <p className="mt-1 text-sm text-gray-600">
                Manage your organizations and access your Kubernetes workspaces
              </p>
            </div>
            <Button
              onClick={() => setIsCreateDialogOpen(true)}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              <svg
                className="h-4 w-4 mr-2"
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

        {/* Error Message */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <svg
                className="h-5 w-5 text-red-400"
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
              <div className="ml-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
              <div className="ml-auto pl-3">
                <Button
                  onClick={() => setError(null)}
                  variant="ghost"
                  size="sm"
                  className="text-red-400 hover:text-red-500"
                >
                  ×
                </Button>
              </div>
            </div>
          </div>
        )}

        {/* Organizations List */}
        {isLoading ? (
          <div className="flex justify-center items-center py-12">
            <LoadingSpinner size="lg" className="mr-3" />
            <span className="text-gray-600">Loading organizations...</span>
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
      </main>
    </div>
  );
}