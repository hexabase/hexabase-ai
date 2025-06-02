"use client";

import { useState, useEffect } from "react";
import { type Organization } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { LoadingSpinner } from "@/components/ui/loading";

interface EditOrganizationDialogProps {
  organization: Organization;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: (name: string) => Promise<void>;
}

export default function EditOrganizationDialog({
  organization,
  isOpen,
  onClose,
  onUpdate,
}: EditOrganizationDialogProps) {
  const [name, setName] = useState(organization.name);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setName(organization.name);
    setError(null);
  }, [organization.name]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name.trim()) {
      setError("Organization name is required");
      return;
    }

    if (name.trim().length < 3) {
      setError("Organization name must be at least 3 characters");
      return;
    }

    if (name.trim() === organization.name) {
      onClose();
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await onUpdate(name.trim());
      onClose();
    } catch (error: unknown) {
      console.error("Failed to update organization:", error);
      const errorMessage = error && typeof error === 'object' && 'response' in error 
        ? (error as { response?: { data?: { error?: string } } }).response?.data?.error 
        : "Failed to update organization";
      setError(errorMessage || "Failed to update organization");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    if (!isLoading) {
      setName(organization.name);
      setError(null);
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-medium text-gray-900">
              Edit Organization
            </h3>
            <button
              onClick={handleClose}
              disabled={isLoading}
              className="text-gray-400 hover:text-gray-500 disabled:opacity-50"
            >
              <svg
                className="h-6 w-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="px-6 py-4">
            <div className="mb-4">
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Organization Name
              </label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isLoading}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
                placeholder="Enter organization name"
                maxLength={100}
              />
              {error && (
                <p className="mt-2 text-sm text-red-600">{error}</p>
              )}
            </div>

            <div className="text-sm text-gray-600">
              <p>Organization ID: <code className="bg-gray-100 px-1 rounded">{organization.id}</code></p>
            </div>
          </div>

          <div className="px-6 py-4 bg-gray-50 flex justify-end space-x-3 rounded-b-lg">
            <Button
              type="button"
              onClick={handleClose}
              disabled={isLoading}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !name.trim() || name.trim() === organization.name}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              {isLoading ? (
                <>
                  <LoadingSpinner size="sm" className="mr-2" />
                  Updating...
                </>
              ) : (
                "Update Organization"
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}