'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth-context';
import { apiClient, Organization } from '@/lib/api-client';
import { OrganizationCard } from './organization-card';
import { CreateOrganizationDialog } from '../create-organization-dialog';
import { EditOrganizationDialog } from '../edit-organization-dialog';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Plus, RefreshCw, Building2 } from 'lucide-react';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useToast } from '@/hooks/use-toast';

export function OrganizationList() {
  const router = useRouter();
  const { activeOrganization, switchOrganization } = useAuth();
  const { toast } = useToast();
  
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null);
  const [deletingOrgId, setDeletingOrgId] = useState<string | null>(null);

  const fetchOrganizations = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await apiClient.organizations.list();
      setOrganizations(response.data.organizations);
    } catch (err) {
      setError('Failed to load organizations');
      console.error('Error fetching organizations:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchOrganizations();
  }, []);

  const handleOrganizationClick = (orgId: string) => {
    switchOrganization(orgId);
    router.push(`/dashboard/organizations/${orgId}`);
  };

  const handleCreateOrganization = async (name: string) => {
    try {
      const response = await apiClient.organizations.create({ name });
      setOrganizations([...organizations, response.data]);
      setIsCreateOpen(false);
      toast({
        title: 'Organization created',
        description: `${name} has been created successfully.`,
      });
    } catch (err) {
      toast({
        title: 'Error creating organization',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  const handleUpdateOrganization = async (id: string, name: string) => {
    try {
      const response = await apiClient.organizations.update(id, { name });
      setOrganizations(
        organizations.map(org => (org.id === id ? response.data : org))
      );
      setEditingOrg(null);
      toast({
        title: 'Organization updated',
        description: 'The organization has been updated successfully.',
      });
    } catch (err) {
      toast({
        title: 'Error updating organization',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteOrganization = async () => {
    if (!deletingOrgId) return;

    try {
      await apiClient.organizations.delete(deletingOrgId);
      setOrganizations(organizations.filter(org => org.id !== deletingOrgId));
      setDeletingOrgId(null);
      toast({
        title: 'Organization deleted',
        description: 'The organization has been deleted successfully.',
      });
    } catch (err) {
      toast({
        title: 'Error deleting organization',
        description: 'Please try again later.',
        variant: 'destructive',
      });
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-4" data-testid="organizations-skeleton">
        {[1, 2, 3].map(i => (
          <Skeleton key={i} className="h-32 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription className="flex items-center justify-between">
          <span>{error}</span>
          <Button onClick={fetchOrganizations} variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    );
  }

  if (organizations.length === 0) {
    return (
      <div className="text-center py-12">
        <Building2 className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">No organizations found</h3>
        <p className="text-muted-foreground mb-4">
          Create your first organization to get started
        </p>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Organization
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Organizations</h2>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Organization
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {organizations.map(org => (
          <OrganizationCard
            key={org.id}
            organization={org}
            isActive={activeOrganization?.id === org.id}
            onClick={handleOrganizationClick}
            onEdit={org.role === 'admin' ? setEditingOrg : undefined}
            onDelete={org.role === 'admin' ? setDeletingOrgId : undefined}
          />
        ))}
      </div>

      <CreateOrganizationDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onSubmit={handleCreateOrganization}
      />

      {editingOrg && (
        <EditOrganizationDialog
          organization={editingOrg}
          open={!!editingOrg}
          onOpenChange={(open) => !open && setEditingOrg(null)}
          onSubmit={(name) => handleUpdateOrganization(editingOrg.id, name)}
        />
      )}

      <AlertDialog open={!!deletingOrgId} onOpenChange={(open) => !open && setDeletingOrgId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this organization? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDeleteOrganization}>
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}