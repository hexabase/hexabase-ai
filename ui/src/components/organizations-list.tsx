"use client";

import { useState } from "react";
import { type Organization } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { formatDateTime } from "@/lib/utils";
import EditOrganizationDialog from "@/components/edit-organization-dialog";

interface OrganizationsListProps {
  organizations: Organization[];
  onDelete: (id: string) => Promise<void>;
  onUpdate: (id: string, name: string) => Promise<void>;
}

export default function OrganizationsList({
  organizations,
  onDelete,
  onUpdate,
}: OrganizationsListProps) {
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleDelete = async (org: Organization) => {
    if (deletingId) return;
    
    const confirmed = window.confirm(
      `Are you sure you want to delete "${org.name}"? This action cannot be undone.`
    );
    
    if (confirmed) {
      setDeletingId(org.id);
      try {
        await onDelete(org.id);
      } finally {
        setDeletingId(null);
      }
    }
  };

  const handleEdit = (org: Organization) => {
    setEditingOrg(org);
  };

  const handleUpdate = async (name: string) => {
    if (!editingOrg) return;
    
    try {
      await onUpdate(editingOrg.id, name);
      setEditingOrg(null);
    } catch (error) {
      throw error; // Re-throw to let dialog handle error
    }
  };

  if (organizations.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
        <div className="mx-auto h-12 w-12 bg-gray-100 rounded-lg flex items-center justify-center mb-4">
          <svg
            className="h-6 w-6 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
            />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">No organizations yet</h3>
        <p className="text-gray-600 mb-4">
          Get started by creating your first organization to manage your Kubernetes workspaces.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {organizations.map((org) => (
        <div
          key={org.id}
          className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow"
        >
          <div className="flex items-center justify-between">
            <div className="flex-1">
              <div className="flex items-center">
                <h3 className="text-lg font-medium text-gray-900">{org.name}</h3>
                {org.role && (
                  <span className="ml-3 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    {org.role}
                  </span>
                )}
              </div>
              <div className="mt-2 text-sm text-gray-600">
                <p>Created: {formatDateTime(org.created_at)}</p>
                {org.updated_at !== org.created_at && (
                  <p>Updated: {formatDateTime(org.updated_at)}</p>
                )}
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Button
                onClick={() => handleEdit(org)}
                variant="outline"
                size="sm"
                disabled={deletingId === org.id}
              >
                <svg
                  className="h-4 w-4 mr-1"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                Edit
              </Button>

              <Button
                onClick={() => handleDelete(org)}
                variant="outline"
                size="sm"
                disabled={deletingId === org.id}
                className="text-red-600 border-red-200 hover:bg-red-50 hover:border-red-300"
              >
                {deletingId === org.id ? (
                  <div className="h-4 w-4 mr-1 animate-spin rounded-full border-2 border-red-300 border-t-red-600" />
                ) : (
                  <svg
                    className="h-4 w-4 mr-1"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                )}
                Delete
              </Button>
            </div>
          </div>

          {/* Organization Stats */}
          <div className="mt-4 grid grid-cols-3 gap-4 pt-4 border-t border-gray-100">
            <div className="text-center">
              <div className="text-2xl font-semibold text-gray-900">0</div>
              <div className="text-sm text-gray-600">Workspaces</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-semibold text-gray-900">0</div>
              <div className="text-sm text-gray-600">Projects</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-semibold text-gray-900">1</div>
              <div className="text-sm text-gray-600">Members</div>
            </div>
          </div>
        </div>
      ))}

      {/* Edit Organization Dialog */}
      {editingOrg && (
        <EditOrganizationDialog
          organization={editingOrg}
          isOpen={!!editingOrg}
          onClose={() => setEditingOrg(null)}
          onUpdate={handleUpdate}
        />
      )}
    </div>
  );
}