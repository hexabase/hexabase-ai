"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
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
  const router = useRouter();
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
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-12 text-center">
        <div className="mx-auto h-16 w-16 bg-gradient-to-br from-gray-100 to-gray-200 rounded-2xl flex items-center justify-center mb-6">
          <svg
            className="h-8 w-8 text-gray-500"
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
        <h3 className="text-xl font-semibold text-gray-900 mb-3">No organizations yet</h3>
        <p className="text-base text-gray-600 max-w-md mx-auto">
          Get started by creating your first organization to manage your Kubernetes workspaces.
        </p>
      </div>
    );
  }

  return (
    <div className="grid gap-4 md:gap-6">
      {organizations.map((org) => (
        <div
          key={org.id}
          className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-lg hover:border-gray-200 transition-all duration-200 group"
        >
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex-1">
              <div className="flex items-center flex-wrap gap-3">
                <h3 className="text-xl font-semibold text-gray-900 group-hover:text-primary-600 transition-colors">{org.name}</h3>
                {org.role && (
                  <span className="inline-flex items-center px-3 py-1 rounded-lg text-xs font-medium bg-primary-50 text-primary-700 border border-primary-200">
                    {org.role === 'admin' ? (
                      <svg className="h-3 w-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                    ) : null}
                    {org.role.charAt(0).toUpperCase() + org.role.slice(1)}
                  </span>
                )}
              </div>
              <div className="mt-3 flex flex-wrap gap-4 text-sm text-gray-500">
                <div className="flex items-center">
                  <svg className="h-4 w-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  Created {formatDateTime(org.created_at)}
                </div>
                {org.updated_at !== org.created_at && (
                  <div className="flex items-center">
                    <svg className="h-4 w-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    Updated {formatDateTime(org.updated_at)}
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button
                onClick={() => router.push(`/dashboard/organizations/${org.id}`)}
                className="rounded-lg"
                data-testid={`open-organization-${org.id}`}
              >
                <svg
                  className="h-4 w-4 mr-1.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
                Open
              </Button>

              <Button
                onClick={() => handleEdit(org)}
                variant="secondary"
                size="sm"
                disabled={deletingId === org.id}
                className="rounded-lg"
              >
                <svg
                  className="h-4 w-4 mr-1.5"
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
                variant="ghost"
                size="sm"
                disabled={deletingId === org.id}
                className="text-danger-600 hover:text-danger-700 hover:bg-danger-50 rounded-lg"
              >
                {deletingId === org.id ? (
                  <div className="h-4 w-4 mr-1.5 animate-spin rounded-full border-2 border-danger-300 border-t-danger-600" />
                ) : (
                  <svg
                    className="h-4 w-4 mr-1.5"
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
          <div className="mt-6 grid grid-cols-3 gap-4 pt-6 border-t border-gray-100">
            <div className="text-center group/stat">
              <div className="text-3xl font-bold text-gray-900 group-hover/stat:text-primary-600 transition-colors">0</div>
              <div className="text-sm text-gray-500 font-medium mt-1">Workspaces</div>
            </div>
            <div className="text-center group/stat">
              <div className="text-3xl font-bold text-gray-900 group-hover/stat:text-primary-600 transition-colors">0</div>
              <div className="text-sm text-gray-500 font-medium mt-1">Projects</div>
            </div>
            <div className="text-center group/stat">
              <div className="text-3xl font-bold text-gray-900 group-hover/stat:text-primary-600 transition-colors">1</div>
              <div className="text-sm text-gray-500 font-medium mt-1">Members</div>
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